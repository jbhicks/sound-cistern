import { useState, useEffect, useRef } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { ListMusic, Plus, Play, Trash2, X, Download, Edit2, Share2, ArrowLeft, Clock, Music2, Check, Copy } from 'lucide-react'
import { formatDuration } from '../components/TrackCard'
import { useStore } from '../store'

export default function Playlists() {
  const { playlists, setPlaylists, loadPlaylists } = useStore()
  const [loading, setLoading] = useState(true)
  const [showModal, setShowModal] = useState(false)
  const [newName, setNewName] = useState('')
  const [creating, setCreating] = useState(false)

  // Rename state
  const [editingId, setEditingId] = useState(null)
  const [editingName, setEditingName] = useState('')
  const renameRef = useRef(null)

  // Share state
  const [shareModal, setShareModal] = useState(null) // { share_token, share_url }
  const [copied, setCopied] = useState(false)

  // Detail view state
  const [detailPlaylist, setDetailPlaylist] = useState(null) // { id, name, tracks, share_token }
  const [detailLoading, setDetailLoading] = useState(false)

  useEffect(() => {
    fetch('/api/playlists', { credentials: 'include', headers: { Accept: 'application/json' } })
      .then(r => r.ok ? r.json() : [])
      .then(d => setPlaylists(Array.isArray(d) ? d : []))
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [])

  // Focus rename input when editing starts
  useEffect(() => {
    if (editingId && renameRef.current) {
      renameRef.current.focus()
      renameRef.current.select()
    }
  }, [editingId])

  const handleCreate = async (e) => {
    e.preventDefault()
    if (!newName.trim()) return
    setCreating(true)
    try {
      const res = await fetch('/api/playlists', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ name: newName.trim() }),
      })
      if (res.ok) {
        const p = await res.json()
        setPlaylists([p, ...playlists])
        setNewName('')
        setShowModal(false)
      }
    } finally {
      setCreating(false)
    }
  }

  const handleDelete = async (id) => {
    if (!confirm('Delete this playlist?')) return
    const res = await fetch(`/api/playlists/${id}`, { method: 'DELETE', credentials: 'include' })
    if (res.ok) setPlaylists(playlists.filter(p => p.id !== id))
  }

  const startRename = (e, pl) => {
    e.stopPropagation()
    setEditingId(pl.id)
    setEditingName(pl.name)
  }

  const commitRename = async (id) => {
    const name = editingName.trim()
    setEditingId(null)
    if (!name) return
    const res = await fetch(`/api/playlists/${id}`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify({ name }),
    })
    if (res.ok) {
      const updated = await res.json()
      setPlaylists(playlists.map(p => p.id === id ? { ...p, name: updated.name } : p))
    }
  }

  const handleShare = async (e, id) => {
    e.stopPropagation()
    const res = await fetch(`/api/playlists/${id}/share`, {
      method: 'POST',
      credentials: 'include',
    })
    if (res.ok) {
      const data = await res.json()
      const fullUrl = `${window.location.origin}${data.share_url}`
      setShareModal({ ...data, share_url: fullUrl })
      setCopied(false)
    }
  }

  const copyShareUrl = async () => {
    if (shareModal?.share_url) {
      await navigator.clipboard.writeText(shareModal.share_url)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    }
  }

  const openDetail = async (pl) => {
    setDetailLoading(true)
    setDetailPlaylist({ id: pl.id, name: pl.name, tracks: null })
    const res = await fetch(`/api/playlists/${pl.id}`, { credentials: 'include' })
    if (res.ok) {
      const data = await res.json()
      setDetailPlaylist(data)
    }
    setDetailLoading(false)
  }

  const removeTrackFromDetail = async (ptId) => {
    if (!detailPlaylist) return
    const res = await fetch(`/api/playlists/${detailPlaylist.id}/tracks/${ptId}`, {
      method: 'DELETE',
      credentials: 'include',
    })
    if (res.ok) {
      setDetailPlaylist(prev => ({
        ...prev,
        tracks: prev.tracks.filter(t => t.pt_id !== ptId),
      }))
      // Update track count in list
      setPlaylists(playlists.map(p =>
        p.id === detailPlaylist.id
          ? { ...p, track_count: Math.max(0, (p.track_count || 1) - 1) }
          : p
      ))
    }
  }

  // ── DETAIL VIEW ──────────────────────────────────────────────────
  if (detailPlaylist) {
    return (
      <div className="min-h-screen px-4 py-6 md:px-6 lg:px-8">
        <div className="max-w-3xl mx-auto">
          <motion.div
            initial={{ opacity: 0, y: -16 }}
            animate={{ opacity: 1, y: 0 }}
            className="mb-6"
          >
            <button
              onClick={() => setDetailPlaylist(null)}
              className="flex items-center gap-1.5 text-sm text-surface-400 hover:text-surface-200 mb-4 transition-colors"
            >
              <ArrowLeft className="w-4 h-4" />
              Back to playlists
            </button>
            <h1 className="text-2xl font-bold text-gradient">{detailPlaylist.name}</h1>
            {detailPlaylist.tracks && (
              <p className="text-surface-500 text-sm mt-0.5">{detailPlaylist.tracks.length} tracks</p>
            )}
          </motion.div>

          {detailLoading ? (
            <div className="space-y-2">
              {Array.from({ length: 4 }).map((_, i) => (
                <div key={i} className="flex items-center gap-3 p-3 rounded-xl bg-surface-800/40 border border-surface-700/30">
                  <div className="w-12 h-12 rounded-lg bg-surface-700/50 animate-pulse flex-shrink-0" />
                  <div className="flex-1 space-y-2">
                    <div className="h-3.5 w-2/3 bg-surface-700/50 rounded animate-pulse" />
                    <div className="h-3 w-1/3 bg-surface-700/50 rounded animate-pulse" />
                  </div>
                </div>
              ))}
            </div>
          ) : !detailPlaylist.tracks || detailPlaylist.tracks.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-24 text-center">
              <div className="w-16 h-16 rounded-2xl bg-surface-800 flex items-center justify-center mb-4">
                <Music2 className="w-8 h-8 text-surface-600" />
              </div>
              <h3 className="text-lg font-semibold text-surface-300 mb-2">No tracks yet</h3>
              <p className="text-surface-500 text-sm">Add tracks to this playlist from your favorites or stream.</p>
            </div>
          ) : (
            <div className="space-y-1">
              <AnimatePresence>
                {detailPlaylist.tracks.map((t, i) => (
                  <motion.div
                    key={t.pt_id}
                    layout
                    initial={{ opacity: 0, x: -10 }}
                    animate={{ opacity: 1, x: 0 }}
                    exit={{ opacity: 0, x: -10, height: 0 }}
                    transition={{ delay: i * 0.02 }}
                    className="flex items-center gap-3 p-3 rounded-xl bg-surface-800/40 border border-surface-700/30 hover:border-surface-600/40 group transition-all"
                  >
                    <div className="w-12 h-12 rounded-lg overflow-hidden flex-shrink-0 bg-surface-700">
                      {t.artwork_url
                        ? <img src={t.artwork_url.replace(/-t\d+x\d+/g, '-t500x500')} alt="" className="w-full h-full object-cover" />
                        : <div className="w-full h-full flex items-center justify-center"><Music2 className="w-5 h-5 text-surface-600" /></div>
                      }
                    </div>
                    <div className="flex-1 min-w-0">
                      <p className="font-medium text-sm text-surface-100 truncate">{t.track_title}</p>
                      <p className="text-xs text-surface-400 truncate">{t.artist_name}</p>
                    </div>
                    {t.genre && (
                      <span className="hidden md:block px-2 py-0.5 text-xs rounded-full bg-surface-700/80 text-surface-400 flex-shrink-0">
                        {t.genre}
                      </span>
                    )}
                    <span className="text-xs text-surface-500 w-10 text-right flex-shrink-0 flex items-center gap-0.5">
                      <Clock className="w-3 h-3" />
                      {formatDuration(t.track_duration)}
                    </span>
                    <button
                      onClick={() => removeTrackFromDetail(t.pt_id)}
                      className="p-1.5 rounded-lg text-surface-600 hover:text-red-400 hover:bg-red-400/10 opacity-0 group-hover:opacity-100 transition-all flex-shrink-0"
                      title="Remove from playlist"
                    >
                      <Trash2 className="w-3.5 h-3.5" />
                    </button>
                  </motion.div>
                ))}
              </AnimatePresence>
            </div>
          )}
        </div>
      </div>
    )
  }

  // ── LIST VIEW ────────────────────────────────────────────────────
  return (
    <div className="min-h-screen px-4 py-6 md:px-6 lg:px-8">
      <div className="max-w-7xl mx-auto">
        <motion.div
          initial={{ opacity: 0, y: -16 }}
          animate={{ opacity: 1, y: 0 }}
          className="mb-6 flex items-center justify-between"
        >
          <div>
            <h1 className="text-2xl font-bold text-gradient">Playlists</h1>
            <p className="text-surface-500 text-sm mt-0.5">Your collections</p>
          </div>
          <div className="flex items-center gap-2">
            <button
              onClick={() => {
                const a = document.createElement('a')
                a.href = '/api/export/playlists/json'
                a.download = 'playlists.json'
                document.body.appendChild(a)
                a.click()
                document.body.removeChild(a)
              }}
              className="flex items-center gap-1.5 px-3 py-2 rounded-xl border border-surface-700/50 bg-surface-800/60 text-surface-400 hover:text-surface-200 hover:border-surface-600/60 text-sm transition-colors"
            >
              <Download className="w-4 h-4" />
              <span className="hidden sm:inline">Export JSON</span>
            </button>
            <button
              onClick={() => setShowModal(true)}
              className="flex items-center gap-2 px-3 py-2 rounded-xl bg-accent hover:bg-accent-hover text-white text-sm font-medium transition-colors shadow-lg shadow-accent/20"
            >
              <Plus className="w-4 h-4" />
              <span className="hidden sm:inline">New Playlist</span>
            </button>
          </div>
        </motion.div>

        {loading ? (
          <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-3">
            {Array.from({ length: 6 }).map((_, i) => (
              <div key={i} className="rounded-2xl overflow-hidden bg-surface-800/40 border border-surface-700/30">
                <div className="aspect-square animate-pulse bg-surface-700/50" />
                <div className="p-3 space-y-2">
                  <div className="h-3.5 w-3/4 bg-surface-700/50 rounded animate-pulse" />
                  <div className="h-3 w-1/2 bg-surface-700/50 rounded animate-pulse" />
                </div>
              </div>
            ))}
          </div>
        ) : playlists.length === 0 ? (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            className="flex flex-col items-center justify-center py-24 text-center"
          >
            <div className="w-16 h-16 rounded-2xl bg-surface-800 flex items-center justify-center mb-4">
              <ListMusic className="w-8 h-8 text-surface-600" />
            </div>
            <h3 className="text-lg font-semibold text-surface-300 mb-2">No playlists yet</h3>
            <p className="text-surface-500 text-sm mb-6">Create a playlist to organize your music.</p>
            <button
              onClick={() => setShowModal(true)}
              className="flex items-center gap-2 px-4 py-2 rounded-xl bg-accent hover:bg-accent-hover text-white text-sm font-medium transition-colors"
            >
              <Plus className="w-4 h-4" />
              Create Playlist
            </button>
          </motion.div>
        ) : (
          <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-3">
            <AnimatePresence mode="popLayout">
              {playlists.map((pl, i) => (
                <motion.div
                  key={pl.id}
                  layout
                  initial={{ opacity: 0, scale: 0.92 }}
                  animate={{ opacity: 1, scale: 1 }}
                  exit={{ opacity: 0, scale: 0.92 }}
                  transition={{ delay: i * 0.04 }}
                  className="group relative rounded-2xl overflow-hidden bg-surface-800/60 border border-surface-700/40 hover:border-surface-600/60 transition-all cursor-pointer"
                  onClick={() => editingId !== pl.id && openDetail(pl)}
                >
                  <div className="aspect-square overflow-hidden bg-gradient-to-br from-accent/20 to-vapor-pink/20 flex items-center justify-center relative">
                    {pl.artwork_url ? (
                      <img src={pl.artwork_url} alt={pl.name} className="w-full h-full object-cover" />
                    ) : (
                      <ListMusic className="w-12 h-12 text-surface-600" />
                    )}
                    <div className="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center">
                      <div className="w-12 h-12 rounded-full bg-accent flex items-center justify-center shadow-xl">
                        <Play className="w-5 h-5 text-white ml-0.5" />
                      </div>
                    </div>
                  </div>
                  <div className="p-3">
                    {editingId === pl.id ? (
                      <input
                        ref={renameRef}
                        type="text"
                        value={editingName}
                        onChange={e => setEditingName(e.target.value)}
                        onBlur={() => commitRename(pl.id)}
                        onKeyDown={e => {
                          if (e.key === 'Enter') { e.preventDefault(); commitRename(pl.id) }
                          if (e.key === 'Escape') setEditingId(null)
                        }}
                        onClick={e => e.stopPropagation()}
                        className="w-full bg-surface-700 border border-accent/40 rounded-lg px-2 py-1 text-sm text-surface-100 focus:outline-none focus:border-accent/70 focus:ring-1 focus:ring-accent/20"
                      />
                    ) : (
                      <p className="font-semibold text-sm text-surface-100 truncate">{pl.name}</p>
                    )}
                    <p className="text-xs text-surface-500 mt-0.5">{pl.track_count || 0} tracks</p>
                  </div>

                  {/* Action buttons — visible on hover */}
                  <div className="absolute top-2 right-2 flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-all">
                    <button
                      onClick={(e) => { e.stopPropagation(); handleShare(e, pl.id) }}
                      className="p-1.5 rounded-full bg-black/40 text-white/60 hover:text-blue-400 hover:bg-black/60 transition-all"
                      title="Share playlist"
                    >
                      <Share2 className="w-3.5 h-3.5" />
                    </button>
                    <button
                      onClick={(e) => startRename(e, pl)}
                      className="p-1.5 rounded-full bg-black/40 text-white/60 hover:text-yellow-400 hover:bg-black/60 transition-all"
                      title="Rename playlist"
                    >
                      <Edit2 className="w-3.5 h-3.5" />
                    </button>
                    <button
                      onClick={(e) => { e.stopPropagation(); handleDelete(pl.id) }}
                      className="p-1.5 rounded-full bg-black/40 text-white/60 hover:text-red-400 hover:bg-black/60 transition-all"
                      title="Delete playlist"
                    >
                      <Trash2 className="w-3.5 h-3.5" />
                    </button>
                  </div>
                </motion.div>
              ))}
            </AnimatePresence>
          </div>
        )}
      </div>

      {/* Create modal */}
      <AnimatePresence>
        {showModal && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm"
            onClick={() => setShowModal(false)}
          >
            <motion.div
              initial={{ scale: 0.92, opacity: 0 }}
              animate={{ scale: 1, opacity: 1 }}
              exit={{ scale: 0.92, opacity: 0 }}
              transition={{ type: 'spring', stiffness: 400, damping: 30 }}
              className="w-full max-w-sm bg-surface-900 border border-surface-700/60 rounded-2xl p-6 shadow-2xl"
              onClick={e => e.stopPropagation()}
            >
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-lg font-bold text-surface-100">New Playlist</h2>
                <button onClick={() => setShowModal(false)} className="p-1.5 text-surface-500 hover:text-surface-300 rounded-lg transition-colors">
                  <X className="w-4 h-4" />
                </button>
              </div>
              <form onSubmit={handleCreate}>
                <input
                  type="text"
                  placeholder="Playlist name"
                  value={newName}
                  onChange={e => setNewName(e.target.value)}
                  autoFocus
                  className="w-full px-4 py-3 bg-surface-800 border border-surface-700 rounded-xl text-surface-100 placeholder-surface-500 focus:outline-none focus:border-accent/60 focus:ring-1 focus:ring-accent/20 transition-all mb-4 text-sm"
                />
                <div className="flex gap-2 justify-end">
                  <button
                    type="button"
                    onClick={() => setShowModal(false)}
                    className="px-4 py-2 text-sm text-surface-400 hover:text-surface-200 rounded-xl bg-surface-800 hover:bg-surface-700 transition-colors"
                  >
                    Cancel
                  </button>
                  <button
                    type="submit"
                    disabled={creating || !newName.trim()}
                    className="px-4 py-2 text-sm bg-accent hover:bg-accent-hover text-white rounded-xl font-medium transition-colors disabled:opacity-50"
                  >
                    {creating ? 'Creating…' : 'Create'}
                  </button>
                </div>
              </form>
            </motion.div>
          </motion.div>
        )}
      </AnimatePresence>

      {/* Share modal */}
      <AnimatePresence>
        {shareModal && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm"
            onClick={() => setShareModal(null)}
          >
            <motion.div
              initial={{ scale: 0.92, opacity: 0 }}
              animate={{ scale: 1, opacity: 1 }}
              exit={{ scale: 0.92, opacity: 0 }}
              transition={{ type: 'spring', stiffness: 400, damping: 30 }}
              className="w-full max-w-sm bg-surface-900 border border-surface-700/60 rounded-2xl p-6 shadow-2xl"
              onClick={e => e.stopPropagation()}
            >
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-lg font-bold text-surface-100">Share Playlist</h2>
                <button onClick={() => setShareModal(null)} className="p-1.5 text-surface-500 hover:text-surface-300 rounded-lg transition-colors">
                  <X className="w-4 h-4" />
                </button>
              </div>
              <p className="text-sm text-surface-400 mb-3">Anyone with this link can view the playlist.</p>
              <div className="flex gap-2">
                <input
                  readOnly
                  value={shareModal.share_url}
                  className="flex-1 px-3 py-2 bg-surface-800 border border-surface-700 rounded-xl text-surface-300 text-xs focus:outline-none truncate"
                />
                <button
                  onClick={copyShareUrl}
                  className={`flex items-center gap-1.5 px-3 py-2 rounded-xl text-sm font-medium transition-all ${
                    copied
                      ? 'bg-green-600/80 text-white'
                      : 'bg-accent hover:bg-accent-hover text-white'
                  }`}
                >
                  {copied ? <Check className="w-4 h-4" /> : <Copy className="w-4 h-4" />}
                  {copied ? 'Copied' : 'Copy'}
                </button>
              </div>
            </motion.div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  )
}
