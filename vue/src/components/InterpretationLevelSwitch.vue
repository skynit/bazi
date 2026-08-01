<script setup lang="ts">
import { computed } from 'vue'
import type { InterpretationLevel } from '../api/fortune'

const props = defineProps<{ modelValue: InterpretationLevel }>()
const emit = defineEmits<{ 'update:modelValue': [value: InterpretationLevel] }>()

// 三段为递进关系：每一档在上一档基础上继续加深
const levels: Array<{ value: InterpretationLevel; order: string; label: string; hint: string }> = [
  { value: 'basic', order: '壹', label: '简明', hint: '重点关系' },
  { value: 'advanced', order: '贰', label: '详细', hint: '加深 · 原因依据' },
  { value: 'professional', order: '叁', label: '规则明细', hint: '最深 · 计算说明' },
]

const activeIndex = computed(() =>
  Math.max(
    0,
    levels.findIndex((item) => item.value === props.modelValue),
  ),
)

const levelDescriptions: Record<InterpretationLevel, string> = {
  basic: '当前只显示重点关系与白话说明。',
  advanced: '当前在简明基础上增加结构分析、五行口径与中文化依据。',
  professional: '当前在详细基础上增加规则口径和传统日课查表过程。',
}
</script>

<template>
  <div class="level-switch-wrap">
    <div class="level-switch" role="radiogroup" aria-label="解读层级（逐层加深）">
      <span class="level-slider" :class="`pos-${activeIndex}`" aria-hidden="true"></span>
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
        <span class="level-order" aria-hidden="true">{{ item.order }}</span>
        <strong>{{ item.label }}</strong>
        <span class="level-hint">{{ item.hint }}</span>
      </button>
    </div>
    <p class="level-boundary" aria-live="polite">
      {{ levelDescriptions[props.modelValue] }}详细程度变化不代表结论可靠性变化。
    </p>
  </div>
</template>

<style scoped>
.level-switch-wrap {
  display: grid;
}

.level-switch {
  position: relative;
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.25rem;
  padding: var(--space-xs);
  border: 1px solid var(--line-strong);
  border-radius: var(--radius-md);
  background: color-mix(in oklab, var(--accent) 3%, var(--surface-1));
}

/* 滑动选中指示：位置只随 activeIndex 变化，过渡 300ms */
.level-slider {
  position: absolute;
  top: var(--space-xs);
  bottom: var(--space-xs);
  left: var(--space-xs);
  width: calc((100% - 2 * var(--space-xs) - 0.5rem) / 3);
  border: 1px solid var(--line-focus);
  border-radius: var(--radius-sm);
  background: color-mix(in oklab, var(--accent) 7%, var(--surface-0));
  box-shadow: var(--shadow-xs);
  transition: transform 300ms ease;
  pointer-events: none;
}
.level-slider.pos-0 {
  transform: translateX(0);
}
.level-slider.pos-1 {
  transform: translateX(calc(100% + 0.25rem));
}
.level-slider.pos-2 {
  transform: translateX(calc(200% + 0.5rem));
}

.level-option {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: baseline;
  justify-content: center;
  gap: 0.4rem;
  min-height: 44px;
  padding: var(--space-sm) 0.5rem;
  border: 1px solid transparent;
  border-radius: var(--radius-sm);
  color: var(--text-dim);
  background: transparent;
  cursor: pointer;
  transition: color 160ms ease-out;
}

.level-order {
  font-family: var(--font-serif);
  font-size: var(--fs-2xs, 0.68rem);
  color: var(--text-soft);
}

.level-option strong {
  color: inherit;
  font-size: var(--fs-sm, 0.88rem);
}

.level-hint {
  font-size: var(--fs-2xs, 0.68rem);
  color: var(--text-soft);
}

.level-option:hover {
  color: var(--text);
}

.level-option.active {
  color: var(--text);
}

.level-option.active strong {
  color: color-mix(in oklab, var(--accent) 48%, var(--text));
}

.level-option.active .level-order {
  color: color-mix(in oklab, var(--accent) 48%, var(--text));
}

.level-option:focus-visible {
  z-index: 1;
  outline: 2px solid var(--accent);
  outline-offset: 1px;
}

.level-boundary {
  margin: var(--space-xs) 0 0;
  color: var(--text-muted);
  font-size: var(--fs-xs, 0.78rem);
  line-height: 1.6;
}

@media (max-width: 560px) {
  .level-option {
    flex-direction: column;
    align-items: center;
    gap: 0.05rem;
  }
}

@media (prefers-reduced-motion: reduce) {
  .level-slider,
  .level-option {
    transition: none;
  }
}
</style>
