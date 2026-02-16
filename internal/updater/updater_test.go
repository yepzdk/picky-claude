package updater

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestDownloadBinary(t *testing.T) {
	content := []byte("fake-binary-content")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(content)
	}))
	defer srv.Close()

	tmpDir := t.TempDir()
	dest := filepath.Join(tmpDir, "picky")

	err := downloadBinary(srv.URL, dest)
	if err != nil {
		t.Fatalf("downloadBinary() error: %v", err)
	}

	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("ReadFile() error: %v", err)
	}
	if string(data) != string(content) {
		t.Errorf("content = %q, want %q", data, content)
	}

	info, err := os.Stat(dest)
	if err != nil {
		t.Fatalf("Stat() error: %v", err)
	}
	if info.Mode().Perm()&0o111 == 0 {
		t.Error("binary is not executable")
	}
}

func TestDownloadBinary_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	tmpDir := t.TempDir()
	dest := filepath.Join(tmpDir, "picky")

	err := downloadBinary(srv.URL, dest)
	if err == nil {
		t.Error("expected error for server error response")
	}
}

func TestUpdateInstructions(t *testing.T) {
	msg := UpdateInstructions("v1.2.3")
	if msg == "" {
		t.Error("UpdateInstructions() returned empty string")
	}
}
