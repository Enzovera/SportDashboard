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
    // Chat goes straight to coach-chat in dev; in production the Go API fronts it under the same path.
    proxy: {
      '/api/chat': process.env.VITE_CHAT_PROXY ?? 'http://coach-chat:8090',
      '/api': process.env.VITE_API_PROXY ?? 'http://api:8080',
    },
  },
  build: {
    // One shared ECharts chunk for every view that draws charts.
    rollupOptions: {
      output: { manualChunks: { echarts: ['echarts/core', 'echarts/charts', 'echarts/components', 'echarts/renderers'] } },
    },
    chunkSizeWarningLimit: 600,
  },
  test: {
    environment: 'node',
  },
})
