package auth

import (
	"errors"
	"testing"
)

func TestErrNoCredentials(t *testing.T) {
	// Test that ErrNoCredentials can be wrapped and unwrapped correctly
	wrappedErr := errors.New("no credentials for personal")
	testErr := errors.Join(ErrNoCredentials, wrappedErr)

	if !errors.Is(testErr, ErrNoCredentials) {
		t.Error("expected error to be ErrNoCredentials")
	}
}
