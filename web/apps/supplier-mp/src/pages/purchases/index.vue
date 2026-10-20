<script setup lang="ts">
import { onPullDownRefresh, onReachBottom, onShow } from '@dcloudio/uni-app'
import { computed, ref } from 'vue'
import TabBar from '@/components/TabBar.vue'
import { ensureLogin, ensureSupplier, errMsg, supplierApi, toast } from '@/utils/api'
import { FULFILL_METHOD_TEXT, PURCHASE_STATUS_TEXT, fen2yuan } from '@/utils/format'
import type { PurchaseOrder } from '@xc/shared'

const tabs = [
  { status: -1, text: '全部' },
  { status: 0, text: '待接单' },
  { status: 1, text: '备货中' },
  { status: 2, text: '待完成' },
  { status: 3, text: '已完成' },
  { status: 4, text: '已取消' },
]

const active = ref(-1)
const list = ref<PurchaseOrder[]>([])
const page = ref(1)
const size = 10
const total = ref(0)
const loading = ref(false)
const finished = computed(() => total.value > 0 && list.value.length >= total.value)

onShow(async () => {
  if (!ensureLogin()) return
  if (!(await ensureSupplier())) return
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
    const res = await supplierApi.purchases({
      status: active.value >= 0 ? active.value : undefined,
      page: p,
      size,
    })
    page.value = p
    total.value = res.total
    list.value = p === 1 ? res.list : [...list.value, ...res.list]
  } catch (e) {
    toast(errMsg(e))
  } finally {
    loading.value = false
  }
}

function goDetail(o: PurchaseOrder) {
  uni.navigateTo({ url: `/pages/purchases/detail?id=${o.id}` })
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

    <view v-if="list.length === 0 && !loading" class="empty">暂无相关采购单</view>

    <view v-for="o in list" :key="o.id" class="card po-card" @tap="goDetail(o)">
      <view class="row1">
        <text class="po-no">{{ o.purchase_no }}</text>
        <text class="po-status" :class="'st-' + o.status">{{ PURCHASE_STATUS_TEXT[o.status] }}</text>
      </view>
      <view class="row-line">
        <text class="label">{{ FULFILL_METHOD_TEXT[o.fulfill_method] || '-' }}</text>
        <text class="value">¥{{ fen2yuan(o.total_amount) }}</text>
      </view>
      <view class="row-line">
        <text class="label">{{ fmtTime(o.created_at) }}</text>
        <text class="link">查看详情</text>
      </view>
    </view>

    <view v-if="list.length > 0" class="load-more">
      {{ loading ? '加载中...' : finished ? '没有更多了' : '上拉加载更多' }}
    </view>

    <view class="bottom-space" />
    <TabBar current="purchases" />
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
  font-size: 26rpx;
  color: #595959;
  padding: 24rpx 0;
  position: relative;
}

.tab.on {
  color: #16a34a;
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
  background: #16a34a;
}

.po-card {
  margin-top: 20rpx;
}

.row1 {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16rpx;
}

.po-no {
  font-size: 26rpx;
  color: #999999;
}

.po-status {
  font-size: 26rpx;
  font-weight: 600;
}

.st-0 { color: #e6504f; }
.st-1 { color: #d97706; }
.st-2 { color: #2563eb; }
.st-3 { color: #16a34a; }
.st-4 { color: #999999; }

.row-line {
  display: flex;
  justify-content: space-between;
  font-size: 28rpx;
  margin-bottom: 10rpx;
}

.label {
  color: #262626;
}

.value {
  color: #16a34a;
  font-weight: 600;
}

.link {
  color: #16a34a;
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
