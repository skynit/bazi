<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { use } from 'echarts/core'
import { BarChart } from 'echarts/charts'
import { GridComponent, TooltipComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import VChart from 'vue-echarts'

const props = defineProps<{
  chart: {
    id?: number
    year_pillar: { gan: string; zhi: string }
    month_pillar: { gan: string; zhi: string }
    day_pillar: { gan: string; zhi: string }
    hour_pillar: { gan: string; zhi: string }
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
  戊: { name: '土', elemColor: '#fde68a' },
  己: { name: '土', elemColor: '#fde68a' },
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
  辰: { name: '土', elemColor: '#fde68a' },
  戌: { name: '土', elemColor: '#fde68a' },
  丑: { name: '土', elemColor: '#fde68a' },
  未: { name: '土', elemColor: '#fde68a' },
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

const elemColor = (e: string) =>
  ({ 金: '#cbd5e1', 木: '#34d399', 水: '#22d3ee', 火: '#fb7185', 土: '#fde68a' })[e] || '#8a9a8e'

const pillarLabel = (k: string) => ({ year: '年柱', month: '月柱', day: '日柱', hour: '时柱' }[k] || k)

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

const pillarShenShaColor = (p: string) =>
  ({ day: 'var(--accent)', year: '#5BA4CF', month: '#60B89A', hour: '#A182CF' }[p] || '#888')

const pillarShenShaBg = (p: string) =>
  ({ day: 'rgba(203,213,225,0.07)', year: 'rgba(91,164,207,0.06)', month: 'rgba(96,184,154,0.06)', hour: 'rgba(161,130,207,0.06)' }[p] || 'rgba(255,255,255,0.02)')

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
  const mutedColor = cssVar('--text-muted', '#5a6a5e')
  const softColor = cssVar('--text-soft', 'rgba(15, 23, 18, 0.44)')
  const lineColor = cssVar('--line-subtle', 'rgba(15, 23, 18, 0.06)')
  const tooltipBg = cssVar('--surface-1', '#ffffff')
  const fe = props.chart.five_elements
  if (!fe) return null
  const total = Object.values(fe as Record<string, number>).reduce((s, v) => s + v, 0)
  if (total === 0) return null

  // 五行配色 — 科技色盘
  const barColors = ['#34d399', '#fb7185', '#fde68a', '#cbd5e1', '#22d3ee']
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
const chartTabs = [
  { key: 'overview', label: '命盘总览' },
  { key: 'wuxing', label: '五行格局' },
  { key: 'shishen', label: '十神详解' },
  { key: 'pattern', label: '格局古籍' },
  { key: 'shensha', label: '神煞' },
  { key: 'fortune', label: '运势详批' },
]

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

      <!-- Four pillars bento grid -->
      <div class="pillars-bento">
        <!-- 年柱卡片 -->
        <div
          class="bento-card bento-small"
          :class="'bento-hover-' + ganElement[pillars[0].gan]?.name"
        >
          <div class="bento-label">{{ pillars[0].label }}</div>
          <div class="bento-body">
            <div class="bento-gan" :style="{ color: ganElement[pillars[0].gan]?.elemColor }">
              <span class="bento-char">{{ pillars[0].gan }}</span>
              <span
                class="elem-tag"
                :style="{
                  background: ganElement[pillars[0].gan]?.elemColor + '22',
                  color: ganElement[pillars[0].gan]?.elemColor,
                  borderColor: ganElement[pillars[0].gan]?.elemColor + '44',
                }"
              >{{ ganElement[pillars[0].gan]?.name }}</span>
            </div>
            <div class="bento-zhi" :style="{ color: zhiElement[pillars[0].zhi]?.elemColor }">
              <span class="bento-char">{{ pillars[0].zhi }}</span>
              <span
                class="elem-tag"
                :style="{
                  background: zhiElement[pillars[0].zhi]?.elemColor + '22',
                  color: zhiElement[pillars[0].zhi]?.elemColor,
                  borderColor: zhiElement[pillars[0].zhi]?.elemColor + '44',
                }"
              >{{ zhiElement[pillars[0].zhi]?.name }}</span>
            </div>
            <span v-if="chart.ten_gods?.year" class="bento-god-tag">{{ chart.ten_gods.year }}</span>
          </div>
          <div v-if="pillarDetails[0]" class="bento-sub">
            <span class="sheng-xiao-tag">{{ pillarDetails[0].sheng_xiao }}</span>
            <span v-if="pillarDetails[0].empties[0]" class="empties-tag">
              空{{ pillarDetails[0].empties[0] }}{{ pillarDetails[0].empties[1] }}
            </span>
          </div>
        </div>

        <!-- 月柱卡片 -->
        <div
          class="bento-card bento-small"
          :class="'bento-hover-' + ganElement[pillars[1].gan]?.name"
        >
          <div class="bento-label">{{ pillars[1].label }}</div>
          <div class="bento-body">
            <div class="bento-gan" :style="{ color: ganElement[pillars[1].gan]?.elemColor }">
              <span class="bento-char">{{ pillars[1].gan }}</span>
              <span
                class="elem-tag"
                :style="{
                  background: ganElement[pillars[1].gan]?.elemColor + '22',
                  color: ganElement[pillars[1].gan]?.elemColor,
                  borderColor: ganElement[pillars[1].gan]?.elemColor + '44',
                }"
              >{{ ganElement[pillars[1].gan]?.name }}</span>
            </div>
            <div class="bento-zhi" :style="{ color: zhiElement[pillars[1].zhi]?.elemColor }">
              <span class="bento-char">{{ pillars[1].zhi }}</span>
              <span
                class="elem-tag"
                :style="{
                  background: zhiElement[pillars[1].zhi]?.elemColor + '22',
                  color: zhiElement[pillars[1].zhi]?.elemColor,
                  borderColor: zhiElement[pillars[1].zhi]?.elemColor + '44',
                }"
              >{{ zhiElement[pillars[1].zhi]?.name }}</span>
            </div>
            <span v-if="chart.ten_gods?.month" class="bento-god-tag">{{ chart.ten_gods.month }}</span>
          </div>
          <div v-if="pillarDetails[1]" class="bento-sub">
            <span class="sheng-xiao-tag">{{ pillarDetails[1].sheng_xiao }}</span>
            <span v-if="pillarDetails[1].empties[0]" class="empties-tag">
              空{{ pillarDetails[1].empties[0] }}{{ pillarDetails[1].empties[1] }}
            </span>
          </div>
        </div>

        <!-- 日柱主卡片 (col-span-2, 流光边框) -->
        <div class="bento-card bento-day-wrapper">
          <div class="bento-day-beam" aria-hidden="true"></div>
          <div
            class="bento-day-inner"
            :class="'bento-hover-' + ganElement[pillars[2].gan]?.name"
          >
            <div class="bento-label bento-day-label">{{ pillars[2].label }}</div>
            <div class="bento-day-body">
              <div class="bento-day-gan" :style="{ color: ganElement[pillars[2].gan]?.elemColor }">
                <span class="bento-day-char">{{ pillars[2].gan }}</span>
                <span
                  class="elem-tag elem-tag-lg"
                  :style="{
                    background: ganElement[pillars[2].gan]?.elemColor + '22',
                    color: ganElement[pillars[2].gan]?.elemColor,
                    borderColor: ganElement[pillars[2].gan]?.elemColor + '44',
                  }"
                >{{ ganElement[pillars[2].gan]?.name }}</span>
              </div>
              <div class="bento-day-divider"></div>
              <div class="bento-day-zhi" :style="{ color: zhiElement[pillars[2].zhi]?.elemColor }">
                <span class="bento-day-char">{{ pillars[2].zhi }}</span>
                <span
                  class="elem-tag elem-tag-lg"
                  :style="{
                    background: zhiElement[pillars[2].zhi]?.elemColor + '22',
                    color: zhiElement[pillars[2].zhi]?.elemColor,
                    borderColor: zhiElement[pillars[2].zhi]?.elemColor + '44',
                  }"
                >{{ zhiElement[pillars[2].zhi]?.name }}</span>
              </div>
              <span v-if="chart.ten_gods?.day" class="bento-god-tag bento-god-tag-day">{{ chart.ten_gods.day }}</span>
            </div>
            <div v-if="pillarDetails[2]" class="bento-sub bento-day-sub">
              <span class="sheng-xiao-tag">{{ pillarDetails[2].sheng_xiao }}</span>
              <span v-if="pillarDetails[2].empties[0]" class="empties-tag">
                空{{ pillarDetails[2].empties[0] }}{{ pillarDetails[2].empties[1] }}
              </span>
            </div>
          </div>
        </div>

        <!-- 时柱卡片 -->
        <div
          class="bento-card bento-small"
          :class="'bento-hover-' + ganElement[pillars[3].gan]?.name"
        >
          <div class="bento-label">{{ pillars[3].label }}</div>
          <div class="bento-body">
            <div class="bento-gan" :style="{ color: ganElement[pillars[3].gan]?.elemColor }">
              <span class="bento-char">{{ pillars[3].gan }}</span>
              <span
                class="elem-tag"
                :style="{
                  background: ganElement[pillars[3].gan]?.elemColor + '22',
                  color: ganElement[pillars[3].gan]?.elemColor,
                  borderColor: ganElement[pillars[3].gan]?.elemColor + '44',
                }"
              >{{ ganElement[pillars[3].gan]?.name }}</span>
            </div>
            <div class="bento-zhi" :style="{ color: zhiElement[pillars[3].zhi]?.elemColor }">
              <span class="bento-char">{{ pillars[3].zhi }}</span>
              <span
                class="elem-tag"
                :style="{
                  background: zhiElement[pillars[3].zhi]?.elemColor + '22',
                  color: zhiElement[pillars[3].zhi]?.elemColor,
                  borderColor: zhiElement[pillars[3].zhi]?.elemColor + '44',
                }"
              >{{ zhiElement[pillars[3].zhi]?.name }}</span>
            </div>
            <span v-if="chart.ten_gods?.hour" class="bento-god-tag">{{ chart.ten_gods.hour }}</span>
          </div>
          <div v-if="pillarDetails[3]" class="bento-sub">
            <span class="sheng-xiao-tag">{{ pillarDetails[3].sheng_xiao }}</span>
            <span v-if="pillarDetails[3].empties[0]" class="empties-tag">
              空{{ pillarDetails[3].empties[0] }}{{ pillarDetails[3].empties[1] }}
            </span>
          </div>
        </div>

        <!-- 五行雷达图 (中) -->
        <div v-if="fiveElementsOption" class="bento-card bento-radar">
          <div class="bento-label">五行分布</div>
          <v-chart class="bento-radar-chart" :option="fiveElementsOption" autoresize />
        </div>
      </div>

      <!-- 天干地支综合分析 -->
      <div v-if="ganZhi" class="ganzhi-analysis">
        <!-- 天干关系 -->
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

        <!-- 地支关系 -->
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
        <div v-show="activeTab === 'overview'" class="tab-content">
          <!-- Five Elements chart moved to Bento Grid -->

          <!-- Ten Gods -->
          <div v-if="chart.ten_gods" class="analysis-block">
            <div class="block-title">十神</div>
            <span class="block-desc">天干对日主的生克关系，反映人际、性格与命运倾向</span>
            <div class="ten-gods-grid">
              <div v-for="(god, pillar) in chart.ten_gods" :key="pillar" class="god-item">
                <span class="god-pillar">{{ pillar }}</span>
                <span class="god-name">{{ god }}</span>
              </div>
            </div>
          </div>

          <!-- NaYin -->
          <div v-if="chart.na_yin" class="analysis-block">
            <div class="block-title">纳音</div>
            <span class="block-desc">六十甲子纳音五行取象，揭示命局的先天气质与能量场</span>
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

          <!-- DaYun -->
          <div v-if="chart.da_yun && chart.da_yun.start_age" class="analysis-block">
            <div class="block-title">
              大运 ({{ chart.da_yun.direction }} · {{ chart.da_yun.start_age }}岁起运)
            </div>
            <span class="block-desc">十年一步的大运流转，决定人生各阶段的运势起伏</span>
            <div class="dayun-timeline">
              <div v-for="(p, i) in chart.da_yun.pillars" :key="i" class="dayun-node">
                <div class="dayun-age">{{ (Number(chart.da_yun.start_age) || 0) + (Number(i) || 0) * 10 }}岁</div>
                <div class="dayun-dot-line">
                  <span class="dayun-dot"></span>
                  <span v-if="i < chart.da_yun.pillars.length - 1" class="dayun-line"></span>
                </div>
                <div class="dayun-pillar-card">
                  <span class="dayun-gan">{{ p.gan }}</span>
                  <span class="dayun-zhi">{{ p.zhi }}</span>
                </div>
              </div>
            </div>
          </div>
        </div><!-- /overview tab -->

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
          <div v-if="chart.body_strength && chart.pattern_analysis?.pattern_type !== '特殊格局'" class="analysis-block">
            <div class="block-title">身旺喜忌</div>
            <span class="block-desc">扶抑法判断日主强弱，身旺抑之、身弱扶之，确定喜用忌神</span>
            <div class="body-strength">
              <div class="bs-verdict">{{ chart.body_strength.verdict }}</div>
              <div class="bs-tags">
                <span class="bs-like-label">喜</span>
                <span v-for="l in chart.body_strength.like" :key="l" class="bs-like">{{ l }}</span>
                <span class="bs-dislike-label">忌</span>
                <span v-for="d in chart.body_strength.dislike" :key="d" class="bs-dislike">{{ d }}</span>
              </div>
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
          <!-- ShenSha (grouped by pillar when available) -->
          <template v-if="groupedShenSha">
            <template v-for="group in groupedShenSha" :key="group.pillar">
              <div v-if="group.items && group.items.length" class="analysis-block">
                <div class="shen-sha-group-title">
                  <span class="shen-sha-group-dot" :style="{ background: pillarShenShaColor(group.pillar) }"></span>
                  {{ group.label }}神煞
                  <span class="shen-sha-group-role">· {{ group.role }}</span>
                </div>
                <div class="shen-sha-list">
                  <article
                    v-for="sha in group.items"
                    :key="sha.name + sha.target + sha.desc"
                    class="shen-sha-row"
                    :style="{ background: pillarShenShaBg(group.pillar) }"
                  >
                    <span class="shen-sha-name" :style="{ color: pillarShenShaColor(group.pillar) }">{{ sha.name }}</span>
                    <span v-if="sha.target" class="shen-sha-target">{{ sha.target }}</span>
                    <span v-if="sha.desc" class="shen-sha-desc">{{ sha.desc }}</span>
                  </article>
                </div>
              </div>
            </template>
            <div v-if="globalShenSha.length" class="analysis-block">
              <div class="shen-sha-group-title">
                <span class="shen-sha-group-dot" style="background: var(--accent)"></span>
                全局组合神煞
                <span class="shen-sha-group-role">· 多柱配合</span>
              </div>
              <div class="shen-sha-list">
                <article
                  v-for="sha in globalShenSha"
                  :key="sha.name + sha.target + sha.desc"
                  class="shen-sha-row"
                  :style="{ background: 'rgba(203,213,225,0.07)' }"
                >
                  <span class="shen-sha-name" :style="{ color: 'var(--accent)' }">{{ sha.name }}</span>
                  <span v-if="sha.target" class="shen-sha-target">{{ sha.target }}</span>
                  <span v-if="sha.desc" class="shen-sha-desc">{{ sha.desc }}</span>
                </article>
              </div>
            </div>
            <div v-if="showSummary" class="analysis-block shen-sha-summary-block">
              <div class="shen-sha-summary-title">{{ props.chart.shen_sha_summary.title }}</div>
              <ul class="shen-sha-summary-list">
                <li v-for="line in props.chart.shen_sha_summary.description" :key="line.slice(0, 16)">
                  {{ line }}
                </li>
              </ul>
            </div>
          </template>
          <div v-else-if="parsedDayShenSha.length" class="analysis-block">
            <div class="block-title">日柱神煞</div>
            <span class="block-desc">日柱天干地支所临神煞，揭示先天吉凶信息</span>
            <div class="shen-sha-list">
              <article v-for="sha in parsedDayShenSha" :key="sha.name + sha.target + sha.desc" class="shen-sha-row">
                <span class="shen-sha-name">{{ sha.name }}</span>
                <span v-if="sha.target" class="shen-sha-target">{{ sha.target }}</span>
                <span v-if="sha.desc" class="shen-sha-desc">{{ sha.desc }}</span>
              </article>
            </div>
          </div>
        </div><!-- /shensha tab -->

        <!-- ═══ Tab: 运势详批 (fortune) ═══ -->
        <div v-show="activeTab === 'fortune'" class="tab-content">
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

          <!-- DaYunFlow 大运流年 -->
          <div v-if="chart.dayun_flow && chart.dayun_flow.length" class="analysis-block">
            <div class="block-title">流年运势</div>
            <span class="block-desc">大运与流年交替的运势变化，关注关键转折节点</span>
            <div class="dayun-flow-list">
              <div v-for="(item, di) in chart.dayun_flow" :key="di" class="dayun-flow-item" :class="'df-' + item.flow_change">
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
        </div><!-- /fortune tab -->

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

/* Header */
.chart-header {
  background: linear-gradient(180deg, rgba(203,213,225,0.04), transparent);
  border-bottom: 1px solid rgba(255,255,255,0.06);
  padding: 1rem 1.25rem;
  text-align: center;
}

.header-eyebrow {
  font-size: 10px;
  letter-spacing: 3px;
  color: var(--text-soft);
  text-transform: uppercase;
  margin-bottom: 4px;
}

.chart-title {
  font-family: var(--font-serif), serif;
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--text);
  margin: 0;
  letter-spacing: 3px;
}

/* Pillars bento grid */
.pillars-bento {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.5rem;
  padding: 0.75rem;
  border-bottom: 1px solid rgba(255,255,255,0.06);
}

.bento-card {
  background: rgba(255,255,255,0.02);
  border: 1px solid rgba(255,255,255,0.06);
  border-radius: 10px;
  padding: 0.6rem 0.75rem;
  transition: box-shadow 0.3s, background 0.3s;
  position: relative;
  overflow: hidden;
}

.bento-card.bento-small:hover,
.bento-card.bento-day-inner:hover {
  background: rgba(255,255,255,0.04);
}

/* Hover 五行微光 */
.bento-hover-木:hover { box-shadow: inset 0 0 30px rgba(52,211,153,0.07); }
.bento-hover-火:hover { box-shadow: inset 0 0 30px rgba(251,113,133,0.07); }
.bento-hover-土:hover { box-shadow: inset 0 0 30px rgba(253,230,138,0.07); }
.bento-hover-金:hover { box-shadow: inset 0 0 30px rgba(203,213,225,0.07); }
.bento-hover-水:hover { box-shadow: inset 0 0 30px rgba(34,211,238,0.07); }

.bento-label {
  font-size: 0.65rem;
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
  font-size: 1.6rem;
  font-weight: 700;
  line-height: 1;
  text-shadow: 0 0 12px currentColor;
}

/* 日柱主卡片 — col-span-2 流光边框 */
.bento-day-wrapper {
  grid-column: span 2;
  padding: 0;
  border: none;
  background: transparent;
  position: relative;
  overflow: hidden;
  border-radius: 12px;
}

.bento-day-beam {
  position: absolute;
  inset: -150%;
  background: conic-gradient(
    from 0deg,
    transparent 0%,
    #22d3ee 8%,
    transparent 16%
  );
  animation: bento-beam-rotate 4s linear infinite;
  z-index: 0;
}

@keyframes bento-beam-rotate {
  to { transform: rotate(360deg); }
}

.bento-day-inner {
  position: relative;
  z-index: 1;
  margin: 1.5px;
  background: rgba(10,10,14,0.97);
  border-radius: 11px;
  padding: 0.75rem 1rem;
  transition: box-shadow 0.3s, background 0.3s;
}

.bento-day-label {
  text-align: center;
  font-size: 0.7rem;
  color: var(--text-muted);
}

.bento-day-body {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 1.5rem;
}

.bento-day-gan,
.bento-day-zhi {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.2rem;
}

.bento-day-char {
  font-family: var(--font-serif), serif;
  font-size: 2.4rem;
  font-weight: 700;
  line-height: 1;
  text-shadow: 0 0 20px currentColor;
}

.bento-day-divider {
  width: 1px;
  height: 3.5rem;
  background: linear-gradient(180deg, transparent, rgba(255,255,255,0.1), transparent);
}

.elem-tag-lg {
  font-size: 0.68rem;
  padding: 0.15rem 0.5rem;
}

.bento-day-sub {
  margin-top: 0.25rem;
}

/* 五行雷达图卡片 */
.bento-radar {
  display: flex;
  flex-direction: column;
}

.bento-radar-chart {
  height: 148px;
  width: 100%;
}

.bento-sub {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.1rem;
  padding-top: 0.25rem;
  border-top: 1px solid rgba(255,255,255,0.04);
  margin-top: 0.25rem;
}

.elem-tag {
  display: inline-block;
  font-size: 0.6rem;
  padding: 0.1rem 0.35rem;
  border-radius: 3px;
  border: 1px solid;
  font-weight: 600;
  letter-spacing: 0.5px;
}

.bento-god-tag {
  font-size: 0.6rem;
  padding: 0.1rem 0.4rem;
  border-radius: 4px;
  background: rgba(255,255,255,0.05);
  border: 1px solid rgba(255,255,255,0.1);
  color: var(--text-muted);
  font-weight: 600;
  letter-spacing: 0.5px;
  margin-top: 0.2rem;
}

.bento-god-tag-day {
  font-size: 0.65rem;
  padding: 0.15rem 0.5rem;
  background: rgba(203,213,225,0.08);
  border-color: rgba(203,213,225,0.15);
  color: var(--accent);
}

/* ===== 天干地支关系 — 紧凑内联布局 ===== */

.ganzhi-analysis {
  margin-top: 2px;
}

.relations-section {
  padding: 0.4rem 1rem;
  border-bottom: 1px solid rgba(255,255,255, 0.06);
}

.relations-title {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 0.65rem;
  font-weight: 700;
  color: var(--text-muted);
  letter-spacing: 1.5px;
  margin-bottom: 0.35rem;
}

.relations-title-dot {
  width: 5px;
  height: 5px;
  border-radius: 50%;
  background: var(--accent);
  box-shadow: 0 0 6px rgba(203,213,225, 0.3);
}

.relations-title-dot.zhi-dot {
  background: #6495ed;
  box-shadow: 0 0 6px rgba(100, 149, 237, 0.4);
}

.relations-count {
  font-size: 0.55rem;
  font-weight: 400;
  color: var(--text-soft);
  margin-left: auto;
  letter-spacing: 0.5px;
}

.ganzhi-compact {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.gz-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.3rem 0.5rem;
  border-radius: 4px;
  border-left: 2px solid transparent;
  transition: background 0.2s;
}

.gz-item:hover {
  background: rgba(255, 255, 255, 0.03);
}

.gz-item.rel-he { border-left-color: rgba(74, 222, 128, 0.5); }
.gz-item.rel-ke { border-left-color: rgba(248, 113, 113, 0.5); }
.gz-item.rel-sheng { border-left-color: rgba(96, 165, 250, 0.5); }
.gz-item.rel-chong { border-left-color: rgba(248, 113, 113, 0.5); }
.gz-item.rel-hai { border-left-color: rgba(251, 146, 60, 0.5); }
.gz-item.rel-xing { border-left-color: rgba(192, 132, 252, 0.5); }
.gz-item.rel-hui { border-left-color: rgba(96, 165, 250, 0.5); }

.gz-chars {
  display: flex;
  align-items: center;
  gap: 1px;
  flex-shrink: 0;
}

.gz-c {
  font-family: var(--font-serif), serif;
  font-size: 0.95rem;
  font-weight: 700;
  color: var(--text);
  line-height: 1;
}

.gz-sym {
  font-size: 0.65rem;
  font-weight: 700;
  padding: 0 2px;
}

.sym-rel-he { color: #4ade80; }
.sym-rel-ke { color: #f87171; }
.sym-rel-sheng { color: #60a5fa; }
.sym-rel-chong { color: #f87171; }
.sym-rel-hai { color: #fb923c; }
.sym-rel-xing { color: #c084fc; }
.sym-rel-hui { color: #60a5fa; }

.gz-tag {
  font-size: 0.5rem;
  font-weight: 600;
  padding: 0.05rem 0.3rem;
  border-radius: 2px;
  letter-spacing: 0.5px;
  flex-shrink: 0;
}

.tag-rel-he { color: #4ade80; background: rgba(74, 222, 128, 0.1); }
.tag-rel-ke { color: #f87171; background: rgba(248, 113, 113, 0.1); }
.tag-rel-sheng { color: #60a5fa; background: rgba(96, 165, 250, 0.1); }
.tag-rel-chong { color: #f87171; background: rgba(248, 113, 113, 0.1); }
.tag-rel-hai { color: #fb923c; background: rgba(251, 146, 60, 0.1); }
.tag-rel-xing { color: #c084fc; background: rgba(192, 132, 252, 0.1); }
.tag-rel-hui { color: #60a5fa; background: rgba(96, 165, 250, 0.1); }

.gz-text {
  color: var(--text-muted);
  font-size: 0.68rem;
  line-height: 1.3;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.no-relations {
  padding: 0.4rem 1rem;
  text-align: center;
  font-size: 0.65rem;
  color: var(--text-soft);
  border-bottom: 1px solid rgba(255,255,255, 0.04);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
}

.no-rel-icon {
  font-size: 0.5rem;
  color: var(--text-soft);
}

/* Analysis sections */
.analysis-section {
  padding: 1rem 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

/* Tab navigation */
.chart-tabs {
  display: flex;
  gap: 0.25rem;
  border-bottom: 1px solid rgba(255,255,255, 0.1);
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
  font-size: 0.78rem;
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

.block-title {
  font-size: 0.78rem;
  font-weight: 700;
  color: var(--accent);
  letter-spacing: 1px;
  margin-bottom: 0.5rem;
}

.block-desc {
  display: block;
  font-size: 0.65rem;
  color: var(--text-soft);
  line-height: 1.5;
  margin-top: -0.3rem;
  margin-bottom: 0.5rem;
  letter-spacing: 0.3px;
}

/* Ten gods */
.ten-gods-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 0.4rem;
}

.god-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 0.4rem;
  background: rgba(255, 255, 255, 0.03);
  border-radius: 6px;
  border: 1px solid rgba(255,255,255, 0.06);
}

.god-pillar {
  font-size: 0.65rem;
  color: var(--text-muted);
  margin-bottom: 0.2rem;
}

.god-name {
  font-size: 0.8rem;
  font-weight: 700;
  color: #fb7185;
}

/* NaYin */
.nayin-list {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem;
}

.nayin-tag {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  font-size: 0.72rem;
  padding: 0.25rem 0.6rem;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid;
  border-radius: 4px;
  color: var(--text);
  cursor: pointer;
  transition: background 0.2s;
}

.nayin-tag:hover {
  background: rgba(255, 255, 255, 0.08);
}

.nayin-pillar {
  opacity: 0.6;
  font-size: 0.68rem;
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
  border-bottom: 1px solid rgba(255,255,255, 0.15);
}

.nayin-detail-name {
  font-size: 1.1rem;
  font-weight: 700;
  color: var(--accent);
}

.nayin-detail-elem {
  font-size: 0.68rem;
  padding: 0.15rem 0.5rem;
  border-radius: 3px;
  color: #1a1a1a;
  font-weight: 600;
}

.nayin-detail-section {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}

.nayin-detail-label {
  font-size: 0.65rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-dim);
}

.nayin-detail-value {
  font-size: 0.82rem;
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
  font-size: 0.7rem;
  padding: 0.15rem 0.5rem;
  background: rgba(255,255,255, 0.08);
  border: 1px solid rgba(255,255,255, 0.15);
  border-radius: 3px;
  color: var(--accent);
}

/* DaYun — 时间轴风格 */
.dayun-timeline {
  display: flex;
  flex-wrap: wrap;
  gap: 0;
}

.dayun-node {
  display: flex;
  flex-direction: column;
  align-items: center;
  min-width: 64px;
  flex: 1;
}

.dayun-age {
  font-size: 0.6rem;
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
  box-shadow: 0 0 8px rgba(203,213,225, 0.3);
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
  background: linear-gradient(90deg, rgba(203,213,225, 0.2), rgba(203,213,225, 0.05));
  transform: translateY(-50%);
}

.dayun-pillar-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0;
  margin-top: 0.4rem;
  padding: 0.35rem 0.5rem;
  background: rgba(203,213,225, 0.04);
  border: 1px solid rgba(203,213,225, 0.1);
  border-radius: 6px;
}

.dayun-gan {
  font-family: var(--font-serif), serif;
  font-size: 0.95rem;
  font-weight: 700;
  color: var(--accent);
  line-height: 1.2;
}

.dayun-zhi {
  font-family: var(--font-serif), serif;
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--text-muted);
  line-height: 1.2;
}

/* Element Detail Table */
.element-detail-table {
  border: 1px solid rgba(255,255,255, 0.1);
  border-radius: 8px;
  overflow: hidden;
}

.ed-header {
  display: grid;
  grid-template-columns: 1fr 1fr 1.5fr 1fr;
  padding: 0.4rem 0.75rem;
  background: rgba(255, 255, 255, 0.03);
  font-size: 0.65rem;
  color: var(--text-muted);
  letter-spacing: 0.5px;
  border-bottom: 1px solid rgba(255,255,255, 0.08);
}

.ed-row {
  display: grid;
  grid-template-columns: 1fr 1fr 1.5fr 1fr;
  padding: 0.4rem 0.75rem;
  font-size: 0.78rem;
  border-bottom: 1px solid rgba(255,255,255, 0.04);
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
.level-medium .ed-total { color: #9ca3af; }
.level-strong .ed-total { color: #fbbf24; font-weight: 700; }
.level-very-strong .ed-total { color: #f97316; font-weight: 800; }

/* Body Strength */
.body-strength {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.bs-verdict {
  font-size: 0.95rem;
  font-weight: 700;
  color: var(--accent);
}

.bs-tags {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.35rem;
}

.bs-like-label,
.bs-dislike-label {
  font-size: 0.65rem;
  padding: 0.15rem 0.5rem;
  border-radius: 4px;
}

.bs-like-label {
  background: rgba(74, 222, 128, 0.1);
  color: #4ade80;
  border: 1px solid rgba(74, 222, 128, 0.2);
}

.bs-dislike-label {
  background: rgba(251, 113, 133, 0.1);
  color: #fb7185;
  border: 1px solid rgba(251, 113, 133, 0.2);
}

.bs-like {
  font-size: 0.72rem;
  padding: 0.15rem 0.5rem;
  background: rgba(74, 222, 128, 0.06);
  color: #4ade80;
  border: 1px solid rgba(74, 222, 128, 0.12);
  border-radius: 4px;
}

.bs-dislike {
  font-size: 0.72rem;
  padding: 0.15rem 0.5rem;
  background: rgba(251, 113, 133, 0.06);
  color: #fb7185;
  border: 1px solid rgba(251, 113, 133, 0.12);
  border-radius: 4px;
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
  font-size: 0.95rem;
  font-weight: 700;
  color: var(--accent);
  letter-spacing: 0.05em;
}

.pattern-type-tag {
  display: inline-block;
  padding: 0.12rem 0.5rem;
  border-radius: 4px;
  font-size: 0.65rem;
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
  font-size: 0.78rem;
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
  font-size: 0.65rem;
  font-weight: 600;
  padding: 0.1rem 0.4rem;
  border-radius: 3px;
}

.pattern-favor .pattern-elem-label {
  background: rgba(74, 222, 128, 0.1);
  color: #4ade80;
  border: 1px solid rgba(74, 222, 128, 0.2);
}

.pattern-avoid .pattern-elem-label {
  background: rgba(251, 113, 133, 0.1);
  color: #fb7185;
  border: 1px solid rgba(251, 113, 133, 0.2);
}

.pattern-elem-tag {
  font-size: 0.7rem;
  padding: 0.12rem 0.45rem;
  border-radius: 3px;
}

.pattern-elem-tag.favor {
  background: rgba(74, 222, 128, 0.06);
  color: #4ade80;
  border: 1px solid rgba(74, 222, 128, 0.12);
}

.pattern-elem-tag.avoid {
  background: rgba(251, 113, 133, 0.06);
  color: #fb7185;
  border: 1px solid rgba(251, 113, 133, 0.12);
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
  font-size: 1.15rem;
  font-weight: 700;
  color: var(--accent);
  letter-spacing: 0.15em;
}

.ming-gong-nayin {
  font-size: 0.78rem;
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
  font-size: 0.75rem;
  font-weight: 600;
  white-space: nowrap;
  background: rgba(100, 100, 100, 0.25);
  color: var(--text);
}

.shensha-badge.shensha-ji {
  background: rgba(196, 164, 75, 0.18);
  color: var(--accent);
  border: 1px solid rgba(255,255,255, 0.25);
}

.shensha-badge.shensha-xiong {
  background: rgba(251, 113, 133, 0.14);
  color: #f87171;
  border: 1px solid rgba(251, 113, 133, 0.22);
}

.shensha-desc {
  font-size: 0.78rem;
  color: var(--text-muted);
  line-height: 1.55;
}

.ming-gong-zhi {
  font-size: 0.78rem;
  color: var(--text-muted);
  line-height: 1.6;
  border-left: 2px solid rgba(255,255,255, 0.25);
  padding-left: 0.6rem;
}

.tiaohou-source {
  font-size: 0.65rem;
  color: var(--text-muted);
  font-weight: 400;
  margin-left: 0.4rem;
  font-style: italic;
}

.tiaohou-card {
  background: rgba(203,213,225,0.06);
  border: 1px solid rgba(203,213,225,0.18);
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
  font-size: 1.3rem;
  font-weight: 700;
  color: var(--accent);
}

.tiaohou-arrow {
  color: var(--text-muted);
  font-size: 0.8rem;
}

.tiaohou-month {
  font-size: 1.3rem;
  font-weight: 700;
  color: var(--text);
}

.tiaohou-divider {
  color: var(--text-soft);
  margin: 0 0.2rem;
}

.tiaohou-label {
  color: var(--text-soft);
  font-size: 0.78rem;
}

.tiaohou-primary {
  font-size: 1.35rem;
  font-weight: 800;
}

.tiaohou-summary {
  font-size: 0.82rem;
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
  font-size: 0.78rem;
}

.tiaohou-xi {
  color: #34d399;
  font-weight: 600;
  white-space: nowrap;
}

.tiaohou-ji {
  color: #fb7185;
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
  font-size: 0.82rem;
  color: var(--text-muted);
  line-height: 1.7;
  white-space: pre-wrap;
  margin: 0;
}


.sheng-xiao-tag {
  font-size: 0.6rem;
  color: var(--text-muted);
  letter-spacing: 1px;
}

.empties-tag {
  font-size: 0.55rem;
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
  border: 1px solid rgba(255,255,255, 0.1);
  border-radius: 7px;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.035);
}

.shen-sha-name {
  font-family: var(--font-serif), serif;
  font-size: 0.8rem;
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
  background: rgba(255,255,255, 0.1);
  border: 1px solid rgba(255,255,255, 0.18);
  color: var(--text);
  font-size: 0.68rem;
  font-weight: 700;
}

.shen-sha-desc {
  color: var(--text-muted);
  font-size: 0.72rem;
  line-height: 1.45;
}

/* ShenSha group title */
.shen-sha-group-title {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  font-family: var(--font-serif), serif;
  font-size: 0.86rem;
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
  font-size: 0.66rem;
  font-weight: 400;
  color: var(--text-soft);
  letter-spacing: 0.03em;
}

/* ShenSha summary */
.shen-sha-summary-block {
  border-top: 1px solid rgba(255,255,255, 0.08);
  margin-top: 0.5rem;
  padding-top: 0.75rem;
}

.shen-sha-summary-title {
  font-size: 0.78rem;
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
  font-size: 0.7rem;
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
  border: 1px solid rgba(255,255,255, 0.12);
  border-radius: 10px;
  overflow: hidden;
  background:
    linear-gradient(180deg, rgba(203,213,225,0.03) 0%, transparent 100%),
    rgba(255, 255, 255, 0.015);
  padding: 0.75rem 0.5rem 0.5rem;
  box-shadow: inset 0 1px 0 rgba(203,213,225,0.08);
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
  font-size: 13px;
}
.ten-god-tabs .el-tabs__item.is-active {
  color: var(--accent);
}
.ten-god-tabs .el-tabs__nav-wrap::after {
  background: rgba(255,255,255,0.06);
}
.tg-summary {
  background: rgba(203,213,225,0.06);
  border: 1px solid rgba(203,213,225,0.12);
  border-radius: 10px;
  padding: 1rem 1.25rem;
  font-size: 14px;
  line-height: 1.8;
  color: var(--text);
}
.tg-text {
  font-size: 14px;
  line-height: 1.8;
  color: var(--text);
}
.tg-god-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.tg-god-card {
  background: rgba(255,255,255,0.03);
  border: 1px solid rgba(203,213,225,0.1);
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
  font-size: 14px;
  color: var(--accent);
}
.tg-god-pct {
  font-size: 13px;
  color: var(--text-muted);
}
.tg-god-meaning {
  font-size: 13px;
  line-height: 1.6;
  color: var(--text);
  margin-bottom: 4px;
}
.tg-god-advice {
  font-size: 12px;
  line-height: 1.6;
  color: var(--text);
}

/* WuXingFlow 五行流通 */
.wuxing-flow-card {
  display: flex; flex-direction: column; gap: 0.5rem;
  padding: 0.75rem;
  background: rgba(255,255,255,0.015);
  border-radius: 8px;
  border: 1px solid rgba(203,213,225,0.06);
}
.wf-header { display: flex; align-items: center; gap: 0.5rem; flex-wrap: wrap; }
.wf-label { font-size: 0.65rem; color: var(--text-muted); }
.wf-elem { font-size: 0.85rem; font-weight: 700; }
.wf-type { font-size: 0.7rem; color: var(--text-muted); padding: 0.1rem 0.4rem; background: rgba(255,255,255,0.03); border-radius: 4px; }
.wf-status { font-size: 0.65rem; font-weight: 600; padding: 0.1rem 0.5rem; border-radius: 4px; }
.wf-smooth { color: #4ade80; background: rgba(74,222,128,0.1); }
.wf-blocked { color: #fb7185; background: rgba(251,113,133,0.1); }
.wf-paths { display: flex; flex-direction: column; gap: 0.25rem; }
.wf-path { display: flex; align-items: center; gap: 0.4rem; font-size: 0.72rem; color: var(--text-muted); }
.wf-path-dot { width: 4px; height: 4px; border-radius: 50%; background: var(--accent); flex-shrink: 0; }
.wf-blocked-row { display: flex; align-items: center; gap: 0.4rem; font-size: 0.72rem; }
.wf-blocked-label { color: #fb7185; font-size: 0.65rem; }
.wf-blocked-elem { font-weight: 600; }
.wf-verdict { font-size: 0.72rem; color: var(--text-muted); font-style: italic; }
.wf-advice { font-size: 0.72rem; color: var(--accent); opacity: 0.8; }

/* TongGuan 通关用神 */
.tong-guan-card {
  display: flex; align-items: center; gap: 0.75rem;
  padding: 0.75rem;
  background: rgba(255,255,255,0.015);
  border-radius: 8px;
  border: 1px solid rgba(203,213,225,0.06);
}
.tg-elem { font-size: 1.4rem; font-weight: 900; font-family: var(--font-serif); }
.tg-desc { font-size: 0.78rem; color: var(--text-muted); line-height: 1.5; }

/* MissingElements 缺失五行 */
.missing-elem-card {
  display: flex; flex-direction: column; gap: 0.4rem;
  padding: 0.75rem;
  background: rgba(255,255,255,0.015);
  border-radius: 8px;
  border: 1px solid rgba(203,213,225,0.06);
}
.me-row { display: flex; align-items: center; gap: 0.4rem; flex-wrap: wrap; }
.me-label { font-size: 0.65rem; color: var(--text-muted); min-width: 28px; }
.me-tag { font-size: 0.78rem; font-weight: 700; padding: 0.1rem 0.4rem; border: 1px solid; border-radius: 4px; }
.me-missing { background: rgba(251,113,133,0.06); }
.me-remedy { background: rgba(74,222,128,0.06); }
.me-severity { font-size: 0.62rem; padding: 0.1rem 0.4rem; border-radius: 4px; margin-left: 0.25rem; }
.me-轻微 { color: var(--accent); background: rgba(203,213,225,0.1); }
.me-中等 { color: #fb7185; background: rgba(251,113,133,0.1); }
.me-严重 { color: #fb7185; background: rgba(251,113,133,0.2); }

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
  border: 1px solid rgba(255, 255, 255, 0.04);
  transition: all 0.25s ease;
}

.dayun-flow-item:hover {
  background: rgba(255, 255, 255, 0.04);
  border-color: rgba(255,255,255, 0.1);
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
  font-size: 0.6rem;
  font-weight: 600;
  color: var(--text-dim);
  white-space: nowrap;
}

.df-pillar {
  font-family: var(--font-serif), serif;
  font-size: 0.85rem;
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
  font-size: 0.6rem;
  font-weight: 600;
  padding: 0.1rem 0.5rem;
  border-radius: 3px;
  letter-spacing: 0.5px;
}

.dfb-增强 { color: #4ade80; background: rgba(74, 222, 128, 0.1); border: 1px solid rgba(74, 222, 128, 0.18); }
.dfb-减弱 { color: #f87171; background: rgba(248, 113, 113, 0.1); border: 1px solid rgba(248, 113, 113, 0.18); }
.dfb-不变 { color: var(--text-dim); background: rgba(255, 255, 255, 0.04); border: 1px solid rgba(255, 255, 255, 0.08); }

.df-增强 { border-left: 3px solid rgba(74, 222, 128, 0.4); }
.df-减弱 { border-left: 3px solid rgba(248, 113, 113, 0.4); }
.df-不变 { border-left: 3px solid rgba(255,255,255, 0.2); }

.df-impact {
  font-size: 0.72rem;
  color: var(--text-muted);
  line-height: 1.5;
}
</style>

<style>
/* NaYin popover — dark theme to match app */
.nayin-popover {
  background: var(--surface-1) !important;
  border: 1px solid rgba(255,255,255,0.1) !important;
  border-radius: 10px !important;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.6), 0 0 24px rgba(203,213,225, 0.08) !important;
  padding: 14px 16px !important;
  color: var(--text);
}
.nayin-popover .el-popover__title {
  color: var(--text);
}
</style>
