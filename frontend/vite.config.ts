import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from "path"

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      "@local": path.resolve(__dirname, "src/")
    }
  },

  server: {
    port: 8001,
    host: "0.0.0.0",
  }
})
