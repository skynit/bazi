<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'

interface StarInfo {
  name: string
  type: string
  scope: string
  brightness: string
}

interface PalaceData {
  name: string
  branch: string
  heavenly_stem: string
  is_body_palace: boolean
  stars: StarInfo[]
  four_hua: string[]
  adjective_stars?: string[]
  changsheng_12?: string
  boshi_12?: string
  jiang_qian_12?: string
  sui_qian_12?: string
}

interface Props {
  baseChart: {
    palaces: PalaceData[]
    life_master: string
    body_master: string
    five_bureau: string
    earthly_branch_of_soul_palace: string
    earthly_branch_of_body_palace: string
  }
  liunianChart: {
    palaces: PalaceData[]
    year: number
    liu_nian_stars?: string[][]
  }
  availableYears: number[]
}

const props = defineProps<Props>()

const emit = defineEmits<{
  (e: 'year-change', year: number): void
}>()

const mode = ref<'base' | 'overlay'>('base')
const selectedYear = ref<number>(new Date().getFullYear())

const isDark = ref(document.documentElement.classList.contains('dark'))
let themeObserver: MutationObserver | null = null

onMounted(() => {
  themeObserver = new MutationObserver(() => {
    isDark.value = document.documentElement.classList.contains('dark')
  })
  themeObserver.observe(document.documentElement, { attributes: true, attributeFilter: ['class'] })
})

onUnmounted(() => {
  themeObserver?.disconnect()
  themeObserver = null
})

// Gold palette for 本命盘 — light mode
const goldMetaLight: Record<string, { bg: string; text: string }> = {
  '庙': { bg: 'linear-gradient(135deg,#e11d48,#be123c)', text: '#fffaf8' },
  '旺': { bg: 'linear-gradient(135deg,#ea580c,#c2410c)', text: '#fffaf8' },
  '得': { bg: 'linear-gradient(135deg,#eab308,#a16207)', text: '#fffaf8' },
  '利': { bg: 'linear-gradient(135deg,#16a34a,#15803d)', text: '#fffaf8' },
  '平': { bg: 'linear-gradient(135deg,#6b7280,#4b5563)', text: '#fffaf8' },
  '不': { bg: 'linear-gradient(135deg,#0e7490,#0c5a74)', text: '#fffaf8' },
  '陷': { bg: 'linear-gradient(135deg,#44403c,#292524)', text: '#e7e5e4' },
}

// Gold palette for 本命盘 — dark mode
const goldMetaDark: Record<string, { bg: string; text: string }> = {
  '庙': { bg: 'linear-gradient(135deg,#fb7185,#be123c)', text: '#fffaf8' },
  '旺': { bg: 'linear-gradient(135deg,#FF8C00,#CC5500)', text: '#fffaf8' },
  '得': { bg: 'linear-gradient(135deg,#fde68a,#94a3b8)', text: '#10140f' },
  '利': { bg: 'linear-gradient(135deg,#34d399,#059669)', text: '#00140e' },
  '平': { bg: 'linear-gradient(135deg,#808080,#696969)', text: '#fffaf8' },
  '不': { bg: 'linear-gradient(135deg,#5F9EA0,#4682B4)', text: '#fffaf8' },
  '陷': { bg: 'linear-gradient(135deg,#2B3A42,#1a252e)', text: '#dbe4e8' },
}

// Purple palette for 流年盘 — light mode
const purpleMetaLight: Record<string, { bg: string; text: string }> = {
  '庙': { bg: 'linear-gradient(135deg,#7c3aed,#6d28d9)', text: '#fffaf8' },
  '旺': { bg: 'linear-gradient(135deg,#8b5cf6,#7c3aed)', text: '#fffaf8' },
  '得': { bg: 'linear-gradient(135deg,#a78bfa,#8b5cf6)', text: '#1a1030' },
  '利': { bg: 'linear-gradient(135deg,#6d28d9,#5b21b6)', text: '#fffaf8' },
  '平': { bg: 'linear-gradient(135deg,#6b5b95,#5a4a84)', text: '#fffaf8' },
  '不': { bg: 'linear-gradient(135deg,#5b4a84,#4a3a7a)', text: '#fffaf8' },
  '陷': { bg: 'linear-gradient(135deg,#3d3060,#2d2050)', text: '#d4c8f0' },
}

// Purple palette for 流年盘 — dark mode
const purpleMetaDark: Record<string, { bg: string; text: string }> = {
  '庙': { bg: 'linear-gradient(135deg,#7B2D8B,#4B0082)', text: '#fffaf8' },
  '旺': { bg: 'linear-gradient(135deg,#9B59B6,#8E44AD)', text: '#fffaf8' },
  '得': { bg: 'linear-gradient(135deg,#8E6DBB,#6B5B95)', text: '#fffaf8' },
  '利': { bg: 'linear-gradient(135deg,#5D4B8B,#4A3A7A)', text: '#fffaf8' },
  '平': { bg: 'linear-gradient(135deg,#7A6B9B,#655580)', text: '#fffaf8' },
  '不': { bg: 'linear-gradient(135deg,#6B5B95,#5A4A84)', text: '#f0edf7' },
  '陷': { bg: 'linear-gradient(135deg,#3D2B5B,#2D1B4A)', text: '#e3dff0' },
}

function baseMeta(brightness: string) {
  const meta = isDark.value ? goldMetaDark : goldMetaLight
  return meta[brightness] || meta['陷']
}

function overlayMeta(brightness: string) {
  const meta = isDark.value ? purpleMetaDark : purpleMetaLight
  return meta[brightness] || meta['陷']
}

function onYearChange() {
  emit('year-change', selectedYear.value)
}

// Sync selectedYear when prop changes (e.g., after year switch)
watch(() => props.liunianChart?.year, (y) => {
  if (y) selectedYear.value = y
}, { immediate: true })

const baseLookup = computed<Record<string, PalaceData>>(() => {
  const m: Record<string, PalaceData> = {}
  props.baseChart.palaces.forEach((p) => { m[p.branch] = p })
  return m
})

// Use a reactive ref for liunianChart to ensure Vue tracks it properly
const liunianPalaces = computed(() => props.liunianChart?.palaces || [])
const liunianLookup = computed<Record<string, PalaceData>>(() => {
  const m: Record<string, PalaceData> = {}
  liunianPalaces.value.forEach((p) => { m[p.branch] = p })
  return m
})
// liunianStars indexed by palace index
const liunianStarsMap = computed<Record<string, string[]>>(() => {
  const m: Record<string, string[]> = {}
  const stars = props.liunianChart?.liu_nian_stars || []
  const palaces = props.liunianChart?.palaces || []
  for (let i = 0; i < palaces.length; i++) {
    m[palaces[i].branch] = stars[i] || []
  }
  return m
})

const branchIndexMap: Record<string, number> = {
  '子': 0, '丑': 1, '寅': 2, '卯': 3, '辰': 4, '巳': 5,
  '午': 6, '未': 7, '申': 8, '酉': 9, '戌': 10, '亥': 11,
}
const indexBranchMap = ['子', '丑', '寅', '卯', '辰', '巳', '午', '未', '申', '酉', '戌', '亥']

function fixIdx(i: number): number {
  return ((i % 12) + 12) % 12
}

function sanfangBranches(branch: string): { opposite: string; trine1: string; trine2: string } {
  const idx = branchIndexMap[branch]
  if (idx === undefined) return { opposite: '', trine1: '', trine2: '' }
  return {
    opposite: indexBranchMap[fixIdx(idx + 6)],
    trine1: indexBranchMap[fixIdx(idx + 4)],
    trine2: indexBranchMap[fixIdx(idx - 4)],
  }
}

const focusedBranch = ref<string | undefined>(undefined)

function palaceHighlightClass(branch: string): string {
  if (!focusedBranch.value) return ''
  if (branch === focusedBranch.value) return 'zw-focused'
  const sf = sanfangBranches(focusedBranch.value)
  if (branch === sf.opposite) return 'zw-opposite'
  if (branch === sf.trine1 || branch === sf.trine2) return 'zw-surrounded'
  return ''
}

const gridRef = ref<HTMLElement | null>(null)

type Point = { x: number; y: number }

const branchAnchorPercent: Record<string, Point> = {
  '巳': { x: 12.5, y: 12.5 },
  '午': { x: 37.5, y: 12.5 },
  '未': { x: 62.5, y: 12.5 },
  '申': { x: 87.5, y: 12.5 },
  '辰': { x: 12.5, y: 37.5 },
  '酉': { x: 87.5, y: 37.5 },
  '卯': { x: 12.5, y: 62.5 },
  '戌': { x: 87.5, y: 62.5 },
  '寅': { x: 12.5, y: 87.5 },
  '丑': { x: 37.5, y: 87.5 },
  '子': { x: 62.5, y: 87.5 },
  '亥': { x: 87.5, y: 87.5 },
}

const svgLines = computed<{ from: Point; to: Point; type: 'opposite' | 'trine' }[]>(() => {
  if (!focusedBranch.value) return []
  const sf = sanfangBranches(focusedBranch.value)
  const from = branchAnchorPercent[focusedBranch.value]
  if (!from) return []
  const lines: { from: Point; to: Point; type: 'opposite' | 'trine' }[] = []
  const opp = branchAnchorPercent[sf.opposite]
  const t1 = branchAnchorPercent[sf.trine1]
  const t2 = branchAnchorPercent[sf.trine2]
  if (opp) lines.push({ from, to: opp, type: 'opposite' })
  if (t1) lines.push({ from, to: t1, type: 'trine' })
  if (t2) lines.push({ from, to: t2, type: 'trine' })
  return lines
})

const branchOrder = ['巳', '午', '未', '申', '辰', '卯', '酉', '戌', '寅', '丑', '子', '亥']
const row1Branches = branchOrder.slice(0, 4)
const row4Branches = branchOrder.slice(8, 12)

interface RichStar { name: string; brightness: string }
function richStars(p: PalaceData | undefined): RichStar[] {
  if (!p) return []
  return p.stars || []
}
function liunianStars(p: PalaceData | undefined): RichStar[] {
  if (!p) return []
  return p.stars || []
}
function palaceSihua(p: PalaceData | undefined): string[] {
  return p?.four_hua || []
}

function basePalaceAt(b: string): PalaceData | undefined { return baseLookup.value[b] }
function liunianPalaceAt(b: string): PalaceData | undefined {
  if (!props.liunianChart) return undefined
  return liunianLookup.value[b]
}
// liunianStarsAt returns the 流年星耀 for a given branch
function liunianStarsAt(b: string): string[] {
  return liunianStarsMap.value[b] || []
}
</script>

<template>
  <div class="zw-overlay">
    <!-- Controls -->
    <div class="zw-controls">
      <div class="zw-toggle">
        <button class="zw-tab" :class="{ 'is-active': mode === 'base' }" @click="mode = 'base'">
          <span class="zw-tab-dot zw-dot-gold"></span>
          本命盘
        </button>
        <button class="zw-tab" :class="{ 'is-active': mode === 'overlay' }" @click="mode = 'overlay'">
          <span class="zw-tab-dot zw-dot-purple"></span>
          流年叠盘
        </button>
      </div>

      <div v-if="mode === 'overlay'" class="zw-year-select">
        <span class="zw-year-label">流年</span>
        <select v-model="selectedYear" class="zw-select" @change="onYearChange">
          <option v-for="y in availableYears" :key="y" :value="y">{{ y }}年</option>
        </select>
      </div>
    </div>

    <!-- Chart grid -->
    <div class="zw-grid-wrap">
      <div class="zw-grid" :class="{ 'zw-grid-overlay': mode === 'overlay' }" ref="gridRef">

      <!-- Row 1 -->
      <template v-for="branch in row1Branches" :key="'r1-' + branch">
        <div class="zw-cell" :class="[{ 'zw-cell-overlay': mode === 'overlay', 'zw-cell-body': basePalaceAt(branch)?.is_body_palace }, palaceHighlightClass(branch)]" @mouseenter="focusedBranch = branch" @mouseleave="focusedBranch = undefined">
          <div class="zw-cell-header">
            <span class="zw-palace-name">{{ basePalaceAt(branch)?.name || branch }}</span>
            <span class="zw-branch">{{ branch }}<template v-if="basePalaceAt(branch)?.heavenly_stem"> · {{ basePalaceAt(branch)?.heavenly_stem }}</template></span>
            <span v-if="basePalaceAt(branch)?.is_body_palace" class="zw-body-tag">身宫</span>
          </div>
          <div class="zw-stars">
            <span
              v-for="(star, si) in richStars(basePalaceAt(branch))"
              :key="'bs-' + si"
              class="zw-star zw-star-gold"
              :style="{ background: baseMeta(star.brightness).bg }"
            >{{ star.name }}</span>
            <span v-for="(sh, si) in palaceSihua(basePalaceAt(branch))" :key="'bsh-' + si" class="zw-sihua-tag">{{ sh }}</span>
          </div>
          <div v-if="basePalaceAt(branch)?.changsheng_12" class="zw-twelve"><span class="zw-twelve-tag">{{ basePalaceAt(branch)?.changsheng_12 }}</span></div>
          <div v-if="mode === 'overlay' && liunianPalaceAt(branch)" class="zw-overlay-stars">
            <div class="zw-overlay-label">流年</div>
            <span
              v-for="(star, si) in liunianStars(liunianPalaceAt(branch))"
              :key="'ls-' + si"
              class="zw-star zw-star-purple"
              :style="{ background: overlayMeta(star.brightness).bg }"
            >{{ star.name }}</span>
          </div>
          <div v-if="mode === 'overlay' && liunianStarsAt(branch).length" class="zw-liuyao">
            <span
              v-for="(star, si) in liunianStarsAt(branch)"
              :key="'ly-' + si"
              class="zw-liuyao-chip"
            >{{ star }}</span>
          </div>
        </div>
      </template>

      <!-- Row 2: 辰 + center + 酉 -->
      <div class="zw-cell" :class="[{ 'zw-cell-overlay': mode === 'overlay', 'zw-cell-body': basePalaceAt('辰')?.is_body_palace }, palaceHighlightClass('辰')]" @mouseenter="focusedBranch = '辰'" @mouseleave="focusedBranch = undefined">
        <div class="zw-cell-header">
          <span class="zw-palace-name">{{ basePalaceAt('辰')?.name || '辰' }}</span>
          <span class="zw-branch">辰<template v-if="basePalaceAt('辰')?.heavenly_stem"> · {{ basePalaceAt('辰')?.heavenly_stem }}</template></span>
          <span v-if="basePalaceAt('辰')?.is_body_palace" class="zw-body-tag">身宫</span>
        </div>
        <div class="zw-stars">
<span
              v-for="(star, si) in richStars(basePalaceAt('辰'))"
              :key="'bs-' + si"
              class="zw-star zw-star-gold"
              :style="{ background: baseMeta(star.brightness).bg }"
            >{{ star.name }}</span>
        </div>
        <div v-if="basePalaceAt('辰')?.changsheng_12" class="zw-twelve"><span class="zw-twelve-tag">{{ basePalaceAt('辰')?.changsheng_12 }}</span></div>
        <div v-if="mode === 'overlay' && liunianPalaceAt('辰')" class="zw-overlay-stars">
          <div class="zw-overlay-label">流年</div>
          <span
            v-for="(star, si) in (liunianStars(liunianPalaceAt('辰')))"
            :key="'ls-' + si"
            class="zw-star zw-star-purple"
            :style="{ background: overlayMeta(star.brightness).bg }"
          >{{ star.name }}</span>
        </div>
        <div v-if="mode === 'overlay' && liunianStarsAt('辰').length" class="zw-liuyao">
          <span v-for="(star, si) in liunianStarsAt('辰')" :key="'ly-' + si" class="zw-liuyao-chip">{{ star }}</span>
        </div>
      </div>

      <!-- Center: 命宫核心 -->
      <div class="zw-center" :class="{ 'zw-center-overlay': mode === 'overlay' }">
        <div class="zw-center-glow"></div>
        <div class="zw-center-title">命宫核心</div>
        <div class="zw-center-row">
          <span class="zw-center-lbl">命主</span>
          <span class="zw-center-val">{{ baseChart.life_master }}</span>
        </div>
        <div class="zw-center-row">
          <span class="zw-center-lbl">身主</span>
          <span class="zw-center-val">{{ baseChart.body_master }}</span>
        </div>
        <div class="zw-center-row">
          <span class="zw-center-lbl">五行局</span>
          <span class="zw-center-val">{{ baseChart.five_bureau }}</span>
        </div>
        <div v-if="mode === 'overlay'" class="zw-center-row zw-center-year">
          <span class="zw-center-lbl">流年</span>
          <span class="zw-center-val zw-year-val">{{ selectedYear }}</span>
        </div>
      </div>

      <div class="zw-cell" :class="[{ 'zw-cell-overlay': mode === 'overlay', 'zw-cell-body': basePalaceAt('酉')?.is_body_palace }, palaceHighlightClass('酉')]" @mouseenter="focusedBranch = '酉'" @mouseleave="focusedBranch = undefined">
        <div class="zw-cell-header">
          <span class="zw-palace-name">{{ basePalaceAt('酉')?.name || '酉' }}</span>
          <span class="zw-branch">酉<template v-if="basePalaceAt('酉')?.heavenly_stem"> · {{ basePalaceAt('酉')?.heavenly_stem }}</template></span>
          <span v-if="basePalaceAt('酉')?.is_body_palace" class="zw-body-tag">身宫</span>
        </div>
        <div class="zw-stars">
          <span
            v-for="(star, si) in (richStars(basePalaceAt('酉')))"
            :key="'bs-' + si"
            class="zw-star zw-star-gold"
            :style="{ background: baseMeta(star.brightness).bg }"
          >{{ star.name }}</span>
        </div>
        <div v-if="basePalaceAt('酉')?.changsheng_12" class="zw-twelve"><span class="zw-twelve-tag">{{ basePalaceAt('酉')?.changsheng_12 }}</span></div>
        <div v-if="mode === 'overlay' && liunianPalaceAt('酉')" class="zw-overlay-stars">
          <div class="zw-overlay-label">流年</div>
          <span
            v-for="(star, si) in (liunianStars(liunianPalaceAt('酉')))"
            :key="'ls-' + si"
            class="zw-star zw-star-purple"
            :style="{ background: overlayMeta(star.brightness).bg }"
          >{{ star.name }}</span>
        </div>
        <div v-if="mode === 'overlay' && liunianStarsAt('酉').length" class="zw-liuyao">
          <span v-for="(star, si) in liunianStarsAt('酉')" :key="'ly-' + si" class="zw-liuyao-chip">{{ star }}</span>
        </div>
      </div>

      <!-- Row 3: 卯 (cols 2-3 taken by zw-center) -->
      <div class="zw-cell" :class="[{ 'zw-cell-overlay': mode === 'overlay', 'zw-cell-body': basePalaceAt('卯')?.is_body_palace }, palaceHighlightClass('卯')]" @mouseenter="focusedBranch = '卯'" @mouseleave="focusedBranch = undefined">
        <div class="zw-cell-header">
          <span class="zw-palace-name">{{ basePalaceAt('卯')?.name || '卯' }}</span>
          <span class="zw-branch">卯<template v-if="basePalaceAt('卯')?.heavenly_stem"> · {{ basePalaceAt('卯')?.heavenly_stem }}</template></span>
          <span v-if="basePalaceAt('卯')?.is_body_palace" class="zw-body-tag">身宫</span>
        </div>
        <div class="zw-stars">
          <span
            v-for="(star, si) in (richStars(basePalaceAt('卯')))"
            :key="'bs-' + si"
            class="zw-star zw-star-gold"
            :style="{ background: baseMeta(star.brightness).bg }"
          >{{ star.name }}</span>
        </div>
        <div v-if="basePalaceAt('卯')?.changsheng_12" class="zw-twelve"><span class="zw-twelve-tag">{{ basePalaceAt('卯')?.changsheng_12 }}</span></div>
        <div v-if="mode === 'overlay' && liunianPalaceAt('卯')" class="zw-overlay-stars">
          <div class="zw-overlay-label">流年</div>
          <span
            v-for="(star, si) in (liunianStars(liunianPalaceAt('卯')))"
            :key="'ls-' + si"
            class="zw-star zw-star-purple"
            :style="{ background: overlayMeta(star.brightness).bg }"
          >{{ star.name }}</span>
        </div>
        <div v-if="mode === 'overlay' && liunianStarsAt('卯').length" class="zw-liuyao">
          <span v-for="(star, si) in liunianStarsAt('卯')" :key="'ly-' + si" class="zw-liuyao-chip">{{ star }}</span>
        </div>
      </div>

      <!-- cols 2-3 occupied by zw-center -->

      <div class="zw-cell" :class="[{ 'zw-cell-overlay': mode === 'overlay', 'zw-cell-body': basePalaceAt('戌')?.is_body_palace }, palaceHighlightClass('戌')]" @mouseenter="focusedBranch = '戌'" @mouseleave="focusedBranch = undefined">
        <div class="zw-cell-header">
          <span class="zw-palace-name">{{ basePalaceAt('戌')?.name || '戌' }}</span>
          <span class="zw-branch">戌<template v-if="basePalaceAt('戌')?.heavenly_stem"> · {{ basePalaceAt('戌')?.heavenly_stem }}</template></span>
          <span v-if="basePalaceAt('戌')?.is_body_palace" class="zw-body-tag">身宫</span>
        </div>
        <div class="zw-stars">
          <span
            v-for="(star, si) in (richStars(basePalaceAt('戌')))"
            :key="'bs-' + si"
            class="zw-star zw-star-gold"
            :style="{ background: baseMeta(star.brightness).bg }"
          >{{ star.name }}</span>
        </div>
        <div v-if="basePalaceAt('戌')?.changsheng_12" class="zw-twelve"><span class="zw-twelve-tag">{{ basePalaceAt('戌')?.changsheng_12 }}</span></div>
        <div v-if="mode === 'overlay' && liunianPalaceAt('戌')" class="zw-overlay-stars">
          <div class="zw-overlay-label">流年</div>
          <span
            v-for="(star, si) in (liunianStars(liunianPalaceAt('戌')))"
            :key="'ls-' + si"
            class="zw-star zw-star-purple"
            :style="{ background: overlayMeta(star.brightness).bg }"
          >{{ star.name }}</span>
        </div>
        <div v-if="mode === 'overlay' && liunianStarsAt('戌').length" class="zw-liuyao">
          <span v-for="(star, si) in liunianStarsAt('戌')" :key="'ly-' + si" class="zw-liuyao-chip">{{ star }}</span>
        </div>
      </div>

      <!-- Row 4 -->
      <template v-for="branch in row4Branches" :key="'r4-' + branch">
        <div class="zw-cell" :class="[{ 'zw-cell-overlay': mode === 'overlay', 'zw-cell-body': basePalaceAt(branch)?.is_body_palace }, palaceHighlightClass(branch)]" @mouseenter="focusedBranch = branch" @mouseleave="focusedBranch = undefined">
          <div class="zw-cell-header">
            <span class="zw-palace-name">{{ basePalaceAt(branch)?.name || branch }}</span>
            <span class="zw-branch">{{ branch }}<template v-if="basePalaceAt(branch)?.heavenly_stem"> · {{ basePalaceAt(branch)?.heavenly_stem }}</template></span>
            <span v-if="basePalaceAt(branch)?.is_body_palace" class="zw-body-tag">身宫</span>
          </div>
          <div class="zw-stars">
            <span
              v-for="(star, si) in richStars(basePalaceAt(branch))"
              :key="'bs-' + si"
              class="zw-star zw-star-gold"
              :style="{ background: baseMeta(star.brightness).bg }"
            >{{ star.name }}</span>
            <span v-for="(sh, si) in palaceSihua(basePalaceAt(branch))" :key="'bsh-' + si" class="zw-sihua-tag">{{ sh }}</span>
          </div>
          <div v-if="basePalaceAt(branch)?.changsheng_12" class="zw-twelve"><span class="zw-twelve-tag">{{ basePalaceAt(branch)?.changsheng_12 }}</span></div>
          <div v-if="mode === 'overlay' && liunianPalaceAt(branch)" class="zw-overlay-stars">
            <div class="zw-overlay-label">流年</div>
            <span
              v-for="(star, si) in liunianStars(liunianPalaceAt(branch))"
              :key="'ls-' + si"
              class="zw-star zw-star-purple"
              :style="{ background: overlayMeta(star.brightness).bg }"
            >{{ star.name }}</span>
          </div>
          <div v-if="mode === 'overlay' && liunianStarsAt(branch).length" class="zw-liuyao">
            <span
              v-for="(star, si) in liunianStarsAt(branch)"
              :key="'ly-' + si"
              class="zw-liuyao-chip"
            >{{ star }}</span>
          </div>
        </div>
      </template>
      </div>

      <!-- SVG connection lines for sanfang sizheng -->
      <svg
        v-if="svgLines.length"
        class="zw-svg-overlay"
        viewBox="0 0 100 100"
        preserveAspectRatio="none"
      >
        <defs>
          <filter id="glow-line" x="-50%" y="-50%" width="200%" height="200%">
            <feGaussianBlur stdDeviation="0.8" result="blur" />
            <feMerge>
              <feMergeNode in="blur" />
              <feMergeNode in="SourceGraphic" />
            </feMerge>
          </filter>
          <linearGradient id="grad-opposite" x1="0%" y1="0%" x2="100%" y2="0%">
            <stop offset="0%" stop-color="rgba(186,130,255,0.95)" />
            <stop offset="100%" stop-color="rgba(186,130,255,0.3)" />
          </linearGradient>
          <linearGradient id="grad-trine" x1="0%" y1="0%" x2="100%" y2="0%">
            <stop offset="0%" stop-color="rgba(206,168,255,0.85)" />
            <stop offset="100%" stop-color="rgba(206,168,255,0.2)" />
          </linearGradient>
        </defs>
        <line
          v-for="(l, i) in svgLines"
          :key="'line-' + i"
          :x1="l.from.x" :y1="l.from.y"
          :x2="l.to.x" :y2="l.to.y"
          :stroke="l.type === 'opposite' ? 'url(#grad-opposite)' : 'url(#grad-trine)'"
          :stroke-width="l.type === 'opposite' ? 0.7 : 0.5"
          stroke-dasharray="120"
          stroke-linecap="round"
          filter="url(#glow-line)"
          class="zw-svg-line"
          :class="{ 'zw-svg-line-opp': l.type === 'opposite' }"
        />
        <circle
          v-for="(l, i) in svgLines"
          :key="'dot-to-' + i"
          :cx="l.to.x" :cy="l.to.y"
          r="0.8"
          :fill="l.type === 'opposite' ? 'rgba(186,130,255,0.8)' : 'rgba(206,168,255,0.6)'"
          filter="url(#glow-line)"
        />
        <circle
          v-for="(l, i) in svgLines"
          :key="'dot-from-' + i"
          :cx="l.from.x" :cy="l.from.y"
          r="0.8"
          fill="rgba(142,109,187,0.8)"
          filter="url(#glow-line)"
        />
      </svg>
    </div>

    <!-- Legend -->
    <div class="zw-legend">
      <div class="zw-legend-item">
        <span class="zw-legend-swatch zw-swatch-gold"></span>
        <span class="zw-legend-text">本命星曜</span>
      </div>
      <div v-if="mode === 'overlay'" class="zw-legend-item">
        <span class="zw-legend-swatch zw-swatch-purple"></span>
        <span class="zw-legend-text">流年星曜</span>
      </div>
      <div class="zw-legend-item">
        <span class="zw-legend-swatch zw-swatch-focused"></span>
        <span class="zw-legend-text">本宫</span>
      </div>
      <div class="zw-legend-item">
        <span class="zw-legend-swatch zw-swatch-opposite"></span>
        <span class="zw-legend-text">对宫</span>
      </div>
      <div class="zw-legend-item">
        <span class="zw-legend-swatch zw-swatch-surrounded"></span>
        <span class="zw-legend-text">三合</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* ── Page ── */
.zw-overlay {
  width: 100%;
  background: var(--surface-1);
  border: 1px solid var(--line-subtle);
  border-radius: 16px;
  padding: 1.25rem;
  box-shadow: 0 20px 80px rgba(0,0,0,0.12);
}
:global(.dark) .zw-overlay {
  background: linear-gradient(160deg, rgba(20,14,35,0.95) 0%, rgba(8,5,15,0.98) 100%);
  box-shadow: 0 20px 80px rgba(0,0,0,0.5);
}

/* ── Controls ── */
.zw-controls {
  display: flex; align-items: center; justify-content: space-between;
  margin-bottom: 1rem;
  flex-wrap: wrap; gap: 0.75rem;
}
.zw-toggle {
  display: flex;
  background: var(--glass-bg);
  border: 1px solid var(--line-subtle);
  border-radius: 10px;
  padding: 3px;
  gap: 2px;
}
.zw-tab {
  display: flex; align-items: center; gap: 6px;
  padding: 0.45rem 1rem;
  border: none; border-radius: 7px;
  background: transparent;
  color: var(--text-dim);
  font-size: 0.8rem; font-weight: 600;
  cursor: pointer; transition: all 0.3s;
  letter-spacing: 0.5px;
}
.zw-tab:hover { color: var(--text-muted); background: var(--glass-bg-hover); }
.zw-tab.is-active { background: var(--glass-bg-hover); color: var(--accent); }
.zw-tab-dot {
  width: 7px; height: 7px; border-radius: 50%;
  flex-shrink: 0;
}
.zw-dot-gold { background: var(--text-muted); box-shadow: 0 0 8px color-mix(in oklab, var(--text-muted) 50%, transparent); }
.zw-dot-purple { background: var(--accent); box-shadow: 0 0 8px color-mix(in oklab, var(--accent) 50%, transparent); }
.zw-year-select { display: flex; align-items: center; gap: 0.5rem; }
.zw-year-label { font-size: 0.78rem; color: var(--text-dim); letter-spacing: 1px; }
.zw-select {
  background: var(--glass-bg);
  border: 1px solid var(--line-subtle);
  border-radius: 8px;
  color: var(--accent);
  font-size: 0.82rem; font-weight: 700;
  padding: 0.35rem 0.75rem;
  cursor: pointer; outline: none;
  transition: border-color 0.3s;
}
:global(.dark) .zw-select {
  border-color: rgba(142,109,187,0.25);
  color: #8E6DBB;
}
.zw-select:hover { border-color: var(--line-focus); }
:global(.dark) .zw-select:hover { border-color: rgba(142,109,187,0.45); }

/* ── Grid ── */
.zw-grid-wrap {
  position: relative;
  overflow: visible;
}
.zw-grid {
  display: grid;
  grid-template-columns: 1fr 1fr 1fr 1fr;
  gap: 2px;
  background: var(--surface-0);
  border: 1px solid var(--line-subtle);
  border-radius: 8px;
  overflow: hidden;
}
:global(.dark) .zw-grid {
  background: rgba(0,0,0,0.4);
}
.zw-grid-overlay { background: var(--surface-2); border-color: var(--line-subtle); }
:global(.dark) .zw-grid-overlay { background: rgba(20,15,40,0.6); border-color: rgba(142,109,187,0.12); }

/* SVG overlay for connection lines */
.zw-svg-overlay {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
  z-index: 10;
  overflow: visible;
}
.zw-svg-line {
  animation: zw-line-draw 0.4s ease-out both, zw-pulse 2.5s ease-in-out 0.4s infinite;
}
.zw-svg-line-opp {
  animation: zw-line-draw-opp 0.35s ease-out both, zw-pulse-opp 2.5s ease-in-out 0.35s infinite;
}
@keyframes zw-line-draw {
  from { stroke-dashoffset: 120; opacity: 0; }
  to { stroke-dashoffset: 0; opacity: 1; }
}
@keyframes zw-line-draw-opp {
  from { stroke-dashoffset: 120; opacity: 0; }
  to { stroke-dashoffset: 0; opacity: 1; }
}
@keyframes zw-pulse {
  0%, 100% { opacity: 0.85; }
  50% { opacity: 0.55; }
}
@keyframes zw-pulse-opp {
  0%, 100% { opacity: 0.9; }
  50% { opacity: 0.6; }
}

/* ── Cell ── */
.zw-cell {
  background: var(--surface-2);
  min-height: 110px;
  padding: 0.6rem 0.4rem;
  display: flex; flex-direction: column; align-items: center; gap: 3px;
  position: relative;
  transition: background 0.3s;
}
:global(.dark) .zw-cell {
  background: linear-gradient(180deg, rgba(12,12,14,0.85) 0%, rgba(6,6,8,0.92) 100%);
}
.zw-cell:hover { background: var(--surface-3); }
:global(.dark) .zw-cell:hover { background: linear-gradient(180deg, rgba(20,20,24,0.92) 0%, rgba(12,12,16,0.96) 100%); }
.zw-cell-overlay { background: color-mix(in oklab, var(--accent) 8%, var(--surface-2)); }
:global(.dark) .zw-cell-overlay { background: linear-gradient(180deg, rgba(30,25,45,0.85) 0%, rgba(20,15,35,0.9) 100%); }
.zw-cell-overlay:hover { background: color-mix(in oklab, var(--accent) 12%, var(--surface-3)); }
:global(.dark) .zw-cell-overlay:hover { background: linear-gradient(180deg, rgba(40,35,60,0.9) 0%, rgba(25,20,45,0.95) 100%); }

/* Cell header */
.zw-cell-header {
  display: flex; flex-direction: column; align-items: center; gap: 1px;
  margin-bottom: 2px;
}
.zw-palace-name {
  font-size: 0.7rem; font-weight: 800;
  color: var(--accent); letter-spacing: 1px;
}
.zw-branch {
  font-size: 0.58rem; color: var(--text-soft);
}

/* Stars */
.zw-stars { display: flex; flex-direction: row; flex-wrap: wrap; justify-content: center; align-items: center; gap: 2px; flex: 1 1 auto; }
.zw-star {
  display: inline-block;
  border-radius: 4px;
  padding: 1px 5px;
  font-size: 0.68rem; font-weight: 800;
  color: var(--destructive-foreground);
  white-space: nowrap;
  letter-spacing: 0.5px;
  line-height: 1.4;
  transition: transform 0.2s, box-shadow 0.2s;
}
.zw-star-gold {
  box-shadow: 0 2px 8px rgba(0,0,0,0.3);
}
.zw-star:hover { transform: translateY(-1px); box-shadow: 0 4px 12px rgba(0,0,0,0.4); }
.zw-star-purple {
  box-shadow: 0 2px 8px rgba(0,0,0,0.3);
}

/* Sihua */
.zw-sihua-tag {
  font-size: 0.58rem; font-weight: 700;
  padding: 1px 5px;
  background: color-mix(in oklab, var(--crimson) 15%, transparent);
  border: 1px solid color-mix(in oklab, var(--crimson) 30%, transparent);
  border-radius: 20px;
  color: var(--crimson);
}

/* Overlay stars */
.zw-overlay-stars {
  display: flex; flex-direction: row; flex-wrap: wrap; justify-content: center; align-items: center; gap: 2px;
  margin-top: auto;
  padding-top: 4px;
  border-top: 1px dashed var(--line-subtle);
  width: 100%;
}
.zw-overlay-label {
  font-size: 0.55rem; font-weight: 700;
  color: var(--text-soft);
  letter-spacing: 2px;
  text-transform: uppercase;
  width: 100%;
  text-align: center;
  margin-bottom: 1px;
}

/* 流耀 */
.zw-liuyao {
  display: flex; flex-wrap: wrap; justify-content: center; gap: 2px;
  margin-top: 3px; padding-top: 3px;
  border-top: 1px dashed color-mix(in oklab, var(--accent) 15%, var(--line-subtle));
  width: 100%;
}
.zw-liuyao-chip {
  font-size: 0.52rem; font-weight: 600;
  padding: 0px 4px;
  background: color-mix(in oklab, var(--accent) 8%, var(--glass-bg));
  border: 1px solid color-mix(in oklab, var(--accent) 20%, var(--line-subtle));
  border-radius: 3px;
  color: var(--accent);
  white-space: nowrap;
}

/* ── Center ── */
.zw-center {
  background: var(--surface-1);
  grid-row: span 2;
  grid-column: span 2;
  display: flex; flex-direction: column; align-items: center; justify-content: center;
  gap: 6px;
  padding: 0.75rem;
  position: relative; overflow: hidden;
  border-left: 1px solid var(--line-subtle);
  border-right: 1px solid var(--line-subtle);
}
:global(.dark) .zw-center {
  background: linear-gradient(180deg, rgba(12,12,14,0.9) 0%, rgba(3,4,4,0.95) 100%);
  border-left-color: rgba(203, 213, 225,0.06);
  border-right-color: rgba(203, 213, 225,0.06);
}
.zw-center-overlay {
  background: var(--surface-2);
  border-left-color: var(--line-subtle);
  border-right-color: var(--line-subtle);
}
:global(.dark) .zw-center-overlay {
  background: linear-gradient(180deg, rgba(50,40,80,0.9) 0%, rgba(25,20,50,0.95) 100%);
  border-left-color: rgba(142,109,187,0.1);
  border-right-color: rgba(142,109,187,0.1);
}
.zw-center-glow {
  position: absolute; inset: 0;
  background: radial-gradient(circle, color-mix(in oklab, var(--text-muted) 6%, transparent), transparent 70%);
  pointer-events: none;
}
:global(.dark) .zw-center-glow {
  background: radial-gradient(circle, rgba(203, 213, 225,0.06), transparent 70%);
}
.zw-center-overlay .zw-center-glow {
  background: radial-gradient(circle, color-mix(in oklab, var(--accent) 8%, transparent), transparent 70%);
}
:global(.dark) .zw-center-overlay .zw-center-glow {
  background: radial-gradient(circle, rgba(142,109,187,0.08), transparent 70%);
}
.zw-center-title {
  font-family: var(--font-serif);
  font-size: 0.78rem; font-weight: 800;
  color: var(--accent); letter-spacing: 2px;
  text-shadow: 0 0 20px color-mix(in oklab, var(--accent) 30%, transparent);
  position: relative;
}
.zw-center-overlay .zw-center-title {
  color: #6d28d9;
  text-shadow: 0 0 20px color-mix(in oklab, var(--accent) 40%, transparent);
}
:global(.dark) .zw-center-overlay .zw-center-title {
  color: #8E6DBB;
  text-shadow: 0 0 20px rgba(142,109,187,0.4);
}
.zw-center-row { display: flex; flex-direction: column; align-items: center; gap: 0; position: relative; }
.zw-center-lbl { font-size: 0.55rem; color: var(--text-soft); letter-spacing: 1px; text-transform: uppercase; }
.zw-center-val { font-size: 0.78rem; font-weight: 800; color: var(--text); }
.zw-year-val { color: var(--accent); font-size: 0.9rem; text-shadow: 0 0 15px color-mix(in oklab, var(--accent) 40%, transparent); }
:global(.dark) .zw-year-val { color: #8E6DBB; text-shadow: 0 0 15px rgba(142,109,187,0.4); }


/* ── Legend ── */
.zw-legend {
  display: flex; justify-content: center; gap: 2rem;
  margin-top: 1rem; padding-top: 0.75rem;
  border-top: 1px solid var(--line-subtle);
}
.zw-legend-item { display: flex; align-items: center; gap: 0.5rem; }
.zw-legend-swatch {
  width: 10px; height: 10px; border-radius: 3px;
}
.zw-swatch-gold { background: linear-gradient(135deg, var(--text-muted), var(--text-dim)); box-shadow: 0 0 8px color-mix(in oklab, var(--text-muted) 40%, transparent); }
.zw-swatch-purple { background: linear-gradient(135deg, var(--accent), color-mix(in oklab, var(--accent) 70%, var(--text-dim))); box-shadow: 0 0 8px color-mix(in oklab, var(--accent) 40%, transparent); }
.zw-swatch-focused { background: color-mix(in oklab, var(--accent) 50%, transparent); border: 1px solid color-mix(in oklab, var(--accent) 70%, transparent); }
.zw-swatch-opposite { background: color-mix(in oklab, var(--accent) 40%, transparent); border: 1px solid color-mix(in oklab, var(--accent) 60%, transparent); }
.zw-swatch-surrounded { background: color-mix(in oklab, var(--accent) 25%, transparent); border: 1px solid color-mix(in oklab, var(--accent) 50%, transparent); }
.zw-legend-text { font-size: 0.72rem; color: var(--text-dim); letter-spacing: 1px; }

/* Body palace */
.zw-cell-body { border: 1px solid color-mix(in oklab, var(--crimson) 30%, var(--line-subtle)); }
.zw-body-tag { font-size: 0.48rem; font-weight: 700; background: rgba(251, 113, 133, 0.12); color: var(--danger); padding: 0px 3px; border-radius: 2px; margin-left: 2px; }

/* ── Sanfang Sizheng highlight ── */
.zw-focused {
  background: color-mix(in oklab, var(--accent) 20%, transparent) !important;
  box-shadow: inset 0 0 20px color-mix(in oklab, var(--accent) 12%, transparent);
}
.zw-opposite {
  background: color-mix(in oklab, var(--accent) 20%, transparent) !important;
  box-shadow: inset 0 0 20px color-mix(in oklab, var(--accent) 10%, transparent);
}
.zw-surrounded {
  background: color-mix(in oklab, var(--accent) 12%, transparent) !important;
  box-shadow: inset 0 0 20px color-mix(in oklab, var(--accent) 6%, transparent);
}

/* Twelve stars */
.zw-twelve { display: flex; justify-content: center; margin-top: 1px; }
.zw-twelve-tag { font-size: 0.5rem; color: var(--accent); background: var(--glass-bg); padding: 0px 3px; border-radius: 2px; }
</style>
