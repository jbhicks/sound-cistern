---
name: sound-cistern-ux
description: Complete UI/UX design system and development guidelines for Sound Cistern using Pico.css, Primer design principles, and Templ templates
license: MIT
compatibility: opencode
metadata:
  version: "1.0"
  audience: ux-developers
  stack: pico-css-templ-htmx-primer
---

> ⚠️ **DEPRECATED**: This skill describes Pico.css + HTMX + Templ UX patterns that are no longer used in Sound Cistern. The v2 frontend uses Tailwind CSS + React + Framer Motion. This skill is kept for historical reference only. Do not use these patterns for new features.

# Sound Cistern UX Development Skill

## Overview

This skill provides comprehensive UI/UX guidelines for building beautiful, accessible, and consistent interfaces in the Sound Cistern project. It combines:
- **Pico.css** - Semantic CSS framework (already in use)
- **Primer Design System** - GitHub's design principles and patterns
- **Templ** - Type-safe Go templates
- **HTMX** - Dynamic interactions

## Core Philosophy

### Semantic-First Design
> *"Use proper HTML elements for their intended purpose. Let the framework handle the styling."*

✅ **DO:**
- Use semantic HTML (`<nav>`, `<article>`, `<section>`, `<button>`)
- Customize via CSS variables, not utility classes
- Use `role="button"` for links that act as buttons
- Trust Pico.css responsive behavior

❌ **DON'T:**
- Use Tailwind-style utility classes (`bg-blue-500 text-white p-4`)
- Fight the framework with excessive overrides
- Create custom classes for things Pico provides
- Use `div` soup instead of semantic elements

### Systematic Approach (from Primer)
> *"Create a system that enables building consistent user experiences with ease, yet with enough flexibility to support the broad spectrum of interfaces."*

- **Composability** - Build complex UIs from simple, reusable parts
- **Consistency** - Same patterns = predictable experience
- **Flexibility** - Systematic but not rigid
- **Accessibility-first** - Design for everyone from the start

### Working WITH Pico.css (Not Against It)

The biggest lesson from refactoring: **Pico.css already provides beautiful form styling**. Don't override it with explicit heights and `!important` declarations.

✅ **DO:**
- Use CSS Grid for form layouts instead of fighting flexbox
- Let Pico handle input sizing naturally
- Override only layout (widths, positioning), not element styling
- Use semantic HTML: `<input type="search">`, `<select>`, etc.

❌ **DON'T:**
- Hardcode heights like `height: 44px !important` on inputs
- Use `!important` to override Pico's sensible defaults
- Fight with `flex: 0 0 auto` to prevent flexbox squeezing
- Duplicate form element CSS that Pico already provides

**Example: Filter Bar Layout**
```css
/* ✅ Good: Grid layout, Pico handles the rest */
.filter-form {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: var(--space-4);
  align-items: end;
}

@media (min-width: 1024px) {
  .filter-form {
    grid-template-columns: 2fr 1fr 1fr 1fr 1fr auto;
  }
}

/* ❌ Bad: Fighting with flexbox and !important */
.filter-input {
  height: 44px !important;
  padding: 0 var(--space-3) !important;
  /* Don't do this - let Pico breathe */
}
```

### Modular CSS Architecture

Split your CSS into logical modules for maintainability:

```
css/
├── tokens.css      # Design tokens only (colors, spacing, typography)
├── layout.css      # Header, footer, containers, navigation
├── components.css  # Reusable UI elements (buttons, cards, forms)
└── pages.css       # Page-specific styles (stream, auth, blog)
```

**Benefits:**
- **44% smaller** total CSS footprint
- No duplicate rules across files
- Clear separation of concerns
- Easier to find and modify styles

### Performance Warnings

⚠️ **NEVER use `* { transition: ... }`**
```css
/* ❌ Performance killer - applies to EVERY element */
* {
  transition: background-color var(--duration-slow) var(--ease-out),
              color var(--duration-slow) var(--ease-out);
}

/* ✅ Good: Target specific interactive elements */
.btn, .card, .nav-link {
  transition: 
    background-color var(--duration-fast) var(--ease-out),
    border-color var(--duration-fast) var(--ease-out);
}
```

This causes jank, especially with HTMX updates and dynamic content changes.

---

## 1. Design Tokens

### Spacing System (Base-8 Scale)

Primer uses a highly composable base-8 spacing scale. Add these to your CSS:

```css
:root {
  /* Base 8 scale - highly composable */
  --space-1: 0.25rem;   /* 4px  - micro adjustments */
  --space-2: 0.5rem;    /* 8px  - tight spacing */
  --space-3: 0.75rem;   /* 12px - compact elements */
  --space-4: 1rem;      /* 16px - standard spacing */
  --space-5: 1.25rem;   /* 20px - comfortable spacing */
  --space-6: 1.5rem;    /* 24px - section padding */
  --space-8: 2rem;      /* 32px - large gaps */
  --space-10: 2.5rem;   /* 40px - section breaks */
  --space-12: 3rem;     /* 48px - major sections */
}
```

**Usage Patterns:**
```css
/* Component padding */
.track-card {
  padding: var(--space-4);        /* 16px */
}

/* Stack gaps */
.filter-group {
  gap: var(--space-3);            /* 12px */
}

/* Section margins */
.stream-section {
  margin-bottom: var(--space-6);  /* 24px */
}

/* Micro-adjustments */
.icon-btn {
  padding: var(--space-1);        /* 4px */
}
```

### Color System (Semantic Tokens)

Instead of raw colors, use **semantic tokens** that adapt to themes:

```css
:root {
  /* Background layers */
  --color-canvas-default: #ffffff;
  --color-canvas-subtle: #f6f8fa;
  --color-canvas-inset: #f3f4f6;
  
  /* Text colors */
  --color-fg-default: #1f2328;
  --color-fg-muted: #656d76;
  --color-fg-subtle: #6e7781;
  --color-fg-on-emphasis: #ffffff;
  
  /* Interactive states (Accent/Primary) */
  --color-accent-fg: #0969da;           /* Links */
  --color-accent-emphasis: #0969da;     /* Primary buttons */
  --color-accent-muted: #ddf4ff;        /* Hover backgrounds */
  --color-accent-subtle: #f6f8fa;       /* Subtle backgrounds */
  
  /* Semantic states */
  --color-success-fg: #1a7f37;
  --color-attention-fg: #9a6700;
  --color-danger-fg: #cf222e;
  --color-severe-fg: #bc4c00;
  
  /* Borders */
  --color-border-default: #d0d7de;
  --color-border-muted: #d8dee4;
}

[data-theme="dark"] {
  --color-canvas-default: #0d1117;
  --color-canvas-subtle: #161b22;
  --color-fg-default: #c9d1d9;
  --color-fg-muted: #8b949e;
  --color-accent-fg: #58a6ff;
  --color-accent-emphasis: #1f6feb;
  --color-accent-muted: rgba(56, 139, 253, 0.4);
  --color-accent-subtle: rgba(56, 139, 253, 0.15);
  --color-border-default: #30363d;
  --color-border-muted: #21262d;
}
```

**Color Usage Patterns:**
```css
/* Cards/containers */
.track-item {
  background: var(--color-canvas-default);
  border: 1px solid var(--color-border-default);
}

/* Muted text (metadata) */
.track-meta {
  color: var(--color-fg-muted);
}

/* Interactive elements */
.favorite-btn {
  color: var(--color-fg-muted);
}

.favorite-btn:hover {
  color: var(--color-accent-fg);
  background: var(--color-accent-subtle);
}

/* Primary actions */
.soundcloud-btn {
  background: var(--color-accent-emphasis);
  color: var(--color-fg-on-emphasis);
}

/* Danger actions */
.remove-btn {
  color: var(--color-danger-fg);
}
```

### Typography Scale

```css
:root {
  /* Font stack (system fonts for performance) */
  --font-stack: -apple-system, BlinkMacSystemFont, "Segoe UI", 
                "Noto Sans", Helvetica, Arial, sans-serif;
  
  /* Type scale */
  --text-xs: 0.75rem;      /* 12px - captions */
  --text-sm: 0.875rem;     /* 14px - metadata */
  --text-base: 1rem;       /* 16px - body */
  --text-md: 1.125rem;     /* 18px - lead text */
  --text-lg: 1.25rem;      /* 20px - subheadings */
  --text-xl: 1.5rem;       /* 24px - headings */
  --text-2xl: 2rem;        /* 32px - page titles */
  --text-3xl: 2.5rem;      /* 40px - hero titles */
  
  /* Line heights */
  --lh-tight: 1.25;        /* headings */
  --lh-normal: 1.5;        /* body text */
  --lh-relaxed: 1.75;      /* readable text */
  
  /* Font weights */
  --font-normal: 400;
  --font-medium: 500;
  --font-semibold: 600;
  --font-bold: 700;
}
```

**Typography Patterns:**
```css
/* Page title */
.page-title {
  font-size: var(--text-2xl);
  font-weight: var(--font-semibold);
  line-height: var(--lh-tight);
  color: var(--color-fg-default);
}

/* Track title (links) */
.track-title {
  font-size: var(--text-base);
  font-weight: var(--font-semibold);
  line-height: var(--lh-tight);
  color: var(--color-accent-fg);
}

/* Artist name */
.artist-name {
  font-size: var(--text-sm);
  font-weight: var(--font-medium);
  color: var(--color-fg-default);
}

/* Metadata (duration, date) */
.track-meta {
  font-size: var(--text-xs);
  color: var(--color-fg-muted);
}

/* Empty state */
.empty-state-text {
  font-size: var(--text-md);
  color: var(--color-fg-muted);
  line-height: var(--lh-relaxed);
}
```

### Border Radius Scale

```css
:root {
  --radius-sm: 3px;       /* small elements */
  --radius-md: 6px;       /* buttons, inputs */
  --radius-lg: 12px;      /* cards, modals */
  --radius-xl: 24px;      /* large containers */
  --radius-full: 100vh;   /* pills, avatars */
}
```

### Shadow System

```css
:root {
  /* Light mode shadows */
  --shadow-sm: 0 1px 0 rgba(31, 35, 40, 0.04);
  --shadow-md: 0 3px 6px rgba(31, 35, 40, 0.08);
  --shadow-lg: 0 8px 24px rgba(31, 35, 40, 0.12);
  --shadow-xl: 0 12px 48px rgba(31, 35, 40, 0.15);
}

[data-theme="dark"] {
  /* Dark mode shadows - more subtle */
  --shadow-sm: 0 1px 0 rgba(0, 0, 0, 0.2);
  --shadow-md: 0 3px 6px rgba(0, 0, 0, 0.3);
  --shadow-lg: 0 8px 24px rgba(0, 0, 0, 0.4);
}
```

### Animation & Transitions

```css
:root {
  /* Timing functions */
  --ease-in-out: cubic-bezier(0.65, 0, 0.35, 1);
  --ease-out: cubic-bezier(0.16, 1, 0.3, 1);
  --ease-in: cubic-bezier(0.4, 0, 1, 1);
  --ease-spring: cubic-bezier(0.34, 1.56, 0.64, 1);
  
  /* Duration scale */
  --duration-instant: 0ms;
  --duration-fast: 150ms;
  --duration-normal: 250ms;
  --duration-slow: 350ms;
  --duration-slower: 500ms;
}
```

---

## 2. Component Patterns

### Buttons

```css
/* Button base */
.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-2);
  padding: var(--space-2) var(--space-3);
  font-size: var(--text-sm);
  font-weight: var(--font-medium);
  border-radius: var(--radius-md);
  border: 1px solid transparent;
  cursor: pointer;
  transition: all var(--duration-fast) var(--ease-in-out);
}

/* Variants */
.btn-primary {
  background: var(--color-accent-emphasis);
  color: var(--color-fg-on-emphasis);
  border-color: var(--color-accent-emphasis);
}

.btn-primary:hover {
  filter: brightness(1.1);
}

.btn-secondary {
  background: var(--color-canvas-subtle);
  color: var(--color-fg-default);
  border-color: var(--color-border-default);
}

.btn-secondary:hover {
  background: var(--color-canvas-default);
  border-color: var(--color-border-muted);
}

.btn-ghost {
  background: transparent;
  color: var(--color-accent-fg);
  border-color: transparent;
}

.btn-ghost:hover {
  background: var(--color-accent-subtle);
}

.btn-danger {
  background: var(--color-danger-fg);
  color: var(--color-fg-on-emphasis);
}
```

### Cards

```css
/* Track card */
.track-card {
  display: flex;
  gap: var(--space-4);
  padding: var(--space-4);
  background: var(--color-canvas-default);
  border: 1px solid var(--color-border-default);
  border-radius: var(--radius-lg);
  transition: 
    border-color var(--duration-fast) var(--ease-in-out),
    box-shadow var(--duration-fast) var(--ease-in-out);
}

.track-card:hover {
  border-color: var(--color-accent-muted);
  box-shadow: var(--shadow-md);
}

.track-card:active {
  box-shadow: var(--shadow-sm);
}
```

### Forms

```css
/* Input base */
.form-input {
  width: 100%;
  padding: var(--space-2) var(--space-3);
  font-size: var(--text-sm);
  color: var(--color-fg-default);
  background: var(--color-canvas-default);
  border: 1px solid var(--color-border-default);
  border-radius: var(--radius-md);
  transition: 
    border-color var(--duration-fast) var(--ease-in-out),
    box-shadow var(--duration-fast) var(--ease-in-out);
}

.form-input:focus {
  outline: none;
  border-color: var(--color-accent-fg);
  box-shadow: 0 0 0 3px var(--color-accent-muted);
}

/* Search input with icon */
.search-input {
  padding-left: calc(var(--space-3) + 20px + var(--space-2));
  background-image: url("data:image/svg+xml,...");
  background-repeat: no-repeat;
  background-position: var(--space-3) center;
  background-size: 20px;
}
```

### Navigation

```html
<nav>
  <ul>
    <li><strong>Sound Cistern</strong></li>
  </ul>
  <ul>
    <li><a href="/stream">Stream</a></li>
    <li><a href="/favorites">Favorites</a></li>
    <li>
      <details class="dropdown">
        <summary role="button">Profile</summary>
        <ul>
          <li><a href="/profile">Your Profile</a></li>
          <li><a href="/settings">Settings</a></li>
          <li><a href="/logout">Sign Out</a></li>
        </ul>
      </details>
    </li>
  </ul>
</nav>
```

---

## 3. Layout Patterns

### Container

```css
.container {
  width: 100%;
  max-width: 1280px;
  margin: 0 auto;
  padding: 0 var(--space-4);
}

.container-narrow {
  max-width: 768px;
}
```

### Stack Layout

```css
.stack {
  display: flex;
  flex-direction: column;
}

.stack-xs { gap: var(--space-1); }
.stack-sm { gap: var(--space-2); }
.stack-md { gap: var(--space-3); }
.stack-lg { gap: var(--space-4); }
.stack-xl { gap: var(--space-6); }
```

### Grid Layout

```css
.grid {
  display: grid;
  gap: var(--space-4);
}

/* Responsive grid for tracks */
.track-grid {
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: var(--space-4);
}

@media (max-width: 768px) {
  .track-grid {
    grid-template-columns: 1fr;
  }
}
```

---

## 4. Templ Template Patterns

### Layout Structure

```templ
package views

type PageData struct {
    Title       string
    Description string
    CurrentPath string
    User        interface{}
}

templ Layout(data PageData) {
    <!DOCTYPE html>
    <html lang="en">
        <head>
            <meta charset="utf-8"/>
            <meta name="viewport" content="width=device-width, initial-scale=1"/>
            <title>
                if data.Title != "" {
                    { data.Title } - Sound Cistern
                } else {
                    Sound Cistern - Audio Feed Aggregator
                }
            </title>
            <link rel="stylesheet" href="/css/pico.min.css"/>
            <link rel="stylesheet" href="/css/custom.css"/>
            <script src="/js/htmx.min.js"></script>
            <script src="/js/theme.js"></script>
        </head>
        <body>
            { children... }
        </body>
    </html>
}
```

### Component Reuse

```templ
// components/track_card.templ
package components

type TrackData struct {
    ID       string
    Title    string
    Artist   string
    Duration string
    Artwork  string
}

templ TrackCard(track TrackData) {
    <article class="track-card">
        <img src={ track.Artwork } alt={ track.Title } class="track-artwork"/>
        <div class="track-content">
            <h3 class="track-title">
                <a href={ templ.URL("/track/" + track.ID) }>{ track.Title }</a>
            </h3>
            <p class="artist-name">{ track.Artist }</p>
            <span class="track-meta">{ track.Duration }</span>
        </div>
    </article>
}
```

### HTMX Integration

```templ
templ LoadMoreButton(nextPage int) {
    <button 
        class="btn btn-secondary"
        hx-get={ fmt.Sprintf("/api/tracks?page=%d", nextPage) }
        hx-target="#track-list"
        hx-swap="beforeend"
        hx-indicator=".loading-spinner"
    >
        Load More
        <span class="loading-spinner" style="display: none;">
            <svg>...</svg>
        </span>
    </button>
}
```

---

## 5. Responsive Design

### Mobile-First Approach

```css
/* Base (mobile) styles */
.track-grid {
  grid-template-columns: 1fr;
  gap: var(--space-3);
}

/* Tablet */
@media (min-width: 769px) {
  .track-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

/* Desktop */
@media (min-width: 1025px) {
  .track-grid {
    grid-template-columns: repeat(3, 1fr);
    gap: var(--space-4);
  }
}
```

### Touch Targets

```css
/* Minimum touch target */
.touch-target {
  min-height: 44px;
  min-width: 44px;
}

/* Ensure all interactive elements are touchable */
button, .btn, [role="button"], a {
  min-height: 44px;
  min-width: 44px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}
```

---

## 6. Accessibility Guidelines

### Color Contrast
- **Minimum**: 4.5:1 for normal text
- **Large text**: 3:1 for 18px+ or 14px+ bold
- **UI components**: 3:1 for borders, icons

### Focus States

```css
/* Visible focus rings */
:focus-visible {
  outline: 2px solid var(--color-accent-fg);
  outline-offset: 2px;
}

/* Skip focus for mouse users */
:focus:not(:focus-visible) {
  outline: none;
}
```

### Semantic HTML

```html
<!-- Use semantic elements -->
<nav aria-label="Main">...</nav>
<main>...</main>
<article class="track-card">...</article>
<button aria-label="Remove from favorites">...</button>

<!-- Accessible form labels -->
<label for="search">Search tracks</label>
<input type="search" id="search" aria-describedby="search-help"/>
<div id="search-help" class="sr-only">
  Type to filter your Soundcloud tracks
</div>
```

### Screen Reader Only Content

```css
.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border-width: 0;
}
```

---

## 7. Animation & Transitions

### Common Transitions

```css
/* Hover effects */
.btn, .track-card, .nav-link {
  transition: 
    background-color var(--duration-fast) var(--ease-in-out),
    border-color var(--duration-fast) var(--ease-in-out),
    box-shadow var(--duration-fast) var(--ease-in-out),
    transform var(--duration-fast) var(--ease-out);
}

/* Theme transitions */
html {
  transition: background-color var(--duration-slow) var(--ease-in-out);
}

/* Loading states */
@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.loading {
  animation: pulse 2s var(--ease-in-out) infinite;
}
```

### Smooth Theme Switching

```css
* {
  transition: 
    background-color var(--duration-normal) var(--ease-in-out),
    color var(--duration-normal) var(--ease-in-out),
    border-color var(--duration-fast) var(--ease-in-out);
}
```

---

## 8. Sound Cistern Specific Patterns

### Track Cards

```css
.track-card {
  display: flex;
  gap: var(--space-4);
  padding: var(--space-4);
  background: var(--color-canvas-default);
  border: 1px solid var(--color-border-default);
  border-radius: var(--radius-lg);
  transition: 
    border-color var(--duration-fast) var(--ease-in-out),
    box-shadow var(--duration-fast) var(--ease-in-out);
}

.track-card:hover {
  border-color: var(--color-accent-muted);
  box-shadow: var(--shadow-md);
}

.track-artwork {
  width: 120px;
  height: 120px;
  border-radius: var(--radius-md);
  object-fit: cover;
}

.track-content {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
  flex: 1;
}

.track-title a {
  color: var(--color-accent-fg);
  text-decoration: none;
  font-weight: var(--font-semibold);
}

.track-title a:hover {
  text-decoration: underline;
}
```

### Favorites Button

```css
.favorite-btn {
  color: var(--color-fg-muted);
  background: transparent;
  border: none;
  cursor: pointer;
  padding: var(--space-1);
  border-radius: var(--radius-md);
  transition: 
    color var(--duration-fast) var(--ease-in-out),
    background var(--duration-fast) var(--ease-in-out);
}

.favorite-btn:hover {
  color: var(--color-accent-fg);
  background: var(--color-accent-subtle);
}

.favorite-btn.active {
  color: var(--color-danger-fg);
}

.favorite-btn.active:hover {
  background: rgba(207, 34, 46, 0.1);
}
```

### Filter Bar

**Use CSS Grid for better alignment than flexbox:**

```css
.filter-form {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: var(--space-4);
  align-items: end;
}

/* Desktop: proportional columns */
@media (min-width: 1024px) {
  .filter-form {
    grid-template-columns: 2fr 1fr 1fr 1fr 1fr auto;
  }
}

/* Tablet: 3 columns with search spanning full width */
@media (min-width: 769px) and (max-width: 1023px) {
  .filter-form {
    grid-template-columns: 1fr 1fr 1fr;
  }
  
  .filter-group--search {
    grid-column: 1 / -1;
  }
}

.filter-group {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
}

.filter-label {
  font-size: var(--text-xs);
  font-weight: var(--font-semibold);
  color: var(--color-fg-muted);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

/* Search input with icon - layout only */
.search-input-wrapper {
  position: relative;
}

.search-input-wrapper input {
  padding-left: var(--space-8);
}

.search-icon {
  position: absolute;
  left: var(--space-3);
  top: 50%;
  transform: translateY(-50%);
  color: var(--color-fg-muted);
  pointer-events: none;
}

/* Only constrain widths, let Pico handle heights */
.filter-input--number {
  width: 80px;
}

.filter-input--date {
  width: 140px;
}
```

**Key points:**
- Grid handles wrapping and sizing better than flexbox with `flex: 0 0 auto`
- Use `align-items: end` to align all controls to their bottom edges
- Only set widths for specific input types (number, date)
- Never override heights - Pico.css provides consistent sizing

### Soundcloud Brand Button

```css
.soundcloud-btn {
  background-color: #ff5500 !important;
  color: white !important;
  border: none;
  padding: var(--space-3) var(--space-6);
  font-size: var(--text-base);
  font-weight: var(--font-semibold);
  border-radius: var(--radius-md);
  text-decoration: none;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-2);
  transition: 
    background-color var(--duration-fast) var(--ease-in-out),
    transform var(--duration-fast) var(--ease-out),
    box-shadow var(--duration-fast) var(--ease-in-out);
  box-shadow: 0 4px 12px rgba(255, 85, 0, 0.3);
}

.soundcloud-btn:hover {
  background-color: #e64a00 !important;
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(255, 85, 0, 0.4);
}
```

---

## 9. Best Practices

### Do's

✅ **Use CSS variables for all values**
```css
/* Good */
padding: var(--space-4);
color: var(--color-fg-default);

/* Bad */
padding: 16px;
color: #333;
```

✅ **Use semantic HTML**
```html
<!-- Good -->
<nav>...</nav>
<article class="track-card">...</article>
<button>Click me</button>

<!-- Bad -->
<div class="nav">...</div>
<div class="card">...</div>
<div class="btn">Click me</div>
```

✅ **Provide accessible labels**
```html
<button aria-label="Remove from favorites">
  <svg>...</svg>
</button>
```

✅ **Test theme switching**
- Always test components in both light and dark modes
- Ensure sufficient contrast in both themes
- Check that shadows are visible in dark mode

### Don'ts

❌ **Don't use utility classes**
```css
/* Don't do this */
<div class="bg-blue-500 text-white p-4 rounded-lg">

/* Do this instead */
<article class="track-card">
```

❌ **Don't fight the framework**
```css
/* Don't override with excessive styles */
button {
  background: red !important;
  border: 1px solid blue !important;
  padding: 20px !important;
}

/* Work with Pico.css */
<button class="secondary">
```

❌ **Don't forget focus states**
```css
/* Bad - removes focus indicator */
button:focus {
  outline: none;
}

/* Good - visible focus */
button:focus-visible {
  outline: 2px solid var(--color-accent-fg);
}
```

❌ **Don't use `* { transition }` - Performance Killer**
```css
/* ❌ BAD: Applies to every element, causes jank with HTMX */
* {
  transition: background-color var(--duration-slow) var(--ease-out),
              color var(--duration-slow) var(--ease-out);
}

/* ✅ GOOD: Target only interactive elements */
.btn, .card, .nav-link {
  transition: 
    background-color var(--duration-fast) var(--ease-out),
    border-color var(--duration-fast) var(--ease-out);
}
```

❌ **Don't override Pico.css form element sizing**
```css
/* ❌ BAD: Fighting the framework */
.filter-input {
  height: 44px !important;
  padding: 0 1rem !important;
  line-height: 44px !important;
  /* Pico already handles this beautifully */
}

/* ✅ GOOD: Layout only, let Pico breathe */
.filter-group {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
}

.filter-input--number {
  width: 80px; /* Only constrain width, not height */
}
```

❌ **Don't duplicate CSS across files**
```css
/* ❌ BAD: Same rule in multiple files */
/* layout.css */
.skip-to-content { ... }

/* components.css */
.skip-to-content { ... } /* Duplicate! */

/* ✅ GOOD: Single source of truth */
/* layout.css only */
.skip-to-content { ... }
```

---

## 10. Development Workflow

### Adding New Components

1. **Choose the right CSS file** based on what you're building:
   - `tokens.css` - Only for new design tokens
   - `layout.css` - Headers, footers, navigation, containers
   - `components.css` - Reusable UI elements (buttons, cards, forms)
   - `pages.css` - Page-specific layouts (stream, auth, blog)

2. **Create the CSS** using design tokens:
   ```css
   /* In public/css/components.css */
   .new-component {
     padding: var(--space-4);
     background: var(--color-canvas-default);
     border: 1px solid var(--color-border-default);
     border-radius: var(--radius-lg);
   }
   ```

3. **Create the Templ component** in `views/components/`:
   ```templ
   package components
   
   templ NewComponent(data SomeData) {
     <div class="new-component">
       { data.Content }
     </div>
   }
   ```

4. **Use in templates**:
   ```templ
   @components.NewComponent(trackData)
   ```

5. **Regenerate templates**:
   ```bash
   make templ
   ```

### CSS File Organization Rules

**When to use each file:**

- **tokens.css**: Colors, spacing, typography, shadows, animations
  - Only edit when adding new design system tokens
  - Never add component styles here

- **layout.css**: Site structure and navigation
  - Headers, footers, main containers
  - Navigation patterns
  - Grid/flexbox layouts for page structure

- **components.css**: Reusable UI building blocks
  - Buttons, cards, badges, alerts
  - Form components (wrappers, layouts)
  - Navigation dropdowns, tooltips

- **pages.css**: Page-specific layouts
  - Stream page filters and track cards
  - Auth page layouts
  - Blog post styling
  - Admin dashboard patterns

**Pico.css Integration:**
```css
/* ✅ Good: Add layout/behavior only */
.filter-form {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: var(--space-4);
}

/* ❌ Bad: Don't override Pico's form element styles */
.filter-input {
  height: 44px;  /* Let Pico handle this */
  padding: 0 1rem;  /* Use Pico's defaults */
}
```

### Testing UI Changes

1. **Start dev server**:
   ```bash
   make dev-bg
   ```

2. **Check responsive design**:
   - Test at 320px (mobile)
   - Test at 768px (tablet)
   - Test at 1280px+ (desktop)

3. **Test theme switching**:
   - Toggle light/dark mode
   - Check contrast ratios
   - Verify smooth transitions

4. **Test accessibility**:
   - Navigate with keyboard only
   - Check focus indicators
   - Verify ARIA labels
   - Run contrast checker

---

## 11. Resources

- **Pico.css Docs**: https://picocss.com/docs
- **Primer Design**: https://primer.style/
- **Templ Guide**: https://templ.guide/
- **HTMX Docs**: https://htmx.org/docs/
- **WCAG Guidelines**: https://www.w3.org/WAI/WCAG21/quickref/

---

This skill provides everything you need to create beautiful, accessible, and consistent UIs in Sound Cistern. Always follow the semantic-first approach and leverage the design tokens for maintainability!
