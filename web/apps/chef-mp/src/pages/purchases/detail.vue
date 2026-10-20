<script setup lang="ts">
import { onShow } from '@dcloudio/uni-app'
import { ref } from 'vue'
import { errMsg, foodApi, toast } from '@/utils/api'
import { PURCHASE_STATUS_TEXT, fen2yuan } from '@/utils/format'
import type { PurchaseOrder } from '@xc/shared'

const pid = ref(0)
const detail = ref<PurchaseOrder | null>(null)

onShow(() => {
  const pages = getCurrentPages()
  const cur = pages[pages.length - 1] as unknown as { options?: Record<string, string> }
  const id = Number(cur?.options?.id || 0)
  if (id) {
    pid.value = id
    load()
  }
})

async function load() {
  try {
    detail.value = await foodApi.purchase(pid.value)
  } catch (e) {
    toast(errMsg(e))
  }
}

function onCancel() {
  const d = detail.value
  if (!d) return
  const paid = (d.paid_amount || 0) > 0
  uni.showModal({
    title: '取消采购单',
    content: paid
      ? '已支付金额将原路退回，确认取消？'
      : '未支付，取消后如需采购请重新下单，确认取消？',
    success: (r) => {
      if (!r.confirm) return
      foodApi
        .cancelPurchase(pid.value)
        .then(() => {
          toast('已取消')
          load()
        })
        .catch((e: unknown) => toast(errMsg(e)))
    },
  })
}

function fmtTime(s: string): string {
  return (s || '').replace('T', ' ').slice(0, 16)
}

function isUnpaid(d: PurchaseOrder): boolean {
  return (d.paid_amount || 0) === 0 && d.status !== 4
}
</script>

<template>
  <view v-if="detail" class="page">
    <view class="status-head" :class="'h-' + detail.status">
      <text class="sh-text">{{ PURCHASE_STATUS_TEXT[detail.status] }}</text>
      <text class="sh-no">{{ detail.purchase_no }}</text>
    </view>

    <view class="card">
      <view class="c-title">供应商</view>
      <view class="s-row">
        <text class="s-name">{{ detail.supplier_name || '供应商' }}</text>
        <text class="s-hint">建议提前电话沟通取货/送达事宜</text>
      </view>
      <view class="addr-block" v-if="detail.status !== 4">
        <template v-if="detail.fulfill_method === 1">
          <view class="a-label">自取地址</view>
          <view class="a-text">{{ detail.pickup_address || '供应商接单后提供' }}</view>
        </template>
        <template v-else>
          <view class="a-label">收货地址</view>
          <view class="a-text">{{ detail.deliver_address }}</view>
          <view class="a-label mt">送达时间</view>
          <view class="a-text">{{ fmtTime(detail.deliver_time) }}</view>
        </template>
      </view>
    </view>

    <view class="card">
      <view class="c-title">食材明细</view>
      <view v-for="it in detail.items" :key="it.ingredient_id" class="it-row">
        <view class="it-main">
          <text class="it-name">{{ it.name }}</text>
          <text class="it-price">¥{{ fen2yuan(it.price) }}/{{ it.unit }} × {{ it.quantity }}</text>
          <text class="it-sub">小计 ¥{{ fen2yuan(it.price * it.quantity) }}</text>
        </view>
      </view>
      <view class="amount-row">
        <text class="a-k">合计</text>
        <text class="a-v strong">¥{{ fen2yuan(detail.total_amount) }}</text>
      </view>
      <view v-if="(detail.paid_amount || 0) > 0" class="amount-row">
        <text class="a-k">已支付</text>
        <text class="a-v paid">¥{{ fen2yuan(detail.paid_amount) }}</text>
      </view>
      <view v-if="isUnpaid(detail)" class="amount-row">
        <text class="a-k">支付状态</text>
        <text class="a-v unpaid">未支付</text>
      </view>
      <view class="amount-row">
        <text class="a-k">下单时间</text>
        <text class="a-v muted">{{ fmtTime(detail.created_at) }}</text>
      </view>
    </view>

    <view v-if="detail.remark" class="card">
      <view class="c-title">备注</view>
      <view class="remark">{{ detail.remark }}</view>
    </view>

    <view class="foot-area">
      <view v-if="detail.status === 0" class="cancel-wrap">
        <view class="cancel-hint">
          {{ (detail.paid_amount || 0) > 0 ? '取消后已支付金额将原路退回' : '支付窗口已关闭（暂不支持重新支付），如需修改可直接取消后重新下单' }}
        </view>
        <button class="cancel-btn" @tap="onCancel">取消采购单</button>
      </view>
      <view v-else-if="detail.status === 2" class="status-tip">
        等待自取/送达，完成后由供应商确认
      </view>
      <view v-else class="status-tip">当前状态暂无可执行操作</view>
      <view class="bottom-space" />
    </view>
  </view>
  <view v-else class="page loading-page">加载中...</view>
</template>

<style scoped>
.page {
  padding-bottom: 40rpx;
}

.loading-page {
  text-align: center;
  color: #bbbbbb;
  font-size: 26rpx;
  padding: 120rpx 0;
}

.status-head {
  background: #d97706;
  padding: 44rpx 32rpx;
  display: flex;
  flex-direction: column;
  gap: 8rpx;
}

.status-head.h-3,
.status-head.h-4 {
  background: #9ca3af;
}

.sh-text {
  color: #ffffff;
  font-size: 38rpx;
  font-weight: 700;
}

.sh-no {
  color: rgba(255, 255, 255, 0.85);
  font-size: 24rpx;
}

.card {
  margin: 20rpx 24rpx 0;
  padding: 24rpx;
}

.c-title {
  font-size: 28rpx;
  font-weight: 700;
  color: #262626;
  margin-bottom: 16rpx;
}

.s-row {
  display: flex;
  flex-direction: column;
  gap: 8rpx;
}

.s-name {
  font-size: 32rpx;
  font-weight: 600;
  color: #262626;
}

.s-hint {
  font-size: 24rpx;
  color: #999999;
}

.addr-block {
  margin-top: 20rpx;
  padding: 20rpx;
  background: #fdf3e3;
  border-radius: 12rpx;
}

.a-label {
  font-size: 24rpx;
  color: #d97706;
}

.a-label.mt {
  margin-top: 16rpx;
}

.a-text {
  font-size: 28rpx;
  color: #262626;
  margin-top: 8rpx;
}

.it-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14rpx 0;
  border-bottom: 2rpx solid #f5f5f5;
}

.it-name {
  font-size: 28rpx;
  color: #262626;
}

.it-price {
  display: block;
  font-size: 24rpx;
  color: #999999;
  margin-top: 6rpx;
}

.it-sub {
  display: block;
  font-size: 24rpx;
  color: #d97706;
  margin-top: 6rpx;
}

.amount-row {
  display: flex;
  justify-content: space-between;
  padding-top: 14rpx;
  font-size: 26rpx;
}

.a-k {
  color: #999999;
}

.a-v {
  color: #262626;
}

.a-v.strong {
  font-size: 32rpx;
  font-weight: 700;
  color: #d97706;
}

.a-v.paid {
  color: #16a34a;
}

.a-v.unpaid {
  color: #e6504f;
}

.a-v.muted {
  color: #999999;
}

.remark {
  font-size: 26rpx;
  color: #595959;
  line-height: 1.6;
}

.foot-area {
  padding: 30rpx 24rpx 0;
}

.cancel-wrap {
  display: flex;
  flex-direction: column;
  gap: 16rpx;
}

.cancel-hint {
  font-size: 24rpx;
  color: #999999;
  line-height: 1.6;
}

.cancel-btn {
  margin: 0;
  height: 84rpx;
  line-height: 84rpx;
  border-radius: 42rpx;
  background: #ffffff;
  color: #e6504f;
  border: 2rpx solid #e6504f;
  font-size: 30rpx;
}

.cancel-btn::after {
  border: none;
}

.status-tip {
  text-align: center;
  font-size: 26rpx;
  color: #999999;
  padding: 30rpx 0 0;
}

.bottom-space {
  height: 60rpx;
}
</style>
