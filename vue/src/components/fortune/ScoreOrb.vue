<script setup lang="ts">
import { computed, ref } from 'vue'
import { useCountUp } from '../../composables/useCountUp'

interface Props {
  value: number
  label?: string
  caption?: string
  unit?: string
}

const props = withDefaults(defineProps<Props>(), {
  label: '平均命中关系',
  caption: '',
  unit: '条/日',
})

const host = ref<HTMLElement | null>(null)
const target = computed(() => props.value)
const { display } = useCountUp(target, host, { duration: 620, decimals: 1 })
</script>

<template>
  <div ref="host" class="metric">
    <div class="content">
      <div class="figure">
        <span class="number tabular-nums">{{ display }}</span>
        <span class="unit">{{ unit }}</span>
      </div>
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
  padding: 24px 20px;
  border: 1px solid var(--line-strong);
  border-radius: 12px;
  background: var(--surface-1);
  box-shadow: var(--shadow-sm);
  position: relative;
}

.metric::before {
  content: '';
  position: absolute;
  top: 0;
  left: 20px;
  right: 20px;
  height: 2px;
  background: rgba(var(--jade-accent-rgb), 0.55);
}

.content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  text-align: center;
}

.figure {
  display: flex;
  align-items: baseline;
  gap: 6px;
}

.number {
  font-family: var(--font-serif), 'Songti SC', serif;
  font-size: var(--fs-hero-strong);
  font-weight: 800;
  line-height: 1;
  color: var(--text);
  min-width: 2.2em;
  text-align: right;
}

.unit {
  color: var(--text-soft);
  font-size: var(--fs-xs);
}

.label {
  margin-top: 8px;
  color: var(--text-muted);
  font-size: var(--fs-xs);
  font-weight: 700;
  letter-spacing: 0.08em;
}

.caption {
  color: var(--text-soft);
  font-size: var(--fs-xs);
}

.tabular-nums {
  font-variant-numeric: tabular-nums;
}
</style>
