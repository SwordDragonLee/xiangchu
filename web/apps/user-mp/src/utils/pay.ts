// 支付收口：mock 支付弹确认框调 mockPay；真实支付走 uni.requestPayment
import type { PayResult } from '@xc/shared/types'
import { payApi } from './api'
import { errMsg } from './format'

export function handlePayResult(result: PayResult): Promise<void> {
  if (result.params && result.params.mock) {
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
            .catch((e: unknown) => reject(new Error(errMsg(e))))
        },
        fail: () => reject(new Error('已取消支付')),
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
