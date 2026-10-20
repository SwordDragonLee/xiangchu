<template>
  <view class="review page-pad">
    <view class="card">
      <view class="card-title">服务评分</view>
      <view class="stars">
        <text
          v-for="i in 5"
          :key="i"
          class="star"
          :class="{ on: i <= rating }"
          @click="rating = i"
        >
          ★
        </text>
        <text class="star-tip">{{ rating }} 分</text>
      </view>
    </view>

    <view class="card">
      <view class="card-title">选择标签</view>
      <view class="chips">
        <view
          v-for="t in reviewTags"
          :key="t"
          class="chip"
          :class="{ active: picked.includes(t) }"
          @click="toggleTag(t)"
        >
          {{ t }}
        </view>
      </view>
    </view>

    <view class="card">
      <view class="card-title">评价内容</view>
      <textarea v-model="content" class="form-textarea" placeholder="说说乡厨的手艺、服务态度、守时情况吧（一期暂不支持上传图片）" />
      <view class="muted img-tip">图片上传将在二期开放，可先在文字中说明</view>
    </view>

    <button class="btn-primary" :class="{ disabled: submitting }" @click="submit">提交评价</button>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { orderApi } from '@/utils/api'
import { REVIEW_TAGS, errMsg } from '@/utils/format'

const reviewTags = REVIEW_TAGS
const orderId = ref(0)
const rating = ref(5)
const picked = ref<string[]>([])
const content = ref('')
const submitting = ref(false)

onLoad((query) => {
  orderId.value = Number(query?.order_id || 0)
  if (!orderId.value) {
    uni.showToast({ title: '缺少订单参数', icon: 'none' })
    setTimeout(() => uni.navigateBack(), 800)
  }
})

function toggleTag(t: string) {
  const idx = picked.value.indexOf(t)
  if (idx >= 0) picked.value.splice(idx, 1)
  else picked.value.push(t)
}

function submit() {
  if (submitting.value) return
  if (!content.value.trim()) {
    uni.showToast({ title: '请填写评价内容', icon: 'none' })
    return
  }
  submitting.value = true
  orderApi
    .review(orderId.value, {
      rating: rating.value,
      tags: picked.value,
      content: content.value.trim(),
    })
    .then(() => {
      uni.showToast({ title: '评价成功', icon: 'success' })
      setTimeout(() => uni.navigateBack(), 600)
    })
    .catch((e: unknown) => uni.showToast({ title: errMsg(e), icon: 'none' }))
    .finally(() => (submitting.value = false))
}
</script>

<style scoped>
.card-title {
  font-size: 30rpx;
  font-weight: 700;
  margin-bottom: 16rpx;
}
.stars {
  display: flex;
  align-items: center;
}
.star {
  font-size: 56rpx;
  color: #e5e7eb;
  margin-right: 12rpx;
}
.star.on {
  color: #d97706;
}
.star-tip {
  margin-left: 12rpx;
  color: #d97706;
  font-size: 28rpx;
  font-weight: 600;
}
.img-tip {
  margin-top: 12rpx;
}
</style>
