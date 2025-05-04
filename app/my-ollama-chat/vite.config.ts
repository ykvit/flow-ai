// vite.config.ts
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import svgr from 'vite-plugin-svgr' 

export default defineConfig({
  plugins: [react(),
    svgr()
  ],
  server: {
    proxy: {
      '/ollama-api': { 
        target: 'http://localhost:11434', 
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/ollama-api/, '/api'),
      },
    },
  },
})