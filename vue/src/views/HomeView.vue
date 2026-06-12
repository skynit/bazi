<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import {
  GrainGradientShapes,
  ShaderFitOptions,
  ShaderMount,
  getShaderColorFromString,
  getShaderNoiseTexture,
  grainGradientFragmentShader,
  type ShaderMountUniforms
} from '@paper-design/shaders'

type WuxingKey = 'mu' | 'huo' | 'tu' | 'jin' | 'shui'

const router = useRouter()
const authStore = useAuthStore()

const savedChartId = ref<number | null>(null)
const currentYongshen = ref<WuxingKey>('mu')
const mounted = ref(false)
const shaderHost = ref<HTMLDivElement | null>(null)

const wuxingCycle: Array<{ key: WuxingKey; label: string }> = [
  { key: 'mu', label: '木' },
  { key: 'huo', label: '火' },
  { key: 'tu', label: '土' },
  { key: 'jin', label: '金' },
  { key: 'shui', label: '水' }
]

type WuxingTheme = {
  accentRgb: string
  accentHex: string
  accentDark: string
  buttonText: string
  shaderBase: string
  shaderGlow: string
}

function createTheme(theme: WuxingTheme): WuxingTheme {
  return theme
}

const wuxingThemes: Record<WuxingKey, WuxingTheme> = {
  mu: createTheme({
    accentRgb: '34, 211, 153',
    accentHex: '#34d399',
    accentDark: '#059669',
    buttonText: '#00140e',
    shaderBase: '#147b33',
    shaderGlow: '#43dfabcf'
  }),
  huo: createTheme({
    accentRgb: '251, 113, 133',
    accentHex: '#fb7185',
    accentDark: '#e11d48',
    buttonText: '#190005',
    shaderBase: '#7f1d1d',
    shaderGlow: '#fb7185cf'
  }),
  tu: createTheme({
    accentRgb: '252, 211, 77',
    accentHex: '#fcd34d',
    accentDark: '#d97706',
    buttonText: '#1a1000',
    shaderBase: '#7c4f08',
    shaderGlow: '#fcd34dcf'
  }),
  jin: createTheme({
    accentRgb: '226, 232, 240',
    accentHex: '#e2e8f0',
    accentDark: '#94a3b8',
    buttonText: '#030404',
    shaderBase: '#334155',
    shaderGlow: '#e2e8f0cf'
  }),
  shui: createTheme({
    accentRgb: '34, 211, 238',
    accentHex: '#22d3ee',
    accentDark: '#2563eb',
    buttonText: '#001116',
    shaderBase: '#075985',
    shaderGlow: '#22d3eecf'
  })
}

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

const shaderNoiseTexture = getShaderNoiseTexture()
let shaderMount: ShaderMount | null = null

function createShaderUniforms(theme: WuxingTheme): ShaderMountUniforms {
  return {
    u_colorBack: getShaderColorFromString('#000000'),
    u_colors: [
      getShaderColorFromString(theme.shaderBase),
      getShaderColorFromString(theme.shaderGlow),
      getShaderColorFromString('#000000'),
      getShaderColorFromString('#000000')
    ],
    u_colorsCount: 4,
    u_softness: 1,
    u_intensity: 1,
    u_noise: 0,
    u_shape: GrainGradientShapes.sphere,
    u_noiseTexture: shaderNoiseTexture,
    u_fit: ShaderFitOptions.cover,
    u_scale: 0.88,
    u_rotation: 360,
    u_originX: 0.5,
    u_originY: 0.5,
    u_offsetX: 0.06,
    u_offsetY: 0,
    u_worldWidth: 0,
    u_worldHeight: 0
  }
}

onMounted(async () => {
  if (authStore.isLoggedIn() && !authStore.user) {
    await authStore.fetchMe().catch(() => {})
  }
  const saved = localStorage.getItem('bazi_last_birth')
  if (saved) {
    try { savedChartId.value = JSON.parse(saved).chartId || null } catch {}
  }
  setTimeout(() => mounted.value = true, 100)

  if (shaderHost.value) {
    shaderMount = new ShaderMount(
      shaderHost.value,
      grainGradientFragmentShader,
      createShaderUniforms(wuxingThemes[currentYongshen.value]),
      { alpha: true, antialias: true },
      2,
      0,
      1.5,
      3200 * 1800
    )
  }
})

onUnmounted(() => {
  shaderMount?.dispose()
  shaderMount = null
})

function startChart() { router.push('/chart/new') }
function continueChart() { if (savedChartId.value) router.push(`/chart/${savedChartId.value}`) }
function switchYongshen(key: WuxingKey) {
  currentYongshen.value = key
  shaderMount?.setUniforms(createShaderUniforms(wuxingThemes[key]))
}
</script>

<template>
  <div class="home-page" :class="{ visible: mounted }" :style="homeThemeStyles">
    <div ref="shaderHost" class="home-grain-shader" aria-hidden="true"></div>

    <!-- 前景内容区域 -->
    <main class="hero-main">
      <div class="title-block">
        <div class="eyebrow">
          <span class="eyebrow-line"></span>
          Bazi · Ziwei Astral System
          <span class="eyebrow-line"></span>
        </div>
        <h1 class="hero-title">
          以天干地支<br />
          <span class="title-accent">推演命局流转</span>
        </h1>
        <p class="hero-sub">See Further · Think Deeper · Act Faster</p>
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
        <button
          v-for="item in wuxingCycle"
          :key="item.key"
          type="button"
          :class="{ active: currentYongshen === item.key }"
          @click="switchYongshen(item.key)"
        >
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
  overflow: hidden;
  padding: clamp(64px, 9vh, 92px) 24px 72px;
  background: #000;
  color: #f4f4f5;
  opacity: 0;
  transform: translateY(-2px);
  transition: opacity 1.2s cubic-bezier(0.16, 1, 0.3, 1), transform 1.2s cubic-bezier(0.16, 1, 0.3, 1);
}

.home-page.visible {
  opacity: 1;
  transform: translateY(0);
}

.home-page::after {
  content: "";
  position: absolute;
  inset: 0;
  background:
    radial-gradient(ellipse at 22% 4%, transparent 0%, rgba(0,0,0,.02) 40%, rgba(0,0,0,.18) 76%, rgba(0,0,0,.5) 100%),
    linear-gradient(90deg, rgba(0,0,0,0) 0%, rgba(0,0,0,.04) 58%, rgba(0,0,0,.5) 100%),
    linear-gradient(180deg, rgba(0,0,0,0) 0%, rgba(0,0,0,.03) 44%, rgba(0,0,0,.62) 100%);
  pointer-events: none;
  z-index: 5;
}

.home-grain-shader {
  position: absolute;
  inset: 0;
  z-index: 0;
  display: block;
  pointer-events: none;
  mix-blend-mode: normal;
  opacity: 1;
  filter: contrast(1.04) saturate(1.02);
}

.home-grain-shader :deep(canvas) {
  display: block;
  width: 100% !important;
  height: 100% !important;
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
  color: rgba(235,255,248,.38);
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
  color: rgba(255,255,255,.99);
  margin: 0;
  line-height: 1.26;
  padding-left: 0.2em;
  text-shadow:
    0 2px 10px rgba(0,0,0,.66),
    0 20px 60px rgba(0,0,0,.62);
}

.title-accent {
  display: block;
  margin-top: 10px;
  color: rgba(var(--jade-accent-rgb), .88);
  text-shadow:
    0 0 20px rgba(var(--jade-accent-rgb), .16),
    0 16px 54px rgba(0,0,0,.78);
}

.hero-sub {
  max-width: 760px;
  font-family: var(--font-serif), serif;
  font-size: clamp(1.1rem, 2.35vw, 2rem);
  color: rgba(245,255,251,.92);
  letter-spacing: 0.035em;
  margin: 0;
  text-shadow: 0 3px 26px rgba(0,0,0,.76);
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
    inset 0 1px 0 rgba(255,255,255,.22);
  transition: all 0.4s cubic-bezier(0.16, 1, 0.3, 1);
}

.btn-primary-base:hover {
  transform: translateY(-2px);
}

.btn-secondary {
  width: 168px;
  padding: 13px 0;
  background: rgba(255,255,255,.02);
  color: rgba(235,255,249,.72);
  font-size: 0.92rem;
  font-weight: 500;
  letter-spacing: 0.08em;
  border: 1px solid rgba(217,255,241,.16);
  border-radius: 50px;
  cursor: pointer;
  backdrop-filter: blur(12px);
  transition: all 0.3s ease;
}

.btn-secondary:hover {
  color: #ffffff;
  border-color: rgba(var(--jade-accent-rgb), .36);
  background: rgba(255,255,255,.05);
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
  color: rgba(229,255,246,.46);
  font-family: var(--font-serif), serif;
  font-size: .85rem;
  cursor: pointer;
  transition: color .3s ease, border-color .3s ease, background .3s ease;
}

.element-rail button:hover,
.element-rail button.active {
  border-color: rgba(var(--jade-accent-rgb), .3);
  background: rgba(0,0,0,.18);
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
