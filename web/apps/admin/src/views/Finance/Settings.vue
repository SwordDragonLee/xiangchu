<template>
  <div>
    <el-card shadow="never" class="page-card">
      <template #header>
        <span>佣金费率设置</span>
      </template>

      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="140px"
        style="max-width: 520px"
        v-loading="loading"
      >
        <el-form-item label="乡厨佣金费率" prop="chef_percent">
          <el-input-number
            v-model="form.chef_percent"
            :min="0"
            :max="100"
            :precision="2"
            :step="0.5"
            controls-position="right"
          />
          <span class="unit">%</span>
          <span class="tip">（{{ bp(form.chef_percent) }} BP）</span>
        </el-form-item>
        <el-form-item label="供应商佣金费率" prop="supplier_percent">
          <el-input-number
            v-model="form.supplier_percent"
            :min="0"
            :max="100"
            :precision="2"
            :step="0.5"
            controls-position="right"
          />
          <span class="unit">%</span>
          <span class="tip">（{{ bp(form.supplier_percent) }} BP）</span>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="saving" @click="onSave">保存设置</el-button>
          <el-button @click="load">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'

import { api } from '@/api/client'

// 后端费率单位为 BP（1/10000），前端展示为百分比：BP = 百分比 * 100
const loading = ref(false)
const saving = ref(false)
const formRef = ref<FormInstance>()
const form = reactive({
  chef_percent: 0,
  supplier_percent: 0,
})

const rules: FormRules = {
  chef_percent: [{ required: true, message: '请输入乡厨佣金费率', trigger: 'blur' }],
  supplier_percent: [{ required: true, message: '请输入供应商佣金费率', trigger: 'blur' }],
}

function bp(percent: number | undefined): number {
  return Math.round((percent ?? 0) * 100)
}

async function load() {
  loading.value = true
  try {
    const res = await api.settings()
    form.chef_percent = Number(((res.chef_bp ?? 0) / 100).toFixed(2))
    form.supplier_percent = Number(((res.supplier_bp ?? 0) / 100).toFixed(2))
  } catch {
    // 拦截器已提示
  } finally {
    loading.value = false
  }
}

async function onSave() {
  const ok = await formRef.value?.validate().catch(() => false)
  if (!ok) return
  saving.value = true
  try {
    await api.updateSettings(bp(form.chef_percent), bp(form.supplier_percent))
    ElMessage.success('保存成功')
  } catch {
    // 拦截器已提示
  } finally {
    saving.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.unit {
  margin-left: 8px;
  color: #606266;
}

.tip {
  margin-left: 12px;
  color: #909399;
  font-size: 12px;
}
</style>
