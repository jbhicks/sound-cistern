# Structured Logging System

This document describes the structured logging system implemented in the Go SaaS template application.

## Overview

The structured logging system provides comprehensive audit trails, security event tracking, and user action logging throughout the application. It uses logrus as the underlying logging library and supports multiple output formats and destinations.

## Features

- **Structured logging** with key-value pairs for better searchability and analysis
- **Multiple log levels**: Debug, Info, Warn, Error, Fatal
- **Specialized logging types**: Audit, UserAction, SecurityEvent
- **Configurable output**: Console, file, or both
- **Multiple formats**: JSON (production) or text (development)
- **Context awareness**: Automatic inclusion of request IDs, user IDs, and HTTP details
- **Environment-based configuration**: LOG_LEVEL, LOG_FORMAT, LOG_OUTPUT, LOG_FILE_PATH

## Architecture

### Core Components

#### Service (`pkg/logging/service.go`)
- Main logging service implementation
- Wraps logrus with structured field support
- Provides specialized methods for different log types

#### Configuration (`pkg/logging/config.go`)
- Environment-based configuration system
- Supports development and production settings
- Configurable log levels, formats, and output destinations

#### Global Instance (`pkg/logging/global.go`)
- Global logging service instance management
- Convenience functions for application-wide logging
- Thread-safe singleton pattern

## Usage

### Basic Logging

```go
import "github.com/jbhicks/sound-cistern/pkg/logging"

// Simple info message
logging.Info("User profile updated")

// Info with structured fields
logging.Info("User profile updated", logging.Fields{
    "user_id": userID,
    "fields_changed": []string{"email", "name"},
})

// Error logging
logging.Error("Database connection failed", err, logging.Fields{
    "database": "primary",
    "retry_count": 3,
})
```

### Context-Aware Logging

```go
// In Buffalo actions, pass the context for automatic field inclusion
func (v UsersResource) Update(c buffalo.Context) error {
    logging.UserAction(c, "profile_update", "user", logging.Fields{
        "fields_changed": changedFields,
    })
    
    // Context automatically includes:
    // - request_id
    // - method (PUT/PATCH)
    // - path (/users/123)
    // - user_id (if authenticated)
    
    return c.Render(http.StatusOK, r.JSON(user))
}
```

### Specialized Logging Types

#### Audit Logging
For administrative and security-critical events:

```go
logging.Audit("admin_user_deletion", logging.Fields{
    "admin_user_id": adminID,
    "target_user_id": targetUserID,
    "reason": "policy_violation",
})
```

#### Security Events
For security-related incidents:

```go
logging.SecurityEvent(c, "failed_login_attempt", logging.Fields{
    "username": username,
    "ip_address": c.Request().RemoteAddr,
    "attempt_number": attemptCount,
})
```

#### User Actions
For tracking user behavior and activities:

```go
logging.UserAction(c, "login", "session", logging.Fields{
    "login_method": "password",
    "remember_me": rememberMe,
})
```

## Configuration

### Environment Variables

| Variable | Description | Default | Examples |
|----------|-------------|---------|----------|
| `LOG_LEVEL` | Minimum log level | `info` | `debug`, `info`, `warn`, `error` |
| `LOG_FORMAT` | Output format | `text` | `text`, `json` |
| `LOG_OUTPUT` | Output destination | `stdout` | `stdout`, `file`, `both` |
| `LOG_FILE_PATH` | Log file path (when using file output) | `./logs/app.log` | `/var/log/app.log` |

### Example Configurations

#### Development
```bash
LOG_LEVEL=debug
LOG_FORMAT=text
LOG_OUTPUT=stdout
```

#### Production
```bash
LOG_LEVEL=info
LOG_FORMAT=json
LOG_OUTPUT=both
LOG_FILE_PATH=/var/log/saas-app/app.log
```

## Log Levels

- **Debug**: Detailed information for debugging issues
- **Info**: General information about application flow
- **Warn**: Warning messages for potentially harmful situations
- **Error**: Error events that might still allow the application to continue
- **Fatal**: Very severe error events that will presumably lead the application to abort

## Structured Fields

### Standard Fields

The logging system automatically includes standard fields in appropriate contexts:

- `request_id`: Unique identifier for each HTTP request
- `method`: HTTP method (GET, POST, etc.)
- `path`: Request path
- `user_id`: Currently authenticated user ID
- `timestamp`: ISO 8601 formatted timestamp
- `level`: Log level
- `msg`: Log message

### Custom Fields

Add context-specific information using the `Fields` type:

```go
logging.Info("Processing payment", logging.Fields{
    "user_id": userID,
    "amount": paymentAmount,
    "currency": "USD",
    "payment_method": "credit_card",
    "transaction_id": transactionID,
})
```

## Integration Points

### Application Startup
- `cmd/app/main.go`: Server startup and fatal error logging
- `models/models.go`: Database connection error logging

### Authentication System
- `actions/auth.go`: Login/logout events, failed authentication attempts
- Security event logging for suspicious activities

### User Management
- `actions/users.go`: User registration, profile updates, authentication checks
- User action logging for audit trails

### Admin Operations
- `actions/admin.go`: Administrative actions, user management operations
- Audit logging for all admin activities

## Security Considerations

### Sensitive Data
- Never log passwords, tokens, or other sensitive credentials
- Use field filtering for PII (personally identifiable information)
- Consider data retention policies for audit logs

### Log Access
- Restrict access to log files in production
- Consider log aggregation services for centralized management
- Implement log rotation to manage disk space

## Testing

The logging system includes comprehensive tests:

```bash
# Run logging system tests
buffalo test pkg/logging

# Run all tests including logging integration
buffalo test
```

Test coverage includes:
- Configuration loading and validation
- All logging methods (Info, Debug, Warn, Error, Fatal, Audit, UserAction, SecurityEvent)
- Field combination and context extraction
- Error handling and edge cases

## Performance Considerations

### Log Levels
- Set appropriate log levels in production to avoid performance impact
- Debug logging should be disabled in production environments

### Structured Fields
- Use consistent field names across the application
- Avoid deeply nested objects in log fields
- Consider field cardinality for log aggregation systems

### File Output
- Enable log rotation when using file output
- Monitor disk space usage
- Consider asynchronous logging for high-throughput applications

## Best Practices

### Naming Conventions
- Use snake_case for field names: `user_id`, `request_id`
- Use descriptive but concise field names
- Maintain consistency across similar operations

### Message Content
- Write clear, actionable log messages
- Include relevant context in the message
- Use structured fields for variable data

### Error Logging
- Always include error details in structured fields
- Log errors at the point where they can be best contextualized
- Include recovery actions or next steps where appropriate

### Security Logging
- Log all authentication attempts (success and failure)
- Log authorization failures
- Log administrative actions
- Include relevant security context (IP addresses, user agents)

## Troubleshooting

### Common Issues

#### Missing Log Output
- Check LOG_LEVEL environment variable
- Verify LOG_OUTPUT configuration
- Ensure file permissions for file output

#### Performance Issues
- Review log level settings
- Check for excessive debug logging in production
- Monitor log file sizes and rotation

#### Missing Context
- Ensure Buffalo context is passed to UserAction/SecurityEvent methods
- Verify middleware is properly configured
- Check authentication middleware for user context

### Debugging

Enable debug logging temporarily:
```bash
LOG_LEVEL=debug buffalo dev
```

Check current configuration:
```go
config := logging.NewConfig()
fmt.Printf("Config: %+v\n", config)
```

## Examples

### Complete Authentication Flow
```go
func (a AuthResource) Attempt(c buffalo.Context) error {
    username := c.Param("username")
    password := c.Param("password")
    
    // Log login attempt
    logging.Info("Login attempt started", logging.Fields{
        "username": username,
        "ip_address": c.Request().RemoteAddr,
    })
    
    user, err := models.FindUserByUsername(username)
    if err != nil {
        // Log failed lookup
        logging.SecurityEvent(c, "failed_login_attempt", logging.Fields{
            "username": username,
            "reason": "user_not_found",
        })
        return c.Render(http.StatusUnauthorized, r.JSON(map[string]string{
            "error": "Invalid credentials",
        }))
    }
    
    if !user.ValidatePassword(password) {
        // Log invalid password
        logging.SecurityEvent(c, "failed_login_attempt", logging.Fields{
            "username": username,
            "user_id": user.ID,
            "reason": "invalid_password",
        })
        return c.Render(http.StatusUnauthorized, r.JSON(map[string]string{
            "error": "Invalid credentials",
        }))
    }
    
    // Successful login
    c.Session().Set("user_id", user.ID)
    
    logging.UserAction(c, "login", "session", logging.Fields{
        "user_id": user.ID,
        "username": user.Username,
        "login_method": "password",
    })
    
    return c.Render(http.StatusOK, r.JSON(user))
}
```

### Admin User Management
```go
func (a AdminResource) DeleteUser(c buffalo.Context) error {
    userID := c.Param("user_id")
    adminUser := c.Value("current_user").(*models.User)
    
    // Log admin action
    logging.Audit("admin_user_deletion_attempt", logging.Fields{
        "admin_user_id": adminUser.ID,
        "target_user_id": userID,
    })
    
    user := &models.User{}
    if err := models.DB.Find(user, userID); err != nil {
        logging.Error("Failed to find user for deletion", err, logging.Fields{
            "user_id": userID,
            "admin_user_id": adminUser.ID,
        })
        return c.Render(http.StatusNotFound, r.JSON(map[string]string{
            "error": "User not found",
        }))
    }
    
    if err := models.DB.Destroy(user); err != nil {
        logging.Error("Failed to delete user", err, logging.Fields{
            "user_id": userID,
            "admin_user_id": adminUser.ID,
        })
        return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{
            "error": "Failed to delete user",
        }))
    }
    
    // Log successful deletion
    logging.Audit("admin_user_deleted", logging.Fields{
        "admin_user_id": adminUser.ID,
        "deleted_user_id": userID,
        "deleted_username": user.Username,
    })
    
    return c.Render(http.StatusOK, r.JSON(map[string]string{
        "message": "User deleted successfully",
    }))
}
```

This structured logging system provides comprehensive observability into your application's behavior, security events, and user actions, making it easier to monitor, debug, and audit your SaaS application.
