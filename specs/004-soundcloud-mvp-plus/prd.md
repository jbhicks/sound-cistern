# Feature Specification: Sound Cistern MVP+ (Phase 2)

**Feature Branch**: `004-soundcloud-mvp-plus`
**Created**: 2026-02-20
**Updated**: 2026-03-07
**Status**: In Progress
**Input**: Product design meeting with user; external PRD review (Grok AI suggestions)

## Executive Summary

Sound Cistern MVP is complete with authentication, stream display, favorites, and search. This PRD covers Phase 2 features to enhance for **Pro/Music Industry users**.

A React v2 UI has been built at `/v2/` (React 18 + Vite + Tailwind + Zustand + Framer Motion) and is ready to become the primary interface. This phase promotes v2 to the main route, retires the HTMX/Templ layer from active development, and adds the highest-value discovery and UX features identified from both internal design work and external PRD review.

### Product Design Decisions

| Question | Answer |
|----------|--------|
| Core Value | Better Filtering (MVP already done) |
| Target User | **Pro/Music Industry** (power users, DJs, producers) |
| #1 Priority | **v2 React UI as primary route** |
| Philosophy | **Fewer, Better** (polished over quantity) |
| Social | Share via links only (no follow system) |
| Technical | **Performance & Reliability** first |
| Frontend Stack | **React 18 + Vite + Tailwind + Zustand + Framer Motion** (v2) |

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
| React v2 UI at `/v2/` | ✅ Done |

---

## 🚀 MVP+ Features (Prioritized)

---

### PRIORITY 0: Promote v2 React UI to Primary Route
**Rationale**: The React v2 UI is complete and superior to the HTMX/Templ layer. It should be the default experience. The old Templ routes should be preserved temporarily as a fallback but removed from active development and documentation.

**User Stories**:
- As a user, I want to land on the modern React UI by default so I get the best experience without navigating to `/v2/`
- As a developer, I want documentation and agent skills to reflect the actual tech stack so I don't get confused by outdated HTMX/Templ references

**Tasks**:
- [ ] Change Go backend root route (`/`) to redirect to `/v2/` (or serve v2 directly)
- [ ] Update `BrowserRouter basename` in `v2/src/App.jsx` from `/v2` to `/` once root is remapped
- [ ] Update `vite.config.js` `base` from `/v2/` to `/` and `outDir` accordingly
- [ ] Update Go static file serving to serve built v2 assets from root
- [ ] Keep old Templ routes accessible at `/legacy/` for rollback safety (temporary)
- [ ] Update `AGENTS.md` — remove HTMX references, update stack to React/Vite/Tailwind/Zustand
- [ ] Update `sound-cistern-dev` skill — remove Templ/HTMX references, add v2 dev workflow (`cd v2 && npm run dev`)
- [ ] Update `sound-cistern-ux` skill — replace Pico.css/HTMX/Templ patterns with Tailwind/React/Framer Motion patterns
- [ ] Archive or clearly mark `pocketbase-htmx` skill as legacy/deprecated
- [ ] Update `Makefile` to include `v2-build` and `v2-dev` targets
- [ ] Update `README.md` to reflect current stack

---

### PRIORITY 1: Track Card Richness — Restore Original Information Density
**Rationale**: The v2 `TrackCard` is visually polished but informationally sparse compared to the original proto designs. The originals showed plays, likes, reposts, comments, genre badges, BPM tags, duration badges, and download indicators — all the metadata a pro user needs to evaluate a track at a glance. The v2 card currently shows only title and artist on the front face, hiding everything behind a flip interaction. The flip mechanic is clever but shouldn't be the only way to see basic stats.

**Reference**: Original proto designs in `views/components/track_card_protos.templ` — specifically `TrackCardGradient1`–`TrackCardGradient10` (full-artwork overlay with stats), `TrackCardVerticalProto4` (artwork + overlay with plays/likes/genre/BPM), and `TrackCardBento` (structured stats grid).

**User Stories**:
- As a listener, I want to see play count, like count, genre, and duration on the card front so I can evaluate tracks without extra clicks
- As a DJ, I want BPM visible on the card so I can quickly identify tracks that fit my set
- As a user, I want a duration badge on the artwork so I can instantly see if a track is a mix-length set
- As a user, I want the card to feel as rich as the original proto designs while keeping the modern v2 aesthetic

**Tasks**:
- [ ] Add duration badge overlay to artwork (bottom-left, pill style) on card front face
- [ ] Add play count and like count stats row below artist name on card front
- [ ] Add genre pill/badge on card front (conditionally rendered when genre exists)
- [ ] Add BPM badge on card front (conditionally rendered when BPM > 0)
- [ ] Add download indicator icon when `downloadable: true` in track metadata
- [ ] Preserve the vinyl flip-back as an optional detail view (keep the mechanic, just enrich the front)
- [ ] Ensure list-view row also shows duration, plays, likes, and genre (already has duration — add stats)
- [ ] Expose `playback_count`, `favoritings_count`, `reposts_count`, `comment_count`, `bpm`, `downloadable` fields in the `/api/stream` and `/api/favorites` JSON responses if not already present

---

### PRIORITY 2: PWA / Offline Mode
**Rationale**: Pro users need offline access to cached tracks for mobile/DJ use

- [ ] Service Worker registration
- [ ] Cache stream tracks for offline playback
- [ ] Cache artwork/images
- [ ] Manifest.json for installability
- [ ] Offline indicator UI
- [ ] Background sync when back online

---

### PRIORITY 3: Data Export
**Rationale**: Pro users own their data - must be exportable

- [ ] Export favorites as JSON
- [ ] Export favorites as CSV
- [ ] Export playlists (JSON)
- [ ] Download button in UI

---

### PRIORITY 4: Playlist System
**Rationale**: Pro users curate/organize music. Basic playlist creation already exists in v2 (`Playlists.jsx`) — this story completes it.

- [ ] Add tracks from stream/favorites to a playlist (currently missing)
- [ ] Reorder tracks within a playlist (drag-and-drop)
- [ ] Rename playlists
- [ ] Share playlist as a public link
- [ ] Playlist detail page showing tracks with full card UI

---

### PRIORITY 5: Related Tracks Explorer
**Rationale**: Core to the mix discovery mission. SoundCloud's `/tracks/{id}/related` endpoint returns similar tracks. Surfacing these filtered by duration gives listeners a "more like this" discovery path that SoundCloud's own UI handles poorly.

**User Stories**:
- As a listener, I want to see tracks similar to one I'm enjoying, filtered to mix-length, so I can keep discovering without leaving the app

**Tasks**:
- [ ] Add "Related" button/section to track card (or player bar)
- [ ] Backend: proxy `/api/track/:id/related` → SoundCloud `/tracks/{id}/related`
- [ ] Frontend: render related tracks in a slide-out panel or inline section
- [ ] Apply existing duration filters to related results
- [ ] Cache related results in PocketBase to avoid repeated API calls

---

### PRIORITY 6: Listening Analytics Dashboard
**Rationale**: Pro users want insights into their listening. Reframe as a dashboard that combines analytics with personalized content — addresses both the internal "Listening Analytics" backlog item and the external PRD's "Personalized Dashboard" suggestion.

- [ ] Track play history (localStorage + PocketBase sync)
- [ ] Most played tracks (top 10 list)
- [ ] Most played artists
- [ ] Genre distribution chart
- [ ] Total listening time stats
- [ ] "Recently Played" section on stream page
- [ ] Dashboard page (`/analytics`) surfacing all of the above

---

### PRIORITY 7: RSS Feed (Share Only)
**Rationale**: Share curated feeds, no full social system. Also covers the external PRD's "Third-Party Integrations" use case (smart home, podcast apps) without building anything custom.

- [ ] Generate RSS feed URL for user's favorites or a playlist
- [ ] Include track metadata (title, artist, duration, artwork, permalink)
- [ ] Link to Soundcloud for playback (no direct audio in RSS)
- [ ] Share button in UI

---

### PRIORITY 8: Download Link Exposure
**Rationale**: Low-effort win — track metadata already stored in PocketBase includes `downloadable` flag. UI-only change.

- [ ] Show download icon/button on cards and player bar when `downloadable: true`
- [ ] Link directly to SoundCloud download URL from track metadata
- [ ] No backend changes required

---

## Deferred / Out of Scope

| Item | Reason |
|------|--------|
| Follow system (full social) | Out of scope — share via links only |
| See friends' tracks in stream | Out of scope |
| Timed comments viewer | Interesting but low priority; revisit post-Analytics |
| Upload/edit helpers | Requires SC write API; not our primary persona |
| Track analytics for owned tracks | DJ/uploader persona; not primary |
| Real-time sync | Performance cost outweighs benefit |
| Crossfade / advanced audio | Scope creep; basic queue sufficient |
| Genre/tag aggregation view | Covered by existing search + filter panel |

---

## Tech Stack Reference (Current)

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.23 + PocketBase v0.22.0 + SQLite |
| Frontend | React 18 + Vite + Tailwind CSS v3 |
| State | Zustand |
| Animation | Framer Motion |
| Icons | Lucide React |
| Auth | SoundCloud OAuth 2.1 + PKCE |
| Deployment | Cloudflare Tunnel → localhost:8090 (prod = dev) |

> **Note**: Templ/HTMX/Pico.css are legacy. Do not use for new features. The `views/` directory and `pocketbase-htmx` skill are preserved for reference only.
