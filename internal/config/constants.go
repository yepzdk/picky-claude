package config

// version is set at build time via -ldflags.
var version = "dev"

const (
	// DefaultPort is the default console server port.
	DefaultPort = 41777

	// DefaultLogLevel is the default structured log level.
	DefaultLogLevel = "info"
)

// Version returns the build version string.
func Version() string {
	return version
}
