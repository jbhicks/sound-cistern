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

		// Add user_id field to link PocketBase users to Soundcloud users
		collection.Schema.AddField(&schema.SchemaField{
			Name:     "user_id",
			Type:     schema.FieldTypeRelation,
			Required: true,
			Options: &schema.RelationOptions{
				MaxSelect:     types.Pointer(1),
				CollectionId:  "_pb_users_auth_",
				CascadeDelete: true,
			},
		})

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db)

		collection, err := dao.FindCollectionByNameOrId("soundcloud_users")
		if err != nil {
			return err
		}

		// Remove the user_id field
		for i := len(collection.Schema.Fields()) - 1; i >= 0; i-- {
			field := collection.Schema.Fields()[i]
			if field.Name == "user_id" {
				collection.Schema.RemoveField(field.Id)
			}
		}

		return dao.SaveCollection(collection)
	})
}
