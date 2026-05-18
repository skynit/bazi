<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import client from '../api/client'
import FortuneChart, { type TrendPoint } from '../components/FortuneChart.vue'

interface ElementImage {
  element: string
  image_url: string
  description: string
}

interface FortuneDay {
  solar_date: string
  day_gan_zhi: string
  yi_ji?: string
  element_images?: ElementImage[]
}

interface WeeklyResponse {
  daily_fortunes: FortuneDay[]
  weekly_score: number
  element_trend: string
}

const route = useRoute()

const data = ref<WeeklyResponse | null>(null)
const loading = ref(true)
const error = ref('')

const trendData = computed<TrendPoint[]>(() => {
  if (!data.value?.element_trend) return []
  try {
    return JSON.parse(data.value.element_trend) as TrendPoint[]
  } catch {
    return []
  }
})

const weekRange = computed(() => {
  if (!data.value?.daily_fortunes?.length) return ''
  const first = data.value.daily_fortunes[0].solar_date
  const last = data.value.daily_fortunes[data.value.daily_fortunes.length - 1].solar_date
  return `${first} ~ ${last}`
})

function todayStr(): string {
  const d = new Date()
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

function scoreColor(score: number): string {
  if (score >= 80) return '#228B22'
  if (score >= 60) return '#DAA520'
  return '#DC143C'
}

async function fetchWeekly() {
  const chartId = route.query.chart_id
  if (!chartId) {
    error.value = '请提供 chart_id 参数'
    loading.value = false
    return
  }

  try {
    const { data: res } = await client.post<WeeklyResponse>('/fortune/weekly', {
      chart_id: Number(chartId),
      start_date: todayStr(),
    })
    data.value = res
  } catch (e: any) {
    error.value = e.response?.data?.error || '加载周运势失败'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchWeekly()
})
</script>

<template>
  <div class="weekly-page">
    <!-- Loading -->
    <div v-if="loading" class="p-8 space-y-4">
      <div class="skeleton h-8 w-48 mx-auto mb-2"></div>
      <div class="skeleton h-56 rounded-xl"></div>
      <div class="skeleton h-14 rounded-lg" v-for="i in 4" :key="i"></div>
    </div>
    <div v-else-if="error" class="state-box">
      <el-result icon="error" title="加载失败" sub-title="请检查网络连接后重试">
        <template #extra>
          <el-button type="primary" @click="fetchWeekly">重试</el-button>
        </template>
      </el-result>
    </div>

    <template v-else-if="data">
      <!-- Header -->
      <div class="weekly-header">
        <h1 class="page-title">本周运势</h1>
        <p class="week-range">{{ weekRange }}</p>
        <div class="score-display">
          <span
            class="score-number"
            :style="{ color: scoreColor(data.weekly_score) }"
          >
            {{ data.weekly_score }}
          </span>
          <span class="score-label">综合评分</span>
        </div>
      </div>

      <!-- Chart -->
      <div class="chart-section">
        <FortuneChart :daily-data="trendData" height="300px" />
      </div>

      <!-- Daily Mini Cards -->
      <div class="daily-cards">
        <h3 class="section-label">每日概况</h3>
        <div
          v-for="(day, idx) in data.daily_fortunes"
          :key="idx"
          class="day-card"
        >
          <div class="day-card-header">
            <span class="day-date">{{ day.solar_date }}</span>
            <span class="day-pillar">{{ day.day_gan_zhi }}</span>
          </div>
          <p v-if="day.yi_ji" class="day-yiji">{{ day.yi_ji }}</p>
        </div>
      </div>

      <div class="bottom-nav">
        <router-link
          :to="`/fortune?chart_id=${route.query.chart_id}`"
          class="nav-link"
        >
          查看今日运势
        </router-link>
      </div>
    </template>
  </div>
</template>

<style scoped>
.weekly-page {
  min-height: 100vh;
  background: #FAF8F3;
  padding: 1.25rem 1rem;
  max-width: 540px;
  margin: 0 auto;
}

/* States */
.state-box {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 60vh;
  gap: 0.75rem;
}

.state-text {
  font-size: 1rem;
  color: var(--color-bazi-ink);
  margin: 0;
}

.state-text.error {
  color: var(--color-bazi-red);
}

.back-link {
  color: var(--color-bazi-red);
  text-decoration: none;
  font-size: 0.9rem;
}

/* Header */
.weekly-header {
  text-align: center;
  margin-bottom: 1.5rem;
}

.page-title {
  font-size: 1.35rem;
  font-weight: 800;
  color: var(--color-bazi-ink);
  margin: 0 0 0.25rem 0;
}

.week-range {
  font-size: 0.82rem;
  color: #999;
  margin: 0 0 0.75rem 0;
}

.score-display {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.score-number {
  font-size: 3rem;
  font-weight: 900;
  line-height: 1.1;
}

.score-label {
  font-size: 0.8rem;
  color: #999;
  margin-top: 0.2rem;
}

/* Chart */
.chart-section {
  background: white;
  border: 1px solid #E8E3D5;
  border-radius: 0.75rem;
  padding: 0.75rem;
  margin-bottom: 1.25rem;
}

/* Daily Mini Cards */
.section-label {
  font-size: 0.95rem;
  font-weight: 700;
  color: var(--color-bazi-ink);
  margin: 0 0 0.625rem 0;
}

.daily-cards {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  margin-bottom: 1.25rem;
}

.day-card {
  background: white;
  border: 1px solid #E8E3D5;
  border-radius: 0.625rem;
  padding: 0.625rem 0.875rem;
  transition: border-color 0.15s;
}

.day-card:hover {
  border-color: var(--color-bazi-red);
}

.day-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.day-date {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--color-bazi-ink);
}

.day-pillar {
  font-size: 0.9rem;
  font-weight: 700;
  color: var(--color-bazi-red);
}

.day-yiji {
  font-size: 0.72rem;
  color: #888;
  margin: 0.3rem 0 0 0;
  line-height: 1.3;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* Bottom Nav */
.bottom-nav {
  text-align: center;
  padding: 0.5rem 0;
}

.nav-link {
  color: var(--color-bazi-red);
  text-decoration: none;
  font-size: 0.85rem;
  font-weight: 500;
}
</style>
