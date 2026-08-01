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
  <article
    class="card"
    :class="{ highest: isHighest, lowest: isLowest }"
    :title="`${date} ${dayPillar}`"
  >
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
  gap: 8px;
  padding: 14px 16px;
  border-radius: 16px;
  background: var(--glass-bg);
  border: 1px solid var(--line-subtle);
  backdrop-filter: blur(14px) saturate(140%);
  transition:
    transform 0.25s cubic-bezier(0.16, 1, 0.3, 1),
    border-color 0.25s ease,
    box-shadow 0.25s ease;
  min-width: 0;
}
.card:hover {
  transform: translateY(-2px);
  border-color: rgba(var(--jade-accent-rgb), 0.42);
  box-shadow: 0 14px 36px rgba(0, 0, 0, 0.06);
}
.card.highest {
  border-color: rgba(var(--jade-accent-rgb), 0.6);
  box-shadow: 0 0 0 1px rgba(var(--jade-accent-rgb), 0.25) inset;
}
.card.lowest {
  border-color: var(--line-strong);
  box-shadow: 0 0 0 1px var(--line-subtle) inset;
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
  text-transform: uppercase;
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
  color: var(--accent);
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
}

.tabular-nums {
  font-variant-numeric: tabular-nums;
}
</style>
