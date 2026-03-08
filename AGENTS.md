# sound-cistern Development Guidelines

Auto-generated from all feature plans. Last updated: 2025-10-21

## ⚠️ Important: Dev Server Management
**DO NOT start or stop the dev server.** The user manages the dev server manually. If you need to check if the server is running, use `curl` to check the health endpoint. Never use `pkill`, `kill`, or similar commands to stop the server.

## Active Technologies
- Go 1.23 + PocketBase v0.22.0, SQLite (embedded), React 18 + Vite + Tailwind CSS, Soundcloud API

## Project Structure
```
views/           # Templ templates (*.templ) — legacy, kept for reference
v2/              # React 18 frontend (Vite + Tailwind + Zustand)
pb_data/         # PocketBase data (SQLite DB)
pb_migrations/   # Database migrations (Go files)
public/          # Static assets (CSS, JS, images)
specs/           # Feature specifications
docs/            # Documentation
skill/           # OpenCode skills for development assistance
```

## Commands
```bash
make templ       # Generate Templ templates
make build       # Build binary: ./sound-cistern
make dev         # Development server with --dev flag
make serve       # Production server
make clean       # Clean build artifacts
make v2-build    # Build React v2 frontend
make v2-dev      # React dev server (port 5173, proxies to :8090)
```

## Testing
```bash
make test              # Run unit tests (mock functions)
make test-integration  # Run integration tests (API mocks)
make test-all          # Run all Go tests (unit + integration)
make test-browser      # Run browser automation tests (Chromedp)
make test-e2e          # Run E2E tests (starts test server on port 8090)
make test-server-start # Start test server in TEST_MODE
```

### E2E Testing Infrastructure
- `TEST_MODE=true` enables auto-authentication (no JWT token needed)
- `pb_data_test/` directory for isolated test database
- Mock Soundcloud API responses for testing without external dependencies
- Test server auto-starts/stops with `make test-e2e`

## Code Style
Go 1.23: Follow standard conventions
- Use React + JSX for frontend components
- PocketBase patterns for database operations
- Use Zustand for state management, Framer Motion for animations

## Recent Changes
- Migrated from Buffalo framework to PocketBase v0.22.0
- Migrated from Plush templates to Templ v0.3.960
- Migrated from PostgreSQL to SQLite (embedded)
- Updated from Go 1.21 to Go 1.23
- Migrated frontend from HTMX/Templ/Pico.css to React 18/Vite/Tailwind (v2 UI)

## Development Philosophy
- Write code like a river - clean, flowing, always moving
- Every change should make the code better
- Performance is a feature, not an afterthought
- Favor conventions over configurations
- Documentation is as important as code

## Browser Management (IMPORTANT)
- Use curl instead of Playwright for quick checks (e.g., checking if server is up, testing API responses)
- After every Playwright use, call playwright_browser_close to prevent memory leaks
- Kill runaway Chrome processes if they freeze: `pkill -9 -f "chrome.*mcp"`
- Check for stuck processes: `ps aux | grep chrome`

## Ralph Wiggum Autonomous Loop Support

This project supports Ralph Wiggum-style autonomous development loops. The skills are discovered automatically but must be loaded on-demand - see:
- `ralph-autonomous-loops` skill for loop implementation
- `ralph-commands` skill for user control commands

## Available Agent Skills

The following skills are available for development assistance (use `/skill <name>` to load):

1. **pocketbase-templ** - PocketBase development with Templ templates
2. **go-saas-template** - Building SaaS applications with Go and PocketBase
3. **soundcloud-oauth** - Soundcloud OAuth 2.0/2.1 integration
4. **pocketbase-htmx** - ⚠️ LEGACY: HTMX integration (deprecated, kept for reference only)
5. **sound-cistern-ux** - ⚠️ LEGACY: UI/UX design system using Pico.css (deprecated, v2 uses Tailwind)
6. **soundcloud-api** - Complete Soundcloud API reference with endpoint documentation
7. **soundcloud-api-favorites** - Soundcloud Favorites/Likes API endpoints
8. **soundcloud-api-stream** - Soundcloud Stream/Activities API endpoints
9. **soundcloud-api-tracks** - Soundcloud Tracks API endpoints

**Note:** Skills are discovered automatically from `.agents/skills/` but must be loaded into context using the `/skill` command when needed.

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->