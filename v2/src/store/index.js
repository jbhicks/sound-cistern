import { create } from 'zustand'

// Debounced play history sync to PocketBase
let syncTimer = null
const SYNC_DELAY = 5000 // 5 seconds after last play

// Genre color palette - predefined colors for consistent assignment
const GENRE_COLOR_PALETTE = [
  { bg: 'rgba(0, 212, 255, 0.35)', border: 'rgba(0, 212, 255, 0.5)', text: '#c7f3ff' },     // Cyan (primary accent)
  { bg: 'rgba(59, 130, 246, 0.35)', border: 'rgba(59, 130, 246, 0.5)', text: '#dbeafe' },   // Blue
  { bg: 'rgba(16, 185, 129, 0.35)', border: 'rgba(16, 185, 129, 0.5)', text: '#d1fae5' },   // Green
  { bg: 'rgba(245, 158, 11, 0.35)', border: 'rgba(245, 158, 11, 0.5)', text: '#fef3c7' },   // Amber
  { bg: 'rgba(239, 68, 68, 0.35)', border: 'rgba(239, 68, 68, 0.5)', text: '#fee2e2' },     // Red
  { bg: 'rgba(236, 72, 153, 0.35)', border: 'rgba(236, 72, 153, 0.5)', text: '#fce7f3' },   // Pink
  { bg: 'rgba(14, 165, 233, 0.35)', border: 'rgba(14, 165, 233, 0.5)', text: '#e0f2fe' },   // Sky
  { bg: 'rgba(139, 92, 246, 0.35)', border: 'rgba(139, 92, 246, 0.5)', text: '#e9d5ff' },   // Purple (moved down)
  { bg: 'rgba(249, 115, 22, 0.35)', border: 'rgba(249, 115, 22, 0.5)', text: '#ffedd5' },   // Orange
  { bg: 'rgba(20, 184, 166, 0.35)', border: 'rgba(20, 184, 166, 0.5)', text: '#ccfbf1' },   // Teal
  { bg: 'rgba(99, 102, 241, 0.35)', border: 'rgba(99, 102, 241, 0.5)', text: '#e0e7ff' },   // Indigo
  { bg: 'rgba(234, 179, 8, 0.35)', border: 'rgba(234, 179, 8, 0.5)', text: '#fef9c3' },     // Yellow
]

// Known genre mappings for consistent colors
const KNOWN_GENRE_COLORS = {
  'electronic': 0, 'house': 0, 'techno': 0, 'trance': 0, 'dubstep': 0, 'edm': 0, 'ambient': 0,
  'drum & bass': 1, 'drum and bass': 1, 'dnb': 1, 'drum\'n\'bass': 1,
  'hip hop': 2, 'hip-hop': 2, 'rap': 2, 'trap': 2, 'r&b': 2, 'rb': 2,
  'rock': 3, 'alternative': 3, 'indie rock': 3, 'metal': 3, 'punk': 3,
  'pop': 4, 'dance pop': 4, 'indie pop': 4, 'synth pop': 4,
  'jazz': 5, 'soul': 5, 'funk': 5, 'blues': 5,
  'folk': 6, 'acoustic': 6, 'country': 6, 'indie': 6,
  'classical': 7, 'orchestral': 7, 'piano': 7,
  'reggae': 8, 'dancehall': 8, 'ska': 8,
  'latin': 9, 'salsa': 9, 'reggaeton': 9, 'bachata': 9,
  'world': 10, 'ethnic': 10, 'traditional': 10,
  'comedy': 11, 'spoken word': 11, 'podcast': 11,
}

async function syncPlayHistoryToPocketBase(entries) {
  if (!entries || entries.length === 0) return
  try {
    await fetch('/api/play-history', {
      method: 'POST',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(entries),
    })
  } catch {
    // silently ignore — localStorage is the source of truth
  }
}

export const useStore = create((set, get) => ({
  user: null,
  tracks: [],
  favorites: new Set(),
  playlists: [],

  // Play history — stored in localStorage, max 500 entries
  // Also synced to PocketBase when user is authenticated
  playHistory: JSON.parse(localStorage.getItem('sc-play-history') || '[]'),
  // Track which entries have been synced to PocketBase (by played_at timestamp)
  _syncedPlayedAts: new Set(JSON.parse(localStorage.getItem('sc-synced-played-ats') || '[]')),

  // Player state
  currentTrack: null,
  isPlaying: false,
  audioEl: null,
  streamQuality: localStorage.getItem('sc-stream-quality') || 'auto',
  playbackRate: 1,

  // Shared Web Audio API nodes — created once by Player.jsx, read by visualizers
  audioCtx: null,
  analyserNode: null,

  filters: {
    q: '',
    sort: 'newest',
    duration_min: 0,
    limit: 100,
  },

  // Genre colors - Map of normalized genre names to color objects
  genreColors: new Map(),

  setUser: (user) => set({ user }),
  
  setTracks: (tracks) => {
    const { genreColors } = get()
    const newColors = new Map(genreColors)
    let colorIndex = 0
    
    // Collect all unique genres from tracks
    const genres = new Set()
    tracks.forEach(track => {
      if (track.genre) {
        genres.add(track.genre.toLowerCase().trim())
      }
    })
    
    // Assign colors to new genres
    genres.forEach(genre => {
      if (!newColors.has(genre)) {
        // Check if it's a known genre
        const knownIndex = KNOWN_GENRE_COLORS[genre]
        if (knownIndex !== undefined) {
          newColors.set(genre, GENRE_COLOR_PALETTE[knownIndex])
        } else {
          // Assign next available color, cycling through palette
          newColors.set(genre, GENRE_COLOR_PALETTE[colorIndex % GENRE_COLOR_PALETTE.length])
          colorIndex++
        }
      }
    })
    
    set({ tracks, genreColors: newColors })
  },
  
  getGenreColor: (genre) => {
    if (!genre) return GENRE_COLOR_PALETTE[0]
    const { genreColors } = get()
    const normalized = genre.toLowerCase().trim()
    return genreColors.get(normalized) || GENRE_COLOR_PALETTE[0]
  },
  
  setPlaylists: (playlists) => set({ playlists }),
  setFavorites: (ids) => set({ favorites: new Set(ids) }),

  recordPlay: (track) => {
    const { playHistory, _syncedPlayedAts } = get()
    const entry = {
      track_id: track.track_id,
      track_title: track.track_title,
      artist_name: track.artist_name,
      artwork_url: track.artwork_url,
      track_duration: track.track_duration,
      genre: track.genre || '',
      played_at: new Date().toISOString(),
    }
    const next = [entry, ...playHistory].slice(0, 500)
    localStorage.setItem('sc-play-history', JSON.stringify(next))
    set({ playHistory: next })

    // Debounce-sync to PocketBase: collect unsent entries and flush after idle
    clearTimeout(syncTimer)
    syncTimer = setTimeout(() => {
      const { playHistory: current, _syncedPlayedAts: synced } = get()
      const unsent = current.filter(e => !synced.has(e.played_at))
      if (unsent.length === 0) return
      syncPlayHistoryToPocketBase(unsent).then(() => {
        const nextSynced = new Set([...synced, ...unsent.map(e => e.played_at)])
        // Keep synced set bounded to last 500 entries
        const syncedArr = [...nextSynced].slice(-500)
        localStorage.setItem('sc-synced-played-ats', JSON.stringify(syncedArr))
        set({ _syncedPlayedAts: new Set(syncedArr) })
      })
    }, SYNC_DELAY)
  },

  // Load play history from PocketBase and merge with localStorage
  loadPlayHistory: async () => {
    try {
      const res = await fetch('/api/play-history?limit=500', { credentials: 'include' })
      if (!res.ok) return
      const data = await res.json()
      const remote = data.entries || []
      if (remote.length === 0) return

      const { playHistory: local } = get()
      // Merge: remote entries not already in local (by played_at)
      const localAts = new Set(local.map(e => e.played_at))
      const merged = [...local]
      for (const e of remote) {
        if (!localAts.has(e.played_at)) merged.push(e)
      }
      // Sort newest first, cap at 500
      merged.sort((a, b) => new Date(b.played_at) - new Date(a.played_at))
      const final = merged.slice(0, 500)
      localStorage.setItem('sc-play-history', JSON.stringify(final))
      // Mark all remote entries as synced
      const syncedArr = [...new Set([...final.map(e => e.played_at)])].slice(-500)
      localStorage.setItem('sc-synced-played-ats', JSON.stringify(syncedArr))
      set({ playHistory: final, _syncedPlayedAts: new Set(syncedArr) })
    } catch {
      // silently ignore
    }
  },

  loadPlaylists: async () => {
    try {
      const res = await fetch('/api/playlists', { credentials: 'include' })
      if (res.ok) {
        const data = await res.json()
        set({ playlists: Array.isArray(data) ? data : [] })
      }
    } catch {
      // silently ignore
    }
  },

  setFilter: (key, value) => set((state) => ({
    filters: { ...state.filters, [key]: value }
  })),

  resetFilters: () => set({
    filters: { q: '', sort: 'newest', duration_min: 0, limit: 100 }
  }),

  // Play a track - sets currentTrack and kicks off audio
  playTrack: (track) => {
    const { audioEl, currentTrack, recordPlay } = get()
    if (currentTrack?.track_id === track.track_id) {
      // Toggle play/pause on same track — don't record
      if (audioEl) {
        if (audioEl.paused) {
          audioEl.play()
          set({ isPlaying: true })
        } else {
          audioEl.pause()
          set({ isPlaying: false })
        }
      }
      return
    }
    recordPlay(track)
    set({ currentTrack: track, isPlaying: true })
  },

  setAudioEl: (el) => set({ audioEl: el }),
  setAudioContext: (audioCtx, analyserNode) => set({ audioCtx, analyserNode }),
  setIsPlaying: (isPlaying) => set({ isPlaying }),
  setStreamQuality: (quality) => {
    localStorage.setItem('sc-stream-quality', quality)
    set({ streamQuality: quality })
  },
  setPlaybackRate: (rate) => set({ playbackRate: rate }),

  toggleFavorite: async (trackId) => {
    const { favorites } = get()
    const isFav = favorites.has(trackId)
    // Optimistic update
    const next = new Set(favorites)
    if (isFav) next.delete(trackId)
    else next.add(trackId)
    set({ favorites: next })

    try {
      const res = await fetch('/api/favorites/toggle', {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ track_id: trackId }),
      })
      if (!res.ok) {
        set({ favorites })
      }
    } catch {
      set({ favorites })
    }
  },
}))
