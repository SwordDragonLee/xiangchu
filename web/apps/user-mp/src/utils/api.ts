// 全局 API 客户端：基于 @xc/shared 的 createClient + uniRequester
import { createClient, uniRequester } from '@xc/shared/request'
import { createAuthApi, createUserApi } from '@xc/shared/api/user'
import { createChefPublicApi, createChefApi } from '@xc/shared/api/chef'
import { createFoodApi, createSupplierApi } from '@xc/shared/api/food'
import { createOrderApi } from '@xc/shared/api/order'
import { createPayApi } from '@xc/shared/api/pay'

const base = (import.meta.env.VITE_API_BASE as string | undefined) || 'http://localhost:8080'

const client = createClient({
  base,
  requester: uniRequester(),
  getToken: () => uni.getStorageSync('token') || undefined,
  onUnauthorized: () => {
    uni.removeStorageSync('token')
    uni.reLaunch({ url: '/pages/login/index' })
  },
})

export const apiBase = base

export const authApi = createAuthApi(client)
export const userApi = createUserApi(client)
export const chefPublicApi = createChefPublicApi(client)
// 入驻申请用（user 角色 JWT 即可调用 /chef/apply、/supplier/apply）
export const chefApi = createChefApi(client)
export const supplierApi = createSupplierApi(client)
export const foodApi = createFoodApi(client)
export const orderApi = createOrderApi(client)
export const payApi = createPayApi(client)
