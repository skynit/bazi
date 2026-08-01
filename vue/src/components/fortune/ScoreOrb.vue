<script setup lang="ts">
interface Props {
  value: number
  label?: string
  caption?: string
  unit?: string
}

withDefaults(defineProps<Props>(), {
  label: '平均命中关系',
  caption: '',
  unit: '条/日',
})
</script>

<template>
  <div class="metric">
    <div class="content">
      <span class="number tabular-nums">{{ value.toFixed(1) }}</span>
      <span class="unit">{{ unit }}</span>
      <span class="label">{{ label }}</span>
      <span v-if="caption" class="caption">{{ caption }}</span>
    </div>
  </div>
</template>

<style scoped>
.metric {
  width: min(220px, 100%);
  min-height: 150px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
  border: 1px solid var(--line-strong);
  border-radius: 8px;
  background: color-mix(in oklab, var(--surface-1) 78%, transparent);
}

.content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 3px;
  text-align: center;
}

.number {
  font-family: var(--font-serif), 'Songti SC', serif;
  font-size: var(--fs-hero-strong);
  font-weight: 800;
  line-height: 1;
  color: var(--accent);
}

.unit,
.caption {
  color: var(--text-soft);
  font-size: var(--fs-xs);
}

.label {
  margin-top: 8px;
  color: var(--text-muted);
  font-size: var(--fs-xs);
  font-weight: 700;
}

.caption {
  margin-top: 2px;
}

.tabular-nums {
  font-variant-numeric: tabular-nums;
}
</style>
