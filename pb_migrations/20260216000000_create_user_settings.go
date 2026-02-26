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
			Name:       "user_settings",
			Type:       models.CollectionTypeBase,
			ListRule:   types.Pointer("@request.auth.id = user_id"),
			ViewRule:   types.Pointer("@request.auth.id = user_id"),
			CreateRule: types.Pointer("@request.auth.id = user_id"),
			UpdateRule: types.Pointer("@request.auth.id = user_id"),
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
					Name:     "active_tab",
					Type:     schema.FieldTypeSelect,
					Required: false,
					Options: &schema.SelectOptions{
						Values: []string{"stream", "favorites"},
					},
				},
				&schema.SchemaField{
					Name:     "filters",
					Type:     schema.FieldTypeJson,
					Required: false,
				},
				&schema.SchemaField{
					Name:     "view_mode",
					Type:     schema.FieldTypeSelect,
					Required: false,
					Options: &schema.SelectOptions{
						Values: []string{"grid", "list"},
					},
				},
			),
		}

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db)
		collection, err := dao.FindCollectionByNameOrId("user_settings")
		if err != nil {
			return err
		}
		return dao.DeleteCollection(collection)
	})
}
