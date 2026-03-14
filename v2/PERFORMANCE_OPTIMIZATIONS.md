# React Performance Optimizations - Summary

This document summarizes all the performance optimizations applied to fix the performance issues, especially while the fullscreen visualizer is running.

## Changes Made

### 1. Player.jsx - Throttled Progress Updates

**Problem:** Progress state was updating 4-10 times per second, causing the entire component tree to re-render constantly.

**Solution:**
- Use refs (`progressRef`, `currentTimeRef`, `durationRef`) for high-frequency updates (no React re-renders)
- Throttle state updates to 4fps (250ms) using `throttleTimeoutRef`
- Use `displayProgress`, `displayCurrentTime`, `displayDuration` for UI display only
- Pre-defined animation configs to avoid creating new objects every render
- Added `memo()` wrapper to prevent unnecessary re-renders
- Use selective store subscriptions with selectors to prevent re-renders on unrelated state changes

**Impact:** Reduced React re-renders from 10/sec to 4/sec during playback.

### 2. Store Subscriptions - Selectors Pattern

**Problem:** Components were subscribing to the entire store, causing re-renders on ANY state change.

**Solution:**
Updated all components to use selector pattern:
```javascript
// Before (bad)
const { analyserNode, audioCtx } = useStore()

// After (good)
const analyserNode = useStore(state => state.analyserNode)
const audioCtx = useStore(state => state.audioCtx)
```

**Files Changed:**
- `Player.jsx`
- `ButterchurnVisualizer.jsx` (both Mini and Fullscreen)
- `EQVisualizer.jsx`
- `TrackCard.jsx`

**Impact:** Components only re-render when their specific subscribed state changes.

### 3. TrackCard.jsx - Memoization

**Problem:** All visible TrackCards re-rendered on every progress tick with expensive Framer Motion layout animations.

**Solution:**
- Wrapped `TrackCard` with `React.memo()` and custom comparison function
- Custom comparison only checks: `track_id`, `viewMode`, `isNew`, `isTopArtist`
- Added `useMemo` for genre style calculation
- Use selectors instead of destructuring entire store

**Impact:** TrackCards only re-render when their essential props change, not on every store update.

### 4. ButterchurnFullscreen.jsx - Split Effects & Visibility Handling

**Problem:** 
- Single massive useEffect with too many dependencies caused unnecessary re-initializations
- RAF loop continued running even when tab was hidden

**Solution:**
- Split initialization useEffect into two focused effects:
  1. Effect 1: Handle open/close state (cleanup RAF when closed)
  2. Effect 2: Initialize visualizer and start RAF loop (only depends on `open` and `canvasKey`)
- Added `visibilitychange` handler to pause RAF when tab is hidden
- Callbacks already use `useCallback` with empty deps (stable references)

**Impact:** 
- Visualizer doesn't re-initialize when unrelated props change
- Saves CPU when user switches to another tab

### 5. EQVisualizer.jsx - Visibility Handling

**Problem:** RAF loop continued running even when tab was hidden.

**Solution:**
- Added `visibilitychange` handler
- When `document.hidden` is true: cancel RAF loop
- When tab becomes visible again: restart RAF loop
- Uses closure variable `isVisible` to track state

**Impact:** Saves CPU when tab is not visible.

## Testing

### Automated Tests
Created Playwright E2E tests in `v2/tests/performance-regression.spec.js` covering:
- Player core functionality (playback, seeking, volume)
- Visualizer functionality (fullscreen, preset cycling)
- Track card interactions
- Performance verification (no excessive re-renders)
- Tab visibility handling

### Manual Testing Checklist
Created `v2/tests/REGRESSION_CHECKLIST.md` with comprehensive checklist for:
- Player functionality
- Track card interactions
- Store/state management
- Performance checks
- Browser compatibility
- Console error checks

## Build Verification

All builds pass successfully:
- ✅ Frontend build: `npm run build` (Vite)
- ✅ Backend build: `go build`

## Performance Improvements

Expected improvements:
1. **Reduced Re-renders:** From ~10/sec to ~4/sec during playback
2. **Better Memory Usage:** Components don't re-render unnecessarily
3. **CPU Savings:** Visualizer pauses when tab hidden
4. **Smoother Animations:** Less React reconciliation blocking the main thread

## Backward Compatibility

All changes are backward compatible:
- No API changes
- No prop changes
- No breaking changes to component interfaces
- Existing functionality preserved

## Files Modified

1. `v2/src/components/Player.jsx`
2. `v2/src/components/ButterchurnVisualizer.jsx`
3. `v2/src/components/EQVisualizer.jsx`
4. `v2/src/components/TrackCard.jsx`

## Files Added

1. `v2/tests/performance-regression.spec.js` - Automated E2E tests
2. `v2/tests/REGRESSION_CHECKLIST.md` - Manual testing checklist
3. `PERFORMANCE_OPTIMIZATIONS.md` - This document