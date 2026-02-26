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

		usersCollection, err := dao.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}

		collection := &models.Collection{
			Name:       "playlists",
			Type:       models.CollectionTypeBase,
			ListRule:   types.Pointer("@request.auth.id = user_id"),
			ViewRule:   types.Pointer("@request.auth.id = user_id || share_token != ''"),
			CreateRule: types.Pointer("@request.auth.id = user_id"),
			UpdateRule: types.Pointer("@request.auth.id = user_id"),
			DeleteRule: types.Pointer("@request.auth.id = user_id"),
			Schema: schema.NewSchema(
				&schema.SchemaField{
					Name:     "name",
					Type:     schema.FieldTypeText,
					Required: true,
					Options: &schema.TextOptions{
						Max: types.Pointer(100),
					},
				},
				&schema.SchemaField{
					Name:     "description",
					Type:     schema.FieldTypeText,
					Required: false,
					Options: &schema.TextOptions{
						Max: types.Pointer(500),
					},
				},
				&schema.SchemaField{
					Name:     "user_id",
					Type:     schema.FieldTypeRelation,
					Required: true,
					Options: &schema.RelationOptions{
						MaxSelect:     types.Pointer(1),
						CollectionId:  usersCollection.Id,
						CascadeDelete: true,
					},
				},
				&schema.SchemaField{
					Name:     "share_token",
					Type:     schema.FieldTypeText,
					Required: false,
					Options: &schema.TextOptions{
						Max: types.Pointer(64),
					},
				},
				&schema.SchemaField{
					Name:     "track_count",
					Type:     schema.FieldTypeNumber,
					Required: false,
					Options: &schema.NumberOptions{
						Min: types.Pointer(0.0),
					},
				},
			),
			Indexes: types.JsonArray[string]{
				"CREATE UNIQUE INDEX idx_playlists_user_name ON playlists (user_id, name)",
				"CREATE INDEX idx_playlists_share_token ON playlists (share_token) WHERE share_token != ''",
			},
		}

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db)
		collection, err := dao.FindCollectionByNameOrId("playlists")
		if err != nil {
			return err
		}
		return dao.DeleteCollection(collection)
	})
}
