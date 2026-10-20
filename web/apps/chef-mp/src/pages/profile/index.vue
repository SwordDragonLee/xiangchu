<script setup lang="ts">
import { onShow } from '@dcloudio/uni-app'
import { computed, reactive, ref } from 'vue'
import TabBar from '@/components/TabBar.vue'
import { chefApi, clearSession, errMsg, getToken, toast } from '@/utils/api'
import { CHEF_STATUS_TEXT, CUISINE_OPTIONS, splitCuisines } from '@/utils/format'
import type { Chef } from '@xc/shared'

const chef = ref<Chef | null>(null)
const loading = ref(false)
const saving = ref(false)
const editing = ref(false)

const form = reactive({
  real_name: '',
  phone: '',
  region: '',
  experience: '',
})
const cuisines = ref<string[]>([])

const statusText = computed(() => (chef.value ? CHEF_STATUS_TEXT[chef.value.status] || '未知' : ''))

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
    const p = await chefApi.profile()
    chef.value = p
    fillForm(p)
  } catch (e) {
    toast(errMsg(e))
  } finally {
    loading.value = false
  }
}

function fillForm(p: Chef) {
  form.real_name = p.real_name || ''
  form.phone = p.phone || ''
  form.region = p.region || ''
  form.experience = p.experience || ''
  cuisines.value = splitCuisines(p.cuisines)
}

function startEdit() {
  if (chef.value) fillForm(chef.value)
  editing.value = true
}

function toggleCuisine(c: string) {
  const i = cuisines.value.indexOf(c)
  if (i >= 0) cuisines.value.splice(i, 1)
  else cuisines.value.push(c)
}

async function save() {
  if (!form.real_name.trim()) {
    toast('请输入姓名')
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
  if (cuisines.value.length === 0) {
    toast('请至少选择一个擅长菜系')
    return
  }
  saving.value = true
  try {
    await chefApi.updateProfile({
      real_name: form.real_name.trim(),
      phone: form.phone,
      region: form.region.trim(),
      cuisines: cuisines.value.join(','),
      experience: form.experience.trim(),
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
    <view v-if="chef" class="head">
      <view class="avatar">{{ (chef.real_name || '厨').slice(0, 1) }}</view>
      <view class="head-info">
        <view class="head-name">
          {{ chef.real_name }}
          <text class="status-badge" :class="statusClass(chef.status)">{{ statusText }}</text>
        </view>
        <view class="head-sub">接单 {{ chef.order_count }} 次 · 评分 {{ (chef.rating ?? 0).toFixed(1) }}</view>
      </view>
    </view>

    <view v-if="chef && (chef.status === 2 || chef.status === 3)" class="card audit-card">
      <text class="audit-label">{{ chef.status === 3 ? '禁用原因' : '拒绝原因' }}</text>
      <text class="audit-content">{{ chef.audit_remark || '未填写' }}</text>
    </view>

    <view v-if="chef && !editing" class="card">
      <view class="detail-row"><text class="dl">姓名</text><text class="dv">{{ chef.real_name }}</text></view>
      <view class="detail-row"><text class="dl">手机号</text><text class="dv">{{ chef.phone }}</text></view>
      <view class="detail-row"><text class="dl">所在地区</text><text class="dv">{{ chef.region }}</text></view>
      <view class="detail-row">
        <text class="dl">擅长菜系</text>
        <text class="dv">{{ splitCuisines(chef.cuisines).join('、') || '暂无' }}</text>
      </view>
      <view class="detail-row"><text class="dl">从业经验</text><text class="dv">{{ chef.experience || '暂无' }}</text></view>
      <view class="detail-row">
        <text class="dl">入驻时间</text>
        <text class="dv">{{ (chef.created_at || '').replace('T', ' ').slice(0, 10) }}</text>
      </view>
      <button class="btn-primary mt24" @tap="startEdit">编辑档案</button>
    </view>

    <!-- 编辑态 -->
    <view v-if="editing" class="card">
      <view class="card-title">编辑档案</view>
      <view class="form-item">
        <text class="form-label">姓名</text>
        <input v-model="form.real_name" class="form-input" placeholder="请输入真实姓名" />
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
        <text class="form-label">擅长菜系（多选）</text>
        <view class="cuisines">
          <view
            v-for="c in CUISINE_OPTIONS"
            :key="c"
            class="cuisine-chip"
            :class="{ on: cuisines.includes(c) }"
            @tap="toggleCuisine(c)"
          >
            {{ c }}
          </view>
        </view>
      </view>
      <view class="form-item">
        <text class="form-label">从业经验</text>
        <textarea v-model="form.experience" class="form-textarea" placeholder="请填写从业经验" />
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
  background: #d97706;
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

.cuisines {
  display: flex;
  flex-wrap: wrap;
}

.cuisine-chip {
  padding: 12rpx 30rpx;
  background: #f7f7f8;
  border-radius: 30rpx;
  font-size: 26rpx;
  color: #595959;
  margin: 0 16rpx 16rpx 0;
  border: 2rpx solid transparent;
}

.cuisine-chip.on {
  background: #fffbeb;
  color: #d97706;
  border-color: #d97706;
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
