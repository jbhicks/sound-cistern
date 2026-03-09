import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { useState, useEffect } from 'react'
import Layout from './components/Layout'
import Stream from './pages/Stream'
import Favorites from './pages/Favorites'
import Playlists from './pages/Playlists'
import Analytics from './pages/Analytics'
import DebugViz from './pages/DebugViz'
import { useStore } from './store'

export default function App() {
  const [loading, setLoading] = useState(true)
  const { user, setUser, setFavorites, loadPlaylists, loadPlayHistory } = useStore()

  useEffect(() => {
    // Try to silently refresh the session before checking auth.
    // This extends the cookie and rotates the SoundCloud token while the
    // session is still technically valid — preventing mid-session logouts.
    const tryRefresh = () =>
      fetch('/api/auth/refresh', { method: 'POST', credentials: 'include' })
        .then(r => r.ok)
        .catch(() => false)

    const loadUser = () =>
      fetch('/api/user', { credentials: 'include' })
        .then(res => res.ok ? res.json() : null)

    loadUser()
      .then(async data => {
        if (!data) {
          // 401 — try a token refresh then retry once
          const refreshed = await tryRefresh()
          if (refreshed) return loadUser()
          return null
        }
        // Session is valid — kick off a background refresh to reset the expiry
        // timer so active users never see an unexpected logout
        tryRefresh()
        return data
      })
      .then(data => {
        if (data) {
          setUser(data)
          return Promise.all([
            fetch('/api/favorites', { credentials: 'include' })
              .then(r => r.ok ? r.json() : { tracks: [] })
              .then(d => setFavorites((d.tracks || []).map(t => t.track_id))),
            loadPlaylists(),
            loadPlayHistory(),
          ])
        }
      })
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [setUser, setFavorites])

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-surface-950">
        <div className="flex flex-col items-center gap-4">
          <div className="w-12 h-12 border-4 border-accent border-t-transparent rounded-full animate-spin" />
          <p className="text-surface-400 text-sm">Loading Sound Cistern...</p>
        </div>
      </div>
    )
  }

  if (!user) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-surface-950">
        <div className="fixed inset-0 pointer-events-none">
          <div className="absolute top-0 left-1/4 w-96 h-96 bg-accent/10 rounded-full blur-3xl" />
          <div className="absolute bottom-0 right-1/4 w-96 h-96 bg-vapor-pink/10 rounded-full blur-3xl" />
        </div>
        <div className="relative text-center px-6">
          <div className="w-20 h-20 mx-auto mb-6 rounded-2xl bg-gradient-to-br from-accent to-vapor-pink flex items-center justify-center shadow-2xl shadow-accent/30">
            <svg className="w-10 h-10 text-white" viewBox="0 0 24 24" fill="currentColor">
              <path d="M12 3v10.55c-.59-.34-1.27-.55-2-.55-2.21 0-4 1.79-4 4s1.79 4 4 4 4-1.79 4-4V7h4V3h-6z"/>
            </svg>
          </div>
          <h1 className="text-4xl font-bold text-gradient mb-3">Sound Cistern</h1>
          <p className="text-surface-400 mb-8 text-lg">Your personal SoundCloud feed</p>
          <a
            href="/auth/soundcloud"
            className="inline-flex items-center gap-3 px-8 py-4 bg-accent hover:bg-accent-hover text-white rounded-2xl font-semibold text-lg transition-all duration-200 shadow-xl shadow-accent/30 hover:shadow-accent/50 hover:scale-105"
          >
            <svg className="w-6 h-6" viewBox="0 0 24 24" fill="currentColor">
              <path d="M11.56 8.87V17h8.76c.97 0 1.75-.78 1.75-1.75 0-.83-.57-1.53-1.35-1.71.01-.1.02-.21.02-.31 0-1.94-1.57-3.5-3.5-3.5-.05 0-.09 0-.14.01C16.55 8.23 15.12 7 13.38 7c-.96 0-1.82.38-2.46 1H11a.56.56 0 00-.56.56.57.57 0 00.12.31zm-1.93 8.13H4.19C3.42 17 2.8 16.38 2.8 15.61c0-.58.35-1.07.86-1.28-.01-.06-.01-.12-.01-.17 0-.87.7-1.57 1.57-1.57.1 0 .2.01.3.03C5.53 11.75 6.46 11 7.57 11c.32 0 .62.07.89.19.24-.11.5-.19.78-.19.82 0 1.52.52 1.78 1.25H11c.51 0 .93.42.93.93v3.82c0 .51-.42.93-.93.93h-.37z"/>
            </svg>
            Connect SoundCloud
          </a>
        </div>
      </div>
    )
  }

  return (
    <BrowserRouter basename="/">
      <Routes>
        {/* Debug route - isolated visualizer testing */}
        <Route path="/debug/viz" element={<DebugViz />} />
        
        {/* Main app routes */}
        <Route path="/*" element={
          <Layout>
            <Routes>
              <Route path="/" element={<Navigate to="/stream" replace />} />
              <Route path="/stream" element={<Stream />} />
              <Route path="/favorites" element={<Favorites />} />
              <Route path="/playlists" element={<Playlists />} />
              <Route path="/analytics" element={<Analytics />} />
            </Routes>
          </Layout>
        } />
      </Routes>
    </BrowserRouter>
  )
}
