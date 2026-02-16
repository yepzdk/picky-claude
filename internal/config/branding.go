// Package config provides centralized configuration for the application.
//
// To rename the product, change the constants in this file, update the go.mod
// module path, and the BINARY_NAME in the Makefile. Everything else propagates
// automatically.
package config

const (
	// BinaryName is the CLI executable name.
	BinaryName = "picky"

	// DisplayName is the human-readable product name used in banners and docs.
	DisplayName = "Picky Claude"

	// EnvPrefix is prepended to environment variable names (e.g. PICKY_HOME).
	EnvPrefix = "PICKY"

	// ConfigDirName is the directory name under $HOME for storing data.
	// Resolved to a full path by paths.go.
	ConfigDirName = ".picky"
)
