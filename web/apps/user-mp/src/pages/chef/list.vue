<template>
  <view class="chef-list page-pad">
    <!-- 筛选区 -->
    <view class="card filter">
      <input
        v-model="keyword"
        class="filter-input"
        placeholder="搜索乡厨姓名 / 地区"
        confirm-type="search"
        @confirm="reload"
      />
      <view class="chips">
        <view class="chip" :class="{ active: cuisine === '' }" @click="pickCuisine('')">全部菜系</view>
        <view
          v-for="c in cuisines"
          :key="c"
          class="chip"
          :class="{ active: cuisine === c }"
          @click="pickCuisine(c)"
        >
          {{ c }}
        </view>
      </view>
      <view class="chips sort-chips">
        <view
          v-for="o in orders"
          :key="o.value"
          class="chip"
          :class="{ active: orderBy === o.value }"
          @click="pickOrder(o.value)"
        >
          {{ o.label }}
        </view>
      </view>
    </view>

    <!-- 列表 -->
    <view v-for="chef in list" :key="chef.id" class="card chef-card" @click="goDetail(chef.id)">
      <view class="chef-avatar">{{ chef.real_name.slice(0, 1) }}</view>
      <view class="grow chef-info">
        <view class="row">
          <text class="chef-name">{{ chef.real_name }}</text>
          <text class="chef-rating">★ {{ chef.rating.toFixed(1) }}</text>
        </view>
        <view class="muted">{{ chef.region }}</view>
        <view class="muted">擅长：{{ chef.cuisines || '-' }}</view>
        <view class="row-between">
          <text class="muted">已接单 {{ chef.order_count }} 单</text>
          <text class="tag">经验 {{ chef.experience || '-' }}</text>
        </view>
      </view>
    </view>

    <view v-if="!list.length && !loading" class="empty">没有找到符合条件的乡厨</view>
    <view v-if="loading" class="empty">加载中...</view>
    <view v-else-if="finished && list.length" class="empty load-tip">没有更多了</view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onReachBottom } from '@dcloudio/uni-app'
import type { Chef } from '@xc/shared/types'
import { chefPublicApi } from '@/utils/api'
import { CUISINES, errMsg } from '@/utils/format'

const cuisines = CUISINES
const orders = [
  { value: 'rating_desc', label: '评分最高' },
  { value: 'order_count_desc', label: '接单最多' },
]

const keyword = ref('')
const cuisine = ref('')
const orderBy = ref('rating_desc')

const list = ref<Chef[]>([])
const page = ref(1)
const size = 10
const total = ref(0)
const loading = ref(false)
const finished = ref(false)

function params() {
  return {
    keyword: keyword.value || undefined,
    cuisine: cuisine.value || undefined,
    order_by: orderBy.value,
    page: page.value,
    size,
  }
}

function loadPage() {
  if (loading.value || finished.value) return
  loading.value = true
  chefPublicApi
    .list(params())
    .then((res) => {
      total.value = res.total
      const items = res.list || []
      list.value = page.value === 1 ? items : list.value.concat(items)
      finished.value = list.value.length >= total.value || items.length < size
    })
    .catch((e: unknown) => uni.showToast({ title: errMsg(e), icon: 'none' }))
    .finally(() => (loading.value = false))
}

function reload() {
  page.value = 1
  finished.value = false
  list.value = []
  loadPage()
}

function pickCuisine(c: string) {
  cuisine.value = c
  reload()
}
function pickOrder(v: string) {
  orderBy.value = v
  reload()
}

onReachBottom(() => {
  if (!finished.value && !loading.value) {
    page.value += 1
    loadPage()
  }
})

reload()

function goDetail(id: number) {
  uni.navigateTo({ url: `/pages/chef/detail?id=${id}` })
}
</script>

<style scoped>
.filter-input {
  background: #f5f5f5;
  border-radius: 32rpx;
  height: 64rpx;
  padding: 0 28rpx;
  font-size: 26rpx;
  margin-bottom: 20rpx;
}
.sort-chips {
  margin-top: 16rpx;
}
.chef-card {
  display: flex;
  align-items: flex-start;
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
.load-tip {
  padding: 24rpx 0;
}
</style>
