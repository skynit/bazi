<script setup lang="ts">
/**
 * FortuneHeatStrip — horizontal structural-index map of N consecutive days.
 * Each cell color follows the active wuxing theme so the visual language stays
 * consistent with ScoreOrb and the trend charts.
 */
import { computed } from 'vue'

interface Day {
  date: string
  score: number
  dayPillar?: string
  isHighest?: boolean
  isLowest?: boolean
}
interface Props {
  days: Day[]
  /** Optional weekday labels (length must match days). */
  weekdayLabels?: string[]
}
const props = defineProps<Props>()

function tone(score: number): string {
  const t = Math.max(0, Math.min(1, score / 100))
  return `color-mix(in oklab, var(--jade-accent) ${Math.round(48 + t * 52)}%, var(--surface-2) ${Math.round((1 - t) * 28)}%)`
}

const cells = computed(() =>
  props.days.map((d, i) => ({
    ...d,
    color: tone(d.score),
    label: props.weekdayLabels?.[i] ?? d.date.slice(5),
  })),
)
</script>

<template>
  <div class="strip" role="list">
    <div
      v-for="cell in cells"
      :key="cell.date"
      role="listitem"
      class="cell"
      :class="{ highest: cell.isHighest, lowest: cell.isLowest }"
      :title="`${cell.date} · 结构指数 ${cell.score}${cell.dayPillar ? ' · ' + cell.dayPillar : ''}`"
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
  transition:
    transform 0.25s cubic-bezier(0.16, 1, 0.3, 1),
    border-color 0.25s ease;
}
.cell:hover {
  transform: translateY(-2px);
  border-color: rgba(var(--jade-accent-rgb), 0.4);
}
.cell.highest {
  border-color: rgba(var(--jade-accent-rgb), 0.65);
  box-shadow:
    0 0 0 1px rgba(var(--jade-accent-rgb), 0.3) inset,
    0 6px 18px rgba(var(--jade-accent-rgb), 0.18);
}
.cell.lowest {
  border-color: var(--line-strong);
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
  font-size: var(--fs-xs);
  letter-spacing: 0.04em;
}

.day {
  color: var(--text-muted);
}
.score {
  color: var(--text);
  font-weight: 700;
}
.tabular-nums {
  font-variant-numeric: tabular-nums;
}
</style>
