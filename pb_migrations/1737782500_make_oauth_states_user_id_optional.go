package migrations

import (
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(db dbx.Builder) error {
		dao := daos.New(db)

		collection, err := dao.FindCollectionByNameOrId("oauth_states")
		if err != nil {
			return err
		}

		userIDField := collection.Schema.GetFieldByName("user_id")
		if userIDField != nil {
			userIDField.Required = false
			if err := dao.SaveCollection(collection); err != nil {
				return err
			}
		}

		return nil
	}, func(db dbx.Builder) error {
		return nil
	})
}
