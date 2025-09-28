# Research: Sound Cistern

## Decisions

### Use my-go-saas-template as base
Decision: Use the template at https://github.com/jbhicks/my-go-saas-template to bootstrap the application.
Rationale: Provides authentication, database setup, and basic Buffalo structure out of the box, adhering to Buffalo Framework Adherence principle.
Alternatives considered: Starting from scratch with Buffalo CLI, but template saves time and follows best practices.

### Soundcloud API Integration
Decision: Integrate Soundcloud API using OAuth for authentication and REST API for fetching user feeds.
Rationale: Required for login and feed access; use official SDK if available, else HTTP client.
Alternatives considered: Custom OAuth implementation, but use template's auth if possible.

### Styling with Pico.css
Decision: Use Pico.css for basic styling and find a beautiful example site to follow.
Rationale: Simple CSS framework for minimal, beautiful sites; avoids JS frameworks.
Alternatives considered: Custom CSS, but Pico.css is lightweight and fits simplicity principle.

### Database Choice
Decision: Use PostgreSQL as provided by the template.
Rationale: Template uses it; suitable for caching feeds.
Alternatives considered: SQLite for simplicity, but Postgres for production.

### No JavaScript Framework
Decision: Stick to server-side rendering with Buffalo templates.
Rationale: Adhere to constraints; no crazy JS.
Alternatives considered: Add React/Vue, but rejected to keep simple.