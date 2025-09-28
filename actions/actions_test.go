package actions

import (
	"os"
	"sync"
	"testing"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/suite/v4"
)

type ActionSuite struct {
	*suite.Action
}

func Test_ActionSuite(t *testing.T) {
	// Ensure we're running in test environment to disable CSRF
	os.Setenv("GO_ENV", "test")

	// Reset the app instance so it gets recreated with test environment
	appOnce = sync.Once{}
	app = nil

	// Create app instance to verify environment
	testApp := App()

	as := &ActionSuite{
		Action: suite.NewAction(testApp),
	}

	// Add middleware to set test_mode flag for debug logging
	as.Action.App.Use(func(next buffalo.Handler) buffalo.Handler {
		return func(c buffalo.Context) error {
			c.Set("test_mode", true)
			return next(c)
		}
	})

	suite.Run(t, as)
}
