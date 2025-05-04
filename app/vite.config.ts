// vite.config.ts
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import svgr from 'vite-plugin-svgr'

export default defineConfig({
  plugins: [react(), svgr()],
  server: {
    proxy: {
      // Проксі для Ollama API
      '/ollama-api': {
        target: 'http://localhost:11434', // Твій Ollama endpoint
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/ollama-api/, '/api'),
      },
      // --- НОВЕ ПРОКСІ для твого бекенду ---
      '/backend-api': {
        target: 'http://localhost:3001', // Порт твого бекенд-сервера
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/backend-api/, '/api'), // Переписуємо шлях на /api
      },
    },
  },
})