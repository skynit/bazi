<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import {
  ArrowLeft,
  CalendarDays,
  Clock3,
  Compass,
  Leaf,
  Pause,
  Play,
  ShieldAlert,
  Sparkles,
} from '@lucide/vue'
import type { ElementAsset, FortuneDay, FortuneGuideItem } from '../api/fortune'
import { fetchDaily } from '../api/fortune'
import { activeGuide, todayString } from '../lib/fortuneGuide'
import {
  blessingProfiles,
  isBlessingElement,
  resolveBlessingElement,
  splitGuideValues,
  type BlessingProfile,
} from '../lib/blessing'

const route = useRoute()
const fortune = ref<FortuneDay | null>(null)
const chartId = ref('')
const loading = ref(true)
const error = ref('')
const presenceSeconds = ref(0)
const fieldActive = ref(true)
let presenceTimer: ReturnType<typeof setInterval> | undefined

const guide = computed(() => (fortune.value ? activeGuide(fortune.value) : undefined))
const element = computed(() => resolveBlessingElement(guide.value, fortune.value?.lucky_color))
const profile = computed(() => blessingProfiles[element.value])
const colorText = computed(
  () => guide.value?.lucky_colors?.[0]?.value || fortune.value?.lucky_color || profile.value.colors,
)
const colorChips = computed(() => splitGuideValues(colorText.value))
const elementAction = computed<FortuneGuideItem | undefined>(() => {
  const actions = guide.value?.recommended_actions ?? []
  return actions.find((item) => item.element === profile.value.element) ?? actions[0]
})
const objectText = computed(() => elementAction.value?.value || profile.value.objects)
const objectChips = computed(() => splitGuideValues(objectText.value))
const actionItems = computed(() => (guide.value?.recommended_actions ?? []).slice(0, 4))
const actionCards = computed(() =>
  actionItems.value.map((item, index) => ({
    item,
    asset: fortune.value?.blessing_assets?.actions?.[index],
  })),
)
const cautionItems = computed(() => (guide.value?.cautions ?? []).slice(0, 3))
const hourItems = computed(() => (guide.value?.best_hours ?? []).slice(0, 4))
const galleryProfiles = computed(() => Object.values(blessingProfiles))
const heroAsset = computed(() => fortune.value?.blessing_assets?.hero)
const galleryItems = computed(() => {
  const assets = fortune.value?.blessing_assets?.gallery ?? []
  if (assets.length) {
    return assets.map((asset) => ({
      key: `asset-${asset.id}`,
      asset,
      profile: profileForElement(asset.element) ?? profile.value,
    }))
  }
  return galleryProfiles.value.map((item) => ({
    key: `fallback-${item.element}`,
    asset: {
      id: 0,
      url: item.image,
      alt_text: item.alt,
      name: item.objectLabel,
      element: item.element,
      orientation: 'square',
      focal_x: 0.5,
      focal_y: 0.5,
    } as ElementAsset,
    profile: item,
  }))
})
const secondaryProfile = computed(() => profileForElement(guide.value?.secondary_element))
const avoidProfile = computed(() => profileForElement(guide.value?.avoid_element))
const pageStyle = computed<Record<string, string>>(() => ({
  '--blessing-accent': profile.value.accent,
  '--blessing-accent-dark': profile.value.accentDark,
  '--blessing-accent-rgb': profile.value.accentRgb,
  '--blessing-backdrop': `url(${heroAsset.value?.url || profile.value.backdrop})`,
}))
const navQuery = computed(() => (chartId.value ? { chart_id: chartId.value } : {}))
const keyStats = computed(() => [
  {
    label: '今日主气',
    value: profile.value.element,
    note: `${profile.value.element}行当令`,
    icon: Sparkles,
  },
  {
    label: '承气方位',
    value: guide.value?.face_direction?.value || profile.value.direction,
    note: guide.value?.wealth_direction?.value
      ? `财位 ${guide.value.wealth_direction.value}`
      : `宜向${profile.value.direction}`,
    icon: Compass,
  },
  {
    label: '今日宜色',
    value: colorChips.value[0] || profile.value.colors,
    note: colorChips.value.slice(1).join(' · ') || profile.value.colors,
    icon: Leaf,
  },
  {
    label: '减量方位',
    value: guide.value?.avoid_direction?.value || profile.value.avoidDirection,
    note: `少引${guide.value?.avoid_element || avoidProfile.value?.element || '杂'}气`,
    icon: ShieldAlert,
  },
])
const presenceText = computed(() => {
  const minutes = Math.floor(presenceSeconds.value / 60)
  const seconds = presenceSeconds.value % 60
  return `${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`
})
const presenceMessage = computed(() => {
  if (presenceSeconds.value < 60) return '静驻片刻，让呼吸先慢下来'
  if (presenceSeconds.value < 180) return '一息一念，正在与今日主气同频'
  if (presenceSeconds.value < 600) return '气定则神闲，把今日所愿放在心上'
  return '久坐不滞，记得起身舒展，再回来承气'
})
const imageModules = computed(() => {
  const modules = [
    {
      key: 'primary',
      badge: '主气',
      title: `${profile.value.element}行入场`,
      text: objectText.value,
      note: elementAction.value?.method || `${profile.value.objects}，用一个小动作承接今日主轴。`,
      profile: profile.value,
      tone: 'primary',
    },
  ]

  if (secondaryProfile.value && secondaryProfile.value.element !== profile.value.element) {
    modules.push({
      key: 'secondary',
      badge: '辅气',
      title: `${secondaryProfile.value.element}行相生`,
      text: secondaryProfile.value.objects,
      note: `用小面积${secondaryProfile.value.colors}或${secondaryProfile.value.objectLabel}补足执行余地。`,
      profile: secondaryProfile.value,
      tone: 'secondary',
    })
  }

  if (avoidProfile.value && avoidProfile.value.element !== profile.value.element) {
    modules.push({
      key: 'avoid',
      badge: '收敛',
      title: `${avoidProfile.value.element}气减量`,
      text: avoidProfile.value.objects,
      note: '相关颜色、方位和物象今日不必过量，以收敛守成为先。',
      profile: avoidProfile.value,
      tone: 'avoid',
    })
  }

  return modules.map((module, index) => ({
    ...module,
    asset: fortune.value?.blessing_assets?.ritual?.[index],
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

function profileForAction(item: FortuneGuideItem) {
  return profileForElement(item.element) ?? profile.value
}

function assetFocusStyle(asset?: ElementAsset) {
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

function scoreWord(score?: number) {
  const value = score ?? 0
  if (value >= 85) return '顺势明显'
  if (value >= 70) return '良好'
  if (value >= 55) return '平稳'
  if (value >= 40) return '承压'
  return '收敛'
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
        <div class="hero-copy">
          <p class="eyebrow"><span></span>{{ fortune.solar_date }} · {{ fortune.day_gan_zhi }}</p>
          <div class="hero-element-mark" aria-hidden="true">{{ profile.element }}</div>
          <h1 id="blessing-title">
            今日宜承<br />
            <em>{{ profile.element }}</em
            >行之气
          </h1>
          <p class="hero-strategy">
            {{
              guide?.strategy ||
              `${profile.element}气为今日主轴。取一抹宜色，择一件宜物，做一件笃定的小事。`
            }}
          </p>
          <div class="hero-meta">
            <span class="score-seal"
              ><b>{{ fortune.score }}</b
              >{{ scoreWord(fortune.score) }}</span
            >
            <span class="meta-divider"></span>
            <span>主气 {{ profile.element }}</span>
            <span>宜向 {{ guide?.face_direction?.value || profile.direction }}</span>
          </div>
          <p class="score-note">传统规则提供的是今日倾向与行动参考，不代表事件发生概率。</p>
        </div>

        <div class="qi-sanctuary" :aria-label="`${profile.element}行持续加持场`">
          <div class="orbit orbit-outer" aria-hidden="true">
            <span v-for="item in galleryProfiles" :key="item.element">{{ item.element }}</span>
          </div>
          <div class="orbit orbit-inner" aria-hidden="true"></div>
          <div class="qi-core" aria-hidden="true">
            <div class="qi-core-glow"></div>
            <span>{{ profile.element }}</span>
            <small>今日主气</small>
          </div>
          <div class="presence-card">
            <div>
              <span class="presence-kicker">此刻 · 气场相伴</span>
              <strong>{{ presenceText }}</strong>
              <p>{{ presenceMessage }}</p>
            </div>
            <button
              type="button"
              class="field-toggle"
              :aria-label="fieldActive ? '暂停气场动画' : '继续气场动画'"
              :aria-pressed="!fieldActive"
              @click="toggleField"
            >
              <Pause v-if="fieldActive" :size="16" />
              <Play v-else :size="16" />
              {{ fieldActive ? '静驻中' : '继续承气' }}
            </button>
          </div>
        </div>
      </section>

      <section class="stats-ribbon" aria-label="今日五行要点">
        <article v-for="stat in keyStats" :key="stat.label">
          <component :is="stat.icon" :size="17" />
          <div>
            <span>{{ stat.label }}</span>
            <strong>{{ stat.value }}</strong>
            <small>{{ stat.note }}</small>
          </div>
        </article>
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
              少引动{{ guide?.avoid_element || '忌神' }}，不急、不满、不逞强。
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
          <p>BEST HOURS · 吉时</p>
          <h2 id="time-title">择时而动，事半功倍</h2>
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
        <div v-else class="hour-empty">今日不拘时，心定即是吉时。</div>
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
          <small v-if="guide?.evidence_completeness"
            >规则证据完整度 {{ guide.evidence_completeness }}%</small
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
  animation: page-enter 0.65s ease both;
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
  display: grid;
  grid-template-columns: minmax(0, 1.15fr) minmax(330px, 0.85fr);
  align-items: stretch;
  gap: 1.5rem;
  min-height: 430px;
  padding: clamp(1.6rem, 4vw, 3rem);
  overflow: hidden;
  border: 1px solid var(--line-strong);
  border-radius: 22px;
  background: color-mix(in oklab, var(--surface-1) 91%, transparent);
  box-shadow:
    var(--shadow-lg),
    inset 0 1px 0 var(--line-subtle);
}

.blessing-hero::before {
  content: '';
  position: absolute;
  inset: 0 auto 0 0;
  width: 3px;
  background: linear-gradient(
    180deg,
    transparent,
    var(--blessing-accent),
    var(--blessing-accent-dark),
    transparent
  );
}

.blessing-hero::after {
  content: '';
  position: absolute;
  z-index: 0;
  inset: 0 0 0 48%;
  background:
    linear-gradient(90deg, var(--surface-1), transparent 48%),
    linear-gradient(180deg, transparent 55%, var(--surface-1)),
    var(--blessing-backdrop) center / cover;
  opacity: 0.11;
  pointer-events: none;
}

.hero-copy,
.qi-sanctuary {
  position: relative;
  z-index: 1;
}

.hero-copy {
  display: flex;
  flex-direction: column;
  justify-content: center;
  min-width: 0;
}

.eyebrow,
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
  display: flex;
  align-items: center;
  gap: 0.65rem;
}

.eyebrow span {
  width: 24px;
  height: 2px;
  border-radius: 99px;
  background: var(--blessing-accent);
}

.hero-element-mark {
  position: absolute;
  top: -2rem;
  right: 0;
  color: rgba(var(--blessing-accent-rgb), 0.07);
  font-family: var(--font-serif);
  font-size: clamp(8rem, 16vw, 12rem);
  font-weight: 900;
  line-height: 1;
  pointer-events: none;
}

.hero-copy h1 {
  position: relative;
  margin: 1rem 0 1.15rem;
  color: var(--text);
  font-family: var(--font-serif);
  font-size: clamp(2.6rem, 5.6vw, 4.6rem);
  font-weight: 900;
  line-height: 1.08;
  letter-spacing: -0.045em;
}

.hero-copy h1 em {
  color: var(--blessing-accent-dark);
  font-style: normal;
  text-shadow: 0 8px 30px rgba(var(--blessing-accent-rgb), 0.18);
}

.hero-strategy {
  max-width: 34rem;
  margin: 0;
  color: var(--text-muted);
  font-size: 0.96rem;
  line-height: 1.85;
}

.hero-meta {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.75rem;
  margin-top: 1.5rem;
  color: var(--text-muted);
  font-size: 0.78rem;
  font-weight: 650;
}

.score-seal {
  display: inline-flex;
  align-items: baseline;
  gap: 0.35rem;
  padding: 0.45rem 0.65rem;
  border: 1px solid rgba(var(--blessing-accent-rgb), 0.22);
  border-radius: 8px;
  color: var(--blessing-accent-dark);
  background: rgba(var(--blessing-accent-rgb), 0.08);
}

.score-seal b {
  font-family: var(--font-serif);
  font-size: 1.15rem;
}

.meta-divider {
  width: 1px;
  height: 18px;
  background: var(--line-strong);
}

.score-note {
  margin: 0.85rem 0 0;
  color: var(--text-soft);
  font-size: 0.7rem;
  line-height: 1.55;
}

.qi-sanctuary {
  display: grid;
  place-items: center;
  min-height: 350px;
}

.orbit {
  position: absolute;
  border: 1px solid rgba(var(--blessing-accent-rgb), 0.2);
  border-radius: 50%;
}

.orbit-outer {
  width: min(310px, 90%);
  aspect-ratio: 1;
  animation: orbit-turn 28s linear infinite;
}

.orbit-outer::before,
.orbit-outer::after {
  content: '';
  position: absolute;
  border-radius: 50%;
  background: var(--blessing-accent);
  box-shadow: 0 0 18px rgba(var(--blessing-accent-rgb), 0.55);
}

.orbit-outer::before {
  top: -3px;
  left: 50%;
  width: 6px;
  height: 6px;
}

.orbit-outer::after {
  right: 8%;
  bottom: 19%;
  width: 4px;
  height: 4px;
}

.orbit-outer span {
  position: absolute;
  display: grid;
  place-items: center;
  width: 28px;
  height: 28px;
  border: 1px solid var(--line-strong);
  border-radius: 50%;
  color: var(--text-muted);
  background: var(--surface-0);
  font-family: var(--font-serif);
  font-size: 0.72rem;
  box-shadow: var(--shadow-sm);
  animation: orbit-counter 28s linear infinite;
}

.orbit-outer span:nth-child(1) {
  top: 3%;
  left: 19%;
}
.orbit-outer span:nth-child(2) {
  top: 17%;
  right: 1%;
}
.orbit-outer span:nth-child(3) {
  right: 8%;
  bottom: 10%;
}
.orbit-outer span:nth-child(4) {
  left: 15%;
  bottom: 2%;
}
.orbit-outer span:nth-child(5) {
  top: 43%;
  left: -5%;
}

.orbit-inner {
  width: min(240px, 70%);
  aspect-ratio: 1;
  border-style: dashed;
  opacity: 0.55;
  animation: orbit-turn 18s linear infinite reverse;
}

.qi-core {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-direction: column;
  width: 156px;
  aspect-ratio: 1;
  border: 1px solid rgba(var(--blessing-accent-rgb), 0.36);
  border-radius: 50%;
  background:
    radial-gradient(circle at 38% 30%, rgba(255, 255, 255, 0.65), transparent 18%),
    radial-gradient(
      circle,
      rgba(var(--blessing-accent-rgb), 0.28),
      rgba(var(--blessing-accent-rgb), 0.07) 58%,
      var(--surface-0)
    );
  box-shadow:
    0 18px 50px rgba(var(--blessing-accent-rgb), 0.18),
    inset 0 0 35px rgba(var(--blessing-accent-rgb), 0.14);
  animation: core-breathe 6s ease-in-out infinite;
}

.qi-core-glow {
  position: absolute;
  inset: -28px;
  border-radius: 50%;
  background: radial-gradient(circle, rgba(var(--blessing-accent-rgb), 0.16), transparent 68%);
  z-index: -1;
}

.qi-core > span {
  color: var(--blessing-accent-dark);
  font-family: var(--font-serif);
  font-size: 4.2rem;
  font-weight: 900;
  line-height: 1;
}

.qi-core small {
  margin-top: 0.35rem;
  color: var(--text-muted);
  font-size: 0.65rem;
  font-weight: 750;
  letter-spacing: 0.18em;
}

.presence-card {
  position: absolute;
  right: 0;
  bottom: 0;
  left: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.9rem 1rem;
  border: 1px solid var(--line-strong);
  border-radius: 12px;
  background: color-mix(in oklab, var(--surface-0) 88%, transparent);
  box-shadow: var(--shadow-md);
  backdrop-filter: blur(14px);
}

.presence-kicker {
  display: block;
  color: var(--text-soft);
  font-size: 0.65rem;
  font-weight: 750;
  letter-spacing: 0.08em;
}

.presence-card strong {
  display: inline-block;
  margin-top: 0.15rem;
  color: var(--text);
  font-family: var(--font-mono);
  font-size: 1.15rem;
}

.presence-card p {
  margin: 0.15rem 0 0;
  color: var(--text-muted);
  font-size: 0.68rem;
}

.field-toggle {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.35rem;
  flex: 0 0 auto;
  min-height: 34px;
  padding: 0 0.75rem;
  border: 1px solid rgba(var(--blessing-accent-rgb), 0.22);
  border-radius: 8px;
  color: var(--blessing-accent-dark);
  background: rgba(var(--blessing-accent-rgb), 0.08);
  font: inherit;
  font-size: 0.7rem;
  font-weight: 750;
  cursor: pointer;
  transition:
    transform 0.2s ease,
    background 0.2s ease;
}

.field-toggle:hover {
  transform: translateY(-1px);
  background: rgba(var(--blessing-accent-rgb), 0.14);
}

.stats-ribbon {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  margin: 1rem 0 4rem;
  overflow: hidden;
  border: 1px solid var(--line-strong);
  border-radius: 16px;
  background: var(--surface-0);
  box-shadow: var(--shadow-sm);
}

.stats-ribbon article {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  min-width: 0;
  padding: 1.15rem;
  border-right: 1px solid var(--line-subtle);
}

.stats-ribbon article:last-child {
  border-right: 0;
}
.stats-ribbon svg {
  flex: 0 0 auto;
  margin-top: 0.12rem;
  color: var(--blessing-accent-dark);
}
.stats-ribbon span,
.stats-ribbon small {
  display: block;
  color: var(--text-soft);
  font-size: 0.66rem;
}
.stats-ribbon strong {
  display: block;
  margin: 0.2rem 0;
  color: var(--text);
  font-family: var(--font-serif);
  font-size: 1rem;
}
.stats-ribbon small {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
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

.field-paused .orbit,
.field-paused .orbit-outer span,
.field-paused .qi-core,
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

@keyframes page-enter {
  from {
    opacity: 0;
    transform: translateY(14px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes orbit-turn {
  to {
    transform: rotate(360deg);
  }
}
@keyframes orbit-counter {
  to {
    transform: rotate(-360deg);
  }
}
@keyframes core-breathe {
  0%,
  100% {
    transform: scale(0.98);
    box-shadow:
      0 14px 42px rgba(var(--blessing-accent-rgb), 0.13),
      inset 0 0 30px rgba(var(--blessing-accent-rgb), 0.1);
  }
  50% {
    transform: scale(1.035);
    box-shadow:
      0 20px 62px rgba(var(--blessing-accent-rgb), 0.24),
      inset 0 0 42px rgba(var(--blessing-accent-rgb), 0.18);
  }
}

@media (max-width: 900px) {
  .blessing-hero {
    grid-template-columns: 1fr;
  }
  .hero-copy {
    text-align: center;
    align-items: center;
  }
  .eyebrow {
    justify-content: center;
  }
  .hero-element-mark {
    right: 50%;
    transform: translateX(50%);
  }
  .hero-meta {
    justify-content: center;
  }
  .qi-sanctuary {
    min-height: 380px;
  }
  .stats-ribbon {
    grid-template-columns: repeat(2, 1fr);
  }
  .stats-ribbon article:nth-child(2) {
    border-right: 0;
  }
  .stats-ribbon article:nth-child(-n + 2) {
    border-bottom: 1px solid var(--line-subtle);
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
    width: min(100% - 24px, 1080px);
    padding-top: 1rem;
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
    min-height: 0;
    padding: 1.45rem 1.1rem;
    border-radius: 17px;
  }
  .blessing-hero::after {
    inset: 46% 0 0;
  }
  .hero-copy h1 {
    font-size: clamp(2.4rem, 14vw, 3.45rem);
  }
  .hero-strategy {
    font-size: 0.88rem;
  }
  .hero-meta {
    gap: 0.55rem;
  }
  .meta-divider {
    display: none;
  }
  .qi-sanctuary {
    min-height: 350px;
  }
  .orbit-outer {
    width: 260px;
  }
  .orbit-inner {
    width: 200px;
  }
  .qi-core {
    width: 138px;
  }
  .qi-core > span {
    font-size: 3.6rem;
  }
  .presence-card {
    align-items: flex-start;
    flex-direction: column;
  }
  .field-toggle {
    width: 100%;
  }
  .stats-ribbon {
    margin-bottom: 3.2rem;
  }
  .stats-ribbon article {
    padding: 0.9rem;
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
  .orbit,
  .orbit-outer span,
  .qi-core,
  .loading-seal,
  .ritual-image-wrap img {
    animation: none !important;
    transition: none !important;
  }
}
</style>
