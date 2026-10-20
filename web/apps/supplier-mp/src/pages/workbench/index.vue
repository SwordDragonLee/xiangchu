<script setup lang="ts">
import { onShow } from '@dcloudio/uni-app'
import { computed, ref } from 'vue'
import TabBar from '@/components/TabBar.vue'
import { ensureLogin, ensureSupplier, errMsg, payApi, supplierApi, toast } from '@/utils/api'
import { fen2yuan } from '@/utils/format'

const available = ref(0)
const frozen = ref(0)
const monthNet = ref(0)
const totalNet = ref(0)
const pendingCount = ref(0)
const supName = ref('')
const loading = ref(false)

const entries = [
  { text: '商品管理', desc: '食材上架维护', url: '/pages/goods/index', cls: 'e-green' },
  { text: '采购履约', desc: '接单备货发货', url: '/pages/purchases/index', cls: 'e-blue' },
  { text: '收入提现', desc: '余额与提现', url: '/pages/income/index', cls: 'e-gold' },
  { text: '我的资料', desc: '资料与审核状态', url: '/pages/profile/index', cls: 'e-purple' },
  { text: '入驻申请', desc: '认证进度查询', url: '/pages/apply/index', cls: 'e-grey' },
]

const greeting = computed(() => {
  const h = new Date().getHours()
  if (h < 6) return '夜深了'
  if (h < 12) return '上午好'
  if (h < 14) return '中午好'
  if (h < 18) return '下午好'
  return '晚上好'
})

onShow(async () => {
  if (!ensureLogin()) return
  if (!(await ensureSupplier())) return
  load()
})

async function load() {
  loading.value = true
  try {
    const income = payApi.supplierIncome().then((i) => {
      available.value = i.available
      frozen.value = i.frozen_amount
      monthNet.value = i.month_net
      totalNet.value = i.total_net
    })
    const pending = supplierApi
      .purchases({ status: 0, page: 1, size: 1 })
      .then((r) => {
        pendingCount.value = r.total
      })
    const profile = supplierApi
      .profile()
      .then((p) => {
        supName.value = p.name
      })
      .catch(() => {
        /* 档案获取失败不影响工作台 */
      })
    await Promise.all([income, pending, profile])
  } catch (e) {
    toast(errMsg(e))
  } finally {
    loading.value = false
  }
}

function go(url: string) {
  uni.navigateTo({ url })
}

function goPurchases() {
  uni.reLaunch({ url: '/pages/purchases/index' })
}
</script>

<template>
  <view class="page">
    <view class="header">
      <view class="hello">
        <text class="hello-name">{{ greeting }}，{{ supName || '供应商' }}</text>
        <text class="hello-sub">食材新鲜口碑好，生意越做越红火</text>
      </view>
    </view>

    <view class="stats card">
      <view class="stat" @tap="go('/pages/income/index')">
        <view class="stat-num">
          <text class="rmb">¥</text>{{ fen2yuan(available) }}
        </view>
        <view class="stat-label">可提余额(元)</view>
      </view>
      <view class="stat" @tap="go('/pages/income/index')">
        <view class="stat-num">
          <text class="rmb">¥</text>{{ fen2yuan(frozen) }}
        </view>
        <view class="stat-label">冻结中(元)</view>
      </view>
      <view class="stat">
        <view class="stat-num">
          <text class="rmb">¥</text>{{ fen2yuan(monthNet) }}
        </view>
        <view class="stat-label">本月结算(元)</view>
      </view>
      <view class="stat">
        <view class="stat-num">
          <text class="rmb">¥</text>{{ fen2yuan(totalNet) }}
        </view>
        <view class="stat-label">累计结算(元)</view>
      </view>
    </view>

    <view v-if="pendingCount > 0" class="card pending-card" @tap="goPurchases">
      <text class="pending-text">您有 {{ pendingCount }} 笔采购单待接单</text>
      <text class="pending-link">去处理 {{ '>' }}</text>
    </view>

    <view class="card">
      <view class="card-title">快捷入口</view>
      <view class="grid">
        <view v-for="it in entries" :key="it.url" class="grid-item" @tap="go(it.url)">
          <view class="grid-icon" :class="it.cls">{{ it.text.slice(0, 1) }}</view>
          <view class="grid-text">{{ it.text }}</view>
          <view class="grid-desc">{{ it.desc }}</view>
        </view>
      </view>
    </view>

    <view class="bottom-space" />
    <TabBar current="workbench" />
  </view>
</template>

<style scoped>
.header {
  background: #16a34a;
  padding: 40rpx 32rpx 80rpx;
}

.hello-name {
  display: block;
  color: #ffffff;
  font-size: 36rpx;
  font-weight: 700;
}

.hello-sub {
  display: block;
  color: rgba(255, 255, 255, 0.85);
  font-size: 24rpx;
  margin-top: 10rpx;
}

.stats {
  margin-top: -50rpx;
  display: flex;
  flex-wrap: wrap;
}

.stat {
  width: 50%;
  text-align: center;
  padding: 16rpx 0;
}

.stat-num {
  font-size: 36rpx;
  font-weight: 700;
  color: #262626;
}

.stat-num .rmb {
  font-size: 24rpx;
  margin-right: 4rpx;
}

.stat-label {
  font-size: 24rpx;
  color: #999999;
  margin-top: 8rpx;
}

.pending-card {
  display: flex;
  justify-content: space-between;
  align-items: center;
  border: 2rpx solid #bbf7d0;
  background: #f0fdf4;
}

.pending-text {
  font-size: 28rpx;
  color: #15803d;
  font-weight: 600;
}

.pending-link {
  font-size: 26rpx;
  color: #16a34a;
}

.grid {
  display: flex;
  flex-wrap: wrap;
}

.grid-item {
  width: 33.33%;
  padding: 20rpx 0;
  text-align: center;
}

.grid-icon {
  width: 84rpx;
  height: 84rpx;
  margin: 0 auto 12rpx;
  border-radius: 24rpx;
  color: #ffffff;
  font-size: 34rpx;
  font-weight: 600;
  line-height: 84rpx;
}

.e-green { background: #16a34a; }
.e-blue { background: #2563eb; }
.e-gold { background: #d97706; }
.e-purple { background: #7c3aed; }
.e-grey { background: #6b7280; }

.grid-text {
  font-size: 26rpx;
  color: #262626;
}

.grid-desc {
  font-size: 20rpx;
  color: #bbbbbb;
  margin-top: 4rpx;
}

.bottom-space {
  height: 140rpx;
}
</style>
