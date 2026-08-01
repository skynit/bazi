<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { ArrowLeft, Pause, Play, ShieldAlert } from '@lucide/vue'
import type { FortuneDay } from '../api/fortune'
import { fetchDaily } from '../api/fortune'
import { getApiErrorMessage } from '../api/client'
import { blessingProfiles, resolveBlessingElement } from '../lib/blessing'
import PeriodNav from '../components/fortune/PeriodNav.vue'
import { vReveal } from '../composables/useReveal'

interface PracticeStep {
  key: string
  label: string
  title: string
  description: string
  image: string
  alt: string
}

interface EvidenceRow {
  label: string
  title: string
  description: string
}

function todayString(date = new Date()) {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
}

const route = useRoute()
const fortune = ref<FortuneDay | null>(null)
const chartId = ref('')
const loading = ref(true)
const error = ref('')
const practiceSeconds = ref(0)
const practiceActive = ref(false)
let practiceTimer: ReturnType<typeof setInterval> | undefined

const element = computed(() => resolveBlessingElement(fortune.value))
const profile = computed(() => blessingProfiles[element.value])
const navQuery = computed(() => (chartId.value ? { chart_id: chartId.value } : {}))
const pageStyle = computed<Record<string, string>>(() => ({
  '--blessing-accent': profile.value.accent,
  '--blessing-accent-dark': profile.value.accentDark,
  '--blessing-accent-rgb': profile.value.accentRgb,
}))
const practiceText = computed(() => {
  const minutes = Math.floor(practiceSeconds.value / 60)
  const seconds = practiceSeconds.value % 60
  return `${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`
})
const heroEvidence = computed(() => {
  const day = fortune.value
  if (!day) return []

  return [
    day.solar_date,
    day.jie_qi || day.seasonal_state?.season,
    `主题：${profile.value.tone}与整理`,
  ].filter(Boolean)
})
const practiceSteps = computed<PracticeStep[]>(() => [
  {
    key: 'prepare',
    label: '整境',
    title: '清出一块空间',
    description: profile.value.objectPrompt,
    image: profile.value.image,
    alt: profile.value.alt,
  },
  {
    key: 'pause',
    label: '停顿',
    title: '安静三分钟',
    description: '放松肩颈，缓慢呼吸。练习计时需要由你主动开始。',
    image: profile.value.ritualImages[1] || profile.value.ritualImages[0] || profile.value.image,
    alt: `${profile.value.element}行主题的安静环境`,
  },
  {
    key: 'act',
    label: '行动',
    title: '推进最小一步',
    description:
      profile.value.actions[0]?.description || '完成一个可以验证的小动作，再根据现实反馈调整方向。',
    image: profile.value.actions[0]?.image || profile.value.ritualImages[2] || profile.value.image,
    alt: `${profile.value.element}行主题行动：推进最小一步`,
  },
])
const mainAction = computed(() => profile.value.actions[0])
const alternativeActions = computed(() => profile.value.actions.slice(1, 2))
const relationshipCount = computed(
  () =>
    (fortune.value?.supporting_evidence?.length ?? 0) +
    (fortune.value?.counter_evidence?.length ?? 0),
)
const evidenceRows = computed<EvidenceRow[]>(() => [
  {
    label: '当日结构',
    title: `${profile.value.element}行主题来自今日五行样本`,
    description: '只用于选择页面主题和通用提示模板，不表示个性化吉凶。',
  },
  {
    label: '关系记录',
    title: relationshipCount.value
      ? `记录到 ${relationshipCount.value} 条传统结构关系`
      : '当前没有额外传统关系记录',
    description: '这些记录不等于事件会发生，也不构成健康、法律、财务或安全建议。',
  },
  {
    label: '现实判断',
    title: '以事实、可控因素和专业意见为准',
    description: '遇到重要决策时，请回到可验证信息，不依据本页提示单独行动。',
  },
])

function readChartId() {
  const raw = route.query.chart_id
  const fromQuery = Array.isArray(raw) ? raw[0] : raw
  if (fromQuery) return String(fromQuery)

  try {
    const saved = localStorage.getItem('bazi_last_birth')
    const savedId = saved ? JSON.parse(saved).chartId : null
    return savedId ? String(savedId) : ''
  } catch {
    return ''
  }
}

async function fetchBlessing() {
  loading.value = true
  error.value = ''
  const cid = readChartId()

  if (!cid) {
    chartId.value = ''
    fortune.value = null
    error.value = '请先创建命盘'
    loading.value = false
    return
  }

  chartId.value = cid
  try {
    fortune.value = await fetchDaily(Number(cid), todayString())
  } catch (reason: unknown) {
    error.value = getApiErrorMessage(reason, '运势加持加载失败，请稍后重试。')
  } finally {
    loading.value = false
  }
}

function togglePractice() {
  practiceActive.value = !practiceActive.value
}

onMounted(() => {
  void fetchBlessing()
  practiceTimer = setInterval(() => {
    if (practiceActive.value) practiceSeconds.value += 1
  }, 1000)
})

onUnmounted(() => {
  if (practiceTimer) clearInterval(practiceTimer)
})
</script>

<template>
  <div class="blessing-page" :style="pageStyle">
    <section v-if="loading" class="state-view" aria-live="polite">
      <div class="state-seal" aria-hidden="true">五行</div>
      <p>正在整理今日五行参考</p>
      <small>内容来自当日结构样本，不代表现实结果</small>
    </section>

    <section v-else-if="error" class="state-view" aria-live="polite">
      <div class="state-icon warning" aria-hidden="true"><ShieldAlert :size="24" /></div>
      <p>{{ error }}</p>
      <small v-if="error === '请先创建命盘'">运势加持需要基于一张已保存的命盘生成。</small>
      <router-link v-if="error === '请先创建命盘'" to="/" class="state-action">
        去创建命盘
      </router-link>
      <button v-else type="button" class="state-action" @click="fetchBlessing">重新加载</button>
    </section>

    <section v-else-if="!fortune" class="state-view" aria-live="polite">
      <div class="state-seal" aria-hidden="true">五行</div>
      <p>今日暂无可显示的加持内容</p>
      <small>可以稍后重新加载，或直接查看今日运势的结构记录。</small>
      <button type="button" class="state-action" @click="fetchBlessing">重新加载</button>
    </section>

    <div v-else class="blessing-shell">
      <section class="blessing-hero" aria-labelledby="blessing-title">
        <img
          class="hero-image"
          :src="profile.backdrop"
          :srcset="`${profile.backdrop} 1600w, ${profile.backdropHd} 3200w`"
          sizes="100vw"
          :alt="`${profile.element}行视觉锚点：${profile.tone}实景`"
          width="3200"
          height="1800"
          loading="eager"
          decoding="async"
          fetchpriority="high"
        />

        <div class="hero-toolbar">
          <router-link :to="{ path: '/fortune', query: navQuery }" class="hero-back">
            <ArrowLeft :size="16" />
            今日运势
          </router-link>
          <PeriodNav current="blessing" :chart-id="chartId" variant="overlay" />
        </div>

        <div class="hero-inner">
          <div class="hero-copy">
            <p class="hero-eyebrow">今日五行样本 · {{ profile.element }}行{{ profile.tone }}</p>
            <h1 id="blessing-title">运势加持</h1>
            <p class="hero-summary">{{ profile.summary }}</p>
            <div class="hero-meta" aria-label="今日参考摘要">
              <span v-for="item in heroEvidence" :key="item">{{ item }}</span>
            </div>
          </div>

          <div class="practice-timer" :aria-label="`${profile.element}行主题练习计时`">
            <div>
              <small>可选练习计时</small>
              <strong>安静三分钟</strong>
              <span>{{ practiceText }}</span>
            </div>
            <button
              type="button"
              class="timer-toggle"
              :title="practiceActive ? '暂停练习计时' : '开始练习计时'"
              :aria-label="practiceActive ? '暂停练习计时' : '开始练习计时'"
              :aria-pressed="practiceActive"
              @click="togglePractice"
            >
              <Pause v-if="practiceActive" :size="18" />
              <Play v-else :size="18" />
            </button>
          </div>
        </div>
      </section>

      <section class="context-band" aria-label="今日主题与说明" v-reveal>
        <div class="context-inner">
          <div class="today-focus">
            <div class="element-seal" aria-hidden="true">{{ profile.element }}</div>
            <div>
              <p class="section-kicker">今日主题</p>
              <h2>{{ mainAction?.title || '整理主线，再推进一步' }}</h2>
              <p>{{ mainAction?.description || profile.summary }}</p>
            </div>
          </div>
          <div class="boundary-copy">
            <p class="section-kicker">如何理解这页</p>
            <h2>生活提示，不是结果预测</h2>
            <p>
              本页由「今日运势」的同一份结构数据生成，只换成可执行的生活提示，不是新的排盘，也不代表事情会如何发展。
            </p>
            <router-link :to="{ path: '/fortune', query: navQuery }" class="boundary-link">
              查看今日运势的完整结构记录 →
            </router-link>
          </div>
        </div>
      </section>

      <main>
        <section class="content-section steps-section" aria-labelledby="steps-title" v-reveal>
          <header class="section-heading">
            <div>
              <p class="section-index">01 / 三步练习</p>
              <h2 id="steps-title">从整理环境，到完成一个具体行动</h2>
            </div>
            <p>不需要完成全部步骤。任选其中一项即可，感觉不适或现实条件不允许时直接跳过。</p>
          </header>

          <div class="steps-grid">
            <article v-for="(step, index) in practiceSteps" :key="step.key" class="step-card">
              <img :src="step.image" :alt="step.alt" loading="lazy" decoding="async" />
              <span class="step-number">0{{ index + 1 }}</span>
              <div class="step-copy">
                <span>{{ step.label }}</span>
                <h3>{{ step.title }}</h3>
                <p>{{ step.description }}</p>
              </div>
            </article>
          </div>
        </section>

        <section class="action-band" aria-labelledby="actions-title" v-reveal>
          <div class="content-section action-layout">
            <div class="action-copy">
              <p class="section-index">02 / 今日行动</p>
              <h2 id="actions-title">一个主行动，两个备选</h2>
              <p>用优先级替代建议堆叠，让页面真正帮助你决定下一步。</p>

              <div class="primary-action">
                <span>建议优先</span>
                <strong>{{ mainAction?.title || '完成一件小事' }}</strong>
                <p>{{ mainAction?.description || profile.summary }}</p>
              </div>

              <dl class="theme-notes">
                <div>
                  <dt>可选主题色</dt>
                  <dd>{{ profile.colors }}</dd>
                </div>
                <div>
                  <dt>环境物件</dt>
                  <dd>{{ profile.objects }}</dd>
                </div>
              </dl>
            </div>

            <div class="action-list">
              <article
                v-for="(action, index) in alternativeActions"
                :key="action.title"
                class="action-item"
              >
                <span>{{ ['壹', '贰'][index] }}</span>
                <div>
                  <strong>{{ action.title }}</strong>
                  <p>{{ action.description }}</p>
                </div>
                <small>{{ index === 0 ? '约 20 分钟' : '约 10 分钟' }}</small>
              </article>
              <article class="action-item">
                <span>贰</span>
                <div>
                  <strong>记录已知与未知</strong>
                  <p>把事实、待核实项和下一步各写一条，不急于补全结论。</p>
                </div>
                <small>约 5 分钟</small>
              </article>
            </div>
          </div>
        </section>

        <section class="content-section evidence-layout" aria-labelledby="evidence-title" v-reveal>
          <div class="evidence-heading">
            <p class="section-index">03 / 依据与边界</p>
            <h2 id="evidence-title">这组提示从哪里来</h2>
            <p>保留必要依据，但不把传统关系记录直接转化为现实行动结论。</p>
          </div>

          <dl class="evidence-list">
            <div v-for="row in evidenceRows" :key="row.label">
              <dt>{{ row.label }}</dt>
              <dd>
                <strong>{{ row.title }}</strong>
                <p>{{ row.description }}</p>
              </dd>
            </div>
          </dl>
        </section>
      </main>

      <footer class="blessing-footer">
        <div class="footer-inner">
          <div>
            <strong>把提示变成一个可验证的小行动</strong>
            <p>传统结构仅作文化参考，现实判断以事实和专业意见为准。</p>
          </div>
          <span aria-hidden="true">行</span>
        </div>
      </footer>
    </div>
  </div>
</template>

<style scoped>
.blessing-page {
  --page-paper: #f7f8f3;
  --page-surface: #ffffff;
  --page-muted: #eaf0e8;
  --page-line: rgba(19, 34, 26, 0.12);
  --page-ink: #13221a;
  --page-soft: #42534a;
  min-height: 100vh;
  margin-top: -80px;
  color: var(--page-ink);
  background: var(--page-paper);
  font-family: var(--font-sans);
}

.blessing-shell {
  overflow: hidden;
}

.blessing-hero {
  position: relative;
  display: grid;
  min-height: 670px;
  overflow: hidden;
  color: #ffffff;
  isolation: isolate;
}

.hero-image {
  position: absolute;
  z-index: -3;
  inset: 0;
  width: 100%;
  height: 100%;
  max-width: none;
  object-fit: cover;
  object-position: 50% 52%;
  filter: brightness(1.06) saturate(1) contrast(1.04);
}

.hero-toolbar,
.hero-inner,
.context-inner,
.content-section {
  width: min(1180px, calc(100% - 48px));
  margin-inline: auto;
}

.hero-toolbar {
  position: absolute;
  z-index: 2;
  top: 96px;
  left: 50%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  transform: translateX(-50%);
}

.hero-back {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  min-height: 44px;
  color: rgba(255, 255, 255, 0.88);
  font-size: 14px;
  font-weight: 650;
  text-decoration: none;
  text-shadow: 0 1px 10px rgba(4, 18, 12, 0.55);
}

.hero-inner {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(240px, 0.42fr);
  align-self: end;
  align-items: end;
  gap: 4rem;
  padding: 180px 0 72px;
}

.hero-copy {
  max-width: 720px;
}

.hero-eyebrow {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  margin: 0 0 1.35rem;
  color: #f0d889;
  font-size: 13px;
  font-weight: 750;
}

.hero-eyebrow::before {
  width: 34px;
  height: 1px;
  background: currentColor;
  content: '';
}

.hero-copy h1 {
  max-width: 7em;
  margin: 0;
  color: #ffffff;
  font-family: var(--font-serif);
  font-size: 72px;
  font-weight: 900;
  line-height: 1.05;
  letter-spacing: 0;
  text-shadow: 0 3px 22px rgba(4, 18, 12, 0.28);
}

.hero-summary {
  max-width: 620px;
  margin: 1.4rem 0 0;
  color: rgba(255, 255, 255, 0.94);
  font-family: var(--font-serif);
  font-size: 21px;
  line-height: 1.8;
  text-shadow: 0 2px 14px rgba(4, 18, 12, 0.34);
}

.hero-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 0.55rem;
  margin-top: 1.65rem;
}

.hero-meta span {
  padding: 0.45rem 0.7rem;
  border: 1px solid rgba(255, 255, 255, 0.28);
  border-radius: 6px;
  color: rgba(255, 255, 255, 0.9);
  background: rgba(24, 46, 35, 0.52);
  font-size: 13px;
}

.practice-timer {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 54px;
  align-items: center;
  gap: 1.4rem;
  padding-left: 1.5rem;
  border-left: 1px solid rgba(255, 255, 255, 0.32);
}

.practice-timer small,
.practice-timer strong,
.practice-timer span {
  display: block;
}

.practice-timer small {
  color: rgba(255, 255, 255, 0.72);
  font-size: 13px;
}

.practice-timer strong {
  margin-top: 0.15rem;
  font-size: 16px;
}

.practice-timer span {
  margin-top: 0.2rem;
  color: rgba(255, 255, 255, 0.76);
  font-family: var(--font-mono);
  font-size: 13px;
}

.timer-toggle {
  display: grid;
  width: 54px;
  height: 54px;
  padding: 0;
  place-items: center;
  border: 1px solid rgba(255, 255, 255, 0.44);
  border-radius: 50%;
  color: #ffffff;
  background: rgba(24, 46, 35, 0.42);
  cursor: pointer;
  transition:
    transform 180ms ease,
    background 180ms ease;
}

.timer-toggle:hover {
  transform: translateY(-2px);
  background: rgba(255, 255, 255, 0.18);
}

.context-band {
  border-bottom: 1px solid var(--page-line);
  background: var(--page-surface);
}

.context-inner {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 0.82fr);
}

.today-focus,
.boundary-copy {
  padding: 2.15rem 0;
}

.today-focus {
  display: grid;
  grid-template-columns: 86px minmax(0, 1fr);
  gap: 1.5rem;
  padding-right: 3.25rem;
}

.element-seal {
  display: grid;
  width: 74px;
  height: 74px;
  place-items: center;
  border: 1px solid rgba(var(--blessing-accent-rgb), 0.28);
  color: var(--blessing-accent-dark);
  background: rgba(var(--blessing-accent-rgb), 0.12);
  font-family: var(--font-serif);
  font-size: 34px;
  font-weight: 900;
}

.boundary-copy {
  padding-left: 2.6rem;
  border-left: 1px solid var(--page-line);
}

.section-kicker,
.section-index {
  margin: 0;
  color: var(--blessing-accent-dark);
  font-size: 13px;
  font-weight: 750;
}

.today-focus h2,
.boundary-copy h2 {
  margin: 0;
  font-family: var(--font-serif);
  font-size: 22px;
  line-height: 1.4;
  letter-spacing: 0;
}

.today-focus p:last-child,
.boundary-copy p:not(.section-kicker) {
  margin: 0.4rem 0 0;
  color: var(--page-soft);
  font-size: 14px;
}

.boundary-link {
  display: inline-flex;
  align-items: center;
  margin-top: 0.75rem;
  color: var(--blessing-accent-dark);
  font-size: 13px;
  font-weight: 650;
  text-decoration: none;
  transition: opacity 160ms ease;
}
.boundary-link:hover {
  text-decoration: underline;
  text-underline-offset: 4px;
}

.content-section {
  padding: 5.4rem 0;
}

.section-heading {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(280px, 0.52fr);
  align-items: end;
  gap: 2.25rem;
  margin-bottom: 2.1rem;
}

.section-heading h2,
.action-copy h2,
.evidence-heading h2 {
  margin: 0.45rem 0 0;
  color: var(--page-ink);
  font-family: var(--font-serif);
  font-size: 38px;
  line-height: 1.3;
  letter-spacing: 0;
}

.section-heading > p,
.action-copy > p:not(.section-index),
.evidence-heading > p:last-child {
  margin: 0;
  color: var(--page-soft);
  font-size: 14px;
}

.steps-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 1rem;
}

.step-card {
  position: relative;
  min-height: 420px;
  overflow: hidden;
  border: 1px solid var(--page-line);
  border-radius: 8px;
  color: #ffffff;
  background: #143f31;
  box-shadow: 0 18px 54px rgba(20, 52, 39, 0.12);
}

.step-card::after {
  position: absolute;
  inset: 0;
  background: linear-gradient(0deg, rgba(9, 30, 20, 0.94), rgba(9, 30, 20, 0.03) 70%);
  content: '';
}

.step-card img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 420ms ease;
}

.step-card:hover img {
  transform: scale(1.035);
}

.step-number {
  position: absolute;
  z-index: 2;
  top: 1.1rem;
  left: 1.1rem;
  display: grid;
  width: 34px;
  height: 34px;
  place-items: center;
  border: 1px solid rgba(255, 255, 255, 0.46);
  font-size: 13px;
  font-weight: 700;
}

.step-copy {
  position: absolute;
  z-index: 2;
  right: 1.4rem;
  bottom: 1.5rem;
  left: 1.4rem;
}

.step-copy > span {
  color: #f0d889;
  font-size: 13px;
  font-weight: 750;
}

.step-copy h3 {
  margin: 0.45rem 0 0.5rem;
  color: #ffffff;
  font-family: var(--font-serif);
  font-size: 24px;
  letter-spacing: 0;
}

.step-copy p {
  margin: 0;
  color: rgba(255, 255, 255, 0.8);
  font-size: 14px;
  line-height: 1.7;
}

.action-band {
  background: var(--page-muted);
}

.action-layout {
  display: grid;
  grid-template-columns: minmax(0, 0.78fr) minmax(0, 1.25fr);
  gap: 4.4rem;
  align-items: start;
}

.action-copy h2,
.evidence-heading h2 {
  margin-top: 0.45rem;
}

.action-copy > p:not(.section-index),
.evidence-heading > p:last-child {
  margin-top: 1rem;
}

.primary-action {
  margin-top: 2rem;
  padding-top: 1.5rem;
  border-top: 2px solid var(--blessing-accent-dark);
}

.primary-action > span {
  color: var(--blessing-accent-dark);
  font-size: 13px;
  font-weight: 750;
}

.primary-action strong {
  display: block;
  margin-top: 0.5rem;
  font-family: var(--font-serif);
  font-size: 24px;
}

.primary-action p {
  margin: 0.45rem 0 0;
  color: var(--page-soft);
  font-size: 14px;
}

.theme-notes {
  display: grid;
  gap: 0.75rem;
  margin: 1.5rem 0 0;
}

.theme-notes div {
  display: grid;
  grid-template-columns: 92px minmax(0, 1fr);
  gap: 0.75rem;
}

.theme-notes dt,
.theme-notes dd {
  margin: 0;
  font-size: 13px;
}

.theme-notes dt {
  color: var(--page-soft);
}

.theme-notes dd {
  color: var(--page-ink);
  font-weight: 650;
}

.action-list {
  border-top: 1px solid var(--page-line);
}

.action-item {
  display: grid;
  grid-template-columns: 42px minmax(0, 1fr) auto;
  align-items: center;
  gap: 1.4rem;
  min-height: 112px;
  border-bottom: 1px solid var(--page-line);
}

.action-item > span {
  color: var(--blessing-accent-dark);
  font-family: var(--font-serif);
  font-size: 19px;
  font-weight: 700;
}

.action-item strong {
  display: block;
  font-size: 16px;
}

.action-item p {
  margin: 0.25rem 0 0;
  color: var(--page-soft);
  font-size: 14px;
}

.action-item small {
  padding: 0.35rem 0.55rem;
  border: 1px solid var(--page-line);
  border-radius: 5px;
  color: var(--page-soft);
  background: color-mix(in oklab, var(--page-surface) 44%, transparent);
  font-size: 13px;
  white-space: nowrap;
}

.evidence-layout {
  display: grid;
  grid-template-columns: minmax(0, 0.65fr) minmax(0, 1.35fr);
  gap: 4.5rem;
}

.evidence-list {
  margin: 0;
}

.evidence-list > div {
  display: grid;
  grid-template-columns: 120px minmax(0, 1fr);
  gap: 1.4rem;
  padding: 1.4rem 0;
  border-bottom: 1px solid var(--page-line);
}

.evidence-list dt {
  color: var(--blessing-accent-dark);
  font-size: 13px;
  font-weight: 750;
}

.evidence-list dd {
  margin: 0;
}

.evidence-list strong {
  display: block;
  font-size: 15px;
}

.evidence-list p {
  margin: 0.3rem 0 0;
  color: var(--page-soft);
  font-size: 14px;
}

.blessing-footer {
  width: 100%;
  color: #dce8df;
  background: #173c2e;
}

.footer-inner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: min(1180px, calc(100% - 48px));
  min-height: 132px;
  margin-inline: auto;
  gap: 2rem;
}

.blessing-footer strong {
  font-family: var(--font-serif);
  font-size: 19px;
}

.blessing-footer p {
  margin: 0.25rem 0 0;
  color: rgba(220, 232, 223, 0.72);
  font-size: 13px;
}

.footer-inner > span {
  flex: 0 0 auto;
  padding-inline: 0.15em;
  font-family: var(--font-serif);
  font-size: 36px;
}

.state-view {
  display: grid;
  min-height: 100vh;
  padding: 9rem 1.5rem 4rem;
  place-items: center;
  align-content: center;
  gap: 0.75rem;
  text-align: center;
}

.state-seal,
.state-icon {
  display: grid;
  width: 64px;
  height: 64px;
  place-items: center;
  border: 1px solid rgba(var(--blessing-accent-rgb), 0.3);
  color: var(--blessing-accent-dark);
  background: rgba(var(--blessing-accent-rgb), 0.1);
  font-family: var(--font-serif);
  font-weight: 700;
}

.state-view p {
  margin: 0.25rem 0 0;
  font-family: var(--font-serif);
  font-size: 18px;
}

.state-view small {
  color: var(--page-soft);
  font-size: 14px;
}

.state-action {
  min-height: 42px;
  margin-top: 0.5rem;
  padding: 0.65rem 1rem;
  border: 1px solid var(--page-line);
  border-radius: 6px;
  color: #ffffff;
  background: var(--blessing-accent-dark);
  font-size: 14px;
  text-decoration: none;
  cursor: pointer;
}

.hero-back:focus-visible,
.timer-toggle:focus-visible,
.state-action:focus-visible {
  outline: 3px solid rgba(240, 216, 137, 0.7);
  outline-offset: 3px;
}

:global(.dark .blessing-page) {
  --page-paper: var(--surface-1);
  --page-surface: var(--surface-2);
  --page-muted: var(--surface-3);
  --page-line: var(--line-strong);
  --page-ink: var(--text);
  --page-soft: var(--text-muted);
}

:global(.dark .blessing-page .element-seal),
:global(.dark .blessing-page .state-seal),
:global(.dark .blessing-page .state-icon) {
  color: color-mix(in oklab, var(--blessing-accent) 76%, #ffffff);
}

:global(.dark .blessing-page .section-kicker),
:global(.dark .blessing-page .section-index),
:global(.dark .blessing-page .primary-action > span),
:global(.dark .blessing-page .action-item > span),
:global(.dark .blessing-page .boundary-link),
:global(.dark .blessing-page .evidence-list dt) {
  color: color-mix(in oklab, var(--blessing-accent) 78%, #ffffff);
}

@media (max-width: 900px) {
  .hero-inner {
    grid-template-columns: 1fr;
    gap: 2rem;
  }

  .practice-timer {
    max-width: 300px;
  }

  .context-inner,
  .section-heading,
  .action-layout,
  .evidence-layout {
    grid-template-columns: 1fr;
  }

  .boundary-copy {
    padding-top: 0;
    padding-left: 110px;
    border-left: 0;
  }

  .steps-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .step-card:last-child {
    grid-column: 1 / -1;
    min-height: 340px;
  }

  .action-layout,
  .evidence-layout {
    gap: 2.5rem;
  }
}

@media (max-width: 560px) {
  .blessing-hero {
    min-height: 720px;
  }

  .hero-image {
    object-position: 62% 50%;
  }

  .hero-toolbar,
  .hero-inner,
  .context-inner,
  .content-section,
  .footer-inner {
    width: calc(100% - 40px);
  }

  .hero-toolbar {
    top: 84px;
  }

  .hero-back {
    font-size: 13px;
  }

  .hero-inner {
    gap: 1.5rem;
    padding: 170px 0 44px;
  }

  .hero-eyebrow {
    margin-bottom: 0.9rem;
    font-size: 13px;
  }

  .hero-copy h1 {
    font-size: 48px;
  }

  .hero-summary {
    margin-top: 1rem;
    font-size: 18px;
    line-height: 1.7;
  }

  .hero-meta {
    gap: 0.4rem;
    margin-top: 1.2rem;
  }

  .hero-meta span {
    padding: 0.4rem 0.55rem;
    font-size: 13px;
  }

  .practice-timer {
    grid-template-columns: minmax(0, 1fr) 48px;
    gap: 1rem;
    max-width: 250px;
    padding-left: 1rem;
  }

  .timer-toggle {
    width: 48px;
    height: 48px;
  }

  .today-focus {
    grid-template-columns: 62px minmax(0, 1fr);
    gap: 1rem;
    padding: 1.6rem 0;
  }

  .element-seal {
    width: 58px;
    height: 58px;
    font-size: 27px;
  }

  .boundary-copy {
    padding: 0 0 1.6rem;
    border-top: 1px solid var(--page-line);
  }

  .boundary-copy .section-kicker {
    padding-top: 1.35rem;
  }

  .today-focus h2,
  .boundary-copy h2 {
    font-size: 19px;
  }

  .content-section {
    padding: 3.6rem 0;
  }

  .section-heading {
    gap: 0.8rem;
    margin-bottom: 1.6rem;
  }

  .section-heading h2,
  .action-copy h2,
  .evidence-heading h2 {
    font-size: 30px;
  }

  .steps-grid {
    grid-template-columns: 1fr;
  }

  .step-card,
  .step-card:last-child {
    grid-column: auto;
    min-height: 390px;
  }

  .action-item {
    grid-template-columns: 34px minmax(0, 1fr);
    gap: 0.9rem;
    padding: 1.1rem 0;
  }

  .action-item small {
    grid-column: 2;
    justify-self: start;
  }

  .evidence-list > div {
    grid-template-columns: 1fr;
    gap: 0.35rem;
  }

  .footer-inner {
    align-items: flex-start;
    min-height: 160px;
    padding: 1.9rem 0;
  }

  .footer-inner > span {
    align-self: center;
  }
}

@media (prefers-reduced-motion: reduce) {
  .timer-toggle,
  .step-card img {
    transition: none !important;
  }
}
</style>
