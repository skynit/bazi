<script setup lang="ts">
import { computed } from 'vue'

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
  palaces: PalaceData[]
  mingZhu: string
  shenZhu: string
  wuxingJu: string
  patterns: string[]
}

const props = defineProps<Props>()

const brightnessColorMap: Record<string, string> = {
  '庙': '#C41E3A',
  '旺': '#FF8C00',
  '得': '#DAA520',
  '利': '#228B22',
  '平': '#808080',
  '不': '#87CEEB',
  '陷': '#191970',
}

function getStarColor(brightness: string): string {
  return brightnessColorMap[brightness] || '#1A1A1A'
}

function getStarTextColor(brightness: string): string {
  if (['陷', '利', '庙'].includes(brightness)) return '#fff'
  return '#1A1A1A'
}

// Build branch → palace lookup
const branchLookup = computed<Record<string, PalaceData>>(() => {
  const m: Record<string, PalaceData> = {}
  props.palaces.forEach((p) => { m[p.branch] = p })
  return m
})

function palaceAt(branch: string): PalaceData | undefined {
  return branchLookup.value[branch]
}
</script>

<template>
  <div class="ziwei-chart-container">
    <!-- Pattern badges -->
    <div v-if="patterns.length" class="mb-4 flex flex-wrap justify-center gap-2">
      <span
        v-for="(pat, idx) in patterns"
        :key="idx"
        class="inline-block rounded-full bg-[var(--color-bazi-red)] px-3 py-1 text-xs font-semibold text-white shadow-sm"
      >
        {{ pat }}
      </span>
    </div>

    <!--
      Traditional square chart layout (4col grid):
      Row 1: 巳  午  未  申
      Row 2: 辰  [center]  酉
      Row 3: 卯  [center]  戌
      Row 4: 寅  丑  子  亥
      center spans col 2-3, row 2-3
    -->
    <div class="chart-grid">
      <!-- Row 1: 巳 午 未 申 -->
      <template v-for="branch in ['巳', '午', '未', '申']" :key="'r1-' + branch">
        <div class="palace-cell">
          <template v-if="palaceAt(branch)">
            <div class="palace-name">{{ palaceAt(branch)!.name }}</div>
            <div class="palace-branch">{{ branch }}</div>
            <div class="star-list main">
              <span
                v-for="(star, si) in palaceAt(branch)!.mainStars"
                :key="'ms-' + si"
                class="star-tag"
                :style="{
                  backgroundColor: getStarColor(star.brightness),
                  color: getStarTextColor(star.brightness),
                }"
              >{{ star.name }}</span>
            </div>
            <div class="star-list aux">
              <span
                v-for="(star, si) in palaceAt(branch)!.auxStars"
                :key="'as-' + si"
                class="aux-star"
              >{{ star.name }}</span>
            </div>
            <div v-if="palaceAt(branch)!.sihua.length" class="sihua-row">
              <span
                v-for="(sh, si) in palaceAt(branch)!.sihua"
                :key="'sh-' + si"
                class="sihua-tag"
              >{{ sh }}</span>
            </div>
          </template>
          <div v-else class="palace-empty">{{ branch }}</div>
        </div>
      </template>

      <!-- Row 2: 辰 + center + 酉 -->
      <div class="palace-cell">
        <template v-if="palaceAt('辰')">
          <div class="palace-name">{{ palaceAt('辰')!.name }}</div>
          <div class="palace-branch">辰</div>
          <div class="star-list main">
            <span
              v-for="(star, si) in palaceAt('辰')!.mainStars"
              :key="'ms-' + si"
              class="star-tag"
              :style="{
                backgroundColor: getStarColor(star.brightness),
                color: getStarTextColor(star.brightness),
              }"
            >{{ star.name }}</span>
          </div>
          <div class="star-list aux">
            <span
              v-for="(star, si) in palaceAt('辰')!.auxStars"
              :key="'as-' + si"
              class="aux-star"
            >{{ star.name }}</span>
          </div>
          <div v-if="palaceAt('辰')?.sihua.length" class="sihua-row">
            <span
              v-for="(sh, si) in palaceAt('辰')!.sihua"
              :key="'sh-' + si"
              class="sihua-tag"
            >{{ sh }}</span>
          </div>
        </template>
        <div v-else class="palace-empty">辰</div>
      </div>

      <!-- Center info (spans row 2-3, col 2-3) -->
      <div class="center-info">
        <div class="center-title">命宫核心</div>
        <div class="center-item">
          <span class="center-label">命主</span>
          <span class="center-value">{{ mingZhu }}</span>
        </div>
        <div class="center-item">
          <span class="center-label">身主</span>
          <span class="center-value">{{ shenZhu }}</span>
        </div>
        <div class="center-item">
          <span class="center-label">五行局</span>
          <span class="center-value">{{ wuxingJu }}</span>
        </div>
      </div>

      <div class="palace-cell">
        <template v-if="palaceAt('酉')">
          <div class="palace-name">{{ palaceAt('酉')!.name }}</div>
          <div class="palace-branch">酉</div>
          <div class="star-list main">
            <span
              v-for="(star, si) in palaceAt('酉')!.mainStars"
              :key="'ms-' + si"
              class="star-tag"
              :style="{
                backgroundColor: getStarColor(star.brightness),
                color: getStarTextColor(star.brightness),
              }"
            >{{ star.name }}</span>
          </div>
          <div class="star-list aux">
            <span
              v-for="(star, si) in palaceAt('酉')!.auxStars"
              :key="'as-' + si"
              class="aux-star"
            >{{ star.name }}</span>
          </div>
          <div v-if="palaceAt('酉')?.sihua.length" class="sihua-row">
            <span
              v-for="(sh, si) in palaceAt('酉')!.sihua"
              :key="'sh-' + si"
              class="sihua-tag"
            >{{ sh }}</span>
          </div>
        </template>
        <div v-else class="palace-empty">酉</div>
      </div>

      <!-- Row 3: 卯 + center (already placed) + 戌 -->
      <div class="palace-cell">
        <template v-if="palaceAt('卯')">
          <div class="palace-name">{{ palaceAt('卯')!.name }}</div>
          <div class="palace-branch">卯</div>
          <div class="star-list main">
            <span
              v-for="(star, si) in palaceAt('卯')!.mainStars"
              :key="'ms-' + si"
              class="star-tag"
              :style="{
                backgroundColor: getStarColor(star.brightness),
                color: getStarTextColor(star.brightness),
              }"
            >{{ star.name }}</span>
          </div>
          <div class="star-list aux">
            <span
              v-for="(star, si) in palaceAt('卯')!.auxStars"
              :key="'as-' + si"
              class="aux-star"
            >{{ star.name }}</span>
          </div>
          <div v-if="palaceAt('卯')?.sihua.length" class="sihua-row">
            <span
              v-for="(sh, si) in palaceAt('卯')!.sihua"
              :key="'sh-' + si"
              class="sihua-tag"
            >{{ sh }}</span>
          </div>
        </template>
        <div v-else class="palace-empty">卯</div>
      </div>

      <div class="palace-cell">
        <template v-if="palaceAt('戌')">
          <div class="palace-name">{{ palaceAt('戌')!.name }}</div>
          <div class="palace-branch">戌</div>
          <div class="star-list main">
            <span
              v-for="(star, si) in palaceAt('戌')!.mainStars"
              :key="'ms-' + si"
              class="star-tag"
              :style="{
                backgroundColor: getStarColor(star.brightness),
                color: getStarTextColor(star.brightness),
              }"
            >{{ star.name }}</span>
          </div>
          <div class="star-list aux">
            <span
              v-for="(star, si) in palaceAt('戌')!.auxStars"
              :key="'as-' + si"
              class="aux-star"
            >{{ star.name }}</span>
          </div>
          <div v-if="palaceAt('戌')?.sihua.length" class="sihua-row">
            <span
              v-for="(sh, si) in palaceAt('戌')!.sihua"
              :key="'sh-' + si"
              class="sihua-tag"
            >{{ sh }}</span>
          </div>
        </template>
        <div v-else class="palace-empty">戌</div>
      </div>

      <!-- Row 4: 寅 丑 子 亥 -->
      <template v-for="branch in ['寅', '丑', '子', '亥']" :key="'r4-' + branch">
        <div class="palace-cell">
          <template v-if="palaceAt(branch)">
            <div class="palace-name">{{ palaceAt(branch)!.name }}</div>
            <div class="palace-branch">{{ branch }}</div>
            <div class="star-list main">
              <span
                v-for="(star, si) in palaceAt(branch)!.mainStars"
                :key="'ms-' + si"
                class="star-tag"
                :style="{
                  backgroundColor: getStarColor(star.brightness),
                  color: getStarTextColor(star.brightness),
                }"
              >{{ star.name }}</span>
            </div>
            <div class="star-list aux">
              <span
                v-for="(star, si) in palaceAt(branch)!.auxStars"
                :key="'as-' + si"
                class="aux-star"
              >{{ star.name }}</span>
            </div>
            <div v-if="palaceAt(branch)!.sihua.length" class="sihua-row">
              <span
                v-for="(sh, si) in palaceAt(branch)!.sihua"
                :key="'sh-' + si"
                class="sihua-tag"
              >{{ sh }}</span>
            </div>
          </template>
          <div v-else class="palace-empty">{{ branch }}</div>
        </div>
      </template>
    </div>
  </div>
</template>

<style scoped>
@reference "tailwindcss";
.ziwei-chart-container {
  @apply w-full max-w-3xl mx-auto;
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
  @apply flex flex-col items-center justify-start p-2;
  background-color: var(--color-bazi-paper);
  min-height: 130px;
  gap: 1px;
}

.palace-name {
  @apply text-xs font-bold leading-tight;
  color: var(--color-bazi-blue);
}

.palace-branch {
  @apply text-[10px] text-gray-400 leading-none;
}

.palace-empty {
  @apply text-gray-300 text-sm self-center my-auto;
}

.star-list {
  @apply flex flex-col items-center gap-px w-full;
}

.star-list.main {
  @apply mt-0.5;
}

.star-tag {
  @apply inline-block rounded-sm px-1 py-px text-[11px] font-bold leading-tight;
  white-space: nowrap;
}

.star-list.aux {
  @apply mt-0.5;
}

.aux-star {
  @apply text-[10px] text-gray-500 leading-tight;
  white-space: nowrap;
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
</style>
