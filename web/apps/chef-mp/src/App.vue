<script setup lang="ts">
import { onLaunch } from '@dcloudio/uni-app'

onLaunch(() => {
  console.log('App launched')
})
</script>

<style>
page,
body {
  background-color: #f5f6f8;
  font-size: 28rpx;
  color: #262626;
}

view,
text,
input,
textarea,
button {
  box-sizing: border-box;
}

/* 通用卡片 */
.card {
  background: #ffffff;
  border-radius: 20rpx;
  padding: 28rpx;
  margin: 20rpx 24rpx;
  box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.04);
}

.card-title {
  font-size: 30rpx;
  font-weight: 600;
  color: #262626;
  margin-bottom: 20rpx;
}

/* 主按钮 */
.btn-primary {
  background: #d97706;
  color: #ffffff;
  border-radius: 44rpx;
  font-size: 30rpx;
  line-height: 84rpx;
  height: 84rpx;
  text-align: center;
  border: none;
}

.btn-primary[disabled] {
  background: #fcd9ac;
  color: #ffffff;
}

.btn-plain {
  background: #fffbeb;
  color: #d97706;
  border: 2rpx solid #d97706;
  border-radius: 44rpx;
  font-size: 30rpx;
  line-height: 80rpx;
  height: 84rpx;
  text-align: center;
}

/* 表单 */
.form-item {
  margin-bottom: 28rpx;
}

.form-label {
  font-size: 26rpx;
  color: #595959;
  margin-bottom: 12rpx;
  display: block;
}

.form-input {
  background: #f7f7f8;
  border-radius: 12rpx;
  height: 80rpx;
  padding: 0 24rpx;
  font-size: 28rpx;
}

.form-textarea {
  background: #f7f7f8;
  border-radius: 12rpx;
  padding: 20rpx 24rpx;
  font-size: 28rpx;
  width: 100%;
  min-height: 160rpx;
}

/* 空状态 */
.empty {
  text-align: center;
  color: #999999;
  font-size: 26rpx;
  padding: 80rpx 0;
}
</style>
