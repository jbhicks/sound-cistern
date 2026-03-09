import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig(({ command }) => ({
  plugins: [react()],
  base: command === 'build' ? '/app/' : '/',
  build: {
    outDir: '../public/app',
    emptyOutDir: true,
  },
  server: {
    port: 5173,
    host: '0.0.0.0',
    // Allow the Cloudflare tunnel host
    allowedHosts: ['soundcistern.jbhicks.dev', '.local', '192.168.1.49'],
    hmr: {
      // Allow HMR to work when accessed from different hosts
      // Vite will auto-detect the host from the request
      host: '0.0.0.0',
    },
    proxy: {
      '/api': {
        target: 'http://localhost:8090',
        changeOrigin: true,
      },
      '/auth': {
        target: 'http://localhost:8090',
        changeOrigin: true,
      },
      '/signout': {
        target: 'http://localhost:8090',
        changeOrigin: true,
      },
      '/js': {
        target: 'http://localhost:8090',
        changeOrigin: true,
      },
    },
  },
}))
