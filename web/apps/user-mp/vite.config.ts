import { defineConfig } from 'vite'
import uni from '@dcloudio/vite-plugin-uni'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [uni()],
  server: {
    port: 5174, // 固定端口，避免与 admin(5173)/chef-mp(5175) 冲突
  },
})
