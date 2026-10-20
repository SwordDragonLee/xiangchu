<script setup lang="ts">
import { onLoad } from '@dcloudio/uni-app'
import { computed, ref } from 'vue'
import { chefApi, errMsg, toast } from '@/utils/api'
import { fmtDate } from '@/utils/format'

interface DayCell {
  date: string
  day: number
  blank?: boolean
  status: number // 0可约 1已锁定 2休息日
  past: boolean
}

const statusMap = ref<Record<string, number>>({})
const loading = ref(false)
const toggling = ref('') // 正在切换的日期

const WEEKS = ['一', '二', '三', '四', '五', '六', '日']

// 未来 60 天窗口：从今天起
const from = new Date()
const to = new Date(Date.now() + 59 * 24 * 3600 * 1000)
const fromStr = fmtDate(from)
const toStr = fmtDate(to)

const todayStr = fromStr

const months = computed<MonthBlock[]>(() => buildMonths())

interface MonthBlock {
  key: string
  title: string
  leadBlanks: number
  days: DayCell[]
}

function buildMonths(): MonthBlock[] {
  const blocks: MonthBlock[] = []
  const d = new Date(from.getTime())
  while (fmtDate(d) <= toStr) {
    const y = d.getFullYear()
    const m = d.getMonth()
    const key = `${y}-${String(m + 1).padStart(2, '0')}`
    const days: DayCell[] = []
    while (d.getFullYear() === y && d.getMonth() === m && fmtDate(d) <= toStr) {
      const date = fmtDate(d)
      days.push({
        date,
        day: d.getDate(),
        status: statusMap.value[date] ?? 0,
        past: date < todayStr,
      })
      d.setDate(d.getDate() + 1)
    }
    const first = new Date(y, m, 1)
    blocks.push({
      key,
      title: `${y}年${m + 1}月`,
      leadBlanks: (first.getDay() + 6) % 7, // 周一为一周开始
      days,
    })
  }
  return blocks
}

onLoad(() => {
  load()
})

async function load() {
  loading.value = true
  try {
    const list = await chefApi.schedules(fromStr, toStr)
    const map: Record<string, number> = {}
    for (const s of list) map[s.date.slice(0, 10)] = s.status
    statusMap.value = map
  } catch (e) {
    toast(errMsg(e))
  } finally {
    loading.value = false
  }
}

async function toggle(day: DayCell) {
  if (day.past) {
    toast('过去日期不可修改')
    return
  }
  if (day.status === 1) {
    toast('该日期已有订单，暂不可修改')
    return
  }
  if (toggling.value) return
  const toRest = day.status !== 2
  toggling.value = day.date
  try {
    await chefApi.setRestDays([day.date], toRest)
    statusMap.value = { ...statusMap.value, [day.date]: toRest ? 2 : 0 }
    toast(toRest ? `已将 ${day.date.slice(5)} 设为休息日` : `已开放 ${day.date.slice(5)} 可接单`)
  } catch (e) {
    toast(errMsg(e))
  } finally {
    toggling.value = ''
  }
}
</script>

<template>
  <view class="page">
    <view class="card legend-card">
      <view class="legend">
        <view class="legend-item"><view class="dot free" />可约</view>
        <view class="legend-item"><view class="dot rest" />休息日</view>
        <view class="legend-item"><view class="dot locked" />已锁定(有订单)</view>
      </view>
      <view class="legend-tip">点击日期可切换「休息 / 可约」，已有订单的日期不可修改</view>
    </view>

    <view v-for="m in months" :key="m.key" class="card">
      <view class="month-title">{{ m.title }}</view>
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
            locked: d.status === 1,
            rest: d.status === 2,
            today: d.date === todayStr,
            busy: toggling === d.date,
          }"
          @tap="toggle(d)"
        >
          <text class="day-num">{{ d.day }}</text>
          <text v-if="d.status === 1" class="day-tag">有单</text>
          <text v-else-if="d.status === 2" class="day-tag rest-tag">休</text>
          <text v-else-if="d.date === todayStr" class="day-tag free-tag">今天</text>
        </view>
      </view>
    </view>

    <view class="bottom-space" />
  </view>
</template>

<style scoped>
.legend-card {
  padding: 24rpx 28rpx;
}

.legend {
  display: flex;
}

.legend-item {
  display: flex;
  align-items: center;
  font-size: 24rpx;
  color: #595959;
  margin-right: 32rpx;
}

.dot {
  width: 20rpx;
  height: 20rpx;
  border-radius: 6rpx;
  margin-right: 10rpx;
}

.dot.free { background: #e8f7ee; border: 2rpx solid #16a34a; }
.dot.rest { background: #f59e0b; }
.dot.locked { background: #e5e7eb; }

.legend-tip {
  font-size: 22rpx;
  color: #bbbbbb;
  margin-top: 14rpx;
}

.month-title {
  font-size: 30rpx;
  font-weight: 700;
  color: #262626;
  margin-bottom: 20rpx;
}

.week-row,
.grid {
  display: flex;
  flex-wrap: wrap;
}

.week-cell {
  width: 14.28%;
  text-align: center;
  font-size: 24rpx;
  color: #999999;
  padding: 10rpx 0;
}

.cell {
  width: 14.28%;
  min-height: 96rpx;
  padding: 4rpx 2rpx;
  margin-bottom: 8rpx;
  border-radius: 12rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

.cell.blank {
  min-height: 0;
}

.cell.day {
  border: 2rpx solid #e8f7ee;
  background: #f2fbf5;
}

.cell.locked {
  background: #e5e7eb;
  border-color: #e5e7eb;
}

.cell.rest {
  background: #fef3c7;
  border-color: #f59e0b;
}

.cell.today {
  border-color: #d97706;
}

.cell.busy {
  opacity: 0.5;
}

.day-num {
  font-size: 28rpx;
  color: #262626;
}

.cell.locked .day-num {
  color: #9ca3af;
}

.day-tag {
  font-size: 18rpx;
  margin-top: 4rpx;
  color: #16a34a;
}

.cell.locked .day-tag {
  color: #6b7280;
}

.rest-tag {
  color: #b45309;
}

.free-tag {
  color: #d97706;
}

.bottom-space {
  height: 60rpx;
}
</style>
