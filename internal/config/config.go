package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

// Config holds runtime configuration resolved from environment variables and defaults.
type Config struct {
	Port     int
	LogLevel slog.Level
}

// Load reads configuration from environment variables, falling back to defaults.
func Load() (*Config, error) {
	port := DefaultPort
	if v := os.Getenv(EnvPrefix + "_PORT"); v != "" {
		p, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid %s_PORT: %w", EnvPrefix, err)
		}
		port = p
	}

	level := parseLogLevel(os.Getenv(EnvPrefix + "_LOG_LEVEL"))

	return &Config{
		Port:     port,
		LogLevel: level,
	}, nil
}

// LevelOff is a log level above LevelError that suppresses all log output.
const LevelOff = slog.Level(16)

func parseLogLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return LevelOff
	}
}
