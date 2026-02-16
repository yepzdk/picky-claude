package session

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestConsoleClientPost(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/api/plans" {
			t.Errorf("path = %q, want /api/plans", r.URL.Path)
		}
		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		if body["path"] != "test.md" {
			t.Errorf("path = %q, want test.md", body["path"])
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]int{"id": 1})
	}))
	defer srv.Close()

	client := NewConsoleClient(srv.URL)
	resp, err := client.Post("/api/plans", map[string]string{
		"path": "test.md", "session_id": "s1", "status": "PENDING",
	})
	if err != nil {
		t.Fatalf("Post: %v", err)
	}
	defer resp.Body.Close()

	if !called {
		t.Error("handler not called")
	}
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("status = %d, want 201", resp.StatusCode)
	}
}

func TestConsoleClientGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}
		json.NewEncoder(w).Encode([]map[string]string{{"id": "s1"}})
	}))
	defer srv.Close()

	client := NewConsoleClient(srv.URL)
	resp, err := client.Get("/api/sessions")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
}

func TestConsoleClientBaseURL(t *testing.T) {
	client := NewConsoleClient("http://localhost:41777")
	if client.BaseURL() != "http://localhost:41777" {
		t.Errorf("BaseURL() = %q", client.BaseURL())
	}
}

func TestDefaultConsoleClient(t *testing.T) {
	client := DefaultConsoleClient(41777)
	if client.BaseURL() != "http://localhost:41777" {
		t.Errorf("BaseURL() = %q", client.BaseURL())
	}
}
