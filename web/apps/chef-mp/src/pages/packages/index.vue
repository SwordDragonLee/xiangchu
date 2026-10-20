<script setup lang="ts">
import { onShow } from '@dcloudio/uni-app'
import { computed, reactive, ref } from 'vue'
import { chefApi, errMsg, toast } from '@/utils/api'
import { BANQUET_TYPES, fen2yuan, parseDishes, yuan2fen } from '@/utils/format'
import type { ChefPackage } from '@xc/shared'

const list = ref<ChefPackage[]>([])
const loading = ref(false)

const showForm = ref(false)
const saving = ref(false)
const editId = ref<number | null>(null)

const form = reactive({
  name: '',
  banquet_type: '',
  price: '', // 元
  min_tables: '10',
  status: true, // true 上架
  dishesText: '', // 每行一道菜
})

const formTitle = computed(() => (editId.value ? '编辑套餐' : '新建套餐'))

onShow(() => {
  load()
})

async function load() {
  loading.value = true
  try {
    list.value = await chefApi.packages()
  } catch (e) {
    toast(errMsg(e))
  } finally {
    loading.value = false
  }
}

function dishesOf(p: ChefPackage): string[] {
  return parseDishes(p.dishes)
}

function openCreate() {
  editId.value = null
  form.name = ''
  form.banquet_type = ''
  form.price = ''
  form.min_tables = '10'
  form.status = true
  form.dishesText = ''
  showForm.value = true
}

function openEdit(p: ChefPackage) {
  editId.value = p.id
  form.name = p.name
  form.banquet_type = p.banquet_type
  form.price = fen2yuan(p.price_per_table)
  form.min_tables = String(p.min_tables ?? 1)
  form.status = p.status === 1
  form.dishesText = dishesOf(p).join('\n')
  showForm.value = true
}

function closeForm() {
  showForm.value = false
}

function pickType(t: string) {
  form.banquet_type = t
}

function validate(): boolean {
  if (!form.name.trim()) {
    toast('请输入套餐名称')
    return false
  }
  if (!form.banquet_type.trim()) {
    toast('请选择或输入宴席类型')
    return false
  }
  const fen = yuan2fen(form.price)
  if (fen <= 0) {
    toast('请输入正确的每桌价格')
    return false
  }
  const mt = parseInt(form.min_tables, 10)
  if (!Number.isFinite(mt) || mt < 1) {
    toast('请输入正确的最少桌数')
    return false
  }
  if (!form.dishesText.trim()) {
    toast('请填写菜品清单')
    return false
  }
  return true
}

async function save() {
  if (!validate()) return
  const dishes = form.dishesText
    .split('\n')
    .map((s) => s.trim())
    .filter(Boolean)
  const payload = {
    name: form.name.trim(),
    banquet_type: form.banquet_type.trim(),
    price_per_table: yuan2fen(form.price),
    min_tables: parseInt(form.min_tables, 10),
    status: form.status ? 1 : 0,
    dishes: JSON.stringify(dishes),
  }
  saving.value = true
  try {
    await chefApi.savePackage(payload, editId.value ?? undefined)
    toast('保存成功')
    showForm.value = false
    await load()
  } catch (e) {
    toast(errMsg(e))
  } finally {
    saving.value = false
  }
}

function toggleStatus(p: ChefPackage) {
  const next = p.status === 1 ? 0 : 1
  uni.showModal({
    title: '提示',
    content: next === 1 ? '确认上架该套餐？' : '确认下架该套餐？下架后用户不可预订',
    success: async (res) => {
      if (!res.confirm) return
      try {
        await chefApi.savePackage(
          {
            name: p.name,
            banquet_type: p.banquet_type,
            price_per_table: p.price_per_table,
            min_tables: p.min_tables,
            status: next,
            dishes: p.dishes,
            sort: p.sort,
          },
          p.id,
        )
        toast(next === 1 ? '已上架' : '已下架')
        await load()
      } catch (e) {
        toast(errMsg(e))
      }
    },
  })
}

function remove(p: ChefPackage) {
  uni.showModal({
    title: '删除套餐',
    content: `确认删除「${p.name}」？删除后不可恢复`,
    confirmColor: '#e6504f',
    success: async (res) => {
      if (!res.confirm) return
      try {
        await chefApi.deletePackage(p.id)
        toast('已删除')
        await load()
      } catch (e) {
        toast(errMsg(e))
      }
    },
  })
}
</script>

<template>
  <view class="page">
    <view class="top-bar">
      <text class="top-tip">共 {{ list.length }} 个套餐</text>
      <view class="add-btn" @tap="openCreate">+ 新建套餐</view>
    </view>

    <view v-if="list.length === 0 && !loading" class="empty">暂无套餐，点击右上角新建</view>

    <view v-for="p in list" :key="p.id" class="card pkg-card">
      <view class="pkg-head">
        <text class="pkg-name">{{ p.name }}</text>
        <text class="pkg-status" :class="p.status === 1 ? 'on' : 'off'">
          {{ p.status === 1 ? '已上架' : '已下架' }}
        </text>
      </view>
      <view class="pkg-line">
        <text class="pkg-price">¥{{ fen2yuan(p.price_per_table) }}</text>
        <text class="pkg-per">/桌</text>
        <text class="pkg-type">{{ p.banquet_type }}</text>
        <text class="pkg-min">最少 {{ p.min_tables }} 桌</text>
      </view>
      <view class="pkg-dishes">菜品（{{ dishesOf(p).length }} 道）：{{ dishesOf(p).join('、') || '暂无' }}</view>
      <view class="pkg-actions">
        <view class="act-btn" @tap="toggleStatus(p)">{{ p.status === 1 ? '下架' : '上架' }}</view>
        <view class="act-btn" @tap="openEdit(p)">编辑</view>
        <view class="act-btn danger" @tap="remove(p)">删除</view>
      </view>
    </view>

    <!-- 新建/编辑弹层 -->
    <view v-if="showForm" class="mask" @tap="closeForm">
      <view class="sheet" @tap.stop>
        <view class="sheet-title">{{ formTitle }}</view>
        <scroll-view scroll-y class="sheet-body">
          <view class="form-item">
            <text class="form-label">套餐名称</text>
            <input v-model="form.name" class="form-input" placeholder="如：吉祥如意宴" />
          </view>
          <view class="form-item">
            <text class="form-label">宴席类型</text>
            <input v-model="form.banquet_type" class="form-input" placeholder="选择或输入，如：婚宴" />
            <view class="type-chips">
              <view
                v-for="t in BANQUET_TYPES"
                :key="t"
                class="type-chip"
                :class="{ on: form.banquet_type === t }"
                @tap="pickType(t)"
              >
                {{ t }}
              </view>
            </view>
          </view>
          <view class="form-row">
            <view class="form-item half">
              <text class="form-label">每桌价格(元)</text>
              <input v-model="form.price" class="form-input" type="digit" placeholder="如：888.00" />
            </view>
            <view class="form-item half">
              <text class="form-label">最少桌数</text>
              <input v-model="form.min_tables" class="form-input" type="number" placeholder="如：10" />
            </view>
          </view>
          <view class="form-item">
            <text class="form-label">是否上架</text>
            <switch :checked="form.status" color="#d97706" @change="(e: any) => (form.status = e.detail.value)" />
          </view>
          <view class="form-item">
            <text class="form-label">菜品清单（每行一道菜名）</text>
            <textarea
              v-model="form.dishesText"
              class="form-textarea dishes"
              placeholder="每行一道菜，如：热菜：粉蒸肉"
            />
          </view>
        </scroll-view>
        <view class="sheet-footer">
          <button class="sheet-btn cancel" @tap="closeForm">取消</button>
          <button class="sheet-btn primary" :disabled="saving" @tap="save">
            {{ saving ? '保存中...' : '保存' }}
          </button>
        </view>
      </view>
    </view>
  </view>
</template>

<style scoped>
.top-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin: 20rpx 24rpx 0;
}

.top-tip {
  font-size: 26rpx;
  color: #999999;
}

.add-btn {
  background: #d97706;
  color: #ffffff;
  font-size: 26rpx;
  padding: 14rpx 28rpx;
  border-radius: 30rpx;
}

.pkg-card {
  margin-top: 20rpx;
}

.pkg-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 14rpx;
}

.pkg-name {
  font-size: 32rpx;
  font-weight: 700;
  color: #262626;
}

.pkg-status {
  font-size: 24rpx;
  padding: 4rpx 20rpx;
  border-radius: 20rpx;
}

.pkg-status.on {
  background: #dcfce7;
  color: #15803d;
}

.pkg-status.off {
  background: #f3f4f6;
  color: #9ca3af;
}

.pkg-line {
  display: flex;
  align-items: baseline;
  margin-bottom: 12rpx;
}

.pkg-price {
  font-size: 36rpx;
  font-weight: 700;
  color: #d97706;
}

.pkg-per {
  font-size: 24rpx;
  color: #999999;
  margin: 0 20rpx 0 6rpx;
}

.pkg-type {
  font-size: 24rpx;
  color: #595959;
  background: #f7f7f8;
  border-radius: 8rpx;
  padding: 4rpx 14rpx;
  margin-right: 16rpx;
}

.pkg-min {
  font-size: 24rpx;
  color: #595959;
}

.pkg-dishes {
  font-size: 24rpx;
  color: #999999;
  line-height: 1.6;
}

.pkg-actions {
  display: flex;
  justify-content: flex-end;
  margin-top: 20rpx;
  padding-top: 20rpx;
  border-top: 2rpx solid #f5f5f5;
}

.act-btn {
  font-size: 26rpx;
  color: #d97706;
  border: 2rpx solid #d97706;
  border-radius: 30rpx;
  padding: 8rpx 32rpx;
  margin-left: 20rpx;
}

.act-btn.danger {
  color: #e6504f;
  border-color: #e6504f;
}

.mask {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  z-index: 1000;
  display: flex;
  align-items: flex-end;
}

.sheet {
  width: 100%;
  max-height: 85vh;
  background: #ffffff;
  border-radius: 32rpx 32rpx 0 0;
  display: flex;
  flex-direction: column;
  padding-bottom: env(safe-area-inset-bottom);
}

.sheet-title {
  text-align: center;
  font-size: 32rpx;
  font-weight: 700;
  padding: 30rpx 0 10rpx;
}

.sheet-body {
  flex: 1;
  padding: 20rpx 32rpx;
  max-height: 62vh;
}

.form-row {
  display: flex;
  justify-content: space-between;
}

.form-item.half {
  width: 48%;
}

.type-chips {
  display: flex;
  flex-wrap: wrap;
  margin-top: 16rpx;
}

.type-chip {
  font-size: 24rpx;
  color: #595959;
  background: #f7f7f8;
  border-radius: 26rpx;
  padding: 8rpx 26rpx;
  margin: 0 14rpx 14rpx 0;
  border: 2rpx solid transparent;
}

.type-chip.on {
  color: #d97706;
  background: #fffbeb;
  border-color: #d97706;
}

.form-textarea.dishes {
  min-height: 220rpx;
}

.sheet-footer {
  display: flex;
  padding: 20rpx 32rpx;
  border-top: 2rpx solid #f0f0f0;
}

.sheet-btn {
  flex: 1;
  height: 84rpx;
  line-height: 84rpx;
  border-radius: 44rpx;
  font-size: 30rpx;
  margin: 0;
}

.sheet-btn.cancel {
  background: #f7f7f8;
  color: #595959;
  margin-right: 24rpx;
}

.sheet-btn.primary {
  background: #d97706;
  color: #ffffff;
}

.sheet-btn[disabled] {
  opacity: 0.6;
}
</style>
