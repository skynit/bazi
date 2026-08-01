<script setup lang="ts">
/**
 * PeriodNav — 运势周期切换导航（今日 / 本周 / 本月 / 加持）。
 * 四个运势页面共用，当前页高亮并标注 aria-current。
 * variant="overlay" 用于加持页头图上的深底场景。
 */
import { computed } from 'vue'

type Period = 'day' | 'week' | 'month' | 'blessing'

const props = withDefaults(
  defineProps<{
    current: Period
    chartId?: string | number | null
    variant?: 'default' | 'overlay'
  }>(),
  { chartId: null, variant: 'default' },
)

const items = computed(() => {
  const query = props.chartId ? { chart_id: String(props.chartId) } : {}
  return [
    { key: 'day' as Period, label: '今日', to: { path: '/fortune', query } },
    { key: 'week' as Period, label: '本周', to: { path: '/fortune/weekly', query } },
    { key: 'month' as Period, label: '本月', to: { path: '/fortune/monthly', query } },
    { key: 'blessing' as Period, label: '加持', to: { path: '/fortune/blessing', query } },
  ]
})
</script>

<template>
  <nav class="period-nav" :class="`is-${variant}`" aria-label="运势周期切换">
    <template v-for="item in items" :key="item.key">
      <span v-if="item.key === current" class="pn-item pn-current" aria-current="page">
        {{ item.label }}
      </span>
      <router-link v-else :to="item.to" class="pn-item pn-link">{{ item.label }}</router-link>
    </template>
  </nav>
</template>

<style scoped>
.period-nav {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  padding: 3px;
  border-radius: 10px;
  border: 1px solid var(--line-strong);
  background: var(--surface-1);
  box-shadow: var(--shadow-xs);
}

.pn-item {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 34px;
  padding: 0 14px;
  border-radius: 7px;
  font-size: var(--fs-xs);
  font-weight: 600;
  letter-spacing: 0.08em;
  text-decoration: none;
  color: var(--text-muted);
  transition:
    color 160ms ease,
    background-color 160ms ease;
}

.pn-link:hover {
  color: var(--text);
  background: var(--surface-2);
}
.pn-link:active {
  transform: translateY(1px);
}

.pn-current {
  color: var(--text);
  background: var(--surface-2);
  box-shadow: inset 0 0 0 1px var(--line-strong);
  font-weight: 700;
}

.pn-link:focus-visible {
  outline: 2px solid var(--line-focus);
  outline-offset: 1px;
}

/* 深底（加持页头图）变体 */
.period-nav.is-overlay {
  border-color: rgba(255, 255, 255, 0.34);
  background: rgba(19, 34, 26, 0.62);
  box-shadow: none;
}
.period-nav.is-overlay .pn-item {
  color: rgba(255, 255, 255, 0.78);
}
.period-nav.is-overlay .pn-link:hover {
  color: #ffffff;
  background: rgba(255, 255, 255, 0.1);
}
.period-nav.is-overlay .pn-current {
  color: #ffffff;
  background: rgba(255, 255, 255, 0.16);
  box-shadow: none;
}

@media (prefers-reduced-motion: reduce) {
  .pn-item,
  .pn-link {
    transition: none;
  }
}
</style>
