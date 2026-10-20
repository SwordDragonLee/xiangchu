<script setup lang="ts">
import { onLoad } from '@dcloudio/uni-app'
import { ref } from 'vue'
import { authApi, chefApi, clearSession, errMsg, getRole, getToken, saveSession, toast } from '@/utils/api'
import { CUISINE_OPTIONS, splitCuisines } from '@/utils/format'
import type { Chef } from '@xc/shared'

// status: 0待审核 1已通过 2已拒绝 3已禁用
const applied = ref(false)
const chefStatus = ref<number | null>(null)
const auditRemark = ref('')
const submitting = ref(false)
const loaded = ref(false)

const form = ref({
  real_name: '',
  phone: '',
  region: '',
  experience: '',
})
const cuisines = ref<string[]>([])

onLoad(() => {
  if (!getToken()) {
    uni.reLaunch({ url: '/pages/login/index' })
    return
  }
  loadProfile()
})

async function loadProfile() {
  try {
    // apply-status 对待审核/被驳回用户同样可见；/chef/profile 需审核通过后的 chef 角色
    const p = await chefApi.applyStatus()
    applied.value = true
    chefStatus.value = p.status
    auditRemark.value = p.audit_remark || ''
    form.value.real_name = p.real_name || ''
    form.value.phone = p.phone || ''
    form.value.region = p.region || ''
    form.value.experience = p.experience || ''
    cuisines.value = splitCuisines(p.cuisines)
  } catch {
    // 尚未提交过入驻申请，显示空白表单
    applied.value = false
  } finally {
    loaded.value = true
  }
}

function toggleCuisine(c: string) {
  const i = cuisines.value.indexOf(c)
  if (i >= 0) cuisines.value.splice(i, 1)
  else cuisines.value.push(c)
}

function validate(): boolean {
  if (!form.value.real_name.trim()) {
    toast('请输入姓名')
    return false
  }
  if (!/^1\d{10}$/.test(form.value.phone)) {
    toast('请输入正确的手机号')
    return false
  }
  if (!form.value.region.trim()) {
    toast('请输入所在地区')
    return false
  }
  if (cuisines.value.length === 0) {
    toast('请至少选择一个擅长菜系')
    return false
  }
  if (!form.value.experience.trim()) {
    toast('请填写从业经验')
    return false
  }
  return true
}

async function submit() {
  if (!validate()) return
  submitting.value = true
  try {
    await chefApi.apply({
      real_name: form.value.real_name.trim(),
      phone: form.value.phone,
      region: form.value.region.trim(),
      cuisines: cuisines.value.join(','),
      experience: form.value.experience.trim(),
    })
    uni.showModal({
      title: '提交成功',
      content: '您的入驻申请已提交，请等待平台审核',
      showCancel: false,
      success: () => {
        if (getRole() === 'chef') {
          uni.reLaunch({ url: '/pages/workbench/index' })
        } else {
          loadProfile()
        }
      },
    })
  } catch (e) {
    toast(errMsg(e))
  } finally {
    submitting.value = false
  }
}

async function goWorkbench() {
  if (getRole() === 'chef') {
    uni.reLaunch({ url: '/pages/workbench/index' })
    return
  }
  // 审核通过但本地角色仍是登录时的旧值：凭 token 换发含最新角色的新 JWT
  try {
    const res = await authApi.refresh()
    saveSession(res.token, res.role)
    if (res.role === 'chef') {
      uni.reLaunch({ url: '/pages/workbench/index' })
    } else {
      toast('身份尚未生效，请重新登录后重试')
    }
  } catch (e) {
    toast(errMsg(e))
  }
}

// 入驻页被 reLaunch 清空了页面栈（无返回键、无 TabBar），未认证账号唯一出口就是这里
function logout() {
  uni.showModal({
    title: '退出登录',
    content: '退出后可更换手机号重新登录',
    success: (r) => {
      if (r.confirm) {
        clearSession()
        uni.reLaunch({ url: '/pages/login/index' })
      }
    },
  })
}
</script>

<template>
  <view class="page">
    <template v-if="loaded">
      <!-- 已提交过申请：展示当前状态 -->
      <view v-if="applied && chefStatus === 0" class="card status-card">
        <view class="status-icon wait">审</view>
        <view class="status-title">入驻申请审核中</view>
        <view class="status-desc">您的资料已提交，平台将在 1-3 个工作日内完成审核</view>
        <view class="status-badge wait-badge">待审核</view>
      </view>

      <view v-else-if="applied && chefStatus === 1" class="card status-card">
        <view class="status-icon pass">过</view>
        <view class="status-title">恭喜，您已通过认证</view>
        <view class="status-desc">您可以正常接单了，请完善档案保持信息最新</view>
        <view class="status-badge pass-badge">已通过</view>
        <button class="btn-primary mt24" @tap="goWorkbench">进入工作台</button>
      </view>

      <view v-else-if="applied" class="card status-card">
        <view class="status-icon fail">拒</view>
        <view class="status-title">{{ chefStatus === 3 ? '账号已被禁用' : '入驻申请未通过' }}</view>
        <view class="status-desc">原因：{{ auditRemark || '未填写' }}</view>
        <view class="status-badge fail-badge">{{ chefStatus === 3 ? '已禁用' : '已拒绝' }}</view>
        <button v-if="chefStatus === 2" class="btn-plain mt24" @tap="applied = false">
          修改资料重新申请
        </button>
      </view>

      <!-- 申请表单（未申请 / 重新申请） -->
      <view v-else class="card">
        <view class="card-title">入驻申请</view>
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
          <textarea
            v-model="form.experience"
            class="form-textarea"
            placeholder="如：从业 15 年，擅长农村坝坝宴，可独立操办 30 桌以上宴席"
          />
        </view>
        <button class="btn-primary" :disabled="submitting" @tap="submit">
          {{ submitting ? '提交中...' : '提交申请' }}
        </button>
      </view>
    </template>
    <!-- 未认证账号在入驻页没有 TabBar/返回键，退出登录是唯一出口，放 loaded 外保证任何状态下可用 -->
    <view class="logout-link" @tap="logout">退出登录</view>
  </view>
</template>

<style scoped>
.page {
  min-height: 100vh;
  padding-bottom: 60rpx;
}

.logout-link {
  text-align: center;
  color: #999999;
  font-size: 26rpx;
  padding: 40rpx 0 20rpx;
}

.status-card {
  text-align: center;
  padding: 60rpx 40rpx;
}

.status-icon {
  width: 110rpx;
  height: 110rpx;
  margin: 0 auto 24rpx;
  border-radius: 50%;
  color: #ffffff;
  font-size: 44rpx;
  line-height: 110rpx;
  font-weight: 700;
}

.status-icon.wait { background: #f59e0b; }
.status-icon.pass { background: #16a34a; }
.status-icon.fail { background: #e6504f; }

.status-title {
  font-size: 34rpx;
  font-weight: 700;
  color: #262626;
}

.status-desc {
  font-size: 26rpx;
  color: #999999;
  margin-top: 16rpx;
  line-height: 1.6;
}

.status-badge {
  display: inline-block;
  margin-top: 24rpx;
  font-size: 24rpx;
  padding: 8rpx 28rpx;
  border-radius: 24rpx;
}

.wait-badge { background: #fef3c7; color: #b45309; }
.pass-badge { background: #dcfce7; color: #15803d; }
.fail-badge { background: #fee2e2; color: #b91c1c; }

.mt24 {
  margin-top: 32rpx;
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
</style>
