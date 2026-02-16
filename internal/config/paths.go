package config

import (
	"os"
	"path/filepath"
)

// HomeDir returns the application's home directory.
// Uses $PICKY_HOME if set, otherwise ~/.picky (derived from ConfigDirName).
func HomeDir() string {
	if v := os.Getenv(EnvPrefix + "_HOME"); v != "" {
		return v
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ConfigDirName
	}
	return filepath.Join(home, ConfigDirName)
}

// DBDir returns the directory for SQLite database files.
func DBDir() string {
	return filepath.Join(HomeDir(), "db")
}

// DBPath returns the path to the main SQLite database.
func DBPath() string {
	return filepath.Join(DBDir(), "picky.db")
}

// SessionsDir returns the directory for session state files.
func SessionsDir() string {
	return filepath.Join(HomeDir(), "sessions")
}

// SessionDir returns the directory for a specific session.
func SessionDir(sessionID string) string {
	return filepath.Join(SessionsDir(), sessionID)
}

// LogDir returns the directory for log files.
func LogDir() string {
	return filepath.Join(HomeDir(), "logs")
}
