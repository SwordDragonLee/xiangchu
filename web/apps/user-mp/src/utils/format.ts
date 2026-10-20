// 展示格式化与业务字典（一期硬编码）

/** 分 -> 元，统一展示 (fen/100).toFixed(2) */
export function yuan(fen: number): string {
  return (Number(fen || 0) / 100).toFixed(2)
}

export interface OrderStatusInfo {
  text: string
  cls: string
}

/** 订单状态：0待支付 1待接单 2已接单 3备菜中 4服务中 5已完成 6已取消 7已拒绝 */
export const ORDER_STATUS: Record<number, OrderStatusInfo> = {
  0: { text: '待支付', cls: 'tag-red' },
  1: { text: '待接单', cls: 'tag' },
  2: { text: '已接单', cls: 'tag-green' },
  3: { text: '备菜中', cls: 'tag-green' },
  4: { text: '服务中', cls: 'tag-green' },
  5: { text: '已完成', cls: 'tag-grey' },
  6: { text: '已取消', cls: 'tag-grey' },
  7: { text: '已拒绝', cls: 'tag-red' },
}

/** 退款状态：0无 1申请中 2已同意 3已拒绝 4已退款 */
export const REFUND_STATUS: Record<number, string> = {
  0: '',
  1: '退款申请中',
  2: '退款已同意，待退款',
  3: '退款被拒绝',
  4: '已退款',
}

/** 宴席类型（一期硬编码字典） */
export const BANQUET_TYPES: string[] = ['婚宴', '寿宴', '白事', '满月', '节庆']

/** 菜系（一期硬编码字典） */
export const CUISINES: string[] = ['川菜', '湘菜', '粤菜', '家常菜', '本帮菜', '东北菜']

/** 档期状态：0可约 1已锁定 2休息日 */
export const SCHEDULE_STATUS: Record<number, { text: string; cls: string }> = {
  0: { text: '可约', cls: 'free' },
  1: { text: '已锁', cls: 'locked' },
  2: { text: '休息', cls: 'rest' },
}

/** 评价标签 */
export const REVIEW_TAGS: string[] = ['味道好', '分量足', '手艺精湛', '守时', '沟通顺畅', '收拾干净']

const pad = (n: number): string => (n < 10 ? `0${n}` : `${n}`)

/** Date -> 'YYYY-MM-DD' */
export function fmtDate(d: Date): string {
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
}

/** 'YYYY-MM-DD HH:mm:ss' / ISO -> 'YYYY-MM-DD HH:mm' */
export function fmtTime(s: string): string {
  if (!s) return ''
  return s.length >= 16 ? s.slice(0, 16).replace('T', ' ') : s
}

/** 'YYYY-MM-DDTHH:mm...' → 'YYYY-MM-DD'（宴席日期只有天，不带时分） */
export function fmtDay(s: string): string {
  if (!s) return ''
  return s.slice(0, 10)
}

/** 今天的 'YYYY-MM-DD' */
export function today(): string {
  return fmtDate(new Date())
}

/** 未来第 n 天的 'YYYY-MM-DD' */
export function dateAfter(days: number): string {
  const d = new Date()
  d.setDate(d.getDate() + days)
  return fmtDate(d)
}

/** 是否晚于今天（含今天返回 true） */
export function isTodayOrLater(dateStr: string): boolean {
  return dateStr >= today()
}

/** 错误信息提取 */
export function errMsg(e: unknown): string {
  if (e && typeof e === 'object' && 'message' in e) return String((e as { message: unknown }).message)
  return '操作失败'
}
