# Buffalo Plush Template Syntax

Based on the official Plush documentation: https://github.com/gobuffalo/plush

## String Manipulation

Plush provides built-in helpers for string operations:

- `capitalize(string)` - capitalizes the first letter
- `len(string)` - gets the length of a string
- String slicing should use helper functions, not Go syntax

## Template Syntax

- Use `<%= %>` for output
- Use `<% %>` for execution without output
- Conditionals: `<%= if (condition) { %> ... <% } %>`
- String comparisons: `string != ""`

## Common Patterns

For getting the first character of a string, use helper functions or create custom helpers.
Avoid Go-style syntax like `string[0:1]` as Plush doesn't support this directly.

## Buffalo Context Variables

- `current_user` - contains the authenticated user object
- Standard string operations should use Plush helpers