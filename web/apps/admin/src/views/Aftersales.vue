<template>
  <div>
    <el-card shadow="never" class="page-card">
      <div class="table-toolbar">
        <el-select v-model="status" placeholder="售后状态" clearable style="width: 160px" @change="onSearch">
          <el-option
            v-for="(meta, key) in aftersaleStatusMap"
            :key="key"
            :label="meta.label"
            :value="Number(key)"
          />
        </el-select>
      </div>

      <el-table v-loading="loading" :data="list" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="after_no" label="售后单号" min-width="190" />
        <el-table-column prop="order_id" label="订单ID" width="110" />
        <el-table-column label="金额" width="120">
          <template #default="{ row }">{{ money(row.amount) }}</template>
        </el-table-column>
        <el-table-column prop="reason" label="原因" min-width="180">
          <template #default="{ row }">{{ row.reason || '-' }}</template>
        </el-table-column>
        <el-table-column label="状态" width="130">
          <template #default="{ row }">
            <el-tag :type="statusMeta(aftersaleStatusMap, row.status).type">
              {{ statusMeta(aftersaleStatusMap, row.status).label }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="申请时间" min-width="170">
          <template #default="{ row }">{{ fmtTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="row.status === 0"
              type="primary"
              link
              @click="openAudit(row)"
            >
              审核
            </el-button>
            <span v-else>{{ row.audit_remark || '-' }}</span>
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
    <el-dialog v-model="auditVisible" title="售后审核" width="460px" destroy-on-close>
      <el-form label-width="80px">
        <el-form-item label="售后单号">{{ current?.after_no }}</el-form-item>
        <el-form-item label="金额">{{ money(current?.amount) }} 元</el-form-item>
        <el-form-item label="审核结果">
          <el-radio-group v-model="auditForm.approve">
            <el-radio :value="true">同意</el-radio>
            <el-radio :value="false">拒绝</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="备注">
          <el-input
            v-model="auditForm.remark"
            type="textarea"
            :rows="3"
            maxlength="200"
            show-word-limit
            placeholder="请输入审核备注"
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
import type { AfterSale } from '@xc/shared/types'

import { api } from '@/api/client'
import { aftersaleStatusMap, fmtTime, money, statusMeta } from '@/utils/format'

const loading = ref(false)
const status = ref<number | undefined>(undefined)
const page = ref(1)
const size = 20
const total = ref(0)
const list = ref<AfterSale[]>([])

const auditVisible = ref(false)
const auditLoading = ref(false)
const current = ref<AfterSale | null>(null)
const auditForm = reactive({ approve: true, remark: '' })

async function load() {
  loading.value = true
  try {
    const res = await api.aftersales({
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

function openAudit(row: AfterSale) {
  current.value = row
  auditForm.approve = true
  auditForm.remark = ''
  auditVisible.value = true
}

async function submitAudit() {
  if (!current.value) return
  auditLoading.value = true
  try {
    await api.auditAftersale(current.value.id, auditForm.approve, auditForm.remark.trim())
    ElMessage.success(auditForm.approve ? '已同意' : '已拒绝')
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
