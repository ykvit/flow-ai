/// <reference types="vitest" />
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      // Proxy all requests starting with /api to the backend server.
      // This is crucial for avoiding CORS issues during development.
      '/api': {
        target: 'http://localhost:8000', // The address of your Go backend
        changeOrigin: true, // Recommended for virtual hosts
      },
    },
  },
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: './src/test/setup.ts',
  },
});
