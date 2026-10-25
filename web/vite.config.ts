import { defineConfig } from 'vitest/config'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: { '@': fileURLToPath(new URL('./src', import.meta.url)) },
  },
  server: {
    port: 5173,
    // Bind mounts from Windows do not deliver fs events into the container.
    watch: { usePolling: !!process.env.VITE_POLL },
    proxy: { '/api': process.env.VITE_API_PROXY ?? 'http://api:8080' },
  },
  test: {
    environment: 'node',
  },
})
