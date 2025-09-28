package models

import (
	"github.com/jbhicks/sound-cistern/pkg/logging"

	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/pop/v6"
)

// DB is a connection to your database to be used
// throughout your application.
var DB *pop.Connection

func init() {
	var err error
	env := envy.Get("GO_ENV", "development")
	DB, err = pop.Connect(env)
	if err != nil {
		logging.Fatal("Failed to connect to database", logging.Fields{
			"environment": env,
			"error":       err.Error(),
		})
	}
	pop.Debug = env == "development"
}
