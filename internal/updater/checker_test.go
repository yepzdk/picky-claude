package updater

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsNewer(t *testing.T) {
	tests := []struct {
		current string
		latest  string
		want    bool
	}{
		{"v0.1.0", "v0.2.0", true},
		{"v0.2.0", "v0.2.0", false},
		{"v0.3.0", "v0.2.0", false},
		{"v1.0.0", "v1.0.1", true},
		{"v1.0.0", "v2.0.0", true},
		{"dev", "v1.0.0", true},
		{"v1.0.0", "dev", false},
		{"dev", "dev", false},
	}

	for _, tt := range tests {
		t.Run(tt.current+"->"+tt.latest, func(t *testing.T) {
			got := IsNewer(tt.current, tt.latest)
			if got != tt.want {
				t.Errorf("IsNewer(%q, %q) = %v, want %v", tt.current, tt.latest, got, tt.want)
			}
		})
	}
}

func TestCheckLatestRelease(t *testing.T) {
	release := githubRelease{
		TagName: "v1.2.3",
		Assets: []releaseAsset{
			{Name: "picky-darwin-arm64", BrowserDownloadURL: "https://example.com/picky-darwin-arm64"},
			{Name: "picky-linux-amd64", BrowserDownloadURL: "https://example.com/picky-linux-amd64"},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(release)
	}))
	defer srv.Close()

	info, err := checkLatestFromURL(srv.URL)
	if err != nil {
		t.Fatalf("checkLatestFromURL() error: %v", err)
	}
	if info.Version != "v1.2.3" {
		t.Errorf("Version = %q, want %q", info.Version, "v1.2.3")
	}
	if len(info.Assets) != 2 {
		t.Errorf("Assets count = %d, want 2", len(info.Assets))
	}
}

func TestCheckLatestRelease_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	_, err := checkLatestFromURL(srv.URL)
	if err == nil {
		t.Error("expected error for server error response")
	}
}

func TestAssetURL(t *testing.T) {
	info := &ReleaseInfo{
		Version: "v1.0.0",
		Assets: []Asset{
			{Name: "picky-darwin-arm64", URL: "https://example.com/a"},
			{Name: "picky-linux-amd64", URL: "https://example.com/b"},
		},
	}

	url := info.AssetURL("darwin", "arm64", "picky")
	if url != "https://example.com/a" {
		t.Errorf("AssetURL(darwin, arm64) = %q, want https://example.com/a", url)
	}

	url = info.AssetURL("linux", "amd64", "picky")
	if url != "https://example.com/b" {
		t.Errorf("AssetURL(linux, amd64) = %q, want https://example.com/b", url)
	}

	url = info.AssetURL("windows", "amd64", "picky")
	if url != "" {
		t.Errorf("AssetURL(windows, amd64) = %q, want empty", url)
	}
}
