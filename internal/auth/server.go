package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/oauth2"
)

const (
	// stateExpiry is how long an OAuth state token remains valid.
	stateExpiry = 5 * time.Minute

	// cleanupInterval is how often expired states are purged.
	cleanupInterval = 1 * time.Minute
)

// pendingState tracks a single in-flight OAuth request.
type pendingState struct {
	loginHint string
	expiresAt time.Time
}

// AuthServer is a persistent HTTP server that provides browser-based re-authentication
// alongside the stdio MCP server. Agents can direct users to /auth to trigger OAuth.
type AuthServer struct {
	manager      *Manager
	port         int
	server       *http.Server
	listener     net.Listener
	states       sync.Map
	stopCleanup  chan struct{}
	shutdownOnce sync.Once
}

// NewAuthServer creates an auth server bound to the given manager and port.
func NewAuthServer(manager *Manager, port int) *AuthServer {
	return &AuthServer{
		manager:     manager,
		port:        port,
		stopCleanup: make(chan struct{}),
	}
}

// Start binds the port and begins serving /auth and /oauth2callback.
// Returns an error if the port cannot be bound.
func (s *AuthServer) Start() error {
	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", s.port))
	if err != nil {
		return fmt.Errorf("binding port %d: %w", s.port, err)
	}
	s.listener = listener

	mux := http.NewServeMux()
	mux.HandleFunc("/auth", s.handleAuth)
	mux.HandleFunc("/oauth2callback", s.handleCallback)

	s.server = &http.Server{Handler: mux}

	go s.server.Serve(listener) //nolint:errcheck // logged by caller; ErrServerClosed expected on shutdown

	go s.cleanupExpiredStates()

	return nil
}

// Shutdown gracefully stops the auth server. Safe to call multiple times.
func (s *AuthServer) Shutdown(ctx context.Context) error {
	s.shutdownOnce.Do(func() { close(s.stopCleanup) })
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}

// handleAuth generates a CSRF state, stores it, and redirects to Google OAuth.
func (s *AuthServer) handleAuth(w http.ResponseWriter, r *http.Request) {
	stateBytes := make([]byte, oauthStateTokenSize)
	if _, err := rand.Read(stateBytes); err != nil {
		sendOAuthError(w, "Server Error", "Failed to generate security token")
		return
	}
	state := hex.EncodeToString(stateBytes)

	loginHint := r.URL.Query().Get("account")

	s.states.Store(state, pendingState{
		loginHint: loginHint,
		expiresAt: time.Now().Add(stateExpiry),
	})

	oauthCfg := s.oauthConfigWithRedirect()

	opts := []oauth2.AuthCodeOption{
		oauth2.AccessTypeOffline,
		oauth2.ApprovalForce,
	}
	if loginHint != "" {
		opts = append(opts, oauth2.SetAuthURLParam("login_hint", loginHint))
	}

	authURL := oauthCfg.AuthCodeURL(state, opts...)
	http.Redirect(w, r, authURL, http.StatusFound)
}

// handleCallback validates the state, exchanges the code, saves the token, and renders the result.
func (s *AuthServer) handleCallback(w http.ResponseWriter, r *http.Request) {
	// Check for Google-reported errors first
	if errMsg := r.URL.Query().Get("error"); errMsg != "" {
		errDesc := r.URL.Query().Get("error_description")
		if errDesc == "" {
			errDesc = errMsg
		}
		sendOAuthError(w, "Authentication Failed", errDesc)
		return
	}

	state := r.URL.Query().Get("state")
	if state == "" {
		sendOAuthError(w, "Authentication Error", "Missing state parameter")
		return
	}

	// LoadAndDelete ensures each state token is used exactly once
	val, ok := s.states.LoadAndDelete(state)
	if !ok {
		sendOAuthError(w, "Authentication Error", "Invalid or expired state token. Please start again from /auth")
		return
	}

	ps := val.(pendingState)
	if time.Now().After(ps.expiresAt) {
		sendOAuthError(w, "Authentication Error", "This authentication link has expired. Please start again from /auth")
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		sendOAuthError(w, "Authentication Error", "No authorization code received from Google")
		return
	}

	oauthCfg := s.oauthConfigWithRedirect()

	token, err := oauthCfg.Exchange(r.Context(), code)
	if err != nil {
		sendOAuthError(w, "Authentication Error", fmt.Sprintf("Failed to exchange code: %v", err))
		return
	}

	client := oauthCfg.Client(r.Context(), token)
	email, err := getAuthenticatedEmail(r.Context(), client)
	if err != nil {
		sendOAuthError(w, "Authentication Error", fmt.Sprintf("Failed to get account email: %v", err))
		return
	}

	if err := s.manager.saveTokenForEmail(email, token); err != nil {
		sendOAuthError(w, "Authentication Error", fmt.Sprintf("Failed to save token: %v", err))
		return
	}

	otherAccounts := getOtherAuthenticatedEmails(email)

	w.Header().Set("Content-Type", "text/html")
	if err := successTmpl.Execute(w, successPageData{
		Email:         email,
		OtherAccounts: otherAccounts,
	}); err != nil {
		http.Error(w, "Failed to render success page", http.StatusInternalServerError)
	}
}

// oauthConfigWithRedirect returns a copy of the OAuth config with the redirect URL set to this server.
func (s *AuthServer) oauthConfigWithRedirect() oauth2.Config {
	cfg := *s.manager.oauthConfig
	cfg.RedirectURL = fmt.Sprintf("http://localhost:%d/oauth2callback", s.port)
	return cfg
}

// cleanupExpiredStates periodically removes expired state tokens.
func (s *AuthServer) cleanupExpiredStates() {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			s.states.Range(func(key, value any) bool {
				ps := value.(pendingState)
				if now.After(ps.expiresAt) {
					s.states.Delete(key)
				}
				return true
			})
		case <-s.stopCleanup:
			return
		}
	}
}
