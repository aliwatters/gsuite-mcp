package auth

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
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
		email    string
		want     bool
	}{
		{"user@gmail.com", false},
		{"user@googlemail.com", false},
		{"USER@GMAIL.COM", false},  // case-insensitive
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
