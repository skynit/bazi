<script setup lang="ts">
import { computed } from 'vue'
import { use } from 'echarts/core'
import { LineChart } from 'echarts/charts'
import {
  GridComponent,
  TooltipComponent,
  LegendComponent,
} from 'echarts/components'
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
      markLine: {
        silent: true,
        symbol: 'none',
        lineStyle: { type: 'dashed', color: '#999', width: 1 },
        data: [{ yAxis: 60, label: { formatter: '60', fontSize: 10 } }],
      },
    },
    ...elementSeries.map((el) => ({
      name: el.name,
      type: 'line',
      yAxisIndex: 1,
      data: props.dailyData.map((d) => d[el.key]),
      lineStyle: { color: el.color, width: 1.5, opacity: 0.8 },
      itemStyle: { color: el.color },
      symbol: 'none',
      smooth: true,
    })),
  ]

  return {
    grid: {
      left: 40,
      right: 50,
      top: 30,
      bottom: 30,
    },
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'cross' },
    },
    legend: {
      bottom: 0,
      textStyle: { fontSize: 10 },
      itemWidth: 12,
      itemHeight: 8,
    },
    xAxis: {
      type: 'category',
      data: dates,
      axisLabel: {
        fontSize: 10,
        color: '#999',
        formatter: (val: string) => {
          // Show only month-day or last digits
          const parts = val.split('-')
          if (parts.length === 3) return `${parts[1]}/${parts[2]}`
          return val
        },
      },
      axisLine: { lineStyle: { color: '#E8E3D5' } },
      axisTick: { show: false },
    },
    yAxis: [
      {
        type: 'value',
        name: '评分',
        min: 0,
        max: 100,
        interval: 20,
        axisLabel: {
          fontSize: 10,
          color: '#999',
        },
        splitLine: { lineStyle: { color: '#F0ECE0', type: 'dashed' } },
      },
      {
        type: 'value',
        name: '百分比',
        min: 0,
        max: 100,
        interval: 20,
        axisLabel: {
          fontSize: 10,
          color: '#999',
        },
        splitLine: { show: false },
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
      autoresize
    />
    <div v-else class="chart-empty">
      <p>暂无数据</p>
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
  align-items: center;
  justify-content: center;
  background: #F5F0E8;
  border-radius: 0.75rem;
  border: 1px dashed #E8E3D5;
}

.chart-empty p {
  color: #bbb;
  font-size: 0.9rem;
  margin: 0;
}
</style>
