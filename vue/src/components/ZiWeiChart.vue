<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'

interface StarInfo {
  name: string
  type: string
  scope: string
  brightness: string
}

interface PalaceData {
  branch: string
  name: string
  heavenly_stem: string
  is_body_palace: boolean
  stars: StarInfo[]
  four_hua: string[]
}

interface Props {
  palaces: PalaceData[]
  life_master: string
  body_master: string
  five_bureau: string
  patterns: string[]
}

const props = defineProps<Props>()

function majorStars(p: PalaceData): StarInfo[] {
  return p.stars.filter(s => s.type === 'major')
}

function auxStars(p: PalaceData): StarInfo[] {
  return p.stars.filter(s => s.type !== 'major')
}

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

// Light-mode brightness palette (brighter, more contrast on light bg)
const brightnessMetaLight: Record<string, { bg: string; text: string; label: string }> = {
  庙: { bg: 'linear-gradient(135deg,#e11d48,#be123c)', text: '#fffaf8', label: '庙' },
  旺: { bg: 'linear-gradient(135deg,#ea580c,#c2410c)', text: '#fffaf8', label: '旺' },
  得: { bg: 'linear-gradient(135deg,#eab308,#a16207)', text: '#fffaf8', label: '得' },
  利: { bg: 'linear-gradient(135deg,#16a34a,#15803d)', text: '#fffaf8', label: '利' },
  平: { bg: 'linear-gradient(135deg,#6b7280,#4b5563)', text: '#fffaf8', label: '平' },
  不: { bg: 'linear-gradient(135deg,#0e7490,#0c5a74)', text: '#fffaf8', label: '不' },
  陷: { bg: 'linear-gradient(135deg,#44403c,#292524)', text: '#e7e5e4', label: '陷' },
}

// Dark-mode brightness palette (original)
const brightnessMetaDark: Record<string, { bg: string; text: string; label: string }> = {
  庙: { bg: 'linear-gradient(135deg,#fb7185,#be123c)', text: '#fffaf8', label: '庙' },
  旺: { bg: 'linear-gradient(135deg,#FF8C00,#CC5500)', text: '#fffaf8', label: '旺' },
  得: { bg: 'linear-gradient(135deg,#fde68a,#94a3b8)', text: '#10140f', label: '得' },
  利: { bg: 'linear-gradient(135deg,#34d399,#059669)', text: '#00140e', label: '利' },
  平: { bg: 'linear-gradient(135deg,#808080,#696969)', text: '#fffaf8', label: '平' },
  不: { bg: 'linear-gradient(135deg,#5F9EA0,#4682B4)', text: '#fffaf8', label: '不' },
  陷: { bg: 'linear-gradient(135deg,#2B3A42,#1a252e)', text: '#dbe4e8', label: '陷' },
}

function starMeta(brightness: string) {
  const meta = isDark.value ? brightnessMetaDark : brightnessMetaLight
  return meta[brightness] || { bg: 'var(--glass-bg)', text: 'var(--text-dim)', label: '' }
}

// Build branch → palace lookup
const branchLookup = computed<Record<string, PalaceData>>(() => {
  const m: Record<string, PalaceData> = {}
  props.palaces.forEach((p) => {
    m[p.branch] = p
  })
  return m
})

function palaceAt(branch: string): PalaceData | undefined {
  return branchLookup.value[branch]
}
</script>

<template>
  <div class="ziwei-wrapper">
    <!-- Section header with explanation -->
    <div class="chart-section-header">
      <div class="chart-section-title-group">
        <span class="chart-section-symbol">◈</span>
        <div>
          <h2 class="chart-section-title">本命盘</h2>
          <p class="chart-section-desc">出生时星曜分布，一生固定不变的命运蓝图</p>
        </div>
      </div>
    </div>

    <!-- Pattern badges -->
    <div v-if="patterns.length" class="patterns-bar">
      <span v-for="(pat, idx) in patterns" :key="idx" class="pattern-badge">{{ pat }}</span>
    </div>

    <!-- Main chart grid -->
    <div v-if="palaces.length > 0" class="chart-wrapper">
      <div class="chart-outer-frame">
        <!-- Corner decorations -->
        <div class="corner corner-tl"></div>
        <div class="corner corner-tr"></div>
        <div class="corner corner-bl"></div>
        <div class="corner corner-br"></div>

        <div class="chart-inner">
          <!-- Row 1: 巳 午 未 申 -->
          <div class="palace-row">
            <div
              v-for="branch in ['巳', '午', '未', '申']"
              :key="'r1-' + branch"
              class="palace-cell"
              :class="{ 'has-sihua': palaceAt(branch)?.four_hua?.length }"
            >
              <template v-if="palaceAt(branch)">
                <div class="palace-header">
                  <span class="palace-name">{{ palaceAt(branch)!.name }}</span>
                  <span class="palace-branch">{{ branch }}</span>
                </div>
                <div class="star-section">
                  <span
                    v-for="(star, si) in majorStars(palaceAt(branch)!)"
                    :key="'ms-' + si"
                    class="main-star"
                    :style="{
                      background: starMeta(star.brightness).bg,
                      color: starMeta(star.brightness).text,
                    }"
                    :title="star.brightness"
                  >
                    {{ star.name
                    }}<span v-if="star.brightness" class="brightness-dot">{{ starMeta(star.brightness).label }}</span>
                  </span>
                </div>
                <div v-if="auxStars(palaceAt(branch)!).length" class="aux-section">
                  <span
                    v-for="(star, si) in auxStars(palaceAt(branch)!)"
                    :key="'as-' + si"
                    class="aux-star"
                    >{{ star.name }}</span
                  >
                </div>
                <div v-if="palaceAt(branch)!.four_hua?.length" class="sihua-section">
                  <span
                    v-for="(sh, si) in palaceAt(branch)!.four_hua"
                    :key="'sh-' + si"
                    class="sihua-tag"
                    >{{ sh }}</span
                  >
                </div>
              </template>
              <div v-else class="palace-empty">{{ branch }}</div>
            </div>
          </div>

          <!-- Row 2: 辰 .. center .. 酉 -->
          <div class="palace-row row2">
            <div class="palace-cell" :class="{ 'has-sihua': palaceAt('辰')?.four_hua?.length }">
              <template v-if="palaceAt('辰')">
                <div class="palace-header">
                  <span class="palace-name">{{ palaceAt('辰')!.name }}</span>
                  <span class="palace-branch">辰</span>
                </div>
                <div class="star-section">
                  <span
                    v-for="(star, si) in majorStars(palaceAt('辰')!)"
                    :key="'ms-' + si"
                    class="main-star"
                    :style="{
                      background: starMeta(star.brightness).bg,
                      color: starMeta(star.brightness).text,
                    }"
                    >{{ star.name
                    }}<span v-if="star.brightness" class="brightness-dot">{{
                      starMeta(star.brightness).label
                    }}</span></span
                  >
                </div>
                <div v-if="auxStars(palaceAt('辰')!).length" class="aux-section">
                  <span
                    v-for="(star, si) in auxStars(palaceAt('辰')!)"
                    :key="'as-' + si"
                    class="aux-star"
                    >{{ star.name }}</span
                  >
                </div>
                <div v-if="palaceAt('辰')!.four_hua?.length" class="sihua-section">
                  <span
                    v-for="(sh, si) in palaceAt('辰')!.four_hua"
                    :key="'sh-' + si"
                    class="sihua-tag"
                    >{{ sh }}</span
                  >
                </div>
              </template>
              <div v-else class="palace-empty">辰</div>
            </div>

            <!-- Center: 命宫核心 -->
            <div class="center-cell">
              <div class="center-glow"></div>
              <div class="center-inner">
                <div class="center-ornament">✦</div>
                <div class="center-title">命宫</div>
                <div class="center-divider"></div>
                <div class="center-item">
                  <span class="center-key">命主</span>
                  <span class="center-val mingzhu">{{ life_master }}</span>
                </div>
                <div class="center-item">
                  <span class="center-key">身主</span>
                  <span class="center-val shenzhu">{{ body_master }}</span>
                </div>
                <div class="center-item">
                  <span class="center-key">五行局</span>
                  <span class="center-val wuxing">{{ five_bureau }}</span>
                </div>
                <div class="center-divider"></div>
                <div class="center-ornament">✦</div>
              </div>
            </div>

            <div class="palace-cell" :class="{ 'has-sihua': palaceAt('酉')?.four_hua?.length }">
              <template v-if="palaceAt('酉')">
                <div class="palace-header">
                  <span class="palace-name">{{ palaceAt('酉')!.name }}</span>
                  <span class="palace-branch">酉</span>
                </div>
                <div class="star-section">
                  <span
                    v-for="(star, si) in majorStars(palaceAt('酉')!)"
                    :key="'ms-' + si"
                    class="main-star"
                    :style="{
                      background: starMeta(star.brightness).bg,
                      color: starMeta(star.brightness).text,
                    }"
                    >{{ star.name
                    }}<span v-if="star.brightness" class="brightness-dot">{{
                      starMeta(star.brightness).label
                    }}</span></span
                  >
                </div>
                <div v-if="auxStars(palaceAt('酉')!).length" class="aux-section">
                  <span
                    v-for="(star, si) in auxStars(palaceAt('酉')!)"
                    :key="'as-' + si"
                    class="aux-star"
                    >{{ star.name }}</span
                  >
                </div>
                <div v-if="palaceAt('酉')!.four_hua?.length" class="sihua-section">
                  <span
                    v-for="(sh, si) in palaceAt('酉')!.four_hua"
                    :key="'sh-' + si"
                    class="sihua-tag"
                    >{{ sh }}</span
                  >
                </div>
              </template>
              <div v-else class="palace-empty">酉</div>
            </div>
          </div>

          <!-- Row 3: 卯 .. center .. 戌 -->
          <div class="palace-row row3">
            <div class="palace-cell" :class="{ 'has-sihua': palaceAt('卯')?.four_hua?.length }">
              <template v-if="palaceAt('卯')">
                <div class="palace-header">
                  <span class="palace-name">{{ palaceAt('卯')!.name }}</span>
                  <span class="palace-branch">卯</span>
                </div>
                <div class="star-section">
                  <span
                    v-for="(star, si) in majorStars(palaceAt('卯')!)"
                    :key="'ms-' + si"
                    class="main-star"
                    :style="{
                      background: starMeta(star.brightness).bg,
                      color: starMeta(star.brightness).text,
                    }"
                    >{{ star.name
                    }}<span v-if="star.brightness" class="brightness-dot">{{
                      starMeta(star.brightness).label
                    }}</span></span
                  >
                </div>
                <div v-if="auxStars(palaceAt('卯')!).length" class="aux-section">
                  <span
                    v-for="(star, si) in auxStars(palaceAt('卯')!)"
                    :key="'as-' + si"
                    class="aux-star"
                    >{{ star.name }}</span
                  >
                </div>
                <div v-if="palaceAt('卯')!.four_hua?.length" class="sihua-section">
                  <span
                    v-for="(sh, si) in palaceAt('卯')!.four_hua"
                    :key="'sh-' + si"
                    class="sihua-tag"
                    >{{ sh }}</span
                  >
                </div>
              </template>
              <div v-else class="palace-empty">卯</div>
            </div>

            <div class="center-cell center-cell-mid">
              <div class="center-glow"></div>
              <div class="center-inner">
                <div class="sky-pointer">
                  <svg width="40" height="40" viewBox="0 0 40 40" fill="none">
                    <circle
                      cx="20"
                      cy="20"
                      r="18"
                      stroke="currentColor"
                      stroke-width="1"
                      stroke-dasharray="2 3"
                    />
                    <circle cx="20" cy="20" r="6" fill="currentColor" opacity="0.3" />
                    <circle cx="20" cy="20" r="3" fill="currentColor" />
                  </svg>
                </div>
                <div class="sky-text">天宫图</div>
              </div>
            </div>

            <div class="palace-cell" :class="{ 'has-sihua': palaceAt('戌')?.four_hua?.length }">
              <template v-if="palaceAt('戌')">
                <div class="palace-header">
                  <span class="palace-name">{{ palaceAt('戌')!.name }}</span>
                  <span class="palace-branch">戌</span>
                </div>
                <div class="star-section">
                  <span
                    v-for="(star, si) in majorStars(palaceAt('戌')!)"
                    :key="'ms-' + si"
                    class="main-star"
                    :style="{
                      background: starMeta(star.brightness).bg,
                      color: starMeta(star.brightness).text,
                    }"
                    >{{ star.name
                    }}<span v-if="star.brightness" class="brightness-dot">{{
                      starMeta(star.brightness).label
                    }}</span></span
                  >
                </div>
                <div v-if="auxStars(palaceAt('戌')!).length" class="aux-section">
                  <span
                    v-for="(star, si) in auxStars(palaceAt('戌')!)"
                    :key="'as-' + si"
                    class="aux-star"
                    >{{ star.name }}</span
                  >
                </div>
                <div v-if="palaceAt('戌')!.four_hua?.length" class="sihua-section">
                  <span
                    v-for="(sh, si) in palaceAt('戌')!.four_hua"
                    :key="'sh-' + si"
                    class="sihua-tag"
                    >{{ sh }}</span
                  >
                </div>
              </template>
              <div v-else class="palace-empty">戌</div>
            </div>
          </div>

          <!-- Row 4: 寅 丑 子 亥 -->
          <div class="palace-row">
            <div
              v-for="branch in ['寅', '丑', '子', '亥']"
              :key="'r4-' + branch"
              class="palace-cell"
              :class="{ 'has-sihua': palaceAt(branch)?.four_hua?.length }"
            >
              <template v-if="palaceAt(branch)">
                <div class="palace-header">
                  <span class="palace-name">{{ palaceAt(branch)!.name }}</span>
                  <span class="palace-branch">{{ branch }}</span>
                </div>
                <div class="star-section">
                  <span
                    v-for="(star, si) in majorStars(palaceAt(branch)!)"
                    :key="'ms-' + si"
                    class="main-star"
                    :style="{
                      background: starMeta(star.brightness).bg,
                      color: starMeta(star.brightness).text,
                    }"
                    >{{ star.name
                    }}<span v-if="star.brightness" class="brightness-dot">{{
                      starMeta(star.brightness).label
                    }}</span></span
                  >
                </div>
                <div v-if="auxStars(palaceAt(branch)!).length" class="aux-section">
                  <span
                    v-for="(star, si) in auxStars(palaceAt(branch)!)"
                    :key="'as-' + si"
                    class="aux-star"
                    >{{ star.name }}</span
                  >
                </div>
                <div v-if="palaceAt(branch)!.four_hua?.length" class="sihua-section">
                  <span
                    v-for="(sh, si) in palaceAt(branch)!.four_hua"
                    :key="'sh-' + si"
                    class="sihua-tag"
                    >{{ sh }}</span
                  >
                </div>
              </template>
              <div v-else class="palace-empty">{{ branch }}</div>
            </div>
          </div>
        </div>

      </div>

      <!-- Legend -->
      <div class="legend-bar">
        <div class="legend-title">星曜亮度</div>
        <div class="legend-items">
          <div v-for="(key, idx) in ['庙','旺','得','利','平','不','陷']" :key="idx" class="legend-item">
            <span
              class="legend-swatch"
              :style="{ background: starMeta(key).bg }"
            ></span>
            <span>{{ key }} — {{ ['最旺','旺盛','得地','有利','中平','不得','陷失'][idx] }}</span>
          </div>
        </div>
        <div class="legend-divider"></div>
        <div class="legend-items">
          <div class="legend-item">
            <span class="legend-swatch sihua-swatch">四化</span>
            <span>化禄·化权·化科·化忌</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Empty state -->
    <div v-else class="empty-state">
      <div class="empty-icon">
        <svg width="64" height="64" viewBox="0 0 64 64" fill="none">
          <circle cx="32" cy="32" r="30" stroke="currentColor" stroke-width="1" stroke-dasharray="3 4" />
          <circle
            cx="32"
            cy="32"
            r="20"
            stroke="currentColor"
            stroke-width="0.5"
            stroke-dasharray="2 3"
          />
          <circle cx="32" cy="32" r="4" fill="currentColor" opacity="0.5" />
        </svg>
      </div>
      <p class="empty-title">暂无命盘数据</p>
      <p class="empty-sub">请确认八字命盘已正确创建</p>
    </div>
  </div>
</template>

<style scoped>
@reference "tailwindcss";

.ziwei-wrapper {
  @apply w-full max-w-4xl mx-auto;
}

/* ── Section header ── */
.chart-section-header {
  @apply mb-5;
}
.chart-section-title-group {
  display: flex; align-items: center; gap: 0.75rem;
}
.chart-section-symbol {
  font-size: var(--fs-3xl); color: var(--accent);
  text-shadow: 0 0 12px var(--accent-glow);
}

:global(.dark) .chart-section-symbol {
  text-shadow: 0 0 12px rgba(203,213,225,0.3);
}
.chart-section-title {
  font-family: var(--font-serif);
  font-size: var(--fs-lg); font-weight: 700;
  color: var(--text); margin: 0;
  letter-spacing: 2px;
}
.chart-section-desc {
  font-size: var(--fs-xs); color: var(--text-muted);
  margin: 0.125rem 0 0;
}

/* ── Patterns ── */
.patterns-bar {
  @apply flex flex-wrap justify-center gap-3 mb-6;
}

.pattern-badge {
  @apply px-4 py-1.5 text-xs font-bold rounded-full;
  background: linear-gradient(135deg, var(--crimson), #be123c);
  color: var(--destructive-foreground);
  box-shadow: 0 0 12px rgba(251, 113, 133, 0.25);
  letter-spacing: 1px;
}

:global(.dark) .pattern-badge {
  box-shadow: 0 0 12px rgba(251, 113, 133, 0.4);
}

/* ── Chart frame ── */
.chart-outer-frame {
  position: relative;
  border: 1px solid var(--line-strong);
  border-radius: 12px;
  background: var(--surface-1);
  box-shadow:
    0 0 40px rgba(203, 213, 225, 0.08),
    inset 0 1px 0 var(--line-subtle);
  padding: 3px;
}

:global(.dark) .chart-outer-frame {
  background: linear-gradient(145deg, rgba(20, 16, 30, 0.95), rgba(10, 8, 20, 0.98));
}

/* Decorative corners */
.corner {
  position: absolute;
  width: 16px;
  height: 16px;
  border-color: var(--accent);
  border-style: solid;
  opacity: 0.6;
}
.corner-tl {
  top: -1px;
  left: -1px;
  border-width: 2px 0 0 2px;
  border-radius: 12px 0 0 0;
}
.corner-tr {
  top: -1px;
  right: -1px;
  border-width: 2px 2px 0 0;
  border-radius: 0 12px 0 0;
}
.corner-bl {
  bottom: -1px;
  left: -1px;
  border-width: 0 0 2px 2px;
  border-radius: 0 0 0 12px;
}
.corner-br {
  bottom: -1px;
  right: -1px;
  border-width: 0 2px 2px 0;
  border-radius: 0 0 12px 0;
}

.chart-inner {
  background: var(--surface-0);
  border-radius: 10px;
  overflow: hidden;
}

:global(.dark) .chart-inner {
  background: rgba(10, 8, 20, 0.9);
}

/* ── Palace rows & cells ── */
.palace-row {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
}

.palace-row.row2,
.palace-row.row3 {
  grid-template-columns: 1fr 2fr 1fr;
}

.palace-row.row2 .palace-cell:first-child {
  grid-column: 1;
}
.palace-row.row2 .center-cell {
  grid-column: 2;
  grid-row: 1;
}
.palace-row.row2 .palace-cell:last-child {
  grid-column: 3;
}

.palace-cell {
  @apply flex flex-col items-center justify-start p-3 relative min-h-[140px];
  background: var(--surface-2);
  border: 1px solid var(--line-subtle);
  gap: 3px;
  transition: all 0.3s ease;
}

.palace-cell:hover {
  background: var(--surface-3);
  border-color: var(--line-focus);
  box-shadow: 0 0 20px var(--accent-dim);
  z-index: 2;
}

:global(.dark) .palace-cell {
  background: linear-gradient(180deg, rgba(12, 12, 14, 0.85), rgba(6, 6, 8, 0.92));
  border-color: rgba(203, 213, 225, 0.08);
}

:global(.dark) .palace-cell:hover {
  background: linear-gradient(180deg, rgba(20, 20, 24, 0.92), rgba(12, 12, 16, 0.96));
  border-color: var(--text-soft);
  box-shadow: 0 0 20px rgba(203, 213, 225, 0.12);
}

.palace-cell.has-sihua {
  border-top: 2px solid rgba(251, 113, 133, 0.6);
}

:global(.dark) .palace-cell.has-sihua {
  border-top-color: rgba(251, 113, 133, 0.6);
}

/* ── Palace content ── */
.palace-header {
  @apply flex flex-col items-center gap-0 w-full mb-1;
  border-bottom: 1px solid var(--line-subtle);
  padding-bottom: 4px;
}

.palace-name {
  @apply text-xs font-bold leading-tight;
  color: var(--accent);
  letter-spacing: 1px;
}

.palace-branch {
  @apply text-[var(--fs-2xs)] leading-tight;
  color: var(--text-muted);
}

.palace-empty {
  @apply flex items-center justify-center flex-1 text-sm;
  color: var(--text-soft);
}

/* ── Stars ── */
.star-section {
  @apply flex flex-col items-center gap-1 w-full;
}

.main-star {
  @apply inline-flex items-center gap-1 px-2 py-0.5 text-[var(--fs-2xs)] font-bold rounded-sm leading-tight;
  white-space: nowrap;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.15);
}

:global(.dark) .main-star {
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.4);
}

.brightness-dot {
  @apply text-[var(--fs-2xs)] opacity-80 font-normal;
}

.aux-section {
  @apply flex flex-wrap justify-center gap-1 mt-1;
}

.aux-star {
  @apply text-[var(--fs-2xs)] leading-tight px-1 py-px rounded;
  color: var(--text-muted);
  background: var(--accent-dim);
  border: 1px solid var(--line-subtle);
}

:global(.dark) .aux-star {
  background: rgba(203, 213, 225, 0.06);
  border-color: rgba(203, 213, 225, 0.1);
}

.sihua-section {
  @apply flex flex-wrap justify-center gap-1 mt-2;
}

.sihua-tag {
  @apply rounded-full px-2 py-px text-[var(--fs-2xs)] font-bold leading-tight;
  background: linear-gradient(135deg, var(--crimson), #be123c);
  color: var(--destructive-foreground);
  box-shadow: 0 0 6px rgba(251, 113, 133, 0.3);
}

:global(.dark) .sihua-tag {
  box-shadow: 0 0 6px rgba(251, 113, 133, 0.5);
}

/* ── Center cell ── */
.center-cell {
  @apply relative flex items-center justify-center p-4;
  background: var(--surface-1);
  border-left: 1px solid var(--line-strong);
  border-right: 1px solid var(--line-strong);
  grid-row: span 1;
}

:global(.dark) .center-cell {
  background: linear-gradient(145deg, rgba(6, 6, 8, 0.95), rgba(3, 4, 4, 0.98));
  border-left-color: rgba(203, 213, 225, 0.15);
  border-right-color: rgba(203, 213, 225, 0.15);
}

.center-glow {
  position: absolute;
  inset: 0;
  background: radial-gradient(ellipse at center, var(--accent-dim), transparent 70%);
  pointer-events: none;
}

:global(.dark) .center-glow {
  background: radial-gradient(ellipse at center, rgba(203, 213, 225, 0.08), transparent 70%);
}

.center-inner {
  @apply relative flex flex-col items-center gap-2 z-10;
}

.center-ornament {
  @apply text-xs;
  color: var(--accent);
  opacity: 0.5;
  animation: pulse-glow 3s ease-in-out infinite;
}

@keyframes pulse-glow {
  0%,
  100% {
    opacity: 0.3;
    text-shadow: none;
  }
  50% {
    opacity: 0.7;
    text-shadow: 0 0 8px var(--accent);
  }
}

.center-title {
  @apply text-sm font-bold tracking-widest;
  color: var(--accent);
  letter-spacing: 4px;
  text-shadow: 0 0 20px var(--accent-glow);
}

:global(.dark) .center-title {
  text-shadow: 0 0 20px rgba(203, 213, 225, 0.4);
}

.center-divider {
  @apply w-16 h-px;
  background: linear-gradient(90deg, transparent, var(--line-focus), transparent);
}

:global(.dark) .center-divider {
  background: linear-gradient(90deg, transparent, rgba(203, 213, 225, 0.4), transparent);
}

.center-item {
  @apply flex flex-col items-center gap-0;
}

.center-key {
  @apply text-[var(--fs-2xs)] leading-tight;
  color: var(--text-muted);
}

.center-val {
  @apply text-sm font-bold leading-tight;
}

.mingzhu,
.shenzhu {
  color: var(--crimson);
  text-shadow: 0 0 12px rgba(251, 113, 133, 0.3);
}

:global(.dark) .mingzhu,
:global(.dark) .shenzhu {
  text-shadow: 0 0 12px rgba(251, 113, 133, 0.5);
}

.wuxing {
  color: var(--accent);
  text-shadow: 0 0 12px var(--accent-glow);
}

:global(.dark) .wuxing {
  text-shadow: 0 0 12px rgba(203, 213, 225, 0.4);
}

/* Mid center (row 3) */
.center-cell-mid {
  background: var(--surface-2);
}

:global(.dark) .center-cell-mid {
  background: linear-gradient(145deg, rgba(15, 12, 25, 0.98), rgba(20, 16, 30, 0.95));
}

.sky-pointer {
  animation: rotate-slow 20s linear infinite;
}
@keyframes rotate-slow {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.sky-text {
  @apply text-xs font-bold tracking-widest mt-1;
  color: var(--text-muted);
}

/* ── Legend ── */
.legend-bar {
  @apply flex flex-wrap items-center justify-center gap-6 mt-4 pt-4 px-4;
  border-top: 1px solid var(--line-subtle);
}

.legend-title {
  @apply text-xs font-bold tracking-widest mr-1;
  color: var(--text-muted);
}

.legend-items {
  @apply flex flex-wrap items-center gap-4;
}

.legend-item {
  @apply flex items-center gap-2 text-[var(--fs-2xs)];
  color: var(--text-muted);
}

.legend-swatch {
  @apply inline-block w-4 h-4 rounded-sm;
}

.sihua-swatch {
  background: linear-gradient(135deg, var(--crimson), #be123c);
  box-shadow: 0 0 6px rgba(251, 113, 133, 0.3);
}

:global(.dark) .sihua-swatch {
  box-shadow: 0 0 6px rgba(251, 113, 133, 0.5);
}

.legend-divider {
  @apply w-px h-4;
  background: var(--line-strong);
}

/* ── Empty state ── */
.empty-state {
  @apply flex flex-col items-center justify-center py-16 gap-4;
}

.empty-icon {
  color: var(--icon-muted);
  opacity: 0.3;
  animation: spin-slow 30s linear infinite;
}

@keyframes spin-slow {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.empty-title {
  @apply text-lg font-bold;
  color: var(--text-muted);
}

.empty-sub {
  @apply text-sm;
  color: var(--text-soft);
}
</style>
