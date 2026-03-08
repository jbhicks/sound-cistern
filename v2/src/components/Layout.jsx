import { NavLink, Link } from 'react-router-dom'
import { motion, AnimatePresence } from 'framer-motion'
import { Home, Heart, ListMusic, BarChart2, LogOut, Menu, X, Zap } from 'lucide-react'
import { useState } from 'react'
import clsx from 'clsx'
import { useStore } from '../store'
import Player from './Player'
import OfflineIndicator from './OfflineIndicator'

const navItems = [
  { path: '/stream', label: 'Stream', icon: Home },
  { path: '/favorites', label: 'Favorites', icon: Heart },
  { path: '/playlists', label: 'Playlists', icon: ListMusic },
  { path: '/analytics', label: 'Analytics', icon: BarChart2 },
]

export default function Layout({ children }) {
  const [mobileOpen, setMobileOpen] = useState(false)
  const { user, currentTrack } = useStore()

  return (
    <div className="min-h-screen bg-surface-950">
      <OfflineIndicator />
      {/* Ambient background */}
      <div className="fixed inset-0 pointer-events-none overflow-hidden">
        <div className="absolute -top-40 left-1/4 w-[500px] h-[500px] bg-accent/8 rounded-full blur-[120px]" />
        <div className="absolute -bottom-40 right-1/4 w-[500px] h-[500px] bg-vapor-pink/6 rounded-full blur-[120px]" />
      </div>

      {/* Header */}
      <header className="fixed top-0 left-0 right-0 z-40 bg-surface-950/80 backdrop-blur-xl border-b border-surface-800/60">
        <nav className="max-w-7xl mx-auto px-4 h-14 flex items-center justify-between">
          {/* Logo */}
          <Link to="/stream" className="flex items-center gap-2.5 group">
            <motion.div
              whileHover={{ rotate: 180 }}
              transition={{ duration: 0.5 }}
              className="w-8 h-8 rounded-lg bg-gradient-to-br from-accent to-vapor-pink flex items-center justify-center shadow-lg shadow-accent/20"
            >
              <Zap className="w-4 h-4 text-white" />
            </motion.div>
            <span className="text-base font-bold text-gradient hidden sm:block tracking-tight">Sound Cistern</span>
          </Link>

          {/* Desktop nav */}
          <div className="hidden md:flex items-center gap-0.5">
            {navItems.map(item => (
              <NavLink
                key={item.path}
                to={item.path}
                className={({ isActive }) => clsx(
                  'flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-sm font-medium transition-all duration-150',
                  isActive
                    ? 'bg-accent/15 text-accent-light'
                    : 'text-surface-400 hover:text-surface-100 hover:bg-surface-800/60'
                )}
              >
                <item.icon className="w-4 h-4" />
                {item.label}
              </NavLink>
            ))}
          </div>

          {/* Right */}
          <div className="flex items-center gap-2">
            {user && (
              <div className="hidden sm:flex items-center gap-2 px-3 py-1.5 rounded-lg bg-surface-800/60 border border-surface-700/50">
                <div className="w-6 h-6 rounded-full bg-gradient-to-br from-accent to-vapor-pink" />
                <span className="text-sm text-surface-300">{user.name?.trim() || user.email}</span>
              </div>
            )}
            <a
              href="/signout"
              className="p-2 text-surface-500 hover:text-surface-300 hover:bg-surface-800 rounded-lg transition-colors"
              title="Sign out"
            >
              <LogOut className="w-4 h-4" />
            </a>
            <button
              onClick={() => setMobileOpen(!mobileOpen)}
              className="md:hidden p-2 text-surface-400 hover:text-surface-100 hover:bg-surface-800 rounded-lg transition-colors"
            >
              {mobileOpen ? <X className="w-5 h-5" /> : <Menu className="w-5 h-5" />}
            </button>
          </div>
        </nav>

        {/* Mobile menu */}
        <AnimatePresence>
          {mobileOpen && (
            <motion.div
              initial={{ opacity: 0, height: 0 }}
              animate={{ opacity: 1, height: 'auto' }}
              exit={{ opacity: 0, height: 0 }}
              className="md:hidden border-t border-surface-800/60"
            >
              <div className="px-4 py-3 space-y-1">
                {navItems.map(item => (
                  <NavLink
                    key={item.path}
                    to={item.path}
                    onClick={() => setMobileOpen(false)}
                    className={({ isActive }) => clsx(
                      'flex items-center gap-3 px-4 py-2.5 rounded-xl text-sm font-medium transition-all',
                      isActive
                        ? 'bg-accent/15 text-accent-light'
                        : 'text-surface-400 hover:text-surface-100 hover:bg-surface-800/60'
                    )}
                  >
                    <item.icon className="w-5 h-5" />
                    {item.label}
                  </NavLink>
                ))}
              </div>
            </motion.div>
          )}
        </AnimatePresence>
      </header>

      {/* Page content */}
      <main className={clsx('pt-14 relative z-10', currentTrack && 'pb-24')}>
        {children}
      </main>

      {/* Audio player */}
      <Player />
    </div>
  )
}
