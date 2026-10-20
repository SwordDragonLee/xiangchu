<script setup lang="ts">
import { onPullDownRefresh, onReachBottom, onShow } from '@dcloudio/uni-app'
import { computed, ref } from 'vue'
import TabBar from '@/components/TabBar.vue'
import { ensureLogin, errMsg, foodApi, toast } from '@/utils/api'
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

onShow(() => {
  if (ensureLogin()) reload()
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
    const res = await foodApi.myPurchases(p, size, active.value >= 0 ? active.value : undefined)
    page.value = p
    total.value = res.total
    list.value = p === 1 ? res.list : [...list.value, ...res.list]
  } catch (e) {
    toast(errMsg(e))
  } finally {
    loading.value = false
  }
}

function itemsSummary(o: PurchaseOrder): string {
  const its = o.items || []
  if (its.length === 0) return '—'
  return its.length === 1 ? its[0].name : `${its[0].name} 等 ${its.length} 种`
}

function isUnpaid(o: PurchaseOrder): boolean {
  return o.paid_amount === 0 && o.status !== 4
}

function goDetail(p: PurchaseOrder) {
  uni.navigateTo({ url: `/pages/purchases/detail?id=${p.id}` })
}

function fmtTime(s: string): string {
  return (s || '').replace('T', ' ').slice(0, 16)
}
</script>

<template>
  <view class="page">
    <view class="p-tabs">
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

    <view v-if="list.length === 0 && !loading" class="empty">暂无采购单</view>

    <view v-for="o in list" :key="o.id" class="card p-card" @tap="goDetail(o)">
      <view class="row1">
        <text class="p-no">{{ o.purchase_no }}</text>
        <text class="p-status" :class="'st-' + o.status">{{ PURCHASE_STATUS_TEXT[o.status] }}</text>
      </view>
      <view class="row2">
        <text class="s-name">{{ o.supplier_name || '供应商' }}</text>
        <text class="s-amount">¥{{ fen2yuan(o.total_amount) }}</text>
      </view>
      <view class="row-line">
        <text class="items-txt">{{ itemsSummary(o) }}</text>
        <view class="tags">
          <text class="tag fulfill">{{ FULFILL_METHOD_TEXT[o.fulfill_method] || '-' }}</text>
          <text v-if="isUnpaid(o)" class="tag unpaid">未支付</text>
          <text v-if="o.status === 0" class="tag warn">待供应商接单</text>
        </view>
      </view>
      <view class="row-time">
        <text class="time">{{ fmtTime(o.created_at) }}</text>
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
.p-tabs {
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

.p-card {
  margin-top: 20rpx;
}

.row1 {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.p-no {
  font-size: 26rpx;
  color: #999999;
}

.p-status {
  font-size: 26rpx;
  font-weight: 600;
}

.st-0 { color: #d97706; }
.st-1 { color: #2563eb; }
.st-2 { color: #16a34a; }
.st-3 { color: #999999; }
.st-4 { color: #bbbbbb; }

.row2 {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 12rpx;
}

.s-name {
  font-size: 30rpx;
  font-weight: 600;
  color: #262626;
}

.s-amount {
  font-size: 30rpx;
  font-weight: 700;
  color: #d97706;
}

.row-line {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 14rpx;
}

.items-txt {
  font-size: 26rpx;
  color: #595959;
}

.row-time {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 14rpx;
  padding-top: 14rpx;
  border-top: 2rpx solid #f5f5f5;
}

.time {
  font-size: 24rpx;
  color: #bbbbbb;
}

.link {
  font-size: 24rpx;
  color: #d97706;
}

.tags {
  display: flex;
  gap: 12rpx;
}

.tag {
  font-size: 20rpx;
  padding: 4rpx 12rpx;
  border-radius: 8rpx;
}

.tag.fulfill {
  background: #f0f0f0;
  color: #595959;
}

.tag.unpaid {
  background: #fdecec;
  color: #e6504f;
}

.tag.warn {
  background: #fdf3e3;
  color: #d97706;
}

.load-more {
  text-align: center;
  color: #bbbbbb;
  font-size: 24rpx;
  padding: 30rpx 0;
}

.empty {
  text-align: center;
  color: #bbbbbb;
  font-size: 26rpx;
  padding: 120rpx 0;
}

.bottom-space {
  height: 140rpx;
}
</style>
