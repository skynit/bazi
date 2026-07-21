<script setup lang="ts">
import { computed, ref, watch, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useThemeStore } from '../stores/theme'
import {
  ShaderMount,
  grainGradientFragmentShader
} from '@paper-design/shaders'
import {
  type ShaderThemeMode,
  createShaderUniforms
} from '../composables/useWuxingThemes'

const router = useRouter()
const authStore = useAuthStore()
const themeStore = useThemeStore()

const savedChartId = ref<number | null>(null)
const mounted = ref(false)
const shaderHost = ref<HTMLDivElement | null>(null)
const themeMode = ref<ShaderThemeMode>(document.documentElement.classList.contains('dark') ? 'dark' : 'light')
const currentElementLabel = computed(() => ({
  mu: '木',
  huo: '火',
  tu: '土',
  jin: '金',
  shui: '水',
})[themeStore.elementTheme] || themeStore.elementTheme)

let shaderMount: ShaderMount | null = null
let themeObserver: MutationObserver | null = null

onMounted(async () => {
  window.scrollTo({ top: 0, left: 0, behavior: 'auto' })

  if (authStore.isLoggedIn() && !authStore.user) {
    await authStore.fetchMe().catch(() => { })
  }

  const saved = localStorage.getItem('bazi_last_birth')
  if (saved) {
    try {
      savedChartId.value = JSON.parse(saved).chartId || null
    } catch {
      // ignore malformed local state
    }
  }

  setTimeout(() => {
    mounted.value = true
  }, 100)

  if (shaderHost.value) {
    shaderMount = new ShaderMount(
      shaderHost.value,
      grainGradientFragmentShader,
      createShaderUniforms('grainGradient', themeStore.elementTheme, undefined, themeMode.value),
      { alpha: true, antialias: true },
      2,
      0,
      Math.min(window.devicePixelRatio || 1, 2),
      undefined
    )
  }

  themeObserver = new MutationObserver(() => {
    themeMode.value = document.documentElement.classList.contains('dark') ? 'dark' : 'light'
    shaderMount?.setUniforms(createShaderUniforms('grainGradient', themeStore.elementTheme, undefined, themeMode.value))
  })
  themeObserver.observe(document.documentElement, { attributes: true, attributeFilter: ['class'] })
})

watch(() => themeStore.elementTheme, (key) => {
  shaderMount?.setUniforms(createShaderUniforms('grainGradient', key, undefined, themeMode.value))
})

onUnmounted(() => {
  themeObserver?.disconnect()
  themeObserver = null
  shaderMount?.dispose()
  shaderMount = null
})

function startChart() {
  router.push('/chart/new')
}

function continueChart() {
  if (savedChartId.value) router.push(`/chart/${savedChartId.value}`)
}
</script>

<template>
  <div ref="shaderHost" class="home-grain-shader" aria-hidden="true"></div>
  <div class="home-page" :class="{ visible: mounted }">
    <main class="hero-main">
      <section class="hero-panel" aria-labelledby="home-hero-title">
        <div class="hero-copy">
          <h1 id="home-hero-title" class="hero-title">
            八字命理
            <span class="title-accent">推演</span>
          </h1>

          <p class="hero-sub">
            一页看八字、流年与紫微。
          </p>

          <div class="cta-group">
            <button @click="startChart" class="btn-primary-base">
              开始排盘
            </button>
            <button v-if="savedChartId" @click="continueChart" class="btn-secondary">
              继续上次
            </button>
          </div>

          <p class="hero-meta" aria-label="当前主题">
            当前五行主题：<strong>{{ currentElementLabel }}</strong>
          </p>
        </div>
      </section>
    </main>
  </div>
</template>

<style scoped>
.home-page {
  position: relative;
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: clamp(72px, 8vh, 104px) 24px 36px;
  background: transparent;
  color: var(--text);
  opacity: 0;
  transform: translateY(-2px);
  transition: opacity 1s cubic-bezier(0.16, 1, 0.3, 1), transform 1s cubic-bezier(0.16, 1, 0.3, 1);
}

.home-page.visible {
  opacity: 1;
  transform: translateY(0);
}

.home-grain-shader {
  position: fixed;
  inset: 0;
  z-index: 0;
  pointer-events: none;
}

.home-grain-shader :deep(canvas) {
  position: fixed !important;
  inset: 0 !important;
  width: 100vw !important;
  height: 100vh !important;
}

.hero-main {
  position: relative;
  z-index: 20;
  width: min(640px, 100%);
  --hero-title-size: 64px;
  --hero-accent-size: 50px;
  --hero-title-tracking: 0.14em;
  --hero-accent-tracking: 0.18em;
  --hero-title-gap: 0.24em;
  --hero-copy-gap: 28px;
  --hero-cta-offset: 22px;
  --hero-meta-offset: 18px;
  transform: translateZ(0);
}

.hero-panel {
  display: grid;
  place-items: center;
  align-content: center;
  justify-items: center;
  gap: 0;
  min-height: min(420px, calc(100vh - 200px));
  text-align: center;
}

.hero-copy {
  display: grid;
  align-content: center;
  justify-items: center;
  gap: var(--hero-copy-gap);
  width: 100%;
}

.hero-title {
  margin: 0;
  font-family: var(--font-serif), 'Songti SC', serif;
  font-size: var(--hero-title-size);
  font-weight: 600;
  line-height: 1.04;
  letter-spacing: var(--hero-title-tracking);
  color: var(--text);
  text-wrap: balance;
}

.title-accent {
  display: block;
  margin-top: var(--hero-title-gap);
  font-size: var(--hero-accent-size);
  font-weight: 700;
  line-height: 1.08;
  letter-spacing: var(--hero-accent-tracking);
  color: rgba(var(--jade-accent-rgb), 1);
}

.hero-sub {
  max-width: 380px;
  margin: 0;
  font-family: var(--font-sans);
  font-size: clamp(13px, 0.95vw, 15px);
  line-height: 1.68;
  letter-spacing: 0.01em;
  color: var(--text-soft);
  text-wrap: balance;
}

.cta-group {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  flex-wrap: wrap;
  margin-top: var(--hero-cta-offset);
}

.btn-primary-base,
.btn-secondary {
  min-width: 154px;
  padding: 13px 22px;
  border-radius: 999px;
  font-size: 13px;
  letter-spacing: 0.08em;
  transition:
    transform 0.25s cubic-bezier(0.16, 1, 0.3, 1),
    border-color 0.25s ease,
    background 0.25s ease,
    color 0.25s ease,
    box-shadow 0.25s ease;
}

.btn-primary-base {
  border: none;
  background: linear-gradient(135deg, var(--jade-accent) 0%, var(--jade-accent-dark) 100%);
  color: var(--jade-button-text);
  font-weight: 700;
  box-shadow:
    0 14px 32px rgba(var(--jade-accent-rgb), 0.24),
    inset 0 1px 0 rgba(255, 255, 255, 0.2);
}

.btn-primary-base:hover {
  transform: translateY(-1px);
  box-shadow:
    0 18px 36px rgba(var(--jade-accent-rgb), 0.28),
    inset 0 1px 0 rgba(255, 255, 255, 0.2);
}

.btn-secondary {
  border: 1px solid color-mix(in oklab, var(--line-subtle) 82%, transparent);
  background: color-mix(in oklab, var(--surface-1) 48%, transparent);
  color: var(--text-soft);
  font-weight: 600;
  backdrop-filter: blur(14px);
}

.btn-secondary:hover {
  transform: translateY(-1px);
  color: var(--text);
  border-color: rgba(var(--jade-accent-rgb), 0.28);
  background: color-mix(in oklab, var(--surface-1) 70%, transparent);
}

.hero-meta {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-wrap: wrap;
  gap: 0;
  margin: var(--hero-meta-offset) 0 0;
  font-size: 12px;
  line-height: 1.4;
  letter-spacing: 0.02em;
  color: var(--text-dim);
}

.hero-meta strong {
  color: var(--jade-accent);
  font-weight: 700;
  font-size: 13px;
}

@media (max-width: 640px) {
  .home-page {
    min-height: 100svh;
    padding: 68px 14px 26px;
  }

  .hero-main {
    width: 100%;
    --hero-title-size: 42px;
    --hero-accent-size: 34px;
    --hero-title-tracking: 0.09em;
    --hero-accent-tracking: 0.14em;
    --hero-title-gap: 0.18em;
    --hero-copy-gap: 22px;
    --hero-cta-offset: 14px;
    --hero-meta-offset: 12px;
  }

  .hero-panel {
    min-height: 0;
  }

  .hero-sub {
    max-width: 260px;
    font-size: 13px;
    line-height: 1.6;
  }

  .cta-group {
    flex-direction: column;
    width: 100%;
  }

  .btn-primary-base,
  .btn-secondary {
    width: 100%;
    min-width: 0;
  }

  .hero-meta {
    max-width: 360px;
  }
}
</style>
