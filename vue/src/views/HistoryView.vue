<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import client from '../api/client'

interface BirthChart {
  id: number
  name: string
  gender: string
  birth_year: number
  birth_month: number
  birth_day: number
  birth_hour: number
  birth_min: number
  calendar_type: string
  created_at: string
}

interface ChartListResponse {
  charts: BirthChart[]
  total: number
  page: number
  page_size: number
}

const router = useRouter()

const charts = ref<BirthChart[]>([])
const loading = ref(true)
const error = ref('')
const total = ref(0)
const page = ref(1)
const pageSize = 10

function formatBirth(c: BirthChart): string {
  const m = String(c.birth_month).padStart(2, '0')
  const d = String(c.birth_day).padStart(2, '0')
  const h = String(c.birth_hour).padStart(2, '0')
  const min = String(c.birth_min).padStart(2, '0')
  return `${c.birth_year}-${m}-${d} ${h}:${min}`
}

function formatDate(iso: string): string {
  if (!iso) return ''
  const d = new Date(iso)
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

function goChart(id: number) {
  router.push(`/chart/${id}`)
}

function goFortuneHistory(chartId: number) {
  router.push(`/fortune?chart_id=${chartId}`)
}

async function fetchCharts(p: number) {
  loading.value = true
  error.value = ''
  try {
    const { data } = await client.get<ChartListResponse>('/charts', {
      params: { page: p, page_size: pageSize },
    })
    charts.value = data.charts
    total.value = data.total
    page.value = data.page
  } catch (e: any) {
    error.value = e.response?.data?.error || '加载列表失败'
  } finally {
    loading.value = false
  }
}

function prevPage() {
  if (page.value > 1) fetchCharts(page.value - 1)
}

function nextPage() {
  if (page.value * pageSize < total.value) fetchCharts(page.value + 1)
}

onMounted(() => {
  fetchCharts(1)
})
</script>

<template>
  <div class="history-page">
    <!-- Constellation background -->
    <div class="bg-constellation" aria-hidden="true">
      <svg viewBox="0 0 800 600" preserveAspectRatio="xMidYMid slice" class="constellation-svg">
        <defs>
          <radialGradient id="hist-nebula" cx="50%" cy="50%" r="50%">
            <stop offset="0%" stop-color="#D4A84B" stop-opacity="0.05" />
            <stop offset="100%" stop-color="#D4A84B" stop-opacity="0" />
          </radialGradient>
        </defs>
        <ellipse cx="400" cy="300" rx="350" ry="250" fill="url(#hist-nebula)" />
        <circle cx="120" cy="100" r="1" fill="#D4A84B" opacity="0.25" />
        <circle cx="680" cy="80" r="1.2" fill="#D4A84B" opacity="0.35" />
        <circle cx="720" cy="480" r="1" fill="#D4A84B" opacity="0.25" />
        <circle cx="60" cy="520" r="0.8" fill="#D4A84B" opacity="0.2" />
        <circle cx="400" cy="300" r="1.5" fill="#D4A84B" opacity="0.4" />
        <circle cx="250" cy="180" r="1" fill="#D4A84B" opacity="0.3" />
        <circle cx="550" cy="400" r="1.3" fill="#D4A84B" opacity="0.35" />
        <line
          x1="120"
          y1="100"
          x2="250"
          y2="180"
          stroke="#D4A84B"
          stroke-width="0.4"
          opacity="0.06"
        />
        <line
          x1="680"
          y1="80"
          x2="550"
          y2="400"
          stroke="#D4A84B"
          stroke-width="0.4"
          opacity="0.05"
        />
      </svg>
    </div>

    <div class="page-inner">
      <div class="history-header">
        <div class="header-eyebrow">BaZi Fortune</div>
        <h1 class="page-title">命盘历史</h1>
        <p v-if="total" class="page-subtitle">共 {{ total }} 条记录</p>
      </div>

      <!-- Loading -->
      <div v-if="loading" class="p-8">
        <div class="skeleton h-6 w-32 mb-4"></div>
        <div class="skeleton h-20 rounded-xl mb-3" v-for="i in 3" :key="i"></div>
      </div>

      <!-- Error -->
      <div v-else-if="error" class="state-box">
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
        <button class="btn-retry" @click="fetchCharts(1)">重新加载</button>
      </div>

      <!-- Chart List -->
      <div v-else class="chart-list">
        <div v-for="chart in charts" :key="chart.id" class="chart-card" @click="goChart(chart.id)">
          <div class="card-glow"></div>
          <div class="card-main">
            <div class="card-avatar">{{ chart.name?.charAt(0) || '?' }}</div>
            <div class="card-info">
              <h3 class="card-name">{{ chart.name || '未命名' }}</h3>
              <p class="card-meta">
                <span class="meta-tag">{{ chart.gender }}</span>
                <span class="meta-sep">·</span>
                <span>{{ formatBirth(chart) }}</span>
              </p>
              <p class="card-date">创建于 {{ formatDate(chart.created_at) }}</p>
            </div>
          </div>
          <div class="card-actions">
            <button class="action-btn fortune-btn" @click.stop="goFortuneHistory(chart.id)">
              <span class="btn-icon">✦</span>
              运势历史
            </button>
          </div>
        </div>
      </div>

      <!-- Pagination -->
      <div v-if="total > pageSize" class="pagination">
        <button class="page-btn" :disabled="page <= 1" @click="prevPage">← 上一页</button>
        <span class="page-info">{{ page }} / {{ Math.ceil(total / pageSize) }}</span>
        <button class="page-btn" :disabled="page * pageSize >= total" @click="nextPage">
          下一页 →
        </button>
      </div>

      <!-- Empty -->
      <div v-if="!loading && !error && charts.length === 0" class="empty-state">
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
            <circle
              cx="40"
              cy="40"
              r="20"
              stroke="#D4A84B"
              stroke-width="0.5"
              stroke-dasharray="1 4"
              opacity="0.2"
            />
            <circle cx="40" cy="40" r="4" fill="#D4A84B" opacity="0.3" />
            <circle cx="20" cy="25" r="2" fill="#D4A84B" opacity="0.4" />
            <circle cx="60" cy="22" r="2.5" fill="#D4A84B" opacity="0.3" />
            <circle cx="62" cy="55" r="2" fill="#D4A84B" opacity="0.35" />
            <circle cx="22" cy="58" r="2.5" fill="#D4A84B" opacity="0.3" />
          </svg>
        </div>
        <p class="empty-title">暂无历史记录</p>
        <p class="empty-sub">开始探索命运的轨迹</p>
        <router-link to="/chart/new" class="btn-primary">
          <span class="btn-icon">✦</span>
          开始排盘
        </router-link>
      </div>
    </div>
  </div>
</template>

<style scoped>
.history-page {
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
  max-width: 560px;
  margin: 0 auto;
  padding: 2rem 1rem;
}

.history-header {
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

.page-subtitle {
  font-size: 12px;
  color: var(--muted);
  margin: 0;
}

/* Chart Cards */
.chart-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.chart-card {
  position: relative;
  background: linear-gradient(160deg, rgba(25, 20, 40, 0.95), rgba(12, 10, 22, 0.98));
  border: 1px solid rgba(212, 168, 75, 0.12);
  border-radius: 16px;
  padding: 1rem 1.125rem;
  cursor: pointer;
  transition: all 0.3s ease;
  overflow: hidden;
}

.card-glow {
  position: absolute;
  inset: 0;
  background: radial-gradient(circle at 50% 0%, rgba(212, 168, 75, 0.04), transparent 70%);
  opacity: 0;
  transition: opacity 0.3s;
}

.chart-card:hover {
  border-color: rgba(212, 168, 75, 0.3);
  transform: translateY(-2px);
  box-shadow:
    0 8px 32px rgba(0, 0, 0, 0.3),
    0 0 20px rgba(212, 168, 75, 0.06);
}

.chart-card:hover .card-glow {
  opacity: 1;
}

.card-main {
  display: flex;
  align-items: center;
  gap: 0.875rem;
  margin-bottom: 0.625rem;
}

.card-avatar {
  width: 42px;
  height: 42px;
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

.card-info {
  flex: 1;
}

.card-name {
  font-size: 1rem;
  font-weight: 700;
  color: var(--text);
  margin: 0 0 0.2rem;
}

.card-meta {
  font-size: 0.78rem;
  color: var(--muted);
  margin: 0;
  display: flex;
  align-items: center;
  gap: 0.3rem;
}

.meta-tag {
  display: inline-block;
  padding: 0.1rem 0.5rem;
  background: rgba(212, 168, 75, 0.08);
  border: 1px solid rgba(212, 168, 75, 0.1);
  border-radius: 4px;
  font-size: 0.7rem;
  color: rgba(212, 168, 75, 0.5);
}

.meta-sep {
  color: rgba(255, 255, 255, 0.08);
}

.card-date {
  font-size: 0.7rem;
  color: rgba(255, 255, 255, 0.2);
  margin: 0.2rem 0 0;
}

.card-actions {
  display: flex;
  justify-content: flex-end;
  padding-top: 0.5rem;
  border-top: 1px solid rgba(212, 168, 75, 0.05);
}

.action-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  font-size: 0.72rem;
  padding: 0.35rem 0.875rem;
  border-radius: 20px;
  border: 1px solid rgba(212, 168, 75, 0.15);
  background: rgba(212, 168, 75, 0.04);
  color: var(--gold);
  cursor: pointer;
  font-weight: 500;
  transition: all 0.2s ease;
}

.action-btn:hover {
  background: rgba(212, 168, 75, 0.1);
  border-color: rgba(212, 168, 75, 0.3);
  box-shadow: 0 0 12px rgba(212, 168, 75, 0.1);
}

.btn-icon {
  font-size: 0.65rem;
  animation: spin-slow 8s linear infinite;
}

@keyframes spin-slow {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

/* States */
.state-box {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 50vh;
  gap: 1rem;
}

.error-icon {
  opacity: 0.6;
}

.error-text {
  font-size: 0.9rem;
  color: var(--muted);
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

/* Pagination */
.pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 1.25rem;
  margin-top: 1.5rem;
}

.page-btn {
  padding: 0.4rem 1rem;
  border: 1px solid rgba(212, 168, 75, 0.1);
  background: rgba(255, 255, 255, 0.02);
  border-radius: 8px;
  font-size: 0.8rem;
  color: var(--muted);
  cursor: pointer;
  transition: all 0.2s;
}

.page-btn:not(:disabled):hover {
  border-color: rgba(212, 168, 75, 0.3);
  color: var(--gold);
  background: rgba(212, 168, 75, 0.04);
}

.page-btn:disabled {
  opacity: 0.35;
  cursor: not-allowed;
}

.page-info {
  font-size: 0.8rem;
  color: var(--muted);
}

/* Empty state */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 4rem 2rem;
  text-align: center;
}

.empty-icon {
  margin-bottom: 1.5rem;
  opacity: 0.5;
}

.empty-title {
  font-size: 1.1rem;
  font-weight: 700;
  color: var(--text);
  margin: 0 0 0.5rem;
}

.empty-sub {
  font-size: 0.85rem;
  color: var(--muted);
  margin: 0 0 2rem;
}

.btn-primary {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem 2rem;
  background: linear-gradient(135deg, #d4a84b, #b8860b);
  color: #0a0815;
  font-weight: 700;
  font-size: 0.9rem;
  border: none;
  border-radius: 50px;
  cursor: pointer;
  transition: all 0.3s;
  box-shadow: 0 4px 20px rgba(212, 168, 75, 0.25);
  text-decoration: none;
  letter-spacing: 1px;
}

.btn-primary:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 32px rgba(212, 168, 75, 0.4);
}
</style>
