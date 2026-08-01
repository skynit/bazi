<script setup lang="ts">
/** Compact view of reproducible daily structural fields. */

interface Props {
  date: string
  dayPillar: string
  relationCount?: number
  tenGod?: string
  isHighest?: boolean
  isLowest?: boolean
  weekday?: string
}
defineProps<Props>()
</script>

<template>
  <article class="card" :class="{ highest: isHighest, lowest: isLowest }" :title="`${date} ${dayPillar}`">
    <header class="head">
      <div class="left">
        <span v-if="weekday" class="weekday">{{ weekday }}</span>
        <span class="date tabular-nums">{{ date.slice(5) }}</span>
      </div>
      <span class="pillar">{{ dayPillar }}</span>
    </header>

    <div class="score-line">
      <div class="index-value">
        <span class="score-num tabular-nums">{{ relationCount ?? '—' }}</span>
        <span class="score-label">条关系</span>
      </div>
      <span v-if="tenGod" class="ten-god">{{ tenGod }}</span>
    </div>
  </article>
</template>

<style scoped>
.card {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 14px 16px;
  border-radius: 12px;
  background: var(--surface-1);
  border: 1px solid var(--line-subtle);
  transition:
    transform 0.2s ease,
    border-color 0.2s ease,
    box-shadow 0.2s ease;
  min-width: 0;
}
@media (prefers-reduced-motion: reduce) {
  .card {
    transition: none;
  }
}
.card:hover {
  transform: translateY(-1px);
  border-color: rgba(var(--jade-accent-rgb), 0.42);
  box-shadow: var(--shadow-md);
}
.card:active {
  transform: translateY(0);
  box-shadow: var(--shadow-sm);
}
.card.highest {
  border-color: rgba(var(--jade-accent-rgb), 0.55);
}
.card.highest::before {
  content: '';
  position: absolute;
  top: -1px;
  left: 16px;
  right: 16px;
  height: 2px;
  background: rgba(var(--jade-accent-rgb), 0.6);
  border-radius: 0 0 2px 2px;
}
.card.lowest {
  border-style: dashed;
  border-color: var(--line-strong);
}

.head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}
.left {
  display: flex;
  align-items: baseline;
  gap: 8px;
}
.weekday {
  font-size: var(--fs-xs);
  letter-spacing: 0.18em;
  color: var(--text-muted);
}
.date {
  font-size: var(--fs-sm);
  font-weight: 700;
  color: var(--text);
  letter-spacing: 0.02em;
}
.pillar {
  font-family: var(--font-serif), serif;
  font-size: var(--fs-sm);
  font-weight: 700;
  color: rgba(var(--jade-accent-rgb), 1);
  letter-spacing: 0.06em;
}

.score-line {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 8px;
}
.index-value {
  display: flex;
  align-items: baseline;
  gap: 8px;
  min-width: 0;
}
.score-num {
  font-family: var(--font-serif), serif;
  font-size: var(--fs-stat);
  font-weight: 800;
  line-height: 1;
  color: var(--text);
}
.score-label {
  font-size: var(--fs-xs);
  color: var(--text-soft);
  white-space: nowrap;
}
.ten-god {
  font-size: var(--fs-xs);
  color: var(--text-muted);
  letter-spacing: 0.06em;
  padding: 2px 8px;
  border-radius: 999px;
  border: 1px solid var(--line-subtle);
  background: var(--surface-2);
}

.tabular-nums {
  font-variant-numeric: tabular-nums;
}
</style>
