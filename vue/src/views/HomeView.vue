<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const savedChartId = ref<number | null>(null)

onMounted(async () => {
  if (authStore.isLoggedIn() && !authStore.user) {
    await authStore.fetchMe().catch(() => {})
  }
  const saved = localStorage.getItem('bazi_last_birth')
  if (saved) {
    try {
      savedChartId.value = JSON.parse(saved).chartId || null
    } catch {}
  }
})

function startChart() {
  router.push('/chart/new')
}

function continueChart() {
  if (savedChartId.value) router.push(`/chart/${savedChartId.value}`)
}
</script>

<template>
  <div class="home-page">
    <!-- Animated star-field background -->
    <div class="star-field" aria-hidden="true">
      <svg class="stars-svg" viewBox="0 0 1440 900" preserveAspectRatio="xMidYMid slice">
        <defs>
          <radialGradient id="nebula-gold" cx="50%" cy="50%" r="50%">
            <stop offset="0%" stop-color="#D4A84B" stop-opacity="0.08" />
            <stop offset="100%" stop-color="#D4A84B" stop-opacity="0" />
          </radialGradient>
          <radialGradient id="nebula-crimson" cx="50%" cy="50%" r="50%">
            <stop offset="0%" stop-color="#C41E3A" stop-opacity="0.06" />
            <stop offset="100%" stop-color="#C41E3A" stop-opacity="0" />
          </radialGradient>
          <filter id="star-glow">
            <feGaussianBlur stdDeviation="1.5" result="blur" />
            <feMerge>
              <feMergeNode in="blur" />
              <feMergeNode in="SourceGraphic" />
            </feMerge>
          </filter>
        </defs>
        <!-- Nebula patches -->
        <ellipse cx="720" cy="400" rx="600" ry="350" fill="url(#nebula-gold)" />
        <ellipse cx="200" cy="200" rx="300" ry="250" fill="url(#nebula-crimson)" />
        <!-- Stars - layer 1 (small) -->
        <circle cx="120" cy="80" r="1" fill="#D4A84B" opacity="0.4" />
        <circle cx="340" cy="150" r="0.8" fill="#fff" opacity="0.3" />
        <circle cx="580" cy="60" r="1.2" fill="#D4A84B" opacity="0.5" />
        <circle cx="800" cy="120" r="0.7" fill="#fff" opacity="0.25" />
        <circle cx="1000" cy="50" r="1" fill="#D4A84B" opacity="0.35" />
        <circle cx="1180" cy="180" r="0.9" fill="#fff" opacity="0.3" />
        <circle cx="1350" cy="90" r="1.1" fill="#D4A84B" opacity="0.4" />
        <circle cx="200" cy="350" r="0.6" fill="#fff" opacity="0.2" />
        <circle cx="450" cy="420" r="1" fill="#D4A84B" opacity="0.3" />
        <circle cx="650" cy="300" r="0.8" fill="#fff" opacity="0.35" />
        <circle cx="900" cy="380" r="1.3" fill="#D4A84B" opacity="0.45" />
        <circle cx="1100" cy="320" r="0.7" fill="#fff" opacity="0.2" />
        <circle cx="1280" cy="450" r="1" fill="#D4A84B" opacity="0.3" />
        <circle cx="80" cy="600" r="0.9" fill="#D4A84B" opacity="0.35" />
        <circle cx="300" cy="700" r="1.1" fill="#fff" opacity="0.3" />
        <circle cx="520" cy="620" r="0.7" fill="#D4A84B" opacity="0.4" />
        <circle cx="750" cy="750" r="1" fill="#fff" opacity="0.25" />
        <circle cx="980" cy="680" r="0.8" fill="#D4A84B" opacity="0.3" />
        <circle cx="1200" cy="720" r="1.2" fill="#fff" opacity="0.35" />
        <circle cx="1400" cy="600" r="0.9" fill="#D4A84B" opacity="0.4" />
        <!-- Stars - layer 2 (medium, brighter) -->
        <circle cx="250" cy="200" r="1.5" fill="#D4A84B" opacity="0.6" filter="url(#star-glow)" />
        <circle cx="720" cy="450" r="2" fill="#D4A84B" opacity="0.7" filter="url(#star-glow)" />
        <circle cx="1150" cy="300" r="1.5" fill="#D4A84B" opacity="0.5" filter="url(#star-glow)" />
        <circle cx="500" cy="800" r="1.8" fill="#D4A84B" opacity="0.6" filter="url(#star-glow)" />
        <circle cx="1000" cy="600" r="1.6" fill="#D4A84B" opacity="0.55" filter="url(#star-glow)" />
        <!-- Constellation lines -->
        <line
          x1="720"
          y1="450"
          x2="900"
          y2="380"
          stroke="#D4A84B"
          stroke-width="0.5"
          opacity="0.15"
        />
        <line
          x1="900"
          y1="380"
          x2="1000"
          y2="50"
          stroke="#D4A84B"
          stroke-width="0.5"
          opacity="0.1"
        />
        <line
          x1="720"
          y1="450"
          x2="1150"
          y2="300"
          stroke="#D4A84B"
          stroke-width="0.5"
          opacity="0.12"
        />
        <line
          x1="250"
          y1="200"
          x2="720"
          y2="450"
          stroke="#D4A84B"
          stroke-width="0.5"
          opacity="0.1"
        />
        <line
          x1="500"
          y1="800"
          x2="750"
          y2="750"
          stroke="#D4A84B"
          stroke-width="0.5"
          opacity="0.08"
        />
      </svg>
    </div>

    <!-- Diagonal accent line -->
    <div class="diagonal-accent" aria-hidden="true"></div>

    <!-- Main content -->
    <div class="hero-content">
      <!-- Symbol with glow -->
      <div class="symbol-wrapper animate-in">
        <div class="symbol-glow"></div>
        <div class="symbol-ring"></div>
        <div class="symbol">☯</div>
      </div>

      <!-- Title block -->
      <div class="title-block animate-in delay-1">
        <div class="eyebrow">
          <span class="eyebrow-line"></span>
          <span class="eyebrow-text">ZiWei · BaZi Fortune</span>
          <span class="eyebrow-line"></span>
        </div>
        <h1 class="hero-title">八字<span class="title-accent">命理</span></h1>
        <p class="hero-sub">命与运</p>
      </div>

      <!-- CTA buttons -->
      <div class="cta-group animate-in delay-2">
        <button class="btn-primary" @click="startChart">
          <span class="btn-icon">✦</span>
          开始排盘
        </button>
        <button v-if="savedChartId" class="btn-secondary" @click="continueChart">
          <span class="btn-icon-secondary">↻</span>
          继续上次
        </button>
      </div>

      <!-- Feature pills -->
      <div class="features animate-in delay-3">
        <div class="feature-pill">
          <span class="pill-dot"></span>
          命盘分析
        </div>
        <div class="pill-sep">·</div>
        <div class="feature-pill">
          <span class="pill-dot"></span>
          运势解读
        </div>
        <div class="pill-sep">·</div>
        <div class="feature-pill">
          <span class="pill-dot"></span>
          紫微斗数
        </div>
      </div>
    </div>

    <!-- Bottom gradient fade -->
    <div class="bottom-fade" aria-hidden="true"></div>
  </div>
</template>

<style scoped>
.home-page {
  position: relative;
  min-height: calc(100vh - 56px);
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}

/* ── Star field ── */
.star-field {
  position: absolute;
  inset: 0;
  z-index: 0;
}

.stars-svg {
  width: 100%;
  height: 100%;
  position: absolute;
  inset: 0;
}

/* ── Diagonal accent ── */
.diagonal-accent {
  position: absolute;
  top: -20%;
  right: -10%;
  width: 60%;
  height: 140%;
  background: linear-gradient(
    135deg,
    transparent 0%,
    rgba(212, 168, 75, 0.018) 40%,
    rgba(212, 168, 75, 0.03) 60%,
    transparent 100%
  );
  transform: skewX(-8deg);
  z-index: 0;
  pointer-events: none;
}

/* ── Bottom fade ── */
.bottom-fade {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 200px;
  background: linear-gradient(to top, var(--bg), transparent);
  z-index: 1;
  pointer-events: none;
}

/* ── Hero content ── */
.hero-content {
  position: relative;
  z-index: 2;
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  padding: 40px 24px;
  max-width: 680px;
  gap: 32px;
}

/* ── Symbol ── */
.symbol-wrapper {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 120px;
  height: 120px;
  animation: float 5s ease-in-out infinite;
}

.symbol-glow {
  position: absolute;
  inset: 0;
  border-radius: 50%;
  background: radial-gradient(circle, rgba(212, 168, 75, 0.12) 0%, transparent 70%);
  animation: pulse-glow 3s ease-in-out infinite;
}

.symbol-ring {
  position: absolute;
  inset: -8px;
  border-radius: 50%;
  border: 1px solid rgba(212, 168, 75, 0.15);
  animation: ring-expand 3s ease-in-out infinite;
}

@keyframes ring-expand {
  0%,
  100% {
    transform: scale(1);
    opacity: 0.15;
  }
  50% {
    transform: scale(1.08);
    opacity: 0.3;
  }
}

.symbol {
  font-size: 4rem;
  color: var(--gold);
  text-shadow:
    0 0 40px rgba(212, 168, 75, 0.5),
    0 0 80px rgba(212, 168, 75, 0.2);
  line-height: 1;
  animation: symbol-pulse 3s ease-in-out infinite;
}

@keyframes symbol-pulse {
  0%,
  100% {
    text-shadow:
      0 0 40px rgba(212, 168, 75, 0.5),
      0 0 80px rgba(212, 168, 75, 0.2);
  }
  50% {
    text-shadow:
      0 0 60px rgba(212, 168, 75, 0.7),
      0 0 120px rgba(212, 168, 75, 0.3);
  }
}

@keyframes float {
  0%,
  100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-10px);
  }
}

@keyframes pulse-glow {
  0%,
  100% {
    opacity: 0.6;
    transform: scale(1);
  }
  50% {
    opacity: 1;
    transform: scale(1.05);
  }
}

/* ── Title block ── */
.title-block {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}

.eyebrow {
  display: flex;
  align-items: center;
  gap: 12px;
}

.eyebrow-line {
  display: block;
  width: 40px;
  height: 1px;
  background: linear-gradient(90deg, transparent, rgba(212, 168, 75, 0.4));
}

.eyebrow-line:last-child {
  background: linear-gradient(90deg, rgba(212, 168, 75, 0.4), transparent);
}

.eyebrow-text {
  font-size: 11px;
  font-weight: 500;
  letter-spacing: 3px;
  color: rgba(212, 168, 75, 0.45);
  text-transform: uppercase;
}

.hero-title {
  font-family: var(--font-serif), 'Songti SC', serif;
  font-size: 4.5rem;
  font-weight: 700;
  color: var(--text);
  letter-spacing: 8px;
  margin: 0;
  line-height: 1.1;
  text-shadow: 0 2px 30px rgba(0, 0, 0, 0.5);
}

.title-accent {
  color: var(--gold);
  text-shadow: 0 0 40px rgba(212, 168, 75, 0.4);
}

.hero-sub {
  font-size: 1rem;
  color: var(--muted);
  letter-spacing: 2px;
  margin: 0;
}

/* ── CTA group ── */
.cta-group {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 14px;
}

.btn-primary {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  padding: 14px 48px;
  background: linear-gradient(135deg, #d4a84b, #b8860b);
  color: #0a0815;
  font-weight: 700;
  font-size: 1rem;
  border: none;
  border-radius: 50px;
  cursor: pointer;
  transition: all 0.3s ease;
  box-shadow:
    0 4px 24px rgba(212, 168, 75, 0.3),
    0 0 0 1px rgba(212, 168, 75, 0.2);
  letter-spacing: 2px;
  position: relative;
  overflow: hidden;
}

.btn-primary::before {
  content: '';
  position: absolute;
  top: 0;
  left: -100%;
  width: 100%;
  height: 100%;
  background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.15), transparent);
  transition: left 0.5s ease;
}

.btn-primary:hover {
  transform: translateY(-3px) scale(1.02);
  box-shadow: 0 12px 40px rgba(212, 168, 75, 0.45);
}

.btn-primary:hover::before {
  left: 100%;
}

.btn-icon {
  font-size: 0.8rem;
  animation: spin-slow 8s linear infinite;
}

@keyframes spin-slow {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.btn-secondary {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 12px 36px;
  background: transparent;
  color: rgba(212, 168, 75, 0.55);
  font-weight: 500;
  font-size: 0.9rem;
  border: 1px solid rgba(212, 168, 75, 0.15);
  border-radius: 50px;
  cursor: pointer;
  transition: all 0.3s ease;
  letter-spacing: 1px;
}

.btn-secondary:hover {
  color: var(--gold);
  border-color: rgba(212, 168, 75, 0.4);
  background: rgba(212, 168, 75, 0.05);
}

.btn-icon-secondary {
  font-size: 1rem;
}

/* ── Feature pills ── */
.features {
  display: flex;
  align-items: center;
  gap: 16px;
}

.feature-pill {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: rgba(255, 255, 255, 0.18);
  letter-spacing: 1px;
  opacity: 0;
  animation: fadeUp 0.6s ease both;
}

.feature-pill:nth-child(1) {
  animation-delay: 0.9s;
  opacity: 0;
}
.feature-pill:nth-child(3) {
  animation-delay: 1.1s;
  opacity: 0;
}
.feature-pill:nth-child(5) {
  animation-delay: 1.3s;
  opacity: 0;
}

.pill-dot {
  display: inline-block;
  width: 4px;
  height: 4px;
  border-radius: 50%;
  background: rgba(212, 168, 75, 0.3);
}

.pill-sep {
  color: rgba(255, 255, 255, 0.08);
  font-size: 12px;
}

@keyframes fadeUp {
  from {
    opacity: 0;
    transform: translateY(12px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* ── Responsive ── */
@media (max-width: 600px) {
  .hero-title {
    font-size: 3rem;
    letter-spacing: 4px;
  }
  .symbol-wrapper {
    width: 90px;
    height: 90px;
  }
  .symbol {
    font-size: 3rem;
  }
}
</style>
