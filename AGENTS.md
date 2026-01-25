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

This project supports Ralph Wiggum-style autonomous development loops for complex, iterative tasks.

### Available Ralph Skills

- **ralph-autonomous-loops** - Custom implementation for OpenCode-based Ralph loops
- **pocketbase-templ** - PocketBase + Templ development patterns
- **pocketbase-htmx** - HTMX integration with PocketBase applications
- **soundcloud-oauth** - Complete Soundcloud OAuth 2.0/2.1 integration
- **go-saas-template** - Multi-tenant SaaS application patterns

### Ralph Project Structure
```
.ralph/
├── config/           # Loop configuration and settings
├── tasks/            # Task definitions and backlog
├── verify/            # Automated verification scripts
├── context/            # Context persistence between iterations
├── progress/          # Session tracking and history
└── state/             # Lock files and current state
```

### Ralph Usage Examples

**Start autonomous development loop:**
```bash
opencode ralph-autonomous-loops "Implement Soundcloud OAuth integration" \
    --max-iterations 30 \
    --cost-limit 15.00
```

**Monitor loop progress:**
```bash
ralph --status
```

**Add context guidance:**
```bash
ralph --add-context "Focus on OAuth callback implementation"
```

### Ralph Benefits

- **Autonomous Development**: AI can work for hours without intervention
- **Iterative Improvement**: Each iteration builds on previous work
- **Context Persistence**: Maintain state across development sessions
- **Quality Assurance**: Automated testing and verification gates
- **Cost Control**: Built-in iteration limits and cost monitoring

### Documentation References

- **Ralph Philosophy**: Based on Geoffrey Huntley's "bash loop" methodology
- **OpenCode Integration**: Specific patterns for opencode CLI agents
- **Best Practices**: Safety controls, progress tracking, troubleshooting

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->