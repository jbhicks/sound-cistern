---
name: sound-cistern-dev
description: Sound Cistern development setup including local dev server, Cloudflare tunnel, and common tasks
license: MIT
compatibility: opencode
metadata:
  version: "1.0"
  audience: sound-cistern-developers
  stack: go-pocketbase-templ
---

# Sound Cistern Development Skill

## Project Overview

Sound Cistern is a Soundcloud aggregator that lets users authenticate via OAuth and browse their liked/favorited tracks. Built with Go, PocketBase, Templ, and HTMX.

## Development Environment

### Running Locally

```bash
# Development server with hot reload
make dev

# Production server
make serve
```

The dev server runs on `http://localhost:8090` by default.

### Cloudflare Tunnel (Prod == Dev Setup)

**CRITICAL**: The public URL https://soundcistern.jbhicks.dev tunnels directly to the local dev server. There is no separate production deployment - prod IS dev.

- **Public URL**: https://soundcistern.jbhicks.dev (tunnels to localhost)
- **Local URL**: http://localhost:8090
- **Database**: Same SQLite database for both (pb_data/)
- **Tokens**: Same JWT tokens work for both URLs

The tunnel must be running for OAuth callbacks to work. All code changes are immediately live on the public URL.

**Start the tunnel**:
```bash
# Using cloudflared (one-time setup)
cloudflared tunnel --url http://localhost:8090

# Or if tunnel is already configured
cloudflared tunnel run sound-cistern
```

**Tunnel Configuration** (if already set up):
- Check `~/.cloudflared/config.yml` or Cloudflare dashboard
- Tunnel name: `sound-cistern`

### Environment Variables

Required for OAuth (check `.env` or set manually):
- `SOUNDCLOUD_CLIENT_ID`
- `SOUNDCLOUD_CLIENT_SECRET`
- `SOUNDCLOUD_REDIRECT_URI` - must be `https://soundcistern.jbhicks.dev/auth/soundcloud/callback`

### Development Implications (Prod == Dev)

**No deployment needed**: Changes are live immediately on https://soundcistern.jbhicks.dev when you run `make dev`.

**Testing against production**: Since prod IS the local dev server:
- Use `serverURL = "https://soundcistern.jbhicks.dev"` for all tests
- No distinction between local and production tokens
- Database is shared (pb_data/)
- No need for separate test databases

**Warning**: This means:
- Database changes affect the "production" site immediately
- All users see dev changes in real-time
- Good for rapid iteration, but be careful with schema migrations

### Database

- **Backend**: SQLite (PocketBase embedded)
- **Data Location**: `pb_data/`
- **Migrations**: `pb_migrations/`

Run migrations after schema changes:
```bash
# PocketBase handles migrations automatically on startup
# Or manually via PocketBase CLI
./sound-cistern migrate
```

## Common Development Tasks

### Generate Templ Templates
```bash
make templ
```

### Build Binary
```bash
make build
```

### Run Tests
```bash
# All tests
go test ./...

# Integration tests (hits live server)
go test -tags=integration -run TestAuthenticatedEndpoints

# With JWT token
TEST_JWT_TOKEN=<token> go test -tags=integration -run TestAuthenticatedEndpoints
```

### Clear Development Data
```bash
# Stop server, then:
rm -rf pb_data/
make serve  # Will recreate fresh DB
```

## API Endpoints

| Endpoint | Description |
|----------|-------------|
| `/` | Home - redirects to stream or login |
| `/stream` | Main track list (requires auth) |
| `/auth/soundcloud` | Start OAuth flow |
| `/auth/soundcloud/callback` | OAuth callback |
| `/api/stream` | Get tracks (HTMX) |
| `/api/sync` | Sync tracks from Soundcloud |
| `/api/favorites` | Get favorites |
| `/api/favorites/:id/toggle` | Toggle favorite |

## Troubleshooting

### 502 Bad Gateway
- Check if `make dev` or `make serve` is running
- Check if Cloudflare tunnel is active

### OAuth Failures
- Ensure tunnel is running: `cloudflared tunnel list`
- Verify redirect URI matches in Soundcloud app settings
- Check logs for `oauth_failed` errors

### Zero Tracks After Login
1. Click "Sync Tracks" or POST to `/api/sync`
2. Check browser console for errors
3. Verify `soundcloud_tracks` collection has records in PocketBase admin

### Token Expiration
JWT tokens expire. Get fresh token from browser cookies (`pb_auth`) and update test env:
```bash
export TEST_JWT_TOKEN='<new-token>'
```
