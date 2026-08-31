<script setup lang="ts">
import { computed, ref, watch, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useRecentChartStore } from '../stores/recentChart'
import { useThemeStore } from '../stores/theme'
import { ShaderMount, grainGradientFragmentShader } from '@paper-design/shaders'
import { type ShaderThemeMode, createShaderUniforms } from '../composables/useWuxingThemes'

const router = useRouter()
const authStore = useAuthStore()
const recentChartStore = useRecentChartStore()
const themeStore = useThemeStore()

const savedChartId = computed(() => recentChartStore.chartId)
const mounted = ref(false)
const shaderHost = ref<HTMLDivElement | null>(null)
const themeMode = ref<ShaderThemeMode>(
  document.documentElement.classList.contains('dark') ? 'dark' : 'light',
)
const currentElementLabel = computed(
  () =>
    ({
      mu: '木',
      huo: '火',
      tu: '土',
      jin: '金',
      shui: '水',
    })[themeStore.elementTheme] || themeStore.elementTheme,
)

let shaderMount: ShaderMount | null = null
let themeObserver: MutationObserver | null = null

onMounted(async () => {
  window.scrollTo({ top: 0, left: 0, behavior: 'auto' })

  if (authStore.isLoggedIn() && !authStore.user) {
    await authStore.fetchMe().catch(() => {})
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
      undefined,
    )
  }

  themeObserver = new MutationObserver(() => {
    themeMode.value = document.documentElement.classList.contains('dark') ? 'dark' : 'light'
    shaderMount?.setUniforms(
      createShaderUniforms('grainGradient', themeStore.elementTheme, undefined, themeMode.value),
    )
  })
  themeObserver.observe(document.documentElement, { attributes: true, attributeFilter: ['class'] })
})

watch(
  () => themeStore.elementTheme,
  (key) => {
    shaderMount?.setUniforms(createShaderUniforms('grainGradient', key, undefined, themeMode.value))
  },
)

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

          <p class="hero-sub">一页看八字、流年与紫微。</p>

          <div class="cta-group">
            <button @click="startChart" class="btn-primary-base">开始排盘</button>
            <button v-if="savedChartId" @click="continueChart" class="btn-secondary">
              继续上次
            </button>
          </div>

          <p class="hero-meta" aria-label="当前主题">
            当前五行主题：<strong>{{ currentElementLabel }}</strong>
          </p>
        </div>
        <a class="scroll-hint" href="#home-capabilities">
          <span>向下了解</span>
          <span class="scroll-hint-chevron" aria-hidden="true">▾</span>
        </a>
      </section>
    </main>

    <div class="home-sections">
      <!-- 核心能力 -->
      <section class="section" id="home-capabilities" aria-labelledby="sec-capabilities">
        <header class="sec-head">
          <span class="sec-eyebrow">核心能力</span>
          <h2 class="sec-title" id="sec-capabilities">排盘、推演、问卦，各守其法</h2>
        </header>
        <div class="cap-grid">
          <div class="cap-item">
            <h3 class="cap-name">八字排盘</h3>
            <p class="cap-desc">
              输入出生年月日时，排出四柱干支、十神藏干与大运流年，命盘结构一页览尽。
            </p>
          </div>
          <div class="cap-item">
            <h3 class="cap-name">紫微斗数</h3>
            <p class="cap-desc">以出生时刻安星定盘，十二宫位与诸星庙旺并陈，命身格局可查可考。</p>
          </div>
          <div class="cap-item">
            <h3 class="cap-name">运势查询</h3>
            <p class="cap-desc">
              日、周、月三档运势，结合日主强弱与流年干支逐层推演，吉凶宜忌各有出处。
            </p>
          </div>
          <div class="cap-item">
            <h3 class="cap-name">卜易问卦</h3>
            <p class="cap-desc">一事一占，起卦、变爻与卦辞并出，临事抉择之际提供一份传统参照。</p>
          </div>
        </div>
      </section>

      <!-- 典籍依据 -->
      <section class="section" aria-labelledby="sec-classics">
        <header class="sec-head">
          <span class="sec-eyebrow">典籍依据</span>
          <h2 class="sec-title" id="sec-classics">论断有本，非凭空之言</h2>
        </header>
        <ol class="book-list">
          <li class="book-row">
            <span class="book-no">01</span>
            <span class="book-name">《渊海子平》</span>
            <span class="book-desc">子平法之渊薮，格局、用神之论多本于此。</span>
          </li>
          <li class="book-row">
            <span class="book-no">02</span>
            <span class="book-name">《三命通会》</span>
            <span class="book-desc">万民英集大成之作，干支、神煞、纳音体系详备。</span>
          </li>
          <li class="book-row">
            <span class="book-no">03</span>
            <span class="book-name">《滴天髓》</span>
            <span class="book-desc">专论旺衰喜忌与气势流通，为推断强弱之圭臬。</span>
          </li>
          <li class="book-row">
            <span class="book-no">04</span>
            <span class="book-name">《穷通宝鉴》</span>
            <span class="book-desc">按月令论五行调候，寒暖燥湿各有宜忌。</span>
          </li>
        </ol>
      </section>

      <!-- 使用流程 -->
      <section class="section" aria-labelledby="sec-flow">
        <header class="sec-head">
          <span class="sec-eyebrow">使用流程</span>
          <h2 class="sec-title" id="sec-flow">三步成盘，循序而读</h2>
        </header>
        <div class="step-grid">
          <div class="step-item">
            <span class="step-no">01</span>
            <h3 class="step-name">填写生辰</h3>
            <p class="step-desc">输入公历出生年月日时，自动换算干支历法，无需自行查表。</p>
          </div>
          <div class="step-item">
            <span class="step-no">02</span>
            <h3 class="step-name">生成命盘</h3>
            <p class="step-desc">四柱、大运、流年即刻排定，命盘存档后可随时回看比对。</p>
          </div>
          <div class="step-item">
            <span class="step-no">03</span>
            <h3 class="step-name">逐层解读</h3>
            <p class="step-desc">从格局喜忌到流年细断，逐层展开，所引据典可溯可查。</p>
          </div>
        </div>
      </section>

      <!-- 底部 CTA -->
      <section class="section cta-section" aria-label="开始排盘">
        <p class="cta-line">生辰既定，格局自见。</p>
        <div class="cta-group cta-group-end">
          <button @click="startChart" class="btn-primary-base">开始排盘</button>
          <router-link
            v-if="!authStore.isLoggedIn()"
            to="/register"
            class="btn-secondary btn-as-link"
          >
            注册账号
          </router-link>
          <button v-else-if="savedChartId" @click="continueChart" class="btn-secondary">
            继续上次
          </button>
        </div>
      </section>
    </div>
  </div>
</template>

<style scoped>
.home-page {
  position: relative;
  padding: 0 24px;
  background: transparent;
  color: var(--text);
  opacity: 0;
  transform: translateY(-2px);
  transition:
    opacity 1s cubic-bezier(0.16, 1, 0.3, 1),
    transform 1s cubic-bezier(0.16, 1, 0.3, 1);
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
  margin: 0 auto;
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
  position: relative;
  display: grid;
  place-items: center;
  align-content: center;
  justify-items: center;
  gap: 0;
  min-height: calc(100svh - 80px);
  text-align: center;
}

.hero-copy {
  display: grid;
  align-content: center;
  justify-items: center;
  gap: var(--hero-copy-gap);
  width: 100%;
  padding-bottom: 56px;
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
  cursor: pointer;
}

.btn-primary-base:hover {
  transform: translateY(-1px);
  box-shadow:
    0 18px 36px rgba(var(--jade-accent-rgb), 0.28),
    inset 0 1px 0 rgba(255, 255, 255, 0.2);
}

.btn-primary-base:active {
  transform: translateY(0);
  box-shadow:
    0 8px 20px rgba(var(--jade-accent-rgb), 0.2),
    inset 0 1px 0 rgba(255, 255, 255, 0.2);
}

.btn-secondary {
  border: 1px solid color-mix(in oklab, var(--line-subtle) 82%, transparent);
  background: color-mix(in oklab, var(--surface-1) 48%, transparent);
  color: var(--text-soft);
  font-weight: 600;
  backdrop-filter: blur(14px);
  cursor: pointer;
}

.btn-secondary:hover {
  transform: translateY(-1px);
  color: var(--text);
  border-color: rgba(var(--jade-accent-rgb), 0.28);
  background: color-mix(in oklab, var(--surface-1) 70%, transparent);
}

.btn-secondary:active {
  transform: translateY(0);
}

.btn-as-link {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  text-decoration: none;
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

.scroll-hint {
  position: absolute;
  bottom: 22px;
  left: 50%;
  transform: translateX(-50%);
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: var(--fs-2xs);
  letter-spacing: 0.14em;
  color: var(--text-dim);
  text-decoration: none;
  transition: color 200ms ease;
}

.scroll-hint:hover {
  color: var(--accent);
}

.scroll-hint-chevron {
  animation: hint-bob 2.4s ease-in-out infinite;
}

@keyframes hint-bob {
  0%,
  100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(3px);
  }
}

/* ── 滚动内容区 ── */
.home-sections {
  position: relative;
  z-index: 10;
  width: min(880px, 100%);
  margin: 0 auto;
  padding-bottom: 72px;
}

.section {
  padding: 56px 0 60px;
  border-top: 1px solid var(--line-subtle);
}

.sec-head {
  display: grid;
  gap: 12px;
  margin-bottom: 36px;
}

.sec-eyebrow {
  font-size: var(--fs-2xs);
  letter-spacing: var(--tracking-meta);
  color: var(--jade-accent);
  font-weight: 600;
}

.sec-title {
  margin: 0;
  font-family: var(--font-serif);
  font-size: clamp(22px, 2.6vw, 28px);
  font-weight: 700;
  line-height: var(--lh-snug);
  letter-spacing: 0.04em;
  color: var(--text);
}

/* 核心能力 */
.cap-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  column-gap: 48px;
  row-gap: 36px;
}

.cap-item {
  border-top: 1px solid var(--line-strong);
  padding-top: 16px;
}

.cap-name {
  margin: 0 0 8px;
  font-family: var(--font-serif);
  font-size: var(--fs-lg);
  font-weight: 600;
  letter-spacing: 0.06em;
  color: var(--text);
}

.cap-desc {
  margin: 0;
  font-size: var(--fs-xs);
  line-height: var(--lh-relaxed);
  color: var(--text-soft);
}

/* 典籍依据 */
.book-list {
  list-style: none;
  margin: 0;
  padding: 0;
}

.book-row {
  display: grid;
  grid-template-columns: 44px minmax(150px, 210px) 1fr;
  align-items: baseline;
  gap: 20px;
  padding: 16px 0;
  border-top: 1px solid var(--line-subtle);
}

.book-row:last-child {
  border-bottom: 1px solid var(--line-subtle);
}

.book-no {
  font-family: var(--font-mono);
  font-size: var(--fs-2xs);
  letter-spacing: 0.08em;
  color: var(--text-dim);
}

.book-name {
  font-family: var(--font-serif);
  font-size: var(--fs-md);
  font-weight: 600;
  letter-spacing: 0.02em;
  color: var(--text);
}

.book-desc {
  font-size: var(--fs-xs);
  line-height: var(--lh-relaxed);
  color: var(--text-soft);
}

/* 使用流程 */
.step-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 40px;
}

.step-item {
  border-top: 1px solid var(--line-strong);
  padding-top: 16px;
}

.step-no {
  display: block;
  font-family: var(--font-mono);
  font-size: var(--fs-2xs);
  letter-spacing: 0.12em;
  color: var(--jade-accent);
  margin-bottom: 10px;
}

.step-name {
  margin: 0 0 8px;
  font-family: var(--font-serif);
  font-size: var(--fs-lg);
  font-weight: 600;
  letter-spacing: 0.06em;
  color: var(--text);
}

.step-desc {
  margin: 0;
  font-size: var(--fs-xs);
  line-height: var(--lh-relaxed);
  color: var(--text-soft);
}

/* 底部 CTA */
.cta-section {
  text-align: center;
  padding-bottom: 80px;
}

.cta-line {
  margin: 0;
  font-family: var(--font-serif);
  font-size: clamp(20px, 2.4vw, 26px);
  font-weight: 600;
  letter-spacing: 0.08em;
  color: var(--text);
}

.cta-group-end {
  margin-top: 28px;
}

@media (max-width: 640px) {
  .home-page {
    padding: 0 18px;
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
    min-height: calc(100svh - 76px);
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

  .cta-group-end {
    align-items: stretch;
  }

  .btn-primary-base,
  .btn-secondary {
    width: 100%;
    min-width: 0;
  }

  .hero-meta {
    max-width: 360px;
  }

  .section {
    padding: 44px 0 48px;
  }

  .sec-head {
    margin-bottom: 28px;
  }

  .cap-grid {
    grid-template-columns: 1fr;
    row-gap: 28px;
  }

  .book-row {
    grid-template-columns: 36px 1fr;
    row-gap: 4px;
    gap: 16px;
  }

  .book-desc {
    grid-column: 2;
  }

  .step-grid {
    grid-template-columns: 1fr;
    gap: 28px;
  }
}
</style>
