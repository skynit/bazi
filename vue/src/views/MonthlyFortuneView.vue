<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import {
  fetchMonthly,
  parseTrend,
  type MonthlyFortuneResponse,
  type FortuneDay,
} from '../api/fortune'
import FortuneChart from '../components/FortuneChart.vue'
import AuroraMeshBackground from '../components/fortune/AuroraMeshBackground.vue'
import ScoreOrb from '../components/fortune/ScoreOrb.vue'
import FortuneRadar from '../components/fortune/FortuneRadar.vue'
import DayFortuneCard from '../components/fortune/DayFortuneCard.vue'
import BestWorstChip from '../components/fortune/BestWorstChip.vue'

const route = useRoute()
const data = ref<MonthlyFortuneResponse | null>(null)
const loading = ref(true)
const error = ref('')
const chartId = ref<number | null>(null)
const expanded = ref(false)

const trendData = computed(() => parseTrend(data.value?.element_trend ?? ''))
const monthDays = computed(() => data.value?.daily_fortunes ?? [])

const monthLabel = computed(() => {
  const days = monthDays.value
  if (!days.length) return ''
  const [y, m] = days[0].solar_date.split('-')
  return `${y} 年 ${parseInt(m, 10)} 月`
})
const monthRange = computed(() => {
  const days = monthDays.value
  if (!days.length) return ''
  return `${days[0].solar_date} – ${days[days.length - 1].solar_date}`
})

const distribution = computed(() => data.value?.summary?.element_distribution ?? {})

function averageScore(days: FortuneDay[]): number {
  if (!days.length) return 0
  return days.reduce((sum, day) => sum + day.score, 0) / days.length
}

function formatMonthDay(date?: string): string {
  return date ? date.slice(5).replace('-', '/') : '—'
}

function rhythmWord(): string {
  const phases = phaseSegments.value
  if (phases.length < 2) return '旬段样本不足'
  const first = phases[0].average
  const last = phases[phases.length - 1].average
  if (last - first >= 8) return '后段均值较高'
  if (first - last >= 8) return '前段均值较高'
  if (data.value && data.value.summary.index_standard_deviation >= 12) return '日值离散度较高'
  return '旬段均值接近'
}

const monthBriefTitle = computed(() => {
  const summary = data.value?.summary
  if (!summary) return '月内结构统计待定'
  return `结构指数分布：${rhythmWord()}`
})

const monthBriefText = computed(() => {
  const summary = data.value?.summary
  if (!summary) return ''
  const phases = phaseSegments.value
  const phaseMeans = phases.map((phase) => `${phase.name} ${phase.average.toFixed(1)}`).join('、')
  const element = summary.dominant_element
    ? `五行频次以${summary.dominant_element}为最高`
    : '五行频次无单一最高项'
  const tenGod = summary.dominant_ten_god ? `，十神频次以${summary.dominant_ten_god}为最高` : ''
  return `本月结构关系指数均值 ${summary.average_index.toFixed(1)}，标准差 ${summary.index_standard_deviation.toFixed(1)}；${phaseMeans}。${element}${tenGod}。以上仅为结构统计，现实结果未裁决。`
})

const scoreSpread = computed(() => {
  const summary = data.value?.summary
  if (!summary) return 0
  return Math.max(0, summary.highest_index - summary.lowest_index)
})

const overviewStats = computed(() => {
  const summary = data.value?.summary
  if (!summary) return []
  return [
    {
      label: '月均指数',
      value: summary.average_index.toFixed(1),
      detail: '结构关系指数均值',
    },
    {
      label: '振幅',
      value: scoreSpread.value.toString(),
      detail: `最高 ${formatMonthDay(summary.highest_index_day)} · 最低 ${formatMonthDay(summary.lowest_index_day)}`,
    },
    {
      label: '标准差',
      value: summary.index_standard_deviation.toFixed(1),
      detail: '月内日指数离散度',
    },
    {
      label: '主气',
      value: summary.dominant_element || '—',
      detail: summary.dominant_ten_god ? `十神 ${summary.dominant_ten_god}` : '五行占优',
    },
  ]
})

interface PhaseSegment {
  name: string
  range: string
  average: number
  driftLabel: string
  highest?: FortuneDay
  lowest?: FortuneDay
  description: string
}

const phaseSegments = computed<PhaseSegment[]>(() => {
  const buckets = [
    { name: '上旬', from: 1, to: 10 },
    { name: '中旬', from: 11, to: 20 },
    { name: '下旬', from: 21, to: 31 },
  ]

  return buckets.flatMap((bucket) => {
    const chunk = monthDays.value.filter((day) => {
      const dateNum = Number(day.solar_date.slice(8, 10))
      return dateNum >= bucket.from && dateNum <= bucket.to
    })
    if (!chunk.length) return []

    const avg = averageScore(chunk)
    const sorted = [...chunk].sort((a, b) => b.score - a.score)
    const drift = chunk[chunk.length - 1].score - chunk[0].score
    const driftLabel = drift >= 4 ? '末值较高' : drift <= -4 ? '末值较低' : '首末接近'
    const highest = sorted[0]
    const lowest = sorted[sorted.length - 1]
    return [
      {
        name: bucket.name,
        range: `${formatMonthDay(chunk[0].solar_date)} - ${formatMonthDay(chunk[chunk.length - 1].solar_date)}`,
        average: avg,
        driftLabel,
        highest,
        lowest,
        description: `旬均值 ${avg.toFixed(1)}；最高 ${formatMonthDay(highest.solar_date)}，最低 ${formatMonthDay(lowest.solar_date)}。`,
      },
    ]
  })
})

interface ExtremeDayCard {
  date: string
  label: string
  score: number
  pillar: string
  variant: 'highest' | 'lowest'
  detail: string
}

const extremeDayCards = computed<ExtremeDayCard[]>(() => {
  const summary = data.value?.summary
  if (!summary) return []
  const byDate = new Map(monthDays.value.map((day) => [day.solar_date, day]))
  const cards: ExtremeDayCard[] = []

  function push(date: string | undefined, label: string, variant: ExtremeDayCard['variant']) {
    if (!date) return
    const day = byDate.get(date)
    if (!day) return
    cards.push({
      date,
      label,
      score: day.score,
      pillar: day.day_gan_zhi,
      variant,
      detail: `干支 ${day.day_gan_zhi} · 结构指数 ${day.score}`,
    })
  }

  push(summary.highest_index_day, '月内最高值', 'highest')
  push(summary.lowest_index_day, '月内最低值', 'lowest')

  return cards
})

const elementFocus = computed(() => {
  const entries = Object.entries(distribution.value)
  const max = Math.max(...entries.map(([, value]) => Number(value)), 1)
  const colors: Record<string, string> = {
    木: 'var(--wuxing-mu)',
    火: 'var(--wuxing-huo)',
    土: 'var(--wuxing-tu)',
    金: 'var(--wuxing-jin)',
    水: 'var(--wuxing-shui)',
  }
  return entries
    .sort((a, b) => Number(b[1]) - Number(a[1]))
    .map(([name, value]) => ({
      name,
      value: Number(value),
      width: `${Math.max(8, (Number(value) / max) * 100)}%`,
      color: colors[name] ?? 'var(--jade-accent)',
    }))
})

const visibleDays = computed<FortuneDay[]>(() => {
  const days = data.value?.daily_fortunes ?? []
  return expanded.value ? days : days.slice(0, 14)
})

function weekdayShort(date: string): string {
  const dt = new Date(date + 'T00:00:00')
  return ['一', '二', '三', '四', '五', '六', '日'][(dt.getDay() + 6) % 7]
}

function currentYearMonth(): { year: number; month: number } {
  const d = new Date()
  return { year: d.getFullYear(), month: d.getMonth() + 1 }
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
  const { year, month } = currentYearMonth()

  try {
    data.value = await fetchMonthly(cid, year, month)
  } catch (e: any) {
    error.value = e?.response?.data?.error || '加载月运势失败'
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<template>
  <div class="monthly-page">
    <AuroraMeshBackground />

    <div v-if="loading" class="state">
      <div class="orb-skeleton" aria-hidden="true"></div>
      <p>本月运势加载中…</p>
    </div>

    <div v-else-if="error" class="state error">
      <p>{{ error }}</p>
      <router-link to="/chart/new" class="btn-link">去排盘 →</router-link>
    </div>

    <template v-else-if="data">
      <main class="page">
        <section class="hero">
          <div class="hero-left">
            <span class="eyebrow">BaZi · Monthly</span>
            <h1 class="title">{{ monthLabel }}</h1>
            <p class="range tabular-nums">{{ monthRange }}</p>
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
                <span class="dot"></span>
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
              label="月均结构指数"
              :caption="`标准差 ${data.summary.index_standard_deviation.toFixed(1)}`"
            />
          </div>
        </section>

        <section class="overview-grid">
          <article v-for="stat in overviewStats" :key="stat.label" class="overview-card">
            <span class="overview-label">{{ stat.label }}</span>
            <strong class="overview-value tabular-nums">{{ stat.value }}</strong>
            <span class="overview-detail">{{ stat.detail }}</span>
          </article>
        </section>

        <section class="month-brief">
          <div class="brief-copy">
            <span class="card-eyebrow">月度总述</span>
            <h2>{{ monthBriefTitle }}</h2>
            <p>{{ monthBriefText }}</p>
          </div>
          <div class="element-bars" aria-label="五行占比">
            <div v-for="item in elementFocus" :key="item.name" class="element-row">
              <span class="element-name">{{ item.name }}</span>
              <div class="element-track">
                <span
                  class="element-fill"
                  :style="{ width: item.width, background: item.color }"
                ></span>
              </div>
              <span class="element-value tabular-nums">{{ item.value }}</span>
            </div>
          </div>
        </section>

        <section class="rhythm-card">
          <header class="card-head">
            <span class="card-eyebrow">上中下旬</span>
            <span class="card-meta">按日历旬段汇总结构指数</span>
          </header>
          <div class="phase-grid">
            <article v-for="phase in phaseSegments" :key="phase.name" class="phase-card">
              <div class="phase-top">
                <div>
                  <span class="phase-name">{{ phase.name }}</span>
                  <span class="phase-range tabular-nums">{{ phase.range }}</span>
                </div>
                <strong class="phase-score tabular-nums">{{ phase.average.toFixed(1) }}</strong>
              </div>
              <div class="phase-meter" aria-hidden="true">
                <span :style="{ width: `${Math.max(6, Math.min(100, phase.average))}%` }"></span>
              </div>
              <div class="phase-days">
                <span v-if="phase.highest"
                  >最高 {{ formatMonthDay(phase.highest.solar_date) }} ·
                  {{ phase.highest.score }}</span
                >
                <span v-if="phase.lowest"
                  >最低 {{ formatMonthDay(phase.lowest.solar_date) }} ·
                  {{ phase.lowest.score }}</span
                >
                <span>{{ phase.driftLabel }}</span>
              </div>
              <p>{{ phase.description }}</p>
            </article>
          </div>
        </section>

        <section class="glass-card trend-card">
          <header class="card-head">
            <span class="card-eyebrow">月内结构指数</span>
            <span class="card-meta">
              均值 {{ data.summary.average_index.toFixed(1) }} · 标准差
              {{ data.summary.index_standard_deviation.toFixed(1) }}
            </span>
          </header>
          <FortuneChart :daily-data="trendData" height="320px" :show-elements="false" />
          <div class="trend-legend" aria-label="月内结构指数说明">
            <span
              ><i class="legend-dot highest"></i>最高
              {{ formatMonthDay(data.summary.highest_index_day) }}</span
            >
            <span
              ><i class="legend-dot lowest"></i>最低
              {{ formatMonthDay(data.summary.lowest_index_day) }}</span
            >
          </div>
        </section>

        <section class="key-days-card">
          <header class="card-head">
            <span class="card-eyebrow">指数极值日期</span>
            <span class="card-meta">仅表示本月内相对最高与最低</span>
          </header>
          <div class="key-day-grid">
            <article
              v-for="day in extremeDayCards"
              :key="day.date"
              class="key-day"
              :class="day.variant"
            >
              <span class="key-date tabular-nums">{{ formatMonthDay(day.date) }}</span>
              <div class="key-main">
                <strong>{{ day.label }}</strong>
                <span>{{ day.pillar }} · 结构指数 {{ day.score }}</span>
              </div>
              <p>{{ day.detail }}</p>
            </article>
          </div>
        </section>

        <section class="grid-2">
          <div class="glass-card">
            <header class="card-head">
              <span class="card-eyebrow">五行雷达</span>
              <span class="card-meta">{{ data.summary.dominant_element }}主导</span>
            </header>
            <FortuneRadar :distribution="distribution" height="280px" />
          </div>
          <div class="glass-card">
            <header class="card-head">
              <span class="card-eyebrow">结构指数与五行走势</span>
              <span class="card-meta tabular-nums"
                >均 {{ data.summary.average_index.toFixed(1) }}</span
              >
            </header>
            <FortuneChart :daily-data="trendData" height="280px" />
          </div>
        </section>

        <section class="day-section">
          <header class="card-head plain">
            <span class="card-eyebrow">每日详记</span>
            <span v-if="!expanded" class="card-meta">先看前 14 天</span>
            <button type="button" class="toggle" @click="expanded = !expanded">
              {{ expanded ? '收起' : `展开全部 ${data.daily_fortunes.length} 天` }}
            </button>
          </header>
          <div class="day-grid">
            <DayFortuneCard
              v-for="d in visibleDays"
              :key="d.solar_date"
              :date="d.solar_date"
              :day-pillar="d.day_gan_zhi"
              :score="d.score"
              :ten-god="d.ten_god?.name"
              :weekday="weekdayShort(d.solar_date)"
              :is-highest="data!.summary.highest_index_day === d.solar_date"
              :is-lowest="data!.summary.lowest_index_day === d.solar_date"
            />
          </div>
        </section>

        <nav class="footer-nav">
          <router-link :to="`/fortune?chart_id=${chartId}`">今日运势</router-link>
          <span aria-hidden="true">·</span>
          <router-link :to="`/fortune/weekly?chart_id=${chartId}`">本周运势</router-link>
        </nav>
      </main>
    </template>
  </div>
</template>

<style scoped>
.monthly-page {
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

/* hero (shared with weekly) */
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
  background: var(--jade-accent);
  box-shadow: 0 0 6px var(--jade-accent);
}

/* glass card */
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

.overview-grid {
  display: grid;
  gap: 12px;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.overview-card {
  display: flex;
  flex-direction: column;
  gap: 8px;
  min-width: 0;
  padding: 18px 18px 16px;
  border-radius: 20px;
  background: linear-gradient(
    135deg,
    var(--glass-bg),
    color-mix(in oklab, var(--surface-2) 70%, transparent)
  );
  border: 1px solid var(--line-subtle);
  box-shadow: var(--shadow-sm);
}
.overview-label,
.overview-detail,
.phase-range,
.phase-days,
.key-main span,
.element-value {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.overview-label {
  font-size: var(--fs-xs);
  letter-spacing: 0.28em;
  color: var(--text-muted);
  text-transform: uppercase;
}
.overview-value {
  font-family: var(--font-serif), serif;
  font-size: var(--fs-display);
  line-height: 1;
  color: var(--text);
}
.overview-detail {
  font-size: var(--fs-xs);
  color: var(--text-soft);
}

.month-brief {
  display: grid;
  gap: 18px;
  padding: 22px;
  border-radius: 24px;
  background:
    linear-gradient(
      135deg,
      color-mix(in oklab, var(--surface-1) 84%, transparent),
      color-mix(in oklab, var(--surface-2) 76%, transparent)
    ),
    radial-gradient(circle at top right, rgba(var(--jade-accent-rgb), 0.08), transparent 32%);
  border: 1px solid var(--line-strong);
  box-shadow: var(--shadow-md);
}
.brief-copy {
  display: flex;
  flex-direction: column;
  gap: 10px;
  min-width: 0;
}
.brief-copy h2 {
  margin: 0;
  font-family: var(--font-serif), serif;
  font-size: var(--fs-display);
  letter-spacing: 0.08em;
}
.brief-copy p {
  margin: 0;
  line-height: 1.8;
  color: var(--text);
  max-width: 66ch;
}
.element-bars {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.element-row {
  display: grid;
  grid-template-columns: 28px minmax(0, 1fr) 34px;
  gap: 10px;
  align-items: center;
}
.element-name,
.element-value {
  font-size: var(--fs-xs);
  color: var(--text-muted);
}
.element-track {
  position: relative;
  height: 10px;
  border-radius: 999px;
  background: rgba(var(--jade-accent-rgb), 0.07);
  overflow: hidden;
}
.element-fill {
  display: block;
  height: 100%;
  border-radius: inherit;
  box-shadow: 0 0 14px rgba(var(--jade-accent-rgb), 0.22);
}

.rhythm-card,
.key-days-card {
  display: flex;
  flex-direction: column;
  gap: 14px;
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

.phase-grid,
.key-day-grid {
  display: grid;
  gap: 12px;
}
.phase-grid {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}
.key-day-grid {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.phase-card,
.key-day {
  min-width: 0;
  border-radius: 18px;
  padding: 16px;
  background: var(--glass-bg);
  border: 1px solid var(--line-subtle);
}
.key-day.highest {
  border-color: rgba(var(--jade-accent-rgb), 0.2);
}
.key-day.lowest {
  border-color: var(--line-strong);
}
.phase-top,
.key-main {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 12px;
}
.phase-name,
.key-main strong {
  font-size: var(--fs-sm);
  color: var(--text);
}
.phase-range,
.key-main span,
.phase-days {
  font-size: var(--fs-xs);
  color: var(--text-soft);
}
.phase-score {
  font-family: var(--font-serif), serif;
  font-size: var(--fs-2xl);
  color: var(--text);
}
.phase-meter {
  height: 8px;
  border-radius: 999px;
  background: rgba(var(--jade-accent-rgb), 0.08);
  overflow: hidden;
  margin: 10px 0 8px;
}
.phase-meter span {
  display: block;
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(
    90deg,
    rgba(var(--jade-accent-rgb), 0.55),
    rgba(var(--jade-accent-rgb), 1)
  );
}
.phase-days {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
  margin-bottom: 10px;
}
.phase-card p,
.key-day p {
  margin: 0;
  line-height: 1.65;
  color: var(--text-muted);
  font-size: var(--fs-sm);
}

.key-day {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.key-date {
  font-size: var(--fs-xs);
  letter-spacing: 0.18em;
  color: var(--text-soft);
  text-transform: uppercase;
}
.key-main {
  flex-direction: column;
  align-items: flex-start;
  gap: 2px;
}

.overview-grid,
.month-brief,
.rhythm-card,
.key-days-card,
.grid-2,
.day-section {
  min-width: 0;
}

.trend-card {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.trend-legend {
  display: flex;
  flex-wrap: wrap;
  gap: 10px 16px;
  align-items: center;
  padding: 0 4px;
  color: var(--text-muted);
  font-size: var(--fs-xs);
  letter-spacing: 0.06em;
}
.trend-legend span {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}
.legend-dot {
  width: 7px;
  height: 7px;
  border-radius: 999px;
  flex: 0 0 auto;
}
.legend-dot.highest {
  background: var(--jade-accent);
  box-shadow: 0 0 8px rgba(var(--jade-accent-rgb), 0.75);
}
.legend-dot.lowest {
  background: var(--text-soft);
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

.day-section {
  display: flex;
  flex-direction: column;
  gap: 14px;
}
.day-grid {
  display: grid;
  gap: 12px;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
}

.day-section .card-head.plain {
  align-items: center;
}

.toggle {
  background: transparent;
  border: 1px solid var(--line-strong);
  color: var(--text);
  padding: 6px 14px;
  border-radius: 999px;
  font-size: var(--fs-xs);
  letter-spacing: 0.06em;
  cursor: pointer;
  transition:
    border-color 0.2s ease,
    color 0.2s ease;
}
.toggle:hover {
  border-color: rgba(var(--jade-accent-rgb), 0.55);
  color: rgba(var(--jade-accent-rgb), 1);
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

@media (min-width: 900px) {
  .overview-grid {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }
  .month-brief {
    grid-template-columns: minmax(0, 1.35fr) minmax(280px, 0.65fr);
    align-items: center;
  }
}

@media (max-width: 1100px) {
  .phase-grid,
  .key-day-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 760px) {
  .page {
    gap: 18px;
  }
  .hero,
  .glass-card,
  .rhythm-card,
  .key-days-card,
  .month-brief {
    padding: 16px;
    border-radius: 20px;
  }
  .overview-grid {
    grid-template-columns: 1fr;
  }
  .phase-grid,
  .key-day-grid {
    grid-template-columns: 1fr;
  }
  .card-head {
    flex-direction: column;
    align-items: flex-start;
  }
  .day-section .card-head.plain {
    gap: 10px;
  }
  .toggle {
    align-self: flex-start;
  }
}
</style>
