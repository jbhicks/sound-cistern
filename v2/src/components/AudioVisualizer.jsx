import { useEffect, useRef } from 'react'
import { useStore } from '../store'

/**
 * AudioVisualizer — canvas-based frequency bar visualizer using Web Audio API.
 * Connects to the shared <audio> element stored in Zustand.
 *
 * Props:
 *   width  — canvas pixel width  (default 200)
 *   height — canvas pixel height (default 48)
 *   bars   — number of frequency bars (default 32)
 *   color  — bar fill color (default '#8b5cf6')
 *   className — extra CSS classes on the canvas wrapper
 *   style  — inline styles on wrapper
 */
export default function AudioVisualizer({
  width = 200,
  height = 48,
  bars = 32,
  color = '#8b5cf6',
  secondaryColor = '#ff6b9d',
  className = '',
  style = {},
}) {
  const canvasRef = useRef(null)
  const animFrameRef = useRef(null)
  const analyserRef = useRef(null)
  const sourceRef = useRef(null)
  const ctxRef = useRef(null) // AudioContext
  const dataRef = useRef(null)
  const smoothRef = useRef(null) // smoothed values for animation

  const { audioEl, isPlaying } = useStore()

  // Build (or re-use) the AudioContext + AnalyserNode connected to the audio element
  useEffect(() => {
    if (!audioEl) return

    // Reuse existing context if we already wired this audio element
    if (!ctxRef.current) {
      ctxRef.current = new (window.AudioContext || window.webkitAudioContext)()
    }
    const ctx = ctxRef.current

    // Only create source once per audio element
    if (!sourceRef.current || sourceRef.current.mediaElement !== audioEl) {
      try {
        // Disconnect old source if any
        if (sourceRef.current) {
          try { sourceRef.current.disconnect() } catch {}
        }
        sourceRef.current = ctx.createMediaElementSource(audioEl)
        sourceRef.current.connect(ctx.destination)
      } catch (err) {
        // InvalidStateError if source was already created elsewhere — skip
        if (err.name !== 'InvalidStateError') throw err
      }
    }

    if (!analyserRef.current) {
      const analyser = ctx.createAnalyser()
      analyser.fftSize = 256
      analyser.smoothingTimeConstant = 0.78
      analyserRef.current = analyser
      sourceRef.current.connect(analyser)
    }

    dataRef.current = new Uint8Array(analyserRef.current.frequencyBinCount)
    smoothRef.current = new Float32Array(bars).fill(0)

    return () => {
      // Don't tear down on every re-render — keep the graph alive
    }
  }, [audioEl, bars])

  // Resume AudioContext on play (browsers require user gesture first)
  useEffect(() => {
    if (isPlaying && ctxRef.current?.state === 'suspended') {
      ctxRef.current.resume().catch(() => {})
    }
  }, [isPlaying])

  // Animation loop
  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas) return

    const dpr = window.devicePixelRatio || 1
    canvas.width = width * dpr
    canvas.height = height * dpr
    canvas.style.width = `${width}px`
    canvas.style.height = `${height}px`

    const gc = canvas.getContext('2d')
    gc.scale(dpr, dpr)

    const step = width / bars
    const gap = Math.max(1, step * 0.15)
    const barW = step - gap

    const DECAY = 0.08  // how fast bars fall when not playing
    const RISE  = 0.55  // how fast bars rise to new value

    function draw() {
      animFrameRef.current = requestAnimationFrame(draw)
      gc.clearRect(0, 0, width, height)

      const analyser = analyserRef.current
      const data = dataRef.current
      const smooth = smoothRef.current

      if (!analyser || !data || !smooth) return

      analyser.getByteFrequencyData(data)

      // Map FFT bins → bar count (use lower ~40% of spectrum — most musical content)
      const usableBins = Math.floor(data.length * 0.4)
      const binStep = usableBins / bars

      for (let i = 0; i < bars; i++) {
        const binStart = Math.floor(i * binStep)
        const binEnd = Math.ceil((i + 1) * binStep)
        let sum = 0
        for (let b = binStart; b < binEnd && b < data.length; b++) sum += data[b]
        const avg = sum / Math.max(1, binEnd - binStart)
        const target = (avg / 255) * height

        // Smooth: fast rise, slow decay
        if (target > smooth[i]) {
          smooth[i] += (target - smooth[i]) * RISE
        } else {
          smooth[i] += (target - smooth[i]) * DECAY
        }

        const barH = Math.max(2, smooth[i])
        const x = i * step + gap / 2
        const y = height - barH

        // Gradient per bar: accent → vapor-pink at higher amplitudes
        const t = smooth[i] / height // 0–1 normalised amplitude
        const r1 = parseInt(color.slice(1, 3), 16)
        const g1 = parseInt(color.slice(3, 5), 16)
        const b1 = parseInt(color.slice(5, 7), 16)
        const r2 = parseInt(secondaryColor.slice(1, 3), 16)
        const g2 = parseInt(secondaryColor.slice(3, 5), 16)
        const b2 = parseInt(secondaryColor.slice(5, 7), 16)
        const r = Math.round(r1 + (r2 - r1) * t)
        const g = Math.round(g1 + (g2 - g1) * t)
        const b = Math.round(b1 + (b2 - b1) * t)

        const grad = gc.createLinearGradient(0, y, 0, height)
        grad.addColorStop(0, `rgba(${r},${g},${b},0.9)`)
        grad.addColorStop(1, `rgba(${r1},${g1},${b1},0.4)`)
        gc.fillStyle = grad

        // Rounded rect bars
        const radius = Math.min(barW / 2, 3)
        gc.beginPath()
        gc.moveTo(x + radius, y)
        gc.lineTo(x + barW - radius, y)
        gc.quadraticCurveTo(x + barW, y, x + barW, y + radius)
        gc.lineTo(x + barW, height - radius)
        gc.quadraticCurveTo(x + barW, height, x + barW - radius, height)
        gc.lineTo(x + radius, height)
        gc.quadraticCurveTo(x, height, x, height - radius)
        gc.lineTo(x, y + radius)
        gc.quadraticCurveTo(x, y, x + radius, y)
        gc.closePath()
        gc.fill()
      }
    }

    draw()
    return () => cancelAnimationFrame(animFrameRef.current)
  }, [width, height, bars, color, secondaryColor])

  return (
    <div className={className} style={style}>
      <canvas ref={canvasRef} />
    </div>
  )
}
