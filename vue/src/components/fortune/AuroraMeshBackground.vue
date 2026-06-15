<script setup lang="ts">
/**
 * Aurora mesh background — three radial gradients drifting in five-element hue.
 * Uses --jade-accent-rgb (set globally by themeStore), so it follows the user's
 * current 五行 selection automatically.
 */
</script>

<template>
  <div class="aurora" aria-hidden="true">
    <div class="layer l1"></div>
    <div class="layer l2"></div>
    <div class="layer l3"></div>
  </div>
</template>

<style scoped>
.aurora {
  position: fixed;
  inset: 0;
  z-index: -1;
  pointer-events: none;
  overflow: hidden;
}

.layer {
  position: absolute;
  inset: -20%;
  filter: blur(70px) saturate(140%);
  mix-blend-mode: screen;
  will-change: transform, opacity;
}

.l1 {
  background: radial-gradient(
    60% 50% at 18% 22%,
    rgba(var(--jade-accent-rgb), 0.42),
    transparent 60%
  );
  animation: drift1 28s ease-in-out infinite alternate;
}

.l2 {
  background: radial-gradient(
    50% 60% at 82% 18%,
    rgba(var(--jade-accent-rgb), 0.30),
    transparent 65%
  );
  animation: drift2 36s ease-in-out infinite alternate;
}

.l3 {
  background: radial-gradient(
    70% 60% at 50% 92%,
    rgba(var(--jade-accent-rgb), 0.22),
    transparent 70%
  );
  animation: drift3 44s ease-in-out infinite alternate;
}

:global(.dark) .layer { mix-blend-mode: lighten; }
:global(.dark) .l1 { background: radial-gradient(60% 50% at 18% 22%, rgba(var(--jade-accent-rgb), 0.22), transparent 60%); }
:global(.dark) .l2 { background: radial-gradient(50% 60% at 82% 18%, rgba(var(--jade-accent-rgb), 0.18), transparent 65%); }
:global(.dark) .l3 { background: radial-gradient(70% 60% at 50% 92%, rgba(var(--jade-accent-rgb), 0.14), transparent 70%); }

@keyframes drift1 {
  from { transform: translate3d(-4%, -3%, 0) scale(1); }
  to   { transform: translate3d(6%, 4%, 0) scale(1.08); }
}
@keyframes drift2 {
  from { transform: translate3d(3%, -2%, 0) scale(1.04); }
  to   { transform: translate3d(-5%, 5%, 0) scale(1); }
}
@keyframes drift3 {
  from { transform: translate3d(0, 4%, 0) scale(1); }
  to   { transform: translate3d(2%, -4%, 0) scale(1.06); }
}

@media (prefers-reduced-motion: reduce) {
  .layer { animation: none; }
}
</style>
