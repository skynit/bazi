<script setup lang="ts">
import { ref, computed, watch } from 'vue'

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

// Sync selectedYear when prop changes (e.g., after year switch)
watch(() => props.liunianChart?.year, (y) => {
  if (y) selectedYear.value = y
}, { immediate: true })

// Gold palette for 本命盘
const goldMeta: Record<string, { bg: string; text: string }> = {
  '庙': { bg: 'linear-gradient(135deg,#C41E3A,#8B0000)', text: '#fff' },
  '旺': { bg: 'linear-gradient(135deg,#FF8C00,#CC5500)', text: '#fff' },
  '得': { bg: 'linear-gradient(135deg,#DAA520,#B8860B)', text: '#fff' },
  '利': { bg: 'linear-gradient(135deg,#228B22,#006400)', text: '#fff' },
  '平': { bg: 'linear-gradient(135deg,#808080,#696969)', text: '#fff' },
  '不': { bg: 'linear-gradient(135deg,#5F9EA0,#4682B4)', text: '#fff' },
  '陷': { bg: 'linear-gradient(135deg,#2B3A42,#1a252e)', text: '#aaa' },
}

// Purple palette for 流年盘
const purpleMeta: Record<string, { bg: string; text: string }> = {
  '庙': { bg: 'linear-gradient(135deg,#7B2D8B,#4B0082)', text: '#fff' },
  '旺': { bg: 'linear-gradient(135deg,#9B59B6,#8E44AD)', text: '#fff' },
  '得': { bg: 'linear-gradient(135deg,#8E6DBB,#6B5B95)', text: '#fff' },
  '利': { bg: 'linear-gradient(135deg,#5D4B8B,#4A3A7A)', text: '#fff' },
  '平': { bg: 'linear-gradient(135deg,#7A6B9B,#655580)', text: '#fff' },
  '不': { bg: 'linear-gradient(135deg,#6B5B95,#5A4A84)', text: '#ddd' },
  '陷': { bg: 'linear-gradient(135deg,#3D2B5B,#2D1B4A)', text: '#999' },
}

function baseMeta(brightness: string) {
  return goldMeta[brightness] || goldMeta['陷']
}

function overlayMeta(brightness: string) {
  return purpleMeta[brightness] || purpleMeta['陷']
}

function onYearChange() {
  emit('year-change', selectedYear.value)
}

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

const branchOrder = ['巳', '午', '未', '申', '辰', '卯', '酉', '戌', '寅', '丑', '子', '亥']
const row1Branches = branchOrder.slice(0, 4)
const row4Branches = branchOrder.slice(8, 12)

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
    <div class="zw-grid" :class="{ 'zw-grid-overlay': mode === 'overlay' }">

      <!-- Row 1 -->
      <template v-for="branch in row1Branches" :key="'r1-' + branch">
        <div class="zw-cell" :class="{ 'zw-cell-overlay': mode === 'overlay' }">
          <div class="zw-cell-header">
            <span class="zw-palace-name">{{ basePalaceAt(branch)?.name || branch }}</span>
            <span class="zw-branch">{{ branch }}</span>
          </div>
          <div class="zw-stars">
            <span
              v-for="(star, si) in (basePalaceAt(branch)?.mainStars || [])"
              :key="'bs-' + si"
              class="zw-star zw-star-gold"
              :style="{ background: baseMeta(star.brightness).bg }"
            >{{ star.name }}</span>
          </div>
          <div v-if="basePalaceAt(branch)?.sihua?.length" class="zw-sihua">
            <span v-for="(sh, si) in (basePalaceAt(branch)?.sihua || [])" :key="'bsh-' + si" class="zw-sihua-tag">{{ sh }}</span>
          </div>
          <div v-if="mode === 'overlay' && liunianPalaceAt(branch)" class="zw-overlay-stars">
            <div class="zw-overlay-label">流年</div>
            <span
              v-for="(star, si) in (liunianPalaceAt(branch)?.mainStars || [])"
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
      <div class="zw-cell" :class="{ 'zw-cell-overlay': mode === 'overlay' }">
        <div class="zw-cell-header">
          <span class="zw-palace-name">{{ basePalaceAt('辰')?.name || '辰' }}</span>
          <span class="zw-branch">辰</span>
        </div>
        <div class="zw-stars">
          <span
            v-for="(star, si) in (basePalaceAt('辰')?.mainStars || [])"
            :key="'bs-' + si"
            class="zw-star zw-star-gold"
            :style="{ background: baseMeta(star.brightness).bg }"
          >{{ star.name }}</span>
        </div>
        <div v-if="mode === 'overlay' && liunianPalaceAt('辰')" class="zw-overlay-stars">
          <div class="zw-overlay-label">流年</div>
          <span
            v-for="(star, si) in (liunianPalaceAt('辰')?.mainStars || [])"
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
          <span class="zw-center-val">{{ baseChart.mingZhu }}</span>
        </div>
        <div class="zw-center-row">
          <span class="zw-center-lbl">身主</span>
          <span class="zw-center-val">{{ baseChart.shenZhu }}</span>
        </div>
        <div class="zw-center-row">
          <span class="zw-center-lbl">五行局</span>
          <span class="zw-center-val">{{ baseChart.wuxingJu }}</span>
        </div>
        <div v-if="mode === 'overlay'" class="zw-center-row zw-center-year">
          <span class="zw-center-lbl">流年</span>
          <span class="zw-center-val zw-year-val">{{ selectedYear }}</span>
        </div>
      </div>

      <div class="zw-cell" :class="{ 'zw-cell-overlay': mode === 'overlay' }">
        <div class="zw-cell-header">
          <span class="zw-palace-name">{{ basePalaceAt('酉')?.name || '酉' }}</span>
          <span class="zw-branch">酉</span>
        </div>
        <div class="zw-stars">
          <span
            v-for="(star, si) in (basePalaceAt('酉')?.mainStars || [])"
            :key="'bs-' + si"
            class="zw-star zw-star-gold"
            :style="{ background: baseMeta(star.brightness).bg }"
          >{{ star.name }}</span>
        </div>
        <div v-if="mode === 'overlay' && liunianPalaceAt('酉')" class="zw-overlay-stars">
          <div class="zw-overlay-label">流年</div>
          <span
            v-for="(star, si) in (liunianPalaceAt('酉')?.mainStars || [])"
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
      <div class="zw-cell" :class="{ 'zw-cell-overlay': mode === 'overlay' }">
        <div class="zw-cell-header">
          <span class="zw-palace-name">{{ basePalaceAt('卯')?.name || '卯' }}</span>
          <span class="zw-branch">卯</span>
        </div>
        <div class="zw-stars">
          <span
            v-for="(star, si) in (basePalaceAt('卯')?.mainStars || [])"
            :key="'bs-' + si"
            class="zw-star zw-star-gold"
            :style="{ background: baseMeta(star.brightness).bg }"
          >{{ star.name }}</span>
        </div>
        <div v-if="mode === 'overlay' && liunianPalaceAt('卯')" class="zw-overlay-stars">
          <div class="zw-overlay-label">流年</div>
          <span
            v-for="(star, si) in (liunianPalaceAt('卯')?.mainStars || [])"
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

      <div class="zw-cell" :class="{ 'zw-cell-overlay': mode === 'overlay' }">
        <div class="zw-cell-header">
          <span class="zw-palace-name">{{ basePalaceAt('戌')?.name || '戌' }}</span>
          <span class="zw-branch">戌</span>
        </div>
        <div class="zw-stars">
          <span
            v-for="(star, si) in (basePalaceAt('戌')?.mainStars || [])"
            :key="'bs-' + si"
            class="zw-star zw-star-gold"
            :style="{ background: baseMeta(star.brightness).bg }"
          >{{ star.name }}</span>
        </div>
        <div v-if="mode === 'overlay' && liunianPalaceAt('戌')" class="zw-overlay-stars">
          <div class="zw-overlay-label">流年</div>
          <span
            v-for="(star, si) in (liunianPalaceAt('戌')?.mainStars || [])"
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
        <div class="zw-cell" :class="{ 'zw-cell-overlay': mode === 'overlay' }">
          <div class="zw-cell-header">
            <span class="zw-palace-name">{{ basePalaceAt(branch)?.name || branch }}</span>
            <span class="zw-branch">{{ branch }}</span>
          </div>
          <div class="zw-stars">
            <span
              v-for="(star, si) in (basePalaceAt(branch)?.mainStars || [])"
              :key="'bs-' + si"
              class="zw-star zw-star-gold"
              :style="{ background: baseMeta(star.brightness).bg }"
            >{{ star.name }}</span>
          </div>
          <div v-if="basePalaceAt(branch)?.sihua?.length" class="zw-sihua">
            <span v-for="(sh, si) in (basePalaceAt(branch)?.sihua || [])" :key="'bsh-' + si" class="zw-sihua-tag">{{ sh }}</span>
          </div>
          <div v-if="mode === 'overlay' && liunianPalaceAt(branch)" class="zw-overlay-stars">
            <div class="zw-overlay-label">流年</div>
            <span
              v-for="(star, si) in (liunianPalaceAt(branch)?.mainStars || [])"
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
    </div>
  </div>
</template>

<style scoped>
/* ── Page ── */
.zw-overlay {
  width: 100%;
  background: linear-gradient(160deg, rgba(20,14,35,0.95) 0%, rgba(8,5,15,0.98) 100%);
  border: 1px solid rgba(212,168,75,0.1);
  border-radius: 16px;
  padding: 1.25rem;
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
  background: rgba(255,255,255,0.03);
  border: 1px solid rgba(212,168,75,0.1);
  border-radius: 10px;
  padding: 3px;
  gap: 2px;
}
.zw-tab {
  display: flex; align-items: center; gap: 6px;
  padding: 0.45rem 1rem;
  border: none; border-radius: 7px;
  background: transparent;
  color: rgba(255,255,255,0.35);
  font-size: 0.8rem; font-weight: 600;
  cursor: pointer; transition: all 0.3s;
  letter-spacing: 0.5px;
}
.zw-tab:hover { color: rgba(255,255,255,0.6); background: rgba(255,255,255,0.04); }
.zw-tab.is-active { background: rgba(212,168,75,0.12); color: #D4A84B; }
.zw-tab-dot {
  width: 7px; height: 7px; border-radius: 50%;
  flex-shrink: 0;
}
.zw-dot-gold { background: #D4A84B; box-shadow: 0 0 8px rgba(212,168,75,0.5); }
.zw-dot-purple { background: #8E6DBB; box-shadow: 0 0 8px rgba(142,109,187,0.5); }
.zw-year-select { display: flex; align-items: center; gap: 0.5rem; }
.zw-year-label { font-size: 0.78rem; color: rgba(255,255,255,0.35); letter-spacing: 1px; }
.zw-select {
  background: rgba(255,255,255,0.04);
  border: 1px solid rgba(142,109,187,0.25);
  border-radius: 8px;
  color: #8E6DBB;
  font-size: 0.82rem; font-weight: 700;
  padding: 0.35rem 0.75rem;
  cursor: pointer; outline: none;
  transition: border-color 0.3s;
}
.zw-select:hover { border-color: rgba(142,109,187,0.45); }

/* ── Grid ── */
.zw-grid {
  display: grid;
  grid-template-columns: 1fr 1fr 1fr 1fr;
  gap: 2px;
  background: rgba(0,0,0,0.4);
  border: 1px solid rgba(212,168,75,0.08);
  border-radius: 8px;
  overflow: hidden;
}
.zw-grid-overlay { background: rgba(20,15,40,0.6); border-color: rgba(142,109,187,0.12); }

/* ── Cell ── */
.zw-cell {
  background: linear-gradient(180deg, rgba(43,37,24,0.85) 0%, rgba(30,25,15,0.9) 100%);
  min-height: 110px;
  padding: 0.6rem 0.4rem;
  display: flex; flex-direction: column; align-items: center; gap: 3px;
  position: relative;
  transition: background 0.3s;
}
.zw-cell:hover { background: linear-gradient(180deg, rgba(60,50,30,0.9) 0%, rgba(40,32,18,0.95) 100%); }
.zw-cell-overlay { background: linear-gradient(180deg, rgba(30,25,45,0.85) 0%, rgba(20,15,35,0.9) 100%); }
.zw-cell-overlay:hover { background: linear-gradient(180deg, rgba(40,35,60,0.9) 0%, rgba(25,20,45,0.95) 100%); }

/* Cell header */
.zw-cell-header {
  display: flex; flex-direction: column; align-items: center; gap: 1px;
  margin-bottom: 2px;
}
.zw-palace-name {
  font-size: 0.7rem; font-weight: 800;
  color: #D4A84B; letter-spacing: 1px;
}
.zw-branch {
  font-size: 0.58rem; color: rgba(212,168,75,0.3);
}

/* Stars */
.zw-stars { display: flex; flex-direction: column; align-items: center; gap: 2px; }
.zw-star {
  display: inline-block;
  border-radius: 4px;
  padding: 1px 5px;
  font-size: 0.68rem; font-weight: 800;
  color: #fff;
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
.zw-sihua { display: flex; flex-wrap: wrap; justify-content: center; gap: 2px; margin-top: 2px; }
.zw-sihua-tag {
  font-size: 0.58rem; font-weight: 700;
  padding: 1px 5px;
  background: rgba(196,30,58,0.15);
  border: 1px solid rgba(196,30,58,0.3);
  border-radius: 20px;
  color: #C41E3A;
}

/* Overlay stars */
.zw-overlay-stars {
  display: flex; flex-direction: column; align-items: center; gap: 2px;
  margin-top: 4px;
  padding-top: 4px;
  border-top: 1px dashed rgba(142,109,187,0.2);
  width: 100%;
}
.zw-overlay-label {
  font-size: 0.55rem; font-weight: 700;
  color: rgba(142,109,187,0.5);
  letter-spacing: 2px;
  text-transform: uppercase;
  margin-bottom: 1px;
}

/* 流耀 */
.zw-liuyao {
  display: flex; flex-wrap: wrap; justify-content: center; gap: 2px;
  margin-top: 3px; padding-top: 3px;
  border-top: 1px dashed rgba(142,109,187,0.15);
  width: 100%;
}
.zw-liuyao-chip {
  font-size: 0.52rem; font-weight: 600;
  padding: 0px 4px;
  background: rgba(142,109,187,0.08);
  border: 1px solid rgba(142,109,187,0.2);
  border-radius: 3px;
  color: #b39ddb;
  white-space: nowrap;
}

/* ── Center ── */
.zw-center {
  background: linear-gradient(180deg, rgba(43,37,24,0.9) 0%, rgba(20,16,28,0.95) 100%);
  grid-row: span 2;
  grid-column: span 2;
  display: flex; flex-direction: column; align-items: center; justify-content: center;
  gap: 6px;
  padding: 0.75rem;
  position: relative; overflow: hidden;
  border-left: 1px solid rgba(212,168,75,0.06);
  border-right: 1px solid rgba(212,168,75,0.06);
}
.zw-center-overlay {
  background: linear-gradient(180deg, rgba(50,40,80,0.9) 0%, rgba(25,20,50,0.95) 100%);
  border-left-color: rgba(142,109,187,0.1);
  border-right-color: rgba(142,109,187,0.1);
}
.zw-center-glow {
  position: absolute; inset: 0;
  background: radial-gradient(circle, rgba(212,168,75,0.06), transparent 70%);
  pointer-events: none;
}
.zw-center-overlay .zw-center-glow {
  background: radial-gradient(circle, rgba(142,109,187,0.08), transparent 70%);
}
.zw-center-title {
  font-family: var(--font-serif);
  font-size: 0.78rem; font-weight: 800;
  color: #D4A84B; letter-spacing: 2px;
  text-shadow: 0 0 20px rgba(212,168,75,0.3);
  position: relative;
}
.zw-center-overlay .zw-center-title {
  color: #8E6DBB;
  text-shadow: 0 0 20px rgba(142,109,187,0.4);
}
.zw-center-row { display: flex; flex-direction: column; align-items: center; gap: 0; position: relative; }
.zw-center-lbl { font-size: 0.55rem; color: rgba(255,255,255,0.2); letter-spacing: 1px; text-transform: uppercase; }
.zw-center-val { font-size: 0.78rem; font-weight: 800; color: rgba(255,255,255,0.75); }
.zw-year-val { color: #8E6DBB; font-size: 0.9rem; text-shadow: 0 0 15px rgba(142,109,187,0.4); }


/* ── Legend ── */
.zw-legend {
  display: flex; justify-content: center; gap: 2rem;
  margin-top: 1rem; padding-top: 0.75rem;
  border-top: 1px solid rgba(212,168,75,0.06);
}
.zw-legend-item { display: flex; align-items: center; gap: 0.5rem; }
.zw-legend-swatch {
  width: 10px; height: 10px; border-radius: 3px;
}
.zw-swatch-gold { background: linear-gradient(135deg, #DAA520, #B8860B); box-shadow: 0 0 8px rgba(212,168,75,0.4); }
.zw-swatch-purple { background: linear-gradient(135deg, #8E6DBB, #6B5B95); box-shadow: 0 0 8px rgba(142,109,187,0.4); }
.zw-legend-text { font-size: 0.72rem; color: rgba(255,255,255,0.35); letter-spacing: 1px; }
</style>