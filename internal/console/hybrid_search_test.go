package console

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestHybridSearch(t *testing.T) {
	srv := testServer(t)

	// Seed observations
	doRequest(t, srv, "POST", "/api/observations", map[string]string{
		"session_id": "s1", "type": "bugfix", "title": "auth bug",
		"text": "Fixed authentication login flow",
	})
	doRequest(t, srv, "POST", "/api/observations", map[string]string{
		"session_id": "s1", "type": "feature", "title": "db update",
		"text": "Database schema migration",
	})

	// Reindex
	rr := doRequest(t, srv, "POST", "/api/search/reindex", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("reindex status = %d, body = %s", rr.Code, rr.Body.String())
	}

	// Hybrid search
	rr = doRequest(t, srv, "GET", "/api/observations/hybrid-search?q=authentication", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("hybrid search status = %d, body = %s", rr.Code, rr.Body.String())
	}

	var results []map[string]any
	json.NewDecoder(rr.Body).Decode(&results)
	if len(results) == 0 {
		t.Error("expected results from hybrid search")
	}
}

func TestHybridSearchMissingQuery(t *testing.T) {
	srv := testServer(t)
	rr := doRequest(t, srv, "GET", "/api/observations/hybrid-search", nil)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestReindexEmpty(t *testing.T) {
	srv := testServer(t)
	rr := doRequest(t, srv, "POST", "/api/search/reindex", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("reindex status = %d, body = %s", rr.Code, rr.Body.String())
	}
}
