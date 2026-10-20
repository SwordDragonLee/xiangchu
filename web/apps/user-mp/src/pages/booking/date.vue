<script setup lang="ts">
// 选择宴席日期：全屏档期日历，选中后回传下单页
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import CalendarPicker from '@/components/CalendarPicker.vue'

const chefId = ref(0)

onLoad((query) => {
  chefId.value = Number(query?.chef_id || 0)
  if (!chefId.value) {
    uni.showToast({ title: '缺少乡厨参数', icon: 'none' })
    setTimeout(() => uni.navigateBack(), 600)
  }
})

function onPick(d: string) {
  uni.$emit('selectDate', d)
  setTimeout(() => uni.navigateBack(), 120)
}
</script>

<template>
  <view class="page-pad">
    <view class="card">
      <view class="card-title">选择宴席日期</view>
      <view class="tip">绿=可约 · 黄=休息 · 灰=已有宴席 · 当天不可约；点可约日期即选中返回</view>
      <CalendarPicker v-if="chefId" :chef-id="chefId" :model-value="''" @update:model-value="onPick" />
    </view>
  </view>
</template>

<style scoped>
.tip {
  font-size: 24rpx;
  color: #999999;
  margin-bottom: 20rpx;
}
</style>
