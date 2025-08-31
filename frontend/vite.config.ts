import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  build: {
    outDir: '../backend/dist'
  },
  server: {
    port: 5081,
    proxy: {
      '/api': {
        target: 'http://localhost:5080',
        changeOrigin: true,
      }
    }
  }
})
