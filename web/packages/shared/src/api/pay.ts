// 财务域 API（收入/提现 + 模拟支付）
import type { ApiClient } from '../request'
import type { Income, Withdrawal } from '../types'

export function createPayApi(c: ApiClient) {
  return {
    mockPay: (paymentNo: string) => c.post(`/pay/mock/${paymentNo}`),
    chefIncome: () => c.get<Income>('/chef/income'),
    chefWithdraw: (amount: number) => c.post('/chef/withdrawals', { amount }),
    chefWithdrawals: () => c.get<Withdrawal[]>('/chef/withdrawals'),
    supplierIncome: () => c.get<Income>('/supplier/income'),
    supplierWithdraw: (amount: number) => c.post('/supplier/withdrawals', { amount }),
    supplierWithdrawals: () => c.get<Withdrawal[]>('/supplier/withdrawals'),
  }
}
