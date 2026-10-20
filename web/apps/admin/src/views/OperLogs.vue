<template>
  <div>
    <el-card shadow="never" class="page-card">
      <el-table v-loading="loading" :data="list" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column label="时间" min-width="170">
          <template #default="{ row }">{{ fmtTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column prop="admin_id" label="管理员" width="100" />
        <el-table-column prop="action" label="动作" min-width="140">
          <template #default="{ row }">{{ row.action || '-' }}</template>
        </el-table-column>
        <el-table-column label="目标" min-width="140">
          <template #default="{ row }">
            {{ row.target_type ? `${row.target_type}#${row.target_id}` : '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="detail" label="详情" min-width="240" show-overflow-tooltip>
          <template #default="{ row }">{{ row.detail || '-' }}</template>
        </el-table-column>
        <el-table-column prop="ip" label="IP" width="140">
          <template #default="{ row }">{{ row.ip || '-' }}</template>
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
import type { OperationLog } from '@xc/shared/types'

import { api } from '@/api/client'
import { fmtTime } from '@/utils/format'

const loading = ref(false)
const page = ref(1)
const size = 20
const total = ref(0)
const list = ref<OperationLog[]>([])

async function load() {
  loading.value = true
  try {
    const res = await api.operLogs({ page: page.value, size })
    list.value = res.list ?? []
    total.value = res.total ?? 0
  } catch {
    // 拦截器已提示
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>
