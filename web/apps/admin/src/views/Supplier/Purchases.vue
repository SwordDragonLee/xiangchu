<template>
  <div>
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
        <el-button @click="load">刷新</el-button>
      </div>

      <el-table v-loading="loading" :data="list" stripe>
        <el-table-column prop="purchase_no" label="采购单号" min-width="190" />
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
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDetail(row)">详情</el-button>
            <el-button v-if="row.status === 0" link type="success" @click="onAccept(row)">接单</el-button>
            <el-button v-if="row.status === 0" link type="danger" @click="onReject(row)">拒单</el-button>
            <el-button v-if="row.status === 1" link type="primary" @click="onReady(row)">备货完成</el-button>
            <el-button v-if="row.status === 2" link type="success" @click="onComplete(row)">
              {{ row.fulfill_method === 2 ? '确认送达' : '确认完成' }}
            </el-button>
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

    <!-- 详情对话框 -->
    <el-dialog v-model="detailVisible" title="采购单详情" width="640px" destroy-on-close>
      <el-descriptions :column="2" border>
        <el-descriptions-item label="采购单号">{{ detail?.purchase_no }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          {{ statusMeta(purchaseStatusMap, detail?.status).label }}
        </el-descriptions-item>
        <el-descriptions-item label="金额">{{ money(detail?.total_amount) }}</el-descriptions-item>
        <el-descriptions-item label="实付">{{ money(detail?.paid_amount) }}</el-descriptions-item>
        <el-descriptions-item label="履约方式">
          {{ statusMeta(fulfillMethodMap, detail?.fulfill_method).label }}
        </el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ fmtTime(detail?.created_at) }}</el-descriptions-item>
        <el-descriptions-item v-if="detail?.fulfill_method === 1" label="自取地址" :span="2">
          {{ detail?.pickup_address || '-' }}
        </el-descriptions-item>
        <el-descriptions-item v-if="detail?.fulfill_method === 2" label="送货地址" :span="2">
          {{ detail?.deliver_address || '-' }}
        </el-descriptions-item>
        <el-descriptions-item v-if="detail?.fulfill_method === 2" label="期望送达">
          {{ fmtTime(detail?.deliver_time) }}
        </el-descriptions-item>
      </el-descriptions>

      <el-table :data="detail?.items ?? []" stripe class="detail-items">
        <el-table-column prop="name" label="食材" min-width="140" />
        <el-table-column label="单价" width="110">
          <template #default="{ row }">{{ money(row.price) }} 元/{{ row.unit }}</template>
        </el-table-column>
        <el-table-column prop="quantity" label="数量" width="90" />
        <el-table-column label="小计" width="110">
          <template #default="{ row }">{{ money(row.subtotal) }}</template>
        </el-table-column>
      </el-table>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { PurchaseOrder } from '@xc/shared/types'

import { supplierApi } from '@/api/client'
import { fmtTime, fulfillMethodMap, money, purchaseStatusMap, statusMeta } from '@/utils/format'

const loading = ref(false)
const status = ref<number | undefined>(undefined)
const page = ref(1)
const size = 20
const total = ref(0)
const list = ref<PurchaseOrder[]>([])

const detailVisible = ref(false)
const detail = ref<PurchaseOrder | null>(null)

async function load() {
  loading.value = true
  try {
    const res = await supplierApi.purchases({
      status: status.value,
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

async function openDetail(row: PurchaseOrder) {
  try {
    detail.value = await supplierApi.purchase(row.id)
    detailVisible.value = true
  } catch {
    // 拦截器已提示
  }
}

async function run(fn: () => Promise<unknown>, successMsg: string) {
  try {
    await fn()
    ElMessage.success(successMsg)
    load()
  } catch {
    // 拦截器已提示
  }
}

function onAccept(row: PurchaseOrder) {
  run(() => supplierApi.acceptPurchase(row.id), '已接单，库存已锁定')
}

function onReady(row: PurchaseOrder) {
  run(() => supplierApi.readyPurchase(row.id), '已标记备货完成')
}

function onComplete(row: PurchaseOrder) {
  const label = row.fulfill_method === 2 ? '确认送达' : '确认自取完成'
  ElMessageBox.confirm(`确认${label}？完成后将进入供应商结算（平台佣金 1%）。`, '完成确认', {
    confirmButtonText: '确认',
    cancelButtonText: '取消',
    type: 'warning',
  })
    .then(() => run(() => supplierApi.completePurchase(row.id), '已完成'))
    .catch(() => undefined)
}

function onReject(row: PurchaseOrder) {
  ElMessageBox.prompt('请输入拒单原因', '拒单', {
    confirmButtonText: '拒单',
    cancelButtonText: '取消',
    inputPattern: /\S+/,
    inputErrorMessage: '拒单原因必填',
    type: 'warning',
  })
    .then(({ value }) => run(() => supplierApi.rejectPurchase(row.id, value ?? ''), '已拒单'))
    .catch(() => undefined)
}

onMounted(load)
</script>

<style scoped>
.detail-items {
  margin-top: 16px;
}
</style>
