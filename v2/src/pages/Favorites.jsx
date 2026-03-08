import { useState, useEffect, useRef, useCallback } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { Heart, Grid, List, Download, Rss, Copy, Check, X, RefreshCw } from 'lucide-react'
import clsx from 'clsx'
import TrackCard from '../components/TrackCard'
import { useStore } from '../store'

const CACHE_KEY = 'sc-favorites-cache'

function readCache() {
  try { return JSON.parse(localStorage.getItem(CACHE_KEY) || '[]') } catch { return [] }
}
function writeCache(tracks) {
  try { localStorage.setItem(CACHE_KEY, JSON.stringify(tracks)) } catch {}
}

export default function Favorites() {
  const cached = useState(() => readCache())[0]
  const [tracks, setTracks] = useState(() => cached)
  const [loading, setLoading] = useState(cached.length === 0)
  const [syncing, setSyncing] = useState(false)
  const [viewMode, setViewMode] = useState('grid')
  const [exportOpen, setExportOpen] = useState(false)
  const [rssModalOpen, setRssModalOpen] = useState(false)
  const [rssCopied, setRssCopied] = useState(false)
  const exportRef = useRef(null)
  const newTrackIdsRef = useRef(new Set())
  const tracksRef = useRef(cached)
  const { favorites, user } = useStore()

  const publicRssUrl = user ? `${window.location.origin}/feed/rss/${user.id}` : null

  // Merge incoming tracks into current list, flag genuinely new ones
  const mergeAndSet = useCallback((incoming) => {
    const cachedIds = new Set(tracksRef.current.map(t => t.track_id))
    const brandNew = new Set(
      incoming.filter(t => !cachedIds.has(t.track_id)).map(t => t.track_id)
    )
    // New tracks at front, then retained tracks not in this batch
    const incomingIds = new Set(incoming.map(t => t.track_id))
    const retained = tracksRef.current.filter(t => !incomingIds.has(t.track_id))
    const merged = [...incoming, ...retained]
    newTrackIdsRef.current = brandNew
    tracksRef.current = merged
    writeCache(merged)
    setTracks(merged)
  }, [])

  // Load from DB on mount — fast, no SoundCloud call
  useEffect(() => {
    fetch('/api/favorites', {
      credentials: 'include',
      headers: { Accept: 'application/json' },
    })
      .then(r => r.ok ? r.json() : { tracks: [] })
      .then(d => mergeAndSet(d.tracks || []))
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [mergeAndSet])

  // Background sync from SoundCloud after DB load settles
  useEffect(() => {
    if (loading) return
    setSyncing(true)
    fetch('/api/favorites/sync', {
      method: 'POST',
      credentials: 'include',
      headers: { Accept: 'application/json' },
    })
      .then(r => r.ok ? r.json() : null)
      .then(d => { if (d?.tracks) mergeAndSet(d.tracks) })
      .catch(() => {})
      .finally(() => setSyncing(false))
  }, [loading, mergeAndSet])

  // Manual re-sync button
  const handleSync = () => {
    if (syncing) return
    setSyncing(true)
    fetch('/api/favorites/sync', {
      method: 'POST',
      credentials: 'include',
      headers: { Accept: 'application/json' },
    })
      .then(r => r.ok ? r.json() : null)
      .then(d => { if (d?.tracks) mergeAndSet(d.tracks) })
      .catch(() => {})
      .finally(() => setSyncing(false))
  }

  // Close dropdown when clicking outside
  useEffect(() => {
    if (!exportOpen) return
    const handler = (e) => {
      if (exportRef.current && !exportRef.current.contains(e.target)) {
        setExportOpen(false)
      }
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [exportOpen])

  const handleCopyRss = () => {
    if (!publicRssUrl) return
    navigator.clipboard.writeText(publicRssUrl).then(() => {
      setRssCopied(true)
      setTimeout(() => setRssCopied(false), 2000)
    })
  }

  const handleExport = (format) => {
    setExportOpen(false)
    const url = format === 'csv'
      ? '/api/export/favorites/csv'
      : '/api/export/favorites/json'
    const a = document.createElement('a')
    a.href = url
    a.download = format === 'csv' ? 'favorites.csv' : 'favorites.json'
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
  }

  return (
    <div className="min-h-screen px-4 py-6 md:px-6 lg:px-8">
      <div className="max-w-screen-2xl mx-auto">
        <motion.div
          initial={{ opacity: 0, y: -16 }}
          animate={{ opacity: 1, y: 0 }}
          className="mb-6 flex items-center justify-between"
        >
          <div>
            <h1 className="text-2xl font-bold text-gradient">Favorites</h1>
            <p className="text-surface-500 text-sm mt-0.5">
              {syncing ? 'Syncing from SoundCloud\u2026' : tracks.length > 0 ? `${tracks.length} tracks` : 'Tracks you\u2019ve loved'}
            </p>
          </div>
          <div className="flex items-center gap-2">
            {/* Sync button */}
            <button
              onClick={handleSync}
              disabled={syncing}
              className={clsx(
                'flex items-center gap-1.5 px-3 py-2 rounded-xl border text-sm transition-colors',
                syncing
                  ? 'border-accent/40 bg-accent/10 text-accent-light cursor-not-allowed'
                  : 'border-surface-700/50 bg-surface-800/60 text-surface-400 hover:text-surface-200 hover:border-surface-600/60'
              )}
              title="Sync from SoundCloud"
            >
              <RefreshCw className={clsx('w-4 h-4', syncing && 'animate-spin')} />
              <span className="hidden sm:inline">Sync</span>
            </button>

            {/* RSS feed button */}
            <button
              onClick={() => setRssModalOpen(true)}
              className="flex items-center gap-1.5 px-3 py-2 rounded-xl border border-surface-700/50 bg-surface-800/60 text-surface-400 hover:text-orange-400 hover:border-orange-500/40 text-sm transition-colors"
              title="RSS Feed"
            >
              <Rss className="w-4 h-4" />
              <span className="hidden sm:inline">RSS</span>
            </button>

            {/* Export dropdown */}
            <div className="relative" ref={exportRef}>
              <button
                onClick={() => setExportOpen(prev => !prev)}
                className="flex items-center gap-1.5 px-3 py-2 rounded-xl border border-surface-700/50 bg-surface-800/60 text-surface-400 hover:text-surface-200 hover:border-surface-600/60 text-sm transition-colors"
              >
                <Download className="w-4 h-4" />
                <span className="hidden sm:inline">Export</span>
              </button>
              <AnimatePresence>
                {exportOpen && (
                  <motion.div
                    initial={{ opacity: 0, scale: 0.95, y: -4 }}
                    animate={{ opacity: 1, scale: 1, y: 0 }}
                    exit={{ opacity: 0, scale: 0.95, y: -4 }}
                    transition={{ duration: 0.1 }}
                    className="absolute right-0 mt-1.5 w-36 bg-surface-900 border border-surface-700/60 rounded-xl shadow-2xl overflow-hidden z-20"
                  >
                    <button
                      onClick={() => handleExport('json')}
                      className="w-full px-4 py-2.5 text-left text-sm text-surface-300 hover:bg-surface-800 hover:text-surface-100 transition-colors"
                    >
                      Export JSON
                    </button>
                    <button
                      onClick={() => handleExport('csv')}
                      className="w-full px-4 py-2.5 text-left text-sm text-surface-300 hover:bg-surface-800 hover:text-surface-100 transition-colors border-t border-surface-700/40"
                    >
                      Export CSV
                    </button>
                  </motion.div>
                )}
              </AnimatePresence>
            </div>

            {/* View mode toggle */}
            <div className="flex items-center rounded-xl border border-surface-700/50 overflow-hidden bg-surface-800/60">
              <button
                onClick={() => setViewMode('grid')}
                className={clsx('p-2.5 transition-colors', viewMode === 'grid' ? 'bg-accent/20 text-accent-light' : 'text-surface-500 hover:text-surface-300')}
              >
                <Grid className="w-4 h-4" />
              </button>
              <button
                onClick={() => setViewMode('list')}
                className={clsx('p-2.5 transition-colors', viewMode === 'list' ? 'bg-accent/20 text-accent-light' : 'text-surface-500 hover:text-surface-300')}
              >
                <List className="w-4 h-4" />
              </button>
            </div>
          </div>
        </motion.div>

        {loading ? (
          <div className={clsx(
            viewMode === 'grid'
              ? 'grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 2xl:grid-cols-7 gap-3'
              : 'space-y-1'
          )}>
            {Array.from({ length: 8 }).map((_, i) => (
              <div key={i} className="rounded-2xl overflow-hidden bg-surface-800/40 border border-surface-700/30">
                <div className="aspect-square animate-pulse bg-surface-700/50" />
                <div className="p-3 space-y-2">
                  <div className="h-3.5 w-3/4 bg-surface-700/50 rounded animate-pulse" />
                  <div className="h-3 w-1/2 bg-surface-700/50 rounded animate-pulse" />
                </div>
              </div>
            ))}
          </div>
        ) : tracks.length === 0 ? (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            className="flex flex-col items-center justify-center py-24 text-center"
          >
            <div className="w-16 h-16 rounded-2xl bg-surface-800 flex items-center justify-center mb-4">
              <Heart className="w-8 h-8 text-surface-600" />
            </div>
            <h3 className="text-lg font-semibold text-surface-300 mb-2">No favorites yet</h3>
            <p className="text-surface-500 text-sm">Heart tracks in your stream to save them here.</p>
          </motion.div>
        ) : (
          <div className={clsx(
            viewMode === 'grid'
              ? 'grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 2xl:grid-cols-7 gap-3'
              : 'space-y-1'
          )}>
            <AnimatePresence mode="popLayout">
              {tracks.map((track, i) => (
                <TrackCard
                  key={track.track_id}
                  track={track}
                  index={i}
                  viewMode={viewMode}
                  isNew={newTrackIdsRef.current.has(track.track_id)}
                />
              ))}
            </AnimatePresence>
          </div>
        )}
      </div>

      {/* RSS Modal */}
      <AnimatePresence>
        {rssModalOpen && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm"
            onClick={() => setRssModalOpen(false)}
          >
            <motion.div
              initial={{ scale: 0.95, opacity: 0 }}
              animate={{ scale: 1, opacity: 1 }}
              exit={{ scale: 0.95, opacity: 0 }}
              transition={{ duration: 0.15 }}
              onClick={e => e.stopPropagation()}
              className="w-full max-w-md bg-surface-900 border border-surface-700/60 rounded-2xl shadow-2xl p-6"
            >
              <div className="flex items-center justify-between mb-4">
                <div className="flex items-center gap-2">
                  <Rss className="w-5 h-5 text-orange-400" />
                  <h2 className="text-base font-semibold text-surface-100">RSS Feed</h2>
                </div>
                <button
                  onClick={() => setRssModalOpen(false)}
                  className="p-1.5 text-surface-500 hover:text-surface-300 transition-colors"
                >
                  <X className="w-4 h-4" />
                </button>
              </div>

              <p className="text-sm text-surface-400 mb-4">
                Subscribe to your favorites feed in any RSS reader or podcast app. This is a public URL — anyone with the link can access it.
              </p>

              <div className="flex gap-2 mb-4">
                <a
                  href="/feed/rss"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="flex items-center gap-1.5 px-3 py-2 rounded-xl border border-surface-700/50 bg-surface-800/60 text-surface-300 hover:text-surface-100 text-sm transition-colors"
                >
                  <Rss className="w-3.5 h-3.5" />
                  Preview feed
                </a>
              </div>

              {publicRssUrl && (
                <div className="space-y-2">
                  <p className="text-xs text-surface-500 font-medium uppercase tracking-wider">Public URL</p>
                  <div className="flex items-center gap-2">
                    <input
                      readOnly
                      value={publicRssUrl}
                      className="flex-1 min-w-0 text-xs bg-surface-800/80 border border-surface-700/50 rounded-xl px-3 py-2 text-surface-300 focus:outline-none focus:border-accent/50 cursor-text"
                      onFocus={e => e.target.select()}
                    />
                    <button
                      onClick={handleCopyRss}
                      className={clsx(
                        'flex-shrink-0 flex items-center gap-1.5 px-3 py-2 rounded-xl border text-sm font-medium transition-colors',
                        rssCopied
                          ? 'border-green-500/40 bg-green-500/10 text-green-400'
                          : 'border-surface-700/50 bg-surface-800/60 text-surface-300 hover:text-surface-100 hover:border-surface-600/60'
                      )}
                    >
                      {rssCopied ? <Check className="w-3.5 h-3.5" /> : <Copy className="w-3.5 h-3.5" />}
                      {rssCopied ? 'Copied' : 'Copy'}
                    </button>
                  </div>
                </div>
              )}
            </motion.div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  )
}
