<script setup lang="ts">
/** Marks the highest or lowest observed relation count in a period. */
interface Props {
  variant: 'highest' | 'lowest'
  date?: string
  count?: number
}
defineProps<Props>()
</script>

<template>
  <span class="chip" :class="variant">
    <span class="dot" aria-hidden="true"></span>
    <span class="label">{{ variant === 'highest' ? '关系较多' : '关系较少' }}</span>
    <span v-if="date" class="date">{{ date.slice(5) }}</span>
    <span v-if="typeof count === 'number'" class="score tabular-nums">{{ count }} 条</span>
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
  width: 6px;
  height: 6px;
  border-radius: 50%;
}

.chip.highest {
  border-color: rgba(var(--jade-accent-rgb), 0.45);
  color: rgba(var(--jade-accent-rgb), 1);
  background: rgba(var(--jade-accent-rgb), 0.1);
}
.chip.highest .dot {
  background: var(--jade-accent);
  box-shadow: 0 0 8px var(--jade-accent);
}

.chip.lowest {
  border-color: var(--line-strong);
  color: var(--text-muted);
  background: var(--glass-bg);
}
.chip.lowest .dot {
  background: var(--text-soft);
}

.score {
  font-weight: 700;
  opacity: 0.9;
}
.date {
  color: var(--text-muted);
}
.tabular-nums {
  font-variant-numeric: tabular-nums;
}
</style>
