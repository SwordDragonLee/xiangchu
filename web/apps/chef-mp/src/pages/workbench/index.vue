<script setup lang="ts">
import { onShow } from '@dcloudio/uni-app'
import { computed, ref } from 'vue'
import TabBar from '@/components/TabBar.vue'
import { chefApi, ensureChef, ensureLogin, errMsg, getRole, toast } from '@/utils/api'
import { fen2yuan } from '@/utils/format'
import type { Chef } from '@xc/shared'

const todayOrders = ref(0)
const pendingOrders = ref(0)
const monthIncome = ref(0)
const chefName = ref('')
const loading = ref(false)

const entries = [
  { text: '接单管理', desc: '待接订单处理', url: '/pages/orders/index', cls: 'e-orange' },
  { text: '食材采购', desc: '市场选菜下单', url: '/pages/food/index', cls: 'e-teal' },
  { text: '档期管理', desc: '休息日设置', url: '/pages/schedules/index', cls: 'e-green' },
  { text: '套餐管理', desc: '宴席套餐维护', url: '/pages/packages/index', cls: 'e-blue' },
  { text: '收入提现', desc: '余额与提现', url: '/pages/income/index', cls: 'e-gold' },
  { text: '我的档案', desc: '资料与审核状态', url: '/pages/profile/index', cls: 'e-purple' },
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
  if (!(await ensureChef())) return
  load()
})

async function load() {
  loading.value = true
  try {
    const wb = await chefApi.workbench()
    todayOrders.value = wb.today_orders
    pendingOrders.value = wb.pending_orders
    monthIncome.value = wb.month_income
    chefApi
      .profile()
      .then((p: Chef) => {
        chefName.value = p.real_name
      })
      .catch(() => {
        /* 档案获取失败不影响工作台 */
      })
  } catch (e) {
    toast(errMsg(e))
  } finally {
    loading.value = false
  }
}

function go(url: string) {
  uni.navigateTo({ url })
}
</script>

<template>
  <view class="page">
    <view class="header">
      <view class="hello">
        <text class="hello-name">{{ greeting }}，{{ chefName || '乡厨师傅' }}</text>
        <text class="hello-sub">今天也元气满满，接单顺利</text>
      </view>
    </view>

    <view class="stats card">
      <view class="stat">
        <view class="stat-num">{{ todayOrders }}</view>
        <view class="stat-label">今日订单</view>
      </view>
      <view class="stat-divider" />
      <view class="stat" @tap="go('/pages/orders/index')">
        <view class="stat-num warn">{{ pendingOrders }}</view>
        <view class="stat-label">待接单</view>
      </view>
      <view class="stat-divider" />
      <view class="stat" @tap="go('/pages/income/index')">
        <view class="stat-num">
          <text class="rmb">¥</text>{{ fen2yuan(monthIncome) }}
        </view>
        <view class="stat-label">本月收入(元)</view>
      </view>
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
  background: #d97706;
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
  align-items: center;
}

.stat {
  flex: 1;
  text-align: center;
}

.stat-divider {
  width: 2rpx;
  height: 60rpx;
  background: #eeeeee;
}

.stat-num {
  font-size: 40rpx;
  font-weight: 700;
  color: #262626;
}

.stat-num.warn {
  color: #e6504f;
}

.stat-num .rmb {
  font-size: 26rpx;
  margin-right: 4rpx;
}

.stat-label {
  font-size: 24rpx;
  color: #999999;
  margin-top: 8rpx;
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

.e-orange { background: #f59e0b; }
.e-green { background: #16a34a; }
.e-blue { background: #2563eb; }
.e-gold { background: #d97706; }
.e-purple { background: #7c3aed; }
.e-grey { background: #6b7280; }
.e-teal { background: #0d9488; }

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
