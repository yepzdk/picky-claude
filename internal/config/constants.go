package config

// version is set at build time via -ldflags.
var version = "dev"

const (
	// DefaultPort is the default console server port.
	DefaultPort = 41777

	// DefaultLogLevel is the default structured log level.
	// "off" disables all log output; users must set PICKY_LOG_LEVEL explicitly to see logs.
	DefaultLogLevel = "off"
)

// Version returns the build version string.
func Version() string {
	return version
}
