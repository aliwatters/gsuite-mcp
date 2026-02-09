// Package auth provides OAuth2 authentication for Google APIs.
// Uses dynamic account discovery - no pre-configuration required.
package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"github.com/aliwatters/gsuite-mcp/internal/config"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// ErrNoCredentials indicates that credentials are missing for an account.
var ErrNoCredentials = errors.New("no credentials")

const (
	// oauthStateTokenSize is the number of random bytes used for CSRF state tokens.
	oauthStateTokenSize = 16

	// oauthCallbackTimeout is the maximum time to wait for OAuth flow completion.
	oauthCallbackTimeout = 5 * time.Minute

	// defaultOAuthPort is the default port for OAuth callback.
	defaultOAuthPort = 8000

	// oauthResultTimeout is the maximum time to wait for the main loop to process the result.
	oauthResultTimeout = 10 * time.Second

	// oauthServerShutdownTimeout is the maximum time to wait for the HTTP server to shut down.
	oauthServerShutdownTimeout = 5 * time.Second

	// credentialFileMode is the file permission for saved credential files.
	credentialFileMode = 0600
)

// authResult carries the result of a successful authentication to be displayed to the user.
type authResult struct {
	email         string
	otherAccounts []string
}

const errorPageTemplate = `<!DOCTYPE html>
<html>
<head><title>Authentication Error</title>
<style>
  * { margin: 0; padding: 0; box-sizing: border-box; }
  body {
    font-family: 'Google Sans', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
    display: flex; justify-content: center; align-items: center;
    min-height: 100vh; background: linear-gradient(135deg, #ef4444 0%%, #dc2626 100%%);
  }
  .card {
    text-align: center; padding: 48px 64px; background: white;
    border-radius: 16px; box-shadow: 0 20px 60px rgba(0,0,0,0.3);
  }
  .icon {
    width: 80px; height: 80px; margin: 0 auto 24px;
    background: linear-gradient(135deg, #ef4444, #dc2626);
    border-radius: 50%%; display: flex; align-items: center; justify-content: center;
    font-size: 40px; color: white;
  }
  h1 { font-size: 24px; font-weight: 500; color: #1a1a1a; margin-bottom: 12px; }
  .message { font-size: 14px; color: #666; margin-bottom: 16px; }
  .hint { font-size: 13px; color: #999; }
</style>
</head>
<body>
<div class="card">
  <div class="icon">✕</div>
  <h1>%s</h1>
  <p class="message">%s</p>
  <p class="hint">Please try again. You can close this window.</p>
</div>
</body>
</html>`

const successPageTemplate = `<!DOCTYPE html>
<html>
<head>
<title>Authentication Successful</title>
<style>
  * { margin: 0; padding: 0; box-sizing: border-box; }
  body {
    font-family: 'Google Sans', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
    display: flex; justify-content: center; align-items: center;
    min-height: 100vh; background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
  }
  .card {
    text-align: center; padding: 48px 64px; background: white;
    border-radius: 16px; box-shadow: 0 20px 60px rgba(0,0,0,0.3);
    animation: slideUp 0.5s ease-out;
  }
  @keyframes slideUp {
    from { opacity: 0; transform: translateY(20px); }
    to { opacity: 1; transform: translateY(0); }
  }
  .checkmark {
    width: 80px; height: 80px; margin: 0 auto 24px;
    background: linear-gradient(135deg, #22c55e, #16a34a);
    border-radius: 50%%; display: flex; align-items: center; justify-content: center;
    animation: pop 0.4s ease-out 0.2s both;
  }
  @keyframes pop {
    0%% { transform: scale(0); }
    80%% { transform: scale(1.1); }
    100%% { transform: scale(1); }
  }
  .checkmark svg { width: 40px; height: 40px; stroke: white; stroke-width: 3; fill: none; }
  .checkmark svg path { stroke-dasharray: 50; stroke-dashoffset: 50; animation: draw 0.5s ease-out 0.5s forwards; }
  @keyframes draw { to { stroke-dashoffset: 0; } }
  h1 { font-size: 28px; font-weight: 500; color: #1a1a1a; margin-bottom: 8px; }
  .saved { font-size: 14px; color: #888; margin-bottom: 8px; }
  .email { font-size: 20px; color: #4285f4; font-weight: 500; margin-bottom: 20px; }
  .others { margin-top: 20px; padding-top: 20px; border-top: 1px solid #eee; }
  .others-label { font-size: 12px; color: #888; margin-bottom: 8px; text-transform: uppercase; letter-spacing: 0.5px; }
  .other-email { font-size: 14px; color: #666; margin: 4px 0; }
  .hint { font-size: 14px; color: #999; margin-top: 20px; }
</style>
</head>
<body>
<div class="card">
  <div class="checkmark"><svg viewBox="0 0 24 24"><path d="M5 13l4 4L19 7"/></svg></div>
  <h1>Authentication Successful</h1>
  <p class="saved">Saved token for</p>
  <p class="email">%s</p>
  %s
  <p class="hint">You can close this window</p>
</div>
</body>
</html>`

// DefaultScopes are the OAuth scopes required for Gmail, Calendar, Docs, Tasks, Drive, Sheets, and Contacts operations.
var DefaultScopes = []string{
	// OpenID Connect scopes (required for getting authenticated user email)
	"openid",
	"https://www.googleapis.com/auth/userinfo.email",
	// Gmail scopes
	"https://www.googleapis.com/auth/gmail.modify",
	"https://www.googleapis.com/auth/gmail.compose",
	"https://www.googleapis.com/auth/gmail.labels",
	"https://www.googleapis.com/auth/gmail.settings.basic",
	// Calendar scopes
	"https://www.googleapis.com/auth/calendar",
	"https://www.googleapis.com/auth/calendar.events",
	// Docs scopes
	"https://www.googleapis.com/auth/documents",
	// Tasks scopes
	"https://www.googleapis.com/auth/tasks",
	// Drive scopes
	"https://www.googleapis.com/auth/drive",
	// Sheets scopes
	"https://www.googleapis.com/auth/spreadsheets",
	// Contacts scopes (People API)
	"https://www.googleapis.com/auth/contacts",
}

// Token represents a stored OAuth token with metadata.
type Token struct {
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	TokenURI     string    `json:"token_uri"`
	ClientID     string    `json:"client_id"`
	ClientSecret string    `json:"client_secret"`
	Scopes       []string  `json:"scopes"`
	Expiry       time.Time `json:"expiry"`
}

// Manager handles OAuth2 authentication and token management.
type Manager struct {
	oauthConfig *oauth2.Config
	// authMu prevents concurrent authentication attempts
	authMu sync.Mutex
}

// NewManager creates a new auth manager.
func NewManager() (*Manager, error) {
	oauthCfg, err := loadOAuthConfig()
	if err != nil {
		return nil, err
	}

	return &Manager{
		oauthConfig: oauthCfg,
	}, nil
}

// getAuthenticatedEmail fetches the email address of the authenticated user.
func getAuthenticatedEmail(ctx context.Context, client *http.Client) (string, error) {
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return "", fmt.Errorf("fetching userinfo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("userinfo API error: %s - %s", resp.Status, string(body))
	}

	var userinfo struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userinfo); err != nil {
		return "", fmt.Errorf("parsing userinfo: %w", err)
	}

	return userinfo.Email, nil
}

// loadOAuthConfig loads the OAuth2 configuration from client_secret.json.
func loadOAuthConfig() (*oauth2.Config, error) {
	path := config.ClientSecretPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("client_secret.json not found at %s - download from Google Cloud Console", path)
		}
		return nil, fmt.Errorf("reading client secret: %w", err)
	}

	cfg, err := google.ConfigFromJSON(data, DefaultScopes...)
	if err != nil {
		return nil, fmt.Errorf("parsing client secret: %w", err)
	}

	return cfg, nil
}

// AuthenticateDynamic performs OAuth2 flow without requiring a pre-configured account.
// It opens the browser, lets the user choose any Google account, and saves the credential
// using the authenticated email as the identifier. Returns the authenticated email.
func (m *Manager) AuthenticateDynamic(ctx context.Context) (string, error) {
	oauthPort := defaultOAuthPort
	if portStr := os.Getenv("GSUITE_MCP_OAUTH_PORT"); portStr != "" {
		if p, err := fmt.Sscanf(portStr, "%d", &oauthPort); err != nil || p != 1 {
			return "", fmt.Errorf("invalid GSUITE_MCP_OAUTH_PORT value: %s", portStr)
		}
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", oauthPort))
	if err != nil {
		return "", fmt.Errorf("failed to listen on port %d: %w", oauthPort, err)
	}
	defer listener.Close()

	oauthCfg := *m.oauthConfig
	oauthCfg.RedirectURL = fmt.Sprintf("http://localhost:%d/oauth2callback", oauthPort)

	stateBytes := make([]byte, oauthStateTokenSize)
	if _, err := rand.Read(stateBytes); err != nil {
		return "", fmt.Errorf("failed to generate state token: %w", err)
	}
	state := hex.EncodeToString(stateBytes)

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)
	resultCh := make(chan authResult, 1)

	server := startOAuthServer(listener, state, codeCh, errCh, resultCh)
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), oauthServerShutdownTimeout)
		defer cancel()
		server.Shutdown(shutdownCtx)
	}()

	authURL := oauthCfg.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	printAuthInstructions(authURL)

	if err := openBrowser(authURL); err != nil {
		fmt.Fprintf(os.Stderr, "warning: couldn't open browser automatically: %v\n", err)
	}

	return m.waitForAuthResult(ctx, codeCh, errCh, &oauthCfg, resultCh)
}

// saveTokenForEmail saves an oauth2.Token using email as the identifier.
func (m *Manager) saveTokenForEmail(email string, oauth2Token *oauth2.Token) error {
	if err := config.EnsureConfigDir(); err != nil {
		return err
	}

	token := &Token{
		Token:        oauth2Token.AccessToken,
		RefreshToken: oauth2Token.RefreshToken,
		TokenURI:     m.oauthConfig.Endpoint.TokenURL,
		ClientID:     m.oauthConfig.ClientID,
		ClientSecret: m.oauthConfig.ClientSecret,
		Scopes:       m.oauthConfig.Scopes,
		Expiry:       oauth2Token.Expiry,
	}

	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling token: %w", err)
	}

	path := config.CredentialPathForEmail(email)
	if err := os.WriteFile(path, data, credentialFileMode); err != nil {
		return fmt.Errorf("writing token: %w", err)
	}

	return nil
}

// GetClientForEmail returns an authenticated HTTP client for the given email.
func (m *Manager) GetClientForEmail(ctx context.Context, email string) (*http.Client, error) {
	token, err := loadTokenForEmail(email)
	if err != nil {
		return nil, err
	}

	oauth2Token := &oauth2.Token{
		AccessToken:  token.Token,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
		TokenType:    "Bearer",
	}

	tokenSource := m.oauthConfig.TokenSource(ctx, oauth2Token)

	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("refreshing token for %s: %w", email, err)
	}

	if newToken.AccessToken != oauth2Token.AccessToken {
		if err := m.saveTokenForEmail(email, newToken); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to save refreshed token for %s: %v\n", email, err)
		}
	}

	return oauth2.NewClient(ctx, tokenSource), nil
}

// loadTokenForEmail loads the stored token for an email address.
func loadTokenForEmail(email string) (*Token, error) {
	path := config.CredentialPathForEmail(email)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w for %s", ErrNoCredentials, email)
		}
		return nil, fmt.Errorf("reading credentials for %s: %w", email, err)
	}

	var token Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("parsing credentials for %s: %w", email, err)
	}

	return &token, nil
}

// GetClientOrAuthenticate returns an authenticated HTTP client for the given email.
// If no email is specified, uses the first available authenticated account.
// If no credentials exist and we're in interactive mode, triggers OAuth flow.
// If no credentials exist and we're in MCP server mode, returns an error with instructions.
func (m *Manager) GetClientOrAuthenticate(ctx context.Context, email string, interactive bool) (*http.Client, error) {
	// If email specified, try to get client for that email
	if email != "" {
		client, err := m.GetClientForEmail(ctx, email)
		if err == nil {
			return client, nil
		}
		if !errors.Is(err, ErrNoCredentials) {
			return nil, err
		}
		// No credentials for specified email
		if !interactive {
			return nil, fmt.Errorf("no credentials for %s; run 'gsuite-mcp auth' and sign in with that account", email)
		}
		// Trigger authentication
		m.authMu.Lock()
		defer m.authMu.Unlock()
		_, err = m.AuthenticateDynamic(ctx)
		if err != nil {
			return nil, fmt.Errorf("authentication failed: %w", err)
		}
		// Check if user authenticated with requested email
		if config.HasCredentialsForEmail(email) {
			return m.GetClientForEmail(ctx, email)
		}
		return nil, fmt.Errorf("authenticated with different account; need credentials for %s", email)
	}

	// No email specified - try to use first available account
	defaultEmail := config.GetDefaultEmail()
	if defaultEmail != "" {
		return m.GetClientForEmail(ctx, defaultEmail)
	}

	// No authenticated accounts
	if !interactive {
		return nil, fmt.Errorf("no authenticated accounts; run 'gsuite-mcp auth' to authenticate")
	}

	// Trigger authentication
	m.authMu.Lock()
	defer m.authMu.Unlock()
	authenticatedEmail, err := m.AuthenticateDynamic(ctx)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}
	return m.GetClientForEmail(ctx, authenticatedEmail)
}

// getOtherAuthenticatedEmails returns a list of all authenticated email accounts except the given one.
func getOtherAuthenticatedEmails(excludeEmail string) []string {
	emails, err := config.GetAuthenticatedEmails()
	if err != nil {
		return nil
	}

	var others []string
	for _, email := range emails {
		if email != excludeEmail {
			others = append(others, email)
		}
	}
	return others
}

// openBrowser opens a URL in the default browser.
func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return errors.New("unsupported platform")
	}

	return cmd.Start()
}

// sendOAuthError sends a formatted HTML error page to the user's browser.
func sendOAuthError(w http.ResponseWriter, title, message string) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprintf(w, errorPageTemplate, title, message)
}

// handleOAuthCallback returns a handler function for the OAuth2 callback.
func handleOAuthCallback(state string, codeCh chan<- string, errCh chan<- error, resultCh <-chan authResult) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != state {
			errCh <- errors.New("invalid state parameter - possible CSRF attack")
			sendOAuthError(w, "Authentication Error", "Invalid or expired OAuth state parameter")
			return
		}

		if errMsg := r.URL.Query().Get("error"); errMsg != "" {
			errDesc := r.URL.Query().Get("error_description")
			errCh <- fmt.Errorf("OAuth error: %s - %s", errMsg, errDesc)
			sendOAuthError(w, "Authentication Failed", errDesc)
			return
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			errCh <- errors.New("no authorization code received")
			sendOAuthError(w, "Authentication Error", "No authorization code received from Google")
			return
		}

		codeCh <- code

		select {
		case result := <-resultCh:
			var otherAccountsHTML string
			if len(result.otherAccounts) > 0 {
				otherAccountsHTML = `<div class="others"><p class="others-label">Other authenticated accounts:</p>`
				for _, email := range result.otherAccounts {
					otherAccountsHTML += fmt.Sprintf(`<p class="other-email">%s</p>`, email)
				}
				otherAccountsHTML += `</div>`
			}

			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintf(w, successPageTemplate, result.email, otherAccountsHTML)

		case <-time.After(oauthResultTimeout):
			sendOAuthError(w, "Timeout", "Token exchange took too long")
		}
	}
}

// startOAuthServer starts a temporary HTTP server to handle the OAuth2 callback.
func startOAuthServer(listener net.Listener, state string, codeCh chan<- string, errCh chan<- error, resultCh <-chan authResult) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/oauth2callback", handleOAuthCallback(state, codeCh, errCh, resultCh))

	server := &http.Server{Handler: mux}

	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("callback server error: %w", err)
		}
	}()

	return server
}

// waitForAuthResult waits for the OAuth2 code and exchanges it for a token.
func (m *Manager) waitForAuthResult(ctx context.Context, codeCh <-chan string, errCh <-chan error, oauthCfg *oauth2.Config, resultCh chan<- authResult) (string, error) {
	select {
	case code := <-codeCh:
		token, err := oauthCfg.Exchange(ctx, code)
		if err != nil {
			resultCh <- authResult{email: "", otherAccounts: nil}
			return "", fmt.Errorf("failed to exchange code for token: %w", err)
		}

		client := oauthCfg.Client(ctx, token)
		actualEmail, err := getAuthenticatedEmail(ctx, client)
		if err != nil {
			resultCh <- authResult{email: "", otherAccounts: nil}
			return "", fmt.Errorf("failed to get authenticated email: %w", err)
		}

		if err := m.saveTokenForEmail(actualEmail, token); err != nil {
			resultCh <- authResult{email: actualEmail, otherAccounts: nil}
			return "", fmt.Errorf("failed to save token: %w", err)
		}

		otherAccounts := getOtherAuthenticatedEmails(actualEmail)
		resultCh <- authResult{email: actualEmail, otherAccounts: otherAccounts}

		fmt.Fprintf(os.Stderr, "✓ Successfully authenticated: %s\n", actualEmail)
		return actualEmail, nil

	case err := <-errCh:
		return "", err

	case <-time.After(oauthCallbackTimeout):
		return "", fmt.Errorf("authentication timed out after %v - please try again", oauthCallbackTimeout)

	case <-ctx.Done():
		return "", ctx.Err()
	}
}

// printAuthInstructions prints the authentication URL and instructions to stderr.
func printAuthInstructions(authURL string) {
	fmt.Fprintf(os.Stderr, "\n=== Authentication Required ===\n")
	fmt.Fprintf(os.Stderr, "Opening browser to authenticate with Google...\n")
	fmt.Fprintf(os.Stderr, "Choose any Google account you want to use.\n")
	fmt.Fprintf(os.Stderr, "If browser doesn't open, visit:\n%s\n", authURL)
	fmt.Fprintf(os.Stderr, "================================\n\n")
}
