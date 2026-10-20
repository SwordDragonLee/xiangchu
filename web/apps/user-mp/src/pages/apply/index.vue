<template>
  <view class="page-pad apply">
    <!-- 未登录提示 -->
    <view v-if="!isLogged" class="card center-box">
      <view class="muted">请先登录后再提交入驻申请</view>
      <button class="btn-primary login-btn" @click="goLogin">去登录</button>
    </view>

    <template v-else>
      <!-- Web 登录密码（审核通过后凭手机号+密码登录供应商/乡厨后台） -->
      <view class="card">
        <view class="section-title">后台登录密码</view>
        <view class="muted tip">审核通过后，凭手机号 + 此密码登录供应商后台（Web）</view>
        <view class="form-item">
          <view class="form-label">登录密码</view>
          <input v-model="pwd" class="form-input" password placeholder="6-32 位" />
        </view>
        <view class="form-item">
          <view class="form-label">确认密码</view>
          <input v-model="pwd2" class="form-input" password placeholder="再次输入" />
        </view>
        <button class="btn-primary" :disabled="savingPwd" @click="savePwd">
          {{ savingPwd ? '保存中...' : '保存密码' }}
        </button>
      </view>

      <!-- 入驻类型 -->
      <view class="card">
        <view class="section-title">申请类型</view>
        <view class="row type-row">
          <view class="type-item" :class="{ active: roleTab === 'chef' }" @click="roleTab = 'chef'">成为乡厨</view>
          <view class="type-item" :class="{ active: roleTab === 'supplier' }" @click="roleTab = 'supplier'">成为供应商</view>
        </view>
      </view>

      <!-- 乡厨入驻表单 -->
      <view v-if="roleTab === 'chef'" class="card">
        <view class="form-item">
          <view class="form-label">真实姓名</view>
          <input v-model="chef.real_name" class="form-input" placeholder="请输入真实姓名" />
        </view>
        <view class="form-item">
          <view class="form-label">联系电话</view>
          <input v-model="chef.phone" class="form-input" type="number" placeholder="请输入手机号" />
        </view>
        <view class="form-item">
          <view class="form-label">所在地区</view>
          <input v-model="chef.region" class="form-input" placeholder="如：华中 / 河南省信阳市" />
        </view>
        <view class="form-item">
          <view class="form-label">擅长菜系</view>
          <input v-model="chef.cuisines" class="form-input" placeholder="如：豫菜,农家菜（逗号分隔）" />
        </view>
        <view class="form-item">
          <view class="form-label">从业经验</view>
          <textarea v-model="chef.experience" class="form-textarea" placeholder="从业年限、拿手宴席、团队规模等" />
        </view>
        <button class="btn-primary" :disabled="submitting" @click="submitChef">
          {{ submitting ? '提交中...' : '提交乡厨申请' }}
        </button>
      </view>

      <!-- 供应商入驻表单 -->
      <view v-else class="card">
        <view class="form-item">
          <view class="form-label">店铺名称</view>
          <input v-model="sup.name" class="form-input" placeholder="如：老王食材铺" />
        </view>
        <view class="form-item">
          <view class="form-label">资质证照</view>
          <input v-model="sup.license" class="form-input" placeholder="营业执照编号 / 图片链接" />
        </view>
        <view class="form-item">
          <view class="form-label">联系人</view>
          <input v-model="sup.contact" class="form-input" placeholder="请输入联系人姓名" />
        </view>
        <view class="form-item">
          <view class="form-label">联系电话</view>
          <input v-model="sup.phone" class="form-input" type="number" placeholder="请输入手机号" />
        </view>
        <view class="form-item">
          <view class="form-label">所在地区</view>
          <input v-model="sup.region" class="form-input" placeholder="如：华中 / 河南省信阳市" />
        </view>
        <view class="form-item">
          <view class="form-label">详细地址</view>
          <input v-model="sup.address" class="form-input" placeholder="店铺 / 仓库地址" />
        </view>
        <view class="form-item">
          <view class="form-label">经营品类</view>
          <input v-model="sup.categories" class="form-input" placeholder="如：肉类,蔬菜（逗号分隔）" />
        </view>
        <button class="btn-primary" :disabled="submitting" @click="submitSupplier">
          {{ submitting ? '提交中...' : '提交供应商申请' }}
        </button>
      </view>

      <view class="muted flow-tip">提交后由平台审核，审核结果将通过通知告知；通过后重新登录即可进入对应工作台。</view>
    </template>
  </view>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { userApi, chefApi, supplierApi } from '@/utils/api'
import { errMsg } from '@/utils/format'
import { useSession } from '@/stores/session'

const { isLogged } = useSession()

const roleTab = ref<'chef' | 'supplier'>('chef')
const submitting = ref(false)
const savingPwd = ref(false)
const pwd = ref('')
const pwd2 = ref('')

const chef = reactive({ real_name: '', phone: '', region: '', cuisines: '', experience: '' })
const sup = reactive({ name: '', license: '', contact: '', phone: '', region: '', address: '', categories: '' })

onShow(() => {
  if (!isLogged.value) return
  // 拉取手机号作为默认联系电话
  userApi
    .profile()
    .then((p) => {
      if (!chef.phone) chef.phone = p.phone || ''
      if (!sup.phone) sup.phone = p.phone || ''
    })
    .catch(() => undefined)
})

function goLogin() {
  uni.navigateTo({ url: '/pages/login/index' })
}

function savePwd() {
  if (pwd.value.length < 6) {
    uni.showToast({ title: '密码至少 6 位', icon: 'none' })
    return
  }
  if (pwd.value !== pwd2.value) {
    uni.showToast({ title: '两次密码不一致', icon: 'none' })
    return
  }
  savingPwd.value = true
  userApi
    .setPassword(pwd.value)
    .then(() => {
      uni.showToast({ title: '密码已保存', icon: 'success' })
      pwd.value = ''
      pwd2.value = ''
    })
    .catch((e: unknown) => uni.showToast({ title: errMsg(e), icon: 'none' }))
    .finally(() => (savingPwd.value = false))
}

function submitChef() {
  if (!chef.real_name || !chef.phone || !chef.region || !chef.cuisines) {
    uni.showToast({ title: '请填写完整必填信息', icon: 'none' })
    return
  }
  submitting.value = true
  chefApi
    .apply({ ...chef })
    .then(() => done('chef'))
    .catch((e: unknown) => {
      uni.showToast({ title: errMsg(e), icon: 'none' })
      submitting.value = false
    })
}

function submitSupplier() {
  if (!sup.name || !sup.license || !sup.contact || !sup.phone || !sup.address) {
    uni.showToast({ title: '请填写完整必填信息', icon: 'none' })
    return
  }
  submitting.value = true
  supplierApi
    .apply({ ...sup })
    .then(() => done('supplier'))
    .catch((e: unknown) => {
      uni.showToast({ title: errMsg(e), icon: 'none' })
      submitting.value = false
    })
}

function done(role: 'chef' | 'supplier') {
  submitting.value = false
  const tip =
    role === 'chef'
      ? '审核通过后，即可在乡厨端（小程序 / H5）登录接单。'
      : '审核通过后，使用手机号 + 密码登录供应商后台（Web）即可开始工作。'
  uni.showModal({
    title: '提交成功',
    content: `已提交，等待平台审核。${tip}`,
    showCancel: false,
    confirmText: '知道了',
    success: () => leavePage(),
  })
}

function leavePage() {
  uni.navigateBack({ fail: () => uni.switchTab({ url: '/pages/profile/index' }) })
}
</script>

<style scoped>
.center-box {
  text-align: center;
  padding: 60rpx 40rpx;
}
.login-btn {
  margin-top: 30rpx;
}
.section-title {
  font-size: 30rpx;
  font-weight: 600;
  margin-bottom: 12rpx;
}
.tip {
  font-size: 24rpx;
  margin-bottom: 16rpx;
}
.type-row {
  gap: 20rpx;
  margin-top: 8rpx;
}
.type-item {
  flex: 1;
  text-align: center;
  padding: 18rpx 0;
  border: 2rpx solid #e5e7eb;
  border-radius: 12rpx;
  font-size: 28rpx;
  color: #6b7280;
}
.type-item.active {
  border-color: #d97706;
  color: #d97706;
  background: #fffbeb;
  font-weight: 600;
}
.form-textarea {
  width: 100%;
  min-height: 140rpx;
  padding: 16rpx 20rpx;
  border: 2rpx solid #e5e7eb;
  border-radius: 12rpx;
  font-size: 28rpx;
  box-sizing: border-box;
}
.flow-tip {
  font-size: 24rpx;
  text-align: center;
  padding: 24rpx 40rpx 60rpx;
}
</style>
