package gmail

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// WebIDKind classifies what kind of web-UI ID was detected.
type WebIDKind string

const (
	// WebIDKindFMfcg is the current Gmail web client form (base64url-encoded 24-byte structure).
	WebIDKindFMfcg WebIDKind = "FMfcg"
	// WebIDKindThreadF is the legacy thread-f:<decimal> form.
	WebIDKindThreadF WebIDKind = "thread-f"
	// WebIDKindMsgF is the legacy msg-f:<decimal> form.
	WebIDKindMsgF WebIDKind = "msg-f"
	// WebIDKindAPIID is already a Gmail API message/thread ID (hex string).
	WebIDKindAPIID WebIDKind = "api-id"
)

// ResolvedWebID holds the resolved API IDs derived from a Gmail web-UI identifier.
type ResolvedWebID struct {
	// Kind is the detected form of the input.
	Kind WebIDKind
	// ThreadID is the Gmail API thread ID (lowercase hex, no leading zeros).
	// May be empty if the input form does not encode a thread ID distinctly.
	ThreadID string
	// MessageID is the Gmail API message ID (lowercase hex, no leading zeros).
	// May be empty for thread-only IDs (thread-f: form).
	MessageID string
}

// hexStr returns a lowercase hex string without leading zeros, but never empty.
// A zero value returns "0".
func hexStr(v uint64) string {
	return fmt.Sprintf("%x", v)
}

// reThreadF matches legacy thread-f:<decimal> IDs.
var reThreadF = regexp.MustCompile(`^thread-f:(\d+)$`)

// reMsgF matches legacy msg-f:<decimal> IDs.
var reMsgF = regexp.MustCompile(`^msg-f:(\d+)$`)

// reAPIID matches a bare Gmail API ID: 1–20 lowercase hex characters.
// Real Gmail API IDs are 16 hex chars; allow 1–20 to handle edge cases.
var reAPIID = regexp.MustCompile(`^[0-9a-f]{1,20}$`)

// reFMfcg matches the current FMfcg... base64url form.
// These IDs start with capital letters and are 32–48 characters long.
var reFMfcg = regexp.MustCompile(`^[A-Za-z0-9_-]{32,64}$`)

// ExtractWebIDFromURL extracts the raw web ID from a Gmail URL fragment.
// Accepts:
//   - Full Gmail URL: https://mail.google.com/mail/u/0/#inbox/FMfcgzQgLrxV…
//   - URL with msg ID: https://mail.google.com/mail/u/0/#inbox/FMfcgzQg…/FMfcgzQg…
//
// Returns the bare ID (the last path segment of the fragment), or the input
// unchanged if it does not look like a Gmail URL.
func ExtractWebIDFromURL(input string) string {
	input = strings.TrimSpace(input)

	// Only attempt URL parsing if it looks like a URL
	if !strings.Contains(input, "mail.google.com") {
		return input
	}

	parsed, err := url.Parse(input)
	if err != nil {
		return input
	}

	fragment := parsed.Fragment // e.g. "inbox/FMfcgzQgLrxV..." or "inbox/FMfcg.../FMfcg..."
	if fragment == "" {
		return input
	}

	// Fragment is like "inbox/ID" or "label/ID" or "inbox/threadID/msgID"
	parts := strings.Split(fragment, "/")
	if len(parts) < 2 {
		return input
	}

	// If the URL contains both thread and message IDs (two long IDs), prefer the last
	// which is the message ID. If only one ID segment, that is the thread/message ID.
	last := parts[len(parts)-1]
	if last == "" {
		return input
	}
	return last
}

// parseUint64Decimal parses a decimal string as uint64.
func parseUint64Decimal(s string) (uint64, error) {
	var v uint64
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return 0, fmt.Errorf("non-digit character %q in decimal string %q", ch, s)
		}
		next := v*10 + uint64(ch-'0')
		if next < v {
			return 0, fmt.Errorf("decimal value %q overflows uint64", s)
		}
		v = next
	}
	return v, nil
}

// decodeFMfcg decodes a FMfcg… base64url ID into (threadID, messageID).
//
// Community-documented structure (24 bytes):
//   - bytes  0– 7: 8-byte header (account/prefix info — not used for ID resolution)
//   - bytes  8–15: 8-byte big-endian uint64 thread ID → lowercase hex = API thread ID
//   - bytes 16–23: 8-byte big-endian uint64 message ID → lowercase hex = API message ID
//
// Some older FMfcg IDs are 8 or 16 bytes. We handle all three sizes gracefully.
func decodeFMfcg(id string) (threadID, messageID string, err error) {
	// Pad to a multiple of 4 for standard base64 decoding
	padded := id
	switch len(id) % 4 {
	case 2:
		padded += "=="
	case 3:
		padded += "="
	}

	raw, err := base64.URLEncoding.DecodeString(padded)
	if err != nil {
		return "", "", fmt.Errorf("base64url decode of %q failed: %w", id, err)
	}

	switch {
	case len(raw) >= 24:
		// Full 24-byte structure: bytes 8–15 = thread ID, bytes 16–23 = message ID
		threadVal := binary.BigEndian.Uint64(raw[8:16])
		msgVal := binary.BigEndian.Uint64(raw[16:24])
		return hexStr(threadVal), hexStr(msgVal), nil

	case len(raw) >= 16:
		// 16-byte structure: bytes 0–7 = thread ID, bytes 8–15 = message ID (older form)
		threadVal := binary.BigEndian.Uint64(raw[0:8])
		msgVal := binary.BigEndian.Uint64(raw[8:16])
		return hexStr(threadVal), hexStr(msgVal), nil

	case len(raw) >= 8:
		// 8-byte structure: single ID (thread only)
		threadVal := binary.BigEndian.Uint64(raw[0:8])
		return hexStr(threadVal), "", nil

	default:
		return "", "", fmt.Errorf("decoded FMfcg ID %q is too short (%d bytes, need ≥8)", id, len(raw))
	}
}

// ParseGmailWebID detects the form of a Gmail web-UI ID and derives the
// corresponding Gmail API thread/message IDs.
//
// Accepted input forms:
//   - Full Gmail URL: https://mail.google.com/mail/u/0/#inbox/FMfcgzQg…
//   - Bare FMfcg… ID (current form, base64url-encoded 24-byte structure)
//   - thread-f:<decimal> (legacy thread URL)
//   - msg-f:<decimal>    (legacy message URL)
//   - Already-API ID     (lowercase hex, 1–20 chars) — passed through unchanged
//
// Returns a non-nil error with full context when resolution fails. Never returns
// a result where both ThreadID and MessageID are empty strings.
func ParseGmailWebID(input string) (*ResolvedWebID, error) {
	if input == "" {
		return nil, fmt.Errorf("gmail web ID: input is empty")
	}

	// Step 1: extract bare ID from URL if needed
	bare := ExtractWebIDFromURL(input)
	if bare == "" {
		return nil, fmt.Errorf("gmail web ID: could not extract ID from URL %q", input)
	}

	// Step 2: match against known forms

	// thread-f:<decimal>
	if m := reThreadF.FindStringSubmatch(bare); m != nil {
		dec, err := parseUint64Decimal(m[1])
		if err != nil {
			return nil, fmt.Errorf("gmail web ID: thread-f decimal parse of %q: %w", bare, err)
		}
		return &ResolvedWebID{
			Kind:     WebIDKindThreadF,
			ThreadID: hexStr(dec),
		}, nil
	}

	// msg-f:<decimal>
	if m := reMsgF.FindStringSubmatch(bare); m != nil {
		dec, err := parseUint64Decimal(m[1])
		if err != nil {
			return nil, fmt.Errorf("gmail web ID: msg-f decimal parse of %q: %w", bare, err)
		}
		return &ResolvedWebID{
			Kind:      WebIDKindMsgF,
			MessageID: hexStr(dec),
		}, nil
	}

	// Already an API ID (hex)
	if reAPIID.MatchString(bare) {
		return &ResolvedWebID{
			Kind:      WebIDKindAPIID,
			ThreadID:  bare,
			MessageID: bare,
		}, nil
	}

	// FMfcg… base64url form
	if reFMfcg.MatchString(bare) {
		threadID, messageID, err := decodeFMfcg(bare)
		if err != nil {
			return nil, fmt.Errorf("gmail web ID: FMfcg decode of %q: %w", bare, err)
		}
		return &ResolvedWebID{
			Kind:      WebIDKindFMfcg,
			ThreadID:  threadID,
			MessageID: messageID,
		}, nil
	}

	return nil, fmt.Errorf("gmail web ID: unrecognised ID form %q (not a Gmail URL, thread-f:, msg-f:, FMfcg…, or API hex ID)", bare)
}
