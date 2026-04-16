---
name: sound-cistern-design
description: UI/UX design patterns, component guidelines, and visual design system for Sound Cistern
---

# Sound Cistern Design System

Use this skill when working on UI/UX design, component creation, or visual styling decisions for the Sound Cistern application.

## Design Principles
- Clean, minimal interfaces that let music take center stage
- Dark-first design optimized for music listening contexts
- Intuitive navigation with minimal cognitive load
- Responsive design that works across desktop and mobile

## Color Palette
- Primary background: `#0a0a0f` (deep black)
- Secondary background: `#1a1a2e` (dark navy)
- Accent: `#ff6b35` (coral/orange for CTAs)
- Text primary: `#ffffff`
- Text secondary: `#a0a0b0`
- Border subtle: `#2a2a3e`

## Typography
- Headings: Inter, sans-serif
- Body: Inter, sans-serif
- Monospace: JetBrains Mono (for code/metadata)
- Scale: Use Tailwind defaults (text-xs to text-7xl)

## Spacing System
- Use Tailwind spacing scale
- Base unit: 4px (0.25rem)
- Common patterns:
  - Component padding: px-4 py-3
  - Card gaps: gap-4
  - Section spacing: py-8 to py-16

## Components

### Buttons
```jsx
// Primary
<button className="bg-orange-500 hover:bg-orange-600 text-white px-4 py-2 rounded-lg transition-colors">

// Secondary
<button className="bg-surface hover:bg-surface-light text-white px-4 py-2 rounded-lg border border-white/10">

// Ghost
<button className="text-gray-400 hover:text-white transition-colors">
```

### Cards
```jsx
<div className="bg-surface rounded-xl p-6 border border-white/5 hover:border-white/10 transition-all">
```

### Forms
```jsx
<input 
  className="w-full bg-surface border border-white/10 rounded-lg px-4 py-3 text-white placeholder-gray-500 focus:outline-none focus:border-orange-500/50"
  placeholder="Enter value..."
/>
```

### Navigation
- Active state: text-orange-500
- Inactive: text-gray-400 hover:text-white
- Mobile: Hamburger menu with slide-out drawer

## Icon System
- Library: Lucide React
- Default size: 20px (w-5 h-5)
- Stroke width: 2

## Animation Guidelines
- Use Framer Motion for transitions
- Default spring: `{ type: "spring", stiffness: 300, damping: 30 }`
- Stagger children: 0.05s delay between items
- Page transitions: 0.3s ease-out

## Responsive Breakpoints
- Mobile: < 640px
- Tablet: 640px - 1024px
- Desktop: > 1024px
- Use Tailwind responsive prefixes: sm:, md:, lg:, xl:

## Common Patterns

### Track List Item
```jsx
<div className="flex items-center gap-4 p-3 hover:bg-white/5 rounded-lg group transition-colors">
  <div className="w-12 h-12 bg-surface rounded-md overflow-hidden">
    <img src={artwork} alt={title} className="w-full h-full object-cover" />
  </div>
  <div className="flex-1 min-w-0">
    <p className="text-white font-medium truncate">{title}</p>
    <p className="text-gray-400 text-sm truncate">{artist}</p>
  </div>
  <button className="opacity-0 group-hover:opacity-100 text-gray-400 hover:text-white transition-all">
    <Play className="w-5 h-5" />
  </button>
</div>
```

### Audio Player (Bottom Bar)
- Fixed to bottom
- Height: 80px
- Background: surface with backdrop-blur
- Progress bar: accent color, 4px height
- Controls: Previous, Play/Pause, Next (centered)
- Track info: left side
- Volume: right side

## Dark Mode
- This app is dark-only
- All colors assume dark background
- Never add light mode variants

## Audio-Specific UX
- Always show loading states for audio
- Display track progress and duration
- Visual feedback on interaction (play/pause)
- Error states for failed audio loads
- Respect user’s system volume