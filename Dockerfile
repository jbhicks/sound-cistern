# Production Dockerfile for Sound Cistern (PocketBase + Templ)
# Single static binary with embedded SQLite database

FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Install templ CLI for template generation
RUN go install github.com/a-h/templ/cmd/templ@latest

# Set working directory
WORKDIR /app

# Copy go mod files for dependency caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download
RUN go mod verify

# Copy source code
COPY . .

# Generate templ files and build
RUN templ generate
RUN CGO_ENABLED=0 go build -o /app/sound-cistern

# Production runtime image
FROM alpine:latest

# Install runtime dependencies (curl for health checks)
RUN apk --no-cache add ca-certificates bash tzdata curl
RUN addgroup -g 1000 appgroup && adduser -u 1000 -G appgroup -s /bin/sh -D appuser

# Create app directory and data directory
WORKDIR /app
RUN mkdir -p /app/pb_data /app/pb_public && chown -R appuser:appgroup /app

# Copy binary from builder stage
COPY --from=builder /app/sound-cistern /app/sound-cistern
COPY --from=builder /app/public /app/public
RUN chmod +x /app/sound-cistern

# Switch to non-root user
USER appuser

# Set environment variables
ENV PORT=8090

# Expose port
EXPOSE 8090

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=20s --retries=3 \
  CMD curl -f http://localhost:8090/health || exit 1

# Start PocketBase
CMD ["/app/sound-cistern", "serve", "--http=0.0.0.0:8090"]
