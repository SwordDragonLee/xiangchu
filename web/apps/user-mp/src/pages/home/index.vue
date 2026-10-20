<template>
  <view class="home">
    <!-- 顶部搜索框 -->
    <view class="search-bar" @click="goSearch">
      <view class="search-box">
        <text class="search-icon">🔍</text>
        <text class="search-placeholder">搜索乡厨 / 菜系 / 地区</text>
      </view>
    </view>

    <!-- banner 占位 -->
    <view class="banner page-pad">
      <view class="banner-inner">
        <view class="banner-title">乡厨到家</view>
        <view class="banner-sub">地道乡厨 · 家宴上门 · 宴席一条龙</view>
      </view>
    </view>

    <!-- 优惠券横幅 -->
    <view v-if="coupons.length" class="coupon-strip">
      <view class="coupon-label">领券</view>
      <scroll-view scroll-x class="coupon-scroll">
        <view class="coupon-row">
          <view
            v-for="c in coupons"
            :key="c.id"
            class="coupon-chip"
            @click="goCouponCenter"
          >
            <text class="coupon-amount">¥{{ yuan(c.amount) }}</text>
            <text class="coupon-cond">满{{ yuan(c.threshold) }}可用</text>
            <text class="coupon-name">{{ c.name }}</text>
          </view>
        </view>
      </scroll-view>
      <view class="coupon-more" @click="goCouponCenter">更多 ›</view>
    </view>

    <!-- 推荐乡厨 -->
    <view class="section page-pad">
      <view class="section-head">
        <text class="section-title">推荐乡厨</text>
        <text class="section-more" @click="goChefList">全部 ›</text>
      </view>

      <view v-for="chef in chefs" :key="chef.id" class="card chef-card" @click="goChefDetail(chef.id)">
        <view class="chef-avatar">{{ chef.real_name.slice(0, 1) }}</view>
        <view class="grow chef-info">
          <view class="row">
            <text class="chef-name">{{ chef.real_name }}</text>
            <text class="chef-rating">★ {{ chef.rating.toFixed(1) }}</text>
          </view>
          <view class="muted">{{ chef.region }} · 已接单 {{ chef.order_count }} 单</view>
          <view class="chef-cuisines">
            <text v-for="(c, i) in cuisineList(chef.cuisines)" :key="i" class="tag chef-cuisine">{{ c }}</text>
          </view>
        </view>
      </view>

      <view v-if="!chefs.length && !loading" class="empty">暂无推荐乡厨</view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import type { Chef, Coupon } from '@xc/shared/types'
import { chefPublicApi, userApi } from '@/utils/api'
import { yuan, errMsg } from '@/utils/format'

const chefs = ref<Chef[]>([])
const coupons = ref<Coupon[]>([])
const loading = ref(false)
let loaded = false

function cuisineList(s: string): string[] {
  return (s || '')
    .split(/[,，、/]/)
    .map((x) => x.trim())
    .filter(Boolean)
    .slice(0, 3)
}

function load() {
  loading.value = true
  chefPublicApi
    .list({ order_by: 'rating_desc', page: 1, size: 10 })
    .then((res) => (chefs.value = res.list || []))
    .catch((e: unknown) => uni.showToast({ title: errMsg(e), icon: 'none' }))
    .finally(() => (loading.value = false))
  userApi
    .couponCenter(1, 3)
    .then((res) => (coupons.value = res.list || []))
    .catch(() => {
      /* 券中心不可用不影响首页 */
    })
}

onShow(() => {
  if (!loaded) {
    loaded = true
    load()
  }
})

function goSearch() {
  uni.navigateTo({ url: '/pages/chef/list' })
}
function goChefList() {
  uni.navigateTo({ url: '/pages/chef/list' })
}
function goChefDetail(id: number) {
  uni.navigateTo({ url: `/pages/chef/detail?id=${id}` })
}
function goCouponCenter() {
  uni.navigateTo({ url: '/pages/coupon/index' })
}
</script>

<style scoped>
.search-bar {
  padding: 20rpx 24rpx;
  background: #d97706;
}
.search-box {
  background: #fff;
  border-radius: 36rpx;
  height: 68rpx;
  display: flex;
  align-items: center;
  padding: 0 28rpx;
}
.search-icon {
  font-size: 26rpx;
  margin-right: 14rpx;
}
.search-placeholder {
  color: #999;
  font-size: 26rpx;
}
.banner {
  padding-top: 24rpx;
}
.banner-inner {
  background: linear-gradient(135deg, #d97706, #f59e0b);
  border-radius: 20rpx;
  padding: 48rpx 40rpx;
  color: #fff;
}
.banner-title {
  font-size: 44rpx;
  font-weight: 700;
}
.banner-sub {
  margin-top: 12rpx;
  font-size: 26rpx;
  opacity: 0.9;
}
.coupon-strip {
  margin: 24rpx;
  background: #fff;
  border-radius: 16rpx;
  padding: 20rpx;
  display: flex;
  align-items: center;
}
.coupon-label {
  flex-shrink: 0;
  color: #d97706;
  font-weight: 700;
  font-size: 28rpx;
  margin-right: 16rpx;
}
.coupon-scroll {
  flex: 1;
  white-space: nowrap;
}
.coupon-row {
  display: inline-flex;
  gap: 16rpx;
}
.coupon-chip {
  display: inline-flex;
  flex-direction: column;
  align-items: center;
  background: #fff7ed;
  border: 1rpx dashed #d97706;
  border-radius: 12rpx;
  padding: 10rpx 20rpx;
  margin-right: 16rpx;
}
.coupon-amount {
  color: #d97706;
  font-weight: 700;
  font-size: 30rpx;
}
.coupon-cond {
  color: #b45309;
  font-size: 20rpx;
}
.coupon-name {
  color: #999;
  font-size: 20rpx;
  max-width: 160rpx;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.coupon-more {
  flex-shrink: 0;
  color: #999;
  font-size: 24rpx;
  margin-left: 12rpx;
}
.section-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20rpx;
}
.section-title {
  font-size: 34rpx;
  font-weight: 700;
}
.section-more {
  color: #999;
  font-size: 26rpx;
}
.chef-card {
  display: flex;
  align-items: center;
}
.chef-avatar {
  width: 96rpx;
  height: 96rpx;
  border-radius: 50%;
  background: #fff7ed;
  color: #d97706;
  font-size: 40rpx;
  font-weight: 700;
  text-align: center;
  line-height: 96rpx;
  flex-shrink: 0;
  margin-right: 20rpx;
}
.chef-info {
  min-width: 0;
}
.chef-name {
  font-size: 32rpx;
  font-weight: 600;
  margin-right: 16rpx;
}
.chef-rating {
  color: #d97706;
  font-size: 26rpx;
}
.chef-cuisines {
  margin-top: 8rpx;
}
.chef-cuisine {
  margin-right: 12rpx;
}
</style>
