package logging

import (
	"os"
	"testing"

	"github.com/gobuffalo/envy"
)

func TestDebugEnvironment(t *testing.T) {
	// Clear any existing GO_ENV
	oldEnv := os.Getenv("GO_ENV")
	defer os.Setenv("GO_ENV", oldEnv)

	// Test 1: Default environment
	os.Unsetenv("GO_ENV")
	env1 := envy.Get("GO_ENV", "development")
	t.Logf("Default environment: %s", env1)

	// Test 2: Set to production
	os.Setenv("GO_ENV", "production")
	env2 := envy.Get("GO_ENV", "development")
	t.Logf("Production environment: %s", env2)

	// Test 3: Create config with production
	config := NewConfig()
	t.Logf("Config environment: %s, LogLevel: %s", config.Environment, config.LogLevel)
}
