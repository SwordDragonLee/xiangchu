// 采购选购车：模块级单例（不持久化，页面栈内存活）
// 后端一张采购单只支持一个 supplier_id，故跨供应商加购时整体替换
import { computed, reactive } from 'vue'
import type { Ingredient } from '@xc/shared'

export interface CartItem {
  ingredient_id: number
  name: string
  unit: string
  price: number
  quantity: number
  stock: number
}

const state = reactive<{ supplierId: number; items: CartItem[] }>({
  supplierId: 0,
  items: [],
})

/** 加购；返回 true 表示因换供应商清空了已选（页面据此提示） */
export function useCart() {
  const count = computed(() => state.items.reduce((n, i) => n + i.quantity, 0))
  const totalFen = computed(() => state.items.reduce((n, i) => n + i.price * i.quantity, 0))

  function add(ing: Ingredient): boolean {
    if ((ing.stock ?? 0) < 1) {
      uni.showToast({ title: '库存不足', icon: 'none' })
      return false
    }
    let switched = false
    if (state.supplierId && state.supplierId !== ing.supplier_id) {
      state.items = []
      switched = true
    }
    state.supplierId = ing.supplier_id
    const exist = state.items.find((i) => i.ingredient_id === ing.id)
    if (exist) {
      if (exist.quantity >= ing.stock) {
        uni.showToast({ title: '已达库存上限', icon: 'none' })
        return switched
      }
      exist.quantity += 1
    } else {
      state.items.push({
        ingredient_id: ing.id,
        name: ing.name,
        unit: ing.unit,
        price: ing.price,
        quantity: 1,
        stock: ing.stock,
      })
    }
    return switched
  }

  function setQty(id: number, qty: number): void {
    const it = state.items.find((i) => i.ingredient_id === id)
    if (!it) return
    it.quantity = Math.min(Math.max(1, qty), it.stock)
  }

  function remove(id: number): void {
    state.items = state.items.filter((i) => i.ingredient_id !== id)
    if (state.items.length === 0) state.supplierId = 0
  }

  function clear(): void {
    state.items = []
    state.supplierId = 0
  }

  return { state, count, totalFen, add, setQty, remove, clear }
}
