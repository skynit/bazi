<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import {
  ArrowLeft,
  CalendarDays,
  Clock3,
  Pause,
  Play,
  ShieldAlert,
  Sparkles,
} from '@lucide/vue'
import type { FortuneDay } from '../api/fortune'
import { fetchDaily } from '../api/fortune'
import {
  blessingProfiles,
  isBlessingElement,
  resolveBlessingElement,
  type BlessingProfile,
} from '../lib/blessing'

interface BlessingAsset {
  url: string
  thumbnail_url?: string
  alt_text: string
  name: string
  description?: string
  element: string
  orientation: 'landscape' | 'portrait' | 'square'
  tone?: string
  focal_x?: number
  focal_y?: number
}

interface PracticeItem {
  label: string
  value: string
  reason: string
  method?: string
  category?: string
  timing?: string
  element?: string
}

function todayString(date = new Date()) {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
}

function splitValues(value?: string) {
  return (value ?? '')
    .split(/[、,，+＋]/)
    .map((item) => item.trim())
    .filter(Boolean)
}

const route = useRoute()
const fortune = ref<FortuneDay | null>(null)
const chartId = ref('')
const loading = ref(true)
const error = ref('')
const presenceSeconds = ref(0)
const fieldActive = ref(true)
let presenceTimer: ReturnType<typeof setInterval> | undefined

const element = computed(() => resolveBlessingElement(fortune.value))
const profile = computed(() => blessingProfiles[element.value])
const guide = computed(() => ({
  strategy: profile.value.summary,
  face_direction: { value: profile.value.direction },
  avoid_direction: { value: profile.value.avoidDirection },
  avoid_element: '',
  analysis: '用短暂静驻、环境整理和一个明确动作承接今日节奏。',
  evidence_completeness: fortune.value?.evidence_completeness,
}))
const colorText = computed(() => profile.value.colors)
const colorChips = computed(() => splitValues(colorText.value))
const objectText = computed(() => profile.value.objects)
const objectChips = computed(() => splitValues(objectText.value))
const actionItems = computed<PracticeItem[]>(() =>
  profile.value.actions.map((action) => ({
    label: action.title,
    value: action.title,
    reason: action.description,
    method: action.description,
    category: `${profile.value.element}行行动`,
    timing: '今日任选一刻',
    element: profile.value.element,
  })),
)
const actionCards = computed(() =>
  actionItems.value.map((item, index) => ({
    item,
    asset: {
      url: profile.value.actions[index]?.image || profile.value.image,
      alt_text: `${profile.value.element}行行动：${item.value}`,
      name: item.value,
      element: profile.value.element,
      orientation: 'square',
    } satisfies BlessingAsset,
  })),
)
const cautionItems = computed<PracticeItem[]>(() =>
  (fortune.value?.counter_evidence ?? []).slice(0, 3).map((item) => ({
    label: item.label,
    value: item.label,
    reason: item.description,
    method: item.description,
  })),
)
const hourItems = computed(() => [
  { label: '开始前', value: '静驻 3 分钟' },
  { label: '进行中', value: '专注 20 分钟' },
  { label: '收尾时', value: '复盘 5 分钟' },
])
const galleryProfiles = computed(() => Object.values(blessingProfiles))
const galleryItems = computed(() => {
  return galleryProfiles.value.map((item) => ({
    key: `fallback-${item.element}`,
    asset: {
      url: item.image,
      thumbnail_url: undefined,
      alt_text: item.alt,
      name: item.objectLabel,
      element: item.element,
      orientation: 'square',
      tone: item.tone,
      description: item.summary,
      focal_x: 0.5,
      focal_y: 0.5,
    } satisfies BlessingAsset,
    profile: item,
  }))
})
const pageStyle = computed<Record<string, string>>(() => ({
  '--blessing-accent': profile.value.accent,
  '--blessing-accent-dark': profile.value.accentDark,
  '--blessing-accent-rgb': profile.value.accentRgb,
  '--blessing-backdrop': `url(${profile.value.backdrop})`,
}))
const navQuery = computed(() => (chartId.value ? { chart_id: chartId.value } : {}))
const heroEvidence = computed(() => {
  const day = fortune.value
  if (!day) return []

  return [
    day.jie_qi || day.seasonal_state?.season,
    day.ten_god?.status === 'observed' ? day.ten_god.name : '',
    `证据完整度 ${Math.round(day.evidence_completeness)}%`,
  ].filter(Boolean)
})
const presenceText = computed(() => {
  const minutes = Math.floor(presenceSeconds.value / 60)
  const seconds = presenceSeconds.value % 60
  return `${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`
})
const imageModules = computed(() => {
  const steps = [
    {
      key: 'prepare',
      badge: '整境',
      title: '先让环境安静下来',
      text: profile.value.objectLabel,
      note: profile.value.objectPrompt,
      tone: 'primary',
    },
    {
      key: 'breathe',
      badge: '调息',
      title: '三分钟静驻承气',
      text: '放松肩颈，缓慢呼吸',
      note: '计时只记录静驻时长；感觉不适时应立即停止。',
      tone: 'secondary',
    },
    {
      key: 'act',
      badge: '行持',
      title: '把主气落到一件小事',
      text: profile.value.actions[0]?.title || profile.value.objects,
      note: profile.value.actions[0]?.description || profile.value.summary,
      tone: 'primary',
    },
  ]

  return steps.map((step, index) => ({
    ...step,
    profile: profile.value,
    asset: {
      url: profile.value.ritualImages[index] || profile.value.image,
      alt_text: `${profile.value.element}行承气步骤：${step.title}`,
      name: step.title,
      element: profile.value.element,
      orientation: 'landscape',
      focal_x: 0.5,
      focal_y: 0.5,
    } satisfies BlessingAsset,
  }))
})

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

function errorMessage(reason: unknown) {
  const response = reason as { response?: { data?: { error?: string } } }
  return response.response?.data?.error || '加载运势加持失败'
}

function profileForElement(elementName?: string): BlessingProfile | undefined {
  return isBlessingElement(elementName) ? blessingProfiles[elementName] : undefined
}

function profileForAction(item: PracticeItem) {
  return profileForElement(item.element) ?? profile.value
}

function assetFocusStyle(asset?: BlessingAsset) {
  if (!asset) return undefined
  return {
    objectPosition: `${(asset.focal_x ?? 0.5) * 100}% ${(asset.focal_y ?? 0.5) * 100}%`,
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
    error.value = errorMessage(reason)
  } finally {
    loading.value = false
  }
}

function toggleField() {
  fieldActive.value = !fieldActive.value
}

onMounted(() => {
  void fetchBlessing()
  presenceTimer = setInterval(() => {
    if (fieldActive.value) presenceSeconds.value += 1
  }, 1000)
})

onUnmounted(() => {
  if (presenceTimer) clearInterval(presenceTimer)
})
</script>

<template>
  <div class="blessing-page" :class="{ 'field-paused': !fieldActive }" :style="pageStyle">
    <div class="ambient-scene" aria-hidden="true">
      <div class="ambient-image"></div>
      <div class="ambient-wash"></div>
      <div class="qi-stream qi-stream-one"></div>
      <div class="qi-stream qi-stream-two"></div>
      <div class="grain"></div>
    </div>

    <div v-if="loading" class="state-view">
      <div class="loading-seal"><span>气</span></div>
      <p>正在感应今日五行</p>
      <small>一呼一吸之间，气场渐明</small>
    </div>

    <div v-else-if="error" class="state-view">
      <div class="state-mark warning"><ShieldAlert :size="24" /></div>
      <p>{{ error }}</p>
      <div class="state-actions">
        <router-link v-if="error === '请先创建命盘'" to="/chart/new" class="primary-action">
          去排盘
        </router-link>
        <button v-else type="button" class="primary-action" @click="fetchBlessing">重新加载</button>
      </div>
    </div>

    <div v-else-if="fortune" class="blessing-shell">
      <nav class="top-nav" aria-label="运势加持导航">
        <router-link :to="{ path: '/fortune', query: navQuery }" class="nav-back">
          <ArrowLeft :size="16" />
          今日运势
        </router-link>
        <div class="nav-periods">
          <span class="active">今日加持</span>
          <router-link :to="{ path: '/fortune/weekly', query: navQuery }">本周</router-link>
          <router-link :to="{ path: '/fortune/monthly', query: navQuery }">本月</router-link>
        </div>
      </nav>

      <section class="blessing-hero" aria-labelledby="blessing-title">
        <img
          class="hero-image"
          :src="profile.backdrop"
          :alt="`${profile.element}行视觉锚点：${profile.tone}实景`"
          decoding="async"
        />

        <div class="hero-inner">
          <div class="hero-copy">
            <p class="eyebrow">今日五行锚点 · {{ profile.element }}行{{ profile.tone }}</p>
            <h1 id="blessing-title">运势加持</h1>
            <p class="hero-strategy">{{ profile.summary }}</p>
            <div class="hero-evidence" aria-label="今日排盘依据摘要">
              <span v-for="item in heroEvidence" :key="item">{{ item }}</span>
            </div>
          </div>

          <div class="hero-presence" :aria-label="`${profile.element}行静驻计时`">
            <span class="hero-element" aria-hidden="true">{{ profile.element }}</span>
            <div class="presence-readout">
              <div>
                <span>静驻时间</span>
                <strong>{{ presenceText }}</strong>
              </div>
              <button
                type="button"
                class="field-toggle"
                :title="fieldActive ? '暂停静驻计时' : '继续静驻计时'"
                :aria-label="fieldActive ? '暂停静驻计时' : '继续静驻计时'"
                :aria-pressed="!fieldActive"
                @click="toggleField"
              >
                <Pause v-if="fieldActive" :size="17" />
                <Play v-else :size="17" />
              </button>
            </div>
          </div>
        </div>
      </section>

      <section class="ritual-section" aria-labelledby="ritual-title">
        <header class="section-heading">
          <div>
            <p>THREE STEPS · 三步承气</p>
            <h2 id="ritual-title">把五行之气，落到今日生活</h2>
          </div>
          <span class="section-number">壹</span>
        </header>

        <div class="ritual-grid">
          <article
            v-for="(module, index) in imageModules"
            :key="module.key"
            class="ritual-card"
            :class="{ avoid: module.tone === 'avoid' }"
            :style="{
              '--module-accent': module.profile.accent,
              '--module-rgb': module.profile.accentRgb,
            }"
          >
            <div class="ritual-image-wrap">
              <img
                :src="module.asset?.url || module.profile.image"
                :alt="module.asset?.alt_text || module.profile.alt"
                loading="lazy"
                decoding="async"
                :style="assetFocusStyle(module.asset)"
              />
              <span>0{{ index + 1 }}</span>
            </div>
            <div class="ritual-body">
              <p>{{ module.badge }}</p>
              <h3>{{ module.title }}</h3>
              <strong>{{ module.text }}</strong>
              <small>{{ module.note }}</small>
            </div>
          </article>
        </div>
      </section>

      <section class="daily-practice" aria-labelledby="practice-title">
        <header class="section-heading compact">
          <div>
            <p>DAILY PRACTICE · 今日行持</p>
            <h2 id="practice-title">顺势而为，知止而安</h2>
          </div>
          <span class="section-number">贰</span>
        </header>

        <div class="practice-columns">
          <article class="practice-panel favorable">
            <div class="panel-title"><Sparkles :size="17" /><span>今日宜行</span></div>
            <div v-if="actionItems.length" class="practice-list">
              <div
                v-for="card in actionCards"
                :key="`${card.item.label}-${card.item.value}`"
                class="practice-item"
              >
                <img
                  :src="card.asset?.url || profileForAction(card.item).image"
                  :alt="card.asset?.alt_text || profileForAction(card.item).alt"
                  :style="assetFocusStyle(card.asset)"
                />
                <div>
                  <span>{{ card.item.category || card.item.label }}</span>
                  <strong>{{ card.item.value }}</strong>
                  <p>{{ card.item.method || card.item.reason }}</p>
                  <small v-if="card.item.timing"><Clock3 :size="13" />{{ card.item.timing }}</small>
                </div>
              </div>
            </div>
            <p v-else class="empty-copy">今日暂无具体宜用动作，守住一件小事即可。</p>
          </article>

          <article class="practice-panel cautious">
            <div class="panel-title"><ShieldAlert :size="17" /><span>今日慎行</span></div>
            <div v-if="cautionItems.length" class="caution-list">
              <div
                v-for="(item, index) in cautionItems"
                :key="`${item.label}-${item.value}`"
                class="caution-item"
              >
                <span>0{{ index + 1 }}</span>
                <div>
                  <strong>{{ item.value }}</strong>
                  <p>{{ item.method || item.reason }}</p>
                </div>
              </div>
            </div>
            <p v-else class="empty-copy">
              当前没有明显反向证据，承气以不过量、不勉强为原则。
            </p>

            <div class="color-ritual">
              <div>
                <span>今日随身宜色</span>
                <strong>{{ colorText }}</strong>
              </div>
              <div class="swatches" aria-hidden="true">
                <i v-for="chip in colorChips.slice(0, 3)" :key="chip"></i>
              </div>
            </div>

            <div class="object-ritual">
              <CalendarDays :size="17" />
              <div>
                <span>随身宜物</span>
                <strong>{{ objectChips.join(' · ') || profile.objects }}</strong>
              </div>
            </div>
          </article>
        </div>
      </section>

      <section class="time-section" aria-labelledby="time-title">
        <div class="time-copy">
          <p>PRACTICE RHYTHM · 承气节律</p>
          <h2 id="time-title">择一刻静驻，再起身行动</h2>
          <span>{{
            guide?.analysis || '在适合的时段完成最重要的一件事，让今日之气有处可落。'
          }}</span>
        </div>
        <div class="hour-track" v-if="hourItems.length">
          <div v-for="(hour, index) in hourItems" :key="`${hour.label}-${hour.value}`">
            <span>0{{ index + 1 }}</span>
            <strong>{{ hour.value }}</strong>
            <small>{{ hour.label }}</small>
          </div>
        </div>
        <div v-else class="hour-empty">今日不拘时，任选不受打扰的一刻即可。</div>
      </section>

      <section class="five-elements" aria-labelledby="elements-title">
        <header class="section-heading compact">
          <div>
            <p>FIVE ELEMENTS · 五行流转</p>
            <h2 id="elements-title">一气流行，生生不息</h2>
          </div>
          <span class="section-number">叁</span>
        </header>
        <div class="element-track">
          <article
            v-for="item in galleryItems"
            :key="item.key"
            :class="[
              { active: item.profile.element === profile.element },
              `orientation-${item.asset.orientation}`,
            ]"
            :style="{
              '--module-accent': item.profile.accent,
              '--module-rgb': item.profile.accentRgb,
            }"
          >
            <img
              :src="item.asset.thumbnail_url || item.asset.url"
              :alt="item.asset.alt_text"
              loading="lazy"
              decoding="async"
              :style="assetFocusStyle(item.asset)"
            />
            <div>
              <span>{{ item.profile.element }}行 · {{ item.asset.tone || '五行物象' }}</span>
              <strong>{{ item.asset.name || item.profile.objectLabel }}</strong>
              <small v-if="item.asset.description && item.asset.description !== item.asset.name">{{
                item.asset.description
              }}</small>
              <small v-else-if="!item.asset.description"
                >{{ item.profile.direction }} · {{ item.profile.colors }}</small
              >
            </div>
          </article>
        </div>
      </section>

      <footer class="blessing-footer">
        <span class="footer-seal">福</span>
        <div>
          <p>愿今日所行，皆有回响</p>
          <small v-if="guide?.evidence_completeness !== undefined"
            >排盘证据完整度 {{ guide.evidence_completeness }}%</small
          >
          <small v-else>以传统命理为参考，以清醒行动为根本</small>
        </div>
      </footer>
    </div>
  </div>
</template>

<style scoped>
.blessing-page {
  --blessing-accent: #22c59e;
  --blessing-accent-dark: #0f8f6e;
  --blessing-accent-rgb: 34, 197, 158;
  --blessing-backdrop: none;
  position: relative;
  min-height: calc(100vh - 72px);
  overflow: hidden;
  color: var(--text);
  background:
    radial-gradient(circle at 78% 6%, rgba(var(--blessing-accent-rgb), 0.09), transparent 27rem),
    linear-gradient(
      180deg,
      color-mix(in oklab, var(--surface-2) 42%, transparent),
      transparent 22rem
    ),
    var(--bg);
}

.ambient-scene {
  position: absolute;
  z-index: 0;
  inset: 0;
  pointer-events: none;
  overflow: hidden;
}

.ambient-image {
  position: absolute;
  top: -6rem;
  right: -8rem;
  width: min(44rem, 58vw);
  aspect-ratio: 1;
  border-radius: 50%;
  background-image: var(--blessing-backdrop);
  background-size: cover;
  background-position: center;
  opacity: 0.07;
  filter: saturate(0.8);
  mask-image: radial-gradient(circle, #000 20%, transparent 70%);
}

.ambient-wash {
  position: absolute;
  top: 5rem;
  left: -12rem;
  width: 30rem;
  height: 30rem;
  border-radius: 50%;
  background: rgba(var(--blessing-accent-rgb), 0.055);
  filter: blur(80px);
}

.qi-stream,
.grain {
  display: none;
}

.blessing-shell {
  position: relative;
  z-index: 1;
  width: min(1080px, calc(100% - 40px));
  margin: 0 auto;
  padding: 2rem 0 5rem;
}

.top-nav {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 1rem;
}

.nav-back,
.nav-periods a,
.nav-periods span {
  color: var(--text-muted);
  text-decoration: none;
  font-size: var(--fs-xs, 0.78rem);
}

.nav-back {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  min-height: 36px;
  padding: 0 0.25rem;
  font-weight: 650;
  transition:
    color 0.2s ease,
    transform 0.2s ease;
}

.nav-back:hover {
  color: var(--text);
  transform: translateX(-2px);
}

.nav-periods {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.25rem;
  border: 1px solid var(--line-strong);
  border-radius: 10px;
  background: color-mix(in oklab, var(--surface-0) 86%, transparent);
  box-shadow: var(--shadow-xs);
}

.nav-periods a,
.nav-periods span {
  padding: 0.5rem 0.8rem;
  border-radius: 7px;
  transition:
    color 0.2s ease,
    background 0.2s ease;
}

.nav-periods .active {
  color: var(--text);
  background: color-mix(in oklab, var(--blessing-accent) 15%, var(--surface-1));
  font-weight: 750;
  box-shadow: inset 0 0 0 1px rgba(var(--blessing-accent-rgb), 0.16);
}

.nav-periods a:hover {
  color: var(--text);
  background: var(--surface-2);
}

.blessing-hero {
  position: relative;
  width: 100%;
  min-height: 420px;
  overflow: visible;
  background: transparent;
  isolation: isolate;
}

.hero-inner {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(230px, 0.42fr);
  align-items: end;
  gap: clamp(2rem, 5vw, 4rem);
  box-sizing: border-box;
  width: 100%;
  min-height: 420px;
  padding: clamp(1.6rem, 4vw, 3rem) 0;
}

.hero-image {
  position: absolute;
  top: 0;
  bottom: 0;
  left: 50%;
  width: 100vw;
  height: 100%;
  transform: translateX(-50%);
}

.hero-image {
  z-index: -2;
  max-width: none;
  object-fit: cover;
  object-position: center;
  filter: saturate(0.82) contrast(0.94);
  transform: translateX(-50%) scale(1.01);
}

.hero-copy {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 1rem;
  min-width: 0;
  max-width: 42rem;
}

.section-heading p,
.time-copy > p {
  margin: 0;
  color: var(--blessing-accent-dark);
  font-size: 0.7rem;
  font-weight: 800;
  letter-spacing: 0.16em;
  text-transform: uppercase;
}

.eyebrow {
  margin: 0;
  color: var(--blessing-accent-dark);
  font-size: 0.76rem;
  font-weight: 800;
  letter-spacing: 0;
}

.hero-copy h1 {
  margin: 0;
  color: var(--text);
  font-family: var(--font-serif);
  font-size: clamp(3rem, 6vw, 5.25rem);
  font-weight: 900;
  line-height: 1;
  letter-spacing: 0;
}

.hero-strategy {
  max-width: 38rem;
  margin: 0;
  color: var(--text);
  font-size: 1rem;
  line-height: 1.75;
}

.hero-evidence {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.hero-evidence span {
  padding: 0.42rem 0.65rem;
  border: 1px solid color-mix(in oklab, var(--blessing-accent) 30%, var(--line-strong));
  border-radius: 6px;
  color: var(--text);
  background: color-mix(in oklab, var(--surface-0) 76%, transparent);
  font-size: 0.72rem;
  font-weight: 700;
}

.hero-presence {
  position: relative;
  display: grid;
  justify-items: end;
  align-content: end;
  gap: 1.2rem;
  min-width: 0;
}

.hero-element {
  color: var(--blessing-accent-dark);
  font-family: var(--font-serif);
  font-size: clamp(5.8rem, 10vw, 8rem);
  font-weight: 900;
  line-height: 0.88;
  text-shadow: 0 8px 28px color-mix(in oklab, var(--surface-0) 80%, transparent);
}

.presence-readout {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1.5rem;
  width: min(100%, 230px);
}

.presence-readout span {
  display: block;
  color: oklch(96% 0.01 155);
  font-size: 0.9rem;
  font-weight: 800;
  line-height: 1.3;
  text-shadow:
    0 1px 2px oklch(18% 0.02 155 / 0.9),
    0 0 8px oklch(18% 0.02 155 / 0.65);
}

.presence-readout strong {
  display: block;
  margin-top: 0.3rem;
  color: oklch(96% 0.01 155);
  font-family: var(--font-mono);
  font-size: 1.45rem;
  letter-spacing: 0;
  text-shadow:
    0 1px 2px oklch(18% 0.02 155 / 0.9),
    0 0 8px oklch(18% 0.02 155 / 0.65);
}

.field-toggle {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 auto;
  width: 44px;
  height: 44px;
  padding: 0;
  border: 1px solid rgba(var(--blessing-accent-rgb), 0.22);
  border-radius: 50%;
  color: var(--blessing-accent-dark);
  background: color-mix(in oklab, var(--surface-0) 78%, transparent);
  font: inherit;
  cursor: pointer;
  transition:
    transform 0.2s ease,
    background 0.2s ease;
}

.field-toggle:hover {
  transform: translateY(-1px);
  background: rgba(var(--blessing-accent-rgb), 0.14);
}

.ritual-section,
.daily-practice,
.five-elements {
  margin-top: 4rem;
}

.section-heading {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 1.35rem;
}

.section-heading h2,
.time-copy h2 {
  margin: 0.35rem 0 0;
  color: var(--text);
  font-family: var(--font-serif);
  font-size: clamp(1.55rem, 3vw, 2.15rem);
  line-height: 1.25;
}

.section-number {
  color: rgba(var(--blessing-accent-rgb), 0.26);
  font-family: var(--font-serif);
  font-size: 2.4rem;
  line-height: 1;
}

.ritual-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 1rem;
  align-items: stretch;
}

.ritual-card {
  position: relative;
  display: flex;
  flex-direction: column;
  min-width: 0;
  overflow: hidden;
  border: 1px solid var(--line-strong);
  border-radius: 16px;
  background: var(--surface-0);
  box-shadow: var(--shadow-sm);
  transition:
    border-color 0.25s ease,
    box-shadow 0.25s ease;
}

.ritual-card:hover {
  border-color: rgba(var(--module-rgb), 0.32);
  box-shadow: var(--shadow-md);
}

.ritual-image-wrap {
  position: relative;
  aspect-ratio: 4 / 5;
  height: auto;
  overflow: hidden;
  background: color-mix(in oklab, var(--module-accent) 8%, var(--surface-2));
}

.ritual-image-wrap::after {
  content: '';
  position: absolute;
  inset: auto 0 0;
  height: 28%;
  background: linear-gradient(180deg, transparent, var(--surface-0));
}

.ritual-image-wrap img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
  opacity: 0.88;
  transition: transform 0.45s ease;
}

.ritual-card:hover img {
  transform: scale(1.035);
}
.ritual-image-wrap span {
  position: absolute;
  z-index: 1;
  top: 0.75rem;
  left: 0.75rem;
  color: var(--text);
  font-family: var(--font-mono);
  font-size: 0.68rem;
  font-weight: 800;
}

.ritual-body {
  flex: 1;
  padding: 0.8rem 1.15rem 1.25rem;
}
.ritual-body p {
  margin: 0;
  color: var(--module-accent);
  font-size: 0.68rem;
  font-weight: 850;
  letter-spacing: 0.12em;
}
.ritual-card.avoid .ritual-body p {
  color: var(--crimson);
}
.ritual-body h3 {
  margin: 0.45rem 0 0.65rem;
  color: var(--text);
  font-family: var(--font-serif);
  font-size: 1.2rem;
}
.ritual-body strong {
  display: block;
  color: var(--text-muted);
  font-size: 0.83rem;
  line-height: 1.65;
}
.ritual-body small {
  display: block;
  margin-top: 0.7rem;
  color: var(--text-soft);
  font-size: 0.72rem;
  line-height: 1.65;
}

.practice-columns {
  display: grid;
  grid-template-columns: 1.45fr 0.8fr;
  gap: 1rem;
}

.practice-panel {
  padding: 1.35rem;
  border: 1px solid var(--line-strong);
  border-radius: 16px;
  background: var(--surface-0);
  box-shadow: var(--shadow-sm);
}

.practice-panel.favorable {
  border-top: 2px solid var(--blessing-accent);
}
.practice-panel.cautious {
  border-top: 2px solid color-mix(in oklab, var(--crimson) 72%, transparent);
}

.panel-title {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 1rem;
  color: var(--text);
  font-family: var(--font-serif);
  font-size: 1rem;
  font-weight: 850;
}

.favorable .panel-title svg {
  color: var(--blessing-accent-dark);
}
.cautious .panel-title svg {
  color: var(--crimson);
}

.practice-list {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 0.7rem;
}
.practice-item {
  display: grid;
  grid-template-columns: 58px minmax(0, 1fr);
  gap: 0.8rem;
  padding: 0.8rem;
  border: 1px solid var(--line-subtle);
  border-radius: 12px;
  background: var(--surface-1);
}
.practice-item img {
  width: 58px;
  height: 58px;
  border-radius: 10px;
  object-fit: cover;
  background: var(--surface-2);
}
.practice-item span {
  display: block;
  color: var(--blessing-accent-dark);
  font-size: 0.62rem;
  font-weight: 800;
}
.practice-item strong {
  display: block;
  margin: 0.15rem 0 0.25rem;
  color: var(--text);
  font-size: 0.82rem;
}
.practice-item p {
  margin: 0;
  color: var(--text-soft);
  font-size: 0.68rem;
  line-height: 1.55;
}
.practice-item small {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  margin-top: 0.35rem;
  color: var(--text-muted);
  font-size: 0.62rem;
}

.caution-list {
  display: grid;
  gap: 0.65rem;
}
.caution-item {
  display: grid;
  grid-template-columns: 28px minmax(0, 1fr);
  gap: 0.55rem;
  padding-bottom: 0.65rem;
  border-bottom: 1px solid var(--line-subtle);
}
.caution-item > span {
  color: color-mix(in oklab, var(--crimson) 70%, var(--text-soft));
  font-family: var(--font-mono);
  font-size: 0.65rem;
}
.caution-item strong {
  color: var(--text);
  font-size: 0.8rem;
}
.caution-item p {
  margin: 0.25rem 0 0;
  color: var(--text-soft);
  font-size: 0.68rem;
  line-height: 1.55;
}
.empty-copy {
  color: var(--text-muted);
  font-size: 0.78rem;
  line-height: 1.7;
}

.color-ritual,
.object-ritual {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  margin-top: 0.8rem;
  padding: 0.8rem;
  border-radius: 10px;
  background: var(--surface-2);
}

.color-ritual span,
.object-ritual span {
  display: block;
  color: var(--text-soft);
  font-size: 0.62rem;
}
.color-ritual strong,
.object-ritual strong {
  display: block;
  margin-top: 0.2rem;
  color: var(--text);
  font-size: 0.75rem;
}
.object-ritual {
  justify-content: flex-start;
}
.object-ritual svg {
  flex: 0 0 auto;
  color: var(--blessing-accent-dark);
}
.swatches {
  display: flex;
}
.swatches i {
  width: 18px;
  height: 18px;
  margin-left: -4px;
  border: 2px solid var(--surface-2);
  border-radius: 50%;
  background: var(--blessing-accent);
}
.swatches i:nth-child(2) {
  opacity: 0.72;
}
.swatches i:nth-child(3) {
  opacity: 0.45;
}

.time-section {
  display: grid;
  grid-template-columns: minmax(220px, 0.8fr) 1.5fr;
  gap: 2rem;
  align-items: center;
  margin-top: 4rem;
  padding: 1.6rem;
  border: 1px solid var(--line-strong);
  border-radius: 16px;
  background:
    linear-gradient(120deg, rgba(var(--blessing-accent-rgb), 0.08), transparent 40%),
    var(--surface-1);
  box-shadow: var(--shadow-sm);
}

.time-copy > span {
  display: block;
  margin-top: 0.7rem;
  color: var(--text-muted);
  font-size: 0.75rem;
  line-height: 1.65;
}
.hour-track {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  overflow: hidden;
  border: 1px solid var(--line-subtle);
  border-radius: 12px;
  background: var(--surface-0);
}
.hour-track > div {
  min-width: 0;
  padding: 1rem 0.8rem;
  border-right: 1px solid var(--line-subtle);
}
.hour-track > div:last-child {
  border-right: 0;
}
.hour-track span {
  display: block;
  color: var(--blessing-accent-dark);
  font-family: var(--font-mono);
  font-size: 0.62rem;
}
.hour-track strong {
  display: block;
  margin: 0.35rem 0;
  overflow: hidden;
  color: var(--text);
  font-family: var(--font-serif);
  font-size: 0.92rem;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.hour-track small {
  color: var(--text-soft);
  font-size: 0.62rem;
}
.hour-empty {
  color: var(--text-muted);
  font-size: 0.8rem;
}

.element-track {
  column-count: 3;
  column-gap: var(--space-sm);
}
.element-track article {
  display: inline-block;
  width: 100%;
  min-width: 0;
  margin: 0 0 var(--space-sm);
  padding: 0.7rem;
  border: 1px solid var(--line-subtle);
  border-radius: 13px;
  background: var(--surface-0);
  break-inside: avoid;
  vertical-align: top;
  transition:
    border-color 0.25s ease,
    box-shadow 0.25s ease;
}
.element-track article.active {
  border-color: rgba(var(--module-rgb), 0.42);
  box-shadow: 0 8px 28px rgba(var(--module-rgb), 0.12);
}
.element-track img {
  display: block;
  width: 100%;
  aspect-ratio: 4 / 3;
  height: auto;
  border-radius: 9px;
  object-fit: cover;
  background: var(--surface-2);
  opacity: 0.82;
  transition: opacity 0.25s ease;
}
.element-track article.active img {
  opacity: 1;
}
.element-track .orientation-landscape img {
  aspect-ratio: 16 / 9;
}
.element-track .orientation-panorama img {
  aspect-ratio: 12 / 5;
}
.element-track .orientation-square img {
  aspect-ratio: 1;
}
.element-track .orientation-portrait img {
  aspect-ratio: 4 / 5;
}
.element-track span {
  display: block;
  margin-top: 0.6rem;
  color: var(--module-accent);
  font-size: 0.62rem;
  font-weight: 800;
}
.element-track strong {
  display: block;
  margin: 0.16rem 0;
  color: var(--text);
  font-family: var(--font-serif);
  font-size: 0.9rem;
}
.element-track small {
  display: block;
  overflow: hidden;
  color: var(--text-soft);
  font-size: 0.6rem;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.blessing-footer {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.8rem;
  margin-top: 4rem;
  padding-top: 1.5rem;
  border-top: 1px solid var(--line-subtle);
  text-align: left;
}

.footer-seal {
  display: grid;
  place-items: center;
  width: 34px;
  height: 34px;
  border: 1px solid rgba(var(--blessing-accent-rgb), 0.3);
  border-radius: 9px;
  color: var(--blessing-accent-dark);
  background: rgba(var(--blessing-accent-rgb), 0.08);
  font-family: var(--font-serif);
  font-weight: 900;
}

.blessing-footer p {
  margin: 0;
  color: var(--text-muted);
  font-family: var(--font-serif);
  font-size: 0.86rem;
}
.blessing-footer small {
  display: block;
  margin-top: 0.15rem;
  color: var(--text-soft);
  font-size: 0.62rem;
}

.state-view {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-direction: column;
  gap: 0.55rem;
  min-height: calc(100vh - 72px);
  padding: 2rem;
  text-align: center;
}

.loading-seal,
.state-mark {
  display: grid;
  place-items: center;
  width: 72px;
  height: 72px;
  margin-bottom: 0.75rem;
  border: 1px solid var(--line-strong);
  border-radius: 50%;
  color: var(--blessing-accent-dark);
  background: var(--surface-0);
  box-shadow:
    0 0 0 12px rgba(var(--blessing-accent-rgb), 0.05),
    var(--shadow-md);
}

.loading-seal {
  animation: core-breathe 2.8s ease-in-out infinite;
}
.loading-seal span {
  font-family: var(--font-serif);
  font-size: 1.8rem;
  font-weight: 900;
}
.state-mark.warning {
  color: var(--crimson);
}
.state-view p {
  margin: 0;
  color: var(--text);
  font-family: var(--font-serif);
  font-size: 1.15rem;
  font-weight: 800;
}
.state-view small {
  color: var(--text-soft);
}
.state-actions {
  margin-top: 1rem;
}
.primary-action {
  display: inline-flex;
  align-items: center;
  min-height: 40px;
  padding: 0 1rem;
  border: 0;
  border-radius: 9px;
  color: #fff;
  background: var(--blessing-accent-dark);
  font: inherit;
  font-size: 0.78rem;
  font-weight: 800;
  text-decoration: none;
  cursor: pointer;
  box-shadow: 0 7px 22px rgba(var(--blessing-accent-rgb), 0.2);
}

.field-paused .loading-seal {
  animation-play-state: paused;
}

.nav-back:focus-visible,
.nav-periods a:focus-visible,
.field-toggle:focus-visible,
.primary-action:focus-visible {
  outline: 3px solid rgba(var(--blessing-accent-rgb), 0.25);
  outline-offset: 3px;
}

@media (max-width: 900px) {
  .hero-inner {
    grid-template-columns: minmax(0, 1fr) 180px;
    min-height: 390px;
  }
  .hero-element {
    font-size: 6rem;
  }
  .presence-readout {
    width: 180px;
  }
  .ritual-grid {
    grid-template-columns: 1fr;
  }
  .ritual-card {
    display: grid;
    grid-template-columns: 190px 1fr;
  }
  .ritual-image-wrap {
    aspect-ratio: auto;
    height: 100%;
    min-height: 180px;
  }
  .ritual-body {
    align-self: center;
    padding: 1.25rem;
  }
  .practice-columns {
    grid-template-columns: 1fr;
  }
  .time-section {
    grid-template-columns: 1fr;
    gap: 1rem;
  }
  .element-track {
    column-count: 2;
  }
}

@media (max-width: 640px) {
  .blessing-shell {
    width: calc(100% - 24px);
    padding: 1rem 0 5rem;
  }
  .top-nav {
    align-items: flex-start;
    flex-direction: column;
  }
  .nav-periods {
    width: 100%;
  }
  .nav-periods a,
  .nav-periods span {
    flex: 1;
    text-align: center;
  }
  .blessing-hero {
    min-height: 520px;
  }
  .hero-inner {
    grid-template-columns: 1fr;
    gap: 2rem;
    width: 100%;
    min-height: 520px;
    padding: 1.75rem 0;
  }
  .hero-copy h1 {
    font-size: 3.2rem;
  }
  .hero-strategy {
    font-size: 0.88rem;
  }
  .hero-evidence {
    gap: 0.4rem;
  }
  .hero-presence {
    display: flex;
    align-items: flex-end;
    justify-content: space-between;
    width: 100%;
  }
  .hero-element {
    font-size: 5.5rem;
  }
  .presence-readout {
    width: 170px;
  }
  .section-heading {
    align-items: flex-start;
  }
  .section-number {
    font-size: 1.8rem;
  }
  .ritual-section,
  .daily-practice,
  .five-elements,
  .time-section {
    margin-top: 3.2rem;
  }
  .ritual-card {
    display: block;
  }
  .ritual-image-wrap {
    aspect-ratio: 4 / 5;
    height: auto;
    min-height: 0;
  }
  .practice-list {
    grid-template-columns: 1fr;
  }
  .hour-track {
    grid-template-columns: repeat(2, 1fr);
  }
  .hour-track > div:nth-child(2) {
    border-right: 0;
  }
  .hour-track > div:nth-child(-n + 2) {
    border-bottom: 1px solid var(--line-subtle);
  }
  .element-track {
    display: flex;
    column-count: auto;
    gap: var(--space-sm);
    overflow-x: auto;
    padding-bottom: 0.5rem;
    scroll-snap-type: x mandatory;
  }
  .element-track article {
    flex: 0 0 176px;
    margin-bottom: 0;
    scroll-snap-align: start;
  }
}

@media (prefers-reduced-motion: reduce) {
  .blessing-shell,
  .loading-seal,
  .ritual-image-wrap img {
    animation: none !important;
    transition: none !important;
  }
}
</style>
