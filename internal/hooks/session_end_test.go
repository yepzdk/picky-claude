package hooks

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jesperpedersen/picky-claude/internal/session"
)

func TestSessionEndHookRegistered(t *testing.T) {
	_, ok := registry["session-end"]
	if !ok {
		t.Error("session-end hook not registered")
	}
}

func TestPostSummary(t *testing.T) {
	tests := []struct {
		name           string
		responseStatus int
		sessionID      string
		wantErr        bool
	}{
		{
			name:           "posts summary successfully",
			responseStatus: http.StatusCreated,
			sessionID:      "test-session-123",
			wantErr:        false,
		},
		{
			name:           "returns error on bad request",
			responseStatus: http.StatusBadRequest,
			sessionID:      "test-session-456",
			wantErr:        true,
		},
		{
			name:           "handles server error gracefully",
			responseStatus: http.StatusInternalServerError,
			sessionID:      "test-session-789",
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var receivedPayload map[string]string

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request
				if r.Method != "POST" {
					t.Errorf("expected POST request, got %s", r.Method)
				}
				if r.URL.Path != "/api/summaries" {
					t.Errorf("expected /api/summaries path, got %s", r.URL.Path)
				}
				if ct := r.Header.Get("Content-Type"); ct != "application/json" {
					t.Errorf("expected Content-Type application/json, got %s", ct)
				}

				// Read payload
				body, _ := io.ReadAll(r.Body)
				json.Unmarshal(body, &receivedPayload)

				w.WriteHeader(tt.responseStatus)
				if tt.responseStatus == http.StatusCreated {
					json.NewEncoder(w).Encode(map[string]int64{"id": 1})
				}
			}))
			defer server.Close()

			// Test postSummary
			client := session.NewConsoleClient(server.URL)
			err := postSummary(client, tt.sessionID)
			if (err != nil) != tt.wantErr {
				t.Errorf("postSummary() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				// Verify payload structure
				if receivedPayload["session_id"] != tt.sessionID {
					t.Errorf("expected session_id=%s, got %s", tt.sessionID, receivedPayload["session_id"])
				}
				expectedText := "Session " + tt.sessionID + " ended"
				if receivedPayload["text"] != expectedText {
					t.Errorf("expected text=%q, got %q", expectedText, receivedPayload["text"])
				}
			}
		})
	}
}

func TestPostSummaryWithUnreachableServer(t *testing.T) {
	// Create and immediately close a server to guarantee connection refusal
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	server.Close()

	client := session.NewConsoleClient(server.URL)
	err := postSummary(client, "test-session")
	if err == nil {
		t.Error("expected error when server is unreachable")
	}
}
