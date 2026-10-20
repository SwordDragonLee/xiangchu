<script setup lang="ts">
import { onShow } from '@dcloudio/uni-app'
import { computed, reactive, ref } from 'vue'
import TabBar from '@/components/TabBar.vue'
import { clearSession, errMsg, getToken, supplierApi, toast } from '@/utils/api'
import { INGREDIENT_CATEGORIES, SUPPLIER_STATUS_TEXT, splitCategories } from '@/utils/format'
import type { Supplier } from '@xc/shared'

const sup = ref<Supplier | null>(null)
const loading = ref(false)
const saving = ref(false)
const editing = ref(false)

const form = reactive({
  name: '',
  license: '',
  contact: '',
  phone: '',
  region: '',
  address: '',
})
const categories = ref<string[]>([])

const statusText = computed(() => (sup.value ? SUPPLIER_STATUS_TEXT[sup.value.status] || '未知' : ''))

onShow(() => {
  if (!getToken()) {
    uni.reLaunch({ url: '/pages/login/index' })
    return
  }
  load()
})

async function load() {
  loading.value = true
  try {
    const p = await supplierApi.profile()
    sup.value = p
    fillForm(p)
  } catch (e) {
    toast(errMsg(e))
  } finally {
    loading.value = false
  }
}

function fillForm(p: Supplier) {
  form.name = p.name || ''
  form.license = p.license || ''
  form.contact = p.contact || ''
  form.phone = p.phone || ''
  form.region = p.region || ''
  form.address = p.address || ''
  categories.value = splitCategories(p.categories)
}

function startEdit() {
  if (sup.value) fillForm(sup.value)
  editing.value = true
}

function toggleCategory(c: string) {
  const i = categories.value.indexOf(c)
  if (i >= 0) categories.value.splice(i, 1)
  else categories.value.push(c)
}

async function save() {
  if (!form.name.trim()) {
    toast('请输入商铺名称')
    return
  }
  if (!form.license.trim()) {
    toast('请填写营业执照/资质证明')
    return
  }
  if (!form.contact.trim()) {
    toast('请输入联系人')
    return
  }
  if (!/^1\d{10}$/.test(form.phone)) {
    toast('请输入正确的手机号')
    return
  }
  if (!form.region.trim()) {
    toast('请输入所在地区')
    return
  }
  if (!form.address.trim()) {
    toast('请输入详细地址')
    return
  }
  if (categories.value.length === 0) {
    toast('请至少选择一个经营品类')
    return
  }
  saving.value = true
  try {
    await supplierApi.updateProfile({
      name: form.name.trim(),
      license: form.license.trim(),
      contact: form.contact.trim(),
      phone: form.phone,
      region: form.region.trim(),
      address: form.address.trim(),
      categories: categories.value.join(','),
    })
    toast('保存成功')
    editing.value = false
    await load()
  } catch (e) {
    toast(errMsg(e))
  } finally {
    saving.value = false
  }
}

function statusClass(s: number): string {
  if (s === 1) return 'pass'
  if (s === 2 || s === 3) return 'fail'
  return 'wait'
}

function logout() {
  uni.showModal({
    title: '退出登录',
    content: '确认退出当前账号？',
    success: (res) => {
      if (!res.confirm) return
      clearSession()
      uni.reLaunch({ url: '/pages/login/index' })
    },
  })
}
</script>

<template>
  <view class="page">
    <view v-if="sup" class="head">
      <view class="avatar">{{ (sup.name || '供').slice(0, 1) }}</view>
      <view class="head-info">
        <view class="head-name">
          {{ sup.name }}
          <text class="status-badge" :class="statusClass(sup.status)">{{ statusText }}</text>
        </view>
        <view class="head-sub">评分 {{ (sup.rating ?? 0).toFixed(1) }} · {{ sup.contact }}</view>
      </view>
    </view>

    <view v-if="sup && (sup.status === 2 || sup.status === 3)" class="card audit-card">
      <text class="audit-label">{{ sup.status === 3 ? '禁用原因' : '拒绝原因' }}</text>
      <text class="audit-content">{{ sup.audit_remark || '未填写' }}</text>
    </view>

    <view v-if="sup && !editing" class="card">
      <view class="detail-row"><text class="dl">商铺名称</text><text class="dv">{{ sup.name }}</text></view>
      <view class="detail-row"><text class="dl">联系人</text><text class="dv">{{ sup.contact }}</text></view>
      <view class="detail-row"><text class="dl">手机号</text><text class="dv">{{ sup.phone }}</text></view>
      <view class="detail-row"><text class="dl">所在地区</text><text class="dv">{{ sup.region }}</text></view>
      <view class="detail-row"><text class="dl">详细地址</text><text class="dv">{{ sup.address }}</text></view>
      <view class="detail-row">
        <text class="dl">经营品类</text>
        <text class="dv">{{ splitCategories(sup.categories).join('、') || '暂无' }}</text>
      </view>
      <view class="detail-row"><text class="dl">资质证照</text><text class="dv">{{ sup.license || '暂无' }}</text></view>
      <view class="detail-row">
        <text class="dl">评分</text>
        <text class="dv">{{ (sup.rating ?? 0).toFixed(1) }} 分</text>
      </view>
      <button class="btn-primary mt24" @tap="startEdit">编辑资料</button>
    </view>

    <!-- 编辑态 -->
    <view v-if="editing" class="card">
      <view class="card-title">编辑资料</view>
      <view class="edit-hint">修改资料后可能需要平台重新审核</view>
      <view class="form-item">
        <text class="form-label">商铺名称</text>
        <input v-model="form.name" class="form-input" placeholder="请输入商铺名称" />
      </view>
      <view class="form-item">
        <text class="form-label">营业执照/资质证明</text>
        <textarea v-model="form.license" class="form-textarea" placeholder="填写证照编号或说明" />
      </view>
      <view class="form-item">
        <text class="form-label">联系人</text>
        <input v-model="form.contact" class="form-input" placeholder="请输入联系人姓名" />
      </view>
      <view class="form-item">
        <text class="form-label">手机号</text>
        <input v-model="form.phone" class="form-input" type="number" maxlength="11" placeholder="请输入手机号" />
      </view>
      <view class="form-item">
        <text class="form-label">所在地区</text>
        <input v-model="form.region" class="form-input" placeholder="如：四川省成都市" />
      </view>
      <view class="form-item">
        <text class="form-label">详细地址</text>
        <input v-model="form.address" class="form-input" placeholder="请输入详细地址" />
      </view>
      <view class="form-item">
        <text class="form-label">经营品类（多选）</text>
        <view class="cats">
          <view
            v-for="c in INGREDIENT_CATEGORIES"
            :key="c"
            class="cat-chip"
            :class="{ on: categories.includes(c) }"
            @tap="toggleCategory(c)"
          >
            {{ c }}
          </view>
        </view>
      </view>
      <view class="btn-row">
        <button class="btn-plain half" @tap="editing = false">取消</button>
        <button class="btn-primary half" :disabled="saving" @tap="save">
          {{ saving ? '保存中...' : '保存' }}
        </button>
      </view>
    </view>

    <view class="card logout-card" @tap="logout">
      <text class="logout-text">退出登录</text>
    </view>

    <view class="bottom-space" />
    <TabBar current="profile" />
  </view>
</template>

<style scoped>
.head {
  background: #16a34a;
  display: flex;
  align-items: center;
  padding: 50rpx 40rpx;
}

.avatar {
  width: 120rpx;
  height: 120rpx;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.25);
  color: #ffffff;
  font-size: 48rpx;
  font-weight: 700;
  text-align: center;
  line-height: 120rpx;
  margin-right: 28rpx;
}

.head-name {
  color: #ffffff;
  font-size: 36rpx;
  font-weight: 700;
  display: flex;
  align-items: center;
}

.status-badge {
  font-size: 22rpx;
  font-weight: 400;
  padding: 4rpx 18rpx;
  border-radius: 20rpx;
  margin-left: 16rpx;
}

.status-badge.pass { background: #dcfce7; color: #15803d; }
.status-badge.wait { background: #fef3c7; color: #b45309; }
.status-badge.fail { background: #fee2e2; color: #b91c1c; }

.head-sub {
  color: rgba(255, 255, 255, 0.85);
  font-size: 24rpx;
  margin-top: 12rpx;
}

.audit-card {
  border: 2rpx solid #fecaca;
  background: #fef2f2;
}

.audit-label {
  display: block;
  font-size: 24rpx;
  color: #b91c1c;
  margin-bottom: 8rpx;
}

.audit-content {
  font-size: 26rpx;
  color: #7f1d1d;
  line-height: 1.6;
}

.detail-row {
  display: flex;
  margin-bottom: 22rpx;
}

.dl {
  width: 160rpx;
  color: #999999;
  font-size: 26rpx;
  flex-shrink: 0;
}

.dv {
  flex: 1;
  color: #262626;
  font-size: 28rpx;
}

.mt24 {
  margin-top: 24rpx;
}

.edit-hint {
  font-size: 24rpx;
  color: #b45309;
  background: #fef3c7;
  border-radius: 10rpx;
  padding: 14rpx 20rpx;
  margin-bottom: 24rpx;
}

.cats {
  display: flex;
  flex-wrap: wrap;
}

.cat-chip {
  padding: 12rpx 30rpx;
  background: #f7f7f8;
  border-radius: 30rpx;
  font-size: 26rpx;
  color: #595959;
  margin: 0 16rpx 16rpx 0;
  border: 2rpx solid transparent;
}

.cat-chip.on {
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

.logout-card {
  text-align: center;
}

.logout-text {
  color: #e6504f;
  font-size: 30rpx;
}

.bottom-space {
  height: 140rpx;
}
</style>
