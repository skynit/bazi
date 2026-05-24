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

// Traditional star brightness → color (庙=most auspicious ... 陷=weakest)
const brightnessMeta: Record<string, { bg: string; text: string; label: string }> = {
  庙: { bg: 'linear-gradient(135deg,#C41E3A,#8B0000)', text: '#fff', label: '庙' },
  旺: { bg: 'linear-gradient(135deg,#FF8C00,#CC5500)', text: '#fff', label: '旺' },
  得: { bg: 'linear-gradient(135deg,#DAA520,#B8860B)', text: '#fff', label: '得' },
  利: { bg: 'linear-gradient(135deg,#228B22,#006400)', text: '#fff', label: '利' },
  平: { bg: 'linear-gradient(135deg,#808080,#696969)', text: '#fff', label: '平' },
  不: { bg: 'linear-gradient(135deg,#5F9EA0,#4682B4)', text: '#fff', label: '不' },
  陷: { bg: 'linear-gradient(135deg,#2B3A42,#1a252e)', text: '#aaa', label: '陷' },
}

function starMeta(brightness: string) {
  return brightnessMeta[brightness] || brightnessMeta['陷']
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
              :class="{ 'has-sihua': palaceAt(branch)?.sihua?.length }"
            >
              <template v-if="palaceAt(branch)">
                <div class="palace-header">
                  <span class="palace-name">{{ palaceAt(branch)!.name }}</span>
                  <span class="palace-branch">{{ branch }}</span>
                </div>
                <div class="star-section">
                  <span
                    v-for="(star, si) in palaceAt(branch)?.mainStars || []"
                    :key="'ms-' + si"
                    class="main-star"
                    :style="{
                      background: starMeta(star.brightness).bg,
                      color: starMeta(star.brightness).text,
                    }"
                    :title="star.brightness"
                  >
                    {{ star.name
                    }}<span class="brightness-dot">{{ starMeta(star.brightness).label }}</span>
                  </span>
                </div>
                <div v-if="(palaceAt(branch)?.auxStars || []).length" class="aux-section">
                  <span
                    v-for="(star, si) in palaceAt(branch)?.auxStars || []"
                    :key="'as-' + si"
                    class="aux-star"
                    >{{ star.name }}</span
                  >
                </div>
                <div v-if="palaceAt(branch)!.sihua?.length" class="sihua-section">
                  <span
                    v-for="(sh, si) in palaceAt(branch)!.sihua"
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
            <div class="palace-cell" :class="{ 'has-sihua': palaceAt('辰')?.sihua?.length }">
              <template v-if="palaceAt('辰')">
                <div class="palace-header">
                  <span class="palace-name">{{ palaceAt('辰')!.name }}</span>
                  <span class="palace-branch">辰</span>
                </div>
                <div class="star-section">
                  <span
                    v-for="(star, si) in palaceAt('辰')?.mainStars || []"
                    :key="'ms-' + si"
                    class="main-star"
                    :style="{
                      background: starMeta(star.brightness).bg,
                      color: starMeta(star.brightness).text,
                    }"
                    >{{ star.name
                    }}<span class="brightness-dot">{{
                      starMeta(star.brightness).label
                    }}</span></span
                  >
                </div>
                <div v-if="(palaceAt('辰')?.auxStars || []).length" class="aux-section">
                  <span
                    v-for="(star, si) in palaceAt('辰')?.auxStars || []"
                    :key="'as-' + si"
                    class="aux-star"
                    >{{ star.name }}</span
                  >
                </div>
                <div v-if="palaceAt('辰')!.sihua?.length" class="sihua-section">
                  <span
                    v-for="(sh, si) in palaceAt('辰')!.sihua"
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
                  <span class="center-val mingzhu">{{ mingZhu }}</span>
                </div>
                <div class="center-item">
                  <span class="center-key">身主</span>
                  <span class="center-val shenzhu">{{ shenZhu }}</span>
                </div>
                <div class="center-item">
                  <span class="center-key">五行局</span>
                  <span class="center-val wuxing">{{ wuxingJu }}</span>
                </div>
                <div class="center-divider"></div>
                <div class="center-ornament">✦</div>
              </div>
            </div>

            <div class="palace-cell" :class="{ 'has-sihua': palaceAt('酉')?.sihua?.length }">
              <template v-if="palaceAt('酉')">
                <div class="palace-header">
                  <span class="palace-name">{{ palaceAt('酉')!.name }}</span>
                  <span class="palace-branch">酉</span>
                </div>
                <div class="star-section">
                  <span
                    v-for="(star, si) in palaceAt('酉')?.mainStars || []"
                    :key="'ms-' + si"
                    class="main-star"
                    :style="{
                      background: starMeta(star.brightness).bg,
                      color: starMeta(star.brightness).text,
                    }"
                    >{{ star.name
                    }}<span class="brightness-dot">{{
                      starMeta(star.brightness).label
                    }}</span></span
                  >
                </div>
                <div v-if="(palaceAt('酉')?.auxStars || []).length" class="aux-section">
                  <span
                    v-for="(star, si) in palaceAt('酉')?.auxStars || []"
                    :key="'as-' + si"
                    class="aux-star"
                    >{{ star.name }}</span
                  >
                </div>
                <div v-if="palaceAt('酉')!.sihua?.length" class="sihua-section">
                  <span
                    v-for="(sh, si) in palaceAt('酉')!.sihua"
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
            <div class="palace-cell" :class="{ 'has-sihua': palaceAt('卯')?.sihua?.length }">
              <template v-if="palaceAt('卯')">
                <div class="palace-header">
                  <span class="palace-name">{{ palaceAt('卯')!.name }}</span>
                  <span class="palace-branch">卯</span>
                </div>
                <div class="star-section">
                  <span
                    v-for="(star, si) in palaceAt('卯')?.mainStars || []"
                    :key="'ms-' + si"
                    class="main-star"
                    :style="{
                      background: starMeta(star.brightness).bg,
                      color: starMeta(star.brightness).text,
                    }"
                    >{{ star.name
                    }}<span class="brightness-dot">{{
                      starMeta(star.brightness).label
                    }}</span></span
                  >
                </div>
                <div v-if="(palaceAt('卯')?.auxStars || []).length" class="aux-section">
                  <span
                    v-for="(star, si) in palaceAt('卯')?.auxStars || []"
                    :key="'as-' + si"
                    class="aux-star"
                    >{{ star.name }}</span
                  >
                </div>
                <div v-if="palaceAt('卯')!.sihua?.length" class="sihua-section">
                  <span
                    v-for="(sh, si) in palaceAt('卯')!.sihua"
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
                      stroke="#D4A84B"
                      stroke-width="1"
                      stroke-dasharray="2 3"
                    />
                    <circle cx="20" cy="20" r="6" fill="#D4A84B" opacity="0.3" />
                    <circle cx="20" cy="20" r="3" fill="#D4A84B" />
                  </svg>
                </div>
                <div class="sky-text">天宫图</div>
              </div>
            </div>

            <div class="palace-cell" :class="{ 'has-sihua': palaceAt('戌')?.sihua?.length }">
              <template v-if="palaceAt('戌')">
                <div class="palace-header">
                  <span class="palace-name">{{ palaceAt('戌')!.name }}</span>
                  <span class="palace-branch">戌</span>
                </div>
                <div class="star-section">
                  <span
                    v-for="(star, si) in palaceAt('戌')?.mainStars || []"
                    :key="'ms-' + si"
                    class="main-star"
                    :style="{
                      background: starMeta(star.brightness).bg,
                      color: starMeta(star.brightness).text,
                    }"
                    >{{ star.name
                    }}<span class="brightness-dot">{{
                      starMeta(star.brightness).label
                    }}</span></span
                  >
                </div>
                <div v-if="(palaceAt('戌')?.auxStars || []).length" class="aux-section">
                  <span
                    v-for="(star, si) in palaceAt('戌')?.auxStars || []"
                    :key="'as-' + si"
                    class="aux-star"
                    >{{ star.name }}</span
                  >
                </div>
                <div v-if="palaceAt('戌')!.sihua?.length" class="sihua-section">
                  <span
                    v-for="(sh, si) in palaceAt('戌')!.sihua"
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
              :class="{ 'has-sihua': palaceAt(branch)?.sihua?.length }"
            >
              <template v-if="palaceAt(branch)">
                <div class="palace-header">
                  <span class="palace-name">{{ palaceAt(branch)!.name }}</span>
                  <span class="palace-branch">{{ branch }}</span>
                </div>
                <div class="star-section">
                  <span
                    v-for="(star, si) in palaceAt(branch)?.mainStars || []"
                    :key="'ms-' + si"
                    class="main-star"
                    :style="{
                      background: starMeta(star.brightness).bg,
                      color: starMeta(star.brightness).text,
                    }"
                    >{{ star.name
                    }}<span class="brightness-dot">{{
                      starMeta(star.brightness).label
                    }}</span></span
                  >
                </div>
                <div v-if="(palaceAt(branch)?.auxStars || []).length" class="aux-section">
                  <span
                    v-for="(star, si) in palaceAt(branch)?.auxStars || []"
                    :key="'as-' + si"
                    class="aux-star"
                    >{{ star.name }}</span
                  >
                </div>
                <div v-if="palaceAt(branch)!.sihua?.length" class="sihua-section">
                  <span
                    v-for="(sh, si) in palaceAt(branch)!.sihua"
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
          <div class="legend-item">
            <span
              class="legend-swatch"
              style="background: linear-gradient(135deg, #c41e3a, #8b0000)"
            ></span>
            <span>庙 — 最旺</span>
          </div>
          <div class="legend-item">
            <span
              class="legend-swatch"
              style="background: linear-gradient(135deg, #ff8c00, #cc5500)"
            ></span>
            <span>旺 — 旺盛</span>
          </div>
          <div class="legend-item">
            <span
              class="legend-swatch"
              style="background: linear-gradient(135deg, #daa520, #b8860b)"
            ></span>
            <span>得 — 得地</span>
          </div>
          <div class="legend-item">
            <span
              class="legend-swatch"
              style="background: linear-gradient(135deg, #228b22, #006400)"
            ></span>
            <span>利 — 有利</span>
          </div>
          <div class="legend-item">
            <span
              class="legend-swatch"
              style="background: linear-gradient(135deg, #808080, #696969)"
            ></span>
            <span>平 — 中平</span>
          </div>
          <div class="legend-item">
            <span
              class="legend-swatch"
              style="background: linear-gradient(135deg, #5f9ea0, #4682b4)"
            ></span>
            <span>不 — 不得</span>
          </div>
          <div class="legend-item">
            <span
              class="legend-swatch"
              style="background: linear-gradient(135deg, #2b3a42, #1a252e)"
            ></span>
            <span>陷 — 陷失</span>
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
          <circle cx="32" cy="32" r="30" stroke="#D4A84B" stroke-width="1" stroke-dasharray="3 4" />
          <circle
            cx="32"
            cy="32"
            r="20"
            stroke="#D4A84B"
            stroke-width="0.5"
            stroke-dasharray="2 3"
          />
          <circle cx="32" cy="32" r="4" fill="#D4A84B" opacity="0.5" />
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
  font-size: 1.5rem; color: var(--gold);
  text-shadow: 0 0 12px rgba(212,168,75,0.3);
}
.chart-section-title {
  font-family: var(--font-serif);
  font-size: 1.05rem; font-weight: 700;
  color: var(--text); margin: 0;
  letter-spacing: 2px;
}
.chart-section-desc {
  font-size: 0.7rem; color: var(--muted);
  margin: 0.125rem 0 0;
}

/* ── Patterns ── */
.patterns-bar {
  @apply flex flex-wrap justify-center gap-3 mb-6;
}

.pattern-badge {
  @apply px-4 py-1.5 text-xs font-bold rounded-full;
  background: linear-gradient(135deg, #c41e3a, #8b0000);
  color: #fff;
  box-shadow: 0 0 12px rgba(196, 30, 58, 0.4);
  letter-spacing: 1px;
}

/* ── Chart frame ── */
.chart-outer-frame {
  position: relative;
  border: 1px solid rgba(212, 168, 75, 0.3);
  border-radius: 12px;
  background: linear-gradient(145deg, rgba(20, 16, 30, 0.95), rgba(10, 8, 20, 0.98));
  box-shadow:
    0 0 40px rgba(212, 168, 75, 0.08),
    inset 0 1px 0 rgba(212, 168, 75, 0.15);
  padding: 3px;
}

/* Decorative corners */
.corner {
  position: absolute;
  width: 16px;
  height: 16px;
  border-color: #d4a84b;
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
  background: rgba(10, 8, 20, 0.9);
  border-radius: 10px;
  overflow: hidden;
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
  background: linear-gradient(180deg, rgba(43, 37, 24, 0.8), rgba(30, 25, 15, 0.9));
  border: 1px solid rgba(212, 168, 75, 0.08);
  gap: 3px;
  transition: all 0.3s ease;
}

.palace-cell:hover {
  background: linear-gradient(180deg, rgba(60, 50, 30, 0.9), rgba(40, 32, 18, 0.95));
  border-color: rgba(212, 168, 75, 0.3);
  box-shadow: 0 0 20px rgba(212, 168, 75, 0.12);
  z-index: 2;
}

.palace-cell.has-sihua {
  border-top: 2px solid rgba(196, 30, 58, 0.6);
}

/* ── Palace content ── */
.palace-header {
  @apply flex flex-col items-center gap-0 w-full mb-1;
  border-bottom: 1px solid rgba(212, 168, 75, 0.1);
  padding-bottom: 4px;
}

.palace-name {
  @apply text-xs font-bold leading-tight;
  color: #d4a84b;
  letter-spacing: 1px;
}

.palace-branch {
  @apply text-[10px] leading-tight;
  color: rgba(212, 168, 75, 0.5);
}

.palace-empty {
  @apply flex items-center justify-center flex-1 text-sm;
  color: rgba(212, 168, 75, 0.2);
}

/* ── Stars ── */
.star-section {
  @apply flex flex-col items-center gap-1 w-full;
}

.main-star {
  @apply inline-flex items-center gap-1 px-2 py-0.5 text-[11px] font-bold rounded-sm leading-tight;
  white-space: nowrap;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.4);
}

.brightness-dot {
  @apply text-[9px] opacity-80 font-normal;
}

.aux-section {
  @apply flex flex-wrap justify-center gap-1 mt-1;
}

.aux-star {
  @apply text-[10px] leading-tight px-1 py-px rounded;
  color: rgba(180, 160, 120, 0.8);
  background: rgba(212, 168, 75, 0.06);
  border: 1px solid rgba(212, 168, 75, 0.1);
}

.sihua-section {
  @apply flex flex-wrap justify-center gap-1 mt-2;
}

.sihua-tag {
  @apply rounded-full px-2 py-px text-[9px] font-bold leading-tight;
  background: linear-gradient(135deg, #c41e3a, #8b0000);
  color: #fff;
  box-shadow: 0 0 6px rgba(196, 30, 58, 0.5);
}

/* ── Center cell ── */
.center-cell {
  @apply relative flex items-center justify-center p-4;
  background: linear-gradient(145deg, rgba(30, 25, 15, 0.95), rgba(20, 16, 28, 0.98));
  border-left: 1px solid rgba(212, 168, 75, 0.15);
  border-right: 1px solid rgba(212, 168, 75, 0.15);
  grid-row: span 1;
}

.center-glow {
  position: absolute;
  inset: 0;
  background: radial-gradient(ellipse at center, rgba(212, 168, 75, 0.08), transparent 70%);
  pointer-events: none;
}

.center-inner {
  @apply relative flex flex-col items-center gap-2 z-10;
}

.center-ornament {
  @apply text-xs;
  color: #d4a84b;
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
    text-shadow: 0 0 8px #d4a84b;
  }
}

.center-title {
  @apply text-sm font-bold tracking-widest;
  color: #d4a84b;
  letter-spacing: 4px;
  text-shadow: 0 0 20px rgba(212, 168, 75, 0.4);
}

.center-divider {
  @apply w-16 h-px;
  background: linear-gradient(90deg, transparent, rgba(212, 168, 75, 0.4), transparent);
}

.center-item {
  @apply flex flex-col items-center gap-0;
}

.center-key {
  @apply text-[10px] leading-tight;
  color: rgba(212, 168, 75, 0.5);
}

.center-val {
  @apply text-sm font-bold leading-tight;
}

.mingzhu,
.shenzhu {
  color: #c41e3a;
  text-shadow: 0 0 12px rgba(196, 30, 58, 0.5);
}
.wuxing {
  color: #d4a84b;
  text-shadow: 0 0 12px rgba(212, 168, 75, 0.4);
}

/* Mid center (row 3) */
.center-cell-mid {
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
  color: rgba(212, 168, 75, 0.4);
}

/* ── Legend ── */
.legend-bar {
  @apply flex flex-wrap items-center justify-center gap-6 mt-4 pt-4 px-4;
  border-top: 1px solid rgba(212, 168, 75, 0.1);
}

.legend-title {
  @apply text-xs font-bold tracking-widest mr-1;
  color: rgba(212, 168, 75, 0.5);
}

.legend-items {
  @apply flex flex-wrap items-center gap-4;
}

.legend-item {
  @apply flex items-center gap-2 text-[11px];
  color: rgba(212, 168, 75, 0.6);
}

.legend-swatch {
  @apply inline-block w-4 h-4 rounded-sm;
}

.sihua-swatch {
  background: linear-gradient(135deg, #c41e3a, #8b0000);
  box-shadow: 0 0 6px rgba(196, 30, 58, 0.5);
}

.legend-divider {
  @apply w-px h-4;
  background: rgba(212, 168, 75, 0.15);
}

/* ── Empty state ── */
.empty-state {
  @apply flex flex-col items-center justify-center py-16 gap-4;
}

.empty-icon {
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
  color: rgba(212, 168, 75, 0.5);
}

.empty-sub {
  @apply text-sm;
  color: rgba(212, 168, 75, 0.3);
}
</style>
