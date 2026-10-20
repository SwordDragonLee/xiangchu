// API 客户端装配：uni.request + token 注入 + 401 统一处理
import {
  ApiError,
  createAuthApi,
  createChefApi,
  createChefOrderApi,
  createClient,
  createFoodApi,
  createPayApi,
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
export const chefApi = createChefApi(client)
export const chefOrderApi = createChefOrderApi(client)
export const foodApi = createFoodApi(client)
export const payApi = createPayApi(client)

/** 未登录则跳登录页，返回是否已登录 */
export function ensureLogin(): boolean {
  if (getToken()) return true
  goLogin()
  return false
}

/** 非 chef 角色则先静默换发（审核刚通过的场景），仍不符再引导去入驻申请页 */
export async function ensureChef(): Promise<boolean> {
  if (getRole() === 'chef') return true
  // 本地角色可能已过期（如刚通过入驻审核）：换发一次拿到最新角色
  try {
    const r = await authApi.refresh()
    saveSession(r.token, r.role)
    if (r.role === 'chef') {
      uni.showToast({ title: '认证已生效', icon: 'success' })
      return true
    }
  } catch {
    /* token 失效由 401 统一处理 */
  }
  uni.showToast({ title: '您还不是认证乡厨', icon: 'none' })
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
