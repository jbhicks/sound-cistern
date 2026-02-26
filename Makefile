.PHONY: help dev dev-bg dev-test dev-test-bg dev-test-stop dev-logs dev-stop dev-status setup build clean migrate test serve admin templ templ-watch

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
	@echo "  dev-test   - 🏃 Start dev server with TEST_MODE (no auth required)"
	@echo "  dev-test-bg - 🏃 Start dev-test server in background"
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
	@echo "  test-e2e          - 🚀 Run E2E tests (starts test server on port 8090)"
	@echo "  test-e2e-quick    - ⚡ Run E2E tests (skip server startup)"
	@echo "  test-server-start - 🚀 Start test server in TEST_MODE on port 8090"
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

# Development server with TEST_MODE enabled (for testing without auth)
dev-test:
	@echo "🚀 Starting PocketBase dev server with TEST_MODE..."
	@echo "📱 Admin UI: http://127.0.0.1:8090/_/"
	@echo "📱 Public API: http://127.0.0.1:8090/api/"
	@echo "🧪 TEST_MODE enabled - stream endpoints work without auth"
	@mkdir -p tmp
	@TEST_MODE=true air -c .air.toml

dev-test-bg: build
	@echo "🚀 Starting PocketBase dev-test server in background..."
	@mkdir -p tmp
	@if [ -f $(PID_FILE) ] && kill -0 $$(cat $(PID_FILE)) 2>/dev/null; then \
		echo "⚠️  Dev-test server already running (PID: $$(cat $(PID_FILE)))"; \
		echo "   Run 'make dev-test-stop' to restart"; \
		exit 1; \
	fi
	@TEST_MODE=true nohup air -c .air.toml > $(LOG_FILE) 2>&1 & echo $$! > $(PID_FILE)
	@echo "✅ Dev-test server started in background (PID: $$(cat $(PID_FILE)))"
	@echo "📱 Admin UI: http://127.0.0.1:8090/_/"
	@echo "🧪 TEST_MODE enabled"

dev-test-stop:
	@if [ -f $(PID_FILE) ]; then \
		PID=$$(cat $(PID_FILE)); \
		if kill -0 $$PID 2>/dev/null; then \
			echo "🛑 stopping dev-test server (PID: $$PID)..."; \
			kill $$PID && rm -f $(PID_FILE); \
			echo "✅ Dev-test server stopped"; \
		else \
			echo "⚠️  Dev-test server not running (stale PID file)"; \
			rm -f $(PID_FILE); \
		fi; \
	else \
		echo "ℹ️  No dev-test server running (no PID file)"; \
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
	@echo "🌐 Running browser automation tests (Chromedp) with mock data..."
	@TEST_MODE=true go test -tags=browser -v ./...
	@echo "✅ Browser tests complete!"

test-e2e: build
	@echo "🚀 Running E2E tests with test server..."
	@if [ -f tmp/dev-server.pid ] && kill -0 $$(cat tmp/dev-server.pid) 2>/dev/null; then \
		echo "⚠️  Stopping running dev server for E2E tests..."; \
		kill $$(cat tmp/dev-server.pid) 2>/dev/null || true; \
		rm -f tmp/dev-server.pid; \
		sleep 2; \
	fi
	@go test -tags=e2e -v -count=1 ./...
	@echo "✅ E2E tests complete!"

test-e2e-quick:
	@echo "🚀 Running E2E tests (skip server startup)..."
	@SKIP_SERVER_START=true go test -tags=e2e -v -count=1 ./...
	@echo "✅ E2E tests complete!"

test-server-start:
	@echo "🚀 Starting test server on port 8090..."
	@mkdir -p pb_data_test
	@TEST_MODE=true ./sound-cistern serve --dir=pb_data_test --http=0.0.0.0:8090

test-db-setup:
	@echo "📦 Setting up test database template..."
	@mkdir -p pb_data_test_template
	@echo "Run 'make dev' first, then visit http://localhost:8090/_/ to create collections and test data"
	@echo "After setup, copy pb_data to pb_data_test_template: cp -r pb_data/* pb_data_test_template/"
