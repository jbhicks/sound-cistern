# Buffalo Database Migration Commands - v0.18.14+

## Important Changes in Buffalo v0.18.14+

Buffalo v0.18.14+ no longer includes the `pop` plugin by default. Database operations now use the `soda` command directly.

### ❌ Old Commands (No Longer Work)
```bash
buffalo pop migrate     # ❌ Error: unknown command "pop" 
buffalo pop create -a   # ❌ Error: unknown command "pop"
buffalo pop reset       # ❌ Error: unknown command "pop"
```

### ✅ New Commands (Correct for v0.18.14+)
```bash
soda migrate up         # Run pending migrations
soda create -a          # Create all databases
soda reset              # Drop, create, and migrate database
soda migrate status     # Check migration status
soda generate migration # Create new migration
soda migrate down       # Rollback migrations
```

## Environment-Specific Commands

### Development Database
```bash
soda migrate up                    # Uses development config by default
soda reset                         # Reset development database
```

### Test Database
```bash
GO_ENV=test soda reset            # Reset test database
GO_ENV=test soda migrate up       # Run migrations on test database
```

### Production Database
```bash
GO_ENV=production soda migrate up  # Run migrations on production
```

## Common Database Tasks

### Setup New Environment
```bash
soda create -a          # Create databases for all environments
soda migrate up         # Run migrations for development
GO_ENV=test soda migrate up  # Run migrations for test
```

### Reset Everything
```bash
soda reset              # Reset development database
GO_ENV=test soda reset  # Reset test database
```

### Check Status
```bash
soda migrate status     # See which migrations have been applied
```

## Buffalo Testing Integration

Buffalo's test suite automatically:
1. Drops and recreates the test database before each test run
2. Runs all migrations to set up the schema
3. Provides each test with a clean database state

You should NOT need to manually run migrations for testing unless:
- You're debugging database schema issues
- You need to reset the test database outside of the test suite
- You're developing new migrations

## Troubleshooting

### "relation does not exist" errors in tests
This usually means:
1. Migration files exist but weren't applied to test database
2. Test database schema is out of sync

**Solution**: Reset test database
```bash
GO_ENV=test soda reset
```

### Migration version conflicts
**Solution**: Check migration status and reset if needed
```bash
soda migrate status
soda reset  # if needed
```

### Database connection errors
1. Ensure PostgreSQL is running
2. Check `database.yml` configuration
3. Verify environment variables are set correctly

## Migration Best Practices

1. **Always test migrations** in development before applying to production
2. **Create reversible migrations** with proper `down` files
3. **Use descriptive migration names** that explain what they do
4. **Keep migrations small and focused** on single changes
5. **Test database resets** to ensure migrations work from scratch

## Integration with Buffalo Development Workflow

- **Starting development**: `make dev` handles database startup
- **Running tests**: `buffalo test` or `make test` handles test database
- **Creating migrations**: `soda generate migration create_new_table`
- **Applying migrations**: `soda migrate up` (development auto-applies via Buffalo dev server)
