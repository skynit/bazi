<script setup lang="ts">
import { onMounted, onUnmounted, ref, watch, nextTick } from 'vue'
import createGlobe from 'cobe'

type WuxingKey = 'mu' | 'huo' | 'tu' | 'jin' | 'shui'

interface WuxingTheme {
  markerColor: number[]
  baseColor: number[]
  glowColor: number[]
  accentText: string
  dropShadow: string
}

const props = withDefaults(defineProps<{
  elementTheme?: WuxingKey
}>(), {
  elementTheme: 'jin' as WuxingKey
})

const canvasRef = ref<HTMLCanvasElement | null>(null)
let globe: ReturnType<typeof createGlobe> | null = null
let currentPhi = 0

const wuxingThemes: Record<WuxingKey, WuxingTheme> = {
  mu: {
    markerColor: [52 / 255, 211 / 255, 153 / 255],
    baseColor: [0.02, 0.05, 0.03],
    glowColor: [0.02, 0.1, 0.05],
    accentText: 'text-emerald-400',
    dropShadow: 'drop-shadow-[0_0_50px_rgba(52,211,153,0.1)]'
  },
  huo: {
    markerColor: [251 / 255, 113 / 255, 133 / 255],
    baseColor: [0.05, 0.02, 0.03],
    glowColor: [0.1, 0.02, 0.04],
    accentText: 'text-rose-400',
    dropShadow: 'drop-shadow-[0_0_50px_rgba(251,113,133,0.1)]'
  },
  tu: {
    markerColor: [252 / 255, 211 / 255, 77 / 255],
    baseColor: [0.04, 0.03, 0.01],
    glowColor: [0.08, 0.06, 0.01],
    accentText: 'text-amber-400',
    dropShadow: 'drop-shadow-[0_0_50px_rgba(252,211,77,0.1)]'
  },
  jin: {
    markerColor: [203 / 255, 213 / 255, 225 / 255],
    baseColor: [0.03, 0.03, 0.04],
    glowColor: [0.04, 0.04, 0.06],
    accentText: 'text-[var(--text-muted)]',
    dropShadow: 'drop-shadow-[0_0_50px_rgba(203,213,225,0.06)]'
  },
  shui: {
    markerColor: [34 / 255, 211 / 255, 238 / 255],
    baseColor: [0.01, 0.03, 0.05],
    glowColor: [0.01, 0.08, 0.12],
    accentText: 'text-cyan-400',
    dropShadow: 'drop-shadow-[0_0_50px_rgba(34,211,238,0.1)]'
  }
}

const PHANTOM_SIZE = 500

const renderGlobe = () => {
  if (!canvasRef.value) return
  if (globe) globe.destroy()

  const theme = wuxingThemes[props.elementTheme] || wuxingThemes.jin
  const dpr = window.devicePixelRatio || 2

  globe = createGlobe(canvasRef.value!, {
    devicePixelRatio: dpr,
    width: PHANTOM_SIZE * dpr,
    height: PHANTOM_SIZE * dpr,
    phi: currentPhi,
    theta: 0.35,
    dark: 1,
    diffuse: 1.1,
    mapSamples: 18000,
    mapBrightness: 5.5,
    baseColor: theme.baseColor,
    markerColor: theme.markerColor,
    glowColor: theme.glowColor,
    markers: [
      { location: [34.3416, 108.9398] as [number, number], size: 0.04 },
      { location: [31.2304, 121.4737] as [number, number], size: 0.03 }
    ],
    onRender: (state: any) => {
      currentPhi += 0.0025
      state.phi = currentPhi
    }
  } as any)
}

watch(() => props.elementTheme, async () => {
  await nextTick()
  renderGlobe()
})

onMounted(() => renderGlobe())
onUnmounted(() => { if (globe) globe.destroy() })
</script>

<template>
  <div class="relative w-[500px] h-[500px] sm:w-[600px] sm:h-[600px] flex items-center justify-center select-none">

    <div class="absolute inset-0 border border-[var(--line-strong)] rounded-full pointer-events-none scale-105 border-dashed animate-[spin_200s_linear_infinite]"></div>
    <div class="absolute w-[85%] h-[85%] border border-[var(--line-subtle)] rounded-full pointer-events-none"></div>

    <canvas ref="canvasRef" class="w-full h-full bg-transparent opacity-75 mix-blend-screen cursor-grab active:cursor-grabbing transition-all duration-1000" :class="wuxingThemes[props.elementTheme]?.dropShadow" />

    <div class="absolute bottom-4 bg-[var(--surface-1)]/90 border border-[var(--line-strong)] backdrop-blur-md px-4 py-1.5 rounded-full text-[var(--fs-2xs)] font-mono tracking-widest text-[var(--text-muted)] shadow-[var(--shadow-sm)] z-20 flex items-center gap-2">
      <span class="w-1.5 h-1.5 rounded-full animate-pulse" :class="wuxingThemes[props.elementTheme]?.accentText || 'bg-zinc-500'"></span>
      THEME: <span class="font-bold uppercase" :class="wuxingThemes[props.elementTheme]?.accentText">{{ props.elementTheme }}</span>
    </div>

  </div>
</template>
