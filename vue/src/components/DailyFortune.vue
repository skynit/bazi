<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import FortuneEvidencePanel from './FortuneEvidencePanel.vue'
import InterpretationLevelSwitch from './InterpretationLevelSwitch.vue'
import type {
  BranchRelation,
  FortuneScoreBreakdown,
  HiddenStemGod,
  InterpretationLevel,
  ScoreEvidence,
  SeasonElementEvidence,
  SeasonalStateEvidence,
  StemRelation,
  TenGodEvidence,
  TraditionalCalendarEvidence,
  TwelveStageEvidence,
} from '../api/fortune'
import type { FortuneLayerSet, ShenShaActivation } from '../api/chart'

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
interface Props {
  solarDate: string
  dayGanZhi: string
  /** 页面级结论区已展示日期/日柱时，可隐藏组件内重复头部 */
  hideHeader?: boolean
  weekDay?: string
  lunarDate?: string
  shengXiao?: string
  elementImages?: ElementImage[]
  scoreBreakdown?: FortuneScoreBreakdown
  supportingEvidence?: ScoreEvidence[]
  counterEvidence?: ScoreEvidence[]
  todayElements?: Record<string, number>
  // 黄历字段
  jiShen?: string
  xiongShen?: string
  taiShen?: string
  pengZu?: string
  gua?: string
  jieQi?: string
  // 分析字段
  shengKeAnalysis?: ShengKeAnalysis
  seasonElement?: SeasonElementEvidence
  // 日课推算
  tenGod?: TenGodEvidence
  twelveStage?: TwelveStageEvidence
  jianChu?: TraditionalCalendarEvidence
  huangDao?: TraditionalCalendarEvidence
  hiddenStems?: HiddenStemGod[]
  stemRelations?: StemRelation[]
  branchRelations?: BranchRelation[]
  activatedShenSha?: ShenShaActivation[]
  seasonalState?: SeasonalStateEvidence
  fortuneLayers?: FortuneLayerSet
}
const props = withDefaults(defineProps<Props>(), {
  hideHeader: false,
  weekDay: '',
  lunarDate: '',
  shengXiao: '',
  elementImages: () => [],
  todayElements: () => ({}),
  scoreBreakdown: undefined,
  supportingEvidence: () => [],
  counterEvidence: () => [],
  jiShen: '',
  xiongShen: '',
  taiShen: '',
  pengZu: '',
  gua: '',
  jieQi: '',
  shengKeAnalysis: undefined,
  seasonElement: undefined,
  // 日课推算
  tenGod: undefined,
  twelveStage: undefined,
  jianChu: undefined,
  huangDao: undefined,
  hiddenStems: () => [],
  stemRelations: () => [],
  branchRelations: () => [],
  activatedShenSha: () => [],
  seasonalState: undefined,
  fortuneLayers: undefined,
})
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
  { key: 'analysis', label: '结构分析', minimum: 'advanced' },
  { key: 'elements', label: '五行结构', minimum: 'advanced' },
  { key: 'rikuyo', label: '日课规则', minimum: 'professional' },
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

function evidenceBasisLabel(evidence?: TwelveStageEvidence | TraditionalCalendarEvidence) {
  if (!evidence) return ''
  if ('reference_stem' in evidence) {
    return `日主 ${evidence.reference_stem} · 今日地支 ${evidence.query_branch}`
  }
  return `当月地支 ${evidence.month_branch} · 今日地支 ${evidence.query_branch}`
}

function tenGodBasisLabel(evidence?: TenGodEvidence) {
  if (!evidence) return ''
  return `日主 ${evidence.reference_stem} · 今日天干 ${evidence.query_stem}`
}

function seasonElementBasisLabel(evidence?: SeasonElementEvidence) {
  if (!evidence) return ''
  return `日主 ${evidence.reference_stem}${evidence.reference_element} · 当月地支 ${evidence.query_month_branch} · ${evidence.season}`
}

function calculationBasisLabel(value?: string) {
  const labels: Record<string, string> = {
    exact_start_time_and_query_time: '依据出生时间与查询日期定位',
    period_pillar_and_natal_chart: '依据周期干支与本命四柱对照',
    period_layer_stem_pair: '依据周期天干关系',
    period_stem_and_target_stem_all_structures: '依据周期天干与本命天干关系',
    period_branch_and_target_branch_all_structures: '依据周期地支与本命地支关系',
  }
  return value ? labels[value] || '依据干支关系计算' : ''
}

function seasonalStateLabel(value?: string): string {
  const labels: Record<string, string> = {
    旺: '旺势（传统季节标签）',
    相: '相势（传统季节标签）',
    休: '休势（传统季节标签）',
    囚: '囚势（传统季节标签）',
    死: '弱势（传统“死”标签）',
  }
  return value ? labels[value] || `${value}（传统季节标签）` : ''
}

function fortuneLayerLabel(value?: string): string {
  const labels: Record<string, string> = {
    dayun: '大运',
    liunian: '流年',
    liuyue: '流月',
    xiaoyun: '小运',
    natal: '本命',
    周期天干: '周期天干',
    周期地支: '周期地支',
    查询日干: '今日天干',
    查询日支: '今日地支',
    日干: '日主',
    年支: '本命年支',
    月支: '本命月支',
    日支: '本命日支',
    时支: '本命时支',
  }
  return value ? labels[value] || '周期层' : '周期层'
}

function natalPillarLabel(value?: string): string {
  const labels: Record<string, string> = {
    year: '本命年柱',
    month: '本命月柱',
    day: '本命日柱',
    hour: '本命时柱',
    年干: '本命年干',
    月干: '本命月干',
    日干: '本命日干',
    时干: '本命时干',
    年支: '本命年支',
    月支: '本命月支',
    日支: '本命日支',
    时支: '本命时支',
  }
  return value ? labels[value] || '本命四柱' : '本命四柱'
}

function fortuneRelationLabel(type?: string, fallback?: string): string {
  const labels: Record<string, string> = {
    shengWo: '生助',
    woSheng: '受生于',
    keWo: '克制',
    woKe: '受制于',
    same: '同支',
    same_stem: '同干',
    same_element: '同五行',
    five_combine: '天干五合',
    clash: '相冲',
    harm: '相害',
    combine: '六合',
    punish: '相刑',
    break: '相破',
    banHe: '半合',
    gongHe: '拱合',
    banHui: '半会',
    sanHe: '三合',
    sanHui: '三会',
  }
  return type ? labels[type] || fallback || '形成关系' : fallback || '形成关系'
}

const fortuneLayerList = computed(() => {
  const layers = props.fortuneLayers
  if (!layers) return []
  return [layers.dayun, layers.liunian, layers.liuyue, layers.xiaoyun].filter(
    (layer) => layer?.status === 'observed',
  )
})
</script>

<template>
  <div class="daily-fortune">
    <!-- Date + Pillar（页面级结论区已展示时可隐藏，避免重复） -->
    <div v-if="!hideHeader" class="df-header glass-card">
      <div class="df-date-col">
        <p class="df-solar">
          {{ solarDate }}<span v-if="weekDay" class="df-weekday">{{ weekDay }}</span>
        </p>
        <p v-if="lunarDate" class="df-lunar">{{ lunarDate }}</p>
      </div>
      <div class="df-pillar-col">
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
      <FortuneEvidencePanel
        :level="interpretationLevel"
        :supporting="supportingEvidence"
        :counter="counterEvidence"
        :breakdown="scoreBreakdown"
      />
    </div>
    <!-- /overview tab -->

    <!-- ═══ Tab: 黄历 ═══ -->
    <div v-show="activeTab === 'almanac'" class="df-tab-content">
      <p class="section-boundary-note almanac-boundary-note">
        黄历展示当天传统历书标签，参考对象是日期本身，不结合个人命盘，也不代表现实风险或行动结果。
      </p>
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

    <!-- ═══ Tab: 结构分析 ═══ -->
    <div v-show="activeTab === 'analysis'" class="df-tab-content">
      <!-- 生克分析 -->
      <div
        v-if="shengKeAnalysis?.summary || seasonElement?.status === 'observed'"
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
          <span class="df-sec-title">结构分析</span>
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
          <div v-if="seasonElement?.status === 'observed'" class="analysis-item">
            <span class="analysis-label">月令季节</span>
            <span class="analysis-value">{{ seasonElementBasisLabel(seasonElement) }}</span>
          </div>
        </div>
      </div>
    </div>
    <!-- /analysis tab -->

    <!-- ═══ Tab: 日课推算 ═══ -->
    <div v-show="activeTab === 'rikuyo'" class="df-tab-content">
      <!-- ═══ 日课推算 ═══ -->

      <!-- 今日十神 + 传统日课结构证据 -->
      <div
        v-if="tenGod?.name || twelveStage?.name || jianChu?.name || huangDao?.name"
        class="df-rikuyo-core glass-card"
      >
        <div class="df-sec-header">
          <span class="df-sec-title">传统日课</span>
        </div>
        <div class="rikuyo-core-grid">
          <!-- 十神 -->
          <div v-if="tenGod?.name" class="rikuyo-core-item rikuyo-evidence-item">
            <span class="rikuyo-core-label">今日十神</span>
            <span class="rikuyo-core-value rikuyo-evidence-value">{{ tenGod.name }}</span>
            <span class="rikuyo-core-desc">{{ tenGodBasisLabel(tenGod) }}</span>
          </div>
          <div v-if="twelveStage?.name" class="rikuyo-core-item rikuyo-evidence-item">
            <span class="rikuyo-core-label">十二长生</span>
            <span class="rikuyo-core-value rikuyo-evidence-value">{{ twelveStage.name }}</span>
            <span class="rikuyo-core-desc">{{ evidenceBasisLabel(twelveStage) }}</span>
          </div>
          <div v-if="jianChu?.name" class="rikuyo-core-item rikuyo-evidence-item">
            <span class="rikuyo-core-label">建除十二神</span>
            <span class="rikuyo-core-value rikuyo-evidence-value">{{ jianChu.name }}</span>
            <span class="rikuyo-core-desc">{{ evidenceBasisLabel(jianChu) }}</span>
          </div>
          <div v-if="huangDao?.name" class="rikuyo-core-item rikuyo-evidence-item">
            <span class="rikuyo-core-label">十二值神</span>
            <span class="rikuyo-core-value rikuyo-evidence-value">{{ huangDao.name }}</span>
            <span class="rikuyo-core-desc">{{ evidenceBasisLabel(huangDao) }}</span>
          </div>
        </div>
        <p class="section-boundary-note">以上为传统日课查表结果，仅供理解当天干支结构。</p>
      </div>

      <!-- 月令状态 -->
      <div v-if="seasonalState?.status === 'observed'" class="df-rikuyo-advance glass-card">
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
          <span class="df-sec-title">月令状态</span>
          <span class="rikuyo-phase-tag">{{ seasonalStateLabel(seasonalState.state) }}</span>
        </div>
        <p class="rikuyo-advance-text">
          当天日干 {{ seasonalState.query_stem }}（{{ seasonalState.query_element }}）在月支
          {{ seasonalState.query_month_branch }}所对应的{{ seasonalState.season }}季状态
        </p>
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
          <div v-for="hs in hiddenStems" :key="hs.stem + hs.type" class="hidden-stem-card">
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
          <div v-for="(sr, i) in stemRelations" :key="'sr' + i" class="relation-item">
            <span class="rel-type-tag">{{ sr.name }}</span>
            <span class="rel-detail"
              >今日天干 {{ sr.query_stem }} · {{ natalPillarLabel(sr.target_pillar) }}
              {{ sr.target_stem }}</span
            >
            <span v-if="sr.combined_element" class="rel-note"
              >合化方向为{{ sr.combined_element }}，仍需结合全盘条件判断</span
            >
          </div>
          <div v-for="(br, i) in branchRelations" :key="'br' + i" class="relation-item">
            <span class="rel-type-tag">{{ br.name }}</span>
            <span class="rel-detail"
              >今日地支 {{ br.query_branch }} · {{ natalPillarLabel(br.target_pillar) }}
              {{ br.target_branch }}</span
            >
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
          <div v-for="ss in activatedShenSha" :key="ss.name" class="shensha-item">
            <span class="ss-name">{{ ss.name }}</span>
            <span class="ss-type-tag">传统查法</span>
            <p class="ss-desc">按当天干支与本命四柱对应传统表项</p>
            <p class="ss-activation">{{ ss.activation }}</p>
          </div>
        </div>
      </div>

      <!-- 周期层结构 -->
      <div v-if="fortuneLayerList.length" class="df-rikuyo-yun glass-card">
        <div class="df-sec-header">
          <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
            <path d="M1 7h12M7 1v12" stroke="currentColor" stroke-width="0.8" opacity="0.3" />
            <circle cx="7" cy="7" r="3" stroke="currentColor" stroke-width="0.8" opacity="0.25" />
          </svg>
          <span class="df-sec-title">周期层结构</span>
        </div>
        <div class="yun-grid">
          <div v-for="layer in fortuneLayerList" :key="layer.key" class="yun-item">
            <span class="yun-label">{{ layer.name }}</span>
            <span class="yun-pillar">{{ layer.pillar }}</span>
            <span v-if="layer.ten_god?.name" class="yun-god">十神 {{ layer.ten_god.name }}</span>
            <span v-if="layer.start_age" class="yun-age"
              >{{ layer.start_age }}-{{ layer.end_age }}岁</span
            >
            <p class="yun-desc">{{ calculationBasisLabel(layer.basis) }}</p>
            <p v-if="layer.relations?.length" class="yun-taisui">
              <span
                v-for="relation in layer.relations.slice(0, 4)"
                :key="`${relation.source}-${relation.target}-${relation.type}`"
              >
                {{ fortuneLayerLabel(relation.source) }}{{ relation.source_value }} ·
                {{ fortuneRelationLabel(relation.type, relation.name) }} ·
                {{ fortuneLayerLabel(relation.target) }}{{ relation.target_value }}
              </span>
            </p>
          </div>
        </div>
        <div v-if="fortuneLayers?.inter_layer_relations?.length" class="yun-inter-layer-relations">
          <strong>岁运月联动结构</strong>
          <span
            v-for="relation in fortuneLayers.inter_layer_relations"
            :key="`${relation.source}-${relation.target}-${relation.type}`"
          >
            {{ fortuneLayerLabel(relation.source) }}{{ relation.source_value }} ·
            {{ fortuneRelationLabel(relation.type, relation.name) }} · {{ fortuneLayerLabel(relation.target)
            }}{{ relation.target_value }}
          </span>
          <small>这里只展示周期之间的干支关系，不据此判断具体事件。</small>
        </div>
      </div>
    </div>
    <!-- /rikuyo tab -->

    <!-- ═══ Tab: 五行结构 ═══ -->
    <div v-show="activeTab === 'elements'" class="df-tab-content">
      <div class="df-bottom-row">
        <div class="df-elems glass-card">
          <div class="df-sec-header">
            <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
              <circle cx="7" cy="7" r="6" stroke="currentColor" stroke-width="1" opacity="0.25" />
              <circle cx="7" cy="7" r="2.5" fill="currentColor" opacity="0.35" />
            </svg>
            <span class="df-sec-title">当日中午样本的五行结构</span>
          </div>
          <p class="section-boundary-note">
            取当天 12:00 的四柱按当前规则权重归一化为百分比，仅作结构参考，不代表全天变化或现实结果。
          </p>
          <div class="df-el-bars">
            <div v-for="[el, clr] in elementEntries" :key="el" class="df-el-row">
              <span class="df-el-name">{{ el }}</span>
              <div class="df-el-track">
                <div
                  class="df-el-fill"
                  :style="{
                    width: elPct(el) + '%',
                    background: clr,
                  }"
                ></div>
              </div>
              <span class="df-el-num">{{ elPct(el) }}%</span>
            </div>
          </div>
        </div>
      </div>
    </div>
    <!-- /elements tab -->
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
  background: var(--surface-1);
  border: 1px solid var(--line-subtle);
  border-radius: 8px;
  color: var(--text-muted);
  font-size: var(--fs-xs);
  font-weight: 600;
  letter-spacing: 1px;
  cursor: pointer;
  white-space: nowrap;
  transition:
    color 0.2s ease,
    border-color 0.2s ease,
    background 0.2s ease;
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
  /* v-show 从 display:none 切回时重新播放一次淡入 */
  animation: df-tab-in 240ms ease;
}

@keyframes df-tab-in {
  from {
    opacity: 0;
    transform: translateY(6px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@media (prefers-reduced-motion: reduce) {
  .df-tab-content {
    animation: none;
    transition: none;
  }
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

.df-pillar-val {
  font-family: var(--font-serif);
  font-size: var(--fs-hero);
  font-weight: 950;
  color: var(--accent);
  letter-spacing: 0.05em;
  line-height: 1;
}

.df-sx {
  font-size: var(--fs-xs);
  color: var(--text-soft);
  margin: 0;
  text-align: right;
}

/* Hours + Elements */
.df-bottom-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.5rem;
}

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
  background: var(--surface-2);
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
  color: var(--crimson);
}

/* Structural analysis */
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
  background: var(--surface-2);
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

/* ═══ 日课推算样式 ═══ */

/* 十神 + 长生核心区 */
.df-rikuyo-core {
  padding: 1rem;
}

.rikuyo-core-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 0.75rem;
  margin-top: 0.65rem;
}

.rikuyo-core-item {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  padding: 0.75rem;
  background: var(--surface-2);
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

.rikuyo-core-desc {
  font-size: var(--fs-xs);
  color: var(--text-muted);
  line-height: 1.5;
}

.section-boundary-note {
  grid-column: 1 / -1;
  margin: 0;
  padding-top: 0.7rem;
  border-top: 1px solid var(--line-subtle);
  color: var(--text-soft);
  font-size: var(--fs-2xs);
  line-height: 1.6;
}

.rikuyo-evidence-meta {
  font-size: var(--fs-xs);
  color: var(--text-soft);
  padding: 0.3rem 0.5rem;
  background: var(--surface-2);
  border-radius: 4px;
  border: 1px solid var(--line-subtle);
  margin-top: 0.25rem;
}

.rikuyo-evidence-value {
  color: var(--text);
  text-shadow: none;
}

/* 月令状态 */
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
  background: var(--accent-dim);
  color: var(--accent);
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
  background: var(--surface-2);
  transition: all 0.25s;
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
  color: var(--text-muted);
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
  background: var(--surface-2);
  border: 1px solid var(--line-subtle);
}

.rel-type-tag {
  font-size: var(--fs-2xs);
  font-weight: 700;
  padding: 0.1rem 0.4rem;
  border-radius: 3px;
  flex-shrink: 0;
  letter-spacing: 0.5px;
  background: var(--accent-dim);
  color: var(--accent);
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
  background: var(--surface-2);
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
  background: rgba(var(--jade-accent-rgb), 0.1);
  color: rgba(var(--jade-accent-rgb), 1);
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
  background: var(--surface-2);
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
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
  font-size: var(--fs-2xs);
  color: var(--text-soft);
  margin: 0.2rem 0 0;
}

.yun-inter-layer-relations {
  display: grid;
  gap: 0.35rem;
  margin-top: 0.75rem;
  padding-top: 0.7rem;
  border-top: 1px solid var(--line-subtle);
}

.yun-inter-layer-relations strong {
  color: var(--text);
  font-size: var(--fs-xs);
}

.yun-inter-layer-relations span {
  color: var(--text-muted);
  font-size: var(--fs-2xs);
  line-height: 1.5;
}

.yun-inter-layer-relations small {
  color: var(--text-soft);
  font-size: var(--fs-2xs);
}

.yun-desc {
  font-size: var(--fs-xs);
  color: var(--text-muted);
  line-height: 1.5;
  margin: 0.2rem 0 0;
}

@media (max-width: 720px) {
  .df-header {
    align-items: flex-start;
    gap: 1rem;
  }

  .df-pillar-val {
    font-size: var(--fs-stat-lg);
  }

  .df-bottom-row,
  .rikuyo-core-grid,
  .yun-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 480px) {
  .df-header {
    flex-direction: column;
  }

  .df-pillar-col {
    align-items: flex-start;
  }

  .almanac-grid {
    grid-template-columns: 1fr;
  }
}
</style>
