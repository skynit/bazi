import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'

const API_TARGET = process.env.VITE_API_TARGET || 'http://localhost:8088'

export default defineConfig({
  plugins: [vue(), tailwindcss()],
  server: {
    port: Number(process.env.VITE_DEV_PORT) || 5173,
    proxy: { '/api': API_TARGET },
  },
})
