<template>
  <div>
    <el-card shadow="never" class="page-card">
      <div class="table-toolbar">
        <el-input
          v-model="keyword"
          placeholder="手机号 / 昵称"
          clearable
          style="width: 240px"
          @keyup.enter="onSearch"
          @clear="onSearch"
        />
        <el-button type="primary" :icon="Search" @click="onSearch">搜索</el-button>
      </div>

      <el-table v-loading="loading" :data="list" stripe>
        <el-table-column prop="id" label="ID" width="90" />
        <el-table-column prop="phone" label="手机号" width="140" />
        <el-table-column prop="nickname" label="昵称" min-width="160">
          <template #default="{ row }">{{ row.nickname || '-' }}</template>
        </el-table-column>
        <el-table-column label="角色" width="160">
          <template #default="{ row }">
            <el-tag
              v-for="r in row.roles"
              :key="r"
              :type="roleMeta(r).type"
              class="role-tag"
            >
              {{ roleMeta(r).label }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="110">
          <template #default="{ row }">
            <el-tag :type="statusMeta(userStatusMap, row.status).type">
              {{ statusMeta(userStatusMap, row.status).label }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="注册时间" min-width="170">
          <template #default="{ row }">{{ fmtTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="启用/禁用" width="110">
          <template #default="{ row }">
            <el-switch
              :model-value="row.status === 1"
              :disabled="busyId === row.id"
              @change="onToggleStatus(row, $event)"
            />
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
import { Search } from '@element-plus/icons-vue'

import { api } from '@/api/client'
import { fmtTime, statusMeta, userStatusMap } from '@/utils/format'

interface AdminUserRow {
  id: number
  phone: string
  nickname: string
  roles: string[] // 后端按档案派生的全部角色，无档案为 ["user"]
  status: number // 1启用 0禁用
  created_at: string
}

const roleMap: Record<string, { label: string; type: 'info' | 'warning' | 'success' }> = {
  user: { label: '普通用户', type: 'info' },
  chef: { label: '乡厨', type: 'warning' },
  supplier: { label: '供应商', type: 'success' },
}
function roleMeta(role?: string) {
  return roleMap[role ?? 'user'] ?? { label: '普通用户', type: 'info' as const }
}

const loading = ref(false)
const keyword = ref('')
const page = ref(1)
const size = 20
const total = ref(0)
const list = ref<AdminUserRow[]>([])
const busyId = ref(-1)

async function load() {
  loading.value = true
  try {
    const res = await api.users({ keyword: keyword.value || undefined, page: page.value, size })
    list.value = (res.list ?? []) as unknown as AdminUserRow[]
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

async function onToggleStatus(row: AdminUserRow, val: string | number | boolean) {
  const status = val ? 1 : 0
  busyId.value = row.id
  try {
    await api.setUserStatus(row.id, status)
    row.status = status
    ElMessage.success(status === 1 ? '已启用' : '已禁用')
  } catch {
    // 失败时刷新以还原开关状态
    await load()
  } finally {
    busyId.value = -1
  }
}

onMounted(load)
</script>

<style scoped>
.role-tag {
  margin-right: 8px;
}
</style>
