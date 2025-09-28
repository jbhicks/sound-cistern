#!/bin/bash
set -e

# Check Go
if ! command -v go >/dev/null 2>&1; then
  echo "❌ Go is not installed"; exit 1;
else
  echo "✅ Go: $(go version)";
fi

# Check Buffalo
if ! command -v buffalo >/dev/null 2>&1; then
  echo "❌ Buffalo CLI not found"; exit 1;
else
  echo "✅ Buffalo: $(buffalo version)";
fi

# Check Soda
if ! command -v soda >/dev/null 2>&1; then
  echo "❌ Soda not found"; exit 1;
else
  echo "✅ Soda: $(soda version)";
fi

# Check database connectivity
if pg_isready -h 127.0.0.1 -p 5432 -U postgres; then
  echo "✅ PostgreSQL is accepting connections on 5432";
else
  echo "❌ PostgreSQL is not accepting connections on 5432";
fi

# Check port 3001
if lsof -i :3001 | grep LISTEN; then
  echo "✅ Port 3001 is in use (Buffalo likely running)";
else
  echo "⚠️  Port 3001 is not in use";
fi

# Check Go bin in PATH
if [[ ":$PATH:" == *":/root/go/bin:"* ]]; then
  echo "✅ /root/go/bin is in PATH";
else
  echo "⚠️  /root/go/bin is NOT in PATH";
fi
