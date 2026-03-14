import { useEffect, useRef } from 'react'
import { useStore } from '../store'

const BAR_COUNT   = 56
const BAR_GAP     = 2
const PEAK_HOLD_MS = 1200
const PEAK_FALL_PX = 0.9
const TARGET_FPS  = 60
const FRAME_BUDGET = 1000 / TARGET_FPS  // ~16.67ms

function buildBinMap(binCount, barCount) {
  const map = []
  for (let i = 0; i < barCount; i++) {
    const lo = Math.floor(binCount * Math.pow(i / barCount, 2.2))
    const hi = Math.floor(binCount * Math.pow((i + 1) / barCount, 2.2))
    map.push({ lo: Math.max(lo, 0), hi: Math.min(hi + 1, binCount) })
  }
  return map
}

export default function EQVisualizer({ height = 40 }) {
  const canvasRef    = useRef(null)
  const rafRef       = useRef(null)
  const peaksRef     = useRef(new Float32Array(BAR_COUNT).fill(0))
  const peakTimesRef = useRef(new Float32Array(BAR_COUNT).fill(0))
  const binMapRef    = useRef(null)
  const gradientRef  = useRef(null)
  const dataRef      = useRef(null)        // reused Uint8Array — no per-frame alloc
  const lastFrameRef = useRef(0)
  const stateRef     = useRef({ analyserNode: null, isPlaying: false })

  // Keep stateRef current without re-running the draw loop effect
  // Use selectors to prevent re-renders on unrelated store changes
  const analyserNode = useStore(state => state.analyserNode)
  const isPlaying = useStore(state => state.isPlaying)
  useEffect(() => { stateRef.current = { analyserNode, isPlaying } }, [analyserNode, isPlaying])

  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas) return

    const ctx = canvas.getContext('2d')
    let cssW = 0
    let cssH = 0
    let dpr = 1

    const resize = () => {
      dpr = window.devicePixelRatio || 1
      const w = canvas.parentElement?.clientWidth || 400
      canvas.width  = Math.round(w * dpr)
      canvas.height = Math.round(height * dpr)
      cssW = canvas.width  / dpr
      cssH = canvas.height / dpr
      // Rebuild cached objects on resize
      binMapRef.current  = null
      gradientRef.current = null
      dataRef.current    = null
    }
    resize()

    const ro = new ResizeObserver(resize)
    ro.observe(canvas.parentElement || canvas)

    // Track visibility state
    let isVisible = true
    let draw = null  // Will be assigned below

    const handleVisibility = () => {
      isVisible = !document.hidden
      if (!isVisible && rafRef.current) {
        cancelAnimationFrame(rafRef.current)
        rafRef.current = null
      } else if (isVisible && !rafRef.current && draw) {
        rafRef.current = requestAnimationFrame(draw)
      }
    }
    document.addEventListener('visibilitychange', handleVisibility)

    draw = (now) => {
      // Only schedule next frame if visible
      if (isVisible) {
        rafRef.current = requestAnimationFrame(draw)
      }

      // Hard cap at TARGET_FPS — skip frames on 120/144/240Hz displays
      if (now - lastFrameRef.current < FRAME_BUDGET) return
      lastFrameRef.current = now

      const { analyserNode, isPlaying } = stateRef.current

      ctx.clearRect(0, 0, cssW, cssH)

      if (!analyserNode || !isPlaying) {
        // Idle: dim stub bars, no gradient needed
        const barW = (cssW - BAR_GAP * (BAR_COUNT - 1)) / BAR_COUNT
        ctx.fillStyle = 'rgba(139, 92, 246, 0.18)'
        for (let i = 0; i < BAR_COUNT; i++) {
          ctx.fillRect(i * (barW + BAR_GAP), cssH - 2, barW, 2)
        }
        return
      }

      // Build bin map once per canvas size
      const binCount = analyserNode.frequencyBinCount
      if (!binMapRef.current) binMapRef.current = buildBinMap(binCount, BAR_COUNT)

      // Build gradient once per canvas size
      if (!gradientRef.current) {
        const g = ctx.createLinearGradient(0, cssH, 0, 0)
        g.addColorStop(0,   'rgba(139, 92, 246, 0.9)')
        g.addColorStop(0.6, 'rgba(167,119, 252, 0.95)')
        g.addColorStop(1.0, 'rgba(236, 72, 153, 0.95)')
        gradientRef.current = g
      }

      // Reuse typed array — no allocation per frame
      if (!dataRef.current || dataRef.current.length !== binCount) {
        dataRef.current = new Uint8Array(binCount)
      }
      analyserNode.getByteFrequencyData(dataRef.current)

      const barW = (cssW - BAR_GAP * (BAR_COUNT - 1)) / BAR_COUNT
      const data = dataRef.current
      const binMap = binMapRef.current

      ctx.fillStyle = gradientRef.current

      for (let i = 0; i < BAR_COUNT; i++) {
        const { lo, hi } = binMap[i]
        let sum = 0
        const count = hi - lo
        for (let b = lo; b < hi; b++) sum += data[b]
        const avg   = count > 0 ? sum / count : 0
        const barH  = Math.max(2, Math.pow(avg / 255, 0.7) * cssH)
        const x     = i * (barW + BAR_GAP)

        ctx.fillStyle = gradientRef.current
        ctx.beginPath()
        ctx.roundRect(x, cssH - barH, barW, barH, [2, 2, 0, 0])
        ctx.fill()

        // Peak dot
        if (barH > peaksRef.current[i]) {
          peaksRef.current[i]     = barH
          peakTimesRef.current[i] = now
        } else if (now - peakTimesRef.current[i] > PEAK_HOLD_MS) {
          peaksRef.current[i] = Math.max(2, peaksRef.current[i] - PEAK_FALL_PX)
        }

        const peakY = cssH - peaksRef.current[i] - 2
        if (peakY > 0 && peakY < cssH - 4) {
          ctx.fillStyle = 'rgba(236, 72, 153, 0.85)'
          ctx.fillRect(x, peakY, barW, 2)
        }
      }
    }

    rafRef.current = requestAnimationFrame(draw)

    return () => {
      cancelAnimationFrame(rafRef.current)
      ro.disconnect()
      document.removeEventListener('visibilitychange', handleVisibility)
    }
  }, [height]) // intentionally omit analyserNode/isPlaying — read via stateRef

  return (
    <canvas ref={canvasRef} style={{ width: '100%', height, display: 'block', pointerEvents: 'none' }} />
  )
}
