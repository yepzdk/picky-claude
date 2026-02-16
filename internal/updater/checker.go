// Package updater checks GitHub Releases for new versions and can self-update
// the binary.
package updater

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	defaultOwner = "jesperpedersen"
	defaultRepo  = "picky-claude"
)

// githubRelease is the JSON structure from the GitHub Releases API.
type githubRelease struct {
	TagName string         `json:"tag_name"`
	Assets  []releaseAsset `json:"assets"`
}

// releaseAsset is a single downloadable asset in a GitHub release.
type releaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// Asset is a simplified release asset with name and download URL.
type Asset struct {
	Name string
	URL  string
}

// ReleaseInfo holds the latest release version and its assets.
type ReleaseInfo struct {
	Version string
	Assets  []Asset
}

// AssetURL returns the download URL for the given OS/arch/binary combination.
// Returns empty string if no matching asset is found.
func (r *ReleaseInfo) AssetURL(goos, goarch, binaryName string) string {
	target := fmt.Sprintf("%s-%s-%s", binaryName, goos, goarch)
	for _, a := range r.Assets {
		if a.Name == target {
			return a.URL
		}
	}
	return ""
}

// CheckLatest queries the GitHub Releases API for the latest release.
func CheckLatest(owner, repo string) (*ReleaseInfo, error) {
	if owner == "" {
		owner = defaultOwner
	}
	if repo == "" {
		repo = defaultRepo
	}
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	return checkLatestFromURL(url)
}

func checkLatestFromURL(url string) (*ReleaseInfo, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github API returned %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("decode release: %w", err)
	}

	info := &ReleaseInfo{Version: release.TagName}
	for _, a := range release.Assets {
		info.Assets = append(info.Assets, Asset{Name: a.Name, URL: a.BrowserDownloadURL})
	}
	return info, nil
}

// IsNewer returns true if latest is a newer semver than current.
// Handles "dev" as the current version (always upgradeable).
func IsNewer(current, latest string) bool {
	if latest == "dev" || latest == "" {
		return false
	}
	if current == "dev" || current == "" {
		return true
	}

	curParts := parseSemver(current)
	latParts := parseSemver(latest)

	for i := 0; i < 3; i++ {
		if latParts[i] > curParts[i] {
			return true
		}
		if latParts[i] < curParts[i] {
			return false
		}
	}
	return false
}

// parseSemver extracts major, minor, patch from a version string like "v1.2.3".
func parseSemver(v string) [3]int {
	v = strings.TrimPrefix(v, "v")
	parts := strings.SplitN(v, ".", 3)
	var result [3]int
	for i, p := range parts {
		if i >= 3 {
			break
		}
		fmt.Sscanf(p, "%d", &result[i])
	}
	return result
}
