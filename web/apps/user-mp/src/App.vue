<script setup lang="ts">
import { onLaunch } from '@dcloudio/uni-app'

onLaunch(() => {
  console.log('App Launch')
})
</script>

<style>
page {
  background-color: #f5f5f5;
  font-size: 28rpx;
  color: #333;
}

/* ---------- 通用布局 ---------- */
.page-pad {
  padding: 24rpx;
  box-sizing: border-box;
}
.card {
  background: #fff;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-bottom: 24rpx;
  box-sizing: border-box;
}
.row {
  display: flex;
  align-items: center;
}
.row-between {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.grow {
  flex: 1;
}
.muted {
  color: #999;
  font-size: 24rpx;
}
.strong {
  font-weight: 600;
}
.price {
  color: #d97706;
  font-weight: 700;
}

/* ---------- 标签 ---------- */
.tag {
  display: inline-block;
  padding: 4rpx 16rpx;
  border-radius: 8rpx;
  font-size: 22rpx;
  line-height: 1.6;
  background: #fff7ed;
  color: #d97706;
}
.tag-grey {
  background: #f3f4f6;
  color: #666;
}
.tag-green {
  background: #ecfdf5;
  color: #059669;
}
.tag-red {
  background: #fef2f2;
  color: #dc2626;
}

/* ---------- 按钮 ---------- */
.btn-primary {
  background: #d97706;
  color: #fff;
  border-radius: 44rpx;
  font-size: 30rpx;
  padding: 20rpx 0;
  text-align: center;
}
.btn-primary.disabled {
  opacity: 0.5;
}
.btn-plain {
  background: #fff;
  color: #d97706;
  border: 1rpx solid #d97706;
  border-radius: 44rpx;
  font-size: 28rpx;
  padding: 16rpx 0;
  text-align: center;
}
.btn-mini {
  display: inline-block;
  font-size: 24rpx;
  padding: 8rpx 28rpx;
  border-radius: 28rpx;
}

/* ---------- 表单 ---------- */
.form-item {
  display: flex;
  align-items: center;
  padding: 24rpx 0;
  border-bottom: 1rpx solid #f0f0f0;
}
.form-label {
  width: 170rpx;
  color: #666;
  font-size: 28rpx;
  flex-shrink: 0;
}
.form-input {
  flex: 1;
  font-size: 28rpx;
  color: #333;
}
.form-textarea {
  width: 100%;
  min-height: 140rpx;
  font-size: 28rpx;
}

/* ---------- 选项 chips ---------- */
.chips {
  display: flex;
  flex-wrap: wrap;
  gap: 16rpx;
}
.chip {
  padding: 10rpx 28rpx;
  border-radius: 28rpx;
  background: #f3f4f6;
  color: #666;
  font-size: 26rpx;
}
.chip.active {
  background: #fff7ed;
  color: #d97706;
  border: 1rpx solid #d97706;
}

/* ---------- 空态 ---------- */
.empty {
  text-align: center;
  color: #999;
  font-size: 26rpx;
  padding: 120rpx 0;
}
</style>
