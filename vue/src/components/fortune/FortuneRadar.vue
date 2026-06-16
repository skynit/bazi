<script setup lang="ts">
/**
 * FortuneRadar — five-element radar chart powered by echarts.
 * Reads `summary.element_distribution` (0-1 floats keyed 木火土金水).
 */
import { computed, ref, watch, onMounted, onUnmounted } from 'vue'
import * as echarts from 'echarts/core'
import { RadarChart } from 'echarts/charts'
import { TooltipComponent, LegendComponent, RadarComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'

echarts.use([RadarChart, TooltipComponent, LegendComponent, RadarComponent, CanvasRenderer])

interface Props {
  distribution: Record<string, number>
  height?: string
}
const props = withDefaults(defineProps<Props>(), { height: '260px' })

const ORDER = ['木', '火', '土', '金', '水'] as const

const host = ref<HTMLDivElement | null>(null)
let inst: echarts.ECharts | null = null
let themeObserver: MutationObserver | null = null

const series = computed(() =>
  ORDER.map(k => Math.round((props.distribution?.[k] ?? 0) * 1000) / 10)
)

function build() {
  if (!host.value) return
  if (!inst) inst = echarts.init(host.value)
  inst.setOption({
    tooltip: { trigger: 'item' },
    radar: {
      indicator: ORDER.map(k => ({ name: k, max: 60 })),
      shape: 'polygon',
      splitNumber: 4,
      axisLine: { lineStyle: { color: 'rgba(127,127,127,0.18)' } },
      splitLine: { lineStyle: { color: 'rgba(127,127,127,0.12)' } },
      splitArea: { areaStyle: { color: ['rgba(127,127,127,0.04)', 'rgba(127,127,127,0.02)'] } },
      axisName: { color: 'var(--text-muted)', fontSize: 12 }
    },
    series: [{
      type: 'radar',
      symbol: 'circle',
      symbolSize: 6,
      lineStyle: { width: 2, color: `rgba(${getRgb()}, 0.9)` },
      areaStyle: { color: `rgba(${getRgb()}, 0.18)` },
      itemStyle: { color: `rgba(${getRgb()}, 1)` },
      data: [{ value: series.value, name: '五行能量' }]
    }]
  })
}

function getRgb(): string {
  if (typeof window === 'undefined') return '52, 211, 153'
  const v = getComputedStyle(document.documentElement).getPropertyValue('--jade-accent-rgb').trim()
  return v || '52, 211, 153'
}

function resize() { inst?.resize() }

onMounted(() => {
  build()
  themeObserver = new MutationObserver(build)
  themeObserver.observe(document.documentElement, { attributes: true, attributeFilter: ['class', 'style', 'data-wuxing'] })
  window.addEventListener('resize', resize)
})
onUnmounted(() => {
  window.removeEventListener('resize', resize)
  themeObserver?.disconnect()
  themeObserver = null
  inst?.dispose()
  inst = null
})

watch(() => [props.distribution], build, { deep: true })
</script>

<template>
  <div ref="host" :style="{ height, width: '100%' }"></div>
</template>
