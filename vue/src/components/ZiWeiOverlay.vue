<script setup lang="ts">
import { ref, computed } from 'vue'

interface StarInfo {
  name: string
  brightness: string
}

interface PalaceData {
  branch: string
  name: string
  mainStars: StarInfo[]
  auxStars: StarInfo[]
  sihua: string[]
}

interface Props {
  baseChart: {
    palaces: PalaceData[]
    mingZhu: string
    shenZhu: string
    wuxingJu: string
  }
  liunianChart: {
    palaces: PalaceData[]
    year: number
  }
  availableYears: number[]
}

const props = defineProps<Props>()

const emit = defineEmits<{
  (e: 'year-change', year: number): void
}>()

const mode = ref<'base' | 'overlay'>('base')
const selectedYear = ref<number>(props.liunianChart?.year || new Date().getFullYear())

const brightnessColorMap: Record<string, string> = {
  '庙': '#C41E3A',
  '旺': '#FF8C00',
  '得': '#DAA520',
  '利': '#228B22',
  '平': '#808080',
  '不': '#87CEEB',
  '陷': '#191970',
}

function starColor(brightness: string): string {
  return brightnessColorMap[brightness] || '#1A1A1A'
}

function overlayColor(brightness: string): string {
  // Liunian stars use semi-transparent teal/purple tones
  const base = brightnessColorMap[brightness] || '#6B5B95'
  return base + '99' // 60% opacity
}

function onYearChange() {
  emit('year-change', selectedYear.value)
}

// Build branch lookup for base and liunian
const baseLookup = computed<Record<string, PalaceData>>(() => {
  const m: Record<string, PalaceData> = {}
  props.baseChart.palaces.forEach((p) => { m[p.branch] = p })
  return m
})

const liunianLookup = computed<Record<string, PalaceData>>(() => {
  const m: Record<string, PalaceData> = {}
  props.liunianChart.palaces.forEach((p) => { m[p.branch] = p })
  return m
})

const branchOrder = ['巳', '午', '未', '申', '辰', '卯', '酉', '戌', '寅', '丑', '子', '亥']
const row1Branches = branchOrder.slice(0, 4)
const row4Branches = branchOrder.slice(8, 12)

function basePalaceAt(b: string): PalaceData | undefined { return baseLookup.value[b] }
function liunianPalaceAt(b: string): PalaceData | undefined { return liunianLookup.value[b] }
</script>

<template>
  <div class="overlay-container">
    <!-- Controls bar -->
    <div class="controls-bar">
      <!-- Mode toggle -->
      <div class="toggle-group">
        <button
          class="toggle-btn"
          :class="{ active: mode === 'base' }"
          @click="mode = 'base'"
        >
          本命盘
        </button>
        <button
          class="toggle-btn"
          :class="{ active: mode === 'overlay' }"
          @click="mode = 'overlay'"
        >
          流年叠盘
        </button>
      </div>

      <!-- Year selector (visible in overlay mode) -->
      <div v-if="mode === 'overlay'" class="year-selector">
        <label class="year-label">流年:</label>
        <select
          v-model="selectedYear"
          class="year-select"
          @change="onYearChange"
        >
          <option
            v-for="y in availableYears"
            :key="y"
            :value="y"
          >
            {{ y }}年
          </option>
        </select>
      </div>
    </div>

    <!-- Chart grid (same 4-col layout as ZiWeiChart) -->
    <div class="chart-grid">
      <!-- Row 1 -->
      <template v-for="branch in row1Branches" :key="'r1-' + branch">
        <div class="palace-cell" :class="{ 'overlay-mode': mode === 'overlay' }">
          <div class="palace-name">{{ basePalaceAt(branch)?.name || branch }}</div>
          <div class="palace-branch">{{ branch }}</div>

          <!-- Base stars (always shown) -->
          <div class="star-list">
            <span
              v-for="(star, si) in (basePalaceAt(branch)?.mainStars || [])"
              :key="'bs-' + si"
              class="star-tag base-star"
              :style="{ backgroundColor: starColor(star.brightness) }"
            >{{ star.name }}</span>
          </div>

          <!-- Liunian overlay stars -->
          <div v-if="mode === 'overlay' && liunianPalaceAt(branch)" class="overlay-stars">
            <span
              v-for="(star, si) in (liunianPalaceAt(branch)?.mainStars || [])"
              :key="'ls-' + si"
              class="star-tag overlay-star"
              :style="{
                backgroundColor: overlayColor(star.brightness),
                borderColor: starColor(star.brightness),
              }"
            >{{ star.name }}</span>
            <span
              v-for="(star, si) in (liunianPalaceAt(branch)?.auxStars || [])"
              :key="'la-' + si"
              class="star-tag overlay-star aux-overlay"
            >{{ star.name }}</span>
          </div>

          <!-- Sihua markers -->
          <div v-if="basePalaceAt(branch)?.sihua.length" class="sihua-row">
            <span
              v-for="(sh, si) in (basePalaceAt(branch)?.sihua || [])"
              :key="'bsh-' + si"
              class="sihua-tag"
            >{{ sh }}</span>
          </div>
        </div>
      </template>

      <!-- Row 2: 辰 + center + 酉 -->
      <div class="palace-cell" :class="{ 'overlay-mode': mode === 'overlay' }">
        <div class="palace-name">{{ basePalaceAt('辰')?.name || '辰' }}</div>
        <div class="palace-branch">辰</div>
        <div class="star-list">
          <span
            v-for="(star, si) in (basePalaceAt('辰')?.mainStars || [])"
            :key="'bs-' + si"
            class="star-tag base-star"
            :style="{ backgroundColor: starColor(star.brightness) }"
          >{{ star.name }}</span>
        </div>
        <div v-if="mode === 'overlay' && liunianPalaceAt('辰')" class="overlay-stars">
          <span
            v-for="(star, si) in (liunianPalaceAt('辰')?.mainStars || [])"
            :key="'ls-' + si"
            class="star-tag overlay-star"
            :style="{
              backgroundColor: overlayColor(star.brightness),
              borderColor: starColor(star.brightness),
            }"
          >{{ star.name }}</span>
        </div>
      </div>

      <!-- Center info -->
      <div class="center-info">
        <div class="center-title">命宫核心</div>
        <div class="center-item">
          <span class="center-label">命主</span>
          <span class="center-value">{{ baseChart.mingZhu }}</span>
        </div>
        <div class="center-item">
          <span class="center-label">身主</span>
          <span class="center-value">{{ baseChart.shenZhu }}</span>
        </div>
        <div class="center-item">
          <span class="center-label">五行局</span>
          <span class="center-value">{{ baseChart.wuxingJu }}</span>
        </div>
        <div v-if="mode === 'overlay'" class="center-item mt-2">
          <span class="center-label">流年</span>
          <span class="center-value text-[var(--color-bazi-red)]">{{ selectedYear }}</span>
        </div>
      </div>

      <div class="palace-cell" :class="{ 'overlay-mode': mode === 'overlay' }">
        <div class="palace-name">{{ basePalaceAt('酉')?.name || '酉' }}</div>
        <div class="palace-branch">酉</div>
        <div class="star-list">
          <span
            v-for="(star, si) in (basePalaceAt('酉')?.mainStars || [])"
            :key="'bs-' + si"
            class="star-tag base-star"
            :style="{ backgroundColor: starColor(star.brightness) }"
          >{{ star.name }}</span>
        </div>
        <div v-if="mode === 'overlay' && liunianPalaceAt('酉')" class="overlay-stars">
          <span
            v-for="(star, si) in (liunianPalaceAt('酉')?.mainStars || [])"
            :key="'ls-' + si"
            class="star-tag overlay-star"
            :style="{
              backgroundColor: overlayColor(star.brightness),
              borderColor: starColor(star.brightness),
            }"
          >{{ star.name }}</span>
        </div>
      </div>

      <!-- Row 3: 卯 + center + 戌 -->
      <div class="palace-cell" :class="{ 'overlay-mode': mode === 'overlay' }">
        <div class="palace-name">{{ basePalaceAt('卯')?.name || '卯' }}</div>
        <div class="palace-branch">卯</div>
        <div class="star-list">
          <span
            v-for="(star, si) in (basePalaceAt('卯')?.mainStars || [])"
            :key="'bs-' + si"
            class="star-tag base-star"
            :style="{ backgroundColor: starColor(star.brightness) }"
          >{{ star.name }}</span>
        </div>
        <div v-if="mode === 'overlay' && liunianPalaceAt('卯')" class="overlay-stars">
          <span
            v-for="(star, si) in (liunianPalaceAt('卯')?.mainStars || [])"
            :key="'ls-' + si"
            class="star-tag overlay-star"
            :style="{
              backgroundColor: overlayColor(star.brightness),
              borderColor: starColor(star.brightness),
            }"
          >{{ star.name }}</span>
        </div>
      </div>

      <div class="palace-cell" :class="{ 'overlay-mode': mode === 'overlay' }">
        <div class="palace-name">{{ basePalaceAt('戌')?.name || '戌' }}</div>
        <div class="palace-branch">戌</div>
        <div class="star-list">
          <span
            v-for="(star, si) in (basePalaceAt('戌')?.mainStars || [])"
            :key="'bs-' + si"
            class="star-tag base-star"
            :style="{ backgroundColor: starColor(star.brightness) }"
          >{{ star.name }}</span>
        </div>
        <div v-if="mode === 'overlay' && liunianPalaceAt('戌')" class="overlay-stars">
          <span
            v-for="(star, si) in (liunianPalaceAt('戌')?.mainStars || [])"
            :key="'ls-' + si"
            class="star-tag overlay-star"
            :style="{
              backgroundColor: overlayColor(star.brightness),
              borderColor: starColor(star.brightness),
            }"
          >{{ star.name }}</span>
        </div>
      </div>

      <!-- Row 4 -->
      <template v-for="branch in row4Branches" :key="'r4-' + branch">
        <div class="palace-cell" :class="{ 'overlay-mode': mode === 'overlay' }">
          <div class="palace-name">{{ basePalaceAt(branch)?.name || branch }}</div>
          <div class="palace-branch">{{ branch }}</div>
          <div class="star-list">
            <span
              v-for="(star, si) in (basePalaceAt(branch)?.mainStars || [])"
              :key="'bs-' + si"
              class="star-tag base-star"
              :style="{ backgroundColor: starColor(star.brightness) }"
            >{{ star.name }}</span>
          </div>
          <div v-if="mode === 'overlay' && liunianPalaceAt(branch)" class="overlay-stars">
            <span
              v-for="(star, si) in (liunianPalaceAt(branch)?.mainStars || [])"
              :key="'ls-' + si"
              class="star-tag overlay-star"
              :style="{
                backgroundColor: overlayColor(star.brightness),
                borderColor: starColor(star.brightness),
              }"
            >{{ star.name }}</span>
          </div>
        </div>
      </template>
    </div>

    <!-- Legend -->
    <div class="legend">
      <div class="legend-item">
        <span class="legend-dot base-dot"></span>
        <span>本命星曜</span>
      </div>
      <div v-if="mode === 'overlay'" class="legend-item">
        <span class="legend-dot overlay-dot"></span>
        <span>流年星曜</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
@reference "tailwindcss";
.overlay-container {
  @apply w-full max-w-3xl mx-auto;
}

.controls-bar {
  @apply flex items-center justify-between mb-3 flex-wrap gap-2;
}

.toggle-group {
  @apply inline-flex rounded-md overflow-hidden border;
  border-color: var(--color-bazi-blue);
}

.toggle-btn {
  @apply px-4 py-1.5 text-sm font-medium transition-colors cursor-pointer border-0;
  background-color: var(--color-bazi-paper);
  color: var(--color-bazi-blue);
}

.toggle-btn.active {
  background-color: var(--color-bazi-blue);
  color: #fff;
}

.toggle-btn:not(.active):hover {
  background-color: rgba(43, 58, 66, 0.08);
}

.year-selector {
  @apply flex items-center gap-2;
}

.year-label {
  @apply text-sm font-medium;
  color: var(--color-bazi-blue);
}

.year-select {
  @apply rounded-md border px-3 py-1 text-sm cursor-pointer;
  background-color: var(--color-bazi-paper);
  border-color: var(--color-bazi-blue);
  color: var(--color-bazi-ink);
}

.chart-grid {
  display: grid;
  grid-template-columns: 1fr 1fr 1fr 1fr;
  gap: 2px;
  background-color: var(--color-bazi-ink);
  border: 2px solid var(--color-bazi-ink);
  border-radius: 4px;
  overflow: hidden;
}

.palace-cell {
  @apply flex flex-col items-center justify-start p-2 relative;
  background-color: var(--color-bazi-paper);
  min-height: 120px;
  gap: 1px;
}

.palace-cell.overlay-mode {
  background-color: #f8f3e8;
}

.palace-name {
  @apply text-xs font-bold leading-tight;
  color: var(--color-bazi-blue);
}

.palace-branch {
  @apply text-[10px] text-gray-400 leading-none;
}

.star-list {
  @apply flex flex-col items-center gap-px w-full mt-0.5;
}

.base-star {
  @apply inline-block rounded-sm px-1 py-px text-[11px] font-bold leading-tight;
  color: #fff;
  white-space: nowrap;
}

.overlay-stars {
  @apply flex flex-col items-center gap-px w-full mt-1;
}

.overlay-star {
  @apply inline-block rounded-sm px-1 py-px text-[11px] font-bold leading-tight border-2 border-dashed;
  color: #1A1A1A;
  white-space: nowrap;
}

.overlay-star.aux-overlay {
  @apply text-[10px] font-normal border;
  background-color: rgba(107, 91, 149, 0.3) !important;
  border-color: rgba(107, 91, 149, 0.5);
}

.sihua-row {
  @apply mt-0.5 flex flex-wrap justify-center gap-0.5;
}

.sihua-tag {
  @apply rounded-full px-1.5 py-px text-[9px] font-semibold leading-tight;
  background-color: var(--color-bazi-red);
  color: white;
}

.center-info {
  @apply flex flex-col items-center justify-center p-3;
  background-color: var(--color-bazi-paper);
  grid-row: span 2;
  grid-column: span 2;
  gap: 6px;
  border-left: 1px solid var(--color-bazi-ink);
  border-right: 1px solid var(--color-bazi-ink);
}

.center-title {
  @apply text-sm font-bold;
  color: var(--color-bazi-red);
}

.center-item {
  @apply flex flex-col items-center gap-0;
}

.center-label {
  @apply text-[10px] text-gray-400 leading-tight;
}

.center-value {
  @apply text-sm font-bold leading-tight;
  color: var(--color-bazi-blue);
}

.legend {
  @apply flex justify-center gap-6 mt-4 pt-3 border-t;
  border-color: rgba(43, 58, 66, 0.1);
}

.legend-item {
  @apply flex items-center gap-2 text-xs;
  color: var(--color-bazi-blue);
}

.legend-dot {
  @apply inline-block w-3 h-3 rounded-sm;
}

.base-dot {
  background-color: var(--color-bazi-red);
}

.overlay-dot {
  @apply border-2 border-dashed;
  background-color: rgba(107, 91, 149, 0.6);
  border-color: #6B5B95;
}
</style>
