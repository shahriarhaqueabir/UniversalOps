/// <reference types="vitest" />
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import path from 'path'
import compression from 'vite-plugin-compression'

export default defineConfig({
  plugins: [
    react(),
    tailwindcss(),
    compression({
      verbose: false,
      disable: false,
      threshold: 10240,
      algorithm: 'gzip',
      ext: '.gz',
    }),
  ],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    host: '127.0.0.1',
    port: 5173,
    strictPort: true,
  },
  build: {
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (!id.includes('node_modules')) return undefined

          if (
            id.includes('node_modules/react/') ||
            id.includes('node_modules/react-dom/') ||
            id.includes('node_modules/scheduler/') ||
            id.includes('node_modules/use-sync-external-store/')
          ) {
            return 'vendor-react'
          }

          if (id.includes('node_modules/@tanstack/')) {
            return 'vendor-tanstack'
          }

          if (id.includes('node_modules/@radix-ui/')) {
            return 'vendor-radix'
          }

          if (id.includes('node_modules/motion/')) {
            return 'vendor-motion'
          }

          if (id.includes('node_modules/recharts/')) {
            return 'vendor-recharts'
          }

          if (id.includes('node_modules/lucide-react/')) {
            return 'vendor-icons'
          }

          if (
            id.includes('node_modules/date-fns/') ||
            id.includes('node_modules/clsx/') ||
            id.includes('node_modules/tailwind-merge/') ||
            id.includes('node_modules/class-variance-authority/')
          ) {
            return 'vendor-style'
          }

          if (id.includes('node_modules/sonner/')) {
            return 'vendor-toast'
          }

          if (id.includes('node_modules/ajv/')) {
            return 'vendor-schema'
          }

          if (id.includes('node_modules/nanoid/')) {
            return 'vendor-id'
          }

          if (
            id.includes('node_modules/lodash/') ||
            id.includes('node_modules/lodash-es/')
          ) {
            return 'vendor-utils'
          }

          if (id.includes('node_modules/zustand/')) {
            return 'vendor-state'
          }

          return 'vendor'
        },
      },
    },
  },
})
