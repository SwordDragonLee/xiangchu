// 会话状态：token + role + userId，持久化在 uni storage
import { computed, reactive } from 'vue'
import type { LoginResult } from '@xc/shared/types'

interface SessionState {
  token: string
  role: string
  userId: number
}

const state = reactive<SessionState>({
  token: (uni.getStorageSync('token') as string) || '',
  role: (uni.getStorageSync('role') as string) || '',
  userId: (uni.getStorageSync('userId') as number) || 0,
})

function persist() {
  if (state.token) uni.setStorageSync('token', state.token)
  else uni.removeStorageSync('token')
  if (state.role) uni.setStorageSync('role', state.role)
  else uni.removeStorageSync('role')
  if (state.userId) uni.setStorageSync('userId', state.userId)
  else uni.removeStorageSync('userId')
}

export function useSession() {
  const isLogged = computed(() => !!state.token)

  function save(result: LoginResult) {
    state.token = result.token
    state.role = result.role
    state.userId = result.user_id
    persist()
  }

  function clear() {
    state.token = ''
    state.role = ''
    state.userId = 0
    persist()
  }

  return { state, isLogged, save, clear }
}
