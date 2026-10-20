// 用户域 API
import type { ApiClient } from '../request'
import type { Address, Coupon, LoginResult, PageData, ProfileVO, UserCoupon } from '../types'

export function createAuthApi(c: ApiClient) {
  return {
    sendCode: (phone: string) => c.post('/user/auth/code', { phone }),
    loginByCode: (phone: string, code: string) => c.post<LoginResult>('/user/auth/login', { phone, code }),
    wxLogin: (code: string) => c.post<LoginResult>('/user/auth/wx', { code }),
    webLogin: (phone: string, password: string) => c.post<LoginResult>('/user/auth/web', { phone, password }),
    // 角色变更（如入驻审核通过）后凭有效 token 换发含最新角色的新 JWT；
    // 双身份账号（既是乡厨又是供应商）可传 wantRole 显式指定按哪侧档案签发
    refresh: (wantRole?: 'chef' | 'supplier') =>
      c.post<LoginResult>('/user/auth/refresh', wantRole ? { role: wantRole } : undefined),
  }
}

export function createUserApi(c: ApiClient) {
  return {
    profile: () => c.get<ProfileVO>('/user/profile'),
    updateProfile: (data: { nickname?: string; avatar?: string }) => c.put('/user/profile', data),
    setPassword: (password: string) => c.put('/user/password', { password }),
    addresses: () => c.get<Address[]>('/user/addresses'),
    saveAddress: (data: Partial<Address> & { contact_name: string; contact_phone: string; address: string }, id?: number) =>
      id ? c.put<{ id: number }>(`/user/addresses/${id}`, data) : c.post<{ id: number }>('/user/addresses', data),
    deleteAddress: (id: number) => c.del(`/user/addresses/${id}`),
    couponCenter: (page = 1, size = 10) => c.get<PageData<Coupon>>('/coupons', { page, size }),
    myCoupons: () => c.get<UserCoupon[]>('/user/coupons'),
    receiveCoupon: (id: number) => c.post(`/user/coupons/${id}/receive`),
  }
}
