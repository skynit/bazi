<script setup lang="ts">
import { computed } from 'vue'

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

// Five-element mapping for heavenly stems
const ganElement: Record<string, { name: string; color: string; bg: string }> = {
  '甲': { name: '木', color: 'text-green-700', bg: 'glass-card border-green-300' },
  '乙': { name: '木', color: 'text-green-700', bg: 'glass-card border-green-300' },
  '丙': { name: '火', color: 'text-red-600', bg: 'glass-card border-red-300' },
  '丁': { name: '火', color: 'text-red-600', bg: 'glass-card border-red-300' },
  '戊': { name: '土', color: 'text-yellow-700', bg: 'glass-card border-yellow-300' },
  '己': { name: '土', color: 'text-yellow-700', bg: 'glass-card border-yellow-300' },
  '庚': { name: '金', color: 'text-amber-600', bg: 'glass-card border-amber-300' },
  '辛': { name: '金', color: 'text-amber-600', bg: 'glass-card border-amber-300' },
  '壬': { name: '水', color: 'text-blue-700', bg: 'glass-card border-blue-300' },
  '癸': { name: '水', color: 'text-blue-700', bg: 'glass-card border-blue-300' },
}

// Five-element mapping for earthly branches
const zhiElement: Record<string, { name: string; color: string; bg: string }> = {
  '寅': { name: '木', color: 'text-green-700', bg: 'glass-card border-green-300' },
  '卯': { name: '木', color: 'text-green-700', bg: 'glass-card border-green-300' },
  '巳': { name: '火', color: 'text-red-600', bg: 'glass-card border-red-300' },
  '午': { name: '火', color: 'text-red-600', bg: 'glass-card border-red-300' },
  '辰': { name: '土', color: 'text-yellow-700', bg: 'glass-card border-yellow-300' },
  '戌': { name: '土', color: 'text-yellow-700', bg: 'glass-card border-yellow-300' },
  '丑': { name: '土', color: 'text-yellow-700', bg: 'glass-card border-yellow-300' },
  '未': { name: '土', color: 'text-yellow-700', bg: 'glass-card border-yellow-300' },
  '申': { name: '金', color: 'text-amber-600', bg: 'glass-card border-amber-300' },
  '酉': { name: '金', color: 'text-amber-600', bg: 'glass-card border-amber-300' },
  '亥': { name: '水', color: 'text-blue-700', bg: 'glass-card border-blue-300' },
  '子': { name: '水', color: 'text-blue-700', bg: 'glass-card border-blue-300' },
}

// Six clashes (六冲) - red dashed border
const clashPairs: [string, string][] = [
  ['子', '午'], ['丑', '未'], ['寅', '申'],
  ['卯', '酉'], ['辰', '戌'], ['巳', '亥'],
]

// Six harmonies (六合) - green
const harmonyPairs: [string, string][] = [
  ['子', '丑'], ['寅', '亥'], ['卯', '戌'],
  ['辰', '酉'], ['巳', '申'], ['午', '未'],
]

function doZhiClash(a: string, b: string): boolean {
  return clashPairs.some(([x, y]) => (a === x && b === y) || (a === y && b === x))
}

function doZhiHarmony(a: string, b: string): boolean {
  return harmonyPairs.some(([x, y]) => (a === x && b === y) || (a === y && b === x))
}

const pillars = computed(() => [
  { label: '年柱', key: 'year' as const, ...props.chart.year_pillar },
  { label: '月柱', key: 'month' as const, ...props.chart.month_pillar },
  { label: '日柱', key: 'day' as const, ...props.chart.day_pillar },
  { label: '时柱', key: 'hour' as const, ...props.chart.hour_pillar },
])

// Pairs of adjacent pillars for clash/harmony checks
const relationships = computed(() => {
  const result: { i: number; j: number; type: 'clash' | 'harmony' }[] = []
  for (let i = 0; i < pillars.value.length; i++) {
    for (let j = i + 1; j < pillars.value.length; j++) {
      if (doZhiClash(pillars.value[i].zhi, pillars.value[j].zhi)) {
        result.push({ i, j, type: 'clash' })
      } else if (doZhiHarmony(pillars.value[i].zhi, pillars.value[j].zhi)) {
        result.push({ i, j, type: 'harmony' })
      }
    }
  }
  return result
})

const elemColor = (e: string) => ({ 金: "#DAA520", 木: "#228B22", 水: "#4169E1", 火: "#DC143C", 土: "#8B4513" }[e] || "#999")

</script>

<template>
  <div class="glass-card overflow-hidden">
    <!-- Title -->
    <div style="background:rgba(255,255,255,0.05);color:#EAE6DD" class="text-center py-3">
      <h2 class="text-lg font-bold tracking-widest">八字命盘</h2>
    </div>

    <!-- Four pillars grid -->
    <div class="grid grid-cols-4 divide-x divide-bazi-red/10">
      <div
        v-for="pillar in pillars"
        :key="pillar.key"
        class="text-center"
      >
        <!-- Pillar label -->
        <div style="background:rgba(255,255,255,0.02)" class="py-2 text-sm font-medium text-bazi-blue/70 tracking-wide">
          {{ pillar.label }}
        </div>

        <!-- Gan (heavenly stem) -->
        <div
          style="border-color:rgba(212,168,75,0.08)" class="py-4 mx-3"
          :class="ganElement[pillar.gan]?.color"
        >
          <div class="text-4xl font-bold font-serif">
            {{ pillar.gan }}
          </div>
          <div class="text-xs mt-1 opacity-70">
            {{ ganElement[pillar.gan]?.name || '—' }}
          </div>
        </div>

        <!-- Zhi (earthly branch) -->
        <div
          class="py-4 mx-3"
          :class="zhiElement[pillar.zhi]?.color"
        >
          <div class="text-4xl font-bold font-serif">
            {{ pillar.zhi }}
          </div>
          <div class="text-xs mt-1 opacity-70">
            {{ zhiElement[pillar.zhi]?.name || '—' }}
          </div>
        </div>
      </div>
    </div>

    <!-- Relationships legend -->
    <div
      v-if="relationships.length > 0"
      style="border-top:1px solid rgba(212,168,75,0.1)" class="px-6 py-4"
    >
      <h4 class="text-sm font-medium text-bazi-blue/70 mb-3">地支关系</h4>
      <div class="flex flex-wrap gap-3">
        <div
          v-for="(rel, ri) in relationships"
          :key="ri"
          class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-full text-sm font-medium"
          :class="
            rel.type === 'clash'
              ? 'glass-card text-red-700 border border-red-200'
              : 'glass-card text-green-700 border border-green-200'
          "
        >
          <span>{{ pillars[rel.i].label }}</span>
          <span class="font-bold">{{ rel.type === 'clash' ? '冲' : '合' }}</span>
          <span>{{ pillars[rel.j].label }}</span>
        </div>
      </div>
    </div>

    <!-- No relationships found -->
    <div v-else style="border-top:1px solid rgba(212,168,75,0.1)" class="px-6 py-4 text-center text-sm text-bazi-blue/40">
      地支无特殊冲合关系
    </div>
    <!-- Analysis sections -->
    <div style="border-top:1px solid rgba(212,168,75,0.1)" class="px-6 py-4 space-y-4">
      <!-- Five Elements -->
      <div v-if="chart.five_elements">
        <h4 class="text-sm font-bold text-bazi-blue mb-2">五行分布</h4>
        <div class="flex gap-2">
          <div v-for="(val, elem) in chart.five_elements" :key="elem"
            class="flex-1 text-center py-2 rounded text-xs font-medium text-white"
            :style="{ background: elemColor(String(elem)) }">
            {{ elem }}{{ val }}
          </div>
        </div>
      </div>

      <!-- Ten Gods -->
      <div v-if="chart.ten_gods">
        <h4 class="text-sm font-bold text-bazi-blue mb-2">十神</h4>
        <div class="grid grid-cols-4 gap-2 text-xs">
          <div v-for="(god, pillar) in chart.ten_gods" :key="pillar"
            style="background:rgba(255,255,255,0.02)" class="rounded px-2 py-1 text-center">
            <span class="text-bazi-blue/50">{{ pillar }}</span>
            <span class="text-bazi-red font-medium ml-1">{{ god }}</span>
          </div>
        </div>
      </div>

      <!-- NaYin -->
      <div v-if="chart.na_yin">
        <h4 class="text-sm font-bold text-bazi-blue mb-2">纳音</h4>
        <div class="flex gap-2 text-xs flex-wrap">
           <span v-for="(v,k) in chart.na_yin" :key="k" style="background:rgba(255,255,255,0.02)" class="rounded px-2 py-1">{{k}}:{{v}}</span>
        </div>
      </div>

      <!-- DaYun -->
      <div v-if="chart.da_yun && chart.da_yun.start_age">
        <h4 class="text-sm font-bold text-bazi-blue mb-2">
          大运（{{ chart.da_yun.direction }} · {{ chart.da_yun.start_age }}岁起运）
        </h4>
        <div class="flex flex-wrap gap-1">
          <span v-for="(p,i) in chart.da_yun.pillars" :key="i"
            class="bg-bazi-red/5 text-bazi-red text-xs rounded px-2 py-1">
            {{ (chart.da_yun.start_age || 0) + i*10 }}岁 {{ p.gan }}{{ p.zhi }}
          </span>
        </div>
      </div>
    </div>

  </div>
</template>
