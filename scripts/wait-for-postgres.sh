#!/bin/bash

# Wait for PostgreSQL to be ready
echo "Checking PostgreSQL readiness..."

# Determine which compose command to use
if command -v podman-compose >/dev/null 2>&1; then
    COMPOSE_CMD="podman-compose"
elif command -v docker-compose >/dev/null 2>&1; then
    COMPOSE_CMD="docker-compose"
else
    echo "ERROR: Neither podman-compose nor docker-compose found"
    exit 1
fi

# Maximum wait time in seconds
MAX_WAIT=30
WAIT_COUNT=0

while [ $WAIT_COUNT -lt $MAX_WAIT ]; do
    # Check if container is running and healthy
    if $COMPOSE_CMD exec postgres pg_isready -U postgres > /dev/null 2>&1; then
        echo "PostgreSQL is ready!"
        exit 0
    fi
    
    echo "Waiting for PostgreSQL... ($((WAIT_COUNT + 1))/$MAX_WAIT)"
    sleep 1
    WAIT_COUNT=$((WAIT_COUNT + 1))
done

echo "ERROR: PostgreSQL failed to become ready within $MAX_WAIT seconds"
echo "Container status:"
$COMPOSE_CMD ps
echo "Container logs:"
$COMPOSE_CMD logs postgres --tail 20
exit 1
