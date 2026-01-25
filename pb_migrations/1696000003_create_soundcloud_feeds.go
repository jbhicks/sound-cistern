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
			Name:       "soundcloud_feeds",
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
					Unique:   true,
					Options: &schema.RelationOptions{
						MaxSelect:     types.Pointer(1),
						CollectionId:  "soundcloud_users",
						CascadeDelete: true,
					},
				},
				&schema.SchemaField{
					Name:     "tracks",
					Type:     schema.FieldTypeJson,
					Required: false,
				},
			),
		}

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db)
		collection, err := dao.FindCollectionByNameOrId("soundcloud_feeds")
		if err != nil {
			return err
		}
		return dao.DeleteCollection(collection)
	})
}
