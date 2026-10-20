<template>
  <view class="chef-detail">
    <template v-if="detail">
      <!-- 档案头部 -->
      <view class="head">
        <view class="head-avatar">{{ chef?.real_name.slice(0, 1) }}</view>
        <view class="grow head-info">
          <view class="head-name">{{ chef?.real_name }}</view>
          <view class="head-meta">
            <text class="head-rating">★ {{ chef?.rating.toFixed(1) }} 分</text>
            <text class="head-orders">已接单 {{ chef?.order_count }} 单</text>
          </view>
          <view class="head-region muted">{{ chef?.region }} · {{ chef?.experience }}</view>
          <view class="head-cuisines">
            <text v-for="(c, i) in cuisineList(chef?.cuisines)" :key="i" class="tag">{{ c }}</text>
          </view>
        </view>
      </view>

      <!-- 套餐列表 -->
      <view class="card">
        <view class="card-title">宴席套餐</view>
        <view v-for="p in packages" :key="p.id" class="pkg" @click="goBooking(p.id)">
          <view class="row-between">
            <text class="pkg-name">{{ p.name }}</text>
            <text class="price">¥{{ yuan(p.price_per_table) }}/桌</text>
          </view>
          <view class="muted">
            {{ p.banquet_type }} · {{ p.min_tables }}桌起
            <text v-if="p.status !== 1" class="tag-red tag pkg-off">已下架</text>
          </view>
          <view class="pkg-dishes">{{ dishSummary(p.dishes) }}</view>
        </view>
        <view v-if="!packages.length" class="empty">暂无套餐</view>
      </view>

      <!-- 档期日历 -->
      <view class="card">
        <view class="row-between">
          <view class="card-title">近期档期（未来30天）</view>
          <view class="legend">
            <text class="lg lg-free">可约</text>
            <text class="lg lg-locked">已锁</text>
            <text class="lg lg-rest">休息</text>
          </view>
        </view>
        <scroll-view scroll-x class="cal-scroll">
          <view class="cal-row">
            <view v-for="d in days" :key="d.date" class="cal-cell" :class="d.cls">
              <text class="cal-date">{{ d.label }}</text>
              <text class="cal-status">{{ d.text }}</text>
            </view>
          </view>
        </scroll-view>
      </view>

      <!-- 案例占位 -->
      <view v-if="cases.length" class="card">
        <view class="card-title">宴席案例</view>
        <view v-for="c in cases" :key="c.id" class="case-item">
          <view class="row-between">
            <text class="strong">{{ c.title }}</text>
          </view>
          <view class="muted">{{ c.description }}</view>
        </view>
      </view>

      <!-- 用户评价 -->
      <view class="card">
        <view class="card-title">用户评价（{{ reviewTotal }}）</view>
        <view v-for="r in reviews" :key="r.id" class="review-item">
          <view class="row-between">
            <text class="review-rating">★ {{ r.rating.toFixed(1) }}</text>
            <text class="muted">{{ fmtTime(r.created_at) }}</text>
          </view>
          <view v-if="tagList(r.tags).length" class="review-tags">
            <text v-for="(t, i) in tagList(r.tags)" :key="i" class="tag tag-grey review-tag">{{ t }}</text>
          </view>
          <view class="review-content">{{ r.content }}</view>
          <view v-if="r.reply" class="review-reply">乡厨回复：{{ r.reply }}</view>
        </view>
        <view v-if="!reviews.length" class="empty">暂无评价</view>
        <view v-else-if="!reviewFinished" class="btn-plain load-more" @click="loadReviews">
          {{ reviewLoading ? '加载中...' : '加载更多评价' }}
        </view>
      </view>

      <view class="bottom-space"></view>

      <!-- 底部预订栏 -->
      <view class="bottom-bar">
        <view class="btn-primary grow" @click="goBooking()">立即预订</view>
      </view>
    </template>
    <view v-else class="empty">{{ loading ? '加载中...' : '乡厨不存在' }}</view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import type { ChefCase, ChefDetail, ChefPackage, ChefSchedule, Review } from '@xc/shared/types'
import { chefPublicApi } from '@/utils/api'
import { SCHEDULE_STATUS, dateAfter, errMsg, fmtTime, today, yuan } from '@/utils/format'

const chefId = ref(0)
const detail = ref<ChefDetail | null>(null)
const loading = ref(true)

const chef = computed(() => detail.value?.chef ?? null)
const packages = computed(() => detail.value?.packages ?? [])
const cases = computed<ChefCase[]>(() => detail.value?.cases ?? [])

// ---- 档期日历 ----
interface CalDay {
  date: string
  label: string
  text: string
  cls: string
}
const days = ref<CalDay[]>([])

function buildDays(schedules: ChefSchedule[]) {
  const map = new Map<string, number>()
  for (const s of schedules || []) map.set(s.date.slice(0, 10), s.status)
  const arr: CalDay[] = []
  for (let i = 0; i < 30; i++) {
    const d = new Date()
    d.setDate(d.getDate() + i)
    const key = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
    const status = map.get(key) ?? 0
    const info = SCHEDULE_STATUS[status] ?? SCHEDULE_STATUS[0]
    arr.push({ date: key, label: `${d.getMonth() + 1}/${d.getDate()}`, text: info.text, cls: info.cls })
  }
  days.value = arr
}

// ---- 评价 ----
const reviews = ref<Review[]>([])
const reviewPage = ref(1)
const reviewTotal = ref(0)
const reviewLoading = ref(false)
const reviewFinished = ref(false)
const reviewSize = 10

function loadReviews() {
  if (reviewLoading.value || reviewFinished.value) return
  reviewLoading.value = true
  chefPublicApi
    .reviews(chefId.value, reviewPage.value, reviewSize)
    .then((res) => {
      reviewTotal.value = res.total
      const items = res.list || []
      reviews.value = reviewPage.value === 1 ? items : reviews.value.concat(items)
      reviewFinished.value = reviews.value.length >= reviewTotal.value || items.length < reviewSize
    })
    .catch((e: unknown) => uni.showToast({ title: errMsg(e), icon: 'none' }))
    .finally(() => (reviewLoading.value = false))
}

onLoad((query) => {
  chefId.value = Number(query?.id || 0)
  if (!chefId.value) {
    loading.value = false
    return
  }
  chefPublicApi
    .detail(chefId.value)
    .then((res) => (detail.value = res))
    .catch((e: unknown) => uni.showToast({ title: errMsg(e), icon: 'none' }))
    .finally(() => (loading.value = false))
  chefPublicApi
    .schedules(chefId.value, today(), dateAfter(29))
    .then((res) => buildDays(res || []))
    .catch(() => buildDays([]))
  loadReviews()
})

function cuisineList(s: string | undefined): string[] {
  return (s || '')
    .split(/[,，、/]/)
    .map((x) => x.trim())
    .filter(Boolean)
}
function tagList(s: string): string[] {
  return (s || '')
    .split(/[,，、/]/)
    .map((x) => x.trim())
    .filter(Boolean)
}
function dishSummary(s: string): string {
  try {
    const arr = JSON.parse(s || '[]') as unknown
    if (Array.isArray(arr)) return arr.slice(0, 8).join(' / ')
  } catch {
    /* 非法 JSON 原样展示 */
  }
  return s || ''
}
function goBooking(packageId?: number) {
  const url = packageId
    ? `/pages/booking/create?chef_id=${chefId.value}&package_id=${packageId}`
    : `/pages/booking/create?chef_id=${chefId.value}`
  uni.navigateTo({ url })
}
</script>

<style scoped>
.head {
  display: flex;
  padding: 32rpx 24rpx;
  background: #fff;
}
.head-avatar {
  width: 120rpx;
  height: 120rpx;
  border-radius: 50%;
  background: #fff7ed;
  color: #d97706;
  font-size: 48rpx;
  font-weight: 700;
  text-align: center;
  line-height: 120rpx;
  flex-shrink: 0;
  margin-right: 24rpx;
}
.head-name {
  font-size: 38rpx;
  font-weight: 700;
}
.head-meta {
  margin: 8rpx 0;
  display: flex;
  gap: 24rpx;
}
.head-rating {
  color: #d97706;
  font-size: 28rpx;
  font-weight: 600;
}
.head-orders {
  color: #666;
  font-size: 26rpx;
}
.head-cuisines .tag {
  margin-right: 12rpx;
}
.card-title {
  font-size: 32rpx;
  font-weight: 700;
  margin-bottom: 20rpx;
}
.pkg {
  border: 1rpx solid #f0f0f0;
  border-radius: 12rpx;
  padding: 20rpx;
  margin-bottom: 16rpx;
}
.pkg-name {
  font-size: 30rpx;
  font-weight: 600;
}
.pkg-dishes {
  color: #666;
  font-size: 24rpx;
  margin-top: 8rpx;
}
.pkg-off {
  margin-left: 12rpx;
}
.legend {
  display: flex;
  gap: 16rpx;
}
.lg {
  font-size: 20rpx;
  padding: 2rpx 12rpx;
  border-radius: 8rpx;
}
.lg-free {
  background: #ecfdf5;
  color: #059669;
}
.lg-locked {
  background: #f3f4f6;
  color: #999;
}
.lg-rest {
  background: #fef2f2;
  color: #dc2626;
}
.cal-scroll {
  white-space: nowrap;
}
.cal-row {
  display: inline-flex;
  gap: 12rpx;
  padding-bottom: 8rpx;
}
.cal-cell {
  display: inline-flex;
  flex-direction: column;
  align-items: center;
  width: 84rpx;
  padding: 12rpx 0;
  border-radius: 12rpx;
  flex-shrink: 0;
}
.cal-cell.free {
  background: #ecfdf5;
}
.cal-cell.locked {
  background: #f3f4f6;
}
.cal-cell.rest {
  background: #fef2f2;
}
.cal-date {
  font-size: 24rpx;
  color: #333;
}
.cal-status {
  font-size: 20rpx;
  margin-top: 6rpx;
}
.free .cal-status {
  color: #059669;
}
.locked .cal-status {
  color: #999;
}
.rest .cal-status {
  color: #dc2626;
}
.case-item {
  border-bottom: 1rpx solid #f5f5f5;
  padding: 12rpx 0;
}
.review-item {
  border-bottom: 1rpx solid #f5f5f5;
  padding: 16rpx 0;
}
.review-rating {
  color: #d97706;
  font-size: 28rpx;
  font-weight: 600;
}
.review-tags {
  margin: 8rpx 0;
}
.review-tag {
  margin-right: 12rpx;
}
.review-content {
  font-size: 26rpx;
  color: #333;
}
.review-reply {
  margin-top: 8rpx;
  background: #fafafa;
  border-radius: 8rpx;
  padding: 12rpx;
  font-size: 24rpx;
  color: #666;
}
.load-more {
  margin-top: 20rpx;
}
.bottom-space {
  height: 140rpx;
}
.bottom-bar {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  background: #fff;
  padding: 20rpx 24rpx calc(20rpx + env(safe-area-inset-bottom));
  box-shadow: 0 -4rpx 16rpx rgba(0, 0, 0, 0.06);
  display: flex;
  z-index: 10;
}
</style>
