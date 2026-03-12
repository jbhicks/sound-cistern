import { useMemo } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { X, SlidersHorizontal, TrendingUp, Clock, Heart, Play, Repeat2, Zap } from 'lucide-react'
import clsx from 'clsx'

export const DEFAULT_FILTERS = {
  q: '',
  sort: 'newest',
  durationMin: 0,   // minutes
  durationMax: 0,   // 0 = no limit
  bpmMin: 0,        // 0 = no min
  bpmMax: 0,        // 0 = no max
  genre: '',
  onlyFavorites: false,
  hideReposts: false,
}

const SORT_OPTIONS = [
  { value: 'newest',    label: 'Newest',      icon: TrendingUp },
  { value: 'oldest',   label: 'Oldest',       icon: Clock },
  { value: 'plays',    label: 'Most Played',  icon: Play },
  { value: 'likes',    label: 'Most Liked',   icon: Heart },
  { value: 'duration', label: 'Duration',     icon: Clock },
  { value: 'bpm',      label: 'BPM',          icon: Zap },
]

const DURATION_PRESETS = [
  { label: 'Any',    min: 0,  max: 0  },
  { label: '< 2m',  min: 0,  max: 2  },
  { label: '2–5m',  min: 2,  max: 5  },
  { label: '5–10m', min: 5,  max: 10 },
  { label: '10–30m',min: 10, max: 30 },
  { label: '30m+',  min: 30, max: 0  },
]

const BPM_PRESETS = [
  { label: 'Any', min: 0, max: 0 },
  { label: '< 100', min: 0, max: 100 },
  { label: '100–120', min: 100, max: 120 },
  { label: '120–128', min: 120, max: 128 },
  { label: '128–140', min: 128, max: 140 },
  { label: '140–175', min: 140, max: 175 },
  { label: '175+', min: 175, max: 0 },
]

export function applyFilters(tracks, filters, favorites) {
  let out = [...tracks]

  if (filters.q) {
    const q = filters.q.toLowerCase()
    out = out.filter(t =>
      t.track_title?.toLowerCase().includes(q) ||
      t.artist_name?.toLowerCase().includes(q) ||
      t.genre?.toLowerCase().includes(q)
    )
  }

  if (filters.genre) {
    out = out.filter(t => t.genre?.toLowerCase() === filters.genre.toLowerCase())
  }

  if (filters.durationMin > 0) {
    out = out.filter(t => t.track_duration >= filters.durationMin * 60 * 1000)
  }
  if (filters.durationMax > 0) {
    out = out.filter(t => t.track_duration <= filters.durationMax * 60 * 1000)
  }

  if (filters.bpmMin > 0) {
    out = out.filter(t => (t.bpm || 0) >= filters.bpmMin)
  }
  if (filters.bpmMax > 0) {
    out = out.filter(t => (t.bpm || 0) <= filters.bpmMax)
  }

  if (filters.onlyFavorites) {
    out = out.filter(t => favorites.has(t.track_id))
  }

  if (filters.hideReposts) {
    out = out.filter(t => !t.is_repost)
  }

  switch (filters.sort) {
    case 'oldest':   out.sort((a, b) => a.track_id - b.track_id); break
    case 'plays':    out.sort((a, b) => (b.playback_count || 0) - (a.playback_count || 0)); break
    case 'likes':    out.sort((a, b) => (b.favoritings_count || 0) - (a.favoritings_count || 0)); break
    case 'duration': out.sort((a, b) => (b.track_duration || 0) - (a.track_duration || 0)); break
    case 'bpm':      out.sort((a, b) => (b.bpm || 0) - (a.bpm || 0)); break
    default: break // newest = API order
  }

  return out
}

export function countActiveFilters(f) {
  let n = 0
  if (f.q) n++
  if (f.genre) n++
  if (f.durationMin > 0 || f.durationMax > 0) n++
  if (f.bpmMin > 0 || f.bpmMax > 0) n++
  if (f.onlyFavorites) n++
  if (f.hideReposts) n++
  if (f.sort !== 'newest') n++
  return n
}

export default function FilterPanel({ open, onClose, filters, onChange, tracks }) {
  const genres = useMemo(() => {
    const seen = new Set()
    const list = []
    for (const t of tracks) {
      if (t.genre && !seen.has(t.genre)) {
        seen.add(t.genre)
        list.push(t.genre)
      }
    }
    return list.sort()
  }, [tracks])

  const set = (key, value) => onChange({ ...filters, [key]: value })
  const reset = () => onChange({ ...DEFAULT_FILTERS })

  const activeDurationPreset = DURATION_PRESETS.find(
    p => p.min === filters.durationMin && p.max === filters.durationMax
  )

  return (
    <AnimatePresence>
      {open && (
        <>
          {/* Backdrop */}
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="fixed inset-0 z-40 bg-black/40 backdrop-blur-sm"
            onClick={onClose}
          />

          {/* Panel */}
          <motion.div
            initial={{ x: '100%' }}
            animate={{ x: 0 }}
            exit={{ x: '100%' }}
            transition={{ type: 'spring', stiffness: 380, damping: 38 }}
            className="fixed right-0 top-0 bottom-0 z-50 w-full max-w-sm bg-surface-950 border-l border-surface-800/80 flex flex-col shadow-2xl"
          >
            {/* Header */}
            <div className="flex items-center justify-between px-5 py-4 border-b border-surface-800/60">
              <div className="flex items-center gap-2.5">
                <SlidersHorizontal className="w-4 h-4 text-accent" />
                <span className="font-semibold text-surface-100">Filters</span>
              </div>
              <div className="flex items-center gap-2">
                {countActiveFilters(filters) > 0 && (
                  <button
                    onClick={reset}
                    className="text-xs text-accent hover:text-accent-light transition-colors px-2 py-1 rounded-lg hover:bg-accent/10"
                  >
                    Reset all
                  </button>
                )}
                <button onClick={onClose} className="p-1.5 rounded-lg text-surface-500 hover:text-surface-200 hover:bg-surface-800 transition-colors">
                  <X className="w-4 h-4" />
                </button>
              </div>
            </div>

            {/* Scrollable body */}
            <div className="flex-1 overflow-y-auto px-5 py-5 space-y-7">

              {/* Sort */}
              <section>
                <Label>Sort by</Label>
                <div className="grid grid-cols-2 gap-2 mt-2">
                  {SORT_OPTIONS.map(opt => (
                    <button
                      key={opt.value}
                      onClick={() => set('sort', opt.value)}
                      className={clsx(
                        'flex items-center gap-2 px-3 py-2.5 rounded-xl text-sm font-medium transition-all border',
                        filters.sort === opt.value
                          ? 'bg-accent/15 border-accent/40 text-accent-light'
                          : 'bg-surface-800/50 border-surface-700/40 text-surface-400 hover:text-surface-200 hover:border-surface-600'
                      )}
                    >
                      <opt.icon className="w-3.5 h-3.5 flex-shrink-0" />
                      {opt.label}
                    </button>
                  ))}
                </div>
              </section>

              {/* Duration */}
              <section>
                <Label>Duration</Label>
                <div className="flex flex-wrap gap-2 mt-2">
                  {DURATION_PRESETS.map(p => {
                    const isActive = p.min === filters.durationMin && p.max === filters.durationMax
                    return (
                      <button
                        key={p.label}
                        onClick={() => onChange({ ...filters, durationMin: p.min, durationMax: p.max })}
                        className={clsx(
                          'px-3 py-1.5 rounded-full text-xs font-medium transition-all border',
                          isActive
                            ? 'bg-accent/15 border-accent/40 text-accent-light'
                            : 'bg-surface-800/50 border-surface-700/40 text-surface-400 hover:text-surface-200 hover:border-surface-600'
                        )}
                      >
                        {p.label}
                      </button>
                    )
                  })}
                </div>

                {/* Custom range */}
                {!activeDurationPreset && (filters.durationMin > 0 || filters.durationMax > 0) && (
                  <p className="text-xs text-accent mt-2">
                    {filters.durationMin}m – {filters.durationMax > 0 ? `${filters.durationMax}m` : '∞'}
                  </p>
                )}

                <div className="grid grid-cols-2 gap-3 mt-3">
                  <div>
                    <span className="text-xs text-surface-500 mb-1 block">Min (min)</span>
                    <input
                      type="number"
                      min="0"
                      max="120"
                      value={filters.durationMin || ''}
                      placeholder="0"
                      onChange={e => set('durationMin', Math.max(0, parseInt(e.target.value) || 0))}
                      className="w-full px-3 py-2 bg-surface-800 border border-surface-700/60 rounded-xl text-sm text-surface-100 placeholder-surface-600 focus:outline-none focus:border-accent/60 focus:ring-1 focus:ring-accent/20 transition-all"
                    />
                  </div>
                  <div>
                    <span className="text-xs text-surface-500 mb-1 block">Max (min)</span>
                    <input
                      type="number"
                      min="0"
                      max="180"
                      value={filters.durationMax || ''}
                      placeholder="∞"
                      onChange={e => set('durationMax', Math.max(0, parseInt(e.target.value) || 0))}
                      className="w-full px-3 py-2 bg-surface-800 border border-surface-700/60 rounded-xl text-sm text-surface-100 placeholder-surface-600 focus:outline-none focus:border-accent/60 focus:ring-1 focus:ring-accent/20 transition-all"
                    />
                  </div>
                </div>
              </section>

              {/* BPM */}
              <section>
                <Label>BPM (Tempo)</Label>
                <div className="flex flex-wrap gap-2 mt-2">
                  {BPM_PRESETS.map(p => {
                    const isActive = p.min === filters.bpmMin && p.max === filters.bpmMax
                    return (
                      <button
                        key={p.label}
                        onClick={() => onChange({ ...filters, bpmMin: p.min, bpmMax: p.max })}
                        className={clsx(
                          'px-3 py-1.5 rounded-full text-xs font-medium transition-all border',
                          isActive
                            ? 'bg-accent/15 border-accent/40 text-accent-light'
                            : 'bg-surface-800/50 border-surface-700/40 text-surface-400 hover:text-surface-200 hover:border-surface-600'
                        )}
                      >
                        {p.label}
                      </button>
                    )
                  })}
                </div>

                {/* Custom range */}
                {!BPM_PRESETS.find(p => p.min === filters.bpmMin && p.max === filters.bpmMax) && (filters.bpmMin > 0 || filters.bpmMax > 0) && (
                  <p className="text-xs text-accent mt-2">
                    {filters.bpmMin} – {filters.bpmMax > 0 ? filters.bpmMax : '∞'} BPM
                  </p>
                )}

                <div className="grid grid-cols-2 gap-3 mt-3">
                  <div>
                    <span className="text-xs text-surface-500 mb-1 block">Min BPM</span>
                    <input
                      type="number"
                      min="0"
                      max="300"
                      value={filters.bpmMin || ''}
                      placeholder="0"
                      onChange={e => set('bpmMin', Math.max(0, parseInt(e.target.value) || 0))}
                      className="w-full px-3 py-2 bg-surface-800 border border-surface-700/60 rounded-xl text-sm text-surface-100 placeholder-surface-600 focus:outline-none focus:border-accent/60 focus:ring-1 focus:ring-accent/20 transition-all"
                    />
                  </div>
                  <div>
                    <span className="text-xs text-surface-500 mb-1 block">Max BPM</span>
                    <input
                      type="number"
                      min="0"
                      max="300"
                      value={filters.bpmMax || ''}
                      placeholder="∞"
                      onChange={e => set('bpmMax', Math.max(0, parseInt(e.target.value) || 0))}
                      className="w-full px-3 py-2 bg-surface-800 border border-surface-700/60 rounded-xl text-sm text-surface-100 placeholder-surface-600 focus:outline-none focus:border-accent/60 focus:ring-1 focus:ring-accent/20 transition-all"
                    />
                  </div>
                </div>
              </section>

              {/* Genre */}
              {genres.length > 0 && (
                <section>
                  <Label>Genre</Label>
                  <div className="flex flex-wrap gap-2 mt-2 max-h-40 overflow-y-auto pr-1">
                    {genres.map(g => (
                      <button
                        key={g}
                        onClick={() => set('genre', filters.genre === g ? '' : g)}
                        className={clsx(
                          'px-3 py-1.5 rounded-full text-xs font-medium transition-all border',
                          filters.genre === g
                            ? 'bg-accent/15 border-accent/40 text-accent-light'
                            : 'bg-surface-800/50 border-surface-700/40 text-surface-400 hover:text-surface-200 hover:border-surface-600'
                        )}
                      >
                        {g}
                      </button>
                    ))}
                  </div>
                </section>
              )}

              {/* Toggles */}
              <section>
                <Label>Show</Label>
                <div className="space-y-2 mt-2">
                  <Toggle
                    label="Favorites only"
                    icon={Heart}
                    active={filters.onlyFavorites}
                    onClick={() => set('onlyFavorites', !filters.onlyFavorites)}
                  />
                  <Toggle
                    label="Hide reposts"
                    icon={Repeat2}
                    active={filters.hideReposts}
                    onClick={() => set('hideReposts', !filters.hideReposts)}
                  />
                </div>
              </section>

            </div>

            {/* Footer — active filter count */}
            <div className="px-5 py-4 border-t border-surface-800/60">
              <button
                onClick={onClose}
                className="w-full py-2.5 rounded-xl bg-accent hover:bg-accent-hover text-white text-sm font-medium transition-colors shadow-lg shadow-accent/20"
              >
                {countActiveFilters(filters) > 0
                  ? `Apply ${countActiveFilters(filters)} filter${countActiveFilters(filters) > 1 ? 's' : ''}`
                  : 'Done'}
              </button>
            </div>
          </motion.div>
        </>
      )}
    </AnimatePresence>
  )
}

function Label({ children }) {
  return <p className="text-xs font-semibold text-surface-500 uppercase tracking-wider">{children}</p>
}

function Toggle({ label, icon: Icon, active, onClick }) {
  return (
    <button
      onClick={onClick}
      className={clsx(
        'w-full flex items-center justify-between px-4 py-3 rounded-xl border transition-all',
        active
          ? 'bg-accent/10 border-accent/30 text-accent-light'
          : 'bg-surface-800/50 border-surface-700/40 text-surface-400 hover:text-surface-200 hover:border-surface-600'
      )}
    >
      <span className="flex items-center gap-2.5 text-sm font-medium">
        <Icon className="w-4 h-4" />
        {label}
      </span>
      {/* Toggle pill */}
      <div className={clsx(
        'w-9 h-5 rounded-full transition-colors relative flex-shrink-0',
        active ? 'bg-accent' : 'bg-surface-700'
      )}>
        <div className={clsx(
          'absolute top-0.5 w-4 h-4 rounded-full bg-white shadow transition-transform',
          active ? 'translate-x-4' : 'translate-x-0.5'
        )} />
      </div>
    </button>
  )
}
