<script setup lang="ts">
/**
 * FortuneStateView — 运势页统一的加载 / 错误 / 空态呈现。
 * 结构一致：符号 → 说明句 →（可选）补充说明 → 下一步操作。
 */
withDefaults(
  defineProps<{
    kind: 'loading' | 'error' | 'empty'
    title: string
    description?: string
    /** 主操作：内部跳转（优先）或点击重试 */
    actionLabel?: string
    actionTo?: string
    retryLabel?: string
  }>(),
  { description: '', actionLabel: '', actionTo: '', retryLabel: '' },
)

const emit = defineEmits<{ retry: [] }>()

const sigils: Record<string, string> = {
  loading: '五',
  error: '✕',
  empty: '◈',
}
</script>

<template>
  <div class="fsv" :class="`is-${kind}`" aria-live="polite">
    <div v-if="kind === 'loading'" class="fsv-spinner" aria-hidden="true"></div>
    <div v-else class="fsv-sigil" aria-hidden="true">{{ sigils[kind] }}</div>
    <p class="fsv-title">{{ title }}</p>
    <p v-if="description" class="fsv-desc">{{ description }}</p>
    <router-link v-if="actionLabel && actionTo" :to="actionTo" class="fsv-action">
      {{ actionLabel }}
    </router-link>
    <button
      v-else-if="retryLabel"
      type="button"
      class="fsv-action"
      @click="emit('retry')"
    >
      {{ retryLabel }}
    </button>
  </div>
</template>

<style scoped>
.fsv {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 60vh;
  gap: 14px;
  padding: 2rem 1.5rem;
  text-align: center;
}

.fsv-sigil {
  display: grid;
  width: 60px;
  height: 60px;
  place-items: center;
  border: 1px solid var(--line-strong);
  border-radius: 12px;
  color: var(--text-soft);
  background: var(--surface-1);
  font-family: var(--font-serif);
  font-size: var(--fs-2xl);
}
.is-error .fsv-sigil {
  border-color: color-mix(in oklab, var(--crimson) 35%, transparent);
  color: var(--crimson);
  background: color-mix(in oklab, var(--crimson) 6%, transparent);
}

.fsv-spinner {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  border: 2px solid var(--line-subtle);
  border-top-color: rgba(var(--jade-accent-rgb, 22, 207, 140), 0.85);
  animation: fsv-spin 0.9s linear infinite;
}
@keyframes fsv-spin {
  to {
    transform: rotate(360deg);
  }
}

.fsv-title {
  margin: 0;
  font-family: var(--font-serif);
  font-size: var(--fs-lg);
  font-weight: 700;
  letter-spacing: 0.04em;
  color: var(--text);
}
.is-error .fsv-title {
  color: var(--crimson);
}

.fsv-desc {
  margin: 0;
  max-width: 42ch;
  font-size: var(--fs-sm);
  line-height: 1.7;
  color: var(--text-muted);
}

.fsv-action {
  margin-top: 6px;
  min-height: 42px;
  padding: 0.6rem 1.6rem;
  border: 1px solid var(--line-strong);
  border-radius: 8px;
  background: var(--surface-1);
  color: var(--text);
  font-size: var(--fs-sm);
  font-weight: 600;
  letter-spacing: 0.06em;
  text-decoration: none;
  cursor: pointer;
  transition:
    border-color 180ms ease,
    color 180ms ease;
}
.fsv-action:hover {
  border-color: rgba(var(--jade-accent-rgb, 22, 207, 140), 0.55);
  color: rgba(var(--jade-accent-rgb, 22, 207, 140), 1);
}
.fsv-action:active {
  transform: translateY(1px);
}
.fsv-action:focus-visible {
  outline: 2px solid var(--line-focus);
  outline-offset: 1px;
}

@media (prefers-reduced-motion: reduce) {
  .fsv-spinner {
    animation: none;
  }
  .fsv-action {
    transition: none;
  }
}
</style>
