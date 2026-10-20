<template>
  <div>
    <el-card shadow="never" class="page-card">
      <el-table v-loading="loading" :data="list" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="order_id" label="订单ID" width="110" />
        <el-table-column label="评分" width="170">
          <template #default="{ row }">
            <el-rate :model-value="Number(row.rating ?? 0)" disabled allow-half />
          </template>
        </el-table-column>
        <el-table-column prop="content" label="内容" min-width="240">
          <template #default="{ row }">{{ row.content || '-' }}</template>
        </el-table-column>
        <el-table-column prop="tags" label="标签" min-width="140">
          <template #default="{ row }">{{ row.tags || '-' }}</template>
        </el-table-column>
        <el-table-column label="时间" min-width="170">
          <template #default="{ row }">{{ fmtTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusMeta(reviewStatusMap, row.status).type">
              {{ statusMeta(reviewStatusMap, row.status).label }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="row.status === 1"
              type="danger"
              link
              :loading="busyId === row.id"
              @click="onSetStatus(row, 0)"
            >
              隐藏
            </el-button>
            <el-button
              v-else
              type="success"
              link
              :loading="busyId === row.id"
              @click="onSetStatus(row, 1)"
            >
              恢复
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
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import type { Review } from '@xc/shared/types'

import { api, client } from '@/api/client'
import { fmtTime, reviewStatusMap, statusMeta } from '@/utils/format'

const loading = ref(false)
const page = ref(1)
const size = 20
const total = ref(0)
const list = ref<Review[]>([])
const busyId = ref(-1)

async function load() {
  loading.value = true
  try {
    const res = await api.reviews({ page: page.value, size })
    list.value = res.list ?? []
    total.value = res.total ?? 0
  } catch {
    // 拦截器已提示
  } finally {
    loading.value = false
  }
}

// 隐藏/恢复评价：POST /admin/reviews/:id/status { status: 0|1 }
async function onSetStatus(row: Review, status: 0 | 1) {
  busyId.value = row.id
  try {
    await client.post(`/admin/reviews/${row.id}/status`, { status })
    row.status = status
    ElMessage.success(status === 1 ? '已恢复显示' : '已隐藏')
  } catch {
    // 拦截器已提示
  } finally {
    busyId.value = -1
  }
}

onMounted(load)
</script>
