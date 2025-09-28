package logging

import (
	"sync"

	"github.com/gobuffalo/buffalo"
)

var (
	// Default is the global logging service instance
	Default *Service
	once    sync.Once
)

// Init initializes the global logging service with the provided configuration
// If config is nil, it will use NewConfig() for defaults
func Init(config *Config) error {
	var err error
	once.Do(func() {
		if config == nil {
			config = NewConfig()
		}

		Default, err = NewService(config)
	})
	return err
}

// MustInit initializes the global logging service and panics on error
func MustInit(config *Config) {
	if err := Init(config); err != nil {
		panic("Failed to initialize logging service: " + err.Error())
	}
}

// GetDefault returns the global logging service instance
// If not initialized, it will initialize with default configuration
func GetDefault() *Service {
	if Default == nil {
		MustInit(nil)
	}
	return Default
}

// Convenience functions using the default logger

// Info logs an info level message using the default logger
func Info(msg string, fields ...Fields) {
	GetDefault().Info(msg, fields...)
}

// Debug logs a debug level message using the default logger
func Debug(msg string, fields ...Fields) {
	GetDefault().Debug(msg, fields...)
}

// Warn logs a warning level message using the default logger
func Warn(msg string, fields ...Fields) {
	GetDefault().Warn(msg, fields...)
}

// Error logs an error level message using the default logger
func Error(msg string, err error, fields ...Fields) {
	GetDefault().Error(msg, err, fields...)
}

// Fatal logs a fatal level message and exits the program using the default logger
func Fatal(msg string, fields ...Fields) {
	GetDefault().Fatal(msg, fields...)
}

// Audit logs security and administrative events using the default logger
func Audit(action string, fields Fields) {
	GetDefault().Audit(action, fields)
}

// UserAction logs user actions using the default logger
func UserAction(c buffalo.Context, actor string, action string, details string, fields ...Fields) {
	GetDefault().UserAction(c, actor, action, details, fields...)
}

// SecurityEvent logs security events using the default logger
func SecurityEvent(c buffalo.Context, eventType string, outcome string, reason string, fields ...Fields) {
	GetDefault().SecurityEvent(c, eventType, outcome, reason, fields...)
}
