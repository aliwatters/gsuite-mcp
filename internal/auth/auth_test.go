package auth

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aliwatters/gsuite-mcp/internal/config"
	"golang.org/x/oauth2"
)

func TestErrNoCredentials(t *testing.T) {
	// Test that ErrNoCredentials can be wrapped and unwrapped correctly
	wrappedErr := errors.New("no credentials for personal")
	testErr := errors.Join(ErrNoCredentials, wrappedErr)

	if !errors.Is(testErr, ErrNoCredentials) {
		t.Error("expected error to be ErrNoCredentials")
	}
}

func TestHandleOAuthCallback_ErrorResultShowsErrorPage(t *testing.T) {
	// Verify that when authResult contains an error, the callback handler
	// renders the error page instead of the success page with an empty email.
	state := "test-state"
	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)
	resultCh := make(chan authResult, 1)

	handler := handleOAuthCallback(state, codeCh, errCh, resultCh)

	// Simulate: the main loop sends an authResult with an error
	// (e.g., token exchange failed)
	resultCh <- authResult{err: fmt.Errorf("failed to exchange code for token: connection refused")}

	req := httptest.NewRequest(http.MethodGet, "/oauth2callback?state=test-state&code=valid-code", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	resp := w.Result()
	body := w.Body.String()

	// Should render error page (400), not success page (200)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for auth failure, got %d", resp.StatusCode)
	}

	// Should NOT contain success page indicators
	if strings.Contains(body, "Authentication Successful") {
		t.Error("expected error page, but got success page")
	}

	// Should contain the error message
	if !strings.Contains(body, "failed to exchange code") {
		t.Errorf("expected error message in body, got: %s", body)
	}
}

func TestHandleOAuthCallback_SuccessResultShowsSuccessPage(t *testing.T) {
	// Verify that when authResult has no error, success page is rendered.
	state := "test-state"
	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)
	resultCh := make(chan authResult, 1)

	handler := handleOAuthCallback(state, codeCh, errCh, resultCh)

	resultCh <- authResult{email: "user@example.com", otherAccounts: []string{"other@example.com"}}

	req := httptest.NewRequest(http.MethodGet, "/oauth2callback?state=test-state&code=valid-code", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	resp := w.Result()
	body := w.Body.String()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for success, got %d", resp.StatusCode)
	}

	if !strings.Contains(body, "user@example.com") {
		t.Errorf("expected email in success page, got: %s", body)
	}
}

// === Unverified app / Workspace domain tests ===

func TestHandleOAuthCallback_AccessDeniedNoDesc_IsUnverifiedApp(t *testing.T) {
	// access_denied with no error_description is typically an admin-policy block.
	state := "test-state"
	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)
	resultCh := make(chan authResult, 1)

	handler := handleOAuthCallback(state, codeCh, errCh, resultCh)

	req := httptest.NewRequest(http.MethodGet,
		"/oauth2callback?state=test-state&error=access_denied",
		nil)
	w := httptest.NewRecorder()

	handler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}

	// Should propagate ErrUnverifiedApp
	select {
	case err := <-errCh:
		if !errors.Is(err, ErrUnverifiedApp) {
			t.Errorf("expected ErrUnverifiedApp, got %v", err)
		}
	default:
		t.Error("expected an error on errCh")
	}

	body := w.Body.String()
	if !strings.Contains(body, "Unverified App") {
		t.Errorf("expected unverified-app message in error page, got: %s", body)
	}
}

func TestHandleOAuthCallback_AccessDeniedWithAdminDesc_IsUnverifiedApp(t *testing.T) {
	state := "test-state"
	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)
	resultCh := make(chan authResult, 1)

	handler := handleOAuthCallback(state, codeCh, errCh, resultCh)

	req := httptest.NewRequest(http.MethodGet,
		"/oauth2callback?state=test-state&error=access_denied&error_description=Admin+policy+restricts+access",
		nil)
	w := httptest.NewRecorder()

	handler(w, req)

	select {
	case err := <-errCh:
		if !errors.Is(err, ErrUnverifiedApp) {
			t.Errorf("expected ErrUnverifiedApp, got %v", err)
		}
	default:
		t.Error("expected an error on errCh")
	}
}

func TestHandleOAuthCallback_AccessDeniedUserCancelled_IsNotUnverifiedApp(t *testing.T) {
	// access_denied can also happen when the user clicks "Cancel" — not an admin block.
	state := "test-state"
	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)
	resultCh := make(chan authResult, 1)

	handler := handleOAuthCallback(state, codeCh, errCh, resultCh)

	req := httptest.NewRequest(http.MethodGet,
		"/oauth2callback?state=test-state&error=access_denied&error_description=User+denied+access",
		nil)
	w := httptest.NewRecorder()

	handler(w, req)

	select {
	case err := <-errCh:
		if errors.Is(err, ErrUnverifiedApp) {
			t.Errorf("user-denied access_denied should NOT be ErrUnverifiedApp, got %v", err)
		}
	default:
		t.Error("expected an error on errCh")
	}
}

func TestIsWorkspaceDomain(t *testing.T) {
	cases := []struct {
		email string
		want  bool
	}{
		{"user@gmail.com", false},
		{"user@googlemail.com", false},
		{"USER@GMAIL.COM", false}, // case-insensitive
		{"user@company.com", true},
		{"user@school.edu", true},
		{"user@corp.google.com", true},
		{"", false},
		{"notanemail", false},
	}

	for _, tc := range cases {
		got := isWorkspaceDomain(tc.email)
		if got != tc.want {
			t.Errorf("isWorkspaceDomain(%q) = %v, want %v", tc.email, got, tc.want)
		}
	}
}

func TestHandleOAuthCallback_TimeoutShowsErrorPage(t *testing.T) {
	// This test verifies the timeout path. We use a very short timeout
	// by not sending anything on resultCh.
	// Note: This test is slow because it waits for the oauthResultTimeout.
	// We skip it in short mode.
	if testing.Short() {
		t.Skip("skipping timeout test in short mode")
	}

	state := "test-state"
	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)
	resultCh := make(chan authResult) // unbuffered, nothing will be sent

	handler := handleOAuthCallback(state, codeCh, errCh, resultCh)

	req := httptest.NewRequest(http.MethodGet, "/oauth2callback?state=test-state&code=valid-code", nil)
	w := httptest.NewRecorder()

	start := time.Now()
	handler(w, req)
	elapsed := time.Since(start)

	// Should have waited roughly oauthResultTimeout
	if elapsed < 5*time.Second {
		t.Errorf("expected handler to wait at least 5s, waited %v", elapsed)
	}

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for timeout, got %d", resp.StatusCode)
	}
}

// === Refresh token preservation tests ===

// newTestManager creates a Manager with a minimal oauth2.Config and the given config dir.
func newTestManager(t *testing.T, configDir string) *Manager {
	t.Helper()
	config.SetConfigDir(configDir)
	t.Cleanup(func() { config.SetConfigDir("") })
	return &Manager{
		oauthConfig: &oauth2.Config{
			ClientID:     "test-client",
			ClientSecret: "test-secret",
			Endpoint: oauth2.Endpoint{
				TokenURL: "https://oauth2.googleapis.com/token",
			},
			Scopes: []string{"openid"},
		},
	}
}

func TestSaveTokenForEmail_PreservesExistingRefreshToken(t *testing.T) {
	// When a refresh response omits refresh_token (the normal Google behaviour),
	// saveTokenForEmail must keep the original refresh token from disk.
	dir := t.TempDir()
	m := newTestManager(t, dir)

	email := "user@example.com"

	// First save: initial grant with refresh token.
	initial := &oauth2.Token{
		AccessToken:  "access1",
		RefreshToken: "refresh-original",
		Expiry:       time.Now().Add(time.Hour),
	}
	if err := m.saveTokenForEmail(email, initial); err != nil {
		t.Fatalf("initial save failed: %v", err)
	}

	// Simulate a silent refresh: Google returns new access token but NO refresh token.
	refreshed := &oauth2.Token{
		AccessToken:  "access2",
		RefreshToken: "", // empty — normal for refresh responses
		Expiry:       time.Now().Add(time.Hour),
	}
	if err := m.saveTokenForEmail(email, refreshed); err != nil {
		t.Fatalf("refresh save failed: %v", err)
	}

	// Read back and verify refresh token was preserved.
	stored, err := loadTokenForEmail(email)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if stored.RefreshToken != "refresh-original" {
		t.Errorf("expected refresh_token %q to be preserved, got %q", "refresh-original", stored.RefreshToken)
	}
	if stored.Token != "access2" {
		t.Errorf("expected access token to be updated to %q, got %q", "access2", stored.Token)
	}
}

func TestSaveTokenForEmail_NewRefreshTokenOverridesOld(t *testing.T) {
	// When a refresh response includes a new refresh token, it should replace the old one.
	// (Rare but can happen with prompt=consent or forced re-auth.)
	dir := t.TempDir()
	m := newTestManager(t, dir)

	email := "user@example.com"

	initial := &oauth2.Token{
		AccessToken:  "access1",
		RefreshToken: "refresh-original",
		Expiry:       time.Now().Add(time.Hour),
	}
	if err := m.saveTokenForEmail(email, initial); err != nil {
		t.Fatalf("initial save failed: %v", err)
	}

	updated := &oauth2.Token{
		AccessToken:  "access2",
		RefreshToken: "refresh-new",
		Expiry:       time.Now().Add(time.Hour),
	}
	if err := m.saveTokenForEmail(email, updated); err != nil {
		t.Fatalf("updated save failed: %v", err)
	}

	stored, err := loadTokenForEmail(email)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if stored.RefreshToken != "refresh-new" {
		t.Errorf("expected new refresh token %q, got %q", "refresh-new", stored.RefreshToken)
	}
}

func TestIsAuthExpiredError(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"invalid_grant", fmt.Errorf("oauth2: %q %q", "invalid_grant", "Token has been expired or revoked."), true},
		{"invalid_grant bare", fmt.Errorf("oauth2: invalid_grant"), true},
		{"token has been expired lower", fmt.Errorf("token has been expired or revoked"), true},
		{"unrelated error", fmt.Errorf("connection refused"), false},
		{"401 but not grant", fmt.Errorf("oauth2: 401 Unauthorized"), false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := isAuthExpiredError(tc.err)
			if got != tc.want {
				t.Errorf("isAuthExpiredError(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}

func TestErrAuthExpired_IsDetectable(t *testing.T) {
	// Verify ErrAuthExpired wraps correctly so callers can use errors.Is.
	wrapped := fmt.Errorf("%w: token for user@example.com has expired: %w",
		ErrAuthExpired, fmt.Errorf("oauth2: invalid_grant"))
	if !errors.Is(wrapped, ErrAuthExpired) {
		t.Error("expected errors.Is(err, ErrAuthExpired) to be true")
	}
}

func TestSaveTokenForEmail_CredentialFilePermissions(t *testing.T) {
	// Credential files must be 0600 — no group/world read.
	dir := t.TempDir()
	m := newTestManager(t, dir)

	email := "perm@example.com"
	token := &oauth2.Token{
		AccessToken:  "access",
		RefreshToken: "refresh",
		Expiry:       time.Now().Add(time.Hour),
	}
	if err := m.saveTokenForEmail(email, token); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	path := filepath.Join(config.CredentialsDir(), email+".json")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}
	if perm := info.Mode().Perm(); perm != credentialFileMode {
		t.Errorf("expected file mode %04o, got %04o", credentialFileMode, perm)
	}
}
