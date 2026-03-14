# Performance Optimization Regression Checklist

Use this checklist to verify that essential functionality still works after applying performance optimizations.

## Player Component

- [ ] **Track Playback**
  - [ ] Clicking a track card opens the player
  - [ ] Play/Pause button toggles playback state
  - [ ] Track info (title, artist) displays correctly
  - [ ] Audio actually plays (check volume/speakers)

- [ ] **Progress Bar**
  - [ ] Progress bar fills as track plays
  - [ ] Clicking progress bar seeks to that position (not restart)
  - [ ] Time display updates (current time / total duration)
  - [ ] Progress bar hover shows seek thumb

- [ ] **Waveform Visualizer**
  - [ ] EQ bars animate when track is playing
  - [ ] Clicking waveform seeks to position
  - [ ] Waveform shows in player bar

- [ ] **Volume Control**
  - [ ] Volume slider adjusts audio volume
  - [ ] Mute button toggles mute state
  - [ ] Volume persists between tracks

- [ ] **Quality Selector**
  - [ ] Quality dropdown opens/closes
  - [ ] Selecting different quality changes stream
  - [ ] Selected quality is remembered

- [ ] **Related Tracks**
  - [ ] Related tracks button opens panel
  - [ ] Related tracks load and display
  - [ ] Clicking related track plays it
  - [ ] Panel closes properly

- [ ] **Fullscreen Visualizer**
  - [ ] Clicking artwork opens fullscreen visualizer
  - [ ] Visualizer animates
  - [ ] Arrow keys cycle presets (next/prev)
  - [ ] 'R' key selects random preset
  - [ ] ESC key closes visualizer
  - [ ] Close button works
  - [ ] Stats display shows FPS

- [ ] **BPM Controls**
  - [ ] BPM controls visible in player
  - [ ] Playback rate can be adjusted
  - [ ] Vinyl spin animation matches playback rate

- [ ] **Close Button**
  - [ ] Close button stops playback
  - [ ] Close button clears current track
  - [ ] Player disappears

## Track Cards

- [ ] **Grid View**
  - [ ] Track cards render correctly
  - [ ] Artwork displays
  - [ ] Track title and artist show
  - [ ] Genre badge displays with correct color
  - [ ] Stats show (plays, favorites, BPM)
  - [ ] Duration shows
  - [ ] Hover effects work
  - [ ] Play button appears on hover

- [ ] **List View**
  - [ ] List view displays tracks correctly
  - [ ] Same info as grid view visible
  - [ ] Clicking plays track

- [ ] **Interactions**
  - [ ] Clicking card plays track
  - [ ] Favorite button adds/removes favorite
  - [ ] Flip button flips card (if applicable)
  - [ ] Add to playlist button works
  - [ ] Download link works (if track is downloadable)

- [ ] **Animations**
  - [ ] New tracks have highlight animation
  - [ ] Top artist glow animation works
  - [ ] Vinyl spin animation when playing
  - [ ] Now playing indicator animates

## Store / State Management

- [ ] **Play History**
  - [ ] Playing a track adds to history
  - [ ] History persists after refresh (localStorage)
  - [ ] History syncs to server (if authenticated)

- [ ] **Favorites**
  - [ ] Favoriting a track persists
  - [ ] Favorites show on track cards
  - [ ] Favorites page displays correctly

- [ ] **Filters**
  - [ ] Search/filter works
  - [ ] Sort options work
  - [ ] Duration filter works
  - [ ] Filters persist or reset appropriately

## Performance Checks

- [ ] **No Excessive Re-renders**
  - [ ] Player doesn't stutter during playback
  - [ ] Track cards don't constantly re-render
  - [ ] Visualizer maintains consistent FPS
  - [ ] No React warnings in console about re-renders

- [ ] **Tab Visibility**
  - [ ] Switching to another tab pauses visualizer (saves CPU)
  - [ ] Returning to tab resumes visualizer
  - [ ] Audio continues playing when tab hidden

- [ ] **Memory Usage**
  - [ ] No memory leaks after playing multiple tracks
  - [ ] Closing visualizer frees resources
  - [ ] Track list scrolls smoothly

## Browser Compatibility

- [ ] **Chrome/Edge**
- [ ] **Firefox**
- [ ] **Safari** (if applicable)

## Mobile/Responsive (if applicable)

- [ ] **Mobile Layout**
  - [ ] Player adapts to mobile screen
  - [ ] Track cards display correctly
  - [ ] Touch interactions work
  - [ ] Visualizer works on mobile

## Console Checks

Open browser dev tools and check:

- [ ] No React errors or warnings
- [ ] No audio playback errors
- [ ] No WebGL errors in visualizer
- [ ] Network requests succeed (200/206)
- [ ] No CORS errors

## Test Commands

Run automated tests:

```bash
# Run Playwright E2E tests
cd v2
npm run test:e2e

# Run specific test file
npx playwright test tests/performance-regression.spec.js

# Run with UI mode for debugging
npm run test:e2e:ui
```

## Known Issues / Limitations

Document any issues found during testing:

- 
- 
- 