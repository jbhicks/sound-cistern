#!/bin/bash
# Script to clear Go build and gopls caches to resolve stale errors

echo "🧹 Clearing Go build cache..."
go clean -cache

echo "🧹 Clearing module cache..."
go clean -modcache

if command -v gopls >/dev/null 2>&1; then
  echo "🧹 Clearing gopls cache..."
  gopls clean
else
  echo "⚠️ gopls not found, skipping gopls cache clean"
fi

echo "✅ Caches cleared. Please restart your editor or reload the Go language server."
