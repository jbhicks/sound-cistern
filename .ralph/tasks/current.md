# Sound Cistern MVP+ - Ralph Tasks

## 🎉 ALL STORIES COMPLETE

**Session**: 2026-03-07
**Status**: ✅ DONE

---

## Completed This Session

### ✅ P0: Promote v2 React UI to Primary Route
- Root `/` redirects to `/v2/` (302)
- `BrowserRouter basename` updated to `"/"`
- `Makefile` has `v2-build` and `v2-dev` targets
- `AGENTS.md` updated — React/Vite/Tailwind stack, HTMX removed
- `sound-cistern-dev` skill updated with v2 dev workflow
- `pocketbase-htmx` and `sound-cistern-ux` skills marked ⚠️ DEPRECATED

### ✅ P1: Track Card Richness
- Duration badge on artwork (bottom-left)
- Play count + like count stats row
- Genre pill + BPM badge (conditional)
- Download indicator icon (conditional)
- List view enriched with plays/likes
- `downloadable` field added to Go API response

### ✅ P2: PWA / Offline Mode
- `v2/public/manifest.json` — installable PWA
- `v2/public/sw.js` — service worker (cache-first static, network-first API)
- `v2/src/components/OfflineIndicator.jsx` — animated offline banner
- SW registered in `main.jsx`
- Manifest linked in `index.html` with Apple PWA meta tags
- Icon placeholders in `v2/public/icons/` (PNGs need to be added)

### ✅ P3: Data Export
- `ExportFavoritesJSON` — real implementation, downloads JSON
- `ExportFavoritesCSV` — real implementation, downloads CSV
- `ExportPlaylistsJSON` — real implementation, nested JSON
- Export dropdown in Favorites page (JSON + CSV)
- Export button in Playlists page

### ✅ P4: Playlist System (Complete)
- All playlist handlers fully implemented (ListPlaylists, CreatePlaylist, GetPlaylist, AddTrackToPlaylist, RemoveTrackFromPlaylist, DeletePlaylist, RenamePlaylist, SharePlaylist, GetSharedPlaylist)
- Playlists page: rename (inline edit), share (modal + copy), detail view with tracks, remove from playlist
- TrackCard back face: "Add to Playlist" button with popover listing playlists
- Zustand store: `playlists`, `loadPlaylists` — loaded on app init

### ✅ P5: Related Tracks Explorer
- Go handler `relatedTracksHandler` with 1-hour in-memory cache
- Calls SoundCloud v2 API (`/tracks/{id}/related`), falls back to v1
- Player bar: Sparkles button toggles slide-up related tracks panel
- Panel shows 10 related tracks, click to play

### ✅ P6: Listening Analytics Dashboard
- Zustand `playHistory` (localStorage, max 500 entries)
- `recordPlay()` called on every new track play
- `/analytics` page: overview cards, top tracks, top artists, genre bar chart, recently played
- Analytics added to nav (BarChart2 icon)
- Stream page: recently-played horizontal strip (last 5 unique tracks)

### ✅ P7: RSS Feed
- `GetUserRSSFeed` — authenticated, generates RSS 2.0 XML from favorites
- `GetSharedRSSFeed` — public, uses PocketBase user ID as share token
- Favorites page: RSS button → modal with public URL + copy

### ✅ P8: Download Link Exposure
- `download_url` field added to Go API response + DB migration
- TrackCard download indicator is now a clickable link
- Player bar: Download button when `currentTrack.downloadable` is true

---

## Previously Completed (MVP)

### Phase 1–3: OAuth, Stream, Favorites, Search, UX Design System
### Phase 4: React v2 UI (Layout, Stream, Favorites, Playlists, Player, TrackCard, FilterPanel)

---

## Next Steps (Future Phases)

- Add real PWA icons (192px + 512px PNGs from the Zap/accent-purple logo)
- Timed comments viewer (deferred)
- Background sync for favorites when back online (SW placeholder exists)
- PocketBase sync for play history (currently localStorage only)
