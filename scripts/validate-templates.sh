#!/bin/bash

# Pre-commit hook to validate admin templates
# This script runs before each commit to ensure admin templates have proper structure

echo "🔍 Validating admin template structure..."

# Run the template validation test
cd "$(git rev-parse --show-toplevel)"

# Check if we're in a Buffalo project
if [ ! -f "app.go" ] && [ ! -f "main.go" ] && [ ! -d "actions" ]; then
    echo "⚠️  Not in a Buffalo project root, skipping template validation"
    exit 0
fi

# Run template validation test
if buffalo test -run "Test_AdminTemplateStructure" >/dev/null 2>&1; then
    echo "✅ Admin template validation passed"
else
    echo "❌ Admin template validation failed!"
    echo ""
    echo "Some admin templates are missing required navigation headers."
    echo "Please ensure all admin templates include:"
    echo "  - <nav class=\"container-fluid dashboard-nav\">"
    echo "  - <strong>My Go SaaS</strong>"
    echo "  - <main class=\"container\">"
    echo ""
    echo "Run 'buffalo test -run Test_AdminTemplateStructure -v' for details."
    echo ""
    echo "See docs/admin-template-requirements.md for guidance."
    exit 1
fi

echo "✅ Admin template validation passed"

# Check for any remaining CSRF form issues in admin templates
if grep -r "authenticity_token" templates/admin/ 2>/dev/null | grep -v "<%#" | grep -q "authenticity_token"; then
    echo "⚠️  Warning: Found uncommented authenticity_token usage in admin templates"
    echo "    These may need to be properly implemented or commented out for testing"
fi

exit 0
