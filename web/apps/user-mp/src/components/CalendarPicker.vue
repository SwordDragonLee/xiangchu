<script setup lang="ts">
// 档期日历选择器：绿=可约（可点选），黄=休息日，灰=已有订单，均不可点
import { computed, ref, watch } from 'vue'
import { chefPublicApi } from '@/utils/api'
import { errMsg } from '@/utils/format'

interface DayCell {
  date: string
  day: number
  status: number // 0可约 1已锁定 2休息日
  past: boolean
  today: boolean
}

interface MonthBlock {
  key: string
  title: string
  leadBlanks: number
  days: DayCell[]
}

const props = defineProps<{ chefId: number; modelValue: string }>()
const emit = defineEmits<{ (e: 'update:modelValue', v: string): void }>()

const WEEKS = ['一', '二', '三', '四', '五', '六', '日']

const from = new Date()
const to = new Date(Date.now() + 59 * 24 * 3600 * 1000)
const fromStr = fmtLocal(from)
const toStr = fmtLocal(to)
const todayStr = fromStr

const statusMap = ref<Record<string, number>>({})
const months = computed<MonthBlock[]>(() => buildMonths())

watch(
  () => props.chefId,
  (id) => {
    if (id) load(id)
  },
  { immediate: true },
)

function fmtLocal(d: Date): string {
  const m = d.getMonth() + 1
  const day = d.getDate()
  return `${d.getFullYear()}-${m < 10 ? '0' + m : m}-${day < 10 ? '0' + day : day}`
}

async function load(id: number) {
  try {
    const list = await chefPublicApi.schedules(id, fromStr, toStr)
    const map: Record<string, number> = {}
    for (const s of list) map[s.date.slice(0, 10)] = s.status
    statusMap.value = map
  } catch (e) {
    // 档期加载失败不阻塞下单，仅日历无标注
    console.log('[calendar] schedules load fail', errMsg(e))
  }
}

function buildMonths(): MonthBlock[] {
  const blocks: MonthBlock[] = []
  const d = new Date(from.getTime())
  while (fmtLocal(d) <= toStr) {
    const y = d.getFullYear()
    const m = d.getMonth()
    const days: DayCell[] = []
    while (d.getFullYear() === y && d.getMonth() === m && fmtLocal(d) <= toStr) {
      const date = fmtLocal(d)
      days.push({
        date,
        day: d.getDate(),
        status: statusMap.value[date] ?? 0,
        past: date <= todayStr,
        today: date === todayStr,
      })
      d.setDate(d.getDate() + 1)
    }
    blocks.push({
      key: `${y}-${m + 1}`,
      title: `${y}年${m + 1}月`,
      leadBlanks: (new Date(y, m, 1).getDay() + 6) % 7,
      days,
    })
  }
  return blocks
}

function pick(day: DayCell) {
  if (day.past) {
    uni.showToast({ title: '宴席日期须晚于今天', icon: 'none' })
    return
  }
  if (day.status === 1) {
    uni.showToast({ title: '该日乡厨已有宴席，请换一天', icon: 'none' })
    return
  }
  if (day.status === 2) {
    uni.showToast({ title: '该日乡厨休息，请换一天', icon: 'none' })
    return
  }
  emit('update:modelValue', day.date)
}
</script>

<template>
  <view class="cal">
    <view class="legend">
      <view class="lg-item"><view class="lg-dot free" />可约</view>
      <view class="lg-item"><view class="lg-dot rest" />休息</view>
      <view class="lg-item"><view class="lg-dot locked" />已有宴席</view>
    </view>
    <view v-for="m in months" :key="m.key" class="month">
      <view class="m-title">{{ m.title }}</view>
      <view class="week-row">
        <view v-for="w in WEEKS" :key="w" class="week-cell">{{ w }}</view>
      </view>
      <view class="grid">
        <view v-for="i in m.leadBlanks" :key="'b' + i" class="cell blank" />
        <view
          v-for="d in m.days"
          :key="d.date"
          class="cell day"
          :class="{
            free: d.status === 0 && !d.past,
            rest: d.status === 2,
            locked: d.status === 1,
            picked: d.date === modelValue,
            today: d.today,
          }"
          @tap="pick(d)"
        >
          <text class="day-num">{{ d.day }}</text>
          <text v-if="d.today" class="day-tag">今天</text>
          <text v-else-if="d.status === 1" class="day-tag">有席</text>
          <text v-else-if="d.status === 2" class="day-tag">休</text>
        </view>
      </view>
    </view>
  </view>
</template>

<style scoped>
.legend {
  display: flex;
  margin-bottom: 16rpx;
}

.lg-item {
  display: flex;
  align-items: center;
  font-size: 22rpx;
  color: #999999;
  margin-right: 28rpx;
}

.lg-dot {
  width: 18rpx;
  height: 18rpx;
  border-radius: 6rpx;
  margin-right: 8rpx;
}

.lg-dot.free {
  background: #f2fbf5;
  border: 2rpx solid #16a34a;
}

.lg-dot.rest {
  background: #fef3c7;
  border: 2rpx solid #f59e0b;
}

.lg-dot.locked {
  background: #e5e7eb;
}

.month {
  margin-bottom: 24rpx;
}

.m-title {
  font-size: 26rpx;
  font-weight: 600;
  color: #595959;
  margin-bottom: 12rpx;
}

.week-row,
.grid {
  display: flex;
  flex-wrap: wrap;
}

.week-cell {
  width: 14.28%;
  text-align: center;
  font-size: 22rpx;
  color: #bbbbbb;
  padding: 8rpx 0;
}

.cell {
  width: 14.28%;
  min-height: 84rpx;
  padding: 4rpx 2rpx;
  margin-bottom: 8rpx;
  border-radius: 10rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  border: 2rpx solid transparent;
}

.cell.blank {
  min-height: 0;
}

.cell.day.free {
  background: #f2fbf5;
  border-color: #e8f7ee;
}

.cell.day.rest {
  background: #fef3c7;
}

.cell.day.locked {
  background: #e5e7eb;
}

.cell.day.today {
  background: #f5f5f5;
}

.cell.day.today .day-num {
  color: #bbbbbb;
}

.cell.day.today .day-tag {
  color: #bbbbbb;
}

.cell.day.picked {
  background: #fffbeb;
  border-color: #d97706;
}

.cell.day.picked .day-num {
  color: #d97706;
  font-weight: 700;
}

.day-num {
  font-size: 26rpx;
  color: #262626;
}

.cell.day.rest .day-num,
.cell.day.locked .day-num {
  color: #9ca3af;
}

.day-tag {
  font-size: 18rpx;
  margin-top: 2rpx;
  color: #b45309;
}

.cell.day.locked .day-tag {
  color: #6b7280;
}
</style>
