# Admin Template Checklist

When creating or modifying admin templates, use this checklist:

## ✅ Required Elements

### Navigation Header
- [ ] `<nav class="container-fluid dashboard-nav">` is present
- [ ] Site logo `<strong>My Go SaaS</strong>` is included
- [ ] Admin Dashboard link is available
- [ ] Theme toggle functionality is preserved

### Main Content Structure
- [ ] `<main class="container">` wraps the content
- [ ] Navigation comes before main content
- [ ] Proper HTML structure is maintained

### HTMX Compatibility
- [ ] Template works for direct page loads
- [ ] Template works when loaded via HTMX
- [ ] No nested main containers when loaded into `#htmx-content`

## 🧪 Testing

### Automated Validation
- [ ] Run `make validate-templates` 
- [ ] All template structure tests pass
- [ ] No validation warnings

### Manual Testing
- [ ] Direct page access works (if publicly accessible)
- [ ] HTMX navigation loads correctly
- [ ] Navigation elements are functional
- [ ] Theme switching works

### Browser Testing (Public Pages Only)
- [ ] **NEVER test protected admin pages in browser without login**
- [ ] Use `buffalo test` for protected page verification
- [ ] Only test public pages like `/auth/new`, `/users/new`

## 🔄 After Changes

1. **Validate Structure**: `make validate-templates`
2. **Run Tests**: `buffalo test`
3. **Check HTMX**: Verify partial loading works
4. **Document Changes**: Update any relevant documentation

## 🚨 Common Mistakes to Avoid

- **Missing navigation**: Always include the dashboard nav
- **Testing protected pages**: Use tests, not browser
- **Breaking HTMX**: Ensure templates work both ways
- **Inconsistent structure**: Follow existing patterns
- **Forgetting validation**: Always run checks before committing

## 📚 References

- `docs/admin-template-requirements.md` - Detailed requirements
- `docs/template-safeguards.md` - Prevention mechanisms
- `templates/admin/dashboard.plush.html` - Reference implementation
- `actions/template_validation_test.go` - Automated checks
