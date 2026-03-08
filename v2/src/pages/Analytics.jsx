import { useMemo } from 'react'
import { Link } from 'react-router-dom'
import { motion } from 'framer-motion'
import { BarChart2, Music, Users, Clock, Play, Radio } from 'lucide-react'
import { useStore } from '../store'

// ─── helpers ────────────────────────────────────────────────────────────────

function formatDuration(ms) {
  const totalSecs = Math.floor((ms || 0) / 1000)
  const h = Math.floor(totalSecs / 3600)
  const m = Math.floor((totalSecs % 3600) / 60)
  if (h === 0) return `${m}m`
  return `${h}h ${m}m`
}

function relativeTime(isoString) {
  const diff = Date.now() - new Date(isoString).getTime()
  const secs = Math.floor(diff / 1000)
  if (secs < 60) return 'just now'
  const mins = Math.floor(secs / 60)
  if (mins < 60) return `${mins}m ago`
  const hrs = Math.floor(mins / 60)
  if (hrs < 24) return `${hrs}h ago`
  const days = Math.floor(hrs / 24)
  if (days === 1) return 'yesterday'
  if (days < 7) return `${days}d ago`
  return new Date(isoString).toLocaleDateString()
}

function artworkUrl(url) {
  return url?.replace(/-t\d+x\d+/g, '-t50x50') || ''
}

// ─── sub-components ─────────────────────────────────────────────────────────

function StatCard({ icon: Icon, label, value, delay = 0 }) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 16 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay, duration: 0.35 }}
      className="bg-surface-900/70 border border-surface-800/60 rounded-2xl p-4 flex items-start gap-3"
    >
      <div className="w-9 h-9 rounded-xl bg-accent/15 flex items-center justify-center flex-shrink-0 mt-0.5">
        <Icon className="w-4 h-4 text-accent-light" />
      </div>
      <div className="min-w-0">
        <p className="text-2xl font-bold text-surface-100 leading-tight truncate">{value}</p>
        <p className="text-xs text-surface-500 mt-0.5">{label}</p>
      </div>
    </motion.div>
  )
}

function SectionTitle({ children, delay = 0 }) {
  return (
    <motion.h2
      initial={{ opacity: 0, x: -8 }}
      animate={{ opacity: 1, x: 0 }}
      transition={{ delay, duration: 0.3 }}
      className="text-base font-semibold text-surface-200 mb-3"
    >
      {children}
    </motion.h2>
  )
}

function RankBadge({ rank }) {
  const color =
    rank === 1 ? 'text-yellow-400' :
    rank === 2 ? 'text-surface-300' :
    rank === 3 ? 'text-amber-600' :
    'text-surface-600'
  return (
    <span className={`w-6 text-xs font-bold text-right flex-shrink-0 ${color}`}>
      {rank}
    </span>
  )
}

function EmptyState() {
  return (
    <motion.div
      initial={{ opacity: 0, scale: 0.97 }}
      animate={{ opacity: 1, scale: 1 }}
      className="flex flex-col items-center justify-center py-24 text-center"
    >
      <div className="w-16 h-16 rounded-2xl bg-surface-800 flex items-center justify-center mb-4">
        <BarChart2 className="w-8 h-8 text-surface-600" />
      </div>
      <h3 className="text-lg font-semibold text-surface-300 mb-2">No listening history yet</h3>
      <p className="text-surface-500 text-sm mb-6 max-w-xs">
        Start listening to build your history — stats will appear here automatically.
      </p>
      <Link
        to="/stream"
        className="inline-flex items-center gap-2 px-5 py-2.5 bg-accent hover:bg-accent-hover text-white rounded-xl font-medium text-sm transition-colors shadow-lg shadow-accent/20"
      >
        <Radio className="w-4 h-4" />
        Go to Stream
      </Link>
    </motion.div>
  )
}

// ─── main page ───────────────────────────────────────────────────────────────

export default function Analytics() {
  const { playHistory, playTrack } = useStore()

  const stats = useMemo(() => {
    if (!playHistory.length) return null

    // Overview
    const totalPlays = playHistory.length
    const uniqueTracks = new Set(playHistory.map(e => e.track_id)).size
    const uniqueArtists = new Set(playHistory.map(e => e.artist_name)).size
    const totalListeningMs = playHistory.reduce((sum, e) => sum + (e.track_duration || 0), 0)

    // Top tracks
    const trackMap = {}
    for (const e of playHistory) {
      if (!trackMap[e.track_id]) {
        trackMap[e.track_id] = { ...e, count: 0 }
      }
      trackMap[e.track_id].count++
    }
    const topTracks = Object.values(trackMap)
      .sort((a, b) => b.count - a.count)
      .slice(0, 10)

    // Top artists
    const artistMap = {}
    for (const e of playHistory) {
      const key = e.artist_name || 'Unknown'
      if (!artistMap[key]) artistMap[key] = { name: key, count: 0, totalMs: 0 }
      artistMap[key].count++
      artistMap[key].totalMs += e.track_duration || 0
    }
    const topArtists = Object.values(artistMap)
      .sort((a, b) => b.count - a.count)
      .slice(0, 10)

    // Genre distribution
    const genreMap = {}
    for (const e of playHistory) {
      const g = (e.genre || '').trim()
      if (!g) continue
      genreMap[g] = (genreMap[g] || 0) + 1
    }
    const genres = Object.entries(genreMap)
      .map(([genre, count]) => ({ genre, count }))
      .sort((a, b) => b.count - a.count)
    const maxGenreCount = genres[0]?.count || 1

    // Recently played (last 20 unique-ified by recency)
    const recentlyPlayed = playHistory.slice(0, 20)

    return {
      totalPlays,
      uniqueTracks,
      uniqueArtists,
      totalListeningMs,
      topTracks,
      topArtists,
      genres,
      maxGenreCount,
      recentlyPlayed,
    }
  }, [playHistory])

  return (
    <div className="min-h-screen px-4 py-6 md:px-6 lg:px-8">
      <div className="max-w-5xl mx-auto">

        {/* Page header */}
        <motion.div
          initial={{ opacity: 0, y: -12 }}
          animate={{ opacity: 1, y: 0 }}
          className="mb-6"
        >
          <h1 className="text-2xl font-bold text-gradient">Listening Analytics</h1>
          <p className="text-surface-500 text-sm mt-0.5">
            {playHistory.length > 0
              ? `${playHistory.length} plays tracked locally`
              : 'Your personal listening stats'}
          </p>
        </motion.div>

        {!stats ? (
          <EmptyState />
        ) : (
          <div className="space-y-8">

            {/* ── Overview cards ── */}
            <section>
              <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
                <StatCard icon={Play}    label="Total Plays"       value={stats.totalPlays.toLocaleString()}    delay={0} />
                <StatCard icon={Music}   label="Unique Tracks"     value={stats.uniqueTracks.toLocaleString()}  delay={0.05} />
                <StatCard icon={Clock}   label="Listening Time"    value={formatDuration(stats.totalListeningMs)} delay={0.1} />
                <StatCard icon={Users}   label="Unique Artists"    value={stats.uniqueArtists.toLocaleString()} delay={0.15} />
              </div>
            </section>

            {/* ── Top two columns ── */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">

              {/* Most Played Tracks */}
              <motion.section
                initial={{ opacity: 0, y: 16 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.2 }}
              >
                <SectionTitle delay={0.2}>Most Played Tracks</SectionTitle>
                <div className="bg-surface-900/50 border border-surface-800/50 rounded-2xl overflow-hidden divide-y divide-surface-800/40">
                  {stats.topTracks.map((track, i) => (
                    <button
                      key={track.track_id}
                      onClick={() => playTrack(track)}
                      className="w-full flex items-center gap-3 px-4 py-2.5 hover:bg-surface-800/40 transition-colors text-left"
                    >
                      <RankBadge rank={i + 1} />
                      {artworkUrl(track.artwork_url) ? (
                        <img
                          src={artworkUrl(track.artwork_url)}
                          alt=""
                          className="w-8 h-8 rounded-md object-cover flex-shrink-0"
                        />
                      ) : (
                        <div className="w-8 h-8 rounded-md bg-surface-800 flex items-center justify-center flex-shrink-0">
                          <Music className="w-3.5 h-3.5 text-surface-600" />
                        </div>
                      )}
                      <div className="flex-1 min-w-0">
                        <p className="text-sm font-medium text-surface-100 truncate leading-tight">
                          {track.track_title}
                        </p>
                        <p className="text-xs text-surface-500 truncate">{track.artist_name}</p>
                      </div>
                      <span className="flex-shrink-0 text-xs font-semibold px-2 py-0.5 rounded-full bg-accent/15 text-accent-light">
                        {track.count}×
                      </span>
                    </button>
                  ))}
                </div>
              </motion.section>

              {/* Most Played Artists */}
              <motion.section
                initial={{ opacity: 0, y: 16 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.25 }}
              >
                <SectionTitle delay={0.25}>Most Played Artists</SectionTitle>
                <div className="bg-surface-900/50 border border-surface-800/50 rounded-2xl overflow-hidden divide-y divide-surface-800/40">
                  {stats.topArtists.map((artist, i) => (
                    <div
                      key={artist.name}
                      className="flex items-center gap-3 px-4 py-2.5"
                    >
                      <RankBadge rank={i + 1} />
                      <div className="w-8 h-8 rounded-full bg-gradient-to-br from-accent/40 to-vapor-pink/30 flex items-center justify-center flex-shrink-0 text-xs font-bold text-accent-light">
                        {artist.name.charAt(0).toUpperCase()}
                      </div>
                      <div className="flex-1 min-w-0">
                        <p className="text-sm font-medium text-surface-100 truncate">{artist.name}</p>
                        <p className="text-xs text-surface-500">{formatDuration(artist.totalMs)}</p>
                      </div>
                      <span className="flex-shrink-0 text-xs font-semibold px-2 py-0.5 rounded-full bg-accent/15 text-accent-light">
                        {artist.count}×
                      </span>
                    </div>
                  ))}
                </div>
              </motion.section>
            </div>

            {/* ── Genre Distribution ── */}
            {stats.genres.length > 0 && (
              <motion.section
                initial={{ opacity: 0, y: 16 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.3 }}
              >
                <SectionTitle delay={0.3}>Genre Distribution</SectionTitle>
                <div className="bg-surface-900/50 border border-surface-800/50 rounded-2xl px-5 py-4 space-y-3">
                  {stats.genres.slice(0, 15).map(({ genre, count }) => {
                    const pct = Math.round((count / stats.maxGenreCount) * 100)
                    return (
                      <div key={genre} className="flex items-center gap-3">
                        <span className="w-28 text-xs text-surface-400 truncate flex-shrink-0">{genre}</span>
                        <div className="flex-1 bg-surface-800 rounded-full h-2">
                          <motion.div
                            initial={{ width: 0 }}
                            animate={{ width: `${pct}%` }}
                            transition={{ delay: 0.4, duration: 0.5, ease: 'easeOut' }}
                            className="bg-accent rounded-full h-2"
                          />
                        </div>
                        <span className="text-xs text-surface-500 w-8 text-right flex-shrink-0">{count}</span>
                      </div>
                    )
                  })}
                </div>
              </motion.section>
            )}

            {/* ── Recently Played ── */}
            <motion.section
              initial={{ opacity: 0, y: 16 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.35 }}
            >
              <SectionTitle delay={0.35}>Recently Played</SectionTitle>
              <div className="bg-surface-900/50 border border-surface-800/50 rounded-2xl overflow-hidden divide-y divide-surface-800/40">
                {stats.recentlyPlayed.map((entry, i) => (
                  <button
                    key={`${entry.track_id}-${i}`}
                    onClick={() => playTrack(entry)}
                    className="w-full flex items-center gap-3 px-4 py-2.5 hover:bg-surface-800/40 transition-colors text-left"
                  >
                    {artworkUrl(entry.artwork_url) ? (
                      <img
                        src={artworkUrl(entry.artwork_url)}
                        alt=""
                        className="w-8 h-8 rounded-md object-cover flex-shrink-0"
                      />
                    ) : (
                      <div className="w-8 h-8 rounded-md bg-surface-800 flex items-center justify-center flex-shrink-0">
                        <Music className="w-3.5 h-3.5 text-surface-600" />
                      </div>
                    )}
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium text-surface-100 truncate leading-tight">
                        {entry.track_title}
                      </p>
                      <p className="text-xs text-surface-500 truncate">{entry.artist_name}</p>
                    </div>
                    <span className="text-xs text-surface-600 flex-shrink-0 whitespace-nowrap">
                      {relativeTime(entry.played_at)}
                    </span>
                  </button>
                ))}
              </div>
            </motion.section>

          </div>
        )}
      </div>
    </div>
  )
}
