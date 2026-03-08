import { create } from 'zustand'

// Debounced play history sync to PocketBase
let syncTimer = null
const SYNC_DELAY = 5000 // 5 seconds after last play

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

  // Shared Web Audio API nodes — created once by Player.jsx, read by visualizers
  audioCtx: null,
  analyserNode: null,

  filters: {
    q: '',
    sort: 'newest',
    duration_min: 0,
    limit: 100,
  },

  setUser: (user) => set({ user }),
  setTracks: (tracks) => set({ tracks }),
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
