<template>
  <view class="aftersale page-pad">
    <view class="card">
      <view class="card-title">申请退款</view>
      <view class="form-item">
        <view class="form-label">退款金额（元）</view>
        <input v-model="amountText" class="form-input" type="digit" placeholder="请输入退款金额" />
        <text class="muted max-tip">最多 ¥{{ yuan(maxAmount) }}</text>
      </view>
      <view class="form-item">
        <view class="form-label">退款原因</view>
        <input v-model="reason" class="form-input" placeholder="请填写退款原因" />
      </view>
      <view class="muted tip">一期仅支持整单/部分金额退款申请，图片凭证二期开放</view>
    </view>

    <button class="btn-primary" :class="{ disabled: submitting }" @click="submit">提交申请</button>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import type { Order } from '@xc/shared/types'
import { orderApi } from '@/utils/api'
import { errMsg, yuan } from '@/utils/format'

const orderId = ref(0)
const maxAmount = ref(0)
const amountText = ref('')
const reason = ref('')
const submitting = ref(false)

onLoad((query) => {
  orderId.value = Number(query?.order_id || 0)
  if (!orderId.value) {
    uni.showToast({ title: '缺少订单参数', icon: 'none' })
    setTimeout(() => uni.navigateBack(), 800)
    return
  }
  orderApi
    .detail(orderId.value)
    .then((o: Order) => {
      maxAmount.value = Math.max(0, o.paid_amount)
      amountText.value = (maxAmount.value / 100).toFixed(2)
    })
    .catch((e: unknown) => uni.showToast({ title: errMsg(e), icon: 'none' }))
})

function submit() {
  if (submitting.value) return
  const fen = Math.round(parseFloat(amountText.value) * 100)
  if (!fen || fen <= 0) {
    uni.showToast({ title: '请输入正确的退款金额', icon: 'none' })
    return
  }
  if (fen > maxAmount.value) {
    uni.showToast({ title: `退款金额不能超过已付 ¥${yuan(maxAmount.value)}`, icon: 'none' })
    return
  }
  if (!reason.value.trim()) {
    uni.showToast({ title: '请填写退款原因', icon: 'none' })
    return
  }
  submitting.value = true
  orderApi
    .applyAfterSale(orderId.value, { amount: fen, reason: reason.value.trim() })
    .then(() => {
      uni.showToast({ title: '申请已提交', icon: 'success' })
      setTimeout(() => uni.navigateBack(), 600)
    })
    .catch((e: unknown) => uni.showToast({ title: errMsg(e), icon: 'none' }))
    .finally(() => (submitting.value = false))
}
</script>

<style scoped>
.card-title {
  font-size: 30rpx;
  font-weight: 700;
  margin-bottom: 16rpx;
}
.max-tip {
  margin-left: 16rpx;
  flex-shrink: 0;
}
.tip {
  margin-top: 8rpx;
}
</style>
