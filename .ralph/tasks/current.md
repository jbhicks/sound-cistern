# Sound Cistern MVP+ - Ralph Tasks

## Current Task: pwa_offline_mode
**Status**: 🔄 IN PROGRESS
**Priority**: P1 (TOP)
**Started**: 2026-02-20

### Completion Promise
Pro users can access their Soundcloud stream offline via PWA, export their data, and create playlists

### Success Criteria (4 Features) - In Progress

- [ ] **Story 1: PWA Installability**
  - Service Worker registration
  - manifest.json for installability  
  - Offline indicator UI
  - Verification: `pwa_install_test`

- [ ] **Story 2: Offline Track Access**
  - Cache stream tracks for offline playback
  - Cache artwork/images
  - Background sync when back online
  - Verification: `offline_playback_test`

- [ ] **Story 3: Data Export**
  - Export favorites as JSON
  - Export favorites as CSV
  - Export playlists as JSON
  - Download buttons in UI
  - Verification: `data_export_test`

- [ ] **Story 4: Playlist System**
  - Create new playlist
  - Name/rename playlists
  - Add tracks from stream/favorites
  - Reorder tracks in playlist
  - Delete playlists
  - Share playlist as link
  - Verification: `playlist_test`

---

## Completed Tasks (MVP)

### Phase 1: OAuth & Stream (Session 002)
- [x] OAuth Authentication (Soundcloud 2.1 + PKCE)
- [x] Login Splash Page
- [x] OAuth Callback Handling
- [x] Stream Display
- [x] Session Management

### Phase 2: Core Features (Previous Sessions)
- [x] Favorites System
- [x] Search Integration  
- [x] Dark Mode Theme
- [x] Mobile Responsive Design
- [x] Soundcloud Stream Integration

### Phase 3: UX Design System (Session UX_001)
- [x] Design Tokens (spacing, colors, typography)
- [x] Global Layout & Navigation
- [x] Authentication Pages UX
- [x] Stream Page UX
- [x] Favorites Page UX
- [x] Blog Pages UX
- [x] Interactive Components & HTMX
- [x] Accessibility Audit (WCAG 2.1 AA)

---

## Backlog (Future Phases)

### Priority 4: Listening Analytics
- Track play history (local storage)
- Most played tracks (top 10)
- Most played artists
- Genre distribution
- Listening time stats

### Priority 5: RSS Feed (Share Only)
- Generate RSS feed URL for user
- Include track metadata
- Link to Soundcloud for playback
- Share button

---

## Implementation Notes

### Dependencies
- OAuth must be working (✅ Done)
- Stream display must be working (✅ Done)
- Favorites system must be working (✅ Done)

### Technical Approach
1. PWA first - critical for pro users
2. Service Worker with Cache API
3. IndexedDB for track metadata offline
4. Background Sync API for when back online

### Constraints
- Fewer features, but polished
- Performance & Reliability first
- Share via links only (no full social)

---

## Ralph Loop State
- **Budget**: $20.00 total
- **Iterations**: 30 max
- **Current Focus**: PWA/Offline Mode
- **Next**: Data Export → Playlist → Analytics → RSS
