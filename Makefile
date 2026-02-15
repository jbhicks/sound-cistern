.PHONY: help dev dev-bg dev-logs dev-stop dev-status setup build clean migrate test serve admin templ templ-watch

# Process management for background dev server
PID_FILE := tmp/dev-server.pid
LOG_FILE := tmp/dev-server.log

help:
	@echo "🚀 Sound Cistern - PocketBase Development Commands"
	@echo ""
	@echo "Quick Start:"
	@echo "  setup      - 🔧 Build application"
	@echo "  dev        - 🏃 Start PocketBase development server (foreground)"
	@echo "  dev-bg     - 🏃 Start dev server in background (for agents)"
	@echo "  serve      - 🌐 Start PocketBase production server"
	@echo "  migrate    - 🔀 Run database migrations"
	@echo ""
	@echo "Development:"
	@echo "  build      - 🔨 Build the application"
	@echo "  templ      - 🎨 Generate templ files"
	@echo "  templ-watch - 👀 Watch and regenerate templ files"
	@echo "  clean      - 🧹 Clean build artifacts"
	@echo "  admin      - 👑 Create admin user"
	@echo ""
	@echo "Dev Server Management:"
	@echo "  dev-logs   - 📋 View dev server logs (tail -f)"
	@echo "  dev-stop   - 🛑 Stop the background dev server"
	@echo "  dev-status - ℹ️  Check if dev server is running"
	@echo ""
	@echo "Testing:"
	@echo "  test              - 🧪 Run Go unit tests (mock functions)"
	@echo "  test-integration  - 🔗 Run Go integration tests (API mocks)"
	@echo "  test-all          - 🧪 Run all Go tests (unit + integration)"
	@echo "  test-browser      - 🌐 Run browser automation tests (Chromedp)"
	@echo ""

setup: build
	@echo "✅ Setup complete!"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Run 'make dev' to start the development server"
	@echo "  2. Visit http://127.0.0.1:8090/_/ to access the admin UI"
	@echo "  3. Create an admin account through the web interface"
	@echo ""

templ:
	@echo "🎨 Generating templ files..."
	@templ generate
	@echo "✅ Templ generation complete!"

templ-watch:
	@echo "👀 Watching templ files..."
	@templ generate --watch

build: templ
	@echo "🔨 Building PocketBase application..."
	@CGO_ENABLED=0 go build -o sound-cistern
	@echo "✅ Build complete!"

dev:
	@echo "🚀 Starting PocketBase development server with hot reload..."
	@echo "📱 Admin UI: http://127.0.0.1:8090/_/"
	@echo "📱 Public API: http://127.0.0.1:8090/api/"
	@echo "🔥 Hot reload enabled - edit .go or .templ files to trigger rebuild"
	@mkdir -p tmp
	@air -c .air.toml

# Background dev server (non-blocking for agents)
dev-bg: build
	@echo "🚀 Starting PocketBase dev server in background..."
	@mkdir -p tmp
	@if [ -f $(PID_FILE) ] && kill -0 $$(cat $(PID_FILE)) 2>/dev/null; then \
		echo "⚠️  Dev server already running (PID: $$(cat $(PID_FILE)))"; \
		echo "   Run 'make dev-logs' to see output or 'make dev-stop' to restart"; \
		exit 1; \
	fi
	@nohup air -c .air.toml > $(LOG_FILE) 2>&1 & echo $$! > $(PID_FILE)
	@echo "✅ Dev server started in background (PID: $$(cat $(PID_FILE)))"
	@echo "📱 Admin UI: http://127.0.0.1:8090/_/"
	@echo "📋 Logs: make dev-logs"
	@echo "🛑 Stop: make dev-stop"

dev-logs:
	@tail -f $(LOG_FILE)

dev-stop:
	@if [ -f $(PID_FILE) ]; then \
		PID=$$(cat $(PID_FILE)); \
		if kill -0 $$PID 2>/dev/null; then \
			echo "🛑 Stopping dev server (PID: $$PID)..."; \
			kill $$PID && rm -f $(PID_FILE); \
			echo "✅ Dev server stopped"; \
		else \
			echo "⚠️  Dev server not running (stale PID file)"; \
			rm -f $(PID_FILE); \
		fi; \
	else \
		echo "ℹ️  No dev server running (no PID file)"; \
	fi

dev-status:
	@if [ -f $(PID_FILE) ] && kill -0 $$(cat $(PID_FILE)) 2>/dev/null; then \
		echo "✅ Dev server running (PID: $$(cat $(PID_FILE)))"; \
		echo "📱 http://127.0.0.1:8090/_/"; \
	else \
		echo "❌ Dev server not running"; \
	fi

serve:
	@echo "🌐 Starting PocketBase production server..."
	@./sound-cistern serve

migrate:
	@echo "🔀 Running database migrations..."
	@./sound-cistern migrate collections
	@echo "✅ Migrations complete!"

admin:
	@echo "👑 Creating admin user..."
	@echo "Run the dev server and visit http://127.0.0.1:8090/_/ to create an admin"
	@echo "Or use: ./sound-cistern admin create <email> <password>"

clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -f sound-cistern
	@rm -rf pb_data
	@echo "✅ Clean complete!"

test:
	@echo "🧪 Running Go unit tests..."
	@go test -tags=unit ./...
	@echo "✅ Unit tests complete!"

test-integration:
	@echo "🔗 Running Go integration tests..."
	@go test -tags=integration ./...
	@echo "✅ Integration tests complete!"

test-all:
	@echo "🧪 Running all Go tests..."
	@go test -tags="unit integration" ./...
	@echo "✅ All tests complete!"

test-browser:
	@echo "🌐 Running browser automation tests (Chromedp)..."
	@go test -tags=browser ./...
	@echo "✅ Browser tests complete!"
