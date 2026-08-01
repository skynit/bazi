<script setup lang="ts">
/**
 * FortuneRadar — five-element radar chart powered by echarts.
 * Reads `summary.element_distribution` (0-1 floats keyed 木火土金水).
 * 所有颜色在 build 时从 CSS 变量解析，随主题与五行选择自适应。
 */
import { computed, ref, watch, onMounted, onUnmounted } from 'vue'
import * as echarts from 'echarts/core'
import { RadarChart } from 'echarts/charts'
import { AriaComponent, TooltipComponent, LegendComponent, RadarComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'

echarts.use([RadarChart, TooltipComponent, LegendComponent, RadarComponent, AriaComponent, CanvasRenderer])

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
  ORDER.map((key) => Math.round((props.distribution?.[key] ?? 0) * 1000) / 10),
)
const indicatorMax = computed(() =>
  Math.max(20, Math.ceil(Math.max(...series.value, 0) / 10) * 10),
)
const chartDescription = computed(() =>
  ORDER.map((key, index) => `${key}${series.value[index]}%`).join('，'),
)

function cssVar(name: string, fallback: string): string {
  if (typeof window === 'undefined') return fallback
  return getComputedStyle(document.documentElement).getPropertyValue(name).trim() || fallback
}

function getRgb(): string {
  return cssVar('--jade-accent-rgb', '52, 211, 153')
}

function build() {
  if (!host.value) return
  if (!inst) inst = echarts.init(host.value)
  const muted = cssVar('--text-muted', '#5a6a5e')
  const line = cssVar('--line-subtle', 'rgba(127,127,127,0.14)')
  const rgb = getRgb()
  inst.setOption({
    aria: { enabled: true, decal: { show: true }, description: chartDescription.value },
    tooltip: { trigger: 'item' },
    radar: {
      indicator: ORDER.map((key) => ({ name: key, max: indicatorMax.value })),
      shape: 'polygon',
      splitNumber: 4,
      radius: '68%',
      axisLine: { lineStyle: { color: line } },
      splitLine: { lineStyle: { color: line } },
      splitArea: { show: false },
      axisName: { color: muted, fontSize: 12, fontFamily: cssVar('--font-serif', 'serif') },
      axisNameGap: 10,
    },
    series: [
      {
        type: 'radar',
        symbol: 'circle',
        symbolSize: 5,
        lineStyle: { width: 2, color: `rgba(${rgb}, 0.9)` },
        areaStyle: { color: `rgba(${rgb}, 0.14)` },
        itemStyle: { color: `rgba(${rgb}, 1)` },
        data: [{ value: series.value, name: '五行样本频次' }],
      },
    ],
  })
}

function resize() {
  inst?.resize()
}

onMounted(() => {
  build()
  themeObserver = new MutationObserver(build)
  themeObserver.observe(document.documentElement, {
    attributes: true,
    attributeFilter: ['class', 'style', 'data-wuxing'],
  })
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
  <div
    ref="host"
    role="img"
    :aria-label="`五行样本频次雷达图：${chartDescription}`"
    :style="{ height, width: '100%' }"
  ></div>
</template>
