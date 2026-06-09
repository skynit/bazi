<script setup lang="ts">
import { ref } from 'vue'

const props = withDefaults(defineProps<{
  text?: string
  size?: 'sm' | 'md' | 'lg'
  disabled?: boolean
}>(), {
  text: '',
  size: 'md',
  disabled: false
})

const emit = defineEmits<{
  click: []
}>()

const shinePosition = ref(-100)

function handleMouseMove(e: MouseEvent) {
  if (props.disabled) return
  const rect = (e.currentTarget as HTMLElement).getBoundingClientRect()
  shinePosition.value = e.clientX - rect.left
}

function handleMouseLeave() {
  shinePosition.value = -100
}

const sizeClasses = {
  sm: 'px-4 py-2 text-sm',
  md: 'px-8 py-3 text-base',
  lg: 'px-12 py-4 text-lg'
}
</script>

<template>
  <button
    :class="[
      'shiny-btn',
      sizeClasses[props.size],
      { 'shiny-btn-disabled': props.disabled }
    ]"
    :disabled="props.disabled"
    @click="emit('click')"
    @mousemove="handleMouseMove"
    @mouseleave="handleMouseLeave"
  >
    <span
      class="shiny-reflect"
      :style="{ left: shinePosition + 'px' }"
    ></span>
    <span class="shiny-content">
      <slot>{{ text }}</slot>
    </span>
  </button>
</template>

<style scoped>
.shiny-btn {
  position: relative;
  overflow: hidden;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  font-weight: 700;
  letter-spacing: 1px;
  border: none;
  border-radius: 50px;
  cursor: pointer;
  transition: all 0.3s ease;
  background: linear-gradient(135deg, #cbd5e1, #94a3b8);
  color: #030404;
  box-shadow: 0 4px 24px rgba(203, 213, 225, 0.15);
}

.shiny-btn:hover:not(.shiny-btn-disabled) {
  transform: translateY(-2px) scale(1.02);
  box-shadow: 0 12px 40px rgba(203, 213, 225, 0.25);
}

.shiny-btn-disabled {
  opacity: 0.5;
  cursor: not-allowed;
  transform: none;
}

.shiny-reflect {
  position: absolute;
  top: 0;
  width: 80px;
  height: 100%;
  background: linear-gradient(
    90deg,
    transparent 0%,
    rgba(255, 255, 255, 0.25) 50%,
    transparent 100%
  );
  transform: skewX(-20deg);
  transition: left 0.1s ease;
  pointer-events: none;
}

.shiny-content {
  position: relative;
  z-index: 1;
}
</style>