<template>
  <div>
    <el-row :gutter="16">
      <el-col v-for="card in statCards" :key="card.label" :xs="12" :sm="12" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-label">{{ card.label }}</div>
          <div class="stat-value">{{ card.value }}</div>
        </el-card>
      </el-col>
    </el-row>

    <el-card shadow="never" class="page-card">
      <div class="table-toolbar">
        <span class="section-title">提现记录</span>
        <div class="toolbar-spacer"></div>
        <el-button type="primary" @click="openWithdraw">申请提现</el-button>
        <el-button @click="loadAll">刷新</el-button>
      </div>

      <el-table v-loading="loading" :data="list" stripe>
        <el-table-column prop="withdraw_no" label="提现单号" min-width="190" />
        <el-table-column label="金额" width="120">
          <template #default="{ row }">{{ money(row.amount) }}</template>
        </el-table-column>
        <el-table-column label="渠道" width="90">
          <template #default="{ row }">
            {{ statusMeta(channelMap, row.channel).label }}
          </template>
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
      </el-table>

      <el-empty v-if="!loading && list.length === 0" description="暂无提现记录" />
    </el-card>

    <!-- 提现申请对话框 -->
    <el-dialog v-model="dialogVisible" title="申请提现" width="420px" destroy-on-close>
      <el-form label-width="90px">
        <el-form-item label="可提余额">
          <span class="available">¥ {{ money(income?.available) }}</span>
        </el-form-item>
        <el-form-item label="提现金额">
          <el-input-number v-model="amountYuan" :min="0.01" :precision="2" style="width: 200px" />
          <span style="margin-left: 8px">元</span>
        </el-form-item>
        <el-form-item label="打款渠道">
          <el-tag>线下打款</el-tag>
          <span class="channel-tip">一期为线下转账，平台审核后打款并标记</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="onWithdraw">提交申请</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import type { Income, Withdrawal } from '@xc/shared/types'

import { payApi } from '@/api/client'
import { channelMap, fmtTime, money, statusMeta, withdrawStatusMap } from '@/utils/format'

const loading = ref(false)
const saving = ref(false)
const list = ref<Withdrawal[]>([])
const income = ref<Income | null>(null)

const dialogVisible = ref(false)
const amountYuan = ref(1)

const statCards = computed(() => [
  { label: '累计结算（元）', value: money(income.value?.total_net) },
  { label: '本月结算（元）', value: money(income.value?.month_net) },
  { label: '累计已提现（元）', value: money(income.value?.total_withdrawn) },
  { label: '可提余额（元）', value: money(income.value?.available) },
])

async function loadIncome() {
  try {
    income.value = await payApi.supplierIncome()
  } catch {
    // 拦截器已提示
  }
}

async function loadList() {
  loading.value = true
  try {
    list.value = (await payApi.supplierWithdrawals()) ?? []
  } catch {
    // 拦截器已提示
  } finally {
    loading.value = false
  }
}

function loadAll() {
  loadIncome()
  loadList()
}

function openWithdraw() {
  amountYuan.value = 1
  dialogVisible.value = true
}

async function onWithdraw() {
  const fen = Math.round(amountYuan.value * 100)
  const available = income.value?.available ?? 0
  if (fen <= 0) {
    ElMessage.warning('请输入正确的提现金额')
    return
  }
  if (fen > available) {
    ElMessage.warning('提现金额不能超过可提余额')
    return
  }
  saving.value = true
  try {
    await payApi.supplierWithdraw(fen)
    ElMessage.success('已提交，等待平台审核打款')
    dialogVisible.value = false
    loadAll()
  } catch {
    // 拦截器已提示
  } finally {
    saving.value = false
  }
}

onMounted(loadAll)
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
}

.toolbar-spacer {
  flex: 1;
}

.section-title {
  font-weight: 600;
  color: #303133;
}

.available {
  font-size: 18px;
  font-weight: 600;
  color: #67c23a;
}

.channel-tip {
  margin-left: 8px;
  font-size: 12px;
  color: #909399;
}
</style>
