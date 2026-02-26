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
			Name:       "oauth_states",
			Type:       models.CollectionTypeBase,
			ListRule:   nil,
			ViewRule:   nil,
			CreateRule: nil,
			UpdateRule: nil,
			DeleteRule: nil,
			Schema: schema.NewSchema(
				&schema.SchemaField{
					Name:     "state",
					Type:     schema.FieldTypeText,
					Required: true,
					Unique:   true,
					Options: &schema.TextOptions{
						Max: types.Pointer(255),
					},
				},
				&schema.SchemaField{
					Name:     "code_verifier",
					Type:     schema.FieldTypeText,
					Required: true,
					Options: &schema.TextOptions{
						Max: types.Pointer(255),
					},
				},
				&schema.SchemaField{
					Name:     "user_id",
					Type:     schema.FieldTypeRelation,
					Required: false,
					Options: &schema.RelationOptions{
						MaxSelect:     types.Pointer(1),
						CollectionId:  "_pb_users_auth_",
						CascadeDelete: true,
					},
				},
				&schema.SchemaField{
					Name:     "expires_at",
					Type:     schema.FieldTypeDate,
					Required: true,
				},
			),
		}

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db)
		collection, err := dao.FindCollectionByNameOrId("oauth_states")
		if err != nil {
			return err
		}
		return dao.DeleteCollection(collection)
	})
}
