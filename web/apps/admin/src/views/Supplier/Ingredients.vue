<template>
  <div>
    <el-card shadow="never" class="page-card">
      <div class="table-toolbar">
        <el-input
          v-model="keyword"
          placeholder="食材名称搜索"
          clearable
          style="width: 200px"
          @keyup.enter="onSearch"
          @clear="onSearch"
        />
        <el-select v-model="category" placeholder="分类" clearable style="width: 140px" @change="onSearch">
          <el-option v-for="cat in categories" :key="cat" :label="cat" :value="cat" />
        </el-select>
        <div class="toolbar-spacer"></div>
        <el-button type="primary" @click="openCreate">新增食材</el-button>
      </div>

      <el-table v-loading="loading" :data="list" stripe>
        <el-table-column prop="name" label="名称" min-width="140" />
        <el-table-column prop="category" label="分类" width="110" />
        <el-table-column prop="unit" label="单位" width="70" />
        <el-table-column label="单价" width="110">
          <template #default="{ row }">{{ money(row.price) }} 元/{{ row.unit }}</template>
        </el-table-column>
        <el-table-column label="库存" width="120">
          <template #default="{ row }">
            <span>{{ row.stock }} {{ row.unit }}</span>
            <el-tag v-if="row.stock < 10" type="danger" size="small">偏低</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="origin" label="产地" width="110" />
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'">
              {{ row.status === 1 ? '在售' : '已下架' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="140" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
            <el-button v-if="row.status === 1" link type="danger" @click="onOffShelf(row)">下架</el-button>
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

    <!-- 新增/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="editId ? '编辑食材' : '新增食材'"
      width="520px"
      destroy-on-close
    >
      <el-form ref="formRef" :model="form" :rules="rules" label-width="80px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" placeholder="如：土猪肉（五花）" maxlength="64" />
        </el-form-item>
        <el-form-item label="分类" prop="category">
          <el-input v-model="form.category" placeholder="如：肉类" maxlength="32" />
        </el-form-item>
        <el-form-item label="单位" prop="unit">
          <el-input v-model="form.unit" placeholder="如：斤" maxlength="16" />
        </el-form-item>
        <el-form-item label="单价(元)" prop="priceYuan">
          <el-input-number v-model="form.priceYuan" :min="0.01" :precision="2" :step="1" style="width: 180px" />
        </el-form-item>
        <el-form-item label="库存" prop="stock">
          <el-input-number v-model="form.stock" :min="0" :step="1" style="width: 180px" />
        </el-form-item>
        <el-form-item label="产地">
          <el-input v-model="form.origin" placeholder="如：本地黑毛猪" maxlength="64" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="onSave">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import type { Ingredient } from '@xc/shared/types'

import { supplierApi } from '@/api/client'
import { money } from '@/utils/format'

const loading = ref(false)
const saving = ref(false)
const keyword = ref('')
const category = ref('')
const page = ref(1)
const size = 20
const total = ref(0)
const list = ref<Ingredient[]>([])
const categories = ['肉类', '蔬菜', '水产', '干货', '调味品', '粮油', '其他']

const dialogVisible = ref(false)
const editId = ref<number | undefined>(undefined)
const formRef = ref<FormInstance>()
const form = reactive({
  name: '',
  category: '',
  unit: '',
  priceYuan: 1,
  stock: 0,
  origin: '',
})

const rules: FormRules = {
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  category: [{ required: true, message: '请输入分类', trigger: 'blur' }],
  unit: [{ required: true, message: '请输入单位', trigger: 'blur' }],
}

async function load() {
  loading.value = true
  try {
    const res = await supplierApi.ingredients({
      keyword: keyword.value || undefined,
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

function openCreate() {
  editId.value = undefined
  Object.assign(form, { name: '', category: '', unit: '', priceYuan: 1, stock: 0, origin: '' })
  dialogVisible.value = true
}

function openEdit(row: Ingredient) {
  editId.value = row.id
  Object.assign(form, {
    name: row.name,
    category: row.category,
    unit: row.unit,
    priceYuan: (row.price ?? 0) / 100,
    stock: row.stock ?? 0,
    origin: row.origin ?? '',
  })
  dialogVisible.value = true
}

async function onSave() {
  const ok = await formRef.value?.validate().catch(() => false)
  if (!ok) return
  saving.value = true
  try {
    await supplierApi.saveIngredient(
      {
        name: form.name.trim(),
        category: form.category.trim(),
        unit: form.unit.trim(),
        price: Math.round(form.priceYuan * 100),
        stock: form.stock,
        origin: form.origin.trim(),
      },
      editId.value,
    )
    ElMessage.success(editId.value ? '已更新' : '已新增')
    dialogVisible.value = false
    load()
  } catch {
    // 拦截器已提示
  } finally {
    saving.value = false
  }
}

async function onOffShelf(row: Ingredient) {
  try {
    await ElMessageBox.confirm(`确定下架「${row.name}」吗？下架后乡厨将无法采购该食材。`, '下架确认', {
      confirmButtonText: '下架',
      cancelButtonText: '取消',
      type: 'warning',
    })
  } catch {
    return
  }
  try {
    await supplierApi.offShelfIngredient(row.id)
    ElMessage.success('已下架')
    load()
  } catch {
    // 拦截器已提示
  }
}

onMounted(load)
</script>

<style scoped>
.toolbar-spacer {
  flex: 1;
}

.stock-low {
  color: #f56c6c;
}
</style>
