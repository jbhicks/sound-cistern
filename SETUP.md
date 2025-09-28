# Setup Guide

## Prerequisites
- Go 1.21+
- Podman or Docker
- Make

## Quick Start
```bash
# Clone and setup
git clone <template-repo>
cd <project-name>

# Complete setup (installs all dependencies)
make setup-complete

# Start development server
make dev
```

## Port Configuration
- Development server: http://localhost:3001
- Database: localhost:5432
- Admin panel: http://localhost:3001/admin

## Notes
- If you see `soda: command not found`, run `make install-soda` or source your shell config.
- If Go binaries are not in your PATH, run `bash scripts/setup-shell.sh` and restart your shell.
- For Docker-only environments, use `make db-up-docker`.
