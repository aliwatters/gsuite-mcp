package gmail

import (
	"context"
	"strings"
	"testing"
)

// === ParseGmailWebID unit tests ===

func TestParseGmailWebID_Empty(t *testing.T) {
	_, err := ParseGmailWebID("")
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}

// TestParseGmailWebID_APIID verifies that an already-API hex ID is passed through.
func TestParseGmailWebID_APIID(t *testing.T) {
	cases := []string{
		"552634d52b0bc546",
		"06dd2cf16cc577e3",
		"1a2b3c4d",
		"0",
	}
	for _, id := range cases {
		t.Run(id, func(t *testing.T) {
			got, err := ParseGmailWebID(id)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Kind != WebIDKindAPIID {
				t.Errorf("kind: got %q, want %q", got.Kind, WebIDKindAPIID)
			}
			if got.ThreadID != id {
				t.Errorf("thread_id: got %q, want %q", got.ThreadID, id)
			}
			if got.MessageID != id {
				t.Errorf("message_id: got %q, want %q", got.MessageID, id)
			}
		})
	}
}

// TestParseGmailWebID_ThreadF verifies legacy thread-f:<decimal> decoding.
func TestParseGmailWebID_ThreadF(t *testing.T) {
	cases := []struct {
		input    string
		wantHex  string
	}{
		// thread-f:1821570065795440641 -> hex: 19478452e138e001
		{"thread-f:1821570065795440641", "19478452e138e001"},
		// thread-f:0 -> hex: 0
		{"thread-f:0", "0"},
		// thread-f:255 -> hex: ff
		{"thread-f:255", "ff"},
		// thread-f:256 -> hex: 100
		{"thread-f:256", "100"},
		// Larger realistic value
		{"thread-f:1780000000000000", "652e68bb34000"},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got, err := ParseGmailWebID(tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Kind != WebIDKindThreadF {
				t.Errorf("kind: got %q, want %q", got.Kind, WebIDKindThreadF)
			}
			if got.ThreadID != tc.wantHex {
				t.Errorf("thread_id: got %q, want %q", got.ThreadID, tc.wantHex)
			}
			if got.MessageID != "" {
				t.Errorf("message_id: expected empty, got %q", got.MessageID)
			}
		})
	}
}

// TestParseGmailWebID_MsgF verifies legacy msg-f:<decimal> decoding.
func TestParseGmailWebID_MsgF(t *testing.T) {
	cases := []struct {
		input   string
		wantHex string
	}{
		{"msg-f:1821570065795440641", "19478452e138e001"},
		{"msg-f:255", "ff"},
		{"msg-f:0", "0"},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got, err := ParseGmailWebID(tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Kind != WebIDKindMsgF {
				t.Errorf("kind: got %q, want %q", got.Kind, WebIDKindMsgF)
			}
			if got.MessageID != tc.wantHex {
				t.Errorf("message_id: got %q, want %q", got.MessageID, tc.wantHex)
			}
			if got.ThreadID != "" {
				t.Errorf("thread_id: expected empty, got %q", got.ThreadID)
			}
		})
	}
}

// TestParseGmailWebID_FMfcg verifies the current base64url FMfcg form.
//
// The example ID "FMfcgzQgLrxVJjTVKwvFRgbdLPFsxXfj" decodes to 24 bytes:
//   14c7dc8334202ebc 552634d52b0bc546 06dd2cf16cc577e3
// bytes 8-15 = thread ID = 552634d52b0bc546
// bytes 16-23 = message ID = 6dd2cf16cc577e3
func TestParseGmailWebID_FMfcg_24bytes(t *testing.T) {
	id := "FMfcgzQgLrxVJjTVKwvFRgbdLPFsxXfj"
	got, err := ParseGmailWebID(id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Kind != WebIDKindFMfcg {
		t.Errorf("kind: got %q, want %q", got.Kind, WebIDKindFMfcg)
	}
	wantThread := "552634d52b0bc546"
	wantMsg := "6dd2cf16cc577e3"
	if got.ThreadID != wantThread {
		t.Errorf("thread_id: got %q, want %q", got.ThreadID, wantThread)
	}
	if got.MessageID != wantMsg {
		t.Errorf("message_id: got %q, want %q", got.MessageID, wantMsg)
	}
}

// TestParseGmailWebID_FMfcg_BadBase64 ensures a clear error for corrupt IDs.
// Note: IDs with non-base64url characters (like !) fail the reFMfcg regex and
// are reported as "unrecognised". IDs that pass the regex but fail base64 decode
// produce an "FMfcg decode" error.
func TestParseGmailWebID_FMfcg_BadBase64(t *testing.T) {
	// This ID has valid base64url characters but decodes to too few bytes (< 8)
	// Use a short base64url string that passes the length check for reFMfcg
	// but then fails inside decodeFMfcg (too short after decode).
	// A 43-char base64url string with all valid chars but invalid content:
	// "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA" decodes to 32 bytes
	// which is >= 24, so it would succeed. Let's use a valid-looking but semantically
	// bad ID that's exactly at a boundary.
	// Actually the simplest approach: verify that a valid-looking FMfcg ID
	// containing non-alphabet chars (which URL encoding might cause) is rejected.
	// Characters ! are not in [A-Za-z0-9_-], so they are rejected before decode.
	id := "FMfcgzQgLrxVJjTVKwvFRgbdLPFsxX!!!"
	_, err := ParseGmailWebID(id)
	if err == nil {
		t.Fatal("expected error for ID with invalid characters")
	}
	// Should report as unrecognised (non-base64url chars fail the regex)
	if !strings.Contains(err.Error(), "unrecognised") {
		t.Errorf("expected 'unrecognised' in error, got: %v", err)
	}
}

// TestParseGmailWebID_FMfcg_TooShort ensures a clear error when the decoded bytes are too few.
// A 33-char base64url string of all 'A's decodes to ~24 bytes (fine), but a very short one
// (e.g. 12 chars = 9 bytes decoded) is too short and triggers the "too short" error.
func TestParseGmailWebID_FMfcg_TooShort(t *testing.T) {
	// 12 valid base64url chars → 9 bytes after decode → < 8 bytes → error
	// But 12 chars is < 32 minimum for reFMfcg, so won't match.
	// Use 32 valid base64url chars that decode to exactly 24 bytes (fine) — hard.
	// Instead verify the decodeFMfcg function directly.
	_, _, err := decodeFMfcg("AAAA") // 4 chars → 3 bytes → too short
	if err == nil {
		t.Fatal("expected error for decoded bytes < 8")
	}
	if !strings.Contains(err.Error(), "too short") {
		t.Errorf("expected 'too short' in error, got: %v", err)
	}
}

// TestParseGmailWebID_Unrecognised ensures a clear error for unrecognised forms.
func TestParseGmailWebID_Unrecognised(t *testing.T) {
	cases := []string{
		"not-a-valid-id",
		"thread-g:12345",
		"INBOX",
		"abc XYZ",
	}
	for _, id := range cases {
		t.Run(id, func(t *testing.T) {
			_, err := ParseGmailWebID(id)
			if err == nil {
				t.Fatalf("expected error for unrecognised ID %q", id)
			}
		})
	}
}

// === ExtractWebIDFromURL unit tests ===

func TestExtractWebIDFromURL_BareID(t *testing.T) {
	id := "FMfcgzQgLrxVJjTVKwvFRgbdLPFsxXfj"
	got := ExtractWebIDFromURL(id)
	if got != id {
		t.Errorf("got %q, want %q", got, id)
	}
}

func TestExtractWebIDFromURL_FullURL(t *testing.T) {
	u := "https://mail.google.com/mail/u/0/#inbox/FMfcgzQgLrxVJjTVKwvFRgbdLPFsxXfj"
	got := ExtractWebIDFromURL(u)
	want := "FMfcgzQgLrxVJjTVKwvFRgbdLPFsxXfj"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestExtractWebIDFromURL_ThreadLabelURL(t *testing.T) {
	// #label/<id> style URL
	u := "https://mail.google.com/mail/u/0/#label/Work/FMfcgzQgLrxVJjTVKwvFRgbdLPFsxXfj"
	got := ExtractWebIDFromURL(u)
	want := "FMfcgzQgLrxVJjTVKwvFRgbdLPFsxXfj"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestExtractWebIDFromURL_SentURL(t *testing.T) {
	u := "https://mail.google.com/mail/u/0/#sent/FMfcgzQgLrxVJjTVKwvFRgbdLPFsxXfj"
	got := ExtractWebIDFromURL(u)
	want := "FMfcgzQgLrxVJjTVKwvFRgbdLPFsxXfj"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// TestExtractWebIDFromURL_TwoSegmentFragment simulates a URL where the fragment
// ends with /threadID/msgID — the last segment (message ID) is returned.
func TestExtractWebIDFromURL_TwoSegmentFragment(t *testing.T) {
	// Format: #inbox/<threadID>/<msgID>
	u := "https://mail.google.com/mail/u/0/#inbox/AAAAAAAAAABthread/AAAAAAAAAABmsg"
	got := ExtractWebIDFromURL(u)
	want := "AAAAAAAAAABmsg"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// === Integration: ParseGmailWebID from full URL ===

func TestParseGmailWebID_FromFullURL(t *testing.T) {
	u := "https://mail.google.com/mail/u/0/#inbox/FMfcgzQgLrxVJjTVKwvFRgbdLPFsxXfj"
	got, err := ParseGmailWebID(u)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Kind != WebIDKindFMfcg {
		t.Errorf("kind: got %q, want %q", got.Kind, WebIDKindFMfcg)
	}
	if got.ThreadID == "" {
		t.Error("expected non-empty thread_id")
	}
	if got.MessageID == "" {
		t.Error("expected non-empty message_id")
	}
}

func TestParseGmailWebID_ThreadFFromURL(t *testing.T) {
	u := "https://mail.google.com/mail/u/0/#all/thread-f:1821570065795440641"
	got, err := ParseGmailWebID(u)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Kind != WebIDKindThreadF {
		t.Errorf("kind: got %q, want %q", got.Kind, WebIDKindThreadF)
	}
	if got.ThreadID != "19478452e138e001" {
		t.Errorf("thread_id: got %q, want %q", got.ThreadID, "19478452e138e001")
	}
}

// === TestableGmailResolveWebID handler tests ===

func TestGmailResolveWebID_FMfcgSuccess(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	request := makeRequest(map[string]any{
		"id": "FMfcgzQgLrxVJjTVKwvFRgbdLPFsxXfj",
	})

	result, err := TestableGmailResolveWebID(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	response := extractResponse(t, result)
	if response["id_kind"] != "FMfcg" {
		t.Errorf("id_kind: got %v, want FMfcg", response["id_kind"])
	}
	if response["thread_id"] == nil || response["thread_id"] == "" {
		t.Error("expected thread_id in response")
	}
	if response["message_id"] == nil || response["message_id"] == "" {
		t.Error("expected message_id in response")
	}
	if response["source_id"] != "FMfcgzQgLrxVJjTVKwvFRgbdLPFsxXfj" {
		t.Errorf("source_id: got %v", response["source_id"])
	}
	if response["hint"] == nil {
		t.Error("expected hint in response")
	}
}

func TestGmailResolveWebID_ThreadFSuccess(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	request := makeRequest(map[string]any{
		"id": "thread-f:1821570065795440641",
	})

	result, err := TestableGmailResolveWebID(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	response := extractResponse(t, result)
	if response["id_kind"] != "thread-f" {
		t.Errorf("id_kind: got %v, want thread-f", response["id_kind"])
	}
	if response["thread_id"] != "19478452e138e001" {
		t.Errorf("thread_id: got %v, want 19478452e138e001", response["thread_id"])
	}
	if _, hasMsg := response["message_id"]; hasMsg {
		t.Error("message_id should be absent for thread-f: form")
	}
}

func TestGmailResolveWebID_MsgFSuccess(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	request := makeRequest(map[string]any{
		"id": "msg-f:1821570065795440641",
	})

	result, err := TestableGmailResolveWebID(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	response := extractResponse(t, result)
	if response["id_kind"] != "msg-f" {
		t.Errorf("id_kind: got %v, want msg-f", response["id_kind"])
	}
	if response["message_id"] != "19478452e138e001" {
		t.Errorf("message_id: got %v, want 19478452e138e001", response["message_id"])
	}
	if _, hasThread := response["thread_id"]; hasThread {
		t.Error("thread_id should be absent for msg-f: form")
	}
}

func TestGmailResolveWebID_APIIDPassthrough(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	request := makeRequest(map[string]any{
		"id": "552634d52b0bc546",
	})

	result, err := TestableGmailResolveWebID(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	response := extractResponse(t, result)
	if response["id_kind"] != "api-id" {
		t.Errorf("id_kind: got %v, want api-id", response["id_kind"])
	}
	if response["thread_id"] != "552634d52b0bc546" {
		t.Errorf("thread_id: got %v, want 552634d52b0bc546", response["thread_id"])
	}
	if response["message_id"] != "552634d52b0bc546" {
		t.Errorf("message_id: got %v, want 552634d52b0bc546", response["message_id"])
	}
}

func TestGmailResolveWebID_FullURL(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	request := makeRequest(map[string]any{
		"id": "https://mail.google.com/mail/u/0/#inbox/FMfcgzQgLrxVJjTVKwvFRgbdLPFsxXfj",
	})

	result, err := TestableGmailResolveWebID(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	response := extractResponse(t, result)
	if response["id_kind"] != "FMfcg" {
		t.Errorf("id_kind: got %v, want FMfcg", response["id_kind"])
	}
	if response["thread_id"] == nil || response["thread_id"] == "" {
		t.Error("expected thread_id in response")
	}
}

func TestGmailResolveWebID_MissingID(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	request := makeRequest(map[string]any{})

	result, err := TestableGmailResolveWebID(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for missing id parameter")
	}
}

func TestGmailResolveWebID_InvalidID(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	request := makeRequest(map[string]any{
		"id": "not-a-valid-gmail-id",
	})

	result, err := TestableGmailResolveWebID(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for unrecognised ID form")
	}
}

func TestGmailResolveWebID_EmptyID(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	request := makeRequest(map[string]any{
		"id": "",
	})

	result, err := TestableGmailResolveWebID(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for empty id")
	}
}

// TestGmailResolveWebID_WhitespaceStripped verifies leading/trailing whitespace is stripped.
func TestGmailResolveWebID_WhitespaceStripped(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	request := makeRequest(map[string]any{
		"id": "  thread-f:1821570065795440641  ",
	})

	result, err := TestableGmailResolveWebID(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success after whitespace strip, got error: %v", result.Content)
	}
}
