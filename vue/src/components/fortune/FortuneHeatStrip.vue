<script setup lang="ts">
/**
 * FortuneHeatStrip — horizontal score-heatmap of N consecutive days.
 * Each cell color comes from the same oklch ramp as ScoreOrb so the visual
 * language stays consistent.
 */
import { computed } from 'vue'

interface Day {
  date: string
  score: number
  dayPillar?: string
  isBest?: boolean
  isWorst?: boolean
}
interface Props {
  days: Day[]
  /** Optional weekday labels (length must match days). */
  weekdayLabels?: string[]
}
const props = defineProps<Props>()

function tone(score: number): string {
  const t = Math.max(0, Math.min(1, score / 100))
  const L = 0.35 + t * 0.40
  const C = 0.05 + t * 0.13
  const h = 155 - t * 5
  return `oklch(${L.toFixed(3)} ${C.toFixed(3)} ${h.toFixed(1)})`
}

const cells = computed(() =>
  props.days.map((d, i) => ({
    ...d,
    color: tone(d.score),
    label: props.weekdayLabels?.[i] ?? d.date.slice(5)
  }))
)
</script>

<template>
  <div class="strip" role="list">
    <div
      v-for="cell in cells"
      :key="cell.date"
      role="listitem"
      class="cell"
      :class="{ best: cell.isBest, worst: cell.isWorst }"
      :title="`${cell.date} · ${cell.score} 分${cell.dayPillar ? ' · ' + cell.dayPillar : ''}`"
    >
      <div class="bar" :style="{ background: cell.color }"></div>
      <div class="meta">
        <span class="day">{{ cell.label }}</span>
        <span class="score tabular-nums">{{ cell.score }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.strip {
  display: grid;
  grid-template-columns: repeat(v-bind('days.length'), minmax(0, 1fr));
  gap: 6px;
  width: 100%;
}

.cell {
  display: flex;
  flex-direction: column;
  align-items: stretch;
  gap: 6px;
  padding: 8px 6px;
  border-radius: 12px;
  background: var(--glass-bg);
  border: 1px solid var(--line-subtle);
  transition: transform 0.25s cubic-bezier(0.16, 1, 0.3, 1), border-color 0.25s ease;
}
.cell:hover {
  transform: translateY(-2px);
  border-color: rgba(var(--jade-accent-rgb), 0.40);
}
.cell.best {
  border-color: rgba(var(--jade-accent-rgb), 0.65);
  box-shadow: 0 0 0 1px rgba(var(--jade-accent-rgb), 0.30) inset, 0 6px 18px rgba(var(--jade-accent-rgb), 0.18);
}
.cell.worst {
  border-color: rgba(232, 64, 87, 0.55);
}

.bar {
  height: 56px;
  border-radius: 10px;
  filter: drop-shadow(0 4px 14px color-mix(in oklab, currentColor 30%, transparent));
}

.meta {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  font-size: 0.7rem;
  letter-spacing: 0.04em;
}

.day { color: var(--text-muted); }
.score { color: var(--text); font-weight: 700; }
.tabular-nums { font-variant-numeric: tabular-nums; }
</style>
