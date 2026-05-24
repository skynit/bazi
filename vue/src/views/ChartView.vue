<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import client from '../api/client'
import BaziChart from '../components/BaziChart.vue'
import BirthInputForm from '../components/BirthInputForm.vue'

interface SavedChart {
  id: number
  name: string
  gender: string
  birth_year: number
  birth_month: number
  birth_day: number
  birth_hour: number
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

async function fetchSavedCharts() {
  chartsLoading.value = true
  chartsError.value = ''
  try {
    const { data } = await client.get('/charts', {
      params: { page: 1, page_size: 10 },
    })
    savedCharts.value = data.charts
    showPicker.value = data.charts.length > 0
  } catch (err: any) {
    // Graceful fallback: show the new-chart form instead of blocking
    chartsError.value = ''
    savedCharts.value = []
    showPicker.value = false
  } finally {
    chartsLoading.value = false
  }
}

async function selectChart(chart: SavedChart) {
  loading.value = true
  error.value = ''
  try {
    const res = await client.post('/chart', {
      birth_year: chart.birth_year,
      birth_month: chart.birth_month,
      birth_day: chart.birth_day,
      birth_hour: chart.birth_hour,
      birth_min: 0,
      calendar_type: 'SOLAR',
      gender: chart.gender.toUpperCase(),
      name: chart.name || '',
    })
    chartData.value = res.data
    showPicker.value = false
    // Save as last birth for future use
    localStorage.setItem('bazi_last_birth', JSON.stringify({
      year: chart.birth_year,
      month: chart.birth_month,
      day: chart.birth_day,
      shichen: Math.floor(chart.birth_hour / 2),
      gender: chart.gender.toLowerCase(),
      chartId: res.data.id,
    }))
  } catch (err: any) {
    error.value = err.response?.data?.error || err.message || '排盘失败，请稍后重试'
  } finally {
    loading.value = false
  }
}

function startNewChart() {
  showPicker.value = false
}

function formatBirth(c: SavedChart): string {
  const m = String(c.birth_month).padStart(2, '0')
  const d = String(c.birth_day).padStart(2, '0')
  const h = String(c.birth_hour).padStart(2, '0')
  return `${c.birth_year}-${m}-${d} ${h}:00`
}

async function loadChart() {
  loading.value = true
  error.value = ''
  try {
    const res = await client.get(`/charts/${route.params.id}`)
    chartData.value = res.data.chart || res.data
  } catch (err: any) {
    error.value = err.response?.data?.error || err.message || '加载命盘失败'
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
    <!-- Constellation background -->
    <div class="bg-constellation" aria-hidden="true">
      <svg viewBox="0 0 1440 900" preserveAspectRatio="xMidYMid slice" class="constellation-svg">
        <defs>
          <radialGradient id="chart-nebula" cx="50%" cy="40%" r="60%">
            <stop offset="0%" stop-color="#D4A84B" stop-opacity="0.07" />
            <stop offset="100%" stop-color="#D4A84B" stop-opacity="0" />
          </radialGradient>
          <radialGradient id="chart-nebula2" cx="80%" cy="60%" r="40%">
            <stop offset="0%" stop-color="#C41E3A" stop-opacity="0.04" />
            <stop offset="100%" stop-color="#C41E3A" stop-opacity="0" />
          </radialGradient>
        </defs>
        <ellipse cx="720" cy="360" rx="700" ry="400" fill="url(#chart-nebula)" />
        <ellipse cx="1100" cy="500" rx="400" ry="350" fill="url(#chart-nebula2)" />
        <circle cx="150" cy="120" r="1.2" fill="#D4A84B" opacity="0.35" />
        <circle cx="600" cy="80" r="0.8" fill="#fff" opacity="0.25" />
        <circle cx="900" cy="150" r="1" fill="#D4A84B" opacity="0.4" />
        <circle cx="1200" cy="100" r="1.3" fill="#D4A84B" opacity="0.35" />
        <circle cx="250" cy="350" r="1.8" fill="#D4A84B" opacity="0.5" filter="url(#star-glow)" />
        <circle cx="720" cy="450" r="2.2" fill="#D4A84B" opacity="0.6" filter="url(#star-glow)" />
        <circle cx="1100" cy="380" r="1.5" fill="#D4A84B" opacity="0.45" />
        <circle cx="500" cy="700" r="1.6" fill="#D4A84B" opacity="0.5" />
        <circle cx="1000" cy="650" r="1.4" fill="#D4A84B" opacity="0.4" />
        <circle cx="80" cy="600" r="0.9" fill="#D4A84B" opacity="0.3" />
        <circle cx="1350" cy="700" r="1.1" fill="#D4A84B" opacity="0.35" />
        <line
          x1="250"
          y1="350"
          x2="720"
          y2="450"
          stroke="#D4A84B"
          stroke-width="0.5"
          opacity="0.12"
        />
        <line
          x1="720"
          y1="450"
          x2="1100"
          y2="380"
          stroke="#D4A84B"
          stroke-width="0.5"
          opacity="0.1"
        />
        <line
          x1="600"
          y1="80"
          x2="900"
          y2="150"
          stroke="#D4A84B"
          stroke-width="0.4"
          opacity="0.08"
        />
        <line
          x1="250"
          y1="350"
          x2="500"
          y2="700"
          stroke="#D4A84B"
          stroke-width="0.4"
          opacity="0.06"
        />
        <filter id="star-glow">
          <feGaussianBlur stdDeviation="2" result="blur" />
          <feMerge>
            <feMergeNode in="blur" />
            <feMergeNode in="SourceGraphic" />
          </feMerge>
        </filter>
      </svg>
    </div>

    <!-- Header -->
    <header class="chart-header">
      <div class="header-inner">
        <router-link to="/" class="back-link">
          <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
            <path
              d="M10 3L5 8l5 5"
              stroke="currentColor"
              stroke-width="1.5"
              stroke-linecap="round"
              stroke-linejoin="round"
            />
          </svg>
          返回首页
        </router-link>
        <div class="header-title-block">
          <div class="header-eyebrow">BaZi Fortune</div>
          <h1 class="header-title">八字命盘</h1>
        </div>
        <div class="header-spacer"></div>
      </div>
    </header>

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
                stroke="#D4A84B"
                stroke-width="0.5"
                stroke-dasharray="2 3"
                opacity="0.4"
              />
              <circle
                cx="40"
                cy="40"
                r="20"
                stroke="#D4A84B"
                stroke-width="0.5"
                stroke-dasharray="1 4"
                opacity="0.3"
              />
              <circle cx="40" cy="40" r="4" fill="#D4A84B" opacity="0.3" />
              <circle cx="20" cy="25" r="2" fill="#D4A84B" opacity="0.6" class="star-pulse" />
              <circle
                cx="60"
                cy="22"
                r="2.5"
                fill="#D4A84B"
                opacity="0.5"
                class="star-pulse"
                style="animation-delay: 0.3s"
              />
              <circle
                cx="62"
                cy="55"
                r="2"
                fill="#D4A84B"
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
              stroke="#C41E3A"
              stroke-width="1"
              stroke-dasharray="4 3"
              opacity="0.4"
            />
            <line
              x1="26"
              y1="26"
              x2="54"
              y2="54"
              stroke="#C41E3A"
              stroke-width="2.5"
              opacity="0.5"
            />
            <line
              x1="54"
              y1="26"
              x2="26"
              y2="54"
              stroke="#C41E3A"
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
            <div
              v-for="chart in savedCharts"
              :key="chart.id"
              class="picker-row glass-panel"
              @click="selectChart(chart)"
            >
              <div class="picker-avatar">
                {{ chart.name?.charAt(0) || '?' }}
              </div>
              <div class="picker-info">
                <span class="picker-name">{{ chart.name || '未命名' }}</span>
                <span class="picker-meta">
                  <span class="meta-tag">{{ chart.gender }}</span>
                  <span class="meta-sep">·</span>
                  {{ formatBirth(chart) }}
                </span>
              </div>
              <svg
                class="picker-arrow"
                width="14"
                height="14"
                viewBox="0 0 14 14"
                fill="none"
              >
                <path
                  d="M5 3l4 4-4 4"
                  stroke="currentColor"
                  stroke-width="1.5"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                />
              </svg>
            </div>
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
          <button class="btn-primary" @click="goFortune">
            <span class="btn-icon">✦</span>
            查看运势
          </button>
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
      <div v-else class="empty-state">
        <div class="empty-icon">
          <svg width="80" height="80" viewBox="0 0 80 80" fill="none">
            <circle
              cx="40"
              cy="40"
              r="35"
              stroke="#D4A84B"
              stroke-width="0.5"
              stroke-dasharray="2 3"
              opacity="0.3"
            />
            <circle cx="40" cy="40" r="4" fill="#D4A84B" opacity="0.3" />
            <circle cx="20" cy="25" r="2" fill="#D4A84B" opacity="0.4" />
            <circle cx="60" cy="22" r="2.5" fill="#D4A84B" opacity="0.3" />
            <circle cx="62" cy="55" r="2" fill="#D4A84B" opacity="0.35" />
          </svg>
        </div>
        <p class="empty-title">未找到命盘</p>
        <router-link to="/chart/new" class="btn-primary">
          <span class="btn-icon">✦</span>
          创建新的命盘
        </router-link>
      </div>
    </main>
  </div>
</template>

<style scoped>
.chart-page {
  min-height: 100vh;
  background: var(--bg);
  position: relative;
  overflow: hidden;
}

/* Background */
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

/* Header */
.chart-header {
  position: relative;
  z-index: 1;
  border-bottom: 1px solid rgba(212, 168, 75, 0.08);
  background: rgba(10, 8, 21, 0.6);
  backdrop-filter: blur(20px);
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
  font-size: 0.82rem;
  color: rgba(212, 168, 75, 0.5);
  text-decoration: none;
  transition: color 0.2s;
  letter-spacing: 1px;
}

.back-link:hover {
  color: var(--gold);
}

.header-title-block {
  text-align: center;
}

.header-eyebrow {
  font-size: 9px;
  letter-spacing: 3px;
  color: rgba(212, 168, 75, 0.3);
  text-transform: uppercase;
}

.header-title {
  font-family: var(--font-serif), serif;
  font-size: 1.1rem;
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
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 60vh;
  gap: 1.25rem;
}

.error-icon {
  opacity: 0.6;
}

.error-title {
  font-size: 0.9rem;
  color: var(--muted);
  text-align: center;
  margin: 0;
}

.btn-retry {
  padding: 0.5rem 1.75rem;
  background: linear-gradient(135deg, #c41e3a, #8b0000);
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 0.85rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  box-shadow: 0 4px 16px rgba(196, 30, 58, 0.25);
}

.btn-retry:hover {
  transform: translateY(-1px);
  box-shadow: 0 6px 20px rgba(196, 30, 58, 0.35);
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
  background: rgba(212, 168, 75, 0.06);
  border: 1px solid rgba(212, 168, 75, 0.15);
  border-radius: 20px;
  font-size: 0.75rem;
  color: rgba(212, 168, 75, 0.6);
  letter-spacing: 1px;
}

.badge-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--gold);
  box-shadow: 0 0 8px rgba(212, 168, 75, 0.5);
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
  background: radial-gradient(circle, rgba(212, 168, 75, 0.06), transparent 70%);
  pointer-events: none;
}

.btn-primary {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 0.875rem 2.5rem;
  background: linear-gradient(135deg, #d4a84b, #b8860b);
  color: #0a0815;
  font-weight: 700;
  font-size: 0.95rem;
  border: none;
  border-radius: 50px;
  cursor: pointer;
  transition: all 0.3s;
  box-shadow: 0 4px 24px rgba(212, 168, 75, 0.25);
  text-decoration: none;
  letter-spacing: 1px;
}

.btn-primary:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 36px rgba(212, 168, 75, 0.4);
}

.btn-secondary {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 0.875rem 2rem;
  background: transparent;
  color: var(--gold);
  font-weight: 600;
  font-size: 0.95rem;
  border: 1px solid rgba(212, 168, 75, 0.25);
  border-radius: 50px;
  cursor: pointer;
  transition: all 0.3s;
  letter-spacing: 1px;
}

.btn-secondary:hover {
  border-color: rgba(212, 168, 75, 0.5);
  background: rgba(212, 168, 75, 0.06);
  box-shadow: 0 0 20px rgba(212, 168, 75, 0.1);
}

.btn-icon {
  font-size: 0.75rem;
  animation: spin-slow 8s linear infinite;
}

.btn-icon-secondary {
  font-size: 1rem;
}

/* Rechart link */
.rechart-link {
  text-align: center;
}

.link-text {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 0.82rem;
  color: var(--muted);
  text-decoration: none;
  transition: color 0.2s;
}

.link-text:hover {
  color: var(--gold);
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
  opacity: 0.5;
  margin-bottom: 0.5rem;
}

.empty-title {
  font-size: 1rem;
  color: var(--muted);
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
  font-size: 0.9rem;
  color: var(--muted);
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
  background: rgba(212, 168, 75, 0.06);
  border: 1px solid rgba(212, 168, 75, 0.15);
  border-radius: 20px;
  font-size: 0.75rem;
  color: rgba(212, 168, 75, 0.6);
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
  cursor: pointer;
  transition: all 0.25s ease;
}

.picker-row:hover {
  border-color: rgba(212, 168, 75, 0.3);
  transform: translateY(-1px);
  box-shadow:
    0 4px 16px rgba(0, 0, 0, 0.2),
    0 0 16px rgba(212, 168, 75, 0.06);
}

.picker-avatar {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  background: linear-gradient(135deg, var(--gold), #b8860b);
  color: #0a0815;
  font-size: 1rem;
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
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--text);
}

.picker-meta {
  font-size: 0.75rem;
  color: var(--muted);
  display: flex;
  align-items: center;
  gap: 0.3rem;
}

.picker-arrow {
  color: rgba(212, 168, 75, 0.3);
  flex-shrink: 0;
  transition: color 0.2s, transform 0.2s;
}

.picker-row:hover .picker-arrow {
  color: var(--gold);
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
  background: linear-gradient(90deg, transparent, rgba(212, 168, 75, 0.12), transparent);
}

.divider-text {
  font-size: 0.7rem;
  color: rgba(255, 255, 255, 0.15);
  letter-spacing: 1px;
}

/* New chart button */
.btn-new-chart {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 0.75rem 2.5rem;
  background: transparent;
  color: var(--gold);
  font-weight: 600;
  font-size: 0.9rem;
  border: 1px solid rgba(212, 168, 75, 0.25);
  border-radius: 50px;
  cursor: pointer;
  transition: all 0.3s;
  letter-spacing: 1px;
}

.btn-new-chart:hover {
  border-color: rgba(212, 168, 75, 0.5);
  background: rgba(212, 168, 75, 0.06);
  box-shadow: 0 0 20px rgba(212, 168, 75, 0.12);
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
