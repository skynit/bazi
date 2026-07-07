<script setup lang="ts">
/**
 * BestWorstChip — small badge marking either the best or worst day in a window.
 */
interface Props {
  variant: 'best' | 'worst'
  date?: string
  score?: number
}
defineProps<Props>()
</script>

<template>
  <span class="chip" :class="variant">
    <span class="dot" aria-hidden="true"></span>
    <span class="label">{{ variant === 'best' ? '吉峰' : '低谷' }}</span>
    <span v-if="date" class="date">{{ date.slice(5) }}</span>
    <span v-if="typeof score === 'number'" class="score tabular-nums">{{ score }}</span>
  </span>
</template>

<style scoped>
.chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 10px 4px 8px;
  border-radius: 999px;
  font-size: var(--fs-xs);
  letter-spacing: 0.06em;
  border: 1px solid var(--line-subtle);
  background: var(--glass-bg);
  backdrop-filter: blur(8px);
  color: var(--text);
}

.chip .dot {
  width: 6px; height: 6px; border-radius: 50%;
}

.chip.best {
  border-color: rgba(var(--jade-accent-rgb), 0.45);
  color: rgba(var(--jade-accent-rgb), 1);
  background: rgba(var(--jade-accent-rgb), 0.10);
}
.chip.best .dot { background: var(--jade-accent); box-shadow: 0 0 8px var(--jade-accent); }

.chip.worst {
  border-color: rgba(232, 64, 87, 0.42);
  color: var(--crimson);
  background: rgba(232, 64, 87, 0.08);
}
.chip.worst .dot { background: var(--crimson); box-shadow: 0 0 8px rgba(232, 64, 87, 0.6); }

.score { font-weight: 700; opacity: 0.9; }
.date { color: var(--text-muted); }
.tabular-nums { font-variant-numeric: tabular-nums; }
</style>
