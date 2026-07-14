<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import FortuneEvidencePanel from './FortuneEvidencePanel.vue'
import InterpretationLevelSwitch from './InterpretationLevelSwitch.vue'
import type {
  FortuneGuide,
  FortuneScoreBreakdown,
  InterpretationLevel,
  ScoreEvidence,
} from '../api/fortune'

interface ElementImage {
  element: string
  image_url: string
  description: string
}
interface ShengKeAnalysis {
  day_stem_relation?: string
  day_branch_relation?: string
  summary?: string
}
interface HiddenStemGod {
  stem: string
  type: string
  element: string
  ten_god: string
  favorable: boolean
}
interface StemRelation {
  type: string
  target: string
  detail: string
  is_favorable: boolean
  note?: string
}
interface BranchRelation {
  type: string
  target: string
  detail: string
  is_favorable: boolean
}
interface ShenShaActivation {
  name: string
  type: string
  description: string
  activation: string
}
interface DaYunInfluence {
  current_pillar: string
  start_age: number
  end_age: number
  ten_god: string
  favorable: boolean
  relation: string
  score: number
  description: string
}
interface LiuNianInfluence {
  year_pillar: string
  ten_god: string
  favorable: boolean
  relation: string
  tai_sui_relation: string
  score: number
  description: string
}
interface AdvanceRetreat {
  phase: string
  phase_desc: string
  element: string
  score: number
  description: string
}
interface YongShenImpact {
  tiao_hou_element: string
  tiao_hou_hit: boolean
  tong_guan_element: string
  tong_guan_hit: boolean
  fu_yi_elements: string[]
  fu_yi_hit: boolean
  score: number
  description: string
}
interface FortuneOverall {
  score?: number
  base_score?: number
  detail_score?: number
  stars?: string
  level?: string
  summary?: string
  key_tip?: string
}
interface FortuneCategory {
  name: string
  score?: number
  weight?: number
  stars?: string
  level?: string
  trend?: string
  keywords?: string[]
  analysis?: string
  advice?: string
}

interface Props {
  solarDate: string
  dayGanZhi: string
  weekDay?: string
  lunarDate?: string
  shengXiao?: string
  yiJi?: string
  chongSha?: string
  elementImages?: ElementImage[]
  luckyColor?: string
  luckyNumber?: number
  wealthDir?: string
  fortuneGuide?: FortuneGuide
  fortuneScore?: number
  fortuneOverall?: FortuneOverall
  fortuneCategories?: FortuneCategory[]
  scoreBreakdown?: FortuneScoreBreakdown
  evidenceCompleteness?: number
  supportingEvidence?: ScoreEvidence[]
  counterEvidence?: ScoreEvidence[]
  engineVersion?: string
  ruleVersion?: string
  auspiciousHours?: string[]
  todayElements?: Record<string, number>
  tiaoHou?: string
  // 黄历字段
  jiShen?: string
  xiongShen?: string
  taiShen?: string
  pengZu?: string
  gua?: string
  jieQi?: string
  // 分析字段
  shengKeAnalysis?: ShengKeAnalysis
  flowImpact?: string
  seasonElementAdvice?: string
  // 日课推算
  todayTenGod?: string
  tenGodFavorable?: boolean
  tenGodDesc?: string
  twelveStage?: string
  stageFavorable?: boolean
  stageDesc?: string
  stageFlexible?: string
  hiddenStems?: HiddenStemGod[]
  stemRelations?: StemRelation[]
  branchRelations?: BranchRelation[]
  activatedShenSha?: ShenShaActivation[]
  dayunInfluence?: DaYunInfluence
  liunianInfluence?: LiuNianInfluence
  advanceRetreat?: AdvanceRetreat
  yongshenImpact?: YongShenImpact
  overallVerdict?: string
  favorScore?: number
  patternName?: string
  patternType?: string
  patternFavorable?: string[]
  patternUnfavorable?: string[]
}
const props = withDefaults(defineProps<Props>(), {
  weekDay: '',
  lunarDate: '',
  shengXiao: '',
  yiJi: '',
  chongSha: '',
  elementImages: () => [],
  luckyColor: '',
  luckyNumber: 0,
  wealthDir: '',
  auspiciousHours: () => [],
  todayElements: () => ({}),
  fortuneGuide: undefined,
  fortuneScore: 0,
  fortuneOverall: undefined,
  fortuneCategories: () => [],
  scoreBreakdown: undefined,
  evidenceCompleteness: 0,
  supportingEvidence: () => [],
  counterEvidence: () => [],
  engineVersion: '',
  ruleVersion: '',
  tiaoHou: '',
  jiShen: '',
  xiongShen: '',
  taiShen: '',
  pengZu: '',
  gua: '',
  jieQi: '',
  shengKeAnalysis: undefined,
  flowImpact: '',
  seasonElementAdvice: '',
  // 日课推算
  todayTenGod: '',
  tenGodFavorable: false,
  tenGodDesc: '',
  twelveStage: '',
  stageFavorable: false,
  stageDesc: '',
  stageFlexible: '',
  hiddenStems: () => [],
  stemRelations: () => [],
  branchRelations: () => [],
  activatedShenSha: () => [],
  dayunInfluence: undefined,
  liunianInfluence: undefined,
  advanceRetreat: undefined,
  yongshenImpact: undefined,
  overallVerdict: '',
  favorScore: 0,
  patternName: '',
  patternType: '',
  patternFavorable: () => [],
  patternUnfavorable: () => [],
})
const showAiModal = ref(false)
const activeTab = ref('overview')
const savedLevel = localStorage.getItem('fortune-interpretation-level')
const interpretationLevel = ref<InterpretationLevel>(
  savedLevel === 'advanced' || savedLevel === 'professional' ? savedLevel : 'basic',
)
const isDark = ref(document.documentElement.classList.contains('dark'))
let themeObserver: MutationObserver | null = null
onMounted(() => {
  themeObserver = new MutationObserver(() => {
    isDark.value = document.documentElement.classList.contains('dark')
  })
  themeObserver.observe(document.documentElement, {
    attributes: true,
    attributeFilter: ['class', 'style', 'data-wuxing'],
  })
})
onUnmounted(() => {
  themeObserver?.disconnect()
})
const allDfTabs = [
  { key: 'overview', label: '今日概览', minimum: 'basic' },
  { key: 'almanac', label: '黄历', minimum: 'basic' },
  { key: 'analysis', label: '运势分析', minimum: 'advanced' },
  { key: 'elements', label: '五行吉时', minimum: 'advanced' },
  { key: 'rikuyo', label: '日课推算', minimum: 'professional' },
]
const levelRank: Record<InterpretationLevel, number> = { basic: 0, advanced: 1, professional: 2 }
const dfTabs = computed(() =>
  allDfTabs.filter(
    (tab) => levelRank[interpretationLevel.value] >= levelRank[tab.minimum as InterpretationLevel],
  ),
)
watch(interpretationLevel, (level) => {
  localStorage.setItem('fortune-interpretation-level', level)
  if (!dfTabs.value.some((tab) => tab.key === activeTab.value)) activeTab.value = 'overview'
})
const elementEntries = computed(() => {
  if (isDark.value)
    return [
      ['金', '#cbd5e1'],
      ['木', '#34d399'],
      ['水', '#22d3ee'],
      ['火', '#fb7185'],
      ['土', '#fde68a'],
    ] as [string, string][]
  return [
    ['金', '#94a3b8'],
    ['木', '#16a34a'],
    ['水', '#0891b2'],
    ['火', '#dc2626'],
    ['土', '#a16207'],
  ] as [string, string][]
})
function elPct(el: string) {
  const n = props.todayElements || {},
    t = Object.values(n).reduce((s, v) => s + v, 0)
  return t ? Math.round(((n[el] || 0) / t) * 100) : 0
}

function guideEvidenceCompletenessLabel(value?: number) {
  if (!value) return ''
  if (value >= 80) return '依据较完整'
  if (value >= 65) return '依据基本完整'
  return '依据有限'
}

function guidePrecisionLabel(level?: string) {
  if (level === 'pattern-aware') return '格局感知'
  if (level === 'legacy') return '旧版规则'
  return '基础模式'
}

function scoreWord(score: number) {
  if (score >= 85) return '顺势明显'
  if (score >= 70) return '良好'
  if (score >= 55) return '平稳'
  if (score >= 40) return '欠佳'
  return '低迷'
}

function starCount(stars?: string) {
  return (stars?.match(/★/g) || []).length
}

function clampScore(score?: number) {
  return Math.max(20, Math.min(100, score ?? 0))
}

function categoryScore(category: FortuneCategory) {
  const fallback = starCount(category.stars) * 20 || 60
  return clampScore(category.score ?? fallback)
}

function trendLabel(trend?: string) {
  if (trend === 'up') return '走高'
  if (trend === 'down') return '承压'
  return '平稳'
}
</script>

<template>
  <div class="daily-fortune">
    <!-- Date + Pillar -->
    <div class="df-header glass-card">
      <div class="df-date-col">
        <p class="df-solar">
          {{ solarDate }}<span v-if="weekDay" class="df-weekday">{{ weekDay }}</span>
        </p>
        <p v-if="lunarDate" class="df-lunar">{{ lunarDate }}</p>
      </div>
      <div class="df-pillar-col">
        <div class="df-pillar-glow"></div>
        <span class="df-pillar-val">{{ dayGanZhi }}</span>
        <span v-if="shengXiao" class="df-sx">属{{ shengXiao }}</span>
      </div>
    </div>

    <InterpretationLevelSwitch v-model="interpretationLevel" class="df-level-switch" />

    <!-- Tab navigation -->
    <div class="df-tabs">
      <button
        v-for="tab in dfTabs"
        :key="tab.key"
        class="df-tab-btn"
        :class="{ active: activeTab === tab.key }"
        @click="activeTab = tab.key"
      >
        {{ tab.label }}
      </button>
    </div>

    <!-- ═══ Tab: 今日概览 ═══ -->
    <div v-show="activeTab === 'overview'" class="df-tab-content">
      <div v-if="fortuneGuide" class="df-guide glass-card">
        <div class="df-sec-header">
          <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
            <path
              d="M7 1l1.8 3.6L13 5.3l-3 2.9.7 4.1L7 10.5l-3.7 1.8.7-4.1-3-2.9 4.2-.7L7 1z"
              stroke="currentColor"
              stroke-width="1"
              opacity="0.5"
            />
          </svg>
          <span class="df-sec-title">今日概览</span>
          <span class="guide-precision">{{
            guidePrecisionLabel(fortuneGuide.precision_level)
          }}</span>
          <span v-if="fortuneGuide.evidence_completeness" class="guide-confidence"
            >{{ guideEvidenceCompletenessLabel(fortuneGuide.evidence_completeness) }} ·
            规则证据完整度 {{ fortuneGuide.evidence_completeness }}%</span
          >
        </div>

        <div class="guide-hero">
          <div class="guide-brief">
            <div class="guide-axis">
              <span>主 {{ fortuneGuide.primary_element || '—' }}</span>
              <span>辅 {{ fortuneGuide.secondary_element || '—' }}</span>
              <span>忌 {{ fortuneGuide.avoid_element || '—' }}</span>
            </div>
            <p v-if="fortuneGuide.strategy" class="guide-strategy">{{ fortuneGuide.strategy }}</p>
          </div>

          <div class="guide-essentials">
            <div class="guide-item" v-if="fortuneGuide.lucky_colors?.[0]?.value">
              <span class="guide-label">幸运色</span>
              <span class="guide-value">
                <span
                  class="color-swatch"
                  :style="{ background: fortuneGuide.lucky_colors?.[0]?.value }"
                ></span>
                {{ fortuneGuide.lucky_colors?.[0]?.value }}
              </span>
              <small class="guide-reason">{{ fortuneGuide.lucky_colors?.[0]?.reason }}</small>
            </div>
            <div class="guide-item" v-if="fortuneGuide.lucky_numbers?.[0]?.value">
              <span class="guide-label">幸运数字</span>
              <span class="guide-value guide-value-xl">{{
                fortuneGuide.lucky_numbers?.[0]?.value
              }}</span>
              <small class="guide-reason">{{ fortuneGuide.lucky_numbers?.[0]?.reason }}</small>
            </div>
            <div class="guide-item" v-if="fortuneGuide.wealth_direction?.value">
              <span class="guide-label">财位</span>
              <span class="guide-value">{{ fortuneGuide.wealth_direction.value }}</span>
              <small class="guide-reason">{{ fortuneGuide.wealth_direction.reason }}</small>
            </div>
            <div class="guide-item" v-if="fortuneGuide.face_direction?.value">
              <span class="guide-label">朝向</span>
              <span class="guide-value">{{ fortuneGuide.face_direction.value }}</span>
              <small class="guide-reason">{{ fortuneGuide.face_direction.reason }}</small>
            </div>
          </div>
        </div>

        <div class="guide-action-grid">
          <div v-if="fortuneGuide.recommended_actions?.length" class="guide-list">
            <div class="guide-list-head">
              <div>
                <span class="guide-list-title">宜用</span>
                <strong>优先执行</strong>
              </div>
              <span class="guide-list-count">{{ fortuneGuide.recommended_actions.length }} 项</span>
            </div>
            <article
              v-for="item in fortuneGuide.recommended_actions.slice(0, 3)"
              :key="`action-${item.label}-${item.value}`"
              class="guide-mini"
            >
              <div class="guide-mini-top">
                <span class="guide-mini-label">{{ item.category || item.label }}</span>
                <strong>{{ item.value }}</strong>
                <span v-if="item.intensity" class="guide-intensity">{{ item.intensity }}</span>
              </div>
              <small class="guide-mini-reason">{{ item.reason }}</small>
              <details
                v-if="item.timing || item.source || item.impact || item.method"
                class="guide-more"
              >
                <summary>依据</summary>
                <div class="guide-detail-row">
                  <span v-if="item.timing">{{ item.timing }}</span>
                  <span v-if="item.source">{{ item.source }}</span>
                  <span v-if="item.impact">{{ item.impact }}</span>
                </div>
                <p v-if="item.method" class="guide-method">{{ item.method }}</p>
              </details>
            </article>
            <details v-if="fortuneGuide.recommended_actions.length > 3" class="guide-extra">
              <summary>展开其余 {{ fortuneGuide.recommended_actions.length - 3 }} 项</summary>
              <div class="guide-extra-list">
                <article
                  v-for="item in fortuneGuide.recommended_actions.slice(3)"
                  :key="`action-extra-${item.label}-${item.value}`"
                  class="guide-mini guide-mini-secondary"
                >
                  <div class="guide-mini-top">
                    <span class="guide-mini-label">{{ item.category || item.label }}</span>
                    <strong>{{ item.value }}</strong>
                    <span v-if="item.intensity" class="guide-intensity">{{ item.intensity }}</span>
                  </div>
                  <small class="guide-mini-reason">{{ item.reason }}</small>
                </article>
              </div>
            </details>
          </div>
          <div v-if="fortuneGuide.cautions?.length" class="guide-list">
            <div class="guide-list-head">
              <div>
                <span class="guide-list-title guide-list-warn">避忌</span>
                <strong>先避风险</strong>
              </div>
              <span class="guide-list-count warn">{{ fortuneGuide.cautions.length }} 项</span>
            </div>
            <article
              v-for="item in fortuneGuide.cautions.slice(0, 3)"
              :key="`caution-${item.label}-${item.value}`"
              class="guide-mini warn"
            >
              <div class="guide-mini-top">
                <span class="guide-mini-label warn">{{ item.category || item.label }}</span>
                <strong>{{ item.value }}</strong>
                <span v-if="item.intensity" class="guide-intensity warn">{{ item.intensity }}</span>
              </div>
              <small class="guide-mini-reason">{{ item.reason }}</small>
              <details
                v-if="item.timing || item.source || item.impact || item.method"
                class="guide-more warn"
              >
                <summary>依据</summary>
                <div class="guide-detail-row warn">
                  <span v-if="item.timing">{{ item.timing }}</span>
                  <span v-if="item.source">{{ item.source }}</span>
                  <span v-if="item.impact">{{ item.impact }}</span>
                </div>
                <p v-if="item.method" class="guide-method">{{ item.method }}</p>
              </details>
            </article>
            <details v-if="fortuneGuide.cautions.length > 3" class="guide-extra warn">
              <summary>展开其余 {{ fortuneGuide.cautions.length - 3 }} 项</summary>
              <div class="guide-extra-list">
                <article
                  v-for="item in fortuneGuide.cautions.slice(3)"
                  :key="`caution-extra-${item.label}-${item.value}`"
                  class="guide-mini warn guide-mini-secondary"
                >
                  <div class="guide-mini-top">
                    <span class="guide-mini-label warn">{{ item.category || item.label }}</span>
                    <strong>{{ item.value }}</strong>
                    <span v-if="item.intensity" class="guide-intensity warn">{{
                      item.intensity
                    }}</span>
                  </div>
                  <small class="guide-mini-reason">{{ item.reason }}</small>
                </article>
              </div>
            </details>
          </div>
        </div>

        <div class="guide-footer">
          <div v-if="fortuneGuide.best_hours?.length" class="guide-hours">
            <span
              v-for="item in fortuneGuide.best_hours.slice(0, 4)"
              :key="`hour-${item.label}-${item.value}`"
              >{{ item.value }}</span
            >
          </div>
          <details v-if="fortuneGuide.analysis" class="guide-note">
            <summary>取用逻辑</summary>
            <p class="guide-analysis">{{ fortuneGuide.analysis }}</p>
          </details>
        </div>
      </div>
      <div v-else class="df-guide-empty glass-card">
        <div class="df-sec-header">
          <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
            <circle cx="7" cy="7" r="5" stroke="currentColor" stroke-width="1" opacity="0.35" />
            <path
              d="M7 4v3.4l2 1.2"
              stroke="currentColor"
              stroke-width="1"
              stroke-linecap="round"
              opacity="0.45"
            />
          </svg>
          <span class="df-sec-title">今日概览</span>
        </div>
        <p>开运指南暂未生成，请稍后刷新今日运势。</p>
      </div>

      <FortuneEvidencePanel
        :level="interpretationLevel"
        :completeness="evidenceCompleteness"
        :supporting="supportingEvidence"
        :counter="counterEvidence"
        :breakdown="scoreBreakdown"
        :engine-version="engineVersion"
        :rule-version="ruleVersion"
      />
    </div>
    <!-- /overview tab -->

    <!-- ═══ Tab: 黄历 ═══ -->
    <div v-show="activeTab === 'almanac'" class="df-tab-content">
      <!-- 黄历信息 -->
      <div
        v-if="jiShen || xiongShen || taiShen || pengZu || gua || jieQi"
        class="df-almanac glass-card"
      >
        <div class="df-sec-header">
          <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
            <rect
              x="2"
              y="3"
              width="10"
              height="9"
              rx="1.5"
              stroke="currentColor"
              stroke-width="1"
              opacity="0.4"
            />
            <line
              x1="2"
              y1="6"
              x2="12"
              y2="6"
              stroke="currentColor"
              stroke-width="0.8"
              opacity="0.3"
            />
            <line
              x1="5"
              y1="1"
              x2="5"
              y2="4"
              stroke="currentColor"
              stroke-width="1"
              stroke-linecap="round"
              opacity="0.4"
            />
            <line
              x1="9"
              y1="1"
              x2="9"
              y2="4"
              stroke="currentColor"
              stroke-width="1"
              stroke-linecap="round"
              opacity="0.4"
            />
          </svg>
          <span class="df-sec-title">黄历</span>
        </div>
        <div class="almanac-grid">
          <div v-if="jieQi" class="almanac-item">
            <span class="almanac-label">节气</span>
            <span class="almanac-value">{{ jieQi }}</span>
          </div>
          <div v-if="gua" class="almanac-item">
            <span class="almanac-label">卦象</span>
            <span class="almanac-value">{{ gua }}</span>
          </div>
          <div v-if="taiShen" class="almanac-item">
            <span class="almanac-label">胎神</span>
            <span class="almanac-value">{{ taiShen }}</span>
          </div>
          <div v-if="jiShen" class="almanac-item">
            <span class="almanac-label">吉神</span>
            <span class="almanac-value almanac-ji">{{ jiShen }}</span>
          </div>
          <div v-if="xiongShen" class="almanac-item">
            <span class="almanac-label">凶神</span>
            <span class="almanac-value almanac-xiong">{{ xiongShen }}</span>
          </div>
          <div v-if="pengZu" class="almanac-item almanac-full">
            <span class="almanac-label">彭祖百忌</span>
            <span class="almanac-value">{{ pengZu }}</span>
          </div>
        </div>
      </div>
    </div>
    <!-- /almanac tab -->

    <!-- ═══ Tab: 运势分析 ═══ -->
    <div v-show="activeTab === 'analysis'" class="df-tab-content">
      <!-- 生克分析 -->
      <div
        v-if="shengKeAnalysis?.summary || flowImpact || seasonElementAdvice"
        class="df-analysis glass-card"
      >
        <div class="df-sec-header">
          <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
            <circle cx="7" cy="7" r="5.5" stroke="currentColor" stroke-width="0.8" opacity="0.3" />
            <path
              d="M5 7h4M7 5v4"
              stroke="currentColor"
              stroke-width="1"
              stroke-linecap="round"
              opacity="0.5"
            />
          </svg>
          <span class="df-sec-title">运势分析</span>
        </div>
        <div class="analysis-items">
          <div v-if="shengKeAnalysis?.day_stem_relation" class="analysis-item">
            <span class="analysis-label">日干关系</span>
            <span class="analysis-value">{{ shengKeAnalysis.day_stem_relation }}</span>
          </div>
          <div v-if="shengKeAnalysis?.day_branch_relation" class="analysis-item">
            <span class="analysis-label">日支关系</span>
            <span class="analysis-value">{{ shengKeAnalysis.day_branch_relation }}</span>
          </div>
          <div v-if="shengKeAnalysis?.summary" class="analysis-item">
            <span class="analysis-label">综合</span>
            <span class="analysis-value">{{ shengKeAnalysis.summary }}</span>
          </div>
          <div v-if="flowImpact" class="analysis-item">
            <span class="analysis-label">流通影响</span>
            <span class="analysis-value">{{ flowImpact }}</span>
          </div>
          <div v-if="seasonElementAdvice" class="analysis-item">
            <span class="analysis-label">季节建议</span>
            <span class="analysis-value analysis-gold">{{ seasonElementAdvice }}</span>
          </div>
        </div>
      </div>
    </div>
    <!-- /analysis tab -->

    <!-- ═══ Tab: 日课推算 ═══ -->
    <div v-show="activeTab === 'rikuyo'" class="df-tab-content">
      <!-- ═══ 日课推算 ═══ -->

      <!-- 综合断语 + 评分 -->
      <div v-if="overallVerdict" class="df-rikuyo-verdict glass-card">
        <div class="df-sec-header">
          <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
            <path
              d="M7 1l1.8 3.6L13 5.3l-3 2.9.7 4.1L7 10.5l-3.7 1.8.7-4.1-3-2.9 4.2-.7L7 1z"
              stroke="currentColor"
              stroke-width="1"
              opacity="0.5"
            />
          </svg>
          <span class="df-sec-title">日课推算</span>
          <span
            v-if="favorScore"
            class="rikuyo-score-badge"
            :style="{ '--score-t': Math.max(0, Math.min(1, favorScore / 100)) }"
            >{{ favorScore }}分</span
          >
        </div>
        <p class="rikuyo-verdict-text">{{ overallVerdict }}</p>
      </div>

      <!-- 格局信息（特殊格局时显示） -->
      <div v-if="patternName && patternType === '特殊格局'" class="df-pattern-info glass-card">
        <div class="df-sec-header">
          <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
            <circle cx="7" cy="7" r="5.5" stroke="currentColor" stroke-width="1" opacity="0.5" />
            <path d="M7 4v3l2 1" stroke="currentColor" stroke-width="1" stroke-linecap="round" />
          </svg>
          <span class="df-sec-title">格局喜忌</span>
          <span class="pattern-badge">{{ patternName }}</span>
        </div>
        <div class="pattern-elements">
          <span v-if="patternFavorable?.length" class="pattern-tag pattern-like"
            >喜{{ patternFavorable.join('') }}</span
          >
          <span v-if="patternUnfavorable?.length" class="pattern-tag pattern-dislike"
            >忌{{ patternUnfavorable.join('') }}</span
          >
        </div>
      </div>

      <!-- 今日十神 + 十二长生 -->
      <div v-if="todayTenGod || twelveStage" class="df-rikuyo-core glass-card">
        <div class="rikuyo-core-grid">
          <!-- 十神 -->
          <div v-if="todayTenGod" class="rikuyo-core-item">
            <span class="rikuyo-core-label">今日十神</span>
            <span
              class="rikuyo-core-value"
              :class="{ 'val-fav': tenGodFavorable, 'val-dis': !tenGodFavorable }"
              >{{ todayTenGod }}</span
            >
            <span v-if="tenGodDesc" class="rikuyo-core-desc">{{ tenGodDesc }}</span>
          </div>
          <!-- 十二长生 -->
          <div v-if="twelveStage" class="rikuyo-core-item">
            <span class="rikuyo-core-label">十二长生</span>
            <span
              class="rikuyo-core-value"
              :class="{ 'val-fav': stageFavorable, 'val-dis': !stageFavorable }"
              >{{ twelveStage }}</span
            >
            <span v-if="stageDesc" class="rikuyo-core-desc">{{ stageDesc }}</span>
            <span v-if="stageFlexible" class="rikuyo-flexible">{{ stageFlexible }}</span>
          </div>
        </div>
      </div>

      <!-- 进退气 -->
      <div v-if="advanceRetreat" class="df-rikuyo-advance glass-card">
        <div class="df-sec-header">
          <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
            <path
              d="M2 12L7 2L12 12"
              stroke="currentColor"
              stroke-width="1"
              stroke-linecap="round"
              opacity="0.4"
            />
            <line
              x1="4"
              y1="8"
              x2="10"
              y2="8"
              stroke="currentColor"
              stroke-width="0.8"
              opacity="0.3"
            />
          </svg>
          <span class="df-sec-title">进退气</span>
          <span
            class="rikuyo-phase-tag"
            :class="{
              'phase-adv': advanceRetreat.phase === '进气',
              'phase-peak': advanceRetreat.phase === '当令',
              'phase-ret': advanceRetreat.phase === '退气',
              'phase-dead': advanceRetreat.phase === '无气' || advanceRetreat.phase === '死',
            }"
            >{{ advanceRetreat.phase }}</span
          >
        </div>
        <p class="rikuyo-advance-text">{{ advanceRetreat.description }}</p>
      </div>

      <!-- 藏干分析 -->
      <div v-if="hiddenStems?.length" class="df-rikuyo-hidden glass-card">
        <div class="df-sec-header">
          <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
            <rect
              x="2"
              y="2"
              width="10"
              height="10"
              rx="2"
              stroke="currentColor"
              stroke-width="0.8"
              opacity="0.3"
            />
            <rect
              x="4.5"
              y="4.5"
              width="5"
              height="5"
              rx="1"
              stroke="currentColor"
              stroke-width="0.6"
              opacity="0.2"
            />
          </svg>
          <span class="df-sec-title">地支藏干</span>
        </div>
        <div class="hidden-stems-grid">
          <div
            v-for="hs in hiddenStems"
            :key="hs.stem + hs.type"
            class="hidden-stem-card"
            :class="{ 'hs-fav': hs.favorable, 'hs-dis': !hs.favorable }"
          >
            <span class="hs-stem">{{ hs.stem }}</span>
            <span class="hs-type">{{ hs.type }}</span>
            <span class="hs-god">{{ hs.ten_god }}</span>
            <span class="hs-elem">{{ hs.element }}</span>
          </div>
        </div>
      </div>

      <!-- 干支关系 -->
      <div
        v-if="stemRelations?.length || branchRelations?.length"
        class="df-rikuyo-relations glass-card"
      >
        <div class="df-sec-header">
          <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
            <circle cx="4" cy="7" r="2.5" stroke="currentColor" stroke-width="0.8" opacity="0.3" />
            <circle cx="10" cy="7" r="2.5" stroke="currentColor" stroke-width="0.8" opacity="0.3" />
            <line
              x1="6.5"
              y1="7"
              x2="7.5"
              y2="7"
              stroke="currentColor"
              stroke-width="1"
              opacity="0.4"
            />
          </svg>
          <span class="df-sec-title">干支关系</span>
        </div>
        <div class="relations-list">
          <div
            v-for="(sr, i) in stemRelations"
            :key="'sr' + i"
            class="relation-item"
            :class="{ 'rel-fav': sr.is_favorable, 'rel-dis': !sr.is_favorable }"
          >
            <span class="rel-type-tag">{{ sr.type }}</span>
            <span class="rel-detail">{{ sr.detail }}</span>
            <span v-if="sr.note" class="rel-note">{{ sr.note }}</span>
          </div>
          <div
            v-for="(br, i) in branchRelations"
            :key="'br' + i"
            class="relation-item"
            :class="{ 'rel-fav': br.is_favorable, 'rel-dis': !br.is_favorable }"
          >
            <span class="rel-type-tag">{{ br.type }}</span>
            <span class="rel-detail">{{ br.detail }}</span>
          </div>
        </div>
      </div>

      <!-- 神煞引动 -->
      <div v-if="activatedShenSha?.length" class="df-rikuyo-shensha glass-card">
        <div class="df-sec-header">
          <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
            <path
              d="M7 1l1.5 3 3.5.5-2.5 2.5.6 3.5L7 9l-3.1 1.5.6-3.5L2 4.5l3.5-.5L7 1z"
              stroke="currentColor"
              stroke-width="0.8"
              opacity="0.4"
            />
          </svg>
          <span class="df-sec-title">神煞引动</span>
        </div>
        <div class="shensha-list">
          <div
            v-for="ss in activatedShenSha"
            :key="ss.name"
            class="shensha-item"
            :class="{ 'ss-ji': ss.type === '吉神', 'ss-xiong': ss.type !== '吉神' }"
          >
            <span class="ss-name">{{ ss.name }}</span>
            <span class="ss-type-tag">{{ ss.type }}</span>
            <p class="ss-desc">{{ ss.description }}</p>
            <p class="ss-activation">{{ ss.activation }}</p>
          </div>
        </div>
      </div>

      <!-- 大运流年叠加 -->
      <div v-if="dayunInfluence || liunianInfluence" class="df-rikuyo-yun glass-card">
        <div class="df-sec-header">
          <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
            <path d="M1 7h12M7 1v12" stroke="currentColor" stroke-width="0.8" opacity="0.3" />
            <circle cx="7" cy="7" r="3" stroke="currentColor" stroke-width="0.8" opacity="0.25" />
          </svg>
          <span class="df-sec-title">大运流年</span>
        </div>
        <div class="yun-grid">
          <div v-if="dayunInfluence" class="yun-item">
            <span class="yun-label">当前大运</span>
            <span class="yun-pillar">{{ dayunInfluence.current_pillar }}</span>
            <span
              class="yun-god"
              :class="{ 'val-fav': dayunInfluence.favorable, 'val-dis': !dayunInfluence.favorable }"
              >{{ dayunInfluence.ten_god }}</span
            >
            <span class="yun-age"
              >{{ dayunInfluence.start_age }}-{{ dayunInfluence.end_age }}岁</span
            >
            <p class="yun-desc">{{ dayunInfluence.description }}</p>
          </div>
          <div v-if="liunianInfluence" class="yun-item">
            <span class="yun-label">流年</span>
            <span class="yun-pillar">{{ liunianInfluence.year_pillar }}</span>
            <span
              class="yun-god"
              :class="{
                'val-fav': liunianInfluence.favorable,
                'val-dis': !liunianInfluence.favorable,
              }"
              >{{ liunianInfluence.ten_god }}</span
            >
            <p v-if="liunianInfluence.tai_sui_relation" class="yun-taisui">
              {{ liunianInfluence.tai_sui_relation }}
            </p>
            <p class="yun-desc">{{ liunianInfluence.description }}</p>
          </div>
        </div>
      </div>

      <!-- 用神影响 -->
      <div v-if="yongshenImpact" class="df-rikuyo-yongshen glass-card">
        <div class="df-sec-header">
          <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
            <circle cx="7" cy="7" r="6" stroke="currentColor" stroke-width="0.6" opacity="0.2" />
            <circle cx="7" cy="7" r="3" stroke="currentColor" stroke-width="0.8" opacity="0.35" />
            <circle cx="7" cy="7" r="1" fill="currentColor" opacity="0.5" />
          </svg>
          <span class="df-sec-title">用神影响</span>
        </div>
        <div class="yongshen-items">
          <div
            v-if="yongshenImpact.tiao_hou_element"
            class="yongshen-item"
            :class="{ 'ys-hit': yongshenImpact.tiao_hou_hit }"
          >
            <span class="ys-label">调候用神</span>
            <span class="ys-elem">{{ yongshenImpact.tiao_hou_element }}</span>
            <span class="ys-status">{{ yongshenImpact.tiao_hou_hit ? '得力' : '未触' }}</span>
          </div>
          <div
            v-if="yongshenImpact.tong_guan_element"
            class="yongshen-item"
            :class="{ 'ys-hit': yongshenImpact.tong_guan_hit }"
          >
            <span class="ys-label">通关用神</span>
            <span class="ys-elem">{{ yongshenImpact.tong_guan_element }}</span>
            <span class="ys-status">{{ yongshenImpact.tong_guan_hit ? '得力' : '未触' }}</span>
          </div>
          <div
            v-if="yongshenImpact.fu_yi_elements?.length"
            class="yongshen-item"
            :class="{ 'ys-hit': yongshenImpact.fu_yi_hit }"
          >
            <span class="ys-label">扶抑喜用</span>
            <span class="ys-elem">{{ yongshenImpact.fu_yi_elements.join(' ') }}</span>
            <span class="ys-status">{{ yongshenImpact.fu_yi_hit ? '得力' : '未触' }}</span>
          </div>
        </div>
        <p v-if="yongshenImpact.description" class="yongshen-desc">
          {{ yongshenImpact.description }}
        </p>
      </div>
    </div>
    <!-- /rikuyo tab -->

    <!-- ═══ Tab: 五行吉时 ═══ -->
    <div v-show="activeTab === 'elements'" class="df-tab-content">
      <!-- Hours + Elements -->
      <div class="df-bottom-row">
        <div v-if="auspiciousHours.length" class="df-hours glass-card">
          <div class="df-sec-header">
            <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
              <circle cx="7" cy="7" r="6" stroke="currentColor" stroke-width="1" opacity="0.4" />
              <line
                x1="7"
                y1="3"
                x2="7"
                y2="7"
                stroke="currentColor"
                stroke-width="1.5"
                stroke-linecap="round"
              />
              <line
                x1="7"
                y1="7"
                x2="10"
                y2="9"
                stroke="currentColor"
                stroke-width="1.5"
                stroke-linecap="round"
              />
            </svg>
            <span class="df-sec-title">吉时</span>
          </div>
          <div class="df-hours-list">
            <span v-for="h in auspiciousHours" :key="h" class="df-hour-chip">
              <span class="df-hour-dot"></span>{{ h }}
            </span>
          </div>
        </div>
        <div class="df-elems glass-card">
          <div class="df-sec-header">
            <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
              <circle cx="7" cy="7" r="6" stroke="currentColor" stroke-width="1" opacity="0.25" />
              <circle cx="7" cy="7" r="2.5" fill="currentColor" opacity="0.35" />
            </svg>
            <span class="df-sec-title">今日五行</span>
          </div>
          <div class="df-el-bars">
            <div v-for="[el, clr] in elementEntries" :key="el" class="df-el-row">
              <span class="df-el-name">{{ el }}</span>
              <div class="df-el-track">
                <div
                  class="df-el-fill"
                  :style="{
                    width: elPct(el) + '%',
                    background: clr,
                    boxShadow: `0 0 8px ${clr}66`,
                  }"
                ></div>
              </div>
              <span class="df-el-num">{{ todayElements[el] ?? 0 }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- AI button -->
      <button class="df-ai-btn" disabled title="该功能仍在规划中">
        <span class="df-ai-btn-icon">◈</span>
        深度解析 · 规划中
      </button>
    </div>
    <!-- /elements tab -->

    <section
      v-if="fortuneCategories.length && interpretationLevel !== 'basic'"
      class="dimension-panel"
      aria-label="分项运势"
    >
      <div class="dimension-head">
        <div>
          <span class="dimension-eyebrow">九维刻度</span>
          <h2 class="dimension-title">先看强弱，再展开细节</h2>
        </div>
        <div class="score-disclaimer">传统规则倾向分，表示规则匹配强弱，不是事件发生概率。</div>
        <div class="score-breakdown" v-if="fortuneOverall">
          <span>细项 {{ fortuneOverall.detail_score ?? fortuneScore }}</span>
          <span>基础 {{ fortuneOverall.base_score ?? fortuneScore }}</span>
          <strong>{{ fortuneOverall.level || scoreWord(fortuneScore || 0) }}</strong>
        </div>
      </div>

      <div class="dimension-grid">
        <details
          v-for="(c, ci) in fortuneCategories"
          :key="c.name"
          class="dimension-card"
          :class="`trend-${c.trend || 'flat'}`"
          :style="{ animationDelay: ci * 42 + 'ms' }"
        >
          <summary class="dimension-summary">
            <span class="dimension-rank">{{ String(ci + 1).padStart(2, '0') }}</span>
            <span class="dimension-main">
              <span class="dimension-name">{{ c.name }}</span>
              <span class="dimension-keyword">{{ c.keywords?.[0] || trendLabel(c.trend) }}</span>
            </span>
            <span class="dimension-meter" aria-hidden="true">
              <span :style="{ width: `${categoryScore(c)}%` }"></span>
            </span>
            <span class="dimension-score">{{ categoryScore(c) }}</span>
            <span class="dimension-state">{{ c.level || scoreWord(categoryScore(c)) }}</span>
          </summary>

          <div class="dimension-detail">
            <div class="dimension-meta-row">
              <span>{{ trendLabel(c.trend) }}</span>
              <span>权重 {{ c.weight ?? 0 }}%</span>
              <span>{{ c.stars }}</span>
            </div>

            <p class="dimension-analysis" v-if="c.analysis">{{ c.analysis }}</p>
            <p class="dimension-advice" v-if="c.advice">{{ c.advice }}</p>

            <div v-if="c.keywords?.length" class="dimension-tags">
              <span v-for="kw in c.keywords.slice(0, 4)" :key="`${c.name}-${kw}`">{{ kw }}</span>
            </div>
          </div>
        </details>
      </div>
    </section>

    <!-- Modal -->
    <Teleport to="body">
      <Transition name="df-modal">
        <div v-if="showAiModal" class="df-modal-overlay" @click.self="showAiModal = false">
          <div class="df-modal-box glass-panel">
            <div class="df-modal-hdr">
              <div class="df-modal-title-group">
                <span class="df-modal-orb">☯</span>
                <h2>AI 深度解析</h2>
              </div>
              <button class="df-modal-close" @click="showAiModal = false">✕</button>
            </div>
            <div class="df-modal-body">
              <div class="df-ai-coming">
                <svg width="90" height="90" viewBox="0 0 90 90" fill="none" class="df-ai-svg">
                  <circle
                    cx="45"
                    cy="45"
                    r="42"
                    stroke="currentColor"
                    stroke-width="0.6"
                    stroke-dasharray="2 4"
                    opacity="0.2"
                  />
                  <circle
                    cx="45"
                    cy="45"
                    r="28"
                    stroke="currentColor"
                    stroke-width="0.6"
                    stroke-dasharray="1 5"
                    opacity="0.15"
                  />
                  <circle cx="45" cy="45" r="8" fill="currentColor" opacity="0.2" />
                  <circle
                    cx="45"
                    cy="45"
                    r="13"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="0.5"
                    opacity="0.3"
                  />
                  <circle
                    cx="22"
                    cy="24"
                    r="2.5"
                    fill="currentColor"
                    opacity="0.45"
                    class="df-star-pulse"
                    style="animation-delay: 0s"
                  />
                  <circle
                    cx="68"
                    cy="22"
                    r="2"
                    fill="currentColor"
                    opacity="0.35"
                    class="df-star-pulse"
                    style="animation-delay: 0.6s"
                  />
                  <circle
                    cx="70"
                    cy="66"
                    r="2.5"
                    fill="currentColor"
                    opacity="0.4"
                    class="df-star-pulse"
                    style="animation-delay: 1.2s"
                  />
                </svg>
                <p class="df-ai-title">AI分析功能即将上线</p>
                <p class="df-ai-sub">智能运势深度解读</p>
              </div>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<style scoped>
.daily-fortune {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

/* Tab navigation */
.df-tabs {
  display: flex;
  gap: 0.25rem;
  padding: 0;
  overflow-x: auto;
  scrollbar-width: none;
  margin-bottom: 0.75rem;
}

.df-level-switch {
  margin: 0 0 0.7rem;
}

.df-tabs::-webkit-scrollbar {
  display: none;
}

.df-tab-btn {
  padding: 0.6rem 1rem;
  flex-shrink: 0;
  background: var(--glass-bg);
  border: 1px solid var(--line-subtle);
  border-radius: 8px;
  color: var(--text-muted);
  font-size: var(--fs-xs);
  font-weight: 600;
  letter-spacing: 1px;
  cursor: pointer;
  white-space: nowrap;
  transition: all 0.3s;
}

.df-tab-btn:hover {
  color: var(--text);
  border-color: var(--line-strong);
}

.df-tab-btn.active {
  color: var(--accent);
  border-color: var(--line-focus);
  background: var(--accent-dim);
}

.df-tab-content {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  transition: opacity 0.3s ease;
}

/* Header */
.df-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.25rem 1.5rem;
  position: relative;
  overflow: hidden;
}

.df-header::after {
  content: '';
  position: absolute;
  top: -20px;
  right: -20px;
  width: 100px;
  height: 100px;
  background: radial-gradient(
    circle,
    color-mix(in oklab, var(--crimson) 7%, transparent),
    transparent 70%
  );
  pointer-events: none;
}

.df-date-col {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
}

.df-solar {
  font-size: var(--fs-lg);
  font-weight: 700;
  color: var(--text);
  margin: 0;
  letter-spacing: 1px;
}

.df-weekday {
  font-size: var(--fs-xs);
  font-weight: 400;
  color: var(--text-dim);
  margin-left: 0.5rem;
}

.df-lunar {
  font-size: var(--fs-xs);
  color: var(--text-dim);
  margin: 0;
}

.df-pillar-col {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  position: relative;
  gap: 0.2rem;
}

.df-pillar-glow {
  position: absolute;
  top: -15px;
  right: -15px;
  width: 80px;
  height: 80px;
  background: radial-gradient(
    circle,
    color-mix(in oklab, var(--accent) 8%, transparent),
    transparent 70%
  );
  pointer-events: none;
}

.df-pillar-val {
  font-size: var(--fs-hero);
  font-weight: 950;
  color: var(--accent);
  letter-spacing: 0.05em;
  line-height: 1;
  text-shadow: 0 0 30px color-mix(in oklab, var(--accent) 40%, transparent);
}

.df-sx {
  font-size: var(--fs-xs);
  color: var(--text-soft);
  margin: 0;
  text-align: right;
}

/* Guide */
.df-guide {
  display: grid;
  gap: 0.75rem;
  padding: 1rem;
}

.df-guide-empty {
  padding: 0.9rem;
  color: var(--text-muted);
}

.df-guide-empty p {
  margin: 0;
  font-size: var(--fs-xs);
  line-height: 1.65;
}

.guide-precision {
  margin-left: auto;
  font-size: var(--fs-2xs);
  color: rgba(var(--jade-accent-rgb), 1);
  background: rgba(var(--jade-accent-rgb), 0.08);
  border: 1px solid rgba(var(--jade-accent-rgb), 0.18);
  padding: 0.14rem 0.45rem;
  border-radius: 999px;
  font-weight: 700;
  letter-spacing: 0.08em;
}

.guide-confidence {
  font-size: var(--fs-2xs);
  color: var(--text-soft);
  letter-spacing: 0.08em;
}

.guide-hero {
  display: grid;
  grid-template-columns: minmax(0, 1.05fr) minmax(280px, 0.95fr);
  gap: 0.75rem;
  align-items: stretch;
}

.guide-brief {
  display: grid;
  align-content: start;
  gap: 0.6rem;
  min-width: 0;
  padding: 0.72rem;
  border: 1px solid rgba(var(--jade-accent-rgb), 0.16);
  border-radius: 10px;
  background:
    linear-gradient(135deg, rgba(var(--jade-accent-rgb), 0.07), transparent 58%), var(--glass-bg);
}

.guide-axis {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
  margin: 0;
}

.guide-axis span {
  padding: 0.16rem 0.45rem;
  border-radius: 999px;
  border: 1px solid var(--line-subtle);
  background: var(--glass-bg);
  color: var(--text-muted);
  font-size: var(--fs-2xs);
  font-weight: 700;
}

.guide-essentials {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.4rem;
  min-width: 0;
}

.guide-item {
  display: flex;
  flex-direction: column;
  gap: 0.16rem;
  min-height: 66px;
  padding: 0.5rem 0.55rem;
  background: var(--glass-bg);
  border-radius: 8px;
  border: 1px solid var(--line-subtle);
  transition: background 0.25s;
}

.guide-item:hover {
  background: var(--glass-bg-hover);
}

.guide-label {
  font-size: var(--fs-2xs);
  color: var(--text-soft);
  text-transform: uppercase;
  letter-spacing: 0.1em;
}

.guide-value {
  font-size: var(--fs-sm);
  color: var(--text);
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 0.4rem;
  min-width: 0;
}

.guide-value-xl {
  color: var(--accent);
  font-size: var(--fs-lg);
  font-weight: 900;
  letter-spacing: 1px;
  text-shadow: 0 0 20px var(--accent-glow);
}

.color-swatch {
  display: inline-block;
  width: 16px;
  height: 16px;
  border-radius: 50%;
  border: 1px solid var(--line-strong);
  flex-shrink: 0;
}

.guide-reason {
  font-size: var(--fs-2xs);
  line-height: 1.5;
  color: var(--text-soft);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.guide-strategy {
  margin: 0;
  color: var(--accent);
  border-left: 2px solid var(--line-focus);
  padding-left: 0.55rem;
  font-size: var(--fs-xs);
  line-height: 1.75;
}

.guide-action-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.75rem;
  margin-top: 0;
}

.guide-list {
  display: flex;
  flex-direction: column;
  gap: 0.46rem;
  min-width: 0;
  padding: 0.66rem;
  border: 1px solid rgba(var(--jade-accent-rgb), 0.17);
  border-radius: 10px;
  background: rgba(var(--jade-accent-rgb), 0.035);
}

.guide-list:has(.guide-list-warn) {
  border-color: color-mix(in oklab, var(--crimson) 16%, transparent);
  background: color-mix(in oklab, var(--crimson) 4%, transparent);
}

.guide-list-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
}

.guide-list-head > div {
  display: grid;
  gap: 0.12rem;
  min-width: 0;
}

.guide-list-title {
  font-size: var(--fs-2xs);
  color: rgba(var(--jade-accent-rgb), 1);
  letter-spacing: 0.12em;
}

.guide-list-head strong {
  color: var(--text);
  font-size: var(--fs-sm);
  line-height: 1.2;
}

.guide-list-warn {
  color: var(--crimson);
}

.guide-list-count {
  font-size: var(--fs-2xs);
  color: rgba(var(--jade-accent-rgb), 0.82);
  border: 1px solid rgba(var(--jade-accent-rgb), 0.16);
  background: rgba(var(--jade-accent-rgb), 0.06);
  border-radius: 999px;
  padding: 0.08rem 0.38rem;
}

.guide-list-count.warn {
  color: var(--crimson);
  border-color: color-mix(in oklab, var(--crimson) 16%, transparent);
  background: color-mix(in oklab, var(--crimson) 5%, transparent);
}

.guide-mini {
  display: flex;
  flex-direction: column;
  gap: 0.32rem;
  min-height: 0;
  padding: 0.58rem 0.62rem;
  border-radius: 8px;
  border: 1px solid rgba(var(--jade-accent-rgb), 0.16);
  background: color-mix(in oklab, var(--surface-1) 78%, rgba(var(--jade-accent-rgb), 0.06));
}

.guide-mini.warn {
  border-color: color-mix(in oklab, var(--crimson) 16%, transparent);
  background: color-mix(
    in oklab,
    var(--surface-1) 78%,
    color-mix(in oklab, var(--crimson) 7%, transparent)
  );
}

.guide-mini-secondary {
  padding: 0.48rem 0.54rem;
  opacity: 0.92;
}

.guide-mini-top {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  align-items: center;
  gap: 0.38rem;
}

.guide-mini-label {
  font-size: var(--fs-2xs);
  color: rgba(var(--jade-accent-rgb), 1);
  border: 1px solid rgba(var(--jade-accent-rgb), 0.18);
  background: rgba(var(--jade-accent-rgb), 0.08);
  border-radius: 5px;
  padding: 0.08rem 0.28rem;
  white-space: nowrap;
}

.guide-mini-label.warn {
  color: var(--crimson);
  border-color: color-mix(in oklab, var(--crimson) 18%, transparent);
  background: color-mix(in oklab, var(--crimson) 6%, transparent);
}

.guide-mini-top strong {
  color: var(--text);
  font-size: var(--fs-sm);
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.guide-intensity {
  font-size: var(--fs-2xs);
  color: var(--text-soft);
  border: 1px solid var(--line-subtle);
  background: var(--glass-bg);
  border-radius: 999px;
  padding: 0.06rem 0.3rem;
  white-space: nowrap;
}

.guide-intensity.warn {
  color: var(--crimson);
}

.guide-mini-reason {
  color: var(--text-soft);
  line-height: 1.55;
  font-size: var(--fs-2xs);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.guide-more,
.guide-extra {
  border: 1px solid var(--line-subtle);
  border-radius: 7px;
  background: color-mix(in oklab, var(--surface-1) 70%, transparent);
}

.guide-more summary,
.guide-extra summary {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.4rem;
  padding: 0.34rem 0.45rem;
  color: var(--text-muted);
  cursor: pointer;
  font-size: var(--fs-2xs);
  font-weight: 800;
  list-style: none;
}

.guide-more summary::-webkit-details-marker,
.guide-extra summary::-webkit-details-marker {
  display: none;
}

.guide-more summary::after,
.guide-extra summary::after {
  content: '+';
  color: var(--accent);
  font-size: var(--fs-xs);
}

.guide-more[open] summary::after,
.guide-extra[open] summary::after {
  content: '-';
}

.guide-more.warn summary::after,
.guide-extra.warn summary::after {
  color: var(--crimson);
}

.guide-more .guide-detail-row,
.guide-more .guide-method {
  margin: 0 0.45rem 0.45rem;
}

.guide-extra-list {
  display: grid;
  gap: 0.4rem;
  padding: 0 0.45rem 0.45rem;
}

.guide-detail-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0.24rem;
}

.guide-detail-row span {
  min-width: 0;
  max-width: 100%;
  border: 1px solid rgba(var(--jade-accent-rgb), 0.12);
  background: rgba(var(--jade-accent-rgb), 0.04);
  color: var(--text-muted);
  border-radius: 999px;
  padding: 0.08rem 0.34rem;
  font-size: var(--fs-2xs);
  line-height: 1.25;
}

.guide-detail-row.warn span {
  border-color: color-mix(in oklab, var(--crimson) 13%, transparent);
  background: color-mix(in oklab, var(--crimson) 4%, transparent);
}

.guide-method {
  margin: 0;
  padding-left: 0.45rem;
  border-left: 2px solid rgba(var(--jade-accent-rgb), 0.2);
  color: var(--text-muted);
  font-size: var(--fs-2xs);
  line-height: 1.45;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.guide-mini.warn .guide-method {
  border-left-color: color-mix(in oklab, var(--crimson) 22%, transparent);
}

.guide-footer {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  margin-top: 0.55rem;
}

.guide-hours {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
}

.guide-hours span {
  padding: 0.24rem 0.55rem;
  border-radius: 999px;
  background: var(--glass-bg);
  border: 1px solid var(--line-subtle);
  color: var(--text-muted);
  font-size: var(--fs-2xs);
}

.guide-note {
  border: 1px solid var(--line-subtle);
  border-radius: 8px;
  background: var(--glass-bg);
}

.guide-note summary {
  padding: 0.45rem 0.6rem;
  color: var(--text-muted);
  cursor: pointer;
  font-size: var(--fs-2xs);
  font-weight: 800;
  list-style: none;
}

.guide-note summary::-webkit-details-marker {
  display: none;
}

.guide-note summary::after {
  content: '+';
  float: right;
  color: var(--accent);
}

.guide-note[open] summary::after {
  content: '-';
}

.guide-note .guide-analysis {
  margin: 0;
  padding: 0 0.6rem 0.6rem;
  border-top: 1px solid var(--line-subtle);
  font-size: var(--fs-2xs);
  color: var(--text-muted);
  line-height: 1.55;
}

/* Hours + Elements */
.df-bottom-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.5rem;
}

.df-hours,
.df-elems {
  padding: 0.9rem;
}

.df-sec-header {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  margin-bottom: 0.6rem;
}

.df-sec-title {
  font-size: var(--fs-xs);
  font-weight: 800;
  color: var(--accent);
  margin: 0;
  letter-spacing: 2px;
}

.df-hours-list {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
}

.df-hour-chip {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 0.25rem 0.65rem;
  background: var(--glass-bg);
  border: 1px solid var(--line-subtle);
  border-radius: 20px;
  font-size: var(--fs-xs);
  color: var(--text-muted);
  transition: all 0.25s;
}

.df-hour-chip:hover {
  background: var(--accent-dim);
  border-color: var(--line-focus);
  color: var(--accent);
}

.df-hour-dot {
  width: 4px;
  height: 4px;
  border-radius: 50%;
  background: var(--accent);
  box-shadow: 0 0 6px var(--accent-glow);
  flex-shrink: 0;
}

.df-el-bars {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.df-el-row {
  display: flex;
  align-items: center;
  gap: 0.4rem;
}

.df-el-name {
  width: 14px;
  font-size: var(--fs-2xs);
  font-weight: 800;
  color: var(--text-soft);
  flex-shrink: 0;
}

.df-el-track {
  flex: 1;
  height: 5px;
  background: var(--line-subtle);
  border-radius: 3px;
  overflow: hidden;
}

.df-el-fill {
  height: 100%;
  border-radius: 3px;
  transition: width 0.8s ease;
}

.df-el-num {
  width: 18px;
  font-size: var(--fs-2xs);
  color: var(--text-soft);
  text-align: right;
  flex-shrink: 0;
}

/* AI btn */
.df-ai-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  width: 100%;
  padding: 0.7rem 1rem;
  background: var(--glass-bg);
  color: var(--text-dim);
  border: 1px solid var(--line-subtle);
  border-radius: 10px;
  font-size: var(--fs-sm);
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s;
  letter-spacing: 1.5px;
}

.df-ai-btn:hover {
  border-color: var(--text-soft);
  color: var(--accent);
  background: var(--accent-dim);
}

.df-ai-btn-icon {
  font-size: var(--fs-body);
}

/* Almanac 黄历 */
.df-almanac {
  padding: 0.9rem;
}

.almanac-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.4rem;
}

.almanac-item {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
  padding: 0.4rem 0.5rem;
  background: var(--glass-bg);
  border-radius: 6px;
  border: 1px solid var(--line-subtle);
}

.almanac-full {
  grid-column: 1 / -1;
}

.almanac-label {
  font-size: var(--fs-2xs);
  color: var(--text-soft);
  letter-spacing: 0.5px;
}

.almanac-value {
  font-size: var(--fs-xs);
  color: var(--text-muted);
  line-height: 1.4;
}

.almanac-ji {
  color: rgba(var(--jade-accent-rgb), 1);
}

.almanac-xiong {
  color: #dc2626;
}

/* Analysis 运势分析 */
.df-analysis {
  padding: 0.9rem;
}

.analysis-items {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.analysis-item {
  display: flex;
  gap: 0.5rem;
  font-size: var(--fs-xs);
  line-height: 1.5;
  padding: 0.3rem 0.5rem;
  background: var(--glass-bg);
  border-radius: 5px;
}

.analysis-label {
  min-width: 56px;
  font-weight: 600;
  color: var(--text-soft);
  flex-shrink: 0;
}

.analysis-value {
  color: var(--text-muted);
}

.analysis-gold {
  color: var(--accent);
}

/* Dimension matrix */
.dimension-panel {
  margin: 0.5rem 0 0;
  padding: 1rem;
  border: 1px solid var(--line-strong);
  border-radius: 12px;
  background:
    linear-gradient(
      135deg,
      color-mix(in oklab, var(--surface-1) 90%, transparent),
      color-mix(in oklab, var(--surface-0) 82%, transparent)
    ),
    linear-gradient(90deg, rgba(var(--jade-accent-rgb), 0.08), transparent 38%);
  box-shadow:
    var(--shadow-lg),
    inset 0 1px 0 var(--line-subtle);
  position: relative;
  overflow: hidden;
}

.dimension-panel::before {
  content: '';
  position: absolute;
  inset: 0 0 auto;
  height: 2px;
  background: linear-gradient(90deg, var(--accent), transparent 76%);
  opacity: 0.7;
}

.dimension-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
  margin-bottom: 0.7rem;
  position: relative;
  z-index: 1;
}

.dimension-eyebrow {
  display: block;
  color: var(--accent);
  font-size: var(--fs-2xs);
  font-weight: 800;
  letter-spacing: 0.12em;
}

.dimension-title {
  margin: 0.18rem 0 0;
  color: var(--text);
  font-family: var(--font-serif);
  font-size: var(--fs-sm);
  font-weight: 800;
  letter-spacing: 0;
}

.score-breakdown {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 0.4rem;
  max-width: 320px;
}

.score-breakdown span,
.score-breakdown strong {
  padding: 0.28rem 0.5rem;
  border-radius: 7px;
  background: var(--glass-bg);
  border: 1px solid var(--line-subtle);
  color: var(--text-muted);
  font-size: var(--fs-2xs);
  font-weight: 700;
}

.score-breakdown strong {
  color: var(--accent);
  background: var(--accent-dim);
  border-color: var(--line-focus);
}

.dimension-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.42rem;
  position: relative;
  z-index: 1;
}

.dimension-card {
  min-height: 0;
  border: 1px solid var(--line-subtle);
  border-radius: 9px;
  background: color-mix(in oklab, var(--surface-1) 82%, transparent);
  box-shadow: inset 0 1px 0 var(--line-subtle);
  animation: dimension-in 0.55s ease both;
  overflow: hidden;
  transition:
    background 0.25s,
    border-color 0.25s,
    box-shadow 0.25s;
}

@keyframes dimension-in {
  from {
    opacity: 0;
    transform: translateY(12px);
  }

  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.dimension-card:hover,
.dimension-card[open] {
  background: color-mix(in oklab, var(--surface-1) 72%, var(--accent-dim));
  border-color: var(--line-focus);
  box-shadow:
    0 10px 26px rgba(var(--jade-accent-rgb), 0.1),
    inset 0 1px 0 var(--line-subtle);
}

.dimension-summary {
  display: grid;
  grid-template-columns: 1.35rem minmax(4.3rem, 1fr) minmax(3.8rem, 0.85fr) 2.4rem 2.8rem;
  align-items: center;
  gap: 0.42rem;
  min-height: 54px;
  padding: 0.48rem 0.55rem;
  cursor: pointer;
  list-style: none;
}

.dimension-summary::-webkit-details-marker {
  display: none;
}

.dimension-rank {
  color: var(--text-soft);
  font-size: var(--fs-2xs);
  font-weight: 800;
  letter-spacing: 0.08em;
}

.dimension-main {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.dimension-name {
  display: block;
  color: var(--text);
  font-size: var(--fs-xs);
  font-weight: 850;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.dimension-keyword {
  display: block;
  margin-top: 0.12rem;
  color: var(--text-soft);
  font-size: var(--fs-2xs);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.dimension-score {
  color: var(--accent);
  font-family: var(--font-serif);
  font-size: var(--fs-lg);
  font-weight: 900;
  line-height: 1;
  text-shadow: 0 0 24px var(--accent-glow);
  text-align: right;
}

.dimension-state {
  padding: 0.18rem 0.32rem;
  border-radius: 6px;
  background: rgba(var(--jade-accent-rgb), 0.08);
  border: 1px solid rgba(var(--jade-accent-rgb), 0.15);
  color: var(--text-muted);
  font-size: var(--fs-2xs);
  font-weight: 800;
  text-align: center;
  white-space: nowrap;
}

.dimension-meter {
  height: 4px;
  border-radius: 999px;
  background: color-mix(in oklab, var(--surface-0) 70%, transparent);
  border: 1px solid var(--line-subtle);
  overflow: hidden;
}

.dimension-meter span {
  display: block;
  height: 100%;
  min-width: 6%;
  border-radius: inherit;
  background: var(--accent);
  box-shadow: 0 0 18px var(--accent-glow);
}

.dimension-detail {
  padding: 0 0.6rem 0.62rem 2.25rem;
  display: grid;
  gap: 0.46rem;
}

.dimension-meta-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
}

.dimension-meta-row span,
.dimension-tags span {
  padding: 0.18rem 0.38rem;
  border-radius: 7px;
  background: var(--glass-bg);
  border: 1px solid var(--line-subtle);
  color: var(--text-muted);
  font-size: var(--fs-2xs);
  font-weight: 700;
}

.dimension-analysis {
  margin: 0;
  color: var(--text-muted);
  font-size: var(--fs-2xs);
  line-height: 1.55;
}

.dimension-advice {
  margin: 0;
  padding-left: 0.55rem;
  border-left: 2px solid var(--line-focus);
  color: var(--accent);
  font-size: var(--fs-2xs);
  line-height: 1.5;
}

.dimension-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
}

.dimension-card.trend-up .dimension-meta-row span:nth-child(2) {
  color: var(--accent);
  border-color: var(--line-focus);
}

.dimension-card.trend-down .dimension-state,
.dimension-card.trend-down .dimension-meta-row span:first-child {
  color: var(--crimson);
  border-color: color-mix(in oklab, var(--crimson) 22%, transparent);
}

/* Modal */
.df-modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 1rem;
  backdrop-filter: blur(8px);
}

.df-modal-box {
  width: 100%;
  max-width: 380px;
  overflow: hidden;
}

.df-modal-hdr {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.25rem 1.5rem;
  border-bottom: 1px solid var(--line-subtle);
}

.df-modal-title-group {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.df-modal-orb {
  font-size: var(--fs-3xl);
  color: var(--accent);
  text-shadow: 0 0 25px var(--accent-glow);
  animation: orb-glow 3s ease-in-out infinite;
}

@keyframes orb-glow {
  0%,
  100% {
    text-shadow: 0 0 20px var(--accent-glow);
  }

  50% {
    text-shadow: 0 0 40px var(--accent-glow);
  }
}

.df-modal-hdr h2 {
  margin: 0;
  font-family: var(--font-serif), serif;
  font-size: var(--fs-lg);
  font-weight: 700;
  color: var(--text);
  letter-spacing: 2px;
}

.df-modal-close {
  background: none;
  border: none;
  font-size: var(--fs-2xl);
  color: var(--text-soft);
  cursor: pointer;
  padding: 0.25rem;
  transition: color 0.2s;
}

.df-modal-close:hover {
  color: var(--accent);
}

.df-modal-body {
  padding: 2.5rem 1.5rem;
}

.df-ai-coming {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
  text-align: center;
}

.df-ai-svg {
  color: var(--icon-muted);
  opacity: 0.75;
  animation: svg-rot 20s linear infinite;
}

@keyframes svg-rot {
  from {
    transform: rotate(0deg);
  }

  to {
    transform: rotate(360deg);
  }
}

.df-star-pulse {
  animation: st-pulse 2.5s ease-in-out infinite;
}

@keyframes st-pulse {
  0%,
  100% {
    opacity: 0.2;
  }

  50% {
    opacity: 0.7;
  }
}

.df-ai-title {
  font-size: var(--fs-body);
  font-weight: 700;
  color: var(--text);
  margin: 0;
}

.df-ai-sub {
  font-size: var(--fs-xs);
  color: var(--text-soft);
  margin: 0;
}

/* Modal transition */
.df-modal-enter-active,
.df-modal-leave-active {
  transition: opacity 0.25s ease;
}

.df-modal-enter-from,
.df-modal-leave-to {
  opacity: 0;
}

.df-modal-enter-active .df-modal-box,
.df-modal-leave-active .df-modal-box {
  transition: transform 0.25s ease;
}

.df-modal-enter-from .df-modal-box {
  transform: scale(0.9) translateY(12px);
}

.df-modal-leave-to .df-modal-box {
  transform: scale(0.9) translateY(12px);
}

/* ═══ 日课推算样式 ═══ */

/* 综合断语 */
.df-rikuyo-verdict {
  padding: 1rem;
  position: relative;
}

.df-rikuyo-verdict::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 2px;
  background: linear-gradient(90deg, transparent, var(--accent), transparent);
}

.rikuyo-score-badge {
  margin-left: auto;
  font-size: var(--fs-xs);
  font-weight: 800;
  padding: 0.15rem 0.6rem;
  border-radius: 20px;
  letter-spacing: 0.5px;
  --score-t: 0.5;
  color: color-mix(in oklab, var(--jade-accent) calc(58% + var(--score-t) * 42%), var(--text));
  background: rgba(var(--jade-accent-rgb), 0.12);
  border: 1px solid rgba(var(--jade-accent-rgb), 0.25);
}

.rikuyo-verdict-text {
  font-size: var(--fs-sm);
  color: var(--text-muted);
  line-height: 1.8;
  margin: 0.5rem 0 0;
  white-space: pre-wrap;
}

/* 格局信息 */
.df-pattern-info {
  padding: 0.8rem 1rem;
}

.pattern-badge {
  font-size: var(--fs-xs);
  color: var(--accent);
  background: var(--accent-dim);
  padding: 0.15rem 0.5rem;
  border-radius: 4px;
  margin-left: auto;
}

.pattern-elements {
  display: flex;
  gap: 0.5rem;
  margin-top: 0.4rem;
}

.pattern-tag {
  font-size: var(--fs-xs);
  padding: 0.2rem 0.6rem;
  border-radius: 4px;
  font-weight: 600;
  letter-spacing: 0.5px;
}

.pattern-like {
  background: rgba(var(--jade-accent-rgb), 0.12);
  color: rgba(var(--jade-accent-rgb), 1);
  border: 1px solid rgba(var(--jade-accent-rgb), 0.24);
}

.pattern-dislike {
  background: rgba(220, 38, 38, 0.1);
  color: #dc2626;
  border: 1px solid rgba(220, 38, 38, 0.2);
}

/* 十神 + 长生核心区 */
.df-rikuyo-core {
  padding: 1rem;
}

.rikuyo-core-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.75rem;
}

.rikuyo-core-item {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  padding: 0.75rem;
  background: var(--glass-bg);
  border-radius: 8px;
  border: 1px solid var(--line-subtle);
}

.rikuyo-core-label {
  font-size: var(--fs-2xs);
  color: var(--text-soft);
  letter-spacing: 0.5px;
  text-transform: uppercase;
}

.rikuyo-core-value {
  font-family: var(--font-serif);
  font-size: var(--fs-2xl);
  font-weight: 900;
  letter-spacing: 2px;
  line-height: 1.2;
}

.val-fav {
  color: rgba(var(--jade-accent-rgb), 1);
  text-shadow: 0 0 12px rgba(var(--jade-accent-rgb), 0.28);
}

.val-dis {
  color: var(--text-soft);
  text-shadow: 0 0 12px color-mix(in oklab, var(--text-soft) 20%, transparent);
}

.rikuyo-core-desc {
  font-size: var(--fs-xs);
  color: var(--text-muted);
  line-height: 1.5;
}

.rikuyo-flexible {
  font-size: var(--fs-xs);
  color: var(--accent);
  font-style: italic;
  padding: 0.3rem 0.5rem;
  background: var(--accent-dim);
  border-radius: 4px;
  border-left: 2px solid var(--line-focus);
  margin-top: 0.25rem;
}

/* 进退气 */
.df-rikuyo-advance {
  padding: 0.9rem;
}

.rikuyo-phase-tag {
  margin-left: auto;
  font-size: var(--fs-2xs);
  font-weight: 700;
  padding: 0.1rem 0.5rem;
  border-radius: 10px;
  letter-spacing: 0.5px;
}

.phase-adv {
  background: rgba(var(--jade-accent-rgb), 0.12);
  color: rgba(var(--jade-accent-rgb), 1);
}

.phase-peak {
  background: var(--accent-dim);
  color: var(--accent);
}

.phase-ret {
  background: rgba(194, 65, 12, 0.1);
  color: #c2410c;
}

.phase-dead {
  background: rgba(220, 38, 38, 0.1);
  color: #dc2626;
}

.rikuyo-advance-text {
  font-size: var(--fs-xs);
  color: var(--text-muted);
  line-height: 1.6;
  margin: 0.4rem 0 0;
}

/* 藏干 */
.df-rikuyo-hidden {
  padding: 0.9rem;
}

.hidden-stems-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.hidden-stem-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.15rem;
  padding: 0.6rem 0.8rem;
  border-radius: 8px;
  min-width: 60px;
  border: 1px solid var(--line-subtle);
  background: var(--glass-bg);
  transition: all 0.25s;
}

.hs-fav {
  border-color: rgba(var(--jade-accent-rgb), 0.18);
  background: rgba(var(--jade-accent-rgb), 0.04);
}

.hs-dis {
  border-color: rgba(220, 38, 38, 0.12);
  background: rgba(220, 38, 38, 0.03);
}

.hs-stem {
  font-family: var(--font-serif);
  font-size: var(--fs-2xl);
  font-weight: 900;
  color: var(--text);
}

.hs-type {
  font-size: var(--fs-2xs);
  color: var(--text-soft);
}

.hs-god {
  font-size: var(--fs-2xs);
  font-weight: 700;
}

.hs-fav .hs-god {
  color: rgba(var(--jade-accent-rgb), 1);
}

.hs-dis .hs-god {
  color: #dc2626;
}

.hs-elem {
  font-size: var(--fs-2xs);
  color: var(--text-soft);
}

/* 干支关系 */
.df-rikuyo-relations {
  padding: 0.9rem;
}

.relations-list {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.relation-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.4rem 0.6rem;
  border-radius: 6px;
  background: var(--glass-bg);
  border: 1px solid var(--line-subtle);
}

.rel-fav {
  border-color: rgba(var(--jade-accent-rgb), 0.18);
}

.rel-dis {
  border-color: rgba(220, 38, 38, 0.12);
}

.rel-type-tag {
  font-size: var(--fs-2xs);
  font-weight: 700;
  padding: 0.1rem 0.4rem;
  border-radius: 3px;
  flex-shrink: 0;
  letter-spacing: 0.5px;
}

.rel-fav .rel-type-tag {
  background: rgba(var(--jade-accent-rgb), 0.1);
  color: rgba(var(--jade-accent-rgb), 1);
}

.rel-dis .rel-type-tag {
  background: rgba(220, 38, 38, 0.08);
  color: #dc2626;
}

.rel-detail {
  font-size: var(--fs-xs);
  color: var(--text-muted);
}

.rel-note {
  font-size: var(--fs-2xs);
  color: var(--accent);
  font-style: italic;
  margin-left: auto;
}

/* 神煞 */
.df-rikuyo-shensha {
  padding: 0.9rem;
}

.shensha-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.shensha-item {
  padding: 0.6rem 0.8rem;
  border-radius: 8px;
  border: 1px solid var(--line-subtle);
  background: var(--glass-bg);
}

.ss-ji {
  border-color: rgba(var(--jade-accent-rgb), 0.18);
}

.ss-xiong {
  border-color: rgba(220, 38, 38, 0.12);
}

.ss-name {
  font-size: var(--fs-sm);
  font-weight: 800;
  color: var(--text);
  margin-right: 0.5rem;
}

.ss-type-tag {
  font-size: var(--fs-2xs);
  padding: 0.1rem 0.4rem;
  border-radius: 3px;
}

.ss-ji .ss-type-tag {
  background: rgba(var(--jade-accent-rgb), 0.1);
  color: rgba(var(--jade-accent-rgb), 1);
}

.ss-xiong .ss-type-tag {
  background: rgba(220, 38, 38, 0.08);
  color: #dc2626;
}

.ss-desc {
  font-size: var(--fs-xs);
  color: var(--text-muted);
  line-height: 1.5;
  margin: 0.3rem 0 0;
}

.ss-activation {
  font-size: var(--fs-2xs);
  color: var(--text-soft);
  margin: 0.2rem 0 0;
  font-style: italic;
}

/* 大运流年 */
.df-rikuyo-yun {
  padding: 0.9rem;
}

.yun-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.75rem;
}

.yun-item {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
  padding: 0.7rem;
  background: var(--glass-bg);
  border-radius: 8px;
  border: 1px solid var(--line-subtle);
}

.yun-label {
  font-size: var(--fs-2xs);
  color: var(--text-soft);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.yun-pillar {
  font-family: var(--font-serif);
  font-size: var(--fs-2xl);
  font-weight: 900;
  color: var(--text);
  letter-spacing: 2px;
}

.yun-god {
  font-size: var(--fs-xs);
  font-weight: 700;
}

.yun-age {
  font-size: var(--fs-2xs);
  color: var(--text-soft);
}

.yun-taisui {
  font-size: var(--fs-2xs);
  color: var(--crimson);
  margin: 0.2rem 0 0;
  font-style: italic;
}

.yun-desc {
  font-size: var(--fs-xs);
  color: var(--text-muted);
  line-height: 1.5;
  margin: 0.2rem 0 0;
}

/* 用神影响 */
.df-rikuyo-yongshen {
  padding: 0.9rem;
}

.yongshen-items {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.yongshen-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 0.7rem;
  border-radius: 6px;
  background: var(--glass-bg);
  border: 1px solid var(--line-subtle);
  transition: all 0.25s;
}

.ys-hit {
  border-color: rgba(var(--jade-accent-rgb), 0.22);
  background: rgba(var(--jade-accent-rgb), 0.07);
}

.ys-label {
  font-size: var(--fs-2xs);
  color: var(--text-soft);
  min-width: 56px;
  flex-shrink: 0;
}

.ys-elem {
  font-size: var(--fs-sm);
  font-weight: 700;
  color: var(--text);
}

.ys-status {
  font-size: var(--fs-2xs);
  margin-left: auto;
  padding: 0.1rem 0.4rem;
  border-radius: 3px;
}

.ys-hit .ys-status {
  background: rgba(var(--jade-accent-rgb), 0.12);
  color: rgba(var(--jade-accent-rgb), 1);
}

.yongshen-desc {
  font-size: var(--fs-xs);
  color: var(--text-muted);
  margin: 0.5rem 0 0;
  line-height: 1.5;
}

/* ═══ Dark mode overrides ═══ */
:global(.dark) .guide-precision {
  background: rgba(var(--jade-accent-rgb), 0.12);
  color: rgba(var(--jade-accent-rgb), 1);
  border-color: rgba(var(--jade-accent-rgb), 0.24);
}

:global(.dark) .guide-axis span,
:global(.dark) .guide-item,
:global(.dark) .guide-mini,
:global(.dark) .guide-note {
  background: var(--glass-bg);
}

:global(.dark) .almanac-ji {
  color: rgba(var(--jade-accent-rgb), 1);
}

:global(.dark) .almanac-xiong {
  color: #f08080;
}

:global(.dark) .rikuyo-score-badge {
  color: color-mix(in oklab, var(--jade-accent) calc(58% + var(--score-t) * 42%), var(--text));
  background: rgba(var(--jade-accent-rgb), 0.12);
  border-color: rgba(var(--jade-accent-rgb), 0.25);
}

:global(.dark) .pattern-like {
  background: rgba(var(--jade-accent-rgb), 0.12);
  color: rgba(var(--jade-accent-rgb), 1);
  border-color: rgba(var(--jade-accent-rgb), 0.24);
}

:global(.dark) .pattern-dislike {
  background: rgba(251, 113, 133, 0.1);
  color: #fb7185;
  border-color: rgba(251, 113, 133, 0.2);
}

:global(.dark) .val-fav {
  color: rgba(var(--jade-accent-rgb), 1);
  text-shadow: 0 0 15px rgba(var(--jade-accent-rgb), 0.3);
}

:global(.dark) .val-dis {
  color: var(--text-soft);
  text-shadow: 0 0 15px color-mix(in oklab, var(--text-soft) 20%, transparent);
}

:global(.dark) .phase-adv {
  background: rgba(var(--jade-accent-rgb), 0.12);
  color: rgba(var(--jade-accent-rgb), 1);
}

:global(.dark) .phase-ret {
  background: rgba(255, 165, 0, 0.1);
  color: #ffa500;
}

:global(.dark) .phase-dead {
  background: rgba(251, 113, 133, 0.1);
  color: #fb7185;
}

:global(.dark) .hs-fav {
  border-color: rgba(var(--jade-accent-rgb), 0.18);
  background: rgba(var(--jade-accent-rgb), 0.04);
}

:global(.dark) .hs-dis {
  border-color: rgba(251, 113, 133, 0.12);
  background: rgba(251, 113, 133, 0.03);
}

:global(.dark) .hs-fav .hs-god {
  color: rgba(var(--jade-accent-rgb), 1);
}

:global(.dark) .hs-dis .hs-god {
  color: #fb7185;
}

:global(.dark) .rel-fav {
  border-color: rgba(var(--jade-accent-rgb), 0.18);
}

:global(.dark) .rel-dis {
  border-color: rgba(251, 113, 133, 0.08);
}

:global(.dark) .rel-fav .rel-type-tag {
  background: rgba(var(--jade-accent-rgb), 0.1);
  color: rgba(var(--jade-accent-rgb), 1);
}

:global(.dark) .rel-dis .rel-type-tag {
  background: rgba(251, 113, 133, 0.08);
  color: #fb7185;
}

:global(.dark) .ss-ji {
  border-color: rgba(var(--jade-accent-rgb), 0.18);
}

:global(.dark) .ss-xiong {
  border-color: rgba(251, 113, 133, 0.1);
}

:global(.dark) .ss-ji .ss-type-tag {
  background: rgba(var(--jade-accent-rgb), 0.1);
  color: rgba(var(--jade-accent-rgb), 1);
}

:global(.dark) .ss-xiong .ss-type-tag {
  background: rgba(251, 113, 133, 0.08);
  color: #fb7185;
}

:global(.dark) .ys-hit {
  border-color: rgba(var(--jade-accent-rgb), 0.22);
  background: rgba(var(--jade-accent-rgb), 0.07);
}

:global(.dark) .ys-hit .ys-status {
  background: rgba(var(--jade-accent-rgb), 0.12);
  color: rgba(var(--jade-accent-rgb), 1);
}

:global(.dark) .df-pillar-val {
  text-shadow: 0 0 30px color-mix(in oklab, var(--accent) 40%, transparent);
}

@media (max-width: 720px) {
  .df-header {
    align-items: flex-start;
    gap: 1rem;
  }

  .df-pillar-val {
    font-size: var(--fs-stat-lg);
  }

  .guide-essentials {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .guide-hero {
    grid-template-columns: 1fr;
  }

  .guide-list-head {
    align-items: flex-start;
  }

  .guide-action-grid {
    grid-template-columns: 1fr;
  }

  .guide-mini-top {
    grid-template-columns: auto minmax(0, 1fr);
  }

  .guide-intensity {
    grid-column: 1 / -1;
    justify-self: start;
  }

  .dimension-panel {
    padding: 0.85rem;
    border-radius: 12px;
  }

  .dimension-head {
    align-items: flex-start;
    flex-direction: column;
  }

  .score-breakdown {
    justify-content: flex-start;
    max-width: none;
  }

  .guide-action-grid,
  .df-bottom-row,
  .dimension-grid,
  .rikuyo-core-grid,
  .yun-grid {
    grid-template-columns: 1fr;
  }

  .dimension-summary {
    grid-template-columns: 1.2rem minmax(4.8rem, 1fr) minmax(4rem, 0.8fr) 2.4rem;
    min-height: 50px;
  }

  .dimension-state {
    display: none;
  }

  .dimension-detail {
    padding-left: 2.05rem;
  }
}

@media (max-width: 480px) {
  .df-header {
    flex-direction: column;
  }

  .df-pillar-col {
    align-items: flex-start;
  }

  .guide-essentials,
  .almanac-grid {
    grid-template-columns: 1fr;
  }

  .guide-brief,
  .guide-list {
    padding: 0.56rem;
  }

  .guide-more summary,
  .guide-extra summary {
    padding: 0.3rem 0.38rem;
  }
}
</style>
