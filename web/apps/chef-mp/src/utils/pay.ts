// 支付收口：mock 支付弹确认框调 mockPay；真实支付走 uni.requestPayment
// （复制自 user-mp/src/utils/pay.ts，适配 chef-mp 的 errMsg 在 ./api）
import type { PayResult } from '@xc/shared/types'
import { errMsg, payApi } from './api'

export function handlePayResult(result: PayResult): Promise<void> {
  // pay 模块未接线时后端可能返回空支付单，直接失败而不是走假支付
  if (!result.payment_no || !result.params) {
    return Promise.reject(new Error('支付单未创建，请稍后重试'))
  }
  if (result.params.mock) {
    return new Promise((resolve, reject) => {
      uni.showModal({
        title: '确认支付',
        content: '开发环境模拟支付，确认完成本次付款？',
        success: (r) => {
          if (!r.confirm) {
            reject(new Error('已取消支付'))
            return
          }
          payApi
            .mockPay(result.payment_no)
            .then(() => {
              uni.showToast({ title: '支付成功', icon: 'success' })
              resolve()
            })
            .catch((e: unknown) => reject(new Error(errMsg(e)))
          )
        },
        fail: () => reject(new Error('已取消支付'))
      })
    })
  }
  const p = result.params || {}
  return new Promise((resolve, reject) => {
    uni.requestPayment({
      provider: 'wxpay',
      timeStamp: p.time_stamp ?? '',
      nonceStr: p.nonce_str ?? '',
      package: p.package ?? '',
      signType: 'RSA',
      paySign: p.pay_sign ?? '',
      success: () => {
        uni.showToast({ title: '支付成功', icon: 'success' })
        resolve()
      },
      fail: (err: { errMsg?: string }) => reject(new Error(err.errMsg || '支付失败')),
    } as unknown as UniApp.RequestPaymentOptions)
  })
}
