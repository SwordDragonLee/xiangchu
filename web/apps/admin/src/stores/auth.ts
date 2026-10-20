import { defineStore } from 'pinia'

import { api, webLogin, TOKEN_KEY } from '@/api/client'

const NAME_KEY = 'xc_admin_username'
const ROLE_KEY = 'xc_admin_role'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem(TOKEN_KEY) || '',
    username: localStorage.getItem(NAME_KEY) || '',
    role: localStorage.getItem(ROLE_KEY) || '',
  }),
  getters: {
    isLoggedIn: (state) => !!state.token,
  },
  actions: {
    // 登录：获取 token，并拉取管理员信息补充用户名
    async login(username: string, password: string) {
      const res = await api.login(username, password)
      this.token = res.token
      this.role = res.role || ''
      localStorage.setItem(TOKEN_KEY, res.token)
      localStorage.setItem(ROLE_KEY, this.role)
      try {
        const me = await api.me()
        this.username = me.username || username
      } catch {
        this.username = username
      }
      localStorage.setItem(NAME_KEY, this.username)
      return res
    },
    // 供应商/乡厨 Web 登录：手机号+密码，后端按已审核档案签发角色
    async webLogin(phone: string, password: string) {
      const res = await webLogin(phone, password)
      if (!res.role || res.role === 'user') {
        throw new Error('该账号暂无后台权限，请先完成入驻并通过审核')
      }
      this.token = res.token
      this.role = res.role
      this.username = phone.replace(/^(\d{3})\d{4}(\d{4})$/, '$1****$2')
      localStorage.setItem(TOKEN_KEY, res.token)
      localStorage.setItem(ROLE_KEY, this.role)
      localStorage.setItem(NAME_KEY, this.username)
      return res
    },
    logout() {
      this.token = ''
      this.username = ''
      this.role = ''
      localStorage.removeItem(TOKEN_KEY)
      localStorage.removeItem(NAME_KEY)
      localStorage.removeItem(ROLE_KEY)
    },
  },
})
