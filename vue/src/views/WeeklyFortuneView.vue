<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { fetchWeekly, parseTrend, type WeeklyFortuneResponse } from '../api/fortune'
import { getApiErrorMessage } from '../api/client'
import FortuneChart from '../components/FortuneChart.vue'
import ScoreOrb from '../components/fortune/ScoreOrb.vue'
import FortuneHeatStrip from '../components/fortune/FortuneHeatStrip.vue'
import FortuneRadar from '../components/fortune/FortuneRadar.vue'
import DayFortuneCard from '../components/fortune/DayFortuneCard.vue'
import BestWorstChip from '../components/fortune/BestWorstChip.vue'
import PeriodNav from '../components/fortune/PeriodNav.vue'
import FortuneStateView from '../components/fortune/FortuneStateView.vue'
import { vReveal } from '../composables/useReveal'

const route = useRoute()
const data = ref<WeeklyFortuneResponse | null>(null)
const loading = ref(true)
const error = ref('')
const errorKind = ref<'missing-chart' | 'request' | ''>('')
const chartId = ref<number | null>(null)

const WEEKDAYS = ['一', '二', '三', '四', '五', '六', '日'] as const

function relationCount(day: WeeklyFortuneResponse['daily_fortunes'][number]) {
  return (day.supporting_evidence?.length ?? 0) + (day.counter_evidence?.length ?? 0)
}

const periodStats = computed(() => {
  const days = data.value?.daily_fortunes ?? []
  const rows = days.map((day) => ({ day, count: relationCount(day) }))
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
  return (data.value?.daily_fortunes ?? []).map((day) => ({
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

const weekRange = computed(() => {
  const days = data.value?.daily_fortunes ?? []
  if (!days.length) return ''
  return `${days[0].solar_date} – ${days[days.length - 1].solar_date}`
})

const heatDays = computed(() => {
  return (data.value?.daily_fortunes ?? []).map((d) => ({
    date: d.solar_date,
    relationCount: relationCount(d),
    dayPillar: d.day_gan_zhi,
    isHighest: periodStats.value.highest?.day.solar_date === d.solar_date,
    isLowest: periodStats.value.lowest?.day.solar_date === d.solar_date,
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
  loading.value = true
  error.value = ''
  errorKind.value = ''
  data.value = null
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
    errorKind.value = 'missing-chart'
    loading.value = false
    return
  }
  chartId.value = cid

  try {
    data.value = await fetchWeekly(cid, todayStr())
  } catch (reason: unknown) {
    const status = (reason as { response?: { status?: number } }).response?.status
    errorKind.value = status === 404 ? 'missing-chart' : 'request'
    error.value =
      status === 404
        ? '命盘不存在或已被删除，请重新排盘。'
        : getApiErrorMessage(reason, '本周内容加载失败，请稍后重试。')
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
    <FortuneStateView
      v-if="loading"
      kind="loading"
      title="本周运势加载中"
      description="按当前规则汇总本周每日干支关系"
    />

    <FortuneStateView
      v-else-if="error"
      kind="error"
      :title="error"
      :description="errorKind === 'missing-chart' ? '本周运势需要基于一张已保存的命盘计算。' : ''"
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
          <PeriodNav current="week" :chart-id="chartId" />
        </div>

        <!-- ① 周期结论区 -->
        <section class="hero" v-reveal="40">
          <div class="hero-left">
            <span class="eyebrow">本周运势 · Weekly</span>
            <h1 class="title">本周关系节奏</h1>
            <p class="range tabular-nums">{{ weekRange }}</p>
            <p class="week-brief">
              本周每天平均命中
              {{ periodStats.average.toFixed(1) }}
              条干支关系。可先查看关系较多日，再结合具体关系类型阅读；条数不代表吉凶。
            </p>
            <div class="chips">
              <BestWorstChip
                v-if="periodStats.highest"
                variant="highest"
                :date="periodStats.highest.day.solar_date"
                :count="periodStats.highest.count"
              />
              <BestWorstChip
                v-if="periodStats.lowest"
                variant="lowest"
                :date="periodStats.lowest.day.solar_date"
                :count="periodStats.lowest.count"
              />
              <span v-if="data.summary.dominant_element" class="meta-chip">
                <span class="dot" :style="{ background: 'var(--jade-accent)' }"></span>
                {{ data.summary.dominant_element }}在样本中出现较多
              </span>
              <span v-if="data.summary.dominant_ten_god" class="meta-chip">
                {{ data.summary.dominant_ten_god }}
              </span>
            </div>
          </div>
          <div class="hero-right">
            <ScoreOrb
              :value="periodStats.average"
              label="每天平均命中关系"
              caption="按实际关系条数统计"
            />
          </div>
        </section>

        <!-- ② 结构化概览 -->
        <section class="surface-card heat-card" v-reveal="80">
          <header class="card-head">
            <span class="card-eyebrow">每日命中关系数</span>
            <span class="card-meta">
              较多 {{ periodStats.highest?.count ?? 0 }} 条 · 较少
              {{ periodStats.lowest?.count ?? 0 }} 条
            </span>
          </header>
          <FortuneHeatStrip :days="heatDays" :weekday-labels="weekdayLabels" />
        </section>

        <section class="grid-2">
          <div class="surface-card" v-reveal="120">
            <header class="card-head">
              <span class="card-eyebrow">五行雷达</span>
              <span class="card-meta">{{ data.summary.dominant_element }}出现较多</span>
            </header>
            <FortuneRadar :distribution="distribution" height="260px" />
          </div>
          <div class="surface-card" v-reveal="160">
            <header class="card-head">
              <span class="card-eyebrow">命中关系数变化</span>
              <span class="card-meta tabular-nums"
                >平均 {{ periodStats.average.toFixed(1) }} 条</span
              >
            </header>
            <FortuneChart :daily-data="trendData" height="260px" />
          </div>
        </section>

        <!-- ③ 明细区 -->
        <section class="day-grid-section" v-reveal="200">
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
              :relation-count="relationCount(d)"
              :ten-god="d.ten_god?.name"
              :weekday="weekdayFor(d.solar_date)"
              :is-highest="periodStats.highest?.day.solar_date === d.solar_date"
              :is-lowest="periodStats.lowest?.day.solar_date === d.solar_date"
            />
          </div>
        </section>
      </main>
    </template>

    <FortuneStateView
      v-else
      kind="empty"
      title="本周暂无可显示的关系记录"
      description="可以稍后重新加载，或先查看今日、本月的记录。"
      retry-label="重新加载"
      @retry="load"
    />
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
.week-brief {
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
}

/* card primitive：实心 surface + 1px 线 + 单层轻阴影 */
.surface-card {
  padding: 20px;
  border-radius: 14px;
  background: var(--surface-1);
  border: 1px solid var(--line-strong);
  box-shadow: var(--shadow-sm);
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

.tabular-nums {
  font-variant-numeric: tabular-nums;
}
</style>
