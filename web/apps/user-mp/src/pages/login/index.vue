<template>
  <view class="login-page page-pad">
    <view class="brand">
      <view class="brand-logo">乡厨</view>
      <view class="brand-name">乡厨到家</view>
      <view class="brand-slogan">地道乡厨 · 家宴上门</view>
    </view>

    <view class="card form">
      <view class="form-item">
        <view class="form-label">手机号</view>
        <input
          v-model="phone"
          class="form-input"
          type="number"
          maxlength="11"
          placeholder="请输入手机号"
        />
      </view>
      <view class="form-item">
        <view class="form-label">验证码</view>
        <input v-model="code" class="form-input" type="number" maxlength="6" placeholder="请输入验证码" />
        <view class="code-btn" :class="{ disabled: counting }" @click="sendCode">
          {{ counting ? `${seconds}s后重发` : '获取验证码' }}
        </view>
      </view>

      <button class="btn-primary submit" :class="{ disabled: submitting }" @click="doLogin">
        登录
      </button>

      <view class="dev-tip">开发环境验证码：123456</view>
    </view>

    <!-- #ifdef H5 -->
    <view class="wx-login" @click="wxLogin">
      <view class="wx-btn">微信一键登录</view>
    </view>
    <!-- #endif -->
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { authApi } from '@/utils/api'
import { useSession } from '@/stores/session'
import { errMsg } from '@/utils/format'

const phone = ref('')
const code = ref('')
const counting = ref(false)
const seconds = ref(60)
const submitting = ref(false)
const { save } = useSession()

function validPhone(): boolean {
  if (!/^1\d{10}$/.test(phone.value)) {
    uni.showToast({ title: '请输入正确的手机号', icon: 'none' })
    return false
  }
  return true
}

function sendCode() {
  if (counting.value) return
  if (!validPhone()) return
  authApi
    .sendCode(phone.value)
    .then(() => {
      uni.showToast({ title: '验证码已发送', icon: 'none' })
      counting.value = true
      seconds.value = 60
      const timer = setInterval(() => {
        seconds.value -= 1
        if (seconds.value <= 0) {
          clearInterval(timer)
          counting.value = false
        }
      }, 1000)
    })
    .catch((e: unknown) => uni.showToast({ title: errMsg(e), icon: 'none' }))
}

function doLogin() {
  if (submitting.value) return
  if (!validPhone()) return
  if (!code.value) {
    uni.showToast({ title: '请输入验证码', icon: 'none' })
    return
  }
  submitting.value = true
  authApi
    .loginByCode(phone.value, code.value)
    .then((res) => {
      save(res)
      uni.showToast({ title: '登录成功', icon: 'success' })
      setTimeout(() => uni.reLaunch({ url: '/pages/home/index' }), 600)
    })
    .catch((e: unknown) => uni.showToast({ title: errMsg(e), icon: 'none' }))
    .finally(() => (submitting.value = false))
}

// H5 端：无法真正调起 wx.login，用 mock code 走微信登录接口
function wxLogin() {
  const mockCode = `h5_${Date.now()}`
  authApi
    .wxLogin(mockCode)
    .then((res) => {
      save(res)
      uni.showToast({ title: '登录成功', icon: 'success' })
      setTimeout(() => uni.reLaunch({ url: '/pages/home/index' }), 600)
    })
    .catch((e: unknown) => uni.showToast({ title: errMsg(e), icon: 'none' }))
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  box-sizing: border-box;
}
.brand {
  text-align: center;
  padding: 100rpx 0 60rpx;
}
.brand-logo {
  width: 140rpx;
  height: 140rpx;
  line-height: 140rpx;
  margin: 0 auto 24rpx;
  border-radius: 32rpx;
  background: #d97706;
  color: #fff;
  font-size: 48rpx;
  font-weight: 700;
}
.brand-name {
  font-size: 40rpx;
  font-weight: 700;
}
.brand-slogan {
  color: #999;
  font-size: 26rpx;
  margin-top: 12rpx;
}
.submit {
  margin-top: 48rpx;
}
.dev-tip {
  text-align: center;
  color: #bbb;
  font-size: 24rpx;
  margin-top: 24rpx;
}
.code-btn {
  flex-shrink: 0;
  color: #d97706;
  font-size: 26rpx;
}
.code-btn.disabled {
  color: #bbb;
}
.wx-login {
  margin-top: 48rpx;
  text-align: center;
}
.wx-btn {
  color: #059669;
  font-size: 28rpx;
  padding: 16rpx;
}
</style>
