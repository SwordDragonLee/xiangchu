// 金额与常量工具（金额单位：分）

/** 分 -> 元字符串，如 12345 -> "123.45" */
export function fen2yuan(fen?: number | null): string {
  return ((fen ?? 0) / 100).toFixed(2)
}

/** 元(字符串/数字) -> 分（四舍五入），非法输入返回 0 */
export function yuan2fen(yuan: string | number): number {
  const n = typeof yuan === 'number' ? yuan : parseFloat(yuan)
  if (!Number.isFinite(n)) return 0
  return Math.round(n * 100)
}

export const CUISINE_OPTIONS = ['川菜', '湘菜', '粤菜', '鲁菜', '苏菜', '浙菜', '闽菜', '徽菜', '家常菜']

export const BANQUET_TYPES = ['婚宴', '寿宴', '满月宴', '乔迁宴', '升学宴', '生日宴', '其他']

export const ORDER_STATUS_TEXT: Record<number, string> = {
  0: '待支付',
  1: '待接单',
  2: '已接单',
  3: '备菜中',
  4: '服务中',
  5: '已完成',
  6: '已取消',
  7: '已拒绝',
}

export const CHEF_STATUS_TEXT: Record<number, string> = {
  0: '待审核',
  1: '已通过',
  2: '已拒绝',
  3: '已禁用',
}

export const WITHDRAW_STATUS_TEXT: Record<number, string> = {
  0: '待审核',
  1: '待打款',
  2: '已拒绝',
  3: '已打款',
}

/** 食材分类（与后端/供应商端一致，服务端按精确字符串过滤） */
export const INGREDIENT_CATEGORIES = ['肉类', '蔬菜', '水产', '干货', '调味品', '粮油', '其他']

export const PURCHASE_STATUS_TEXT: Record<number, string> = {
  0: '待接单',
  1: '备货中',
  2: '待自取/待送达',
  3: '已完成',
  4: '已取消',
}

export const FULFILL_METHOD_TEXT: Record<number, string> = {
  1: '自取',
  2: '送货',
}

/** Date -> "YYYY-MM-DD"（本地时区） */
export function fmtDate(d: Date): string {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

/** 解析逗号分隔的菜系字符串为数组 */
export function splitCuisines(s?: string | null): string[] {
  if (!s) return []
  return s
    .split(/[,，、\s]+/)
    .map((x) => x.trim())
    .filter(Boolean)
}

/** 解析 JSON 数组字符串（套餐菜品清单），失败返回空数组 */
export function parseDishes(s?: string | null): string[] {
  if (!s) return []
  try {
    const arr = JSON.parse(s)
    if (Array.isArray(arr)) return arr.map((x) => String(x))
    return []
  } catch {
    return splitCuisines(s)
  }
}
