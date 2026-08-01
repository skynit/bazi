<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { fetchChart, fetchCharts } from '../api/chart'
import { getApiErrorMessage } from '../api/client'
import BaziChart from '../components/BaziChart.vue'
import BirthInputForm from '../components/BirthInputForm.vue'
import { Button } from '@/components/ui/button'

interface SavedChart {
  id: number
  name: string
  gender: string
  zi_hour_policy: 'late_zi_next_day' | 'late_zi_same_day'
  birth_year: number
  birth_month: number
  birth_day: number
  birth_hour: number
  birth_min: number
  birth_sec: number
  calendar_type?: string
  lunar_leap_month?: boolean
  birth_place?: string
  timezone?: string
  birth_utc_offset_seconds?: number
  longitude?: number
  use_true_solar_time?: boolean
  time_uncertain?: boolean
  uncertainty_seconds?: number
}

const route = useRoute()
const router = useRouter()
const isNew = computed(() => route.params.id === 'new')
const chartData = ref<any>(null)
const loading = ref(false)
const error = ref('')

const savedCharts = ref<SavedChart[]>([])
const showPicker = ref(false)
const chartsLoading = ref(false)
const chartsError = ref('')

onMounted(async () => {
  tryLoadChart()
})

watch(
  () => route.fullPath,
  () => {
    tryLoadChart()
  },
)

function tryLoadChart() {
  if (route.params.id === 'new') {
    const raw = sessionStorage.getItem('lastChart')
    if (raw) {
      chartData.value = JSON.parse(raw)
      sessionStorage.removeItem('lastChart')
    } else {
      fetchSavedCharts()
    }
  } else {
    loadChart()
  }
}

function cacheLastBirthFromChart(chart: any) {
  if (!chart?.id || !chart?.birth_year || !chart?.birth_month || !chart?.birth_day) return
  const birthHour = Number(chart.birth_hour)
  const genderRaw = String(chart.gender || '').toLowerCase()
  localStorage.setItem(
    'bazi_last_birth',
    JSON.stringify({
      year: Number(chart.birth_year),
      month: Number(chart.birth_month),
      day: Number(chart.birth_day),
      hour: Number.isFinite(birthHour) ? birthHour : 8,
      minute: Number(chart.birth_min) || 0,
      second: Number(chart.birth_sec) || 0,
      calendarType: chart.calendar_type === 'LUNAR' ? 'LUNAR' : 'SOLAR',
      lunarLeapMonth: Boolean(chart.lunar_leap_month),
      gender: genderRaw === 'female' || genderRaw === '女' ? 'FEMALE' : 'MALE',
      ziHourPolicy:
        chart.zi_hour_policy === 'late_zi_same_day' ? 'late_zi_same_day' : 'late_zi_next_day',
      birthPlace: chart.birth_place || '',
      timezone: chart.timezone || Intl.DateTimeFormat().resolvedOptions().timeZone,
      birthUTCOffsetSeconds: chart.birth_utc_offset_seconds,
      longitude: chart.longitude,
      useTrueSolarTime: Boolean(chart.use_true_solar_time),
      timeUncertain: Boolean(chart.time_uncertain),
      uncertaintySeconds: Number(chart.uncertainty_seconds) || 0,
      chartId: chart.id,
    }),
  )
}

async function fetchSavedCharts() {
  chartsLoading.value = true
  chartsError.value = ''
  try {
    const data = await fetchCharts(1, 10)
    savedCharts.value = data.charts
    showPicker.value = data.charts.length > 0
  } catch (reason: unknown) {
    chartsError.value = getApiErrorMessage(reason, '已有命盘加载失败，请稍后重试。')
    savedCharts.value = []
    showPicker.value = false
  } finally {
    chartsLoading.value = false
  }
}

async function selectChart(chart: SavedChart) {
  error.value = ''
  try {
    localStorage.setItem(
      'bazi_last_birth',
      JSON.stringify({
        year: chart.birth_year,
        month: chart.birth_month,
        day: chart.birth_day,
        hour: chart.birth_hour,
        minute: chart.birth_min || 0,
        second: chart.birth_sec || 0,
        calendarType: chart.calendar_type === 'LUNAR' ? 'LUNAR' : 'SOLAR',
        lunarLeapMonth: Boolean(chart.lunar_leap_month),
        gender:
          chart.gender.toLowerCase() === 'female' || chart.gender === '女' ? 'FEMALE' : 'MALE',
        ziHourPolicy:
          chart.zi_hour_policy === 'late_zi_same_day' ? 'late_zi_same_day' : 'late_zi_next_day',
        birthPlace: chart.birth_place || '',
        timezone: chart.timezone || Intl.DateTimeFormat().resolvedOptions().timeZone,
        birthUTCOffsetSeconds: chart.birth_utc_offset_seconds,
        longitude: chart.longitude,
        useTrueSolarTime: Boolean(chart.use_true_solar_time),
        timeUncertain: Boolean(chart.time_uncertain),
        uncertaintySeconds: Number(chart.uncertainty_seconds) || 0,
        chartId: chart.id,
      }),
    )
    router.push(`/chart/${chart.id}`)
  } catch {
    error.value = '打开命盘失败，请稍后重试。'
  }
}

function startNewChart() {
  showPicker.value = false
}

function formatBirth(c: SavedChart): string {
  const m = String(c.birth_month).padStart(2, '0')
  const d = String(c.birth_day).padStart(2, '0')
  const h = String(c.birth_hour).padStart(2, '0')
  const min = String(c.birth_min || 0).padStart(2, '0')
  const sec = String(c.birth_sec || 0).padStart(2, '0')
  const calendar = c.calendar_type === 'LUNAR' ? `农历${c.lunar_leap_month ? '闰月' : ''}` : '公历'
  return `${calendar} ${c.birth_year}-${m}-${d} ${h}:${min}:${sec}`
}

function formatGender(value: string): string {
  const gender = String(value || '').toUpperCase()
  return gender === 'FEMALE' || gender === '女' ? '女' : '男'
}

async function loadChart() {
  loading.value = true
  error.value = ''
  try {
    const chart = await fetchChart(String(route.params.id))
    chartData.value = chart
    cacheLastBirthFromChart(chart)
  } catch (reason: unknown) {
    error.value = getApiErrorMessage(reason, '命盘加载失败，请稍后重试。')
  } finally {
    loading.value = false
  }
}

function goFortune() {
  router.push(`/fortune?chart_id=${chartData.value.id}`)
}

function goZiWei() {
  router.push(`/ziwei/${chartData.value.id}`)
}
</script>
<template>
  <div class="chart-page">
    <main class="page-content">
      <!-- Loading state -->
      <div v-if="loading" class="loading-state">
        <div class="loading-inner">
          <div class="loading-constellation">
            <svg width="80" height="80" viewBox="0 0 80 80" fill="none">
              <circle
                cx="40"
                cy="40"
                r="35"
                stroke="currentColor"
                stroke-width="0.5"
                stroke-dasharray="2 3"
                opacity="0.4"
              />
              <circle
                cx="40"
                cy="40"
                r="20"
                stroke="currentColor"
                stroke-width="0.5"
                stroke-dasharray="1 4"
                opacity="0.3"
              />
              <circle cx="40" cy="40" r="4" fill="currentColor" opacity="0.3" />
              <circle cx="20" cy="25" r="2" fill="currentColor" opacity="0.6" class="star-pulse" />
              <circle
                cx="60"
                cy="22"
                r="2.5"
                fill="currentColor"
                opacity="0.5"
                class="star-pulse"
                style="animation-delay: 0.3s"
              />
              <circle
                cx="62"
                cy="55"
                r="2"
                fill="currentColor"
                opacity="0.4"
                class="star-pulse"
                style="animation-delay: 0.6s"
              />
            </svg>
          </div>
          <p class="loading-text">命盘加载中</p>
        </div>
      </div>

      <!-- Error state -->
      <div v-else-if="error" class="error-state">
        <div class="error-icon">
          <svg width="80" height="80" viewBox="0 0 80 80" fill="none">
            <circle
              cx="40"
              cy="40"
              r="35"
              stroke="currentColor"
              stroke-width="1"
              stroke-dasharray="4 3"
              opacity="0.4"
            />
            <line
              x1="26"
              y1="26"
              x2="54"
              y2="54"
              stroke="currentColor"
              stroke-width="2.5"
              opacity="0.5"
            />
            <line
              x1="54"
              y1="26"
              x2="26"
              y2="54"
              stroke="currentColor"
              stroke-width="2.5"
              opacity="0.5"
            />
          </svg>
        </div>
        <p class="error-title">{{ error }}</p>
        <button class="btn-retry" @click="loadChart">重新加载</button>
      </div>

      <!-- New chart: picker or form -->
      <div v-else-if="isNew && !chartData" class="new-chart-state">
        <!-- Loading saved charts -->
        <div v-if="chartsLoading" class="picker-loading">
          <div class="skeleton h-5 w-40 mb-6"></div>
          <div class="skeleton h-[72px] rounded-xl mb-3" v-for="i in 3" :key="i"></div>
        </div>

        <!-- Error loading -->
        <div v-else-if="chartsError" class="picker-error">
          <p class="picker-error-text">{{ chartsError }}</p>
          <button class="btn-retry" @click="fetchSavedCharts">重新加载</button>
        </div>

        <!-- Saved chart picker -->
        <div v-else-if="showPicker" class="picker-section">
          <div class="picker-header">
            <span class="badge-dot"></span>
            选择已有命盘
          </div>
          <div class="picker-list">
            <button
              v-for="chart in savedCharts"
              :key="chart.id"
              type="button"
              class="picker-row glass-panel"
              @click="selectChart(chart)"
            >
              <div class="picker-avatar">
                {{ chart.name?.charAt(0) || '?' }}
              </div>
              <div class="picker-info">
                <span class="picker-name">{{ chart.name || '未命名' }}</span>
                <span class="picker-meta">
                  <span class="meta-tag">{{ formatGender(chart.gender) }}</span>
                  <span class="meta-sep">·</span>
                  {{ formatBirth(chart) }}
                </span>
              </div>
              <svg class="picker-arrow" width="14" height="14" viewBox="0 0 14 14" fill="none">
                <path
                  d="M5 3l4 4-4 4"
                  stroke="currentColor"
                  stroke-width="1.5"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                />
              </svg>
            </button>
          </div>
          <div class="picker-divider">
            <span class="divider-line"></span>
            <span class="divider-text">或者</span>
            <span class="divider-line"></span>
          </div>
          <button class="btn-new-chart" @click="startNewChart">
            <span class="btn-icon">✦</span>
            添加排盘
          </button>
        </div>

        <!-- No saved charts: show form directly -->
        <template v-else>
          <div class="new-chart-badge">
            <span class="badge-dot"></span>
            新建命盘
          </div>
          <BirthInputForm />
        </template>
      </div>

      <!-- Chart display -->
      <div v-else-if="chartData" class="chart-result">
        <BaziChart :chart="chartData" />

        <!-- Action buttons -->
        <div class="action-row">
          <div class="action-glow"></div>
          <Button
            @click="goFortune"
            class="rounded-full h-10 px-6 text-sm font-medium bg-foreground text-background hover:bg-foreground/90"
            >查看运势</Button
          >
          <button class="btn-secondary" @click="goZiWei">
            <span class="btn-icon-secondary">☯</span>
            紫微斗数
          </button>
        </div>

        <div class="rechart-link">
          <router-link to="/chart/new" class="link-text">
            重新排盘
            <svg width="12" height="12" viewBox="0 0 12 12" fill="none">
              <path
                d="M2 6h8M7 3l3 3-3 3"
                stroke="currentColor"
                stroke-width="1.5"
                stroke-linecap="round"
                stroke-linejoin="round"
              />
            </svg>
          </router-link>
        </div>
      </div>

      <!-- Not found -->
      <div v-else-if="!isNew" class="empty-state">
        <div class="empty-icon">
          <svg width="80" height="80" viewBox="0 0 80 80" fill="none">
            <circle
              cx="40"
              cy="40"
              r="35"
              stroke="currentColor"
              stroke-width="0.5"
              stroke-dasharray="2 3"
              opacity="0.3"
            />
            <circle cx="40" cy="40" r="4" fill="currentColor" opacity="0.3" />
            <circle cx="20" cy="25" r="2" fill="currentColor" opacity="0.4" />
            <circle cx="60" cy="22" r="2.5" fill="currentColor" opacity="0.3" />
            <circle cx="62" cy="55" r="2" fill="currentColor" opacity="0.35" />
          </svg>
        </div>
        <p class="empty-title">未找到命盘</p>
        <Button @click="router.push('/chart/new')" variant="outline" class="rounded-full"
          >创建新的命盘</Button
        >
      </div>
    </main>
  </div>
</template>

<style scoped>
.chart-page {
  min-height: 100vh;
  background: transparent;
  position: relative;
  overflow: hidden;
}

/* Background */
.computation-wrapper {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 60vh;
}

/* Header */
.chart-header {
  position: relative;
  z-index: 2;
  border-bottom: 1px solid var(--line-subtle);
  background: var(--glass-nav);
  backdrop-filter: blur(12px);
}

.header-inner {
  max-width: 900px;
  margin: 0 auto;
  padding: 0 2rem;
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.back-link {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: var(--fs-sm);
  color: var(--text-muted);
  text-decoration: none;
  transition: color 0.2s;
  letter-spacing: 1px;
}

.back-link:hover {
  color: var(--accent);
}

.header-title-block {
  text-align: center;
}

.header-eyebrow {
  font-size: var(--fs-2xs);
  letter-spacing: 3px;
  color: var(--text-soft);
  text-transform: uppercase;
}

.header-title {
  font-family: var(--font-serif), serif;
  font-size: var(--fs-lg);
  font-weight: 700;
  color: var(--text);
  margin: 0;
  letter-spacing: 3px;
}

.header-spacer {
  width: 80px;
}

/* Page content */
.page-content {
  position: relative;
  z-index: 1;
  max-width: 860px;
  margin: 0 auto;
  padding: 2rem 1.5rem 3rem;
}

/* Loading */
.loading-state {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 60vh;
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
  font-size: var(--fs-2xs);
  color: var(--text-muted);
  letter-spacing: 2px;
}

/* Error */
.error-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 60vh;
  gap: 1.25rem;
}

.error-icon {
  color: var(--danger);
  opacity: 0.6;
}

.error-title {
  font-size: var(--fs-sm);
  color: var(--text-muted);
  text-align: center;
  margin: 0;
}

.btn-retry {
  padding: 0.5rem 1.75rem;
  background: linear-gradient(135deg, var(--crimson), #be123c);
  color: var(--destructive-foreground);
  border: none;
  border-radius: 8px;
  font-size: var(--fs-sm);
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  box-shadow: 0 4px 16px rgba(251, 113, 133, 0.25);
}

.btn-retry:hover {
  transform: translateY(-1px);
  box-shadow: 0 6px 20px rgba(251, 113, 133, 0.35);
}

/* New chart state */
.new-chart-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1.5rem;
  padding-top: 1rem;
}

.new-chart-badge {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 0.4rem 1rem;
  background: var(--accent-dim);
  border: 1px solid var(--line-strong);
  border-radius: 20px;
  font-size: var(--fs-xs);
  color: var(--text-muted);
  letter-spacing: 1px;
}

.badge-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--accent);
  box-shadow: 0 0 8px var(--accent-glow);
  animation: pulse-dot 2s ease-in-out infinite;
}

@keyframes pulse-dot {
  0%,
  100% {
    opacity: 1;
    transform: scale(1);
  }
  50% {
    opacity: 0.6;
    transform: scale(0.85);
  }
}

/* Chart result */
.chart-result {
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.action-row {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 1rem;
  position: relative;
}

.action-glow {
  position: absolute;
  width: 300px;
  height: 60px;
  background: radial-gradient(circle, var(--accent-dim), transparent 70%);
  pointer-events: none;
}

.btn-primary {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 0.875rem 2.5rem;
  background: linear-gradient(135deg, var(--accent), #94a3b8);
  color: var(--bg);
  font-weight: 700;
  font-size: var(--fs-sm);
  border: none;
  border-radius: 50px;
  cursor: pointer;
  transition: all 0.3s;
  box-shadow: 0 4px 24px var(--accent-glow);
  text-decoration: none;
  letter-spacing: 1px;
}

.btn-primary:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 36px var(--accent-glow);
}

.btn-secondary {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 0.875rem 2rem;
  background: transparent;
  color: var(--accent);
  font-weight: 600;
  font-size: var(--fs-sm);
  border: 1px solid var(--line-strong);
  border-radius: 50px;
  cursor: pointer;
  transition: all 0.3s;
  letter-spacing: 1px;
}

.btn-secondary:hover {
  border-color: var(--text-muted);
  background: var(--accent-dim);
  box-shadow: 0 0 20px var(--accent-dim);
}

.btn-icon {
  font-size: var(--fs-xs);
  animation: spin-slow 8s linear infinite;
}

.btn-icon-secondary {
  font-size: var(--fs-body);
}

/* Rechart link */
.rechart-link {
  text-align: center;
}

.link-text {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: var(--fs-sm);
  color: var(--text-muted);
  text-decoration: none;
  transition: color 0.2s;
}

.link-text:hover {
  color: var(--accent);
}

/* Empty state */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 60vh;
  gap: 1.25rem;
  text-align: center;
}

.empty-icon {
  color: var(--icon-muted);
  opacity: 0.5;
  margin-bottom: 0.5rem;
}

.empty-title {
  font-size: var(--fs-body);
  color: var(--text-muted);
  margin: 0 0 1.5rem;
}

/* Picker states */
.picker-loading {
  width: 100%;
  max-width: 480px;
}

.picker-error {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
  padding: 2rem 0;
}

.picker-error-text {
  font-size: var(--fs-sm);
  color: var(--text-muted);
  text-align: center;
  margin: 0;
}

/* Picker section */
.picker-section {
  width: 100%;
  max-width: 500px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
}

.picker-header {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 0.4rem 1rem;
  background: var(--accent-dim);
  border: 1px solid var(--line-strong);
  border-radius: 20px;
  font-size: var(--fs-xs);
  color: var(--text-muted);
  letter-spacing: 1px;
}

.picker-list {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 0.625rem;
}

.picker-row {
  display: flex;
  align-items: center;
  gap: 0.875rem;
  padding: 0.875rem 1rem;
  width: 100%;
  color: var(--text);
  font: inherit;
  text-align: left;
  cursor: pointer;
  transition: all 0.25s ease;
}

.picker-row:focus-visible {
  outline: 2px solid var(--line-focus);
  outline-offset: 2px;
}

.picker-row:hover {
  border-color: var(--text-soft);
  transform: translateY(-1px);
  box-shadow:
    var(--shadow-md),
    0 0 16px var(--accent-dim);
}

.picker-avatar {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  background: linear-gradient(135deg, var(--accent), #94a3b8);
  color: var(--bg);
  font-size: var(--fs-body);
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.picker-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}

.picker-name {
  font-size: var(--fs-sm);
  font-weight: 600;
  color: var(--text);
}

.picker-meta {
  font-size: var(--fs-xs);
  color: var(--text-muted);
  display: flex;
  align-items: center;
  gap: 0.3rem;
}

.picker-arrow {
  color: var(--text-soft);
  flex-shrink: 0;
  transition:
    color 0.2s,
    transform 0.2s;
}

.picker-row:hover .picker-arrow {
  color: var(--accent);
  transform: translateX(2px);
}

/* Divider */
.picker-divider {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 0.25rem 0;
}

.divider-line {
  flex: 1;
  height: 1px;
  background: linear-gradient(90deg, transparent, var(--line-strong), transparent);
}

.divider-text {
  font-size: var(--fs-xs);
  color: var(--text-soft);
  letter-spacing: 1px;
}

/* New chart button */
.btn-new-chart {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 0.75rem 2.5rem;
  background: transparent;
  color: var(--accent);
  font-weight: 600;
  font-size: var(--fs-sm);
  border: 1px solid var(--line-strong);
  border-radius: 50px;
  cursor: pointer;
  transition: all 0.3s;
  letter-spacing: 1px;
}

.btn-new-chart:hover {
  border-color: var(--text-muted);
  background: var(--accent-dim);
  box-shadow: 0 0 20px var(--accent-dim);
  transform: translateY(-1px);
}

@media (max-width: 640px) {
  .action-row {
    flex-direction: column;
    gap: 0.75rem;
  }
  .btn-primary,
  .btn-secondary {
    width: 100%;
    justify-content: center;
  }
}
</style>
