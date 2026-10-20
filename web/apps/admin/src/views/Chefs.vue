<template>
  <div>
    <el-card shadow="never" class="page-card">
      <div class="table-toolbar">
        <el-input
          v-model="keyword"
          placeholder="姓名 / 手机号"
          clearable
          style="width: 220px"
          @keyup.enter="onSearch"
          @clear="onSearch"
        />
        <el-select v-model="status" placeholder="状态" clearable style="width: 140px" @change="onSearch">
          <el-option
            v-for="(meta, key) in auditStatusMap"
            :key="key"
            :label="meta.label"
            :value="Number(key)"
          />
        </el-select>
        <el-button type="primary" :icon="Search" @click="onSearch">搜索</el-button>
      </div>

      <el-table v-loading="loading" :data="list" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="real_name" label="姓名" width="120">
          <template #default="{ row }">{{ row.real_name || '-' }}</template>
        </el-table-column>
        <el-table-column prop="phone" label="手机号" width="130" />
        <el-table-column prop="region" label="地区" min-width="150">
          <template #default="{ row }">{{ row.region || '-' }}</template>
        </el-table-column>
        <el-table-column prop="cuisines" label="菜系" min-width="140">
          <template #default="{ row }">{{ row.cuisines || '-' }}</template>
        </el-table-column>
        <el-table-column label="评分" width="90">
          <template #default="{ row }">{{ Number(row.rating ?? 0).toFixed(1) }}</template>
        </el-table-column>
        <el-table-column prop="order_count" label="接单数" width="90" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusMeta(auditStatusMap, row.status).type">
              {{ statusMeta(auditStatusMap, row.status).label }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="申请时间" min-width="170">
          <template #default="{ row }">{{ fmtTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="openAudit(row)">审核</el-button>
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

    <!-- 审核对话框 -->
    <el-dialog v-model="auditVisible" title="乡厨审核" width="460px" destroy-on-close>
      <el-form label-width="80px">
        <el-form-item label="乡厨">
          {{ current?.real_name || '-' }}（{{ current?.phone || '-' }}）
        </el-form-item>
        <el-form-item label="审核结果">
          <el-radio-group v-model="auditForm.approve">
            <el-radio :value="true">通过</el-radio>
            <el-radio :value="false">拒绝</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="备注" required>
          <el-input
            v-model="auditForm.remark"
            type="textarea"
            :rows="3"
            maxlength="200"
            show-word-limit
            :placeholder="auditForm.approve ? '请输入备注（选填）' : '拒绝时必填'"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="auditVisible = false">取消</el-button>
        <el-button type="primary" :loading="auditLoading" @click="submitAudit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import type { Chef } from '@xc/shared/types'

import { api } from '@/api/client'
import { auditStatusMap, fmtTime, statusMeta } from '@/utils/format'

const loading = ref(false)
const keyword = ref('')
const status = ref<number | undefined>(undefined)
const page = ref(1)
const size = 20
const total = ref(0)
const list = ref<Chef[]>([])

const auditVisible = ref(false)
const auditLoading = ref(false)
const current = ref<Chef | null>(null)
const auditForm = reactive({ approve: true, remark: '' })

async function load() {
  loading.value = true
  try {
    const res = await api.chefs({
      keyword: keyword.value || undefined,
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

function openAudit(row: Chef) {
  current.value = row
  auditForm.approve = true
  auditForm.remark = ''
  auditVisible.value = true
}

async function submitAudit() {
  if (!current.value) return
  if (!auditForm.approve && !auditForm.remark.trim()) {
    ElMessage.warning('拒绝时必须填写备注')
    return
  }
  auditLoading.value = true
  try {
    await api.auditChef(current.value.id, auditForm.approve, auditForm.remark.trim())
    ElMessage.success(auditForm.approve ? '已通过' : '已拒绝')
    auditVisible.value = false
    await load()
  } catch {
    // 拦截器已提示
  } finally {
    auditLoading.value = false
  }
}

onMounted(load)
</script>
