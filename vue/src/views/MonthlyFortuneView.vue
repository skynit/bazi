<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import client from '../api/client'
import FortuneChart, { type TrendPoint } from '../components/FortuneChart.vue'
import ShaderBackground from '../components/ShaderBackground.vue'

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

interface MonthlyResponse {
  daily_fortunes: FortuneDay[]
  weekly_score: number
  element_trend: string
}

const route = useRoute()

const data = ref<MonthlyResponse | null>(null)
const loading = ref(true)
const error = ref('')
const chartId = ref<string | number>('')

const trendData = computed<TrendPoint[]>(() => {
  if (!data.value?.element_trend) return []
  try {
    return JSON.parse(data.value.element_trend) as TrendPoint[]
  } catch {
    return []
  }
})

const monthLabel = computed(() => {
  if (!data.value?.daily_fortunes?.length) return ''
  const first = data.value.daily_fortunes[0].solar_date
  const parts = first.split('-')
  if (parts.length >= 2) return `${parts[0]}年${parseInt(parts[1])}月`
  return first
})

const monthRange = computed(() => {
  if (!data.value?.daily_fortunes?.length) return ''
  const first = data.value.daily_fortunes[0].solar_date
  const last = data.value.daily_fortunes[data.value.daily_fortunes.length - 1].solar_date
  return `${first} ~ ${last}`
})

function currentYearMonth(): { year: number; month: number } {
  const d = new Date()
  return { year: d.getFullYear(), month: d.getMonth() + 1 }
}

function scoreColor(score: number): string {
  if (score >= 80) return '#4ADE80'
  if (score >= 60) return 'var(--accent)'
  return 'var(--danger)'
}

async function fetchMonthly() {
  let cid: string | number | null = route.query.chart_id as string | null
  if (!cid) {
    try { const s = localStorage.getItem('bazi_last_birth'); if (s) cid = JSON.parse(s).chartId } catch {}
    if (!cid) { error.value = '请先创建命盘'; loading.value = false; return }
  }
  chartId.value = cid

  const { year, month } = currentYearMonth()

  try {
    const { data: res } = await client.post<MonthlyResponse>('/fortune/monthly', {
      chart_id: Number(chartId.value),
      year,
      month,
    })
    data.value = res
  } catch (e: any) {
    error.value = e.response?.data?.error || '加载月运势失败'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchMonthly()
})
</script>

<template>
  <div class="monthly-page">
    <ShaderBackground yongshen="tu" shader-type="grainGradient" :overlay-opacity="0.66" />

    <!-- Loading -->
    <div v-if="loading" class="loading-state">
      <div class="loading-inner">
        <div class="loading-constellation">
          <svg width="60" height="60" viewBox="0 0 60 60" fill="none">
            <circle
              cx="30"
              cy="30"
              r="25"
              stroke="currentColor"
              stroke-width="0.5"
              stroke-dasharray="2 3"
              opacity="0.4"
            />
            <circle cx="30" cy="30" r="12" stroke="currentColor" stroke-width="0.5" opacity="0.3" />
            <circle cx="30" cy="30" r="3" fill="currentColor" opacity="0.3" />
            <circle cx="15" cy="20" r="2" fill="currentColor" opacity="0.5" class="star-pulse" />
            <circle
              cx="45"
              cy="18"
              r="2"
              fill="currentColor"
              opacity="0.4"
              class="star-pulse"
              style="animation-delay: 0.3s"
            />
          </svg>
        </div>
        <p class="loading-text">月运势加载中</p>
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
            stroke="currentColor"
            stroke-width="1"
            stroke-dasharray="3 2"
            opacity="0.4"
          />
          <line x1="20" y1="20" x2="40" y2="40" stroke="currentColor" stroke-width="2" opacity="0.5" />
          <line x1="40" y1="20" x2="20" y2="40" stroke="currentColor" stroke-width="2" opacity="0.5" />
        </svg>
      </div>
      <p class="error-text">{{ error }}</p>
      <router-link to="/chart/new" class="btn-retry">去排盘</router-link>
    </div>

    <template v-else-if="data">
      <div class="page-inner">
        <!-- Header -->
        <div class="monthly-header">
          <div class="header-eyebrow">BaZi Fortune</div>
          <h1 class="page-title">{{ monthLabel }} 月运势</h1>
          <p class="month-range">{{ monthRange }}</p>
          <div class="score-display glass-panel">
            <div class="score-glow"></div>
            <div class="score-inner">
              <span class="score-number" :style="{ color: scoreColor(data.weekly_score) }">
                {{ data.weekly_score }}
              </span>
              <span class="score-label">月综合评分</span>
            </div>
          </div>
        </div>

        <!-- Chart -->
        <div class="chart-section glass-card">
          <FortuneChart :daily-data="trendData" height="280px" />
        </div>

        <!-- Scrollable Daily Cards -->
        <div class="daily-section">
          <h3 class="section-title">每日概况 ({{ data.daily_fortunes.length }}天)</h3>
          <div class="daily-scroll">
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
          <router-link :to="`/fortune?chart_id=${chartId}`" class="nav-link">
            今日运势
          </router-link>
          <span class="nav-sep">·</span>
          <router-link :to="`/fortune/weekly?chart_id=${chartId}`" class="nav-link">
            本周运势
          </router-link>
        </div>
      </div>
    </template>
  </div>
</template>

<style scoped>
.monthly-page {
  min-height: 100vh;
  background: transparent;
  position: relative;
  overflow: hidden;
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
  color: var(--icon-muted);
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
  color: var(--text-muted);
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
  color: var(--danger);
  opacity: 0.6;
}

.error-text {
  font-size: 0.9rem;
  color: var(--text-muted);
  margin: 0;
}

.btn-retry {
  padding: 0.5rem 1.5rem;
  background: linear-gradient(135deg, #fb7185, #be123c);
  color: var(--destructive-foreground);
  border: none;
  border-radius: 8px;
  font-size: 0.85rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  box-shadow: 0 4px 16px rgba(251, 113, 133, 0.2);
}

.btn-retry:hover {
  transform: translateY(-1px);
  box-shadow: 0 6px 20px rgba(251, 113, 133, 0.3);
}

/* Header */
.monthly-header {
  text-align: center;
  margin-bottom: 1.5rem;
}

.header-eyebrow {
  font-size: 10px;
  letter-spacing: 3px;
  color: var(--text-soft);
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

.month-range {
  font-size: 12px;
  color: var(--text-muted);
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
  background: radial-gradient(circle at 50% 50%, rgba(203, 213, 225, 0.06), transparent 70%);
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
  color: var(--text-muted);
  margin-top: 0.3rem;
  letter-spacing: 1px;
}

/* Chart */
.chart-section {
  padding: 1rem;
  margin-bottom: 1.25rem;
}

/* Daily Section */
.section-title {
  font-size: 0.95rem;
  font-weight: 700;
  color: var(--text);
  margin: 0 0 0.75rem;
  padding-bottom: 0.5rem;
  border-bottom: 1px solid rgba(203, 213, 225, 0.1);
  letter-spacing: 1px;
}

.daily-scroll {
  max-height: 480px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
  padding-right: 0.25rem;
}

.daily-scroll::-webkit-scrollbar {
  width: 4px;
}
.daily-scroll::-webkit-scrollbar-track {
  background: transparent;
}
.daily-scroll::-webkit-scrollbar-thumb {
  background: rgba(203, 213, 225, 0.2);
  border-radius: 2px;
}
.daily-scroll::-webkit-scrollbar-thumb:hover {
  background: rgba(203, 213, 225, 0.4);
}

.day-card {
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(203, 213, 225, 0.08);
  border-radius: 10px;
  padding: 0.625rem 1rem;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex-shrink: 0;
  transition: all 0.2s;
}

.day-card:hover {
  border-color: rgba(203, 213, 225, 0.2);
  background: rgba(203, 213, 225, 0.03);
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
  color: var(--danger);
  text-shadow: 0 0 10px rgba(251, 113, 133, 0.3);
  min-width: 48px;
  text-align: center;
}

.day-yiji {
  font-size: 0.7rem;
  color: var(--text-muted);
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
  padding: 1rem 0 0.5rem;
  display: flex;
  justify-content: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.nav-link {
  color: var(--accent);
  text-decoration: none;
  font-size: 0.82rem;
  font-weight: 500;
  transition: all 0.2s;
}

.nav-link:hover {
  text-shadow: 0 0 12px rgba(203, 213, 225, 0.4);
}

.nav-sep {
  color: var(--text-soft);
  font-size: 0.8rem;
}
</style>
