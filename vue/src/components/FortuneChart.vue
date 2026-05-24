<script setup lang="ts">
import { computed } from 'vue'
import { use } from 'echarts/core'
import { LineChart } from 'echarts/charts'
import { GridComponent, TooltipComponent, LegendComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import VChart from 'vue-echarts'

use([CanvasRenderer, LineChart, GridComponent, TooltipComponent, LegendComponent])

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
  }>(),
  {
    dailyData: () => [],
    height: '320px',
  },
)

const elementSeries = [
  { key: 'metal' as const, name: '金', color: '#FFD700' },
  { key: 'wood' as const, name: '木', color: '#228B22' },
  { key: 'water' as const, name: '水', color: '#4169E1' },
  { key: 'fire' as const, name: '火', color: '#DC143C' },
  { key: 'earth' as const, name: '土', color: '#DAA520' },
]

const option = computed(() => {
  const dates = props.dailyData.map((d) => d.date)
  const scores = props.dailyData.map((d) => d.score)

  const series: any[] = [
    {
      name: '运势评分',
      type: 'line',
      yAxisIndex: 0,
      data: scores,
      lineStyle: { color: '#C41E3A', width: 2.5 },
      itemStyle: { color: '#C41E3A' },
      symbol: 'circle',
      symbolSize: 6,
      smooth: true,
    },
    ...elementSeries.map((el) => ({
      name: el.name,
      type: 'line',
      yAxisIndex: 1,
      data: props.dailyData.map((d) => d[el.key]),
      lineStyle: { color: el.color, width: 1.5, opacity: 0.75 },
      itemStyle: { color: el.color },
      symbol: 'none',
      smooth: true,
    })),
  ]

  return {
    backgroundColor: 'transparent',
    grid: {
      left: 45,
      right: 50,
      top: 20,
      bottom: 35,
    },
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'cross' },
      backgroundColor: 'rgba(10, 8, 21, 0.9)',
      borderColor: 'rgba(212, 168, 75, 0.2)',
      textStyle: { color: '#F0EDE4', fontSize: 11 },
    },
    legend: {
      bottom: 0,
      textStyle: { color: 'rgba(212,168,75,0.5)', fontSize: 10 },
      itemWidth: 12,
      itemHeight: 8,
    },
    xAxis: {
      type: 'category',
      data: dates,
      axisLabel: {
        fontSize: 10,
        color: 'rgba(139, 131, 120, 0.6)',
        formatter: (val: string) => {
          const parts = val.split('-')
          if (parts.length === 3) return `${parts[1]}/${parts[2]}`
          return val
        },
      },
      axisLine: { lineStyle: { color: 'rgba(212, 168, 75, 0.1)' } },
      axisTick: { show: false },
    },
    yAxis: [
      {
        type: 'value',
        name: '评分',
        min: 0,
        max: 100,
        interval: 20,
        axisLabel: { fontSize: 10, color: 'rgba(139, 131, 120, 0.6)' },
        splitLine: { lineStyle: { color: 'rgba(212, 168, 75, 0.06)', type: 'dashed' } },
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
    <v-chart v-if="dailyData.length" class="chart-instance" :option="option" autoresize />
    <div v-else class="chart-empty">
      <div class="empty-constellation">
        <svg width="100" height="100" viewBox="0 0 100 100" fill="none">
          <circle
            cx="50"
            cy="50"
            r="42"
            stroke="#D4A84B"
            stroke-width="0.5"
            stroke-dasharray="2 3"
            opacity="0.25"
          />
          <circle
            cx="50"
            cy="50"
            r="28"
            stroke="#D4A84B"
            stroke-width="0.5"
            stroke-dasharray="1 4"
            opacity="0.18"
          />
          <circle cx="50" cy="50" r="5" fill="#D4A84B" opacity="0.2" />
          <circle cx="25" cy="32" r="2.5" fill="#D4A84B" opacity="0.4" class="star-pulse" />
          <circle
            cx="75"
            cy="28"
            r="3"
            fill="#D4A84B"
            opacity="0.35"
            class="star-pulse"
            style="animation-delay: 0.4s"
          />
          <circle
            cx="78"
            cy="68"
            r="2"
            fill="#D4A84B"
            opacity="0.3"
            class="star-pulse"
            style="animation-delay: 0.8s"
          />
          <circle
            cx="22"
            cy="72"
            r="3"
            fill="#D4A84B"
            opacity="0.35"
            class="star-pulse"
            style="animation-delay: 1.2s"
          />
          <line
            x1="25"
            y1="32"
            x2="50"
            y2="50"
            stroke="#D4A84B"
            stroke-width="0.4"
            opacity="0.08"
          />
          <line
            x1="75"
            y1="28"
            x2="50"
            y2="50"
            stroke="#D4A84B"
            stroke-width="0.4"
            opacity="0.06"
          />
        </svg>
      </div>
      <p class="empty-title">暂无数据</p>
      <p class="empty-sub">运势数据将显示在这里</p>
    </div>
  </div>
</template>

<style scoped>
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
  background: rgba(255, 255, 255, 0.02);
  border: 1px dashed rgba(212, 168, 75, 0.12);
  border-radius: 12px;
  gap: 0.75rem;
}

.empty-constellation {
  animation: spin-slow 30s linear infinite;
  opacity: 0.7;
}

@keyframes spin-slow {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.star-pulse {
  animation: star-twinkle 2.5s ease-in-out infinite;
}

@keyframes star-twinkle {
  0%,
  100% {
    opacity: 0.25;
  }
  50% {
    opacity: 0.7;
  }
}

.empty-title {
  font-size: 0.9rem;
  font-weight: 600;
  color: rgba(212, 168, 75, 0.4);
  margin: 0;
  letter-spacing: 1px;
}

.empty-sub {
  font-size: 0.75rem;
  color: rgba(139, 131, 120, 0.35);
  margin: 0;
}
</style>
