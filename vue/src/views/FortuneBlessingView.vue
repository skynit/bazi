<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { ArrowLeft, CalendarDays, Clock3, Compass, ShieldAlert, Sparkles } from '@lucide/vue'
import type { FortuneDay, FortuneGuideItem } from '../api/fortune'
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

const guide = computed(() => (fortune.value ? activeGuide(fortune.value) : undefined))
const element = computed(() => resolveBlessingElement(guide.value, fortune.value?.lucky_color))
const profile = computed(() => blessingProfiles[element.value])
const colorText = computed(() => guide.value?.lucky_colors?.[0]?.value || fortune.value?.lucky_color || profile.value.colors)
const colorChips = computed(() => splitGuideValues(colorText.value))
const elementAction = computed<FortuneGuideItem | undefined>(() => {
  const actions = guide.value?.recommended_actions ?? []
  return actions.find((item) => item.element === profile.value.element) ?? actions[0]
})
const objectText = computed(() => elementAction.value?.value || profile.value.objects)
const objectChips = computed(() => splitGuideValues(objectText.value))
const actionItems = computed(() => (guide.value?.recommended_actions ?? []).slice(0, 5))
const cautionItems = computed(() => (guide.value?.cautions ?? []).slice(0, 4))
const hourItems = computed(() => (guide.value?.best_hours ?? []).slice(0, 4))
const galleryProfiles = computed(() => Object.values(blessingProfiles))
const secondaryProfile = computed(() => profileForElement(guide.value?.secondary_element))
const avoidProfile = computed(() => profileForElement(guide.value?.avoid_element))
const pageStyle = computed<Record<string, string>>(() => ({
  '--blessing-accent': profile.value.accent,
  '--blessing-accent-dark': profile.value.accentDark,
  '--blessing-accent-rgb': profile.value.accentRgb,
}))
const navQuery = computed(() => (chartId.value ? { chart_id: chartId.value } : {}))
const keyStats = computed(() => [
  { label: '主气', value: profile.value.element, icon: Sparkles },
  { label: '财位', value: guide.value?.wealth_direction?.value || fortune.value?.wealth_direction || profile.value.direction, icon: Compass },
  { label: '朝向', value: guide.value?.face_direction?.value || profile.value.direction, icon: Compass },
  { label: '避开', value: guide.value?.avoid_direction?.value || profile.value.avoidDirection, icon: ShieldAlert },
])
const imageModules = computed(() => {
  const modules = [
    {
      key: 'primary',
      badge: '主气',
      title: `${profile.value.element}行加持`,
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
      title: `${secondaryProfile.value.element}行辅助`,
      text: secondaryProfile.value.objects,
      note: `用小面积${secondaryProfile.value.colors}或${secondaryProfile.value.objectLabel}补足执行余地。`,
      profile: secondaryProfile.value,
      tone: 'secondary',
    })
  }

  if (avoidProfile.value && avoidProfile.value.element !== profile.value.element) {
    modules.push({
      key: 'avoid',
      badge: '避忌',
      title: `少引动${avoidProfile.value.element}气`,
      text: avoidProfile.value.objects,
      note: `相关颜色、方位和物象今天减量，以避损为先。`,
      profile: avoidProfile.value,
      tone: 'avoid',
    })
  }

  return modules
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

function profileForCaution(item: FortuneGuideItem) {
  return profileForElement(item.element) ?? avoidProfile.value ?? profile.value
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
  if (value >= 85) return '大吉'
  if (value >= 70) return '良好'
  if (value >= 55) return '平稳'
  if (value >= 40) return '承压'
  return '收敛'
}

onMounted(fetchBlessing)
</script>

<template>
  <div class="blessing-page" :style="pageStyle">
    <div v-if="loading" class="state-view">
      <div class="state-mark">
        <Sparkles :size="24" />
      </div>
      <p>运势加持生成中</p>
    </div>

    <div v-else-if="error" class="state-view">
      <div class="state-mark warning">
        <ShieldAlert :size="24" />
      </div>
      <p>{{ error }}</p>
      <div class="state-actions">
        <router-link v-if="error === '请先创建命盘'" to="/chart/new" class="primary-action">
          去排盘
        </router-link>
        <button v-else type="button" class="primary-action" @click="fetchBlessing">
          重新加载
        </button>
      </div>
    </div>

    <main v-else-if="fortune" class="blessing-shell">
      <nav class="top-nav" aria-label="运势加持导航">
        <router-link :to="{ path: '/fortune', query: navQuery }" class="nav-link">
          <ArrowLeft :size="16" />
          今日运势
        </router-link>
        <router-link :to="{ path: '/fortune/weekly', query: navQuery }" class="nav-link muted-link">
          本周
        </router-link>
        <router-link :to="{ path: '/fortune/monthly', query: navQuery }" class="nav-link muted-link">
          本月
        </router-link>
      </nav>

      <section class="blessing-hero" aria-labelledby="blessing-title">
        <div class="hero-copy">
          <p class="eyebrow">{{ fortune.solar_date }} · {{ fortune.day_gan_zhi }}</p>
          <h1 id="blessing-title">今日运势加持</h1>
          <p class="hero-strategy">{{ guide?.strategy || `${profile.element}气为今日主轴，先取色与物，再落到一个可执行动作。` }}</p>
          <div class="hero-meta">
            <span class="score-pill">{{ fortune.score }} · {{ scoreWord(fortune.score) }}</span>
            <span v-if="guide?.confidence" class="score-pill quiet">{{ guide.confidence }}% 置信</span>
          </div>
        </div>

        <div class="hero-image-panel">
          <img :src="profile.image" :alt="profile.alt" class="hero-image" />
        </div>
      </section>

      <section class="image-module-section" aria-label="加持图片模块">
        <article
          v-for="module in imageModules"
          :key="module.key"
          class="image-module-card"
          :class="{ avoid: module.tone === 'avoid' }"
          :style="{ '--module-accent': module.profile.accent, '--module-rgb': module.profile.accentRgb }"
        >
          <img :src="module.profile.image" :alt="module.profile.alt" class="module-image" />
          <div class="module-body">
            <span>{{ module.badge }}</span>
            <strong>{{ module.title }}</strong>
            <p>{{ module.text }}</p>
            <em>{{ module.note }}</em>
          </div>
        </article>
      </section>

      <section class="blessing-grid" aria-label="今日加持信息">
        <article class="color-panel">
          <div class="panel-head">
            <Sparkles :size="18" />
            <span>推荐主色</span>
          </div>
          <strong>{{ colorText }}</strong>
          <div class="swatch-row">
            <span
              v-for="chip in colorChips"
              :key="chip"
              class="color-chip"
            >
              <span class="swatch"></span>{{ chip }}
            </span>
          </div>
          <p v-if="guide?.lucky_colors?.[0]?.reason">{{ guide.lucky_colors[0].reason }}</p>
        </article>

        <article class="object-panel">
          <div class="panel-head">
            <CalendarDays :size="18" />
            <span>加持物象</span>
          </div>
          <strong>{{ profile.objectLabel }}</strong>
          <div class="object-list">
            <span v-for="item in objectChips" :key="item">{{ item }}</span>
          </div>
          <p>{{ elementAction?.reason || `${profile.objects}可作为今日${profile.element}气的轻量承接。` }}</p>
        </article>

        <article v-for="stat in keyStats" :key="stat.label" class="stat-panel">
          <component :is="stat.icon" :size="18" />
          <span>{{ stat.label }}</span>
          <strong>{{ stat.value }}</strong>
        </article>
      </section>

      <section class="gallery-panel" aria-label="五行加持图鉴">
        <div class="section-title">
          <Sparkles :size="18" />
          <h2>五行加持图鉴</h2>
        </div>
        <div class="gallery-track">
          <article
            v-for="item in galleryProfiles"
            :key="item.element"
            class="gallery-card"
            :class="{ active: item.element === profile.element }"
            :style="{ '--module-accent': item.accent, '--module-rgb': item.accentRgb }"
          >
            <img :src="item.image" :alt="item.alt" />
            <div>
              <span>{{ item.element }}行</span>
              <strong>{{ item.objectLabel }}</strong>
              <p>{{ item.objects }}</p>
            </div>
          </article>
        </div>
      </section>

      <section class="practice-layout" aria-label="今日执行与避忌">
        <article class="practice-panel">
          <div class="section-title">
            <Sparkles :size="18" />
            <h2>宜用动作</h2>
          </div>
          <div v-if="actionItems.length" class="practice-list">
            <article v-for="item in actionItems" :key="`${item.label}-${item.value}`" class="practice-item">
              <img :src="profileForAction(item).image" :alt="profileForAction(item).alt" class="practice-thumb" />
              <div class="practice-copy">
                <span>{{ item.category || item.label }}</span>
                <strong>{{ item.value }}</strong>
                <p>{{ item.method || item.reason }}</p>
                <em v-if="item.timing">{{ item.timing }}</em>
              </div>
            </article>
          </div>
          <p v-else class="empty-copy">今日暂无具体宜用动作。</p>
        </article>

        <article class="practice-panel caution-panel">
          <div class="section-title">
            <ShieldAlert :size="18" />
            <h2>先避风险</h2>
          </div>
          <div v-if="cautionItems.length" class="practice-list">
            <article v-for="item in cautionItems" :key="`${item.label}-${item.value}`" class="practice-item caution">
              <img :src="profileForCaution(item).image" :alt="profileForCaution(item).alt" class="practice-thumb" />
              <div class="practice-copy">
                <span>{{ item.category || item.label }}</span>
                <strong>{{ item.value }}</strong>
                <p>{{ item.method || item.reason }}</p>
                <em v-if="item.timing">{{ item.timing }}</em>
              </div>
            </article>
          </div>
          <p v-else class="empty-copy">今日避忌以少引动{{ guide?.avoid_element || '忌神' }}为主。</p>
        </article>
      </section>

      <section class="time-band" aria-label="今日吉时">
        <div class="section-title">
          <Clock3 :size="18" />
          <h2>加持时段</h2>
        </div>
        <div v-if="hourItems.length" class="hour-strip">
          <span v-for="hour in hourItems" :key="`${hour.label}-${hour.value}`">{{ hour.value }}</span>
        </div>
        <p v-if="guide?.analysis" class="analysis-copy">{{ guide.analysis }}</p>
      </section>
    </main>
  </div>
</template>

<style scoped>
.blessing-page {
  --blessing-accent: #22c59e;
  --blessing-accent-dark: #0f8f6e;
  --blessing-accent-rgb: 34, 197, 158;
  min-height: calc(100vh - 80px);
  color: var(--text);
  background:
    linear-gradient(135deg, rgba(var(--blessing-accent-rgb), 0.16), transparent 38%),
    linear-gradient(180deg, color-mix(in oklab, var(--surface-0) 95%, var(--blessing-accent) 5%), var(--bg));
  overflow-x: hidden;
}

.state-view {
  min-height: calc(100vh - 80px);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 16px;
  padding: 24px;
  text-align: center;
  color: var(--text-muted);
}

.state-mark {
  width: 56px;
  height: 56px;
  display: grid;
  place-items: center;
  border-radius: 8px;
  color: var(--blessing-accent);
  background: rgba(var(--blessing-accent-rgb), 0.12);
  border: 1px solid rgba(var(--blessing-accent-rgb), 0.24);
}

.state-mark.warning {
  color: var(--danger);
  background: color-mix(in oklab, var(--danger) 12%, transparent);
  border-color: color-mix(in oklab, var(--danger) 28%, transparent);
}

.state-actions {
  display: flex;
  gap: 10px;
}

.primary-action,
.nav-link {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  min-height: 38px;
  border-radius: 8px;
  text-decoration: none;
  font-size: var(--fs-sm);
  font-weight: 700;
}

.primary-action {
  padding: 0 18px;
  border: 0;
  color: var(--bg);
  background: var(--blessing-accent);
  cursor: pointer;
}

.blessing-shell {
  width: min(1180px, calc(100% - 32px));
  margin: 0 auto;
  padding: 24px 0 72px;
}

.top-nav {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 18px;
}

.nav-link {
  padding: 0 14px;
  color: var(--text);
  background: color-mix(in oklab, var(--surface-1) 84%, transparent);
  border: 1px solid var(--line-subtle);
}

.muted-link {
  color: var(--text-muted);
}

.blessing-hero {
  display: grid;
  grid-template-columns: minmax(0, 0.95fr) minmax(360px, 1.05fr);
  align-items: stretch;
  gap: 18px;
  margin-bottom: 18px;
}

.hero-copy,
.hero-image-panel,
.image-module-card,
.color-panel,
.object-panel,
.stat-panel,
.practice-panel,
.time-band,
.gallery-panel {
  border-radius: 8px;
  border: 1px solid var(--line-strong);
  background: color-mix(in oklab, var(--surface-1) 88%, transparent);
  box-shadow: var(--shadow-lg), inset 0 1px 0 var(--line-subtle);
}

.hero-copy {
  min-height: 430px;
  padding: 38px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  position: relative;
  overflow: hidden;
}

.hero-copy::before {
  content: '';
  position: absolute;
  inset: 0 auto 0 0;
  width: 4px;
  background: linear-gradient(180deg, var(--blessing-accent), var(--crimson), var(--blessing-accent-dark));
}

.eyebrow {
  margin: 0 0 14px;
  color: var(--blessing-accent);
  font-size: var(--fs-sm);
  font-weight: 800;
}

.hero-copy h1 {
  margin: 0;
  color: var(--text);
  font-family: var(--font-serif), serif;
  font-size: var(--fs-hero);
  line-height: var(--lh-tight);
}

.hero-strategy {
  max-width: 560px;
  margin: 20px 0 0;
  color: var(--text-muted);
  font-size: var(--fs-lg);
  line-height: var(--lh-relaxed);
}

.hero-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-top: 26px;
}

.score-pill {
  display: inline-flex;
  align-items: center;
  min-height: 34px;
  padding: 0 12px;
  border-radius: 8px;
  background: rgba(var(--blessing-accent-rgb), 0.14);
  color: var(--blessing-accent);
  font-weight: 800;
}

.score-pill.quiet {
  color: var(--text-muted);
  background: var(--glass-bg);
  border: 1px solid var(--line-subtle);
}

.hero-image-panel {
  min-height: 430px;
  overflow: hidden;
  background:
    linear-gradient(135deg, rgba(var(--blessing-accent-rgb), 0.18), transparent),
    var(--surface-0);
}

.hero-image {
  width: 100%;
  height: 100%;
  min-height: 430px;
  object-fit: cover;
  display: block;
}

.image-module-section {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
  margin-bottom: 18px;
}

.image-module-card {
  --module-accent: var(--blessing-accent);
  --module-rgb: var(--blessing-accent-rgb);
  overflow: hidden;
  background:
    linear-gradient(135deg, rgba(var(--module-rgb), 0.16), transparent),
    color-mix(in oklab, var(--surface-1) 88%, transparent);
}

.image-module-card.avoid {
  background:
    linear-gradient(135deg, color-mix(in oklab, var(--danger) 14%, transparent), transparent),
    color-mix(in oklab, var(--surface-1) 88%, transparent);
}

.module-image {
  width: 100%;
  aspect-ratio: 16 / 10;
  object-fit: cover;
  display: block;
  border-bottom: 1px solid color-mix(in oklab, var(--module-accent) 28%, transparent);
}

.module-body {
  padding: 18px;
  display: grid;
  gap: 8px;
}

.module-body span,
.gallery-card span {
  color: var(--module-accent);
  font-size: var(--fs-xs);
  font-weight: 800;
}

.module-body strong,
.gallery-card strong {
  color: var(--text);
  font-family: var(--font-serif), serif;
  font-size: var(--fs-xl);
}

.module-body p,
.module-body em,
.gallery-card p {
  margin: 0;
  color: var(--text-muted);
  font-size: var(--fs-sm);
  line-height: var(--lh-body);
}

.module-body em {
  font-style: normal;
  color: var(--text-soft);
}

.blessing-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
  margin-bottom: 18px;
}

.color-panel,
.object-panel {
  grid-column: span 2;
  padding: 22px;
}

.panel-head,
.section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--blessing-accent);
  font-size: var(--fs-sm);
  font-weight: 800;
}

.color-panel strong,
.object-panel strong {
  display: block;
  margin: 18px 0 14px;
  color: var(--text);
  font-family: var(--font-serif), serif;
  font-size: var(--fs-4xl);
  line-height: var(--lh-snug);
}

.color-panel p,
.object-panel p,
.practice-item p,
.analysis-copy,
.empty-copy {
  color: var(--text-muted);
  font-size: var(--fs-sm);
  line-height: var(--lh-body);
}

.swatch-row,
.object-list,
.hour-strip {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.color-chip,
.object-list span,
.hour-strip span {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  min-height: 32px;
  padding: 0 10px;
  border-radius: 8px;
  color: var(--text);
  background: var(--glass-bg);
  border: 1px solid var(--line-subtle);
  font-size: var(--fs-sm);
  font-weight: 700;
}

.swatch {
  width: 14px;
  height: 14px;
  border-radius: 50%;
  background: var(--blessing-accent);
  box-shadow: 0 0 0 3px rgba(var(--blessing-accent-rgb), 0.16);
}

.stat-panel {
  min-height: 132px;
  padding: 18px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  color: var(--text-muted);
}

.stat-panel svg {
  color: var(--blessing-accent);
}

.stat-panel strong {
  color: var(--text);
  font-family: var(--font-serif), serif;
  font-size: var(--fs-2xl);
}

.gallery-panel {
  padding: 24px;
  margin-bottom: 18px;
}

.gallery-track {
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  gap: 10px;
  margin-top: 16px;
}

.gallery-card {
  --module-accent: var(--blessing-accent);
  --module-rgb: var(--blessing-accent-rgb);
  min-width: 0;
  overflow: hidden;
  border-radius: 8px;
  border: 1px solid var(--line-subtle);
  background: color-mix(in oklab, var(--surface-2) 70%, transparent);
  transition: border-color 0.2s ease, transform 0.2s ease, box-shadow 0.2s ease;
}

.gallery-card.active,
.gallery-card:hover {
  transform: translateY(-2px);
  border-color: color-mix(in oklab, var(--module-accent) 54%, transparent);
  box-shadow: 0 18px 42px rgba(var(--module-rgb), 0.14);
}

.gallery-card img {
  width: 100%;
  aspect-ratio: 4 / 3;
  object-fit: cover;
  display: block;
}

.gallery-card div {
  padding: 12px;
  display: grid;
  gap: 4px;
}

.gallery-card strong {
  font-size: var(--fs-lg);
}

.gallery-card p {
  display: -webkit-box;
  overflow: hidden;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.practice-layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  gap: 12px;
  margin-bottom: 18px;
}

.practice-panel,
.time-band {
  padding: 24px;
}

.section-title h2 {
  margin: 0;
  color: var(--text);
  font-size: var(--fs-xl);
}

.practice-list {
  display: grid;
  gap: 10px;
  margin-top: 16px;
}

.practice-item {
  padding: 16px;
  border-radius: 8px;
  background: color-mix(in oklab, var(--surface-2) 72%, transparent);
  border: 1px solid rgba(var(--blessing-accent-rgb), 0.2);
  display: grid;
  grid-template-columns: 96px minmax(0, 1fr);
  gap: 14px;
  align-items: start;
}

.practice-item.caution {
  border-color: color-mix(in oklab, var(--danger) 24%, transparent);
}

.practice-thumb {
  width: 96px;
  aspect-ratio: 1 / 1;
  object-fit: cover;
  display: block;
  border-radius: 8px;
  border: 1px solid rgba(var(--blessing-accent-rgb), 0.22);
}

.practice-copy {
  min-width: 0;
}

.practice-item span {
  color: var(--blessing-accent);
  font-size: var(--fs-xs);
  font-weight: 800;
}

.practice-item.caution span {
  color: var(--danger);
}

.practice-item strong {
  display: block;
  margin-top: 6px;
  color: var(--text);
  font-size: var(--fs-lg);
}

.practice-item p {
  margin: 8px 0 0;
}

.practice-item em {
  display: inline-flex;
  margin-top: 10px;
  color: var(--text-soft);
  font-size: var(--fs-xs);
  font-style: normal;
}

.time-band {
  display: grid;
  gap: 16px;
}

.analysis-copy {
  margin: 0;
  border-left: 3px solid var(--blessing-accent);
  padding-left: 14px;
}

:global(.dark) .hero-copy,
:global(.dark) .hero-image-panel,
:global(.dark) .image-module-card,
:global(.dark) .color-panel,
:global(.dark) .object-panel,
:global(.dark) .stat-panel,
:global(.dark) .practice-panel,
:global(.dark) .time-band,
:global(.dark) .gallery-panel {
  background: color-mix(in oklab, var(--surface-1) 80%, #000 20%);
}

@media (max-width: 900px) {
  .blessing-hero,
  .practice-layout {
    grid-template-columns: 1fr;
  }

  .image-module-section,
  .blessing-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .gallery-track {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .hero-copy,
  .hero-image-panel,
  .hero-image {
    min-height: 340px;
  }
}

@media (max-width: 620px) {
  .blessing-shell {
    width: min(100% - 24px, 1180px);
    padding-top: 16px;
  }

  .top-nav {
    overflow-x: auto;
    padding-bottom: 2px;
  }

  .blessing-grid {
    grid-template-columns: 1fr;
  }

  .image-module-section,
  .gallery-track {
    grid-template-columns: 1fr;
  }

  .color-panel,
  .object-panel {
    grid-column: span 1;
  }

  .hero-copy,
  .practice-panel,
  .time-band,
  .gallery-panel {
    padding: 20px;
  }

  .hero-copy h1 {
    font-size: var(--fs-4xl);
  }

  .hero-strategy {
    font-size: var(--fs-md);
  }

  .practice-item {
    grid-template-columns: 76px minmax(0, 1fr);
    gap: 12px;
  }

  .practice-thumb {
    width: 76px;
  }
}

@media (prefers-reduced-motion: reduce) {
  *,
  *::before,
  *::after {
    transition-duration: 0.01ms !important;
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
  }
}
</style>
