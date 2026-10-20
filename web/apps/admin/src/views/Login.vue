<template>
  <div class="login-page">
    <el-card class="login-card">
      <h2 class="title">乡厨管理后台</h2>
      <el-tabs v-model="mode" stretch>
        <!-- 平台管理员：admins 表用户名+密码 -->
        <el-tab-pane label="管理员" name="admin">
          <el-form
            ref="adminFormRef"
            :model="adminForm"
            :rules="adminRules"
            label-position="top"
            @keyup.enter="onAdminSubmit"
          >
            <el-form-item label="用户名" prop="username">
              <el-input v-model="adminForm.username" placeholder="请输入用户名" :prefix-icon="User" clearable />
            </el-form-item>
            <el-form-item label="密码" prop="password">
              <el-input
                v-model="adminForm.password"
                type="password"
                placeholder="请输入密码"
                :prefix-icon="Lock"
                show-password
              />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" class="submit" :loading="loading" @click="onAdminSubmit">登 录</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <!-- 供应商：users 表手机号+密码（入驻审核通过后可用） -->
        <el-tab-pane label="供应商" name="supplier">
          <el-form
            ref="webFormRef"
            :model="webForm"
            :rules="webRules"
            label-position="top"
            @keyup.enter="onWebSubmit"
          >
            <el-form-item label="手机号" prop="phone">
              <el-input v-model="webForm.phone" placeholder="请输入入驻手机号" :prefix-icon="Iphone" clearable maxlength="11" />
            </el-form-item>
            <el-form-item label="密码" prop="password">
              <el-input
                v-model="webForm.password"
                type="password"
                placeholder="入驻时设置的密码"
                :prefix-icon="Lock"
                show-password
              />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" class="submit" :loading="loading" @click="onWebSubmit">登 录</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { Iphone, Lock, User } from '@element-plus/icons-vue'

import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const mode = ref<'admin' | 'supplier'>('admin')
const loading = ref(false)

// 管理员登录
const adminFormRef = ref<FormInstance>()
const adminForm = reactive({ username: '', password: '' })
const adminRules: FormRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
}

// 供应商登录（手机号 + 密码）
const webFormRef = ref<FormInstance>()
const webForm = reactive({ phone: '', password: '' })
const webRules: FormRules = {
  phone: [
    { required: true, message: '请输入手机号', trigger: 'blur' },
    { pattern: /^1\d{10}$/, message: '手机号格式不正确', trigger: 'blur' },
  ],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
}

function gotoHome(role: string) {
  // 按角色跳首页：供应商进工作台，管理员进仪表盘
  router.push(role === 'supplier' ? '/sup/ingredients' : (route.query.redirect as string) || '/')
}

async function onAdminSubmit() {
  const ok = await adminFormRef.value?.validate().catch(() => false)
  if (!ok) return
  loading.value = true
  try {
    const res = await auth.login(adminForm.username.trim(), adminForm.password)
    ElMessage.success('登录成功')
    gotoHome(res.role || 'admin')
  } catch (err) {
    // 错误信息已由拦截器统一提示
    console.error('login failed', err)
  } finally {
    loading.value = false
  }
}

async function onWebSubmit() {
  const ok = await webFormRef.value?.validate().catch(() => false)
  if (!ok) return
  loading.value = true
  try {
    const res = await auth.webLogin(webForm.phone.trim(), webForm.password)
    if (res.role === 'chef') {
      // 乡厨 Web 工作台暂未开放，引导去乡厨端
      auth.logout()
      ElMessage.info('乡厨请使用乡厨端小程序 / H5（apps/chef-mp）操作')
      return
    }
    ElMessage.success('登录成功')
    gotoHome(res.role)
  } catch (err) {
    const msg = (err as Error).message
    if (msg && !msg.includes('Request failed')) ElMessage.error(msg)
    console.error('web login failed', err)
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-page {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #1f2d3d 0%, #2b4a6f 100%);
}

.login-card {
  width: 380px;
  padding: 12px 8px;
}

.title {
  text-align: center;
  margin: 0 0 20px;
  color: #303133;
}

.submit {
  width: 100%;
}
</style>
