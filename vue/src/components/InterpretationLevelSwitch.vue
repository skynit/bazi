<script setup lang="ts">
import type { InterpretationLevel } from '../api/fortune'

const props = defineProps<{ modelValue: InterpretationLevel }>()
const emit = defineEmits<{ 'update:modelValue': [value: InterpretationLevel] }>()

const levels: Array<{ value: InterpretationLevel; label: string; hint: string }> = [
  { value: 'basic', label: '普通', hint: '结论与行动' },
  { value: 'advanced', label: '进阶', hint: '正反证据' },
  { value: 'professional', label: '专业', hint: '规则与版本' },
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
  gap: 0.35rem;
  padding: 0.32rem;
  border: 1px solid color-mix(in oklab, var(--accent) 18%, transparent);
  border-radius: 12px;
  background: color-mix(in oklab, var(--surface, #111827) 90%, transparent);
}

.level-option {
  display: flex;
  align-items: baseline;
  justify-content: center;
  gap: 0.4rem;
  min-height: 42px;
  padding: 0.45rem 0.65rem;
  border: 0;
  border-radius: 9px;
  color: var(--text-muted);
  background: transparent;
  cursor: pointer;
  transition: 160ms ease;
}

.level-option strong {
  color: inherit;
  font-size: var(--fs-sm, 0.88rem);
}

.level-option span {
  font-size: var(--fs-2xs, 0.68rem);
  opacity: 0.72;
}

.level-option:hover,
.level-option.active {
  color: var(--text);
  background: color-mix(in oklab, var(--accent) 12%, transparent);
  box-shadow: inset 0 0 0 1px color-mix(in oklab, var(--accent) 18%, transparent);
}

@media (max-width: 560px) {
  .level-option {
    flex-direction: column;
    align-items: center;
    gap: 0.05rem;
  }
}
</style>
