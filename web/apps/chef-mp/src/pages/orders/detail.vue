<script setup lang="ts">
import { onLoad } from '@dcloudio/uni-app'
import { computed, ref } from 'vue'
import { chefOrderApi, errMsg, toast } from '@/utils/api'
import { ORDER_STATUS_TEXT, fen2yuan } from '@/utils/format'
import type { Order } from '@xc/shared'

const order = ref<Order | null>(null)
const acting = ref(false)
const orderId = ref(0)

onLoad((q) => {
  const id = Number((q as Record<string, string> | undefined)?.id || 0)
  if (!id) {
    toast('缺少订单参数')
    return
  }
  orderId.value = id
  load()
})

async function load() {
  try {
    order.value = await chefOrderApi.detail(orderId.value)
  } catch (e) {
    toast(errMsg(e))
  }
}

const statusText = computed(() => (order.value ? ORDER_STATUS_TEXT[order.value.status] || '未知' : ''))

async function confirmAction(title: string, fn: () => Promise<unknown>) {
  uni.showModal({
    title: '确认操作',
    content: title,
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
  confirmAction('确认接受该订单？接单后请按时备菜开席。', () => chefOrderApi.accept(orderId.value))
}

function startPrep() {
  confirmAction('确认开始备菜？', () => chefOrderApi.startPrep(orderId.value))
}

function startService() {
  confirmAction('确认开席服务？', () => chefOrderApi.startService(orderId.value))
}

function complete() {
  confirmAction('确认该订单服务已完成？完成后订单不可撤销。', () => chefOrderApi.complete(orderId.value))
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
        await chefOrderApi.reject(orderId.value, reason)
        toast('已拒单')
        await load()
      } catch (e) {
        toast(errMsg(e))
      } finally {
        acting.value = false
      }
    },
  })
}

function callContact() {
  if (!order.value?.contact_phone) return
  uni.makePhoneCall({ phoneNumber: order.value.contact_phone })
}

function fmtTime(s: string): string {
  return (s || '').replace('T', ' ').slice(0, 16)
}
</script>

<template>
  <view class="page">
    <view v-if="order" class="card">
      <view class="status-row">
        <text class="status-main" :class="'st-' + order.status">{{ statusText }}</text>
        <text class="order-no">{{ order.order_no }}</text>
      </view>

      <view class="info-block">
        <view class="info-line">
          <text class="info-label">宴席类型</text>
          <text class="info-value">{{ order.banquet_type }}</text>
        </view>
        <view class="info-line">
          <text class="info-label">开席时间</text>
          <text class="info-value">{{ fmtTime(order.banquet_time) }}</text>
        </view>
        <view class="info-line">
          <text class="info-label">桌数</text>
          <text class="info-value">{{ order.table_count }} 桌 / {{ order.people_count || '-' }} 人</text>
        </view>
        <view class="info-line">
          <text class="info-label">地点</text>
          <text class="info-value">{{ order.region }} {{ order.address }}</text>
        </view>
        <view class="info-line" @tap="callContact">
          <text class="info-label">联系人</text>
          <text class="info-value link">{{ order.contact_name }} {{ order.contact_phone }}（点击拨打）</text>
        </view>
        <view class="info-line">
          <text class="info-label">订单金额</text>
          <text class="info-value strong">¥{{ fen2yuan(order.amount) }}</text>
        </view>
        <view v-if="order.paid_amount > 0" class="info-line">
          <text class="info-label">已支付</text>
          <text class="info-value">¥{{ fen2yuan(order.paid_amount) }}</text>
        </view>
        <view class="info-line">
          <text class="info-label">下单时间</text>
          <text class="info-value">{{ fmtTime(order.created_at) }}</text>
        </view>
      </view>

      <view v-if="order.menu_remark" class="remark">
        <text class="remark-title">菜单备注</text>
        <text class="remark-content">{{ order.menu_remark }}</text>
      </view>

      <view v-if="order.cancel_reason" class="remark">
        <text class="remark-title">取消/拒绝原因</text>
        <text class="remark-content">{{ order.cancel_reason }}</text>
      </view>
    </view>

    <!-- 状态操作按钮 -->
    <view v-if="order" class="footer">
      <template v-if="order.status === 1">
        <button class="footer-btn plain" :disabled="acting" @tap="reject">拒单</button>
        <button class="footer-btn primary" :disabled="acting" @tap="accept">接单</button>
      </template>
      <button v-else-if="order.status === 2" class="footer-btn primary full" :disabled="acting" @tap="startPrep">
        开始备菜
      </button>
      <button v-else-if="order.status === 3" class="footer-btn primary full" :disabled="acting" @tap="startService">
        开席服务
      </button>
      <button v-else-if="order.status === 4" class="footer-btn primary full" :disabled="acting" @tap="complete">
        完成订单
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

.st-1 { color: #e6504f; }
.st-2 { color: #d97706; }
.st-3 { color: #2563eb; }
.st-4 { color: #16a34a; }
.st-5 { color: #999999; }

.order-no {
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
  color: #d97706;
}

.info-value.link {
  color: #d97706;
}

.remark {
  background: #f7f7f8;
  border-radius: 12rpx;
  padding: 20rpx;
  margin-top: 10rpx;
}

.remark-title {
  display: block;
  font-size: 24rpx;
  color: #999999;
  margin-bottom: 8rpx;
}

.remark-content {
  font-size: 26rpx;
  color: #262626;
  line-height: 1.6;
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
  background: #d97706;
  color: #ffffff;
}

.footer-btn.plain {
  background: #ffffff;
  color: #e6504f;
  border: 2rpx solid #e6504f;
  margin-right: 24rpx;
}

.footer-btn.full {
  background: #d97706;
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
