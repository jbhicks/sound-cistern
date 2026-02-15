# sound-cistern Development Guidelines

Auto-generated from all feature plans. Last updated: 2025-10-21

## Active Technologies
- Go 1.23 + PocketBase v0.22.0, SQLite (embedded), Templ v0.3.960, Soundcloud API

## Project Structure
```
views/           # Templ templates (*.templ)
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
```

## Code Style
Go 1.23: Follow standard conventions
- Use Templ for type-safe templates
- PocketBase patterns for database operations
- HTMX for dynamic frontend interactions

## Recent Changes
- Migrated from Buffalo framework to PocketBase v0.22.0
- Migrated from Plush templates to Templ v0.3.960
- Migrated from PostgreSQL to SQLite (embedded)
- Updated from Go 1.21 to Go 1.23

## Development Philosophy
- Write code like a river - clean, flowing, always moving
- Every change should make the code better
- Performance is a feature, not an afterthought
- Favor conventions over configurations
- Documentation is as important as code

## Ralph Wiggum Autonomous Loop Support

This project supports Ralph Wiggum-style autonomous development loops. The skills are discovered automatically but must be loaded on-demand - see:
- `ralph-autonomous-loops` skill for loop implementation
- `ralph-commands` skill for user control commands

## Available Agent Skills

The following skills are available for development assistance (use `/skill <name>` to load):

1. **pocketbase-templ** - PocketBase development with Templ templates
2. **go-saas-template** - Building SaaS applications with Go and PocketBase
3. **soundcloud-oauth** - Soundcloud OAuth 2.0/2.1 integration
4. **pocketbase-htmx** - HTMX integration with PocketBase
5. **sound-cistern-ux** - UI/UX design system using Pico.css and Primer design principles

**Note:** Skills are discovered automatically from `.agents/skills/` but must be loaded into context using the `/skill` command when needed.

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->