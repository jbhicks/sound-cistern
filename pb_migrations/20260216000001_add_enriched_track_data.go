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
			Name:     "playback_count",
			Type:     schema.FieldTypeNumber,
			Required: false,
		})

		collection.Schema.AddField(&schema.SchemaField{
			Name:     "favoritings_count",
			Type:     schema.FieldTypeNumber,
			Required: false,
		})

		collection.Schema.AddField(&schema.SchemaField{
			Name:     "comment_count",
			Type:     schema.FieldTypeNumber,
			Required: false,
		})

		collection.Schema.AddField(&schema.SchemaField{
			Name:     "reposts_count",
			Type:     schema.FieldTypeNumber,
			Required: false,
		})

		collection.Schema.AddField(&schema.SchemaField{
			Name:     "user_avatar_url",
			Type:     schema.FieldTypeText,
			Required: false,
			Options: &schema.TextOptions{
				Max: types.Pointer(500),
			},
		})

		collection.Schema.AddField(&schema.SchemaField{
			Name:     "tag_list",
			Type:     schema.FieldTypeText,
			Required: false,
			Options: &schema.TextOptions{
				Max: types.Pointer(1000),
			},
		})

		collection.Schema.AddField(&schema.SchemaField{
			Name:     "bpm",
			Type:     schema.FieldTypeNumber,
			Required: false,
		})

		collection.Schema.AddField(&schema.SchemaField{
			Name:     "release_year",
			Type:     schema.FieldTypeNumber,
			Required: false,
		})

		collection.Schema.AddField(&schema.SchemaField{
			Name:     "downloadable",
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

		fieldsToRemove := []string{
			"playback_count",
			"favoritings_count",
			"comment_count",
			"reposts_count",
			"user_avatar_url",
			"tag_list",
			"bpm",
			"release_year",
			"downloadable",
		}

		for i := len(collection.Schema.Fields()) - 1; i >= 0; i-- {
			field := collection.Schema.Fields()[i]
			for _, fieldName := range fieldsToRemove {
				if field.Name == fieldName {
					collection.Schema.RemoveField(field.Id)
					break
				}
			}
		}

		return dao.SaveCollection(collection)
	})
}
