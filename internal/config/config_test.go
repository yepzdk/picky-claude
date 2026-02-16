package config

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func TestVersion(t *testing.T) {
	v := Version()
	if v == "" {
		t.Error("Version() returned empty string")
	}
}

func TestHomeDir_Default(t *testing.T) {
	t.Setenv(EnvPrefix+"_HOME", "")
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine user home dir")
	}
	got := HomeDir()
	want := filepath.Join(home, ConfigDirName)
	if got != want {
		t.Errorf("HomeDir() = %q, want %q", got, want)
	}
}

func TestHomeDir_EnvOverride(t *testing.T) {
	t.Setenv(EnvPrefix+"_HOME", "/tmp/test-picky")
	got := HomeDir()
	if got != "/tmp/test-picky" {
		t.Errorf("HomeDir() = %q, want %q", got, "/tmp/test-picky")
	}
}

func TestDBPath(t *testing.T) {
	t.Setenv(EnvPrefix+"_HOME", "/tmp/test-picky")
	got := DBPath()
	want := "/tmp/test-picky/db/picky.db"
	if got != want {
		t.Errorf("DBPath() = %q, want %q", got, want)
	}
}

func TestSessionDir(t *testing.T) {
	t.Setenv(EnvPrefix+"_HOME", "/tmp/test-picky")
	got := SessionDir("abc123")
	want := "/tmp/test-picky/sessions/abc123"
	if got != want {
		t.Errorf("SessionDir() = %q, want %q", got, want)
	}
}

func TestLoad_Defaults(t *testing.T) {
	t.Setenv(EnvPrefix+"_PORT", "")
	t.Setenv(EnvPrefix+"_LOG_LEVEL", "")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.Port != DefaultPort {
		t.Errorf("Port = %d, want %d", cfg.Port, DefaultPort)
	}
	if cfg.LogLevel != slog.LevelInfo {
		t.Errorf("LogLevel = %v, want %v", cfg.LogLevel, slog.LevelInfo)
	}
}

func TestLoad_CustomPort(t *testing.T) {
	t.Setenv(EnvPrefix+"_PORT", "9999")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.Port != 9999 {
		t.Errorf("Port = %d, want 9999", cfg.Port)
	}
}

func TestLoad_InvalidPort(t *testing.T) {
	t.Setenv(EnvPrefix+"_PORT", "notanumber")
	_, err := Load()
	if err == nil {
		t.Error("Load() should return error for invalid port")
	}
}

func TestLoad_LogLevels(t *testing.T) {
	tests := []struct {
		env  string
		want slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"DEBUG", slog.LevelDebug},
		{"warn", slog.LevelWarn},
		{"warning", slog.LevelWarn},
		{"error", slog.LevelError},
		{"info", slog.LevelInfo},
		{"", slog.LevelInfo},
		{"unknown", slog.LevelInfo},
	}
	for _, tt := range tests {
		t.Run(tt.env, func(t *testing.T) {
			t.Setenv(EnvPrefix+"_LOG_LEVEL", tt.env)
			cfg, err := Load()
			if err != nil {
				t.Fatalf("Load() error: %v", err)
			}
			if cfg.LogLevel != tt.want {
				t.Errorf("LogLevel = %v, want %v", cfg.LogLevel, tt.want)
			}
		})
	}
}
