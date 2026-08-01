<script setup lang="ts">
interface ElementItem {
  element: string
  image_url: string
  description: string
}

defineProps<{
  elements: ElementItem[]
}>()

const elementMap: Record<string, { color: string; chinese: string; symbol: string }> = {
  金: { color: '#cbd5e1', chinese: '金', symbol: '◇' },
  木: { color: '#34d399', chinese: '木', symbol: '♠' },
  水: { color: '#22d3ee', chinese: '水', symbol: '⬡' },
  火: { color: '#fb7185', chinese: '火', symbol: '▲' },
  土: { color: '#fde68a', chinese: '土', symbol: '◆' },
}
</script>

<template>
  <div class="element-images">
    <div class="elements-header">
      <div class="header-line"></div>
      <span class="header-text">五行元素</span>
      <div class="header-line"></div>
    </div>
    <div class="elements-grid">
      <div
        v-for="(item, idx) in elements"
        :key="idx"
        class="element-card"
        :style="{ '--elem-color': elementMap[item.element]?.color || '#8a9a8e' }"
      >
        <div class="card-glow"></div>
        <div
          class="element-orb"
          :style="{
            background: `radial-gradient(circle, ${elementMap[item.element]?.color}33, transparent)`,
          }"
        >
          <div class="orb-inner" :style="{ color: elementMap[item.element]?.color }">
            {{ elementMap[item.element]?.symbol }}
          </div>
        </div>
        <div class="element-name" :style="{ color: elementMap[item.element]?.color }">
          {{ item.element }}
        </div>
        <p class="element-desc">{{ item.description || '天地五行之一' }}</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.element-images {
  padding: 0.5rem 0;
}

.elements-header {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 1rem;
}

.header-line {
  flex: 1;
  height: 1px;
  background: linear-gradient(90deg, transparent, var(--line-strong), transparent);
}

.header-text {
  font-size: var(--fs-xs);
  letter-spacing: 3px;
  color: var(--text-soft);
  text-transform: uppercase;
}

.elements-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(96px, 1fr));
  gap: 0.5rem;
}

.element-card {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.4rem;
  padding: 0.875rem 0.5rem;
  background: var(--glass-bg);
  border: 1px solid color-mix(in oklab, var(--elem-color) 20%, transparent);
  border-radius: 12px;
  transition: all 0.3s;
  overflow: hidden;
}

.element-card:hover {
  border-color: color-mix(in oklab, var(--elem-color) 45%, transparent);
  transform: translateY(-2px);
}

.element-card:hover .card-glow {
  opacity: 1;
}

.card-glow {
  position: absolute;
  inset: 0;
  background: radial-gradient(
    circle at 50% 30%,
    color-mix(in oklab, var(--elem-color) 7%, transparent),
    transparent 70%
  );
  opacity: 0;
  transition: opacity 0.3s;
}

.element-orb {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.orb-inner {
  font-size: var(--fs-2xl);
  font-weight: 700;
  text-shadow: 0 0 10px color-mix(in oklab, currentColor 55%, transparent);
}

.element-name {
  font-size: var(--fs-sm);
  font-weight: 700;
  letter-spacing: 1px;
}

.element-desc {
  font-size: var(--fs-2xs);
  color: var(--text-muted);
  text-align: center;
  margin: 0;
  line-height: 1.3;
  padding: 0 0.25rem;
}
</style>
