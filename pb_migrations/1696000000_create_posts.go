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
			Name:       "posts",
			Type:       models.CollectionTypeBase,
			ListRule:   types.Pointer("published = true"),
			ViewRule:   types.Pointer("published = true"),
			CreateRule: nil,
			UpdateRule: nil,
			DeleteRule: nil,
			Schema: schema.NewSchema(
				&schema.SchemaField{
					Name:     "title",
					Type:     schema.FieldTypeText,
					Required: true,
					Options: &schema.TextOptions{
						Max: types.Pointer(255),
					},
				},
				&schema.SchemaField{
					Name:     "slug",
					Type:     schema.FieldTypeText,
					Required: true,
					Unique:   true,
					Options: &schema.TextOptions{
						Max:     types.Pointer(255),
						Pattern: "^[a-z0-9-]+$",
					},
				},
				&schema.SchemaField{
					Name:     "content",
					Type:     schema.FieldTypeText,
					Required: false,
					Options: &schema.TextOptions{
						Max: nil,
					},
				},
				&schema.SchemaField{
					Name:     "excerpt",
					Type:     schema.FieldTypeText,
					Required: false,
					Options: &schema.TextOptions{
						Max: types.Pointer(500),
					},
				},
				&schema.SchemaField{
					Name:     "published",
					Type:     schema.FieldTypeBool,
					Required: false,
				},
				&schema.SchemaField{
					Name:     "author_id",
					Type:     schema.FieldTypeRelation,
					Required: true,
					Options: &schema.RelationOptions{
						MaxSelect:     types.Pointer(1),
						CollectionId:  "users",
						CascadeDelete: false,
					},
				},
				&schema.SchemaField{
					Name:     "meta_title",
					Type:     schema.FieldTypeText,
					Required: false,
					Options: &schema.TextOptions{
						Max: types.Pointer(255),
					},
				},
				&schema.SchemaField{
					Name:     "meta_description",
					Type:     schema.FieldTypeText,
					Required: false,
					Options: &schema.TextOptions{
						Max: types.Pointer(500),
					},
				},
				&schema.SchemaField{
					Name:     "meta_keywords",
					Type:     schema.FieldTypeText,
					Required: false,
					Options: &schema.TextOptions{
						Max: types.Pointer(500),
					},
				},
				&schema.SchemaField{
					Name:     "og_title",
					Type:     schema.FieldTypeText,
					Required: false,
					Options: &schema.TextOptions{
						Max: types.Pointer(255),
					},
				},
				&schema.SchemaField{
					Name:     "og_description",
					Type:     schema.FieldTypeText,
					Required: false,
					Options: &schema.TextOptions{
						Max: types.Pointer(500),
					},
				},
				&schema.SchemaField{
					Name:     "og_image",
					Type:     schema.FieldTypeText,
					Required: false,
					Options: &schema.TextOptions{
						Max: types.Pointer(500),
					},
				},
				&schema.SchemaField{
					Name:     "image",
					Type:     schema.FieldTypeFile,
					Required: false,
					Options: &schema.FileOptions{
						MaxSelect: 1,
						MaxSize:   5242880,
						MimeTypes: []string{"image/jpeg", "image/png", "image/webp"},
					},
				},
				&schema.SchemaField{
					Name:     "image_alt",
					Type:     schema.FieldTypeText,
					Required: false,
					Options: &schema.TextOptions{
						Max: types.Pointer(255),
					},
				},
			),
		}

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db)
		collection, err := dao.FindCollectionByNameOrId("posts")
		if err != nil {
			return err
		}
		return dao.DeleteCollection(collection)
	})
}
