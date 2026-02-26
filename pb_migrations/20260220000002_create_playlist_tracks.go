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
			Name:       "playlist_tracks",
			Type:       models.CollectionTypeBase,
			ListRule:   types.Pointer("@request.auth.id = added_by"),
			ViewRule:   types.Pointer("@request.auth.id = added_by || @request.auth.id = playlist_id.user_id"),
			CreateRule: types.Pointer("@request.auth.id = added_by"),
			UpdateRule: types.Pointer("@request.auth.id = playlist_id.user_id"),
			DeleteRule: types.Pointer("@request.auth.id = playlist_id.user_id || @request.auth.id = added_by"),
			Schema: schema.NewSchema(
				&schema.SchemaField{
					Name:     "playlist_id",
					Type:     schema.FieldTypeRelation,
					Required: true,
					Options: &schema.RelationOptions{
						MaxSelect:     types.Pointer(1),
						CollectionId:  "playlists",
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
				&schema.SchemaField{
					Name:     "position",
					Type:     schema.FieldTypeNumber,
					Required: true,
					Options: &schema.NumberOptions{
						Min: types.Pointer(0.0),
					},
				},
				&schema.SchemaField{
					Name:     "added_by",
					Type:     schema.FieldTypeRelation,
					Required: true,
					Options: &schema.RelationOptions{
						MaxSelect:     types.Pointer(1),
						CollectionId:  "users",
						CascadeDelete: true,
					},
				},
			),
			Indexes: types.JsonArray[string]{
				"CREATE UNIQUE INDEX idx_playlist_tracks_playlist_track ON playlist_tracks (playlist_id, track_id)",
				"CREATE INDEX idx_playlist_tracks_playlist_position ON playlist_tracks (playlist_id, position)",
			},
		}

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db)
		collection, err := dao.FindCollectionByNameOrId("playlist_tracks")
		if err != nil {
			return err
		}
		return dao.DeleteCollection(collection)
	})
}
