<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import DailyFortune from '../components/DailyFortune.vue'
import PeriodNav from '../components/fortune/PeriodNav.vue'
import FortuneStateView from '../components/fortune/FortuneStateView.vue'
import ScoreOrb from '../components/fortune/ScoreOrb.vue'
import { vReveal } from '../composables/useReveal'
import type { FortuneDay } from '../api/fortune'
import { fetchDaily } from '../api/fortune'
import { getApiErrorMessage } from '../api/client'
import { useRecentChartStore } from '../stores/recentChart'
import {
  toneBranchOf,
  toneCopyOf,
  scoreTierOf,
  TEN_GOD_GUIDE,
  TWELVE_STAGE_PLAIN,
  SEASONAL_STATE_PLAIN,
} from '../lib/fortuneCopy'

const route = useRoute()
const recentChartStore = useRecentChartStore()
const fortune = ref<FortuneDay | null>(null)
const loading = ref(true)
const error = ref('')
const errorKind = ref<'missing-chart' | 'request' | ''>('')
const chartId = ref<string | number>('')

/** 结构指数：后端流水线的 final_score，中性起分 base_score（通常 50） */
const finalScore = computed(
  () => fortune.value?.score_breakdown?.final_score ?? fortune.value?.score ?? 0,
)
const baseScore = computed(() => fortune.value?.score_breakdown?.base_score ?? 50)
const supportingCount = computed(() => fortune.value?.supporting_evidence?.length ?? 0)
const counterCount = computed(() => fortune.value?.counter_evidence?.length ?? 0)

/** 一句话大白话基调：标题 + 白话解释（术语仅以括号对照出现） */
const tone = computed(() =>
  toneCopyOf(
    toneBranchOf(supportingCount.value, counterCount.value),
    supportingCount.value,
    counterCount.value,
  ),
)

/** 今日状态分的人话刻度：分档标签 + 平常日参照，不让用户自己换算基线 */
const scoreTier = computed(() => scoreTierOf(finalScore.value, baseScore.value))

/** 十神 → 今日关键词 + 行动建议 */
const tenGodName = computed(() =>
  fortune.value?.ten_god?.status === 'observed' ? (fortune.value.ten_god.name ?? '') : '',
)
const tenGodGuide = computed(() => (tenGodName.value ? TEN_GOD_GUIDE[tenGodName.value] : undefined))

/** 节奏/大环境 chips：十二长生与月令状态的白话翻译，原词小字括号对照 */
const rhythmChips = computed(() => {
  const f = fortune.value
  if (!f) return [] as Array<{ label: string; plain: string; raw: string }>
  const chips: Array<{ label: string; plain: string; raw: string }> = []
  if (f.twelve_stage?.status === 'observed' && f.twelve_stage.name) {
    const raw = f.twelve_stage.name
    chips.push({ label: '节奏', plain: TWELVE_STAGE_PLAIN[raw] ?? '自有节奏', raw })
  }
  if (f.seasonal_state?.status === 'observed' && f.seasonal_state.state) {
    const raw = f.seasonal_state.state
    chips.push({ label: '大环境', plain: SEASONAL_STATE_PLAIN[raw] ?? '平缓', raw })
  }
  return chips
})

function todayString(date = new Date()) {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
}

async function fetchFortune() {
  loading.value = true
  error.value = ''
  errorKind.value = ''
  fortune.value = null
  const queryChartId = route.query.chart_id
  const cid =
    typeof queryChartId === 'string' && queryChartId ? queryChartId : recentChartStore.chartId
  if (!cid) {
    error.value = '请先创建命盘'
    errorKind.value = 'missing-chart'
    loading.value = false
    return
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

onMounted(fetchFortune)
</script>

<template>
  <div class="fortune-page">
    <FortuneStateView
      v-if="loading"
      kind="loading"
      title="正在整理今日干支关系"
      description="按当前规则计算今日干支与本命四柱的结构对应"
    />

    <FortuneStateView
      v-else-if="error"
      kind="error"
      :title="error"
      :description="errorKind === 'missing-chart' ? '今日运势需要基于一张已保存的命盘计算。' : ''"
      v-bind="
        errorKind === 'missing-chart'
          ? { actionLabel: '去排盘', actionTo: '/chart/new' }
          : { retryLabel: '重新加载' }
      "
      @retry="fetchFortune"
    />

    <FortuneStateView
      v-else-if="!fortune"
      kind="empty"
      title="今日暂无可显示的关系记录"
      description="可以稍后重新加载，或先查看本周、本月的整体节奏。"
      retry-label="重新加载"
      @retry="fetchFortune"
    />

    <!-- ── Main ── -->
    <main v-else class="fortune-main">
      <div class="top-nav" v-reveal>
        <PeriodNav current="day" :chart-id="chartId" />
      </div>

      <!-- ① 周期结论区 -->
      <header class="hero-panel" v-reveal="40">
        <p class="hero-eyebrow">今日运势 · Daily</p>
        <div class="hero-main">
          <div class="hero-pillar">
            <span class="pillar-prefix">日柱</span>
            <span class="pillar-value">{{ fortune.day_gan_zhi }}</span>
          </div>
          <div class="hero-text">
            <p class="hero-date tabular-nums">
              {{ fortune.solar_date
              }}<span v-if="fortune.week_day" class="hero-weekday">{{ fortune.week_day }}</span>
            </p>
            <p class="hero-meta">
              <span v-if="fortune.lunar_date">农历 {{ fortune.lunar_date }}</span>
              <span v-if="fortune.sheng_xiao">属{{ fortune.sheng_xiao }}</span>
            </p>
          </div>
        </div>
        <div class="hero-overview">
          <ScoreOrb
            :value="finalScore"
            :decimals="0"
            label="今日状态分"
            unit="/ 100"
            :caption="scoreTier.caption"
          />
          <div class="hero-overview-text">
            <p class="hero-headline">{{ tone.headline }}</p>
            <p class="hero-summary">{{ tone.why }}</p>
            <div v-if="tenGodGuide" class="hero-keyword">
              <p class="hero-keyword-eyebrow">今日关键词 · 十神「{{ tenGodName }}」</p>
              <p class="hero-keyword-word">{{ tenGodGuide.keyword }}</p>
              <p class="hero-keyword-advice">{{ tenGodGuide.advice }}</p>
            </div>
            <ul v-if="rhythmChips.length" class="hero-points">
              <li v-for="chip in rhythmChips" :key="chip.label" class="hero-point">
                <span class="hero-point-label">{{ chip.label }}</span>
                <span class="hero-point-value">
                  {{ chip.plain }}<span class="hero-point-raw">（{{ chip.raw }}）</span>
                </span>
              </li>
            </ul>
          </div>
        </div>
        <p class="hero-tip">
          以上内容根据干支结构与传统命理查表生成，只描述状态与节奏倾向，供参考，不代表确定会发生的事。想了解具体依据，可查看下方的结构关系记录。
        </p>
      </header>

      <!-- ② 明细区（解读层级 + 分页明细） -->
      <div class="main-grid" v-reveal="80">
        <DailyFortune
          hide-header
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

/* ── Main ── */
.fortune-main {
  position: relative;
  z-index: 2;
  max-width: 1180px;
  margin: 0 auto;
  padding: 2.5rem 1.5rem 5rem;
}

.top-nav {
  display: flex;
  justify-content: center;
  margin-bottom: 1.5rem;
}

/* ── ① 周期结论区 ── */
.hero-panel {
  position: relative;
  background: var(--surface-1);
  border: 1px solid var(--line-strong);
  border-radius: 16px;
  padding: 2rem 2.5rem 1.75rem;
  margin-bottom: 1.5rem;
  box-shadow: var(--shadow-md);
}
.hero-panel::before {
  content: '';
  position: absolute;
  top: -1px;
  left: 2.5rem;
  right: 2.5rem;
  height: 2px;
  background: rgba(var(--jade-accent-rgb), 0.55);
}

.hero-eyebrow {
  margin: 0 0 1.25rem;
  font-size: var(--fs-2xs);
  letter-spacing: var(--tracking-meta, 0.18em);
  color: rgba(var(--jade-accent-rgb), 1);
  font-weight: 600;
  text-transform: uppercase;
}

.hero-main {
  display: flex;
  align-items: flex-end;
  gap: 2.5rem;
  padding-bottom: 1.5rem;
  border-bottom: 1px solid var(--line-subtle);
}

.hero-pillar {
  display: flex;
  align-items: baseline;
  gap: 0.75rem;
}
.pillar-prefix {
  font-size: var(--fs-xs);
  color: var(--text-soft);
  letter-spacing: 2px;
}
.pillar-value {
  font-family: var(--font-serif);
  font-size: var(--fs-hero);
  font-weight: 900;
  color: rgba(var(--jade-accent-rgb), 1);
  letter-spacing: 4px;
  line-height: 1;
}

.hero-text {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
  padding-bottom: 0.2rem;
}
.hero-date {
  margin: 0;
  font-size: var(--fs-lg);
  font-weight: 700;
  color: var(--text);
  letter-spacing: 1px;
}
.hero-weekday {
  font-size: var(--fs-xs);
  font-weight: 400;
  color: var(--text-dim);
  margin-left: 0.5rem;
}
.hero-meta {
  display: flex;
  gap: 0.75rem;
  margin: 0;
  font-size: var(--fs-xs);
  color: var(--text-soft);
}

.hero-headline {
  margin: 0;
  font-family: var(--font-serif);
  font-size: var(--fs-xl);
  font-weight: 700;
  letter-spacing: 0.04em;
  color: var(--text);
  border-left: 2px solid rgba(var(--jade-accent-rgb), 0.55);
  padding-left: 0.75rem;
}

/* ── ①a 今日运势概览 ── */
.hero-overview {
  display: flex;
  align-items: center;
  gap: 2rem;
  margin-top: 1.5rem;
}

.hero-overview-text {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 0.7rem;
}

.hero-summary {
  margin: 0;
  font-size: var(--fs-sm);
  color: var(--text-muted);
  line-height: 1.8;
}

/* ── 今日关键词（十神的白话翻译） ── */
.hero-keyword {
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
  padding: 0.9rem 1.1rem;
  background: var(--surface-2);
  border: 1px solid var(--line-subtle);
  border-radius: 10px;
}

.hero-keyword-eyebrow {
  margin: 0;
  font-size: var(--fs-2xs);
  letter-spacing: var(--tracking-meta, 0.18em);
  color: var(--text-soft);
  font-weight: 600;
}

.hero-keyword-word {
  margin: 0;
  font-family: var(--font-serif);
  font-size: var(--fs-lg);
  font-weight: 800;
  letter-spacing: 0.06em;
  color: rgba(var(--jade-accent-rgb), 1);
}

.hero-keyword-advice {
  margin: 0;
  font-size: var(--fs-xs);
  color: var(--text-muted);
  line-height: 1.7;
}

.hero-points {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin: 0;
  padding: 0;
  list-style: none;
}

.hero-point {
  display: flex;
  align-items: baseline;
  gap: 0.45rem;
  padding: 0.45rem 0.7rem;
  background: var(--surface-2);
  border: 1px solid var(--line-subtle);
  border-radius: 8px;
  font-size: var(--fs-2xs);
}

.hero-point-label {
  color: var(--text-soft);
  letter-spacing: 0.5px;
  flex-shrink: 0;
}

.hero-point-value {
  color: var(--text-muted);
  font-weight: 600;
}

.hero-point-raw {
  font-weight: 400;
  font-size: 0.92em;
  color: var(--text-soft);
}

.hero-tip {
  margin: 0.6rem 0 0;
  font-size: var(--fs-xs);
  color: var(--text-soft);
  line-height: 1.7;
  padding-left: calc(0.75rem + 2px);
}

/* ── ② 明细区 ── */
.main-grid {
  display: block;
}

@media (max-width: 768px) {
  .fortune-main {
    padding: 1.5rem 1rem 4rem;
  }
  .hero-panel {
    padding: 1.5rem 1.25rem 1.25rem;
  }
  .hero-panel::before {
    left: 1.25rem;
    right: 1.25rem;
  }
  .hero-main {
    flex-direction: column;
    align-items: flex-start;
    gap: 1rem;
  }
  .hero-overview {
    flex-direction: column;
    align-items: stretch;
    gap: 1.25rem;
  }
  .hero-overview .metric {
    width: 100%;
  }
  .pillar-value {
    font-size: var(--fs-stat-lg);
  }
}
</style>
