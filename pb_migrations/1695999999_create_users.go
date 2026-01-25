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

		collection, err := dao.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}

		collection.Schema.AddField(&schema.SchemaField{
			Name:     "first_name",
			Type:     schema.FieldTypeText,
			Required: false,
		})

		collection.Schema.AddField(&schema.SchemaField{
			Name:     "last_name",
			Type:     schema.FieldTypeText,
			Required: false,
		})

		collection.Schema.AddField(&schema.SchemaField{
			Name:     "role",
			Type:     schema.FieldTypeText,
			Required: true,
			Options: &schema.TextOptions{
				Max: types.Pointer(50),
			},
		})

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db)
		collection, err := dao.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}

		for i := len(collection.Schema.Fields()) - 1; i >= 0; i-- {
			field := collection.Schema.Fields()[i]
			if field.Name == "first_name" || field.Name == "last_name" || field.Name == "role" {
				collection.Schema.RemoveField(field.Id)
			}
		}

		return dao.SaveCollection(collection)
	})
}
