package logging

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/gobuffalo/envy"
)

// Config holds the logging configuration
type Config struct {
	// LogLevel sets the minimum log level (debug, info, warn, error)
	LogLevel string

	// LogFilePath is the path to the main log file
	LogFilePath string

	// ErrorLogPath is the path to error-only logs
	ErrorLogPath string

	// AuditLogPath is the path to audit/security logs
	AuditLogPath string

	// EnableFileOutput determines if logs should be written to files
	EnableFileOutput bool

	// EnableConsoleOutput determines if logs should be written to console
	EnableConsoleOutput bool

	// Environment is the current environment (development, test, production)
	Environment string
}

// NewConfig creates a new logging configuration with environment-based defaults
func NewConfig() *Config {
	env := os.Getenv("GO_ENV")
	if env == "" {
		env = "development"
	}

	// Default log directory
	logDir := envy.Get("LOG_DIR", "logs")

	// Ensure log directory exists
	if err := os.MkdirAll(logDir, 0755); err != nil {
		// Fallback to current directory if logs dir can't be created
		logDir = "."
	}

	config := &Config{
		LogLevel:            envy.Get("LOG_LEVEL", getDefaultLogLevel(env)),
		LogFilePath:         envy.Get("LOG_FILE_PATH", filepath.Join(logDir, "application.log")),
		ErrorLogPath:        envy.Get("ERROR_LOG_PATH", filepath.Join(logDir, "error.log")),
		AuditLogPath:        envy.Get("AUDIT_LOG_PATH", filepath.Join(logDir, "audit.log")),
		EnableFileOutput:    envy.Get("LOG_FILE_ENABLED", "true") == "true",
		EnableConsoleOutput: envy.Get("LOG_CONSOLE_ENABLED", getDefaultConsoleOutput(env)) == "true",
		Environment:         env,
	}

	return config
}

// getDefaultLogLevel returns the default log level based on environment
func getDefaultLogLevel(env string) string {
	switch strings.ToLower(env) {
	case "production":
		return "info"
	case "test":
		return "warn"
	default: // development
		return "debug"
	}
}

// getDefaultConsoleOutput returns whether console output should be enabled by default
func getDefaultConsoleOutput(env string) string {
	switch strings.ToLower(env) {
	case "production":
		return "false" // In production, typically only file output
	default:
		return "true" // Development and test environments show console output
	}
}

// IsValidLogLevel checks if the provided log level is valid
func IsValidLogLevel(level string) bool {
	validLevels := []string{"debug", "info", "warn", "error"}
	level = strings.ToLower(level)

	for _, validLevel := range validLevels {
		if level == validLevel {
			return true
		}
	}
	return false
}
