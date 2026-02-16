package console

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestViewerServesHTML(t *testing.T) {
	srv := testServer(t)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET / status = %d, want %d", w.Code, http.StatusOK)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Picky Claude") {
		t.Error("response body does not contain 'Picky Claude'")
	}
}
