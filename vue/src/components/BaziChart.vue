<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { use } from 'echarts/core'
import { BarChart } from 'echarts/charts'
import { GridComponent, TooltipComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import VChart from 'vue-echarts'
import ClassicalInterpretationPanel from './ClassicalInterpretationPanel.vue'
import type { BodyStrengthResult, FortuneLayer, FortuneLayerSet, RuleMeta, ShenShaMeta } from '@/api/chart'

type PillarKey = 'year' | 'month' | 'day' | 'hour'
type PillarShenShaGroup = {
  pillar: PillarKey
  label: string
  gan: string
  zhi: string
  priority: number
  role: string
  items: string[]
  details?: ShenShaMeta[]
}
type DayunFlowItem = {
  start_age?: number
  pillar?: string
  flow_change?: string
  impact?: string
}
type DayunStageView = {
  index: number
  gan: string
  zhi: string
  pillar: string
  startAge: number
  endAge: number
  startYear: number | null
  endYear: number | null
  ganElement: string
  zhiElement: string
  tenGod: string
  flowChange: string
  impact: string
  isCurrent: boolean
}

const props = defineProps<{
  chart: {
    id?: number
    year_pillar: { gan: string; zhi: string }
    month_pillar: { gan: string; zhi: string }
    day_pillar: { gan: string; zhi: string }
    hour_pillar: { gan: string; zhi: string }
    rule_meta?: RuleMeta
    body_strength?: BodyStrengthResult
    day_shen_sha_details?: ShenShaMeta[]
    global_shen_sha_details?: ShenShaMeta[]
    shen_sha_by_pillar?: PillarShenShaGroup[]
    fortune_layers?: FortuneLayerSet
    [key: string]: any
  }
}>()

const themeVersion = ref(0)
let themeObserver: MutationObserver | null = null

function cssVar(name: string, fallback: string): string {
  if (typeof window === 'undefined') return fallback
  return getComputedStyle(document.documentElement).getPropertyValue(name).trim() || fallback
}

onMounted(() => {
  themeObserver = new MutationObserver(() => {
    themeVersion.value += 1
  })
  themeObserver.observe(document.documentElement, { attributes: true, attributeFilter: ['class'] })
})

onUnmounted(() => {
  themeObserver?.disconnect()
  themeObserver = null
})

const ganElement: Record<string, { name: string; elemColor: string }> = {
  甲: { name: '木', elemColor: '#34d399' },
  乙: { name: '木', elemColor: '#34d399' },
  丙: { name: '火', elemColor: '#fb7185' },
  丁: { name: '火', elemColor: '#fb7185' },
  戊: { name: '土', elemColor: '#c76f12' },
  己: { name: '土', elemColor: '#c76f12' },
  庚: { name: '金', elemColor: '#cbd5e1' },
  辛: { name: '金', elemColor: '#cbd5e1' },
  壬: { name: '水', elemColor: '#22d3ee' },
  癸: { name: '水', elemColor: '#22d3ee' },
}

const zhiElement: Record<string, { name: string; elemColor: string }> = {
  寅: { name: '木', elemColor: '#34d399' },
  卯: { name: '木', elemColor: '#34d399' },
  巳: { name: '火', elemColor: '#fb7185' },
  午: { name: '火', elemColor: '#fb7185' },
  辰: { name: '土', elemColor: '#c76f12' },
  戌: { name: '土', elemColor: '#c76f12' },
  丑: { name: '土', elemColor: '#c76f12' },
  未: { name: '土', elemColor: '#c76f12' },
  申: { name: '金', elemColor: '#cbd5e1' },
  酉: { name: '金', elemColor: '#cbd5e1' },
  亥: { name: '水', elemColor: '#22d3ee' },
  子: { name: '水', elemColor: '#22d3ee' },
}

const pillars = computed(() => [
  { label: '年柱', key: 'year' as const, idx: 0, ...props.chart.year_pillar },
  { label: '月柱', key: 'month' as const, idx: 1, ...props.chart.month_pillar },
  { label: '日柱', key: 'day' as const, idx: 2, ...props.chart.day_pillar },
  { label: '时柱', key: 'hour' as const, idx: 3, ...props.chart.hour_pillar },
])

const daYun = computed(() => props.chart.da_yun || props.chart.da_yun_start)

// --- 天干地支分析（从 API 数据读取）---
const ganZhi = computed(() => props.chart.gan_zhi_analysis)

function ganRelClass(type: string): string {
  if (type === '五合') return 'rel-he'
  if (type === '相克') return 'rel-ke'
  if (type === '相生') return 'rel-sheng'
  return ''
}

function zhiRelClass(type: string): string {
  if (type === '六冲') return 'rel-chong'
  if (type === '六合') return 'rel-he'
  if (type === '六害') return 'rel-hai'
  if (type === '相刑') return 'rel-xing'
  if (type === '三会') return 'rel-hui'
  return ''
}

function ganRelSymbol(type: string): string {
  if (type === '五合') return '合'
  if (type === '相克') return '克'
  return '生'
}

function zhiRelSymbol(type: string): string {
  if (type === '六冲') return '冲'
  if (type === '六合') return '合'
  if (type === '六害') return '害'
  if (type === '相刑') return '刑'
  return '会'
}

function relationSummary(detail: string): string {
  return String(detail || '').split('\n')[0] || ''
}

const elemColor = (e: string) => {
  themeVersion.value
  const isDark = typeof document !== 'undefined' && document.documentElement.classList.contains('dark')
  const lightMap: Record<string, string> = {
    金: '#cbd5e1',
    木: '#34d399',
    水: '#22d3ee',
    火: '#fb7185',
    土: '#c76f12',
  }
  const darkMap: Record<string, string> = {
    金: '#cbd5e1',
    木: '#34d399',
    水: '#22d3ee',
    火: '#fb7185',
    土: '#f59e0b',
  }
  return (isDark ? darkMap : lightMap)[e] || '#8a9a8e'
}

const pillarLabel = (k: string) => ({ year: '年柱', month: '月柱', day: '日柱', hour: '时柱' }[k] || k)

const ganPolarity: Record<string, 'yang' | 'yin'> = {
  甲: 'yang', 丙: 'yang', 戊: 'yang', 庚: 'yang', 壬: 'yang',
  乙: 'yin', 丁: 'yin', 己: 'yin', 辛: 'yin', 癸: 'yin',
}

const elementCycle = ['木', '火', '土', '金', '水']

function produces(source: string, target: string): boolean {
  const idx = elementCycle.indexOf(source)
  return idx >= 0 && elementCycle[(idx + 1) % elementCycle.length] === target
}

function controls(source: string, target: string): boolean {
  const idx = elementCycle.indexOf(source)
  return idx >= 0 && elementCycle[(idx + 2) % elementCycle.length] === target
}

function tenGodFor(dayGan?: string, targetGan?: string): string {
  if (!dayGan || !targetGan) return '未判'
  const dayElem = ganElement[dayGan]?.name
  const targetElem = ganElement[targetGan]?.name
  if (!dayElem || !targetElem) return '未判'
  const samePolarity = ganPolarity[dayGan] === ganPolarity[targetGan]

  if (targetElem === dayElem) return samePolarity ? '比肩' : '劫财'
  if (produces(targetElem, dayElem)) return samePolarity ? '偏印' : '正印'
  if (produces(dayElem, targetElem)) return samePolarity ? '食神' : '伤官'
  if (controls(dayElem, targetElem)) return samePolarity ? '偏财' : '正财'
  if (controls(targetElem, dayElem)) return samePolarity ? '七杀' : '正官'
  return '未判'
}

const dayunFlowItems = computed<DayunFlowItem[]>(() => Array.isArray(props.chart.dayun_flow) ? props.chart.dayun_flow : [])

const currentAge = computed(() => {
  const birthYear = Number(props.chart.birth_year || 0)
  if (!birthYear) return null
  const now = new Date()
  const birthMonth = Number(props.chart.birth_month || 0)
  const birthDay = Number(props.chart.birth_day || 0)
  let age = now.getFullYear() - birthYear
  if (birthMonth && birthDay) {
    const monthPassed = now.getMonth() + 1 > birthMonth
    const dayPassed = now.getMonth() + 1 === birthMonth && now.getDate() >= birthDay
    if (!monthPassed && !dayPassed) age -= 1
  }
  return Math.max(0, age)
})

const dayunStages = computed<DayunStageView[]>(() => {
  const info = daYun.value || {}
  const pillarsRaw: Array<{ gan?: string; zhi?: string }> = Array.isArray(info.pillars) ? info.pillars : []
  const flowRaw = dayunFlowItems.value
  const source = pillarsRaw.length ? pillarsRaw : flowRaw.map(item => {
    const pillar = String(item.pillar || '')
    return { gan: pillar.slice(0, 1), zhi: pillar.slice(1, 2) }
  })
  const startAgeBase = Number(info.start_age || flowRaw[0]?.start_age || 0)
  const birthYear = Number(props.chart.birth_year || 0)
  const age = currentAge.value

  return source.map((p, index) => {
    const gan = String(p.gan || '')
    const zhi = String(p.zhi || '')
    const pillar = gan + zhi
    const flow = flowRaw.find(item => item.pillar === pillar) || flowRaw[index]
    const startAge = Number(flow?.start_age || startAgeBase + index * 10)
    const endAge = startAge + 9
    const startYear = birthYear ? birthYear + startAge : null
    const endYear = startYear ? startYear + 9 : null

    return {
      index,
      gan,
      zhi,
      pillar,
      startAge,
      endAge,
      startYear,
      endYear,
      ganElement: ganElement[gan]?.name || '',
      zhiElement: zhiElement[zhi]?.name || '',
      tenGod: tenGodFor(props.chart.day_pillar?.gan, gan),
      flowChange: flow?.flow_change || '未判',
      impact: flow?.impact || '暂无流通判断',
      isCurrent: age !== null && age >= startAge && age <= endAge,
    }
  })
})

const hasDaYun = computed(() => dayunStages.value.length > 0)
const currentDayunStage = computed(() => dayunStages.value.find(stage => stage.isCurrent) || null)
const currentDayunLayer = computed(() => props.chart.fortune_layers?.dayun || null)

function parseShenSha(raw: string) {
  const [head, desc = ''] = raw.split('｜')
  const colonIndex = head.indexOf('：')
  if (colonIndex === -1) return { name: head, target: '', desc }

  const name = head.slice(0, colonIndex)
  const tail = head.slice(colonIndex + 1)
  const target = desc ? tail : ''
  return { name, target, desc: desc || tail }
}

const parsedDayShenSha = computed(() => (props.chart.day_shen_sha || []).map(parseShenSha))

const groupedShenSha = computed(() => {
  const raw = props.chart.shen_sha_by_pillar
  if (!raw || !raw.length) return null
  return raw.map((g: any) => ({ ...g, items: (g.items || []).map(parseShenSha) }))
})

const globalShenSha = computed(() => (props.chart.global_shen_sha || []).map(parseShenSha))

const showSummary = computed(() => !!props.chart.shen_sha_summary)

const ruleMeta = computed(() => props.chart.rule_meta || null)
const ruleTablePreview = computed(() => ruleMeta.value?.tables || [])

const bodyStrengthComponents = computed(() => props.chart.body_strength?.components || [])
const bodyStrengthEvidence = computed(() => (props.chart.body_strength?.evidence || []).slice(0, 4))
const bodyScorePercent = computed(() => {
  const score = Number(props.chart.body_strength?.total_score || 0)
  return Math.max(0, Math.min(100, Math.round(score * 100)))
})

const shenShaDetails = computed(() => {
  const seen = new Set<string>()
  const details: ShenShaMeta[] = []
  const add = (items?: ShenShaMeta[]) => {
    for (const item of items || []) {
      if (!item?.name || seen.has(item.name)) continue
      seen.add(item.name)
      details.push(item)
    }
  }
  add(props.chart.day_shen_sha_details)
  add(props.chart.global_shen_sha_details)
  for (const group of props.chart.shen_sha_by_pillar || []) add(group.details)
  return details.sort((a, b) => Number(a.priority || 9) - Number(b.priority || 9)).slice(0, 8)
})

const shenShaCategorySummary = computed(() => {
  const counts = new Map<string, number>()
  for (const detail of shenShaDetails.value) {
    const key = detail.category || '提示'
    counts.set(key, (counts.get(key) || 0) + 1)
  }
  return Array.from(counts.entries()).map(([category, count]) => ({ category, count }))
})

const fortuneLayerList = computed(() => {
  const layers = props.chart.fortune_layers
  if (!layers) return []
  return (['liunian', 'liuyue', 'xiaoyun'] as const)
    .map(key => layers[key])
    .filter((layer): layer is FortuneLayer => Boolean(layer))
})

function polarityClass(polarity?: string): string {
  if (polarity === '吉') return 'sha-good'
  if (polarity === '凶') return 'sha-risk'
  return 'sha-neutral'
}

function componentWidth(score: number): string {
  return `${Math.max(4, Math.min(100, Math.round(Number(score || 0) * 100)))}%`
}

const pillarShenShaColor = (p: string) =>
  ({ day: 'var(--accent)', year: '#5BA4CF', month: '#60B89A', hour: '#A182CF' }[p] || '#888')

const pillarShenShaBg = (p: string) =>
  ({ day: 'var(--accent-dim)', year: 'rgba(91,164,207,0.06)', month: 'rgba(96,184,154,0.06)', hour: 'rgba(161,130,207,0.06)' }[p] || 'var(--accent-dim)')

function strengthLevel(total: number): string {
  if (total <= 0) return 'none'
  if (total <= 5) return 'weak'
  if (total <= 15) return 'medium'
  if (total <= 25) return 'strong'
  return 'very-strong'
}

// MingGong shensha classification
const jiShenSha = new Set(['天福', '天贵', '天权', '天印', '天艺', '天寿'])
const xiongShenSha = new Set(['天刃', '天破', '天奸', '天孤', '天刑', '天囚'])

function isJiShenSha(name: string): boolean {
  return jiShenSha.has(name)
}

function isXiongShenSha(name: string): boolean {
  return xiongShenSha.has(name)
}

const fiveElementsOption = computed(() => {
  themeVersion.value
  const textColor = cssVar('--text', '#0f1712')
  const softColor = cssVar('--text-soft', 'rgba(15, 23, 18, 0.44)')
  const lineColor = cssVar('--line-subtle', 'rgba(15, 23, 18, 0.06)')
  const tooltipBg = cssVar('--surface-1', '#ffffff')
  const fe = props.chart.five_elements
  if (!fe) return null
  const total = Object.values(fe as Record<string, number>).reduce((s, v) => s + v, 0)
  if (total === 0) return null

  // 五行配色 — 科技色盘
  const barColors = ['#34d399', '#fb7185', '#c76f12', '#cbd5e1', '#22d3ee']
  const labels = ['木', '火', '土', '金', '水']

  return {
    backgroundColor: 'transparent',
    grid: { left: 8, right: 8, bottom: 28, top: 12, containLabel: true },
    xAxis: {
      type: 'category',
      data: labels,
      axisLine: { lineStyle: { color: lineColor } },
      axisTick: { show: false },
      axisLabel: {
        color: textColor,
        fontSize: 12,
        fontWeight: '600',
        fontFamily: 'Noto Serif SC, Songti SC, serif',
      },
      splitLine: { show: false },
    },
    yAxis: {
      type: 'value',
      max: 30,
      axisLabel: { color: softColor, fontSize: 10, formatter: '{value}' },
      splitLine: { lineStyle: { color: lineColor, type: 'dashed' } },
      axisLine: { show: false },
      axisTick: { show: false },
    },
    series: [{
      type: 'bar',
      data: labels.map((l, i) => ({
        value: fe[l] || 0,
        itemStyle: {
          color: {
            type: 'linear', x: 0, y: 0, x2: 0, y2: 1,
            colorStops: [
              { offset: 0, color: barColors[i] },
              { offset: 1, color: barColors[i] + '66' },
            ],
          },
          borderRadius: [4, 4, 0, 0],
        },
        emphasis: {
          itemStyle: {
            color: {
              type: 'linear', x: 0, y: 0, x2: 0, y2: 1,
              colorStops: [
                { offset: 0, color: barColors[i] + 'EE' },
                { offset: 1, color: barColors[i] + '99' },
              ],
            },
          },
        },
      })),
      barMaxWidth: 38,
      barCategoryGap: '30%',
      label: {
        show: true,
        position: 'top',
        formatter: '{c}',
        fontSize: 11,
        fontWeight: '600',
        color: textColor,
        fontFamily: 'DM Mono, Fira Code, monospace',
      },
    }],
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'shadow' },
      backgroundColor: tooltipBg,
      borderColor: lineColor,
      borderWidth: 1,
      padding: [8, 14],
      textStyle: {
        color: textColor,
        fontSize: 13,
        fontFamily: 'Noto Serif SC, Songti SC, serif',
        fontWeight: '600',
      },
      formatter: (params: any[]) => {
        const p = params[0]
        return `<span style="color:${p.color};font-weight:700">${p.name}</span>：<span style="color:${textColor};font-weight:700">${p.value}</span> 分`
      },
    },
    animationDuration: 900,
    animationEasing: 'cubicOut' as const,
  }
})

const pillarDetails = computed(() => props.chart.pillar_details || [])

const birthMonthLabel = computed(() => {
  const m = props.chart.birth_month
  if (!m) return ''
  const labels: Record<number, string> = { 5: '五月', 6: '六月' }
  return labels[m] || ''
})

// tiaohouElem returns the element of the primary tiaohou god for color coding
const tiaohouElem = computed(() => {
  const g = props.chart.tiaohou?.primary_god
  const map: Record<string, string> = { '甲': '木', '乙': '木', '丙': '火', '丁': '火', '戊': '土', '己': '土', '庚': '金', '辛': '金', '壬': '水', '癸': '水' }
  return g ? (map[g] || '金') : '金'
})

use([BarChart, GridComponent, TooltipComponent, CanvasRenderer])

// Tab navigation
const activeTab = ref('overview')
const chartTabs = computed(() => {
  const tabs = [
    { key: 'overview', label: '命盘总览' },
    ...(hasDaYun.value ? [{ key: 'dayun', label: '大运' }] : []),
    { key: 'wuxing', label: '五行格局' },
    { key: 'shishen', label: '十神详解' },
    { key: 'pattern', label: '格局古籍' },
    { key: 'shensha', label: '神煞' },
    { key: 'fortune', label: '运势详批' },
  ]
  if (ruleMeta.value || props.chart.id) tabs.push({ key: 'rules', label: '规则依据' })
  return tabs
})

const tenGodChartOptions = computed(() => {
  themeVersion.value
  const textColor = cssVar('--text', '#0f1712')
  const mutedColor = cssVar('--text-muted', '#5a6a5e')
  const softColor = cssVar('--text-soft', 'rgba(15, 23, 18, 0.44)')
  const lineColor = cssVar('--line-subtle', 'rgba(15, 23, 18, 0.06)')
  const tooltipBg = cssVar('--surface-1', '#ffffff')
  const data = props.chart.ten_god_proportion || []
  // 10 ten gods — cold-tone tech palette
  const barColors = [
    '#cbd5e1', // 比肩 - chrome silver
    '#94a3b8', // 劫财 - slate
    '#9B72CF', // 食神 - purple
    '#C85FCF', // 伤官 - magenta
    '#60a5fa', // 正财 - blue
    '#3b82f6', // 偏财 - deep blue
    '#34d399', // 正官 - emerald
    '#059669', // 七杀 - dark emerald
    '#fb7185', // 正印 - rose
    '#e11d48', // 偏印 - deep rose
  ]
  return {
    backgroundColor: 'transparent',
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'none' },
      formatter: (params: any[]) => {
        const p = params[0]
        return `<span style="color:${p.color};font-weight:700">${p.name}</span>：${p.value}%`
      },
      backgroundColor: tooltipBg,
      borderColor: lineColor,
      borderWidth: 1,
      padding: [6, 10],
      textStyle: { color: textColor, fontSize: 12 },
    },
    grid: {
      left: 8,
      right: 8,
      bottom: data.length > 0 ? 32 : 8,
      top: data.length > 0 ? 20 : 8,
      containLabel: true,
    },
    xAxis: {
      type: 'category',
      data: data.map((d: any) => d.name),
      axisLine: { lineStyle: { color: lineColor } },
      axisTick: { show: false },
      axisLabel: {
        color: mutedColor,
        fontSize: 10,
        fontWeight: '500',
        interval: 0,
        rotate: data.length > 6 ? 30 : 0,
      },
      splitLine: { show: false },
    },
    yAxis: {
      type: 'value',
      max: 100,
      axisLabel: {
        formatter: '{value}%',
        color: softColor,
        fontSize: 9,
      },
      splitLine: {
        lineStyle: { color: lineColor, type: 'dashed' },
      },
      axisLine: { show: false },
      axisTick: { show: false },
    },
    series: [
      {
        type: 'bar',
        data: data.map((d: any, i: number) => ({
          value: d.percent,
          itemStyle: {
            color: {
              type: 'linear',
              x: 0, y: 0, x2: 0, y2: 1,
              colorStops: [
                { offset: 0, color: barColors[i % barColors.length] },
                { offset: 1, color: barColors[i % barColors.length] + '88' },
              ],
            },
            borderRadius: [6, 6, 0, 0],
            shadowBlur: 12,
            shadowColor: barColors[i % barColors.length] + '55',
          },
        })),
        barMaxWidth: data.length > 6 ? 18 : 28,
        barGap: '6px',
        label: {
          show: true,
          position: 'top',
          formatter: '{c}%',
          fontSize: 10,
          fontWeight: '600',
          color: mutedColor,
          distance: 6,
        },
        emphasis: {
          itemStyle: {
            shadowBlur: 20,
            shadowColor: '#cbd5e166',
          },
        },
      },
    ],
    animationDuration: 1000,
    animationEasing: 'cubicOut',
    animationDelay: (idx: number) => idx * 80,
  }
})
</script>

<template>
  <div class="bazi-chart">
    <!-- Constellation decoration -->
    <div class="chart-bg" aria-hidden="true">
      <svg viewBox="0 0 600 200" preserveAspectRatio="xMidYMid slice" class="bg-svg">
        <circle cx="50" cy="30" r="1" fill="currentColor" opacity="0.2" />
        <circle cx="550" cy="40" r="1.2" fill="currentColor" opacity="0.25" />
        <circle cx="300" cy="100" r="1.5" fill="currentColor" opacity="0.15" />
        <line
          x1="50"
          y1="30"
          x2="300"
          y2="100"
          stroke="currentColor"
          stroke-width="0.3"
          opacity="0.05"
        />
        <line
          x1="550"
          y1="40"
          x2="300"
          y2="100"
          stroke="currentColor"
          stroke-width="0.3"
          opacity="0.05"
        />
      </svg>
    </div>

    <div class="chart-card glass-card overflow-hidden">
      <!-- Title -->
      <div class="chart-header">
        <div class="header-eyebrow">BaZi Fortune</div>
        <h2 class="chart-title">八字命盘</h2>
      </div>

      <!-- Four-pillar axis with centered five elements -->
      <div class="pillars-bento">
        <div
          v-for="(pillar, pi) in pillars"
          :key="pillar.key"
          class="bento-card bento-small"
          :class="[
            'bento-hover-' + ganElement[pillar.gan]?.name,
            { 'bento-day-card': pillar.key === 'day' },
          ]"
          :style="{ order: pillar.idx >= 2 ? pillar.idx + 2 : pillar.idx + 1 }"
        >
          <div class="bento-label">{{ pillar.label }}</div>
          <div class="bento-body">
            <div class="bento-gan" :style="{ color: ganElement[pillar.gan]?.elemColor }">
              <span class="bento-char">{{ pillar.gan }}</span>
              <span
                class="elem-tag"
                :style="{
                  background: ganElement[pillar.gan]?.elemColor + '22',
                  color: ganElement[pillar.gan]?.elemColor,
                  borderColor: ganElement[pillar.gan]?.elemColor + '44',
                }"
              >{{ ganElement[pillar.gan]?.name }}</span>
            </div>
            <div class="bento-zhi" :style="{ color: zhiElement[pillar.zhi]?.elemColor }">
              <span class="bento-char">{{ pillar.zhi }}</span>
              <span
                class="elem-tag"
                :style="{
                  background: zhiElement[pillar.zhi]?.elemColor + '22',
                  color: zhiElement[pillar.zhi]?.elemColor,
                  borderColor: zhiElement[pillar.zhi]?.elemColor + '44',
                }"
              >{{ zhiElement[pillar.zhi]?.name }}</span>
            </div>
            <span v-if="chart.ten_gods?.[pillar.key]" class="bento-god-tag">{{ chart.ten_gods[pillar.key] }}</span>
          </div>
          <div v-if="pillarDetails[pi]" class="bento-sub">
            <span class="sheng-xiao-tag">{{ pillarDetails[pi].sheng_xiao }}</span>
            <span v-if="pillarDetails[pi].empties[0]" class="empties-tag">
              空{{ pillarDetails[pi].empties[0] }}{{ pillarDetails[pi].empties[1] }}
            </span>
          </div>
        </div>

        <div v-if="fiveElementsOption" class="bento-card bento-radar">
          <div class="bento-label">五行分布</div>
          <v-chart class="bento-radar-chart" :option="fiveElementsOption" autoresize />
        </div>
      </div>

      <!-- Tab navigation -->
      <div class="chart-tabs">
        <button
          v-for="tab in chartTabs"
          :key="tab.key"
          class="tab-btn"
          :class="{ active: activeTab === tab.key }"
          @click="activeTab = tab.key"
        >{{ tab.label }}</button>
      </div>

      <!-- Analysis sections -->
      <div class="analysis-section">

        <!-- ═══ Tab: 命盘总览 (overview) ═══ -->
        <div v-show="activeTab === 'overview'" class="tab-content overview-layout">
          <!-- Five Elements chart moved to Bento Grid -->

          <!-- Ten Gods -->
          <section v-if="chart.ten_gods" class="analysis-block overview-section overview-ten-gods">
            <div class="overview-section-head">
              <div>
                <div class="block-title">十神</div>
                <span class="block-desc">天干对日主的生克关系，反映人际、性格与命运倾向</span>
              </div>
            </div>
            <div class="overview-section-body">
              <div class="ten-gods-grid">
                <div v-for="(god, pillar) in chart.ten_gods" :key="pillar" class="god-item">
                  <span class="god-pillar">{{ pillar }}</span>
                  <span class="god-name">{{ god }}</span>
                </div>
              </div>
            </div>
          </section>

          <!-- NaYin -->
          <section v-if="chart.na_yin" class="analysis-block overview-section overview-nayin">
            <div class="overview-section-head">
              <div>
                <div class="block-title">纳音</div>
                <span class="block-desc">六十甲子纳音五行取象，揭示命局的先天气质与能量场</span>
              </div>
            </div>
            <div class="overview-section-body">
              <div class="nayin-list">
                <el-popover
                  v-for="(info, key) in chart.na_yin"
                  :key="key"
                  placement="bottom"
                  :width="300"
                  trigger="click"
                  popper-class="nayin-popover"
                >
                  <template #reference>
                    <span class="nayin-tag" :style="{ borderColor: elemColor(info.element) }">
                      <span class="nayin-pillar">{{ pillarLabel(String(key)) }}</span>
                      <span class="nayin-name" :style="{ color: elemColor(info.element) }">{{ info.name }}</span>
                    </span>
                  </template>
                  <div class="nayin-detail">
                    <div class="nayin-detail-header">
                      <span class="nayin-detail-name">{{ info.name }}</span>
                      <span class="nayin-detail-elem" :style="{ background: elemColor(info.element) }">{{ info.element }}</span>
                    </div>
                    <div class="nayin-detail-section">
                      <div class="nayin-detail-label">取象释义</div>
                      <div class="nayin-detail-value">{{ info.image_desc }}</div>
                    </div>
                    <div class="nayin-detail-section">
                      <div class="nayin-detail-label">性格命运</div>
                      <div class="nayin-detail-value">{{ info.personality }}</div>
                    </div>
                    <div class="nayin-detail-section">
                      <div class="nayin-detail-label">能量阶段</div>
                      <div class="nayin-detail-value nayin-energy">{{ info.energy_stage }}</div>
                    </div>
                    <div class="nayin-detail-section">
                      <div class="nayin-detail-label">现代延伸</div>
                      <div class="nayin-detail-value">{{ info.modern_ext }}</div>
                    </div>
                    <div v-if="info.judgments && info.judgments.length" class="nayin-detail-section">
                      <div class="nayin-detail-label">特质断语</div>
                      <div class="nayin-detail-tags">
                        <span v-for="j in info.judgments" :key="j" class="nayin-judgment-tag">{{ j }}</span>
                      </div>
                    </div>
                  </div>
                </el-popover>
              </div>
            </div>
          </section>

          <section v-if="ganZhi" class="analysis-block overview-section overview-relations">
            <div class="overview-section-head">
              <div>
                <div class="block-title">天干地支关系</div>
                <span class="block-desc">先看命盘总览，再看天干五合、生克与地支冲合刑害，便于把关系放回四柱结构中理解</span>
              </div>
            </div>
            <div class="overview-relations-grid">
              <div class="relation-card">
                <div v-if="ganZhi.gan_relations?.length > 0" class="relations-section">
                  <div class="relations-title">
                    <span class="relations-title-dot"></span>
                    天干关系
                    <span class="relations-count">{{ ganZhi.gan_relations.length }}组</span>
                  </div>
                  <div class="ganzhi-compact">
                    <div v-for="(rel, ri) in ganZhi.gan_relations" :key="'g'+ri" class="gz-item" :class="ganRelClass(rel.type)">
                      <span class="gz-chars">
                        <span class="gz-c">{{ rel.pillar1 }}</span>
                        <span class="gz-sym" :class="'sym-' + ganRelClass(rel.type)">{{ ganRelSymbol(rel.type) }}</span>
                        <span class="gz-c">{{ rel.pillar2 }}</span>
                      </span>
                      <span class="gz-tag" :class="'tag-' + ganRelClass(rel.type)">{{ rel.type }}</span>
                      <span class="gz-text">{{ relationSummary(rel.detail) }}</span>
                    </div>
                  </div>
                </div>
                <div v-else class="no-relations">
                  <span class="no-rel-icon">◇</span> 天干无特殊关系
                </div>
              </div>

              <div class="relation-card">
                <div v-if="ganZhi.zhi_relations?.length > 0" class="relations-section">
                  <div class="relations-title">
                    <span class="relations-title-dot zhi-dot"></span>
                    地支关系
                    <span class="relations-count">{{ ganZhi.zhi_relations.length }}组</span>
                  </div>
                  <div class="ganzhi-compact">
                    <div v-for="(rel, ri) in ganZhi.zhi_relations" :key="'z'+ri" class="gz-item" :class="zhiRelClass(rel.type)">
                      <span class="gz-chars">
                        <span class="gz-c">{{ rel.pillar1 }}</span>
                        <span class="gz-sym" :class="'sym-' + zhiRelClass(rel.type)">{{ zhiRelSymbol(rel.type) }}</span>
                        <span class="gz-c">{{ rel.pillar2 }}</span>
                      </span>
                      <span class="gz-tag" :class="'tag-' + zhiRelClass(rel.type)">{{ rel.type }}</span>
                      <span class="gz-text">{{ relationSummary(rel.detail) }}</span>
                    </div>
                  </div>
                </div>
                <div v-else class="no-relations">
                  <span class="no-rel-icon">◇</span> 地支无特殊关系
                </div>
              </div>
            </div>
          </section>

        </div><!-- /overview tab -->

        <!-- ═══ Tab: 大运 (dayun) ═══ -->
        <div v-show="activeTab === 'dayun'" class="tab-content dayun-detail-tab">
          <div v-if="hasDaYun" class="analysis-block dayun-overview-card">
            <div class="block-title">大运总览</div>
            <span class="block-desc">大运以十年为一阶段，重点看起运年龄、顺逆方向、干支十神与命局五行的增减关系</span>
            <div class="dayun-summary-grid">
              <div class="dayun-summary-item">
                <span class="dayun-summary-label">起运年龄</span>
                <strong>{{ daYun?.start_age || dayunStages[0]?.startAge }}岁</strong>
              </div>
              <div class="dayun-summary-item">
                <span class="dayun-summary-label">排运方向</span>
                <strong>{{ daYun?.direction || '未判' }}</strong>
              </div>
              <div class="dayun-summary-item">
                <span class="dayun-summary-label">当前年龄</span>
                <strong>{{ currentAge !== null ? `${currentAge}岁` : '未提供' }}</strong>
              </div>
              <div class="dayun-summary-item">
                <span class="dayun-summary-label">当前大运</span>
                <strong>{{ currentDayunStage?.pillar || currentDayunLayer?.pillar || '未定位' }}</strong>
              </div>
            </div>
            <div class="dayun-method-notes">
              <span>依据出生时刻与节气距离定起运</span>
              <span>按性别与年干阴阳定顺逆</span>
              <span>每十年推进一组干支</span>
            </div>
          </div>

          <div v-if="currentDayunStage || currentDayunLayer" class="analysis-block dayun-current-card">
            <div class="dayun-current-head">
              <div>
                <div class="block-title">当前大运重点</div>
                <span class="block-desc">当前阶段会成为流年、流月判断的底色，后续吉凶仍需叠加流年触发</span>
              </div>
              <span v-if="currentDayunLayer" class="dayun-current-score" :class="currentDayunLayer.favorable ? 'score-good' : 'score-watch'">
                {{ currentDayunLayer.favorable ? '偏有利' : '需留意' }} · {{ currentDayunLayer.score }}
              </span>
            </div>
            <div class="dayun-current-body">
              <div class="dayun-current-pillar">
                <span>{{ currentDayunStage?.gan || currentDayunLayer?.gan }}</span>
                <span>{{ currentDayunStage?.zhi || currentDayunLayer?.zhi }}</span>
              </div>
              <div class="dayun-current-copy">
                <div class="dayun-current-tags">
                  <span v-if="currentDayunStage" class="dayun-mini-chip">{{ currentDayunStage.startAge }}-{{ currentDayunStage.endAge }}岁</span>
                  <span v-if="currentDayunStage?.tenGod" class="dayun-mini-chip">十神 {{ currentDayunStage.tenGod }}</span>
                  <span v-if="currentDayunStage?.ganElement" class="dayun-mini-chip">天干 {{ currentDayunStage.ganElement }}</span>
                  <span v-if="currentDayunStage?.flowChange" class="dayun-mini-chip" :class="'dfb-' + currentDayunStage.flowChange">{{ currentDayunStage.flowChange }}</span>
                </div>
                <p>{{ currentDayunLayer?.description || currentDayunStage?.impact }}</p>
                <div v-if="currentDayunLayer?.evidence?.length" class="dayun-evidence-list">
                  <span v-for="item in currentDayunLayer.evidence" :key="item" class="dayun-evidence-chip">{{ item }}</span>
                </div>
              </div>
            </div>
          </div>

          <div v-if="dayunStages.length" class="analysis-block dayun-stage-block">
            <div class="block-title">十年阶段明细</div>
            <span class="block-desc">每一步展示年龄段、约略年份、干支五行、对日主十神，以及大运流通对命局的作用方向</span>
            <div class="dayun-stage-grid">
              <article
                v-for="stage in dayunStages"
                :key="stage.index + stage.pillar"
                class="dayun-stage-card"
                :class="[{ 'is-current': stage.isCurrent }, 'flow-' + stage.flowChange]"
              >
                <div class="dayun-stage-top">
                  <span>{{ stage.startAge }}-{{ stage.endAge }}岁</span>
                  <span v-if="stage.startYear && stage.endYear">约 {{ stage.startYear }}-{{ stage.endYear }}</span>
                  <span v-if="stage.isCurrent" class="dayun-now-badge">当前</span>
                </div>
                <div class="dayun-stage-main">
                  <div class="dayun-stage-pillar">
                    <span class="dayun-stage-gan" :style="{ color: elemColor(stage.ganElement) }">{{ stage.gan }}</span>
                    <span class="dayun-stage-zhi" :style="{ color: elemColor(stage.zhiElement) }">{{ stage.zhi }}</span>
                  </div>
                  <div class="dayun-stage-info">
                    <div class="dayun-stage-title">{{ stage.pillar }}大运</div>
                    <div class="dayun-stage-tags">
                      <span>十神 {{ stage.tenGod }}</span>
                      <span>干 {{ stage.ganElement || '未知' }}</span>
                      <span>支 {{ stage.zhiElement || '未知' }}</span>
                      <span :class="'dfb-' + stage.flowChange">{{ stage.flowChange }}</span>
                    </div>
                  </div>
                </div>
                <p class="dayun-stage-impact">{{ stage.impact }}</p>
              </article>
            </div>
          </div>

          <div v-if="dayunFlowItems.length" class="analysis-block dayun-flow-panel">
            <div class="block-title">大运流通明细</div>
            <span class="block-desc">从原“运势详批”移入大运：这里专看每步大运对日主五行的增强、减弱或泄秀作用</span>
            <div class="dayun-flow-list dayun-flow-grid">
              <div v-for="(item, di) in dayunFlowItems" :key="di" class="dayun-flow-item" :class="'df-' + item.flow_change">
                <div class="df-left">
                  <span class="df-age">{{ item.start_age }}岁</span>
                  <span class="df-pillar">{{ item.pillar }}</span>
                </div>
                <div class="df-right">
                  <span class="df-change-badge" :class="'dfb-' + item.flow_change">{{ item.flow_change }}</span>
                  <span class="df-impact">{{ item.impact }}</span>
                </div>
              </div>
            </div>
          </div>
        </div><!-- /dayun tab -->

        <!-- ═══ Tab: 五行格局 (wuxing) ═══ -->
        <div v-show="activeTab === 'wuxing'" class="tab-content">
          <!-- Element Detail -->
          <div v-if="chart.element_detail && chart.element_detail.length" class="analysis-block">
            <div class="block-title">五行力量与藏干分析</div>
            <span class="block-desc">天干明透与地支藏干的综合力量，揭示五行的真实强弱</span>
            <div class="element-detail-table">
              <div class="ed-header">
                <span>五行</span>
                <span>天干</span>
                <span>地支藏干</span>
                <span>总力量</span>
              </div>
              <div
                v-for="ed in chart.element_detail"
                :key="ed.element"
                class="ed-row"
                :class="'level-' + strengthLevel(ed.total)"
              >
                <span class="ed-elem" :style="{ color: elemColor(ed.element) }">{{ ed.element }}</span>
                <span class="ed-tg">{{ ed.tian_gan }}</span>
                <span class="ed-zc">{{ ed.cang_gan_list ? ed.cang_gan_list.join('、') : '—' }}</span>
                <span class="ed-total">{{ ed.total }}</span>
              </div>
            </div>
          </div>

          <!-- WuXingFlow 五行流通 -->
          <div v-if="chart.wuxing_flow" class="analysis-block">
            <div class="block-title">五行流通</div>
            <span class="block-desc">五行相生流转是否顺畅，流通则吉、阻滞则凶</span>
            <div class="wuxing-flow-card">
              <div class="wf-header">
                <span class="wf-label">日主</span>
                <span class="wf-elem" :style="{ color: elemColor(chart.wuxing_flow.day_element) }">{{ chart.wuxing_flow.day_element }}</span>
                <span class="wf-type">{{ chart.wuxing_flow.flow_type }}</span>
                <span class="wf-status" :class="chart.wuxing_flow.is_smooth ? 'wf-smooth' : 'wf-blocked'">
                  {{ chart.wuxing_flow.is_smooth ? '流通顺畅' : '流通受阻' }}
                </span>
              </div>
              <div v-if="chart.wuxing_flow.flow_paths?.length" class="wf-paths">
                <div v-for="(path, pi) in chart.wuxing_flow.flow_paths" :key="pi" class="wf-path">
                  <span class="wf-path-dot"></span>{{ path }}
                </div>
              </div>
              <div v-if="chart.wuxing_flow.blocked_element" class="wf-blocked-row">
                <span class="wf-blocked-label">阻滞</span>
                <span class="wf-blocked-elem" :style="{ color: elemColor(chart.wuxing_flow.blocked_element) }">{{ chart.wuxing_flow.blocked_element }}</span>
              </div>
              <div v-if="chart.wuxing_flow.balance_verdict" class="wf-verdict">{{ chart.wuxing_flow.balance_verdict }}</div>
              <div v-if="chart.wuxing_flow.advice" class="wf-advice">{{ chart.wuxing_flow.advice }}</div>
            </div>
          </div>

          <!-- TongGuan 通关用神 -->
          <div v-if="chart.tong_guan && chart.tong_guan.has_tong_guan" class="analysis-block">
            <div class="block-title">通关用神</div>
            <span class="block-desc">两行对峙取中间五行通关调和，化敌为友</span>
            <div class="tong-guan-card">
              <span class="tg-elem" :style="{ color: elemColor(chart.tong_guan.tong_guan_element) }">{{ chart.tong_guan.tong_guan_element }}</span>
              <span class="tg-desc">{{ chart.tong_guan.description }}</span>
            </div>
          </div>

          <!-- MissingElements 缺失五行 -->
          <div v-if="chart.missing_elements && chart.missing_elements.missing_elements?.length" class="analysis-block">
            <div class="block-title">五行缺失</div>
            <span class="block-desc">命局中完全不见的五行，需通过后天补救平衡</span>
            <div class="missing-elem-card">
              <div class="me-row">
                <span class="me-label">缺失</span>
                <span v-for="e in chart.missing_elements.missing_elements" :key="'m'+e" class="me-tag me-missing" :style="{ color: elemColor(e), borderColor: elemColor(e) + '44' }">{{ e }}</span>
                <span v-if="chart.missing_elements.severity" class="me-severity" :class="'me-' + chart.missing_elements.severity">{{ chart.missing_elements.severity }}</span>
              </div>
              <div v-if="chart.missing_elements.remedy_elements?.length" class="me-row">
                <span class="me-label">补救</span>
                <span v-for="e in chart.missing_elements.remedy_elements" :key="'r'+e" class="me-tag me-remedy" :style="{ color: elemColor(e), borderColor: elemColor(e) + '44' }">{{ e }}</span>
              </div>
            </div>
          </div>
        </div><!-- /wuxing tab -->

        <!-- ═══ Tab: 十神详解 (shishen) ═══ -->
        <div v-show="activeTab === 'shishen'" class="tab-content">
          <!-- TenGodProportion -->
          <div v-if="chart.ten_god_proportion && chart.ten_god_proportion.length" class="analysis-block">
            <div class="block-title">十神占比</div>
            <span class="block-desc">各十神在命局中的比重，占比越高影响力越大</span>
            <div class="ten-god-chart-wrap">
              <v-chart class="ten-god-chart" :option="(tenGodChartOptions as any)" autoresize />
            </div>
          </div>

          <!-- TenGodAnalysis -->
          <div v-if="chart.ten_god_analysis" class="analysis-block">
            <div class="block-title">十神综合解读</div>
            <span class="block-desc">综合十神组合对性格、事业、感情、健康的深层影响</span>
            <el-tabs class="ten-god-tabs" tab-position="top">
              <el-tab-pane label="综合概述">
                <div class="tg-summary">{{ chart.ten_god_analysis.summary }}</div>
              </el-tab-pane>
              <el-tab-pane label="性格特点">
                <div class="tg-text">{{ chart.ten_god_analysis.personality }}</div>
              </el-tab-pane>
              <el-tab-pane label="人际关系">
                <div class="tg-text">{{ chart.ten_god_analysis.interpersonal }}</div>
              </el-tab-pane>
              <el-tab-pane label="事业财运">
                <div class="tg-text">{{ chart.ten_god_analysis.career_fortune }}</div>
              </el-tab-pane>
              <el-tab-pane label="感情姻缘">
                <div class="tg-text">{{ chart.ten_god_analysis.emotion_relation }}</div>
              </el-tab-pane>
              <el-tab-pane label="健康提醒">
                <div class="tg-text">{{ chart.ten_god_analysis.health_note }}</div>
              </el-tab-pane>
              <el-tab-pane label="十神详解">
                <div class="tg-god-list">
                  <div v-for="god in chart.ten_god_analysis.god_relations" :key="god.god" class="tg-god-card">
                    <div class="tg-god-header">
                      <span class="tg-god-name">{{ god.god }}</span>
                      <span class="tg-god-pct">{{ god.percent }}</span>
                    </div>
                    <div class="tg-god-meaning">{{ god.meaning }}</div>
                    <div class="tg-god-advice">{{ god.advice }}</div>
                  </div>
                </div>
              </el-tab-pane>
            </el-tabs>
          </div>
        </div><!-- /shishen tab -->

        <!-- ═══ Tab: 格局古籍 (pattern) ═══ -->
        <div v-show="activeTab === 'pattern'" class="tab-content">
          <!-- Pattern Analysis -->
          <div v-if="chart.pattern_analysis" class="analysis-block">
            <div class="block-title">命局格局</div>
            <span class="block-desc">判断命局属于正格还是特殊格局，格局决定全局喜忌方向</span>
            <div class="pattern-detail">
              <div class="pattern-header">
                <span class="pattern-name">{{ chart.pattern_analysis.pattern_name }}</span>
                <span class="pattern-type-tag" :class="chart.pattern_analysis.pattern_type === '特殊格局' ? 'tag-special' : 'tag-normal'">
                  {{ chart.pattern_analysis.pattern_type }}
                </span>
              </div>
              <p class="pattern-desc">{{ chart.pattern_analysis.description }}</p>
              <div class="pattern-elems">
                <div v-if="chart.pattern_analysis.favorable_elements?.length" class="pattern-favor">
                  <span class="pattern-elem-label">喜</span>
                  <span v-for="e in chart.pattern_analysis.favorable_elements" :key="'fav'+e" class="pattern-elem-tag favor">{{ e }}</span>
                </div>
                <div v-if="chart.pattern_analysis.unfavorable_elements?.length" class="pattern-avoid">
                  <span class="pattern-elem-label">忌</span>
                  <span v-for="e in chart.pattern_analysis.unfavorable_elements" :key="'unf'+e" class="pattern-elem-tag avoid">{{ e }}</span>
                </div>
              </div>
            </div>
          </div>

          <!-- Body Strength -->
          <div v-if="chart.body_strength" class="analysis-block">
            <div class="block-title">身强评分与喜忌</div>
            <span class="block-desc">扶抑法判断日主强弱，身旺抑之、身弱扶之；若命局为特殊格局，则以格局取用为主、评分作参考</span>
            <div class="body-strength">
              <div class="bs-verdict">{{ chart.body_strength.verdict }}</div>
              <span class="bs-score">{{ bodyScorePercent }}%</span>
              <div class="bs-tags">
                <span class="bs-like-label">喜</span>
                <span v-for="l in chart.body_strength.like" :key="l" class="bs-like">{{ l }}</span>
                <span class="bs-dislike-label">忌</span>
                <span v-for="d in chart.body_strength.dislike" :key="d" class="bs-dislike">{{ d }}</span>
              </div>
            </div>
            <div v-if="bodyStrengthComponents.length" class="evidence-bars body-strength-bars">
              <div v-for="component in bodyStrengthComponents" :key="component.key" class="evidence-bar-row">
                <span class="evidence-bar-label">{{ component.name }}</span>
                <div class="evidence-bar-track">
                  <span class="evidence-bar-fill" :style="{ width: componentWidth(component.normalized_score) }"></span>
                </div>
                <span class="evidence-bar-value">{{ component.normalized_score }}</span>
              </div>
            </div>
            <div v-if="bodyStrengthEvidence.length" class="evidence-notes body-strength-notes">
              <span v-for="item in bodyStrengthEvidence" :key="item.component + item.source + item.item" class="evidence-note">
                {{ item.source }}·{{ item.item }}{{ item.score >= 0 ? ' +' : ' ' }}{{ item.score.toFixed(2) }}
              </span>
            </div>
          </div>

          <!-- Tiaohou from 《穷通宝鉴》 -->
          <div v-if="chart.tiaohou" class="analysis-block">
            <div class="block-title">调候用神 <span class="tiaohou-source">《穷通宝鉴》</span></div>
            <span class="block-desc">根据日干生于各月的寒暖燥湿，取调候之用神以求中和</span>
            <div class="tiaohou-card">
              <div class="tiaohou-header">
                <span class="tiaohou-stem">{{ chart.tiaohou.stem }}</span>
                <span class="tiaohou-arrow">生</span>
                <span class="tiaohou-month">{{ chart.tiaohou.month }}</span>
                <span class="tiaohou-divider">|</span>
                <span class="tiaohou-label">用神</span>
                <span class="tiaohou-primary" :style="{ color: elemColor(tiaohouElem) }">{{ chart.tiaohou.primary_god }}</span>
              </div>
              <div class="tiaohou-summary">{{ chart.tiaohou.summary }}</div>
              <div v-if="chart.tiaohou.rules && chart.tiaohou.rules.length" class="tiaohou-rules">
                <div v-for="(rule, idx) in chart.tiaohou.rules" :key="idx" class="tiaohou-rule">
                  <span class="tiaohou-xi">喜{{ rule.xi_shen }}</span>
                  <span class="tiaohou-ji">忌{{ rule.ji_shen }}</span>
                  <span class="tiaohou-reason">{{ rule.reason }}</span>
                </div>
              </div>
            </div>
          </div>

          <!-- JinBuHuan -->
          <div v-if="chart.jin_bu_huan" class="analysis-block">
            <div class="block-title">金不换</div>
            <span class="block-desc">以日干查十二月令的富贵贫贱断语，出自古籍《金不换》</span>
            <p class="jin-bu-huan-text">{{ chart.jin_bu_huan }}</p>
          </div>

          <!-- RiZhuDesc (legacy fallback) -->
          <div v-if="chart.ri_zhu_desc && !chart.ri_zhu_poem" class="analysis-block">
            <div class="block-title">日主坐命</div>
            <span class="block-desc">日柱天干坐支的组合断语，揭示命主的核心气质</span>
            <p class="ri-zhu-text">{{ chart.ri_zhu_desc }}</p>
          </div>

          <!-- RiZhuPoem -->
          <div v-if="chart.ri_zhu_poem" class="analysis-block">
            <div class="block-title">日主诗意</div>
            <span class="block-desc">古人以诗赋形式描绘日柱特质，意境深远</span>
            <p class="ri-zhu-text">{{ chart.ri_zhu_poem }}</p>
          </div>

          <!-- RiZhuSource -->
          <div v-if="chart.ri_zhu_source" class="analysis-block">
            <div class="block-title">古籍出处</div>
            <span class="block-desc">诗赋的典籍来源，追溯命理学的文化传承</span>
            <p class="ri-zhu-text">{{ chart.ri_zhu_source }}</p>
          </div>

          <!-- RiZhuComment -->
          <div v-if="chart.ri_zhu_comment" class="analysis-block">
            <div class="block-title">补充判断</div>
            <span class="block-desc">结合现代视角对日柱特质的补充解读</span>
            <p class="ri-zhu-text">{{ chart.ri_zhu_comment }}</p>
          </div>

          <!-- RiZhuHourDetail -->
          <div v-if="chart.ri_zhu_hour_detail" class="analysis-block">
            <div class="block-title">时辰详批</div>
            <span class="block-desc">日柱配合时柱的精细论断，时辰为命局的归宿</span>
            <p class="ri-zhu-text">{{ chart.ri_zhu_hour_detail }}</p>
          </div>
        </div><!-- /pattern tab -->

        <!-- ═══ Tab: 神煞 (shensha) ═══ -->
        <div v-show="activeTab === 'shensha'" class="tab-content">
          <section v-if="shenShaCategorySummary.length" class="analysis-block shensha-overview-card">
            <div class="shensha-overview-head">
              <div>
                <div class="block-title">神煞分类</div>
                <span class="block-desc">按吉凶倾向与用途归类，分类只作为阅读索引，具体判断仍看柱位、触发条件和全局组合</span>
              </div>
              <span class="shensha-overview-count">{{ shenShaDetails.length }} 项</span>
            </div>
            <div class="evidence-table-list shensha-category-pills">
              <span v-for="item in shenShaCategorySummary" :key="item.category" class="evidence-table-pill shensha-pill">
                {{ item.category }} · {{ item.count }}
              </span>
            </div>
            <div class="shensha-grid">
              <article v-for="item in shenShaDetails" :key="item.name + item.source" class="shen-sha-detail-row shensha-detail-card" :class="polarityClass(item.polarity)">
                <div class="shen-sha-detail-top">
                  <span class="shen-sha-detail-name">{{ item.name }}</span>
                  <span class="shen-sha-detail-meta">{{ item.category }}</span>
                  <span class="shen-sha-detail-priority">优先级 {{ item.priority }}</span>
                </div>
                <span class="shen-sha-detail-desc">{{ item.description }}</span>
              </article>
            </div>
          </section>

          <!-- ShenSha (grouped by pillar when available) -->
          <template v-if="groupedShenSha">
            <template v-for="group in groupedShenSha" :key="group.pillar">
              <section v-if="group.items && group.items.length" class="analysis-block shensha-group-card">
                <div class="shen-sha-group-title shensha-group-head">
                  <span class="shen-sha-group-dot" :style="{ background: pillarShenShaColor(group.pillar) }"></span>
                  <span class="shen-sha-group-label">{{ group.label }}神煞</span>
                  <span class="shen-sha-group-role">· {{ group.role }}</span>
                </div>
                <div class="shen-sha-list">
                  <article
                    v-for="sha in group.items"
                    :key="sha.name + sha.target + sha.desc"
                    class="shen-sha-row shensha-row-card"
                    :style="{ background: pillarShenShaBg(group.pillar) }"
                  >
                    <span class="shen-sha-name" :style="{ color: pillarShenShaColor(group.pillar) }">{{ sha.name }}</span>
                    <span v-if="sha.target" class="shen-sha-target">{{ sha.target }}</span>
                    <span v-if="sha.desc" class="shen-sha-desc">{{ sha.desc }}</span>
                  </article>
                </div>
              </section>
            </template>
            <section v-if="globalShenSha.length" class="analysis-block shensha-group-card">
              <div class="shen-sha-group-title shensha-group-head">
                <span class="shen-sha-group-dot" style="background: var(--accent)"></span>
                <span class="shen-sha-group-label">全局组合神煞</span>
                <span class="shen-sha-group-role">· 多柱配合</span>
              </div>
              <div class="shen-sha-list">
                <article
                  v-for="sha in globalShenSha"
                  :key="sha.name + sha.target + sha.desc"
                  class="shen-sha-row shensha-row-card"
                  :style="{ background: 'var(--accent-dim)' }"
                >
                  <span class="shen-sha-name" :style="{ color: 'var(--accent)' }">{{ sha.name }}</span>
                  <span v-if="sha.target" class="shen-sha-target">{{ sha.target }}</span>
                  <span v-if="sha.desc" class="shen-sha-desc">{{ sha.desc }}</span>
                </article>
              </div>
            </section>
            <section v-if="showSummary" class="analysis-block shensha-summary-card">
              <div class="shen-sha-summary-title">{{ props.chart.shen_sha_summary.title }}</div>
              <ul class="shen-sha-summary-list">
                <li v-for="line in props.chart.shen_sha_summary.description" :key="line.slice(0, 16)">
                  {{ line }}
                </li>
              </ul>
            </section>
          </template>
          <section v-else-if="parsedDayShenSha.length" class="analysis-block shensha-group-card">
            <div class="block-title">日柱神煞</div>
            <span class="block-desc">日柱天干地支所临神煞，揭示先天吉凶信息</span>
            <div class="shen-sha-list">
              <article v-for="sha in parsedDayShenSha" :key="sha.name + sha.target + sha.desc" class="shen-sha-row shensha-row-card">
                <span class="shen-sha-name">{{ sha.name }}</span>
                <span v-if="sha.target" class="shen-sha-target">{{ sha.target }}</span>
                <span v-if="sha.desc" class="shen-sha-desc">{{ sha.desc }}</span>
              </article>
            </div>
          </section>
        </div><!-- /shensha tab -->

        <!-- ═══ Tab: 运势详批 (fortune) ═══ -->
        <div v-show="activeTab === 'fortune'" class="tab-content fortune-detail-tab">
          <div v-if="fortuneLayerList.length" class="analysis-block">
            <div class="block-title">运势层依据</div>
            <span class="block-desc">流年、流月、小运分层叠加；大运已移入独立标签，作为这些短周期判断的底色</span>
            <div class="fortune-layer-list">
              <div v-for="layer in fortuneLayerList" :key="layer.key" class="fortune-layer-row">
                <div class="fortune-layer-top">
                  <span class="fortune-layer-name">{{ layer.name }}</span>
                  <span class="fortune-layer-pillar">{{ layer.pillar }}</span>
                  <span class="fortune-layer-god">{{ layer.ten_god }}</span>
                </div>
                <div class="fortune-layer-sub">
                  <span v-if="layer.start_age !== undefined" class="fortune-layer-chip">{{ layer.start_age }}岁起</span>
                  <span v-if="layer.year" class="fortune-layer-chip">{{ layer.year }}年</span>
                  <span v-if="layer.month" class="fortune-layer-chip">{{ layer.month }}月</span>
                  <span class="fortune-layer-chip" :class="layer.favorable ? 'sha-good' : 'sha-risk'">{{ layer.favorable ? '有利' : '需察' }}</span>
                </div>
                <div v-if="layer.evidence?.length" class="fortune-layer-evidence">
                  {{ layer.evidence[0] }}
                </div>
              </div>
            </div>
          </div>

          <!-- MingGong -->
          <div v-if="chart.ming_gong && chart.ming_gong.gan_zhi" class="analysis-block">
            <div class="block-title">命宫 · 第五柱</div>
            <span class="block-desc">以出生时辰推算命宫，被视为八字之外的第五柱</span>
            <div class="ming-gong-detail">
              <div class="ming-gong-main">
                <span class="ming-gong-ganzhi">{{ chart.ming_gong.gan_zhi }}</span>
                <span v-if="chart.ming_gong.nayin" class="ming-gong-nayin">({{ chart.ming_gong.nayin }})</span>
              </div>
              <div v-if="chart.ming_gong.shen_sha" class="ming-gong-shensha">
                <span class="shensha-badge" :class="{ 'shensha-ji': isJiShenSha(chart.ming_gong.shen_sha), 'shensha-xiong': isXiongShenSha(chart.ming_gong.shen_sha) }">
                  {{ chart.ming_gong.shen_sha }}星
                </span>
                <span class="shensha-desc">{{ chart.ming_gong.shen_sha_desc }}</span>
              </div>
              <div v-if="chart.ming_gong.zhi_detail" class="ming-gong-zhi">
                {{ chart.ming_gong.zhi_detail }}
              </div>
            </div>
          </div>

          <!-- SeasonText -->
          <div v-if="chart.season_text" class="analysis-block">
            <div class="block-title">季节解读</div>
            <span class="block-desc">日主生于当令季节的旺衰状态与调候要点</span>
            <p class="season-text">{{ chart.season_text }}</p>
          </div>

          <!-- SeasonTextMonth -->
          <div v-if="chart.season_text_month" class="analysis-block">
            <div class="block-title">{{ birthMonthLabel }}解读</div>
            <span class="block-desc">该月特有的节气特征与调候细则</span>
            <p class="season-text">{{ chart.season_text_month }}</p>
          </div>

          <!-- FlowPatternDesc 流通格局 -->
          <div v-if="chart.flow_pattern_desc" class="analysis-block">
            <div class="block-title">流通格局</div>
            <span class="block-desc">五行流通形成的特殊格局类型与化解之道</span>
            <p class="season-text">{{ chart.flow_pattern_desc }}</p>
          </div>

        </div><!-- /fortune tab -->

        <!-- ═══ Tab: 规则依据 (rules) ═══ -->
        <div v-show="activeTab === 'rules'" class="tab-content">
          <div v-if="chart.id" class="classical-rules-block">
            <ClassicalInterpretationPanel :chart-id="chart.id" />
          </div>

          <div v-if="ruleMeta" class="analysis-block">
            <div class="block-title">bazi-rules</div>
            <span class="block-desc">后端确定性规则集版本，用来标记本次排盘采用的规则表、权重和流派，不是新的命理结论</span>
            <div class="rule-meta-card">
              <div class="rule-meta-row">
                <span class="rule-meta-label">版本</span>
                <span class="rule-meta-value">{{ ruleMeta.rule_version }}</span>
              </div>
              <div class="rule-meta-row">
                <span class="rule-meta-label">流派</span>
                <span class="rule-meta-value">{{ ruleMeta.school }}</span>
              </div>
            </div>
          </div>

          <div v-if="ruleMeta?.body_strength" class="analysis-block">
            <div class="block-title">身强评分权重</div>
            <span class="block-desc">这里展示评分规则本身；实际得分已放入“格局古籍”的身旺喜忌卡片</span>
            <div class="rule-weight-grid">
              <span v-for="(value, key) in ruleMeta.body_strength.weights" :key="'w'+key" class="rule-weight-chip">
                {{ key }} · {{ value }}
              </span>
            </div>
          </div>

          <div v-if="ruleTablePreview.length" class="analysis-block">
            <div class="block-title">规则表清单</div>
            <span class="block-desc">这些表分别支撑十神、藏干、纳音、神煞、调候、身强和运势分层的确定性计算</span>
            <div class="rule-table-list">
              <article v-for="table in ruleTablePreview" :key="table.key" class="rule-table-card">
                <div class="rule-table-head">
                  <span class="rule-table-name">{{ table.name }}</span>
                  <span class="rule-table-version">v{{ table.version }}</span>
                </div>
                <div class="rule-table-source">{{ table.school }} · {{ table.source }}</div>
                <p class="rule-table-desc">{{ table.description }}</p>
                <span v-if="table.count" class="rule-table-count">{{ table.count }} 条</span>
              </article>
            </div>
          </div>
        </div><!-- /rules tab -->

      </div><!-- /analysis-section -->
    </div>
  </div>
</template>

<style scoped>
.bazi-chart {
  position: relative;
}

.chart-bg {
  position: absolute;
  inset: 0;
  pointer-events: none;
  overflow: hidden;
}

.bg-svg {
  width: 100%;
  height: 100%;
  position: absolute;
  top: 0;
  left: 0;
}

.chart-card {
  position: relative;
  z-index: 1;
}

.chart-card.glass-card:hover {
  box-shadow: 0 0 0 1px var(--line-strong), var(--shadow-xs);
  transform: none;
}

/* Header */
.chart-header {
  background: linear-gradient(180deg, var(--line-subtle), transparent);
  border-bottom: 1px solid var(--line-subtle);
  padding: 1rem 1.25rem;
  text-align: center;
}

.header-eyebrow {
  font-size: var(--fs-2xs);
  letter-spacing: 3px;
  color: var(--text-soft);
  text-transform: uppercase;
  margin-bottom: 4px;
}

.chart-title {
  font-family: var(--font-serif), serif;
  font-size: var(--fs-2xl);
  font-weight: 700;
  color: var(--text);
  margin: 0;
  letter-spacing: 3px;
}

/* Pillars bento grid */
.pillars-bento {
  display: grid;
  grid-template-columns:
    minmax(96px, 0.86fr)
    minmax(96px, 0.86fr)
    minmax(210px, 1.35fr)
    minmax(96px, 0.86fr)
    minmax(96px, 0.86fr);
  gap: 0.55rem;
  padding: 0.75rem;
  border-bottom: 1px solid var(--line-subtle);
  align-items: stretch;
}

.pillars-bento > .bento-card {
  min-height: 188px;
}

.bento-card {
  background: var(--accent-dim);
  border: 1px solid var(--line-subtle);
  border-radius: 10px;
  padding: 0.6rem 0.75rem;
  transition: box-shadow 0.3s, background 0.3s;
  position: relative;
  overflow: hidden;
}

.bento-small {
  display: flex;
  flex-direction: column;
}

.bento-card.bento-small:hover,
.bento-card.bento-day-card:hover {
  background: var(--surface-2);
}

/* Hover 五行微光 */
.bento-hover-木:hover { box-shadow: inset 0 0 30px color-mix(in oklab, var(--wuxing-mu) 7%, transparent); }
.bento-hover-火:hover { box-shadow: inset 0 0 30px color-mix(in oklab, var(--wuxing-huo) 7%, transparent); }
.bento-hover-土:hover { box-shadow: inset 0 0 30px color-mix(in oklab, var(--wuxing-tu) 7%, transparent); }
.bento-hover-金:hover { box-shadow: inset 0 0 30px color-mix(in oklab, var(--wuxing-jin) 7%, transparent); }
.bento-hover-水:hover { box-shadow: inset 0 0 30px color-mix(in oklab, var(--wuxing-shui) 7%, transparent); }

.bento-label {
  font-size: var(--fs-2xs);
  color: var(--text-muted);
  letter-spacing: 1px;
  margin-bottom: 0.4rem;
  font-weight: 600;
}

.bento-body {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.35rem;
  flex: 1;
  justify-content: center;
}

.bento-gan,
.bento-zhi {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.15rem;
}

.bento-char {
  font-family: var(--font-serif), serif;
  font-size: var(--fs-3xl);
  font-weight: 700;
  line-height: 1;
  text-shadow: 0 0 12px currentColor;
}

.bento-day-card {
  border-color: var(--line-focus);
  background:
    linear-gradient(180deg, var(--accent-dim), transparent 145%),
    var(--glass-bg);
}

/* 五行雷达图卡片 */
.bento-radar {
  display: flex;
  flex-direction: column;
  order: 3;
  min-height: 188px;
  padding-bottom: 0.35rem;
}

.bento-radar-chart {
  flex: 1;
  min-height: 148px;
  height: 148px;
  width: 100%;
}

.bento-sub {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.1rem;
  padding-top: 0.25rem;
  border-top: 1px solid var(--line-subtle);
  margin-top: 0.25rem;
}

.elem-tag {
  display: inline-block;
  font-size: var(--fs-2xs);
  padding: 0.1rem 0.35rem;
  border-radius: 3px;
  border: 1px solid;
  font-weight: 600;
  letter-spacing: 0.5px;
}

.bento-god-tag {
  font-size: var(--fs-2xs);
  padding: 0.1rem 0.4rem;
  border-radius: 4px;
  background: var(--surface-2);
  border: 1px solid var(--line-strong);
  color: var(--text-muted);
  font-weight: 600;
  letter-spacing: 0.5px;
  margin-top: 0.2rem;
}

.bento-god-tag-day {
  font-size: var(--fs-2xs);
  padding: 0.15rem 0.5rem;
  background: var(--line-strong);
  border-color: var(--line-focus);
  color: var(--accent);
}

/* ===== 天干地支关系 — information ledger ===== */

.ganzhi-analysis {
  margin-top: 2px;
}

.overview-relations {
  --rel-he: #16845a;
  --rel-ke: #b94a48;
  --rel-sheng: #2f6fb7;
  --rel-hai: #a86622;
  --rel-xing: #7e5aa8;
}

.relations-section {
  padding: 0.9rem 1rem 1rem;
  border-bottom: 1px solid var(--line-subtle);
}

.relations-title {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  padding-bottom: 0.68rem;
  border-bottom: 1px solid var(--line-subtle);
  font-size: var(--fs-xs);
  font-weight: 700;
  color: var(--text);
  letter-spacing: 0.04em;
  margin-bottom: 0.65rem;
}

.relations-title-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--rel-he);
  box-shadow: none;
}

.relations-title-dot.zhi-dot {
  background: var(--rel-sheng);
  box-shadow: none;
}

.relations-count {
  display: inline-flex;
  align-items: center;
  min-height: 22px;
  padding: 0 0.48rem;
  border: 1px solid var(--line-subtle);
  border-radius: 999px;
  background: color-mix(in oklab, var(--surface-0) 82%, transparent);
  font-size: var(--fs-2xs);
  font-weight: 600;
  color: var(--text-soft);
  margin-left: auto;
  letter-spacing: 0.02em;
}

.ganzhi-compact {
  display: flex;
  flex-direction: column;
  gap: 0.38rem;
}

.gz-item {
  --rel-color: var(--line-focus);
  display: grid;
  grid-template-columns: max-content max-content minmax(0, 1fr);
  align-items: center;
  gap: 0.55rem;
  min-height: 40px;
  padding: 0.48rem 0.62rem;
  border-radius: 7px;
  border-left: 3px solid color-mix(in oklab, var(--rel-color) 62%, transparent);
  background: color-mix(in oklab, var(--surface-0) 58%, transparent);
  transition: background 0.18s ease, border-color 0.18s ease;
}

.gz-item:hover {
  background: color-mix(in oklab, var(--rel-color) 7%, var(--surface-0));
}

.gz-item.rel-he { --rel-color: var(--rel-he); }
.gz-item.rel-ke { --rel-color: var(--rel-ke); }
.gz-item.rel-sheng { --rel-color: var(--rel-sheng); }
.gz-item.rel-chong { --rel-color: var(--rel-ke); }
.gz-item.rel-hai { --rel-color: var(--rel-hai); }
.gz-item.rel-xing { --rel-color: var(--rel-xing); }
.gz-item.rel-hui { --rel-color: var(--rel-sheng); }

.gz-chars {
  display: flex;
  align-items: center;
  gap: 0.15rem;
  min-width: 6.25rem;
  flex-shrink: 0;
}

.gz-c {
  font-family: var(--font-serif), serif;
  font-size: var(--fs-md);
  font-weight: 700;
  color: var(--text);
  line-height: 1.2;
}

.gz-sym {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 1.35rem;
  height: 1.35rem;
  border-radius: 999px;
  border: 1px solid transparent;
  font-size: var(--fs-2xs);
  font-weight: 700;
  line-height: 1;
}

.sym-rel-he { --rel-color: var(--rel-he); }
.sym-rel-ke { --rel-color: var(--rel-ke); }
.sym-rel-sheng { --rel-color: var(--rel-sheng); }
.sym-rel-chong { --rel-color: var(--rel-ke); }
.sym-rel-hai { --rel-color: var(--rel-hai); }
.sym-rel-xing { --rel-color: var(--rel-xing); }
.sym-rel-hui { --rel-color: var(--rel-sheng); }

.gz-tag {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 24px;
  font-size: var(--fs-2xs);
  font-weight: 700;
  padding: 0.08rem 0.42rem;
  border: 1px solid transparent;
  border-radius: 6px;
  letter-spacing: 0.02em;
  flex-shrink: 0;
}

.tag-rel-he { --rel-color: var(--rel-he); }
.tag-rel-ke { --rel-color: var(--rel-ke); }
.tag-rel-sheng { --rel-color: var(--rel-sheng); }
.tag-rel-chong { --rel-color: var(--rel-ke); }
.tag-rel-hai { --rel-color: var(--rel-hai); }
.tag-rel-xing { --rel-color: var(--rel-xing); }
.tag-rel-hui { --rel-color: var(--rel-sheng); }

.gz-sym,
.gz-tag {
  color: var(--rel-color);
  background: color-mix(in oklab, var(--rel-color) 11%, transparent);
  border-color: color-mix(in oklab, var(--rel-color) 18%, transparent);
}

.gz-text {
  color: var(--text-muted);
  font-size: var(--fs-xs);
  line-height: 1.5;
  min-width: 0;
  white-space: normal;
}

.no-relations {
  min-height: 132px;
  padding: 1rem;
  text-align: center;
  font-size: var(--fs-xs);
  color: var(--text-soft);
  border-bottom: 1px solid var(--line-subtle);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
}

.no-rel-icon {
  font-size: var(--fs-2xs);
  color: var(--text-soft);
}

/* Analysis sections */
.analysis-section {
  padding: 1rem 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.evidence-table-list,
.evidence-notes {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
}

.evidence-table-pill,
.evidence-note,
.fortune-layer-chip {
  display: inline-flex;
  align-items: center;
  min-height: 24px;
  padding: 0.15rem 0.48rem;
  border: 1px solid var(--line-subtle);
  border-radius: 4px;
  color: var(--text-muted);
  font-size: var(--fs-2xs);
  line-height: 1.2;
  background: var(--surface-0);
}

.evidence-bars {
  display: flex;
  flex-direction: column;
  gap: 0.42rem;
}

.evidence-bar-row {
  display: grid;
  grid-template-columns: 3.2rem minmax(80px, 1fr) 2.4rem;
  align-items: center;
  gap: 0.5rem;
}

.evidence-bar-label,
.evidence-bar-value {
  color: var(--text-muted);
  font-size: var(--fs-2xs);
}

.evidence-bar-value {
  text-align: right;
  font-family: var(--font-mono);
}

.evidence-bar-track {
  height: 6px;
  border-radius: 999px;
  background: var(--line-subtle);
  overflow: hidden;
}

.evidence-bar-fill {
  display: block;
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(90deg, var(--accent), #22d3ee);
}

.shen-sha-detail-list,
.fortune-layer-list {
  display: flex;
  flex-direction: column;
  gap: 0.45rem;
}

.shen-sha-detail-list {
  margin-top: 0.6rem;
}

.shen-sha-detail-row,
.fortune-layer-row {
  border-left: 2px solid var(--line-strong);
  padding-left: 0.55rem;
  min-width: 0;
}

.shen-sha-detail-row.sha-good,
.fortune-layer-chip.sha-good {
  border-color: rgba(22, 163, 74, 0.35);
  color: #16a34a;
}

.shen-sha-detail-row.sha-risk,
.fortune-layer-chip.sha-risk {
  border-color: rgba(220, 38, 38, 0.35);
  color: #dc2626;
}

.shen-sha-detail-row.sha-neutral {
  border-color: rgba(100, 149, 237, 0.35);
}

.shen-sha-detail-name,
.fortune-layer-name {
  display: inline-block;
  margin-right: 0.4rem;
  font-size: var(--fs-xs);
  font-weight: 700;
  color: var(--text);
}

.shen-sha-detail-meta,
.fortune-layer-god {
  font-size: var(--fs-2xs);
  color: var(--text-soft);
}

.shen-sha-detail-desc,
.fortune-layer-evidence {
  display: block;
  margin-top: 0.18rem;
  color: var(--text-muted);
  font-size: var(--fs-2xs);
  line-height: 1.45;
}

.shensha-overview-card,
.shensha-group-card,
.shensha-summary-card {
  padding: 0.85rem 0.95rem;
  border: 1px solid var(--line-subtle);
  border-radius: 10px;
  background: color-mix(in oklab, var(--surface-2) 38%, transparent);
}

.shensha-overview-card {
  display: flex;
  flex-direction: column;
  gap: 0.8rem;
}

.shensha-overview-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.75rem;
  min-width: 0;
}

.shensha-overview-head .block-desc {
  max-width: 54rem;
}

.shensha-overview-count {
  display: inline-flex;
  align-items: center;
  min-height: 24px;
  padding: 0 0.48rem;
  border: 1px solid var(--line-subtle);
  border-radius: 999px;
  background: var(--surface-0);
  color: var(--text-soft);
  font-size: var(--fs-2xs);
  font-weight: 600;
  white-space: nowrap;
  flex-shrink: 0;
}

.shensha-category-pills {
  gap: 0.28rem;
  margin-bottom: 0;
}

.shensha-pill {
  min-height: 22px;
  padding: 0.12rem 0.42rem;
  border-radius: 4px;
}

.shensha-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.42rem 0.55rem;
}

.shensha-detail-card {
  display: flex;
  flex-direction: column;
  gap: 0.22rem;
  padding: 0.48rem 0.55rem;
  border-radius: 8px;
  background: var(--surface-0);
}

.shen-sha-detail-top {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.3rem 0.45rem;
  min-width: 0;
}

.shen-sha-detail-priority {
  font-size: var(--fs-2xs);
  color: var(--text-soft);
}

.shensha-group-card {
  display: flex;
  flex-direction: column;
  gap: 0.65rem;
}

.shensha-group-head {
  margin-bottom: 0;
}

.shen-sha-group-label {
  min-width: 0;
}

.shensha-row-card {
  padding: 0.48rem 0.55rem;
  border-radius: 8px;
  border-left-width: 3px;
  background: var(--surface-0);
}

.shensha-summary-card {
  border-top: 1px solid var(--line-subtle);
  margin-top: 0;
  padding-top: 0.85rem;
}

.fortune-layer-top {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  min-width: 0;
}

.fortune-layer-pillar {
  font-family: var(--font-serif), serif;
  color: var(--accent);
  font-weight: 700;
}

.fortune-layer-sub {
  display: flex;
  flex-wrap: wrap;
  gap: 0.3rem;
  margin-top: 0.35rem;
}

.shensha-category-pills {
  margin-bottom: 0.75rem;
}

.shensha-meta-list {
  margin-top: 0;
}

.rule-meta-card,
.rule-table-card {
  border: 1px solid var(--line-subtle);
  border-radius: 8px;
  background: color-mix(in oklab, var(--surface-2) 42%, transparent);
}

.rule-meta-card {
  display: grid;
  gap: 0.45rem;
  padding: 0.75rem;
}

.rule-meta-row,
.rule-table-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  min-width: 0;
}

.rule-meta-label,
.rule-table-source,
.rule-table-count {
  font-size: var(--fs-2xs);
  color: var(--text-soft);
}

.rule-meta-value {
  min-width: 0;
  text-align: right;
  color: var(--text);
  font-size: var(--fs-xs);
  font-family: var(--font-mono);
  overflow-wrap: anywhere;
}

.rule-weight-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem;
}

.rule-weight-chip {
  display: inline-flex;
  align-items: center;
  min-height: 24px;
  padding: 0.15rem 0.5rem;
  border: 1px solid var(--line-subtle);
  border-radius: 4px;
  background: var(--surface-0);
  color: var(--text-muted);
  font-size: var(--fs-2xs);
  font-family: var(--font-mono);
}

.rule-table-list {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.7rem;
}

.rule-table-card {
  position: relative;
  padding: 0.75rem;
  min-width: 0;
}

.rule-table-name {
  color: var(--text);
  font-size: var(--fs-xs);
  font-weight: 700;
}

.rule-table-version {
  flex-shrink: 0;
  color: var(--accent);
  font-size: var(--fs-2xs);
  font-family: var(--font-mono);
}

.rule-table-source {
  margin-top: 0.35rem;
}

.rule-table-desc {
  margin: 0.4rem 0 0;
  color: var(--text-muted);
  font-size: var(--fs-2xs);
  line-height: 1.55;
}

.rule-table-count {
  display: inline-flex;
  margin-top: 0.55rem;
  padding: 0.1rem 0.42rem;
  border: 1px solid var(--line-subtle);
  border-radius: 4px;
  background: var(--surface-0);
}

/* Tab navigation */
.chart-tabs {
  display: flex;
  gap: 0.25rem;
  border-bottom: 1px solid var(--line-strong);
  padding: 0 1.25rem;
  overflow-x: auto;
  scrollbar-width: none;
}
.chart-tabs::-webkit-scrollbar { display: none; }
.tab-btn {
  padding: 0.75rem 1.25rem;
  background: none;
  border: none;
  border-bottom: 2px solid transparent;
  color: var(--text-dim);
  font-size: var(--fs-xs);
  font-weight: 600;
  letter-spacing: 1px;
  cursor: pointer;
  white-space: nowrap;
  transition: all 0.3s;
}
.tab-btn:hover { color: var(--text); }
.tab-btn.active { color: var(--accent); border-bottom-color: var(--accent); }
.tab-content {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.overview-layout {
  padding-top: 0.15rem;
}

.overview-section {
  padding: 1.1rem 0 1.25rem;
}

.overview-section:first-child {
  padding-top: 0.35rem;
}

.overview-section + .overview-section {
  border-top: 1px solid var(--line-subtle);
}

.overview-section-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 0.9rem;
}

.overview-section .block-title {
  margin-bottom: 0.35rem;
  font-size: var(--fs-sm);
}

.overview-section .block-desc {
  max-width: 46rem;
  margin: 0;
  font-size: var(--fs-2xs);
  line-height: 1.7;
}

.overview-section-body {
  min-width: 0;
}

.overview-relations-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 1rem;
}

.relation-card {
  min-width: 0;
  border: 1px solid color-mix(in oklab, var(--text) 10%, transparent);
  border-radius: 10px;
  background:
    linear-gradient(180deg, color-mix(in oklab, var(--surface-0) 78%, transparent), color-mix(in oklab, var(--surface-2) 42%, transparent)),
    var(--surface-0);
  box-shadow: inset 0 1px 0 color-mix(in oklab, var(--text) 4%, transparent), var(--shadow-xs);
  overflow: hidden;
}

.relation-card .relations-section,
.relation-card .no-relations {
  border-bottom: 0;
}

.block-title {
  font-size: var(--fs-xs);
  font-weight: 700;
  color: var(--accent);
  letter-spacing: 1px;
  margin-bottom: 0.5rem;
}

.block-desc {
  display: block;
  font-size: var(--fs-2xs);
  color: var(--text-soft);
  line-height: 1.5;
  margin-top: -0.3rem;
  margin-bottom: 0.5rem;
  letter-spacing: 0.3px;
}

/* Ten gods */
.ten-gods-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 0.75rem;
}

.god-item {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  justify-content: space-between;
  min-height: 76px;
  padding: 0.72rem 0.78rem;
  background: color-mix(in oklab, var(--surface-2) 55%, transparent);
  border-radius: 8px;
  border: 1px solid var(--line-subtle);
  box-shadow: inset 0 1px 0 color-mix(in oklab, var(--text) 4%, transparent);
}

.god-pillar {
  font-size: var(--fs-2xs);
  color: var(--text-muted);
  margin-bottom: 0.35rem;
}

.god-name {
  font-size: var(--fs-sm);
  font-weight: 700;
  line-height: 1.2;
  color: var(--crimson);
}

/* NaYin */
.nayin-list {
  display: flex;
  flex-wrap: wrap;
  gap: 0.65rem;
}

.nayin-tag {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  min-height: 38px;
  padding: 0.45rem 0.72rem;
  background: color-mix(in oklab, var(--accent-dim) 72%, transparent);
  border: 1px solid;
  border-radius: 8px;
  color: var(--text);
  cursor: pointer;
  line-height: 1.2;
  transition: background 0.2s, border-color 0.2s, transform 0.2s;
}

.nayin-tag:hover {
  background: var(--surface-2);
  transform: translateY(-1px);
}

.nayin-pillar {
  opacity: 0.6;
  font-size: var(--fs-2xs);
}

.nayin-name {
  font-weight: 600;
}

/* NaYin Detail Popover */
.nayin-detail {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  padding: 0.25rem 0;
}

.nayin-detail-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-bottom: 0.5rem;
  border-bottom: 1px solid var(--line-focus);
}

.nayin-detail-name {
  font-size: var(--fs-lg);
  font-weight: 700;
  color: var(--accent);
}

.nayin-detail-elem {
  font-size: var(--fs-2xs);
  padding: 0.15rem 0.5rem;
  border-radius: 3px;
  color: var(--bg);
  font-weight: 600;
}

.nayin-detail-section {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}

.nayin-detail-label {
  font-size: var(--fs-2xs);
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-dim);
}

.nayin-detail-value {
  font-size: var(--fs-sm);
  line-height: 1.5;
  color: var(--text);
}

.nayin-energy {
  color: var(--accent);
  font-weight: 600;
}

.nayin-detail-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.3rem;
  margin-top: 0.15rem;
}

.nayin-judgment-tag {
  font-size: var(--fs-xs);
  padding: 0.15rem 0.5rem;
  background: var(--line-strong);
  border: 1px solid var(--line-focus);
  border-radius: 3px;
  color: var(--accent);
}

/* DaYun — 时间轴风格 */
.dayun-timeline {
  display: flex;
  flex-wrap: nowrap;
  gap: 0.25rem;
  overflow-x: auto;
  padding: 0.2rem 0.1rem 0.75rem;
  scrollbar-width: thin;
}

.dayun-node {
  display: flex;
  flex-direction: column;
  align-items: center;
  min-width: 76px;
  flex: 0 0 76px;
}

.dayun-age {
  font-size: var(--fs-2xs);
  color: var(--text-dim);
  margin-bottom: 0.3rem;
  white-space: nowrap;
}

.dayun-dot-line {
  display: flex;
  align-items: center;
  width: 100%;
  position: relative;
}

.dayun-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--accent);
  box-shadow: 0 0 8px var(--accent-glow);
  flex-shrink: 0;
  margin: 0 auto;
  position: relative;
  z-index: 1;
}

.dayun-line {
  position: absolute;
  top: 50%;
  left: 50%;
  width: 100%;
  height: 1px;
  background: linear-gradient(90deg, var(--line-focus), var(--line-subtle));
  transform: translateY(-50%);
}

.dayun-pillar-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0;
  min-width: 48px;
  margin-top: 0.5rem;
  padding: 0.48rem 0.58rem;
  background: color-mix(in oklab, var(--accent-dim) 76%, transparent);
  border: 1px solid var(--line-strong);
  border-radius: 8px;
}

.dayun-gan {
  font-family: var(--font-serif), serif;
  font-size: var(--fs-sm);
  font-weight: 700;
  color: var(--accent);
  line-height: 1.2;
}

.dayun-zhi {
  font-family: var(--font-serif), serif;
  font-size: var(--fs-sm);
  font-weight: 600;
  color: var(--text-muted);
  line-height: 1.2;
}

.dayun-detail-tab {
  gap: 0.95rem;
}

.dayun-overview-card,
.dayun-current-card,
.dayun-stage-block,
.dayun-flow-panel {
  border: 1px solid var(--line-subtle);
  border-radius: 10px;
  background: color-mix(in oklab, var(--surface-2) 42%, transparent);
  padding: 1rem;
}

.dayun-summary-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 0.75rem;
  margin-top: 0.9rem;
}

.dayun-summary-item {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  min-width: 0;
  padding: 0.8rem;
  border: 1px solid var(--line-subtle);
  border-radius: 8px;
  background: var(--surface-0);
}

.dayun-summary-label {
  color: var(--text-soft);
  font-size: var(--fs-2xs);
}

.dayun-summary-item strong {
  min-width: 0;
  color: var(--text);
  font-size: var(--fs-sm);
  line-height: 1.25;
  overflow-wrap: anywhere;
}

.dayun-method-notes {
  display: flex;
  flex-wrap: wrap;
  gap: 0.45rem;
  margin-top: 0.85rem;
}

.dayun-method-notes span,
.dayun-mini-chip,
.dayun-evidence-chip {
  display: inline-flex;
  align-items: center;
  min-height: 24px;
  padding: 0.14rem 0.5rem;
  border: 1px solid var(--line-subtle);
  border-radius: 4px;
  background: var(--surface-0);
  color: var(--text-muted);
  font-size: var(--fs-2xs);
  line-height: 1.25;
}

.dayun-current-head {
  display: flex;
  justify-content: space-between;
  gap: 0.9rem;
  align-items: flex-start;
  margin-bottom: 0.9rem;
}

.dayun-current-score {
  display: inline-flex;
  flex-shrink: 0;
  align-items: center;
  min-height: 28px;
  padding: 0.18rem 0.65rem;
  border-radius: 999px;
  border: 1px solid var(--line-subtle);
  font-size: var(--fs-2xs);
  font-weight: 700;
}

.dayun-current-score.score-good {
  color: #16a34a;
  background: rgba(22, 163, 74, 0.08);
  border-color: rgba(22, 163, 74, 0.18);
}

.dayun-current-score.score-watch {
  color: #dc2626;
  background: rgba(220, 38, 38, 0.08);
  border-color: rgba(220, 38, 38, 0.18);
}

.dayun-current-body {
  display: grid;
  grid-template-columns: 76px minmax(0, 1fr);
  gap: 0.9rem;
  align-items: stretch;
}

.dayun-current-pillar {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.05rem;
  border-radius: 9px;
  border: 1px solid var(--line-focus);
  background: var(--accent-dim);
  color: var(--accent);
  font-family: var(--font-serif), serif;
  font-size: var(--fs-3xl);
  font-weight: 800;
  line-height: 1.05;
}

.dayun-current-copy {
  min-width: 0;
}

.dayun-current-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem;
  margin-bottom: 0.55rem;
}

.dayun-current-copy p,
.dayun-stage-impact {
  margin: 0;
  color: var(--text-muted);
  font-size: var(--fs-xs);
  line-height: 1.65;
}

.dayun-evidence-list {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem;
  margin-top: 0.65rem;
}

.dayun-stage-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.8rem;
  margin-top: 0.9rem;
}

.dayun-stage-card {
  min-width: 0;
  padding: 0.85rem;
  border: 1px solid var(--line-subtle);
  border-left: 3px solid var(--line-focus);
  border-radius: 9px;
  background: var(--surface-0);
  transition: border-color 0.2s, background 0.2s, transform 0.2s;
}

.dayun-stage-card:hover {
  border-color: var(--line-strong);
  transform: translateY(-1px);
}

.dayun-stage-card.is-current {
  border-color: var(--line-focus);
  background: color-mix(in oklab, var(--accent-dim) 58%, var(--surface-0));
}

.dayun-stage-card.flow-增强 {
  border-left-color: rgba(22, 163, 74, 0.52);
}

.dayun-stage-card.flow-减弱 {
  border-left-color: rgba(220, 38, 38, 0.52);
}

.dayun-stage-card.flow-泄秀 {
  border-left-color: rgba(37, 99, 235, 0.52);
}

.dayun-stage-top {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.35rem;
  color: var(--text-soft);
  font-size: var(--fs-2xs);
  margin-bottom: 0.7rem;
}

.dayun-now-badge {
  color: var(--accent);
  background: var(--accent-dim);
  border: 1px solid var(--line-focus);
  border-radius: 999px;
  padding: 0.08rem 0.42rem;
  font-weight: 700;
}

.dayun-stage-main {
  display: grid;
  grid-template-columns: 48px minmax(0, 1fr);
  gap: 0.7rem;
  align-items: center;
  margin-bottom: 0.7rem;
}

.dayun-stage-pillar {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 60px;
  border-radius: 8px;
  border: 1px solid var(--line-subtle);
  background: color-mix(in oklab, var(--surface-2) 70%, transparent);
  font-family: var(--font-serif), serif;
  font-weight: 800;
  line-height: 1.05;
}

.dayun-stage-gan {
  font-size: var(--fs-2xl);
}

.dayun-stage-zhi {
  font-size: var(--fs-lg);
}

.dayun-stage-info {
  min-width: 0;
}

.dayun-stage-title {
  color: var(--text);
  font-size: var(--fs-sm);
  font-weight: 800;
  margin-bottom: 0.45rem;
}

.dayun-stage-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.32rem;
}

.dayun-stage-tags span {
  display: inline-flex;
  align-items: center;
  min-height: 22px;
  padding: 0.1rem 0.42rem;
  border: 1px solid var(--line-subtle);
  border-radius: 4px;
  color: var(--text-muted);
  background: color-mix(in oklab, var(--surface-2) 62%, transparent);
  font-size: var(--fs-2xs);
  line-height: 1.2;
}

.dayun-flow-list.dayun-flow-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.65rem;
  margin-top: 0.9rem;
}

.fortune-detail-tab {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.85rem;
}

.fortune-detail-tab > .analysis-block {
  min-width: 0;
  padding: 0.9rem;
  border: 1px solid var(--line-subtle);
  border-radius: 10px;
  background: color-mix(in oklab, var(--surface-2) 36%, transparent);
}

.fortune-detail-tab > .analysis-block:first-child {
  grid-column: 1 / -1;
}

.fortune-detail-tab .fortune-layer-list {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.65rem;
}

.fortune-detail-tab .fortune-layer-row {
  min-width: 0;
  padding: 0.7rem;
  border: 1px solid var(--line-subtle);
  border-left: 3px solid var(--line-focus);
  border-radius: 8px;
  background: var(--surface-0);
}

.classical-rules-block {
  margin-bottom: 0.95rem;
}

/* Element Detail Table */
.element-detail-table {
  border: 1px solid var(--line-strong);
  border-radius: 8px;
  overflow: hidden;
}

.ed-header {
  display: grid;
  grid-template-columns: 1fr 1fr 1.5fr 1fr;
  padding: 0.4rem 0.75rem;
  background: rgba(255, 255, 255, 0.03);
  font-size: var(--fs-2xs);
  color: var(--text-muted);
  letter-spacing: 0.5px;
  border-bottom: 1px solid var(--line-subtle);
}

.ed-row {
  display: grid;
  grid-template-columns: 1fr 1fr 1.5fr 1fr;
  padding: 0.4rem 0.75rem;
  font-size: var(--fs-xs);
  border-bottom: 1px solid var(--line-subtle);
}

.ed-row:last-child {
  border-bottom: none;
}

.ed-elem {
  font-weight: 700;
}

.ed-tg,
.ed-zc,
.ed-total {
  color: var(--text);
}

.level-none { background: rgba(255, 255, 255, 0.01); }
.level-weak .ed-total { color: var(--text-dim); }
.level-medium .ed-total { color: var(--text-muted); }
.level-strong .ed-total { color: #c76f12; font-weight: 700; }
.level-very-strong .ed-total { color: #a6540a; font-weight: 800; }

/* Body Strength */
.body-strength {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.bs-verdict {
  display: inline-flex;
  align-self: flex-start;
  font-size: var(--fs-sm);
  font-weight: 700;
  color: var(--accent);
}

.bs-score {
  display: inline-flex;
  align-self: flex-start;
  min-height: 24px;
  align-items: center;
  padding: 0.12rem 0.55rem;
  border-radius: 4px;
  border: 1px solid var(--line-focus);
  background: var(--accent-dim);
  color: var(--accent);
  font-size: var(--fs-2xs);
  font-family: var(--font-mono);
}

.bs-tags {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.35rem;
}

.bs-like-label,
.bs-dislike-label {
  font-size: var(--fs-2xs);
  padding: 0.15rem 0.5rem;
  border-radius: 4px;
}

.bs-like-label {
  background: rgba(22, 163, 74, 0.1);
  color: #16a34a;
  border: 1px solid rgba(22, 163, 74, 0.2);
}

.bs-dislike-label {
  background: rgba(220, 38, 38, 0.1);
  color: #dc2626;
  border: 1px solid rgba(220, 38, 38, 0.2);
}

.bs-like {
  font-size: var(--fs-xs);
  padding: 0.15rem 0.5rem;
  background: rgba(22, 163, 74, 0.06);
  color: #16a34a;
  border: 1px solid rgba(22, 163, 74, 0.12);
  border-radius: 4px;
}

.bs-dislike {
  font-size: var(--fs-xs);
  padding: 0.15rem 0.5rem;
  background: rgba(220, 38, 38, 0.06);
  color: #dc2626;
  border: 1px solid rgba(220, 38, 38, 0.12);
  border-radius: 4px;
}

.body-strength-bars {
  margin-top: 0.25rem;
  padding-top: 0.65rem;
  border-top: 1px solid var(--line-subtle);
}

.body-strength-notes {
  margin-top: 0.25rem;
}

/* Pattern Analysis */
.pattern-detail {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.pattern-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.pattern-name {
  font-size: var(--fs-sm);
  font-weight: 700;
  color: var(--accent);
  letter-spacing: 0.05em;
}

.pattern-type-tag {
  display: inline-block;
  padding: 0.12rem 0.5rem;
  border-radius: 4px;
  font-size: var(--fs-2xs);
  font-weight: 600;
  letter-spacing: 0.5px;
}

.tag-special {
  background: rgba(186, 85, 211, 0.12);
  color: #ba55d3;
  border: 1px solid rgba(186, 85, 211, 0.22);
}

.tag-normal {
  background: rgba(100, 149, 237, 0.1);
  color: #6495ed;
  border: 1px solid rgba(100, 149, 237, 0.18);
}

.pattern-desc {
  font-size: var(--fs-xs);
  color: var(--text-muted);
  line-height: 1.65;
  margin: 0;
}

.pattern-elems {
  display: flex;
  flex-wrap: wrap;
  gap: 0.6rem;
}

.pattern-favor,
.pattern-avoid {
  display: flex;
  align-items: center;
  gap: 0.3rem;
  flex-wrap: wrap;
}

.pattern-elem-label {
  font-size: var(--fs-2xs);
  font-weight: 600;
  padding: 0.1rem 0.4rem;
  border-radius: 3px;
}

.pattern-favor .pattern-elem-label {
  background: rgba(22, 163, 74, 0.1);
  color: #16a34a;
  border: 1px solid rgba(22, 163, 74, 0.2);
}

.pattern-avoid .pattern-elem-label {
  background: rgba(220, 38, 38, 0.1);
  color: #dc2626;
  border: 1px solid rgba(220, 38, 38, 0.2);
}

.pattern-elem-tag {
  font-size: var(--fs-xs);
  padding: 0.12rem 0.45rem;
  border-radius: 3px;
}

.pattern-elem-tag.favor {
  background: rgba(22, 163, 74, 0.06);
  color: #16a34a;
  border: 1px solid rgba(22, 163, 74, 0.12);
}

.pattern-elem-tag.avoid {
  background: rgba(220, 38, 38, 0.06);
  color: #dc2626;
  border: 1px solid rgba(220, 38, 38, 0.12);
}

.ming-gong-detail {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.ming-gong-main {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.ming-gong-ganzhi {
  font-size: var(--fs-xl);
  font-weight: 700;
  color: var(--accent);
  letter-spacing: 0.15em;
}

.ming-gong-nayin {
  font-size: var(--fs-xs);
  color: var(--text-muted);
}

.ming-gong-shensha {
  display: flex;
  align-items: flex-start;
  gap: 0.5rem;
}

.shensha-badge {
  display: inline-block;
  padding: 0.12rem 0.45rem;
  border-radius: 3px;
  font-size: var(--fs-xs);
  font-weight: 600;
  white-space: nowrap;
  background: color-mix(in oklab, var(--text) 15%, transparent);
  color: var(--text);
}

.shensha-badge.shensha-ji {
  background: color-mix(in oklab, #b8860b 12%, transparent);
  color: var(--accent);
  border: 1px solid var(--line-focus);
}

.shensha-badge.shensha-xiong {
  background: color-mix(in oklab, var(--crimson) 10%, transparent);
  color: var(--crimson);
  border: 1px solid color-mix(in oklab, var(--crimson) 18%, transparent);
}

.shensha-desc {
  font-size: var(--fs-xs);
  color: var(--text-muted);
  line-height: 1.55;
}

.ming-gong-zhi {
  font-size: var(--fs-xs);
  color: var(--text-muted);
  line-height: 1.6;
  border-left: 2px solid var(--line-focus);
  padding-left: 0.6rem;
}

.tiaohou-source {
  font-size: var(--fs-2xs);
  color: var(--text-muted);
  font-weight: 400;
  margin-left: 0.4rem;
  font-style: italic;
}

.tiaohou-card {
  background: var(--accent-dim);
  border: 1px solid var(--line-focus);
  border-radius: 10px;
  padding: 1rem 1.1rem;
  margin-top: 0.4rem;
}

.tiaohou-header {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  margin-bottom: 0.7rem;
  flex-wrap: wrap;
}

.tiaohou-stem {
  font-size: var(--fs-2xl);
  font-weight: 700;
  color: var(--accent);
}

.tiaohou-arrow {
  color: var(--text-muted);
  font-size: var(--fs-sm);
}

.tiaohou-month {
  font-size: var(--fs-2xl);
  font-weight: 700;
  color: var(--text);
}

.tiaohou-divider {
  color: var(--text-soft);
  margin: 0 0.2rem;
}

.tiaohou-label {
  color: var(--text-soft);
  font-size: var(--fs-xs);
}

.tiaohou-primary {
  font-size: var(--fs-2xl);
  font-weight: 800;
}

.tiaohou-summary {
  font-size: var(--fs-sm);
  color: var(--text);
  line-height: 1.65;
  margin-bottom: 0.8rem;
}

.tiaohou-rules {
  display: flex;
  flex-direction: column;
  gap: 0.45rem;
}

.tiaohou-rule {
  display: flex;
  align-items: baseline;
  gap: 0.5rem;
  font-size: var(--fs-xs);
}

.tiaohou-xi {
  color: #16a34a;
  font-weight: 600;
  white-space: nowrap;
}

.tiaohou-ji {
  color: #dc2626;
  font-weight: 600;
  white-space: nowrap;
}

.tiaohou-reason {
  color: var(--text-muted);
  line-height: 1.5;
}

.ri-zhu-text,
.jin-bu-huan-text,
.season-text {
  font-size: var(--fs-sm);
  color: var(--text-muted);
  line-height: 1.7;
  white-space: pre-wrap;
  margin: 0;
}


.sheng-xiao-tag {
  font-size: var(--fs-2xs);
  color: var(--text-muted);
  letter-spacing: 1px;
}

.empties-tag {
  font-size: var(--fs-2xs);
  color: rgba(251, 113, 133, 0.5);
}

/* ShenSha list */
.shen-sha-list {
  display: flex;
  flex-direction: column;
  gap: 0.42rem;
}

.shen-sha-row {
  display: grid;
  grid-template-columns: 4.6rem auto 1fr;
  align-items: center;
  gap: 0.5rem;
  min-height: 2.1rem;
  padding: 0.35rem 0.65rem;
  background: rgba(255, 255, 255, 0.028);
  border: 1px solid var(--line-strong);
  border-radius: 7px;
  box-shadow: inset 0 1px 0 color-mix(in oklab, var(--text) 4%, transparent);
}

.shen-sha-name {
  font-family: var(--font-serif), serif;
  font-size: var(--fs-sm);
  font-weight: 700;
  color: var(--accent);
  letter-spacing: 0.06em;
  white-space: nowrap;
}

.shen-sha-target {
  min-width: 1.35rem;
  height: 1.35rem;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  background: var(--line-strong);
  border: 1px solid var(--line-focus);
  color: var(--text);
  font-size: var(--fs-2xs);
  font-weight: 700;
}

.shen-sha-desc {
  color: var(--text-muted);
  font-size: var(--fs-xs);
  line-height: 1.45;
}

/* ShenSha group title */
.shen-sha-group-title {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  font-family: var(--font-serif), serif;
  font-size: var(--fs-sm);
  font-weight: 700;
  color: var(--text);
  margin-bottom: 0.55rem;
  letter-spacing: 0.06em;
}

.shen-sha-group-dot {
  width: 0.42rem;
  height: 0.42rem;
  border-radius: 50%;
  flex-shrink: 0;
}

.shen-sha-group-role {
  font-family: var(--font-sans), sans-serif;
  font-size: var(--fs-2xs);
  font-weight: 400;
  color: var(--text-soft);
  letter-spacing: 0.03em;
}

/* ShenSha summary */
.shen-sha-summary-block {
  border-top: 1px solid var(--line-subtle);
  margin-top: 0.5rem;
  padding-top: 0.75rem;
}

.shen-sha-summary-title {
  font-size: var(--fs-xs);
  font-weight: 600;
  color: var(--text-muted);
  margin-bottom: 0.45rem;
}

.shen-sha-summary-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
}

.shen-sha-summary-list li {
  font-size: var(--fs-xs);
  color: var(--text-muted);
  line-height: 1.5;
  padding-left: 0.65rem;
  position: relative;
}

.shen-sha-summary-list li::before {
  content: '–';
  position: absolute;
  left: 0;
  color: var(--text-soft);
}

.ten-god-chart-wrap {
  border: 1px solid var(--line-strong);
  border-radius: 10px;
  overflow: hidden;
  background:
    linear-gradient(180deg, var(--line-subtle) 0%, transparent 100%),
    rgba(255, 255, 255, 0.015);
  padding: 0.75rem 0.5rem 0.5rem;
  box-shadow: inset 0 1px 0 var(--line-subtle);
}

.ten-god-chart {
  width: 100%;
  height: 220px;
}
.ten-god-tabs {
  --el-color-primary: #cbd5e1;
}
.ten-god-tabs .el-tabs__item {
  color: var(--text-muted);
  font-size: var(--fs-xs);
}
.ten-god-tabs .el-tabs__item.is-active {
  color: var(--accent);
}
.ten-god-tabs .el-tabs__nav-wrap::after {
  background: var(--accent-dim);
}
.tg-summary {
  background: var(--accent-dim);
  border: 1px solid var(--line-strong);
  border-radius: 10px;
  padding: 1rem 1.25rem;
  font-size: var(--fs-sm);
  line-height: 1.8;
  color: var(--text);
}
.tg-text {
  font-size: var(--fs-sm);
  line-height: 1.8;
  color: var(--text);
}
.tg-god-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.tg-god-card {
  background: var(--accent-dim);
  border: 1px solid var(--line-strong);
  border-radius: 10px;
  padding: 0.875rem 1rem;
}
.tg-god-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 6px;
}
.tg-god-name {
  font-weight: 600;
  font-size: var(--fs-sm);
  color: var(--accent);
}
.tg-god-pct {
  font-size: var(--fs-xs);
  color: var(--text-muted);
}
.tg-god-meaning {
  font-size: var(--fs-xs);
  line-height: 1.6;
  color: var(--text);
  margin-bottom: 4px;
}
.tg-god-advice {
  font-size: var(--fs-2xs);
  line-height: 1.6;
  color: var(--text);
}

/* WuXingFlow 五行流通 */
.wuxing-flow-card {
  display: flex; flex-direction: column; gap: 0.5rem;
  padding: 0.75rem;
  background: color-mix(in oklab, var(--text) 2%, transparent);
  border-radius: 8px;
  border: 1px solid var(--line-subtle);
}
.wf-header { display: flex; align-items: center; gap: 0.5rem; flex-wrap: wrap; }
.wf-label { font-size: var(--fs-2xs); color: var(--text-muted); }
.wf-elem { font-size: var(--fs-sm); font-weight: 700; }
.wf-type { font-size: var(--fs-xs); color: var(--text-muted); padding: 0.1rem 0.4rem; background: var(--accent-dim); border-radius: 4px; }
.wf-status { font-size: var(--fs-2xs); font-weight: 600; padding: 0.1rem 0.5rem; border-radius: 4px; }
.wf-smooth { color: #16a34a; background: rgba(22,163,74,0.1); }
.wf-blocked { color: #dc2626; background: rgba(220,38,38,0.1); }
.wf-paths { display: flex; flex-direction: column; gap: 0.25rem; }
.wf-path { display: flex; align-items: center; gap: 0.4rem; font-size: var(--fs-xs); color: var(--text-muted); }
.wf-path-dot { width: 4px; height: 4px; border-radius: 50%; background: var(--accent); flex-shrink: 0; }
.wf-blocked-row { display: flex; align-items: center; gap: 0.4rem; font-size: var(--fs-xs); }
.wf-blocked-label { color: #dc2626; font-size: var(--fs-2xs); }
.wf-blocked-elem { font-weight: 600; }
.wf-verdict { font-size: var(--fs-xs); color: var(--text-muted); font-style: italic; }
.wf-advice { font-size: var(--fs-xs); color: var(--accent); opacity: 0.8; }

/* TongGuan 通关用神 */
.tong-guan-card {
  display: flex; align-items: center; gap: 0.75rem;
  padding: 0.75rem;
  background: color-mix(in oklab, var(--text) 2%, transparent);
  border-radius: 8px;
  border: 1px solid var(--line-subtle);
}
.tg-elem { font-size: var(--fs-2xl); font-weight: 900; font-family: var(--font-serif); }
.tg-desc { font-size: var(--fs-xs); color: var(--text-muted); line-height: 1.5; }

/* MissingElements 缺失五行 */
.missing-elem-card {
  display: flex; flex-direction: column; gap: 0.4rem;
  padding: 0.75rem;
  background: color-mix(in oklab, var(--text) 2%, transparent);
  border-radius: 8px;
  border: 1px solid var(--line-subtle);
}
.me-row { display: flex; align-items: center; gap: 0.4rem; flex-wrap: wrap; }
.me-label { font-size: var(--fs-2xs); color: var(--text-muted); min-width: 28px; }
.me-tag { font-size: var(--fs-xs); font-weight: 700; padding: 0.1rem 0.4rem; border: 1px solid; border-radius: 4px; }
.me-missing { background: rgba(220,38,38,0.06); }
.me-remedy { background: rgba(22,163,74,0.06); }
.me-severity { font-size: var(--fs-2xs); padding: 0.1rem 0.4rem; border-radius: 4px; margin-left: 0.25rem; }
.me-轻微 { color: var(--accent); background: var(--line-strong); }
.me-中等 { color: #dc2626; background: rgba(220,38,38,0.1); }
.me-严重 { color: #dc2626; background: rgba(220,38,38,0.2); }

/* DaYunFlow 流年运势 */
.dayun-flow-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.dayun-flow-item {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  padding: 0.55rem 0.7rem;
  background: rgba(255, 255, 255, 0.02);
  border-radius: 8px;
  border: 1px solid var(--line-subtle);
  transition: all 0.25s ease;
}

.dayun-flow-item:hover {
  background: rgba(255, 255, 255, 0.04);
  border-color: var(--line-strong);
}

.df-left {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.15rem;
  min-width: 44px;
  flex-shrink: 0;
}

.df-age {
  font-size: var(--fs-2xs);
  font-weight: 600;
  color: var(--text-dim);
  white-space: nowrap;
}

.df-pillar {
  font-family: var(--font-serif), serif;
  font-size: var(--fs-sm);
  font-weight: 700;
  color: var(--accent);
  letter-spacing: 1px;
}

.df-right {
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
  min-width: 0;
  flex: 1;
}

.df-change-badge {
  display: inline-block;
  width: fit-content;
  font-size: var(--fs-2xs);
  font-weight: 600;
  padding: 0.1rem 0.5rem;
  border-radius: 3px;
  letter-spacing: 0.5px;
}

.dfb-增强 { color: #16a34a; background: rgba(22, 163, 74, 0.1); border: 1px solid rgba(22, 163, 74, 0.18); }
.dfb-减弱 { color: #dc2626; background: rgba(220, 38, 38, 0.1); border: 1px solid rgba(220, 38, 38, 0.18); }
.dfb-泄秀 { color: #2563eb; background: rgba(37, 99, 235, 0.1); border: 1px solid rgba(37, 99, 235, 0.18); }
.dfb-不变 { color: var(--text-dim); background: var(--surface-2); border: 1px solid var(--line-subtle); }

.df-增强 { border-left: 3px solid rgba(22, 163, 74, 0.4); }
.df-减弱 { border-left: 3px solid rgba(220, 38, 38, 0.4); }
.df-泄秀 { border-left: 3px solid rgba(37, 99, 235, 0.4); }
.df-不变 { border-left: 3px solid var(--line-focus); }

.df-impact {
  font-size: var(--fs-xs);
  color: var(--text-muted);
  line-height: 1.5;
}

/* ===== Dark mode overrides — keep relation colors readable without neon glare ===== */
:global(.dark) .overview-relations {
  --rel-he: #6fcf97;
  --rel-ke: #e58a8a;
  --rel-sheng: #8fb9ee;
  --rel-hai: #d8a56c;
  --rel-xing: #b9a0dd;
}

:global(.dark) .bs-like-label { background: rgba(74, 222, 128, 0.1); color: #4ade80; border-color: rgba(74, 222, 128, 0.2); }
:global(.dark) .bs-dislike-label { background: rgba(251, 113, 133, 0.1); color: #fb7185; border-color: rgba(251, 113, 133, 0.2); }
:global(.dark) .bs-like { background: rgba(74, 222, 128, 0.06); color: #4ade80; border-color: rgba(74, 222, 128, 0.12); }
:global(.dark) .bs-dislike { background: rgba(251, 113, 133, 0.06); color: #fb7185; border-color: rgba(251, 113, 133, 0.12); }

:global(.dark) .pattern-favor .pattern-elem-label { background: rgba(74, 222, 128, 0.1); color: #4ade80; border-color: rgba(74, 222, 128, 0.2); }
:global(.dark) .pattern-avoid .pattern-elem-label { background: rgba(251, 113, 133, 0.1); color: #fb7185; border-color: rgba(251, 113, 133, 0.2); }
:global(.dark) .pattern-elem-tag.favor { background: rgba(74, 222, 128, 0.06); color: #4ade80; border-color: rgba(74, 222, 128, 0.12); }
:global(.dark) .pattern-elem-tag.avoid { background: rgba(251, 113, 133, 0.06); color: #fb7185; border-color: rgba(251, 113, 133, 0.12); }

:global(.dark) .wf-smooth { color: #4ade80; background: rgba(74, 222, 128, 0.1); }
:global(.dark) .wf-blocked { color: #fb7185; background: rgba(251, 113, 133, 0.1); }
:global(.dark) .wf-blocked-label { color: #fb7185; }

:global(.dark) .dfb-增强 { color: #4ade80; background: rgba(74, 222, 128, 0.1); border-color: rgba(74, 222, 128, 0.18); }
:global(.dark) .dfb-减弱 { color: #f87171; background: rgba(248, 113, 113, 0.1); border-color: rgba(248, 113, 113, 0.18); }
:global(.dark) .dfb-泄秀 { color: #60a5fa; background: rgba(96, 165, 250, 0.1); border-color: rgba(96, 165, 250, 0.18); }
:global(.dark) .df-增强 { border-left-color: rgba(74, 222, 128, 0.4); }
:global(.dark) .df-减弱 { border-left-color: rgba(248, 113, 113, 0.4); }
:global(.dark) .df-泄秀 { border-left-color: rgba(96, 165, 250, 0.4); }

:global(.dark) .me-missing { background: rgba(251, 113, 133, 0.06); }
:global(.dark) .me-remedy { background: rgba(74, 222, 128, 0.06); }
:global(.dark) .me-中等 { color: #fb7185; background: rgba(251, 113, 133, 0.1); }
:global(.dark) .me-严重 { color: #fb7185; background: rgba(251, 113, 133, 0.2); }

:global(.dark) .god-name { color: #fb7185; }

:global(.dark) .level-strong .ed-total { color: #f59e0b; }
:global(.dark) .level-very-strong .ed-total { color: #f97316; }

:global(.dark) .bento-day-card {
  background:
    linear-gradient(180deg, rgba(var(--jade-accent-rgb), 0.12), transparent 150%),
    var(--glass-bg);
}

:global(.dark) .shensha-badge.shensha-xiong { color: #f87171; }

:global(.dark) .tiaohou-xi { color: #34d399; }
:global(.dark) .tiaohou-ji { color: #fb7185; }

:global(.dark) .shen-sha-detail-row.sha-good,
:global(.dark) .fortune-layer-chip.sha-good {
  border-color: rgba(74, 222, 128, 0.35);
  color: #4ade80;
}

:global(.dark) .shen-sha-detail-row.sha-risk,
:global(.dark) .fortune-layer-chip.sha-risk {
  border-color: rgba(248, 113, 113, 0.35);
  color: #f87171;
}

@media (max-width: 980px) {
  .pillars-bento {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .bento-radar {
    grid-column: 1 / -1;
    min-height: 184px;
  }

  .bento-radar-chart {
    min-height: 148px;
    height: 148px;
  }
}

@media (max-width: 860px) {
  .pillars-bento {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .bento-radar {
    grid-column: 1 / -1;
  }

  .ten-gods-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .overview-relations-grid,
  .dayun-summary-grid,
  .dayun-stage-grid,
  .dayun-flow-list.dayun-flow-grid,
  .fortune-detail-tab,
  .fortune-detail-tab .fortune-layer-list {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .shensha-grid {
    grid-template-columns: 1fr;
  }

  .rule-table-list {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 560px) {
  .pillars-bento {
    grid-template-columns: 1fr;
  }

  .pillars-bento > .bento-card {
    min-height: auto;
  }

  .bento-radar {
    grid-column: auto;
    min-height: 178px;
  }

  .bento-radar-chart {
    min-height: 142px;
    height: 142px;
  }

  .overview-section {
    padding: 0.95rem 0 1.05rem;
  }

  .overview-section-head {
    margin-bottom: 0.7rem;
  }

  .shensha-overview-head {
    flex-direction: column;
  }

  .shensha-overview-count {
    align-self: flex-start;
  }

  .shensha-grid {
    grid-template-columns: 1fr;
  }

  .ten-gods-grid {
    gap: 0.55rem;
  }

  .god-item {
    min-height: 68px;
    padding: 0.65rem;
  }

  .nayin-list {
    display: grid;
    grid-template-columns: 1fr;
    gap: 0.5rem;
  }

  .nayin-tag {
    width: 100%;
    justify-content: space-between;
  }

  .dayun-node {
    min-width: 68px;
    flex-basis: 68px;
  }

  .dayun-pillar-card {
    min-width: 44px;
    padding: 0.42rem 0.52rem;
  }

  .dayun-overview-card,
  .dayun-current-card,
  .dayun-stage-block,
  .dayun-flow-panel {
    padding: 0.8rem;
  }

  .overview-relations-grid,
  .dayun-summary-grid,
  .dayun-stage-grid,
  .dayun-flow-list.dayun-flow-grid,
  .fortune-detail-tab,
  .fortune-detail-tab .fortune-layer-list {
    grid-template-columns: 1fr;
    gap: 0.65rem;
  }

  .relations-section {
    padding: 0.82rem;
  }

  .gz-item {
    grid-template-columns: minmax(0, 1fr) max-content;
    align-items: start;
    gap: 0.42rem 0.55rem;
    padding: 0.58rem 0.62rem;
  }

  .gz-chars {
    min-width: 0;
  }

  .gz-text {
    grid-column: 1 / -1;
  }

  .dayun-current-head,
  .dayun-current-body {
    display: flex;
    flex-direction: column;
  }

  .dayun-current-score {
    align-self: flex-start;
  }

  .dayun-current-pillar {
    min-height: 72px;
    flex-direction: row;
    gap: 0.25rem;
  }

  .fortune-layer-top,
  .rule-meta-row,
  .rule-table-head {
    align-items: flex-start;
    flex-direction: column;
    gap: 0.35rem;
  }

  .rule-meta-value {
    text-align: left;
  }

  .evidence-bar-row {
    grid-template-columns: 2.8rem minmax(72px, 1fr) 2.2rem;
    gap: 0.38rem;
  }
}
</style>

<style>
/* NaYin popover — dark theme to match app */
.nayin-popover {
  background: var(--surface-1) !important;
  border: 1px solid var(--line-strong) !important;
  border-radius: 10px !important;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.6), 0 0 24px var(--accent-dim) !important;
  padding: 14px 16px !important;
  color: var(--text);
}
.nayin-popover .el-popover__title {
  color: var(--text);
}
</style>
