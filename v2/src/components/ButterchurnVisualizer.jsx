import { useEffect, useRef, useState, useCallback } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { X, Shuffle, ChevronLeft, ChevronRight } from 'lucide-react'
import { useStore } from '../store'
import { vizConfig } from '../pages/DebugViz'

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
  
  // Load base presets
  const basePkg = window.butterchurnPresets?.default ?? window.butterchurnPresets
  if (basePkg) {
    const basePresets = typeof basePkg.getPresets === 'function' ? basePkg.getPresets() : basePkg
    shared.presets = { ...basePresets }
  }
  
  // Load extra presets if available
  const extraPkg = window.butterchurnPresetsExtra?.default ?? window.butterchurnPresetsExtra
  if (extraPkg) {
    const extraPresets = typeof extraPkg.getPresets === 'function' ? extraPkg.getPresets() : extraPkg
    shared.presets = { ...shared.presets, ...extraPresets }
  }
  
  shared.keys = Object.keys(shared.presets)
  console.log(`[Butterchurn] Loaded ${shared.keys.length} presets`)
}

function applySharedPreset(idx, ...vizInstances) {
  const keys = shared.keys
  if (!keys.length) {
    console.log('[Butterchurn] Cannot apply preset - no keys loaded')
    return ''
  }
  const key = keys[idx]
  console.log(`[Butterchurn] Applying preset: ${key} (idx=${idx})`)
  for (const viz of vizInstances) {
    if (viz) {
      viz.loadPreset(shared.presets[key], 2.0)
      console.log(`[Butterchurn] Preset loaded into visualizer`)
    } else {
      console.log('[Butterchurn] Visualizer instance is null!')
    }
  }
  notifyPresetChange(key)
  return key
}

// ── createViz helper ─────────────────────────────────────────────────────────
function createViz(canvas, audioCtx, analyserNode, dpr = 1, speed = 1.0) {
  if (!canvas || !audioCtx || !analyserNode) return null
  let bc = window.butterchurn?.default ?? window.butterchurn
  if (!bc || typeof bc.createVisualizer !== 'function') {
    console.warn('[Butterchurn] library not ready'); return null
  }
  try {
    // Use vizConfig for quality settings
    const viz = bc.createVisualizer(audioCtx, canvas, {
      width: canvas.width, 
      height: canvas.height, 
      pixelRatio: dpr, 
      textureRatio: vizConfig?.textureRatio ?? 1,
      meshWidth: vizConfig?.meshWidth ?? 48,
      meshHeight: vizConfig?.meshHeight ?? 36,
      speed: speed, // Animation speed multiplier
    })
    viz.connectAudio(analyserNode)
    loadSharedPresets()
    if (shared.keys.length > 0) {
      viz.loadPreset(shared.presets[shared.keys[shared.idx]], 2.0) // 2 second smooth blend
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
    const viz = createViz(canvas, audioCtx, analyserNode, dpr, 1.0)
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
  const mockAudioRef = useRef(null) // Store mock audio to prevent recreation
  const initializedRef = useRef(false) // Track if we've initialized in this component instance
  const [presetName, setPresetName] = useState('')
  const [presetCount, setPresetCount] = useState(0)

  // ── Stats ──────────────────────────────────────────────────────────────────
  const [stats, setStats] = useState(null) // { fps, frameMs, resW, resH }
  const statsRef = useRef({ times: [], lastFlush: 0, resW: 0, resH: 0 })

  // ── Auto-cycle ────────────────────────────────────────────────────────────
  const [autoCycle, setAutoCycle] = useState(true) // Auto-cycle through presets
  const [cycleInterval, setCycleInterval] = useState(15000) // 15 seconds default
  const autoCycleRef = useRef(null)

  // ── Animation speed ─────────────────────────────────────────────────────────
  const [animSpeed, setAnimSpeed] = useState(0.6) // 0.6 = 60% speed (slower)

  // Subscribe to preset name updates
  useEffect(() => {
    const cb = (name) => setPresetName(name)
    shared.listeners.add(cb)
    return () => shared.listeners.delete(cb)
  }, [])

  const next   = useCallback(() => {
    if (!shared.keys.length) {
      console.log('[Butterchurn] No presets loaded yet')
      return
    }
    shared.idx = (shared.idx + 1) % shared.keys.length
    console.log(`[Butterchurn] Next preset: ${shared.keys[shared.idx]}`)
    applySharedPreset(shared.idx, vizRef.current)
  }, [])

  const prev   = useCallback(() => {
    if (!shared.keys.length) {
      console.log('[Butterchurn] No presets loaded yet')
      return
    }
    shared.idx = (shared.idx - 1 + shared.keys.length) % shared.keys.length
    console.log(`[Butterchurn] Previous preset: ${shared.keys[shared.idx]}`)
    applySharedPreset(shared.idx, vizRef.current)
  }, [])

  const random = useCallback(() => {
    if (!shared.keys.length) {
      console.log('[Butterchurn] No presets loaded yet')
      return
    }
    shared.idx = Math.floor(Math.random() * shared.keys.length)
    console.log(`[Butterchurn] Random preset: ${shared.keys[shared.idx]}`)
    applySharedPreset(shared.idx, vizRef.current)
  }, [])

  // Initialize visualizer only once when component mounts
  useEffect(() => {
    if (!open) {
      cancelAnimationFrame(rafRef.current)
      rafRef.current = null
      // Reset stats so stale timestamps don't corrupt the next open
      statsRef.current.times    = []
      statsRef.current.lastFlush = 0
      setStats(null)
      return
    }

    // If RAF loop is running, visualizer is already set up - just continue
    if (rafRef.current) {
      return
    }

    const canvas = canvasRef.current
    if (!canvas) return

    // Create mock audio context/analyser if none exists (for testing without playing audio)
    let ctx = audioCtx
    let analyser = analyserNode
    if (!ctx || !analyser) {
      // Reuse existing mock audio if available
      if (!mockAudioRef.current) {
        ctx = new (window.AudioContext || window.webkitAudioContext)()
        analyser = ctx.createAnalyser()
        analyser.fftSize = 512
        
        // Create oscillator for mock audio data
        const oscillator = ctx.createOscillator()
        const gain = ctx.createGain()
        oscillator.connect(gain)
        gain.connect(analyser)
        oscillator.start()
        gain.gain.value = 0 // Silent but generates data for visualizer
        
        // Resume audio context (required by Chrome autoplay policy)
        ctx.resume().catch(() => {})
        
        // Store in ref to prevent recreation
        mockAudioRef.current = { ctx, analyser, oscillator, gain }
      } else {
        ctx = mockAudioRef.current.ctx
        analyser = mockAudioRef.current.analyser
      }
      
      // Fill with synthetic data
      const dataArray = new Uint8Array(analyser.frequencyBinCount)
      for (let i = 0; i < dataArray.length; i++) {
        dataArray[i] = Math.random() * 255
      }
      analyser.getByteFrequencyData = () => dataArray
      analyser.getByteTimeDomainData = () => dataArray
    }

    const MAX_W = vizConfig?.maxWidth ?? 1920
    const MAX_H = vizConfig?.maxHeight ?? 1080
    const dpr = window.devicePixelRatio || 1
    const panel = panelRef.current
    const panelW = panel?.clientWidth  || window.innerWidth
    const panelH = panel?.clientHeight || window.innerHeight
    const scale  = Math.min(dpr, MAX_W / panelW, MAX_H / panelH)
    canvas.width  = Math.round(panelW * scale)
    canvas.height = Math.round(panelH * scale)
    statsRef.current.resW = canvas.width
    statsRef.current.resH = canvas.height

    // Only create visualizer if it doesn't exist (don't recreate on StrictMode double-invoke)
    let viz = vizRef.current
    if (!viz) {
      viz = createViz(canvas, ctx, analyser, 1, animSpeed)
      if (!viz) return
      vizRef.current = viz
      
      // Sync to current shared preset
      if (shared.keys.length > 0) {
        viz.loadPreset(shared.presets[shared.keys[shared.idx]], 2.0) // 2 second smooth blend
        setPresetName(shared.keys[shared.idx])
      }
      setPresetCount(shared.keys.length)
    } else {
      // Update speed if visualizer already exists and supports setSpeed
      if (typeof viz.setSpeed === 'function') {
        viz.setSpeed(animSpeed)
      }
    }

    // Re-measure after enter animation settles (in case panel was 0×0 at mount)
    const resizeTimer = setTimeout(() => {
      const panel = panelRef.current
      const pw = panel ? panel.clientWidth  : window.innerWidth
      const ph = panel ? panel.clientHeight : window.innerHeight
      if (pw > 0 && ph > 0) {
        const s = Math.min(dpr, MAX_W / pw, MAX_H / ph)
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
      const s  = Math.min(dpr, MAX_W / pw, MAX_H / ph)
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

    // Frame skipping for speed control when setSpeed is not available (e.g., minified builds)
    let lastRenderTime = performance.now()
    let accumulatedFrameTime = 0
    const usesSetSpeed = typeof viz.setSpeed === 'function'
    
    const loop = (now) => {
      rafRef.current = requestAnimationFrame(loop)
      
      if (usesSetSpeed) {
        // Native speed control via butterchurn
        vizRef.current?.render()
      } else {
        // Fallback: frame skipping based on speed
        const delta = now - lastRenderTime
        lastRenderTime = now
        accumulatedFrameTime += delta
        
        // At 100% speed, render every frame (~16ms)
        // At 50% speed, render every 32ms
        // At 10% speed, render every 160ms
        const targetInterval = 16.667 / animSpeed
        
        if (accumulatedFrameTime >= targetInterval) {
          accumulatedFrameTime = Math.max(0, accumulatedFrameTime - targetInterval)
          vizRef.current?.render()
        }
      }
    }
    loop(performance.now())

    return () => {
      clearTimeout(resizeTimer)
      cancelAnimationFrame(rafRef.current)
      rafRef.current = null
      window.removeEventListener('resize', onResize)
      window.removeEventListener('keydown', onKey)
    }
  }, [open, analyserNode, audioCtx, next, prev, random, onClose, animSpeed])
  
  // Update animation speed when slider changes
  useEffect(() => {
    if (vizRef.current && typeof vizRef.current.setSpeed === 'function') {
      vizRef.current.setSpeed(animSpeed)
    }
  }, [animSpeed])
  
  // Cleanup on unmount
  useEffect(() => {
    return () => {
      vizRef.current = null
      initializedRef.current = false
      
      // Clean up mock audio if we created it
      if (mockAudioRef.current) {
        try {
          mockAudioRef.current.oscillator?.stop()
          mockAudioRef.current.ctx?.close()
        } catch (e) {
          // Ignore cleanup errors
        }
        mockAudioRef.current = null
      }
    }
  }, [])

  useEffect(() => {
    if (isPlaying && audioCtx?.state === 'suspended') audioCtx.resume().catch(() => {})
  }, [isPlaying, audioCtx])

  // Auto-cycle through presets
  useEffect(() => {
    if (!open || !autoCycle) {
      if (autoCycleRef.current) {
        clearInterval(autoCycleRef.current)
        autoCycleRef.current = null
      }
      return
    }

    // Start auto-cycling
    autoCycleRef.current = setInterval(() => {
      if (shared.keys.length > 0) {
        shared.idx = Math.floor(Math.random() * shared.keys.length)
        applySharedPreset(shared.idx, vizRef.current)
      }
    }, cycleInterval)

    return () => {
      if (autoCycleRef.current) {
        clearInterval(autoCycleRef.current)
        autoCycleRef.current = null
      }
    }
  }, [open, autoCycle, cycleInterval])

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
          className="fixed top-14 left-0 right-0 bottom-16 z-[45] bg-black overflow-hidden"
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
            <div className="absolute top-16 right-6 flex flex-col items-end gap-2">
              <p className="text-white/30 text-[10px]">{presetCount} presets</p>
              <button
                onClick={() => setAutoCycle(!autoCycle)}
                className={`text-[10px] px-2 py-1 rounded transition-colors ${
                  autoCycle ? 'bg-accent/50 text-white' : 'bg-white/10 text-white/50'
                }`}
                title={autoCycle ? 'Auto-cycle ON' : 'Auto-cycle OFF'}
              >
                {autoCycle ? 'Auto' : 'Manual'}
              </button>
              {/* Speed control */}
              <div className="flex flex-col items-end gap-1 mt-1">
                <span className="text-white/40 text-[10px]">Speed: {(animSpeed * 100).toFixed(0)}%</span>
                <input
                  type="range"
                  min="0.1"
                  max="2.0"
                  step="0.1"
                  value={animSpeed}
                  onChange={(e) => setAnimSpeed(parseFloat(e.target.value))}
                  className="w-20 h-1 bg-white/20 rounded-lg appearance-none cursor-pointer accent-accent"
                  title="Animation speed"
                />
              </div>
            </div>
          )}

          {/* Stats are now displayed in debug menu */}
        </motion.div>
      )}
    </AnimatePresence>
  )
}

// Default export kept for any remaining imports
export default function ButterchurnVisualizer(props) {
  return <ButterchurnMini {...props} />
}
