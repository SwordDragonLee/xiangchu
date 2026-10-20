<script setup lang="ts">
import { onShow } from '@dcloudio/uni-app'
import { computed, ref } from 'vue'
import { ensureLogin, errMsg, payApi, toast } from '@/utils/api'
import { WITHDRAW_STATUS_TEXT, fen2yuan, yuan2fen } from '@/utils/format'
import type { Income, Withdrawal } from '@xc/shared'

const income = ref<Income | null>(null)
const withdrawals = ref<Withdrawal[]>([])
const amount = ref('')
const submitting = ref(false)
const loading = ref(false)

const availableYuan = computed(() => (income.value ? fen2yuan(income.value.available) : '0.00'))

onShow(async () => {
  // 收入页从工作台/宫格进入，此处只做登录兜底（角色守卫已在 ensureSupplier 完成）
  if (!ensureLogin()) return
  await load()
})

async function load() {
  loading.value = true
  try {
    const [inc, ws] = await Promise.all([payApi.supplierIncome(), payApi.supplierWithdrawals()])
    income.value = inc
    withdrawals.value = ws
  } catch (e) {
    toast(errMsg(e))
  } finally {
    loading.value = false
  }
}

function fillAll() {
  if (income.value) amount.value = fen2yuan(income.value.available)
}

function withdraw() {
  if (!income.value) return
  const fen = yuan2fen(amount.value)
  if (fen <= 0) {
    toast('请输入正确的提现金额')
    return
  }
  if (fen > income.value.available) {
    toast('超出可提现余额')
    return
  }
  uni.showModal({
    title: '申请提现',
    content: `确认申请提现 ¥${fen2yuan(fen)}？提交后系统自动打款，大额或异常申请需平台审核`,
    success: async (res) => {
      if (!res.confirm) return
      submitting.value = true
      try {
        await payApi.supplierWithdraw(fen)
        toast('提现申请已提交')
        amount.value = ''
        await load()
      } catch (e) {
        toast(errMsg(e))
      } finally {
        submitting.value = false
      }
    },
  })
}

function fmtTime(s?: string): string {
  return (s || '').replace('T', ' ').slice(0, 16)
}
</script>

<template>
  <view class="page">
    <view class="income-hero">
      <view class="hero-label">可提现余额(元)</view>
      <view class="hero-amount">{{ availableYuan }}</view>
      <view class="hero-line">
        <text>累计净收入 ¥{{ fen2yuan(income?.total_net) }}</text>
        <text>本月净收入 ¥{{ fen2yuan(income?.month_net) }}</text>
      </view>
      <view v-if="(income?.frozen_amount || 0) > 0" class="hero-line frozen-line">
        <text>冻结中 ¥{{ fen2yuan(income?.frozen_amount) }}</text>
        <text>满冻结期自动可提</text>
      </view>
    </view>

    <view class="card">
      <view class="card-title">申请提现</view>
      <view class="wd-row">
        <view class="wd-input-wrap">
          <text class="rmb">¥</text>
          <input v-model="amount" class="wd-input" type="digit" placeholder="请输入提现金额(元)" />
          <text class="all-btn" @tap="fillAll">全部</text>
        </view>
        <button class="wd-btn" :disabled="submitting" @tap="withdraw">
          {{ submitting ? '提交中...' : '提现' }}
        </button>
      </view>
      <view class="wd-tip">提现金额需不小于 1 元，且不超过可提现余额</view>
    </view>

    <view class="card">
      <view class="card-title">提现记录</view>
      <view v-if="withdrawals.length === 0" class="empty">暂无提现记录</view>
      <view v-for="w in withdrawals" :key="w.id" class="wd-item">
        <view class="wd-left">
          <view class="wd-no">{{ w.withdraw_no }}</view>
          <view class="wd-time">{{ fmtTime(w.applied_at) }}</view>
          <view v-if="w.status === 2 && w.audit_remark" class="wd-remark">原因：{{ w.audit_remark }}</view>
        </view>
        <view class="wd-right">
          <view class="wd-amount">¥{{ fen2yuan(w.amount) }}</view>
          <view class="wd-status" :class="'ws-' + w.status">{{ WITHDRAW_STATUS_TEXT[w.status] }}</view>
        </view>
      </view>
    </view>
  </view>
</template>

<style scoped>
.income-hero {
  background: #16a34a;
  padding: 50rpx 40rpx;
  color: #ffffff;
}

.hero-label {
  font-size: 26rpx;
  opacity: 0.9;
}

.hero-amount {
  font-size: 72rpx;
  font-weight: 700;
  margin: 16rpx 0 24rpx;
}

.hero-line {
  display: flex;
  justify-content: space-between;
  font-size: 24rpx;
  opacity: 0.9;
}

.frozen-line {
  margin-top: 12rpx;
}

.wd-row {
  display: flex;
  align-items: center;
}

.wd-input-wrap {
  flex: 1;
  display: flex;
  align-items: center;
  background: #f7f7f8;
  border-radius: 12rpx;
  height: 80rpx;
  padding: 0 24rpx;
}

.rmb {
  color: #16a34a;
  font-size: 32rpx;
  font-weight: 700;
  margin-right: 8rpx;
}

.wd-input {
  flex: 1;
  font-size: 30rpx;
}

.all-btn {
  color: #16a34a;
  font-size: 26rpx;
  padding-left: 20rpx;
}

.wd-btn {
  margin-left: 20rpx;
  background: #16a34a;
  color: #ffffff;
  border-radius: 12rpx;
  font-size: 28rpx;
  line-height: 80rpx;
  height: 80rpx;
  padding: 0 40rpx;
}

.wd-btn[disabled] {
  background: #a7dcb8;
}

.wd-tip {
  font-size: 22rpx;
  color: #bbbbbb;
  margin-top: 16rpx;
}

.wd-item {
  display: flex;
  justify-content: space-between;
  padding: 24rpx 0;
  border-bottom: 2rpx solid #f5f5f5;
}

.wd-item:last-child {
  border-bottom: none;
}

.wd-no {
  font-size: 26rpx;
  color: #262626;
}

.wd-time {
  font-size: 22rpx;
  color: #999999;
  margin-top: 8rpx;
}

.wd-remark {
  font-size: 22rpx;
  color: #e6504f;
  margin-top: 8rpx;
}

.wd-right {
  text-align: right;
}

.wd-amount {
  font-size: 32rpx;
  font-weight: 700;
  color: #262626;
}

.wd-status {
  font-size: 24rpx;
  margin-top: 8rpx;
}

.ws-0 { color: #b45309; }
.ws-1 { color: #2563eb; }
.ws-2 { color: #e6504f; }
.ws-3 { color: #16a34a; }
</style>
