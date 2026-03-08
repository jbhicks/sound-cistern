import { useEffect, useRef, useState, useCallback } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { X, Shuffle, ChevronLeft, ChevronRight } from 'lucide-react'
import { useStore } from '../store'

/**
 * ButterchurnVisualizer — two separate exported components:
 *
 *   <ButterchurnMini />      — the small artwork-slot canvas in the player bar
 *   <ButterchurnFullscreen /> — the expanded panel, rendered at player root level
 *
 * They share preset state via module-level refs so navigating presets in one
 * is reflected in the other. Only one WebGL RAF loop runs at a time:
 * the mini loop pauses while the fullscreen panel is open.
 */

// ── Module-level shared preset state ────────────────────────────────────────
// These live outside React so both components can read/write them without
// prop-drilling or a context.
const shared = {
  presets:    {},
  keys:       [],
  idx:        0,
  listeners:  new Set(), // callbacks to notify on preset change
}

function notifyPresetChange(name) {
  shared.listeners.forEach(fn => fn(name))
}

function loadSharedPresets() {
  if (shared.keys.length > 0) return
  const pkg = window.butterchurnPresets?.default ?? window.butterchurnPresets
  if (!pkg) return
  const presets = typeof pkg.getPresets === 'function' ? pkg.getPresets() : pkg
  shared.presets = presets
  shared.keys    = Object.keys(presets)
}

function applySharedPreset(idx, ...vizInstances) {
  const keys = shared.keys
  if (!keys.length) return ''
  const key = keys[idx]
  for (const viz of vizInstances) {
    if (viz) viz.loadPreset(shared.presets[key], 2.0)
  }
  notifyPresetChange(key)
  return key
}

// ── createViz helper ─────────────────────────────────────────────────────────
function createViz(canvas, audioCtx, analyserNode) {
  if (!canvas || !audioCtx || !analyserNode) return null
  let bc = window.butterchurn?.default ?? window.butterchurn
  if (!bc || typeof bc.createVisualizer !== 'function') {
    console.warn('[Butterchurn] library not ready'); return null
  }
  try {
    const viz = bc.createVisualizer(audioCtx, canvas, {
      width: canvas.width, height: canvas.height, pixelRatio: 1, textureRatio: 0.5,
    })
    viz.connectAudio(analyserNode)
    loadSharedPresets()
    if (shared.keys.length > 0) {
      viz.loadPreset(shared.presets[shared.keys[shared.idx]], 0.0)
    }
    return viz
  } catch (e) {
    console.error('[Butterchurn] createVisualizer failed:', e); return null
  }
}

// ── Mini component ────────────────────────────────────────────────────────────
export function ButterchurnMini({ width = 40, height = 40, paused = false }) {
  const { analyserNode, audioCtx } = useStore()
  const canvasRef  = useRef(null)
  const vizRef     = useRef(null)
  const rafRef     = useRef(null)
  const initRef    = useRef(false)
  const pausedRef  = useRef(paused)

  // Keep pausedRef current so the RAF loop always reads the latest value.
  // When paused, also cancel the RAF so the GL context isn't competing.
  useEffect(() => {
    pausedRef.current = paused
    if (paused) cancelAnimationFrame(rafRef.current)
  }, [paused])

  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas || !analyserNode || !audioCtx || initRef.current) return
    const dpr = window.devicePixelRatio || 1
    canvas.width  = width  * dpr
    canvas.height = height * dpr
    const viz = createViz(canvas, audioCtx, analyserNode)
    if (!viz) return
    vizRef.current  = viz
    initRef.current = true

    const loop = () => {
      rafRef.current = requestAnimationFrame(loop)
      if (!pausedRef.current) vizRef.current?.render()
    }
    loop()

    return () => {
      cancelAnimationFrame(rafRef.current)
      vizRef.current  = null
      initRef.current = false
    }
  }, [analyserNode, audioCtx, width, height]) // eslint-disable-line react-hooks/exhaustive-deps

  // Subscribe to preset changes from fullscreen
  const [, forceRender] = useState(0)
  useEffect(() => {
    const cb = () => forceRender(n => n + 1)
    shared.listeners.add(cb)
    return () => shared.listeners.delete(cb)
  }, [])

  return (
    <canvas
      ref={canvasRef}
      style={{ width, height, display: 'block' }}
      className="rounded-lg"
    />
  )
}

// ── Fullscreen component ──────────────────────────────────────────────────────
export function ButterchurnFullscreen({ open, onClose }) {
  const { analyserNode, audioCtx, currentTrack, isPlaying } = useStore()
  const canvasRef  = useRef(null)
  const panelRef   = useRef(null)
  const vizRef     = useRef(null)
  const rafRef     = useRef(null)
  const [presetName, setPresetName] = useState('')
  const [presetCount, setPresetCount] = useState(0)

  // ── Stats ──────────────────────────────────────────────────────────────────
  const [stats, setStats] = useState(null) // { fps, frameMs, resW, resH }
  const statsRef = useRef({ times: [], lastFlush: 0, resW: 0, resH: 0 })

  // Subscribe to preset name updates
  useEffect(() => {
    const cb = (name) => setPresetName(name)
    shared.listeners.add(cb)
    return () => shared.listeners.delete(cb)
  }, [])

  const next   = useCallback(() => {
    if (!shared.keys.length) return
    shared.idx = (shared.idx + 1) % shared.keys.length
    applySharedPreset(shared.idx, vizRef.current)
  }, [])

  const prev   = useCallback(() => {
    if (!shared.keys.length) return
    shared.idx = (shared.idx - 1 + shared.keys.length) % shared.keys.length
    applySharedPreset(shared.idx, vizRef.current)
  }, [])

  const random = useCallback(() => {
    if (!shared.keys.length) return
    shared.idx = Math.floor(Math.random() * shared.keys.length)
    applySharedPreset(shared.idx, vizRef.current)
  }, [])

  useEffect(() => {
    if (!open) {
      cancelAnimationFrame(rafRef.current)
      vizRef.current = null
      // Reset stats so stale timestamps don't corrupt the next open
      statsRef.current.times    = []
      statsRef.current.lastFlush = 0
      setStats(null)
      return
    }

    const canvas = canvasRef.current
    if (!canvas || !analyserNode || !audioCtx) return

    // Cap render resolution — butterchurn is shader-heavy and doesn't need
    // native 4K/5K. 1280×720 looks great and runs at 60fps on any GPU.
    const MAX_W = 1280
    const MAX_H = 720
    const panel = panelRef.current
    const panelW = panel?.clientWidth  || window.innerWidth
    const panelH = panel?.clientHeight || window.innerHeight
    const scale  = Math.min(1, MAX_W / panelW, MAX_H / panelH)
    canvas.width  = Math.round(panelW * scale)
    canvas.height = Math.round(panelH * scale)
    statsRef.current.resW = canvas.width
    statsRef.current.resH = canvas.height

    const viz = createViz(canvas, audioCtx, analyserNode)
    if (!viz) return
    vizRef.current = viz

    // Sync to current shared preset
    if (shared.keys.length > 0) {
      viz.loadPreset(shared.presets[shared.keys[shared.idx]], 0.0)
      setPresetName(shared.keys[shared.idx])
    }
    setPresetCount(shared.keys.length)

    // Re-measure after enter animation settles (in case panel was 0×0 at mount)
    const resizeTimer = setTimeout(() => {
      const panel = panelRef.current
      const pw = panel ? panel.clientWidth  : window.innerWidth
      const ph = panel ? panel.clientHeight : window.innerHeight
      if (pw > 0 && ph > 0) {
        const s = Math.min(1, MAX_W / pw, MAX_H / ph)
        const w = Math.round(pw * s)
        const h = Math.round(ph * s)
        if (canvas.width !== w || canvas.height !== h) {
          canvas.width  = w
          canvas.height = h
          try { viz.setRendererSize(w, h) } catch {}
        }
      }
    }, 250)

    const onResize = () => {
      const panel = panelRef.current
      const pw = panel?.clientWidth  ?? window.innerWidth
      const ph = panel?.clientHeight ?? window.innerHeight
      const s  = Math.min(1, MAX_W / pw, MAX_H / ph)
      const w  = Math.round(pw * s)
      const h  = Math.round(ph * s)
      canvas.width  = w
      canvas.height = h
      statsRef.current.resW = w
      statsRef.current.resH = h
      try { vizRef.current?.setRendererSize(w, h) } catch {}
    }
    window.addEventListener('resize', onResize)

    const onKey = (e) => {
      if (e.key === 'Escape')             onClose?.()
      if (e.key === 'ArrowRight')         next()
      if (e.key === 'ArrowLeft')          prev()
      if (e.key === 'r' || e.key === 'R') random()
    }
    window.addEventListener('keydown', onKey)

    const loop = (now) => {
      rafRef.current = requestAnimationFrame(loop)
      vizRef.current?.render()

      // Track frame times for stats
      const s = statsRef.current
      s.times.push(now)
      // Keep a rolling 60-frame window
      if (s.times.length > 60) s.times.shift()
      // Flush to React state ~4×/sec so the display stays readable
      if (now - s.lastFlush > 250 && s.times.length >= 2) {
        s.lastFlush = now
        const span = s.times[s.times.length - 1] - s.times[0]
        const fps  = Math.round((s.times.length - 1) / (span / 1000))
        const frameMs = span / (s.times.length - 1)
        setStats({ fps, frameMs: frameMs.toFixed(1), resW: s.resW, resH: s.resH })
      }
    }
    loop(performance.now())

    return () => {
      clearTimeout(resizeTimer)
      cancelAnimationFrame(rafRef.current)
      window.removeEventListener('resize', onResize)
      window.removeEventListener('keydown', onKey)
      vizRef.current = null
    }
  }, [open, analyserNode, audioCtx, next, prev, random, onClose])

  useEffect(() => {
    if (isPlaying && audioCtx?.state === 'suspended') audioCtx.resume().catch(() => {})
  }, [isPlaying, audioCtx])

  return (
    <AnimatePresence>
      {open && (
        <motion.div
          ref={panelRef}
          key="bc-fullscreen"
          initial={{ opacity: 0, scaleY: 0.97 }}
          animate={{ opacity: 1, scaleY: 1 }}
          exit={{ opacity: 0, scaleY: 0.97 }}
          style={{ transformOrigin: 'bottom center' }}
          transition={{ duration: 0.2, ease: 'easeOut' }}
          className="fixed top-14 left-0 right-0 bottom-[76px] z-[45] bg-black overflow-hidden"
        >
          <canvas ref={canvasRef} style={{ width: '100%', height: '100%', display: 'block' }} />

          {currentTrack && (
            <motion.div
              initial={{ opacity: 0, x: -16 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ delay: 0.25 }}
              className="absolute top-6 left-6 flex items-center gap-3 pointer-events-none"
            >
              {currentTrack.artwork_url && (
                <img
                  src={currentTrack.artwork_url.replace(/-t\d+x\d+/g, '-t200x200')}
                  alt=""
                  className="w-12 h-12 rounded-lg shadow-2xl object-cover opacity-80"
                />
              )}
              <div>
                <p className="text-white font-semibold text-sm drop-shadow-lg leading-tight">{currentTrack.track_title}</p>
                <p className="text-white/60 text-xs drop-shadow-lg mt-0.5">{currentTrack.artist_name}</p>
              </div>
            </motion.div>
          )}

          <motion.div
            initial={{ opacity: 0, y: 12 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.25 }}
            className="absolute bottom-8 left-1/2 -translate-x-1/2 flex flex-col items-center gap-2"
          >
            <div className="flex items-center gap-2">
              <button onClick={prev}   className="p-2 rounded-full bg-white/10 hover:bg-white/20 text-white/70 hover:text-white transition-colors backdrop-blur-sm" title="Previous (←)"><ChevronLeft  className="w-4 h-4" /></button>
              <button onClick={random} className="p-2 rounded-full bg-white/10 hover:bg-white/20 text-white/70 hover:text-white transition-colors backdrop-blur-sm" title="Random (R)">   <Shuffle      className="w-4 h-4" /></button>
              <button onClick={next}   className="p-2 rounded-full bg-white/10 hover:bg-white/20 text-white/70 hover:text-white transition-colors backdrop-blur-sm" title="Next (→)">      <ChevronRight className="w-4 h-4" /></button>
            </div>
            {presetName && <p className="text-white/40 text-[10px] max-w-xs text-center truncate px-2">{presetName}</p>}
          </motion.div>

          <motion.button
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.2 }}
            onClick={onClose}
            className="absolute top-6 right-6 p-2.5 rounded-full bg-white/10 hover:bg-white/20 text-white/70 hover:text-white transition-colors backdrop-blur-sm"
            title="Close (Esc)"
          >
            <X className="w-5 h-5" />
          </motion.button>

          {presetCount > 0 && (
            <p className="absolute top-16 right-6 text-white/30 text-[10px]">{presetCount} presets</p>
          )}

          {/* Render stats — bottom right */}
          {stats && (
            <div className="absolute bottom-6 right-6 text-right pointer-events-none select-none">
              <p style={{ fontFamily: 'monospace', fontSize: 11, lineHeight: '1.6' }}
                 className={stats.fps >= 55 ? 'text-green-400/70' : stats.fps >= 30 ? 'text-yellow-400/70' : 'text-red-400/80'}>
                {stats.fps} fps
              </p>
              <p style={{ fontFamily: 'monospace', fontSize: 10, lineHeight: '1.6' }}
                 className="text-white/25">
                {stats.frameMs} ms
              </p>
              <p style={{ fontFamily: 'monospace', fontSize: 10, lineHeight: '1.6' }}
                 className="text-white/25">
                {stats.resW}×{stats.resH}
              </p>
            </div>
          )}
        </motion.div>
      )}
    </AnimatePresence>
  )
}

// Default export kept for any remaining imports
export default function ButterchurnVisualizer(props) {
  return <ButterchurnMini {...props} />
}
