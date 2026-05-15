<script setup lang="ts">
interface ElementItem {
  element: string
  image_url: string
  description: string
}

defineProps<{
  elements: ElementItem[]
}>()

const elementColors: Record<string, string> = {
  金: '#FFD700',
  木: '#228B22',
  水: '#4169E1',
  火: '#DC143C',
  土: '#DAA520',
}

function getColor(element: string): string {
  return elementColors[element] || '#999'
}
</script>

<template>
  <div class="w-full">
    <div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-5 gap-4">
      <div
        v-for="(item, idx) in elements"
        :key="idx"
        class="flex flex-col items-center rounded-lg border border-[var(--color-bazi-paper)] bg-[var(--color-bazi-paper)] p-4 shadow-sm transition-shadow hover:shadow-md"
      >
        <!-- Element name -->
        <span class="mb-2 text-lg font-bold text-[var(--color-bazi-ink)]">
          {{ item.element }}
        </span>

        <!-- Colored placeholder block -->
        <div
          class="mb-2 flex h-28 w-full items-center justify-center rounded-md"
          :style="{ backgroundColor: getColor(item.element) }"
        >
          <span class="text-sm text-white font-medium drop-shadow-md">
            图片资源待补充
          </span>
        </div>

        <!-- Description -->
        <p class="text-center text-xs text-[var(--color-bazi-blue)] leading-relaxed">
          {{ item.description }}
        </p>
      </div>
    </div>
  </div>
</template>
