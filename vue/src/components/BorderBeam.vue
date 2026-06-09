<script setup lang="ts">
const props = withDefaults(defineProps<{
  color?: string
  duration?: number
  size?: number
}>(), {
  color: 'var(--wuxing-shui)',
  duration: 4,
  size: 2
})
</script>

<template>
  <div class="border-beam-wrapper" :style="{ '--beam-size': props.size + 'px', '--beam-duration': props.duration + 's', '--beam-color': props.color }">
    <div class="border-beam"></div>
  </div>
</template>

<style scoped>
.border-beam-wrapper {
  position: absolute;
  inset: 0;
  pointer-events: none;
  overflow: hidden;
  border-radius: inherit;
}

.border-beam {
  position: absolute;
  inset: 0;
  border-radius: inherit;
  mask-image: linear-gradient(
    var(--beam-angle, 0deg),
    transparent calc(50% - var(--beam-size)),
    black calc(50% - var(--beam-size) / 2),
    black calc(50% + var(--beam-size) / 2),
    transparent calc(50% + var(--beam-size))
  );
  -webkit-mask-image: linear-gradient(
    var(--beam-angle, 0deg),
    transparent calc(50% - var(--beam-size)),
    black calc(50% - var(--beam-size) / 2),
    black calc(50% + var(--beam-size) / 2),
    transparent calc(50% + var(--beam-size))
  );
  background: conic-gradient(
    from 0deg,
    transparent 0%,
    var(--beam-color) 10%,
    transparent 20%
  );
  animation: beam-spin var(--beam-duration) linear infinite;
}

@keyframes beam-spin {
  from {
    --beam-angle: 0deg;
  }
  to {
    --beam-angle: 360deg;
  }
}

@property --beam-angle {
  syntax: '<angle>';
  initial-value: 0deg;
  inherits: false;
}
</style>