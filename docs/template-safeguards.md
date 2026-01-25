## Common Patterns

### HTMX Integration
```templ
templ HXButton(url string, target string) {
    <button 
        hx-get={ url }
        hx-target={ target }
        hx-swap="innerHTML"
    >
        Load Content
    </button>
}
```

### Form Components
```templ
templ SignInForm(action string) {
    <form method="POST" action={ templ.URL(action) }>
        <label for="email">Email:</label>
        <input type="email" id="email" name="email" required/>
        
        <label for="password">Password:</label>
        <input type="password" id="password" name="password" required/>
        
        <button type="submit">Sign In</button>
    </form>
}
```

### Navigation Component
```templ
templ Nav(currentUser *User) {
    <nav class="container-fluid">
        <ul>
            <li><a href="/">Home</a></li>
            if currentUser != nil {
                <li><a href="/dashboard">Dashboard</a></li>
                <li><a href="/logout">Logout</a></li>
            } else {
                <li><a href="/signin">Sign In</a></li>
                <li><a href="/signup">Sign Up</a></li>
            }
        </ul>
    </nav>
}
```

## Error Prevention

### 1. Type Safety
- Templ catches type errors at compile time
- No runtime template parsing errors
- Go compiler enforces correctness

### 2. Build Process
```bash
# Always regenerate templates after changes
make templ

# Build will fail if templates have errors
make build
```

### 3. Testing Templates
```go
// Test template rendering
func TestUserCard(t *testing.T) {
    user := User{Name: "John", Age: 30}
    
    var buf bytes.Buffer
    err := views.UserCard(user).Render(context.Background(), &buf)
    
    assert.NoError(t, err)
    assert.Contains(t, buf.String(), "John")
}
```

## Troubleshooting

### Template Won't Generate
```bash
# Clean and regenerate
rm views/*_templ.go
make templ
```

### Compile Errors After Template Changes
1. Run `make templ` to regenerate
2. Check for syntax errors in `.templ` files
3. Verify all referenced Go types exist
4. Check imports in template file

### Missing Imports
```templ
package views

import "strconv"

templ MyComponent(count int) {
    <p>Count: { strconv.Itoa(count) }</p>
}
```

## Resources

- **Templ Documentation**: https://templ.guide/
- **Templ GitHub**: https://github.com/a-h/templ
- **Pico CSS** (used for styling): https://picocss.com/
- **HTMX** (for dynamic interactions): https://htmx.org/

## Migration Notes

### From Plush/Buffalo Templates
- Replace `<%= %>` with `{ }` for expressions
- Move logic to Go code or templ conditionals
- Convert helpers to Go functions
- Use type-safe props instead of context maps

This type-safe template system prevents many common template errors at compile time, making development faster and more reliable.