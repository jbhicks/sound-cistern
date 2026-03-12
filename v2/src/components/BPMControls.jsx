import { useState, useEffect } from 'react'
import { motion } from 'framer-motion'
import { Zap, RotateCcw } from 'lucide-react'
import { useStore } from '../store'

export default function BPMControls() {
  const { currentTrack, audioEl, playbackRate, setPlaybackRate } = useStore()
  const [showControls, setShowControls] = useState(false)

  const originalBPM = currentTrack?.bpm || 0
  const playbackBPM = originalBPM > 0 ? Math.round(originalBPM * playbackRate) : 0

  // Update audio playback rate
  useEffect(() => {
    if (audioEl) {
      audioEl.playbackRate = playbackRate
    }
  }, [playbackRate, audioEl])

  // Reset when track changes
  useEffect(() => {
    setPlaybackRate(1)
  }, [currentTrack?.track_id, setPlaybackRate])

  const handleSpeedChange = (value) => {
    const rate = parseFloat(value)
    setPlaybackRate(rate)
  }

  const resetSpeed = () => {
    setPlaybackRate(1)
  }

  const percentChange = Math.round((playbackRate - 1) * 100)

  if (!currentTrack) return null

  return (
    <div className="relative">
      {/* Toggle button */}
      <button
        onClick={() => setShowControls(!showControls)}
        className={`
          flex items-center gap-1.5 px-2 py-1 rounded-lg text-xs font-medium transition-all border
          ${showControls 
            ? 'bg-accent/15 border-accent/40 text-accent-light' 
            : 'bg-surface-800/60 border-surface-700/40 text-surface-400 hover:text-surface-200'}
        `}
        title="Adjust playback speed"
      >
        <Zap className="w-3.5 h-3.5" />
        <span className="hidden lg:inline">
          {playbackRate !== 1 ? `${percentChange > 0 ? '+' : ''}${percentChange}%` : 'Speed'}
        </span>
      </button>

      {/* Dropdown panel */}
      {showControls && (
        <motion.div
          initial={{ opacity: 0, y: 4, scale: 0.97 }}
          animate={{ opacity: 1, y: 0, scale: 1 }}
          exit={{ opacity: 0, y: 4, scale: 0.97 }}
          transition={{ duration: 0.12 }}
          className="absolute bottom-full mb-2 right-0 w-64 bg-surface-900 border border-surface-700/60 rounded-xl shadow-2xl overflow-hidden z-50 p-4"
        >
          {/* BPM Display */}
          <div className="flex items-center justify-center gap-4 mb-4">
            <div className="text-center">
              <div className="text-2xl font-bold text-surface-300 tabular-nums">
                {originalBPM > 0 ? Math.round(originalBPM) : '--'}
              </div>
              <div className="text-[10px] text-surface-500 uppercase tracking-wider">Original BPM</div>
            </div>
            <div className="text-surface-600">→</div>
            <div className="text-center">
              <div className="text-2xl font-bold text-accent-light tabular-nums">
                {playbackBPM > 0 ? playbackBPM : '--'}
              </div>
              <div className="text-[10px] text-surface-500 uppercase tracking-wider">Now Playing</div>
            </div>
          </div>

          {/* Speed slider */}
          <div className="mb-4">
            <div className="flex justify-between text-[10px] text-surface-500 mb-1">
              <span>Slower</span>
              <span className={percentChange !== 0 ? 'text-accent-light font-medium' : ''}>
                {playbackRate.toFixed(2)}x
              </span>
              <span>Faster</span>
            </div>
            <input
              type="range"
              min="0.8"
              max="1.2"
              step="0.01"
              value={playbackRate}
              onChange={(e) => handleSpeedChange(e.target.value)}
              className="w-full accent-accent h-1.5 cursor-pointer"
            />
            <div className="flex justify-between text-[10px] text-surface-600 mt-1">
              <span>-20%</span>
              <span>Normal</span>
              <span>+20%</span>
            </div>
          </div>

          {/* Reset button */}
          <button
            onClick={resetSpeed}
            disabled={playbackRate === 1}
            className="w-full flex items-center justify-center gap-1.5 px-3 py-2 rounded-lg bg-surface-800 hover:bg-surface-700 disabled:opacity-40 disabled:cursor-not-allowed text-xs text-surface-300 transition-colors"
          >
            <RotateCcw className="w-3 h-3" />
            Reset to Normal Speed
          </button>

          {/* Note */}
          <p className="mt-3 text-[10px] text-surface-500 text-center">
            Adjust playback speed (±20%)<br />
            Changes both tempo and pitch
          </p>
        </motion.div>
      )}
    </div>
  )
}
