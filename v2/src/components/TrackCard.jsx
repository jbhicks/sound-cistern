import { useState, forwardRef, useRef, useEffect } from 'react'
import { motion } from 'framer-motion'
import { Heart, Play, Pause, Disc3, Download, ListPlus, Check, Clock } from 'lucide-react'
import clsx from 'clsx'
import { useStore } from '../store'

// Use store's genre color system

export function formatDuration(ms) {
  if (!ms) return '0:00'
  const s = Math.floor(ms / 1000)
  const m = Math.floor(s / 60)
  const sec = s % 60
  return `${m}:${sec.toString().padStart(2, '0')}`
}

export function formatCompact(n) {
  if (!n) return '0'
  if (n >= 1_000_000) return (n / 1_000_000).toFixed(1) + 'M'
  if (n >= 1_000) return (n / 1_000).toFixed(1) + 'K'
  return String(n)
}

export function upgradeArtwork(url) {
  if (!url) return null
  return url.replace(/-t\d+x\d+/g, '-t500x500').replace('-large.', '-t500x500.')
}

function AddToPlaylistButton({ track }) {
  const { playlists } = useStore()
  const [open, setOpen] = useState(false)
  const [added, setAdded] = useState(null) // playlist id that was just added
  const ref = useRef(null)

  // Close on outside click
  useEffect(() => {
    if (!open) return
    const handler = (e) => {
      if (ref.current && !ref.current.contains(e.target)) setOpen(false)
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [open])

  const handleAdd = async (e, playlist) => {
    e.stopPropagation()
    const res = await fetch(`/api/playlists/${playlist.id}/tracks`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify({ track_id: track.track_id }),
    })
    if (res.ok) {
      setAdded(playlist.id)
      setTimeout(() => {
        setAdded(null)
        setOpen(false)
      }, 1000)
    }
  }

  if (!playlists || playlists.length === 0) return null

  return (
    <div ref={ref} className="relative" onClick={e => e.stopPropagation()}>
      <button
        onClick={(e) => { e.stopPropagation(); setOpen(o => !o) }}
        className={clsx(
          'p-1.5 rounded-full transition-colors',
          open ? 'text-accent' : 'text-surface-500 hover:text-surface-300'
        )}
        title="Add to playlist"
      >
        <ListPlus className="w-3.5 h-3.5" />
      </button>

      {open && (
        <div className="absolute bottom-full right-0 mb-1 w-44 bg-surface-800 border border-surface-700/60 rounded-xl shadow-2xl overflow-hidden z-50">
          <div className="px-3 py-2 border-b border-surface-700/40">
            <p className="text-[10px] font-semibold text-surface-400 uppercase tracking-wider">Add to playlist</p>
          </div>
          <div className="max-h-40 overflow-y-auto">
            {playlists.map(pl => (
              <button
                key={pl.id}
                onClick={(e) => handleAdd(e, pl)}
                className={clsx(
                  'w-full flex items-center justify-between gap-2 px-3 py-2 text-left text-xs hover:bg-surface-700/60 transition-colors',
                  added === pl.id ? 'text-green-400' : 'text-surface-200'
                )}
              >
                <span className="truncate">{pl.name}</span>
                {added === pl.id && <Check className="w-3 h-3 flex-shrink-0 text-green-400" />}
              </button>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}

const TrackCard = forwardRef(function TrackCard({ track, index, viewMode = 'grid', isNew = false, isTopArtist = false }, ref) {
  const [flipped, setFlipped] = useState(false)
  const [imgError, setImgError] = useState(false)
  const { currentTrack, isPlaying, playbackRate, playTrack, favorites, toggleFavorite, getGenreColor } = useStore()

  const isActive = currentTrack?.track_id === track.track_id
  const isThisPlaying = isActive && isPlaying
  const isFav = favorites.has(track.track_id)
  const art = imgError ? null : upgradeArtwork(track.artwork_url)

  // Helper to get genre style from store
  const getGenreStyle = (genre) => {
    const color = getGenreColor(genre)
    return {
      bg: color.bg,
      border: color.border,
      text: color.text
    }
  }

  const handleCardClick = () => {
    playTrack(track)
  }

  const handleFav = (e) => {
    e.stopPropagation()
    toggleFavorite(track.track_id)
  }

  const handlePlayBtn = (e) => {
    e.stopPropagation()
    playTrack(track)
  }

  // ── LIST VIEW ──────────────────────────────────────────────────
  if (viewMode === 'list') {
    return (
      <motion.div
        ref={ref}
        layout
        initial={isNew ? { opacity: 0, x: 40, backgroundColor: 'rgba(139,92,246,0.15)' } : { opacity: 0, x: -10 }}
        animate={{ opacity: 1, x: 0, backgroundColor: 'rgba(139,92,246,0)' }}
        exit={{ opacity: 0, x: -10 }}
        transition={isNew
          ? { duration: 0.4, ease: 'easeOut', backgroundColor: { duration: 1.5, delay: 0.3 } }
          : { delay: Math.min(index * 0.015, 0.3) }}
        onClick={handleCardClick}
        className={clsx(
          'flex items-center gap-3 p-3 rounded-xl cursor-pointer group transition-all duration-200',
          isActive
            ? 'bg-accent/10 border border-accent/30'
            : 'hover:bg-surface-800/60 border border-transparent hover:border-surface-700/40'
        )}
        style={isTopArtist && !isActive ? { animation: 'topArtistGlow 3s ease-in-out infinite' } : undefined}
      >
        <div className="relative w-12 h-12 rounded-lg overflow-hidden flex-shrink-0 bg-surface-700">
          {art
            ? <img src={art} alt="" className="w-full h-full object-cover" onError={() => setImgError(true)} />
            : <MusicIcon />
          }
          <div className={clsx(
            'absolute inset-0 bg-black/50 flex items-center justify-center transition-opacity',
            isActive ? 'opacity-100' : 'opacity-0 group-hover:opacity-100'
          )}>
            {isThisPlaying ? <Pause className="w-5 h-5 text-white" /> : <Play className="w-5 h-5 text-white ml-0.5" />}
          </div>
        </div>

        <div className="flex-1 min-w-0">
          <p className={clsx('font-medium truncate text-sm', isActive ? 'text-accent-light' : 'text-surface-100')}>
            {track.track_title}
          </p>
          <p className="text-xs text-surface-400 truncate">{track.artist_name}</p>
        </div>

        {track.genre && (
          <span 
            className="hidden md:block px-2 py-0.5 text-[10px] rounded-md flex-shrink-0 backdrop-blur-md"
            style={{
              backgroundColor: getGenreStyle(track.genre).bg,
              border: `1px solid ${getGenreStyle(track.genre).border}`,
              color: getGenreStyle(track.genre).text
            }}
          >
            {track.genre}
          </span>
        )}

        {track.playback_count > 0 && (
          <span className="hidden lg:flex items-center gap-0.5 text-xs text-cyan-400/80 flex-shrink-0">
            <Play className="w-3 h-3" />
            {formatCompact(track.playback_count)}
          </span>
        )}
        {track.favoritings_count > 0 && (
          <span className="hidden lg:flex items-center gap-0.5 text-xs text-pink-400/80 flex-shrink-0">
            <Heart className="w-3 h-3" />
            {formatCompact(track.favoritings_count)}
          </span>
        )}
        {track.bpm > 0 && (
          <span className="hidden lg:flex items-center gap-0.5 text-xs text-amber-400/80 flex-shrink-0">
            <span 
              className="text-[10px]"
              style={{
                animation: `bpmPulse ${60 / track.bpm}s ease-in-out infinite`,
                textShadow: '0 0 4px rgba(251, 191, 36, 0.5)'
              }}
            >
              ♫
            </span>
            {Math.round(track.bpm)}
          </span>
        )}

        <span className="flex items-center justify-end gap-0.5 text-xs text-surface-500 w-12 text-right flex-shrink-0">
          <Clock className="w-3 h-3" />
          {formatDuration(track.track_duration)}
        </span>

        <button onClick={handleFav} className={clsx(
          'p-1.5 rounded-lg transition-colors flex-shrink-0',
          isFav ? 'text-red-400' : 'text-surface-600 hover:text-surface-300 opacity-0 group-hover:opacity-100'
        )}>
          <Heart className={clsx('w-4 h-4', isFav && 'fill-current')} />
        </button>
      </motion.div>
    )
  }

  // ── GRID VIEW — flip card ──────────────────────────────────────
  return (
    <motion.div
      ref={ref}
      layout
      initial={isNew ? { opacity: 0, scale: 0.85, y: -16 } : { opacity: 0, scale: 0.9 }}
      animate={{ opacity: 1, scale: 1, y: 0 }}
      exit={{ opacity: 0, scale: 0.9 }}
      transition={isNew
        ? { type: 'spring', stiffness: 320, damping: 22 }
        : { delay: Math.min(index * 0.02, 0.4), type: 'spring', stiffness: 280, damping: 24 }}
      className="relative cursor-pointer"
      style={{ perspective: '1000px' }}
      onClick={handleCardClick}
    >
      {/* Top-artist ambient glow */}
      {isTopArtist && !isActive && (
        <div
          className="absolute inset-0 rounded-2xl pointer-events-none z-20"
          style={{ animation: 'topArtistGlow 3s ease-in-out infinite' }}
        />
      )}

      {/* New-track highlight ring — fades out after entry */}
      {isNew && (
        <motion.div
          className="absolute inset-0 rounded-2xl pointer-events-none z-20"
          initial={{ boxShadow: '0 0 0 2px rgba(139,92,246,0.7)' }}
          animate={{ boxShadow: '0 0 0 2px rgba(139,92,246,0)' }}
          transition={{ duration: 1.8, delay: 0.2, ease: 'easeOut' }}
        />
      )}

      {/* Flip container */}
      <div
        className="relative transition-transform duration-500"
        style={{
          transformStyle: 'preserve-3d',
          transform: flipped ? 'rotateY(180deg)' : 'rotateY(0deg)',
          aspectRatio: '1 / 1',
        }}
      >
        {/* ── FRONT ── */}
        <div
          className={clsx(
            'absolute inset-0 rounded-2xl overflow-hidden',
            'border',
            isActive ? 'border-accent/50' : 'border-surface-700/30',
            'group'
          )}
          style={{ backfaceVisibility: 'hidden' }}
        >
          {/* Static artwork — shown when not playing */}
          <div className={clsx(
            'absolute inset-0 transition-opacity duration-500',
            isThisPlaying ? 'opacity-0' : 'opacity-100'
          )}>
            {art
              ? <img src={art} alt={track.track_title} className="w-full h-full object-cover transition-transform duration-500 group-hover:scale-105" loading="lazy" onError={() => setImgError(true)} />
              : <div className="w-full h-full bg-surface-800 flex items-center justify-center"><MusicIcon large /></div>
            }
          </div>

          {/* Spinning vinyl — shown when playing */}
          <div className={clsx(
            'absolute inset-0 flex items-center justify-center bg-surface-950 transition-opacity duration-500',
            isThisPlaying ? 'opacity-100' : 'opacity-0 pointer-events-none'
          )}>
            <div
              className="relative rounded-full"
              style={{
                width: '86%',
                aspectRatio: '1/1',
                background: 'repeating-radial-gradient(circle at center, transparent 0px, transparent 3px, rgba(40,40,40,0.85) 3px, rgba(40,40,40,0.85) 4px), radial-gradient(circle at center, #1a1a1a 35%, #050505 100%)',
                animation: isThisPlaying ? `vinylSpin ${1.8 / playbackRate}s linear infinite` : 'none',
                boxShadow: '0 0 50px rgba(139,92,246,0.25), 0 0 20px rgba(139,92,246,0.1), 0 8px 40px rgba(0,0,0,0.8)',
              }}
            >
              {/* Shine */}
              <div className="absolute inset-0 rounded-full" style={{ background: 'linear-gradient(135deg, rgba(255,255,255,0.07) 0%, transparent 55%)' }} />
              {/* Cover art label */}
              <div
                className="absolute rounded-full overflow-hidden border-2 border-surface-700/40"
                style={{ width: '38%', height: '38%', top: '31%', left: '31%' }}
              >
                {art
                  ? <img src={art} alt="" className="w-full h-full object-cover" />
                  : <div className="w-full h-full bg-gradient-to-br from-accent to-vapor-pink" />
                }
              </div>
              {/* Center hole */}
              <div className="absolute rounded-full bg-surface-950" style={{ width: '7%', height: '7%', top: '46.5%', left: '46.5%' }} />
            </div>
          </div>

          {/* Top row: fav + flip buttons */}
          <div className="absolute top-0 left-0 right-0 flex items-center justify-between p-2 z-10">
            <button
              onClick={handleFav}
              className={clsx(
                'p-1.5 rounded-full backdrop-blur-sm transition-all duration-200',
                isFav ? 'bg-red-500/80 text-white opacity-100' : 'bg-black/40 text-white/70 opacity-0 group-hover:opacity-100'
              )}
            >
              <Heart className={clsx('w-3.5 h-3.5', isFav && 'fill-current')} />
            </button>
            <button
              onClick={e => { e.stopPropagation(); setFlipped(f => !f) }}
              className="p-1.5 rounded-full bg-black/40 backdrop-blur-sm text-white/70 opacity-0 group-hover:opacity-100 transition-all duration-200"
              title="Flip"
            >
              <Disc3 className="w-3.5 h-3.5" />
            </button>
          </div>

          {/* Centered play button — only on hover when not playing */}
          {!isThisPlaying && (
            <div className="absolute inset-0 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity duration-200 z-10">
              <div className="w-12 h-12 rounded-full bg-white/20 backdrop-blur-sm border border-white/30 flex items-center justify-center shadow-xl">
                <Play className="w-5 h-5 text-white ml-0.5" />
              </div>
            </div>
          )}

          {/* Bottom gradient + info overlay */}
          <div className="absolute bottom-0 left-0 right-0 z-10"
            style={{ background: 'linear-gradient(to top, rgba(0,0,0,0.9) 0%, rgba(0,0,0,0.6) 40%, transparent 100%)' }}
          >
            <div className="px-3 pt-8 pb-3">
              {/* Title - larger and more prominent */}
              <p className={clsx('font-semibold text-[15px] truncate leading-tight', isActive ? 'text-accent-light' : 'text-white')}>
                {track.track_title}
              </p>
              
              {/* Artist - clear hierarchy */}
              <p className="text-xs text-white/60 truncate mt-0.5">{track.artist_name}</p>

              {/* Metadata row: Genre + Stats + Duration */}
              <div className="flex items-center justify-between mt-2">
                <div className="flex items-center gap-2 flex-1 min-w-0">
                  {/* Genre pill - compact, integrated */}
                  {track.genre && (
                    <span 
                      className="px-1.5 py-0.5 rounded-md backdrop-blur-md text-[9px] font-medium truncate flex-shrink-0"
                      style={{
                        backgroundColor: getGenreStyle(track.genre).bg,
                        border: `1px solid ${getGenreStyle(track.genre).border}`,
                        color: getGenreStyle(track.genre).text
                      }}
                    >
                      {track.genre}
                    </span>
                  )}
                  
                  {/* Stats with semantic colors */}
                  <div className="flex items-center gap-2 text-[10px]">
                    {track.playback_count > 0 && (
                      <span className="flex items-center gap-0.5 text-cyan-300/80">
                        <Play className="w-2.5 h-2.5" />
                        {formatCompact(track.playback_count)}
                      </span>
                    )}
                    {track.favoritings_count > 0 && (
                      <span className="flex items-center gap-0.5 text-pink-300/80">
                        <Heart className="w-2.5 h-2.5" />
                        {formatCompact(track.favoritings_count)}
                      </span>
                    )}
                    {track.downloadable && (
                      <a
                        href={track.download_url || track.permalink_url}
                        target="_blank"
                        rel="noopener noreferrer"
                        onClick={e => e.stopPropagation()}
                        className="flex items-center gap-0.5 text-emerald-300/80 hover:text-emerald-200 transition-colors"
                        title={track.download_url ? "Download track" : "Download on SoundCloud"}
                      >
                        <Download className="w-2.5 h-2.5" />
                      </a>
                    )}
                    {/* BPM badge with breathing glow */}
                    {track.bpm > 0 && (
                      <span className="flex items-center gap-0.5 text-amber-300/80">
                        <span 
                          className="text-[8px]"
                          style={{
                            animation: `bpmPulse ${60 / track.bpm}s ease-in-out infinite`,
                            textShadow: '0 0 4px rgba(251, 191, 36, 0.5)'
                          }}
                        >
                          ♫
                        </span>
                        {Math.round(track.bpm)}
                      </span>
                    )}
                  </div>
                </div>
                
                {/* Duration - right aligned with icon */}
                <span className="flex items-center gap-0.5 text-[10px] text-white/50 tabular-nums font-medium flex-shrink-0 ml-2">
                  <Clock className="w-2.5 h-2.5" />
                  {formatDuration(track.track_duration)}
                </span>
              </div>
            </div>
          </div>
        </div>

        {/* ── BACK — vinyl record fills the full square ── */}
        <div
          className={clsx(
            'absolute inset-0 rounded-2xl overflow-hidden',
            'bg-surface-950 border',
            isActive ? 'border-accent/50 shadow-lg shadow-accent/20' : 'border-surface-700/60',
          )}
          style={{ backfaceVisibility: 'hidden', transform: 'rotateY(180deg)' }}
        >
          {/* Full-card vinyl disc */}
          <div className="absolute inset-0 flex items-center justify-center bg-black/60">
            <div
              className="relative rounded-full shadow-2xl"
              style={{
                width: '88%',
                aspectRatio: '1/1',
                background: 'repeating-radial-gradient(circle at center, transparent 0px, transparent 3px, rgba(40,40,40,0.85) 3px, rgba(40,40,40,0.85) 4px), radial-gradient(circle at center, #1a1a1a 35%, #050505 100%)',
                animation: isThisPlaying ? `vinylSpin ${1.8 / playbackRate}s linear infinite` : 'none',
                boxShadow: isActive
                  ? '0 0 50px rgba(139,92,246,0.35), 0 0 20px rgba(139,92,246,0.15), 0 8px 40px rgba(0,0,0,0.7)'
                  : '0 8px 40px rgba(0,0,0,0.7)',
              }}
            >
              {/* Vinyl shine */}
              <div className="absolute inset-0 rounded-full" style={{ background: 'linear-gradient(135deg, rgba(255,255,255,0.07) 0%, transparent 55%)' }} />

              {/* Cover art label — 38% of disc diameter */}
              <div
                className="absolute rounded-full overflow-hidden border-2 border-surface-700/40"
                style={{ width: '38%', height: '38%', top: '31%', left: '31%' }}
              >
                {art
                  ? <img src={art} alt="" className="w-full h-full object-cover" />
                  : <div className="w-full h-full bg-gradient-to-br from-accent to-vapor-pink" />
                }
              </div>

              {/* Center hole */}
              <div className="absolute rounded-full bg-surface-950"
                style={{ width: '7%', height: '7%', top: '46.5%', left: '46.5%' }}
              />
            </div>
          </div>

          {/* Bottom overlay: track info + controls */}
          <div
            className="absolute bottom-0 left-0 right-0 px-3 py-3 flex items-center justify-between gap-2"
            style={{ background: 'linear-gradient(to top, rgba(0,0,0,0.9) 0%, rgba(0,0,0,0.5) 60%, transparent 100%)' }}
          >
            <div className="min-w-0 flex-1">
              <p className={clsx('font-semibold text-xs truncate leading-tight', isActive ? 'text-accent-light' : 'text-surface-100')}>
                {track.track_title}
              </p>
              <p className="text-[10px] text-white/50 truncate mt-0.5">{track.artist_name}</p>
            </div>
            <div className="flex items-center gap-1.5 flex-shrink-0">
              <button
                onClick={handleFav}
                className={clsx('p-1.5 rounded-full transition-colors', isFav ? 'text-red-400' : 'text-white/50 hover:text-white/80')}
              >
                <Heart className={clsx('w-3.5 h-3.5', isFav && 'fill-current')} />
              </button>
              <AddToPlaylistButton track={track} />
              <button
                onClick={handlePlayBtn}
                className="w-8 h-8 rounded-full bg-accent hover:bg-accent-hover flex items-center justify-center shadow-lg shadow-accent/40 transition-colors"
              >
                {isThisPlaying ? <Pause className="w-3.5 h-3.5 text-white" /> : <Play className="w-3.5 h-3.5 text-white ml-0.5" />}
              </button>
            </div>
          </div>
        </div>
      </div>
    </motion.div>
  )
})

export default TrackCard

function MusicIcon({ large }) {
  const size = large ? 'w-12 h-12' : 'w-6 h-6'
  return (
    <div className="w-full h-full flex items-center justify-center">
      <svg className={clsx(size, 'text-surface-600')} viewBox="0 0 24 24" fill="currentColor">
        <path d="M12 3v10.55c-.59-.34-1.27-.55-2-.55-2.21 0-4 1.79-4 4s1.79 4 4 4 4-1.79 4-4V7h4V3h-6z"/>
      </svg>
    </div>
  )
}

function NowPlayingBars() {
  return (
    <div className="flex items-end gap-0.5 h-6">
      {[1, 2, 3, 4].map(i => (
        <div
          key={i}
          className="w-1 bg-accent rounded-full"
          style={{
            height: `${[55, 100, 70, 85][i - 1]}%`,
            animation: `vinylBar 0.8s ease-in-out infinite alternate`,
            animationDelay: `${i * 0.12}s`,
          }}
        />
      ))}
    </div>
  )
}
