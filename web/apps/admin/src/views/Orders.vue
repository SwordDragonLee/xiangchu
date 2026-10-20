<template>
  <div>
    <el-card shadow="never" class="page-card">
      <div class="table-toolbar">
        <el-input
          v-model="keyword"
          placeholder="订单号"
          clearable
          style="width: 240px"
          @keyup.enter="onSearch"
          @clear="onSearch"
        />
        <el-button type="primary" :icon="Search" @click="onSearch">搜索</el-button>
      </div>

      <el-tabs v-model="statusTab" @tab-change="onTabChange">
        <el-tab-pane label="全部" name="all" />
        <el-tab-pane
          v-for="(meta, key) in orderStatusMap"
          :key="key"
          :label="meta.label"
          :name="String(key)"
        />
      </el-tabs>

      <el-table v-loading="loading" :data="list" stripe>
        <el-table-column prop="order_no" label="订单号" min-width="190" />
        <el-table-column label="类型" width="110">
          <template #default="{ row }">{{ banquetTypeLabel(row.banquet_type) }}</template>
        </el-table-column>
        <el-table-column label="宴席时间" min-width="160">
          <template #default="{ row }">{{ fmtTime(row.banquet_time) }}</template>
        </el-table-column>
        <el-table-column prop="table_count" label="桌数" width="80" />
        <el-table-column label="金额" width="110">
          <template #default="{ row }">{{ money(row.amount) }}</template>
        </el-table-column>
        <el-table-column label="实付" width="110">
          <template #default="{ row }">{{ money(row.paid_amount) }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusMeta(orderStatusMap, row.status).type">
              {{ statusMeta(orderStatusMap, row.status).label }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="退款状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusMeta(refundStatusMap, row.refund_status).type">
              {{ statusMeta(refundStatusMap, row.refund_status).label }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="90" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="openDetail(row)">详情</el-button>
          </template>
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

    <!-- 详情抽屉 -->
    <el-drawer v-model="detailVisible" title="订单详情" size="480px">
      <el-descriptions v-loading="detailLoading" :column="1" border>
        <el-descriptions-item label="订单号">{{ detail?.order_no }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="statusMeta(orderStatusMap, detail?.status).type">
            {{ orderStatusName(detail?.status) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="退款状态">
          {{ statusMeta(refundStatusMap, detail?.refund_status).label }}
        </el-descriptions-item>
        <el-descriptions-item label="类型">{{ banquetTypeLabel(detail?.banquet_type) }}</el-descriptions-item>
        <el-descriptions-item label="宴席时间">{{ fmtTime(detail?.banquet_time) }}</el-descriptions-item>
        <el-descriptions-item label="桌数">{{ detail?.table_count ?? '-' }}</el-descriptions-item>
        <el-descriptions-item label="人数">{{ detail?.people_count ?? '-' }}</el-descriptions-item>
        <el-descriptions-item label="金额">{{ money(detail?.amount) }} 元</el-descriptions-item>
        <el-descriptions-item label="实付">{{ money(detail?.paid_amount) }} 元</el-descriptions-item>
        <el-descriptions-item label="联系人">{{ detail?.contact_name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="联系电话">{{ detail?.contact_phone || '-' }}</el-descriptions-item>
        <el-descriptions-item label="地区">{{ detail?.region || '-' }}</el-descriptions-item>
        <el-descriptions-item label="地址">{{ detail?.address || '-' }}</el-descriptions-item>
        <el-descriptions-item label="菜单备注">{{ detail?.menu_remark || '-' }}</el-descriptions-item>
        <el-descriptions-item label="取消原因">{{ detail?.cancel_reason || '-' }}</el-descriptions-item>
        <el-descriptions-item label="下单时间">{{ fmtTime(detail?.created_at) }}</el-descriptions-item>
      </el-descriptions>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { Search } from '@element-plus/icons-vue'
import type { Dicts, Order } from '@xc/shared/types'

import { api } from '@/api/client'
import { fmtTime, money, orderStatusMap, refundStatusMap, statusMeta } from '@/utils/format'

const loading = ref(false)
const keyword = ref('')
const statusTab = ref('all')
const page = ref(1)
const size = 20
const total = ref(0)
const list = ref<Order[]>([])

const dicts = ref<Dicts | null>(null)

const detailVisible = ref(false)
const detailLoading = ref(false)
const detail = ref<Order | null>(null)

async function load() {
  loading.value = true
  try {
    const res = await api.orders({
      keyword: keyword.value || undefined,
      status: statusTab.value === 'all' ? undefined : Number(statusTab.value),
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

function onSearch() {
  page.value = 1
  load()
}

function onTabChange() {
  page.value = 1
  load()
}

function orderStatusName(status?: number): string {
  if (status === undefined || status === null) return '-'
  const key = String(status)
  if (dicts.value?.order_status && dicts.value.order_status[key]) {
    return dicts.value.order_status[key]
  }
  return statusMeta(orderStatusMap, status).label
}

function banquetTypeLabel(val?: string): string {
  if (!val) return '-'
  return val
}

async function openDetail(row: Order) {
  detailVisible.value = true
  detailLoading.value = true
  try {
    detail.value = await api.order(row.id)
  } catch {
    // 拦截器已提示
  } finally {
    detailLoading.value = false
  }
}

onMounted(async () => {
  load()
  try {
    dicts.value = await api.dicts()
  } catch {
    // 字典加载失败时使用本地映射
  }
})
</script>
