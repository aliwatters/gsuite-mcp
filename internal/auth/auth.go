// Package auth provides OAuth2 authentication for Google APIs.
// Uses dynamic account discovery - no pre-configuration required.
package auth

import (
	"context"
	"crypto/rand"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aliwatters/gsuite-mcp/internal/config"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

//go:embed templates/error.html
var errorTemplateFS embed.FS

//go:embed templates/success.html
var successTemplateFS embed.FS

var (
	errorTmpl   = template.Must(template.ParseFS(errorTemplateFS, "templates/error.html"))
	successTmpl = template.Must(template.ParseFS(successTemplateFS, "templates/success.html"))
)

// ErrNoCredentials indicates that credentials are missing for an account.
var ErrNoCredentials = errors.New("no credentials")

// ErrUnverifiedApp indicates that the OAuth app has not been verified by Google and the
// Google Workspace admin has blocked unverified apps for the domain.
var ErrUnverifiedApp = errors.New("unverified app blocked by Google Workspace admin")

const (
	// oauthStateTokenSize is the number of random bytes used for CSRF state tokens.
	oauthStateTokenSize = 16

	// oauthCallbackTimeout is the maximum time to wait for OAuth flow completion.
	oauthCallbackTimeout = 5 * time.Minute

	// oauthResultTimeout is the maximum time to wait for the main loop to process the result.
	oauthResultTimeout = 10 * time.Second

	// oauthServerShutdownTimeout is the maximum time to wait for the HTTP server to shut down.
	oauthServerShutdownTimeout = 5 * time.Second

	// credentialFileMode is the file permission for saved credential files.
	credentialFileMode = 0600
)

// googleUserInfoURL is the Google OAuth2 userinfo endpoint.
const googleUserInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"

// authResult carries the result of an authentication attempt to be displayed to the user.
type authResult struct {
	email         string
	otherAccounts []string
	err           error
}

// errorPageData holds the template data for the error page.
type errorPageData struct {
	Title   string
	Message string
}

// successPageData holds the template data for the success page.
type successPageData struct {
	Email         string
	OtherAccounts []string
}

// ScopesByService documents the OAuth scopes required per Google service.
// All scopes are requested together in a single consent screen — this map exists
// for documentation and debugging, not for selective scoping.
//
// | Service      | Scope(s)                                                   | Why                                      |
// |--------------|-----------------------------------------------------------|------------------------------------------|
// | Identity     | openid, userinfo.email                                    | Determine the authenticated email        |
// | Gmail        | gmail.modify, gmail.compose, gmail.labels                 | Read/write messages, manage labels       |
// |              | gmail.settings.basic                                      | Manage filters, send-as, vacation        |
// | Calendar     | calendar, calendar.events                                 | Full calendar and event access           |
// | Drive        | drive                                                     | Read/write files and metadata            |
// | Docs         | documents                                                 | Read/write Google Docs                   |
// | Sheets       | spreadsheets                                              | Read/write Google Sheets                 |
// | Slides       | presentations                                             | Read/write Google Slides                 |
// | Forms        | forms.body, forms.responses.readonly                      | Manage forms and read responses          |
// | Tasks        | tasks                                                     | Read/write task lists and tasks          |
// | Contacts     | contacts                                                  | Read/write Google Contacts               |
var ScopesByService = map[string][]string{
	"identity": {
		"openid",
		"https://www.googleapis.com/auth/userinfo.email",
	},
	"gmail": {
		"https://www.googleapis.com/auth/gmail.modify",
		"https://www.googleapis.com/auth/gmail.compose",
		"https://www.googleapis.com/auth/gmail.labels",
		"https://www.googleapis.com/auth/gmail.settings.basic",
	},
	"calendar": {
		"https://www.googleapis.com/auth/calendar",
		"https://www.googleapis.com/auth/calendar.events",
	},
	"drive": {
		"https://www.googleapis.com/auth/drive",
	},
	"docs": {
		"https://www.googleapis.com/auth/documents",
	},
	"sheets": {
		"https://www.googleapis.com/auth/spreadsheets",
	},
	"slides": {
		"https://www.googleapis.com/auth/presentations",
	},
	"forms": {
		"https://www.googleapis.com/auth/forms.body",
		"https://www.googleapis.com/auth/forms.responses.readonly",
	},
	"tasks": {
		"https://www.googleapis.com/auth/tasks",
	},
	"contacts": {
		"https://www.googleapis.com/auth/contacts",
	},
}

// DefaultScopes aggregates all service scopes for the single-consent OAuth flow.
// All scopes are requested together to avoid multiple consent screens as users add services.
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
	// Drive scopes (broad access needed for citation feature and file operations)
	"https://www.googleapis.com/auth/drive",
	// Sheets scopes
	"https://www.googleapis.com/auth/spreadsheets",
	// Slides scopes
	"https://www.googleapis.com/auth/presentations",
	// Forms scopes
	"https://www.googleapis.com/auth/forms.body",
	"https://www.googleapis.com/auth/forms.responses.readonly",
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
	// AuthServerURL is set when the HTTP auth server is running (e.g. "http://localhost:38917/auth").
	AuthServerURL string
}

// OAuthConfig returns the OAuth2 configuration (needed for reading ClientID).
func (m *Manager) OAuthConfig() *oauth2.Config {
	return m.oauthConfig
}

// NewManager creates a new auth manager.
func NewManager() (*Manager, error) {
	oauthCfg, err := loadOAuthConfig()
	if err != nil {
		return nil, fmt.Errorf("loading OAuth config: %w", err)
	}

	return &Manager{
		oauthConfig: oauthCfg,
	}, nil
}

// getAuthenticatedEmail fetches the email address of the authenticated user.
func getAuthenticatedEmail(ctx context.Context, client *http.Client) (string, error) {
	resp, err := client.Get(googleUserInfoURL)
	if err != nil {
		return "", fmt.Errorf("fetching userinfo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return "", fmt.Errorf("reading userinfo response: %w", readErr)
		}
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

// resolveOAuthPort determines the OAuth callback port.
// Resolution order: GSUITE_MCP_OAUTH_PORT env var → config.json → default (38917).
func resolveOAuthPort() (int, error) {
	if portStr := os.Getenv("GSUITE_MCP_OAUTH_PORT"); portStr != "" {
		port, err := strconv.Atoi(portStr)
		if err != nil || port < 1 || port > 65535 {
			return 0, fmt.Errorf("invalid GSUITE_MCP_OAUTH_PORT value %q: must be 1-65535", portStr)
		}
		return port, nil
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return 0, err
	}

	if cfg.OAuthPort < 1 || cfg.OAuthPort > 65535 {
		return 0, fmt.Errorf("invalid oauth_port %d in %s: must be 1-65535", cfg.OAuthPort, config.ConfigPath())
	}

	return cfg.OAuthPort, nil
}

// ResolveOAuthPort returns the resolved OAuth callback port and whether it was overridden by env var.
func ResolveOAuthPort() (port int, envOverride bool, err error) {
	if os.Getenv("GSUITE_MCP_OAUTH_PORT") != "" {
		p, err := resolveOAuthPort()
		return p, true, err
	}
	p, err := resolveOAuthPort()
	return p, false, err
}

// AuthenticateDynamic performs OAuth2 flow without requiring a pre-configured account.
// It opens the browser, lets the user choose any Google account, and saves the credential
// using the authenticated email as the identifier. Returns the authenticated email.
func (m *Manager) AuthenticateDynamic(ctx context.Context) (string, error) {
	oauthPort, err := resolveOAuthPort()
	if err != nil {
		return "", err
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", oauthPort))
	if err != nil {
		return "", fmt.Errorf("port %d is in use — set a different port in %s (oauth_port) or via GSUITE_MCP_OAUTH_PORT env var: %w",
			oauthPort, config.ConfigPath(), err)
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
		return fmt.Errorf("ensuring config dir: %w", err)
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
		return nil, fmt.Errorf("loading token for %s: %w", email, err)
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
			return nil, fmt.Errorf("getting client for %s: %w", email, err)
		}
		// No credentials for specified email
		if !interactive {
			if m.AuthServerURL != "" {
				return nil, fmt.Errorf("no credentials for %s; open %s?account=%s to authenticate", email, m.AuthServerURL, url.QueryEscape(email))
			}
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
		if m.AuthServerURL != "" {
			return nil, fmt.Errorf("no authenticated accounts; open %s to authenticate", m.AuthServerURL)
		}
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
	errorTmpl.Execute(w, errorPageData{Title: title, Message: message})
}

// isWorkspaceDomain returns true when the email is not a personal Google account.
// This is a heuristic: gmail.com and googlemail.com are consumer domains; everything
// else is assumed to be a Google Workspace (formerly G Suite) domain.
func isWorkspaceDomain(email string) bool {
	if email == "" {
		return false
	}
	at := strings.LastIndex(email, "@")
	if at < 0 {
		return false
	}
	domain := strings.ToLower(email[at+1:])
	return domain != "gmail.com" && domain != "googlemail.com"
}

// printWorkspaceDomainWarning prints a pre-flight warning when the user is attempting
// to authenticate with a Google Workspace account. Workspace admins may have restricted
// OAuth access to verified apps only, which would cause access_denied.
func printWorkspaceDomainWarning(email string) {
	fmt.Fprintf(os.Stderr, "\n⚠  Workspace domain detected (%s)\n", email)
	fmt.Fprintf(os.Stderr, "   Google Workspace admins can restrict OAuth to verified apps only.\n")
	fmt.Fprintf(os.Stderr, "   If authentication fails with 'access denied', see the options below.\n\n")
}

// printUnverifiedAppHelp prints actionable guidance when Google blocks an unverified app.
func printUnverifiedAppHelp() {
	fmt.Fprintln(os.Stderr, "\n"+strings.Repeat("=", 60))
	fmt.Fprintln(os.Stderr, "ACCESS DENIED — Unverified app blocked by Google Workspace")
	fmt.Fprintln(os.Stderr, strings.Repeat("=", 60))
	fmt.Fprintln(os.Stderr, "Your Google Workspace admin has restricted OAuth access to")
	fmt.Fprintln(os.Stderr, "verified applications only. You have three options:")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "  1. Ask your admin to allow the app:")
	fmt.Fprintln(os.Stderr, "     Admin Console → Security → API Controls → App Access Control")
	fmt.Fprintln(os.Stderr, "     → Add the OAuth client ID to the trusted apps list")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "  2. Complete OAuth app verification (if you own the app):")
	fmt.Fprintln(os.Stderr, "     https://support.google.com/cloud/answer/13463073")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "  3. Use a personal Google account (gmail.com) instead:")
	fmt.Fprintln(os.Stderr, "     Run 'gsuite-mcp auth' again and sign in with a personal account")
	fmt.Fprintln(os.Stderr, strings.Repeat("=", 60)+"\n")
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

			// Detect unverified-app blocks: Google returns error=access_denied and an
			// error_description mentioning admin policy or app verification.
			if errMsg == "access_denied" {
				descLower := strings.ToLower(errDesc)
				isWorkspaceBlock := strings.Contains(descLower, "admin") ||
					strings.Contains(descLower, "unverified") ||
					strings.Contains(descLower, "policy") ||
					strings.Contains(descLower, "restricted") ||
					errDesc == "" // access_denied with no description is also typically an admin block
				if isWorkspaceBlock {
					printUnverifiedAppHelp()
					errCh <- fmt.Errorf("%w: %s", ErrUnverifiedApp, errDesc)
					sendOAuthError(w, "Access Denied — Unverified App",
						"Your Google Workspace admin has blocked access to unverified apps. "+
							"Check the terminal for options.")
					return
				}
			}

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
			if result.err != nil {
				sendOAuthError(w, "Authentication Error", result.err.Error())
				return
			}
			w.Header().Set("Content-Type", "text/html")
			successTmpl.Execute(w, successPageData{
				Email:         result.email,
				OtherAccounts: result.otherAccounts,
			})

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
			exchangeErr := fmt.Errorf("failed to exchange code for token: %w", err)
			resultCh <- authResult{err: exchangeErr}
			return "", exchangeErr
		}

		client := oauthCfg.Client(ctx, token)
		actualEmail, err := getAuthenticatedEmail(ctx, client)
		if err != nil {
			emailErr := fmt.Errorf("failed to get authenticated email: %w", err)
			resultCh <- authResult{err: emailErr}
			return "", emailErr
		}

		if err := m.saveTokenForEmail(actualEmail, token); err != nil {
			saveErr := fmt.Errorf("failed to save token: %w", err)
			resultCh <- authResult{err: saveErr}
			return "", saveErr
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
	fmt.Fprintf(os.Stderr, "\nAuthentication required — opening browser...\n")
	fmt.Fprintf(os.Stderr, "Choose any Google account to authenticate.\n")
	fmt.Fprintf(os.Stderr, "If the browser does not open, visit:\n  %s\n\n", authURL)
}
