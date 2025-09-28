package logging

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewConfig(t *testing.T) {
	// Clean up any existing env vars
	os.Unsetenv("LOG_LEVEL")
	os.Unsetenv("LOG_FILE_PATH")
	os.Unsetenv("GO_ENV")

	config := NewConfig()

	// Should have default values
	if config.LogLevel != "debug" {
		t.Errorf("Expected default log level 'debug', got '%s'", config.LogLevel)
	}

	if config.Environment != "development" {
		t.Errorf("Expected default environment 'development', got '%s'", config.Environment)
	}

	if !config.EnableConsoleOutput {
		t.Error("Expected console output to be enabled by default in development")
	}

	if !config.EnableFileOutput {
		t.Error("Expected file output to be enabled by default")
	}
}

func TestNewConfigProduction(t *testing.T) {
	// Store original environment
	oldEnv := os.Getenv("GO_ENV")
	defer func() {
		if oldEnv == "" {
			os.Unsetenv("GO_ENV")
		} else {
			os.Setenv("GO_ENV", oldEnv)
		}
	}()

	// Set production environment
	os.Setenv("GO_ENV", "production")

	// Create new config
	config := NewConfig()

	if config.LogLevel != "info" {
		t.Errorf("Expected production log level 'info', got '%s'", config.LogLevel)
	}

	if config.Environment != "production" {
		t.Errorf("Expected environment 'production', got '%s'", config.Environment)
	}
}

func TestNewService(t *testing.T) {
	// Create a temporary directory for test logs
	tempDir, err := os.MkdirTemp("", "test-logs")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := &Config{
		LogLevel:            "debug",
		LogFilePath:         filepath.Join(tempDir, "test.log"),
		ErrorLogPath:        filepath.Join(tempDir, "error.log"),
		AuditLogPath:        filepath.Join(tempDir, "audit.log"),
		EnableFileOutput:    true,
		EnableConsoleOutput: false,
		Environment:         "test",
	}

	service, err := NewService(config)
	if err != nil {
		t.Fatalf("Failed to create logging service: %v", err)
	}

	if service == nil {
		t.Fatal("Expected non-nil logging service")
	}

	// Test basic logging
	service.Info("Test info message")
	service.Debug("Test debug message")
	service.Warn("Test warning message")

	// Check that log file was created
	if _, err := os.Stat(config.LogFilePath); os.IsNotExist(err) {
		t.Error("Expected log file to be created")
	}
}

func TestLogLevels(t *testing.T) {
	validLevels := []string{"debug", "info", "warn", "error"}
	invalidLevels := []string{"trace", "fatal", "invalid"}

	for _, level := range validLevels {
		if !IsValidLogLevel(level) {
			t.Errorf("Expected '%s' to be a valid log level", level)
		}
	}

	for _, level := range invalidLevels {
		if IsValidLogLevel(level) {
			t.Errorf("Expected '%s' to be an invalid log level", level)
		}
	}
}
