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

## ⚠️ CRITICAL: NEVER RUN THE DEV SERVER

**You (the agent) must NEVER run the development server.**

The user will handle running `make dev`, `make serve`, or any server commands. Your role is only to:
- Make code changes
- Run build commands (`make build`, `make templ`)
- Run test commands (`make test`, etc.)
- Read and analyze code

Do NOT run `make dev`, `make serve`, `make test-e2e`, or any command that starts the server. Wait for the user to do that.

## Project Overview

Sound Cistern is a Soundcloud aggregator that lets users authenticate via OAuth and browse their liked/favorited tracks. Built with Go, PocketBase, Templ, and HTMX.

## Development Environment

### ⚠️ DO NOT RUN - User handles this

The dev server runs on `http://localhost:8090` by default. **DO NOT run it yourself.**

If you need to verify changes work:
1. Make your code changes
2. Run `make build` to compile
3. Tell the user to restart the server to test

### Building the Project

```bash
# Generate Templ templates
make templ

# Build the binary
make build
```

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

**No deployment needed**: Changes are live immediately on https://soundcistern.jbhicks.dev when the user runs `make dev`.

**Testing against production**: Since prod IS the local dev server (run by the user):
- Use `serverURL = "https://soundcistern.jbhicks.dev"` for all tests
- No distinction between local and production tokens
- Database is shared (pb_data/)
- No need for separate test databases

**Warning**: This means:
- Database changes affect the "production" site immediately when the server is running
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

### Test Mode
Test mode is currently **disabled** in the codebase (hardcoded to return `false` in `main.go`). To re-enable:
1. Change `isTestMode()` to check `os.Getenv("TEST_MODE") == "true"`
2. Set `TEST_MODE=true` environment variable

Test mode bypasses OAuth and uses mock data for development without a SoundCloud account.

### Debug Parameter

Many routes support `?debug=1` to bypass authentication and use mock/fallback data:

| Route | Debug Behavior |
|-------|----------------|
| `/proto?debug=1` | Bypasses auth, uses tracks from database or fallback mock data |
| `/api/track/:id/stream?debug=1` | Uses first available SoundCloud token instead of requiring user auth |

This is useful for:
- Testing prototypes without logging in
- Browser automation testing
- Development when OAuth tokens are expired

Example:
```bash
curl "http://localhost:8090/proto?debug=1"
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
**DO NOT run this yourself.** If the user wants to clear data, they will do it themselves:
```bash
# User will run: Stop server, then:
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
| `/proto` | Unified prototype page |

## Prototyping

**Single route for all prototypes**: `/proto`

Usage:
- `/proto?type=track-cards` - Track card layout experiments
- `/proto?type=gradient` - Gradient vignette card designs
- `/proto?type=pill` - Pill-shaped card variations
- `/proto?type=ui` - General UI component experiments

To add a new prototype:
1. Add a new `ProtoType` constant in `views/proto.templ`
2. Add a case in the switch statement to render the prototype
3. Components go in `views/components/`

The template uses HTMX boost for seamless navigation between prototype types.

## Troubleshooting

### 502 Bad Gateway
**This is a user issue, not something you should fix.** The user is responsible for running the server. If they report this, remind them to start the server.

### OAuth Failures
**This is a user issue, not something you should fix.** The tunnel and server are the user's responsibility. If they report OAuth issues, remind them to:
- Ensure tunnel is running: `cloudflared tunnel list`
- Restart the server

### Zero Tracks After Login
1. Click "Sync Tracks" or POST to `/api/sync`
2. Check browser console for errors
3. Verify `soundcloud_tracks` collection has records in PocketBase admin

### Token Expiration
JWT tokens expire. Get fresh token from browser cookies (`pb_auth`) and update test env:
```bash
export TEST_JWT_TOKEN='<new-token>'
```
