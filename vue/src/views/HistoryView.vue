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
    <div class="history-header">
      <h1 class="page-title">八字排盘历史</h1>
      <p v-if="total" class="page-subtitle">共 {{ total }} 条记录</p>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="p-8">
      <div class="skeleton h-6 w-32 mb-4"></div>
      <div class="skeleton h-20 rounded-xl mb-3" v-for="i in 3" :key="i"></div>
    </div>

    <!-- Error -->
    <div v-else-if="error" class="state-box">
      <el-result icon="error" title="加载失败" sub-title="请检查网络连接后重试">
        <template #extra>
          <el-button type="primary" @click="fetchCharts(1)">重试</el-button>
        </template>
      </el-result>
    </div>

    <!-- Chart List -->
    <div v-else class="chart-list">
      <div
        v-for="chart in charts"
        :key="chart.id"
        class="chart-card"
        @click="goChart(chart.id)"
      >
        <div class="card-main">
          <h3 class="card-name">{{ chart.name || '未命名' }}</h3>
          <p class="card-meta">
            <span>{{ chart.gender }}</span>
            <span class="meta-sep">·</span>
            <span>{{ formatBirth(chart) }}</span>
          </p>
          <p class="card-date">创建于 {{ formatDate(chart.created_at) }}</p>
        </div>
        <div class="card-actions">
          <button
            class="action-btn fortune-btn"
            @click.stop="goFortuneHistory(chart.id)"
          >
            查看运势历史
          </button>
        </div>
      </div>
    </div>

    <!-- Pagination -->
    <div v-if="total > pageSize" class="pagination">
      <button
        class="page-btn"
        :disabled="page <= 1"
        @click="prevPage"
      >
        上一页
      </button>
      <span class="page-info">{{ page }} / {{ Math.ceil(total / pageSize) }}</span>
      <button
        class="page-btn"
        :disabled="page * pageSize >= total"
        @click="nextPage"
      >
        下一页
      </button>
    </div>

    <!-- Empty -->
    <div v-if="!loading && !error && charts.length === 0" class="text-center py-20 text-muted">
      <div class="text-4xl mb-4">📋</div>
      <p class="mb-4">暂无历史记录</p>
      <router-link to="/chart/new">
        <el-button type="primary">去排盘</el-button>
      </router-link>
    </div>
  </div>
</template>

<style scoped>
.history-page {
  min-height: 100vh;
  background: #FAF8F3;
  padding: 1.25rem 1rem;
  max-width: 540px;
  margin: 0 auto;
}

.history-header {
  margin-bottom: 1.25rem;
}

.page-title {
  font-size: 1.35rem;
  font-weight: 800;
  color: var(--color-bazi-ink);
  margin: 0;
}

.page-subtitle {
  font-size: 0.8rem;
  color: #999;
  margin: 0.25rem 0 0 0;
}

/* Chart Cards */
.chart-list {
  display: flex;
  flex-direction: column;
  gap: 0.625rem;
}

.chart-card {
  background: white;
  border: 1px solid #E8E3D5;
  border-radius: 0.75rem;
  padding: 0.875rem 1rem;
  cursor: pointer;
  transition: border-color 0.15s, box-shadow 0.15s;
}

.chart-card:hover {
  border-color: var(--color-bazi-red);
  box-shadow: 0 2px 12px rgba(196, 30, 58, 0.08);
}

.card-main {
  margin-bottom: 0.5rem;
}

.card-name {
  font-size: 1.05rem;
  font-weight: 700;
  color: var(--color-bazi-ink);
  margin: 0 0 0.25rem 0;
}

.card-meta {
  font-size: 0.82rem;
  color: #666;
  margin: 0;
}

.meta-sep {
  margin: 0 0.35rem;
  color: #ccc;
}

.card-date {
  font-size: 0.72rem;
  color: #bbb;
  margin: 0.15rem 0 0 0;
}

.card-actions {
  display: flex;
  justify-content: flex-end;
}

.action-btn {
  font-size: 0.75rem;
  padding: 0.3rem 0.75rem;
  border-radius: 1rem;
  border: 1px solid var(--color-bazi-red);
  background: white;
  color: var(--color-bazi-red);
  cursor: pointer;
  font-weight: 500;
  transition: background 0.15s, color 0.15s;
}

.action-btn:hover {
  background: var(--color-bazi-red);
  color: white;
}

/* States */
.state-box {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 50vh;
  gap: 0.75rem;
}

.state-text {
  font-size: 0.95rem;
  color: var(--color-bazi-ink);
  margin: 0;
}

.state-text.empty {
  color: #bbb;
}

.error-box .state-text {
  color: var(--color-bazi-red);
}

.back-link {
  color: var(--color-bazi-red);
  text-decoration: none;
  font-size: 0.85rem;
}

/* Pagination */
.pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 1rem;
  margin-top: 1.25rem;
}

.page-btn {
  padding: 0.4rem 1rem;
  border: 1px solid #E8E3D5;
  background: white;
  border-radius: 0.5rem;
  font-size: 0.8rem;
  color: var(--color-bazi-ink);
  cursor: pointer;
  transition: border-color 0.15s;
}

.page-btn:disabled {
  color: #ccc;
  cursor: not-allowed;
}

.page-btn:not(:disabled):hover {
  border-color: var(--color-bazi-red);
}

.page-info {
  font-size: 0.8rem;
  color: #999;
}
</style>
