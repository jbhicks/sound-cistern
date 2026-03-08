# PWA Icons

This directory needs two PNG icon files for the web app manifest:

- `icon-192.png` — 192×192 px
- `icon-512.png` — 512×512 px

## Recommended design

Use the Zap icon (from lucide-react) centered on a square canvas with a
gradient background matching the app's accent colors:

- Background: gradient from `#8b5cf6` (accent purple) to `#ec4899` (vapor-pink)
- Icon: white Zap SVG, scaled to ~55% of the canvas

## Quick generation with Inkscape or ImageMagick

```bash
# If you have the SVG source, convert with ImageMagick:
convert -background "#8b5cf6" icon.svg -resize 192x192 icon-192.png
convert -background "#8b5cf6" icon.svg -resize 512x512 icon-512.png
```

The manifest will still load and the app will still be installable without
these files — they only affect the home-screen icon appearance.
