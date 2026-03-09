import { useState, useEffect, useRef, useCallback, useMemo } from 'react'
import { useVirtualizer } from '@tanstack/react-virtual'
import { motion, AnimatePresence } from 'framer-motion'
import { Search, RefreshCw, Grid, List, X, Music, SlidersHorizontal, Clock } from 'lucide-react'
import clsx from 'clsx'
import TrackCard from '../components/TrackCard'
import FilterPanel, { DEFAULT_FILTERS, applyFilters, countActiveFilters } from '../components/FilterPanel'
import { useStore } from '../store'

const LIMIT = 50
const MIN_VISIBLE = 100 // keep fetching until this many filtered results are visible

function SkeletonCard({ viewMode }) {
  if (viewMode === 'list') {
    return (
      <div className="flex items-center gap-3 p-3 rounded-xl">
        <div className="w-12 h-12 rounded-lg bg-surface-800 animate-pulse flex-shrink-0" />
        <div className="flex-1 space-y-2">
          <div className="h-3.5 w-2/3 bg-surface-800 rounded animate-pulse" />
          <div className="h-3 w-1/2 bg-surface-800 rounded animate-pulse" />
        </div>
      </div>
    )
  }
  return (
    <div className="rounded-2xl overflow-hidden bg-surface-800/40 border border-surface-700/30 relative" style={{ aspectRatio: '1/1' }}>
      <div className="absolute inset-0 animate-pulse bg-surface-700/50" />
      <div className="absolute bottom-0 left-0 right-0 p-3 space-y-1.5" style={{ background: 'linear-gradient(to top, rgba(0,0,0,0.6) 0%, transparent 100%)' }}>
        <div className="h-3.5 w-3/4 bg-white/20 rounded animate-pulse" />
        <div className="h-3 w-1/2 bg-white/15 rounded animate-pulse" />
      </div>
    </div>
  )
}

// Human-readable label for an active filter chip
function filterChips(filters) {
  const chips = []
  if (filters.sort !== 'newest') {
    const labels = { oldest: 'Oldest', plays: 'Most Played', likes: 'Most Liked', duration: 'By Duration', bpm: 'By BPM' }
    chips.push({ key: 'sort', label: labels[filters.sort] || filters.sort })
  }
  if (filters.durationMin > 0 || filters.durationMax > 0) {
    const lo = filters.durationMin > 0 ? `${filters.durationMin}m` : '0'
    const hi = filters.durationMax > 0 ? `${filters.durationMax}m` : '∞'
    chips.push({ key: 'duration', label: `${lo} – ${hi}` })
  }
  if (filters.genre) chips.push({ key: 'genre', label: filters.genre })
  if (filters.onlyFavorites) chips.push({ key: 'onlyFavorites', label: 'Favorites' })
  if (filters.hideReposts) chips.push({ key: 'hideReposts', label: 'No Reposts' })
  return chips
}

const TRACKS_CACHE_KEY = 'sc-tracks-cache'

function readTracksCache() {
  try {
    const raw = localStorage.getItem(TRACKS_CACHE_KEY)
    return raw ? JSON.parse(raw) : []
  } catch { return [] }
}

function writeTracksCache(tracks) {
  try {
    // Cap at 500 tracks to avoid blowing the localStorage quota
    localStorage.setItem(TRACKS_CACHE_KEY, JSON.stringify(tracks.slice(0, 500)))
  } catch {}
}

// Calculate columns based on container width (must match Tailwind grid classes)
function getColumnCount(width) {
  if (width >= 1536) return 7 // 2xl:grid-cols-7
  if (width >= 1280) return 6 // xl:grid-cols-6
  if (width >= 1024) return 5 // lg:grid-cols-5
  if (width >= 768) return 4  // md:grid-cols-4
  if (width >= 640) return 3  // sm:grid-cols-3
  return 2                    // grid-cols-2
}

// Virtualized grid row component
function VirtualGridRow({ tracks, columns, gap, viewMode, topArtistNames, newTrackIdsRef }) {
  return (
    <div 
      className="grid"
      style={{
        gridTemplateColumns: `repeat(${columns}, 1fr)`,
        gap: `${gap}px`,
      }}
    >
      {tracks.map((track, i) => (
        <TrackCard
          key={track.track_id}
          track={track}
          index={i}
          viewMode={viewMode}
          isNew={newTrackIdsRef.current.has(track.track_id)}
          isTopArtist={topArtistNames.has(track.artist_name)}
        />
      ))}
    </div>
  )
}

export default function Stream() {
  const cachedTracks = useState(() => readTracksCache())[0]
  const [tracks, setTracks] = useState(() => cachedTracks)
  const [loading, setLoading] = useState(cachedTracks.length === 0)
  const [loadingMore, setLoadingMore] = useState(false)
  const [syncing, setSyncing] = useState(false)
  const [viewMode, setViewMode] = useState(
    () => localStorage.getItem('sc-view-mode') || 'grid'
  )
  const [query, setQuery] = useState(
    () => localStorage.getItem('sc-query') || ''
  )
  const [hasMore, setHasMore] = useState(true)
  const [filterOpen, setFilterOpen] = useState(false)
  const [filters, setFilters] = useState(() => {
    try {
      const saved = localStorage.getItem('sc-filters')
      return saved ? { ...DEFAULT_FILTERS, ...JSON.parse(saved) } : DEFAULT_FILTERS
    } catch { return DEFAULT_FILTERS }
  })
  const { favorites, playHistory, playTrack } = useStore()
  const recentlyPlayed = [...new Map(playHistory.map(e => [e.track_id, e])).values()].slice(0, 5)

  // Refs for virtualization
  const scrollContainerRef = useRef(null)
  const [containerWidth, setContainerWidth] = useState(0)

  // Track container width for column calculation
  useEffect(() => {
    const container = scrollContainerRef.current
    if (!container) return
    
    const updateWidth = () => {
      setContainerWidth(container.clientWidth)
    }
    
    updateWidth()
    const ro = new ResizeObserver(updateWidth)
    ro.observe(container)
    
    return () => ro.disconnect()
  }, [viewMode])

  // Top 10 most-played artists by play count — same logic as Analytics page
  const topArtistNames = useMemo(() => {
    const counts = {}
    for (const e of playHistory) {
      const key = e.artist_name || ''
      if (!key) continue
      counts[key] = (counts[key] || 0) + 1
    }
    return new Set(
      Object.entries(counts)
        .sort((a, b) => b[1] - a[1])
        .slice(0, 10)
        .map(([name]) => name)
    )
  }, [playHistory])

  const offsetRef = useRef(0)
  const loadingMoreRef = useRef(false)
  const hasMoreRef = useRef(true)
  const sentinelRef = useRef(null)
  const queryRef = useRef(query)
  const filtersRef = useRef(filters)
  const tracksRef = useRef(cachedTracks)
  const newTrackIdsRef = useRef(new Set())

  useEffect(() => {
    queryRef.current = query
    localStorage.setItem('sc-query', query)
  }, [query])

  useEffect(() => {
    filtersRef.current = filters
    localStorage.setItem('sc-filters', JSON.stringify(filters))
  }, [filters])

  useEffect(() => {
    localStorage.setItem('sc-view-mode', viewMode)
  }, [viewMode])

  const doFetch = useCallback(async (replace = true) => {
    const currentOffset = replace ? 0 : offsetRef.current
    if (!replace && (loadingMoreRef.current || !hasMoreRef.current)) return

    if (replace) {
      // Only show the loading skeleton if there's nothing cached to show yet
      if (tracksRef.current.length === 0) setLoading(true)
      offsetRef.current = 0
    } else {
      loadingMoreRef.current = true
      setLoadingMore(true)
    }

    try {
      const params = new URLSearchParams({
        format: 'json',
        limit: LIMIT,
        offset: currentOffset,
      })
      if (queryRef.current) params.set('q', queryRef.current)

      const res = await fetch(`/api/stream?${params}`, {
        credentials: 'include',
        headers: { Accept: 'application/json' },
      })

      if (res.status === 401) { window.location.href = '/auth/soundcloud'; return }
      if (!res.ok) return

      const data = await res.json()
      const incoming = Array.isArray(data) ? data : (data.tracks || [])
      // Use has_more from backend when available; fall back to length heuristic
      const more = data.has_more !== undefined ? data.has_more : incoming.length === LIMIT

      hasMoreRef.current = more
      setHasMore(more)

      let nextTracks
      if (replace) {
        // Diff against what was previously cached to identify genuinely new tracks
        const cachedIds = new Set(tracksRef.current.map(t => t.track_id))
        const brandNew = new Set(incoming.filter(t => !cachedIds.has(t.track_id)).map(t => t.track_id))
        // Merge: new tracks at front, then existing tracks not in incoming (preserves cached tail)
        const incomingIds = new Set(incoming.map(t => t.track_id))
        const retained = tracksRef.current.filter(t => !incomingIds.has(t.track_id))
        nextTracks = [...incoming, ...retained]
        offsetRef.current = incoming.length
        newTrackIdsRef.current = brandNew
        writeTracksCache(nextTracks)
      } else {
        const seen = new Set(tracksRef.current.map(t => t.track_id))
        const unique = incoming.filter(t => !seen.has(t.track_id))
        // Mark all appended tracks as new
        unique.forEach(t => newTrackIdsRef.current.add(t.track_id))
        nextTracks = [...tracksRef.current, ...unique]
        offsetRef.current = nextTracks.length
        writeTracksCache(nextTracks)
      }
      tracksRef.current = nextTracks
      setTracks(nextTracks)

      // Always keep fetching until MIN_VISIBLE filtered results or API exhausted.
      // Clear the lock first so the recursive call isn't blocked.
      const filtered = applyFilters(nextTracks, { ...filtersRef.current, q: queryRef.current }, new Set())
      const needsMore = more && filtered.length < MIN_VISIBLE
      if (needsMore) {
        loadingMoreRef.current = false // release lock before re-entering
        setLoadingMore(false)
        doFetch(false)
        return // skip the finally clear
      }
    } catch (err) {
      console.error('fetch tracks:', err)
    } finally {
      setLoading(false)
      loadingMoreRef.current = false
      setLoadingMore(false)
    }
  }, [])

  // Refetch when search query changes
  useEffect(() => {
    const t = setTimeout(() => doFetch(true), query ? 300 : 0)
    return () => clearTimeout(t)
  }, [query, doFetch])

  // Infinite scroll — also fires immediately if page isn't tall enough to scroll
  useEffect(() => {
    const sentinel = sentinelRef.current
    if (!sentinel) return
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting && hasMoreRef.current && !loadingMoreRef.current) {
          doFetch(false)
        }
      },
      // Large rootMargin means it fires as soon as sentinel is within 600px
      // of the viewport — catches the "not enough to scroll" case too
      { rootMargin: '600px' }
    )
    observer.observe(sentinel)
    return () => observer.disconnect()
  }, [doFetch])

  useEffect(() => {
    hasMoreRef.current = hasMore
    // If sentinel is still visible after a batch and there's more, trigger next fetch
    if (hasMore && !loadingMore && sentinelRef.current) {
      const rect = sentinelRef.current.getBoundingClientRect()
      if (rect.top < window.innerHeight + 600) {
        doFetch(false)
      }
    }
  }, [hasMore, loadingMore, doFetch])

  const handleSync = async () => {
    setSyncing(true)
    try {
      await fetch('/api/sync', { method: 'POST', credentials: 'include' })
      await doFetch(true)
    } finally {
      setSyncing(false)
    }
  }

  const removeFilter = (key) => {
    if (key === 'duration') setFilters(f => ({ ...f, durationMin: 0, durationMax: 0 }))
    else if (key === 'sort') setFilters(f => ({ ...f, sort: 'newest' }))
    else setFilters(f => ({ ...f, [key]: DEFAULT_FILTERS[key] }))
  }

  const displayed = applyFilters(tracks, { ...filters, q: query }, favorites)
  const chips = filterChips(filters)
  const activeCount = countActiveFilters(filters)

  // Virtualization setup
  const columns = viewMode === 'grid' ? getColumnCount(containerWidth) : 1
  const gap = 12 // gap-3 = 12px
  
  // Calculate row data for grid view
  const rowData = useMemo(() => {
    if (viewMode === 'list') {
      // In list view, each track is its own "row"
      return displayed.map(track => [track])
    }
    // In grid view, group tracks into rows
    const rows = []
    for (let i = 0; i < displayed.length; i += columns) {
      rows.push(displayed.slice(i, i + columns))
    }
    return rows
  }, [displayed, columns, viewMode])

  // Estimate row height for virtualizer
  const estimateRowHeight = useCallback(() => {
    if (viewMode === 'list') return 64 // Approximate height of list item
    // Grid item height = width based on aspect ratio 1:1
    if (containerWidth === 0) return 200
    const itemWidth = (containerWidth - (columns - 1) * gap) / columns
    return itemWidth // 1:1 aspect ratio
  }, [viewMode, containerWidth, columns, gap])

  const virtualizer = useVirtualizer({
    count: rowData.length,
    getScrollElement: () => scrollContainerRef.current,
    estimateSize: estimateRowHeight,
    overscan: 5, // Render 5 extra rows above/below viewport
    gap,
  })

  const virtualItems = virtualizer.getVirtualItems()

  return (
    <div className="min-h-screen px-4 py-6 md:px-6 lg:px-8">
      <div className="max-w-screen-2xl mx-auto">
        {/* Header */}
        <motion.div initial={{ opacity: 0, y: -12 }} animate={{ opacity: 1, y: 0 }} className="mb-6">
          <h1 className="text-2xl font-bold text-gradient">Stream</h1>
          <p className="text-surface-500 text-sm mt-0.5">Your SoundCloud feed</p>
        </motion.div>

        {/* 2xl layout: side-by-side. Below 2xl: stacked. */}
        <div className="2xl:flex 2xl:gap-6 2xl:items-start">

        {/* Main content column */}
        <div className="2xl:flex-1 2xl:min-w-0" ref={scrollContainerRef} style={{ height: 'calc(100vh - 200px)', overflow: 'auto' }}>

        {/* Recently Played — inline strip (hidden at 2xl, shown in sidebar instead) */}
        {recentlyPlayed.length > 0 && (
          <motion.div
            initial={{ opacity: 0, y: -8 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.04 }}
            className="mb-4 2xl:hidden"
          >
            <div className="flex items-center gap-1.5 mb-2">
              <Clock className="w-3.5 h-3.5 text-surface-500" />
              <span className="text-xs font-medium text-surface-500 uppercase tracking-wider">Recently Played</span>
            </div>
            <div className="flex gap-2 overflow-x-auto pb-1 scrollbar-none">
              {recentlyPlayed.map((entry) => (
                <button
                  key={entry.track_id}
                  onClick={() => playTrack(entry)}
                  title={`${entry.track_title} — ${entry.artist_name}`}
                  className="flex-shrink-0 flex items-center gap-2 pl-1.5 pr-3 py-1.5 rounded-xl bg-surface-800/60 border border-surface-700/40 hover:bg-surface-700/60 hover:border-surface-600/50 transition-all group max-w-[160px]"
                >
                  {entry.artwork_url ? (
                    <img
                      src={entry.artwork_url.replace(/-t\d+x\d+/g, '-t50x50')}
                      alt=""
                      className="w-7 h-7 rounded-md object-cover flex-shrink-0"
                    />
                  ) : (
                    <div className="w-7 h-7 rounded-md bg-surface-700 flex items-center justify-center flex-shrink-0">
                      <Music className="w-3 h-3 text-surface-500" />
                    </div>
                  )}
                  <span className="text-xs text-surface-300 truncate group-hover:text-surface-100 transition-colors">
                    {entry.track_title}
                  </span>
                </button>
              ))}
            </div>
          </motion.div>
        )}

        {/* Toolbar */}
        <motion.div
          initial={{ opacity: 0, y: -8 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.05 }}
          className="flex items-center gap-3 mb-3"
        >
          {/* Search */}
          <div className="relative flex-1">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-surface-500" />
            <input
              type="text"
              placeholder="Search tracks, artists, genres..."
              value={query}
              onChange={e => setQuery(e.target.value)}
              className="w-full pl-9 pr-8 py-2.5 bg-surface-800/70 border border-surface-700/50 rounded-xl text-sm text-surface-100 placeholder-surface-500 focus:outline-none focus:border-accent/60 focus:ring-1 focus:ring-accent/20 transition-all"
            />
            {query && (
              <button onClick={() => setQuery('')} className="absolute right-3 top-1/2 -translate-y-1/2 text-surface-500 hover:text-surface-300">
                <X className="w-3.5 h-3.5" />
              </button>
            )}
          </div>

          {/* Filter button */}
          <button
            onClick={() => setFilterOpen(true)}
            className={clsx(
              'relative flex items-center gap-2 px-3 py-2.5 rounded-xl text-sm font-medium transition-all border',
              activeCount > 0
                ? 'bg-accent/15 border-accent/40 text-accent-light'
                : 'bg-surface-800/70 border-surface-700/50 text-surface-400 hover:text-surface-200 hover:border-surface-600'
            )}
          >
            <SlidersHorizontal className="w-4 h-4" />
            <span className="hidden sm:inline">Filter</span>
            {activeCount > 0 && (
              <span className="absolute -top-1.5 -right-1.5 w-4 h-4 rounded-full bg-accent text-white text-[10px] flex items-center justify-center font-bold">
                {activeCount}
              </span>
            )}
          </button>

          {/* Sync */}
          <motion.button
            whileHover={{ scale: 1.02 }}
            whileTap={{ scale: 0.97 }}
            onClick={handleSync}
            disabled={syncing}
            className={clsx(
              'flex items-center gap-2 px-3 py-2.5 rounded-xl text-sm font-medium transition-all',
              'bg-accent hover:bg-accent-hover text-white shadow-lg shadow-accent/20',
              syncing && 'opacity-60 cursor-not-allowed'
            )}
          >
            <RefreshCw className={clsx('w-4 h-4', syncing && 'animate-spin')} />
            <span className="hidden sm:inline">{syncing ? 'Syncing…' : 'Sync'}</span>
          </motion.button>

          {/* View toggle */}
          <div className="flex items-center rounded-xl border border-surface-700/50 overflow-hidden bg-surface-800/60">
            <button
              onClick={() => setViewMode('grid')}
              className={clsx('p-2.5 transition-colors', viewMode === 'grid' ? 'bg-accent/20 text-accent-light' : 'text-surface-500 hover:text-surface-300')}
            >
              <Grid className="w-4 h-4" />
            </button>
            <button
              onClick={() => setViewMode('list')}
              className={clsx('p-2.5 transition-colors', viewMode === 'list' ? 'bg-accent/20 text-accent-light' : 'text-surface-500 hover:text-surface-300')}
            >
              <List className="w-4 h-4" />
            </button>
          </div>
        </motion.div>

        {/* Active filter chips */}
        <AnimatePresence>
          {chips.length > 0 && (
            <motion.div
              initial={{ opacity: 0, height: 0 }}
              animate={{ opacity: 1, height: 'auto' }}
              exit={{ opacity: 0, height: 0 }}
              className="flex flex-wrap gap-2 mb-4 overflow-hidden"
            >
              {chips.map(chip => (
                <motion.button
                  key={chip.key}
                  initial={{ opacity: 0, scale: 0.85 }}
                  animate={{ opacity: 1, scale: 1 }}
                  exit={{ opacity: 0, scale: 0.85 }}
                  onClick={() => removeFilter(chip.key)}
                  className="flex items-center gap-1.5 pl-3 pr-2 py-1.5 rounded-full bg-accent/15 border border-accent/30 text-accent-light text-xs font-medium hover:bg-accent/25 transition-colors"
                >
                  {chip.label}
                  <X className="w-3 h-3" />
                </motion.button>
              ))}
              <motion.button
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                onClick={() => setFilters(DEFAULT_FILTERS)}
                className="px-3 py-1.5 rounded-full bg-surface-800/60 border border-surface-700/40 text-surface-500 hover:text-surface-300 text-xs transition-colors"
              >
                Clear all
              </motion.button>
            </motion.div>
          )}
        </AnimatePresence>

        {/* Track count */}
        {!loading && displayed.length > 0 && (
          <p className="text-surface-600 text-xs mb-4">
            {displayed.length}{hasMore && tracks.length === displayed.length ? '+' : ''} tracks
            {displayed.length !== tracks.length && ` (filtered from ${tracks.length})`}
          </p>
        )}

        {/* Content */}
        {loading ? (
          <div className={clsx(
            viewMode === 'grid'
              ? 'grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 2xl:grid-cols-7 gap-3'
              : 'space-y-1'
          )}>
            {Array.from({ length: 12 }).map((_, i) => <SkeletonCard key={i} viewMode={viewMode} />)}
          </div>
        ) : displayed.length === 0 ? (
          <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="flex flex-col items-center justify-center py-24 text-center">
            <div className="w-16 h-16 rounded-2xl bg-surface-800 flex items-center justify-center mb-4">
              <Music className="w-8 h-8 text-surface-600" />
            </div>
            <h3 className="text-lg font-semibold text-surface-300 mb-2">
              {query || activeCount > 0 ? 'No matches' : 'No tracks yet'}
            </h3>
            <p className="text-surface-500 text-sm mb-6 max-w-xs">
              {query || activeCount > 0
                ? 'Try adjusting your filters.'
                : 'Sync your SoundCloud account to load your feed.'}
            </p>
            {query || activeCount > 0 ? (
              <button
                onClick={() => { setQuery(''); setFilters(DEFAULT_FILTERS) }}
                className="btn-secondary text-sm"
              >
                Clear filters
              </button>
            ) : (
              <button onClick={handleSync} className="btn-primary flex items-center gap-2">
                <RefreshCw className="w-4 h-4" />Sync Now
              </button>
            )}
          </motion.div>
        ) : (
          <div
            style={{
              height: `${virtualizer.getTotalSize()}px`,
              width: '100%',
              position: 'relative',
            }}
          >
            {virtualItems.map((virtualItem) => {
              const rowTracks = rowData[virtualItem.index]
              return (
                <div
                  key={virtualItem.key}
                  data-index={virtualItem.index}
                  ref={virtualizer.measureElement}
                  style={{
                    position: 'absolute',
                    top: 0,
                    left: 0,
                    width: '100%',
                    transform: `translateY(${virtualItem.start}px)`,
                  }}
                >
                  {viewMode === 'grid' ? (
                    <VirtualGridRow
                      tracks={rowTracks}
                      columns={columns}
                      gap={gap}
                      viewMode={viewMode}
                      topArtistNames={topArtistNames}
                      newTrackIdsRef={newTrackIdsRef}
                    />
                  ) : (
                    <div className="space-y-1">
                      {rowTracks.map((track) => (
                        <TrackCard
                          key={track.track_id}
                          track={track}
                          viewMode={viewMode}
                          isNew={newTrackIdsRef.current.has(track.track_id)}
                          isTopArtist={topArtistNames.has(track.artist_name)}
                        />
                      ))}
                    </div>
                  )}
                </div>
              )
            })}
          </div>
        )}

        {/* Infinite scroll sentinel */}
        <div ref={sentinelRef} className="h-16 flex items-center justify-center mt-4">
          {loadingMore && (
            <div className="flex items-center gap-2 text-surface-500 text-sm">
              <div className="w-4 h-4 border-2 border-accent/40 border-t-accent rounded-full animate-spin" />
              Loading more…
            </div>
          )}
          {!hasMore && tracks.length > 0 && !loading && (
            <p className="text-surface-700 text-xs">All tracks loaded</p>
          )}
        </div>

        </div>{/* end main content column */}

        {/* Recently Played sidebar — only visible at 2xl */}
        {recentlyPlayed.length > 0 && (
          <motion.aside
            initial={{ opacity: 0, x: 16 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ delay: 0.08 }}
            className="hidden 2xl:block w-52 flex-shrink-0 sticky top-20 self-start"
          >
            <div className="flex items-center gap-1.5 mb-3">
              <Clock className="w-3.5 h-3.5 text-surface-500" />
              <span className="text-xs font-medium text-surface-500 uppercase tracking-wider">Recently Played</span>
            </div>
            <div className="flex flex-col gap-1.5">
              {recentlyPlayed.map((entry) => (
                <button
                  key={entry.track_id}
                  onClick={() => playTrack(entry)}
                  title={`${entry.track_title} — ${entry.artist_name}`}
                  className="flex items-center gap-2 pl-1.5 pr-2 py-1.5 rounded-xl bg-surface-800/60 border border-surface-700/40 hover:bg-surface-700/60 hover:border-surface-600/50 transition-all group text-left w-full"
                >
                  {entry.artwork_url ? (
                    <img
                      src={entry.artwork_url.replace(/-t\d+x\d+/g, '-t50x50')}
                      alt=""
                      className="w-8 h-8 rounded-md object-cover flex-shrink-0"
                    />
                  ) : (
                    <div className="w-8 h-8 rounded-md bg-surface-700 flex items-center justify-center flex-shrink-0">
                      <Music className="w-3.5 h-3.5 text-surface-500" />
                    </div>
                  )}
                  <div className="min-w-0">
                    <p className="text-xs text-surface-300 truncate group-hover:text-surface-100 transition-colors leading-tight">
                      {entry.track_title}
                    </p>
                    <p className="text-[10px] text-surface-600 truncate leading-tight mt-0.5">
                      {entry.artist_name}
                    </p>
                  </div>
                </button>
              ))}
            </div>
          </motion.aside>
        )}

        </div>{/* end 2xl flex row */}
      </div>

      {/* Filter panel */}
      <FilterPanel
        open={filterOpen}
        onClose={() => setFilterOpen(false)}
        filters={filters}
        onChange={setFilters}
        tracks={tracks}
      />
    </div>
  )
}
