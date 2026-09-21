import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'
import tailwindcss from '@tailwindcss/vite'
import pkg from './package.json'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  // single source of truth for the app version — injected from package.json
  define: {
    __APP_VERSION__: JSON.stringify(pkg.version),
  },
})
