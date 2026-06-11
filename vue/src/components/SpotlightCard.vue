<script setup lang="ts">
import { ref } from 'vue'

const props = withDefaults(defineProps<{
  spotlightColor?: string
}>(), {
  spotlightColor: 'rgba(203, 213, 225, 0.06)'
})

const mouseX = ref(0)
const mouseY = ref(0)
const isHovered = ref(false)

function handleMouseMove(e: MouseEvent) {
  const rect = (e.currentTarget as HTMLElement).getBoundingClientRect()
  mouseX.value = e.clientX - rect.left
  mouseY.value = e.clientY - rect.top
}

function handleMouseEnter() {
  isHovered.value = true
}

function handleMouseLeave() {
  isHovered.value = false
}
</script>

<template>
  <div
    class="spotlight-card"
    :style="{
      '--mouse-x': mouseX + 'px',
      '--mouse-y': mouseY + 'px',
      '--spotlight-color': props.spotlightColor,
      '--spotlight-opacity': isHovered ? '1' : '0',
    }"
    @mousemove="handleMouseMove"
    @mouseenter="handleMouseEnter"
    @mouseleave="handleMouseLeave"
  >
    <div class="spotlight-glow"></div>
    <div class="spotlight-content">
      <slot />
    </div>
  </div>
</template>

<style scoped>
.spotlight-card {
  position: relative;
  border-radius: inherit;
  overflow: hidden;
}

.spotlight-glow {
  position: absolute;
  inset: 0;
  pointer-events: none;
  z-index: 1;
  opacity: var(--spotlight-opacity);
  background: radial-gradient(
    650px circle at var(--mouse-x) var(--mouse-y),
    var(--spotlight-color),
    transparent 40%
  );
  transition: opacity 0.4s ease;
}

.spotlight-content {
  position: relative;
  z-index: 2;
}
</style>
