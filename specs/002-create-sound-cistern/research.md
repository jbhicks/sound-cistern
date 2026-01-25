# Research: Sound Cistern

## Decisions

### Use PocketBase as backend framework
Decision: Use PocketBase v0.22.0 as the application backend framework.
Rationale: Provides embedded SQLite database, built-in authentication, real-time subscriptions, and admin UI out of the box. Single binary deployment simplifies infrastructure.
Alternatives considered: Buffalo framework (original plan), but migrated to PocketBase for simplicity and modern Go patterns.

### Soundcloud API Integration
Decision: Integrate Soundcloud API using OAuth 2.0 for authentication and REST API for fetching user feeds.
Rationale: Required for login and feed access; implement OAuth flow in PocketBase routes using standard Go HTTP client.
Alternatives considered: Third-party OAuth libraries, but PocketBase has built-in auth record patterns that simplify integration.

### Templ for Type-Safe Templates
Decision: Use Templ v0.3.960 for HTML template generation.
Rationale: Type-safe templates with compile-time checking, better IDE support than text templates, seamless Go integration.
Alternatives considered: html/template (Go standard library), but Templ provides superior developer experience and type safety.

### Styling with Pico.css
Decision: Use Pico.css for basic styling and find a beautiful example site to follow.
Rationale: Simple CSS framework for minimal, beautiful sites; class-less defaults work well with server-side rendering.
Alternatives considered: Custom CSS or Tailwind, but Pico.css is lightweight and fits simplicity principle.

### Database Choice
Decision: Use SQLite embedded with PocketBase.
Rationale: Single-file database, zero configuration, sufficient for small scale (1000 users), built-in to PocketBase.
Alternatives considered: PostgreSQL (original plan), but SQLite simplifies deployment and matches scale requirements.

### Minimal JavaScript with HTMX
Decision: Use HTMX for dynamic interactions, avoid heavy JavaScript frameworks.
Rationale: Server-side rendering remains primary approach; HTMX provides progressive enhancement for filtering without page reloads.
Alternatives considered: React/Vue (rejected for complexity), vanilla JS (HTMX provides cleaner declarative approach).