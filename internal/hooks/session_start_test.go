package hooks

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jesperpedersen/picky-claude/internal/session"
)

func TestSessionStartHookRegistered(t *testing.T) {
	_, ok := registry["session-start"]
	if !ok {
		t.Error("session-start hook not registered")
	}
}

func TestFetchContext(t *testing.T) {
	tests := []struct {
		name           string
		responseBody   map[string]string
		rawBody        string
		responseStatus int
		wantContext    string
		wantErr        bool
	}{
		{
			name:           "returns context from server",
			responseBody:   map[string]string{"context": "Test context from server"},
			responseStatus: http.StatusOK,
			wantContext:    "Test context from server",
			wantErr:        false,
		},
		{
			name:           "returns empty string when context is empty",
			responseBody:   map[string]string{"context": ""},
			responseStatus: http.StatusOK,
			wantContext:    "",
			wantErr:        false,
		},
		{
			name:           "returns error on server failure",
			responseStatus: http.StatusInternalServerError,
			wantContext:    "",
			wantErr:        true,
		},
		{
			name:           "returns error on invalid JSON",
			rawBody:        "invalid json",
			responseStatus: http.StatusOK,
			wantContext:    "",
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request
				if r.Method != "GET" {
					t.Errorf("expected GET request, got %s", r.Method)
				}
				if r.URL.Path != "/api/context/inject" {
					t.Errorf("expected /api/context/inject path, got %s", r.URL.Path)
				}

				// Send response
				w.WriteHeader(tt.responseStatus)
				if tt.responseBody != nil {
					json.NewEncoder(w).Encode(tt.responseBody)
				} else if tt.rawBody != "" {
					w.Write([]byte(tt.rawBody))
				}
			}))
			defer server.Close()

			// Test fetchContext
			client := session.NewConsoleClient(server.URL)
			got, err := fetchContext(client, "test-session")
			if (err != nil) != tt.wantErr {
				t.Errorf("fetchContext() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.wantContext {
				t.Errorf("fetchContext() = %q, want %q", got, tt.wantContext)
			}
		})
	}
}

func TestFetchContextIncludesSessionID(t *testing.T) {
	tests := []struct {
		name      string
		sessionID string
	}{
		{"plain session ID", "my-session-id"},
		{"session ID with spaces", "session with spaces"},
		{"session ID with special chars", "id&foo=bar#baz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var receivedSessionID string

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				receivedSessionID = r.URL.Query().Get("session_id")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]string{"context": "test"})
			}))
			defer server.Close()

			client := session.NewConsoleClient(server.URL)
			fetchContext(client, tt.sessionID)

			if receivedSessionID != tt.sessionID {
				t.Errorf("expected session_id=%q, got %q", tt.sessionID, receivedSessionID)
			}
		})
	}
}
