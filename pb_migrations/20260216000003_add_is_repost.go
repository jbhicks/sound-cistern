package migrations

import (
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/models/schema"
)

func init() {
	m.Register(func(db dbx.Builder) error {
		dao := daos.New(db)

		collection, err := dao.FindCollectionByNameOrId("soundcloud_tracks")
		if err != nil {
			return err
		}

		collection.Schema.AddField(&schema.SchemaField{
			Name:     "is_repost",
			Type:     schema.FieldTypeBool,
			Required: false,
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
			if field.Name == "is_repost" {
				collection.Schema.RemoveField(field.Id)
				break
			}
		}

		return dao.SaveCollection(collection)
	})
}
