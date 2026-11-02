package migrations

import (
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/models/schema"
	"github.com/pocketbase/pocketbase/tools/types"
)

// Security hardening (idempotent; safe to run on a DB that is already fixed):
//
//  1. Repair relation fields whose options.collectionId holds a collection
//     NAME instead of an id (older migrations used e.g. "soundcloud_users",
//     "users", "soundcloud_tracks"). PocketBase only resolves ids there, so the
//     admin UI could not save those collections and relation-traversing API
//     rules could not work. Targets are looked up by name, never hardcoded.
//  2. soundcloud_tracks: list/view were public (""), exposing every user's
//     synced tracks including stream/download URLs. soundcloud_tracks.user_id
//     references a soundcloud_users record (not a users record), so the owner
//     rule traverses that relation to the PocketBase user.
//  3. soundcloud_feeds: list/view were public; the app never reads it through
//     the REST API (server-side DAO access bypasses rules), so lock it.
//  4. users: enforce in code what was set by hand in the admin UI — no public
//     sign-up (createRule = admin only) and no username/email password auth
//     (SoundCloud OAuth only) — and forbid users from changing their own role.
const (
	soundcloudTracksOwnerRule = "@request.auth.id != '' && user_id.user_id = @request.auth.id"
	usersSelfUpdateRule       = "id = @request.auth.id && @request.data.role:isset = false"
)

func init() {
	m.Register(func(db dbx.Builder) error {
		dao := daos.New(db)

		if err := repairRelationCollectionIds(dao); err != nil {
			return err
		}

		if tracks, err := dao.FindCollectionByNameOrId("soundcloud_tracks"); err == nil {
			tracks.ListRule = types.Pointer(soundcloudTracksOwnerRule)
			tracks.ViewRule = types.Pointer(soundcloudTracksOwnerRule)
			tracks.CreateRule = nil
			tracks.UpdateRule = nil
			tracks.DeleteRule = nil
			if err := dao.SaveCollection(tracks); err != nil {
				return err
			}
		}

		if feeds, err := dao.FindCollectionByNameOrId("soundcloud_feeds"); err == nil {
			feeds.ListRule = nil
			feeds.ViewRule = nil
			feeds.CreateRule = nil
			feeds.UpdateRule = nil
			feeds.DeleteRule = nil
			if err := dao.SaveCollection(feeds); err != nil {
				return err
			}
		}

		users, err := dao.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}
		users.CreateRule = nil // admin only
		if users.Schema.GetFieldByName("role") != nil {
			users.UpdateRule = types.Pointer(usersSelfUpdateRule)
		}
		if users.IsAuth() {
			opts := users.AuthOptions()
			opts.AllowUsernameAuth = false
			opts.AllowEmailAuth = false
			opts.AllowOAuth2Auth = true
			if err := users.SetOptions(opts); err != nil {
				return err
			}
		}
		return dao.SaveCollection(users)
	}, func(db dbx.Builder) error {
		// Intentionally not reverting to the previous, insecure rules.
		return nil
	})
}

// repairRelationCollectionIds rewrites relation options.collectionId values
// that reference a collection by name to that collection's real id.
func repairRelationCollectionIds(dao *daos.Dao) error {
	collections := []*models.Collection{}
	if err := dao.CollectionQuery().All(&collections); err != nil {
		return err
	}

	byID := make(map[string]*models.Collection, len(collections))
	byName := make(map[string]*models.Collection, len(collections))
	for _, c := range collections {
		byID[c.Id] = c
		byName[c.Name] = c
	}

	for _, c := range collections {
		changed := false
		for _, f := range c.Schema.Fields() {
			if f.Type != schema.FieldTypeRelation {
				continue
			}
			if err := f.InitOptions(); err != nil {
				return err
			}
			opts, ok := f.Options.(*schema.RelationOptions)
			if !ok || opts == nil {
				continue
			}
			if _, valid := byID[opts.CollectionId]; valid {
				continue
			}
			if target, found := byName[opts.CollectionId]; found {
				opts.CollectionId = target.Id
				changed = true
			}
		}
		if changed {
			if err := dao.SaveCollection(c); err != nil {
				return err
			}
		}
	}
	return nil
}
