<script setup lang="ts">
import { ref } from 'vue'

/**
 * ChartSection — 命盘页长滚动分区容器。
 * 分区头模型：eyebrow（编号）+ 衬线题 + 说明句；可折叠分区在折叠时仍显示摘要句，
 * 让用户不展开也能获得该分区的核心信息。展开/收起使用 grid-rows 过渡（≤350ms），
 * 尊重 prefers-reduced-motion。展开状态仅保留在组件内（会话级）。
 */
const props = withDefaults(
  defineProps<{
    id: string
    eyebrow?: string
    title: string
    desc?: string
    /** 折叠时展示的一句话内容摘要 */
    summary?: string
    defaultOpen?: boolean
    collapsible?: boolean
  }>(),
  {
    eyebrow: '',
    desc: '',
    summary: '',
    defaultOpen: true,
    collapsible: true,
  },
)

const open = ref(props.defaultOpen)

function toggle() {
  if (!props.collapsible) return
  open.value = !open.value
}
</script>

<template>
  <section
    :id="id"
    class="chart-section"
    :class="{ 'is-open': open, 'is-collapsible': collapsible }"
  >
    <component
      :is="collapsible ? 'button' : 'div'"
      class="section-head"
      v-bind="
        collapsible
          ? { type: 'button', 'aria-expanded': open, 'aria-controls': `${id}-body` }
          : {}
      "
      @click="toggle"
    >
      <span class="section-head-text">
        <span v-if="eyebrow" class="section-eyebrow">{{ eyebrow }}</span>
        <span class="section-title-row">
          <span class="section-title">{{ title }}</span>
          <span v-if="collapsible" class="section-toggle" aria-hidden="true">
            <span class="section-toggle-label">{{ open ? '收起' : '展开' }}</span>
            <svg
              class="section-chevron"
              :class="{ 'is-open': open }"
              width="12"
              height="12"
              viewBox="0 0 12 12"
              fill="none"
            >
              <path
                d="M3 4.5l3 3 3-3"
                stroke="currentColor"
                stroke-width="1.5"
                stroke-linecap="round"
                stroke-linejoin="round"
              />
            </svg>
          </span>
        </span>
        <span v-if="desc" class="section-desc">{{ desc }}</span>
        <span v-if="collapsible && !open && summary" class="section-summary">{{ summary }}</span>
      </span>
    </component>
    <div
      v-if="collapsible"
      :id="`${id}-body`"
      class="section-collapse"
      :class="{ 'is-closed': !open }"
    >
      <div class="section-body">
        <slot />
      </div>
    </div>
    <div v-else class="section-body section-body-static">
      <slot />
    </div>
  </section>
</template>

<style scoped>
.chart-section {
  scroll-margin-top: 96px;
  padding: 1.35rem 0 1.5rem;
  border-top: 1px solid var(--line-subtle);
}

.chart-section:first-child {
  border-top: 0;
  padding-top: 0.35rem;
}

.section-head {
  display: block;
  width: 100%;
  padding: 0;
  margin: 0;
  background: none;
  border: 0;
  font: inherit;
  color: inherit;
  text-align: left;
}

.is-collapsible > .section-head {
  cursor: pointer;
  border-radius: 6px;
}

.is-collapsible > .section-head:focus-visible {
  outline: 2px solid var(--line-focus);
  outline-offset: 3px;
}

.section-head-text {
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
  min-width: 0;
}

.section-eyebrow {
  font-size: var(--fs-2xs);
  letter-spacing: 2.5px;
  color: var(--text-soft);
  text-transform: uppercase;
}

.section-title-row {
  display: flex;
  align-items: baseline;
  gap: 0.75rem;
  min-width: 0;
}

.section-title {
  font-family: var(--font-serif), serif;
  font-size: var(--fs-lg);
  font-weight: 700;
  color: var(--text);
  letter-spacing: 2px;
  line-height: 1.3;
}

.section-toggle {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  flex-shrink: 0;
  color: var(--text-soft);
  transition: color 0.2s;
}

.is-collapsible > .section-head:hover .section-toggle {
  color: var(--accent);
}

.section-toggle-label {
  font-size: var(--fs-2xs);
  letter-spacing: 1px;
}

.section-chevron {
  transition: transform 0.28s cubic-bezier(0.16, 1, 0.3, 1);
}

.section-chevron.is-open {
  transform: rotate(180deg);
}

.section-desc {
  max-width: 46rem;
  font-size: var(--fs-xs);
  color: var(--text-muted);
  line-height: 1.7;
}

.section-summary {
  display: block;
  margin-top: 0.35rem;
  padding: 0.55rem 0.75rem;
  border-left: 2px solid var(--line-strong);
  background: color-mix(in oklab, var(--surface-2) 42%, transparent);
  border-radius: 0 6px 6px 0;
  font-size: var(--fs-xs);
  color: var(--text-muted);
  line-height: 1.6;
  overflow-wrap: anywhere;
}

.section-collapse {
  display: grid;
  grid-template-rows: 1fr;
  transition: grid-template-rows 320ms cubic-bezier(0.16, 1, 0.3, 1);
}

.section-collapse.is-closed {
  grid-template-rows: 0fr;
}

.section-collapse > .section-body {
  overflow: hidden;
  min-height: 0;
}

.section-body {
  padding-top: 1.1rem;
}

@media (prefers-reduced-motion: reduce) {
  .section-collapse,
  .section-chevron {
    transition: none;
  }
}

@media (max-width: 640px) {
  .chart-section {
    scroll-margin-top: 128px;
    padding: 1.1rem 0 1.2rem;
  }

  .section-title {
    font-size: var(--fs-body);
  }
}
</style>
