<script setup lang="ts">
import { onLoad } from '@dcloudio/uni-app'
import { ref } from 'vue'
import { authApi, clearSession, errMsg, getRole, getToken, saveSession, supplierApi, toast } from '@/utils/api'
import { INGREDIENT_CATEGORIES, splitCategories } from '@/utils/format'
import type { Supplier } from '@xc/shared'

// status: 0待审核 1已通过 2已拒绝 3已禁用
const applied = ref(false)
const supStatus = ref<number | null>(null)
const auditRemark = ref('')
const submitting = ref(false)
const loaded = ref(false)

const form = ref({
  name: '',
  license: '',
  contact: '',
  phone: '',
  region: '',
  address: '',
})
const categories = ref<string[]>([])

onLoad(() => {
  if (!getToken()) {
    uni.reLaunch({ url: '/pages/login/index' })
    return
  }
  loadProfile()
})

async function loadProfile() {
  try {
    // apply-status 对待审核/被驳回用户同样可见；/supplier/profile 需审核通过后的 supplier 角色
    const p = await supplierApi.applyStatus()
    applied.value = true
    supStatus.value = p.status
    auditRemark.value = p.audit_remark || ''
    fillForm(p)
  } catch {
    // 尚未提交过入驻申请，显示空白表单
    applied.value = false
  } finally {
    loaded.value = true
  }
}

function fillForm(p: Supplier) {
  form.value.name = p.name || ''
  form.value.license = p.license || ''
  form.value.contact = p.contact || ''
  form.value.phone = p.phone || ''
  form.value.region = p.region || ''
  form.value.address = p.address || ''
  categories.value = splitCategories(p.categories)
}

function toggleCategory(c: string) {
  const i = categories.value.indexOf(c)
  if (i >= 0) categories.value.splice(i, 1)
  else categories.value.push(c)
}

function validate(): boolean {
  if (!form.value.name.trim()) {
    toast('请输入商铺名称')
    return false
  }
  if (!form.value.license.trim()) {
    toast('请填写营业执照/资质证明')
    return false
  }
  if (!form.value.contact.trim()) {
    toast('请输入联系人')
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
  if (!form.value.address.trim()) {
    toast('请输入详细地址')
    return false
  }
  if (categories.value.length === 0) {
    toast('请至少选择一个经营品类')
    return false
  }
  return true
}

async function submit() {
  if (!validate()) return
  submitting.value = true
  try {
    await supplierApi.apply({
      name: form.value.name.trim(),
      license: form.value.license.trim(),
      contact: form.value.contact.trim(),
      phone: form.value.phone,
      region: form.value.region.trim(),
      address: form.value.address.trim(),
      categories: categories.value.join(','),
    })
    uni.showModal({
      title: '提交成功',
      content: '您的入驻申请已提交，请等待平台审核',
      showCancel: false,
      success: () => {
        if (getRole() === 'supplier') {
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
  if (getRole() === 'supplier') {
    uni.reLaunch({ url: '/pages/workbench/index' })
    return
  }
  // 审核通过但本地角色仍是登录时的旧值（或双身份账号）：凭 token 换发 supplier 角色的 JWT
  try {
    const res = await authApi.refresh('supplier')
    saveSession(res.token, res.role)
    if (res.role === 'supplier') {
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
      <view v-if="applied && supStatus === 0" class="card status-card">
        <view class="status-icon wait">审</view>
        <view class="status-title">入驻申请审核中</view>
        <view class="status-desc">您的资料已提交，平台将在 1-3 个工作日内完成审核</view>
        <view class="status-badge wait-badge">待审核</view>
      </view>

      <view v-else-if="applied && supStatus === 1" class="card status-card">
        <view class="status-icon pass">过</view>
        <view class="status-title">恭喜，您已通过认证</view>
        <view class="status-desc">您可以上架商品、承接采购订单了</view>
        <view class="status-badge pass-badge">已通过</view>
        <button class="btn-primary mt24" @tap="goWorkbench">进入工作台</button>
      </view>

      <view v-else-if="applied" class="card status-card">
        <view class="status-icon fail">拒</view>
        <view class="status-title">{{ supStatus === 3 ? '账号已被禁用' : '入驻申请未通过' }}</view>
        <view class="status-desc">原因：{{ auditRemark || '未填写' }}</view>
        <view class="status-badge fail-badge">{{ supStatus === 3 ? '已禁用' : '已拒绝' }}</view>
        <button v-if="supStatus === 2" class="btn-plain mt24" @tap="applied = false">
          修改资料重新申请
        </button>
      </view>

      <!-- 申请表单（未申请 / 重新申请） -->
      <view v-else class="card">
        <view class="card-title">入驻申请</view>
        <view class="form-item">
          <text class="form-label">商铺名称</text>
          <input v-model="form.name" class="form-input" placeholder="如：老王家蔬菜批发" />
        </view>
        <view class="form-item">
          <text class="form-label">营业执照/资质证明</text>
          <textarea
            v-model="form.license"
            class="form-textarea"
            placeholder="填写证照编号或说明，如：统一社会信用代码 9251XXXXXXXXXXXX（图片上传即将开放）"
          />
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
          <input v-model="form.address" class="form-input" placeholder="如：XX镇农贸市场 12 号摊位" />
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
</style>
