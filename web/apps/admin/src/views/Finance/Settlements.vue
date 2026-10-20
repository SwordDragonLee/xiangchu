<template>
  <div>
    <el-card shadow="never" class="page-card">
      <div class="table-toolbar">
        <el-select v-model="role" placeholder="角色" clearable style="width: 140px" @change="onSearch">
          <el-option
            v-for="(meta, key) in roleMap"
            :key="key"
            :label="meta.label"
            :value="Number(key)"
          />
        </el-select>
      </div>

      <el-table v-loading="loading" :data="list" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="settle_no" label="结算单号" min-width="190" />
        <el-table-column label="角色" width="100">
          <template #default="{ row }">
            <el-tag :type="statusMeta(roleMap, row.role).type">
              {{ statusMeta(roleMap, row.role).label }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="biz_no" label="业务单号" min-width="180">
          <template #default="{ row }">{{ row.biz_no || '-' }}</template>
        </el-table-column>
        <el-table-column label="毛利" width="120">
          <template #default="{ row }">{{ money(row.gross_amount) }}</template>
        </el-table-column>
        <el-table-column label="佣金" width="120">
          <template #default="{ row }">{{ money(row.commission_amount) }}</template>
        </el-table-column>
        <el-table-column label="净额" width="120">
          <template #default="{ row }">{{ money(row.net_amount) }}</template>
        </el-table-column>
        <el-table-column label="结算时间" min-width="170">
          <template #default="{ row }">{{ fmtTime(row.settle_time) }}</template>
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
import type { Settlement } from '@xc/shared/types'

import { api } from '@/api/client'
import { fmtTime, money, roleMap, statusMeta } from '@/utils/format'

const loading = ref(false)
const role = ref<number | undefined>(undefined)
const page = ref(1)
const size = 20
const total = ref(0)
const list = ref<Settlement[]>([])

async function load() {
  loading.value = true
  try {
    const res = await api.settlements({
      role: role.value,
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

onMounted(load)
</script>
