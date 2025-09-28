# Pico.css Implementation Guide

> **For Buffalo SaaS Template Development**

This guide outlines how to effectively use Pico.css in the Buffalo SaaS template, focusing on semantic HTML and minimal CSS classes.

## Core Philosophy

Pico.css follows a **semantic-first** approach:
1. Use proper HTML elements for their intended purpose
2. Minimal CSS classes - let Pico.css handle the styling automatically
3. Customize with CSS variables, not utility classes
4. Embrace automatic theming with dark/light modes

## Common Patterns

### Navigation & Headers

```html
<!-- Use semantic nav element -->
<nav>
  <ul>
    <li><strong>Brand Name</strong></li>
  </ul>
  <ul>
    <li><a href="/dashboard">Dashboard</a></li>
    <li><a href="/account">Account</a></li>
  </ul>
</nav>
```

### Buttons & CTAs

```html
<!-- Primary action -->
<a href="/signup" role="button">Sign Up</a>

<!-- Secondary action -->
<a href="/login" role="button" class="secondary">Login</a>

<!-- Outlined button -->
<a href="/learn-more" role="button" class="outline">Learn More</a>

<!-- Contrast button (adapts to theme) -->
<a href="/dashboard" role="button" class="contrast">Dashboard</a>
```

### Dropdowns & Menus

```html
<!-- Use details/summary for dropdowns -->
<details class="dropdown">
  <summary role="button">Profile</summary>
  <ul>
    <li><a href="/profile">Your Profile</a></li>
    <li><a href="/settings">Settings</a></li>
    <li><a href="/logout">Sign out</a></li>
  </ul>
</details>
```

### Forms

```html
<!-- Pico.css automatically styles form elements -->
<form>
  <label for="email">Email</label>
  <input type="email" id="email" name="email" required>
  
  <label for="password">Password</label>
  <input type="password" id="password" name="password" required>
  
  <button type="submit">Sign In</button>
</form>
```

### Grid Layouts

```html
<!-- Simple responsive grid -->
<div class="grid">
  <div>Column 1</div>
  <div>Column 2</div>
  <div>Column 3</div>
</div>
```

### Cards & Containers

```html
<!-- Use article for content cards -->
<article>
  <header>
    <h2>Card Title</h2>
  </header>
  <p>Card content goes here...</p>
  <footer>
    <a href="#" role="button">Action</a>
  </footer>
</article>
```

## Theme Customization

### Custom Colors

```css
:root {
  --pico-primary: #0066cc;
  --pico-secondary: #6c757d;
}

/* Dark mode overrides */
[data-theme="dark"] {
  --pico-primary: #4da6ff;
}
```

### Typography

```css
:root {
  --pico-font-family: 'Inter', system-ui, sans-serif;
  --pico-font-size: 1rem;
  --pico-line-height: 1.6;
}
```

### Spacing

```css
:root {
  --pico-spacing: 1rem;
  --pico-typography-spacing-vertical: 1.5rem;
}
```

## Theme Switching Implementation

Our template uses localStorage-based theme switching:

```javascript
function setTheme(theme) {
  const themeIcon = document.getElementById('theme-icon');
  
  if (theme === 'auto') {
    localStorage.removeItem('picoPreferredColorScheme');
    document.documentElement.removeAttribute('data-theme');
    themeIcon.textContent = '🔄';
  } else {
    localStorage.setItem('picoPreferredColorScheme', theme);
    document.documentElement.setAttribute('data-theme', theme);
    themeIcon.textContent = theme === 'dark' ? '🌙' : '☀️';
  }
}

// Initialize on page load
document.addEventListener('DOMContentLoaded', function() {
  const savedTheme = localStorage.getItem('picoPreferredColorScheme');
  if (savedTheme) {
    document.documentElement.setAttribute('data-theme', savedTheme);
  }
});
```

## Anti-Patterns to Avoid

### ❌ Don't Use Utility Classes

```html
<!-- Avoid this Tailwind-style approach -->
<div class="bg-blue-500 text-white p-4 rounded-lg">
  Content
</div>
```

### ✅ Use Semantic HTML + CSS Variables

```html
<!-- Do this instead -->
<article style="background: var(--pico-primary); color: var(--pico-primary-inverse);">
  Content
</article>
```

### ❌ Don't Fight the Framework

```html
<!-- Don't override with inline styles excessively -->
<button style="background: red; border: 1px solid blue; padding: 20px;">
  Bad Button
</button>
```

### ✅ Use Classes and CSS Variables

```html
<!-- Work with Pico.css -->
<button class="secondary">Good Button</button>
```

## Responsive Design

Pico.css handles responsive design automatically:

- Typography scales appropriately
- Grid layouts adapt to screen size
- Form elements remain touch-friendly
- Navigation collapses on mobile

### Custom Breakpoints

```css
/* Mobile first approach */
.custom-layout {
  display: block;
}

@media (min-width: 768px) {
  .custom-layout {
    display: flex;
  }
}
```

## Accessibility

Pico.css includes built-in accessibility features:

- Proper focus states
- High contrast ratios
- Screen reader friendly markup
- Keyboard navigation support

### Additional Accessibility

```html
<!-- Use proper ARIA labels -->
<button aria-label="Close dialog" class="close">×</button>

<!-- Semantic headings -->
<h1>Main Title</h1>
<h2>Section Title</h2>
<h3>Subsection</h3>
```

## Performance

- **Minimal CSS**: Pico.css is lightweight (~10KB gzipped)
- **No JavaScript required**: Pure CSS framework
- **Tree-shakeable**: Use only what you need
- **CDN available**: Fast loading from jsDelivr

## Common Customizations for SaaS

### Brand Colors

```css
:root {
  --pico-primary: #2563eb;        /* Brand blue */
  --pico-secondary: #64748b;      /* Neutral gray */
}

[data-theme="dark"] {
  --pico-primary: #3b82f6;        /* Lighter blue for dark mode */
}
```

### Professional Typography

```css
:root {
  --pico-font-family: 'Inter', -apple-system, BlinkMacSystemFont, sans-serif;
  --pico-font-size: 0.95rem;
  --pico-line-height: 1.5;
}
```

### Subtle Shadows

```css
:root {
  --pico-card-box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  --pico-dropdown-box-shadow: 0 8px 32px rgba(0, 0, 0, 0.12);
}
```

## Integration with Buffalo Templates

### Plush Template Example

```html
<!-- Authentication-aware navigation -->
<nav>
  <ul>
    <li><strong>SaaS App</strong></li>
  </ul>
  <ul>
    <% if (current_user) { %>
      <li><a href="/dashboard">Dashboard</a></li>
      <li>
        <details class="dropdown">
          <summary role="button">Profile</summary>
          <ul>
            <li><a href="/profile">Profile</a></li>
            <li><a href="/signout">Sign Out</a></li>
          </ul>
        </details>
      </li>
    <% } else { %>
      <li><a href="/login" role="button">Login</a></li>
      <li><a href="/signup" role="button" class="contrast">Sign Up</a></li>
    <% } %>
  </ul>
</nav>
```

## Resources

- [Pico.css Official Documentation](https://picocss.com/docs)
- [CSS Variables Reference](./pico-css-variables.md)
- [Semantic HTML Guide](https://developer.mozilla.org/en-US/docs/Web/HTML/Element)
