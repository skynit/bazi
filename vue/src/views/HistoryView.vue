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
              stroke="currentColor"
              stroke-width="0.5"
              stroke-dasharray="2 3"
              opacity="0.3"
            />
            <circle
              cx="40"
              cy="40"
              r="20"
              stroke="currentColor"
              stroke-width="0.5"
              stroke-dasharray="1 4"
              opacity="0.2"
            />
            <circle cx="40" cy="40" r="4" fill="currentColor" opacity="0.3" />
            <circle cx="20" cy="25" r="2" fill="currentColor" opacity="0.4" />
            <circle cx="60" cy="22" r="2.5" fill="currentColor" opacity="0.3" />
            <circle cx="62" cy="55" r="2" fill="currentColor" opacity="0.35" />
            <circle cx="22" cy="58" r="2.5" fill="currentColor" opacity="0.3" />
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
  background: transparent;
  position: relative;
  overflow: hidden;
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

.page-subtitle {
  font-size: 12px;
  color: var(--text-muted);
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
  background: color-mix(in oklab, var(--surface-1) 84%, transparent);
  border: 1px solid var(--line-strong);
  border-radius: 16px;
  padding: 1rem 1.125rem;
  cursor: pointer;
  transition: all 0.3s ease;
  overflow: hidden;
}

.card-glow {
  position: absolute;
  inset: 0;
  background: radial-gradient(circle at 50% 0%, var(--accent-dim), transparent 70%);
  opacity: 0;
  transition: opacity 0.3s;
}

.chart-card:hover {
  border-color: var(--line-focus);
  transform: translateY(-2px);
  box-shadow:
    var(--shadow-md),
    0 0 20px var(--accent-dim);
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
  background: linear-gradient(135deg, var(--accent), #94a3b8);
  color: #030404;
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
  color: var(--text-muted);
  margin: 0;
  display: flex;
  align-items: center;
  gap: 0.3rem;
}

.meta-tag {
  display: inline-block;
  padding: 0.1rem 0.5rem;
  background: var(--accent-dim);
  border: 1px solid var(--line-subtle);
  border-radius: 4px;
  font-size: 0.7rem;
  color: var(--text-muted);
}

.meta-sep {
  color: var(--text-soft);
}

.card-date {
  font-size: 0.7rem;
  color: var(--text-dim);
  margin: 0.2rem 0 0;
}

.card-actions {
  display: flex;
  justify-content: flex-end;
  padding-top: 0.5rem;
  border-top: 1px solid rgba(203, 213, 225, 0.05);
}

.action-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  font-size: 0.72rem;
  padding: 0.35rem 0.875rem;
  border-radius: 20px;
  border: 1px solid rgba(203, 213, 225, 0.15);
  background: rgba(203, 213, 225, 0.04);
  color: var(--accent);
  cursor: pointer;
  font-weight: 500;
  transition: all 0.2s ease;
}

.action-btn:hover {
  background: rgba(203, 213, 225, 0.1);
  border-color: var(--text-soft);
  box-shadow: 0 0 12px rgba(203, 213, 225, 0.1);
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
  color: var(--danger);
  opacity: 0.6;
}

.error-text {
  font-size: 0.9rem;
  color: var(--text-muted);
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
  border: 1px solid var(--line-subtle);
  background: var(--glass-bg);
  border-radius: 8px;
  font-size: 0.8rem;
  color: var(--text-muted);
  cursor: pointer;
  transition: all 0.2s;
}

.page-btn:not(:disabled):hover {
  border-color: var(--line-focus);
  color: var(--accent);
  background: var(--glass-bg-hover);
}

.page-btn:disabled {
  opacity: 0.35;
  cursor: not-allowed;
}

.page-info {
  font-size: 0.8rem;
  color: var(--text-muted);
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
  color: var(--icon-muted);
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
  color: var(--text-muted);
  margin: 0 0 2rem;
}

.btn-primary {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem 2rem;
  background: linear-gradient(135deg, #cbd5e1, #94a3b8);
  color: #030404;
  font-weight: 700;
  font-size: 0.9rem;
  border: none;
  border-radius: 50px;
  cursor: pointer;
  transition: all 0.3s;
  box-shadow: 0 4px 20px rgba(203, 213, 225, 0.25);
  text-decoration: none;
  letter-spacing: 1px;
}

.btn-primary:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 32px rgba(203, 213, 225, 0.4);
}
</style>
