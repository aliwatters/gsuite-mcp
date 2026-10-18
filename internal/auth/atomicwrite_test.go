package auth

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aliwatters/gsuite-mcp/internal/config"
	"golang.org/x/oauth2"
)

// assertNoTempLeftovers fails the test if any atomic-write temp file remains in dir.
func assertNoTempLeftovers(t *testing.T, dir string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("reading dir %s: %v", dir, err)
	}
	for _, e := range entries {
		if strings.Contains(e.Name(), ".tmp-") {
			t.Errorf("temp file left behind: %s", filepath.Join(dir, e.Name()))
		}
	}
}

func TestSaveTokenForEmail_InterruptedWritePreservesPreviousFile(t *testing.T) {
	// A failure partway through writing new credentials must leave the previous
	// credential file byte-identical and must not leave a temp file behind.
	dir := t.TempDir()
	m := newTestManager(t, dir)

	email := "interrupt@example.com"
	tokenA := &oauth2.Token{
		AccessToken:  "access-A",
		RefreshToken: "refresh-A",
		Expiry:       time.Now().Add(time.Hour),
	}
	if err := m.saveTokenForEmail(email, tokenA); err != nil {
		t.Fatalf("saving token A: %v", err)
	}
	path := config.CredentialPathForEmail(email)
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading token A file: %v", err)
	}

	errInjected := errors.New("injected partial write")
	orig := atomicWriteData
	atomicWriteData = func(f *os.File, b []byte) error {
		if _, err := f.Write(b[:len(b)/2]); err != nil {
			return err
		}
		return errInjected
	}
	t.Cleanup(func() { atomicWriteData = orig })

	tokenB := &oauth2.Token{
		AccessToken:  "access-B",
		RefreshToken: "refresh-B",
		Expiry:       time.Now().Add(2 * time.Hour),
	}
	err = m.saveTokenForEmail(email, tokenB)
	if err == nil {
		t.Fatal("expected saving token B to fail after an interrupted write, got nil")
	}
	if !errors.Is(err, errInjected) {
		t.Errorf("expected error to wrap the injected failure, got %v", err)
	}

	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading credential file after failed save: %v", err)
	}
	if string(after) != string(before) {
		t.Errorf("credential file changed after failed save:\n got %q\nwant %q", after, before)
	}

	loaded, err := loadTokenFromFile(email)
	if err != nil {
		t.Fatalf("loading credentials after failed save: %v", err)
	}
	if loaded.RefreshToken != "refresh-A" {
		t.Errorf("expected refresh token %q after failed save, got %q", "refresh-A", loaded.RefreshToken)
	}

	assertNoTempLeftovers(t, config.CredentialsDir())
}

func TestSaveTokenForEmail_ConcurrentReadersNeverSeePartialFile(t *testing.T) {
	// Readers racing a writer must always see a complete, parseable credential
	// file: either the previous contents or the new contents, never a
	// truncated or partially written one.
	dir := t.TempDir()
	m := newTestManager(t, dir)

	email := "concurrent@example.com"
	// A large payload widens the window between truncate and write completion
	// for a non-atomic writer.
	padding := strings.Repeat("x", 64*1024)
	makeToken := func(i int) *oauth2.Token {
		return &oauth2.Token{
			AccessToken:  fmt.Sprintf("access-%d", i),
			RefreshToken: fmt.Sprintf("refresh-%d-%s", i, padding),
			Expiry:       time.Now().Add(time.Hour),
		}
	}
	if err := m.saveTokenForEmail(email, makeToken(0)); err != nil {
		t.Fatalf("initial save: %v", err)
	}

	const (
		writes  = 200
		readers = 4
	)
	var (
		done       atomic.Bool
		readErrors atomic.Int64
		firstErr   atomic.Value
		wg         sync.WaitGroup
	)
	for r := 0; r < readers; r++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for !done.Load() {
				if _, err := loadTokenFromFile(email); err != nil {
					readErrors.Add(1)
					firstErr.CompareAndSwap(nil, err.Error())
				}
			}
		}()
	}

	var writeErr error
	for i := 1; i <= writes; i++ {
		if err := m.saveTokenForEmail(email, makeToken(i)); err != nil {
			writeErr = fmt.Errorf("save %d: %w", i, err)
			break
		}
	}
	done.Store(true)
	wg.Wait()

	if writeErr != nil {
		t.Fatalf("writer failed: %v", writeErr)
	}
	if n := readErrors.Load(); n != 0 {
		t.Errorf("readers saw %d read/parse errors; first: %v", n, firstErr.Load())
	}
	assertNoTempLeftovers(t, config.CredentialsDir())
}

func TestSaveTokenForEmail_NormalizesWiderModeTo0600(t *testing.T) {
	// An existing credential file with a wider mode must come back as 0600
	// after a save: every credential write normalizes the mode.
	dir := t.TempDir()
	m := newTestManager(t, dir)

	email := "widen@example.com"
	token := &oauth2.Token{
		AccessToken:  "access",
		RefreshToken: "refresh",
		Expiry:       time.Now().Add(time.Hour),
	}
	if err := m.saveTokenForEmail(email, token); err != nil {
		t.Fatalf("initial save: %v", err)
	}
	path := config.CredentialPathForEmail(email)
	if err := os.Chmod(path, 0o644); err != nil {
		t.Fatalf("widening mode: %v", err)
	}

	if err := m.saveTokenForEmail(email, token); err != nil {
		t.Fatalf("second save: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}
	// Hardcoded literal on purpose; see TestSaveTokenForEmail_CredentialFilePermissions.
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("expected file mode 0600 after save, got %04o", perm)
	}
}
