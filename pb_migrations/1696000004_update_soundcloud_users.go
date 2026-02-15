package migrations

import (
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/models/schema"
	"github.com/pocketbase/pocketbase/tools/types"
)

func init() {
	m.Register(func(db dbx.Builder) error {
		dao := daos.New(db)

		collection, err := dao.FindCollectionByNameOrId("soundcloud_users")
		if err != nil {
			return err
		}

		// Add refresh_token field for OAuth 2.1 with PKCE
		collection.Schema.AddField(&schema.SchemaField{
			Name:     "refresh_token",
			Type:     schema.FieldTypeText,
			Required: false,
			Options: &schema.TextOptions{
				Max: types.Pointer(500),
			},
		})

		// Add expires_at field for token expiration tracking
		collection.Schema.AddField(&schema.SchemaField{
			Name:     "expires_at",
			Type:     schema.FieldTypeDate,
			Required: false,
		})

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db)

		collection, err := dao.FindCollectionByNameOrId("soundcloud_users")
		if err != nil {
			return err
		}

		// Remove the fields we added
		for i := len(collection.Schema.Fields()) - 1; i >= 0; i-- {
			field := collection.Schema.Fields()[i]
			if field.Name == "refresh_token" || field.Name == "expires_at" {
				collection.Schema.RemoveField(field.Id)
			}
		}

		return dao.SaveCollection(collection)
	})
}
