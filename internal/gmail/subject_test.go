package gmail

import (
	"context"
	"mime"
	"net/mail"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/gmail/v1"
)

func assertSubject(t *testing.T, raw, want string) string {
	t.Helper()
	msg, err := mail.ReadMessage(strings.NewReader(decodeRawEmail(t, raw)))
	if err != nil {
		t.Fatal(err)
	}
	header := msg.Header.Get("Subject")
	for _, c := range header {
		if c > 127 {
			t.Errorf("raw Subject is not ASCII: %q", header)
			break
		}
	}
	decoded, err := new(mime.WordDecoder).DecodeHeader(header)
	if err != nil || decoded != want {
		t.Fatalf("decoded subject %q, error %v; want %q", decoded, err, want)
	}
	return header
}

func TestOutgoingUnicodeSubject(t *testing.T) {
	for _, subject := range []string{"ASCII subject", "Meeting — tomorrow", "Café résumé", "会議の予定", strings.Repeat("長い件名", 40)} {
		for name, handler := range map[string]func(context.Context, mcp.CallToolRequest, *GmailHandlerDeps) (*mcp.CallToolResult, error){
			"send": TestableGmailSend, "draft": TestableGmailDraft, "update_draft": TestableGmailUpdateDraft,
		} {
			t.Run(name+"/"+subject, func(t *testing.T) {
				fixtures := NewGmailTestFixtures()
				fixtures.MockService.AddDraft(&gmail.Draft{Id: "draft1", Message: &gmail.Message{Id: "draft-message"}})
				result, err := handler(context.Background(), makeRequest(map[string]any{"to": "recipient@example.com", "subject": subject, "body": "Synthetic body", "draft_id": "draft1"}), fixtures.Deps)
				if err != nil || result.IsError {
					t.Fatalf("result %v, error %v", result, err)
				}
				var raw string
				switch name {
				case "send":
					raw = lastGmailMessageArg(t, fixtures, "SendMessage").Raw
				case "draft":
					raw = lastGmailDraftArg(t, fixtures, "CreateDraft").Message.Raw
				default:
					raw = fixtures.MockService.GetLastCall().Args[1].(*gmail.Draft).Message.Raw
				}
				assertSubject(t, raw, subject)
			})
		}
		t.Run("multipart/"+subject, func(t *testing.T) {
			raw, err := buildEmailMessageWithAttachments(EmailMessage{Subject: subject, Body: "Body"}, []emailAttachment{{Filename: "test.txt", MIMEType: "text/plain", Data: []byte("fixture")}})
			if err != nil {
				t.Fatal(err)
			}
			assertSubject(t, raw, subject)
		})
	}
}

func TestReplySubjectRoundTrip(t *testing.T) {
	for _, original := range []string{"Meeting — café 会議", mime.QEncoding.Encode("utf-8", "Re: Meeting — café 会議"), mime.BEncoding.Encode("utf-8", "Meeting — café 会議")} {
		t.Run(original, func(t *testing.T) {
			subject := original
			for i := 0; i < 3; i++ {
				fixtures := NewGmailTestFixtures()
				msg := newTestMessage("original", "thread", subject, "sender@example.com", "recipient@example.com", "Body", nil)
				fixtures.MockService.AddMessage(msg)
				result, err := TestableGmailReply(context.Background(), makeRequest(map[string]any{"message_id": "original", "body": "Reply"}), fixtures.Deps)
				if err != nil || result.IsError {
					t.Fatalf("result %v, error %v", result, err)
				}
				subject = assertSubject(t, lastGmailMessageArg(t, fixtures, "SendMessage").Raw, "Re: Meeting — café 会議")
				received := newTestMessage("received", "thread", subject, "sender@example.com", "recipient@example.com", "Body", nil)
				formatted := FormatMessageWithOptions(received, FormatMessageOptions{HeaderMode: HeaderModeRaw})
				if got := formatted["headers"].(map[string]string)["subject"]; got != "Re: Meeting — café 会議" {
					t.Errorf("displayed subject %q", got)
				}
				if got := formatted["payload_headers"].([]map[string]string)[0]["value"]; got != subject {
					t.Errorf("raw header modified: %q", got)
				}
			}
		})
	}
}

func TestDecodeSubjectOnce(t *testing.T) {
	for _, tc := range []struct{ input, want string }{
		{"ASCII subject", "ASCII subject"},
		{"=?utf-8?q?caf=C3=A9?=", "café"},
		{"=?iso-8859-1?q?caf=E9?=", "café"},
		{"=?utf-8?b?not-valid-base64!?=", "=?utf-8?b?not-valid-base64!?="},
		{"=?unknown?q?abc?=", "=?unknown?q?abc?="},
		{"=?utf-8?q?=3D=3Futf-8=3Fq=3Finner=3F=3D?=", "=?utf-8?q?inner?="},
	} {
		if got := decodeSubject(tc.input); got != tc.want {
			t.Errorf("decodeSubject(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}
