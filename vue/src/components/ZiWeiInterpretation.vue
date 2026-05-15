<script setup lang="ts">
import { ref } from 'vue'

interface SectionData {
  title: string
  content: string
  tags: string[]
}

interface PalaceReading {
  palaceName: string
  mainStarAnalysis: SectionData
  auxStarInfluence: SectionData
  sihuaInfluence: SectionData
  sanFangSiZheng: SectionData
  patternAnnotations: SectionData
}

defineProps<{
  palaceReading: PalaceReading
}>()

const expanded = ref(false)

function toggle() {
  expanded.value = !expanded.value
}
</script>

<template>
  <div
    class="interpretation-card"
    :class="{ expanded }"
  >
    <!-- Card header (always visible, clickable) -->
    <div class="card-header" @click="toggle">
      <div class="flex items-center gap-3">
        <span class="text-xl">{{ expanded ? '📖' : '📘' }}</span>
        <div>
          <h3 class="card-title">{{ palaceReading.palaceName }} 详解</h3>
          <p class="card-subtitle">点击{{ expanded ? '收起' : '展开' }}查看完整解读</p>
        </div>
      </div>
      <span class="toggle-icon">{{ expanded ? '▲' : '▼' }}</span>
    </div>

    <!-- Expandable content -->
    <transition name="expand">
      <div v-if="expanded" class="card-body">
        <!-- Section: 主星特性 -->
        <section class="reading-section">
          <h4 class="section-heading">
            <span class="section-marker">✦</span>
            {{ palaceReading.mainStarAnalysis.title || '主星特性' }}
          </h4>
          <p class="section-text">{{ palaceReading.mainStarAnalysis.content }}</p>
          <div v-if="palaceReading.mainStarAnalysis.tags.length" class="tag-row">
            <span
              v-for="(tag, i) in palaceReading.mainStarAnalysis.tags"
              :key="'ms-' + i"
              class="reading-tag main-star"
            >{{ tag }}</span>
          </div>
        </section>

        <!-- Section: 辅星影响 -->
        <section class="reading-section">
          <h4 class="section-heading">
            <span class="section-marker">◇</span>
            {{ palaceReading.auxStarInfluence.title || '辅星影响' }}
          </h4>
          <p class="section-text">{{ palaceReading.auxStarInfluence.content }}</p>
          <div v-if="palaceReading.auxStarInfluence.tags.length" class="tag-row">
            <span
              v-for="(tag, i) in palaceReading.auxStarInfluence.tags"
              :key="'as-' + i"
              class="reading-tag aux-star"
            >{{ tag }}</span>
          </div>
        </section>

        <!-- Section: 四化影响 -->
        <section class="reading-section">
          <h4 class="section-heading">
            <span class="section-marker">◈</span>
            {{ palaceReading.sihuaInfluence.title || '四化影响' }}
          </h4>
          <p class="section-text">{{ palaceReading.sihuaInfluence.content }}</p>
          <div v-if="palaceReading.sihuaInfluence.tags.length" class="tag-row">
            <span
              v-for="(tag, i) in palaceReading.sihuaInfluence.tags"
              :key="'sh-' + i"
              class="reading-tag sihua-star"
            >{{ tag }}</span>
          </div>
        </section>

        <!-- Section: 三方四正 -->
        <section class="reading-section">
          <h4 class="section-heading">
            <span class="section-marker">△</span>
            {{ palaceReading.sanFangSiZheng.title || '三方四正' }}
          </h4>
          <p class="section-text">{{ palaceReading.sanFangSiZheng.content }}</p>
          <div v-if="palaceReading.sanFangSiZheng.tags.length" class="tag-row">
            <span
              v-for="(tag, i) in palaceReading.sanFangSiZheng.tags"
              :key="'sf-' + i"
              class="reading-tag sanfang-star"
            >{{ tag }}</span>
          </div>
        </section>

        <!-- Section: 格局标注 -->
        <section class="reading-section">
          <h4 class="section-heading">
            <span class="section-marker">★</span>
            {{ palaceReading.patternAnnotations.title || '格局标注' }}
          </h4>
          <p class="section-text">{{ palaceReading.patternAnnotations.content }}</p>
          <div v-if="palaceReading.patternAnnotations.tags.length" class="tag-row">
            <span
              v-for="(tag, i) in palaceReading.patternAnnotations.tags"
              :key="'pa-' + i"
              class="reading-tag pattern-star"
            >{{ tag }}</span>
          </div>
        </section>
      </div>
    </transition>
  </div>
</template>

<style scoped>
@reference "tailwindcss";
.interpretation-card {
  @apply rounded-lg border overflow-hidden transition-shadow duration-200;
  background-color: var(--color-bazi-paper);
  border-color: var(--color-bazi-blue);
}

.interpretation-card:hover {
  box-shadow: 0 4px 12px rgba(43, 58, 66, 0.15);
}

.card-header {
  @apply flex items-center justify-between p-4 cursor-pointer select-none;
  border-bottom: 1px solid transparent;
}

.expanded .card-header {
  border-bottom-color: rgba(43, 58, 66, 0.15);
}

.card-title {
  @apply text-base font-bold m-0;
  color: var(--color-bazi-blue);
}

.card-subtitle {
  @apply text-xs m-0 mt-0.5;
  color: #999;
}

.toggle-icon {
  @apply text-sm;
  color: var(--color-bazi-blue);
}

.card-body {
  @apply p-4 flex flex-col gap-4;
}

.reading-section {
  @apply pb-3 border-b border-dashed;
  border-color: rgba(43, 58, 66, 0.1);
}

.reading-section:last-child {
  @apply border-b-0 pb-0;
}

.section-heading {
  @apply flex items-center gap-1.5 text-sm font-bold mb-2 m-0;
  color: var(--color-bazi-blue);
}

.section-marker {
  color: var(--color-bazi-red);
  font-size: 12px;
}

.section-text {
  @apply text-sm leading-relaxed m-0 mb-2;
  color: var(--color-bazi-ink);
}

.tag-row {
  @apply flex flex-wrap gap-1.5;
}

.reading-tag {
  @apply inline-block rounded-full px-2.5 py-0.5 text-xs font-medium;
}

.main-star {
  background-color: rgba(196, 30, 58, 0.1);
  color: var(--color-bazi-red);
}

.aux-star {
  background-color: rgba(43, 58, 66, 0.08);
  color: var(--color-bazi-blue);
}

.sihua-star {
  background-color: rgba(218, 165, 32, 0.15);
  color: #8B6914;
}

.sanfang-star {
  background-color: rgba(34, 139, 34, 0.1);
  color: #228B22;
}

.pattern-star {
  background-color: rgba(128, 0, 128, 0.1);
  color: #800080;
}

/* Expand/collapse transition */
.expand-enter-active,
.expand-leave-active {
  transition: all 0.3s ease;
  overflow: hidden;
}

.expand-enter-from,
.expand-leave-to {
  opacity: 0;
  max-height: 0;
}

.expand-enter-to,
.expand-leave-from {
  opacity: 1;
  max-height: 2000px;
}
</style>
