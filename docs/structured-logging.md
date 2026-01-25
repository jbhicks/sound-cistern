# Structured Logging System

This document describes the structured logging system for Sound Cistern.

## Overview

The structured logging system provides comprehensive audit trails, security event tracking, and user action logging throughout the application. PocketBase provides built-in logging capabilities with SQLite-based log storage.

## Features

- **Built-in PocketBase Logging**: Automatic request and system logging
- **SQLite Log Storage**: Logs stored in `/pb_data/logs.db`
- **Admin Dashboard**: View and filter logs through PocketBase admin UI at `/_/`
- **Custom Application Logging**: Add custom log entries for business logic
- **Go Standard Library**: Use `log` package or PocketBase's logger
- **Environment-based configuration**: `LOG_LEVEL` environment variable

## Architecture

### PocketBase Logging System

#### Built-in Request Logger
- Automatically logs all HTTP requests
- Stores in `/pb_data/logs.db`
- Accessible via admin UI at `/_/logs`

#### Custom Application Logging
```go
// Access PocketBase's logger
app.Logger().Info("Custom log message", "key", "value")
app.Logger().Error("Error occurred", "error", err)
```

#### Log Storage
- SQLite database: `/pb_data/logs.db`
- Automatically managed by PocketBase
- Queryable through admin interface

## Usage

### Basic Logging with PocketBase

```go
// In main.go or route handlers
app.Logger().Info("User profile updated", 
    "user_id", userID,
    "fields_changed", []string{"email", "name"},
)

// Error logging
app.Logger().Error("Database operation failed", 
    "error", err,
    "collection", "users",
)
```

### Custom Route Handlers

```go
func handleUserUpdate(app *pocketbase.PocketBase) echo.HandlerFunc {
    return func(c echo.Context) error {
        userID := c.Param("id")
        
        // Log the action
        app.Logger().Info("User update request",
            "user_id", userID,
            "ip", c.RealIP(),
        )
        
        // Your logic here
        
        return c.JSON(200, map[string]string{"status": "ok"})
    }
}
```

### Logging Best Practices

#### Security Events
```go
// Failed login attempt
app.Logger().Warn("Failed login attempt",
    "username", username,
    "ip", c.RealIP(),
    "attempt", attemptCount,
)
```

#### User Actions
```go
// Successful login
app.Logger().Info("User login",
    "user_id", user.Id,
    "username", user.Username(),
    "method", "password",
)
```

#### System Events
```go
// Application startup
app.Logger().Info("Application started",
    "port", 8090,
    "environment", os.Getenv("GO_ENV"),
)
```

## Configuration

### Environment Variables

| Variable | Description | Default | Examples |
|----------|-------------|---------|----------|
| `LOG_LEVEL` | Minimum log level | `info` | `debug`, `info`, `warn`, `error` |

### Log Storage Location

- **Database**: `/pb_data/logs.db` (SQLite)
- **Automatic Cleanup**: PocketBase manages log retention
- **Admin Access**: View logs at `http://localhost:8090/_/logs`

### Example Configurations

#### Development
```bash
LOG_LEVEL=debug
./sound-cistern serve --dev
```

#### Production
```bash
LOG_LEVEL=info
./sound-cistern serve
```

## Log Levels

PocketBase supports standard log levels:

- **Debug**: Detailed debugging information (use `--dev` flag)
- **Info**: General informational messages
- **Warn**: Warning messages for potential issues
- **Error**: Error events requiring attention

## Viewing Logs

### Admin Dashboard
1. Start application: `./sound-cistern serve`
2. Navigate to: `http://localhost:8090/_/`
3. Login with admin credentials
4. Go to Logs section
5. Filter by level, date, or search terms

### Programmatic Access
```go
// Query logs.db directly if needed (advanced)
// PocketBase admin API provides log access endpoints
```

### Production Log Access
- SSH to server
- View logs in `/pb_data/logs.db`
- Use PocketBase admin UI (recommended)
- Export logs for analysis

## Integration Points

### Application Startup
- `main.go`: Server startup and initialization logging
- PocketBase bootstrap events automatically logged

### Authentication System
- PocketBase built-in auth automatically logged
- Custom auth handlers can add additional logging
- Failed authentication attempts tracked

### API Routes
- All HTTP requests automatically logged by PocketBase
- Custom route handlers can add business logic logging
- Response times and status codes tracked

## Security Considerations

### Sensitive Data
- Never log passwords, tokens, or API secrets
- Be careful with PII (personally identifiable information)
- PocketBase admin UI access should be restricted in production

### Log Access
- Protect `/pb_data/logs.db` file permissions
- Secure admin panel with strong password
- Consider firewall rules to restrict `/_/` access in production
- Use HTTPS in production for admin access

## Performance Considerations

### Log Levels
- Use `info` or `warn` in production
- Enable `debug` only during development with `--dev` flag

### Database Size
- Monitor `/pb_data/logs.db` size
- PocketBase handles log rotation automatically
- Consider periodic cleanup for very high-traffic applications

### Query Performance
- Logs indexed by PocketBase for fast searches
- Use admin UI filters for efficient log retrieval

## Best Practices

### Logging Key-Value Pairs
```go
// Good: Clear key-value pairs
app.Logger().Info("User action", 
    "action", "profile_update",
    "user_id", userID,
)

// Avoid: Mixing message and data
app.Logger().Info(fmt.Sprintf("User %s updated profile", userID))
```

### Error Logging
```go
// Always include error details
if err != nil {
    app.Logger().Error("Operation failed",
        "error", err.Error(),
        "operation", "user_update",
        "user_id", userID,
    )
    return err
}
```

### Security Events
- Log authentication attempts (success and failure)
- Log authorization failures
- Track IP addresses for suspicious activity
- Monitor admin actions

## Troubleshooting

### Common Issues

#### No Logs Appearing
- Check that application is running
- Verify `/pb_data/logs.db` exists
- Check LOG_LEVEL environment variable
- Access admin UI to confirm logs are being written

#### Can't Access Admin UI
- Ensure admin user created (first startup)
- Check URL: `http://localhost:8090/_/`
- Verify port 8090 is accessible
- Check for firewall/network issues

### Debugging

Enable debug logging:
```bash
LOG_LEVEL=debug ./sound-cistern serve --dev
```

Check log database:
```bash
ls -lh /home/josh/sound-cistern/pb_data/logs.db
```

## Examples

### Custom Route with Logging
```go
func registerCustomRoutes(app *pocketbase.PocketBase) {
    app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
        e.Router.POST("/api/custom", func(c echo.Context) error {
            app.Logger().Info("Custom API called",
                "ip", c.RealIP(),
                "user_agent", c.Request().UserAgent(),
            )
            
            // Your business logic here
            
            return c.JSON(200, map[string]string{"status": "ok"})
        })
        
        return nil
    })
}
```

### Error Handling with Logging
```go
func handleOperation(app *pocketbase.PocketBase, c echo.Context) error {
    userID := c.Param("id")
    
    user, err := app.Dao().FindRecordById("users", userID)
    if err != nil {
        app.Logger().Error("User lookup failed",
            "error", err.Error(),
            "user_id", userID,
            "ip", c.RealIP(),
        )
        return c.JSON(404, map[string]string{
            "error": "User not found",
        })
    }
    
    app.Logger().Info("User operation successful",
        "user_id", userID,
        "operation", "profile_view",
    )
    
    return c.JSON(200, user)
}
```

This logging system leverages PocketBase's built-in capabilities while allowing custom application logging for business logic, providing comprehensive observability into your application's behavior.
