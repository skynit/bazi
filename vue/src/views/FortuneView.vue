<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import DailyFortune from '../components/DailyFortune.vue'
import type { FortuneDay } from '../api/fortune'
import { fetchDaily } from '../api/fortune'
import { getApiErrorMessage } from '../api/client'

const route = useRoute()
const fortune = ref<FortuneDay | null>(null)
const loading = ref(true)
const error = ref('')
const errorKind = ref<'missing-chart' | 'request' | ''>('')
const mounted = ref(false)
const chartId = ref<string | number>('')
const relationCount = computed(
  () =>
    (fortune.value?.supporting_evidence?.length ?? 0) +
    (fortune.value?.counter_evidence?.length ?? 0),
)

function todayString(date = new Date()) {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
}

async function fetchFortune() {
  loading.value = true
  error.value = ''
  errorKind.value = ''
  fortune.value = null
  let cid: string | number | null = route.query.chart_id as string | null
  if (!cid) {
    try {
      const s = localStorage.getItem('bazi_last_birth')
      if (s) cid = JSON.parse(s).chartId
    } catch {}
    if (!cid) {
      error.value = '请先创建命盘'
      errorKind.value = 'missing-chart'
      loading.value = false
      return
    }
  }
  chartId.value = cid
  try {
    fortune.value = await fetchDaily(Number(chartId.value), todayString())
  } catch (reason: unknown) {
    const status = (reason as { response?: { status?: number } }).response?.status
    errorKind.value = status === 404 ? 'missing-chart' : 'request'
    error.value =
      status === 404
        ? '命盘不存在或已被删除，请重新排盘。'
        : getApiErrorMessage(reason, '今日运势加载失败，请稍后重试。')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchFortune()
  setTimeout(() => (mounted.value = true), 100)
})
</script>

<template>
  <div class="fortune-page">
    <!-- ── Loading ── -->
    <div v-if="loading" class="loading-state">
      <div class="loading-orbs">
        <div class="l-orb l-orb-1"></div>
        <div class="l-orb l-orb-2"></div>
        <div class="l-orb l-orb-3"></div>
        <div class="l-core"></div>
      </div>
      <p class="loading-text">正在整理今日干支关系</p>
    </div>

    <!-- ── Error ── -->
    <div v-else-if="error" class="error-state">
      <div class="error-sigil">✕</div>
      <p class="error-text">{{ error }}</p>
      <router-link v-if="errorKind === 'missing-chart'" to="/chart/new" class="retry-btn"
        >去排盘</router-link
      >
      <button v-else class="retry-btn" @click="fetchFortune">重新加载</button>
    </div>

    <!-- ── Empty ── -->
    <div v-else-if="!fortune" class="empty-state">
      <div class="empty-sigil">◈</div>
      <p class="empty-title">今日暂无可显示的关系记录</p>
      <button type="button" class="go-chart-btn" @click="fetchFortune">重新加载</button>
    </div>

    <!-- ── Main ── -->
    <main v-else class="fortune-main" :class="{ visible: mounted }">
      <!-- Relationship summary + date -->
      <header class="hero-panel">
        <div class="hero-inner">
          <div class="relation-summary" aria-label="今日命中关系条数">
            <div class="relation-value">{{ relationCount }}</div>
            <div class="relation-label">条结构关系</div>
          </div>

          <!-- Info block -->
          <div class="hero-text">
            <div class="hero-date-row">
              <span class="date-label">{{ fortune.solar_date }}</span>
            </div>
            <div class="hero-pillar-display">
              <span class="pillar-prefix">日柱</span>
              <span class="pillar-value">{{ fortune.day_gan_zhi }}</span>
            </div>
            <p class="hero-summary">今日干支关系概览</p>
            <p class="hero-tip">条数只表示命中的规则关系数量，不代表吉凶、强弱或事件概率。</p>
          </div>
        </div>
      </header>

      <!-- Content -->
      <div class="main-grid">
        <DailyFortune
          :solar-date="fortune.solar_date"
          :day-gan-zhi="fortune.day_gan_zhi"
          :week-day="fortune.week_day"
          :lunar-date="fortune.lunar_date"
          :sheng-xiao="fortune.sheng_xiao"
          :score-breakdown="fortune.score_breakdown"
          :supporting-evidence="fortune.supporting_evidence"
          :counter-evidence="fortune.counter_evidence"
          :chong-sha="fortune.clash_zodiac"
          :element-images="fortune.element_images"
          :today-elements="fortune.today_elements"
          :ji-shen="fortune.ji_shen"
          :xiong-shen="fortune.xiong_shen"
          :tai-shen="fortune.tai_shen"
          :peng-zu="fortune.peng_zu"
          :gua="fortune.gua"
          :jie-qi="fortune.jie_qi"
          :sheng-ke-analysis="fortune.sheng_ke_analysis"
          :season-element="fortune.season_element"
          :ten-god="fortune.ten_god"
          :twelve-stage="fortune.twelve_stage"
          :jian-chu="fortune.jian_chu"
          :huang-dao="fortune.huang_dao"
          :hidden-stems="fortune.hidden_stems"
          :stem-relations="fortune.stem_relations"
          :branch-relations="fortune.branch_relations"
          :activated-shen-sha="fortune.activated_shen_sha"
          :seasonal-state="fortune.seasonal_state"
          :fortune-layers="fortune.fortune_layers"
        />
      </div>

      <div class="fortune-nav">
        <router-link :to="`/fortune/blessing?chart_id=${chartId}`" class="fortune-nav-link">
          运势加持 →
        </router-link>
        <router-link :to="`/fortune/weekly?chart_id=${chartId}`" class="fortune-nav-link">
          本周运势 →
        </router-link>
        <router-link :to="`/fortune/monthly?chart_id=${chartId}`" class="fortune-nav-link">
          本月运势 →
        </router-link>
      </div>
    </main>
  </div>
</template>

<style scoped>
/* ── Page ── */
.fortune-page {
  min-height: 100vh;
  background: transparent;
  position: relative;
  overflow-x: hidden;
}

/* ── Loading ── */
.loading-state {
  position: relative;
  z-index: 2;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  gap: 2.5rem;
}
.loading-orbs {
  position: relative;
  width: 100px;
  height: 100px;
}
.l-orb {
  position: absolute;
  inset: 0;
  border-radius: 50%;
  border: 1px solid var(--line-strong);
  animation: l-spin linear infinite;
}
.l-orb-2 {
  inset: 15px;
  border-color: var(--line-subtle);
  animation-duration: 5s;
  animation-direction: reverse;
}
.l-orb-3 {
  inset: 30px;
  border-color: var(--line-subtle);
  animation-duration: 8s;
}
.l-core {
  position: absolute;
  inset: 40px;
  border-radius: 50%;
  background: var(--accent-dim);
  animation: l-pulse 2s ease-in-out infinite;
}
@keyframes l-spin {
  to {
    transform: rotate(360deg);
  }
}
@keyframes l-pulse {
  0%,
  100% {
    transform: scale(1);
    opacity: 0.3;
  }
  50% {
    transform: scale(1.6);
    opacity: 0.8;
  }
}
.loading-text {
  color: var(--text-soft);
  font-size: var(--fs-2xs);
  letter-spacing: 5px;
  text-transform: uppercase;
}

/* ── Error ── */
.error-state {
  position: relative;
  z-index: 2;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  gap: 1.5rem;
}
.error-sigil {
  font-size: var(--fs-hero-strong);
  color: var(--crimson);
  opacity: 0.5;
}
.error-text {
  color: var(--text-muted);
  font-size: var(--fs-sm);
}
.retry-btn {
  padding: 0.7rem 2rem;
  background: var(--crimson);
  color: var(--destructive-foreground);
  border: none;
  border-radius: 8px;
  font-size: var(--fs-sm);
  font-weight: 700;
  cursor: pointer;
  box-shadow: 0 4px 20px color-mix(in oklab, var(--crimson) 30%, transparent);
  transition: all 0.3s;
}
.retry-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 30px color-mix(in oklab, var(--crimson) 50%, transparent);
}

/* ── Empty ── */
.empty-state {
  position: relative;
  z-index: 2;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  gap: 1.5rem;
}
.empty-sigil {
  font-size: var(--fs-hero-strong);
  color: var(--text-soft);
}
.empty-title {
  color: var(--text-muted);
  font-size: var(--fs-lg);
}
.go-chart-btn {
  padding: 0.8rem 2.5rem;
  background: var(--accent);
  color: var(--bg);
  font-weight: 800;
  border: none;
  border-radius: 50px;
  cursor: pointer;
  text-decoration: none;
  font-size: var(--fs-sm);
  letter-spacing: 1px;
  box-shadow: 0 4px 30px color-mix(in oklab, var(--accent) 40%, transparent);
  transition: all 0.3s;
}
.go-chart-btn:hover {
  transform: translateY(-3px);
  box-shadow: 0 8px 50px color-mix(in oklab, var(--accent) 60%, transparent);
}

/* ── Main ── */
.fortune-main {
  position: relative;
  z-index: 2;
  max-width: 960px;
  margin: 0 auto;
  padding: 2.5rem 1.5rem 5rem;
  opacity: 0;
  transform: translateY(30px);
  transition:
    opacity 1s ease,
    transform 1s ease;
}
.fortune-main.visible {
  opacity: 1;
  transform: translateY(0);
}

/* ── Hero panel ── */
.hero-panel {
  background: color-mix(in oklab, var(--surface-1) 88%, transparent);
  border: 1px solid var(--line-strong);
  border-radius: 20px;
  padding: 2.5rem;
  margin-bottom: 1.5rem;
  position: relative;
  overflow: hidden;
  box-shadow:
    var(--shadow-lg),
    inset 0 1px 0 var(--line-subtle);
}
.hero-panel::before {
  content: none;
}
.hero-panel::after {
  content: '';
  position: absolute;
  top: -50%;
  right: -10%;
  width: 400px;
  height: 400px;
  background: radial-gradient(circle, var(--accent-dim), transparent 60%);
  pointer-events: none;
}
.hero-inner {
  display: flex;
  align-items: center;
  gap: 3rem;
}

.relation-summary {
  flex: 0 0 180px;
  min-height: 150px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.35rem;
  padding: 1.25rem;
  border: 1px solid var(--line-strong);
  border-radius: 8px;
  background: color-mix(in oklab, var(--surface-1) 74%, transparent);
}

.relation-value {
  color: var(--accent);
  font-family: var(--font-serif);
  font-size: var(--fs-hero-strong);
  font-weight: 900;
  line-height: 1;
}

.relation-label {
  color: var(--text-muted);
  font-size: var(--fs-xs);
  font-weight: 700;
}

/* ── Score sphere ── */
.score-sphere {
  flex-shrink: 0;
  width: 180px;
  height: 180px;
  border-radius: 50%;
  background: var(--surface-0);
  border: 2px solid var(--line-strong);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  position: relative;
  box-shadow:
    0 0 0 1px var(--line-subtle),
    0 0 60px var(--sg, var(--line-subtle)),
    0 0 120px var(--sg, var(--line-subtle)),
    inset 0 0 60px rgba(0, 0, 0, 0.15);
}
.sphere-ring {
  position: absolute;
  border-radius: 50%;
  border: 1px solid var(--line-subtle);
  animation: ring-spin linear infinite;
}
.sphere-ring-1 {
  inset: -12px;
  animation-duration: 20s;
}
.sphere-ring-2 {
  inset: -24px;
  animation-duration: 35s;
  animation-direction: reverse;
  border-color: color-mix(in oklab, var(--crimson) 5%, transparent);
}
.sphere-ring-3 {
  inset: -40px;
  animation-duration: 50s;
  border-color: var(--line-subtle);
}
@keyframes ring-spin {
  to {
    transform: rotate(360deg);
  }
}
.sphere-glow-a {
  position: absolute;
  inset: -30px;
  border-radius: 50%;
  background: radial-gradient(circle, var(--sg, var(--line-subtle)) 0%, transparent 70%);
  animation: glow-pulse 3s ease-in-out infinite;
}
.sphere-glow-b {
  position: absolute;
  inset: 10%;
  border-radius: 50%;
  background: radial-gradient(circle, var(--sc, #cbd5e1) 0%, transparent 70%);
  opacity: 0.06;
  animation: glow-inner 4s ease-in-out infinite;
}
@keyframes glow-pulse {
  0%,
  100% {
    opacity: 0.6;
    transform: scale(1);
  }
  50% {
    opacity: 1;
    transform: scale(1.08);
  }
}
@keyframes glow-inner {
  0%,
  100% {
    opacity: 0.04;
  }
  50% {
    opacity: 0.1;
  }
}
.sphere-value {
  font-family: var(--font-serif);
  font-size: var(--fs-hero-strong);
  font-weight: 900;
  color: var(--sc, #cbd5e1);
  line-height: 1;
  position: relative;
  z-index: 2;
  text-shadow:
    0 0 60px var(--sg, var(--accent-glow)),
    0 0 20px var(--sc, var(--accent));
  transition:
    color 0.6s,
    text-shadow 0.6s;
  letter-spacing: -2px;
}
.sphere-label {
  font-size: var(--fs-2xs);
  font-weight: 700;
  color: var(--text-muted);
  letter-spacing: 4px;
  text-transform: uppercase;
  position: relative;
  z-index: 2;
  margin-top: 2px;
}

/* ── Hero text ── */
.hero-text {
  flex: 1;
}
.hero-date-row {
  margin-bottom: 0.75rem;
}
.date-label {
  font-size: var(--fs-xs);
  color: var(--text-muted);
  letter-spacing: 3px;
  text-transform: uppercase;
}
.hero-pillar-display {
  display: flex;
  align-items: baseline;
  gap: 0.75rem;
  margin-bottom: 1rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid var(--line-subtle);
}
.pillar-prefix {
  font-size: var(--fs-xs);
  color: var(--text-soft);
  letter-spacing: 2px;
}
.pillar-value {
  font-family: var(--font-serif);
  font-size: var(--fs-stat-lg);
  font-weight: 900;
  color: var(--accent);
  letter-spacing: 4px;
  text-shadow: 0 0 40px var(--accent-glow);
}
.hero-summary {
  font-size: var(--fs-sm);
  color: var(--text-muted);
  line-height: 1.8;
  margin: 0 0 0.75rem;
  border-left: 2px solid var(--line-strong);
  padding-left: 0.75rem;
}
.hero-tip {
  font-size: var(--fs-sm);
  color: var(--accent);
  font-weight: 600;
  margin: 0;
  opacity: 0.85;
}

/* ── Main grid ── */
.main-grid {
  display: block;
}

@media (max-width: 768px) {
  .hero-inner {
    flex-direction: column;
    align-items: center;
    text-align: center;
    gap: 1.5rem;
  }
  .hero-panel {
    padding: 1.75rem 1.5rem;
  }
  .score-sphere {
    width: 140px;
    height: 140px;
  }
  .relation-summary {
    flex-basis: auto;
    width: min(220px, 100%);
    min-height: 120px;
  }
  .sphere-value {
    font-size: var(--fs-hero);
  }
  .pillar-value {
    font-size: var(--fs-stat);
  }
  .fortune-main {
    padding: 1.5rem 1rem 4rem;
  }
}

/* ── Fortune nav ── */
.fortune-nav {
  display: flex;
  justify-content: center;
  gap: 1.5rem;
  padding: 1.5rem 0 0;
}
.fortune-nav-link {
  color: var(--accent);
  text-decoration: none;
  font-size: var(--fs-sm);
  font-weight: 500;
  transition: all 0.2s;
}
.fortune-nav-link:hover {
  text-shadow: 0 0 12px var(--accent-glow);
}

/* ── Dark mode overrides ── */
:global(.dark) .score-sphere {
  background: radial-gradient(circle at 35% 35%, #1a1530 0%, #050308 100%);
  box-shadow:
    0 0 0 1px rgba(203, 213, 225, 0.05),
    0 0 60px var(--sg, rgba(203, 213, 225, 0.2)),
    0 0 120px var(--sg, rgba(203, 213, 225, 0.1)),
    inset 0 0 60px rgba(0, 0, 0, 0.6);
}
:global(.dark) .sphere-ring-2 {
  border-color: rgba(251, 113, 133, 0.05);
}
</style>
