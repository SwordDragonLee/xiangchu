<script setup lang="ts">
import { ref } from 'vue'
import { authApi, errMsg, getRole, getToken, saveSession, toast } from '@/utils/api'

const phone = ref('')
const code = ref('')
const sending = ref(false)
const countdown = ref(0)
const logging = ref(false)

// 已登录直接进对应页
if (getToken()) {
  const role = getRole()
  if (role === 'supplier') {
    uni.reLaunch({ url: '/pages/workbench/index' })
  } else if (role) {
    uni.reLaunch({ url: '/pages/apply/index' })
  }
}

function fallbackToApply(): void {
  uni.reLaunch({ url: '/pages/apply/index' })
}

async function sendCode() {
  if (!/^1\d{10}$/.test(phone.value)) {
    toast('请输入正确的手机号')
    return
  }
  if (countdown.value > 0 || sending.value) return
  sending.value = true
  try {
    await authApi.sendCode(phone.value)
    toast('验证码已发送（开发环境固定 123456）')
    countdown.value = 60
    const timer = setInterval(() => {
      countdown.value -= 1
      if (countdown.value <= 0) clearInterval(timer)
    }, 1000)
  } catch (e) {
    toast(errMsg(e))
  } finally {
    sending.value = false
  }
}

async function doLogin() {
  if (!/^1\d{10}$/.test(phone.value)) {
    toast('请输入正确的手机号')
    return
  }
  if (!/^\d{4,6}$/.test(code.value)) {
    toast('请输入验证码')
    return
  }
  logging.value = true
  try {
    const res = await authApi.loginByCode(phone.value, code.value)
    saveSession(res.token, res.role)
    if (res.role !== 'supplier') {
      // 双身份账号（如既是认证乡厨又是供应商）登录拿到的是另一角色，按 supplier 档案换发重试
      try {
        const r = await authApi.refresh('supplier')
        saveSession(r.token, r.role)
      } catch {
        /* 换发失败按未认证处理 */
      }
      if (getRole() === 'supplier') {
        uni.reLaunch({ url: '/pages/workbench/index' })
        return
      }
      uni.showModal({
        title: '提示',
        content: '您还不是认证供应商，请先提交入驻申请',
        showCancel: false,
        confirmText: '去申请',
        success: fallbackToApply,
      })
    } else {
      uni.reLaunch({ url: '/pages/workbench/index' })
    }
  } catch (e) {
    toast(errMsg(e))
  } finally {
    logging.value = false
  }
}
</script>

<template>
  <view class="login-page">
    <view class="logo">
      <view class="logo-badge">供</view>
      <view class="logo-title">供应商工作台</view>
      <view class="logo-sub">食材供应 · 采购履约 · 收入提现</view>
    </view>

    <view class="card">
      <view class="form-item">
        <text class="form-label">手机号</text>
        <input
          v-model="phone"
          class="form-input"
          type="number"
          maxlength="11"
          placeholder="请输入手机号"
        />
      </view>
      <view class="form-item">
        <text class="form-label">验证码</text>
        <view class="code-row">
          <input
            v-model="code"
            class="form-input code-input"
            type="number"
            maxlength="6"
            placeholder="请输入验证码"
          />
          <view class="code-btn" :class="{ disabled: countdown > 0 || sending }" @tap="sendCode">
            {{ countdown > 0 ? `${countdown}s后重发` : '获取验证码' }}
          </view>
        </view>
      </view>
      <button class="btn-primary" :disabled="logging" @tap="doLogin">
        {{ logging ? '登录中...' : '登录' }}
      </button>
      <view class="dev-hint">开发环境验证码：123456</view>
    </view>
  </view>
</template>

<style scoped>
.login-page {
  min-height: 100vh;
  padding: 100rpx 0 60rpx;
}

.logo {
  text-align: center;
  margin-bottom: 60rpx;
}

.logo-badge {
  width: 140rpx;
  height: 140rpx;
  margin: 0 auto 24rpx;
  border-radius: 32rpx;
  background: #16a34a;
  color: #ffffff;
  font-size: 48rpx;
  font-weight: 700;
  line-height: 140rpx;
}

.logo-title {
  font-size: 40rpx;
  font-weight: 700;
  color: #262626;
}

.logo-sub {
  font-size: 26rpx;
  color: #999999;
  margin-top: 12rpx;
}

.code-row {
  display: flex;
  align-items: center;
}

.code-input {
  flex: 1;
}

.code-btn {
  margin-left: 20rpx;
  background: #f0fdf4;
  color: #16a34a;
  border: 2rpx solid #16a34a;
  border-radius: 12rpx;
  font-size: 26rpx;
  padding: 18rpx 24rpx;
  white-space: nowrap;
}

.code-btn.disabled {
  color: #999999;
  border-color: #dddddd;
  background: #f7f7f8;
}

.dev-hint {
  text-align: center;
  color: #bbbbbb;
  font-size: 24rpx;
  margin-top: 30rpx;
}
</style>
