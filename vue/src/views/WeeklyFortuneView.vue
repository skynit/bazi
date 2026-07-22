<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { fetchWeekly, parseTrend, type WeeklyFortuneResponse } from '../api/fortune'
import { getApiErrorMessage } from '../api/client'
import FortuneChart from '../components/FortuneChart.vue'
import AuroraMeshBackground from '../components/fortune/AuroraMeshBackground.vue'
import ScoreOrb from '../components/fortune/ScoreOrb.vue'
import FortuneHeatStrip from '../components/fortune/FortuneHeatStrip.vue'
import FortuneRadar from '../components/fortune/FortuneRadar.vue'
import DayFortuneCard from '../components/fortune/DayFortuneCard.vue'
import BestWorstChip from '../components/fortune/BestWorstChip.vue'

const route = useRoute()
const data = ref<WeeklyFortuneResponse | null>(null)
const loading = ref(true)
const error = ref('')
const chartId = ref<number | null>(null)

const WEEKDAYS = ['一', '二', '三', '四', '五', '六', '日'] as const

const trendData = computed(() => parseTrend(data.value?.element_trend ?? ''))

const weekRange = computed(() => {
  const days = data.value?.daily_fortunes ?? []
  if (!days.length) return ''
  return `${days[0].solar_date} – ${days[days.length - 1].solar_date}`
})

const heatDays = computed(() => {
  const summary = data.value?.summary
  return (data.value?.daily_fortunes ?? []).map((d) => ({
    date: d.solar_date,
    score: d.score,
    dayPillar: d.day_gan_zhi,
    isHighest: summary?.highest_index_day === d.solar_date,
    isLowest: summary?.lowest_index_day === d.solar_date,
  }))
})

const weekdayLabels = computed(() => {
  return (data.value?.daily_fortunes ?? []).map((d) => {
    const dt = new Date(d.solar_date + 'T00:00:00')
    return dt.toLocaleDateString('zh-CN', { weekday: 'short' }).replace('星期', '')
  })
})

const distribution = computed(() => data.value?.summary?.element_distribution ?? {})

function todayStr(): string {
  const d = new Date()
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

async function load() {
  let cid: number | null = null
  const q = route.query.chart_id
  if (typeof q === 'string' && q) cid = Number(q)
  if (!cid) {
    try {
      const s = localStorage.getItem('bazi_last_birth')
      if (s) cid = Number(JSON.parse(s).chartId) || null
    } catch {
      /* ignore */
    }
  }
  if (!cid) {
    error.value = '请先创建命盘'
    loading.value = false
    return
  }
  chartId.value = cid

  try {
    data.value = await fetchWeekly(cid, todayStr())
  } catch (reason: unknown) {
    error.value = getApiErrorMessage(reason, '本周内容加载失败，请稍后重试。')
  } finally {
    loading.value = false
  }
}

function weekdayFor(date: string): string {
  const dt = new Date(date + 'T00:00:00')
  return WEEKDAYS[(dt.getDay() + 6) % 7] // 周一为首
}

onMounted(load)
</script>

<template>
  <div class="weekly-page">
    <AuroraMeshBackground />

    <div v-if="loading" class="state">
      <div class="orb-skeleton" aria-hidden="true"></div>
      <p>本周运势加载中…</p>
    </div>

    <div v-else-if="error" class="state error">
      <p>{{ error }}</p>
      <router-link to="/chart/new" class="btn-link">去排盘 →</router-link>
    </div>

    <template v-else-if="data">
      <main class="page">
        <!-- Hero -->
        <section class="hero">
          <div class="hero-left">
            <span class="eyebrow">BaZi · Weekly</span>
            <h1 class="title">本周关系节奏</h1>
            <p class="range tabular-nums">{{ weekRange }}</p>
            <div class="chips">
              <BestWorstChip
                v-if="data.summary.highest_index_day"
                variant="highest"
                :date="data.summary.highest_index_day"
                :score="data.summary.highest_index"
              />
              <BestWorstChip
                v-if="data.summary.lowest_index_day"
                variant="lowest"
                :date="data.summary.lowest_index_day"
                :score="data.summary.lowest_index"
              />
              <span v-if="data.summary.dominant_element" class="meta-chip">
                <span class="dot" :style="{ background: 'var(--jade-accent)' }"></span>
                {{ data.summary.dominant_element }}主气
              </span>
              <span v-if="data.summary.dominant_ten_god" class="meta-chip">
                {{ data.summary.dominant_ten_god }}
              </span>
            </div>
          </div>
          <div class="hero-right">
            <ScoreOrb
              :score="data.structural_relation_index"
              label="周均关系活跃度"
              caption="用于日期间比较"
            />
          </div>
        </section>

        <!-- Heat strip -->
        <section class="glass-card heat-card">
          <header class="card-head">
            <span class="card-eyebrow">每日关系活跃度</span>
            <span class="card-meta">
              较多 {{ data.summary.highest_index }} · 较少 {{ data.summary.lowest_index }}
            </span>
          </header>
          <FortuneHeatStrip :days="heatDays" :weekday-labels="weekdayLabels" />
        </section>

        <!-- Radar + trend grid -->
        <section class="grid-2">
          <div class="glass-card">
            <header class="card-head">
              <span class="card-eyebrow">五行雷达</span>
              <span class="card-meta">{{ data.summary.dominant_element }}主导</span>
            </header>
            <FortuneRadar :distribution="distribution" height="260px" />
          </div>
          <div class="glass-card">
            <header class="card-head">
              <span class="card-eyebrow">关系变化曲线</span>
              <span class="card-meta tabular-nums"
                >均 {{ data.summary.average_index.toFixed(1) }}</span
              >
            </header>
            <FortuneChart :daily-data="trendData" height="260px" />
          </div>
        </section>

        <!-- Day cards -->
        <section class="day-grid-section">
          <header class="card-head plain">
            <span class="card-eyebrow">每日详记</span>
            <span class="card-meta">{{ data.daily_fortunes.length }} 天</span>
          </header>
          <div class="day-grid">
            <DayFortuneCard
              v-for="d in data.daily_fortunes"
              :key="d.solar_date"
              :date="d.solar_date"
              :day-pillar="d.day_gan_zhi"
              :score="d.score"
              :ten-god="d.ten_god?.name"
              :weekday="weekdayFor(d.solar_date)"
              :is-highest="data!.summary.highest_index_day === d.solar_date"
              :is-lowest="data!.summary.lowest_index_day === d.solar_date"
            />
          </div>
        </section>

        <nav class="footer-nav">
          <router-link :to="`/fortune?chart_id=${chartId}`">今日运势</router-link>
          <span aria-hidden="true">·</span>
          <router-link :to="`/fortune/monthly?chart_id=${chartId}`">本月运势</router-link>
        </nav>
      </main>
    </template>
  </div>
</template>

<style scoped>
.weekly-page {
  position: relative;
  min-height: 100vh;
  color: var(--text);
  padding: 24px 16px 64px;
}

.page {
  position: relative;
  z-index: 1;
  max-width: 1200px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 24px;
}

/* states */
.state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 70vh;
  gap: 16px;
  color: var(--text-muted);
  font-size: var(--fs-sm);
}
.state.error p {
  color: var(--crimson);
}
.btn-link {
  color: rgba(var(--jade-accent-rgb), 1);
  text-decoration: none;
  font-weight: 600;
  letter-spacing: 0.04em;
}
.orb-skeleton {
  width: 100px;
  height: 100px;
  border-radius: 50%;
  background: radial-gradient(closest-side, rgba(var(--jade-accent-rgb), 0.45), transparent 70%);
  filter: blur(4px);
  animation: pulse 1.6s ease-in-out infinite;
}
@keyframes pulse {
  0%,
  100% {
    opacity: 0.4;
  }
  50% {
    opacity: 0.95;
  }
}

/* hero */
.hero {
  display: grid;
  grid-template-columns: 1fr;
  gap: 24px;
  padding: 24px;
  border-radius: 28px;
  background: linear-gradient(
    135deg,
    color-mix(in oklab, var(--surface-1) 86%, transparent),
    color-mix(in oklab, var(--surface-2) 78%, transparent)
  );
  border: 1px solid var(--line-strong);
  backdrop-filter: blur(22px) saturate(140%);
  box-shadow: var(--shadow-lg);
}
:global(.dark) .hero {
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.04), rgba(255, 255, 255, 0.02));
}
@media (min-width: 900px) {
  .hero {
    grid-template-columns: 1.4fr 1fr;
    align-items: center;
  }
}
.hero-left {
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-width: 0;
}
.hero-right {
  display: flex;
  align-items: center;
  justify-content: center;
}

.eyebrow {
  font-family: var(--font-mono), monospace;
  font-size: var(--fs-xs);
  letter-spacing: 0.42em;
  color: var(--text-muted);
  text-transform: uppercase;
}
.title {
  font-family: var(--font-serif), 'Songti SC', serif;
  font-size: var(--fs-stat);
  font-weight: 800;
  letter-spacing: 0.18em;
  margin: 0;
  color: var(--text);
}
.range {
  color: var(--text-muted);
  margin: 0;
  letter-spacing: 0.06em;
  font-size: var(--fs-sm);
}
.chips {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}
.meta-chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 10px;
  border-radius: 999px;
  border: 1px solid var(--line-subtle);
  font-size: var(--fs-xs);
  color: var(--text-muted);
  background: var(--glass-bg);
}
.meta-chip .dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  box-shadow: 0 0 6px currentColor;
}

/* glass card primitive */
.glass-card {
  padding: 20px;
  border-radius: 22px;
  background: linear-gradient(
    135deg,
    color-mix(in oklab, var(--surface-1) 88%, transparent),
    color-mix(in oklab, var(--surface-2) 78%, transparent)
  );
  border: 1px solid var(--line-strong);
  backdrop-filter: blur(22px) saturate(140%);
  box-shadow: var(--shadow-md);
}
:global(.dark) .glass-card {
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.04), rgba(255, 255, 255, 0.02));
}
.heat-card {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.card-head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 12px;
}
.card-head.plain {
  padding: 0 4px;
}
.card-eyebrow {
  font-size: var(--fs-xs);
  letter-spacing: 0.32em;
  color: var(--text-muted);
  font-family: var(--font-mono), monospace;
  text-transform: uppercase;
}
.card-meta {
  font-size: var(--fs-xs);
  color: var(--text-soft);
  letter-spacing: 0.06em;
}

.grid-2 {
  display: grid;
  gap: 24px;
  grid-template-columns: 1fr;
}
@media (min-width: 900px) {
  .grid-2 {
    grid-template-columns: 1fr 1fr;
  }
}

.day-grid-section {
  display: flex;
  flex-direction: column;
  gap: 14px;
}
.day-grid {
  display: grid;
  gap: 12px;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
}

.footer-nav {
  display: flex;
  gap: 14px;
  justify-content: center;
  padding: 8px 0;
  font-size: var(--fs-sm);
  color: var(--text-muted);
}
.footer-nav a {
  color: rgba(var(--jade-accent-rgb), 1);
  text-decoration: none;
  font-weight: 500;
  letter-spacing: 0.04em;
}
.footer-nav a:hover {
  text-shadow: 0 0 12px rgba(var(--jade-accent-rgb), 0.55);
}

.tabular-nums {
  font-variant-numeric: tabular-nums;
}
</style>
