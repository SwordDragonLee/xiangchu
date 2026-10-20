<script setup lang="ts">
import { onLoad } from '@dcloudio/uni-app'
import { computed, ref } from 'vue'
import { errMsg, supplierApi, toast } from '@/utils/api'
import { FULFILL_METHOD_TEXT, PURCHASE_STATUS_TEXT, fen2yuan } from '@/utils/format'
import type { PurchaseOrder } from '@xc/shared'

const po = ref<PurchaseOrder | null>(null)
const acting = ref(false)
const poId = ref(0)

onLoad((q) => {
  const id = Number((q as Record<string, string> | undefined)?.id || 0)
  if (!id) {
    toast('缺少采购单参数')
    return
  }
  poId.value = id
  load()
})

async function load() {
  try {
    po.value = await supplierApi.purchase(poId.value)
  } catch (e) {
    toast(errMsg(e))
  }
}

const statusText = computed(() =>
  po.value ? PURCHASE_STATUS_TEXT[po.value.status] || '未知' : '',
)

async function confirmAction(content: string, fn: () => Promise<unknown>) {
  uni.showModal({
    title: '确认操作',
    content,
    success: async (res) => {
      if (!res.confirm) return
      acting.value = true
      try {
        await fn()
        toast('操作成功')
        await load()
      } catch (e) {
        toast(errMsg(e))
      } finally {
        acting.value = false
      }
    },
  })
}

function accept() {
  confirmAction('确认接单？接单后将锁定对应库存，请按时备货。', () =>
    supplierApi.acceptPurchase(poId.value),
  )
}

function ready() {
  confirmAction('确认备货完成？完成后等待乡厨自取/确认送达。', () =>
    supplierApi.readyPurchase(poId.value),
  )
}

function complete() {
  confirmAction('确认买家已提货/收货？完成后订单进入结算，平台收取 1% 佣金。', () =>
    supplierApi.completePurchase(poId.value),
  )
}

function reject() {
  uni.showModal({
    title: '拒单',
    editable: true,
    placeholderText: '请填写拒单原因（必填）',
    success: async (res) => {
      if (!res.confirm) return
      const reason = (res.content || '').trim()
      if (!reason) {
        toast('请填写拒单原因')
        return
      }
      acting.value = true
      try {
        await supplierApi.rejectPurchase(poId.value, reason)
        toast('已拒单，货款将原路退回')
        await load()
      } catch (e) {
        toast(errMsg(e))
      } finally {
        acting.value = false
      }
    },
  })
}

function fmtTime(s: string): string {
  return (s || '').replace('T', ' ').slice(0, 16)
}
</script>

<template>
  <view class="page">
    <view v-if="po" class="card">
      <view class="status-row">
        <text class="status-main" :class="'st-' + po.status">{{ statusText }}</text>
        <text class="po-no">{{ po.purchase_no }}</text>
      </view>

      <view class="info-block">
        <view class="info-line">
          <text class="info-label">订单金额</text>
          <text class="info-value strong">¥{{ fen2yuan(po.total_amount) }}</text>
        </view>
        <view class="info-line">
          <text class="info-label">实付金额</text>
          <text class="info-value">¥{{ fen2yuan(po.paid_amount) }}</text>
        </view>
        <view class="info-line">
          <text class="info-label">履约方式</text>
          <text class="info-value">{{ FULFILL_METHOD_TEXT[po.fulfill_method] || '-' }}</text>
        </view>
        <view v-if="po.fulfill_method === 1" class="info-line">
          <text class="info-label">自取地址</text>
          <text class="info-value">{{ po.pickup_address || '-' }}</text>
        </view>
        <view v-if="po.fulfill_method === 2" class="info-line">
          <text class="info-label">送货地址</text>
          <text class="info-value">{{ po.deliver_address || '-' }}</text>
        </view>
        <view v-if="po.fulfill_method === 2 && po.deliver_time" class="info-line">
          <text class="info-label">期望送达</text>
          <text class="info-value">{{ fmtTime(po.deliver_time) }}</text>
        </view>
        <view class="info-line">
          <text class="info-label">下单时间</text>
          <text class="info-value">{{ fmtTime(po.created_at) }}</text>
        </view>
      </view>
    </view>

    <!-- 商品明细 -->
    <view v-if="po && po.items && po.items.length" class="card">
      <view class="card-title">商品明细</view>
      <view v-for="it in po.items" :key="it.id" class="item-row">
        <view class="item-main">
          <text class="item-name">{{ it.name }}</text>
          <text class="item-sub">¥{{ fen2yuan(it.price) }}/{{ it.unit }} × {{ it.quantity }}</text>
        </view>
        <text class="item-subtotal">¥{{ fen2yuan(it.subtotal) }}</text>
      </view>
    </view>

    <!-- 状态操作按钮 -->
    <view v-if="po" class="footer">
      <template v-if="po.status === 0">
        <button class="footer-btn plain" :disabled="acting" @tap="reject">拒单</button>
        <button class="footer-btn primary" :disabled="acting" @tap="accept">接单</button>
      </template>
      <button v-else-if="po.status === 1" class="footer-btn primary full" :disabled="acting" @tap="ready">
        备货完成
      </button>
      <button v-else-if="po.status === 2" class="footer-btn primary full" :disabled="acting" @tap="complete">
        确认完成
      </button>
      <view v-else class="no-action">当前状态暂无可执行操作</view>
    </view>

    <view style="height: 180rpx" />
  </view>
</template>

<style scoped>
.status-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: 20rpx;
  border-bottom: 2rpx solid #f0f0f0;
  margin-bottom: 20rpx;
}

.status-main {
  font-size: 34rpx;
  font-weight: 700;
}

.st-0 { color: #e6504f; }
.st-1 { color: #d97706; }
.st-2 { color: #2563eb; }
.st-3 { color: #16a34a; }
.st-4 { color: #999999; }

.po-no {
  font-size: 24rpx;
  color: #999999;
}

.info-line {
  display: flex;
  margin-bottom: 18rpx;
}

.info-label {
  width: 160rpx;
  color: #999999;
  font-size: 26rpx;
  flex-shrink: 0;
}

.info-value {
  flex: 1;
  color: #262626;
  font-size: 28rpx;
}

.info-value.strong {
  font-weight: 700;
  color: #16a34a;
}

.item-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 18rpx 0;
  border-bottom: 2rpx solid #f5f5f5;
}

.item-row:last-child {
  border-bottom: none;
}

.item-main {
  display: flex;
  flex-direction: column;
}

.item-name {
  font-size: 28rpx;
  color: #262626;
}

.item-sub {
  font-size: 24rpx;
  color: #999999;
  margin-top: 6rpx;
}

.item-subtotal {
  font-size: 28rpx;
  font-weight: 600;
  color: #262626;
}

.footer {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  background: #ffffff;
  display: flex;
  padding: 20rpx 24rpx calc(20rpx + env(safe-area-inset-bottom));
  border-top: 2rpx solid #f0f0f0;
}

.footer-btn {
  flex: 1;
  height: 84rpx;
  line-height: 84rpx;
  border-radius: 44rpx;
  font-size: 30rpx;
  margin: 0;
}

.footer-btn.primary {
  background: #16a34a;
  color: #ffffff;
}

.footer-btn.plain {
  background: #ffffff;
  color: #e6504f;
  border: 2rpx solid #e6504f;
  margin-right: 24rpx;
}

.footer-btn.full {
  background: #16a34a;
  color: #ffffff;
}

.footer-btn[disabled] {
  opacity: 0.6;
}

.no-action {
  flex: 1;
  text-align: center;
  color: #bbbbbb;
  font-size: 26rpx;
  line-height: 84rpx;
}
</style>
