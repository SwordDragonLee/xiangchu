<template>
  <view class="order-list">
    <!-- 状态 tab -->
    <view class="tabs">
      <view
        v-for="t in tabs"
        :key="t.key"
        class="tab"
        :class="{ active: current === t.key }"
        @click="switchTab(t.key)"
      >
        <text>{{ t.label }}</text>
        <view v-if="current === t.key" class="tab-line"></view>
      </view>
    </view>

    <view class="list page-pad">
      <view v-for="o in list" :key="o.id" class="card order-card" @click="goDetail(o.id)">
        <view class="row-between">
          <text class="order-no">单号 {{ o.order_no }}</text>
          <text class="tag" :class="statusInfo(o.status).cls">{{ statusInfo(o.status).text }}</text>
        </view>
        <view class="order-body">
          <view class="strong">{{ o.banquet_type }} · {{ o.table_count }}桌</view>
          <view class="muted">{{ fmtDay(o.banquet_time) }} 开席</view>
          <view class="muted">{{ o.region }} {{ o.address }}</view>
        </view>
        <view class="row-between">
          <text class="muted">已付 ¥{{ yuan(o.paid_amount) }} / 应付 ¥{{ yuan(o.amount) }}</text>
          <text v-if="o.refund_status" class="tag tag-red">{{ refundText(o.refund_status) }}</text>
        </view>
      </view>

      <view v-if="!list.length && !loading" class="empty">暂无相关订单</view>
      <view v-if="loading" class="empty">加载中...</view>
      <view v-else-if="finished && list.length" class="empty load-tip">没有更多了</view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onReachBottom, onShow } from '@dcloudio/uni-app'
import type { Order } from '@xc/shared/types'
import { orderApi } from '@/utils/api'
import { ORDER_STATUS, REFUND_STATUS, errMsg, fmtDay, fmtTime, yuan } from '@/utils/format'
import { useSession } from '@/stores/session'

type TabKey = 'all' | 'unpaid' | 'pending' | 'ongoing' | 'done'

const tabs: { key: TabKey; label: string; statuses: number[] }[] = [
  { key: 'all', label: '全部', statuses: [] },
  { key: 'unpaid', label: '待支付', statuses: [0] },
  { key: 'pending', label: '待接单', statuses: [1] },
  { key: 'ongoing', label: '进行中', statuses: [2, 3, 4] },
  { key: 'done', label: '已完成', statuses: [5] },
]

const current = ref<TabKey>('all')
const list = ref<Order[]>([])
const page = ref(1)
const size = 10
const loading = ref(false)
const finished = ref(false)

const { isLogged } = useSession()

function statusInfo(status: number) {
  return ORDER_STATUS[status] ?? { text: `状态${status}`, cls: 'tag-grey' }
}
function refundText(s: number) {
  return REFUND_STATUS[s] || ''
}

// 进行中 tab 含多个状态，合并三次请求结果
function fetchStatuses(pageNo: number, statuses: number[]): Promise<{ items: Order[]; total: number }> {
  if (!statuses.length) {
    return orderApi.list({ page: pageNo, size }).then((res) => ({ items: res.list || [], total: res.total }))
  }
  return Promise.all(statuses.map((s) => orderApi.list({ status: s, page: pageNo, size }))).then((rs) => {
    const items: Order[] = []
    let total = 0
    for (const r of rs) {
      items.push(...(r.list || []))
      total += r.total
    }
    items.sort((a, b) => (a.created_at < b.created_at ? 1 : -1))
    return { items, total }
  })
}

function loadPage() {
  if (loading.value || finished.value) return
  const tab = tabs.find((t) => t.key === current.value)
  if (!tab) return
  loading.value = true
  fetchStatuses(page.value, tab.statuses)
    .then(({ items, total }) => {
      list.value = page.value === 1 ? items : list.value.concat(items)
      finished.value = items.length < size || (total > 0 && list.value.length >= total)
    })
    .catch((e: unknown) => uni.showToast({ title: errMsg(e), icon: 'none' }))
    .finally(() => (loading.value = false))
}

function reload() {
  if (!isLogged.value) {
    list.value = []
    finished.value = true
    return
  }
  page.value = 1
  finished.value = false
  loadPage()
}

function switchTab(key: TabKey) {
  if (current.value === key) return
  current.value = key
  reload()
}

onShow(() => reload())

onReachBottom(() => {
  if (!finished.value && !loading.value) {
    page.value += 1
    loadPage()
  }
})

function goDetail(id: number) {
  uni.navigateTo({ url: `/pages/order/detail?id=${id}` })
}
</script>

<style scoped>
.tabs {
  display: flex;
  background: #fff;
  position: sticky;
  top: 0;
  z-index: 5;
}
.tab {
  flex: 1;
  text-align: center;
  padding: 22rpx 0;
  font-size: 28rpx;
  color: #666;
  position: relative;
}
.tab.active {
  color: #d97706;
  font-weight: 700;
}
.tab-line {
  position: absolute;
  left: 50%;
  transform: translateX(-50%);
  bottom: 0;
  width: 48rpx;
  height: 6rpx;
  border-radius: 4rpx;
  background: #d97706;
}
.order-no {
  color: #999;
  font-size: 22rpx;
}
.order-body {
  padding: 12rpx 0;
}
.load-tip {
  padding: 24rpx 0;
}
</style>
