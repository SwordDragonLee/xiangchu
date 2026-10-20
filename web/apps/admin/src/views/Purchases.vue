<template>
  <div>
    <!-- 顶部统计卡 -->
    <el-row :gutter="16">
      <el-col v-for="card in statCards" :key="card.label" :xs="12" :sm="8" :md="8" :lg="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-label">{{ card.label }}</div>
          <div class="stat-value">{{ card.value }}</div>
        </el-card>
      </el-col>
    </el-row>

    <el-card shadow="never" class="page-card">
      <div class="table-toolbar">
        <el-select v-model="status" placeholder="状态" clearable style="width: 160px" @change="onSearch">
          <el-option
            v-for="(meta, key) in purchaseStatusMap"
            :key="key"
            :label="meta.label"
            :value="Number(key)"
          />
        </el-select>
        <el-select
          v-model="fulfillMethod"
          placeholder="履约方式"
          clearable
          style="width: 140px"
          @change="onSearch"
        >
          <el-option
            v-for="(meta, key) in fulfillMethodMap"
            :key="key"
            :label="meta.label"
            :value="Number(key)"
          />
        </el-select>
      </div>

      <el-table v-loading="loading" :data="list" stripe>
        <el-table-column prop="purchase_no" label="采购单号" min-width="190" />
        <el-table-column label="供应商" min-width="160">
          <template #default="{ row }">{{ row.supplier_name || row.supplier_id || '-' }}</template>
        </el-table-column>
        <el-table-column label="金额" width="110">
          <template #default="{ row }">{{ money(row.total_amount) }}</template>
        </el-table-column>
        <el-table-column label="实付" width="110">
          <template #default="{ row }">{{ money(row.paid_amount) }}</template>
        </el-table-column>
        <el-table-column label="履约方式" width="100">
          <template #default="{ row }">
            <el-tag :type="statusMeta(fulfillMethodMap, row.fulfill_method).type">
              {{ statusMeta(fulfillMethodMap, row.fulfill_method).label }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="130">
          <template #default="{ row }">
            <el-tag :type="statusMeta(purchaseStatusMap, row.status).type">
              {{ statusMeta(purchaseStatusMap, row.status).label }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" min-width="170">
          <template #default="{ row }">{{ fmtTime(row.created_at) }}</template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination
          v-model:current-page="page"
          layout="total, prev, pager, next"
          :total="total"
          :page-size="size"
          @current-change="load"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import type { PurchaseOrder } from '@xc/shared/types'

import { api } from '@/api/client'
import { fmtTime, fulfillMethodMap, money, purchaseStatusMap, statusMeta } from '@/utils/format'

const loading = ref(false)
const status = ref<number | undefined>(undefined)
const fulfillMethod = ref<number | undefined>(undefined)
const page = ref(1)
const size = 20
const total = ref(0)
const list = ref<PurchaseOrder[]>([])

const stats = reactive({
  total: 0,
  pending: 0,
  completed: 0,
  by_fulfill: {} as Record<string, number>,
  gmv: 0,
})

const statCards = computed(() => [
  { label: '采购单总数', value: String(stats.total) },
  { label: '待接单', value: String(stats.pending) },
  { label: '已完成', value: String(stats.completed) },
  { label: '自取单量', value: String(stats.by_fulfill['1'] ?? 0) },
  { label: '送货单量', value: String(stats.by_fulfill['2'] ?? 0) },
  { label: '采购 GMV（元）', value: money(stats.gmv) },
])

async function load() {
  loading.value = true
  try {
    const res = await api.purchases({
      status: status.value,
      fulfill_method: fulfillMethod.value,
      page: page.value,
      size,
    })
    list.value = res.list ?? []
    total.value = res.total ?? 0
  } catch {
    // 拦截器已提示
  } finally {
    loading.value = false
  }
}

async function loadStats() {
  try {
    Object.assign(stats, await api.purchaseStats())
  } catch {
    // 拦截器已提示
  }
}

function onSearch() {
  page.value = 1
  load()
}

onMounted(() => {
  load()
  loadStats()
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
  font-size: 24px;
  font-weight: 600;
  color: #303133;
  word-break: break-all;
}
</style>
