<template>
  <view class="profile">
    <!-- 用户卡片 -->
    <view class="user-head">
      <view class="avatar">{{ avatarText }}</view>
      <view class="grow user-info">
        <template v-if="isLogged">
          <view class="nickname">{{ profile?.nickname || '乡厨用户' }}</view>
          <view class="muted phone">{{ profile?.phone || '' }}</view>
        </template>
        <template v-else>
          <view class="nickname" @click="goLogin">点击登录 ›</view>
          <view class="muted phone">登录后可预订乡厨</view>
        </template>
      </view>
    </view>

    <!-- 昵称编辑 -->
    <view v-if="isLogged && editing" class="card page-pad">
      <view class="form-item">
        <view class="form-label">新昵称</view>
        <input v-model="nicknameDraft" class="form-input" placeholder="请输入昵称" />
      </view>
      <view class="row edit-actions">
        <view class="btn-plain btn-mini" @click="editing = false">取消</view>
        <view class="btn-primary btn-mini" @click="saveNickname">保存</view>
      </view>
    </view>

    <!-- 菜单 -->
    <view class="card menu">
      <view v-if="isLogged" class="menu-item" @click="startEditNickname">
        <text>资料（昵称 / 头像）</text>
        <text class="arrow">›</text>
      </view>
      <view class="menu-item" @click="go('/pages/address/index')">
        <text>我的地址</text>
        <text class="arrow">›</text>
      </view>
      <view class="menu-item" @click="go('/pages/coupon/index?tab=mine')">
        <text>我的优惠券</text>
        <text class="arrow">›</text>
      </view>
      <view class="menu-item" @click="goOrder">
        <text>我的订单</text>
        <text class="arrow">›</text>
      </view>
      <view v-if="isLogged" class="menu-item" @click="go('/pages/apply/index')">
        <text>成为乡厨 / 供应商</text>
        <text class="arrow">›</text>
      </view>
      <view class="menu-item" @click="about">
        <text>关于</text>
        <text class="arrow">›</text>
      </view>
    </view>

    <view v-if="isLogged" class="page-pad">
      <button class="btn-plain logout" @click="logout">退出登录</button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import type { ProfileVO } from '@xc/shared/types'
import { userApi } from '@/utils/api'
import { errMsg } from '@/utils/format'
import { useSession } from '@/stores/session'

const { isLogged, state, clear } = useSession()
const profile = ref<ProfileVO | null>(null)
const editing = ref(false)
const nicknameDraft = ref('')

const avatarText = computed(() => {
  const n = profile.value?.nickname || ''
  return n ? n.slice(0, 1) : '客'
})

function loadProfile() {
  if (!isLogged.value) {
    profile.value = null
    return
  }
  userApi
    .profile()
    .then((res) => (profile.value = res))
    .catch(() => {
      /* 401 时由 request 拦截统一处理 */
    })
}

onShow(() => loadProfile())

function go(url: string) {
  uni.navigateTo({ url })
}
function goLogin() {
  uni.navigateTo({ url: '/pages/login/index' })
}
function goOrder() {
  uni.switchTab({ url: '/pages/order/list' })
}

function startEditNickname() {
  nicknameDraft.value = profile.value?.nickname || ''
  editing.value = true
}

function saveNickname() {
  const name = nicknameDraft.value.trim()
  if (!name) {
    uni.showToast({ title: '请输入昵称', icon: 'none' })
    return
  }
  userApi
    .updateProfile({ nickname: name })
    .then(() => {
      uni.showToast({ title: '已保存', icon: 'success' })
      editing.value = false
      loadProfile()
    })
    .catch((e: unknown) => uni.showToast({ title: errMsg(e), icon: 'none' }))
}

function about() {
  uni.showModal({
    title: '关于乡厨到家',
    content: '乡厨到家用户端 v1.0.0\n地道乡厨 · 家宴上门',
    showCancel: false,
  })
}

function logout() {
  uni.showModal({
    title: '退出登录',
    content: '确认退出当前账号？',
    success: (r) => {
      if (!r.confirm) return
      clear()
      profile.value = null
      uni.showToast({ title: '已退出', icon: 'none' })
    },
  })
}
</script>

<style scoped>
.user-head {
  display: flex;
  align-items: center;
  padding: 60rpx 40rpx;
  background: linear-gradient(135deg, #d97706, #f59e0b);
}
.avatar {
  width: 120rpx;
  height: 120rpx;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.9);
  color: #d97706;
  font-size: 48rpx;
  font-weight: 700;
  text-align: center;
  line-height: 120rpx;
  margin-right: 24rpx;
  flex-shrink: 0;
}
.user-info {
  color: #fff;
}
.nickname {
  font-size: 36rpx;
  font-weight: 700;
}
.phone {
  color: rgba(255, 255, 255, 0.85) !important;
  margin-top: 8rpx;
}
.menu {
  padding: 0 24rpx;
}
.menu-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 30rpx 0;
  border-bottom: 1rpx solid #f5f5f5;
  font-size: 28rpx;
}
.menu-item:last-child {
  border-bottom: none;
}
.arrow {
  color: #ccc;
}
.edit-actions {
  gap: 20rpx;
  margin-top: 20rpx;
}
.logout {
  margin-top: 8rpx;
}
</style>
