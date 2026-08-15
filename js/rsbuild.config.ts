import { defineConfig } from '@rsbuild/core'
import { pluginReact } from '@rsbuild/plugin-react'

// The dev server proxies /api to goapi so the browser makes same-origin
// requests. Override the target with GOAPI_URL when goapi runs elsewhere.
const goapiUrl = process.env.GOAPI_URL ?? 'http://localhost:8080'

export default defineConfig({
  plugins: [pluginReact()],
  html: {
    title: 'Today',
  },
  source: {
    entry: {
      index: './src/main.tsx',
    },
  },
  server: {
    host: '0.0.0.0',
    port: 3000,
    proxy: {
      '/api': {
        target: goapiUrl,
        changeOrigin: true,
      },
    },
  },
})
