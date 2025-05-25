import tailwindcss from '@tailwindcss/vite'
import path from 'path'
import devtools from 'solid-devtools/vite'
import { defineConfig, ServerOptions } from 'vite'
import solid from 'vite-plugin-solid'

const useProxy = process.env.USE_VITE_PROXY == 'true'

let proxyServer: ServerOptions | undefined
if (useProxy) {
  proxyServer = {
    proxy: {
      '/api': {
        target: 'https://api-staging.treenq.com',
        changeOrigin: true,
        secure: false,
        rewrite: (path) => path.replace(/^\/api/, ''),
        configure: (proxy, options) => {
          proxy.on('proxyRes', (proxyRes, req, res) => {
            const cookies = proxyRes.headers['set-cookie']
            if (cookies) {
              proxyRes.headers['set-cookie'] = cookies.map((cookie) =>
                cookie.replace(/;\s*Secure/i, '').replace(/domain=[^;]+/i, 'domain=localhost'),
              )
            }
          })
        },
      },
    },
  }
}

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
  server: proxyServer,
})
