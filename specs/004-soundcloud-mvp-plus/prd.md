# Feature Specification: Sound Cistern MVP+ (Phase 2)

**Feature Branch**: `004-soundcloud-mvp-plus`
**Created**: 2026-02-20
**Status**: In Progress
**Input**: Product design meeting with user

## Executive Summary

Sound Cistern MVP is complete with authentication, stream display, favorites, and search. This PRD covers Phase 2 features to enhance for **Pro/Music Industry users**.

### Product Design Decisions

| Question | Answer |
|----------|--------|
| Core Value | Better Filtering (MVP already done) |
| Target User | **Pro/Music Industry** (power users, DJs, producers) |
| #1 Priority | **PWA / Offline Mode** |
| Philosophy | **Fewer, Better** (polished over quantity) |
| Social | Share via links only (no follow system) |
| Technical | **Performance & Reliability** first |

---

## 🎯 Features Completed (MVP)

| Feature | Status |
|---------|--------|
| Soundcloud OAuth 2.1 + PKCE | ✅ Done |
| Login Splash Page | ✅ Done |
| Stream Display | ✅ Done |
| Favorites System | ✅ Done |
| Search Integration | ✅ Done |
| UX Design System | ✅ Done |
| Dark Mode | ✅ Done |
| Mobile Responsive | ✅ Done |

---

## 🚀 MVP+ Features (Prioritized)

### PRIORITY 1: PWA / Offline Mode (TOP)
**Rationale**: Pro users need offline access to cached tracks for mobile/DJ use

- [ ] Service Worker registration
- [ ] Cache stream tracks for offline playback
- [ ] Cache artwork/images
- [ ] Manifest.json for installability
- [ ] Offline indicator UI
- [ ] Background sync when back online

### PRIORITY 2: Data Export
**Rationale**: Pro users own their data - must be exportable

- [ ] Export favorites as JSON
- [ ] Export favorites as CSV  
- [ ] Export playlists (JSON)
- [ ] Download button in UI

### PRIORITY 3: Playlist System
**Rationale**: Pro users curate/organize music

- [ ] Create new playlist
- [ ] Name/rename playlists
- [ ] Add tracks from stream/favorites
- [ ] Reorder tracks in playlist
- [ ] Delete playlists
- [ ] Share playlist as link

### PRIORITY 4: Listening Analytics
**Rationale**: Pro users want insights into their listening

- [ ] Track play history (local storage)
- [ ] Most played tracks (top 10)
- [ ] Most played artists
- [ ] Genre distribution
- [ ] Listening time stats

### PRIORITY 5: RSS Feed (Share Only)
**Rationale**: Share curated feeds, no full social system

- [ ] Generate RSS feed URL for user
- [ ] Include track metadata
- [ ] Link to Soundcloud for playback
- [ ] Share button

---

## Out of Scope (v1)
- ❌ Follow system (full social)
- ❌ See friend's tracks in stream
- ❌ Comments on tracks
- ❌ Real-time sync
