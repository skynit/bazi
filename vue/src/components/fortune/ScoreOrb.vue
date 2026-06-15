<script setup lang="ts">
/**
 * ScoreOrb — large central score with aurora halo and progress ring.
 * Score is 0-100. Color shifts from cool jade-low to warm jade-high.
 */
import { computed } from 'vue'

interface Props {
  score: number
  label?: string
  caption?: string
}
const props = withDefaults(defineProps<Props>(), {
  label: '综合评分',
  caption: ''
})

const clamped = computed(() => Math.max(0, Math.min(100, props.score)))
const progress = computed(() => clamped.value / 100)
const ringDash = computed(() => `${progress.value * 283} 283`) // 2πr ≈ 283 for r=45
const tone = computed(() => {
  const t = progress.value
  // Lightness eases from 0.42 → 0.78; chroma scales with score
  const L = 0.42 + t * 0.36
  const C = 0.05 + t * 0.13
  const h = 155 - t * 5
  return `oklch(${L.toFixed(3)} ${C.toFixed(3)} ${h.toFixed(1)})`
})
</script>

<template>
  <div class="orb">
    <div class="halo" aria-hidden="true"></div>
    <svg class="ring" viewBox="0 0 100 100" aria-hidden="true">
      <circle class="ring-track" cx="50" cy="50" r="45" />
      <circle
        class="ring-fill"
        cx="50"
        cy="50"
        r="45"
        :stroke-dasharray="ringDash"
        :style="{ stroke: tone }"
      />
    </svg>
    <div class="content">
      <span class="number tabular-nums" :style="{ color: tone }">{{ clamped }}</span>
      <span class="unit">分</span>
      <span class="label">{{ label }}</span>
      <span v-if="caption" class="caption">{{ caption }}</span>
    </div>
  </div>
</template>

<style scoped>
.orb {
  position: relative;
  width: 220px;
  height: 220px;
  display: grid;
  place-items: center;
}

.halo {
  position: absolute;
  inset: -30%;
  background:
    radial-gradient(closest-side, rgba(var(--jade-accent-rgb), 0.32), transparent 70%);
  filter: blur(20px);
  pointer-events: none;
  animation: pulse 5s ease-in-out infinite alternate;
}

.ring {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  transform: rotate(-90deg);
}

.ring-track {
  fill: none;
  stroke: var(--line-strong);
  stroke-width: 2.4;
  opacity: 0.55;
}

.ring-fill {
  fill: none;
  stroke-width: 3.2;
  stroke-linecap: round;
  filter: drop-shadow(0 0 8px rgba(var(--jade-accent-rgb), 0.55));
  transition: stroke-dasharray 0.8s cubic-bezier(0.16, 1, 0.3, 1);
}

.content {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
}

.number {
  font-family: var(--font-serif), 'Songti SC', serif;
  font-size: 4rem;
  font-weight: 800;
  line-height: 1;
  letter-spacing: 0.02em;
  text-shadow: 0 0 24px rgba(var(--jade-accent-rgb), 0.45);
}

.unit {
  font-size: 0.78rem;
  color: var(--text-soft);
  letter-spacing: 0.2em;
  margin-top: -2px;
}

.label {
  margin-top: 6px;
  font-size: 0.72rem;
  letter-spacing: 0.32em;
  color: var(--text-muted);
  text-transform: uppercase;
}

.caption {
  margin-top: 2px;
  font-size: 0.7rem;
  color: var(--text-soft);
  letter-spacing: 0.06em;
}

@keyframes pulse {
  from { opacity: 0.6; transform: scale(1); }
  to   { opacity: 1;   transform: scale(1.08); }
}

@media (prefers-reduced-motion: reduce) {
  .halo { animation: none; }
  .ring-fill { transition: none; }
}

.tabular-nums { font-variant-numeric: tabular-nums; }
</style>
