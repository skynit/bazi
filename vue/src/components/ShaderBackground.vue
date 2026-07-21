<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, computed } from 'vue'
import { ShaderMount } from '@paper-design/shaders'
import {
  type WuxingKey,
  type ShaderType,
  type ShaderThemeMode,
  bgImages,
  getShaderFragment,
  createShaderUniforms,
  getShaderSpeed
} from '../composables/useWuxingThemes'

const props = withDefaults(defineProps<{
  elementTheme?: WuxingKey
  shaderType?: ShaderType
  overlayOpacity?: number
}>(), {
  elementTheme: 'mu',
  shaderType: 'grainGradient',
  overlayOpacity: 0.4
})

const shaderHost = ref<HTMLDivElement | null>(null)
const loaded = ref(false)
const themeMode = ref<ShaderThemeMode>(document.documentElement.classList.contains('dark') ? 'dark' : 'light')
let shaderMount: ShaderMount | null = null
let currentImage: HTMLImageElement | null = null
let themeObserver: MutationObserver | null = null

function loadImage(src: string): Promise<HTMLImageElement> {
  return new Promise((resolve) => {
    const img = new Image()
    img.crossOrigin = 'anonymous'
    img.onload = () => resolve(img)
    img.onerror = () => resolve(img)
    img.src = src
  })
}

const shaderMountConfig = {
  alpha: true,
  antialias: true
} as const
const shaderMinPixelRatio = Math.min(window.devicePixelRatio || 2, 2)
const shaderMaxPixels = 4096 * 4096

async function mountHalftoneCmyk() {
  if (!shaderHost.value) return

  shaderMount?.dispose()
  shaderMount = null
  loaded.value = false

  const src = bgImages[props.elementTheme]
  currentImage = await loadImage(src)

  if (!shaderHost.value) return

  shaderMount = new ShaderMount(
    shaderHost.value,
    getShaderFragment(props.shaderType),
    createShaderUniforms(props.shaderType, props.elementTheme, currentImage, themeMode.value),
    shaderMountConfig,
    getShaderSpeed(props.shaderType),
    0,
    shaderMinPixelRatio,
    shaderMaxPixels
  )
  loaded.value = true
}

function mountShader() {
  if (!shaderHost.value) return
  shaderMount = new ShaderMount(
    shaderHost.value,
    getShaderFragment(props.shaderType),
    createShaderUniforms(props.shaderType, props.elementTheme, undefined, themeMode.value),
    shaderMountConfig,
    getShaderSpeed(props.shaderType),
    0,
    shaderMinPixelRatio,
    shaderMaxPixels
  )
  loaded.value = true
}

onMounted(() => {
  themeObserver = new MutationObserver(() => {
    themeMode.value = document.documentElement.classList.contains('dark') ? 'dark' : 'light'
  })
  themeObserver.observe(document.documentElement, { attributes: true, attributeFilter: ['class'] })

  if (props.shaderType === 'halftoneCmyk') {
    mountHalftoneCmyk()
  } else {
    mountShader()
  }
})

onUnmounted(() => {
  themeObserver?.disconnect()
  themeObserver = null
  shaderMount?.dispose()
  shaderMount = null
})

watch([() => props.elementTheme, themeMode], () => {
  if (props.shaderType === 'halftoneCmyk') {
    mountHalftoneCmyk()
  } else {
    shaderMount?.setUniforms(createShaderUniforms(props.shaderType, props.elementTheme, undefined, themeMode.value))
  }
})

const overlayStyle = computed(() => ({
  opacity: props.overlayOpacity
}))
</script>

<template>
  <div class="shader-bg-container">
    <div ref="shaderHost" class="shader-canvas-host" aria-hidden="true"></div>
    <div class="shader-overlay" :style="overlayStyle"></div>
  </div>
</template>

<style scoped>
.shader-bg-container {
  position: absolute;
  inset: 0;
  z-index: 0;
  pointer-events: none;
  overflow: hidden;
  background: var(--bg);
  opacity: 0.5;
}

.shader-canvas-host {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
}

.shader-canvas-host :deep(canvas) {
  display: block;
  width: 100% !important;
  height: 100% !important;
}

.shader-overlay {
  position: absolute;
  inset: 0;
  background:
    radial-gradient(900px circle at 14% 18%, rgba(255,255,255,.38), transparent 54%),
    radial-gradient(820px circle at 82% 22%, rgba(255,255,255,.22), transparent 58%),
    linear-gradient(180deg, rgba(250,251,248,.22) 0%, rgba(250,251,248,.54) 72%, rgba(250,251,248,.78) 100%);
  pointer-events: none;
}

:global(.dark) .shader-bg-container {
  opacity: 0.54;
}

:global(.dark) .shader-overlay {
  background:
    radial-gradient(ellipse at 22% 4%, transparent 0%, rgba(0,0,0,.02) 40%, rgba(0,0,0,.18) 76%, rgba(0,0,0,.5) 100%),
    linear-gradient(90deg, rgba(0,0,0,0) 0%, rgba(0,0,0,.04) 58%, rgba(0,0,0,.5) 100%),
    linear-gradient(180deg, rgba(0,0,0,0) 0%, rgba(0,0,0,.03) 44%, rgba(0,0,0,.62) 100%);
}
</style>
