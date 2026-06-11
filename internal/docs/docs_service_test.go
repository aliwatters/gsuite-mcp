package docs

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	googledocs "google.golang.org/api/docs/v1"
)

func TestRealDocsServiceBatchUpdateRawUsesDocsAPIEndpoint(t *testing.T) {
	requestsJSON := json.RawMessage(`[{"insertText":{"text":"Hello","location":{"index":1}}}]`)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want %s", r.Method, http.MethodPost)
		}
		if r.URL.Path != "/v1/documents/doc-123:batchUpdate" {
			t.Errorf("path = %s, want /v1/documents/doc-123:batchUpdate", r.URL.Path)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", got)
		}

		var body struct {
			Requests []json.RawMessage `json:"requests"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if len(body.Requests) != 1 {
			t.Fatalf("request count = %d, want 1", len(body.Requests))
		}
		if string(body.Requests[0]) != string(requestsJSON[1:len(requestsJSON)-1]) {
			t.Errorf("request body = %s, want %s", body.Requests[0], requestsJSON[1:len(requestsJSON)-1])
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"documentId":"doc-123","replies":[{}]}`))
	}))
	defer server.Close()

	srv := NewRealDocsServiceWithHTTP(&googledocs.Service{
		BasePath: server.URL + "/",
	}, nil, server.Client())

	resp, err := srv.BatchUpdateRaw(context.Background(), "doc-123", requestsJSON)
	if err != nil {
		t.Fatalf("BatchUpdateRaw returned error: %v", err)
	}
	if resp.DocumentId != "doc-123" {
		t.Errorf("DocumentId = %q, want doc-123", resp.DocumentId)
	}
	if len(resp.Replies) != 1 {
		t.Errorf("reply count = %d, want 1", len(resp.Replies))
	}
}
