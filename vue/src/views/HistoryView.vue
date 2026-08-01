<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import {
  fetchCharts as fetchChartList,
  fetchFortuneHistory,
  type ChartSummary,
  type FortuneHistoryItem,
} from '../api/chart'
import { getApiErrorMessage } from '../api/client'

const router = useRouter()

const charts = ref<ChartSummary[]>([])
const loading = ref(true)
const error = ref('')
const total = ref(0)
const page = ref(1)
const pageSize = 10
const expandedHistoryChartId = ref<number | null>(null)
const fortuneHistory = ref<Record<number, FortuneHistoryItem[]>>({})
const historyLoadingChartId = ref<number | null>(null)
const historyError = ref('')

function formatBirth(c: ChartSummary): string {
  const m = String(c.birth_month).padStart(2, '0')
  const d = String(c.birth_day).padStart(2, '0')
  const h = String(c.birth_hour).padStart(2, '0')
  const min = String(c.birth_min).padStart(2, '0')
  const calendar = c.calendar_type === 'LUNAR' ? (c.lunar_leap_month ? '农历闰月' : '农历') : '公历'
  return `${calendar} ${c.birth_year}-${m}-${d} ${h}:${min}`
}

function genderLabel(value: string): string {
  return value === 'FEMALE' || value === '女' ? '女' : '男'
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

async function toggleFortuneHistory(chartId: number) {
  if (expandedHistoryChartId.value === chartId) {
    expandedHistoryChartId.value = null
    return
  }
  expandedHistoryChartId.value = chartId
  historyError.value = ''
  if (fortuneHistory.value[chartId]) return
  historyLoadingChartId.value = chartId
  try {
    const data = await fetchFortuneHistory(chartId, 1, 10)
    fortuneHistory.value = { ...fortuneHistory.value, [chartId]: data.items }
  } catch (reason: unknown) {
    historyError.value = getApiErrorMessage(reason, '运势记录加载失败，请稍后重试。')
  } finally {
    historyLoadingChartId.value = null
  }
}

async function loadCharts(p: number) {
  loading.value = true
  error.value = ''
  try {
    const data = await fetchChartList(p, pageSize)
    charts.value = data.charts
    total.value = data.total
    page.value = data.page
  } catch (reason: unknown) {
    error.value = getApiErrorMessage(reason, '命盘列表加载失败，请稍后重试。')
  } finally {
    loading.value = false
  }
}

function prevPage() {
  if (page.value > 1) loadCharts(page.value - 1)
}

function nextPage() {
  if (page.value * pageSize < total.value) loadCharts(page.value + 1)
}

onMounted(() => {
  loadCharts(1)
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
            <line
              x1="20"
              y1="20"
              x2="40"
              y2="40"
              stroke="currentColor"
              stroke-width="2"
              opacity="0.5"
            />
            <line
              x1="40"
              y1="20"
              x2="20"
              y2="40"
              stroke="currentColor"
              stroke-width="2"
              opacity="0.5"
            />
          </svg>
        </div>
        <p class="error-text">{{ error }}</p>
        <button class="btn-retry" @click="loadCharts(1)">重新加载</button>
      </div>

      <!-- Chart List -->
      <div v-else class="chart-list">
        <article v-for="chart in charts" :key="chart.id" class="chart-card">
          <button type="button" class="card-main-button" @click="goChart(chart.id)">
            <div class="card-main">
              <div class="card-avatar" aria-hidden="true">{{ chart.name?.charAt(0) || '?' }}</div>
              <div class="card-info">
                <div class="card-title-row">
                  <h3 class="card-name">{{ chart.name || '未命名' }}</h3>
                  <span class="card-date">{{ formatDate(chart.created_at || '') }} 创建</span>
                </div>
                <p class="card-meta">
                  <span class="meta-tag">{{ genderLabel(chart.gender) }}</span>
                  <span class="meta-sep">·</span>
                  <span>{{ formatBirth(chart) }}</span>
                </p>
              </div>
              <span class="card-chevron" aria-hidden="true">›</span>
            </div>
          </button>
          <div class="card-actions">
            <button
              type="button"
              class="action-btn fortune-btn"
              :aria-expanded="expandedHistoryChartId === chart.id"
              @click="toggleFortuneHistory(chart.id)"
            >
              <span class="btn-icon" aria-hidden="true">✦</span>
              运势记录
            </button>
          </div>
          <div v-if="expandedHistoryChartId === chart.id" class="fortune-history-panel">
            <p v-if="historyLoadingChartId === chart.id" class="history-status">
              正在加载运势记录…
            </p>
            <div v-else-if="historyError" class="history-status history-error">
              <span>{{ historyError }}</span>
              <button type="button" @click="toggleFortuneHistory(chart.id)">关闭</button>
            </div>
            <p v-else-if="!fortuneHistory[chart.id]?.length" class="history-status">
              该命盘还没有已保存的运势查询记录。
            </p>
            <ol v-else class="fortune-history-list">
              <li v-for="item in fortuneHistory[chart.id]" :key="item.id">
                <div>
                  <time>{{ item.query_date }}</time>
                  <strong>{{ item.day_gan_zhi }}</strong>
                </div>
                <p>{{ item.summary || '已保存当日干支关系记录' }}</p>
              </li>
            </ol>
          </div>
        </article>
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
  max-width: 640px;
  margin: 0 auto;
  padding: 2.5rem 1rem;
}

.history-header {
  margin-bottom: 1.75rem;
  padding-bottom: 1.25rem;
  border-bottom: 1px solid var(--line-subtle);
}

.header-eyebrow {
  font-size: var(--fs-2xs);
  letter-spacing: var(--tracking-eyebrow);
  color: var(--text-soft);
  text-transform: uppercase;
  margin-bottom: 8px;
}

.page-title {
  font-family: var(--font-serif), serif;
  font-size: var(--fs-4xl);
  font-weight: 700;
  color: var(--text);
  margin: 0 0 6px;
  letter-spacing: 0.12em;
}

.page-subtitle {
  font-size: var(--fs-2xs);
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
  background: var(--bg-elevated);
  border: 1px solid var(--line-subtle);
  border-radius: var(--radius-lg);
  padding: 1rem 1.125rem;
  transition:
    border-color 0.2s ease,
    box-shadow 0.2s ease,
    transform 0.2s ease;
}

.card-main-button {
  width: 100%;
  padding: 0;
  border: 0;
  background: transparent;
  color: inherit;
  text-align: left;
  cursor: pointer;
}

.card-main-button:focus-visible {
  outline: 2px solid var(--line-focus);
  outline-offset: 4px;
  border-radius: var(--radius-sm);
}

.chart-card:hover {
  border-color: var(--line-strong);
  transform: translateY(-1px);
  box-shadow: var(--shadow-md);
}

.chart-card:hover .card-chevron {
  color: var(--text);
  transform: translateX(2px);
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
  border-radius: var(--radius-md);
  background: var(--accent-dim);
  border: 1px solid var(--line-subtle);
  color: var(--text);
  font-family: var(--font-serif), serif;
  font-size: var(--fs-lg);
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.card-info {
  flex: 1;
  min-width: 0;
}

.card-title-row {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 0.75rem;
  margin-bottom: 0.2rem;
}

.card-name {
  font-size: var(--fs-md);
  font-weight: 600;
  color: var(--text);
  margin: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.card-date {
  font-size: var(--fs-2xs);
  color: var(--text-dim);
  flex-shrink: 0;
}

.card-meta {
  font-size: var(--fs-xs);
  color: var(--text-muted);
  margin: 0;
  display: flex;
  align-items: center;
  gap: 0.3rem;
}

.meta-tag {
  display: inline-block;
  padding: 0.05rem 0.45rem;
  background: var(--accent-dim);
  border-radius: var(--radius-sm);
  font-size: var(--fs-2xs);
  color: var(--text-muted);
}

.meta-sep {
  color: var(--text-soft);
}

.card-chevron {
  color: var(--text-soft);
  font-size: var(--fs-xl);
  line-height: 1;
  flex-shrink: 0;
  transition:
    color 0.2s ease,
    transform 0.2s ease;
}

.card-actions {
  display: flex;
  justify-content: flex-end;
  padding-top: 0.625rem;
  border-top: 1px solid var(--line-subtle);
}

.fortune-history-panel {
  margin-top: 0.75rem;
  padding-top: 0.75rem;
  border-top: 1px solid var(--line-subtle);
}

.history-status {
  margin: 0;
  color: var(--text-muted);
  font-size: var(--fs-xs);
  line-height: 1.6;
}

.history-error {
  display: flex;
  justify-content: space-between;
  gap: 0.75rem;
  color: var(--danger);
}

.history-error button {
  color: var(--text-muted);
  text-decoration: underline;
  text-underline-offset: 3px;
  cursor: pointer;
}

.fortune-history-list {
  display: grid;
  gap: 0.55rem;
  margin: 0;
  padding: 0;
  list-style: none;
}

.fortune-history-list li {
  padding: 0.6rem 0.75rem;
  border: 1px solid var(--line-subtle);
  border-radius: var(--radius-md);
  background: var(--surface-1);
}

.fortune-history-list li > div {
  display: flex;
  justify-content: space-between;
  gap: 0.75rem;
  color: var(--text);
  font-size: var(--fs-xs);
}

.fortune-history-list time {
  font-family: var(--font-mono), monospace;
  font-size: var(--fs-2xs);
  color: var(--text-dim);
}

.fortune-history-list p {
  margin: 0.3rem 0 0;
  color: var(--text-soft);
  font-size: var(--fs-xs);
  line-height: 1.5;
}

.action-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  font-size: var(--fs-xs);
  padding: 0.3rem 0.875rem;
  border-radius: 999px;
  border: 1px solid var(--line-strong);
  background: transparent;
  color: var(--text-muted);
  cursor: pointer;
  font-weight: 500;
  transition:
    color 0.2s ease,
    border-color 0.2s ease,
    background-color 0.2s ease;
}

.action-btn:hover {
  color: var(--text);
  border-color: var(--line-focus);
  background: var(--surface-1);
}

.btn-icon {
  font-size: var(--fs-2xs);
  color: var(--accent);
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
  font-size: var(--fs-sm);
  color: var(--text-muted);
}

.btn-retry {
  padding: 0.5rem 1.5rem;
  background: transparent;
  color: var(--danger);
  border: 1px solid color-mix(in oklab, var(--danger) 35%, transparent);
  border-radius: var(--radius-md);
  font-size: var(--fs-sm);
  font-weight: 500;
  cursor: pointer;
  transition:
    background-color 0.2s,
    border-color 0.2s;
}

.btn-retry:hover {
  background: color-mix(in oklab, var(--danger) 8%, transparent);
  border-color: color-mix(in oklab, var(--danger) 55%, transparent);
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
  border: 1px solid var(--line-strong);
  background: var(--bg-elevated);
  border-radius: var(--radius-md);
  font-size: var(--fs-sm);
  color: var(--text-muted);
  cursor: pointer;
  transition:
    color 0.2s,
    border-color 0.2s,
    background-color 0.2s;
}

.page-btn:not(:disabled):hover {
  border-color: var(--line-focus);
  color: var(--text);
  background: var(--surface-1);
}

.page-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.page-info {
  font-size: var(--fs-sm);
  color: var(--text-muted);
  font-variant-numeric: tabular-nums;
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
  font-size: var(--fs-lg);
  font-weight: 700;
  color: var(--text);
  margin: 0 0 0.5rem;
}

.empty-sub {
  font-size: var(--fs-sm);
  color: var(--text-muted);
  margin: 0 0 2rem;
}

.btn-primary {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.7rem 1.75rem;
  background: var(--text);
  color: var(--bg);
  font-weight: 600;
  font-size: var(--fs-sm);
  border: none;
  border-radius: 999px;
  cursor: pointer;
  transition:
    transform 0.2s,
    box-shadow 0.2s;
  box-shadow: var(--shadow-sm);
  text-decoration: none;
  letter-spacing: 0.1em;
}

.btn-primary:hover {
  transform: translateY(-1px);
  box-shadow: var(--shadow-md);
}
</style>
