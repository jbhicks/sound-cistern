# Production Dockerfile for Sound Cistern with Go 1.21+ support
# Uses modern Go version to support all dependencies

FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata wget

# Install Buffalo CLI (compatible with Go 1.21+)
RUN go install github.com/gobuffalo/cli/cmd/buffalo@latest

# Set working directory
WORKDIR /app

# Copy go mod files for dependency caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download
RUN go mod verify

# Copy source code
COPY . .

# Build the application
RUN buffalo build --skip-template-validation -o /app/sound-cistern

# Production runtime image
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates bash tzdata wget
RUN addgroup -g 1000 appgroup && adduser -u 1000 -G appgroup -s /bin/sh -D appuser

# Create app directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/sound-cistern /app/sound-cistern
RUN chmod +x /app/sound-cistern

# Create non-root user for security
RUN chown -R appuser:appgroup /app
USER appuser

# Set container environment (disable SSL redirect for container networking)
ENV GO_ENV=development
ENV ADDR=0.0.0.0
ENV PORT=3000

# Expose port
EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:3000/health || exit 1

# Start application (migrations handled separately)
CMD /app/sound-cistern
