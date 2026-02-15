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

		collection, err := dao.FindCollectionByNameOrId("soundcloud_tracks")
		if err != nil {
			return err
		}

		collection.Schema.AddField(&schema.SchemaField{
			Name:     "artist_name",
			Type:     schema.FieldTypeText,
			Required: false,
			Options: &schema.TextOptions{
				Max: types.Pointer(255),
			},
		})

		collection.Schema.AddField(&schema.SchemaField{
			Name:     "permalink_url",
			Type:     schema.FieldTypeText,
			Required: false,
			Options: &schema.TextOptions{
				Max: types.Pointer(500),
			},
		})

		collection.Schema.AddField(&schema.SchemaField{
			Name:     "artwork_url",
			Type:     schema.FieldTypeText,
			Required: false,
			Options: &schema.TextOptions{
				Max: types.Pointer(500),
			},
		})

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db)
		collection, err := dao.FindCollectionByNameOrId("soundcloud_tracks")
		if err != nil {
			return err
		}

		for i := len(collection.Schema.Fields()) - 1; i >= 0; i-- {
			field := collection.Schema.Fields()[i]
			if field.Name == "artist_name" || field.Name == "permalink_url" || field.Name == "artwork_url" {
				collection.Schema.RemoveField(field.Id)
			}
		}

		return dao.SaveCollection(collection)
	})
}
