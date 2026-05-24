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
  if (score >= 80) return '#4ADE80'
  if (score >= 60) return 'var(--gold)'
  return 'var(--crimson)'
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
    <!-- Constellation background -->
    <div class="bg-constellation" aria-hidden="true">
      <svg viewBox="0 0 800 600" preserveAspectRatio="xMidYMid slice" class="constellation-svg">
        <circle cx="100" cy="80" r="1" fill="#D4A84B" opacity="0.3" />
        <circle cx="600" cy="60" r="1.2" fill="#D4A84B" opacity="0.35" />
        <circle cx="700" cy="400" r="1" fill="#D4A84B" opacity="0.25" />
        <circle cx="200" cy="500" r="0.8" fill="#D4A84B" opacity="0.2" />
        <circle cx="400" cy="300" r="1.5" fill="#D4A84B" opacity="0.4" />
        <line
          x1="100"
          y1="80"
          x2="400"
          y2="300"
          stroke="#D4A84B"
          stroke-width="0.4"
          opacity="0.06"
        />
      </svg>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="loading-state">
      <div class="loading-inner">
        <div class="loading-constellation">
          <svg width="60" height="60" viewBox="0 0 60 60" fill="none">
            <circle
              cx="30"
              cy="30"
              r="25"
              stroke="#D4A84B"
              stroke-width="0.5"
              stroke-dasharray="2 3"
              opacity="0.4"
            />
            <circle cx="30" cy="30" r="12" stroke="#D4A84B" stroke-width="0.5" opacity="0.3" />
            <circle cx="30" cy="30" r="3" fill="#D4A84B" opacity="0.3" />
            <circle cx="15" cy="20" r="2" fill="#D4A84B" opacity="0.5" class="star-pulse" />
            <circle
              cx="45"
              cy="18"
              r="2"
              fill="#D4A84B"
              opacity="0.4"
              class="star-pulse"
              style="animation-delay: 0.3s"
            />
          </svg>
        </div>
        <p class="loading-text">本周运势加载中</p>
      </div>
    </div>

    <!-- Error -->
    <div v-else-if="error" class="error-state">
      <div class="error-icon">
        <svg width="60" height="60" viewBox="0 0 60 60" fill="none">
          <circle
            cx="30"
            cy="30"
            r="26"
            stroke="#C41E3A"
            stroke-width="1"
            stroke-dasharray="3 2"
            opacity="0.4"
          />
          <line x1="20" y1="20" x2="40" y2="40" stroke="#C41E3A" stroke-width="2" opacity="0.5" />
          <line x1="40" y1="20" x2="20" y2="40" stroke="#C41E3A" stroke-width="2" opacity="0.5" />
        </svg>
      </div>
      <p class="error-text">{{ error }}</p>
      <button class="btn-retry" @click="fetchWeekly">重新加载</button>
    </div>

    <template v-else-if="data">
      <div class="page-inner">
        <!-- Header -->
        <div class="weekly-header">
          <div class="header-eyebrow">BaZi Fortune</div>
          <h1 class="page-title">本周运势</h1>
          <p class="week-range">{{ weekRange }}</p>
          <div class="score-display glass-panel">
            <div class="score-glow"></div>
            <div class="score-inner">
              <span class="score-number" :style="{ color: scoreColor(data.weekly_score) }">
                {{ data.weekly_score }}
              </span>
              <span class="score-label">综合评分</span>
            </div>
          </div>
        </div>

        <!-- Chart -->
        <div class="chart-section glass-card">
          <FortuneChart :daily-data="trendData" height="280px" />
        </div>

        <!-- Daily Cards -->
        <div class="daily-section">
          <h3 class="section-title">每日概况</h3>
          <div class="daily-cards">
            <div v-for="(day, idx) in data.daily_fortunes" :key="idx" class="day-card">
              <div class="day-card-left">
                <span class="day-date">{{ day.solar_date }}</span>
              </div>
              <span class="day-pillar">{{ day.day_gan_zhi }}</span>
              <p v-if="day.yi_ji" class="day-yiji">{{ day.yi_ji }}</p>
            </div>
          </div>
        </div>

        <div class="bottom-nav">
          <router-link :to="`/fortune?chart_id=${route.query.chart_id}`" class="nav-link">
            查看今日运势 →
          </router-link>
        </div>
      </div>
    </template>
  </div>
</template>

<style scoped>
.weekly-page {
  min-height: 100vh;
  background: var(--bg);
  position: relative;
  overflow: hidden;
}

.bg-constellation {
  position: absolute;
  inset: 0;
  pointer-events: none;
  z-index: 0;
}

.constellation-svg {
  width: 100%;
  height: 100%;
  position: absolute;
  inset: 0;
}

.page-inner {
  position: relative;
  z-index: 1;
  max-width: 540px;
  margin: 0 auto;
  padding: 2rem 1rem;
}

/* Loading */
.loading-state {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 70vh;
}

.loading-inner {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
}

.loading-constellation {
  animation: spin-slow 20s linear infinite;
}

@keyframes spin-slow {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.star-pulse {
  animation: star-twinkle 2s ease-in-out infinite;
}

@keyframes star-twinkle {
  0%,
  100% {
    opacity: 0.3;
    r: 2;
  }
  50% {
    opacity: 0.9;
    r: 3;
  }
}

.loading-text {
  font-size: 12px;
  color: rgba(212, 168, 75, 0.5);
  letter-spacing: 2px;
}

/* Error */
.error-state {
  position: relative;
  z-index: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 70vh;
  gap: 1rem;
}

.error-icon {
  opacity: 0.6;
}

.error-text {
  font-size: 0.9rem;
  color: var(--muted);
  margin: 0;
}

.btn-retry {
  padding: 0.5rem 1.5rem;
  background: linear-gradient(135deg, #c41e3a, #8b0000);
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 0.85rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  box-shadow: 0 4px 16px rgba(196, 30, 58, 0.2);
}

.btn-retry:hover {
  transform: translateY(-1px);
  box-shadow: 0 6px 20px rgba(196, 30, 58, 0.3);
}

/* Header */
.weekly-header {
  text-align: center;
  margin-bottom: 1.5rem;
}

.header-eyebrow {
  font-size: 10px;
  letter-spacing: 3px;
  color: rgba(212, 168, 75, 0.35);
  text-transform: uppercase;
  margin-bottom: 8px;
}

.page-title {
  font-family: var(--font-serif), serif;
  font-size: 1.8rem;
  font-weight: 700;
  color: var(--text);
  margin: 0 0 6px;
  letter-spacing: 3px;
}

.week-range {
  font-size: 12px;
  color: var(--muted);
  margin: 0 0 1rem 0;
}

.score-display {
  display: inline-block;
  padding: 1.25rem 2.5rem;
  text-align: center;
  position: relative;
  overflow: hidden;
}

.score-glow {
  position: absolute;
  inset: 0;
  background: radial-gradient(circle at 50% 50%, rgba(212, 168, 75, 0.06), transparent 70%);
  pointer-events: none;
}

.score-inner {
  position: relative;
}

.score-number {
  font-size: 3.5rem;
  font-weight: 900;
  line-height: 1;
  text-shadow: 0 0 30px currentColor;
}

.score-label {
  display: block;
  font-size: 0.75rem;
  color: var(--muted);
  margin-top: 0.3rem;
  letter-spacing: 1px;
}

/* Chart */
.chart-section {
  padding: 1rem;
  margin-bottom: 1.25rem;
}

/* Daily Cards */
.daily-section {
  margin-bottom: 1.25rem;
}

.section-title {
  font-size: 0.95rem;
  font-weight: 700;
  color: var(--text);
  margin: 0 0 0.75rem;
  padding-bottom: 0.5rem;
  border-bottom: 1px solid rgba(212, 168, 75, 0.1);
  letter-spacing: 1px;
}

.daily-cards {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.day-card {
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(212, 168, 75, 0.08);
  border-radius: 10px;
  padding: 0.75rem 1rem;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  transition: all 0.2s;
}

.day-card:hover {
  border-color: rgba(212, 168, 75, 0.2);
  background: rgba(212, 168, 75, 0.03);
}

.day-card-left {
  flex: 1;
}

.day-date {
  font-size: 0.82rem;
  font-weight: 600;
  color: var(--text);
}

.day-pillar {
  font-size: 1rem;
  font-weight: 800;
  color: var(--crimson);
  text-shadow: 0 0 10px rgba(196, 30, 58, 0.3);
  min-width: 48px;
  text-align: center;
}

.day-yiji {
  font-size: 0.7rem;
  color: var(--muted);
  margin: 0;
  flex: 2;
  text-align: right;
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
  color: var(--gold);
  text-decoration: none;
  font-size: 0.85rem;
  font-weight: 500;
  transition: all 0.2s;
}

.nav-link:hover {
  text-shadow: 0 0 12px rgba(212, 168, 75, 0.4);
}
</style>
