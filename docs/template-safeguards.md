# Template Development Safeguards

This document outlines the safeguards in place to prevent template structure issues and how to maintain them.

## Problem Prevention

### What We're Preventing
- **Missing Navigation Headers**: Admin templates without proper navigation
- **Inconsistent Structure**: Admin pages that don't match the expected layout
- **Silent Failures**: Template issues that aren't caught until manual testing

### How We Prevent It

#### 1. Automated Template Validation
- **File**: `actions/template_validation_test.go`
- **Function**: `Test_AdminTemplateStructure`
- **What it checks**:
  - All admin templates have required navigation elements
  - Navigation comes before main content
  - Proper HTML structure is maintained

#### 2. Development Workflow Integration
- **Make Command**: `make validate-templates` 
- **Test Integration**: Runs automatically with `make test`
- **Quick Feedback**: Immediate validation during development

#### 3. Pre-commit Hook Support
- **Script**: `scripts/validate-templates.sh`
- **Optional Setup**: Can be installed as git pre-commit hook
- **Prevents Bad Commits**: Catches issues before they enter the repository

## Template Development Guidelines

### When Creating New Admin Templates

1. **Always start with navigation**:
   ```html
   <nav class="container-fluid dashboard-nav">
     <!-- Navigation content -->
   </nav>
   ```

2. **Include main container**:
   ```html
   <main class="container">
     <!-- Page content -->
   </main>
   ```

3. **Test structure immediately**:
   ```bash
   make validate-templates
   ```

### When Modifying Existing Templates

1. **Preserve navigation structure**
2. **Run validation after changes**
3. **Check both direct access and HTMX loading**

## Troubleshooting

### Template Validation Fails

If `make validate-templates` fails:

1. **Check the error message** - it will specify which template and what's missing
2. **Run with verbose output**: `buffalo test -v -run Test_AdminTemplateStructure`
3. **Compare with working templates** in `templates/admin/`
4. **Refer to documentation**: `docs/admin-template-requirements.md`

### Adding New Template Types

If you add a new category of admin templates:

1. **Update the test** in `template_validation_test.go` if needed
2. **Add exclude patterns** for partials or special cases
3. **Test the validation** to ensure new templates are covered

## Maintenance

### Regular Checks
- Template validation runs with every test
- Should pass before any commit
- Automated in CI/CD if configured

### Updating Safeguards
- Modify `template_validation_test.go` for new requirements
- Update `scripts/validate-templates.sh` for new checks
- Keep documentation current with actual implementation

## Integration Points

### With Buffalo Development
- Works with Buffalo's hot reload
- Compatible with existing test suite
- No impact on development server performance

### With Git Workflow
- Optional pre-commit hook available
- Can be integrated into CI/CD pipelines
- Provides clear error messages for fixes

### With VS Code
- Language server respects these validations
- Test failures show in Problems panel
- Quick fixes available through Go extension
