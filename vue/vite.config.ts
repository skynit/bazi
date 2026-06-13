import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import { fileURLToPath, URL } from 'node:url'

const API_TARGET = process.env.VITE_API_TARGET || 'http://localhost:8088'

export default defineConfig({
  plugins: [vue(), tailwindcss()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  },
  server: {
    port: Number(process.env.VITE_DEV_PORT) || 5174,
    proxy: { '/api': API_TARGET },
  },
})
