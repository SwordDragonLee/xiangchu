<script setup lang="ts">
// 自定义底部导航（图标 + 文字，图片资源在 src/static）
type TabKey = 'workbench' | 'orders' | 'purchases' | 'profile'
const props = defineProps<{ current: TabKey }>()

const tabs: { key: TabKey; text: string; url: string; icon: string; iconActive: string }[] = [
  { key: 'workbench', text: '工作台', url: '/pages/workbench/index', icon: '/static/tab-workbench.png', iconActive: '/static/tab-workbench-active.png' },
  { key: 'orders', text: '订单', url: '/pages/orders/index', icon: '/static/tab-orders.png', iconActive: '/static/tab-orders-active.png' },
  { key: 'purchases', text: '采购', url: '/pages/purchases/index', icon: '/static/tab-purchases.png', iconActive: '/static/tab-purchases-active.png' },
  { key: 'profile', text: '我的', url: '/pages/profile/index', icon: '/static/tab-profile.png', iconActive: '/static/tab-profile-active.png' },
]

function go(url: string, key: TabKey) {
  if (key === props.current) return
  uni.reLaunch({ url })
}
</script>

<template>
  <view class="tabbar">
    <view
      v-for="t in tabs"
      :key="t.key"
      class="tabbar-item"
      :class="{ active: t.key === props.current }"
      @tap="go(t.url, t.key)"
    >
      <image class="tabbar-icon" :src="t.key === props.current ? t.iconActive : t.icon" mode="aspectFit" />
      <text class="tabbar-text">{{ t.text }}</text>
    </view>
  </view>
</template>

<style scoped>
.tabbar {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  background: #ffffff;
  border-top: 2rpx solid #eeeeee;
  padding-bottom: env(safe-area-inset-bottom);
  z-index: 999;
}

.tabbar-item {
  flex: 1;
  height: 104rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 4rpx;
}

.tabbar-icon {
  width: 46rpx;
  height: 46rpx;
}

.tabbar-text {
  font-size: 20rpx;
  color: #999999;
}

.tabbar-item.active .tabbar-text {
  color: #d97706;
  font-weight: 600;
}
</style>
