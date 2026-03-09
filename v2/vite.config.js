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
    allowedHosts: ['soundcistern.com', '.local', '192.168.1.49'],
    hmr: {
      // Use the same host as the page for HMR WebSocket
      // This makes it work whether accessed via localhost or IP
      protocol: 'ws',
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
