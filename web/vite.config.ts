import tailwindcss from '@tailwindcss/vite'
import path from 'path'
import devtools from 'solid-devtools/vite'
import { defineConfig } from 'vite'
import solid from 'vite-plugin-solid'

export default defineConfig({
  plugins: [
    devtools({
      autoname: true,
      locator: true,
    }),
    solid(),
    tailwindcss(),
  ],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  envPrefix: 'APP_',
})
