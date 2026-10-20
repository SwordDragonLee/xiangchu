import axios from 'axios'
import { ElMessage } from 'element-plus'

import { createClient, axiosRequester } from '@xc/shared/request'
import { createAdminApi } from '@xc/shared/api/admin'
import { createSupplierApi } from '@xc/shared/api/food'
import { createPayApi } from '@xc/shared/api/pay'

import router from '@/router'

export const TOKEN_KEY = 'xc_admin_token'

// axios 实例
const instance = axios.create({
  baseURL: (import.meta.env.VITE_API_BASE as string | undefined) || 'http://localhost:8080',
  timeout: 15000,
})

// 请求拦截器：注入 Bearer Token
instance.interceptors.request.use((config) => {
  const token = localStorage.getItem(TOKEN_KEY)
  if (token) {
    config.headers = config.headers ?? {}
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// 响应拦截器：code !== 0 统一弹出错误信息（40100 由 onUnauthorized 处理）
instance.interceptors.response.use(
  (res) => {
    const body = res?.data as { code?: number; msg?: string } | undefined
    if (body && typeof body === 'object' && 'code' in body && body.code !== 0 && body.code !== 40100) {
      ElMessage.error(body.msg || '请求失败')
    }
    return res
  },
  (err: unknown) => {
    const e = err as { response?: { status: number; data?: unknown } }
    if (!e?.response) {
      ElMessage.error('网络异常，请稍后重试')
    }
    return Promise.reject(err)
  },
)

function gotoLogin() {
  localStorage.removeItem(TOKEN_KEY)
  if (router.currentRoute.value.path !== '/login') {
    router.push('/login')
  }
}

// 底层 ApiClient（可访问 createAdminApi 之外的接口）
export const client = createClient({
  base: instance.defaults.baseURL as string,
  requester: axiosRequester(instance),
  getToken: () => localStorage.getItem(TOKEN_KEY) || undefined,
  onUnauthorized: () => {
    ElMessage.error('登录已失效，请重新登录')
    gotoLogin()
  },
})

// 管理端 API
export const api = createAdminApi(client)

// 供应商工作台 API（/supplier/*，supplier 角色 JWT）
export const supplierApi = createSupplierApi(client)

// 财务 API（收入/提现/模拟支付）
export const payApi = createPayApi(client)

// 供应商/乡厨 Web 登录（users 表手机号+密码，后端按已审核档案签发角色 JWT）
export function webLogin(phone: string, password: string) {
  return client.post<{ token: string; role: string; user_id: number }>('/user/auth/web', { phone, password })
}

export default api
