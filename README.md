# Sound Cistern

A PocketBase web application for tracking and managing Soundcloud content with user authentication and blog capabilities.

## ✅ Features Checklist

### Core Infrastructure
- [x] **PocketBase application** - Go web framework with embedded SQLite
- [x] **SQLite database** - Embedded database with PocketBase
- [x] **Database migrations** - User, post, and Soundcloud data tables
- [x] **Templ templates** - Type-safe Go templating
- [x] **Development workflow** - Make commands for common tasks

### Authentication & Authorization
- [x] **Authentication system** - User registration, login with PocketBase
- [x] **User management** - Built-in PocketBase admin UI
- [x] **Password security** - PocketBase bcrypt hashing

### User Interface & Experience
- [x] **Templ templates** - Type-safe Go templates with Pico.css styling
- [x] **HTMX integration** - Dynamic content loading
- [x] **Theme system** - Dark/light/auto modes
- [x] **Responsive design** - Mobile-friendly layout

### Blog System
- [x] **Blog post system** - Posts with author attribution
- [x] **SEO-friendly URLs** - Slug-based routing
- [x] **Post excerpts** - Content summaries

### Soundcloud Integration
- [ ] **Soundcloud OAuth** - User authentication via Soundcloud
- [ ] **Track tracking** - Monitor Soundcloud tracks
- [ ] **Feed generation** - RSS/podcast feeds from Soundcloud content

## Quick Start

### Prerequisites
- **Go 1.23+** - [Download Go](https://golang.org/dl/)
- **Templ CLI** - `go install github.com/a-h/templ/cmd/templ@latest`

### Setup

```console
# Clone the repository
git clone <your-repo-url>
cd sound-cistern

# Install dependencies
go mod download

# Generate templates
make templ

# Build application
make build

# Run in development mode
make dev
```

After setup, visit [http://127.0.0.1:8090](http://127.0.0.1:8090) to see your application running.

## Development

### Auto-Reload Development

Run PocketBase in dev mode for automatic reloading:

```console
make dev
```

This starts PocketBase with the `--dev` flag, which provides:
- Admin UI at `/_/` (no authentication required in dev mode)
- Auto-reload on file changes
- Debug logging

### Template Development

This project uses Templ for type-safe Go templates:

```console
# Generate templates after changes
make templ

# Watch mode (requires templ CLI)
templ generate --watch
```

### First Admin User

PocketBase creates an admin UI automatically in dev mode at `http://127.0.0.1:8090/_/`

For production:
```console
# Create admin user interactively
./sound-cistern admin create
```

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
# Run all tests
make test

# Clean up after development
make clean
```

### Advanced Commands

#### Database Operations
```console
# PocketBase handles migrations automatically
# Migration files are in pb_migrations/
# View admin UI at http://localhost:8090/_/

# Manual migration creation (Go files)
# Create new migration file in pb_migrations/
```

#### Admin Management
PocketBase provides a built-in admin dashboard at `/_/` where you can:
- Manage users and collections
- View logs and analytics
- Configure authentication settings
- Run database queries

No manual admin promotion needed - create admin account on first run.

#### Building & Production
```console
# Build for production
make build                     # Creates binary: ./sound-cistern

# Run production server
make serve                     # Starts server on port 8090

# Single binary includes:
# - Application code
# - Embedded SQLite database (pb_data/)
# - Static assets (public/)
```

### Development Tips

#### PocketBase Development Server
- **Automatic reload** - Use `--dev` flag for debug mode and auto-restart
- **Port 8090** - Default server port
- **Template changes** - Run `make templ` to regenerate Templ files
- **Database file** - SQLite database stored in `pb_data/data.db`

#### Database Development
- **Embedded SQLite** - No separate database container needed
- **Migrations** - Go files in `pb_migrations/` run automatically on startup
- **Admin UI** - Access at `http://localhost:8090/_/` for data management
- **Backup** - Simple file copy of `pb_data/` directory

#### Template Development
- **Templ templates** - Type-safe Go templates (`.templ` files)
- **Generation required** - Run `make templ` after editing `.templ` files
- **HTMX integration** - Templates support dynamic content loading
- **Pico.css styling** - Semantic HTML with automatic styling

### Troubleshooting

#### Common Issues

**Port Conflicts**
```console
# Check what's using port 8090
lsof -i :8090

# Kill process if needed
kill -9 $(lsof -t -i:8090)
```

**Template Compilation Errors**
```console
# Regenerate Templ templates
make templ

# Check for syntax errors in .templ files
# Error messages will show specific file and line number
```

**Database Issues**
```console
# Check database file exists
ls -lh pb_data/data.db

# View logs
ls -lh pb_data/logs.db

# Reset database (caution: deletes all data)
rm -rf pb_data/
# Restart server to recreate with migrations
```

**Build Errors**
```console
# Clean Go module cache
go clean -modcache
go mod tidy

# Rebuild everything
make templ && make build
```

## 🔐 Authentication Features

### Built-in PocketBase Authentication
- **Admin Dashboard**: `/_/` - Built-in admin UI for user management
- **Session-based Auth**: PocketBase handles authentication and session management
- **User Records**: Users stored in PocketBase collections with built-in validation
- **Protected Routes**: Middleware checks authentication status for protected pages

### User Interface
- **Persistent Header/Footer**: Main site layout with navigation and theme toggle
- **Dynamic Content**: HTMX loads page content without full page reloads
- **Modal Forms**: Pico.css modals for login/signup triggered via HTMX
- **Theme Switching**: Dark/light/auto mode with localStorage persistence

## ✨ HTMX Integration

This template uses HTMX for dynamic page updates without full page reloads.

- **Dynamic Content**: Navigation and forms use HTMX (`hx-get`, `hx-post`, `hx-target`) to update page sections
- **Server-Side Rendering**: Templ templates generate HTML fragments for HTMX responses
- **Progressive Enhancement**: Full page loads work without JavaScript, HTMX enhances the experience
- **Reduced Complexity**: No complex JavaScript frameworks needed

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
- **Framework**: PocketBase v0.22.0 (Go-based BaaS)
- **Database**: SQLite (embedded, file-based)
- **Authentication**: PocketBase built-in auth with sessions
- **API**: RESTful API with PocketBase hooks and custom routes
- **Realtime**: WebSocket support via PocketBase subscriptions

### Frontend Architecture
- **Templating**: Templ v0.3.960 - Type-safe Go templates
- **Styling**: Pico.css v1.5.13 - Semantic CSS framework with theming
- **Interactions**: HTMX - Dynamic content loading
- **Theme System**: Dark/light/auto modes with localStorage
- **Responsive Design**: Mobile-first semantic HTML

### Database Schema

PocketBase collections are defined in `pb_migrations/` as Go files. Key collections:

#### Users Collection
- Built-in PocketBase auth collection
- Email/password authentication
- Role-based access control
- Avatar support

#### Soundcloud Collections
- `soundcloud_users` - Soundcloud user profiles
- `soundcloud_tracks` - Track metadata and links
- `soundcloud_feeds` - Feed URLs and sync status

### Application Structure

```
sound-cistern/
├── views/              # Templ template files (*.templ)
│   ├── layout.templ   # Main layout wrapper
│   ├── home.templ     # Homepage
│   ├── signin.templ   # Login page
│   └── signup.templ   # Registration page
├── pb_migrations/      # Database migrations (Go files)
├── pb_data/           # SQLite database and runtime data
│   ├── data.db       # Main database file
│   └── logs.db       # Application logs
├── public/            # Static assets
│   ├── css/          # Pico.css and custom styles
│   ├── js/           # HTMX and theme switcher
│   └── uploads/      # User-uploaded files
├── main.go            # Application entry point
└── docs/              # Documentation
```

### Security Architecture

#### Authentication Security
- **PocketBase Auth** - Built-in secure authentication system
- **Password Hashing** - Bcrypt hashing with appropriate cost
- **Session Management** - Secure cookie-based sessions
- **CSRF Protection** - Built-in request validation

#### Authorization Security
- **Collection Rules** - PocketBase rule-based access control
- **API Security** - Record-level permissions
- **Admin Dashboard** - Separate admin authentication at `/_/`

#### Database Security
- **Prepared Statements** - All queries use parameterization
- **Validation** - Schema validation on all records
- **Backup Strategy** - Simple file-based database backup

### Performance Optimizations

#### Frontend Performance
- **Minimal JavaScript** - HTMX provides interactivity with minimal JS
- **Semantic CSS** - Pico.css without utility class bloat
- **Type-safe Templates** - Templ compilation catches errors early
- **Static Asset Optimization** - Minified CSS/JS

#### Backend Performance
- **Compiled Binary** - Single optimized Go executable
- **SQLite Performance** - Fast embedded database
- **Template Compilation** - Templ templates compiled to Go
- **Connection Pooling** - Efficient database access

## 🤖 Development Assistant Instructions

When working with this PocketBase + Templ application, follow these patterns:

#### Template Development & HTMX Integration
1. **Templ Files**: Edit `.templ` files in `views/` directory, then run `make templ`
2. **Type Safety**: Templ provides compile-time checking - use proper Go types
3. **HTMX Patterns**: Use `hx-get`, `hx-post`, `hx-target`, `hx-swap` for dynamic updates
4. **Component Reuse**: Create reusable Templ components for repeated UI elements

#### Styling with Pico.css Framework

**CRITICAL: Always consult `/docs/` folder before making ANY styling changes**

1. **Documentation First**: Check `/docs/pico-css-variables.md` and `/docs/pico-implementation-guide.md`
2. **CSS Variables Only**: Use `--pico-primary`, `--pico-background-color`, etc.
3. **Semantic HTML**: Use proper HTML elements (`<nav>`, `<article>`, `<section>`, `<details>`)
4. **Minimal Classes**: Prefer `role="button"`, `class="secondary"` over custom styles
5. **Theme Compatibility**: Use CSS variables for dark/light mode support
6. **Responsive Design**: Trust Pico.css responsive behavior

#### PocketBase Patterns
1. **Collections**: Define in `pb_migrations/` as Go files
2. **Hooks**: Use `OnRecordBefore*` and `OnRecordAfter*` for validation/logic
3. **Auth**: Use `c.Get("authRecord")` to access authenticated user
4. **Admin UI**: Access at `/_/` for data management
5. **Realtime**: Use PocketBase subscriptions for live updates

#### Database & Migration Patterns
1. **Migration Files**: Create Go files in `pb_migrations/`
2. **Auto-run**: Migrations run automatically on application startup
3. **Collection Rules**: Define access rules in collection schema
4. **Validation**: Use PocketBase validators in collection fields

#### Common Development Patterns
- **HTMX Forms**: Use HTMX for form submission without page reloads
- **Error Handling**: Return proper HTTP status codes and error messages
- **File Uploads**: Use PocketBase file fields with validation
- **API Routes**: Use `e.Router` to add custom routes

#### Testing & Quality Assurance
- **Test Coverage**: Write tests for custom routes and business logic
- **HTMX Testing**: Test both HTMX and direct URL access
- **Database Testing**: Use separate SQLite file for tests

#### Troubleshooting Guidelines
- **Template Errors**: Check `make templ` output for compilation errors
- **HTMX Issues**: Inspect Network tab for request/response headers
- **Database Issues**: Check `pb_data/` directory permissions and file integrity
- **Auth Issues**: Verify collection rules and authentication settings

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
│   ├── views/                    # Templ template files
│   │   ├── layout.templ         # Main layout wrapper
│   │   ├── home.templ           # Homepage
│   │   ├── blog.templ           # Blog pages
│   │   ├── signin.templ         # Login page
│   │   └── signup.templ         # Registration page
│   └── *_templ.go               # Generated template files (from Templ)
│
├── 🎨 Frontend & Static Assets
│   └── public/                   # Static assets
│       ├── css/
│       │   ├── pico.min.css     # Pico.css framework
│       │   └── custom.css       # Custom CSS variables
│       ├── js/
│       │   ├── htmx.min.js      # HTMX library
│       │   ├── theme.js         # Theme switcher
│       │   └── icons.js         # Icon utilities
│       ├── images/              # Static images
│       └── uploads/             # User-uploaded files
│
├── 🛠️  Development & Deployment
│   ├── Makefile                  # Build and dev commands
│   ├── go.mod                   # Go dependencies
│   ├── go.sum                   # Dependency checksums
│   ├── Dockerfile               # Production container
│   └── sound-cistern            # Compiled binary (after build)
│
└── 📖 Documentation
    ├── README.md                # This guide
    ├── SETUP.md                 # Setup instructions
    ├── CUSTOMIZATION.md         # Customization guide
    └── docs/                    # Additional documentation
        ├── pico-css-variables.md
        ├── pico-implementation-guide.md
        ├── seo-implementation.md
        └── buffalo/             # Legacy Buffalo docs (archived)
```

### Key File Descriptions

#### Core Application Files
- **`main.go`** - PocketBase application setup, hooks, and custom routes
- **`views/*.templ`** - Type-safe template files compiled to Go
- **`pb_migrations/*.go`** - Database schema and migration logic

#### Database Files
- **`pb_data/data.db`** - SQLite database file
- **`pb_data/logs.db`** - Application logs
- **`pb_migrations/`** - Migration files define collections and schema

#### Frontend Architecture
- **`public/css/pico.min.css`** - Semantic CSS framework
- **`public/js/theme.js`** - Theme switching with localStorage
- **`public/js/htmx.min.js`** - Dynamic content loading

#### Development Infrastructure  
- **`Makefile`** - Build, dev, and deployment commands
- **`Dockerfile`** - Production container configuration
- **`.env`** - Environment configuration (not in git)

## 📝 Development Roadmap

### Planned Features

#### Phase 1: Core Soundcloud Integration (In Progress)
- [x] PocketBase migration from Buffalo
- [x] Templ template system
- [ ] Soundcloud API integration
- [ ] Feed URL ingestion
- [ ] Track metadata storage

#### Phase 2: User Features
- [ ] User authentication (PocketBase built-in)
- [ ] Feed subscription management
- [ ] Personalized dashboards
- [ ] Mobile-responsive UI

#### Phase 3: Advanced Features
- [ ] Search and filtering
- [ ] Playlist management
- [ ] Offline support
- [ ] PWA capabilities

## 🤝 Contributing

This is a personal project, but suggestions and bug reports are welcome via GitHub issues.

## 📄 License

MIT License - see LICENSE file for details.

## 🙏 Acknowledgments

- **PocketBase** - Backend framework
- **Templ** - Type-safe Go templates
- **Pico.css** - Semantic CSS framework
- **HTMX** - Dynamic interactions
- **Soundcloud** - Audio platform API
