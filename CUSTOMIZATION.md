# Template Customization Guide

## Module Name
- Change the module name in `go.mod` to match your project.

## Database Names
- Update database names in `database.yml` and `config/buffalo-app.toml` as needed.
- Run `make db-create` to create new databases.

## Project Customization
- Update project metadata in `README.md` and `SETUP.md`.
- Adjust environment variables in `.env` or config files as required.

## Migration Workflow
- To update from template changes, use `git merge` or `git cherry-pick` for new features.
- Run `make migrate` after pulling updates to apply new migrations.

## Additional Notes
- See `/docs/` for more on Buffalo, Pico.css, and dependency management.
- For further customization, review the Makefile and scripts for automation points.
