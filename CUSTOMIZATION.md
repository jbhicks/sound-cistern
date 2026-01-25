# Template Customization Guide

## Module Name
- Change the module name in `go.mod` to match your project.

## Database
- PocketBase uses embedded SQLite database stored in `/pb_data/data.db`
- No configuration needed - database is created automatically on first run
- For production use with PostgreSQL, see [PocketBase documentation](https://pocketbase.io/docs/)

## Project Customization
- Update project metadata in `README.md` and `SETUP.md`
- Adjust environment variables in `.env` file as required
- Configure Soundcloud OAuth credentials (see `.env.example`)

## Migration Workflow
- Database migrations are in `/pb_migrations/` directory (Go files)
- Migrations run automatically when starting the application
- To create new migrations, use PocketBase migration commands or manually create Go migration files

## Template Development
- Templ templates are in `/views/` directory (`.templ` files)
- Run `make templ` to generate Go code from Templ templates
- Generated files: `*_templ.go` (do not edit these directly)
- Edit `.templ` files and regenerate

## Additional Notes
- See `/docs/` for more on PocketBase, Templ, Pico.css, and dependency management
- For further customization, review the Makefile and scripts for automation points
