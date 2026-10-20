// 采购供应域 API（食材市场 + 采购单 + 供应商工作台）
import type { ApiClient } from '../request'
import type { Ingredient, PageData, PurchaseOrder, Supplier } from '../types'

export function createFoodApi(c: ApiClient) {
  return {
    ingredients: (params: { keyword?: string; category?: string; supplier_id?: number; page?: number; size?: number }) =>
      c.get<PageData<Ingredient>>('/ingredients', params),
    createPurchase: (data: {
      supplier_id: number
      order_id?: number
      items: { ingredient_id: number; quantity: number }[]
      fulfill_method: 1 | 2
      deliver_address?: string
      deliver_time?: string
      remark?: string
    }) => c.post<{ purchase_id: number; purchase_no: string; total_amount: number; payment_no: string; params: unknown }>('/purchases', data),
    myPurchases: (page = 1, size = 10, status?: number) =>
      c.get<PageData<PurchaseOrder>>('/purchases', { page, size, status: status ?? undefined }),
    purchase: (id: number) => c.get<PurchaseOrder>(`/purchases/${id}`),
    cancelPurchase: (id: number) => c.post(`/purchases/${id}/cancel`),
  }
}

export function createSupplierApi(c: ApiClient) {
  return {
    profile: () => c.get<Supplier>('/supplier/profile'),
    applyStatus: () => c.get<Supplier>('/supplier/apply-status'),
    apply: (data: { name: string; license: string; contact: string; phone: string; region: string; address: string; categories: string }) => c.post('/supplier/apply', data),
    updateProfile: (data: Partial<{ name: string; license: string; contact: string; phone: string; region: string; address: string; categories: string }>) => c.put('/supplier/profile', data),
    ingredients: (params: { keyword?: string; category?: string; page?: number; size?: number } = {}) => c.get<PageData<Ingredient>>('/supplier/ingredients', params),
    saveIngredient: (data: Partial<Ingredient> & { name: string; category: string; unit: string; price: number }, id?: number) =>
      id ? c.put<{ id: number }>(`/supplier/ingredients/${id}`, data) : c.post<{ id: number }>('/supplier/ingredients', data),
    offShelfIngredient: (id: number) => c.del(`/supplier/ingredients/${id}`),
    purchases: (params: { status?: number; page?: number; size?: number } = {}) => c.get<PageData<PurchaseOrder>>('/supplier/purchases', params),
    purchase: (id: number) => c.get<PurchaseOrder>(`/supplier/purchases/${id}`),
    acceptPurchase: (id: number) => c.post(`/supplier/purchases/${id}/accept`),
    rejectPurchase: (id: number, reason: string) => c.post(`/supplier/purchases/${id}/reject`, { reason }),
    readyPurchase: (id: number) => c.post(`/supplier/purchases/${id}/ready`),
    completePurchase: (id: number) => c.post(`/supplier/purchases/${id}/complete`),
  }
}
