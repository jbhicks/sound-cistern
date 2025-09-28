# Pico.css CSS Variables Documentation

> **Source**: https://picocss.com/docs/css-variables

Customize Pico's design system with over 130 CSS variables to create a unique look and feel.

## Overview

Pico includes many custom properties (variables) that allow easy access to frequently used values such as:
- `font-family`
- `font-size`
- `border-radius`
- `margin`
- `padding`
- Colors and color schemes
- Spacing and typography

## Key Principles

### Prefixed Variables
All CSS variables are prefixed with `--pico-` to avoid collisions with other CSS frameworks or your own vars. You can remove or customize this prefix by recompiling the CSS files with SASS.

### Global vs Local Application
- **Global**: Define CSS variables within the `:root` selector to apply changes globally
- **Local**: Overwrite CSS variables on specific selectors to apply changes locally

## Example Usage

```css
:root {
  --pico-border-radius: 2rem;
  --pico-typography-spacing-vertical: 1.5rem;
  --pico-form-element-spacing-vertical: 1rem;
  --pico-form-element-spacing-horizontal: 1.25rem;
}

h1 {
  --pico-font-family: Pacifico, cursive;
  --pico-font-weight: 400;
  --pico-typography-spacing-vertical: 0.5rem;
}

button {
  --pico-font-weight: 700;
}
```

## Color Schemes

### Light Mode (Default)
To add or edit CSS variables for light mode only:

```css
/* Light color scheme (Default) */
/* Can be forced with data-theme="light" */
[data-theme="light"],
:root:not([data-theme="dark"]) {
  --pico-color: #000;
  --pico-background-color: #fff;
  /* ... other light mode variables */
}
```

### Dark Mode
To add or edit CSS variables for dark mode, define them twice:

1. **Auto Dark Mode** (based on user's device settings):
```css
/* Dark color scheme (Auto) */
/* Automatically enabled if user has Dark mode enabled */
@media only screen and (prefers-color-scheme: dark) {
  :root:not([data-theme]) {
    --pico-color: #fff;
    --pico-background-color: #000;
    /* ... other dark mode variables */
  }
}
```

2. **Forced Dark Mode** (manual toggle):
```css
/* Dark color scheme (Forced) */
/* Enabled if forced with data-theme="dark" */
[data-theme="dark"] {
  --pico-color: #fff;
  --pico-background-color: #000;
  /* ... other dark mode variables */
}
```

## Variable Categories

There are two main categories of CSS variables:

1. **Style variables** - Do not depend on color scheme (typography, spacing, borders)
2. **Color variables** - Depend on color scheme (backgrounds, text colors, borders)

## Common CSS Variables

### Typography
- `--pico-font-family`
- `--pico-font-size`
- `--pico-font-weight`
- `--pico-line-height`
- `--pico-typography-spacing-vertical`

### Colors
- `--pico-color` - Main text color
- `--pico-background-color` - Main background
- `--pico-primary` - Primary brand color
- `--pico-secondary` - Secondary color
- `--pico-muted-color` - Muted text
- `--pico-muted-border-color` - Subtle borders

### Spacing
- `--pico-spacing`
- `--pico-form-element-spacing-vertical`
- `--pico-form-element-spacing-horizontal`

### Borders
- `--pico-border-radius`
- `--pico-border-width`
- `--pico-outline-width`

### Form Elements
- `--pico-form-element-background-color`
- `--pico-form-element-border-color`
- `--pico-form-element-color`

## Theme Implementation in Buffalo SaaS Template

Our template implements theme switching using:

```javascript
function setTheme(theme) {
  if (theme === 'auto') {
    localStorage.removeItem('picoPreferredColorScheme');
    document.documentElement.removeAttribute('data-theme');
  } else {
    localStorage.setItem('picoPreferredColorScheme', theme);
    document.documentElement.setAttribute('data-theme', theme);
  }
}
```

## Best Practices

1. **Use semantic variable names** - Prefer `--pico-primary` over hardcoded colors
2. **Test both themes** - Always verify customizations work in light and dark modes
3. **Override sparingly** - Use Pico's design system as much as possible
4. **Maintain accessibility** - Ensure color contrast ratios remain compliant
5. **Consider auto theme** - Respect user's system preferences when possible

## Resources

- [Pico.css Documentation](https://picocss.com/docs)
- [Pico.css GitHub](https://github.com/picocss/pico)
- [CSS Variables Specification](https://developer.mozilla.org/en-US/docs/Web/CSS/Using_CSS_custom_properties)

# Default Styles CSS Variables
```css
:root,
:host {
  --pico-font-family-emoji: "Apple Color Emoji", "Segoe UI Emoji", "Segoe UI Symbol", "Noto Color Emoji";
  --pico-font-family-sans-serif: system-ui, "Segoe UI", Roboto, Oxygen, Ubuntu, Cantarell, Helvetica, Arial, "Helvetica Neue", sans-serif, var(--pico-font-family-emoji);
  --pico-font-family-monospace: ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas, "Liberation Mono", monospace, var(--pico-font-family-emoji);
  --pico-font-family: var(--pico-font-family-sans-serif);
  --pico-line-height: 1.5;
  --pico-font-weight: 400;
  --pico-font-size: 100%;
  --pico-text-underline-offset: 0.1rem;
  --pico-border-radius: 0.25rem;
  --pico-border-width: 0.0625rem;
  --pico-outline-width: 0.125rem;
  --pico-transition: 0.2s ease-in-out;
  --pico-spacing: 1rem;
  --pico-typography-spacing-vertical: 1rem;
  --pico-block-spacing-vertical: var(--pico-spacing);
  --pico-block-spacing-horizontal: var(--pico-spacing);
  --pico-grid-column-gap: var(--pico-spacing);
  --pico-grid-row-gap: var(--pico-spacing);
  --pico-form-element-spacing-vertical: 0.75rem;
  --pico-form-element-spacing-horizontal: 1rem;
  --pico-group-box-shadow: 0 0 0 rgba(0, 0, 0, 0);
  --pico-group-box-shadow-focus-with-button: 0 0 0 var(--pico-outline-width) var(--pico-primary-focus);
  --pico-group-box-shadow-focus-with-input: 0 0 0 0.0625rem var(--pico-form-element-border-color);
  --pico-modal-overlay-backdrop-filter: blur(0.375rem);
  --pico-nav-element-spacing-vertical: 1rem;
  --pico-nav-element-spacing-horizontal: 0.5rem;
  --pico-nav-link-spacing-vertical: 0.5rem;
  --pico-nav-link-spacing-horizontal: 0.5rem;
  --pico-nav-breadcrumb-divider: ">";
  --pico-icon-checkbox: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='24' height='24' viewBox='0 0 24 24' fill='none' stroke='rgb(255, 255, 255)' stroke-width='4' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpolyline points='20 6 9 17 4 12'%3E%3C/polyline%3E%3C/svg%3E");
  --pico-icon-minus: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='24' height='24' viewBox='0 0 24 24' fill='none' stroke='rgb(255, 255, 255)' stroke-width='4' stroke-linecap='round' stroke-linejoin='round'%3E%3Cline x1='5' y1='12' x2='19' y2='12'%3E%3C/line%3E%3C/svg%3E");
  --pico-icon-chevron: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='24' height='24' viewBox='0 0 24 24' fill='none' stroke='rgb(136, 145, 164)' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpolyline points='6 9 12 15 18 9'%3E%3C/polyline%3E%3C/svg%3E");
  --pico-icon-date: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='24' height='24' viewBox='0 0 24 24' fill='none' stroke='rgb(136, 145, 164)' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Crect x='3' y='4' width='18' height='18' rx='2' ry='2'%3E%3C/rect%3E%3Cline x1='16' y1='2' x2='16' y2='6'%3E%3C/line%3E%3Cline x1='8' y1='2' x2='8' y2='6'%3E%3C/line%3E%3Cline x1='3' y1='10' x2='21' y2='10'%3E%3C/line%3E%3C/svg%3E");
  --pico-icon-time: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='24' height='24' viewBox='0 0 24 24' fill='none' stroke='rgb(136, 145, 164)' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Ccircle cx='12' cy='12' r='10'%3E%3C/circle%3E%3Cpolyline points='12 6 12 12 16 14'%3E%3C/polyline%3E%3C/svg%3E");
  --pico-icon-search: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='24' height='24' viewBox='0 0 24 24' fill='none' stroke='rgb(136, 145, 164)' stroke-width='1.5' stroke-linecap='round' stroke-linejoin='round'%3E%3Ccircle cx='11' cy='11' r='8'%3E%3C/circle%3E%3Cline x1='21' y1='21' x2='16.65' y2='16.65'%3E%3C/line%3E%3C/svg%3E");
  --pico-icon-close: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='24' height='24' viewBox='0 0 24 24' fill='none' stroke='rgb(136, 145, 164)' stroke-width='3' stroke-linecap='round' stroke-linejoin='round'%3E%3Cline x1='18' y1='6' x2='6' y2='18'%3E%3C/line%3E%3Cline x1='6' y1='6' x2='18' y2='18'%3E%3C/line%3E%3C/svg%3E");
  --pico-icon-loading: url("data:image/svg+xml,%3Csvg fill='none' height='24' width='24' viewBox='0 0 24 24' xmlns='http://www.w3.org/2000/svg' %3E%3Cstyle%3E g %7B animation: rotate 2s linear infinite; transform-origin: center center; %7D circle %7B stroke-dasharray: 75,100; stroke-dashoffset: -5; animation: dash 1.5s ease-in-out infinite; stroke-linecap: round; %7D @keyframes rotate %7B 0%25 %7B transform: rotate(0deg); %7D 100%25 %7B transform: rotate(360deg); %7D %7D @keyframes dash %7B 0%25 %7B stroke-dasharray: 1,100; stroke-dashoffset: 0; %7D 50%25 %7B stroke-dasharray: 44.5,100; stroke-dashoffset: -17.5; %7D 100%25 %7B stroke-dasharray: 44.5,100; stroke-dashoffset: -62; %7D %7D %3C/style%3E%3Cg%3E%3Ccircle cx='12' cy='12' r='10' fill='none' stroke='rgb(136, 145, 164)' stroke-width='4' /%3E%3C/g%3E%3C/svg%3E");
}
@media (min-width: 576px) {
  :root,
  :host {
    --pico-font-size: 106.25%;
  }
}
@media (min-width: 768px) {
  :root,
  :host {
    --pico-font-size: 112.5%;
  }
}
@media (min-width: 1024px) {
  :root,
  :host {
    --pico-font-size: 118.75%;
  }
}
@media (min-width: 1280px) {
  :root,
  :host {
    --pico-font-size: 125%;
  }
}
@media (min-width: 1536px) {
  :root,
  :host {
    --pico-font-size: 131.25%;
  }
}

a {
  --pico-text-decoration: underline;
}
a.secondary, a.contrast {
  --pico-text-decoration: underline;
}

small {
  --pico-font-size: 0.875em;
}

h1,
h2,
h3,
h4,
h5,
h6 {
  --pico-font-weight: 700;
}

h1 {
  --pico-font-size: 2rem;
  --pico-line-height: 1.125;
  --pico-typography-spacing-top: 3rem;
}

h2 {
  --pico-font-size: 1.75rem;
  --pico-line-height: 1.15;
  --pico-typography-spacing-top: 2.625rem;
}

h3 {
  --pico-font-size: 1.5rem;
  --pico-line-height: 1.175;
  --pico-typography-spacing-top: 2.25rem;
}

h4 {
  --pico-font-size: 1.25rem;
  --pico-line-height: 1.2;
  --pico-typography-spacing-top: 1.874rem;
}

h5 {
  --pico-font-size: 1.125rem;
  --pico-line-height: 1.225;
  --pico-typography-spacing-top: 1.6875rem;
}

h6 {
  --pico-font-size: 1rem;
  --pico-line-height: 1.25;
  --pico-typography-spacing-top: 1.5rem;
}

thead th,
thead td,
tfoot th,
tfoot td {
  --pico-font-weight: 600;
  --pico-border-width: 0.1875rem;
}

pre,
code,
kbd,
samp {
  --pico-font-family: var(--pico-font-family-monospace);
}

kbd {
  --pico-font-weight: bolder;
}

input:not([type=submit],
[type=button],
[type=reset],
[type=checkbox],
[type=radio],
[type=file]),
:where(select, textarea) {
  --pico-outline-width: 0.0625rem;
}

[type=search] {
  --pico-border-radius: 5rem;
}

[type=checkbox],
[type=radio] {
  --pico-border-width: 0.125rem;
}

[type=checkbox][role=switch] {
  --pico-border-width: 0.1875rem;
}

details.dropdown summary:not([role=button]) {
  --pico-outline-width: 0.0625rem;
}

nav details.dropdown summary:focus-visible {
  --pico-outline-width: 0.125rem;
}

[role=search] {
  --pico-border-radius: 5rem;
}

[role=search]:has(button.secondary:focus,
[type=submit].secondary:focus,
[type=button].secondary:focus,
[role=button].secondary:focus),
[role=group]:has(button.secondary:focus,
[type=submit].secondary:focus,
[type=button].secondary:focus,
[role=button].secondary:focus) {
  --pico-group-box-shadow-focus-with-button: 0 0 0 var(--pico-outline-width) var(--pico-secondary-focus);
}
[role=search]:has(button.contrast:focus,
[type=submit].contrast:focus,
[type=button].contrast:focus,
[role=button].contrast:focus),
[role=group]:has(button.contrast:focus,
[type=submit].contrast:focus,
[type=button].contrast:focus,
[role=button].contrast:focus) {
  --pico-group-box-shadow-focus-with-button: 0 0 0 var(--pico-outline-width) var(--pico-contrast-focus);
}
[role=search] button,
[role=search] [type=submit],
[role=search] [type=button],
[role=search] [role=button],
[role=group] button,
[role=group] [type=submit],
[role=group] [type=button],
[role=group] [role=button] {
  --pico-form-element-spacing-horizontal: 2rem;
}

details summary[role=button]:not(.outline)::after {
  filter: brightness(0) invert(1);
}

[aria-busy=true]:not(input, select, textarea):is(button, [type=submit], [type=button], [type=reset], [role=button]):not(.outline)::before {
  filter: brightness(0) invert(1);
}
```
 
# Default Colors CSS Variables

```css
[data-theme=light],
:root:not([data-theme=dark]),
:host(:not([data-theme=dark])) {
  color-scheme: light;
  --pico-background-color: #fff;
  --pico-color: #373c44;
  --pico-text-selection-color: rgba(2, 154, 232, 0.25);
  --pico-muted-color: #646b79;
  --pico-muted-border-color: rgb(231, 234, 239.5);
  --pico-primary: #0172ad;
  --pico-primary-background: #0172ad;
  --pico-primary-border: var(--pico-primary-background);
  --pico-primary-underline: rgba(1, 114, 173, 0.5);
  --pico-primary-hover: #015887;
  --pico-primary-hover-background: #02659a;
  --pico-primary-hover-border: var(--pico-primary-hover-background);
  --pico-primary-hover-underline: var(--pico-primary-hover);
  --pico-primary-focus: rgba(2, 154, 232, 0.5);
  --pico-primary-inverse: #fff;
  --pico-secondary: #5d6b89;
  --pico-secondary-background: #525f7a;
  --pico-secondary-border: var(--pico-secondary-background);
  --pico-secondary-underline: rgba(93, 107, 137, 0.5);
  --pico-secondary-hover: #48536b;
  --pico-secondary-hover-background: #48536b;
  --pico-secondary-hover-border: var(--pico-secondary-hover-background);
  --pico-secondary-hover-underline: var(--pico-secondary-hover);
  --pico-secondary-focus: rgba(93, 107, 137, 0.25);
  --pico-secondary-inverse: #fff;
  --pico-contrast: #181c25;
  --pico-contrast-background: #181c25;
  --pico-contrast-border: var(--pico-contrast-background);
  --pico-contrast-underline: rgba(24, 28, 37, 0.5);
  --pico-contrast-hover: #000;
  --pico-contrast-hover-background: #000;
  --pico-contrast-hover-border: var(--pico-contrast-hover-background);
  --pico-contrast-hover-underline: var(--pico-secondary-hover);
  --pico-contrast-focus: rgba(93, 107, 137, 0.25);
  --pico-contrast-inverse: #fff;
  --pico-box-shadow: 0.0145rem 0.029rem 0.174rem rgba(129, 145, 181, 0.01698), 0.0335rem 0.067rem 0.402rem rgba(129, 145, 181, 0.024), 0.0625rem 0.125rem 0.75rem rgba(129, 145, 181, 0.03), 0.1125rem 0.225rem 1.35rem rgba(129, 145, 181, 0.036), 0.2085rem 0.417rem 2.502rem rgba(129, 145, 181, 0.04302), 0.5rem 1rem 6rem rgba(129, 145, 181, 0.06), 0 0 0 0.0625rem rgba(129, 145, 181, 0.015);
  --pico-h1-color: #2d3138;
  --pico-h2-color: #373c44;
  --pico-h3-color: #424751;
  --pico-h4-color: #4d535e;
  --pico-h5-color: #5c6370;
  --pico-h6-color: #646b79;
  --pico-mark-background-color: rgb(252.5, 230.5, 191.5);
  --pico-mark-color: #0f1114;
  --pico-ins-color: rgb(28.5, 105.5, 84);
  --pico-del-color: rgb(136, 56.5, 53);
  --pico-blockquote-border-color: var(--pico-muted-border-color);
  --pico-blockquote-footer-color: var(--pico-muted-color);
  --pico-button-box-shadow: 0 0 0 rgba(0, 0, 0, 0);
  --pico-button-hover-box-shadow: 0 0 0 rgba(0, 0, 0, 0);
  --pico-table-border-color: var(--pico-muted-border-color);
  --pico-table-row-stripped-background-color: rgba(111, 120, 135, 0.0375);
  --pico-code-background-color: rgb(243, 244.5, 246.75);
  --pico-code-color: #646b79;
  --pico-code-kbd-background-color: var(--pico-color);
  --pico-code-kbd-color: var(--pico-background-color);
  --pico-form-element-background-color: rgb(251, 251.5, 252.25);
  --pico-form-element-selected-background-color: #dfe3eb;
  --pico-form-element-border-color: #cfd5e2;
  --pico-form-element-color: #23262c;
  --pico-form-element-placeholder-color: var(--pico-muted-color);
  --pico-form-element-active-background-color: #fff;
  --pico-form-element-active-border-color: var(--pico-primary-border);
  --pico-form-element-focus-color: var(--pico-primary-border);
  --pico-form-element-disabled-opacity: 0.5;
  --pico-form-element-invalid-border-color: rgb(183.5, 105.5, 106.5);
  --pico-form-element-invalid-active-border-color: rgb(200.25, 79.25, 72.25);
  --pico-form-element-invalid-focus-color: var(--pico-form-element-invalid-active-border-color);
  --pico-form-element-valid-border-color: rgb(76, 154.5, 137.5);
  --pico-form-element-valid-active-border-color: rgb(39, 152.75, 118.75);
  --pico-form-element-valid-focus-color: var(--pico-form-element-valid-active-border-color);
  --pico-switch-background-color: #bfc7d9;
  --pico-switch-checked-background-color: var(--pico-primary-background);
  --pico-switch-color: #fff;
  --pico-switch-thumb-box-shadow: 0 0 0 rgba(0, 0, 0, 0);
  --pico-range-border-color: #dfe3eb;
  --pico-range-active-border-color: #bfc7d9;
  --pico-range-thumb-border-color: var(--pico-background-color);
  --pico-range-thumb-color: var(--pico-secondary-background);
  --pico-range-thumb-active-color: var(--pico-primary-background);
  --pico-accordion-border-color: var(--pico-muted-border-color);
  --pico-accordion-active-summary-color: var(--pico-primary-hover);
  --pico-accordion-close-summary-color: var(--pico-color);
  --pico-accordion-open-summary-color: var(--pico-muted-color);
  --pico-card-background-color: var(--pico-background-color);
  --pico-card-border-color: var(--pico-muted-border-color);
  --pico-card-box-shadow: var(--pico-box-shadow);
  --pico-card-sectioning-background-color: rgb(251, 251.5, 252.25);
  --pico-dropdown-background-color: #fff;
  --pico-dropdown-border-color: #eff1f4;
  --pico-dropdown-box-shadow: var(--pico-box-shadow);
  --pico-dropdown-color: var(--pico-color);
  --pico-dropdown-hover-background-color: #eff1f4;
  --pico-loading-spinner-opacity: 0.5;
  --pico-modal-overlay-background-color: rgba(232, 234, 237, 0.75);
  --pico-progress-background-color: #dfe3eb;
  --pico-progress-color: var(--pico-primary-background);
  --pico-tooltip-background-color: var(--pico-contrast-background);
  --pico-tooltip-color: var(--pico-contrast-inverse);
  --pico-icon-valid: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='24' height='24' viewBox='0 0 24 24' fill='none' stroke='rgb(76, 154.5, 137.5)' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpolyline points='20 6 9 17 4 12'%3E%3C/polyline%3E%3C/svg%3E");
  --pico-icon-invalid: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='24' height='24' viewBox='0 0 24 24' fill='none' stroke='rgb(200.25, 79.25, 72.25)' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Ccircle cx='12' cy='12' r='10'%3E%3C/circle%3E%3Cline x1='12' y1='8' x2='12' y2='12'%3E%3C/line%3E%3Cline x1='12' y1='16' x2='12.01' y2='16'%3E%3C/line%3E%3C/svg%3E");
}
[data-theme=light] input:is([type=submit],
[type=button],
[type=reset],
[type=checkbox],
[type=radio],
[type=file]),
:root:not([data-theme=dark]) input:is([type=submit],
[type=button],
[type=reset],
[type=checkbox],
[type=radio],
[type=file]),
:host(:not([data-theme=dark])) input:is([type=submit],
[type=button],
[type=reset],
[type=checkbox],
[type=radio],
[type=file]) {
  --pico-form-element-focus-color: var(--pico-primary-focus);
}

@media only screen and (prefers-color-scheme: dark) {
  :root:not([data-theme]),
  :host(:not([data-theme])) {
    color-scheme: dark;
    --pico-background-color: rgb(19, 22.5, 30.5);
    --pico-color: #c2c7d0;
    --pico-text-selection-color: rgba(1, 170, 255, 0.1875);
    --pico-muted-color: #7b8495;
    --pico-muted-border-color: #202632;
    --pico-primary: #01aaff;
    --pico-primary-background: #0172ad;
    --pico-primary-border: var(--pico-primary-background);
    --pico-primary-underline: rgba(1, 170, 255, 0.5);
    --pico-primary-hover: #79c0ff;
    --pico-primary-hover-background: #017fc0;
    --pico-primary-hover-border: var(--pico-primary-hover-background);
    --pico-primary-hover-underline: var(--pico-primary-hover);
    --pico-primary-focus: rgba(1, 170, 255, 0.375);
    --pico-primary-inverse: #fff;
    --pico-secondary: #969eaf;
    --pico-secondary-background: #525f7a;
    --pico-secondary-border: var(--pico-secondary-background);
    --pico-secondary-underline: rgba(150, 158, 175, 0.5);
    --pico-secondary-hover: #b3b9c5;
    --pico-secondary-hover-background: #5d6b89;
    --pico-secondary-hover-border: var(--pico-secondary-hover-background);
    --pico-secondary-hover-underline: var(--pico-secondary-hover);
    --pico-secondary-focus: rgba(144, 158, 190, 0.25);
    --pico-secondary-inverse: #fff;
    --pico-contrast: #dfe3eb;
    --pico-contrast-background: #eff1f4;
    --pico-contrast-border: var(--pico-contrast-background);
    --pico-contrast-underline: rgba(223, 227, 235, 0.5);
    --pico-contrast-hover: #fff;
    --pico-contrast-hover-background: #fff;
    --pico-contrast-hover-border: var(--pico-contrast-hover-background);
    --pico-contrast-hover-underline: var(--pico-contrast-hover);
    --pico-contrast-focus: rgba(207, 213, 226, 0.25);
    --pico-contrast-inverse: #000;
    --pico-box-shadow: 0.0145rem 0.029rem 0.174rem rgba(7, 8.5, 12, 0.01698), 0.0335rem 0.067rem 0.402rem rgba(7, 8.5, 12, 0.024), 0.0625rem 0.125rem 0.75rem rgba(7, 8.5, 12, 0.03), 0.1125rem 0.225rem 1.35rem rgba(7, 8.5, 12, 0.036), 0.2085rem 0.417rem 2.502rem rgba(7, 8.5, 12, 0.04302), 0.5rem 1rem 6rem rgba(7, 8.5, 12, 0.06), 0 0 0 0.0625rem rgba(7, 8.5, 12, 0.015);
    --pico-h1-color: #f0f1f3;
    --pico-h2-color: #e0e3e7;
    --pico-h3-color: #c2c7d0;
    --pico-h4-color: #b3b9c5;
    --pico-h5-color: #a4acba;
    --pico-h6-color: #8891a4;
    --pico-mark-background-color: #014063;
    --pico-mark-color: #fff;
    --pico-ins-color: #62af9a;
    --pico-del-color: rgb(205.5, 126, 123);
    --pico-blockquote-border-color: var(--pico-muted-border-color);
    --pico-blockquote-footer-color: var(--pico-muted-color);
    --pico-button-box-shadow: 0 0 0 rgba(0, 0, 0, 0);
    --pico-button-hover-box-shadow: 0 0 0 rgba(0, 0, 0, 0);
    --pico-table-border-color: var(--pico-muted-border-color);
    --pico-table-row-stripped-background-color: rgba(111, 120, 135, 0.0375);
    --pico-code-background-color: rgb(26, 30.5, 40.25);
    --pico-code-color: #8891a4;
    --pico-code-kbd-background-color: var(--pico-color);
    --pico-code-kbd-color: var(--pico-background-color);
    --pico-form-element-background-color: rgb(28, 33, 43.5);
    --pico-form-element-selected-background-color: #2a3140;
    --pico-form-element-border-color: #2a3140;
    --pico-form-element-color: #e0e3e7;
    --pico-form-element-placeholder-color: #8891a4;
    --pico-form-element-active-background-color: rgb(26, 30.5, 40.25);
    --pico-form-element-active-border-color: var(--pico-primary-border);
    --pico-form-element-focus-color: var(--pico-primary-border);
    --pico-form-element-disabled-opacity: 0.5;
    --pico-form-element-invalid-border-color: rgb(149.5, 74, 80);
    --pico-form-element-invalid-active-border-color: rgb(183.25, 63.5, 59);
    --pico-form-element-invalid-focus-color: var(--pico-form-element-invalid-active-border-color);
    --pico-form-element-valid-border-color: #2a7b6f;
    --pico-form-element-valid-active-border-color: rgb(22, 137, 105.5);
    --pico-form-element-valid-focus-color: var(--pico-form-element-valid-active-border-color);
    --pico-switch-background-color: #333c4e;
    --pico-switch-checked-background-color: var(--pico-primary-background);
    --pico-switch-color: #fff;
    --pico-switch-thumb-box-shadow: 0 0 0 rgba(0, 0, 0, 0);
    --pico-range-border-color: #202632;
    --pico-range-active-border-color: #2a3140;
    --pico-range-thumb-border-color: var(--pico-background-color);
    --pico-range-thumb-color: var(--pico-secondary-background);
    --pico-range-thumb-active-color: var(--pico-primary-background);
    --pico-accordion-border-color: var(--pico-muted-border-color);
    --pico-accordion-active-summary-color: var(--pico-primary-hover);
    --pico-accordion-close-summary-color: var(--pico-color);
    --pico-accordion-open-summary-color: var(--pico-muted-color);
    --pico-card-background-color: #181c25;
    --pico-card-border-color: var(--pico-card-background-color);
    --pico-card-box-shadow: var(--pico-box-shadow);
    --pico-card-sectioning-background-color: rgb(26, 30.5, 40.25);
    --pico-dropdown-background-color: #181c25;
    --pico-dropdown-border-color: #202632;
    --pico-dropdown-box-shadow: var(--pico-box-shadow);
    --pico-dropdown-color: var(--pico-color);
    --pico-dropdown-hover-background-color: #202632;
    --pico-loading-spinner-opacity: 0.5;
    --pico-modal-overlay-background-color: rgba(7.5, 8.5, 10, 0.75);
    --pico-progress-background-color: #202632;
    --pico-progress-color: var(--pico-primary-background);
    --pico-tooltip-background-color: var(--pico-contrast-background);
    --pico-tooltip-color: var(--pico-contrast-inverse);
    --pico-icon-valid: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='24' height='24' viewBox='0 0 24 24' fill='none' stroke='rgb(42, 123, 111)' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpolyline points='20 6 9 17 4 12'%3E%3C/polyline%3E%3C/svg%3E");
    --pico-icon-invalid: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='24' height='24' viewBox='0 0 24 24' fill='none' stroke='rgb(149.5, 74, 80)' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Ccircle cx='12' cy='12' r='10'%3E%3C/circle%3E%3Cline x1='12' y1='8' x2='12' y2='12'%3E%3C/line%3E%3Cline x1='12' y1='16' x2='12.01' y2='16'%3E%3C/line%3E%3C/svg%3E");
  }
  :root:not([data-theme]) input:is([type=submit],
  [type=button],
  [type=reset],
  [type=checkbox],
  [type=radio],
  [type=file]),
  :host(:not([data-theme])) input:is([type=submit],
  [type=button],
  [type=reset],
  [type=checkbox],
  [type=radio],
  [type=file]) {
    --pico-form-element-focus-color: var(--pico-primary-focus);
  }
  :root:not([data-theme]) details summary[role=button].contrast:not(.outline)::after,
  :host(:not([data-theme])) details summary[role=button].contrast:not(.outline)::after {
    filter: brightness(0);
  }
  :root:not([data-theme]) [aria-busy=true]:not(input, select, textarea).contrast:is(button,
  [type=submit],
  [type=button],
  [type=reset],
  [role=button]):not(.outline)::before,
  :host(:not([data-theme])) [aria-busy=true]:not(input, select, textarea).contrast:is(button,
  [type=submit],
  [type=button],
  [type=reset],
  [role=button]):not(.outline)::before {
    filter: brightness(0);
  }
}
[data-theme=dark] {
  color-scheme: dark;
  --pico-background-color: rgb(19, 22.5, 30.5);
  --pico-color: #c2c7d0;
  --pico-text-selection-color: rgba(1, 170, 255, 0.1875);
  --pico-muted-color: #7b8495;
  --pico-muted-border-color: #202632;
  --pico-primary: #01aaff;
  --pico-primary-background: #0172ad;
  --pico-primary-border: var(--pico-primary-background);
  --pico-primary-underline: rgba(1, 170, 255, 0.5);
  --pico-primary-hover: #79c0ff;
  --pico-primary-hover-background: #017fc0;
  --pico-primary-hover-border: var(--pico-primary-hover-background);
  --pico-primary-hover-underline: var(--pico-primary-hover);
  --pico-primary-focus: rgba(1, 170, 255, 0.375);
  --pico-primary-inverse: #fff;
  --pico-secondary: #969eaf;
  --pico-secondary-background: #525f7a;
  --pico-secondary-border: var(--pico-secondary-background);
  --pico-secondary-underline: rgba(150, 158, 175, 0.5);
  --pico-secondary-hover: #b3b9c5;
  --pico-secondary-hover-background: #5d6b89;
  --pico-secondary-hover-border: var(--pico-secondary-hover-background);
  --pico-secondary-hover-underline: var(--pico-secondary-hover);
  --pico-secondary-focus: rgba(144, 158, 190, 0.25);
  --pico-secondary-inverse: #fff;
  --pico-contrast: #dfe3eb;
  --pico-contrast-background: #eff1f4;
  --pico-contrast-border: var(--pico-contrast-background);
  --pico-contrast-underline: rgba(223, 227, 235, 0.5);
  --pico-contrast-hover: #fff;
  --pico-contrast-hover-background: #fff;
  --pico-contrast-hover-border: var(--pico-contrast-hover-background);
  --pico-contrast-hover-underline: var(--pico-contrast-hover);
  --pico-contrast-focus: rgba(207, 213, 226, 0.25);
  --pico-contrast-inverse: #000;
  --pico-box-shadow: 0.0145rem 0.029rem 0.174rem rgba(7, 8.5, 12, 0.01698), 0.0335rem 0.067rem 0.402rem rgba(7, 8.5, 12, 0.024), 0.0625rem 0.125rem 0.75rem rgba(7, 8.5, 12, 0.03), 0.1125rem 0.225rem 1.35rem rgba(7, 8.5, 12, 0.036), 0.2085rem 0.417rem 2.502rem rgba(7, 8.5, 12, 0.04302), 0.5rem 1rem 6rem rgba(7, 8.5, 12, 0.06), 0 0 0 0.0625rem rgba(7, 8.5, 12, 0.015);
  --pico-h1-color: #f0f1f3;
  --pico-h2-color: #e0e3e7;
  --pico-h3-color: #c2c7d0;
  --pico-h4-color: #b3b9c5;
  --pico-h5-color: #a4acba;
  --pico-h6-color: #8891a4;
  --pico-mark-background-color: #014063;
  --pico-mark-color: #fff;
  --pico-ins-color: #62af9a;
  --pico-del-color: rgb(205.5, 126, 123);
  --pico-blockquote-border-color: var(--pico-muted-border-color);
  --pico-blockquote-footer-color: var(--pico-muted-color);
  --pico-button-box-shadow: 0 0 0 rgba(0, 0, 0, 0);
  --pico-button-hover-box-shadow: 0 0 0 rgba(0, 0, 0, 0);
  --pico-table-border-color: var(--pico-muted-border-color);
  --pico-table-row-stripped-background-color: rgba(111, 120, 135, 0.0375);
  --pico-code-background-color: rgb(26, 30.5, 40.25);
  --pico-code-color: #8891a4;
  --pico-code-kbd-background-color: var(--pico-color);
  --pico-code-kbd-color: var(--pico-background-color);
  --pico-form-element-background-color: rgb(28, 33, 43.5);
  --pico-form-element-selected-background-color: #2a3140;
  --pico-form-element-border-color: #2a3140;
  --pico-form-element-color: #e0e3e7;
  --pico-form-element-placeholder-color: #8891a4;
  --pico-form-element-active-background-color: rgb(26, 30.5, 40.25);
  --pico-form-element-active-border-color: var(--pico-primary-border);
  --pico-form-element-focus-color: var(--pico-primary-border);
  --pico-form-element-disabled-opacity: 0.5;
  --pico-form-element-invalid-border-color: rgb(149.5, 74, 80);
  --pico-form-element-invalid-active-border-color: rgb(183.25, 63.5, 59);
  --pico-form-element-invalid-focus-color: var(--pico-form-element-invalid-active-border-color);
  --pico-form-element-valid-border-color: #2a7b6f;
  --pico-form-element-valid-active-border-color: rgb(22, 137, 105.5);
  --pico-form-element-valid-focus-color: var(--pico-form-element-valid-active-border-color);
  --pico-switch-background-color: #333c4e;
  --pico-switch-checked-background-color: var(--pico-primary-background);
  --pico-switch-color: #fff;
  --pico-switch-thumb-box-shadow: 0 0 0 rgba(0, 0, 0, 0);
  --pico-range-border-color: #202632;
  --pico-range-active-border-color: #2a3140;
  --pico-range-thumb-border-color: var(--pico-background-color);
  --pico-range-thumb-color: var(--pico-secondary-background);
  --pico-range-thumb-active-color: var(--pico-primary-background);
  --pico-accordion-border-color: var(--pico-muted-border-color);
  --pico-accordion-active-summary-color: var(--pico-primary-hover);
  --pico-accordion-close-summary-color: var(--pico-color);
  --pico-accordion-open-summary-color: var(--pico-muted-color);
  --pico-card-background-color: #181c25;
  --pico-card-border-color: var(--pico-card-background-color);
  --pico-card-box-shadow: var(--pico-box-shadow);
  --pico-card-sectioning-background-color: rgb(26, 30.5, 40.25);
  --pico-dropdown-background-color: #181c25;
  --pico-dropdown-border-color: #202632;
  --pico-dropdown-box-shadow: var(--pico-box-shadow);
  --pico-dropdown-color: var(--pico-color);
  --pico-dropdown-hover-background-color: #202632;
  --pico-loading-spinner-opacity: 0.5;
  --pico-modal-overlay-background-color: rgba(7.5, 8.5, 10, 0.75);
  --pico-progress-background-color: #202632;
  --pico-progress-color: var(--pico-primary-background);
  --pico-tooltip-background-color: var(--pico-contrast-background);
  --pico-tooltip-color: var(--pico-contrast-inverse);
  --pico-icon-valid: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='24' height='24' viewBox='0 0 24 24' fill='none' stroke='rgb(42, 123, 111)' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpolyline points='20 6 9 17 4 12'%3E%3C/polyline%3E%3C/svg%3E");
  --pico-icon-invalid: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='24' height='24' viewBox='0 0 24 24' fill='none' stroke='rgb(149.5, 74, 80)' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Ccircle cx='12' cy='12' r='10'%3E%3C/circle%3E%3Cline x1='12' y1='8' x2='12' y2='12'%3E%3C/line%3E%3Cline x1='12' y1='16' x2='12.01' y2='16'%3E%3C/line%3E%3C/svg%3E");
}
[data-theme=dark] input:is([type=submit],
[type=button],
[type=reset],
[type=checkbox],
[type=radio],
[type=file]) {
  --pico-form-element-focus-color: var(--pico-primary-focus);
}
[data-theme=dark] details summary[role=button].contrast:not(.outline)::after {
  filter: brightness(0);
}
[data-theme=dark] [aria-busy=true]:not(input, select, textarea).contrast:is(button,
[type=submit],
[type=button],
[type=reset],
[role=button]):not(.outline)::before {
  filter: brightness(0);
}

progress,
[type=checkbox],
[type=radio],
[type=range] {
  accent-color: var(--pico-primary);
}
```