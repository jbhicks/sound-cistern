import { useState, useMemo } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { Sparkles, X, Play, Clock, Zap, Shuffle, Save, Check, Loader2 } from 'lucide-react'
import { useStore } from '../store'
import clsx from 'clsx'

const BPM_PRESETS = [
  { label: 'Chill', bpm: 90, desc: 'Downtempo, ambient' },
  { label: 'Deep House', bpm: 122, desc: 'Classic house vibes' },
  { label: 'Techno', bpm: 128, desc: 'Driving, energetic' },
  { label: 'Trance', bpm: 138, desc: 'Uplifting, melodic' },
  { label: 'DnB', bpm: 174, desc: 'Fast, intense' },
]

export default function AutoMix({ tracks, onPlayTrack }) {
  const [isOpen, setIsOpen] = useState(false)
  const [targetBPM, setTargetBPM] = useState(128)
  const [tolerance, setTolerance] = useState(6)
  const [generatedMix, setGeneratedMix] = useState(null)
  const [saveDialogOpen, setSaveDialogOpen] = useState(false)
  const [playlistName, setPlaylistName] = useState('')
  const [isSaving, setIsSaving] = useState(false)
  const [saveSuccess, setSaveSuccess] = useState(false)
  const { playTrack } = useStore()

  // Calculate compatible tracks
  const compatibleTracks = useMemo(() => {
    if (!tracks || tracks.length === 0) return []
    
    const minBPM = targetBPM * (1 - tolerance / 100)
    const maxBPM = targetBPM * (1 + tolerance / 100)
    
    return tracks
      .filter(t => t.bpm > 0 && t.bpm >= minBPM && t.bpm <= maxBPM)
      .sort((a, b) => {
        // Sort by closeness to target BPM
        const diffA = Math.abs(a.bpm - targetBPM)
        const diffB = Math.abs(b.bpm - targetBPM)
        return diffA - diffB
      })
  }, [tracks, targetBPM, tolerance])

  const generateMix = () => {
    if (compatibleTracks.length === 0) return
    
    // Create a mix that flows well
    const mix = [...compatibleTracks].sort((a, b) => {
      // Sort by BPM for smooth transitions
      return a.bpm - b.bpm
    })
    
    setGeneratedMix(mix)
  }

  const clearMix = () => {
    setGeneratedMix(null)
    setSaveSuccess(false)
  }

  const saveMixAsPlaylist = async () => {
    if (!generatedMix || generatedMix.length === 0) return
    
    setIsSaving(true)
    setSaveSuccess(false)
    
    try {
      const name = playlistName.trim() || `AutoMix ${targetBPM} BPM`
      const trackIds = generatedMix.map(t => t.track_id)
      
      const res = await fetch('/api/playlists', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({
          name,
          track_ids: trackIds,
          description: `Auto-generated mix: ${stats.min}-${stats.max} BPM, ${stats.count} tracks`
        })
      })
      
      if (res.ok) {
        setSaveSuccess(true)
        setSaveDialogOpen(false)
        setPlaylistName('')
        // Refresh playlists in store
        const { loadPlaylists } = useStore.getState()
        if (loadPlaylists) loadPlaylists()
      } else {
        throw new Error('Failed to save playlist')
      }
    } catch (err) {
      console.error('Error saving mix:', err)
      alert('Failed to save playlist. Please try again.')
    } finally {
      setIsSaving(false)
    }
  }

  const playMixTrack = (track) => {
    if (onPlayTrack) {
      onPlayTrack(track)
    } else {
      playTrack(track)
    }
  }

  const stats = useMemo(() => {
    if (!generatedMix || generatedMix.length === 0) return null
    
    const bpms = generatedMix.map(t => t.bpm)
    const min = Math.min(...bpms)
    const max = Math.max(...bpms)
    const avg = bpms.reduce((a, b) => a + b, 0) / bpms.length
    const totalDuration = generatedMix.reduce((sum, t) => sum + (t.track_duration || 0), 0)
    
    return {
      count: generatedMix.length,
      min,
      max,
      avg: Math.round(avg),
      duration: Math.round(totalDuration / 1000 / 60) // minutes
    }
  }, [generatedMix])

  return (
    <>
      {/* Trigger button */}
      <button
        onClick={() => setIsOpen(true)}
        className="flex items-center gap-2 px-3 py-2 rounded-xl bg-gradient-to-r from-accent/20 to-vapor-pink/20 border border-accent/30 text-accent-light text-sm font-medium hover:from-accent/30 hover:to-vapor-pink/30 transition-all"
      >
        <Sparkles className="w-4 h-4" />
        <span className="hidden sm:inline">AutoMix</span>
      </button>

      {/* Panel */}
      <AnimatePresence>
        {isOpen && (
          <>
            {/* Backdrop */}
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              className="fixed inset-0 z-40 bg-black/60 backdrop-blur-sm"
              onClick={() => setIsOpen(false)}
            />

            {/* Content */}
            <motion.div
              initial={{ opacity: 0, scale: 0.95, y: 20 }}
              animate={{ opacity: 1, scale: 1, y: 0 }}
              exit={{ opacity: 0, scale: 0.95, y: 20 }}
              transition={{ type: 'spring', stiffness: 380, damping: 38 }}
              className="fixed inset-4 md:inset-auto md:top-20 md:left-1/2 md:-translate-x-1/2 md:w-[600px] md:max-h-[80vh] z-50 bg-surface-900 border border-surface-700/60 rounded-2xl shadow-2xl overflow-hidden flex flex-col"
            >
              {/* Header */}
              <div className="flex items-center justify-between px-5 py-4 border-b border-surface-800/60 bg-surface-900/50">
                <div className="flex items-center gap-2.5">
                  <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-accent to-vapor-pink flex items-center justify-center">
                    <Sparkles className="w-4 h-4 text-white" />
                  </div>
                  <div>
                    <h2 className="font-semibold text-surface-100">AutoMix Generator</h2>
                    <p className="text-xs text-surface-500">Create seamless BPM-matched playlists</p>
                  </div>
                </div>
                <button
                  onClick={() => setIsOpen(false)}
                  className="p-2 rounded-lg text-surface-500 hover:text-surface-300 hover:bg-surface-800 transition-colors"
                >
                  <X className="w-4 h-4" />
                </button>
              </div>

              {/* Body */}
              <div className="flex-1 overflow-y-auto p-5 space-y-5">
                {/* Presets */}
                <div>
                  <label className="text-xs font-semibold text-surface-500 uppercase tracking-wider mb-2 block">
                    Quick Presets
                  </label>
                  <div className="grid grid-cols-2 sm:grid-cols-5 gap-2">
                    {BPM_PRESETS.map(preset => (
                      <button
                        key={preset.label}
                        onClick={() => setTargetBPM(preset.bpm)}
                        className={clsx(
                          'p-2 rounded-xl border text-left transition-all',
                          targetBPM === preset.bpm
                            ? 'bg-accent/15 border-accent/40'
                            : 'bg-surface-800/50 border-surface-700/40 hover:border-surface-600'
                        )}
                      >
                        <div className="text-xs font-semibold text-surface-200">{preset.label}</div>
                        <div className="text-[10px] text-surface-500">{preset.bpm} BPM</div>
                      </button>
                    ))}
                  </div>
                </div>

                {/* Custom controls */}
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="text-xs font-semibold text-surface-500 uppercase tracking-wider mb-2 block">
                      Target BPM
                    </label>
                    <div className="flex items-center gap-2">
                      <input
                        type="range"
                        min="60"
                        max="200"
                        value={targetBPM}
                        onChange={(e) => setTargetBPM(parseInt(e.target.value))}
                        className="flex-1 accent-accent h-1.5"
                      />
                      <span className="text-lg font-bold text-accent-light w-12 text-right tabular-nums">
                        {targetBPM}
                      </span>
                    </div>
                  </div>

                  <div>
                    <label className="text-xs font-semibold text-surface-500 uppercase tracking-wider mb-2 block">
                      Tolerance
                    </label>
                    <div className="flex items-center gap-2">
                      <input
                        type="range"
                        min="3"
                        max="12"
                        step="1"
                        value={tolerance}
                        onChange={(e) => setTolerance(parseInt(e.target.value))}
                        className="flex-1 accent-accent h-1.5"
                      />
                      <span className="text-sm font-medium text-surface-300 w-12 text-right">
                        ±{tolerance}%
                      </span>
                    </div>
                  </div>
                </div>

                {/* Compatibility info */}
                <div className="p-3 rounded-xl bg-surface-800/50 border border-surface-700/40">
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-surface-400">Compatible tracks:</span>
                    <span className="text-lg font-bold text-accent-light">
                      {compatibleTracks.length}
                    </span>
                  </div>
                  <div className="text-xs text-surface-500 mt-1">
                    BPM range: {Math.round(targetBPM * (1 - tolerance / 100))} - {Math.round(targetBPM * (1 + tolerance / 100))}
                  </div>
                </div>

                {/* Generate button */}
                <button
                  onClick={generateMix}
                  disabled={compatibleTracks.length === 0}
                  className="w-full py-3 rounded-xl bg-accent hover:bg-accent-hover disabled:opacity-50 disabled:cursor-not-allowed text-white font-medium transition-all flex items-center justify-center gap-2"
                >
                  <Shuffle className="w-4 h-4" />
                  Generate Mix ({compatibleTracks.length} tracks)
                </button>

                {/* Generated mix */}
                {generatedMix && stats && (
                  <div className="border-t border-surface-800/60 pt-4">
                    {/* Stats */}
                    <div className="grid grid-cols-4 gap-2 mb-4">
                      <div className="p-2 rounded-lg bg-surface-800/50 text-center">
                        <div className="text-lg font-bold text-surface-200">{stats.count}</div>
                        <div className="text-[10px] text-surface-500">Tracks</div>
                      </div>
                      <div className="p-2 rounded-lg bg-surface-800/50 text-center">
                        <div className="text-lg font-bold text-surface-200">{stats.min}-{stats.max}</div>
                        <div className="text-[10px] text-surface-500">BPM Range</div>
                      </div>
                      <div className="p-2 rounded-lg bg-surface-800/50 text-center">
                        <div className="text-lg font-bold text-accent-light">{stats.avg}</div>
                        <div className="text-[10px] text-surface-500">Avg BPM</div>
                      </div>
                      <div className="p-2 rounded-lg bg-surface-800/50 text-center">
                        <div className="text-lg font-bold text-surface-200">{stats.duration}m</div>
                        <div className="text-[10px] text-surface-500">Duration</div>
                      </div>
                    </div>

                    {/* Track list */}
                    <div className="space-y-1 max-h-48 overflow-y-auto">
                      {generatedMix.map((track, i) => (
                        <button
                          key={track.track_id}
                          onClick={() => playMixTrack(track)}
                          className="w-full flex items-center gap-3 p-2 rounded-lg hover:bg-surface-800/70 transition-colors group text-left"
                        >
                          <span className="text-xs text-surface-500 w-5">{i + 1}</span>
                          <img
                            src={track.artwork_url?.replace(/-t\d+x\d+/g, '-t50x50')}
                            alt=""
                            className="w-8 h-8 rounded object-cover"
                          />
                          <div className="flex-1 min-w-0">
                            <div className="text-xs font-medium text-surface-200 truncate">{track.track_title}</div>
                            <div className="text-[10px] text-surface-500 truncate">{track.artist_name}</div>
                          </div>
                          <div className="flex items-center gap-2">
                            <span className="text-xs font-medium text-accent-light tabular-nums">
                              {Math.round(track.bpm)}
                            </span>
                            <span className="text-[10px] text-surface-600">BPM</span>
                            <Play className="w-3 h-3 text-surface-600 opacity-0 group-hover:opacity-100 transition-opacity" />
                          </div>
                        </button>
                      ))}
                    </div>

                    {/* Action buttons */}
                    <div className="flex gap-2 mt-3">
                      <button
                        onClick={() => {
                          // Play all tracks starting from first
                          if (generatedMix && generatedMix.length > 0) {
                            // Play first track
                            playMixTrack(generatedMix[0])
                            // Close the panel
                            setIsOpen(false)
                          }
                        }}
                        className="flex-1 flex items-center justify-center gap-2 py-2 rounded-lg bg-accent hover:bg-accent-hover text-xs text-white transition-colors"
                      >
                        <Play className="w-3.5 h-3.5" />
                        Play All
                      </button>
                      <button
                        onClick={() => setSaveDialogOpen(true)}
                        className="flex-1 flex items-center justify-center gap-2 py-2 rounded-lg bg-surface-800 hover:bg-surface-700 text-xs text-surface-300 transition-colors"
                      >
                        <Save className="w-3.5 h-3.5" />
                        Save
                      </button>
                      <button
                        onClick={clearMix}
                        className="px-3 py-2 rounded-lg bg-surface-800 hover:bg-surface-700 text-xs text-surface-400 transition-colors"
                        title="Clear"
                      >
                        <X className="w-3.5 h-3.5" />
                      </button>
                    </div>

                    {saveSuccess && (
                      <div className="mt-2 p-2 rounded-lg bg-green-500/20 border border-green-500/40 text-green-400 text-xs text-center flex items-center justify-center gap-1.5">
                        <Check className="w-3.5 h-3.5" />
                        Playlist saved successfully!
                      </div>
                    )}
                  </div>
                )}
              </div>
            </motion.div>

            {/* Save Dialog */}
            <AnimatePresence>
              {saveDialogOpen && (
                <>
                  <motion.div
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    exit={{ opacity: 0 }}
                    className="fixed inset-0 z-[60] bg-black/70 backdrop-blur-sm"
                    onClick={() => !isSaving && setSaveDialogOpen(false)}
                  />
                  <motion.div
                    initial={{ opacity: 0, scale: 0.95 }}
                    animate={{ opacity: 1, scale: 1 }}
                    exit={{ opacity: 0, scale: 0.95 }}
                    className="fixed inset-0 flex items-center justify-center z-[70] pointer-events-none"
                  >
                    <div className="bg-surface-900 border border-surface-700/60 rounded-2xl shadow-2xl p-6 w-full max-w-sm pointer-events-auto">
                      <h3 className="text-lg font-semibold text-surface-100 mb-1">Save Playlist</h3>
                      <p className="text-xs text-surface-500 mb-4">
                        Save this {stats?.count} track mix to your playlists
                      </p>
                      
                      <input
                        type="text"
                        placeholder={`AutoMix ${targetBPM} BPM`}
                        value={playlistName}
                        onChange={(e) => setPlaylistName(e.target.value)}
                        className="w-full px-3 py-2 bg-surface-800 border border-surface-700/60 rounded-xl text-sm text-surface-100 placeholder-surface-600 focus:outline-none focus:border-accent/60 mb-4"
                        disabled={isSaving}
                      />
                      
                      <div className="flex gap-2">
                        <button
                          onClick={() => setSaveDialogOpen(false)}
                          disabled={isSaving}
                          className="flex-1 py-2 rounded-xl bg-surface-800 hover:bg-surface-700 text-sm text-surface-400 transition-colors disabled:opacity-50"
                        >
                          Cancel
                        </button>
                        <button
                          onClick={saveMixAsPlaylist}
                          disabled={isSaving}
                          className="flex-1 py-2 rounded-xl bg-accent hover:bg-accent-hover text-sm text-white transition-colors flex items-center justify-center gap-2 disabled:opacity-50"
                        >
                          {isSaving ? (
                            <>
                              <Loader2 className="w-4 h-4 animate-spin" />
                              Saving...
                            </>
                          ) : (
                            <>
                              <Save className="w-4 h-4" />
                              Save
                            </>
                          )}
                        </button>
                      </div>
                    </div>
                  </motion.div>
                </>
              )}
            </AnimatePresence>
          </>
        )}
      </AnimatePresence>
    </>
  )
}
