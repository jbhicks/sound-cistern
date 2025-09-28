# My Go SaaS Template

A Buffalo web application template with PostgreSQL database, user authentication, admin panel, and HTMX-based UI.

## ✅ Features Checklist

### Core Infrastructure
- [x] **Buffalo application** - Go web framework with hot reload
- [x] **PostgreSQL database** - Podman Compose setup with PostgreSQL 15
- [x] **Database configuration** - Development, test, production environments
- [x] **Database migrations** - User table with role field
- [x] **Development workflow** - Make commands for common tasks

### Authentication & Authorization
- [x] **Authentication system** - User registration, login, logout with sessions
- [x] **Role-based access control** - User and admin roles
- [x] **User profile management** - Profile editing
- [x] **Password security** - bcrypt hashing

### Admin Management System
- [x] **Admin dashboard** - User statistics and basic admin interface
- [x] **User management** - User listing, editing, and deletion
- [x] **Role assignment** - Change user roles between user and admin
- [x] **Admin promotion** - Promote first user to admin via command
- [x] **Authorization middleware** - Admin-only route protection
- [x] **Safety controls** - Prevent admin self-deletion

### User Interface & Experience
- [x] **Template system** - Plush templates with Pico.css styling
- [x] **HTMX integration** - Dynamic content loading without full page refreshes
- [x] **Persistent UI shell** - Header and footer persist across navigation
- [x] **Modal authentication** - Login and signup forms in modals
- [x] **Theme system** - Dark/light/auto modes
- [x] **Responsive design** - Mobile-friendly layout
- [x] **UI components** - User dropdowns, admin tables, basic interface elements

### SEO & Performance
- [x] **SEO optimization** - Meta tags, Open Graph, structured data, XML sitemap
- [x] **Performance optimization** - Minimal CSS/JS footprint
- [x] **Accessibility** - Semantic HTML and keyboard navigation
- [x] **Search engine friendly** - robots.txt and canonical URLs

### Development & Testing
- [x] **Hot reload development** - Buffalo dev server with automatic recompilation
- [x] **Testing framework** - Go testing with database integration
- [x] **Database health checks** - PostgreSQL readiness verification
- [x] **Error handling** - Basic error handling in commands and scripts

### Blog & CMS System
- [x] **Blog post system** - CRUD operations for blog posts with admin management
- [x] **Post validation** - Title, content, and slug validation with excerpt generation
- [x] **Admin blog management** - Full admin interface for post creation, editing, and deletion
- [x] **SEO-friendly URLs** - Automatic slug generation from titles
- [x] **Post excerpts** - Automatic excerpt generation for blog listings

### CMS Enhancement Roadmap
- [x] **Rich Text Editor** - WYSIWYG editor (Quill.js) for post content with locally served assets
- [ ] **Media Library** - File upload system with image management and thumbnails
- [ ] **Content Categories/Tags** - Taxonomy system for content organization
- [ ] **Content Scheduling** - Publish date functionality for scheduled posts
- [x] **SEO Meta Fields** - Title tags, meta descriptions, keywords, and OpenGraph data for social media
- [x] **Bulk Operations** - Select multiple posts for bulk actions (publish/unpublish/delete)
- [x] **Content Search/Filtering** - Search posts by title, content, or author in admin with status filtering
- [x] **Draft System** - Save drafts before publishing with status management
- [ ] **Content Analytics** - View counts and popular posts dashboard
- [ ] **Content Types** - Extend beyond blog posts to pages, FAQs, etc.
- [ ] **Revision History** - Track content changes over time
- [ ] **Multi-language Support** - Content translation capabilities

### Pending Features
- [ ] **Billing/subscription features** - Payment processing and subscription management
- [ ] **Email services** - Transactional emails and notifications
- [ ] **Production deployment** - Docker containers and cloud deployment guides

## Quick Start

### Prerequisites
- **Go 1.19+** - [Download Go](https://golang.org/dl/)
- **Podman/Docker** - For PostgreSQL container ([Install Podman](https://podman.io/getting-started/installation))
- **Buffalo CLI** - `go install github.com/gobuffalo/cli/cmd/buffalo@latest`

### Setup

Setup steps for the application:

```console
# Clone the repository
git clone <your-repo-url>
cd my-go-saas-template

# Complete setup (database + migrations + first run)
make setup

# Start development mode
make dev
```

After setup, visit [http://127.0.0.1:3000](http://127.0.0.1:3000) to see your application running.

## Buffalo Auto-Reload Development

**Important**: Buffalo has built-in hot reload that automatically handles all file changes. Once you run `make dev`, the server stays running and automatically reloads when you make changes.

### How Auto-Reload Works
- **Go code changes** → Buffalo automatically recompiles and restarts the server
- **Template changes** → Templates reload instantly without server restart
- **Static assets** → CSS/JS changes update automatically via the asset pipeline
- **Database migrations** → Run migrations with `soda migrate up` while server runs

### Development Best Practices
- **Start once**: Run `make dev` at the beginning of your development session
- **Keep running**: Leave Buffalo running in the background throughout development
- **Just code**: Make your changes and refresh the browser to see updates
- **No manual restarts**: Buffalo handles all recompilation automatically

### When to Restart Buffalo
- **Compilation errors**: If Go code has syntax errors preventing compilation
- **Database issues**: If you need to reset the database state
- **Explicit request**: Only restart if specifically needed for debugging

**🚨 Never kill the Buffalo process during normal development** - it's designed to handle all changes automatically!

### First Admin User

To set up your first admin user:

1. **Register a user account** through the web interface
2. **Promote to admin** with one command:
   ```console
   make admin
   ```

This automatically promotes the first registered user to admin role, giving them access to the admin panel.

## 👑 Admin Management System

This template includes a role-based admin management system with basic CRUD operations and safety controls.

### Admin System Features

#### User Management
- **Basic CRUD Operations** - Create, read, update, and delete users
- **Role Assignment** - Change user roles between admin and user
- **User Management** - User listing with basic pagination
- **Safety Controls** - Admins cannot delete their own accounts

#### Admin Interface
- **Basic Dashboard** - User statistics and system overview
- **User Management Table** - User list with edit/delete actions
- **Role Management Forms** - Simple role assignment interface
- **Responsive Design** - Works on desktop and mobile devices

#### Security Features
- **Authorization Middleware** - Admin routes protected with `AdminRequired` middleware
- **Role-Based Access** - UI shows/hides features based on user permissions
- **Session Security** - Session management with role verification
- **Input Validation** - Basic validation for admin operations

### Setting Up Admin Access

#### Automatic Admin Promotion
```console
# Promote the first registered user to admin
make admin
```

This grift task finds the first user (by creation date) and promotes them to admin role.

#### Manual Admin Promotion
```console
# Using Buffalo task directly
buffalo task db:promote_admin

# Or promote a specific user via database
psql -d sound_cistern_development -c "UPDATE users SET role = 'admin' WHERE email = 'your-email@example.com';" 
```

### Admin Routes & API

| Route | Method | Description | Access Level |
|-------|--------|-------------|--------------|
| `/admin` | GET | Admin dashboard with statistics | Admin Only |
| `/admin/users` | GET | User management list (paginated) | Admin Only |
| `/admin/users/{id}` | GET | Edit user form | Admin Only |
| `/admin/users/{id}` | POST | Update user (including role) | Admin Only |
| `/admin/users/{id}` | DELETE | Delete user (with safety checks) | Admin Only |

### Role System Details

#### User Roles
- **`user`** (default) - Standard application access
  - Profile management
  - Dashboard access
  - Standard features

- **`admin`** - Full administrative privileges
  - All user permissions
  - Admin panel access (`/admin`)
  - User management capabilities
  - Role assignment permissions
  - System administration

#### Role Enforcement
- **Database Level** - Role field with proper constraints and validation
- **Middleware Level** - `AdminRequired` middleware protects admin routes
- **Template Level** - Conditional rendering based on user role
- **UI Level** - Dynamic navigation and feature visibility

### Admin Development Patterns

#### Adding New Admin Features
```go
// In actions/app.go - Add new admin routes
adminGroup := app.Group("/admin")
adminGroup.Use(AdminRequired)
adminGroup.GET("/new-feature", AdminNewFeatureHandler)
```

#### Template Access Control
```html
<!-- In templates - Check admin role -->
<%= if (current_user.Role == "admin") { %>
  <a href="/admin">Admin Panel</a>
<% } %>
```

#### Safety Checks Example
```go
// Prevent self-deletion
if userToDelete.ID == currentUser.ID {
    return c.Error(400, errors.New("cannot delete your own account"))
}
```

## 🛠️ Development Commands

### Quick Reference

| Command | Purpose | Description |
|---------|---------|-------------|
| `make setup` | First-time setup | Creates database, runs migrations |
| `make dev` | Development mode | Starts database + Buffalo dev server |
| `make admin` | Admin setup | Promotes first user to admin role |
| `make test` | Run tests | Executes full test suite with database |
| `make clean` | Cleanup | Stops services and cleans containers |
| `make db-status` | Health check | Shows database container status |

### Development Workflow

#### First Time Setup
```console
# Clone and setup the project
git clone <your-repo-url>
cd my-go-saas-template
make setup

# Create your first user account via the web interface
# Then promote to admin
make admin
```

#### Daily Development
```console
# Start development (runs database + Buffalo dev server)
make dev

# Buffalo automatically reloads on file changes
# Visit http://127.0.0.1:3000 to see your changes
```

#### Testing & Quality Assurance
```console
# Run all tests (includes database setup)
make test

# Check database health
make db-status

# Clean up after development
make clean
```

### Advanced Commands

#### Database Operations
```console
# Manual database management
make db-up                     # Start database only
make db-down                   # Stop database
make db-reset                  # Reset database (drop/create/migrate)
make migrate                   # Run migrations only

# Buffalo database commands
soda create -a                 # Create all databases
soda migrate up                # Run migrations
soda generate migration        # Create new migration
soda drop -e development       # Drop development database
```

#### Admin Management
```console
# Admin user management
buffalo task db:promote_admin  # Promote first user to admin
make admin                     # Same as above (via make)

# Manual role assignment via database
psql -d github.com/jbhicks/sound-cistern_development \
  -c "UPDATE users SET role = 'admin' WHERE email = 'user@example.com';"
```

#### Building & Production
```console
# Build for production
make build                     # Creates binary in bin/
buffalo build                  # Direct Buffalo build
buffalo build --static        # Static binary build

# Production database setup
GO_ENV=production soda create
GO_ENV=production soda migrate up
```

### Development Tips

#### Buffalo Development Server
- **Automatic reload** - Buffalo watches files and reloads automatically
- **Port 3000** - Default development port
- **Hot reload** - Template and Go code changes reload automatically
- **Keep running** - Leave Buffalo running, it handles recompilation

#### Database Development
- **Container persistence** - Database data persists between restarts
- **Health checks** - Make commands wait for database readiness
- **Multiple environments** - Development, test, and production databases
- **Migration tracking** - Buffalo tracks applied migrations automatically

#### Template Development
- **HTMX integration** - Templates support dynamic content loading
- **Pico.css styling** - Semantic HTML with automatic styling
- **Plush templating** - Buffalo's template engine with Go-like syntax
- **Live reload** - Template changes appear immediately

### Troubleshooting

#### Common Issues

**Database Connection Issues**
```console
# Check container status
make db-status
podman-compose ps

# Check logs
podman-compose logs postgres

# Reset database if corrupted
make db-reset
```

**Buffalo Issues**
```console
# Check if Buffalo is running
ps aux | grep buffalo
lsof -i :3000

# Restart Buffalo if needed
# Ctrl+C to stop, then: make dev
```

**Port Conflicts**
```console
# Check what's using port 3000
lsof -i :3000

# Kill process if needed
kill -9 $(lsof -t -i:3000)
```

**Template Errors**
- Check Buffalo console output for Plush syntax errors
- Ensure proper variable names and template structure
- Verify HTMX attributes and targets are correct

## 🔐 Authentication Features

### User Registration & Login
- **Registration**: `/users/new` - Create new user accounts (form loads in a modal via HTMX)
- **Login**: `/auth/new` - Sign in with email/password (form loads in a modal via HTMX)
- **Dashboard**: `/dashboard` - Protected area for authenticated users (content loads via HTMX)
- **Logout**: Available via user dropdown menu (uses HTMX POST)

### User Interface
- **Persistent Header/Footer**: The main site header (with navigation, theme toggle, profile/auth links) and footer are defined in `templates/home/index.plush.html` and persist across page views using `hx-preserve="true"`.
- **Dynamic Content Area**: The `<main id="htmx-content">` in `index.plush.html` is where page-specific content is dynamically loaded by HTMX.
- **Modal Authentication Forms**: "Login" and "Sign Up" buttons in the header trigger Pico.css modals. The respective forms are loaded into these modals via HTMX.
- **Landing Page**: Marketing page with conditional CTAs. Login/Signup CTAs open modals.
- **User Dropdown**: Professional dropdown menu in the persistent header for authenticated users with:
  - User avatar (initials with gradient background)
  - User name display
  - Profile Settings (placeholder, loads via HTMX)
  - Account Settings (placeholder, loads via HTMX)  
  - Sign Out functionality (HTMX POST request)

### Authentication Flow
1. Unauthenticated users see the landing page. "Login" and "Sign Up" buttons in the header open modals with the respective forms, loaded via HTMX.
2. After successful login/signup from a modal, the server typically responds with an `HX-Refresh: true` header, causing a full page refresh. This updates the header to the logged-in state and closes the modal.
3. Authenticated users see the persistent header with their profile dropdown. Navigating to areas like `/dashboard` loads the content into the `#htmx-content` area.
4. Logout (via an HTMX POST request from the dropdown) clears the session and typically triggers an `HX-Refresh: true` or `HX-Redirect` to the landing page.

## ✨ HTMX Integration

This template heavily utilizes HTMX for a modern, single-page application feel without complex JavaScript frameworks.

- **Core Principle**: The main layout (`templates/home/index.plush.html`) acts as a persistent shell. Navigation links and form submissions use HTMX attributes (`hx-get`, `hx-post`, `hx-target`, `hx-swap`) to fetch HTML fragments from the server and swap them into the `#htmx-content` div.
- **Initial Page Load**: The homepage (`/`) loads `index.plush.html`, and the `<main id="htmx-content">` tag has an `hx-trigger="load"` attribute that immediately makes an HTMX request to fetch the initial homepage content (`_index_content.plush.html`).
- **Server-Side Handling**:
    - Go handlers in the `actions` package detect HTMX requests (using `IsHTMX(c.Request())` from `actions/render.go`).
    - For HTMX requests, handlers use a specific render engine (`rHTMX`) that employs a minimal layout (`templates/htmx.plush.html`, which is just `<%= yield %>`). This ensures only the necessary HTML fragment is sent to the client.
    - Standard (non-HTMX) requests render full pages using the default engine (`r`) and `templates/application.plush.html` (which now primarily serves `index.plush.html` for the main app view).
- **Benefits**: Reduced page flicker, faster perceived load times for content changes, and simpler server-rendered HTML.

##  SEO & Performance Features

### Search Engine Optimization
- **Search Engine Friendly**: robots.txt configured to allow crawling while protecting private areas
- **Dynamic Meta Tags**: Page-specific titles, descriptions, and keywords
- **Open Graph**: Social media preview tags for Facebook, Twitter, and LinkedIn
- **Structured Data**: JSON-LD schema markup for SaaS applications
- **Canonical URLs**: Prevent duplicate content issues
- **XML Sitemap**: Basic sitemap for search engines

### Performance & Accessibility
- **Semantic HTML**: Proper HTML5 structure with Pico.css styling
- **HTMX for Dynamic Updates**: Updates page sections without full refreshes
- **Mobile-First**: Responsive design with proper viewport settings
- **Theme Support**: Dark/light/auto modes with system preference detection
- **Fast Loading**: Minimal CSS/JS footprint
- **Accessibility**: Semantic markup and keyboard navigation

## 📊 Architecture & Technology Stack

### Backend Architecture
- **Framework**: Buffalo (Go web framework)
- **Database**: PostgreSQL 15 (containerized)
- **Authentication**: Session-based with bcrypt password hashing
- **Authorization**: Role-based access control with middleware
- **Background Jobs**: Buffalo workers (available for future use)
- **Testing**: Go testing framework with database integration

### Frontend Architecture
- **Templating**: Plush templates - Buffalo's template engine
- **Styling**: Pico.css - Semantic CSS framework with automatic theming
- **Interactions**: HTMX - Dynamic content loading without complex JavaScript
- **Theme System**: Dark/light/auto modes with localStorage persistence
- **Responsive Design**: Mobile-first approach with semantic HTML

### Database Schema

#### Users Table
```sql
users (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) DEFAULT 'user' -- 'user' or 'admin'
);
```

#### Key Features
- **UUID Primary Keys** - Non-enumerable identifiers
- **Timestamps** - Automatic created_at/updated_at tracking
- **Email Uniqueness** - Prevents duplicate accounts
- **Password Security** - bcrypt hashing
- **Role System** - User and admin roles

### Application Structure

```
my-go-saas-template/
├── actions/          # HTTP handlers and middleware
│   ├── app.go       # Main application and routing
│   ├── auth.go      # Authentication handlers
│   ├── admin.go     # Admin management system
│   ├── users.go     # User profile management
│   └── render.go    # Template rendering utilities
├── models/          # Database models and validation
│   └── user.go      # User model with role support
├── templates/       # Plush template files
│   ├── application.plush.html  # Main layout
│   ├── home/        # Homepage and dashboard
│   ├── auth/        # Authentication forms
│   ├── users/       # User profile management
│   └── admin/       # Admin panel templates
├── migrations/      # Database migration files
├── grifts/         # Background tasks and utilities
├── public/         # Static assets (CSS, JS, images)
└── docs/           # Documentation and guides
```

### HTMX Integration Architecture

#### Core Concept
The application uses a persistent shell architecture where the main layout stays loaded and content areas are dynamically updated via HTMX.

#### Key Components
1. **Persistent Shell** (`templates/home/index.plush.html`)
   - Header with navigation and user menu
   - Footer with site information
   - `<main id="htmx-content">` target area

2. **Content Partials** (Various template files)
   - Dashboard content (`templates/home/dashboard.plush.html`)
   - Admin interfaces (`templates/admin/*.plush.html`)
   - User forms (`templates/users/*.plush.html`)

3. **HTMX Response Engine** (`actions/render.go`)
   - Detects HTMX requests via headers
   - Uses minimal layout for partial responses
   - Maintains full page rendering for direct access

#### Benefits
- **Faster Navigation** - Only content area updates, header/footer persist
- **Better UX** - No page flicker during navigation
- **SEO Friendly** - Full pages still render for search engines
- **Progressive Enhancement** - Works without JavaScript as fallback

### Security Architecture

#### Authentication Security
- **Session Management** - Session cookies with expiration
- **Password Hashing** - bcrypt with appropriate cost factor
- **CSRF Protection** - Built-in Buffalo CSRF middleware
- **Input Validation** - Basic validation on user inputs

#### Authorization Security
- **Role-Based Access** - Middleware-enforced role checking
- **Route Protection** - Admin routes protected with `AdminRequired` middleware
- **Template Security** - Role-based conditional rendering
- **API Security** - Authorization checks on endpoints

#### Database Security
- **Prepared Statements** - All queries use parameterization
- **Connection Pooling** - Database connection management
- **Migration Tracking** - Database schema version control
- **Data Validation** - Model-level validation before database operations

### Performance Optimizations

#### Frontend Performance
- **Minimal JavaScript** - HTMX provides dynamic behavior with minimal JS
- **Semantic CSS** - Pico.css provides styling without utility class bloat
- **Template Rendering** - Plush templates for server-side rendering
- **Static Asset Optimization** - Minified CSS and optimized images

#### Backend Performance
- **Compiled Go Binary** - High-performance compiled application
- **Connection Pooling** - Database connection management
- **Session Optimization** - Session storage and retrieval
- **Template Caching** - Plush templates cached in production

#### Database Performance
- **Indexed Queries** - Indexing on frequently queried columns
- **Query Optimization** - Efficient queries
- **Connection Limits** - Connection pool sizing
- **Migration Efficiency** - Non-blocking migrations where possible

## 🛠️ Development

### Template Development

This project uses Buffalo's Plush templating engine, Pico.css for styling, and HTMX for dynamic interactions.

**Important: For ALL styling changes, always consult `/docs/` folder FIRST**

- **Main Shell**: `templates/home/index.plush.html` is the primary persistent layout containing the header, footer, and the `<main id="htmx-content">` target.
- **Content Partials**: Most page-specific content is in separate partial files (e.g., `templates/home/_index_content.plush.html`, `templates/home/dashboard.plush.html`). These are loaded into `#htmx-content`.
- **HTMX Fragments Layout**: `templates/htmx.plush.html` (containing just `<%= yield %>`) is used by the `rHTMX` render engine for HTMX responses.
- **Plush Syntax**: See `/docs/buffalo-template-syntax.md`.
- **Pico.css Styling**: **CRITICAL** - See `/docs/pico-implementation-guide.md` and `/docs/pico-css-variables.md` - Use Pico CSS variables instead of custom CSS
- **Modals**: Pico.css `<dialog>` elements are used for login/signup, triggered by JavaScript and populated by HTMX.
- **Theme Switching**: Built-in dark/light/auto mode support, works with the persistent header.

#### Pico.css Styling Guidelines
- **Always use CSS variables**: Modify `--pico-primary`, `--pico-background-color`, etc. instead of writing custom CSS
- **Check documentation first**: Consult `/docs/pico-css-variables.md` for all available Pico variables
- **Use semantic HTML**: Follow patterns in `/docs/pico-implementation-guide.md` for proper Pico.css usage
- **Never override Pico directly**: Work within Pico's variable system for all customization

### Database Management
The application includes a PostgreSQL container configured via `docker-compose.yml`:
- **Development DB**: `github.com/jbhicks/sound-cistern_development`
- **Test DB**: `github.com/jbhicks/sound-cistern_test`
- **Production DB**: `github.com/jbhicks/sound-cistern_production`

### Common Commands
```console
# Development shortcuts (recommended)
make dev                       # Start everything for development
make setup                     # First-time setup
make test                      # Run tests
make clean                     # Stop and cleanup
make admin                     # Promote first user to admin

# Manual commands
# Database operations
soda create -a                 # Create databases
soda migrate up                # Run migrations
soda generate migration        # Create new migration

# Admin management
buffalo task db:promote_admin  # Promote first user to admin role

# Development
buffalo dev                    # Start dev server with hot reload
buffalo build                  # Build production binary
buffalo test                   # Run tests

# Container management (Podman/Docker)
podman-compose up -d           # Start database
podman-compose down            # Stop database
podman-compose ps              # Check container status
```
User management endpoints (`/users`, `/auth`) are still the same, but interactions are now primarily via HTMX from modals or links.

## 🤖 Bot Instructions

When working with this Buffalo SaaS template:

### Template Development & HTMX
1.  **Understand the Shell**: `templates/home/index.plush.html` is the persistent shell. Most new content should be a partial loaded into `#htmx-content`.
2.  **HTMX Attributes**: Use `hx-get`, `hx-post`, `hx-target="#htmx-content"`, `hx-swap` for navigation and forms. For modals, target the modal's content div.
3.  **Server-Side**:
    *   Check for HTMX requests using `IsHTMX(c.Request())`.
    *   Use `rHTMX.HTML("path/to/partial.plush.html")` for HTMX responses.
    *   Use `r.HTML("home/index.plush.html")` for initial full page loads of the main app view.
4.  **Plush Syntax**: Refer to `/docs/buffalo-template-syntax.md`. Avoid Go-style operations.
5.  **Modals**: Login/signup forms are in modals. Ensure HTMX attributes on trigger buttons target modal content divs. Server responses for modal forms (e.g., validation errors) should re-render the form fragment. Successful modal submissions often use `HX-Refresh: true`.

### Styling with Pico.css

**CRITICAL: Always use Pico.css variables instead of custom CSS**

1.  **Semantic HTML**: Key for Pico.css - use proper HTML elements
2.  **CSS Variables Only**: Modify `--pico-primary`, `--pico-background-color`, etc. instead of writing custom CSS
3.  **Documentation First**: Always check `/docs/pico-css-variables.md` and `/docs/pico-implementation-guide.md` BEFORE making styling changes
4.  **Modals**: Use `<dialog>` and `<article>` structure as documented in `/docs/`
5.  **Theme Support**: Use CSS variables to ensure compatibility with dark/light modes

### Authentication & Authorization
1.  **Modal Forms**: Login/signup are via modals loaded with HTMX.
2.  **Session Management**: `current_user_id` in session, `current_user` in templates.
3.  **Role-based Access**: Check `current_user.Role` for admin functionality.
4.  **Admin Middleware**: Use `AdminRequired` middleware for admin-only routes.
5.  **Post-Login/Signup**: Usually `HX-Refresh: true` from server.

### Admin System Patterns
1.  **Admin Routes**: Group under `/admin` with `AdminRequired` middleware.
2.  **Role Checks**: Use `current_user.Role == "admin"` in templates for conditional content.
3.  **User Management**: CRUD operations follow Buffalo conventions with proper validation.
4.  **Safety Checks**: Always prevent users from deleting themselves or escalating beyond their permissions.

### Common Patterns
- **Persistent Elements**: Header/footer in `index.plush.html` use `hx-preserve="true"`.
- **Conditional Content**: Check `current_user` for auth-specific content, often within partials.
- **Form Handling**: Standard Buffalo form helpers can be used, but HTMX attributes handle submission.

### Troubleshooting
- **500 errors**: Often Plush syntax. Check Buffalo logs.
- **HTMX Issues**: Use browser dev tools (Network tab) to inspect HTMX requests and responses. Check `HX-Request` headers and what HTML fragments are being returned. Ensure `hx-target` and `hx-swap` are correct.

## 🤖 Development Assistant Instructions

When working with this Buffalo SaaS template, follow these patterns and guidelines:

#### Template Development & HTMX Integration
1. **Understand the Shell Architecture**: `templates/home/index.plush.html` is the persistent shell. Most new content should be a partial template loaded into `#htmx-content`.

2. **HTMX Response Patterns**: 
   - Use `hx-get`, `hx-post`, `hx-target="#htmx-content"`, `hx-swap` for navigation and forms
   - For modals, target the modal's content div specifically
   - Server responses use `rHTMX.HTML("path/to/partial.plush.html")` for HTMX requests

3. **Template Engine Guidelines**:
   - Check for HTMX requests using `IsHTMX(c.Request())` in handlers
   - Use `r.HTML("home/index.plush.html")` for initial full page loads
   - Reference `/docs/buffalo-template-syntax.md` for Plush syntax patterns

#### Styling with Pico.css Framework

**CRITICAL: Always consult `/docs/` folder before making ANY styling changes**

1. **Documentation First**: Check `/docs/pico-css-variables.md` and `/docs/pico-implementation-guide.md` BEFORE styling
2. **CSS Variables Only**: Use `--pico-primary`, `--pico-background-color`, etc. - NEVER write custom CSS rules
3. **Semantic HTML First**: Use proper HTML elements (`<nav>`, `<article>`, `<section>`, `<details>`) as shown in `/docs/`
4. **Minimal CSS Classes**: Prefer `role="button"`, `class="secondary"`, `class="dropdown"` over custom styles
5. **Theme Compatibility**: Use CSS variables to ensure dark/light mode compatibility
6. **Responsive Design**: Trust Pico.css responsive behavior, avoid custom breakpoints unless necessary

#### Authentication & Authorization Patterns
1. **Modal Authentication**: Login/signup forms load via HTMX into modal dialogs
2. **Session Management**: Use `current_user_id` in session, `current_user` available in templates
3. **Role-Based Access**: Check `current_user.Role` for admin functionality in templates
4. **Admin Middleware**: Always use `AdminRequired` middleware for admin-only routes
5. **Post-Authentication**: Use `HX-Refresh: true` header for successful modal form submissions

#### Admin System Development
1. **Route Structure**: Group admin routes under `/admin` with `AdminRequired` middleware protection
2. **Role Checking**: Use `current_user.Role == "admin"` in templates for conditional content
3. **CRUD Operations**: Follow Buffalo conventions with proper validation and error handling
4. **Safety Controls**: Always prevent users from deleting themselves or escalating beyond permissions

#### Database & Migration Patterns
1. **Migration Safety**: Use non-blocking migrations where possible for production deployments
2. **Model Validation**: Implement comprehensive validation at the model level
3. **Role System**: Use `role` field with proper constraints and default values
4. **UUID Primary Keys**: Maintain UUID usage for security and scalability

#### Common Development Patterns
- **Persistent Elements**: Header/footer use `hx-preserve="true"` to maintain state
- **Conditional Content**: Check `current_user` for authentication-specific content rendering
- **Form Handling**: Use HTMX attributes for submission while maintaining Buffalo form helpers
- **Error Handling**: Provide comprehensive error messages and proper HTTP status codes

#### Testing & Quality Assurance
- **Test Coverage**: Maintain tests for all authentication and admin functionality
- **Database Testing**: Use test database environment for isolated test runs
- **HTMX Testing**: Test both HTMX and direct URL access for all routes
- **Role Testing**: Verify proper authorization for all role-based features

#### Troubleshooting Guidelines
- **500 Errors**: Usually Plush template syntax issues - check Buffalo console output
- **HTMX Issues**: Use browser dev tools Network tab to inspect HTMX requests/responses
- **Database Issues**: Use `make db-status` and `make db-logs` for diagnostics
- **Permission Issues**: Verify middleware application and role assignments

## 📁 Project File Structure

```
my-go-saas-template/
├── 🗄️  Database & Configuration
│   ├── database.yml              # Database configuration for all environments
│   ├── docker-compose.yml        # PostgreSQL container configuration
│   └── migrations/               # Database migration files
│       ├── *_create_users.up.fizz   # Initial user table creation
│       └── *_add_role_to_users.*.fizz # Role system addition
│
├── 🏗️  Application Core
│   ├── main.go                   # Application entry point
│   ├── app                       # Buffalo application instance
│   ├── actions/                  # HTTP handlers and middleware
│   │   ├── app.go               # Main routing and application setup
│   │   ├── auth.go              # Authentication handlers (login/logout)
│   │   ├── users.go             # User profile management
│   │   ├── admin.go             # Admin management system (CRUD)
│   │   ├── home.go              # Homepage and dashboard handlers
│   │   └── render.go            # Template rendering utilities
│   │
│   ├── models/                   # Database models and validation
│   │   ├── models.go            # Database connection and base models
│   │   └── user.go              # User model with role support
│   │
│   └── grifts/                   # Background tasks and utilities
│       └── db.go                # Admin promotion and database tasks
│
├── 🎨 Frontend & Templates
│   ├── templates/                # Plush template files
│   │   ├── application.plush.html    # Base application layout
│   │   ├── home/                     # Homepage and dashboard templates
│   │   │   ├── index.plush.html      # Persistent shell (header/footer)
│   │   │   └── dashboard.plush.html  # User dashboard with admin section
│   │   ├── auth/                     # Authentication form templates
│   │   │   ├── new.plush.html        # Login form (modal)
│   │   │   └── landing.plush.html    # Marketing landing page
│   │   ├── users/                    # User management templates
│   │   │   ├── profile.plush.html    # User profile editing
│   │   │   └── new.plush.html        # User registration (modal)
│   │   └── admin/                    # Admin panel templates
│   │       ├── users.plush.html      # User management table
│   │       └── user_edit.plush.html  # User editing form
│   │
│   └── public/                   # Static assets
│       ├── css/
│       │   ├── pico.min.css     # Pico.css framework (semantic styling)
│       │   └── custom.css       # Custom CSS variables and overrides
│       ├── js/
│       │   ├── theme.js         # Dark/light mode switching
│       │   └── icons.js         # Icon system utilities
│       └── images/              # Static images and favicon
│
├── 🛠️  Development & Deployment
│   ├── Makefile                  # Robust development commands
│   ├── scripts/
│   │   └── wait-for-postgres.sh # Database health check script
│   ├── go.mod                   # Go module dependencies
│   ├── go.sum                   # Dependency checksums
│   └── bin/                     # Compiled binaries (created by build)
│
├── 🧪 Testing & Quality
│   ├── *_test.go                # Go test files throughout project
│   ├── fixtures/                # Test data fixtures
│   └── tmp/                     # Temporary build files
│
└── 📖 Documentation
    ├── README.md                # This comprehensive guide
    └── docs/                    # Additional documentation
        ├── buffalo-template-syntax.md     # Plush templating guide
        ├── pico-implementation-guide.md   # Semantic CSS patterns
        ├── pico-css-variables.md          # CSS customization guide
        └── seo-implementation.md          # SEO optimization guide
```

### Key File Descriptions

#### Core Application Files
- **`actions/app.go`** - Main application setup, routing configuration, and middleware stack
- **`actions/admin.go`** - Complete admin management system with CRUD operations and safety controls  
- **`models/user.go`** - User model with role support, validation, and authentication methods
- **`templates/home/index.plush.html`** - Persistent application shell with HTMX content area

#### Database Files
- **`migrations/*.fizz`** - Database schema evolution with role-based user system
- **`database.yml`** - Multi-environment database configuration
- **`grifts/db.go`** - Administrative tasks including user promotion to admin role

#### Frontend Architecture
- **`public/css/pico.min.css`** - Semantic CSS framework providing automatic styling
- **`public/js/theme.js`** - Theme switching functionality with localStorage persistence
- **`templates/admin/*.plush.html`** - Professional admin interface templates

#### Development Infrastructure  
- **`Makefile`** - Comprehensive development commands with health checks and error handling
- **`scripts/wait-for-postgres.sh`** - Database readiness verification for reliable automation
- **`docker-compose.yml`** - PostgreSQL container configuration for development environment

This file structure supports a maintainable SaaS application with clear separation of concerns.

## 📝 Development Roadmap

### 🚀 Unified Logging Implementation Plan

**Status**: Planning Phase - Not Started

Buffalo already has a solid logging foundation via `gobuffalo/logger` (logrus-based). This plan enhances it with configurability and structured business event logging.

#### **📋 Current State Analysis**

**✅ What Buffalo Already Provides:**
- [x] Built-in structured logging with request IDs, timing, status codes
- [x] paramlogger middleware for HTTP request logging  
- [x] Context-aware logger via `c.Logger()`
- [x] Log levels (info, debug, error, etc.)
- [x] JSON-like structured output

**❌ What's Missing:**
- [ ] Configurable file output location
- [ ] Consistent application-level logging
- [ ] Business event logging (user actions, errors)
- [ ] Centralized logging configuration

#### **🎯 Implementation Phases**

##### **Phase 1: Configuration & File Output** 
- [ ] Create logging configuration structure
  - [ ] Environment-based log levels (`LOG_LEVEL`)
  - [ ] Configurable file paths with sensible defaults (`LOG_FILE_PATH`)
  - [ ] Development vs production settings
- [ ] Add file output support
  - [ ] Default: `/logs/application.log`
  - [ ] Log rotation support
- [ ] Enhance Buffalo's existing logger
  - [ ] Keep Buffalo's middleware logging (already good)
  - [ ] Add custom fields for business context
  - [ ] Configure log level via environment

##### **Phase 2: Structured Application Logging**
- [ ] Create centralized logging service
  - [ ] Wrapper around Buffalo's logger
  - [ ] Consistent field names and formats
  - [ ] Request correlation ID support
- [ ] Add business event logging
  - [ ] User registration/login/logout events
  - [ ] Admin actions (user management, role changes)
  - [ ] Error tracking with context
  - [ ] Security events (failed login attempts, etc.)

##### **Phase 3: Integration & Standards**
- [ ] Update existing codebase
  - [ ] Replace scattered `c.Logger().Debugf()` calls
  - [ ] Add structured logging to key business flows
  - [ ] Standardize error logging
- [ ] Documentation and guidelines
  - [ ] Logging standards for the team
  - [ ] Examples and best practices

#### **🔧 Technical Implementation Details**

**Directory Structure:**
```
logs/
├── application.log          # Main application logs
├── access.log              # HTTP request logs (optional)
├── error.log               # Error-only logs
└── audit.log               # Security/admin events
```

**Configuration Approach:**
- Use Buffalo's existing logger infrastructure (don't reinvent)
- Environment variables for configuration
- Sensible defaults that work out of the box
- Compatible with Docker/container deployments

**Log Levels & Events:**
- **INFO**: User actions, business events
- **WARN**: Unusual but handled conditions
- **ERROR**: Application errors, failed operations
- **DEBUG**: Development debugging (current usage)

**Structured Fields Standard:**
- `user_id`: Current user context
- `request_id`: Buffalo's existing request IDs
- `action`: Business action being performed
- `resource`: What resource is being acted upon
- `ip_address`: Client IP for security events

---

## 📝 Content Management System (CMS)

This template includes a comprehensive blog and content management system with advanced features for content creation, management, and SEO optimization.

### CMS Features Overview

#### Content Creation & Editing
- **Rich Text Editor** - Professional WYSIWYG editor powered by Quill.js
- **Draft System** - Save content as drafts before publishing
- **SEO Optimization** - Complete meta tags and Open Graph support
- **Automatic Slug Generation** - URL-friendly slugs generated from titles
- **Content Excerpts** - Auto-generated or custom excerpts for listings

#### Content Management
- **Search & Filter** - Find posts by title, content, author, or publication status
- **Bulk Operations** - Manage multiple posts simultaneously
- **Status Management** - Published/Draft status with visual indicators
- **Author Attribution** - Posts linked to user accounts with proper attribution

### Using the CMS

#### Creating Blog Posts

1. **Access Admin Panel** - Log in as an admin user and navigate to `/admin`
2. **Create New Post** - Click "Blog Posts" → "New Post"
3. **Content Creation**:
   - **Title**: Enter a descriptive title (slug auto-generates)
   - **Content**: Use the rich text editor for formatted content
   - **Excerpt**: Add custom excerpt or leave blank for auto-generation
   - **Publication Status**: Check "Published" to make live, uncheck for draft

#### Rich Text Editor Features

The Quill.js editor provides:
- **Text Formatting**: Bold, italic, underline, strikethrough
- **Headers**: H1, H2, H3 for content structure
- **Lists**: Numbered and bulleted lists with indentation
- **Links**: Insert and edit hyperlinks
- **Quotes**: Blockquotes for emphasized content
- **Code**: Inline code and code blocks
- **Cleanup**: Remove formatting tool

#### SEO & Social Media Optimization

Each post includes comprehensive SEO fields accessible via the "SEO & Social Media Settings" section:

##### Meta Tags (SEO)
- **Meta Title**: Custom title for search engines (50-60 chars recommended)
- **Meta Description**: Search result snippet (150-160 chars recommended)  
- **Meta Keywords**: Comma-separated keywords for search engines

##### Open Graph (Social Media)
- **OG Title**: Title for social media shares
- **OG Description**: Description for social media previews
- **OG Image**: Image URL for social media previews (1200x630px recommended)

**Best Practices:**
- Leave fields blank to use post title/excerpt as defaults
- Optimize meta descriptions for click-through rates
- Use high-quality, relevant Open Graph images
- Test social media previews before publishing

#### Content Search & Filtering

The admin posts interface provides powerful search capabilities:

##### Search Options
- **Text Search**: Search across post titles, content, and author names
- **Status Filter**: Filter by Published, Draft, or All Posts
- **Combined Filters**: Use search text and status filter together

##### Usage Tips
- Use specific keywords to quickly find posts
- Filter by status to review drafts or published content
- Clear filters to return to full post listing

#### Bulk Operations

Efficiently manage multiple posts with bulk actions:

##### Available Actions
- **Bulk Publish**: Make multiple drafts live simultaneously
- **Bulk Unpublish**: Convert published posts to drafts
- **Bulk Delete**: Remove multiple posts (with confirmation)

##### How to Use Bulk Operations
1. **Select Posts**: Check boxes next to posts you want to modify
2. **Select All**: Use the header checkbox to select all visible posts
3. **Choose Action**: Select desired action from dropdown
4. **Apply**: Click "Apply" and confirm the action
5. **Confirmation**: Review the confirmation dialog before proceeding

**Safety Features:**
- Confirmation dialogs for all bulk actions
- Special confirmation for destructive delete operations
- Flash messages confirm successful operations

### CMS Administration

#### Post Management Workflow

##### Content Creation Process
1. **Draft Creation**: Create posts as drafts for review
2. **Content Development**: Use rich text editor for professional formatting
3. **SEO Optimization**: Complete meta tags and social media fields
4. **Review Process**: Preview content before publishing
5. **Publication**: Publish when ready or schedule for later

##### Content Maintenance
- **Regular Reviews**: Use search/filter to find content needing updates
- **Bulk Updates**: Use bulk operations for status changes
- **SEO Monitoring**: Review and update meta descriptions periodically
- **Content Audits**: Use draft status for content under revision

#### Admin Routes for CMS

| Route | Method | Description | Access Level |
|-------|--------|-------------|--------------|
| `/admin/posts` | GET | Post listing with search/filter | Admin Only |
| `/admin/posts/new` | GET | New post creation form | Admin Only |
| `/admin/posts` | POST | Create new post | Admin Only |
| `/admin/posts/bulk` | POST | Bulk operations on posts | Admin Only |
| `/admin/posts/{id}` | GET | View single post details | Admin Only |
| `/admin/posts/{id}/edit` | GET | Edit post form | Admin Only |
| `/admin/posts/{id}` | POST | Update existing post | Admin Only |
| `/admin/posts/{id}` | DELETE | Delete single post | Admin Only |

#### Database Schema for Posts

```sql
posts (
    id INTEGER PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    content TEXT NOT NULL,
    excerpt TEXT,
    published BOOLEAN DEFAULT false,
    author_id UUID NOT NULL REFERENCES users(id),
    
    -- SEO Fields
    meta_title VARCHAR(255),
    meta_description TEXT,
    meta_keywords TEXT,
    og_title VARCHAR(255),
    og_description TEXT,
    og_image VARCHAR(255),
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### CMS Best Practices

#### Content Strategy
- **Consistent Publishing**: Maintain regular content publication schedule
- **SEO Optimization**: Always complete meta descriptions and titles
- **Draft Workflow**: Use drafts for collaborative content creation
- **Content Organization**: Use descriptive titles and proper excerpts

#### Technical Considerations
- **Image Optimization**: Optimize images before adding to Open Graph fields
- **Link Management**: Regularly check and update external links
- **Performance**: Monitor content length for page load performance
- **Backup**: Regular database backups to protect content

#### Security & Access Control
- **Admin Access**: Only trusted users should have admin privileges
- **Content Review**: Implement content review process for published materials
- **Draft Protection**: Use draft status for sensitive content under development
- **Audit Trail**: Monitor who creates and modifies content

### Troubleshooting CMS Issues

#### Common Problems

**Rich Text Editor Not Loading**
- Check that Quill.js assets are properly served from `/public/js/` and `/public/css/`
- Verify JavaScript console for loading errors
- Ensure proper MIME types for static assets

**Search Not Working**
- Verify database connection and query parameters
- Check PostgreSQL ILIKE support for case-insensitive search
- Review search term encoding and special characters

**Bulk Operations Failing**
- Confirm POST request includes CSRF token
- Check that post IDs are properly submitted in form data
- Verify admin permissions for bulk operation routes

**SEO Fields Not Saving**
- Ensure database migration for SEO fields completed successfully
- Check form field names match model properties
- Verify model validation allows optional SEO fields
