<template>
  <div class="absolute inset-0 pointer-events-none isolate z-0 overflow-hidden bg-[var(--bg)]">

    <div
      class="absolute -bottom-[35%] -left-[20%] w-[140vw] h-[140vw] rounded-full
             bg-gradient-to-tr from-[#011612] via-[#045841] to-[#2fd397]
             opacity-80 blur-[1px] transition-all duration-1000 ease-in-out"
      :class="nebulaClass"
      style="mask-image: radial-gradient(circle at 38% 38%, black 25%, transparent 68%);
             -webkit-mask-image: radial-gradient(circle at 38% 38%, black 25%, transparent 68%);"
    ></div>

    <div class="absolute inset-0 opacity-[0.15] mix-blend-overlay">
      <svg xmlns="http://www.w3.org/2000/svg" width="100%" height="100%">
        <filter id="jadense-grain">
          <feTurbulence
            type="fractalNoise"
            baseFrequency="0.65"
            numOctaves="3"
            stitchTiles="stitch"
          />
          <feColorMatrix type="matrix" values="1 0 0 0 0  0 1 0 0 0  0 0 1 0 0  0 0 0 1.4 -0.4" />
        </filter>
        <rect width="100%" height="100%" filter="url(#jadense-grain)" />
      </svg>
    </div>

    <div class="absolute inset-0 bg-gradient-to-b from-transparent via-[var(--bg)]/20 to-[var(--bg)]"></div>

  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

type WuxingKey = 'mu' | 'huo' | 'tu' | 'jin' | 'shui'

const props = withDefaults(defineProps<{
  elementTheme?: WuxingKey
}>(), {
  elementTheme: 'mu' as WuxingKey
})

const nebulaOverrides: Record<string, string> = {
  huo: 'from-[#160105] via-[#7f1d1d] to-[#f43f5e]',
  tu: 'from-[#161405] via-[#78661d] to-[#fde68a]',
  jin: 'from-[#0d0d10] via-[#3f3f50] to-[#cbd5e1]',
  shui: 'from-[#011316] via-[#044058] to-[#22d3ee]'
}

const nebulaClass = computed(() => nebulaOverrides[props.elementTheme] || '')
</script>
