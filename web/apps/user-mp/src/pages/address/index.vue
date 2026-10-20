<template>
  <view class="address page-pad">
    <!-- 选择模式提示条 -->
    <view v-if="selectMode" class="card row-between select-tip">
      <text class="muted">点击地址卡片即可选用该地址</text>
      <view class="btn-plain btn-mini" @click="exitSelect">退出选择</view>
    </view>

    <!-- 地址列表 -->
    <view v-for="a in list" :key="a.id" class="card addr-card" @click="onPick(a)">
      <view class="row-between">
        <view class="row">
          <text class="strong">{{ a.contact_name }} {{ a.contact_phone }}</text>
          <text v-if="a.is_default" class="tag default-tag">默认</text>
        </view>
        <text v-if="!selectMode" class="muted" @click.stop="edit(a)">编辑</text>
        <text v-else class="tag tag-green">选用</text>
      </view>
      <view class="muted">{{ a.region }} {{ a.address }}</view>
    </view>

    <!-- 编辑/新增表单 card -->
    <view v-if="formOpen" class="card">
      <view class="card-title">{{ editingId ? '编辑地址' : '新增地址' }}</view>
      <view class="form-item">
        <view class="form-label">联系人</view>
        <input v-model="form.contact_name" class="form-input" placeholder="姓名" />
      </view>
      <view class="form-item">
        <view class="form-label">联系电话</view>
        <input v-model="form.contact_phone" class="form-input" type="number" maxlength="11" placeholder="手机号" />
      </view>
      <view class="form-item">
        <view class="form-label">所在地区</view>
        <picker mode="region" @change="onRegionChange">
          <view class="form-input picker-value">{{ form.region || '选择省 / 市 / 区' }}</view>
        </picker>
      </view>
      <view class="form-item">
        <view class="form-label">详细地址</view>
        <input v-model="form.address" class="form-input" placeholder="村镇 / 街道 / 门牌" />
      </view>
      <view class="form-item">
        <view class="form-label">默认地址切换</view>
        <switch :checked="form.is_default" color="#d97706" @change="onDefaultChange" />
      </view>
      <view class="row edit-actions">
        <view class="btn-plain btn-mini" @click="closeForm">取消</view>
        <view class="btn-primary btn-mini" @click="save">保存</view>
        <view v-if="editingId" class="btn-plain btn-mini danger" @click="remove">删除</view>
      </view>
    </view>
    <view v-else class="btn-primary submit" @click="openCreate">+ 新增地址</view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import type { Address } from '@xc/shared/types'
import { userApi } from '@/utils/api'
import { errMsg } from '@/utils/format'

const selectMode = ref(false)
const list = ref<Address[]>([])
const formOpen = ref(false)
const editingId = ref(0)
const submitting = ref(false)

const form = ref({
  contact_name: '',
  contact_phone: '',
  region: '',
  address: '',
  is_default: false,
})

function load() {
  userApi
    .addresses()
    .then((res) => (list.value = res || []))
    .catch((e: unknown) => uni.showToast({ title: errMsg(e), icon: 'none' }))
}

onShow(() => {
  const pages = getCurrentPages()
  const current = pages[pages.length - 1]
  const opts = (current as unknown as { options?: Record<string, string> }).options || {}
  selectMode.value = opts['select'] === '1'
  load()
})

function exitSelect() {
  uni.navigateBack()
}

function onPick(a: Address) {
  if (!selectMode.value) return
  uni.$emit('selectAddress', a)
  uni.navigateBack()
}

function openCreate() {
  editingId.value = 0
  form.value = { contact_name: '', contact_phone: '', region: '', address: '', is_default: false }
  formOpen.value = true
}

function closeForm() {
  formOpen.value = false
}

function edit(a: Address) {
  editingId.value = a.id
  form.value = {
    contact_name: a.contact_name,
    contact_phone: a.contact_phone,
    region: a.region,
    address: a.address,
    is_default: a.is_default,
  }
  formOpen.value = true
}

function onRegionChange(e: { detail: { value: string[] } }) {
  const v = e.detail.value || []
  form.value.region = v.filter(Boolean).join(' ')
}

function onDefaultChange(e: Event | { detail: { value: boolean } }) {
  const detail = (e as { detail?: { value?: boolean } }).detail
  form.value.is_default = !!detail?.value
}

function save() {
  if (submitting.value) return
  if (!form.value.contact_name) {
    uni.showToast({ title: '请填写联系人', icon: 'none' })
    return
  }
  if (!/^1\d{10}$/.test(form.value.contact_phone)) {
    uni.showToast({ title: '请填写正确的手机号', icon: 'none' })
    return
  }
  if (!form.value.region || !form.value.address) {
    uni.showToast({ title: '请填写完整地址', icon: 'none' })
    return
  }
  submitting.value = true
  userApi
    .saveAddress(
      {
        contact_name: form.value.contact_name,
        contact_phone: form.value.contact_phone,
        region: form.value.region,
        address: form.value.address,
        is_default: form.value.is_default,
      },
      editingId.value || undefined,
    )
    .then(() => {
      uni.showToast({ title: '已保存', icon: 'success' })
      formOpen.value = false
      load()
    })
    .catch((e: unknown) => uni.showToast({ title: errMsg(e), icon: 'none' }))
    .finally(() => (submitting.value = false))
}

function remove() {
  if (submitting.value || !editingId.value) return
  uni.showModal({
    title: '删除地址',
    content: '确认删除该地址吗？',
    success: (r) => {
      if (!r.confirm) return
      userApi
        .deleteAddress(editingId.value)
        .then(() => {
          uni.showToast({ title: '已删除', icon: 'success' })
          formOpen.value = false
          load()
        })
        .catch((e: unknown) => uni.showToast({ title: errMsg(e), icon: 'none' }))
    },
  })
}
</script>

<style scoped>
.select-tip {
  margin-bottom: 24rpx;
}
.addr-card {
  margin-bottom: 24rpx;
}
.default-tag {
  margin-left: 12rpx;
}
.edit-actions {
  gap: 20rpx;
  margin-top: 20rpx;
}
.danger {
  color: #dc2626;
  border-color: #dc2626;
}
.submit {
  margin-top: 40rpx;
}
</style>
