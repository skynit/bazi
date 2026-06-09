<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import CosmicGrainBackground from '../components/CosmicGrainBackground.vue'

type WuxingKey = 'mu' | 'huo' | 'tu' | 'jin' | 'shui'

const router = useRouter()
const authStore = useAuthStore()

const savedChartId = ref<number | null>(null)
const currentYongshen = ref<WuxingKey>('mu')
const mounted = ref(false)

const earthlyBranches = ['子', '丑', '寅', '卯', '辰', '巳', '午', '未', '申', '酉', '戌', '亥']

/**
 * 五行主题配置
 * ⚡️ 坐标重构：将光源精准下放到圆盘底弧（Y: 90% - 98%），使其在屏幕顶部完美显现
 * ⚡️ 错位混色：引入邻近色（如木局配蔚蓝，火局配妖姬紫），打造极光般高级的数字流动感
 */
const wuxingThemes: Record<string, {
  accentText: string
  lineColor: string
  glyphColor: string
  domeBg: string
  btnClass: string
  dotColor: string
}> = {
  mu: {
    accentText: 'text-emerald-400 font-bold tracking-widest drop-shadow-[0_0_12px_rgba(52,211,153,0.6)]',
    lineColor: 'stroke-emerald-400/50',
    glyphColor: 'fill-emerald-300 font-medium drop-shadow-[0_0_4px_rgba(52,211,153,0.4)]',
    domeBg: `
      radial-gradient(750px circle at 30% 95%, #00ffaa 0%, rgba(0,255,170,0.3) 50%, transparent 100%),
      radial-gradient(650px circle at 70% 92%, #0ea5e9 0%, rgba(14,165,233,0.2) 50%, transparent 100%),
      radial-gradient(750px circle at 50% 98%, #10b981 0%, rgba(16,185,129,0.4) 50%, transparent 100%)
    `,
    btnClass: 'bg-emerald-400 text-zinc-950 font-bold shadow-[0_10px_35px_rgba(0,255,170,0.5)] hover:bg-emerald-300 hover:scale-[1.02]',
    dotColor: 'bg-emerald-400 shadow-[0_0_10px_rgba(0,255,170,1)]'
  },
  huo: {
    accentText: 'text-rose-400 font-bold tracking-widest drop-shadow-[0_0_12px_rgba(251,113,133,0.6)]',
    lineColor: 'stroke-rose-400/50',
    glyphColor: 'fill-rose-300 font-medium',
    domeBg: `
      radial-gradient(750px circle at 25% 95%, #ff3366 0%, rgba(255,51,102,0.3) 50%, transparent 100%),
      radial-gradient(650px circle at 75% 92%, #a855f7 0%, rgba(168,85,247,0.2) 50%, transparent 100%),
      radial-gradient(750px circle at 50% 98%, #f43f5e 0%, rgba(244,63,94,0.4) 50%, transparent 100%)
    `,
    btnClass: 'bg-rose-400 text-zinc-950 font-bold shadow-[0_10px_35px_rgba(251,113,133,0.5)] hover:bg-rose-300',
    dotColor: 'bg-rose-400 shadow-[0_0_10px_rgba(255,51,102,1)]'
  },
  tu: {
    accentText: 'text-amber-400 font-bold tracking-widest drop-shadow-[0_0_12px_rgba(252,211,77,0.6)]',
    lineColor: 'stroke-amber-400/50',
    glyphColor: 'fill-amber-300 font-medium',
    domeBg: `
      radial-gradient(750px circle at 30% 95%, #ffaa00 0%, rgba(255,170,0,0.3) 50%, transparent 100%),
      radial-gradient(650px circle at 70% 92%, #f97316 0%, rgba(249,115,22,0.2) 50%, transparent 100%),
      radial-gradient(750px circle at 50% 98%, #d97706 0%, rgba(217,119,6,0.4) 50%, transparent 100%)
    `,
    btnClass: 'bg-amber-400 text-zinc-950 font-bold shadow-[0_10px_35px_rgba(252,211,77,0.5)] hover:bg-amber-300',
    dotColor: 'bg-amber-400 shadow-[0_0_10px_rgba(255,170,0,1)]'
  },
  jin: {
    accentText: 'text-zinc-100 font-bold tracking-widest drop-shadow-[0_0_8px_rgba(255,255,255,0.5)]',
    lineColor: 'stroke-zinc-300/60',
    glyphColor: 'fill-zinc-100 font-medium',
    domeBg: `
      radial-gradient(750px circle at 25% 95%, #ffffff 0%, rgba(255,255,255,0.4) 50%, transparent 100%),
      radial-gradient(650px circle at 75% 92%, #e4e4e7 0%, rgba(228,228,231,0.2) 50%, transparent 100%),
      radial-gradient(750px circle at 50% 98%, #a1a1aa 0%, rgba(161,161,170,0.3) 50%, transparent 100%)
    `,
    btnClass: 'bg-zinc-100 text-zinc-950 font-bold shadow-[0_10px_35px_rgba(244,244,245,0.4)] hover:bg-white',
    dotColor: 'bg-zinc-100 shadow-[0_0_10px_rgba(255,255,255,0.8)]'
  },
  shui: {
    accentText: 'text-cyan-400 font-bold tracking-widest drop-shadow-[0_0_12px_rgba(34,211,238,0.6)]',
    lineColor: 'stroke-cyan-400/50',
    glyphColor: 'fill-cyan-300 font-medium',
    domeBg: `
      radial-gradient(750px circle at 28% 95%, #00f0ff 0%, rgba(0,240,255,0.4) 50%, transparent 100%),
      radial-gradient(650px circle at 72% 92%, #3b82f6 0%, rgba(59,130,246,0.2) 50%, transparent 100%),
      radial-gradient(750px circle at 50% 98%, #1d4ed8 0%, rgba(29,78,216,0.3) 50%, transparent 100%)
    `,
    btnClass: 'bg-cyan-400 text-zinc-950 font-bold shadow-[0_10px_35px_rgba(34,211,238,0.5)] hover:bg-cyan-300',
    dotColor: 'bg-cyan-400 shadow-[0_0_10px_rgba(0,240,255,1)]'
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
})

function startChart() { router.push('/chart/new') }
function continueChart() { if (savedChartId.value) router.push(`/chart/${savedChartId.value}`) }
</script>

<template>
  <div class="home-page" :class="{ visible: mounted }">
    <CosmicGrainBackground :yongshen="currentYongshen" />

    <div class="astrolabe-canopy-layer">
      <div
        class="astrolabe-dome"
        :style="{ '--dome-bg': wuxingThemes[currentYongshen]?.domeBg }"
      >
        <div class="sharp-blade-dots"></div>
        <svg viewBox="0 0 800 800" class="w-full h-full font-serif relative z-10">
          <g class="animate-[spin_240s_linear_infinite]" style="transform-origin:400px 400px">
            <circle cx="400" cy="400" r="385" :class="wuxingThemes[currentYongshen]?.lineColor" stroke-width="0.5" fill="none" />
            <g v-for="(branch, index) in earthlyBranches" :key="branch" :transform="`rotate(${index * 30} 400 400)`">
              <line x1="400" y1="775" x2="400" y2="785" :class="wuxingThemes[currentYongshen]?.lineColor" stroke-width="0.75" />
              <text x="400" y="758" text-anchor="middle" :class="wuxingThemes[currentYongshen]?.glyphColor" font-size="13">
                {{ branch }}
              </text>
            </g>
          </g>
        </svg>
      </div>
    </div>

    <main class="hero-main">
      <div class="status-badge">
        <span class="status-dot" :class="wuxingThemes[currentYongshen]?.dotColor"></span>
        <span class="uppercase tracking-widest text-zinc-400">
          SYSTEM STATUS: <span :class="wuxingThemes[currentYongshen]?.accentText">{{ currentYongshen }}局 ACTIVE</span>
        </span>
      </div>

      <div class="title-block">
        <div class="eyebrow">
          <span class="eyebrow-line"></span>
          Ziwei · Bazi Fullstack Reality
          <span class="eyebrow-line"></span>
        </div>
        <h1 class="hero-title">
          以数字之理<br />
          <span class="title-accent">解构命运之网</span>
        </h1>
        <p class="hero-sub">See Further · Think Deeper · Act Faster</p>
      </div>

      <div class="cta-group">
        <button @click="startChart" :class="wuxingThemes[currentYongshen]?.btnClass" class="btn-primary-base">
          启动命运解构
        </button>
        <button v-if="savedChartId" @click="continueChart" class="btn-secondary">
          继续上次
        </button>
      </div>
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
  overflow: hidden;
  background: #010201;
  color: #f4f4f5;
  opacity: 0;
  transform: translateY(-2px);
  transition: opacity 1.2s cubic-bezier(0.16, 1, 0.3, 1), transform 1.2s cubic-bezier(0.16, 1, 0.3, 1);
}

.home-page.visible {
  opacity: 1;
  transform: translateY(0);
}

.astrolabe-canopy-layer {
  position: absolute;
  inset: 0;
  z-index: 1;
  display: flex;
  justify-content: center;
  pointer-events: none;
  overflow: hidden;
}

.astrolabe-dome {
  width: 1500px;
  height: 1500px;
  position: absolute;
  top: -70vw;
  left: 50%;
  transform: translateX(-50%);
  border-radius: 50%;
  overflow: hidden;
  background: #010201; 
  box-shadow: inset 0 0 120px rgba(0, 0, 0, 1);
}

/* 光晕主体 */
.astrolabe-dome::before {
  content: '';
  position: absolute;
  inset: -10%; 
  background: var(--dome-bg);
  filter: blur(60px); /* 适度降低模糊半径，防止高饱和色彩被过度稀释 */
  opacity: 0.88; /* 增强色彩通透度 */
  z-index: 0;
  pointer-events: none;
  transition: background 1s ease;
}

/* ⚡️ 修复后的暗角融合：改为线性向下过渡，让顶部（虚无区）深邃，底弧（可见区）彻底释放光芒 */
.astrolabe-dome::after {
  content: '';
  position: absolute;
  inset: 0;
  background: linear-gradient(
    to bottom,
    rgba(1, 2, 1, 0.9) 0%,
    rgba(1, 2, 1, 0.4) 70%,
    transparent 100%
  );
  z-index: 1;
  pointer-events: none;
}

/* 颗粒纹理层 */
.sharp-blade-dots {
  position: absolute;
  inset: 0;
  z-index: 2;
  background-image: url("data:image/svg+xml,%3Csvg viewBox='0 0 100 100' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='razorGrain'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.99' numOctaves='1' result='noise'/%3E%3CfeColorMatrix type='matrix' values='-100 0 0 0 72 -100 0 0 0 72 -100 0 0 0 72 0 0 0 1 0'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23razorGrain)'/%3E%3C/svg%3E");
  background-size: 140px 140px;
  background-repeat: repeat;
  mix-blend-mode: multiply;
  opacity: 0.52;
  image-rendering: pixelated;
  pointer-events: none;
}

.hero-main {
  position: relative;
  z-index: 10;
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  padding: 40px 24px;
  max-width: 800px;
  gap: 48px;
  margin-top: 120px;
  transform: translateZ(0);
}

.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  padding: 6px 16px;
  background: rgba(0, 0, 0, 0.65);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 50px;
  font-size: 10px;
  font-family: monospace;
  backdrop-filter: blur(8px);
}

.status-dot {
  width: 5px;
  height: 5px;
  border-radius: 50%;
  animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
}

.title-block {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 20px;
}

.eyebrow {
  display: flex;
  align-items: center;
  gap: 14px;
  font-size: 10px;
  font-family: monospace;
  letter-spacing: 0.35em;
  text-transform: uppercase;
  color: #444446;
}

.eyebrow-line {
  display: block;
  width: 24px;
  height: 1px;
  background: rgba(63, 63, 70, 0.3);
}

.hero-title {
  font-family: var(--font-serif), 'Songti SC', serif;
  font-size: 3.8rem;
  font-weight: 300;
  letter-spacing: 0.25em;
  color: #f4f4f5;
  margin: 0;
  line-height: 1.4;
  padding-left: 0.25em;
  text-shadow: 0 4px 24px rgba(0, 0, 0, 0.95);
}

.title-accent {
  display: block;
  margin-top: 8px;
  color: #ffffff;
}

.hero-sub {
  font-size: 11px;
  color: #52525b;
  letter-spacing: 0.4em;
  margin: 0;
  font-family: monospace;
  text-transform: uppercase;
}

.btn-primary-base {
  width: 210px;
  padding: 14px 0;
  border: none;
  border-radius: 50px;
  font-size: 0.85rem;
  letter-spacing: 0.15em;
  cursor: pointer;
  transition: all 0.4s cubic-bezier(0.16, 1, 0.3, 1);
}

.btn-primary-base:hover {
  transform: translateY(-2px);
}

.btn-secondary {
  width: 210px;
  padding: 14px 0;
  background: rgba(255, 255, 255, 0.03);
  color: #a1a1aa;
  font-size: 0.85rem;
  font-weight: 400;
  letter-spacing: 0.15em;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 50px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.btn-secondary:hover {
  color: #ffffff;
  border-color: rgba(255, 255, 255, 0.2);
  background: rgba(255, 255, 255, 0.06);
}

@keyframes pulse {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: .4; transform: scale(0.9); }
}

@media (max-width: 1024px) {
  .astrolabe-dome { width: 1300px; height: 1300px; top: -65vw; }
  .hero-main { margin-top: 80px; }
  .hero-title { font-size: 2.8rem; }
}

@media (max-width: 640px) {
  .astrolabe-dome { width: 900px; height: 900px; top: -85vw; }
  .hero-main { margin-top: 60px; gap: 36px; }
  .hero-title { font-size: 2.1rem; letter-spacing: 0.15em; }
}
</style>