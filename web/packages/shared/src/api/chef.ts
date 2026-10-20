// 乡厨域 API（公开浏览 + 乡厨工作台）
import type { ApiClient } from '../request'
import type { Chef, ChefDetail, ChefPackage, ChefSchedule, PageData, Review } from '../types'

export function createChefPublicApi(c: ApiClient) {
  return {
    list: (params: { region?: string; cuisine?: string; keyword?: string; min_rating?: number; order_by?: string; page?: number; size?: number }) =>
      c.get<PageData<Chef>>('/chefs', params),
    detail: (id: number) => c.get<ChefDetail>(`/chefs/${id}`),
    schedules: (id: number, from: string, to: string) => c.get<ChefSchedule[]>(`/chefs/${id}/schedules`, { from, to }),
    reviews: (id: number, page = 1, size = 10) => c.get<PageData<Review>>(`/chefs/${id}/reviews`, { page, size }),
  }
}

export function createChefApi(c: ApiClient) {
  return {
    workbench: () => c.get<{ today_orders: number; pending_orders: number; month_income: number }>('/chef/workbench'),
    profile: () => c.get<Chef>('/chef/profile'),
    // 本人入驻申请（含待审核/被驳回，user 角色即可调用）
    applyStatus: () => c.get<Chef>('/chef/apply-status'),
    apply: (data: { real_name: string; phone: string; region: string; cuisines: string; experience: string }) => c.post('/chef/apply', data),
    updateProfile: (data: Partial<{ real_name: string; phone: string; region: string; cuisines: string; experience: string }>) => c.put('/chef/profile', data),
    schedules: (from: string, to: string) => c.get<ChefSchedule[]>('/chef/schedules', { from, to }),
    setRestDays: (dates: string[], rest: boolean) => c.put('/chef/schedules/rest-days', { dates, rest }),
    packages: () => c.get<ChefPackage[]>('/chef/packages'),
    savePackage: (data: Partial<ChefPackage> & { name: string; banquet_type: string; price_per_table: number }, id?: number) =>
      id ? c.put<{ id: number }>(`/chef/packages/${id}`, data) : c.post<{ id: number }>('/chef/packages', data),
    deletePackage: (id: number) => c.del(`/chef/packages/${id}`),
  }
}
