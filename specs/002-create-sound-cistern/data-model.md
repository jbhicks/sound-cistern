# Data Model: Sound Cistern

## PocketBase Collections

### soundcloud_users
**Collection Type**: Base
**Migration**: `pb_migrations/1696000001_create_soundcloud_users.go`

Fields:
- id: Auto-generated (PocketBase 15-char ID)
- soundcloud_id: Text, required, unique (max 255)
- access_token: Text, required (max 500, encrypted at app layer)
- created: Auto-timestamp (PocketBase default)
- updated: Auto-timestamp (PocketBase default)

Rules:
- ListRule: null (admin only)
- ViewRule: null (admin only)
- CreateRule: null (server-side only)
- UpdateRule: null (server-side only)
- DeleteRule: null (admin only)

Relationships: Has many soundcloud_tracks, has one soundcloud_feed

### soundcloud_tracks
**Collection Type**: Base
**Migration**: `pb_migrations/1696000002_create_soundcloud_tracks.go`

Fields:
- id: Auto-generated (PocketBase 15-char ID)
- user_id: Relation (soundcloud_users), required, cascade delete
- soundcloud_id: Text, required (max 255)
- title: Text, required (max 500)
- length: Number, optional (min 0, seconds)
- genre: Text, optional (max 255, comma-separated)
- post_time: Date, optional
- created: Auto-timestamp (PocketBase default)
- updated: Auto-timestamp (PocketBase default)

Rules:
- ListRule: "" (authenticated users can list)
- ViewRule: "" (authenticated users can view)
- CreateRule: null (server-side only)
- UpdateRule: null (server-side only)
- DeleteRule: null (server-side only)

Relationships: Belongs to soundcloud_users

### soundcloud_feeds
**Collection Type**: Base
**Migration**: `pb_migrations/1696000003_create_soundcloud_feeds.go`

Fields:
- id: Auto-generated (PocketBase 15-char ID)
- user_id: Relation (soundcloud_users), required, unique, cascade delete
- tracks: JSON, optional (array of track IDs or embedded objects)
- created: Auto-timestamp (PocketBase default)
- updated: Auto-timestamp (PocketBase default)

Rules:
- ListRule: "" (authenticated users can list)
- ViewRule: "" (authenticated users can view)
- CreateRule: null (server-side only)
- UpdateRule: null (server-side only)
- DeleteRule: null (server-side only)

Relationships: Belongs to soundcloud_users (1:1 via unique constraint)

## Validation Rules
- soundcloud_users.soundcloud_id: Unique (PocketBase enforced)
- soundcloud_tracks.length: Min 0 (PocketBase enforced)
- soundcloud_feeds.updated: Auto-managed by PocketBase for cache validity checks

## State Transitions
- soundcloud_users: Created on Soundcloud OAuth callback
- soundcloud_tracks: Created/updated on feed fetch from Soundcloud API
- soundcloud_feeds: Refreshed periodically, trimmed to tracks from last 2 weeks

## Implementation Status
✅ All three collections have migrations created
❌ OAuth flow, API client, feed logic not implemented