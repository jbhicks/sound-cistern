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
			Name:       "soundcloud_tracks",
			Type:       models.CollectionTypeBase,
			ListRule:   types.Pointer(""),
			ViewRule:   types.Pointer(""),
			CreateRule: nil,
			UpdateRule: nil,
			DeleteRule: nil,
			Schema: schema.NewSchema(
				&schema.SchemaField{
					Name:     "user_id",
					Type:     schema.FieldTypeRelation,
					Required: true,
					Options: &schema.RelationOptions{
						MaxSelect:     types.Pointer(1),
						CollectionId:  "soundcloud_users",
						CascadeDelete: true,
					},
				},
				&schema.SchemaField{
					Name:     "soundcloud_id",
					Type:     schema.FieldTypeText,
					Required: true,
					Options: &schema.TextOptions{
						Max: types.Pointer(255),
					},
				},
				&schema.SchemaField{
					Name:     "title",
					Type:     schema.FieldTypeText,
					Required: true,
					Options: &schema.TextOptions{
						Max: types.Pointer(500),
					},
				},
				&schema.SchemaField{
					Name:     "length",
					Type:     schema.FieldTypeNumber,
					Required: false,
					Options: &schema.NumberOptions{
						Min: types.Pointer(0.0),
					},
				},
				&schema.SchemaField{
					Name:     "genre",
					Type:     schema.FieldTypeText,
					Required: false,
					Options: &schema.TextOptions{
						Max: types.Pointer(255),
					},
				},
				&schema.SchemaField{
					Name:     "post_time",
					Type:     schema.FieldTypeDate,
					Required: false,
				},
			),
		}

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db)
		collection, err := dao.FindCollectionByNameOrId("soundcloud_tracks")
		if err != nil {
			return err
		}
		return dao.DeleteCollection(collection)
	})
}
