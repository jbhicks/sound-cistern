# Buffalo Development Workflow

## Starting the Development Server

Buffalo provides a built-in development server with hot reloading capabilities.

### Primary Development Command

```bash
buffalo dev
```

This command:
- Starts the development server on port 3000 by default
- Automatically watches for file changes
- Recompiles and restarts the server when changes are detected
- Handles Go files, templates, and assets

### Development Server Features

- **Hot Reload**: Automatically rebuilds when files change
- **Asset Pipeline**: Processes CSS, JS, and other assets
- **Template Watching**: Reloads when Plush templates change
- **Database Integration**: Connects to configured database

### Configuration

The development server reads configuration from:
- `config/buffalo-app.toml` - Main application configuration
- Environment variables
- Database configuration from `database.yml`

### Port Configuration

Default port is 3000. To change:

```bash
# Via environment variable
PORT=8080 buffalo dev

# Or set in buffalo-app.toml
```

### Development Environment

Buffalo sets `GO_ENV=development` automatically when using `buffalo dev`.

## Stopping the Server

- **Graceful shutdown**: `Ctrl+C` in the terminal running `buffalo dev`
- **Force kill**: `buffalo dev` handles most cleanup automatically

## Troubleshooting Development Server

### Port Already in Use

If you get "address already in use" errors:

```bash
# Check what's using port 3000
lsof -ti:3000

# Kill specific processes if needed
pkill -f "my-go-saas-template"
```

### Server Not Reloading

1. Check file permissions
2. Verify you're in the project root directory
3. Check for syntax errors in Go files
4. Review Buffalo logs for specific errors

### Database Connection Issues

1. Ensure PostgreSQL is running: `docker-compose ps`
2. Check database configuration in `database.yml`
3. Verify migrations are up to date: `soda migrate up`

## Best Practices

1. **Always use `buffalo dev`** for development - don't run the binary directly
2. **Keep one terminal dedicated** to the Buffalo dev server
3. **Check logs** in the Buffalo dev terminal for errors
4. **Use `Ctrl+C`** to stop the server cleanly before making major changes
5. **Restart only when necessary** - the hot reload should handle most changes

## Alternative Commands

### Building for Production
```bash
buffalo build
```

### Running Tests
```bash
buffalo test
```

### Database Operations

**IMPORTANT: Buffalo v0.18.14+ uses `soda` for database migrations, NOT `buffalo pop`**

```bash
# Run migrations
soda migrate up

# Create databases  
soda create -a

# Reset database (drop, create, migrate)
soda reset

# Reset test database specifically
GO_ENV=test soda reset

# Check migration status
soda migrate status

# Create new migration
soda generate migration create_posts

# Rollback migrations
soda migrate down
```

**Legacy Documentation Note**: Older Buffalo documentation may reference `buffalo pop` commands, but these are not available in Buffalo v0.18.14+. Always use `soda` commands directly.

**Testing Database Setup**: 
- Buffalo tests automatically run migrations before each test suite
- Use `GO_ENV=test soda reset` if you need to manually reset the test database
- The test environment uses a separate database specified in `database.yml`

### Asset Management
```bash
buffalo generate webpack  # Generate webpack config
```

## File Watching

Buffalo watches these file types by default:
- `.go` files (actions, models, etc.)
- `.plush.html` files (templates)  
- `.js` files in assets
- `.css` files in assets

Changes to these files trigger automatic rebuilds.

## Environment Variables

Key environment variables for development:

```bash
GO_ENV=development     # Set automatically by buffalo dev
PORT=3000             # Server port
DATABASE_URL=         # Override database connection
```

## Docker Integration

When using Docker for services like PostgreSQL:

1. Start services: `docker-compose up -d`
2. Start Buffalo: `buffalo dev`
3. Buffalo will connect to containerized services

## Logs and Debugging

Buffalo dev server shows:
- HTTP requests and responses
- Template compilation errors
- Database queries (if enabled)
- Asset pipeline status
- Go compilation errors