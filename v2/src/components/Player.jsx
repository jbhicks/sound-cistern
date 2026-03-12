import { useEffect, useRef, useState, useCallback } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { Play, Pause, X, Volume2, VolumeX, ExternalLink, Radio, Sparkles, Download, Activity } from 'lucide-react'
import { useStore } from '../store'
import { ButterchurnMini, ButterchurnFullscreen } from './ButterchurnVisualizer'
import EQVisualizer from './EQVisualizer'
import BPMControls from './BPMControls'
import clsx from 'clsx'

function formatDurationMs(ms) {
  if (!ms) return '0:00'
  const s = Math.floor(ms / 1000)
  const m = Math.floor(s / 60)
  const sec = s % 60
  return `${m}:${sec.toString().padStart(2, '0')}`
}

const QUALITY_OPTIONS = [
  { value: 'auto',        label: 'Auto',      desc: 'Best available' },
  { value: 'hls_aac_160', label: 'AAC 160',   desc: 'Highest quality' },
  { value: 'http_mp3_128', label: 'MP3 128',  desc: 'Progressive MP3' },
  { value: 'hls_mp3_128', label: 'HLS 128',   desc: 'HLS MP3 stream' },
]

function formatTime(secs) {
  if (!isFinite(secs) || isNaN(secs)) return '0:00'
  const m = Math.floor(secs / 60)
  const s = Math.floor(secs % 60)
  return `${m}:${s.toString().padStart(2, '0')}`
}

export default function Player() {
  const { currentTrack, isPlaying, setIsPlaying, setCurrentTrack, setAudioEl, setAudioContext, streamQuality, setStreamQuality, playTrack } = useStore()
  const audioRef = useRef(null)
  const [progress, setProgress] = useState(0)
  const [duration, setDuration] = useState(0)
  const [currentTime, setCurrentTime] = useState(0)
  const [volume, setVolume] = useState(1)
  const [muted, setMuted] = useState(false)
  const [loading, setLoading] = useState(false)
  const [qualityOpen, setQualityOpen] = useState(false)
  const progressBarRef = useRef(null)
  const vizSeekRef = useRef(null)

  // Visualizer state
  const [vizFullscreen, setVizFullscreen] = useState(false)

  // Related tracks state
  const [relatedOpen, setRelatedOpen] = useState(false)
  const [relatedTracks, setRelatedTracks] = useState([])
  const [relatedLoading, setRelatedLoading] = useState(false)

  // Register audio element with store
  useEffect(() => {
    setAudioEl(audioRef.current)
  }, [setAudioEl])

  // Build shared AudioContext + AnalyserNode (once) and publish to store
  // Must happen after a user gesture (play) due to browser autoplay policy.
  // We lazily create it on first play.
  const audioCtxRef = useRef(null)
  const analyserRef = useRef(null)
  const sourceRef = useRef(null)

  const ensureAudioContext = useCallback(() => {
    const audio = audioRef.current
    if (!audio || audioCtxRef.current) return
    try {
      const ctx = new (window.AudioContext || window.webkitAudioContext)()
      const analyser = ctx.createAnalyser()
      analyser.fftSize = 2048
      analyser.smoothingTimeConstant = 0.8
      const source = ctx.createMediaElementSource(audio)
      source.connect(analyser)
      analyser.connect(ctx.destination)
      audioCtxRef.current = ctx
      analyserRef.current = analyser
      sourceRef.current = source
      setAudioContext(ctx, analyser)
    } catch (e) {
      console.warn('[Player] AudioContext init failed:', e)
    }
  }, [setAudioContext])

  // Resume context on play
  useEffect(() => {
    if (!isPlaying) return
    ensureAudioContext()
    if (audioCtxRef.current?.state === 'suspended') {
      audioCtxRef.current.resume().catch(() => {})
    }
  }, [isPlaying, ensureAudioContext])

  // Fetch related tracks when panel opens or track changes while open
  useEffect(() => {
    if (!relatedOpen || !currentTrack?.track_id) return
    let cancelled = false
    setRelatedLoading(true)
    setRelatedTracks([])
    fetch(`/api/track/${currentTrack.track_id}/related`, { credentials: 'include' })
      .then(r => r.ok ? r.json() : Promise.reject(r.status))
      .then(data => { if (!cancelled) setRelatedTracks(Array.isArray(data) ? data : []) })
      .catch(() => { if (!cancelled) setRelatedTracks([]) })
      .finally(() => { if (!cancelled) setRelatedLoading(false) })
    return () => { cancelled = true }
  }, [relatedOpen, currentTrack?.track_id])

  // Load and play when track or quality changes
  useEffect(() => {
    const audio = audioRef.current
    if (!audio || !currentTrack) return

    const wasPlaying = !audio.paused
    const prevTime = audio.currentTime
    const isReload = audio.src.includes(currentTrack.track_id) // same track, quality change

    setLoading(true)
    if (!isReload) {
      setProgress(0)
      setCurrentTime(0)
      setDuration(0)
    }

    const q = streamQuality && streamQuality !== 'auto' ? `?quality=${streamQuality}` : ''
    audio.src = `/api/track/${currentTrack.track_id}/stream${q}`
    audio.volume = volume
    audio.muted = muted
    audio.load()

    if (isReload && prevTime > 0) {
      audio.addEventListener('canplay', () => { audio.currentTime = prevTime }, { once: true })
    }

    if (!isReload || wasPlaying) {
      audio.play()
        .then(() => setIsPlaying(true))
        .catch(() => setIsPlaying(false))
        .finally(() => setLoading(false))
    } else {
      setLoading(false)
    }
  }, [currentTrack?.track_id, streamQuality]) // eslint-disable-line

  // Sync play/pause from store
  useEffect(() => {
    const audio = audioRef.current
    if (!audio) return
    if (isPlaying) {
      audio.play().catch(() => {})
    } else {
      audio.pause()
    }
  }, [isPlaying])

  const handleTimeUpdate = useCallback(() => {
    const audio = audioRef.current
    if (!audio) return
    setCurrentTime(audio.currentTime)
    setDuration(audio.duration || 0)
    setProgress(audio.duration ? (audio.currentTime / audio.duration) * 100 : 0)
  }, [])

  const handleEnded = useCallback(() => {
    setIsPlaying(false)
    setProgress(0)
    setCurrentTime(0)
  }, [setIsPlaying])

  const handleProgressClick = (e) => {
    const audio = audioRef.current
    const bar = progressBarRef.current
    if (!audio || !bar || !audio.duration) return
    const rect = bar.getBoundingClientRect()
    const pct = Math.max(0, Math.min(1, (e.clientX - rect.left) / rect.width))
    audio.currentTime = pct * audio.duration
  }

  const handleVizSeek = (e) => {
    const audio = audioRef.current
    const bar = vizSeekRef.current
    if (!audio || !bar || !audio.duration) return
    const rect = bar.getBoundingClientRect()
    const pct = Math.max(0, Math.min(1, (e.clientX - rect.left) / rect.width))
    audio.currentTime = pct * audio.duration
  }

  const handleVolumeChange = (e) => {
    const val = parseFloat(e.target.value)
    setVolume(val)
    if (audioRef.current) audioRef.current.volume = val
    if (val > 0) setMuted(false)
  }

  const toggleMute = () => {
    const next = !muted
    setMuted(next)
    if (audioRef.current) audioRef.current.muted = next
  }

  const close = () => {
    const audio = audioRef.current
    if (audio) { audio.pause(); audio.src = '' }
    setCurrentTrack(null)
    setIsPlaying(false)
  }

  const artwork = currentTrack?.artwork_url?.replace(/-t\d+x\d+/g, '-t500x500') || ''

  return (
    <>
      {/* Hidden audio element always in DOM */}
      <audio
        ref={audioRef}
        onTimeUpdate={handleTimeUpdate}
        onEnded={handleEnded}
        onCanPlay={() => setLoading(false)}
        onWaiting={() => setLoading(true)}
        preload="none"
        crossOrigin="anonymous"
      />

      {/* Fullscreen visualizer — lives outside the player bar so it's never nested */}
      <ButterchurnFullscreen open={vizFullscreen} onClose={() => setVizFullscreen(false)} />

      <AnimatePresence>
        {currentTrack && (
          <motion.div
            initial={{ y: 100, opacity: 0 }}
            animate={{ y: 0, opacity: 1 }}
            exit={{ y: 100, opacity: 0 }}
            transition={{ type: 'spring', stiffness: 400, damping: 40 }}
            className="fixed bottom-0 left-0 right-0 z-50"
          >
            {/* Related tracks panel */}
            <AnimatePresence>
              {relatedOpen && (
                <motion.div
                  key="related-panel"
                  initial={{ height: 0, opacity: 0 }}
                  animate={{ height: 'auto', opacity: 1 }}
                  exit={{ height: 0, opacity: 0 }}
                  transition={{ type: 'spring', stiffness: 380, damping: 38 }}
                  className="overflow-hidden bg-surface-900/98 backdrop-blur-2xl border-t border-surface-700/50"
                >
                  <div className="max-w-7xl mx-auto px-4 py-3">
                    {/* Header */}
                    <div className="flex items-center justify-between mb-3">
                      <div className="flex items-center gap-2">
                        <Sparkles className="w-3.5 h-3.5 text-accent" />
                        <span className="text-xs font-semibold text-surface-300 uppercase tracking-wider">Related Tracks</span>
                      </div>
                      <button
                        onClick={() => setRelatedOpen(false)}
                        className="p-1 text-surface-500 hover:text-surface-300 transition-colors"
                      >
                        <X className="w-3.5 h-3.5" />
                      </button>
                    </div>

                    {/* Content */}
                    {relatedLoading ? (
                      <div className="flex items-center justify-center py-6">
                        <div className="w-5 h-5 border-2 border-accent border-t-transparent rounded-full animate-spin" />
                      </div>
                    ) : relatedTracks.length === 0 ? (
                      <p className="text-center text-xs text-surface-500 py-6">No related tracks found</p>
                    ) : (
                      <div className="space-y-1 max-h-64 overflow-y-auto pr-1">
                        {relatedTracks.map((track) => {
                          const art = track.artwork_url?.replace(/-t\d+x\d+/g, '-t500x500') || ''
                          const isActive = currentTrack?.track_id === track.track_id
                          return (
                            <button
                              key={track.track_id}
                              onClick={() => { playTrack(track); setRelatedOpen(false) }}
                              className={clsx(
                                'w-full flex items-center gap-3 px-2 py-1.5 rounded-lg text-left transition-colors group',
                                isActive
                                  ? 'bg-accent/10 border border-accent/30'
                                  : 'hover:bg-surface-800/70 border border-transparent'
                              )}
                            >
                              {/* Artwork */}
                              <div className="w-8 h-8 rounded flex-shrink-0 overflow-hidden bg-surface-700">
                                {art
                                  ? <img src={art} alt="" className="w-full h-full object-cover" />
                                  : <div className="w-full h-full bg-surface-700" />
                                }
                              </div>

                              {/* Info */}
                              <div className="flex-1 min-w-0">
                                <p className={clsx('text-xs font-medium truncate leading-tight', isActive ? 'text-accent-light' : 'text-surface-100')}>
                                  {track.track_title}
                                </p>
                                <p className="text-[10px] text-surface-400 truncate">{track.artist_name}</p>
                              </div>

                              {/* Duration */}
                              <span className="text-[10px] text-surface-500 flex-shrink-0 tabular-nums">
                                {formatDurationMs(track.track_duration)}
                              </span>
                            </button>
                          )
                        })}
                      </div>
                    )}
                  </div>
                </motion.div>
              )}
            </AnimatePresence>

            {/* Progress bar at very top of player */}
            <div
              ref={progressBarRef}
              onClick={handleProgressClick}
              className="h-1 bg-surface-700 cursor-pointer group relative"
            >
              <div
                className="h-full bg-gradient-to-r from-accent to-vapor-pink transition-all duration-100"
                style={{ width: `${progress}%` }}
              />
              <div
                className="absolute top-1/2 -translate-y-1/2 w-3 h-3 rounded-full bg-white shadow-lg opacity-0 group-hover:opacity-100 transition-opacity"
                style={{ left: `calc(${progress}% - 6px)` }}
              />
            </div>

            {/* Player body */}
            <div className="bg-surface-900/95 backdrop-blur-2xl border-t border-surface-700/50 px-4 py-3">
              <div className="max-w-7xl mx-auto flex items-center gap-3">

                {/* Artwork thumbnail → opens fullscreen visualizer */}
                <div
                  className="relative w-10 h-10 rounded-lg overflow-hidden flex-shrink-0 shadow-lg cursor-pointer group"
                  onClick={() => setVizFullscreen(true)}
                  title="Open visualizer"
                >
                  {artwork && (
                    <img src={artwork} alt="" className="absolute inset-0 w-full h-full object-cover" />
                  )}
                  <div className={clsx(
                    'absolute inset-0 transition-opacity duration-300',
                    isPlaying ? 'opacity-100' : 'opacity-0'
                  )}>
                    <ButterchurnMini width={40} height={40} paused={vizFullscreen} />
                  </div>
                  <div className="absolute inset-0 bg-black/0 group-hover:bg-black/40 transition-colors flex items-center justify-center pointer-events-none">
                    <Activity className="w-3.5 h-3.5 text-white opacity-0 group-hover:opacity-100 transition-opacity" />
                  </div>
                  {loading && (
                    <div className="absolute inset-0 bg-black/60 flex items-center justify-center">
                      <div className="w-3.5 h-3.5 border-2 border-accent border-t-transparent rounded-full animate-spin" />
                    </div>
                  )}
                </div>

                {/* Track info */}
                <div className="min-w-0 flex-shrink-0 w-36 lg:w-52">
                  <p className="font-semibold text-surface-100 truncate text-sm leading-tight">
                    {currentTrack.track_title}
                  </p>
                  <p className="text-surface-400 text-xs truncate mt-0.5">
                    {currentTrack.artist_name}
                  </p>
                </div>

                {/* Play / Pause — left of the waveform */}
                <motion.button
                  whileHover={{ scale: 1.05 }}
                  whileTap={{ scale: 0.95 }}
                  onClick={() => setIsPlaying(!isPlaying)}
                  className="w-9 h-9 rounded-full bg-accent hover:bg-accent-hover flex items-center justify-center shadow-lg shadow-accent/30 transition-colors flex-shrink-0"
                >
                  {isPlaying ? (
                    <Pause className="w-4 h-4 text-white" />
                  ) : (
                    <Play className="w-4 h-4 text-white ml-0.5" />
                  )}
                </motion.button>

                {/* Waveform + time — fills remaining center space */}
                <div className="flex-1 min-w-0 flex items-center gap-2">
                  <span className="hidden sm:block text-[11px] text-surface-600 tabular-nums w-7 text-right flex-shrink-0">
                    {formatTime(currentTime)}
                  </span>
                  {/* Visualizer + seekable progress overlay */}
                  <div
                    ref={vizSeekRef}
                    onClick={handleVizSeek}
                    className="relative flex-1 min-w-0 h-10 rounded-lg overflow-hidden bg-surface-800/40 cursor-pointer group/viz"
                  >
                    <EQVisualizer height={40} />

                    {/* Filled progress — subtle tint over the played portion */}
                    <div
                      className="absolute inset-y-0 left-0 bg-accent/15 pointer-events-none transition-none"
                      style={{ width: `${progress}%` }}
                    />

                    {/* Playhead line */}
                    <div
                      className="absolute inset-y-0 w-px bg-accent/70 pointer-events-none"
                      style={{ left: `${progress}%` }}
                    />

                    {/* Seek thumb — appears on hover */}
                    <div
                      className="absolute top-1/2 -translate-y-1/2 -translate-x-1/2 w-2.5 h-2.5 rounded-full bg-white shadow-lg pointer-events-none opacity-0 group-hover/viz:opacity-100 transition-opacity"
                      style={{ left: `${progress}%` }}
                    />
                  </div>
                  <span className="hidden sm:block text-[11px] text-surface-600 tabular-nums w-7 flex-shrink-0">
                    {formatTime(duration || (currentTrack.track_duration / 1000))}
                  </span>
                </div>

                {/* Right: volume + actions */}
                <div className="flex items-center gap-2 flex-shrink-0">
                  {/* Volume */}
                  <div className="hidden sm:flex items-center gap-2">
                    <button
                      onClick={toggleMute}
                      className="p-1.5 text-surface-400 hover:text-surface-100 transition-colors"
                    >
                      {muted || volume === 0 ? (
                        <VolumeX className="w-4 h-4" />
                      ) : (
                        <Volume2 className="w-4 h-4" />
                      )}
                    </button>
                    <input
                      type="range"
                      min="0"
                      max="1"
                      step="0.05"
                      value={muted ? 0 : volume}
                      onChange={handleVolumeChange}
                      className="w-20 accent-accent h-1 cursor-pointer"
                    />
                  </div>

                  {/* Quality selector */}
                  <div className="relative hidden sm:block">
                    <button
                      onClick={() => setQualityOpen(o => !o)}
                      className={clsx(
                        'flex items-center gap-1.5 px-2 py-1 rounded-lg text-xs font-medium transition-colors border',
                        qualityOpen
                          ? 'bg-accent/15 border-accent/40 text-accent-light'
                          : 'bg-surface-800/60 border-surface-700/40 text-surface-400 hover:text-surface-200'
                      )}
                      title="Stream quality"
                    >
                      <Radio className="w-3 h-3" />
                      <span>{QUALITY_OPTIONS.find(o => o.value === streamQuality)?.label ?? 'Auto'}</span>
                    </button>
                    <AnimatePresence>
                      {qualityOpen && (
                        <motion.div
                          initial={{ opacity: 0, y: 4, scale: 0.97 }}
                          animate={{ opacity: 1, y: 0, scale: 1 }}
                          exit={{ opacity: 0, y: 4, scale: 0.97 }}
                          transition={{ duration: 0.12 }}
                          className="absolute bottom-full mb-2 right-0 w-44 bg-surface-900 border border-surface-700/60 rounded-xl shadow-2xl overflow-hidden z-50"
                        >
                          {QUALITY_OPTIONS.map(opt => (
                            <button
                              key={opt.value}
                              onClick={() => { setStreamQuality(opt.value); setQualityOpen(false) }}
                              className={clsx(
                                'w-full flex items-center justify-between px-3 py-2.5 text-xs transition-colors',
                                streamQuality === opt.value
                                  ? 'bg-accent/15 text-accent-light'
                                  : 'text-surface-300 hover:bg-surface-800'
                              )}
                            >
                              <div>
                                <div className="font-medium">{opt.label}</div>
                                <div className="text-surface-500 text-[10px]">{opt.desc}</div>
                              </div>
                              {streamQuality === opt.value && (
                                <div className="w-1.5 h-1.5 rounded-full bg-accent" />
                              )}
                            </button>
                          ))}
                        </motion.div>
                      )}
                    </AnimatePresence>
                  </div>

                  {/* Related tracks button */}
                  <button
                    onClick={() => setRelatedOpen(o => !o)}
                    className={clsx(
                      'p-1.5 rounded-lg transition-colors',
                      relatedOpen
                        ? 'text-accent-light bg-accent/15'
                        : 'text-surface-500 hover:text-surface-300'
                    )}
                    title="Related tracks"
                  >
                    <Sparkles className="w-4 h-4" />
                  </button>

                  {/* BPM Controls */}
                  <BPMControls />

				  {/* SC link */}
                  {currentTrack.permalink_url && (
                    <a
                      href={currentTrack.permalink_url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="p-1.5 text-surface-500 hover:text-surface-300 transition-colors"
                      title="Open on SoundCloud"
                    >
                      <ExternalLink className="w-4 h-4" />
                    </a>
                  )}

                  {/* Download link */}
                  {currentTrack.downloadable && (
                    <a
                      href={currentTrack.download_url || currentTrack.permalink_url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="p-1.5 text-surface-500 hover:text-accent-light transition-colors"
                      title={currentTrack.download_url ? "Download track" : "Download on SoundCloud"}
                    >
                      <Download className="w-4 h-4" />
                    </a>
                  )}

                  {/* Close */}
                  <button
                    onClick={close}
                    className="p-1.5 text-surface-500 hover:text-surface-300 transition-colors"
                  >
                    <X className="w-4 h-4" />
                  </button>
                </div>
              </div>
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </>
  )
}
