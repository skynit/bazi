<script setup lang="ts">
/**
 * FortuneHeatStrip — horizontal relation-count map of N consecutive days.
 * 颜色由 --jade-accent 与 --surface-3 按命中数混合，随主题与五行选择自适应。
 */
import { computed } from 'vue'

interface Day {
  date: string
  relationCount: number
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

const maxCount = computed(() => Math.max(...props.days.map((day) => day.relationCount), 1))

function tone(count: number): string {
  const t = Math.max(0, Math.min(1, count / maxCount.value))
  return `color-mix(in oklab, var(--jade-accent) ${Math.round(24 + t * 64)}%, var(--surface-3))`
}

const cells = computed(() =>
  props.days.map((d, i) => ({
    ...d,
    color: tone(d.relationCount),
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
      :title="`${cell.date} · 命中关系 ${cell.relationCount} 条${cell.dayPillar ? ' · ' + cell.dayPillar : ''}`"
    >
      <div class="bar" :style="{ background: cell.color }"></div>
      <div class="meta">
        <span class="day">{{ cell.label }}</span>
        <span class="score tabular-nums">{{ cell.relationCount }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.strip {
  display: grid;
  grid-template-columns: repeat(v-bind('days.length'), minmax(0, 1fr));
  gap: 8px;
  width: 100%;
}

.cell {
  display: flex;
  flex-direction: column;
  align-items: stretch;
  gap: 8px;
  padding: 10px 10px 8px;
  border-radius: 10px;
  background: var(--surface-1);
  border: 1px solid var(--line-subtle);
  transition:
    transform 0.2s ease,
    border-color 0.2s ease;
}
@media (prefers-reduced-motion: reduce) {
  .cell {
    transition: none;
  }
}
.cell:hover {
  transform: translateY(-1px);
  border-color: rgba(var(--jade-accent-rgb), 0.4);
}
.cell:active {
  transform: translateY(0);
}
.cell.highest {
  border-color: rgba(var(--jade-accent-rgb), 0.55);
}
.cell.highest .day {
  color: rgba(var(--jade-accent-rgb), 1);
  font-weight: 700;
}
.cell.lowest {
  border-style: dashed;
  border-color: var(--line-strong);
}

.bar {
  height: 52px;
  border-radius: 6px;
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
