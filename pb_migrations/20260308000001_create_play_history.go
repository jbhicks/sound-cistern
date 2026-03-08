package migrations

import (
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/models/schema"
	"github.com/pocketbase/pocketbase/tools/types"
)

func init() {
	m.Register(func(db dbx.Builder) error {
		dao := daos.New(db)

		collection := &models.Collection{
			Name:       "play_history",
			Type:       models.CollectionTypeBase,
			ListRule:   types.Pointer("@request.auth.id = user_id"),
			ViewRule:   types.Pointer("@request.auth.id = user_id"),
			CreateRule: types.Pointer("@request.auth.id = user_id"),
			UpdateRule: nil,
			DeleteRule: types.Pointer("@request.auth.id = user_id"),
			Schema: schema.NewSchema(
				&schema.SchemaField{
					Name:     "user_id",
					Type:     schema.FieldTypeText,
					Required: true,
				},
				&schema.SchemaField{
					Name:     "track_id",
					Type:     schema.FieldTypeText,
					Required: true,
				},
				&schema.SchemaField{
					Name:     "track_title",
					Type:     schema.FieldTypeText,
					Required: false,
				},
				&schema.SchemaField{
					Name:     "artist_name",
					Type:     schema.FieldTypeText,
					Required: false,
				},
				&schema.SchemaField{
					Name:     "artwork_url",
					Type:     schema.FieldTypeText,
					Required: false,
				},
				&schema.SchemaField{
					Name:     "track_duration",
					Type:     schema.FieldTypeNumber,
					Required: false,
				},
				&schema.SchemaField{
					Name:     "genre",
					Type:     schema.FieldTypeText,
					Required: false,
				},
				&schema.SchemaField{
					Name:     "played_at",
					Type:     schema.FieldTypeText,
					Required: true,
				},
			),
		}

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db)
		collection, err := dao.FindCollectionByNameOrId("play_history")
		if err != nil {
			return err
		}
		return dao.DeleteCollection(collection)
	})
}
