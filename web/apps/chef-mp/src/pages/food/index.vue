<script setup lang="ts">
import { onPullDownRefresh, onReachBottom, onShow } from '@dcloudio/uni-app'
import { computed, ref } from 'vue'
import { ensureLogin, ensureChef, errMsg, foodApi, toast } from '@/utils/api'
import { INGREDIENT_CATEGORIES, fen2yuan } from '@/utils/format'
import { useCart } from '@/stores/cart'
import type { Ingredient } from '@xc/shared'

const cats = ['', ...INGREDIENT_CATEGORIES] // '' = 全部

const keyword = ref('')
const activeCat = ref('')
const list = ref<Ingredient[]>([])
const page = ref(1)
const size = 10
const total = ref(0)
const loading = ref(false)
const finished = computed(() => total.value > 0 && list.value.length >= total.value)

const cart = useCart()

onShow(async () => {
  ensureAndLoad()
})

async function ensureAndLoad() {
  if (!ensureLogin()) return
  if (!(await ensureChef())) return
  reload()
}

onPullDownRefresh(async () => {
  await reload()
  uni.stopPullDownRefresh()
})

onReachBottom(() => {
  if (!finished.value && !loading.value) load(page.value + 1)
})

async function switchCat(c: string) {
  if (activeCat.value === c) return
  activeCat.value = c
  await reload()
}

async function reload() {
  list.value = []
  total.value = 0
  await load(1)
}

async function load(p: number) {
  loading.value = true
  try {
    const res = await foodApi.ingredients({
      keyword: keyword.value.trim() || undefined,
      category: activeCat.value || undefined,
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

function onAdd(it: Ingredient) {
  const switched = cart.add(it)
  if (switched) toast('不同供应商的食材不可合并下单，已更新选择')
}

function qtyOf(it: Ingredient): number {
  return cart.state.items.find((i) => i.ingredient_id === it.id)?.quantity ?? 0
}

function onMinus(it: Ingredient) {
  const q = qtyOf(it)
  if (q <= 1) {
    cart.remove(it.id)
  } else {
    cart.setQty(it.id, q - 1)
  }
}

function onPlus(it: Ingredient) {
  cart.setQty(it.id, qtyOf(it) + 1)
}

function goCheckout() {
  uni.navigateTo({ url: '/pages/food/checkout' })
}

function clearCart() {
  uni.showModal({
    title: '清空已选',
    content: '确认清空当前已选食材？',
    success: (r) => {
      if (r.confirm) cart.clear()
    },
  })
}
</script>

<template>
  <view class="page">
    <view class="filter-bar">
      <view class="search-row">
        <input
          v-model="keyword"
          class="search-input"
          placeholder="搜索食材名称/产地"
          confirm-type="search"
          @confirm="reload"
        />
        <view class="search-btn" @tap="reload">搜索</view>
      </view>
      <scroll-view scroll-x class="cat-scroll">
        <view class="cat-row">
          <view
            v-for="c in cats"
            :key="c || 'all'"
            class="chip"
            :class="{ on: activeCat === c }"
            @tap="switchCat(c)"
          >
            {{ c || '全部' }}
          </view>
        </view>
      </scroll-view>
    </view>

    <view v-if="list.length === 0 && !loading" class="empty">暂无相关食材</view>

    <view v-for="it in list" :key="it.id" class="card ing-card" :class="{ off: it.stock <= 0 }">
      <view class="row1">
        <text class="g-name">{{ it.name }}</text>
        <text class="g-price">¥{{ fen2yuan(it.price) }}<text class="g-unit">/{{ it.unit }}</text></text>
      </view>
      <view class="row-line">
        <text class="g-cat">{{ it.category }}<template v-if="it.origin"> · 产地 {{ it.origin }}</template></text>
        <text class="g-stock" :class="{ low: it.stock > 0 && it.stock < 10 }">
          {{ it.stock <= 0 ? '已售罄' : `库存 ${it.stock}${it.unit}` }}
        </text>
      </view>
      <view class="op-row">
        <template v-if="it.stock > 0">
          <view v-if="qtyOf(it) > 0" class="stepper">
            <view class="step-btn" @tap="onMinus(it)">−</view>
            <text class="step-num">{{ qtyOf(it) }}</text>
            <view class="step-btn" @tap="onPlus(it)">+</view>
          </view>
          <button v-else class="add-btn" @tap="onAdd(it)">加入</button>
        </template>
      </view>
    </view>

    <view v-if="list.length > 0" class="load-more">
      {{ loading ? '市场加载中...' : finished ? '没有更多了' : '上拉加载更多' }}
    </view>

    <view v-if="cart.count.value > 0" class="cart-bar">
      <view class="cart-info" @tap="clearCart">
        <text class="cart-count">已选 {{ cart.count.value }} 项</text>
        <text class="cart-total">¥{{ fen2yuan(cart.totalFen.value) }}</text>
      </view>
      <button class="checkout-btn" @tap="goCheckout">去结算</button>
    </view>

    <view class="bottom-space" />
  </view>
</template>

<style scoped>
.filter-bar {
  position: sticky;
  top: 0;
  z-index: 10;
  background: #f5f6f8;
  padding: 20rpx 24rpx 12rpx;
}

.search-row {
  display: flex;
  gap: 16rpx;
}

.search-input {
  flex: 1;
  height: 72rpx;
  background: #ffffff;
  border-radius: 36rpx;
  padding: 0 30rpx;
  font-size: 26rpx;
}

.search-btn {
  width: 110rpx;
  height: 72rpx;
  border-radius: 36rpx;
  background: #d97706;
  color: #ffffff;
  font-size: 26rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}

.cat-scroll {
  margin-top: 16rpx;
  white-space: nowrap;
}

.cat-row {
  display: inline-flex;
  gap: 16rpx;
  padding: 4rpx 0;
}

.chip {
  display: inline-block;
  padding: 10rpx 28rpx;
  border-radius: 30rpx;
  background: #ffffff;
  color: #666666;
  font-size: 24rpx;
}

.chip.on {
  background: #d97706;
  color: #ffffff;
  font-weight: 600;
}

.ing-card {
  margin: 16rpx 24rpx 0;
  padding: 24rpx;
}

.off {
  opacity: 0.5;
}

.row1 {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  gap: 16rpx;
}

.g-name {
  font-size: 30rpx;
  font-weight: 600;
  color: #262626;
}

.g-price {
  font-size: 30rpx;
  font-weight: 700;
  color: #d97706;
  white-space: nowrap;
}

.g-unit {
  font-size: 22rpx;
  font-weight: 400;
  color: #999999;
}

.row-line {
  display: flex;
  justify-content: space-between;
  margin-top: 10rpx;
  font-size: 24rpx;
  color: #999999;
}

.g-stock.low {
  color: #e6504f;
}

.op-row {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  margin-top: 14rpx;
  min-height: 56rpx;
}

.stepper {
  display: flex;
  align-items: center;
}

.step-btn {
  width: 52rpx;
  height: 52rpx;
  border-radius: 8rpx;
  background: #f5f6f8;
  color: #262626;
  font-size: 32rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}

.step-num {
  min-width: 70rpx;
  text-align: center;
  font-size: 28rpx;
  color: #262626;
}

.add-btn {
  margin: 0;
  padding: 0 32rpx;
  height: 56rpx;
  line-height: 56rpx;
  border-radius: 28rpx;
  background: #d97706;
  color: #ffffff;
  font-size: 24rpx;
}

.cart-bar {
  position: fixed;
  left: 16rpx;
  right: 16rpx;
  bottom: 20rpx;
  z-index: 99;
  display: flex;
  align-items: center;
  background: #262626;
  border-radius: 48rpx;
  padding: 12rpx 12rpx 12rpx 32rpx;
  box-shadow: 0 8rpx 24rpx rgba(0, 0, 0, 0.25);
}

.cart-info {
  flex: 1;
  display: flex;
  align-items: baseline;
  gap: 12rpx;
}

.cart-count {
  font-size: 24rpx;
  color: #cccccc;
}

.cart-total {
  font-size: 34rpx;
  font-weight: 700;
  color: #ffffff;
}

.checkout-btn {
  margin: 0;
  padding: 0 44rpx;
  height: 72rpx;
  line-height: 72rpx;
  border-radius: 36rpx;
  background: #d97706;
  color: #ffffff;
  font-size: 28rpx;
  font-weight: 600;
}

.load-more {
  text-align: center;
  font-size: 24rpx;
  color: #bbbbbb;
  padding: 24rpx 0;
}

.empty {
  text-align: center;
  color: #bbbbbb;
  font-size: 26rpx;
  padding: 120rpx 0;
}

.bottom-space {
  height: 160rpx;
}
</style>
