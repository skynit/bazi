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

function scoreWord(score: number): string {
  if (score >= 85) return '大吉'
  if (score >= 72) return '顺行'
  if (score >= 60) return '平稳'
  if (score >= 48) return '蓄势'
  return '谨慎'
}

function scoreMood(score: number): string {
  if (score >= 72) return 'good'
  if (score >= 58) return 'steady'
  return 'caution'
}

function rhythmWord(): string {
  const phases = phaseSegments.value
  if (phases.length < 2) return '节奏平缓'
  const first = phases[0].average
  const last = phases[phases.length - 1].average
  if (last - first >= 8) return '后劲渐起'
  if (first - last >= 8) return '前强后收'
  if (scoreSpread.value >= 35) return '起伏明显'
  return '节奏平稳'
}

const monthBriefTitle = computed(() => {
  const summary = data.value?.summary
  if (!summary) return '月内节奏待定'
  const element = summary.dominant_element ? `${summary.dominant_element}气` : '主气'
  return `${rhythmWord()}，${element}当令`
})

const monthBriefText = computed(() => {
  const summary = data.value?.summary
  if (!summary) return ''
  const phases = phaseSegments.value
  const strongest = [...phases].sort((a, b) => b.average - a.average)[0]
  const softest = [...phases].sort((a, b) => a.average - b.average)[0]
  const good = scoreBands.value.good
  const caution = scoreBands.value.caution
  const tenGod = summary.dominant_ten_god ? `，十神以${summary.dominant_ten_god}露头` : ''
  const action = good > caution
    ? '可把沟通、推进与定案放在高分段完成'
    : caution > good
      ? '宜先稳住节奏，减少仓促承诺与高风险决策'
      : '宜稳中求进，把重要事项拆小处理'
  const phaseNote = strongest && softest && strongest.name !== softest.name
    ? `${strongest.name}承接力较强，${softest.name}更适合复盘和留白`
    : '三旬节奏差异不大，重在持续执行'
  return `本月${summary.dominant_element || '五行'}气偏显${tenGod}，整体呈${rhythmWord()}。${phaseNote}；${action}。`
})

function phaseAdvice(avg: number, drift: number): string {
  if (avg >= 76) return '适合主动推进、定计划、谈合作，把机会窗口用满。'
  if (avg >= 64 && drift >= 4) return '走势渐开，先稳住节奏，再把重点事项放到后半段。'
  if (avg >= 60) return '以稳定执行为主，适合整理资源、复盘账目与补足短板。'
  if (avg >= 50) return '少做高风险决策，优先处理确定性强的小事。'
  return '宜静不宜躁，重要承诺、投资和冲突场景都要留缓冲。'
}

function topCounts(values: Array<string | number | undefined>, limit = 3) {
  const counts = new Map<string, number>()
  values.forEach(value => {
    if (value === undefined || value === null || value === '') return
    const key = String(value)
    counts.set(key, (counts.get(key) ?? 0) + 1)
  })
  return [...counts.entries()]
    .sort((a, b) => b[1] - a[1] || a[0].localeCompare(b[0], 'zh-CN'))
    .slice(0, limit)
    .map(([label, count]) => ({ label, count }))
}

function colorSwatch(name: string): string {
  const map: Record<string, string> = {
    '红色': '#e84057', '红': '#e84057', '朱红': '#dc2626',
    '橙色': '#f97316', '黄色': '#fcd34d', '金色': '#d4a017', '金': '#d4a017',
    '绿色': '#22c55e', '青色': '#10b981', '翠': '#34d399',
    '蓝色': '#3b82f6', '青蓝': '#0ea5e9',
    '紫色': '#a855f7', '紫': '#a855f7',
    '黑色': '#1f2937', '黑': '#1f2937',
    '白色': '#f8fafc', '白': '#f8fafc',
    '灰色': '#94a3b8',
  }
  return map[name] ?? '#94a3b8'
}

const scoreBands = computed(() => {
  const days = monthDays.value
  return {
    good: days.filter(day => day.score >= 72).length,
    steady: days.filter(day => day.score >= 58 && day.score < 72).length,
    caution: days.filter(day => day.score < 58).length,
  }
})

const scoreSpread = computed(() => {
  const summary = data.value?.summary
  if (!summary) return 0
  return Math.max(0, summary.best_score - summary.worst_score)
})

const overviewStats = computed(() => {
  const summary = data.value?.summary
  if (!summary) return []
  return [
    {
      label: '月均分',
      value: summary.average_score.toFixed(1),
      detail: `${scoreWord(summary.average_score)} · ${scoreBands.value.good} 天顺行`,
      mood: scoreMood(summary.average_score),
    },
    {
      label: '振幅',
      value: scoreSpread.value.toString(),
      detail: `${formatMonthDay(summary.best_day)} 至 ${formatMonthDay(summary.worst_day)}`,
      mood: scoreSpread.value > 30 ? 'caution' : 'steady',
    },
    {
      label: '连吉',
      value: `${summary.good_streak}d`,
      detail: `低谷连段 ${summary.bad_streak}d`,
      mood: summary.good_streak >= summary.bad_streak ? 'good' : 'caution',
    },
    {
      label: '主气',
      value: summary.dominant_element || '—',
      detail: summary.dominant_ten_god ? `十神 ${summary.dominant_ten_god}` : '五行占优',
      mood: 'good',
    },
  ]
})

interface PhaseSegment {
  name: string
  range: string
  average: number
  mood: string
  driftLabel: string
  best?: FortuneDay
  low?: FortuneDay
  advice: string
}

const phaseSegments = computed<PhaseSegment[]>(() => {
  const buckets = [
    { name: '上旬', from: 1, to: 10 },
    { name: '中旬', from: 11, to: 20 },
    { name: '下旬', from: 21, to: 31 },
  ]

  return buckets.flatMap(bucket => {
    const chunk = monthDays.value.filter(day => {
      const dateNum = Number(day.solar_date.slice(8, 10))
      return dateNum >= bucket.from && dateNum <= bucket.to
    })
    if (!chunk.length) return []

    const avg = averageScore(chunk)
    const sorted = [...chunk].sort((a, b) => b.score - a.score)
    const drift = chunk[chunk.length - 1].score - chunk[0].score
    const driftLabel = drift >= 4 ? '走升' : drift <= -4 ? '回落' : '平稳'
    return [{
      name: bucket.name,
      range: `${formatMonthDay(chunk[0].solar_date)} - ${formatMonthDay(chunk[chunk.length - 1].solar_date)}`,
      average: avg,
      mood: scoreMood(avg),
      driftLabel,
      best: sorted[0],
      low: sorted[sorted.length - 1],
      advice: phaseAdvice(avg, drift),
    }]
  })
})

interface KeyDayCard {
  date: string
  label: string
  score: number
  pillar: string
  variant: 'best' | 'worst' | 'peak' | 'low'
  detail: string
}

const keyDayCards = computed<KeyDayCard[]>(() => {
  const summary = data.value?.summary
  if (!summary) return []
  const byDate = new Map(monthDays.value.map(day => [day.solar_date, day]))
  const used = new Set<string>()
  const cards: KeyDayCard[] = []

  function push(date: string | undefined, label: string, variant: KeyDayCard['variant']) {
    if (!date || used.has(date)) return
    const day = byDate.get(date)
    if (!day) return
    used.add(date)
    const isLow = variant === 'worst' || variant === 'low'
    const detail = isLow
      ? (day.ji?.slice(0, 2).join('、') || '放慢节奏，重要事项留余地')
      : (day.yi?.slice(0, 2).join('、') || '适合推进重点事项')
    cards.push({
      date,
      label,
      score: day.score,
      pillar: day.day_gan_zhi,
      variant,
      detail,
    })
  }

  push(summary.best_day, '月内吉峰', 'best')
  push(summary.worst_day, '谨慎低谷', 'worst')
  summary.peak_days.slice(0, 4).forEach(date => push(date, '高能日', 'peak'))
  summary.low_days.slice(0, 4).forEach(date => push(date, '提醒日', 'low'))

  return cards.slice(0, 8)
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

const guideGroups = computed(() => {
  const days = monthDays.value
  return [
    { title: '幸运色', kind: 'color', items: topCounts(days.map(day => day.lucky_color), 4) },
    { title: '幸运数', kind: 'number', items: topCounts(days.map(day => day.lucky_number), 4) },
    { title: '财位', kind: 'direction', items: topCounts(days.map(day => day.wealth_direction), 4) },
    { title: '十神', kind: 'ten-god', items: topCounts(days.map(day => day.today_ten_god), 4) },
  ]
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
    } catch { /* ignore */ }
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
            <p class="advice">{{ data.summary.key_advice }}</p>
            <div class="chips">
              <BestWorstChip
                v-if="data.summary.best_day"
                variant="best"
                :date="data.summary.best_day"
                :score="data.summary.best_score"
              />
              <BestWorstChip
                v-if="data.summary.worst_day"
                variant="worst"
                :date="data.summary.worst_day"
                :score="data.summary.worst_score"
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
              :score="data.monthly_score"
              label="月度综合"
              :caption="`波动 ±${data.summary.volatility.toFixed(1)}`"
            />
          </div>
        </section>

        <section class="overview-grid">
          <article
            v-for="stat in overviewStats"
            :key="stat.label"
            class="overview-card"
            :class="`mood-${stat.mood}`"
          >
            <span class="overview-label">{{ stat.label }}</span>
            <strong class="overview-value tabular-nums">{{ stat.value }}</strong>
            <span class="overview-detail">{{ stat.detail }}</span>
          </article>
        </section>

        <section class="month-brief">
          <div class="brief-copy">
            <span class="card-eyebrow">月度总述</span>
            <h2>{{ monthBriefTitle }}</h2>
            <p>{{ monthBriefText || data.summary.key_advice }}</p>
            <div class="score-bands" aria-label="分数分布">
              <span class="band good">
                <strong class="tabular-nums">{{ scoreBands.good }}</strong>
                顺行
              </span>
              <span class="band steady">
                <strong class="tabular-nums">{{ scoreBands.steady }}</strong>
                平稳
              </span>
              <span class="band caution">
                <strong class="tabular-nums">{{ scoreBands.caution }}</strong>
                谨慎
              </span>
            </div>
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
            <span class="card-meta">按月内评分自动切段</span>
          </header>
          <div class="phase-grid">
            <article
              v-for="phase in phaseSegments"
              :key="phase.name"
              class="phase-card"
              :class="`mood-${phase.mood}`"
            >
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
                <span v-if="phase.best">高 {{ formatMonthDay(phase.best.solar_date) }} · {{ phase.best.score }}</span>
                <span v-if="phase.low">低 {{ formatMonthDay(phase.low.solar_date) }} · {{ phase.low.score }}</span>
                <span>{{ phase.driftLabel }}</span>
              </div>
              <p>{{ phase.advice }}</p>
            </article>
          </div>
        </section>

        <section class="glass-card trend-card">
          <header class="card-head">
            <span class="card-eyebrow">月历强度</span>
            <span class="card-meta">
              评分折线 · 峰 {{ data.summary.peak_days.length }} · 谷 {{ data.summary.low_days.length }}
            </span>
          </header>
          <FortuneChart :daily-data="trendData" height="320px" :show-elements="false" />
          <div class="trend-legend" aria-label="月历强度说明">
            <span><i class="legend-dot best"></i>吉峰 {{ formatMonthDay(data.summary.best_day) }}</span>
            <span><i class="legend-dot worst"></i>低谷 {{ formatMonthDay(data.summary.worst_day) }}</span>
            <span class="tabular-nums">连吉 {{ data.summary.good_streak }}d</span>
          </div>
        </section>

        <section class="key-days-card">
          <header class="card-head">
            <span class="card-eyebrow">关键日期</span>
            <span class="card-meta">吉峰、高能与低谷提醒</span>
          </header>
          <div class="key-day-grid">
            <article
              v-for="day in keyDayCards"
              :key="day.date"
              class="key-day"
              :class="day.variant"
            >
              <span class="key-date tabular-nums">{{ formatMonthDay(day.date) }}</span>
              <div class="key-main">
                <strong>{{ day.label }}</strong>
                <span>{{ day.pillar }} · {{ day.score }}分</span>
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
              <span class="card-eyebrow">五行走势</span>
              <span class="card-meta tabular-nums">均 {{ data.summary.average_score.toFixed(1) }}</span>
            </header>
            <FortuneChart :daily-data="trendData" height="280px" />
          </div>
        </section>

        <section class="guide-card">
          <header class="card-head">
            <span class="card-eyebrow">开运频次</span>
            <span class="card-meta">按每日幸运项统计</span>
          </header>
          <div class="guide-grid">
            <article v-for="group in guideGroups" :key="group.title" class="guide-group">
              <h3>{{ group.title }}</h3>
              <div class="guide-items">
                <span
                  v-for="item in group.items"
                  :key="`${group.title}-${item.label}`"
                  class="guide-chip"
                >
                  <span
                    v-if="group.kind === 'color'"
                    class="guide-swatch"
                    :style="{ background: colorSwatch(item.label) }"
                  ></span>
                  <strong>{{ item.label }}</strong>
                  <em class="tabular-nums">{{ item.count }}次</em>
                </span>
              </div>
            </article>
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
              :lucky-color="d.lucky_color"
              :lucky-number="d.lucky_number"
              :wealth-dir="d.wealth_direction"
              :yi-items="d.yi"
              :ji-items="d.ji"
              :today-ten-god="d.today_ten_god"
              :weekday="weekdayShort(d.solar_date)"
              :is-best="data!.summary.best_day === d.solar_date"
              :is-worst="data!.summary.worst_day === d.solar_date"
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
  display: flex; flex-direction: column; align-items: center; justify-content: center;
  min-height: 70vh; gap: 16px; color: var(--text-muted); font-size: 0.95rem;
}
.state.error p { color: var(--crimson); }
.btn-link { color: rgba(var(--jade-accent-rgb), 1); text-decoration: none; font-weight: 600; letter-spacing: 0.04em; }
.orb-skeleton {
  width: 100px; height: 100px; border-radius: 50%;
  background: radial-gradient(closest-side, rgba(var(--jade-accent-rgb), 0.45), transparent 70%);
  filter: blur(4px);
  animation: pulse 1.6s ease-in-out infinite;
}
@keyframes pulse { 0%,100% { opacity: 0.4 } 50% { opacity: 0.95 } }

/* hero (shared with weekly) */
.hero {
  display: grid; grid-template-columns: 1fr; gap: 24px; padding: 24px;
  border-radius: 28px;
  background: linear-gradient(135deg, color-mix(in oklab, var(--surface-1) 86%, transparent), color-mix(in oklab, var(--surface-2) 78%, transparent));
  border: 1px solid var(--line-strong);
  backdrop-filter: blur(22px) saturate(140%);
  box-shadow: var(--shadow-lg);
}
:global(.dark) .hero { background: linear-gradient(135deg, rgba(255,255,255,0.04), rgba(255,255,255,0.02)); }
@media (min-width: 900px) { .hero { grid-template-columns: 1.4fr 1fr; align-items: center; } }
.hero-left { display: flex; flex-direction: column; gap: 12px; min-width: 0; }
.hero-right { display: flex; align-items: center; justify-content: center; }

.eyebrow {
  font-family: var(--font-mono), monospace;
  font-size: 0.7rem; letter-spacing: 0.42em; color: var(--text-muted); text-transform: uppercase;
}
.title {
  font-family: var(--font-serif), 'Songti SC', serif;
  font-size: clamp(2rem, 4vw, 3rem); font-weight: 800; letter-spacing: 0.18em;
  margin: 0; color: var(--text);
}
.range { color: var(--text-muted); margin: 0; letter-spacing: 0.06em; font-size: 0.9rem; }
.advice {
  font-family: var(--font-serif), serif; font-size: 1rem; line-height: 1.75;
  color: var(--text); margin: 4px 0 6px; letter-spacing: 0.02em;
}
.chips { display: flex; gap: 8px; flex-wrap: wrap; }
.meta-chip {
  display: inline-flex; align-items: center; gap: 6px;
  padding: 4px 10px; border-radius: 999px;
  border: 1px solid var(--line-subtle); font-size: 0.72rem;
  color: var(--text-muted); background: var(--glass-bg);
}
.meta-chip .dot { width: 6px; height: 6px; border-radius: 50%; background: var(--jade-accent); box-shadow: 0 0 6px var(--jade-accent); }

/* glass card */
.glass-card {
  padding: 20px; border-radius: 22px;
  background: linear-gradient(135deg, color-mix(in oklab, var(--surface-1) 88%, transparent), color-mix(in oklab, var(--surface-2) 78%, transparent));
  border: 1px solid var(--line-strong);
  backdrop-filter: blur(22px) saturate(140%);
  box-shadow: var(--shadow-md);
}
:global(.dark) .glass-card { background: linear-gradient(135deg, rgba(255,255,255,0.04), rgba(255,255,255,0.02)); }

.card-head { display: flex; align-items: baseline; justify-content: space-between; gap: 12px; }
.card-head.plain { padding: 0 4px; }
.card-eyebrow {
  font-size: 0.72rem; letter-spacing: 0.32em; color: var(--text-muted);
  font-family: var(--font-mono), monospace; text-transform: uppercase;
}
.card-meta { font-size: 0.72rem; color: var(--text-soft); letter-spacing: 0.06em; }

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
  background: linear-gradient(135deg, var(--glass-bg), color-mix(in oklab, var(--surface-2) 70%, transparent));
  border: 1px solid var(--line-subtle);
  box-shadow: var(--shadow-sm);
}
.overview-card.mood-good { border-color: rgba(var(--jade-accent-rgb), 0.36); }
.overview-card.mood-steady { border-color: rgba(130, 145, 160, 0.24); }
.overview-card.mood-caution { border-color: rgba(232, 64, 87, 0.24); }
.overview-label,
.overview-detail,
.phase-range,
.phase-days,
.guide-group h3,
.guide-chip,
.key-main span,
.element-value {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.overview-label {
  font-size: 0.72rem;
  letter-spacing: 0.28em;
  color: var(--text-muted);
  text-transform: uppercase;
}
.overview-value {
  font-family: var(--font-serif), serif;
  font-size: clamp(1.5rem, 2vw, 2rem);
  line-height: 1;
  color: var(--text);
}
.overview-detail {
  font-size: 0.78rem;
  color: var(--text-soft);
}

.month-brief {
  display: grid;
  gap: 18px;
  padding: 22px;
  border-radius: 24px;
  background:
    linear-gradient(135deg, color-mix(in oklab, var(--surface-1) 84%, transparent), color-mix(in oklab, var(--surface-2) 76%, transparent)),
    radial-gradient(circle at top right, rgba(var(--jade-accent-rgb), 0.08), transparent 32%);
  border: 1px solid var(--line-strong);
  box-shadow: var(--shadow-md);
}
.brief-copy { display: flex; flex-direction: column; gap: 10px; min-width: 0; }
.brief-copy h2 {
  margin: 0;
  font-family: var(--font-serif), serif;
  font-size: clamp(1.2rem, 2vw, 1.55rem);
  letter-spacing: 0.08em;
}
.brief-copy p {
  margin: 0;
  line-height: 1.8;
  color: var(--text);
  max-width: 66ch;
}
.score-bands {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
.band {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 6px 12px;
  border-radius: 999px;
  border: 1px solid var(--line-subtle);
  background: var(--glass-bg);
  font-size: 0.72rem;
  color: var(--text-muted);
}
.band strong { font-size: 0.98rem; color: var(--text); }
.band.good { border-color: rgba(var(--jade-accent-rgb), 0.28); }
.band.steady { border-color: rgba(130, 145, 160, 0.22); }
.band.caution { border-color: rgba(232, 64, 87, 0.24); }

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
  font-size: 0.74rem;
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
.key-days-card,
.guide-card {
  display: flex;
  flex-direction: column;
  gap: 14px;
  padding: 20px;
  border-radius: 22px;
  background: linear-gradient(135deg, color-mix(in oklab, var(--surface-1) 88%, transparent), color-mix(in oklab, var(--surface-2) 78%, transparent));
  border: 1px solid var(--line-strong);
  backdrop-filter: blur(22px) saturate(140%);
  box-shadow: var(--shadow-md);
}

.phase-grid,
.key-day-grid,
.guide-grid {
  display: grid;
  gap: 12px;
}
.phase-grid { grid-template-columns: repeat(3, minmax(0, 1fr)); }
.key-day-grid { grid-template-columns: repeat(4, minmax(0, 1fr)); }
.guide-grid { grid-template-columns: repeat(4, minmax(0, 1fr)); }

.phase-card,
.key-day,
.guide-group {
  min-width: 0;
  border-radius: 18px;
  padding: 16px;
  background: var(--glass-bg);
  border: 1px solid var(--line-subtle);
}
.phase-card.mood-good,
.key-day.best,
.key-day.peak,
.guide-group {
  border-color: rgba(var(--jade-accent-rgb), 0.20);
}
.phase-card.mood-caution,
.key-day.worst,
.key-day.low {
  border-color: rgba(232, 64, 87, 0.22);
}
.phase-top,
.key-main {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 12px;
}
.phase-name,
.key-main strong,
.guide-group h3 {
  font-size: 0.9rem;
  color: var(--text);
}
.phase-range,
.key-main span,
.phase-days,
.guide-chip em {
  font-size: 0.72rem;
  color: var(--text-soft);
}
.phase-score {
  font-family: var(--font-serif), serif;
  font-size: 1.4rem;
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
  background: linear-gradient(90deg, rgba(var(--jade-accent-rgb), 0.55), rgba(var(--jade-accent-rgb), 1));
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
  font-size: 0.82rem;
}

.key-day {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.key-date {
  font-size: 0.72rem;
  letter-spacing: 0.18em;
  color: var(--text-soft);
  text-transform: uppercase;
}
.key-main {
  flex-direction: column;
  align-items: flex-start;
  gap: 2px;
}

.guide-group { display: flex; flex-direction: column; gap: 10px; }
.guide-group h3 {
  margin: 0;
  font-size: 0.82rem;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: var(--text-muted);
}
.guide-items {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
.guide-chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  max-width: 100%;
  padding: 6px 10px;
  border-radius: 999px;
  border: 1px solid var(--line-subtle);
  background: rgba(var(--jade-accent-rgb), 0.06);
  font-size: 0.74rem;
}
.guide-chip strong { min-width: 0; }
.guide-chip em { font-style: normal; }
.guide-swatch {
  width: 10px;
  height: 10px;
  border-radius: 999px;
  border: 1px solid var(--line-strong);
  flex: 0 0 auto;
}

.overview-grid,
.month-brief,
.rhythm-card,
.key-days-card,
.guide-card,
.grid-2,
.day-section {
  min-width: 0;
}

.trend-card { display: flex; flex-direction: column; gap: 12px; }
.trend-legend {
  display: flex;
  flex-wrap: wrap;
  gap: 10px 16px;
  align-items: center;
  padding: 0 4px;
  color: var(--text-muted);
  font-size: 0.72rem;
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
.legend-dot.best {
  background: var(--jade-accent);
  box-shadow: 0 0 8px rgba(var(--jade-accent-rgb), 0.75);
}
.legend-dot.worst {
  background: var(--crimson);
  box-shadow: 0 0 8px rgba(232, 64, 87, 0.55);
}

.grid-2 { display: grid; gap: 24px; grid-template-columns: 1fr; }
@media (min-width: 900px) { .grid-2 { grid-template-columns: 1fr 1fr; } }

.day-section { display: flex; flex-direction: column; gap: 14px; }
.day-grid {
  display: grid; gap: 12px;
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
  font-size: 0.74rem;
  letter-spacing: 0.06em;
  cursor: pointer;
  transition: border-color 0.2s ease, color 0.2s ease;
}
.toggle:hover { border-color: rgba(var(--jade-accent-rgb), 0.55); color: rgba(var(--jade-accent-rgb), 1); }

.footer-nav {
  display: flex; gap: 14px; justify-content: center; padding: 8px 0;
  font-size: 0.85rem; color: var(--text-muted);
}
.footer-nav a {
  color: rgba(var(--jade-accent-rgb), 1);
  text-decoration: none; font-weight: 500; letter-spacing: 0.04em;
}
.footer-nav a:hover { text-shadow: 0 0 12px rgba(var(--jade-accent-rgb), 0.55); }

.tabular-nums { font-variant-numeric: tabular-nums; }

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
  .key-day-grid,
  .guide-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 760px) {
  .page { gap: 18px; }
  .hero,
  .glass-card,
  .rhythm-card,
  .key-days-card,
  .guide-card,
  .month-brief {
    padding: 16px;
    border-radius: 20px;
  }
  .overview-grid {
    grid-template-columns: 1fr;
  }
  .phase-grid,
  .key-day-grid,
  .guide-grid {
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
