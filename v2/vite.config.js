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
    // Allow the Cloudflare tunnel host
    allowedHosts: ['soundcistern.jbhicks.dev'],
    hmr: {
      // When accessed via the Cloudflare tunnel, HMR WebSocket must also
      // go through the tunnel (port 443, wss). Vite will use the page's
      // host by default which is correct — just need clientPort to match.
      clientPort: 443,
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
