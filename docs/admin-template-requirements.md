# Admin Template Requirements

## Navigation Header Requirement

**ALL admin templates MUST include a navigation header** to maintain consistent user experience and site navigation.

### Required Navigation Structure

Every admin template (`templates/admin/**/*.plush.html`) must include:

```html
<nav class="container-fluid dashboard-nav">
    <ul>
        <li><strong><a href="/" style="text-decoration: none; color: inherit;">My Go SaaS</a></strong></li>
    </ul>
    <ul>
        <li><a href="/admin/dashboard">Admin Dashboard</a></li>
        <li>
            <details class="dropdown">
                <summary role="button" class="secondary">☰</summary>
                <ul dir="rtl">
                    <li><a href="/admin/users">Users</a></li>
                    <li><a href="/admin/posts">Blog Posts</a></li>
                    <li><a href="/logout">Logout</a></li>
                </ul>
            </details>
        </li>
        <li>
            <details class="dropdown">
                <summary role="button" class="secondary" id="theme-toggle">🌙</summary>
                <ul dir="rtl">
                    <li><a href="#" onclick="setTheme('light')">☀️ Light</a></li>
                    <li><a href="#" onclick="setTheme('dark')">🌙 Dark</a></li>
                    <li><a href="#" onclick="setTheme('auto')">🔄 Auto</a></li>
                </ul>
            </details>
        </li>
    </ul>
</nav>

<main class="container">
    <!-- Template content goes here -->
</main>
```

### Why This Is Required

- Admin templates use `r.HTML()` rendering with `application.plush.html` layout
- The basic layout only provides HTML structure, NOT navigation
- Each admin template must include its own navigation to maintain site consistency
- Without navigation, users get "headerless" pages that break the user experience

### Template Structure Pattern

1. **Navigation Header**: Always include the full navigation structure
2. **Main Container**: Wrap content in `<main class="container">`
3. **Consistent Styling**: Use the same CSS classes and structure as other admin templates

### Examples

✅ **Correct Structure** (see `templates/admin/dashboard.plush.html`):
- Includes full navigation header
- Has main container wrapper
- Provides complete user experience

❌ **Incorrect Structure**:
- Missing navigation header
- Content directly in template without navigation
- Results in "headerless" pages

### Testing

The `Test_AdminPostPagesHaveNavigation()` test in `actions/blog_test.go` validates that all admin post pages include proper navigation headers.

**When creating new admin templates, ensure they follow this pattern to avoid navigation issues.**

## Safeguards and Prevention

### Automated Template Validation

This template includes safeguards to prevent navigation headers from being missing:

#### 1. Template Structure Test
- **Test**: `Test_AdminTemplateStructure` in `actions/template_validation_test.go`
- **Validates**: All admin templates include required navigation elements
- **Required Elements**:
  - `<nav class="container-fluid dashboard-nav">`
  - `<strong>My Go SaaS</strong>` (site logo)
  - `<main class="container">` (content wrapper)
- **Coverage**: Automatically discovers and validates all admin templates

#### 2. Make Command Integration
- **Command**: `make validate-templates`
- **Purpose**: Quick validation of template structure
- **Integration**: Runs automatically as part of `make test`

#### 3. Validation Script
- **Script**: `scripts/validate-templates.sh`
- **Usage**: Can be used as a pre-commit hook
- **Features**: 
  - Validates template structure
  - Warns about CSRF token issues
  - Provides clear error messages with guidance

### Setting Up Pre-commit Hook (Optional)

To automatically validate templates before each commit:

```bash
# Create git hooks directory if it doesn't exist
mkdir -p .git/hooks

# Create pre-commit hook
cat > .git/hooks/pre-commit << 'EOF'
#!/bin/bash
exec ./scripts/validate-templates.sh
EOF

# Make it executable
chmod +x .git/hooks/pre-commit
```

### Running Validation

```bash
# Validate templates manually
make validate-templates

# Run full test suite (includes template validation)
make test

# Run validation script directly
./scripts/validate-templates.sh
```
