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
			Name:       "favorites",
			Type:       models.CollectionTypeBase,
			ListRule:   types.Pointer("@request.auth.id = user_id"),
			ViewRule:   types.Pointer("@request.auth.id = user_id"),
			CreateRule: types.Pointer("@request.auth.id = user_id"),
			UpdateRule: nil,
			DeleteRule: types.Pointer("@request.auth.id = user_id"),
			Schema: schema.NewSchema(
				&schema.SchemaField{
					Name:     "user_id",
					Type:     schema.FieldTypeRelation,
					Required: true,
					Options: &schema.RelationOptions{
						MaxSelect:     types.Pointer(1),
						CollectionId:  "users",
						CascadeDelete: true,
					},
				},
				&schema.SchemaField{
					Name:     "track_id",
					Type:     schema.FieldTypeRelation,
					Required: true,
					Options: &schema.RelationOptions{
						MaxSelect:     types.Pointer(1),
						CollectionId:  "soundcloud_tracks",
						CascadeDelete: true,
					},
				},
			),
		}

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db)
		collection, err := dao.FindCollectionByNameOrId("favorites")
		if err != nil {
			return err
		}
		return dao.DeleteCollection(collection)
	})
}
