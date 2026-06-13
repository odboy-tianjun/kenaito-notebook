import {defineConfig} from 'vite'
import { createVuePlugin } from 'vite-plugin-vue2'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [createVuePlugin()],
  server: {
    // Wails dev 模式下的配置
    host: '0.0.0.0',
    port: 5173,
    strictPort: true,
    // 不启用 HMR 的 WebSocket 代理，让 Wails 处理
  },
  // 确保资源路径正确
  base: './',
  build: {
    // 确保输出目录正确
    outDir: 'dist',
    assetsDir: 'assets'
  }
})
