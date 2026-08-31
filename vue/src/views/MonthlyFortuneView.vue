<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import {
  fetchMonthly,
  parseTrend,
  type MonthlyFortuneResponse,
  type FortuneDay,
} from '../api/fortune'
import { getApiErrorMessage } from '../api/client'
import FortuneChart from '../components/FortuneChart.vue'
import ScoreOrb from '../components/fortune/ScoreOrb.vue'
import DayFortuneCard from '../components/fortune/DayFortuneCard.vue'
import BestWorstChip from '../components/fortune/BestWorstChip.vue'
import PeriodNav from '../components/fortune/PeriodNav.vue'
import FortuneStateView from '../components/fortune/FortuneStateView.vue'
import { vReveal } from '../composables/useReveal'
import { useRecentChartStore } from '../stores/recentChart'

const route = useRoute()
const recentChartStore = useRecentChartStore()
const data = ref<MonthlyFortuneResponse | null>(null)
const loading = ref(true)
const error = ref('')
const errorKind = ref<'missing-chart' | 'request' | ''>('')
const chartId = ref<number | null>(null)
const expanded = ref(false)

const monthDays = computed(() => data.value?.daily_fortunes ?? [])

function relationCount(day: FortuneDay): number {
  return (day.supporting_evidence?.length ?? 0) + (day.counter_evidence?.length ?? 0)
}

const monthStats = computed(() => {
  const rows = monthDays.value.map((day) => ({ day, count: relationCount(day) }))
  const sorted = [...rows].sort((left, right) => right.count - left.count)
  const highest = sorted[0]
  const lowest = sorted[sorted.length - 1]
  const average = rows.length ? rows.reduce((sum, row) => sum + row.count, 0) / rows.length : 0
  return { highest, lowest, average }
})

const trendData = computed(() => {
  const elementByDate = new Map(
    parseTrend(data.value?.element_trend ?? '').map((item) => [item.date, item]),
  )
  return monthDays.value.map((day) => ({
    ...(elementByDate.get(day.solar_date) ?? {
      date: day.solar_date,
      metal: 0,
      wood: 0,
      water: 0,
      fire: 0,
      earth: 0,
    }),
    date: day.solar_date,
    score: relationCount(day),
  }))
})

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

function averageRelationCount(days: FortuneDay[]): number {
  if (!days.length) return 0
  return days.reduce((sum, day) => sum + relationCount(day), 0) / days.length
}

function asPercentage(value: number): number {
  return value <= 1 ? value * 100 : value
}

function formatMonthDay(date?: string): string {
  return date ? date.slice(5).replace('-', '/') : '—'
}

function rhythmWord(): string {
  const phases = phaseSegments.value
  if (phases.length < 2) return '旬段样本不足'
  const first = phases[0].average
  const last = phases[phases.length - 1].average
  if (last - first >= 1) return '后段命中关系较多'
  if (first - last >= 1) return '前段命中关系较多'
  return '各旬命中关系数接近'
}

const monthBriefTitle = computed(() => {
  const summary = data.value?.summary
  if (!summary) return '月内结构统计待定'
  return `本月关系节奏：${rhythmWord()}`
})

const monthBriefText = computed(() => {
  const summary = data.value?.summary
  if (!summary) return ''
  const phases = phaseSegments.value
  const phaseMeans = phases
    .map((phase) => `${phase.name}平均 ${phase.average.toFixed(1)} 条`)
    .join('、')
  const element = summary.dominant_element
    ? `日柱样本中${summary.dominant_element}出现较多`
    : '五行频次无单一最高项'
  const tenGod = summary.dominant_ten_god ? `，十神样本中${summary.dominant_ten_god}出现较多` : ''
  return `本月每天平均命中 ${monthStats.value.average.toFixed(1)} 条干支关系；${phaseMeans}。${element}${tenGod}。关系条数只反映规则命中数量，不代表吉凶。`
})

const heroBrief = computed(() => {
  if (!data.value?.summary) return ''
  return `本月每天平均命中 ${monthStats.value.average.toFixed(1)} 条干支关系，${rhythmWord()}。条数不代表吉凶，可先查看关系较多日，再结合具体关系类型阅读。`
})

const overviewStats = computed(() => {
  const summary = data.value?.summary
  if (!summary) return []
  const highest = monthStats.value.highest
  const lowest = monthStats.value.lowest
  return [
    {
      label: '每天平均命中',
      value: `${monthStats.value.average.toFixed(1)} 条`,
      detail: '按实际关系条数统计',
    },
    {
      label: '关系较多日',
      value: formatMonthDay(highest?.day.solar_date),
      detail: highest ? `${highest.day.day_gan_zhi} · ${highest.count} 条关系` : '—',
    },
    {
      label: '关系较少日',
      value: formatMonthDay(lowest?.day.solar_date),
      detail: lowest ? `${lowest.day.day_gan_zhi} · ${lowest.count} 条关系` : '—',
    },
    {
      label: '样本中较多五行',
      value: summary.dominant_element || '—',
      detail: summary.dominant_ten_god ? `十神频次：${summary.dominant_ten_god}` : '按日柱样本累计',
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

    const avg = averageRelationCount(chunk)
    const sorted = [...chunk].sort((a, b) => relationCount(b) - relationCount(a))
    const drift = relationCount(chunk[chunk.length - 1]) - relationCount(chunk[0])
    const driftLabel = drift >= 1 ? '旬末关系较多' : drift <= -1 ? '旬初关系较多' : '旬初旬末相近'
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
        description: `每天平均 ${avg.toFixed(1)} 条；较多日 ${formatMonthDay(highest.solar_date)}，较少日 ${formatMonthDay(lowest.solar_date)}。`,
      },
    ]
  })
})

const elementFocus = computed(() => {
  const entries = Object.entries(distribution.value)
  const max = Math.max(...entries.map(([, value]) => Number(value)), 1)
  const colors: Record<string, string> = {
    木: 'var(--wuxing-mu)',
    火: 'var(--wuxing-huo)',
    土: 'var(--wuxing-tu)',
    金: 'color-mix(in oklab, var(--wuxing-jin) 55%, var(--text-muted))',
    水: 'var(--wuxing-shui)',
  }
  return entries
    .sort((a, b) => Number(b[1]) - Number(a[1]))
    .map(([name, value]) => ({
      name,
      value: asPercentage(Number(value)),
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
  loading.value = true
  error.value = ''
  errorKind.value = ''
  data.value = null
  let cid: number | null = null
  const q = route.query.chart_id
  if (typeof q === 'string' && q) cid = Number(q)
  if (!cid) cid = recentChartStore.chartId
  if (!cid) {
    error.value = '请先创建命盘'
    errorKind.value = 'missing-chart'
    loading.value = false
    return
  }
  chartId.value = cid
  const { year, month } = currentYearMonth()

  try {
    data.value = await fetchMonthly(cid, year, month)
  } catch (reason: unknown) {
    const status = (reason as { response?: { status?: number } }).response?.status
    errorKind.value = status === 404 ? 'missing-chart' : 'request'
    error.value =
      status === 404
        ? '命盘不存在或已被删除，请重新排盘。'
        : getApiErrorMessage(reason, '本月内容加载失败，请稍后重试。')
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<template>
  <div class="monthly-page">
    <FortuneStateView
      v-if="loading"
      kind="loading"
      title="本月运势加载中"
      description="按当前规则汇总本月每日干支关系"
    />

    <FortuneStateView
      v-else-if="error"
      kind="error"
      :title="error"
      :description="errorKind === 'missing-chart' ? '本月运势需要基于一张已保存的命盘计算。' : ''"
      v-bind="
        errorKind === 'missing-chart'
          ? { actionLabel: '去排盘', actionTo: '/chart/new' }
          : { retryLabel: '重新加载' }
      "
      @retry="load"
    />

    <template v-else-if="data">
      <main class="page">
        <div class="top-nav" v-reveal>
          <PeriodNav current="month" :chart-id="chartId" />
        </div>

        <!-- ① 周期结论区 -->
        <section class="hero" v-reveal="40">
          <div class="hero-left">
            <span class="eyebrow">本月运势 · Monthly</span>
            <h1 class="title">{{ monthLabel }}</h1>
            <p class="range tabular-nums">{{ monthRange }}</p>
            <p class="month-brief-line">{{ heroBrief }}</p>
            <div class="chips">
              <BestWorstChip
                v-if="monthStats.highest"
                variant="highest"
                :date="monthStats.highest.day.solar_date"
                :count="monthStats.highest.count"
              />
              <BestWorstChip
                v-if="monthStats.lowest"
                variant="lowest"
                :date="monthStats.lowest.day.solar_date"
                :count="monthStats.lowest.count"
              />
              <span v-if="data.summary.dominant_element" class="meta-chip">
                <span class="dot"></span>
                {{ data.summary.dominant_element }}在样本中出现较多
              </span>
              <span v-if="data.summary.dominant_ten_god" class="meta-chip">
                {{ data.summary.dominant_ten_god }}
              </span>
            </div>
          </div>
          <div class="hero-right">
            <ScoreOrb
              :value="monthStats.average"
              label="每天平均命中关系"
              caption="按实际关系条数统计"
            />
          </div>
        </section>

        <!-- ② 结构化概览 -->
        <section class="overview-grid" v-reveal="80">
          <article v-for="stat in overviewStats" :key="stat.label" class="overview-card">
            <span class="overview-label">{{ stat.label }}</span>
            <strong class="overview-value tabular-nums">{{ stat.value }}</strong>
            <span class="overview-detail">{{ stat.detail }}</span>
          </article>
        </section>

        <section class="rhythm-card" v-reveal="120">
          <header class="card-head">
            <span class="card-eyebrow">上中下旬</span>
            <span class="card-meta">按上旬、中旬、下旬比较</span>
          </header>
          <div class="phase-grid">
            <article v-for="phase in phaseSegments" :key="phase.name" class="phase-card">
              <div class="phase-top">
                <div>
                  <span class="phase-name">{{ phase.name }}</span>
                  <span class="phase-range tabular-nums">{{ phase.range }}</span>
                </div>
                <strong class="phase-score tabular-nums"
                  >{{ phase.average.toFixed(1) }} 条/日</strong
                >
              </div>
              <div class="phase-meter" aria-hidden="true">
                <span
                  :style="{
                    width: `${Math.max(6, (phase.average / Math.max(monthStats.highest?.count ?? 1, 1)) * 100)}%`,
                  }"
                ></span>
              </div>
              <div class="phase-days">
                <span v-if="phase.highest"
                  >较多 {{ formatMonthDay(phase.highest.solar_date) }} ·
                  {{ relationCount(phase.highest) }} 条</span
                >
                <span v-if="phase.lowest"
                  >较少 {{ formatMonthDay(phase.lowest.solar_date) }} ·
                  {{ relationCount(phase.lowest) }} 条</span
                >
                <span>{{ phase.driftLabel }}</span>
              </div>
              <p>{{ phase.description }}</p>
            </article>
          </div>
        </section>

        <section class="surface-card trend-card" v-reveal="160">
          <header class="card-head">
            <span class="card-eyebrow">月内命中关系数</span>
            <span class="card-meta"> 平均 {{ monthStats.average.toFixed(1) }} 条 </span>
          </header>
          <FortuneChart :daily-data="trendData" height="320px" :show-elements="false" />
          <div class="trend-legend" aria-label="月内命中关系数说明">
            <span
              ><i class="legend-dot highest"></i>关系较多
              {{ formatMonthDay(monthStats.highest?.day.solar_date) }}</span
            >
            <span
              ><i class="legend-dot lowest"></i>关系较少
              {{ formatMonthDay(monthStats.lowest?.day.solar_date) }}</span
            >
          </div>
        </section>

        <!-- 月度总述 + 五行频次 -->
        <section class="month-brief" v-reveal="200">
          <div class="brief-copy">
            <span class="card-eyebrow">月度总述</span>
            <h2>{{ monthBriefTitle }}</h2>
            <p>{{ monthBriefText }}</p>
          </div>
          <div class="element-bars" aria-label="月内日柱样本五行频次">
            <div v-for="item in elementFocus" :key="item.name" class="element-row">
              <span class="element-name">{{ item.name }}</span>
              <div class="element-track">
                <span
                  class="element-fill"
                  :style="{ width: item.width, background: item.color }"
                ></span>
              </div>
              <span class="element-value tabular-nums">{{ item.value.toFixed(1) }}%</span>
            </div>
          </div>
        </section>

        <!-- ③ 明细区 -->
        <section class="day-section" v-reveal="240">
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
              :relation-count="relationCount(d)"
              :ten-god="d.ten_god?.name"
              :weekday="weekdayShort(d.solar_date)"
              :is-highest="monthStats.highest?.day.solar_date === d.solar_date"
              :is-lowest="monthStats.lowest?.day.solar_date === d.solar_date"
            />
          </div>
        </section>
      </main>
    </template>

    <FortuneStateView
      v-else
      kind="empty"
      title="本月暂无可显示的关系记录"
      description="可以稍后重新加载，或先查看今日、本周的记录。"
      retry-label="重新加载"
      @retry="load"
    />
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
  max-width: 1180px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.top-nav {
  display: flex;
  justify-content: center;
}

/* hero */
.hero {
  position: relative;
  display: grid;
  grid-template-columns: 1fr;
  gap: 24px;
  padding: 28px;
  border-radius: 16px;
  background: var(--surface-1);
  border: 1px solid var(--line-strong);
  box-shadow: var(--shadow-md);
}
.hero::before {
  content: '';
  position: absolute;
  top: -1px;
  left: 28px;
  right: 28px;
  height: 2px;
  background: rgba(var(--jade-accent-rgb), 0.55);
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
  font-size: var(--fs-2xs);
  letter-spacing: var(--tracking-meta, 0.18em);
  color: rgba(var(--jade-accent-rgb), 1);
  font-weight: 600;
  text-transform: uppercase;
}
.title {
  font-family: var(--font-serif), 'Songti SC', serif;
  font-size: clamp(22px, 2.6vw, 28px);
  font-weight: 800;
  letter-spacing: 0.04em;
  line-height: var(--lh-snug, 1.3);
  margin: 0;
  color: var(--text);
}
.range {
  color: var(--text-muted);
  margin: 0;
  letter-spacing: 0.06em;
  font-size: var(--fs-sm);
}
.month-brief-line {
  max-width: 64ch;
  margin: 0;
  color: var(--text-muted);
  font-size: var(--fs-sm);
  line-height: 1.7;
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
  background: var(--surface-2);
}
.meta-chip .dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--jade-accent);
}

/* card primitive：实心 surface + 1px 线 + 单层轻阴影 */
.surface-card {
  padding: 20px;
  border-radius: 14px;
  background: var(--surface-1);
  border: 1px solid var(--line-strong);
  box-shadow: var(--shadow-sm);
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
  font-size: var(--fs-2xs);
  letter-spacing: var(--tracking-meta, 0.18em);
  color: var(--text-muted);
  font-weight: 600;
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
  border-radius: 12px;
  background: var(--surface-1);
  border: 1px solid var(--line-subtle);
  box-shadow: var(--shadow-sm);
  transition:
    transform 180ms ease,
    border-color 180ms ease,
    box-shadow 180ms ease;
}
.overview-card:hover {
  transform: translateY(-1px);
  border-color: var(--line-strong);
  box-shadow: var(--shadow-md);
}
.overview-label,
.overview-detail,
.phase-range,
.phase-days {
  overflow-wrap: anywhere;
  white-space: normal;
}
.overview-label {
  font-size: var(--fs-2xs);
  letter-spacing: var(--tracking-meta, 0.18em);
  color: var(--text-muted);
  font-weight: 600;
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
  padding: 24px;
  border-radius: 14px;
  background: var(--surface-1);
  border: 1px solid var(--line-strong);
  box-shadow: var(--shadow-sm);
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
  font-size: var(--fs-sm);
}
.element-bars {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.element-row {
  display: grid;
  grid-template-columns: 28px minmax(0, 1fr) 48px;
  gap: 10px;
  align-items: center;
}
.element-name,
.element-value {
  font-size: var(--fs-xs);
  color: var(--text-muted);
}
.element-value {
  text-align: right;
  white-space: nowrap;
}
.element-track {
  position: relative;
  height: 8px;
  border-radius: 999px;
  background: var(--surface-3);
  overflow: hidden;
}
.element-fill {
  display: block;
  height: 100%;
  border-radius: inherit;
}

.rhythm-card {
  display: flex;
  flex-direction: column;
  gap: 14px;
  padding: 20px;
  border-radius: 14px;
  background: var(--surface-1);
  border: 1px solid var(--line-strong);
  box-shadow: var(--shadow-sm);
}

.phase-grid {
  display: grid;
  gap: 12px;
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.phase-card {
  min-width: 0;
  border-radius: 10px;
  padding: 16px;
  background: var(--surface-2);
  border: 1px solid var(--line-subtle);
}
.phase-top {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 12px;
}
.phase-name {
  font-size: var(--fs-sm);
  color: var(--text);
  margin-right: 8px;
  font-weight: 700;
}
.phase-range,
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
  height: 6px;
  border-radius: 999px;
  background: var(--surface-3);
  overflow: hidden;
  margin: 10px 0 8px;
}
.phase-meter span {
  display: block;
  height: 100%;
  border-radius: inherit;
  background: rgba(var(--jade-accent-rgb), 0.85);
}
.phase-days {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
  margin-bottom: 10px;
}
.phase-card p {
  margin: 0;
  line-height: 1.65;
  color: var(--text-muted);
  font-size: var(--fs-sm);
}

.overview-grid,
.month-brief,
.rhythm-card,
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
}
.legend-dot.lowest {
  background: var(--text-soft);
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

@media (max-width: 720px) {
  .phase-grid {
    grid-template-columns: 1fr;
  }
}

@media (prefers-reduced-motion: reduce) {
  .overview-card,
  .toggle {
    transition: none;
  }
}
</style>
