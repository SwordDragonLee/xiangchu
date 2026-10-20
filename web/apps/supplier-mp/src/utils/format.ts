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

/** 食材分类（与后端/管理端一致） */
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

export const SUPPLIER_STATUS_TEXT: Record<number, string> = {
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

/** 解析逗号分隔的品类字符串为数组 */
export function splitCategories(s?: string | null): string[] {
  if (!s) return []
  return s
    .split(/[,，、\s]+/)
    .map((x) => x.trim())
    .filter(Boolean)
}
