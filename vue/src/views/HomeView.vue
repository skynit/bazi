<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import {
  ShaderMount,
  grainGradientFragmentShader
} from '@paper-design/shaders'
import {
  type WuxingKey,
  type ShaderThemeMode,
  wuxingThemes,
  createShaderUniforms
} from '../composables/useWuxingThemes'

const router = useRouter()
const authStore = useAuthStore()

const savedChartId = ref<number | null>(null)
const currentYongshen = ref<WuxingKey>('mu')
const mounted = ref(false)
const shaderHost = ref<HTMLDivElement | null>(null)
const themeMode = ref<ShaderThemeMode>(document.documentElement.classList.contains('dark') ? 'dark' : 'light')

const wuxingCycle: Array<{ key: WuxingKey; label: string }> = [
  { key: 'mu', label: '木' },
  { key: 'huo', label: '火' },
  { key: 'tu', label: '土' },
  { key: 'jin', label: '金' },
  { key: 'shui', label: '水' }
]

const homeThemeStyles = computed(() => {
  const key = currentYongshen.value
  const theme = wuxingThemes[key]
  return {
    '--jade-accent-rgb': theme.accentRgb,
    '--jade-accent': theme.accentHex,
    '--jade-accent-dark': theme.accentDark,
    '--jade-button-text': theme.buttonText
  }
})

let shaderMount: ShaderMount | null = null
let themeObserver: MutationObserver | null = null

onMounted(async () => {
  if (authStore.isLoggedIn() && !authStore.user) {
    await authStore.fetchMe().catch(() => { })
  }
  const saved = localStorage.getItem('bazi_last_birth')
  if (saved) {
    try { savedChartId.value = JSON.parse(saved).chartId || null } catch { }
  }
  setTimeout(() => mounted.value = true, 100)

  if (shaderHost.value) {
    shaderMount = new ShaderMount(
      shaderHost.value,
      grainGradientFragmentShader,
      createShaderUniforms('grainGradient', currentYongshen.value, undefined, themeMode.value),
      { alpha: true, antialias: true },
      2,
      0,
      Math.min(window.devicePixelRatio || 1, 2),
      undefined
    )
  }

  themeObserver = new MutationObserver(() => {
    themeMode.value = document.documentElement.classList.contains('dark') ? 'dark' : 'light'
    shaderMount?.setUniforms(createShaderUniforms('grainGradient', currentYongshen.value, undefined, themeMode.value))
  })
  themeObserver.observe(document.documentElement, { attributes: true, attributeFilter: ['class'] })
})

onUnmounted(() => {
  themeObserver?.disconnect()
  themeObserver = null
  shaderMount?.dispose()
  shaderMount = null
})

function startChart() { router.push('/chart/new') }
function continueChart() { if (savedChartId.value) router.push(`/chart/${savedChartId.value}`) }
function switchYongshen(key: WuxingKey) {
  currentYongshen.value = key
  shaderMount?.setUniforms(createShaderUniforms('grainGradient', key, undefined, themeMode.value))
}
</script>

<template>
  <div ref="shaderHost" class="home-grain-shader" aria-hidden="true"></div>
  <div class="home-page" :class="{ visible: mounted }" :style="homeThemeStyles">

    <!-- 前景内容区域 -->
    <main class="hero-main">
      <div class="title-block">
        <div class="eyebrow">
          <span class="eyebrow-line"></span>
          Bazi · Ziwei Astral System
          <span class="eyebrow-line"></span>
        </div>
        <h1 class="hero-title">
          天干地支<br />
          <span class="title-accent">推演</span>
        </h1>
      </div>

      <div class="cta-group">
        <button @click="startChart" class="btn-primary-base">
          开始排盘
        </button>
        <button v-if="savedChartId" @click="continueChart" class="btn-secondary">
          继续上次
        </button>
      </div>

      <div class="element-rail" aria-label="五行切换">
        <button v-for="item in wuxingCycle" :key="item.key" type="button"
          :class="{ active: currentYongshen === item.key }" @click="switchYongshen(item.key)">
          {{ item.label }}
        </button>
      </div>
    </main>
  </div>
</template>

<style scoped>
/* =============================================
   页面容器
   ============================================= */
.home-page {
  position: relative;
  min-height: 100vh;
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding: clamp(64px, 9vh, 92px) 24px 72px;
  background: transparent;
  color: var(--text);
  opacity: 0;
  transform: translateY(-2px);
  transition: opacity 1.2s cubic-bezier(0.16, 1, 0.3, 1), transform 1.2s cubic-bezier(0.16, 1, 0.3, 1);
}

.home-page.visible {
  opacity: 1;
  transform: translateY(0);
}

/* .home-page::after {
  content: "";
  position: absolute;
  inset: 0;
  background:
    radial-gradient(ellipse at 22% 4%, transparent 0%, rgba(0, 0, 0, .02) 40%, rgba(0, 0, 0, .18) 76%, rgba(0, 0, 0, .5) 100%),
    linear-gradient(90deg, rgba(0, 0, 0, 0) 0%, rgba(0, 0, 0, .04) 58%, rgba(0, 0, 0, .5) 100%),
    linear-gradient(180deg, rgba(0, 0, 0, 0) 0%, rgba(0, 0, 0, .03) 44%, rgba(0, 0, 0, .62) 100%);
  pointer-events: none;
  z-index: -5;
} */

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
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  width: min(920px, 100%);
  padding: 0 20px;
  gap: 24px;
  margin-top: clamp(10px, 2.4vh, 30px);
  transform: translateZ(0);
}

.title-block {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 22px;
}

.eyebrow {
  display: flex;
  align-items: center;
  gap: 14px;
  font-size: 10px;
  font-family: var(--font-mono), monospace;
  letter-spacing: 0.45em;
  text-transform: uppercase;
  color: color-mix(in oklab, var(--text) 48%, transparent);
}

.eyebrow-line {
  display: block;
  width: 28px;
  height: 1px;
  background: rgba(var(--jade-accent-rgb), .18);
}

.hero-title {
  font-family: var(--font-serif), 'Songti SC', serif;
  font-size: clamp(3.4rem, 6.25vw, 6.2rem);
  font-weight: 400;
  letter-spacing: 0.2em;
  color: var(--text);
  margin: 0;
  line-height: 1.26;
  padding-left: 0.2em;
  text-shadow:
    0 2px 10px color-mix(in oklab, var(--bg) 54%, transparent),
    0 20px 60px color-mix(in oklab, var(--bg) 58%, transparent);
}

.title-accent {
  display: block;
  margin-top: 10px;
  color: rgba(var(--jade-accent-rgb), .88);
  text-shadow:
    0 0 20px rgba(var(--jade-accent-rgb), .16),
    0 16px 54px rgba(0, 0, 0, .78);
}

.hero-sub {
  max-width: 760px;
  font-family: var(--font-serif), serif;
  font-size: clamp(1.1rem, 2.35vw, 2rem);
  color: var(--text);
  letter-spacing: 0.035em;
  margin: 0;
  text-shadow: 0 3px 26px rgba(0, 0, 0, .76);
}

.cta-group {
  display: flex;
  gap: 22px;
  flex-wrap: wrap;
  justify-content: center;
  margin-top: 6px;
}

.btn-primary-base {
  width: 168px;
  padding: 13px 0;
  border: none;
  border-radius: 50px;
  background: linear-gradient(135deg, var(--jade-accent) 0%, var(--jade-accent-dark) 100%);
  color: var(--jade-button-text);
  font-size: 0.92rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  cursor: pointer;
  box-shadow:
    0 12px 38px rgba(var(--jade-accent-rgb), .42),
    inset 0 1px 0 rgba(255, 255, 255, .22);
  transition: all 0.4s cubic-bezier(0.16, 1, 0.3, 1);
}

.btn-primary-base:hover {
  transform: translateY(-2px);
}

.btn-secondary {
  width: 168px;
  padding: 13px 0;
  background: rgba(255, 255, 255, .02);
  color: var(--text-muted);
  font-size: 0.92rem;
  font-weight: 500;
  letter-spacing: 0.08em;
  border: 1px solid var(--line-strong);
  border-radius: 50px;
  cursor: pointer;
  backdrop-filter: blur(12px);
  transition: all 0.3s ease;
}

.btn-secondary:hover {
  color: var(--text);
  border-color: rgba(var(--jade-accent-rgb), .36);
  background: rgba(255, 255, 255, .05);
}

.element-rail {
  display: flex;
  justify-content: center;
  gap: 12px;
  margin-top: 2px;
  opacity: .55;
}

.element-rail button {
  width: 28px;
  height: 28px;
  padding: 0;
  border: 1px solid transparent;
  border-radius: 50%;
  background: transparent;
  color: var(--text-muted);
  font-family: var(--font-serif), serif;
  font-size: .85rem;
  cursor: pointer;
  transition: color .3s ease, border-color .3s ease, background .3s ease;
}

.element-rail button:hover,
.element-rail button.active {
  border-color: rgba(var(--jade-accent-rgb), .3);
  background: var(--glass-bg);
  color: rgba(var(--jade-accent-rgb), .9);
}

/* =============================================
   响应式
   ============================================= */
@media (max-width: 1024px) {
  .hero-main {
    margin-top: 0;
  }
}

@media (max-width: 640px) {
  .home-page {
    min-height: 100svh;
    padding: 58px 18px 42px;
  }

  .hero-main {
    padding: 0;
    margin-top: 0;
    gap: 22px;
  }

  .title-block {
    gap: 18px;
  }

  .hero-title {
    font-size: clamp(2.55rem, 12vw, 3.55rem);
    letter-spacing: 0.1em;
    padding-left: 0.1em;
    line-height: 1.28;
  }

  .hero-sub {
    max-width: 340px;
    font-size: 1rem;
    line-height: 1.45;
  }

  .eyebrow {
    gap: 10px;
    font-size: 8px;
    letter-spacing: .18em;
  }

  .cta-group {
    flex-direction: column;
    gap: 12px;
  }

  .btn-primary-base,
  .btn-secondary {
    width: min(260px, calc(100vw - 48px));
  }

  .element-rail {
    gap: 8px;
  }
}
</style>
