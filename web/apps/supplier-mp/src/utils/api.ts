// API 客户端装配：uni.request + token 注入 + 401 统一处理
import {
  ApiError,
  createAuthApi,
  createClient,
  createPayApi,
  createSupplierApi,
  uniRequester,
} from '@xc/shared'

export { ApiError }

export const API_BASE: string = import.meta.env.VITE_API_BASE || 'http://localhost:8080'

export const TOKEN_KEY = 'token'
export const ROLE_KEY = 'role'

export function getToken(): string {
  const t = uni.getStorageSync(TOKEN_KEY)
  return typeof t === 'string' ? t : ''
}

export function getRole(): string {
  const r = uni.getStorageSync(ROLE_KEY)
  return typeof r === 'string' ? r : ''
}

export function saveSession(token: string, role: string): void {
  uni.setStorageSync(TOKEN_KEY, token)
  uni.setStorageSync(ROLE_KEY, role)
}

export function clearSession(): void {
  uni.removeStorageSync(TOKEN_KEY)
  uni.removeStorageSync(ROLE_KEY)
}

function goLogin(): void {
  clearSession()
  uni.reLaunch({ url: '/pages/login/index' })
}

const client = createClient({
  base: API_BASE,
  requester: uniRequester(),
  getToken: () => getToken() || undefined,
  onUnauthorized: goLogin,
})

export const authApi = createAuthApi(client)
export const supplierApi = createSupplierApi(client)
export const payApi = createPayApi(client)

/** 未登录则跳登录页，返回是否已登录 */
export function ensureLogin(): boolean {
  if (getToken()) return true
  goLogin()
  return false
}

/** 非 supplier 角色则先静默换发（审核刚通过/双身份账号的场景），仍不符再引导去入驻申请页 */
export async function ensureSupplier(): Promise<boolean> {
  if (getRole() === 'supplier') return true
  // 本地角色可能已过期或为双身份账号（resolveRole chef 优先）：显式按 supplier 档案换发
  try {
    const r = await authApi.refresh('supplier')
    saveSession(r.token, r.role)
    if (r.role === 'supplier') {
      uni.showToast({ title: '认证已生效', icon: 'success' })
      return true
    }
  } catch {
    /* token 失效由 401 统一处理 */
  }
  uni.showToast({ title: '您还不是认证供应商', icon: 'none' })
  setTimeout(() => {
    uni.reLaunch({ url: '/pages/apply/index' })
  }, 800)
  return false
}

export function toast(msg: string): void {
  uni.showToast({ title: msg, icon: 'none' })
}

export function errMsg(e: unknown): string {
  if (e instanceof ApiError) return e.message
  if (e instanceof Error) return e.message
  return '操作失败，请重试'
}
