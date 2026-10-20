<script setup lang="ts">
import { onShow } from '@dcloudio/uni-app'
import { computed, ref } from 'vue'
import { errMsg, foodApi, toast } from '@/utils/api'
import { fen2yuan } from '@/utils/format'
import { useCart } from '@/stores/cart'
import { handlePayResult } from '@/utils/pay'
import type { PayResult } from '@xc/shared/types'

const cart = useCart()

const fulfill = ref<1 | 2>(1) // 1 自取 2 送货
const address = ref('')
const date = ref('')
const time = ref('')
const remark = ref('')
const submitting = ref(false)

function pad2(n: number): string {
  return n < 10 ? `0${n}` : `${n}`
}

const dateStart = computed(() => {
  const n = new Date()
  return `${n.getFullYear()}-${pad2(n.getMonth() + 1)}-${pad2(n.getDate())}`
})

onShow(() => {
  if (cart.state.items.length === 0) {
    toast('采购单为空')
    setTimeout(() => uni.navigateBack(), 600)
  }
})

function onDateChange(e: { detail: { value: string } }) {
  date.value = e.detail.value
}

function onTimeChange(e: { detail: { value: string } }) {
  time.value = e.detail.value
}

function onMinus(id: number) {
  const it = cart.state.items.find((i) => i.ingredient_id === id)
  if (!it) return
  if (it.quantity <= 1) cart.remove(id)
  else cart.setQty(id, it.quantity - 1)
}

function onPlus(id: number) {
  const it = cart.state.items.find((i) => i.ingredient_id === id)
  if (it) cart.setQty(id, it.quantity + 1)
}

async function submit() {
  if (submitting.value) return
  if (cart.state.items.length === 0) {
    toast('采购单为空')
    return
  }
  let deliverTime = ''
  if (fulfill.value === 2) {
    if (!address.value.trim()) {
      toast('送货需填写收货地址')
      return
    }
    if (!date.value || !time.value) {
      toast('请选择送达时间')
      return
    }
    const d = date.value.split('-').map(Number)
    const t = time.value.split(':').map(Number)
    const want = new Date(d[0], d[1] - 1, d[2], t[0], t[1])
    if (want.getTime() <= Date.now()) {
      toast('送达时间必须晚于当前时间')
      return
    }
    deliverTime = want.toISOString()
  }
  submitting.value = true
  try {
    const res = await foodApi.createPurchase({
      supplier_id: cart.state.supplierId,
      items: cart.state.items.map((i) => ({ ingredient_id: i.ingredient_id, quantity: i.quantity })),
      fulfill_method: fulfill.value,
      deliver_address: fulfill.value === 2 ? address.value.trim() : undefined,
      deliver_time: deliverTime || undefined,
      remark: remark.value.trim() || undefined,
    })
    const payRes: PayResult = { payment_no: res.payment_no, params: res.params as PayResult['params'] }
    try {
      await handlePayResult(payRes)
      cart.clear()
      uni.redirectTo({ url: `/pages/purchases/detail?id=${res.purchase_id}` })
    } catch {
      // 已创建未支付：仍进入详情（详情页可取消后重新下单）
      cart.clear()
      toast('订单已创建但未支付，可在详情页取消后重新下单')
      setTimeout(() => uni.redirectTo({ url: `/pages/purchases/detail?id=${res.purchase_id}` }), 800)
    }
  } catch (e) {
    toast(errMsg(e))
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <view class="page">
    <view class="card">
      <view class="card-title">食材清单</view>
      <view v-for="it in cart.state.items" :key="it.ingredient_id" class="item-row">
        <view class="i-main">
          <text class="i-name">{{ it.name }}</text>
          <text class="i-price">¥{{ fen2yuan(it.price) }}/{{ it.unit }}</text>
        </view>
        <view class="stepper">
          <view class="step-btn" @tap="onMinus(it.ingredient_id)">−</view>
          <text class="step-num">{{ it.quantity }}</text>
          <view class="step-btn" @tap="onPlus(it.ingredient_id)">+</view>
        </view>
      </view>
      <view class="sum-row">
        共 {{ cart.count.value }} 件 · 合计
        <text class="sum-num">¥{{ fen2yuan(cart.totalFen.value) }}</text>
      </view>
    </view>

    <view class="card">
      <view class="card-title">配送方式</view>
      <view class="method-chips">
        <view class="m-chip" :class="{ on: fulfill === 1 }" @tap="fulfill = 1">自取（到店提货）</view>
        <view class="m-chip" :class="{ on: fulfill === 2 }" @tap="fulfill = 2">送货上门</view>
      </view>
      <view v-if="fulfill === 2" class="deliver-form">
        <view class="form-item">
          <text class="f-label">收货地址</text>
          <textarea
            v-model="address"
            class="f-textarea"
            placeholder="请填写村组/门牌等详细地址"
            maxlength="140"
          />
        </view>
        <view class="form-item">
          <text class="f-label">送达时间</text>
          <view class="pickers">
            <picker mode="date" :start="dateStart" :value="date" @change="onDateChange">
              <view class="pick-cell" :class="{ ghost: !date }">{{ date || '选择日期' }}</view>
            </picker>
            <picker mode="time" :value="time" @change="onTimeChange">
              <view class="pick-cell" :class="{ ghost: !time }">{{ time || '选择时间' }}</view>
            </picker>
          </view>
        </view>
      </view>
      <view v-else class="pickup-hint">供应商接单后请到其门店自提，地址见采购单详情</view>
      <view class="form-item">
        <text class="f-label">备注</text>
        <textarea
          v-model="remark"
          class="f-textarea"
          placeholder="选填，给供应商的留言（如加工要求）"
          maxlength="255"
        />
      </view>
    </view>

    <view class="footer">
      <view class="foot-amount">
        合计 <text class="fa-num">¥{{ fen2yuan(cart.totalFen.value) }}</text>
      </view>
      <button class="submit-btn" :disabled="submitting" @tap="submit">提交并支付</button>
    </view>
    <view class="bottom-space" />
  </view>
</template>

<style scoped>
.card {
  margin: 20rpx 24rpx 0;
  padding: 24rpx;
}

.card-title {
  font-size: 30rpx;
  font-weight: 700;
  color: #262626;
  margin-bottom: 20rpx;
}

.item-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16rpx 0;
  border-bottom: 2rpx solid #f5f5f5;
}

.i-main {
  display: flex;
  flex-direction: column;
  gap: 6rpx;
}

.i-name {
  font-size: 28rpx;
  color: #262626;
}

.i-price {
  font-size: 24rpx;
  color: #d97706;
}

.stepper {
  display: flex;
  align-items: center;
}

.step-btn {
  width: 48rpx;
  height: 48rpx;
  border-radius: 8rpx;
  background: #f5f6f8;
  color: #262626;
  font-size: 30rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}

.step-num {
  min-width: 64rpx;
  text-align: center;
  font-size: 28rpx;
  color: #262626;
}

.sum-row {
  padding-top: 20rpx;
  text-align: right;
  font-size: 26rpx;
  color: #666666;
}

.sum-num {
  color: #d97706;
  font-size: 32rpx;
  font-weight: 700;
}

.method-chips {
  display: flex;
  gap: 20rpx;
}

.m-chip {
  flex: 1;
  height: 76rpx;
  border-radius: 12rpx;
  border: 2rpx solid #e5e5e5;
  background: #ffffff;
  color: #666666;
  font-size: 26rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}

.m-chip.on {
  border-color: #d97706;
  background: #fdf3e3;
  color: #d97706;
  font-weight: 600;
}

.deliver-form {
  margin-top: 24rpx;
}

.pickup-hint {
  margin-top: 20rpx;
  font-size: 24rpx;
  color: #999999;
}

.form-item {
  margin-top: 24rpx;
}

.f-label {
  display: block;
  font-size: 26rpx;
  color: #666666;
  margin-bottom: 12rpx;
}

.f-textarea {
  box-sizing: border-box;
  width: 100%;
  height: 140rpx;
  background: #f5f6f8;
  border-radius: 12rpx;
  padding: 16rpx 20rpx;
  font-size: 26rpx;
}

.pickers {
  display: flex;
  gap: 16rpx;
}

.pickers > picker {
  flex: 1;
}

.pick-cell {
  height: 76rpx;
  border-radius: 12rpx;
  background: #f5f6f8;
  color: #262626;
  font-size: 26rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}

.pick-cell.ghost {
  color: #bbbbbb;
}

.footer {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 99;
  display: flex;
  align-items: center;
  background: #ffffff;
  border-top: 2rpx solid #eeeeee;
  padding: 16rpx 24rpx;
  padding-bottom: calc(16rpx + env(safe-area-inset-bottom));
}

.foot-amount {
  flex: 1;
  font-size: 26rpx;
  color: #666666;
}

.fa-num {
  font-size: 38rpx;
  font-weight: 700;
  color: #d97706;
}

.submit-btn {
  margin: 0;
  padding: 0 56rpx;
  height: 80rpx;
  line-height: 80rpx;
  border-radius: 40rpx;
  background: #d97706;
  color: #ffffff;
  font-size: 30rpx;
  font-weight: 600;
}

.bottom-space {
  height: 160rpx;
}
</style>
