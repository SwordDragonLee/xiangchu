<template>
  <div class="dashboard">
    <el-row :gutter="16">
      <el-col v-for="card in cards" :key="card.label" :xs="12" :sm="12" :md="8" :lg="6" :xl="3">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-label">{{ card.label }}</div>
          <div class="stat-value">{{ card.value }}</div>
          <div v-if="card.sub" class="stat-sub">{{ card.sub }}</div>
        </el-card>
      </el-col>
    </el-row>

    <el-card shadow="never" class="page-card">
      <template #header>
        <span>订单区域分布</span>
      </template>
      <el-table v-loading="loading" :data="regions" stripe>
        <el-table-column type="index" label="序号" width="80" />
        <el-table-column prop="region" label="区域" min-width="180" />
        <el-table-column prop="count" label="订单数" width="160" />
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'

import { api } from '@/api/client'
import { money } from '@/utils/format'

const loading = ref(false)

const orderStats = reactive({
  gmv: 0,
  total: 0,
  pending: 0,
  completed: 0,
  repurchase_rate: 0,
})
const chefStats = reactive({ total: 0, pending: 0 })
const supplierStats = reactive({ total: 0, pending: 0 })
const financeStats = reactive({
  gmv: 0,
  settlement_count: 0,
  settlement_amount: 0,
  withdraw_pending: 0,
})

const regions = ref<{ region: string; count: number }[]>([])

const cards = computed(() => [
  { label: 'GMV（元）', value: money(orderStats.gmv) },
  { label: '订单总数', value: String(orderStats.total) },
  { label: '进行中订单', value: String(orderStats.pending) },
  { label: '已完成订单', value: String(orderStats.completed) },
  { label: '复购率', value: `${(orderStats.repurchase_rate * 100).toFixed(2)}%` },
  {
    label: '乡厨总数',
    value: String(chefStats.total),
    sub: `待审核 ${chefStats.pending}`,
  },
  {
    label: '供应商总数',
    value: String(supplierStats.total),
    sub: `待审核 ${supplierStats.pending}`,
  },
  { label: '待处理提现', value: String(financeStats.withdraw_pending) },
])

onMounted(async () => {
  loading.value = true
  const [o, c, s, f] = await Promise.allSettled([
    api.orderStats(),
    api.chefStats(),
    api.supplierStats(),
    api.financeStats(),
  ])
  if (o.status === 'fulfilled') Object.assign(orderStats, o.value)
  if (c.status === 'fulfilled') Object.assign(chefStats, c.value)
  if (s.status === 'fulfilled') Object.assign(supplierStats, s.value)
  if (f.status === 'fulfilled') Object.assign(financeStats, f.value)
  regions.value = o.status === 'fulfilled' ? (o.value.regions ?? []) : []
  loading.value = false
})
</script>

<style scoped>
.stat-card {
  margin-bottom: 16px;
  text-align: center;
}

.stat-label {
  color: #909399;
  font-size: 13px;
}

.stat-value {
  margin-top: 8px;
  font-size: 26px;
  font-weight: 600;
  color: #303133;
  word-break: break-all;
}

.stat-sub {
  margin-top: 6px;
  font-size: 12px;
  color: #e6a23c;
}
</style>
