<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import type {
  ZiWeiDayunStageAnalysis,
  ZiWeiOverlayAnalysis,
  ZiWeiOverlayFocusPalace,
  ZiWeiOverlayTrigger,
} from '../api/ziwei'

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
    liu_nian_four_hua?: string[][]
    overlay_analysis?: ZiWeiOverlayAnalysis
  }
  overlayAnalysis?: ZiWeiOverlayAnalysis
  dayunStages?: ZiWeiDayunStageAnalysis[]
  availableYears: number[]
  birthYearBranch: string
  gender: string
}

const props = defineProps<Props>()

const emit = defineEmits<{
  (e: 'year-change', year: number): void
}>()

const mode = ref<'base' | 'overlay'>('base')
const selectedYear = ref<number>(new Date().getFullYear())
const focusedBranch = ref<string | undefined>(undefined)
const overlaySideTab = ref<'four_hua' | 'annual_stars' | 'focus_palaces'>('four_hua')

const isDark = ref(document.documentElement.classList.contains('dark'))
let themeObserver: MutationObserver | null = null

function list<T>(items: T[] | null | undefined): T[] {
  return Array.isArray(items) ? items : []
}

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

const goldMetaLight: Record<string, { bg: string; text: string }> = {
  '庙': { bg: 'linear-gradient(135deg,#b91c1c,#7f1d1d)', text: '#fffaf8' },
  '旺': { bg: 'linear-gradient(135deg,#c2410c,#9a3412)', text: '#fffaf8' },
  '得': { bg: 'linear-gradient(135deg,#ca8a04,#854d0e)', text: '#fffaf8' },
  '利': { bg: 'linear-gradient(135deg,#15803d,#166534)', text: '#fffaf8' },
  '平': { bg: 'linear-gradient(135deg,#64748b,#475569)', text: '#fffaf8' },
  '不': { bg: 'linear-gradient(135deg,#0e7490,#155e75)', text: '#fffaf8' },
  '陷': { bg: 'linear-gradient(135deg,#44403c,#292524)', text: '#e7e5e4' },
}

const goldMetaDark: Record<string, { bg: string; text: string }> = {
  '庙': { bg: 'linear-gradient(135deg,#fb7185,#be123c)', text: '#fffaf8' },
  '旺': { bg: 'linear-gradient(135deg,#fb923c,#c2410c)', text: '#fffaf8' },
  '得': { bg: 'linear-gradient(135deg,#facc15,#a16207)', text: '#17120a' },
  '利': { bg: 'linear-gradient(135deg,#34d399,#047857)', text: '#02140e' },
  '平': { bg: 'linear-gradient(135deg,#94a3b8,#64748b)', text: '#07111f' },
  '不': { bg: 'linear-gradient(135deg,#38bdf8,#0369a1)', text: '#06111a' },
  '陷': { bg: 'linear-gradient(135deg,#334155,#1e293b)', text: '#dbe4e8' },
}

const brightnessLevels = ['庙', '旺', '得', '利', '平', '不', '陷']

function baseMeta(brightness: string) {
  const meta = isDark.value ? goldMetaDark : goldMetaLight
  return meta[brightness] || meta['陷']
}

function onYearChange() {
  emit('year-change', selectedYear.value)
}

watch(() => props.liunianChart?.year, (year) => {
  if (year) selectedYear.value = year
}, { immediate: true })

const analysis = computed(() => props.overlayAnalysis || props.liunianChart?.overlay_analysis)
const dayunContext = computed(() => analysis.value?.dayun_context)

const baseLookup = computed<Record<string, PalaceData>>(() => {
  const map: Record<string, PalaceData> = {}
  list(props.baseChart?.palaces).forEach((palace) => {
    map[palace.branch] = palace
  })
  return map
})

const liunianStarsMap = computed<Record<string, string[]>>(() => {
  const map: Record<string, string[]> = {}
  const stars = list<string[]>(props.liunianChart?.liu_nian_stars)
  const palaces = list(props.liunianChart?.palaces)
  for (let i = 0; i < palaces.length; i++) {
    map[palaces[i].branch] = list(stars[i])
  }
  return map
})

const branchIndexMap: Record<string, number> = {
  '子': 0, '丑': 1, '寅': 2, '卯': 3, '辰': 4, '巳': 5,
  '午': 6, '未': 7, '申': 8, '酉': 9, '戌': 10, '亥': 11,
}
const indexBranchMap = ['子', '丑', '寅', '卯', '辰', '巳', '午', '未', '申', '酉', '戌', '亥']
const xiaoxianAgeCount = 8

const branchOrder = ['巳', '午', '未', '申', '辰', '酉', '卯', '戌', '寅', '丑', '子', '亥']
const branchGridPosition: Record<string, { row: number; col: number }> = {
  '巳': { row: 1, col: 1 },
  '午': { row: 1, col: 2 },
  '未': { row: 1, col: 3 },
  '申': { row: 1, col: 4 },
  '辰': { row: 2, col: 1 },
  '酉': { row: 2, col: 4 },
  '卯': { row: 3, col: 1 },
  '戌': { row: 3, col: 4 },
  '寅': { row: 4, col: 1 },
  '丑': { row: 4, col: 2 },
  '子': { row: 4, col: 3 },
  '亥': { row: 4, col: 4 },
}

const branchAnchorPercent: Record<string, { x: number; y: number }> = {
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

const allTriggers = computed<ZiWeiOverlayTrigger[]>(() => [
  ...list(analysis.value?.four_hua),
  ...list(analysis.value?.annual_stars),
])

const triggersByBranch = computed<Record<string, ZiWeiOverlayTrigger[]>>(() => {
  const map: Record<string, ZiWeiOverlayTrigger[]> = {}
  allTriggers.value.forEach((trigger) => {
    if (!map[trigger.branch]) map[trigger.branch] = []
    map[trigger.branch].push(trigger)
  })
  return map
})

const fourHuaTargets = computed(() => list(analysis.value?.four_hua))
const annualStarTargets = computed(() => list(analysis.value?.annual_stars))
const focusPalaces = computed(() => list(analysis.value?.focus_palaces))
const dayunLabelByBranch = computed<Record<string, string>>(() => {
  const labels: Record<string, string> = {}
  list(props.dayunStages).forEach((stage) => {
    if (!stage.branch) return
    labels[stage.branch] = `${stage.start_age} - ${stage.end_age}`
  })
  return labels
})

const overlaySideTabs = computed(() => [
  { key: 'four_hua' as const, label: '流年四化', count: fourHuaTargets.value.length },
  { key: 'annual_stars' as const, label: '流禄羊陀马', count: annualStarTargets.value.length },
  { key: 'focus_palaces' as const, label: '触发宫位', count: focusPalaces.value.length },
])

const centerTitle = computed(() => (
  mode.value === 'overlay' ? (analysis.value?.gan_zhi || `${selectedYear.value}年`) : '本命盘'
))

const centerSubtitle = computed(() => {
  if (mode.value === 'overlay') {
    const context = dayunContext.value
    if (context) {
      return `${selectedYear.value}年流年叠盘 · ${context.gan_zhi}大限 ${context.start_age}-${context.end_age}岁`
    }
    return `${selectedYear.value}年流年叠盘`
  }
  const soulPalace = props.baseChart.earthly_branch_of_soul_palace || '—'
  const bodyPalace = props.baseChart.earthly_branch_of_body_palace || '—'
  return `命宫 ${soulPalace} · 身宫 ${bodyPalace}`
})

function fixIdx(index: number): number {
  return ((index % 12) + 12) % 12
}

function sanfangBranches(branch: string): { opposite: string; trine1: string; trine2: string } {
  const idx = branchIndexMap[branch]
  if (idx === undefined) return { opposite: '', trine1: '', trine2: '' }
  return {
    opposite: indexBranchMap[fixIdx(idx + 6)],
    trine1: indexBranchMap[fixIdx(idx + 4)],
    trine2: indexBranchMap[fixIdx(idx + 8)],
  }
}

const svgLines = computed<{ from: { x: number; y: number }; to: { x: number; y: number }; type: 'opposite' | 'trine' }[]>(() => {
  if (!focusedBranch.value) return []
  const from = branchAnchorPercent[focusedBranch.value]
  const sf = sanfangBranches(focusedBranch.value)
  if (!from) return []
  const lines: { from: { x: number; y: number }; to: { x: number; y: number }; type: 'opposite' | 'trine' }[] = []
  const opposite = branchAnchorPercent[sf.opposite]
  const trine1 = branchAnchorPercent[sf.trine1]
  const trine2 = branchAnchorPercent[sf.trine2]
  if (opposite) lines.push({ from, to: opposite, type: 'opposite' })
  if (trine1) lines.push({ from, to: trine1, type: 'trine' })
  if (trine2) lines.push({ from, to: trine2, type: 'trine' })
  return lines
})

function basePalaceAt(branch: string): PalaceData | undefined {
  return baseLookup.value[branch]
}

function dayunLabelAt(branch: string): string {
  return dayunLabelByBranch.value[branch] || ''
}

function xiaoxianStartBranchIndex(yearBranch: string): number | undefined {
  if (['寅', '午', '戌'].includes(yearBranch)) return branchIndexMap['辰']
  if (['申', '子', '辰'].includes(yearBranch)) return branchIndexMap['戌']
  if (['巳', '酉', '丑'].includes(yearBranch)) return branchIndexMap['未']
  if (['亥', '卯', '未'].includes(yearBranch)) return branchIndexMap['丑']
  return undefined
}

function xiaoxianDirection(): 1 | -1 | 0 {
  const gender = props.gender.trim().toUpperCase()
  if (gender === '男' || gender === 'MALE' || gender === 'M') return 1
  if (gender === '女' || gender === 'FEMALE' || gender === 'F') return -1
  return 0
}

// 小限一岁由生年支三合墓库的对冲位起，男顺女逆，每十二年回到同一宫。
function palaceAgesAt(branch: string): number[] {
  const startBranchIndex = xiaoxianStartBranchIndex(props.birthYearBranch)
  const targetBranchIndex = branchIndexMap[branch]
  const direction = xiaoxianDirection()
  if (startBranchIndex === undefined || targetBranchIndex === undefined || direction === 0) return []

  const offset = direction === 1
    ? fixIdx(targetBranchIndex - startBranchIndex)
    : fixIdx(startBranchIndex - targetBranchIndex)
  const firstAge = offset + 1
  return Array.from({ length: xiaoxianAgeCount }, (_, index) => firstAge + index * 12)
}

function palaceAgesTitleAt(branch: string): string {
  const palaceName = basePalaceAt(branch)?.name || `${branch}宫`
  return `${palaceName}小限经过年龄：${palaceAgesAt(branch).join('、')}岁`
}

function richStars(palace: PalaceData | undefined): StarInfo[] {
  return list(palace?.stars)
}

function displayStars(palace: PalaceData | undefined): StarInfo[] {
  const stars = richStars(palace)
  return mode.value === 'overlay' ? stars.slice(0, 5) : stars
}

function hiddenStarCount(palace: PalaceData | undefined): number {
  const count = richStars(palace).length - displayStars(palace).length
  return count > 0 ? count : 0
}

function palaceSihua(palace: PalaceData | undefined): string[] {
  return list(palace?.four_hua)
}

function liunianStarsAt(branch: string): string[] {
  return list(liunianStarsMap.value[branch])
}

function triggerChipsAt(branch: string): ZiWeiOverlayTrigger[] {
  return list(triggersByBranch.value[branch])
}

function branchStyle(branch: string) {
  const pos = branchGridPosition[branch]
  return {
    gridColumn: String(pos.col),
    gridRow: String(pos.row),
  }
}

function palaceHighlightClass(branch: string): string {
  if (!focusedBranch.value) return ''
  if (branch === focusedBranch.value) return 'zw-focused'
  const sf = sanfangBranches(focusedBranch.value)
  if (branch === sf.opposite) return 'zw-opposite'
  if (branch === sf.trine1 || branch === sf.trine2) return 'zw-surrounded'
  return ''
}

function overlayImpactClass(branch: string): string {
  if (mode.value !== 'overlay') return ''
  const triggers = triggerChipsAt(branch)
  if (triggers.some((trigger) => trigger.polarity === 'constraint' || trigger.type === '化忌')) return 'zw-impact-watch'
  if (triggers.some((trigger) => trigger.polarity === 'resource')) return 'zw-impact-good'
  if (triggers.some((trigger) => trigger.polarity === 'movement')) return 'zw-impact-move'
  if (triggers.length) return 'zw-impact-neutral'
  return ''
}

function triggerClass(trigger: ZiWeiOverlayTrigger): string {
  if (trigger.polarity === 'constraint' || trigger.type === '化忌') return 'is-watch'
  if (trigger.polarity === 'resource') return 'is-good'
  if (trigger.polarity === 'movement') return 'is-move'
  return 'is-neutral'
}

function triggerLabel(trigger: ZiWeiOverlayTrigger): string {
  return trigger.star ? `${trigger.star}${trigger.type}` : trigger.type
}

function fallbackChipClass(label: string): string {
  if (label.includes('化忌') || label.includes('流羊') || label.includes('流陀')) return 'is-watch'
  if (label.includes('化禄') || label.includes('化科') || label.includes('流禄')) return 'is-good'
  if (label.includes('流马')) return 'is-move'
  return 'is-neutral'
}

function focusPalaceClass(item: ZiWeiOverlayFocusPalace): string {
  const triggers = list(item.triggers)
  if (triggers.some((trigger) => trigger.polarity === 'constraint' || trigger.type === '化忌')) return 'is-watch'
  if (triggers.some((trigger) => trigger.polarity === 'resource')) return 'is-good'
  if (triggers.some((trigger) => trigger.polarity === 'movement')) return 'is-move'
  return 'is-neutral'
}

function overlaySummary(): string {
  return analysis.value?.summary || `${selectedYear.value}年叠盘会标出流年四化、流禄、流羊、流陀和流马落宫，记录被时间层触发的宫位结构。`
}
</script>

<template>
  <section class="zw-overlay">
    <div class="zw-controls">
      <div class="zw-toggle" aria-label="紫微盘显示模式">
        <button class="zw-tab" :class="{ 'is-active': mode === 'base' }" type="button" @click="mode = 'base'">
          <span class="zw-tab-dot zw-dot-gold"></span>
          本命盘
        </button>
        <button class="zw-tab" :class="{ 'is-active': mode === 'overlay' }" type="button" @click="mode = 'overlay'">
          <span class="zw-tab-dot zw-dot-year"></span>
          流年叠盘
        </button>
      </div>

      <div v-if="mode === 'overlay'" class="zw-year-select">
        <span class="zw-year-label">流年</span>
        <select v-model="selectedYear" class="zw-select" @change="onYearChange">
          <option v-for="year in availableYears" :key="year" :value="year">{{ year }}年</option>
        </select>
      </div>
    </div>

    <section v-if="mode === 'overlay'" class="zw-overlay-guide">
      <div class="zw-guide-main">
        <span class="zw-kicker">年度叠盘依据</span>
        <div class="zw-guide-title-row">
          <h3>{{ selectedYear }}年 <span>{{ analysis?.gan_zhi || '流年' }}</span></h3>
          <div class="zw-contract-status">
            <strong>结构</strong>
            <span>参考</span>
          </div>
        </div>
        <div class="zw-guide-reading">
          <p class="zw-guide-summary">{{ overlaySummary() }}</p>
          <div class="zw-guide-meta">
            <span v-if="analysis?.shi_shen">十神：{{ analysis.shi_shen }}</span>
            <span v-if="analysis?.relation_to_ming">命局关系：{{ analysis.relation_to_ming }}</span>
            <span v-if="dayunContext">
              大限：{{ dayunContext.gan_zhi }} · {{ dayunContext.palace }} · 虚岁
              {{ dayunContext.nominal_age }}
            </span>
            <span v-if="analysis?.review_note">{{ analysis.review_note }}</span>
          </div>
        </div>
      </div>
      <div class="zw-method-grid">
        <article v-for="step in list(analysis?.method)" :key="step.label" class="zw-method-card">
          <span>{{ step.label }}</span>
          <strong>{{ step.value }}</strong>
          <p>{{ step.meaning }}</p>
        </article>
      </div>
    </section>

    <div class="zw-workspace" :class="{ 'is-overlay': mode === 'overlay' }">
      <div class="zw-grid-wrap">
        <div class="zw-grid" :class="{ 'zw-grid-overlay': mode === 'overlay' }">
          <button
            v-for="branch in branchOrder"
            :key="branch"
            class="zw-cell"
            :class="[
              { 'zw-cell-overlay': mode === 'overlay', 'zw-cell-body': basePalaceAt(branch)?.is_body_palace },
              palaceHighlightClass(branch),
              overlayImpactClass(branch),
            ]"
            :style="branchStyle(branch)"
            :data-branch="branch"
            type="button"
            @mouseenter="focusedBranch = branch"
            @mouseleave="focusedBranch = undefined"
            @focus="focusedBranch = branch"
            @blur="focusedBranch = undefined"
          >
            <div class="zw-cell-header">
              <span class="zw-palace-name">{{ basePalaceAt(branch)?.name || branch }}</span>
              <span class="zw-branch">{{ branch }}<template v-if="basePalaceAt(branch)?.heavenly_stem"> · {{ basePalaceAt(branch)?.heavenly_stem }}</template></span>
              <span v-if="basePalaceAt(branch)?.is_body_palace" class="zw-body-tag">身宫</span>
            </div>

            <div class="zw-stars">
              <span
                v-for="(star, index) in displayStars(basePalaceAt(branch))"
                :key="star.name + index"
                class="zw-star"
                :style="{ background: baseMeta(star.brightness).bg, color: baseMeta(star.brightness).text }"
              >
                {{ star.name }}
              </span>
              <span v-if="hiddenStarCount(basePalaceAt(branch))" class="zw-more-star">+{{ hiddenStarCount(basePalaceAt(branch)) }}</span>
              <span v-for="hua in palaceSihua(basePalaceAt(branch))" :key="hua" class="zw-sihua-tag">
                {{ mode === 'overlay' ? `本命·${hua}` : hua }}
              </span>
            </div>

            <div v-if="basePalaceAt(branch)?.changsheng_12" class="zw-twelve">
              <span>{{ basePalaceAt(branch)?.changsheng_12 }}</span>
            </div>

            <div v-if="mode === 'overlay'" class="zw-trigger-strip">
              <template v-if="triggerChipsAt(branch).length">
                <span
                  v-for="trigger in triggerChipsAt(branch)"
                  :key="trigger.type + trigger.star + trigger.palace + trigger.branch"
                  class="zw-trigger-chip"
                  :class="triggerClass(trigger)"
                  :title="trigger.meaning"
                >
                  流年·{{ triggerLabel(trigger) }}
                </span>
              </template>
              <template v-else-if="liunianStarsAt(branch).length">
                <span
                  v-for="star in liunianStarsAt(branch)"
                  :key="star"
                  class="zw-trigger-chip"
                  :class="fallbackChipClass(star)"
                >
                  流年·{{ star }}
                </span>
              </template>
              <span v-else class="zw-no-trigger">本年未重点触发</span>
            </div>

            <span v-if="mode === 'base' && dayunLabelAt(branch)" class="zw-dayun-range">
              {{ dayunLabelAt(branch) }}
            </span>
            <span
              v-if="mode === 'base' && palaceAgesAt(branch).length"
              class="zw-palace-ages"
              :title="palaceAgesTitleAt(branch)"
            >
              <span class="zw-palace-age-list">
                <b v-for="age in palaceAgesAt(branch)" :key="age">{{ age }}</b>
              </span>
            </span>
          </button>

          <div class="zw-center" :class="{ 'zw-center-overlay': mode === 'overlay' }">
            <div class="zw-center-head">
              <span class="zw-center-kicker">{{ mode === 'overlay' ? '年度结构' : '命宫核心' }}</span>
              <strong class="zw-center-title">{{ centerTitle }}</strong>
              <span class="zw-center-subtitle">{{ centerSubtitle }}</span>
            </div>

            <div class="zw-center-duo">
              <span class="zw-center-major">
                <small>命主</small>
                <b>{{ baseChart.life_master || '—' }}</b>
              </span>
              <span class="zw-center-major">
                <small>身主</small>
                <b>{{ baseChart.body_master || '—' }}</b>
              </span>
            </div>

            <div class="zw-center-grid">
              <span><small>五行局</small><b>{{ baseChart.five_bureau || '—' }}</b></span>
              <span v-if="mode === 'overlay'"><small>流年</small><b>{{ selectedYear }}</b></span>
            </div>
          </div>
        </div>

        <svg
          v-if="svgLines.length"
          class="zw-svg-overlay"
          viewBox="0 0 100 100"
          preserveAspectRatio="none"
        >
          <line
            v-for="(line, index) in svgLines"
            :key="'line-' + index"
            :x1="line.from.x"
            :y1="line.from.y"
            :x2="line.to.x"
            :y2="line.to.y"
            :class="line.type === 'opposite' ? 'is-opposite' : 'is-trine'"
            stroke-linecap="round"
          />
        </svg>

      </div>

      <aside v-if="mode === 'overlay'" class="zw-overlay-side">
        <section class="zw-evidence-block">
          <div class="zw-side-heading">
            <span>流年触发</span>
            <small>{{ selectedYear }}年</small>
          </div>

          <div class="zw-side-tabs" aria-label="流年触发类型">
            <button
              v-for="tab in overlaySideTabs"
              :key="tab.key"
              type="button"
              class="zw-side-tab"
              :class="{ 'is-active': overlaySideTab === tab.key }"
              @click="overlaySideTab = tab.key"
            >
              <span>{{ tab.label }}</span>
              <small>{{ tab.count }}</small>
            </button>
          </div>

          <div v-if="overlaySideTab === 'four_hua'" class="zw-trigger-list">
            <article
              v-for="trigger in fourHuaTargets"
              :key="trigger.type + trigger.star + trigger.branch"
              class="zw-trigger-card"
              :class="triggerClass(trigger)"
            >
              <div>
                <strong>{{ triggerLabel(trigger) }}</strong>
                <span>{{ trigger.branch }} · {{ trigger.palace }}</span>
              </div>
              <p>{{ trigger.meaning }}</p>
            </article>
            <p v-if="!fourHuaTargets.length" class="zw-empty-line">本年未形成四化飞星触发。</p>
          </div>

          <div v-else-if="overlaySideTab === 'annual_stars'" class="zw-trigger-list compact">
            <article
              v-for="trigger in annualStarTargets"
              :key="trigger.type + trigger.branch"
              class="zw-trigger-card"
              :class="triggerClass(trigger)"
            >
              <div>
                <strong>{{ trigger.type }}</strong>
                <span>{{ trigger.branch }} · {{ trigger.palace }}</span>
              </div>
              <p>{{ trigger.meaning }}</p>
            </article>
            <p v-if="!annualStarTargets.length" class="zw-empty-line">本年未见流禄羊陀马落宫。</p>
          </div>

          <div v-else class="zw-focus-list">
            <article
              v-for="item in focusPalaces"
              :key="item.palace + item.branch"
              class="zw-focus-card"
              :class="focusPalaceClass(item)"
              @mouseenter="focusedBranch = item.branch"
              @mouseleave="focusedBranch = undefined"
            >
              <header>
                <strong>{{ item.palace }}</strong>
                <span>{{ item.branch }}宫 · {{ list(item.triggers).length }} 项触发</span>
              </header>
              <div v-if="list(item.main_stars).length" class="zw-mini-tags">
                <span v-for="star in list(item.main_stars)" :key="star">{{ star }}</span>
              </div>
              <p>{{ item.review_note }}</p>
            </article>
            <p v-if="!focusPalaces.length" class="zw-empty-line">本年未记录流年四化或流禄羊陀马触发宫位。</p>
          </div>
        </section>
      </aside>
    </div>

    <div class="zw-legend">
      <span><i class="swatch base"></i>本命星曜</span>
      <span v-if="mode === 'overlay'"><i class="swatch good"></i>资源触发</span>
      <span v-if="mode === 'overlay'"><i class="swatch watch"></i>约束触发</span>
      <span v-if="mode === 'overlay'"><i class="swatch move"></i>移动变化</span>
      <span><i class="swatch focus"></i>悬停三方四正</span>
      <div class="zw-brightness-legend" aria-label="星曜亮度：庙、旺、得、利、平、不、陷">
        <div class="zw-brightness-row">
          <span
            v-for="level in brightnessLevels"
            :key="'dot-' + level"
            class="zw-brightness-dot"
            :style="{ background: baseMeta(level).bg }"
          ></span>
        </div>
        <div class="zw-brightness-row zw-brightness-text">
          <span v-for="level in brightnessLevels" :key="'label-' + level">{{ level }}</span>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.zw-overlay {
  width: 100%;
  padding: 1.2rem;
  border: 1px solid var(--line-subtle);
  border-radius: 12px;
  background:
    linear-gradient(145deg, color-mix(in oklab, var(--surface-1) 92%, transparent), var(--surface-0)),
    var(--surface-1);
  box-shadow: 0 20px 70px rgba(15, 23, 42, 0.12);
}

:global(.dark) .zw-overlay {
  background:
    linear-gradient(145deg, rgba(8, 13, 20, 0.98), rgba(15, 18, 27, 0.96)),
    var(--surface-1);
  box-shadow: 0 24px 80px rgba(0, 0, 0, 0.42);
}

.zw-controls {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  margin-bottom: 1rem;
  flex-wrap: wrap;
}

.zw-toggle {
  display: flex;
  gap: 3px;
  padding: 3px;
  border: 1px solid var(--line-subtle);
  border-radius: 9px;
  background: var(--glass-bg);
}

.zw-tab {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  min-height: 34px;
  padding: 0.35rem 0.85rem;
  border: 0;
  border-radius: 6px;
  background: transparent;
  color: var(--text-dim);
  font-size: var(--fs-sm);
  font-weight: 700;
  cursor: pointer;
}

.zw-tab:hover,
.zw-tab.is-active {
  background: var(--glass-bg-hover);
  color: var(--text);
}

.zw-tab-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.zw-dot-gold { background: #d6a44b; }
.zw-dot-year { background: #22c55e; }

.zw-year-select {
  display: flex;
  align-items: center;
  gap: 0.55rem;
}

.zw-year-label {
  color: var(--text-dim);
  font-size: var(--fs-xs);
  font-weight: 700;
}

.zw-select {
  min-height: 34px;
  padding: 0.35rem 0.7rem;
  border: 1px solid var(--line-subtle);
  border-radius: 7px;
  background: var(--glass-bg);
  color: var(--text);
  font-weight: 800;
  outline: none;
}

.zw-select:focus {
  border-color: color-mix(in oklab, #22c55e 55%, var(--line-subtle));
  box-shadow: 0 0 0 3px rgba(34, 197, 94, 0.12);
}

.zw-overlay-guide {
  display: grid;
  grid-template-columns: minmax(0, 1.05fr) minmax(300px, 0.95fr);
  gap: 0.9rem;
  margin-bottom: 1rem;
  padding: 1rem;
  border: 1px solid var(--line-subtle);
  border-radius: 10px;
  background: var(--surface-1);
}

.zw-guide-main {
  display: flex;
  min-width: 0;
  min-height: 100%;
  flex-direction: column;
}

.zw-guide-reading {
  position: relative;
  margin-top: clamp(3.5rem, 7vw, 5.25rem);
  padding: 0.8rem 0.85rem 0.9rem;
  border: 1px solid color-mix(in oklab, var(--line-subtle) 82%, transparent);
  border-radius: 8px;
  background:
    linear-gradient(135deg, rgba(var(--jade-accent-rgb), 0.045), transparent 62%),
    color-mix(in oklab, var(--surface-1) 72%, transparent);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.28);
}

.zw-kicker,
.zw-center-kicker {
  color: var(--text-dim);
  font-size: var(--fs-xs);
  font-weight: 800;
  letter-spacing: 0.12em;
}

.zw-guide-main > .zw-kicker {
  font-size: var(--fs-sm);
  letter-spacing: 0.14em;
}

.zw-guide-title-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  margin-top: 0.35rem;
}

.zw-guide-title-row h3 {
  margin: 0;
  color: var(--text);
  font-family: var(--font-serif), serif;
  font-size: var(--fs-stat-lg);
  line-height: 1.15;
  letter-spacing: 0;
}

.zw-guide-title-row h3 span {
  color: #16a34a;
}

:global(.dark) .zw-guide-title-row h3 span {
  color: #86efac;
}

.zw-contract-status {
  display: grid;
  place-items: center;
  min-width: 104px;
  min-height: 86px;
  border: 1px solid var(--line-subtle);
  border-radius: 8px;
  background: var(--surface-1);
}

.zw-contract-status strong {
  color: var(--text);
  font-size: var(--fs-stat);
  line-height: 1;
}

.zw-contract-status span {
  color: var(--text-dim);
  font-size: var(--fs-xs);
  font-weight: 700;
}

.zw-guide-summary {
  margin: 0;
  color: var(--text-muted);
  line-height: 1.75;
}

.zw-guide-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin-top: 0.75rem;
}

.zw-guide-meta span {
  min-width: 0;
  max-width: 100%;
  padding: 0.28rem 0.55rem;
  border: 1px solid var(--line-subtle);
  border-radius: 6px;
  background: color-mix(in oklab, var(--surface-1) 78%, transparent);
  color: var(--text-soft);
  font-size: var(--fs-xs);
  line-height: 1.45;
  overflow-wrap: anywhere;
}

.zw-method-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.65rem;
}

.zw-method-card {
  padding: 0.7rem;
  border: 1px solid var(--line-subtle);
  border-radius: 8px;
  background: color-mix(in oklab, var(--surface-1) 78%, transparent);
}

.zw-method-card span {
  color: var(--text-dim);
  font-size: var(--fs-xs);
  font-weight: 800;
}

.zw-method-card strong {
  display: block;
  margin: 0.24rem 0;
  color: var(--text);
  font-size: var(--fs-sm);
  line-height: 1.45;
}

.zw-method-card p {
  margin: 0;
  color: var(--text-soft);
  font-size: var(--fs-xs);
  line-height: 1.6;
}

.zw-workspace {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 1rem;
}

.zw-workspace.is-overlay {
  grid-template-columns: minmax(560px, 1fr) minmax(300px, 0.38fr);
  align-items: start;
}

.zw-grid-wrap {
  position: relative;
  min-width: 0;
}

.zw-grid {
  display: grid;
  grid-template-columns: 1fr 1fr 1fr 1fr;
  grid-template-rows: repeat(4, minmax(110px, auto));
  gap: 2px;
  overflow: hidden;
  border: 1px solid var(--line-subtle);
  border-radius: 8px;
  background: var(--surface-0);
}

.zw-cell,
.zw-center {
  min-width: 0;
  border: 0;
  background: var(--surface-2);
}

.zw-cell {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 3px;
  min-height: 110px;
  padding: 0.82rem 0.4rem;
  color: var(--text);
  text-align: center;
  cursor: pointer;
  transition: background 0.18s ease, box-shadow 0.18s ease, transform 0.18s ease;
}

.zw-cell:hover,
.zw-cell:focus-visible {
  z-index: 2;
  background: var(--surface-3);
  outline: none;
  box-shadow: inset 0 0 0 1px color-mix(in oklab, var(--accent) 18%, transparent);
}

.zw-cell-overlay {
  background: color-mix(in oklab, var(--accent) 8%, var(--surface-2));
}

:global(.dark) .zw-cell {
  background: linear-gradient(180deg, rgba(16, 20, 28, 0.92), rgba(9, 12, 18, 0.96));
}

:global(.dark) .zw-cell-overlay {
  background: linear-gradient(180deg, rgba(30,25,45,0.85) 0%, rgba(20,15,35,0.9) 100%);
}

.zw-cell-header {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1px;
  margin-bottom: 2px;
}

.zw-palace-name {
  color: var(--accent);
  font-size: var(--fs-xs);
  font-weight: 900;
  letter-spacing: 1px;
}

.zw-branch {
  color: var(--text-dim);
  font-size: var(--fs-2xs);
  font-weight: 700;
}

.zw-body-tag {
  width: fit-content;
  padding: 0.05rem 0.28rem;
  border-radius: 4px;
  background: rgba(244, 63, 94, 0.12);
  color: #e11d48;
  font-size: var(--fs-2xs);
  font-weight: 800;
}

.zw-stars {
  display: flex;
  flex-wrap: wrap;
  gap: 0.22rem;
  align-items: center;
  justify-content: center;
  min-height: 28px;
  flex: 1 1 auto;
}

.zw-star,
.zw-more-star,
.zw-sihua-tag {
  display: inline-flex;
  align-items: center;
  max-width: 100%;
  min-height: 20px;
  padding: 0.08rem 0.38rem;
  border-radius: 4px;
  font-size: var(--fs-2xs);
  font-weight: 850;
  line-height: 1.25;
  white-space: nowrap;
}

.zw-star {
  color: var(--destructive-foreground);
  box-shadow: 0 2px 8px rgba(0,0,0,0.3);
  transition: transform 0.2s, box-shadow 0.2s;
}

.zw-star:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(0,0,0,0.4);
}

.zw-more-star {
  border: 1px solid var(--line-subtle);
  background: var(--glass-bg);
  color: var(--text-dim);
}

.zw-sihua-tag {
  border: 1px solid color-mix(in oklab, var(--crimson) 30%, transparent);
  background: color-mix(in oklab, var(--crimson) 15%, transparent);
  color: var(--crimson);
  border-radius: 20px;
}

:global(.dark) .zw-sihua-tag {
  color: #fbbf24;
}

.zw-twelve {
  margin-top: auto;
  display: flex;
  justify-content: center;
}

.zw-twelve span {
  color: var(--accent);
  font-size: var(--fs-2xs);
  background: var(--glass-bg);
  padding: 0px 3px;
  border-radius: 2px;
}

.zw-dayun-range {
  position: absolute;
  right: 0.38rem;
  bottom: 0.32rem;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 1.15rem;
  color: var(--accent);
  font-size: var(--fs-2xs);
  font-weight: 900;
  line-height: 1;
  pointer-events: none;
  white-space: nowrap;
}

.zw-palace-ages {
  position: absolute;
  z-index: 1;
  top: 0.34rem;
  left: 0.3rem;
  display: grid;
  max-width: 2.9rem;
  justify-items: start;
  color: var(--text-dim);
  font-size: 0.58rem;
  line-height: 1.05;
  text-align: left;
  pointer-events: none;
}

.zw-palace-age-list {
  display: grid;
  grid-template-rows: repeat(4, auto);
  grid-auto-flow: column;
  grid-auto-columns: max-content;
  align-items: flex-start;
  justify-content: start;
  column-gap: 0.26rem;
  row-gap: 0.08rem;
}

.zw-palace-age-list b {
  color: var(--accent);
  font-size: inherit;
  font-weight: 850;
  line-height: 1;
}

.zw-trigger-strip {
  display: flex;
  flex-wrap: wrap;
  gap: 0.24rem;
  min-height: 24px;
  padding-top: 0.38rem;
  border-top: 1px dashed var(--line-subtle);
}

.zw-trigger-chip {
  display: inline-flex;
  max-width: 100%;
  min-height: 20px;
  align-items: center;
  padding: 0.05rem 0.35rem;
  border: 1px solid var(--line-subtle);
  border-radius: 5px;
  font-size: var(--fs-2xs);
  font-weight: 850;
  line-height: 1.25;
  white-space: nowrap;
}

.is-good {
  border-color: rgba(34, 197, 94, 0.35) !important;
  background: rgba(34, 197, 94, 0.11) !important;
  color: #15803d !important;
}

.is-watch {
  border-color: rgba(244, 63, 94, 0.34) !important;
  background: rgba(244, 63, 94, 0.11) !important;
  color: #be123c !important;
}

.is-move {
  border-color: rgba(14, 165, 233, 0.34) !important;
  background: rgba(14, 165, 233, 0.11) !important;
  color: #0369a1 !important;
}

.is-neutral {
  border-color: rgba(148, 163, 184, 0.32) !important;
  background: rgba(148, 163, 184, 0.11) !important;
  color: var(--text-muted) !important;
}

:global(.dark) .is-good { color: #86efac !important; }
:global(.dark) .is-watch { color: #fda4af !important; }
:global(.dark) .is-move { color: #7dd3fc !important; }

.zw-no-trigger {
  color: var(--text-dim);
  font-size: var(--fs-2xs);
}

.zw-impact-good {
  box-shadow: inset 0 0 0 1px rgba(34, 197, 94, 0.22);
}

.zw-impact-watch {
  box-shadow: inset 0 0 0 1px rgba(244, 63, 94, 0.24);
}

.zw-impact-move {
  box-shadow: inset 0 0 0 1px rgba(14, 165, 233, 0.22);
}

.zw-impact-neutral {
  box-shadow: inset 0 0 0 1px rgba(148, 163, 184, 0.2);
}

.zw-cell-body {
  box-shadow: inset 0 0 0 1px rgba(245, 158, 11, 0.32);
}

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

.zw-center {
  grid-column: 2 / 4;
  grid-row: 2 / 4;
  position: relative;
  display: grid;
  align-content: center;
  justify-content: center;
  gap: 0.68rem;
  overflow: hidden;
  padding: 1rem 1.05rem;
  background:
    radial-gradient(circle at 50% 42%, color-mix(in oklab, var(--accent) 10%, transparent), transparent 58%),
    linear-gradient(180deg, color-mix(in oklab, var(--accent) 5%, transparent), transparent 48%),
    color-mix(in oklab, var(--surface-2) 96%, var(--surface-0));
  text-align: center;
}

.zw-center::before,
.zw-center::after {
  content: '';
  position: absolute;
  pointer-events: none;
}

.zw-center::before {
  inset: 0.62rem;
  border: 1px solid color-mix(in oklab, var(--accent) 14%, transparent);
  border-radius: 7px;
}

.zw-center::after {
  top: 0.8rem;
  right: 0.72rem;
  bottom: 0.8rem;
  width: 2px;
  border-radius: 999px;
  background: linear-gradient(180deg, var(--accent), color-mix(in oklab, var(--crimson) 68%, var(--accent)));
  opacity: 0.48;
}

.zw-center-overlay {
  background:
    radial-gradient(circle at 50% 42%, rgba(34, 197, 94, 0.12), transparent 58%),
    linear-gradient(145deg, rgba(34, 197, 94, 0.08), rgba(14, 165, 233, 0.06)),
    color-mix(in oklab, var(--surface-2) 96%, var(--surface-0));
}

.zw-center-head,
.zw-center-duo,
.zw-center-grid {
  position: relative;
  z-index: 1;
}

.zw-center-head {
  display: grid;
  justify-items: center;
  gap: 0.22rem;
}

.zw-center-title {
  color: var(--text);
  font-family: var(--font-serif), serif;
  font-size: var(--fs-2xl);
  line-height: 1.15;
  letter-spacing: 0;
}

.zw-center-subtitle {
  max-width: 100%;
  color: var(--text-soft);
  font-size: var(--fs-xs);
  font-weight: 750;
  line-height: 1.35;
}

.zw-center-duo {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.46rem;
}

.zw-center-major {
  display: grid;
  justify-items: center;
  gap: 0.16rem;
  min-width: 0;
  min-height: 56px;
  padding: 0.48rem 0.58rem;
  border: 1px solid color-mix(in oklab, var(--accent) 20%, var(--line-subtle));
  border-radius: 8px;
  background:
    linear-gradient(180deg, color-mix(in oklab, var(--surface-1) 72%, transparent), color-mix(in oklab, var(--surface-2) 86%, transparent));
  box-shadow: inset 0 1px 0 color-mix(in oklab, var(--text) 5%, transparent);
}

.zw-center-major small,
.zw-center-grid small {
  color: var(--text-dim);
  font-size: var(--fs-2xs);
  font-weight: 850;
  line-height: 1.2;
}

.zw-center-major b {
  max-width: 100%;
  color: var(--accent);
  font-family: var(--font-serif), serif;
  font-size: var(--fs-xl);
  line-height: 1.1;
  letter-spacing: 0;
  overflow-wrap: anywhere;
}

.zw-center-grid {
  display: inline-flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 0.35rem;
  max-width: 100%;
}

.zw-center-grid span {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.32rem;
  min-width: 0;
  padding: 0.22rem 0.48rem;
  border: 1px solid color-mix(in oklab, var(--line-subtle) 72%, transparent);
  border-radius: 999px;
  background: color-mix(in oklab, var(--surface-1) 48%, transparent);
  color: var(--text-dim);
}

.zw-center-grid b {
  max-width: 100%;
  color: var(--text);
  font-size: var(--fs-xs);
  line-height: 1.2;
  overflow-wrap: anywhere;
}

.zw-svg-overlay {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
  z-index: 5;
}

.zw-svg-overlay line {
  stroke-width: 0.55;
  stroke-dasharray: 2 1.4;
  opacity: 0.82;
}

.zw-svg-overlay .is-opposite {
  stroke: rgba(34, 197, 94, 0.78);
}

.zw-svg-overlay .is-trine {
  stroke: rgba(245, 158, 11, 0.68);
}

.zw-brightness-legend {
  position: absolute;
  right: 0;
  bottom: 0.42rem;
  display: grid;
  justify-items: end;
  gap: 0.18rem;
  padding: 0 0.12rem;
  pointer-events: none;
}

.zw-brightness-row {
  display: grid;
  grid-template-columns: repeat(7, 0.72rem);
  justify-items: center;
  align-items: center;
  gap: 0.08rem;
}

.zw-brightness-dot {
  width: 0.44rem;
  height: 0.44rem;
  border-radius: 999px;
  box-shadow:
    inset 0 1px 1px rgba(255, 255, 255, 0.28),
    0 0 0 1px rgba(255, 255, 255, 0.18);
}

.zw-brightness-text {
  color: var(--text-soft);
  font-size: 0.58rem;
  font-weight: 850;
  line-height: 1;
}

.zw-brightness-text span {
  display: block;
}

.zw-overlay-side {
  display: grid;
  gap: 0.85rem;
}

.zw-evidence-block {
  border: 1px solid var(--line-subtle);
  border-radius: 10px;
  background: color-mix(in oklab, var(--surface-1) 86%, transparent);
  overflow: hidden;
}

.zw-side-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  padding: 0.75rem 0.85rem;
  border-bottom: 1px solid var(--line-subtle);
}

.zw-side-heading span {
  color: var(--text);
  font-size: var(--fs-sm);
  font-weight: 900;
}

.zw-side-heading small {
  color: var(--text-dim);
  font-size: var(--fs-xs);
  font-weight: 800;
}

.zw-side-tabs {
  display: flex;
  flex-wrap: wrap;
  gap: 0.45rem;
  padding: 0.72rem 0.75rem 0;
}

.zw-side-tab {
  display: inline-flex;
  align-items: center;
  gap: 0.34rem;
  min-height: 30px;
  padding: 0.26rem 0.56rem;
  border: 1px solid var(--line-subtle);
  border-radius: 999px;
  background: var(--glass-bg);
  color: var(--text-dim);
  cursor: pointer;
  font-size: var(--fs-xs);
  font-weight: 850;
  line-height: 1.2;
  transition: background 0.18s ease, border-color 0.18s ease, color 0.18s ease, transform 0.18s ease;
}

.zw-side-tab:hover,
.zw-side-tab.is-active {
  border-color: color-mix(in oklab, var(--accent) 32%, var(--line-subtle));
  background: color-mix(in oklab, var(--accent) 12%, var(--glass-bg));
  color: var(--text);
}

.zw-side-tab:hover {
  transform: translateY(-1px);
}

.zw-side-tab small {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 18px;
  height: 18px;
  padding: 0 0.28rem;
  border-radius: 999px;
  background: color-mix(in oklab, var(--surface-0) 74%, transparent);
  color: var(--text-soft);
  font-size: var(--fs-2xs);
  font-weight: 900;
}

.zw-side-tab.is-active small {
  background: var(--accent);
  color: var(--bg);
}

.zw-trigger-list,
.zw-focus-list {
  display: grid;
  gap: 0.55rem;
  padding: 0.75rem;
}

.zw-trigger-list.compact {
  gap: 0.45rem;
}

.zw-trigger-card,
.zw-focus-card {
  display: grid;
  gap: 0.45rem;
  padding: 0.68rem;
  border: 1px solid var(--line-subtle);
  border-radius: 8px;
  background: var(--surface-2);
}

.zw-trigger-card div,
.zw-focus-card header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.65rem;
}

.zw-trigger-card strong,
.zw-focus-card strong {
  color: inherit;
  font-size: var(--fs-sm);
}

.zw-trigger-card span,
.zw-focus-card header span {
  color: var(--text-dim);
  font-size: var(--fs-xs);
  white-space: nowrap;
}

.zw-trigger-card p,
.zw-focus-card p {
  margin: 0;
  color: var(--text-soft);
  font-size: var(--fs-xs);
  line-height: 1.65;
}

.zw-empty-line {
  margin: 0;
  padding: 0.7rem;
  border: 1px dashed var(--line-subtle);
  border-radius: 8px;
  color: var(--text-dim);
  background: var(--glass-bg);
  font-size: var(--fs-xs);
  line-height: 1.65;
}

.zw-mini-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.25rem;
}

.zw-mini-tags span {
  padding: 0.08rem 0.36rem;
  border-radius: 4px;
  background: var(--glass-bg);
  color: var(--text-muted);
  font-size: var(--fs-2xs);
  font-weight: 800;
}

.zw-legend {
  position: relative;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.8rem 1.3rem;
  margin-top: 1rem;
  padding-top: 0.8rem;
  border-top: 1px solid var(--line-subtle);
  color: var(--text-dim);
  font-size: var(--fs-xs);
  font-weight: 750;
  justify-content: center;
  min-height: 2.05rem;
}

.zw-legend > span {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
}

.swatch {
  width: 10px;
  height: 10px;
  border-radius: 3px;
  display: inline-block;
}

.swatch.base { background: #d6a44b; }
.swatch.good { background: #22c55e; }
.swatch.watch { background: #f43f5e; }
.swatch.move { background: #0ea5e9; }
.swatch.focus { background: #f59e0b; }

@media (max-width: 1180px) {
  .zw-workspace.is-overlay,
  .zw-overlay-guide {
    grid-template-columns: 1fr;
  }

  .zw-workspace.is-overlay {
    display: grid;
  }

  .zw-guide-reading {
    margin-top: 1rem;
  }
}

@media (max-width: 720px) {
  .zw-overlay {
    padding: 0.8rem;
  }

  .zw-method-grid {
    grid-template-columns: 1fr;
  }

  .zw-grid {
    grid-template-columns: repeat(4, minmax(72px, 1fr));
    grid-template-rows: repeat(4, minmax(104px, auto));
  }

  .zw-cell {
    min-height: 104px;
    padding: 0.62rem 0.42rem;
  }

  .zw-palace-name {
    font-size: var(--fs-xs);
  }

  .zw-star,
  .zw-trigger-chip,
  .zw-more-star,
  .zw-sihua-tag {
    font-size: var(--fs-2xs);
    padding-inline: 0.25rem;
  }

  .zw-dayun-range {
    right: 0.24rem;
    bottom: 0.24rem;
  }

  .zw-palace-ages {
    top: 0.28rem;
    left: 0.2rem;
    max-width: 2.55rem;
    font-size: 0.52rem;
  }

  .zw-palace-age-list {
    column-gap: 0.2rem;
    row-gap: 0.06rem;
  }

  .zw-center {
    gap: 0.34rem;
    padding: 0.48rem;
  }

  .zw-center::before {
    inset: 0.32rem;
    border-radius: 5px;
  }

  .zw-center::after {
    top: 0.32rem;
    right: 0.32rem;
    bottom: 0.32rem;
    width: 2px;
  }

  .zw-center-head {
    gap: 0.08rem;
  }

  .zw-center-kicker {
    font-size: var(--fs-2xs);
    letter-spacing: 0.08em;
  }

  .zw-center-title {
    font-size: var(--fs-body);
  }

  .zw-center-subtitle {
    font-size: var(--fs-2xs);
    line-height: 1.2;
  }

  .zw-center-duo {
    gap: 0.24rem;
  }

  .zw-center-major {
    min-height: 42px;
    padding: 0.28rem 0.22rem;
    border-radius: 6px;
  }

  .zw-center-major small,
  .zw-center-grid small {
    font-size: var(--fs-2xs);
  }

  .zw-center-major b {
    font-size: var(--fs-sm);
  }

  .zw-center-grid {
    justify-content: center;
    gap: 0.22rem;
  }

  .zw-center-grid span {
    gap: 0.22rem;
    padding: 0.18rem 0.34rem;
  }

  .zw-center-grid b {
    font-size: var(--fs-2xs);
  }

  .zw-brightness-legend {
    position: static;
    flex-basis: 100%;
    margin-left: auto;
    gap: 0.16rem;
    padding: 0;
  }

  .zw-brightness-row {
    grid-template-columns: repeat(7, 0.58rem);
    gap: 0.06rem;
  }

  .zw-brightness-dot {
    width: 0.38rem;
    height: 0.38rem;
  }

  .zw-brightness-text {
    font-size: 0.5rem;
  }

  .zw-guide-title-row {
    flex-direction: column;
  }
}
</style>
