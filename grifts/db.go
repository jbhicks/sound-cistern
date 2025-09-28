package grifts

import (
	"fmt"
	"github.com/jbhicks/sound-cistern/models"

	"github.com/gobuffalo/grift/grift"
	"github.com/gobuffalo/pop/v6"
)

var _ = grift.Namespace("db", func() {

	grift.Desc("seed", "Seeds a database")
	grift.Add("seed", func(c *grift.Context) error {
		// Add DB seeding stuff here
		return nil
	})

	grift.Desc("promote_admin", "Promotes the first user (by email) to admin role")
	grift.Add("promote_admin", func(c *grift.Context) error {
		db, err := pop.Connect("development")
		if err != nil {
			return err
		}
		defer db.Close()

		// Find the first user
		user := &models.User{}
		if err := db.Order("created_at asc").First(user); err != nil {
			fmt.Println("No users found to promote")
			return err
		}

		// Update user role to admin
		user.Role = "admin"
		if err := db.Update(user); err != nil {
			return err
		}

		fmt.Printf("Successfully promoted user %s (%s) to admin\n", user.Email, user.FirstName+" "+user.LastName)
		return nil
	})

})
