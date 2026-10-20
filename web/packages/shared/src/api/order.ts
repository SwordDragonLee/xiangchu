// 订单域 API（用户下单 + 乡厨接单 + 管理端）
import type { ApiClient } from '../request'
import type { AfterSale, Order, PageData, PayResult, Review } from '../types'

export function createOrderApi(c: ApiClient) {
  return {
    create: (data: {
      chef_id: number
      package_id: number
      banquet_type: string
      banquet_time: string
      table_count: number
      people_count?: number
      region: string
      address: string
      contact_name: string
      contact_phone: string
      menu_remark?: string
      user_coupon_id?: number
      pay_type: 1 | 2
    }) => c.post<{ order_id: number; order_no: string; deposit: number; payment_no: string; params: PayResult['params'] }>('/orders', data),
    list: (params: { status?: number; page?: number; size?: number } = {}) => c.get<PageData<Order>>('/orders', params),
    detail: (id: number) => c.get<Order>(`/orders/${id}`),
    cancel: (id: number) => c.post(`/orders/${id}/cancel`),
    payBalance: (id: number) => c.post<PayResult>(`/orders/${id}/pay-balance`),
    review: (id: number, data: { rating: number; tags?: string[]; content: string; images?: string[] }) => c.post(`/orders/${id}/review`, data),
    applyAfterSale: (id: number, data: { amount: number; reason: string; images?: string[] }) => c.post(`/orders/${id}/aftersale`, data),
  }
}

export function createChefOrderApi(c: ApiClient) {
  return {
    list: (params: { status?: number; page?: number; size?: number } = {}) => c.get<PageData<Order>>('/chef/orders', params),
    detail: (id: number) => c.get<Order>(`/chef/orders/${id}`),
    accept: (id: number) => c.post(`/chef/orders/${id}/accept`),
    reject: (id: number, reason: string) => c.post(`/chef/orders/${id}/reject`, { reason }),
    startPrep: (id: number) => c.post(`/chef/orders/${id}/start-prep`),
    startService: (id: number) => c.post(`/chef/orders/${id}/start-service`),
    complete: (id: number) => c.post(`/chef/orders/${id}/complete`),
  }
}
