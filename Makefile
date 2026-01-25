.PHONY: help dev setup build clean migrate test serve admin templ templ-watch

help:
	@echo "🚀 Sound Cistern - PocketBase Development Commands"
	@echo ""
	@echo "Quick Start:"
	@echo "  setup      - 🔧 Build application"
	@echo "  dev        - 🏃 Start PocketBase development server"
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
	@echo "🧪 Running tests..."
	@go test ./...
