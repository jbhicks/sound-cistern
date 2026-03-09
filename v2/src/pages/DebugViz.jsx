import { useState, useEffect, useRef, useCallback } from 'react'
import { useStore } from '../store'
import { ButterchurnFullscreen } from '../components/ButterchurnVisualizer'

// Shared stats context
let globalStats = { fps: 0, frameMs: 0, resW: 0, resH: 0 }

// Global quality config - can be modified on the fly
export const vizConfig = {
  textureRatio: 4,
  meshWidth: 128,
  meshHeight: 96,
  maxWidth: 3840,
  maxHeight: 2160,
}

// Debug visualizer with incremental React integration testing
export default function DebugViz() {
  const store = useStore()
  const [mode, setMode] = useState('component') // 'pure' | 'store' | 'component'
  const [vizOpen, setVizOpen] = useState(true)
  const [stats, setStats] = useState({ fps: 0, frameMs: 0 })
  const [configVersion, setConfigVersion] = useState(0) // Force re-render when config changes
  
  // Update stats from global or component
  useEffect(() => {
    // Reset viz open state when switching modes
    setVizOpen(mode === 'component')
    
    const interval = setInterval(() => {
      // In component mode, read from window._vizStats (set by ButterchurnFullscreen)
      // In other modes, read from globalStats
      if (mode === 'component' && window._vizStats) {
        setStats(window._vizStats)
      } else {
        setStats({ ...globalStats })
      }
    }, 500)
    return () => clearInterval(interval)
  }, [mode])

  // Config update helpers
  const updateConfig = (key, value) => {
    vizConfig[key] = parseFloat(value)
    setConfigVersion(v => v + 1)
    // Restart visualizer to apply changes
    setVizOpen(false)
    setTimeout(() => setVizOpen(true), 100)
  }

  return (
    <div className="fixed inset-0 bg-black">
      {/* Mode selector */}
      <div className="absolute top-4 right-4 z-50 flex flex-col gap-2 bg-black/80 p-4 rounded text-white max-w-xs">
        <div className="text-sm font-semibold mb-2">Integration Level:</div>
        
        <button 
          onClick={() => setMode('pure')}
          className={`px-3 py-2 rounded text-left ${mode === 'pure' ? 'bg-accent' : 'bg-white/20'}`}
        >
          <div className="font-medium">1. Pure Butterchurn</div>
          <div className="text-xs opacity-70">No React wrapper</div>
        </button>
        
        <button 
          onClick={() => setMode('store')}
          className={`px-3 py-2 rounded text-left ${mode === 'store' ? 'bg-accent' : 'bg-white/20'}`}
        >
          <div className="font-medium">2. + Store Connection</div>
          <div className="text-xs opacity-70">Connect to audio/analyser</div>
        </button>
        
        <button 
          onClick={() => setMode('component')}
          className={`px-3 py-2 rounded text-left ${mode === 'component' ? 'bg-accent' : 'bg-white/20'}`}
        >
          <div className="font-medium">3. + ButterchurnFullscreen</div>
          <div className="text-xs opacity-70">Full React component</div>
        </button>
        
        <div className="mt-4 pt-4 border-t border-white/20 font-mono text-sm">
          <div className="text-green-400 text-lg">{stats.fps || 0} FPS</div>
          <div className="text-gray-400">{stats.frameMs || 0}ms/frame</div>
          <div className="text-gray-500 text-xs mt-1">{stats.resW || 0}×{stats.resH || 0}</div>
        </div>
        
        {/* Audio Status */}
        <div className="mt-4 pt-4 border-t border-white/20">
          <div className="text-sm font-semibold mb-2">Audio Source:</div>
          <div className={`text-xs ${store.analyserNode ? 'text-green-400' : 'text-yellow-400'}`}>
            {store.analyserNode ? '🔊 Real Audio (Active)' : '🔇 Mock Audio (No Track Playing)'}
          </div>
          {store.currentTrack && (
            <div className="text-xs text-gray-400 mt-1 truncate">
              Now: {store.currentTrack.track_title}
            </div>
          )}
          {!store.analyserNode && (
            <a 
              href="/stream" 
              className="block mt-2 px-3 py-1.5 bg-accent hover:bg-accent-hover text-white text-xs rounded text-center transition-colors"
            >
              Play a Track →
            </a>
          )}
        </div>
        
        {/* Quality Config Panel */}
        {mode === 'component' && (
          <div className="mt-4 pt-4 border-t border-white/20">
            <div className="text-sm font-semibold mb-2">Quality Settings:</div>
            
            <div className="space-y-2">
              <div>
                <label className="text-xs text-gray-400">Texture Ratio</label>
                <select 
                  value={vizConfig.textureRatio}
                  onChange={(e) => updateConfig('textureRatio', e.target.value)}
                  className="w-full mt-1 px-2 py-1 bg-white/20 rounded text-sm"
                >
                  <option value="0.5">0.5 (Low)</option>
                  <option value="1">1 (Normal)</option>
                  <option value="2">2 (High)</option>
                  <option value="4">4 (Ultra)</option>
                </select>
              </div>
              
              <div>
                <label className="text-xs text-gray-400">Mesh Width</label>
                <select 
                  value={vizConfig.meshWidth}
                  onChange={(e) => updateConfig('meshWidth', e.target.value)}
                  className="w-full mt-1 px-2 py-1 bg-white/20 rounded text-sm"
                >
                  <option value="32">32 (Low)</option>
                  <option value="48">48 (Normal)</option>
                  <option value="96">96 (High)</option>
                  <option value="128">128 (Ultra)</option>
                </select>
              </div>
              
              <div>
                <label className="text-xs text-gray-400">Mesh Height</label>
                <select 
                  value={vizConfig.meshHeight}
                  onChange={(e) => updateConfig('meshHeight', e.target.value)}
                  className="w-full mt-1 px-2 py-1 bg-white/20 rounded text-sm"
                >
                  <option value="24">24 (Low)</option>
                  <option value="36">36 (Normal)</option>
                  <option value="72">72 (High)</option>
                  <option value="96">96 (Ultra)</option>
                </select>
              </div>
              
              <div>
                <label className="text-xs text-gray-400">Max Resolution</label>
                <select 
                  value={`${vizConfig.maxWidth}x${vizConfig.maxHeight}`}
                  onChange={(e) => {
                    const [w, h] = e.target.value.split('x').map(Number)
                    vizConfig.maxWidth = w
                    vizConfig.maxHeight = h
                    setConfigVersion(v => v + 1)
                    setVizOpen(false)
                    setTimeout(() => setVizOpen(true), 100)
                  }}
                  className="w-full mt-1 px-2 py-1 bg-white/20 rounded text-sm"
                >
                  <option value="1920x1080">1920×1080 (FHD)</option>
                  <option value="2560x1440">2560×1440 (QHD)</option>
                  <option value="3840x2160">3840×2160 (4K)</option>
                </select>
              </div>
            </div>
          </div>
        )}
      </div>
      
      {/* Test modes */}
      {mode === 'pure' && <PureViz />}
      {mode === 'store' && <StoreViz store={store} />}
      {/* Only render component in component mode */}
      {mode === 'component' && vizOpen && (
        <ButterchurnFullscreen 
          open={true} 
          onClose={() => {}}
          key={configVersion} // Force re-mount when config changes
        />
      )}
    </div>
  )
}

// Hook to track FPS
function useFPSTracker() {
  const rafRef = useRef(null)
  const frameCountRef = useRef(0)
  const lastTimeRef = useRef(performance.now())
  const lastFrameTimeRef = useRef(performance.now())
  
  useEffect(() => {
    const track = () => {
      rafRef.current = requestAnimationFrame(track)
      
      const now = performance.now()
      frameCountRef.current++
      
      // Calculate frame time
      const frameTime = now - lastFrameTimeRef.current
      lastFrameTimeRef.current = now
      
      // Update stats every 500ms
      if (now - lastTimeRef.current > 500) {
        const elapsed = now - lastTimeRef.current
        const fps = Math.round((frameCountRef.current * 1000) / elapsed)
        
        globalStats = { 
          fps, 
          frameMs: frameTime.toFixed(1),
          resW: window.innerWidth,
          resH: window.innerHeight
        }
        
        frameCountRef.current = 0
        lastTimeRef.current = now
      }
    }
    
    track()
    
    return () => {
      cancelAnimationFrame(rafRef.current)
    }
  }, [])
}

// Pure butterchurn - no React wrapper
function PureViz() {
  const canvasRef = useRef(null)
  const vizRef = useRef(null)
  const rafRef = useRef(null)
  
  useFPSTracker()
  
  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas) return
    
    canvas.width = window.innerWidth
    canvas.height = window.innerHeight
    
    const bc = window.butterchurn?.default ?? window.butterchurn
    if (!bc) {
      console.error('[DebugViz] Butterchurn not loaded')
      return
    }
    
    const audioCtx = new (window.AudioContext || window.webkitAudioContext)()
    const viz = bc.createVisualizer(audioCtx, canvas, {
      width: canvas.width,
      height: canvas.height,
      pixelRatio: 1,
      textureRatio: 1
    })
    
    vizRef.current = viz
    
    const pkg = window.butterchurnPresets?.default ?? window.butterchurnPresets
    if (pkg) {
      const presets = typeof pkg.getPresets === 'function' ? pkg.getPresets() : pkg
      const keys = Object.keys(presets)
      if (keys.length > 0) {
        viz.loadPreset(presets[keys[0]], 0)
      }
    }
    
    const loop = () => {
      rafRef.current = requestAnimationFrame(loop)
      vizRef.current?.render()
    }
    loop()
    
    return () => {
      cancelAnimationFrame(rafRef.current)
    }
  }, [])
  
  return <canvas ref={canvasRef} className="w-full h-full" />
}

// Butterchurn + Store connection (audio/analyser)
function StoreViz({ store }) {
  const canvasRef = useRef(null)
  const vizRef = useRef(null)
  const rafRef = useRef(null)
  
  useFPSTracker()
  
  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas) return
    
    canvas.width = window.innerWidth
    canvas.height = window.innerHeight
    
    const bc = window.butterchurn?.default ?? window.butterchurn
    if (!bc) return
    
    const audioCtx = store.audioCtx || new (window.AudioContext || window.webkitAudioContext)()
    const viz = bc.createVisualizer(audioCtx, canvas, {
      width: canvas.width,
      height: canvas.height,
      pixelRatio: 1,
      textureRatio: 1
    })
    
    vizRef.current = viz
    
    if (store.analyserNode) {
      viz.connectAudio(store.analyserNode)
    }
    
    const pkg = window.butterchurnPresets?.default ?? window.butterchurnPresets
    if (pkg) {
      const presets = typeof pkg.getPresets === 'function' ? pkg.getPresets() : pkg
      const keys = Object.keys(presets)
      if (keys.length > 0) {
        viz.loadPreset(presets[keys[0]], 0)
      }
    }
    
    const loop = () => {
      rafRef.current = requestAnimationFrame(loop)
      vizRef.current?.render()
    }
    loop()
    
    return () => {
      cancelAnimationFrame(rafRef.current)
    }
  }, [store.audioCtx, store.analyserNode])
  
  return <canvas ref={canvasRef} className="w-full h-full" />
}
