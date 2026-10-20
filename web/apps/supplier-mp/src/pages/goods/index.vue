<script setup lang="ts">
import { onPullDownRefresh, onReachBottom, onShow } from '@dcloudio/uni-app'
import { computed, reactive, ref } from 'vue'
import TabBar from '@/components/TabBar.vue'
import { ensureLogin, ensureSupplier, errMsg, supplierApi, toast } from '@/utils/api'
import { INGREDIENT_CATEGORIES, fen2yuan, yuan2fen } from '@/utils/format'
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

// 新增/编辑表单面板
const showForm = ref(false)
const editId = ref<number | null>(null)
const saving = ref(false)
const form = reactive({
  name: '',
  category: '',
  unit: '',
  price: '',
  stock: '',
  origin: '',
})

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

async function switchCat(c: string) {
  if (activeCat.value === c) return
  activeCat.value = c
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
    const res = await supplierApi.ingredients({
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

function openNew() {
  editId.value = null
  form.name = ''
  form.category = ''
  form.unit = ''
  form.price = ''
  form.stock = ''
  form.origin = ''
  showForm.value = true
}

function openEdit(it: Ingredient) {
  editId.value = it.id
  form.name = it.name
  form.category = it.category
  form.unit = it.unit
  form.price = fen2yuan(it.price)
  form.stock = String(it.stock)
  form.origin = it.origin || ''
  showForm.value = true
}

function closeForm() {
  showForm.value = false
  editId.value = null
}

function pickCategory(c: string) {
  form.category = c
}

async function save() {
  if (!form.name.trim()) {
    toast('请输入商品名称')
    return
  }
  if (!form.category) {
    toast('请选择分类')
    return
  }
  if (!form.unit.trim()) {
    toast('请输入单位，如 斤/份/箱')
    return
  }
  const price = yuan2fen(form.price)
  if (price <= 0) {
    toast('请输入正确的单价')
    return
  }
  const stock = Math.floor(Number(form.stock))
  if (!Number.isFinite(stock) || stock < 0) {
    toast('请输入正确的库存')
    return
  }
  saving.value = true
  try {
    await supplierApi.saveIngredient(
      {
        name: form.name.trim(),
        category: form.category,
        unit: form.unit.trim(),
        price,
        stock,
        origin: form.origin.trim(),
      },
      editId.value ?? undefined,
    )
    toast(editId.value ? '已保存' : '已上架')
    closeForm()
    await reload()
  } catch (e) {
    toast(errMsg(e))
  } finally {
    saving.value = false
  }
}

function offShelf(it: Ingredient) {
  uni.showModal({
    title: '下架商品',
    content: `确认下架「${it.name}」？下架后买家不可再下单`,
    success: async (res) => {
      if (!res.confirm) return
      try {
        await supplierApi.offShelfIngredient(it.id)
        toast('已下架')
        await reload()
      } catch (e) {
        toast(errMsg(e))
      }
    },
  })
}
</script>

<template>
  <view class="page">
    <!-- 新增/编辑表单面板 -->
    <view v-if="showForm" class="card form-card">
      <view class="card-title">{{ editId ? '编辑商品' : '新增商品' }}</view>
      <view class="form-item">
        <text class="form-label">名称</text>
        <input v-model="form.name" class="form-input" placeholder="如：本地土猪肉" />
      </view>
      <view class="form-item">
        <text class="form-label">分类</text>
        <view class="chips">
          <view
            v-for="c in INGREDIENT_CATEGORIES"
            :key="c"
            class="chip"
            :class="{ on: form.category === c }"
            @tap="pickCategory(c)"
          >
            {{ c }}
          </view>
        </view>
      </view>
      <view class="form-cols">
        <view class="form-item col">
          <text class="form-label">单位</text>
          <input v-model="form.unit" class="form-input" placeholder="斤/份/箱" />
        </view>
        <view class="form-item col">
          <text class="form-label">单价(元)</text>
          <input v-model="form.price" class="form-input" type="digit" placeholder="如 18.5" />
        </view>
      </view>
      <view class="form-cols">
        <view class="form-item col">
          <text class="form-label">库存</text>
          <input v-model="form.stock" class="form-input" type="number" placeholder="如 100" />
        </view>
        <view class="form-item col">
          <text class="form-label">产地(选填)</text>
          <input v-model="form.origin" class="form-input" placeholder="如 本地/XX县" />
        </view>
      </view>
      <view class="btn-row">
        <button class="btn-plain half" @tap="closeForm">取消</button>
        <button class="btn-primary half" :disabled="saving" @tap="save">
          {{ saving ? '保存中...' : '保存' }}
        </button>
      </view>
    </view>

    <!-- 搜索 + 分类筛选 -->
    <view class="filter-bar">
      <view class="search-row">
        <input
          v-model="keyword"
          class="search-input"
          placeholder="搜索商品名称"
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

    <!-- 商品列表 -->
    <view v-if="list.length === 0 && !loading" class="empty">暂无商品，点击下方按钮新增</view>
    <view v-for="it in list" :key="it.id" class="card goods-card">
      <view class="row1">
        <text class="g-name">{{ it.name }}</text>
        <text class="g-status" :class="it.status === 1 ? 'on' : 'off'">
          {{ it.status === 1 ? '在售' : '已下架' }}
        </text>
      </view>
      <view class="row-line">
        <text class="g-cat">{{ it.category }}</text>
        <text class="g-price">¥{{ fen2yuan(it.price) }}<text class="g-unit">/{{ it.unit }}</text></text>
      </view>
      <view class="row-line">
        <text class="g-stock" :class="{ low: it.stock < 10 }">库存 {{ it.stock }}{{ it.unit }}</text>
        <text v-if="it.origin" class="g-origin">产地：{{ it.origin }}</text>
      </view>
      <view class="op-row">
        <button class="op-btn plain" @tap="openEdit(it)">编辑</button>
        <button v-if="it.status === 1" class="op-btn danger" @tap="offShelf(it)">下架</button>
      </view>
    </view>

    <view v-if="list.length > 0" class="load-more">
      {{ loading ? '加载中...' : finished ? '没有更多了' : '上拉加载更多' }}
    </view>

    <!-- 新增按钮 -->
    <view v-if="!showForm" class="fab" @tap="openNew">+ 新增商品</view>

    <view class="bottom-space" />
    <TabBar current="goods" />
  </view>
</template>

<style scoped>
.form-card {
  border: 2rpx solid #bbf7d0;
}

.form-cols {
  display: flex;
  justify-content: space-between;
}

.form-item.col {
  width: 48%;
}

.chips {
  display: flex;
  flex-wrap: wrap;
}

.chip {
  padding: 10rpx 26rpx;
  background: #f7f7f8;
  border-radius: 28rpx;
  font-size: 24rpx;
  color: #595959;
  margin: 0 14rpx 14rpx 0;
  border: 2rpx solid transparent;
}

.chip.on {
  background: #f0fdf4;
  color: #16a34a;
  border-color: #16a34a;
}

.btn-row {
  display: flex;
  justify-content: space-between;
}

.half {
  width: 48%;
  padding: 0;
}

.filter-bar {
  background: #ffffff;
  padding: 20rpx 24rpx 4rpx;
  position: sticky;
  top: 0;
  z-index: 10;
  border-bottom: 2rpx solid #f0f0f0;
}

.search-row {
  display: flex;
  align-items: center;
  margin-bottom: 16rpx;
}

.search-input {
  flex: 1;
  background: #f7f7f8;
  border-radius: 34rpx;
  height: 68rpx;
  padding: 0 28rpx;
  font-size: 26rpx;
}

.search-btn {
  margin-left: 16rpx;
  color: #16a34a;
  font-size: 28rpx;
  font-weight: 600;
  padding: 0 8rpx;
}

.cat-scroll {
  white-space: nowrap;
}

.cat-row {
  display: inline-flex;
  padding-bottom: 16rpx;
}

.cat-row .chip {
  display: inline-block;
  margin: 0 14rpx 0 0;
}

.goods-card {
  margin-top: 20rpx;
}

.row1 {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12rpx;
}

.g-name {
  font-size: 30rpx;
  font-weight: 600;
  color: #262626;
}

.g-status {
  font-size: 22rpx;
  padding: 4rpx 16rpx;
  border-radius: 18rpx;
}

.g-status.on {
  background: #dcfce7;
  color: #15803d;
}

.g-status.off {
  background: #f0f0f0;
  color: #999999;
}

.row-line {
  display: flex;
  justify-content: space-between;
  font-size: 26rpx;
  margin-bottom: 8rpx;
}

.g-cat {
  color: #595959;
}

.g-price {
  color: #16a34a;
  font-weight: 700;
  font-size: 30rpx;
}

.g-unit {
  font-size: 22rpx;
  color: #999999;
  font-weight: 400;
}

.g-stock {
  color: #595959;
}

.g-stock.low {
  color: #e6504f;
}

.g-origin {
  color: #999999;
  font-size: 24rpx;
}

.op-row {
  display: flex;
  justify-content: flex-end;
  margin-top: 12rpx;
}

.op-btn {
  font-size: 24rpx;
  line-height: 56rpx;
  height: 56rpx;
  padding: 0 32rpx;
  margin: 0 0 0 16rpx;
  border-radius: 30rpx;
}

.op-btn.plain {
  background: #f0fdf4;
  color: #16a34a;
  border: 2rpx solid #16a34a;
}

.op-btn.danger {
  background: #ffffff;
  color: #e6504f;
  border: 2rpx solid #e6504f;
}

.load-more {
  text-align: center;
  color: #bbbbbb;
  font-size: 24rpx;
  padding: 30rpx 0;
}

.fab {
  position: fixed;
  right: 32rpx;
  bottom: 180rpx;
  background: #16a34a;
  color: #ffffff;
  font-size: 28rpx;
  font-weight: 600;
  padding: 24rpx 36rpx;
  border-radius: 44rpx;
  box-shadow: 0 8rpx 20rpx rgba(22, 163, 74, 0.35);
  z-index: 99;
}

.bottom-space {
  height: 140rpx;
}
</style>
