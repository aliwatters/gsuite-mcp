package auth

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

// newTestAuthServer creates an AuthServer with a fake oauth2 config for testing.
func newTestAuthServer() *AuthServer {
	mgr := &Manager{
		oauthConfig: &oauth2.Config{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://accounts.google.com/o/oauth2/auth",
				TokenURL: "https://oauth2.googleapis.com/token",
			},
			Scopes: []string{"openid", "email"},
		},
	}
	return NewAuthServer(mgr, 8100)
}

func TestHandleAuth_RedirectsToGoogle(t *testing.T) {
	s := newTestAuthServer()

	req := httptest.NewRequest(http.MethodGet, "/auth", nil)
	w := httptest.NewRecorder()

	s.handleAuth(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("expected 302 redirect, got %d", resp.StatusCode)
	}

	location := resp.Header.Get("Location")
	if location == "" {
		t.Fatal("expected Location header")
	}

	u, err := url.Parse(location)
	if err != nil {
		t.Fatalf("invalid redirect URL: %v", err)
	}

	if u.Host != "accounts.google.com" {
		t.Errorf("expected redirect to accounts.google.com, got %s", u.Host)
	}

	state := u.Query().Get("state")
	if state == "" {
		t.Error("expected state parameter in redirect URL")
	}

	// Verify state was stored
	if _, ok := s.states.Load(state); !ok {
		t.Error("state was not stored in map")
	}

	// Should not have login_hint
	if u.Query().Get("login_hint") != "" {
		t.Error("did not expect login_hint without account param")
	}
}

func TestHandleAuth_WithAccountParam(t *testing.T) {
	s := newTestAuthServer()

	req := httptest.NewRequest(http.MethodGet, "/auth?account=user@gmail.com", nil)
	w := httptest.NewRecorder()

	s.handleAuth(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("expected 302, got %d", resp.StatusCode)
	}

	location := resp.Header.Get("Location")
	u, err := url.Parse(location)
	if err != nil {
		t.Fatalf("invalid redirect URL: %v", err)
	}

	if hint := u.Query().Get("login_hint"); hint != "user@gmail.com" {
		t.Errorf("expected login_hint=user@gmail.com, got %q", hint)
	}
}

func TestHandleCallback_UnknownState(t *testing.T) {
	s := newTestAuthServer()

	req := httptest.NewRequest(http.MethodGet, "/oauth2callback?state=bogus&code=abc", nil)
	w := httptest.NewRecorder()

	s.handleCallback(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Invalid or expired state token") {
		t.Errorf("expected error about invalid state, got: %s", body)
	}
}

func TestHandleCallback_ExpiredState(t *testing.T) {
	s := newTestAuthServer()

	// Store a state that already expired
	s.states.Store("expired-state", pendingState{
		expiresAt: time.Now().Add(-1 * time.Minute),
	})

	req := httptest.NewRequest(http.MethodGet, "/oauth2callback?state=expired-state&code=abc", nil)
	w := httptest.NewRecorder()

	s.handleCallback(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}

	body := w.Body.String()
	if !strings.Contains(body, "expired") {
		t.Errorf("expected error about expiry, got: %s", body)
	}
}

func TestHandleCallback_MissingCode(t *testing.T) {
	s := newTestAuthServer()

	s.states.Store("valid-state", pendingState{
		expiresAt: time.Now().Add(5 * time.Minute),
	})

	req := httptest.NewRequest(http.MethodGet, "/oauth2callback?state=valid-state", nil)
	w := httptest.NewRecorder()

	s.handleCallback(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}

	body := w.Body.String()
	if !strings.Contains(body, "No authorization code") {
		t.Errorf("expected error about missing code, got: %s", body)
	}
}

func TestHandleCallback_GoogleError(t *testing.T) {
	s := newTestAuthServer()

	req := httptest.NewRequest(http.MethodGet, "/oauth2callback?error=access_denied&error_description=User+denied+access", nil)
	w := httptest.NewRecorder()

	s.handleCallback(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}

	body := w.Body.String()
	if !strings.Contains(body, "User denied access") {
		t.Errorf("expected Google error description, got: %s", body)
	}
}

func TestHandleCallback_MissingState(t *testing.T) {
	s := newTestAuthServer()

	req := httptest.NewRequest(http.MethodGet, "/oauth2callback?code=abc", nil)
	w := httptest.NewRecorder()

	s.handleCallback(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Missing state") {
		t.Errorf("expected error about missing state, got: %s", body)
	}
}

func TestCleanupExpiredStates(t *testing.T) {
	s := newTestAuthServer()

	// Add an expired state and a valid state
	s.states.Store("expired", pendingState{
		expiresAt: time.Now().Add(-1 * time.Minute),
	})
	s.states.Store("valid", pendingState{
		expiresAt: time.Now().Add(5 * time.Minute),
	})

	// Run cleanup manually (simulate one tick)
	now := time.Now()
	s.states.Range(func(key, value any) bool {
		ps := value.(pendingState)
		if now.After(ps.expiresAt) {
			s.states.Delete(key)
		}
		return true
	})

	// Expired should be gone
	if _, ok := s.states.Load("expired"); ok {
		t.Error("expected expired state to be cleaned up")
	}

	// Valid should remain
	if _, ok := s.states.Load("valid"); !ok {
		t.Error("expected valid state to still exist")
	}
}

func TestHandleCallback_StateUsedOnce(t *testing.T) {
	s := newTestAuthServer()

	s.states.Store("one-time", pendingState{
		expiresAt: time.Now().Add(5 * time.Minute),
	})

	// First call consumes the state (will fail at code exchange, but state is consumed)
	req := httptest.NewRequest(http.MethodGet, "/oauth2callback?state=one-time&code=fake", nil)
	w := httptest.NewRecorder()
	s.handleCallback(w, req)

	// Second call with same state should fail with invalid state
	req2 := httptest.NewRequest(http.MethodGet, "/oauth2callback?state=one-time&code=fake", nil)
	w2 := httptest.NewRecorder()
	s.handleCallback(w2, req2)

	if w2.Result().StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 on reuse, got %d", w2.Result().StatusCode)
	}
	if !strings.Contains(w2.Body.String(), "Invalid or expired state token") {
		t.Error("expected invalid state error on reuse")
	}
}
