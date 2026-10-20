<script setup lang="ts">
import { onPullDownRefresh, onReachBottom, onShow } from '@dcloudio/uni-app'
import { computed, ref } from 'vue'
import TabBar from '@/components/TabBar.vue'
import { chefOrderApi, errMsg, toast } from '@/utils/api'
import { ORDER_STATUS_TEXT, fen2yuan } from '@/utils/format'
import type { Order } from '@xc/shared'

const tabs = [
  { status: 1, text: '待接单' },
  { status: 2, text: '已接单' },
  { status: 3, text: '备菜中' },
  { status: 4, text: '服务中' },
  { status: 5, text: '已完成' },
]

const active = ref(1)
const list = ref<Order[]>([])
const page = ref(1)
const size = 10
const total = ref(0)
const loading = ref(false)
const finished = computed(() => total.value > 0 && list.value.length >= total.value)

onShow(() => {
  reload()
})

onPullDownRefresh(async () => {
  await reload()
  uni.stopPullDownRefresh()
})

onReachBottom(() => {
  if (!finished.value && !loading.value) load(page.value + 1)
})

async function switchTab(s: number) {
  if (active.value === s) return
  active.value = s
  await reload()
}

async function reload() {
  list.value = []
  total.value = 0
  page.value = 1
  await load(1)
}

async function load(p: number) {
  loading.value = true
  try {
    const res = await chefOrderApi.list({ status: active.value, page: p, size })
    page.value = p
    total.value = res.total
    list.value = p === 1 ? res.list : [...list.value, ...res.list]
  } catch (e) {
    toast(errMsg(e))
  } finally {
    loading.value = false
  }
}

function goDetail(o: Order) {
  uni.navigateTo({ url: `/pages/orders/detail?id=${o.id}` })
}

function fmtTime(s: string): string {
  return (s || '').replace('T', ' ').slice(0, 16)
}
</script>

<template>
  <view class="page">
    <view class="tabs">
      <view
        v-for="t in tabs"
        :key="t.status"
        class="tab"
        :class="{ on: active === t.status }"
        @tap="switchTab(t.status)"
      >
        {{ t.text }}
      </view>
    </view>

    <view v-if="list.length === 0 && !loading" class="empty">暂无相关订单</view>

    <view v-for="o in list" :key="o.id" class="card order-card" @tap="goDetail(o)">
      <view class="row1">
        <text class="order-no">{{ o.order_no }}</text>
        <text class="order-status" :class="'st-' + o.status">{{ ORDER_STATUS_TEXT[o.status] }}</text>
      </view>
      <view class="row-line">
        <text class="label">{{ o.banquet_type }}</text>
        <text class="value">{{ fmtTime(o.banquet_time) }}</text>
      </view>
      <view class="row-line">
        <text class="label">{{ o.table_count }} 桌</text>
        <text class="value">¥{{ fen2yuan(o.amount) }}</text>
      </view>
      <view class="row-line">
        <text class="label addr">{{ o.region }} {{ o.address }}</text>
      </view>
      <view class="row-line">
        <text class="label">{{ o.contact_name }} {{ o.contact_phone }}</text>
        <text class="link">查看详情</text>
      </view>
    </view>

    <view v-if="list.length > 0" class="load-more">
      {{ loading ? '加载中...' : finished ? '没有更多了' : '上拉加载更多' }}
    </view>

    <view class="bottom-space" />
    <TabBar current="orders" />
  </view>
</template>

<style scoped>
.tabs {
  display: flex;
  background: #ffffff;
  position: sticky;
  top: 0;
  z-index: 10;
  border-bottom: 2rpx solid #f0f0f0;
}

.tab {
  flex: 1;
  text-align: center;
  font-size: 28rpx;
  color: #595959;
  padding: 24rpx 0;
  position: relative;
}

.tab.on {
  color: #d97706;
  font-weight: 600;
}

.tab.on::after {
  content: '';
  position: absolute;
  left: 50%;
  transform: translateX(-50%);
  bottom: 0;
  width: 48rpx;
  height: 6rpx;
  border-radius: 4rpx;
  background: #d97706;
}

.order-card {
  margin-top: 20rpx;
}

.row1 {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16rpx;
}

.order-no {
  font-size: 26rpx;
  color: #999999;
}

.order-status {
  font-size: 26rpx;
  font-weight: 600;
}

.st-1 { color: #e6504f; }
.st-2 { color: #d97706; }
.st-3 { color: #2563eb; }
.st-4 { color: #16a34a; }
.st-5 { color: #999999; }

.row-line {
  display: flex;
  justify-content: space-between;
  font-size: 28rpx;
  margin-bottom: 10rpx;
}

.label {
  color: #262626;
}

.label.addr {
  color: #595959;
  font-size: 26rpx;
}

.value {
  color: #262626;
  font-weight: 600;
}

.link {
  color: #d97706;
  font-size: 26rpx;
}

.load-more {
  text-align: center;
  color: #bbbbbb;
  font-size: 24rpx;
  padding: 30rpx 0;
}

.bottom-space {
  height: 140rpx;
}
</style>
