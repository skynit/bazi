<script setup lang="ts">
import type { InterpretationLevel } from '../api/fortune'

const props = defineProps<{ modelValue: InterpretationLevel }>()
const emit = defineEmits<{ 'update:modelValue': [value: InterpretationLevel] }>()

const levels: Array<{ value: InterpretationLevel; label: string; hint: string }> = [
  { value: 'basic', label: '简明', hint: '重点关系' },
  { value: 'advanced', label: '详细', hint: '原因与依据' },
  { value: 'professional', label: '推算', hint: '计算过程' },
]
</script>

<template>
  <div class="level-switch" role="radiogroup" aria-label="解读层级">
    <button
      v-for="item in levels"
      :key="item.value"
      type="button"
      role="radio"
      class="level-option"
      :class="{ active: props.modelValue === item.value }"
      :aria-checked="props.modelValue === item.value"
      :data-level="item.value"
      @click="emit('update:modelValue', item.value)"
    >
      <strong>{{ item.label }}</strong>
      <span>{{ item.hint }}</span>
    </button>
  </div>
</template>

<style scoped>
.level-switch {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.25rem;
  padding: var(--space-xs);
  border: 1px solid var(--line-strong);
  border-radius: var(--radius-md);
  background: color-mix(in oklab, var(--accent) 3%, var(--surface-1));
}

.level-option {
  display: flex;
  align-items: baseline;
  justify-content: center;
  gap: 0.4rem;
  min-height: 44px;
  padding: var(--space-sm) 0.75rem;
  border: 1px solid transparent;
  border-radius: var(--radius-sm);
  color: var(--text-dim);
  background: transparent;
  cursor: pointer;
  transition:
    color 160ms ease-out,
    border-color 160ms ease-out,
    background-color 160ms ease-out,
    box-shadow 160ms ease-out;
}

.level-option strong {
  color: inherit;
  font-size: var(--fs-sm, 0.88rem);
}

.level-option span {
  font-size: var(--fs-2xs, 0.68rem);
  color: var(--text-soft);
}

.level-option:hover {
  color: var(--text);
  background: var(--surface-1);
}

.level-option.active {
  color: var(--text);
  border-color: var(--line-focus);
  background: color-mix(in oklab, var(--accent) 7%, var(--surface-0));
  box-shadow: var(--shadow-xs);
}

.level-option.active strong {
  color: color-mix(in oklab, var(--accent) 48%, var(--text));
}

.level-option:focus-visible {
  z-index: 1;
  outline: 2px solid var(--accent);
  outline-offset: 1px;
}

@media (max-width: 560px) {
  .level-option {
    flex-direction: column;
    align-items: center;
    gap: 0.05rem;
  }
}
</style>
