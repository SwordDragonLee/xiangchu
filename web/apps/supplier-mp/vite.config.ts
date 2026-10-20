import { defineConfig } from 'vite'
import uni from '@dcloudio/vite-plugin-uni'

export default defineConfig({
  plugins: [uni()],
  server: {
    port: 5176, // 固定端口，避免与 admin(5173)/user-mp(5174)/chef-mp(5175) 冲突
  },
})
