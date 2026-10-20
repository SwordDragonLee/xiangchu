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
        <el-select v-model="status" placeholder="状态" clearable style="width: 160px" @change="onSearch">
          <el-option
            v-for="(meta, key) in withdrawStatusMap"
            :key="key"
            :label="meta.label"
            :value="Number(key)"
          />
        </el-select>
      </div>

      <el-table v-loading="loading" :data="list" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="withdraw_no" label="提现单号" min-width="190" />
        <el-table-column label="角色" width="100">
          <template #default="{ row }">
            <el-tag :type="statusMeta(roleMap, row.role).type">
              {{ statusMeta(roleMap, row.role).label }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="金额" width="120">
          <template #default="{ row }">{{ money(row.amount) }}</template>
        </el-table-column>
        <el-table-column label="渠道" width="100">
          <template #default="{ row }">{{ statusMeta(channelMap, row.channel).label }}</template>
        </el-table-column>
        <el-table-column label="状态" width="130">
          <template #default="{ row }">
            <el-tag :type="statusMeta(withdrawStatusMap, row.status).type">
              {{ statusMeta(withdrawStatusMap, row.status).label }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="申请时间" min-width="170">
          <template #default="{ row }">{{ fmtTime(row.applied_at) }}</template>
        </el-table-column>
        <el-table-column label="备注" min-width="140">
          <template #default="{ row }">{{ row.audit_remark || '-' }}</template>
        </el-table-column>
        <el-table-column label="操作" width="170" fixed="right">
          <template #default="{ row }">
            <template v-if="row.status === 0">
              <el-button type="success" link @click="openAudit(row, true)">通过</el-button>
              <el-button type="danger" link @click="openAudit(row, false)">拒绝</el-button>
            </template>
            <el-button
              v-else-if="row.status === 1"
              type="primary"
              link
              :loading="busyId === row.id"
              @click="onMarkPaid(row)"
            >
              标记已打款
            </el-button>
            <span v-else>-</span>
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
    <el-dialog v-model="auditVisible" title="提现审核" width="460px" destroy-on-close>
      <el-form label-width="80px">
        <el-form-item label="提现单号">{{ current?.withdraw_no }}</el-form-item>
        <el-form-item label="金额">{{ money(current?.amount) }} 元</el-form-item>
        <el-form-item label="审核结果">
          <el-radio-group v-model="auditForm.approve">
            <el-radio :value="true">通过</el-radio>
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
import { ElMessage, ElMessageBox } from 'element-plus'
import type { Withdrawal } from '@xc/shared/types'

import { api } from '@/api/client'
import { channelMap, fmtTime, money, roleMap, statusMeta, withdrawStatusMap } from '@/utils/format'

const loading = ref(false)
const role = ref<number | undefined>(undefined)
const status = ref<number | undefined>(undefined)
const page = ref(1)
const size = 20
const total = ref(0)
const list = ref<Withdrawal[]>([])
const busyId = ref(-1)

const auditVisible = ref(false)
const auditLoading = ref(false)
const current = ref<Withdrawal | null>(null)
const auditForm = reactive({ approve: true, remark: '' })

async function load() {
  loading.value = true
  try {
    const res = await api.withdrawals({
      role: role.value,
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

function openAudit(row: Withdrawal, approve: boolean) {
  current.value = row
  auditForm.approve = approve
  auditForm.remark = ''
  auditVisible.value = true
}

async function submitAudit() {
  if (!current.value) return
  auditLoading.value = true
  try {
    await api.auditWithdrawal(current.value.id, auditForm.approve, auditForm.remark.trim())
    ElMessage.success(auditForm.approve ? '已通过' : '已拒绝')
    auditVisible.value = false
    await load()
  } catch {
    // 拦截器已提示
  } finally {
    auditLoading.value = false
  }
}

// 已通过待打款 -> 标记已打款
async function onMarkPaid(row: Withdrawal) {
  try {
    await ElMessageBox.confirm(
      `确认已向提现单 ${row.withdraw_no} 打款 ${money(row.amount)} 元？`,
      '标记已打款',
      { confirmButtonText: '确认', cancelButtonText: '取消', type: 'warning' },
    )
  } catch {
    return
  }
  busyId.value = row.id
  try {
    await api.markWithdrawalPaid(row.id)
    ElMessage.success('已标记打款')
    await load()
  } catch {
    // 拦截器已提示
  } finally {
    busyId.value = -1
  }
}

onMounted(load)
</script>
