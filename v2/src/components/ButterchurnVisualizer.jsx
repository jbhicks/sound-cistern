import { useEffect, useRef, useState, useCallback } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { X, Shuffle, ChevronLeft, ChevronRight, Settings, Sliders } from 'lucide-react'
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
function createViz(canvas, audioCtx, analyserNode, options = {}) {
  return createVizWithDimensions(canvas, audioCtx, analyserNode, canvas.width, canvas.height, options)
}

function createVizWithDimensions(canvas, audioCtx, analyserNode, width, height, options = {}) {
  if (!canvas || !audioCtx || !analyserNode) {
    console.log('[Butterchurn] Missing required params:', { hasCanvas: !!canvas, hasAudioCtx: !!audioCtx, hasAnalyser: !!analyserNode })
    return null
  }
  let bc = window.butterchurn?.default ?? window.butterchurn
  if (!bc || typeof bc.createVisualizer !== 'function') {
    console.warn('[Butterchurn] library not ready'); return null
  }
  
  // Check canvas dimensions and layout
  const rect = canvas.getBoundingClientRect()
  console.log('[Butterchurn] Canvas state:', {
    widthAttr: canvas.width,
    heightAttr: canvas.height,
    cssWidth: rect.width,
    cssHeight: rect.height,
    desiredWidth: width,
    desiredHeight: height,
    inDOM: document.contains(canvas)
  })
  
  // Verify we have valid dimensions
  if (width <= 0 || height <= 0) {
    console.warn('[Butterchurn] Invalid dimensions:', width, height)
    return null
  }
  
  // Check if canvas already has a WebGL context that might be lost/invalid
  const existingGL = canvas.getContext('webgl2') || canvas.getContext('webgl')
  if (existingGL) {
    // Check if context is lost - if so, we can't use this canvas
    if (existingGL.isContextLost()) {
      console.warn('[Butterchurn] Canvas has lost WebGL context, cannot reuse')
      return null
    }
  }
  
  try {
    // Use provided options or fall back to vizConfig defaults
    const textureRatio = options.textureRatio ?? vizConfig?.textureRatio ?? 1
    const meshWidth = options.meshWidth ?? vizConfig?.meshWidth ?? 48
    const meshHeight = options.meshHeight ?? meshWidth
    
    // Round to nearest power of 2 for better WebGL compatibility on macOS
    const pow2Width = Math.pow(2, Math.ceil(Math.log2(Math.max(16, width))))
    const pow2Height = Math.pow(2, Math.ceil(Math.log2(Math.max(16, height))))
    
    // Ensure canvas dimensions are set BEFORE butterchurn creates its WebGL context
    // This is critical - butterchurn reads canvas.width/height internally
    canvas.width = pow2Width
    canvas.height = pow2Height
    
    console.log('[Butterchurn] Creating visualizer:', { 
      requested: `${width}x${height}`, 
      pow2: `${pow2Width}x${pow2Height}`,
      textureRatio, 
      meshWidth, 
      meshHeight 
    })
    
    // Create visualizer - butterchurn creates its own WebGL context
    // Use the original dimensions in options, not the pow2 ones
    const viz = bc.createVisualizer(audioCtx, canvas, {
      width: width, 
      height: height, 
      pixelRatio: 1, 
      textureRatio: textureRatio,
      meshWidth: meshWidth,
      meshHeight: meshHeight,
    })
    
    console.log('[Butterchurn] Visualizer created successfully')
    viz.connectAudio(analyserNode)
    loadSharedPresets()
    if (shared.keys.length > 0) {
      viz.loadPreset(shared.presets[shared.keys[shared.idx]], 2.0)
    }
    return viz
  } catch (e) {
    console.error('[Butterchurn] createVisualizer failed:', e)
    return null
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
  const [canvasKey, setCanvasKey] = useState(0) // Force new canvas element on failure

  // Keep pausedRef current so the RAF loop always reads the latest value.
  // When paused, also cancel the RAF so the GL context isn't competing.
  useEffect(() => {
    pausedRef.current = paused
    if (paused) cancelAnimationFrame(rafRef.current)
  }, [paused])

  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas || !analyserNode || !audioCtx || initRef.current) return
    
    // Guard against zero dimensions
    if (width <= 0 || height <= 0) {
      console.warn('[Butterchurn] Mini canvas has zero dimensions')
      return
    }
    
    // Check if butterchurn library is loaded
    const bc = window.butterchurn?.default ?? window.butterchurn
    if (!bc || typeof bc.createVisualizer !== 'function') {
      console.warn('[Butterchurn] Library not loaded yet, deferring...')
      return
    }
    
    const dpr = window.devicePixelRatio || 1
    const w = Math.max(1, Math.round(width  * dpr))
    const h = Math.max(1, Math.round(height * dpr))
    const viz = createVizWithDimensions(canvas, audioCtx, analyserNode, w, h)
    if (!viz) {
      // Failed to create visualizer - likely due to lost WebGL context
      // Force a new canvas element by updating the key
      console.log('[Butterchurn] Failed to create mini visualizer, will retry with fresh canvas')
      setTimeout(() => setCanvasKey(k => k + 1), 100)
      return
    }
    vizRef.current  = viz
    initRef.current = true

    const loop = () => {
      rafRef.current = requestAnimationFrame(loop)
      if (!pausedRef.current) vizRef.current?.render()
    }
    loop()

    return () => {
      cancelAnimationFrame(rafRef.current)
      // Try to clean up WebGL context properly
      try {
        if (vizRef.current) {
          // Some versions of butterchurn have a dispose method
          if (typeof vizRef.current.dispose === 'function') {
            vizRef.current.dispose()
          }
        }
      } catch (e) {
        // Ignore cleanup errors
      }
      vizRef.current  = null
      initRef.current = false
    }
  }, [analyserNode, audioCtx, width, height, canvasKey]) // eslint-disable-line react-hooks/exhaustive-deps

  // Subscribe to preset changes from fullscreen
  const [, forceRender] = useState(0)
  useEffect(() => {
    const cb = () => forceRender(n => n + 1)
    shared.listeners.add(cb)
    return () => shared.listeners.delete(cb)
  }, [])

  return (
    <canvas
      key={canvasKey}
      ref={canvasRef}
      width={width}
      height={height}
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
  const [canvasKey, setCanvasKey] = useState(0) // Force new canvas on repeated failures
  const failureCountRef = useRef(0)

  // ── Stats ──────────────────────────────────────────────────────────────────
  const [stats, setStats] = useState(null) // { fps, frameMs, resW, resH }
  const statsRef = useRef({ times: [], lastFlush: 0, resW: 0, resH: 0 })

  // ── Auto-cycle ────────────────────────────────────────────────────────────
  const [autoCycle, setAutoCycle] = useState(true) // Auto-cycle through presets
  const [cycleInterval, setCycleInterval] = useState(15000) // 15 seconds default
  const autoCycleRef = useRef(null)

  // ── Settings Panel ────────────────────────────────────────────────────────
  const [showSettings, setShowSettings] = useState(false)
  const [vizSpeed, setVizSpeed] = useState(1.0) // Animation speed multiplier
  const [meshQuality, setMeshQuality] = useState(vizConfig?.meshWidth ?? 48) // Mesh resolution
  const [texQuality, setTexQuality] = useState(vizConfig?.textureRatio ?? 1) // Texture quality
  const [maxRes, setMaxRes] = useState(720) // Max resolution height

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

    // Reset failure count when opening
    failureCountRef.current = 0

    // If RAF loop is running, visualizer is already set up - just continue
    if (rafRef.current) {
      return
    }

    const canvas = canvasRef.current
    const panel = panelRef.current
    if (!canvas || !panel) return

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

    // Use user settings for max resolution (aspect ratio preserved)
    const aspectRatio = panel.clientWidth / Math.max(1, panel.clientHeight)
    const MAX_H = maxRes
    const MAX_W = Math.round(maxRes * aspectRatio)
    const dpr = Math.min(window.devicePixelRatio || 1, 1.5)  // Cap DPR to avoid large textures

    // Track all cleanup functions
    const cleanupFns = []

    // Initialize with proper dimensions using ResizeObserver
    const tryInitViz = () => {
      // Already initialized?
      if (vizRef.current) return true
      
      const panelW = panel.clientWidth
      const panelH = panel.clientHeight
      
      // Guard against zero dimensions - wait for layout
      if (panelW <= 0 || panelH <= 0) {
        return false
      }
      
      // Check if canvas is actually in the DOM and has layout
      const rect = canvas.getBoundingClientRect()
      if (rect.width === 0 || rect.height === 0) {
        console.log('[Butterchurn] Canvas not yet laid out, waiting...')
        return false
      }
      
      // Get WebGL max texture size BEFORE creating visualizer
      // butterchurn internally multiplies dimensions by textureRatio
      const textureRatio = vizConfig?.textureRatio ?? 1
      let gl = canvas.getContext('webgl') || canvas.getContext('webgl2')
      const maxTextureSize = gl ? gl.getParameter(gl.MAX_TEXTURE_SIZE) : 4096
      
      // Calculate safe dimensions accounting for texture ratio
      const safeMaxW = Math.floor(maxTextureSize / textureRatio)
      const safeMaxH = Math.floor(maxTextureSize / textureRatio)
      
      const scale  = Math.min(dpr, MAX_W / panelW, MAX_H / panelH, safeMaxW / panelW, safeMaxH / panelH)
      const rawW = Math.max(16, Math.round(panelW * scale))
      const rawH = Math.max(16, Math.round(panelH * scale))
      
      // Store the raw dimensions for display
      statsRef.current.resW = rawW
      statsRef.current.resH = rawH
      
      console.log(`[Butterchurn] Creating visualizer at ${rawW}x${rawH} (maxTextureSize: ${maxTextureSize}, textureRatio: ${textureRatio})`)

      // Create visualizer with current quality settings (power-of-2 handled internally)
      const viz = createVizWithDimensions(canvas, ctx, analyser, rawW, rawH, {
        meshWidth: meshQuality,
        meshHeight: meshQuality,
        textureRatio: texQuality
      })
      if (!viz) {
        failureCountRef.current++
        if (failureCountRef.current > 3) {
          console.log('[Butterchurn] Too many failures, forcing canvas refresh')
          setCanvasKey(k => k + 1)
          failureCountRef.current = 0
        }
        return false
      }
      failureCountRef.current = 0
      vizRef.current = viz
      
      // Sync to current shared preset
      if (shared.keys.length > 0) {
        viz.loadPreset(shared.presets[shared.keys[shared.idx]], 2.0)
        setPresetName(shared.keys[shared.idx])
      }
      setPresetCount(shared.keys.length)
      
      // Start animation loop after a brief delay to let butterchurn initialize
      setTimeout(() => {
        if (!vizRef.current) return
        
        let lastFrameTime = performance.now()
        let frameCount = 0
        
        // Ensure canvas has valid dimensions before starting (use power-of-2 for macOS)
        if (canvas.width <= 0 || canvas.height <= 0) {
          console.log('[Butterchurn] Fixing canvas dimensions before render loop')
          const w = Math.max(16, statsRef.current.resW || 640)
          const h = Math.max(16, statsRef.current.resH || 360)
          canvas.width = Math.pow(2, Math.ceil(Math.log2(w)))
          canvas.height = Math.pow(2, Math.ceil(Math.log2(h)))
        }
        
        const loop = (now) => {
          rafRef.current = requestAnimationFrame(loop)
          
          // Skip rendering if visualizer isn't ready
          if (!vizRef.current) return
          
          // Ensure canvas always has valid power-of-2 dimensions
          if (canvas.width <= 0 || canvas.height <= 0) {
            const w = Math.max(16, statsRef.current.resW || 640)
            const h = Math.max(16, statsRef.current.resH || 360)
            canvas.width = Math.pow(2, Math.ceil(Math.log2(w)))
            canvas.height = Math.pow(2, Math.ceil(Math.log2(h)))
          }
          
          frameCount++
          
          // Speed control: skip frames based on speed setting
          // Speed 0.5 = render every 2nd frame, 0.25 = every 4th frame
          const speed = vizSpeed
          if (speed < 1.0) {
            const skipInterval = Math.max(1, Math.round(1 / speed))
            if (frameCount % skipInterval !== 0) return
          }
          
          // Measure actual time between frames
          lastFrameTime = now
          
          // Render the visualizer
          try {
            vizRef.current.render()
          } catch (e) {
            console.warn('[Butterchurn] Render error:', e)
          }

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
            const stats = { fps, frameMs: frameMs.toFixed(1), resW: s.resW, resH: s.resH }
            setStats(stats)
            // Export for debug page
            window._vizStats = stats
          }
        }
        loop(performance.now())
      }, 100)
      
      return true
    }

    // Set up resize handler
    const onResize = () => {
      const pw = panel?.clientWidth  ?? window.innerWidth
      const ph = panel?.clientHeight ?? window.innerHeight
      if (pw <= 0 || ph <= 0) return
      
      // Account for texture ratio
      const textureRatio = vizConfig?.textureRatio ?? 1
      let gl = canvas.getContext('webgl') || canvas.getContext('webgl2')
      const maxTextureSize = gl ? gl.getParameter(gl.MAX_TEXTURE_SIZE) : 4096
      const safeMaxW = Math.floor(maxTextureSize / textureRatio)
      const safeMaxH = Math.floor(maxTextureSize / textureRatio)
      
      const s  = Math.min(dpr, MAX_W / pw, MAX_H / ph, safeMaxW / pw, safeMaxH / ph)
      const rawW = Math.max(16, Math.round(pw * s))
      const rawH = Math.max(16, Math.round(ph * s))
      
      // Use power-of-2 dimensions for macOS WebGL compatibility
      const w = Math.pow(2, Math.ceil(Math.log2(rawW)))
      const h = Math.pow(2, Math.ceil(Math.log2(rawH)))
      
      // Don't resize if visualizer isn't ready
      if (!vizRef.current) return
      
      statsRef.current.resW = rawW
      statsRef.current.resH = rawH
      
      // Use butterchurn's setRendererSize instead of modifying canvas dimensions
      // This avoids destroying and recreating the WebGL context
      try { 
        vizRef.current.setRendererSize(w, h) 
        console.log(`[Butterchurn] Resized to ${w}x${h} (from ${rawW}x${rawH})`)
      } catch (e) {
        console.warn('[Butterchurn] Resize failed:', e)
      }
    }
    window.addEventListener('resize', onResize)
    cleanupFns.push(() => window.removeEventListener('resize', onResize))

    // Set up keyboard handler
    const onKey = (e) => {
      if (e.key === 'Escape')             onClose?.()
      if (e.key === 'ArrowRight')         next()
      if (e.key === 'ArrowLeft')          prev()
      if (e.key === 'r' || e.key === 'R') random()
    }
    window.addEventListener('keydown', onKey)
    cleanupFns.push(() => window.removeEventListener('keydown', onKey))

    // Try immediate initialization first
    if (!tryInitViz()) {
      // If that fails, use ResizeObserver to wait for dimensions
      console.log('[Butterchurn] Using ResizeObserver to wait for dimensions...')
      
      const ro = new ResizeObserver((entries) => {
        for (const entry of entries) {
          const { width, height } = entry.contentRect
          if (width > 0 && height > 0 && tryInitViz()) {
            ro.disconnect()
            break
          }
        }
      })
      ro.observe(panel)
      cleanupFns.push(() => ro.disconnect())
      
      // Also try after animation settles
      const fallbackTimer = setTimeout(() => {
        if (!vizRef.current) {
          console.log('[Butterchurn] Fallback initialization attempt...')
          tryInitViz()
        }
      }, 350)
      cleanupFns.push(() => clearTimeout(fallbackTimer))
    }

    return () => {
      cancelAnimationFrame(rafRef.current)
      rafRef.current = null
      cleanupFns.forEach(fn => fn())
    }
  }, [open, analyserNode, audioCtx, next, prev, random, onClose, canvasKey])
  
  // Cleanup on unmount
  useEffect(() => {
    return () => {
      cancelAnimationFrame(rafRef.current)
      
      // Try to clean up WebGL context properly
      try {
        if (vizRef.current) {
          if (typeof vizRef.current.dispose === 'function') {
            vizRef.current.dispose()
          }
        }
      } catch (e) {
        // Ignore cleanup errors
      }
      
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
          className="fixed top-14 left-0 right-0 bottom-[76px] z-[45] bg-black overflow-hidden"
        >
          <canvas key={canvasKey} ref={canvasRef} style={{ width: '100%', height: '100%', display: 'block' }} />

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
            <div className="absolute top-16 right-6 flex flex-col items-end gap-1">
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
            </div>
          )}

          {/* Settings Toggle Button */}
          <motion.button
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.3 }}
            onClick={() => setShowSettings(!showSettings)}
            className="absolute top-6 right-20 p-2.5 rounded-full bg-white/10 hover:bg-white/20 text-white/70 hover:text-white transition-colors backdrop-blur-sm"
            title="Settings"
          >
            <Settings className="w-5 h-5" />
          </motion.button>

          {/* Settings Panel */}
          <AnimatePresence>
            {showSettings && (
              <motion.div
                initial={{ opacity: 0, x: 20 }}
                animate={{ opacity: 1, x: 0 }}
                exit={{ opacity: 0, x: 20 }}
                transition={{ duration: 0.2 }}
                className="absolute top-20 right-6 w-64 bg-black/80 backdrop-blur-md rounded-xl border border-white/10 p-4 shadow-2xl"
              >
                <div className="flex items-center justify-between mb-4">
                  <h3 className="text-white text-sm font-semibold flex items-center gap-2">
                    <Sliders className="w-4 h-4" />
                    Visualizer Settings
                  </h3>
                  <button 
                    onClick={() => setShowSettings(false)}
                    className="text-white/50 hover:text-white text-xs"
                  >
                    ✕
                  </button>
                </div>

                {/* Animation Speed */}
                <div className="mb-4">
                  <label className="text-white/60 text-xs mb-1 block">
                    Speed: {vizSpeed.toFixed(2)}x
                  </label>
                  <input
                    type="range"
                    min="0.1"
                    max="2.0"
                    step="0.1"
                    value={vizSpeed}
                    onChange={(e) => setVizSpeed(parseFloat(e.target.value))}
                    className="w-full h-1 bg-white/20 rounded-lg appearance-none cursor-pointer accent-accent"
                  />
                  <div className="flex justify-between text-[10px] text-white/40 mt-1">
                    <span>Slow</span>
                    <span>Normal</span>
                    <span>Fast</span>
                  </div>
                </div>

                {/* Auto-cycle Interval */}
                <div className="mb-4">
                  <label className="text-white/60 text-xs mb-1 block">
                    Auto-cycle: {Math.round(cycleInterval / 1000)}s
                  </label>
                  <input
                    type="range"
                    min="5000"
                    max="60000"
                    step="5000"
                    value={cycleInterval}
                    onChange={(e) => setCycleInterval(parseInt(e.target.value))}
                    className="w-full h-1 bg-white/20 rounded-lg appearance-none cursor-pointer accent-accent"
                  />
                  <div className="flex justify-between text-[10px] text-white/40 mt-1">
                    <span>5s</span>
                    <span>30s</span>
                    <span>60s</span>
                  </div>
                </div>

                {/* Mesh Quality */}
                <div className="mb-4">
                  <label className="text-white/60 text-xs mb-1 block">
                    Mesh Quality: {meshQuality}x{meshQuality}
                  </label>
                  <input
                    type="range"
                    min="24"
                    max="96"
                    step="8"
                    value={meshQuality}
                    onChange={(e) => setMeshQuality(parseInt(e.target.value))}
                    className="w-full h-1 bg-white/20 rounded-lg appearance-none cursor-pointer accent-accent"
                  />
                  <div className="flex justify-between text-[10px] text-white/40 mt-1">
                    <span>Low</span>
                    <span>Med</span>
                    <span>High</span>
                  </div>
                  <p className="text-[10px] text-white/30 mt-1">Requires restart</p>
                </div>

                {/* Texture Quality */}
                <div className="mb-4">
                  <label className="text-white/60 text-xs mb-1 block">
                    Texture Quality: {texQuality.toFixed(1)}x
                  </label>
                  <input
                    type="range"
                    min="0.5"
                    max="2.0"
                    step="0.5"
                    value={texQuality}
                    onChange={(e) => setTexQuality(parseFloat(e.target.value))}
                    className="w-full h-1 bg-white/20 rounded-lg appearance-none cursor-pointer accent-accent"
                  />
                  <div className="flex justify-between text-[10px] text-white/40 mt-1">
                    <span>Low</span>
                    <span>Med</span>
                    <span>High</span>
                  </div>
                  <p className="text-[10px] text-white/30 mt-1">Requires restart</p>
                </div>

                {/* Max Resolution */}
                <div className="mb-2">
                  <label className="text-white/60 text-xs mb-1 block">
                    Max Resolution: {maxRes}p
                  </label>
                  <input
                    type="range"
                    min="480"
                    max="1080"
                    step="120"
                    value={maxRes}
                    onChange={(e) => setMaxRes(parseInt(e.target.value))}
                    className="w-full h-1 bg-white/20 rounded-lg appearance-none cursor-pointer accent-accent"
                  />
                  <div className="flex justify-between text-[10px] text-white/40 mt-1">
                    <span>480p</span>
                    <span>720p</span>
                    <span>1080p</span>
                  </div>
                </div>

                {/* Current Stats */}
                {stats && (
                  <div className="mt-4 pt-3 border-t border-white/10">
                    <p className="text-[10px] text-white/40 mb-1">Performance</p>
                    <div className="text-[10px] text-white/60 font-mono">
                      <div className="flex justify-between">
                        <span>FPS:</span>
                        <span>{stats.fps}</span>
                      </div>
                      <div className="flex justify-between">
                        <span>Frame:</span>
                        <span>{stats.frameMs}ms</span>
                      </div>
                      <div className="flex justify-between">
                        <span>Res:</span>
                        <span>{stats.resW}×{stats.resH}</span>
                      </div>
                    </div>
                  </div>
                )}
              </motion.div>
            )}
          </AnimatePresence>
        </motion.div>
      )}
    </AnimatePresence>
  )
}

// Default export kept for any remaining imports
export default function ButterchurnVisualizer(props) {
  return <ButterchurnMini {...props} />
}
