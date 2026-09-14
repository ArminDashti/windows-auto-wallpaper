import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'node:path'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  return {
    base: '/',
    plugins: [vue()],
    resolve: {
      alias: {
        '@': path.resolve(__dirname, './src'),
      },
    },
    server: {
      host: true,
      port: 5181,
      proxy: {
        '/api': {
          target: env.VITE_API_PROXY || 'http://127.0.0.1:8101',
          changeOrigin: true,
        },
        '/health': {
          target: env.VITE_API_PROXY || 'http://127.0.0.1:8101',
          changeOrigin: true,
        },
      },
    },
  }
})
