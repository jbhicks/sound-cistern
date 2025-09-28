.PHONY: help dev setup db-up db-down db-reset test clean build admin migrate db-status db-logs health check-deps install-deps update-deps
# Add clean-caches target for clearing Go and gopls caches
.PHONY: clean-caches

# Default target
help:
	@echo "🚀 My Go SaaS Template - Development Commands"
	@echo ""
	@echo "Quick Start:"
	@echo "  setup      - 🔧 Initial setup: start database, run migrations, install deps"
	@echo "  dev        - 🏃 Start database and run Buffalo development server"
	@echo "  admin      - 👑 Promote first user to admin role"
	@echo ""
	@echo "Database Commands:"
	@echo "  db-up      - 🗄️  Start PostgreSQL database with Podman"
	@echo "  db-down    - ⬇️  Stop PostgreSQL database"
	@echo "  db-reset   - 🔄 Reset database (drop, create, migrate)"
	@echo "  db-status  - 📊 Check database container status"
	@echo "  db-logs    - 📋 Show database container logs"
	@echo "  migrate    - 🔀 Run database migrations"
	@echo ""
	@echo "Development:"
	@echo "  test            - 🧪 Run all tests with Buffalo (recommended)"
	@echo "  test-fast       - ⚡ Run Buffalo tests without database setup"
	@echo "  test-resilient  - 🛡️  Run tests with automatic database startup"
	@echo "  validate-templates - 🎨 Validate admin template structure"
	@echo "  build           - 🔨 Build the application for production"
	@echo "  health          - 🏥 Check system health (dependencies, database, etc.)"
	@echo "  clean           - 🧹 Stop all services and clean up containers"
	@echo "  clean-caches    - 🧹 Clear Go build, module, and gopls caches"
	@echo ""
	@echo "Dependencies:"
	@echo "  check-deps  - ✅ Check if all required dependencies are installed"
	@echo "  install-deps - 📦 Install missing dependencies (where possible)"
	@echo "  update-deps - 🔄 Update all frontend dependencies (JS/CSS) to latest versions"

# Check if all required dependencies are installed
check-deps:
	@echo "🔍 Checking required dependencies..."
	@error_count=0; \
	if ! command -v go >/dev/null 2>&1; then \
		echo "❌ Go is not installed. Please install Go 1.19+ from https://golang.org/dl/"; \
		error_count=$$((error_count + 1)); \
	else \
		echo "✅ Go is installed: $$(go version)"; \
	fi; \
	if ! command -v buffalo >/dev/null 2>&1; then \
		echo "❌ Buffalo CLI is not installed. Run: go install github.com/gobuffalo/cli/cmd/buffalo@latest"; \
		error_count=$$((error_count + 1)); \
	else \
		echo "✅ Buffalo CLI is installed: $$(buffalo version)"; \
	fi; \
	if ! command -v podman-compose >/dev/null 2>&1; then \
		if ! command -v docker-compose >/dev/null 2>&1; then \
			echo "❌ Neither podman-compose nor docker-compose found. Please install Podman or Docker."; \
			error_count=$$((error_count + 1)); \
		else \
			echo "✅ Docker Compose is installed: $$(docker-compose version)"; \
		fi; \
	else \
		echo "✅ Podman Compose is installed: $$(podman-compose version)"; \
	fi; \
	if [ $$error_count -gt 0 ]; then \
		echo ""; \
		echo "❌ $$error_count dependencies are missing. Please install them before continuing."; \
		echo "Run 'make install-deps' to install dependencies where possible."; \
		exit 1; \
	else \
		echo ""; \
		echo "✅ All dependencies are installed and ready!"; \
	fi

# Install missing dependencies where possible
install-deps:
	@echo "📦 Installing missing dependencies..."
	@if ! command -v buffalo >/dev/null 2>&1; then \
		echo "Installing Buffalo CLI..."; \
		go install github.com/gobuffalo/cli/cmd/buffalo@latest || echo "Failed to install Buffalo CLI"; \
	fi
	@echo "✅ Dependency installation complete. Run 'make check-deps' to verify."

# --- Soda installation and checks ---
install-soda:
	@echo "📦 Installing soda database tool..."
	@go install github.com/gobuffalo/pop/v6/soda@latest
	@echo "✅ Soda installed successfully"

check-soda:
	@which soda > /dev/null || (echo "❌ Soda not found. Run 'make install-soda'" && exit 1)
	@echo "✅ Soda is available"

db-create:
	@echo "🗄️ Creating databases..."
	@soda create -a
	@echo "✅ Databases created successfully"

# Start database and development server with full health checks
dev: check-deps db-up
	@echo "🔍 Waiting for database to be ready..."
	@if ! ./scripts/wait-for-postgres.sh; then \
		echo "❌ Database failed to start. Check 'make db-logs' for details."; \
		exit 1; \
	fi
	@echo "🚀 Starting Buffalo development server..."
	@echo "📱 Visit http://127.0.0.1:3000 to see your application"
	@echo "🔥 Hot reload is enabled - changes will be reflected automatically"
	@echo "🔍 Checking for processes on port 3000..."
	@lsof -ti:3000 | xargs -r kill -9 2>/dev/null || echo "No processes on port 3000 to kill."
	HOST=0.0.0.0:3000 buffalo dev || (echo "❌ Buffalo failed to start. Check the output above for errors." && exit 1)

# Initial setup with comprehensive checks
setup: check-deps db-up migrate
	@echo "🎉 Setup complete!"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Run 'make dev' to start the development server"
	@echo "  2. Visit http://127.0.0.1:3000 to see your application"
	@echo "  3. Create a user account through the web interface"
	@echo "  4. Run 'make admin' to promote your user to admin"
	@echo ""
	@echo "🔧 Development commands available: make help"
	@echo "🎉 Complete setup finished!"

# Promote first user to admin with better error handling
admin: db-up
	@echo "👑 Setting up admin user..."
	@if ! ./scripts/wait-for-postgres.sh; then \
		echo "❌ Database is not ready. Cannot promote user to admin."; \
		exit 1; \
	fi
	@echo "🔍 Looking for users to promote..."
	@if buffalo task db:promote_admin 2>/dev/null; then \
		echo "✅ User successfully promoted to admin role!"; \
		echo "🎯 You can now access the admin panel at http://127.0.0.1:3000/admin"; \
	else \
		echo "⚠️  No users found to promote. Please:"; \
		echo "   1. Create a user account through the web interface first"; \
		echo "   2. Then run 'make admin' again"; \
	fi

# Run database migrations with better error handling
migrate: db-up
	@echo "🔀 Running database migrations..."
	@if ! ./scripts/wait-for-postgres.sh; then \
		echo "❌ Database is not ready. Cannot run migrations."; \
		exit 1; \
	fi
	@echo "📊 Checking migration status..."
	@if buffalo pop migrate 2>/dev/null || soda migrate 2>/dev/null; then \
		echo "✅ Migrations completed successfully!"; \
	else \
		echo "❌ Migration failed. Check database connection and migration files."; \
		exit 1; \
	fi

# Start PostgreSQL database with comprehensive checks
db-up:
	@echo "🗄️  Starting PostgreSQL database..."
	@if ! command -v podman-compose >/dev/null 2>&1; then \
		if ! command -v docker-compose >/dev/null 2>&1; then \
			echo "❌ Neither podman-compose nor docker-compose found."; \
			echo "Please install Podman (recommended) or Docker."; \
			echo "Podman: https://podman.io/getting-started/installation"; \
			echo "Docker: https://docs.docker.com/get-docker/"; \
			exit 1; \
		else \
			echo "🐳 Using Docker Compose..."; \
			docker-compose up -d postgres || (echo "❌ Failed to start database with Docker Compose" && exit 1); \
		fi; \
	else \
		echo "🔷 Using Podman Compose..."; \
		podman-compose up -d postgres || (echo "❌ Failed to start database with Podman Compose" && exit 1); \
	fi
	@echo "✅ Database container started successfully."

# Stop PostgreSQL database
db-down:
	@echo "⬇️  Stopping PostgreSQL database..."
	@if command -v podman-compose >/dev/null 2>&1; then \
		podman-compose down || echo "Database was not running."; \
	elif command -v docker-compose >/dev/null 2>&1; then \
		docker-compose down || echo "Database was not running."; \
	else \
		echo "❌ No compose command found."; \
	fi
	@echo "✅ Database stopped."

# Check database status with detailed information
db-status:
	@echo "📊 Database container status:"
	@if command -v podman-compose >/dev/null 2>&1; then \
		podman-compose ps postgres 2>/dev/null || echo "❌ Database container not found (Podman)"; \
		echo ""; \
		echo "📡 Container health:"; \
		if podman-compose exec postgres pg_isready -U postgres >/dev/null 2>&1; then \
			echo "✅ PostgreSQL is ready and accepting connections"; \
		else \
			echo "❌ PostgreSQL is not ready"; \
		fi; \
	elif command -v docker-compose >/dev/null 2>&1; then \
		docker-compose ps postgres 2>/dev/null || echo "❌ Database container not found (Docker)"; \
		echo ""; \
		echo "📡 Container health:"; \
		if docker-compose exec postgres pg_isready -U postgres >/dev/null 2>&1; then \
			echo "✅ PostgreSQL is ready and accepting connections"; \
		else \
			echo "❌ PostgreSQL is not ready"; \
		fi; \
	else \
		echo "❌ No compose command found."; \
	fi

# Show database logs
db-logs:
	@echo "📋 Database container logs (last 50 lines):"
	@if command -v podman-compose >/dev/null 2>&1; then \
		podman-compose logs postgres --tail 50 || echo "❌ Cannot access database logs"; \
	elif command -v docker-compose >/dev/null 2>&1; then \
		docker-compose logs postgres --tail 50 || echo "❌ Cannot access database logs"; \
	else \
		echo "❌ No compose command found."; \
	fi

# Reset database with safety confirmations
db-reset: 
	@echo "🔄 Database Reset - This will DELETE ALL DATA!"
	@echo "Are you sure you want to reset the database? [y/N]" && read ans && [ $${ans:-N} = y ]
	@echo "🗄️  Starting database..."
	@$(MAKE) db-up
	@echo "⏳ Waiting for database to be ready..."
	@if ! ./scripts/wait-for-postgres.sh; then \
		echo "❌ Database failed to start. Cannot reset."; \
		exit 1; \
	fi
	@echo "🗑️  Dropping development database..."
	@buffalo pop drop -e development 2>/dev/null || echo "Database drop failed (may not exist)"
	@echo "🏗️  Creating development database..."
	@buffalo pop create -e development || (echo "❌ Database create failed" && exit 1)
	@echo "🔀 Running migrations..."
	@buffalo pop migrate -e development || soda migrate || (echo "❌ Migration failed" && exit 1)
	@echo "✅ Database reset complete!"
	@echo "🎯 You can now run 'make dev' to start the development server"

# Run tests with comprehensive setup
test: check-deps db-up validate-templates
	@echo "🧪 Running test suite with Buffalo..."
	@if ! ./scripts/wait-for-postgres.sh; then \
		echo "❌ Database is not ready. Cannot run tests."; \
		exit 1; \
	fi
	@echo "🔄 Setting up test database..."
	@GO_ENV=test soda create -a >/dev/null 2>&1 || true
	@GO_ENV=test soda migrate up >/dev/null 2>&1 || true
	@echo "🏃 Executing Buffalo tests..."
	@if buffalo test; then \
		echo "✅ All tests passed!"; \
	else \
		echo "❌ Some tests failed. Check the output above for details."; \
		exit 1; \
	fi

# Run Buffalo tests quickly (assumes database is already running)
test-fast: check-deps
	@echo "⚡ Running Buffalo tests (fast mode)..."
	@echo "🏃 Executing Buffalo tests..."
	@if buffalo test; then \
		echo "✅ All tests passed!"; \
	else \
		echo "❌ Some tests failed. Check the output above for details."; \
		exit 1; \
	fi

# Resilient test command that handles database startup automatically
test-resilient: check-deps
	@echo "🔄 Running resilient test suite..."
	@echo "🔍 Checking if database is running..."
	@if ! podman-compose ps | grep -q "postgres.*Up" 2>/dev/null; then \
		echo "🗄️  Database not running, starting it..."; \
		$(MAKE) db-up; \
		sleep 3; \
	else \
		echo "✅ Database is already running"; \
	fi
	@if ! ./scripts/wait-for-postgres.sh; then \
		echo "❌ Database failed to start or become ready. Cannot run tests."; \
		exit 1; \
	fi
	@echo "🔄 Setting up test database..."
	@GO_ENV=test soda create -a >/dev/null 2>&1 || true
	@GO_ENV=test soda migrate up >/dev/null 2>&1 || true
	@echo "🏃 Executing Buffalo tests..."
	@if buffalo test; then \
		echo "✅ All tests passed!"; \
	else \
		echo "❌ Some tests failed. Check the output above for details."; \
		exit 1; \
	fi

# Template validation
validate-templates:
	@echo "🎨 Validating admin template structure..."
	@./scripts/validate-templates.sh

# Clear Go build, module, and gopls caches
clean-caches:
	@echo "🧹 Clearing Go and language server caches..."
	@go clean -cache || echo "Go cache already clean"
	@go clean -modcache || echo "Module cache already clean" 
	@echo "💡 If VS Code still shows errors, restart the Go language server:"
	@echo "   Ctrl+Shift+P -> 'Go: Restart Language Server'"
	@echo "✅ Cache cleanup complete!"

# Clean up everything with confirmation
clean:
	@echo "🧹 Cleaning up development environment..."
	@echo "This will stop all services and remove containers. Continue? [y/N]" && read ans && [ $${ans:-N} = y ]
	@echo "🛑 Stopping all services..."
	@if command -v podman-compose >/dev/null 2>&1; then \
		podman-compose down || echo "Services were not running."; \
		echo "🗑️  Cleaning up containers and volumes..."; \
		podman system prune -f --volumes 2>/dev/null || echo "Cleanup completed with warnings."; \
	elif command -v docker-compose >/dev/null 2>&1; then \
		docker-compose down || echo "Services were not running."; \
		echo "🗑️  Cleaning up containers and volumes..."; \
		docker system prune -f --volumes 2>/dev/null || echo "Cleanup completed with warnings."; \
	else \
		echo "❌ No compose command found."; \
	fi
	@echo "✅ Clean complete!"

# --- Docker alternative for db-up ---
db-up-docker:
	@echo "🐳 Starting PostgreSQL with Docker..."
	@docker-compose up -d postgres
	@echo "✅ PostgreSQL container started"

# Update all frontend dependencies to latest versions
update-deps:
	@echo "🔄 Updating frontend dependencies to latest versions..."
	@echo ""
	
	# Check for required tools
	@if ! command -v curl >/dev/null 2>&1; then \
		echo "❌ curl is required but not installed."; \
		exit 1; \
	fi
	
	@echo "📦 Checking latest versions..."
	
	# Get latest Quill.js version
	@echo "🔍 Checking Quill.js..."
	@QUILL_VERSION=$$(curl -s "https://registry.npmjs.org/quill/latest" | grep '"version"' | head -1 | sed 's/.*"version":"\([^"]*\)".*/\1/'); \
	echo "   Latest Quill.js version: $$QUILL_VERSION"; \
	echo "   📥 Downloading Quill.js $$QUILL_VERSION..."; \
	curl -s -o public/css/quill.snow.css "https://cdn.jsdelivr.net/npm/quill@$$QUILL_VERSION/dist/quill.snow.css" && \
	curl -s -o public/js/quill.min.js "https://cdn.jsdelivr.net/npm/quill@$$QUILL_VERSION/dist/quill.js" && \
	echo "   ✅ Quill.js updated to $$QUILL_VERSION"
	
	# Get latest HTMX version
	@echo "🔍 Checking HTMX..."
	@HTMX_VERSION=$$(curl -s "https://api.github.com/repos/bigskysoftware/htmx/releases/latest" | grep '"tag_name"' | sed 's/.*"tag_name": "v\([^"]*\)".*/\1/'); \
	echo "   Latest HTMX version: $$HTMX_VERSION"; \
	echo "   📥 Downloading HTMX $$HTMX_VERSION..."; \
	curl -s -o public/js/htmx.min.js "https://unpkg.com/htmx.org@$$HTMX_VERSION/dist/htmx.min.js" && \
	echo "   ✅ HTMX updated to $$HTMX_VERSION"
	
	# Get latest Pico.css version
	@echo "🔍 Checking Pico.css..."
	@PICO_VERSION=$$(curl -s "https://api.github.com/repos/picocss/pico/releases/latest" | grep '"tag_name"' | sed 's/.*"tag_name": "v\([^"]*\)".*/\1/'); \
	echo "   Latest Pico.css version: $$PICO_VERSION"; \
	echo "   📥 Downloading Pico.css $$PICO_VERSION..."; \
	curl -s -o public/css/pico.min.css "https://cdn.jsdelivr.net/npm/@picocss/pico@$$PICO_VERSION/css/pico.min.css" && \
	echo "   ✅ Pico.css updated to $$PICO_VERSION"
	
	@echo ""
	@echo "🎉 All frontend dependencies updated successfully!"
	@echo "📝 Updated files:"
	@echo "   - public/css/quill.snow.css"
	@echo "   - public/css/pico.min.css" 
	@echo "   - public/js/quill.min.js"
	@echo "   - public/js/htmx.min.js"
	@echo ""
	@echo "💡 Tip: Restart Buffalo dev server to see changes: make dev"

health:
	@bash scripts/health-check.sh
