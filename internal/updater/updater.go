package updater

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/jesperpedersen/picky-claude/internal/config"
)

// SelfUpdate checks for a new version and replaces the current binary.
// Returns the new version string if updated, or empty string if already current.
func SelfUpdate() (string, error) {
	info, err := CheckLatest("", "")
	if err != nil {
		return "", fmt.Errorf("check latest: %w", err)
	}

	if !IsNewer(config.Version(), info.Version) {
		return "", nil
	}

	url := info.AssetURL(runtime.GOOS, runtime.GOARCH, config.BinaryName)
	if url == "" {
		return "", fmt.Errorf("no binary available for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("locate current binary: %w", err)
	}

	if err := downloadBinary(url, execPath); err != nil {
		return "", fmt.Errorf("download update: %w", err)
	}

	return info.Version, nil
}

// downloadBinary downloads a URL to a destination file path, making it executable.
func downloadBinary(url, dest string) error {
	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned %d", resp.StatusCode)
	}

	// Write to a temp file first, then rename for atomicity.
	tmpFile := dest + ".tmp"
	f, err := os.Create(tmpFile)
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}

	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		os.Remove(tmpFile)
		return fmt.Errorf("write binary: %w", err)
	}
	f.Close()

	if err := os.Chmod(tmpFile, 0o755); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("chmod: %w", err)
	}

	if err := os.Rename(tmpFile, dest); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("rename: %w", err)
	}

	return nil
}

// UpdateInstructions returns a human-readable message with manual update steps.
func UpdateInstructions(latestVersion string) string {
	return fmt.Sprintf(`A new version of %s is available: %s (current: %s)

Update manually:
  go install github.com/%s/%s/cmd/%s@latest

Or download from:
  https://github.com/%s/%s/releases/latest`,
		config.DisplayName, latestVersion, config.Version(),
		defaultOwner, defaultRepo, config.BinaryName,
		defaultOwner, defaultRepo,
	)
}
