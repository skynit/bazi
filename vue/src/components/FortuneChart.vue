<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { use } from 'echarts/core'
import { LineChart } from 'echarts/charts'
import { AriaComponent, GridComponent, TooltipComponent, LegendComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import VChart from 'vue-echarts'

use([CanvasRenderer, LineChart, GridComponent, TooltipComponent, LegendComponent, AriaComponent])

export interface TrendPoint {
  date: string
  score: number
  metal: number
  wood: number
  water: number
  fire: number
  earth: number
}

const props = withDefaults(
  defineProps<{
    dailyData?: TrendPoint[]
    height?: string
    showElements?: boolean
    metricLabel?: string
  }>(),
  {
    dailyData: () => [],
    height: '320px',
    showElements: true,
    metricLabel: '命中关系数',
  },
)

const isDark = ref(document.documentElement.classList.contains('dark'))
const chartDescription = computed(() =>
  props.dailyData.length
    ? props.dailyData.map((item) => `${item.date} ${props.metricLabel}${item.score}条`).join('；')
    : '暂无趋势数据',
)

const elementSeries = computed(() => {
  themeVersion.value
  if (isDark.value) {
    return [
      { key: 'metal' as const, name: '金', color: cssVar('--wuxing-jin', '#cbd5e1') },
      { key: 'wood' as const, name: '木', color: cssVar('--wuxing-mu', '#34d399') },
      { key: 'water' as const, name: '水', color: cssVar('--wuxing-shui', '#22d3ee') },
      { key: 'fire' as const, name: '火', color: cssVar('--wuxing-huo', '#fb7185') },
      { key: 'earth' as const, name: '土', color: cssVar('--wuxing-tu', '#fde68a') },
    ]
  }
  return [
    { key: 'metal' as const, name: '金', color: '#94a3b8' },
    { key: 'wood' as const, name: '木', color: '#16a34a' },
    { key: 'water' as const, name: '水', color: '#0891b2' },
    { key: 'fire' as const, name: '火', color: '#dc2626' },
    { key: 'earth' as const, name: '土', color: '#a16207' },
  ]
})

const themeVersion = ref(0)
let themeObserver: MutationObserver | null = null

function cssVar(name: string, fallback: string): string {
  if (typeof window === 'undefined') return fallback
  return getComputedStyle(document.documentElement).getPropertyValue(name).trim() || fallback
}

onMounted(() => {
  themeObserver = new MutationObserver(() => {
    themeVersion.value += 1
    isDark.value = document.documentElement.classList.contains('dark')
  })
  themeObserver.observe(document.documentElement, {
    attributes: true,
    attributeFilter: ['class', 'style', 'data-wuxing'],
  })
})

onUnmounted(() => {
  themeObserver?.disconnect()
  themeObserver = null
})

const option = computed(() => {
  themeVersion.value
  const textColor = cssVar('--text', '#0f1712')
  const mutedColor = cssVar('--text-muted', '#5a6a5e')
  const lineColor = cssVar('--line-subtle', 'rgba(15, 23, 18, 0.06)')
  const tooltipBg = cssVar('--surface-1', '#ffffff')
  const accentColor = cssVar('--jade-accent', isDark.value ? '#4ade80' : '#16a34a')
  const dates = props.dailyData.map((d) => d.date)
  const scores = props.dailyData.map((d) => d.score)
  const metricMax = Math.max(...scores, 1)

  const series: any[] = [
    {
      name: props.metricLabel,
      type: 'line',
      yAxisIndex: 0,
      data: scores,
      lineStyle: { color: accentColor, width: 2 },
      itemStyle: { color: accentColor },
      symbol: 'circle',
      symbolSize: 5,
      smooth: 0.4,
    },
    ...(props.showElements
      ? elementSeries.value.map((el) => ({
          name: el.name,
          type: 'line',
          yAxisIndex: 1,
          data: props.dailyData.map((d) => (d as unknown as Record<string, number>)[el.key]),
          lineStyle: { color: el.color, width: 1.5, opacity: 0.75 },
          itemStyle: { color: el.color },
          symbol: 'none',
          smooth: true,
        }))
      : []),
  ]

  return {
    aria: {
      enabled: true,
      decal: { show: true },
      description: chartDescription.value,
    },
    backgroundColor: 'transparent',
    grid: {
      left: 45,
      right: props.showElements ? 50 : 22,
      top: 20,
      bottom: props.showElements ? 35 : 26,
    },
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'cross' },
      backgroundColor: tooltipBg,
      borderColor: lineColor,
      textStyle: { color: textColor, fontSize: 11 },
    },
    legend: {
      show: props.showElements,
      bottom: 0,
      textStyle: { color: mutedColor, fontSize: 10 },
      itemWidth: 12,
      itemHeight: 8,
    },
    xAxis: {
      type: 'category',
      data: dates,
      axisLabel: {
        fontSize: 10,
        color: mutedColor,
        formatter: (val: string) => {
          const parts = val.split('-')
          if (parts.length === 3) return `${parts[1]}/${parts[2]}`
          return val
        },
      },
      axisLine: { lineStyle: { color: lineColor } },
      axisTick: { show: false },
    },
    yAxis: [
      {
        type: 'value',
        name: '条',
        min: 0,
        max: Math.max(2, Math.ceil(metricMax)),
        minInterval: 1,
        axisLabel: { fontSize: 10, color: mutedColor },
        splitLine: { lineStyle: { color: lineColor, type: 'dashed' } },
        axisLine: { show: false },
        axisTick: { show: false },
      },
      {
        type: 'value',
        name: '',
        min: 0,
        max: 100,
        axisLabel: { show: false },
        splitLine: { show: false },
        axisLine: { show: false },
        axisTick: { show: false },
      },
    ],
    series,
  }
})
</script>

<template>
  <div class="fortune-chart" :style="{ height }">
    <v-chart
      v-if="dailyData.length"
      class="chart-instance"
      :option="option"
      :aria-label="chartDescription"
      autoresize
    />
    <div v-else class="chart-empty">
      <div class="empty-constellation">
        <svg width="100" height="100" viewBox="0 0 100 100" fill="none">
          <circle
            cx="50"
            cy="50"
            r="42"
            stroke="currentColor"
            stroke-width="0.5"
            stroke-dasharray="2 3"
            opacity="0.25"
          />
          <circle
            cx="50"
            cy="50"
            r="28"
            stroke="currentColor"
            stroke-width="0.5"
            stroke-dasharray="1 4"
            opacity="0.18"
          />
          <circle cx="50" cy="50" r="5" fill="currentColor" opacity="0.2" />
          <circle cx="25" cy="32" r="2.5" fill="currentColor" opacity="0.4" class="star-pulse" />
          <circle
            cx="75"
            cy="28"
            r="3"
            fill="currentColor"
            opacity="0.35"
            class="star-pulse"
            style="animation-delay: 0.4s"
          />
          <circle
            cx="78"
            cy="68"
            r="2"
            fill="currentColor"
            opacity="0.3"
            class="star-pulse"
            style="animation-delay: 0.8s"
          />
          <circle
            cx="22"
            cy="72"
            r="3"
            fill="currentColor"
            opacity="0.35"
            class="star-pulse"
            style="animation-delay: 1.2s"
          />
          <line
            x1="25"
            y1="32"
            x2="50"
            y2="50"
            stroke="currentColor"
            stroke-width="0.4"
            opacity="0.08"
          />
          <line
            x1="75"
            y1="28"
            x2="50"
            y2="50"
            stroke="currentColor"
            stroke-width="0.4"
            opacity="0.06"
          />
        </svg>
      </div>
      <p class="empty-title">暂无数据</p>
      <p class="empty-sub">结构趋势数据将显示在这里</p>
    </div>
  </div>
  <details v-if="dailyData.length" class="chart-data-details">
    <summary>查看图表数据</summary>
    <div class="chart-data-table-wrap">
      <table>
        <thead>
          <tr>
            <th>日期</th>
            <th>{{ metricLabel }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in dailyData" :key="item.date">
            <td>{{ item.date }}</td>
            <td>{{ item.score }} 条</td>
          </tr>
        </tbody>
      </table>
    </div>
  </details>
</template>

<style scoped>
.chart-data-details {
  margin-top: 0.65rem;
  color: var(--text-muted);
  font-size: var(--fs-xs);
}

.chart-data-details summary {
  width: fit-content;
  min-height: 36px;
  cursor: pointer;
}

.chart-data-table-wrap {
  max-height: 15rem;
  overflow: auto;
}

.chart-data-details table {
  width: 100%;
  border-collapse: collapse;
}

.chart-data-details th,
.chart-data-details td {
  padding: 0.45rem;
  border-bottom: 1px solid var(--line-subtle);
  text-align: left;
}
.fortune-chart {
  width: 100%;
}

.chart-instance {
  width: 100%;
  height: 100%;
}

.chart-empty {
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: var(--surface-2);
  border: 1px dashed var(--line-subtle);
  border-radius: 12px;
  gap: 0.75rem;
}

.empty-constellation {
  color: var(--icon-muted);
  opacity: 0.7;
}

.empty-title {
  font-size: var(--fs-sm);
  font-weight: 600;
  color: var(--text-muted);
  margin: 0;
  letter-spacing: 1px;
}

.empty-sub {
  font-size: var(--fs-xs);
  color: var(--text-soft);
  margin: 0;
}
</style>
