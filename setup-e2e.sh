#!/bin/bash
# E2E Test Setup and Runner Script

set -e

echo "🚀 Setting up E2E tests for Sound Cistern..."

# Check if Node.js is installed
if ! command -v node &> /dev/null; then
    echo "❌ Node.js is required for E2E tests. Please install Node.js first."
    exit 1
fi

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is required. Please install Go first."
    exit 1
fi

# Install Playwright dependencies
echo "📦 Installing Playwright dependencies..."
npm install

# Install Playwright browsers
echo "🌐 Installing Playwright browsers..."
npx playwright install

# Build the Go application
echo "🔨 Building Go application..."
go build -o sound-cistern .

echo "✅ E2E test setup complete!"
echo ""
echo "📋 Available test commands:"
echo "  npm test              - Run all E2E tests"
echo "  npm run test:headed   - Run tests with browser visible"
echo "  npm run test:debug    - Run tests in debug mode"
echo "  npm run test:ui       - Run tests with Playwright UI"
echo "  npm run test:report   - Show test report"
echo ""
echo "🎯 To run tests:"
echo "  npm test"