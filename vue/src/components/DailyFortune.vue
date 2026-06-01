<script setup lang="ts">
import { ref, computed } from 'vue'

interface ElementImage { element: string; image_url: string; description: string }
interface ShengKeAnalysis { day_stem_relation?: string; day_branch_relation?: string; summary?: string }
interface HiddenStemGod { stem: string; type: string; element: string; ten_god: string; favorable: boolean }
interface StemRelation { type: string; target: string; detail: string; is_favorable: boolean; note?: string }
interface BranchRelation { type: string; target: string; detail: string; is_favorable: boolean }
interface ShenShaActivation { name: string; type: string; description: string; activation: string }
interface DaYunInfluence { current_pillar: string; start_age: number; end_age: number; ten_god: string; favorable: boolean; relation: string; score: number; description: string }
interface LiuNianInfluence { year_pillar: string; ten_god: string; favorable: boolean; relation: string; tai_sui_relation: string; score: number; description: string }
interface AdvanceRetreat { phase: string; phase_desc: string; element: string; score: number; description: string }
interface YongShenImpact { tiao_hou_element: string; tiao_hou_hit: boolean; tong_guan_element: string; tong_guan_hit: boolean; fu_yi_elements: string[]; fu_yi_hit: boolean; score: number; description: string }

interface Props {
  solarDate: string; dayGanZhi: string; weekDay?: string; lunarDate?: string
  shengXiao?: string; yiJi?: string; chongSha?: string; elementImages?: ElementImage[]
  luckyColor?: string; luckyNumber?: number; wealthDir?: string
  auspiciousHours?: string[]; todayElements?: Record<string, number>
  tiaoHou?: string
  // 黄历字段
  jiShen?: string; xiongShen?: string; taiShen?: string
  pengZu?: string; gua?: string; jieQi?: string
  // 分析字段
  shengKeAnalysis?: ShengKeAnalysis; flowImpact?: string; seasonElementAdvice?: string
  // 日课推算
  todayTenGod?: string; tenGodFavorable?: boolean; tenGodDesc?: string
  twelveStage?: string; stageFavorable?: boolean; stageDesc?: string; stageFlexible?: string
  hiddenStems?: HiddenStemGod[]
  stemRelations?: StemRelation[]; branchRelations?: BranchRelation[]
  activatedShenSha?: ShenShaActivation[]
  dayunInfluence?: DaYunInfluence; liunianInfluence?: LiuNianInfluence
  advanceRetreat?: AdvanceRetreat; yongshenImpact?: YongShenImpact
  overallVerdict?: string; favorScore?: number
  patternName?: string; patternType?: string
  patternFavorable?: string[]; patternUnfavorable?: string[]
}
const props = withDefaults(defineProps<Props>(), {
  weekDay: '', lunarDate: '', shengXiao: '', yiJi: '',
  chongSha: '', elementImages: () => [],
  luckyColor: '', luckyNumber: 0, wealthDir: '', auspiciousHours: () => [], todayElements: () => ({}),
  tiaoHou: '',
  jiShen: '', xiongShen: '', taiShen: '',
  pengZu: '', gua: '', jieQi: '',
  shengKeAnalysis: undefined, flowImpact: '', seasonElementAdvice: '',
  // 日课推算
  todayTenGod: '', tenGodFavorable: false, tenGodDesc: '',
  twelveStage: '', stageFavorable: false, stageDesc: '', stageFlexible: '',
  hiddenStems: () => [],
  stemRelations: () => [], branchRelations: () => [],
  activatedShenSha: () => [],
  dayunInfluence: undefined, liunianInfluence: undefined,
  advanceRetreat: undefined, yongshenImpact: undefined,
  overallVerdict: '', favorScore: 0,
  patternName: '', patternType: '',
  patternFavorable: () => [], patternUnfavorable: () => [],
})
const showAiModal = ref(false)
const activeTab = ref('overview')
const dfTabs = [
  { key: 'overview', label: '今日概览' },
  { key: 'almanac', label: '黄历' },
  { key: 'analysis', label: '运势分析' },
  { key: 'rikuyo', label: '日课推算' },
  { key: 'elements', label: '五行吉时' },
]
const elementEntries = [['金','#FFD700'],['木','#3CB371'],['水','#4169E1'],['火','#DC143C'],['土','#DAA520']] as [string,string][]
function elPct(el: string) {
  const n = props.todayElements || {}, t = Object.values(n).reduce((s,v) => s + v, 0)
  return t ? Math.round(((n[el]||0)/t)*100) : 0
}
const yiItems = computed(() => {
  if (!props.yiJi) return []
  const m = props.yiJi.match(/宜[:：]?\s*(.+?)(?:忌|$)/)
  return m ? m[1].split(/[、，,]/).filter(Boolean).map(s => s.trim()) : []
})
const jiItems = computed(() => {
  if (!props.yiJi) return []
  const m = props.yiJi.match(/忌[:：]?\s*(.+)/)
  return m ? m[1].split(/[、，,]/).filter(Boolean).map(s => s.trim()) : []
})
</script>

<template>
  <div class="daily-fortune">

    <!-- Date + Pillar -->
    <div class="df-header glass-card">
      <div class="df-date-col">
        <p class="df-solar">{{ solarDate }}<span v-if="weekDay" class="df-weekday">{{ weekDay }}</span></p>
        <p v-if="lunarDate" class="df-lunar">{{ lunarDate }}</p>
      </div>
      <div class="df-pillar-col">
        <div class="df-pillar-glow"></div>
        <span class="df-pillar-val">{{ dayGanZhi }}</span>
        <span v-if="shengXiao" class="df-sx">属{{ shengXiao }}</span>
      </div>
    </div>

    <!-- Tab navigation -->
    <div class="df-tabs">
      <button
        v-for="tab in dfTabs"
        :key="tab.key"
        class="df-tab-btn"
        :class="{ active: activeTab === tab.key }"
        @click="activeTab = tab.key"
      >{{ tab.label }}</button>
    </div>

    <!-- ═══ Tab: 今日概览 ═══ -->
    <div v-show="activeTab === 'overview'" class="df-tab-content">
    <!-- Lucky 4-grid -->
    <div class="df-lucky-grid">
      <div class="df-lucky-cell glass-card">
        <div class="lc-icon">
          <div class="lc-color-dot" :style="{ background: luckyColor||'#C41E3A', boxShadow: `0 0 18px ${luckyColor||'#C41E3A'}88` }"></div>
        </div>
        <span class="lc-lbl">幸运色</span>
        <span class="lc-val">{{ luckyColor || '—' }}</span>
      </div>
      <div class="df-lucky-cell glass-card">
        <div class="lc-icon">
          <svg width="32" height="32" viewBox="0 0 32 32" fill="none">
            <rect x="3" y="3" width="26" height="26" rx="5" stroke="#D4A84B" stroke-width="1.5" opacity="0.35" />
            <text x="16" y="22" text-anchor="middle" font-size="13" font-weight="900" fill="#D4A84B" opacity="0.7">{{ luckyNumber || '?' }}</text>
          </svg>
        </div>
        <span class="lc-lbl">幸运数字</span>
        <span class="lc-val lc-val-gold">{{ luckyNumber || '—' }}</span>
      </div>
      <div class="df-lucky-cell glass-card">
        <div class="lc-icon">
          <svg width="32" height="32" viewBox="0 0 32 32" fill="none">
            <circle cx="16" cy="16" r="13" stroke="#D4A84B" stroke-width="1.2" opacity="0.3" />
            <line x1="16" y1="3" x2="16" y2="16" stroke="#D4A84B" stroke-width="2" stroke-linecap="round" />
            <line x1="16" y1="16" x2="24" y2="22" stroke="#D4A84B" stroke-width="2" stroke-linecap="round" />
          </svg>
        </div>
        <span class="lc-lbl">财神方位</span>
        <span class="lc-val">{{ wealthDir || '—' }}</span>
      </div>
      <div class="df-lucky-cell glass-card">
        <div class="lc-icon">
          <svg width="32" height="32" viewBox="0 0 32 32" fill="none">
            <path d="M16 2L21 12H30L23 18L26 28L16 22L6 28L9 18L2 12H11L16 2Z" stroke="#C41E3A" stroke-width="1.3" opacity="0.45" />
          </svg>
        </div>
        <span class="lc-lbl">冲煞</span>
        <span class="lc-val lc-val-red">{{ chongSha || '—' }}</span>
      </div>
    </div>

    <!-- Yi Ji -->
    <div class="df-yiji glass-card">
      <div class="yj-col yj-yi">
        <div class="yj-header">
          <span class="yj-tag yj-tag-yi">宜</span>
        </div>
        <div v-if="yiItems.length" class="yj-tags">
          <span v-for="it in yiItems" :key="it" class="yj-tag-item yj-tag-yi-item">{{ it }}</span>
        </div>
        <p v-else class="yj-empty">—</p>
      </div>
      <div class="yj-divider"></div>
      <div class="yj-col yj-ji">
        <div class="yj-header">
          <span class="yj-tag yj-tag-ji">忌</span>
        </div>
        <div v-if="jiItems.length" class="yj-tags">
          <span v-for="it in jiItems" :key="it" class="yj-tag-item yj-tag-ji-item">{{ it }}</span>
        </div>
        <p v-else class="yj-empty">—</p>
      </div>
    </div>

    <!-- TiaoHou -->
    <div v-if="tiaoHou" class="df-tiaohou glass-card">
      <div class="df-sec-header">
        <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
          <path d="M7 1L9 5H13L9.5 7.5L11 12L7 9.5L3 12L4.5 7.5L1 5H5L7 1Z" stroke="#D4A84B" stroke-width="1" opacity="0.5"/>
        </svg>
        <span class="df-sec-title">调候吉言</span>
      </div>
      <p class="tiaohou-text">{{ tiaoHou }}</p>
    </div>
    </div><!-- /overview tab -->

    <!-- ═══ Tab: 黄历 ═══ -->
    <div v-show="activeTab === 'almanac'" class="df-tab-content">
    <!-- 黄历信息 -->
    <div v-if="jiShen || xiongShen || taiShen || pengZu || gua || jieQi" class="df-almanac glass-card">
      <div class="df-sec-header">
        <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
          <rect x="2" y="3" width="10" height="9" rx="1.5" stroke="#D4A84B" stroke-width="1" opacity="0.4"/>
          <line x1="2" y1="6" x2="12" y2="6" stroke="#D4A84B" stroke-width="0.8" opacity="0.3"/>
          <line x1="5" y1="1" x2="5" y2="4" stroke="#D4A84B" stroke-width="1" stroke-linecap="round" opacity="0.4"/>
          <line x1="9" y1="1" x2="9" y2="4" stroke="#D4A84B" stroke-width="1" stroke-linecap="round" opacity="0.4"/>
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
    </div><!-- /almanac tab -->

    <!-- ═══ Tab: 运势分析 ═══ -->
    <div v-show="activeTab === 'analysis'" class="df-tab-content">
    <!-- 生克分析 -->
    <div v-if="shengKeAnalysis?.summary || flowImpact || seasonElementAdvice" class="df-analysis glass-card">
      <div class="df-sec-header">
        <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
          <circle cx="7" cy="7" r="5.5" stroke="#D4A84B" stroke-width="0.8" opacity="0.3"/>
          <path d="M5 7h4M7 5v4" stroke="#D4A84B" stroke-width="1" stroke-linecap="round" opacity="0.5"/>
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
    </div><!-- /analysis tab -->

    <!-- ═══ Tab: 日课推算 ═══ -->
    <div v-show="activeTab === 'rikuyo'" class="df-tab-content">
    <!-- ═══ 日课推算 ═══ -->

    <!-- 综合断语 + 评分 -->
    <div v-if="overallVerdict" class="df-rikuyo-verdict glass-card">
      <div class="df-sec-header">
        <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
          <path d="M7 1l1.8 3.6L13 5.3l-3 2.9.7 4.1L7 10.5l-3.7 1.8.7-4.1-3-2.9 4.2-.7L7 1z" stroke="#D4A84B" stroke-width="1" opacity="0.5"/>
        </svg>
        <span class="df-sec-title">日课推算</span>
        <span v-if="favorScore" class="rikuyo-score-badge" :class="{ 'score-good': favorScore >= 60, 'score-mid': favorScore >= 40 && favorScore < 60, 'score-bad': favorScore < 40 }">{{ favorScore }}分</span>
      </div>
      <p class="rikuyo-verdict-text">{{ overallVerdict }}</p>
    </div>

    <!-- 格局信息（特殊格局时显示） -->
    <div v-if="patternName && patternType === '特殊格局'" class="df-pattern-info glass-card">
      <div class="df-sec-header">
        <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
          <circle cx="7" cy="7" r="5.5" stroke="#E8B84B" stroke-width="1" opacity="0.5"/>
          <path d="M7 4v3l2 1" stroke="#E8B84B" stroke-width="1" stroke-linecap="round"/>
        </svg>
        <span class="df-sec-title">格局喜忌</span>
        <span class="pattern-badge">{{ patternName }}</span>
      </div>
      <div class="pattern-elements">
        <span v-if="patternFavorable?.length" class="pattern-tag pattern-like">喜{{ patternFavorable.join('') }}</span>
        <span v-if="patternUnfavorable?.length" class="pattern-tag pattern-dislike">忌{{ patternUnfavorable.join('') }}</span>
      </div>
    </div>

    <!-- 今日十神 + 十二长生 -->
    <div v-if="todayTenGod || twelveStage" class="df-rikuyo-core glass-card">
      <div class="rikuyo-core-grid">
        <!-- 十神 -->
        <div v-if="todayTenGod" class="rikuyo-core-item">
          <span class="rikuyo-core-label">今日十神</span>
          <span class="rikuyo-core-value" :class="{ 'val-fav': tenGodFavorable, 'val-dis': !tenGodFavorable }">{{ todayTenGod }}</span>
          <span v-if="tenGodDesc" class="rikuyo-core-desc">{{ tenGodDesc }}</span>
        </div>
        <!-- 十二长生 -->
        <div v-if="twelveStage" class="rikuyo-core-item">
          <span class="rikuyo-core-label">十二长生</span>
          <span class="rikuyo-core-value" :class="{ 'val-fav': stageFavorable, 'val-dis': !stageFavorable }">{{ twelveStage }}</span>
          <span v-if="stageDesc" class="rikuyo-core-desc">{{ stageDesc }}</span>
          <span v-if="stageFlexible" class="rikuyo-flexible">{{ stageFlexible }}</span>
        </div>
      </div>
    </div>

    <!-- 进退气 -->
    <div v-if="advanceRetreat" class="df-rikuyo-advance glass-card">
      <div class="df-sec-header">
        <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
          <path d="M2 12L7 2L12 12" stroke="#D4A84B" stroke-width="1" stroke-linecap="round" opacity="0.4"/>
          <line x1="4" y1="8" x2="10" y2="8" stroke="#D4A84B" stroke-width="0.8" opacity="0.3"/>
        </svg>
        <span class="df-sec-title">进退气</span>
        <span class="rikuyo-phase-tag" :class="{ 'phase-adv': advanceRetreat.phase === '进气', 'phase-peak': advanceRetreat.phase === '当令', 'phase-ret': advanceRetreat.phase === '退气', 'phase-dead': advanceRetreat.phase === '无气' || advanceRetreat.phase === '死' }">{{ advanceRetreat.phase }}</span>
      </div>
      <p class="rikuyo-advance-text">{{ advanceRetreat.description }}</p>
    </div>

    <!-- 藏干分析 -->
    <div v-if="hiddenStems?.length" class="df-rikuyo-hidden glass-card">
      <div class="df-sec-header">
        <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
          <rect x="2" y="2" width="10" height="10" rx="2" stroke="#D4A84B" stroke-width="0.8" opacity="0.3"/>
          <rect x="4.5" y="4.5" width="5" height="5" rx="1" stroke="#D4A84B" stroke-width="0.6" opacity="0.2"/>
        </svg>
        <span class="df-sec-title">地支藏干</span>
      </div>
      <div class="hidden-stems-grid">
        <div v-for="hs in hiddenStems" :key="hs.stem + hs.type" class="hidden-stem-card" :class="{ 'hs-fav': hs.favorable, 'hs-dis': !hs.favorable }">
          <span class="hs-stem">{{ hs.stem }}</span>
          <span class="hs-type">{{ hs.type }}</span>
          <span class="hs-god">{{ hs.ten_god }}</span>
          <span class="hs-elem">{{ hs.element }}</span>
        </div>
      </div>
    </div>

    <!-- 干支关系 -->
    <div v-if="stemRelations?.length || branchRelations?.length" class="df-rikuyo-relations glass-card">
      <div class="df-sec-header">
        <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
          <circle cx="4" cy="7" r="2.5" stroke="#D4A84B" stroke-width="0.8" opacity="0.3"/>
          <circle cx="10" cy="7" r="2.5" stroke="#D4A84B" stroke-width="0.8" opacity="0.3"/>
          <line x1="6.5" y1="7" x2="7.5" y2="7" stroke="#D4A84B" stroke-width="1" opacity="0.4"/>
        </svg>
        <span class="df-sec-title">干支关系</span>
      </div>
      <div class="relations-list">
        <div v-for="(sr, i) in stemRelations" :key="'sr'+i" class="relation-item" :class="{ 'rel-fav': sr.is_favorable, 'rel-dis': !sr.is_favorable }">
          <span class="rel-type-tag">{{ sr.type }}</span>
          <span class="rel-detail">{{ sr.detail }}</span>
          <span v-if="sr.note" class="rel-note">{{ sr.note }}</span>
        </div>
        <div v-for="(br, i) in branchRelations" :key="'br'+i" class="relation-item" :class="{ 'rel-fav': br.is_favorable, 'rel-dis': !br.is_favorable }">
          <span class="rel-type-tag">{{ br.type }}</span>
          <span class="rel-detail">{{ br.detail }}</span>
        </div>
      </div>
    </div>

    <!-- 神煞引动 -->
    <div v-if="activatedShenSha?.length" class="df-rikuyo-shensha glass-card">
      <div class="df-sec-header">
        <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
          <path d="M7 1l1.5 3 3.5.5-2.5 2.5.6 3.5L7 9l-3.1 1.5.6-3.5L2 4.5l3.5-.5L7 1z" stroke="#D4A84B" stroke-width="0.8" opacity="0.4"/>
        </svg>
        <span class="df-sec-title">神煞引动</span>
      </div>
      <div class="shensha-list">
        <div v-for="ss in activatedShenSha" :key="ss.name" class="shensha-item" :class="{ 'ss-ji': ss.type === '吉神', 'ss-xiong': ss.type !== '吉神' }">
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
          <path d="M1 7h12M7 1v12" stroke="#D4A84B" stroke-width="0.8" opacity="0.3"/>
          <circle cx="7" cy="7" r="3" stroke="#D4A84B" stroke-width="0.8" opacity="0.25"/>
        </svg>
        <span class="df-sec-title">大运流年</span>
      </div>
      <div class="yun-grid">
        <div v-if="dayunInfluence" class="yun-item">
          <span class="yun-label">当前大运</span>
          <span class="yun-pillar">{{ dayunInfluence.current_pillar }}</span>
          <span class="yun-god" :class="{ 'val-fav': dayunInfluence.favorable, 'val-dis': !dayunInfluence.favorable }">{{ dayunInfluence.ten_god }}</span>
          <span class="yun-age">{{ dayunInfluence.start_age }}-{{ dayunInfluence.end_age }}岁</span>
          <p class="yun-desc">{{ dayunInfluence.description }}</p>
        </div>
        <div v-if="liunianInfluence" class="yun-item">
          <span class="yun-label">流年</span>
          <span class="yun-pillar">{{ liunianInfluence.year_pillar }}</span>
          <span class="yun-god" :class="{ 'val-fav': liunianInfluence.favorable, 'val-dis': !liunianInfluence.favorable }">{{ liunianInfluence.ten_god }}</span>
          <p v-if="liunianInfluence.tai_sui_relation" class="yun-taisui">{{ liunianInfluence.tai_sui_relation }}</p>
          <p class="yun-desc">{{ liunianInfluence.description }}</p>
        </div>
      </div>
    </div>

    <!-- 用神影响 -->
    <div v-if="yongshenImpact" class="df-rikuyo-yongshen glass-card">
      <div class="df-sec-header">
        <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
          <circle cx="7" cy="7" r="6" stroke="#D4A84B" stroke-width="0.6" opacity="0.2"/>
          <circle cx="7" cy="7" r="3" stroke="#D4A84B" stroke-width="0.8" opacity="0.35"/>
          <circle cx="7" cy="7" r="1" fill="#D4A84B" opacity="0.5"/>
        </svg>
        <span class="df-sec-title">用神影响</span>
      </div>
      <div class="yongshen-items">
        <div v-if="yongshenImpact.tiao_hou_element" class="yongshen-item" :class="{ 'ys-hit': yongshenImpact.tiao_hou_hit }">
          <span class="ys-label">调候用神</span>
          <span class="ys-elem">{{ yongshenImpact.tiao_hou_element }}</span>
          <span class="ys-status">{{ yongshenImpact.tiao_hou_hit ? '得力' : '未触' }}</span>
        </div>
        <div v-if="yongshenImpact.tong_guan_element" class="yongshen-item" :class="{ 'ys-hit': yongshenImpact.tong_guan_hit }">
          <span class="ys-label">通关用神</span>
          <span class="ys-elem">{{ yongshenImpact.tong_guan_element }}</span>
          <span class="ys-status">{{ yongshenImpact.tong_guan_hit ? '得力' : '未触' }}</span>
        </div>
        <div v-if="yongshenImpact.fu_yi_elements?.length" class="yongshen-item" :class="{ 'ys-hit': yongshenImpact.fu_yi_hit }">
          <span class="ys-label">扶抑喜用</span>
          <span class="ys-elem">{{ yongshenImpact.fu_yi_elements.join(' ') }}</span>
          <span class="ys-status">{{ yongshenImpact.fu_yi_hit ? '得力' : '未触' }}</span>
        </div>
      </div>
      <p v-if="yongshenImpact.description" class="yongshen-desc">{{ yongshenImpact.description }}</p>
    </div>
    </div><!-- /rikuyo tab -->

    <!-- ═══ Tab: 五行吉时 ═══ -->
    <div v-show="activeTab === 'elements'" class="df-tab-content">
    <!-- Hours + Elements -->
    <div class="df-bottom-row">
      <div v-if="auspiciousHours.length" class="df-hours glass-card">
        <div class="df-sec-header">
          <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
            <circle cx="7" cy="7" r="6" stroke="#D4A84B" stroke-width="1" opacity="0.4"/>
            <line x1="7" y1="3" x2="7" y2="7" stroke="#D4A84B" stroke-width="1.5" stroke-linecap="round"/>
            <line x1="7" y1="7" x2="10" y2="9" stroke="#D4A84B" stroke-width="1.5" stroke-linecap="round"/>
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
            <circle cx="7" cy="7" r="6" stroke="#D4A84B" stroke-width="1" opacity="0.25"/>
            <circle cx="7" cy="7" r="2.5" fill="#D4A84B" opacity="0.35"/>
          </svg>
          <span class="df-sec-title">今日五行</span>
        </div>
        <div class="df-el-bars">
          <div v-for="[el, clr] in elementEntries" :key="el" class="df-el-row">
            <span class="df-el-name">{{ el }}</span>
            <div class="df-el-track">
              <div class="df-el-fill" :style="{ width: elPct(el)+'%', background: clr, boxShadow: `0 0 8px ${clr}66` }"></div>
            </div>
            <span class="df-el-num">{{ todayElements[el] ?? 0 }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- AI button -->
    <button class="df-ai-btn" @click="showAiModal = true">
      <span class="df-ai-btn-icon">◈</span>
      AI 深度解析
    </button>
    </div><!-- /elements tab -->

    <!-- Modal -->
    <Teleport to="body">
      <Transition name="df-modal">
        <div v-if="showAiModal" class="df-modal-overlay" @click.self="showAiModal=false">
          <div class="df-modal-box glass-panel">
            <div class="df-modal-hdr">
              <div class="df-modal-title-group">
                <span class="df-modal-orb">☯</span>
                <h2>AI 深度解析</h2>
              </div>
              <button class="df-modal-close" @click="showAiModal=false">✕</button>
            </div>
            <div class="df-modal-body">
              <div class="df-ai-coming">
                <svg width="90" height="90" viewBox="0 0 90 90" fill="none" class="df-ai-svg">
                  <circle cx="45" cy="45" r="42" stroke="#D4A84B" stroke-width="0.6" stroke-dasharray="2 4" opacity="0.2" />
                  <circle cx="45" cy="45" r="28" stroke="#D4A84B" stroke-width="0.6" stroke-dasharray="1 5" opacity="0.15" />
                  <circle cx="45" cy="45" r="8" fill="#D4A84B" opacity="0.2" />
                  <circle cx="45" cy="45" r="13" fill="none" stroke="#C41E3A" stroke-width="0.5" opacity="0.3" />
                  <circle cx="22" cy="24" r="2.5" fill="#D4A84B" opacity="0.45" class="df-star-pulse" style="animation-delay:0s" />
                  <circle cx="68" cy="22" r="2" fill="#D4A84B" opacity="0.35" class="df-star-pulse" style="animation-delay:0.6s" />
                  <circle cx="70" cy="66" r="2.5" fill="#D4A84B" opacity="0.4" class="df-star-pulse" style="animation-delay:1.2s" />
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
.daily-fortune { display: flex; flex-direction: column; gap: 0.75rem; }

/* Tab navigation */
.df-tabs {
  display: flex; gap: 0.25rem; padding: 0;
  overflow-x: auto; scrollbar-width: none; margin-bottom: 0.75rem;
}
.df-tabs::-webkit-scrollbar { display: none; }
.df-tab-btn {
  padding: 0.6rem 1rem; flex-shrink: 0;
  background: rgba(255,255,255,0.02);
  border: 1px solid rgba(212,168,75,0.08); border-radius: 8px;
  color: rgba(255,255,255,0.4); font-size: 0.72rem; font-weight: 600;
  letter-spacing: 1px; cursor: pointer; white-space: nowrap; transition: all 0.3s;
}
.df-tab-btn:hover { color: rgba(212,168,75,0.7); border-color: rgba(212,168,75,0.2); }
.df-tab-btn.active { color: #D4A84B; border-color: rgba(212,168,75,0.3); background: rgba(212,168,75,0.06); }
.df-tab-content { display: flex; flex-direction: column; gap: 0.75rem; }

/* Header */
.df-header {
  display: flex; justify-content: space-between; align-items: center;
  padding: 1.25rem 1.5rem;
  position: relative; overflow: hidden;
}
.df-header::after {
  content: ''; position: absolute; top: -20px; right: -20px;
  width: 100px; height: 100px;
  background: radial-gradient(circle, rgba(196,30,58,0.07), transparent 70%);
  pointer-events: none;
}
.df-date-col { display: flex; flex-direction: column; gap: 0.15rem; }
.df-solar { font-size: 1.05rem; font-weight: 700; color: rgba(255,255,255,0.9); margin: 0; letter-spacing: 1px; }
.df-weekday { font-size: 0.72rem; font-weight: 400; color: rgba(255,255,255,0.35); margin-left: 0.5rem; }
.df-lunar { font-size: 0.72rem; color: rgba(255,255,255,0.28); margin: 0; }
.df-pillar-col { display: flex; flex-direction: column; align-items: flex-end; position: relative; gap: 0.2rem; }
.df-pillar-glow { position: absolute; top: -15px; right: -15px; width: 80px; height: 80px; background: radial-gradient(circle, rgba(196,30,58,0.08), transparent 70%); pointer-events: none; }
.df-pillar-val {
  font-size: 2.75rem; font-weight: 950;
  color: #C41E3A; letter-spacing: 0.05em; line-height: 1;
  text-shadow: 0 0 30px rgba(196,30,58,0.4);
}
.df-sx { font-size: 0.7rem; color: rgba(255,255,255,0.25); margin: 0; text-align: right; }

/* Lucky grid */
.df-lucky-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 0.5rem; }
.df-lucky-cell {
  display: flex; flex-direction: column; align-items: center; gap: 0.35rem;
  padding: 1rem 0.5rem;
  transition: border-color 0.3s, box-shadow 0.3s;
}
.df-lucky-cell:hover { border-color: rgba(212,168,75,0.2); box-shadow: 0 4px 20px rgba(0,0,0,0.2); }
.lc-icon { display: flex; align-items: center; justify-content: center; height: 40px; }
.lc-color-dot { width: 30px; height: 30px; border-radius: 50%; border: 1.5px solid rgba(255,255,255,0.12); }
.lc-lbl { font-size: 0.58rem; color: rgba(255,255,255,0.22); text-transform: uppercase; letter-spacing: 0.1em; }
.lc-val { font-size: 0.85rem; font-weight: 700; color: rgba(255,255,255,0.75); }
.lc-val-gold { color: #D4A84B; font-size: 1.4rem; font-weight: 900; letter-spacing: 1px; text-shadow: 0 0 20px rgba(212,168,75,0.3); }
.lc-val-red { color: #C41E3A; font-size: 0.78rem; }

/* Yi Ji */
.df-yiji { display: flex; overflow: hidden; }
.yj-col { flex: 1; padding: 0.9rem; }
.yj-header { margin-bottom: 0.5rem; }
.yj-tag {
  display: inline-block; font-size: 0.72rem; font-weight: 800;
  padding: 0.15rem 0.7rem; border-radius: 4px; letter-spacing: 1px;
}
.yj-tag-yi { background: rgba(74,222,128,0.1); color: #4ade80; border: 1px solid rgba(74,222,128,0.2); }
.yj-tag-ji { background: rgba(196,30,58,0.1); color: #C41E3A; border: 1px solid rgba(196,30,58,0.2); }
.yj-tags { display: flex; flex-wrap: wrap; gap: 0.35rem; }
.yj-tag-item {
  display: inline-block; font-size: 0.72rem; padding: 0.2rem 0.5rem;
  border-radius: 4px; transition: all 0.2s; cursor: default;
}
.yj-tag-yi-item { background: rgba(74,222,128,0.04); border: 1px solid rgba(74,222,128,0.1); color: rgba(74,222,128,0.65); }
.yj-tag-yi-item:hover { background: rgba(74,222,128,0.1); color: #4ade80; }
.yj-tag-ji-item { background: rgba(196,30,58,0.04); border: 1px solid rgba(196,30,58,0.1); color: rgba(196,30,58,0.6); }
.yj-tag-ji-item:hover { background: rgba(196,30,58,0.1); color: #C41E3A; }
.yj-divider { width: 1px; background: rgba(212,168,75,0.06); margin: 0.75rem 0; }
.yj-empty { font-size: 0.82rem; color: rgba(139,131,120,0.2); margin: 0.5rem 0; }

/* Hours + Elements */
.df-bottom-row { display: grid; grid-template-columns: 1fr 1fr; gap: 0.5rem; }
.df-hours, .df-elems { padding: 0.9rem; }
.df-sec-header { display: flex; align-items: center; gap: 0.4rem; margin-bottom: 0.6rem; }
.df-sec-title { font-size: 0.72rem; font-weight: 800; color: #D4A84B; margin: 0; letter-spacing: 2px; }
.df-hours-list { display: flex; flex-wrap: wrap; gap: 0.35rem; }
.df-hour-chip {
  display: inline-flex; align-items: center; gap: 5px;
  padding: 0.25rem 0.65rem;
  background: rgba(212,168,75,0.04);
  border: 1px solid rgba(212,168,75,0.1);
  border-radius: 20px; font-size: 0.7rem;
  color: rgba(255,255,255,0.6);
  transition: all 0.25s;
}
.df-hour-chip:hover { background: rgba(212,168,75,0.1); border-color: rgba(212,168,75,0.25); color: #D4A84B; }
.df-hour-dot { width: 4px; height: 4px; border-radius: 50%; background: #D4A84B; box-shadow: 0 0 6px rgba(212,168,75,0.7); flex-shrink: 0; }
.df-el-bars { display: flex; flex-direction: column; gap: 0.35rem; }
.df-el-row { display: flex; align-items: center; gap: 0.4rem; }
.df-el-name { width: 14px; font-size: 0.65rem; font-weight: 800; color: rgba(255,255,255,0.3); flex-shrink: 0; }
.df-el-track { flex: 1; height: 5px; background: rgba(255,255,255,0.04); border-radius: 3px; overflow: hidden; }
.df-el-fill { height: 100%; border-radius: 3px; transition: width 0.8s ease; }
.df-el-num { width: 18px; font-size: 0.6rem; color: rgba(255,255,255,0.2); text-align: right; flex-shrink: 0; }

/* AI btn */
.df-ai-btn {
  display: flex; align-items: center; justify-content: center; gap: 0.5rem;
  width: 100%; padding: 0.7rem 1rem;
  background: rgba(255,255,255,0.02);
  color: rgba(255,255,255,0.28);
  border: 1px solid rgba(212,168,75,0.1);
  border-radius: 10px; font-size: 0.8rem; font-weight: 600;
  cursor: pointer; transition: all 0.3s; letter-spacing: 1.5px;
}
.df-ai-btn:hover { border-color: rgba(212,168,75,0.3); color: #D4A84B; background: rgba(212,168,75,0.05); }
.df-ai-btn-icon { font-size: 1rem; }

/* TiaoHou */
.df-tiaohou { padding: 0.9rem; }
.tiaohou-text {
  font-size: 0.82rem;
  color: rgba(255, 255, 255, 0.6);
  line-height: 1.7;
  white-space: pre-wrap;
  margin: 0.4rem 0 0;
  font-style: italic;
}

/* Almanac 黄历 */
.df-almanac { padding: 0.9rem; }
.almanac-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 0.4rem; }
.almanac-item {
  display: flex; flex-direction: column; gap: 0.15rem;
  padding: 0.4rem 0.5rem;
  background: rgba(255,255,255,0.02);
  border-radius: 6px;
  border: 1px solid rgba(255,255,255,0.03);
}
.almanac-full { grid-column: 1 / -1; }
.almanac-label { font-size: 0.58rem; color: rgba(255,255,255,0.2); letter-spacing: 0.5px; }
.almanac-value { font-size: 0.72rem; color: rgba(255,255,255,0.65); line-height: 1.4; }
.almanac-ji { color: #4ade80; }
.almanac-xiong { color: #f08080; }

/* Analysis 运势分析 */
.df-analysis { padding: 0.9rem; }
.analysis-items { display: flex; flex-direction: column; gap: 0.35rem; }
.analysis-item {
  display: flex; gap: 0.5rem; font-size: 0.72rem; line-height: 1.5;
  padding: 0.3rem 0.5rem;
  background: rgba(255,255,255,0.015);
  border-radius: 5px;
}
.analysis-label { min-width: 56px; font-weight: 600; color: rgba(255,255,255,0.3); flex-shrink: 0; }
.analysis-value { color: rgba(255,255,255,0.6); }
.analysis-gold { color: var(--gold); }

/* Modal */
.df-modal-overlay {
  position: fixed; inset: 0; background: rgba(0,0,0,0.7);
  display: flex; align-items: center; justify-content: center;
  z-index: 1000; padding: 1rem; backdrop-filter: blur(8px);
}
.df-modal-box { width: 100%; max-width: 380px; overflow: hidden; }
.df-modal-hdr {
  display: flex; justify-content: space-between; align-items: center;
  padding: 1.25rem 1.5rem; border-bottom: 1px solid rgba(212,168,75,0.07);
}
.df-modal-title-group { display: flex; align-items: center; gap: 0.75rem; }
.df-modal-orb {
  font-size: 1.6rem; color: #D4A84B;
  text-shadow: 0 0 25px rgba(212,168,75,0.5);
  animation: orb-glow 3s ease-in-out infinite;
}
@keyframes orb-glow { 0%,100%{text-shadow:0 0 20px rgba(212,168,75,0.3)} 50%{text-shadow:0 0 40px rgba(212,168,75,0.7)} }
.df-modal-hdr h2 { margin: 0; font-family: var(--font-serif), serif; font-size: 1.05rem; font-weight: 700; color: rgba(255,255,255,0.85); letter-spacing: 2px; }
.df-modal-close { background: none; border: none; font-size: 1.2rem; color: rgba(255,255,255,0.2); cursor: pointer; padding: 0.25rem; transition: color 0.2s; }
.df-modal-close:hover { color: #D4A84B; }
.df-modal-body { padding: 2.5rem 1.5rem; }
.df-ai-coming { display: flex; flex-direction: column; align-items: center; gap: 1rem; text-align: center; }
.df-ai-svg { opacity: 0.75; animation: svg-rot 20s linear infinite; }
@keyframes svg-rot { from{transform:rotate(0deg)} to{transform:rotate(360deg)} }
.df-star-pulse { animation: st-pulse 2.5s ease-in-out infinite; }
@keyframes st-pulse { 0%,100%{opacity:0.2} 50%{opacity:0.7} }
.df-ai-title { font-size: 1rem; font-weight: 700; color: rgba(255,255,255,0.75); margin: 0; }
.df-ai-sub { font-size: 0.78rem; color: rgba(255,255,255,0.25); margin: 0; }

/* Modal transition */
.df-modal-enter-active, .df-modal-leave-active { transition: opacity 0.25s ease; }
.df-modal-enter-from, .df-modal-leave-to { opacity: 0; }
.df-modal-enter-active .df-modal-box, .df-modal-leave-active .df-modal-box { transition: transform 0.25s ease; }
.df-modal-enter-from .df-modal-box { transform: scale(0.9) translateY(12px); }
.df-modal-leave-to .df-modal-box { transform: scale(0.9) translateY(12px); }

/* ═══ 日课推算样式 ═══ */

/* 综合断语 */
.df-rikuyo-verdict { padding: 1rem; position: relative; }
.df-rikuyo-verdict::before {
  content: ''; position: absolute; top: 0; left: 0; right: 0; height: 2px;
  background: linear-gradient(90deg, transparent, rgba(212,168,75,0.5), transparent);
}
.rikuyo-score-badge {
  margin-left: auto; font-size: 0.72rem; font-weight: 800;
  padding: 0.15rem 0.6rem; border-radius: 20px; letter-spacing: 0.5px;
}
.score-good { background: rgba(74,222,128,0.12); color: #4ade80; border: 1px solid rgba(74,222,128,0.25); }
.score-mid { background: rgba(212,168,75,0.1); color: #D4A84B; border: 1px solid rgba(212,168,75,0.25); }
.score-bad { background: rgba(196,30,58,0.1); color: #C41E3A; border: 1px solid rgba(196,30,58,0.25); }
.rikuyo-verdict-text {
  font-size: 0.82rem; color: rgba(255,255,255,0.65); line-height: 1.8;
  margin: 0.5rem 0 0; white-space: pre-wrap;
}

/* 格局信息 */
.df-pattern-info { padding: 0.8rem 1rem; }
.pattern-badge {
  font-size: 0.7rem; color: #E8B84B; background: rgba(232,184,75,0.1);
  padding: 0.15rem 0.5rem; border-radius: 4px; margin-left: auto;
}
.pattern-elements { display: flex; gap: 0.5rem; margin-top: 0.4rem; }
.pattern-tag {
  font-size: 0.75rem; padding: 0.2rem 0.6rem; border-radius: 4px;
  font-weight: 600; letter-spacing: 0.5px;
}
.pattern-like { background: rgba(74,222,128,0.1); color: #4ade80; border: 1px solid rgba(74,222,128,0.2); }
.pattern-dislike { background: rgba(196,30,58,0.1); color: #C41E3A; border: 1px solid rgba(196,30,58,0.2); }

/* 十神 + 长生核心区 */
.df-rikuyo-core { padding: 1rem; }
.rikuyo-core-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 0.75rem; }
.rikuyo-core-item {
  display: flex; flex-direction: column; gap: 0.25rem;
  padding: 0.75rem; background: rgba(255,255,255,0.02);
  border-radius: 8px; border: 1px solid rgba(255,255,255,0.04);
}
.rikuyo-core-label { font-size: 0.58rem; color: rgba(255,255,255,0.2); letter-spacing: 0.5px; text-transform: uppercase; }
.rikuyo-core-value {
  font-family: var(--font-serif); font-size: 1.4rem; font-weight: 900;
  letter-spacing: 2px; line-height: 1.2;
}
.val-fav { color: #4ade80; text-shadow: 0 0 15px rgba(74,222,128,0.3); }
.val-dis { color: #C41E3A; text-shadow: 0 0 15px rgba(196,30,58,0.3); }
.rikuyo-core-desc { font-size: 0.72rem; color: rgba(255,255,255,0.45); line-height: 1.5; }
.rikuyo-flexible {
  font-size: 0.72rem; color: #D4A84B; font-style: italic;
  padding: 0.3rem 0.5rem; background: rgba(212,168,75,0.06);
  border-radius: 4px; border-left: 2px solid rgba(212,168,75,0.3);
  margin-top: 0.25rem;
}

/* 进退气 */
.df-rikuyo-advance { padding: 0.9rem; }
.rikuyo-phase-tag {
  margin-left: auto; font-size: 0.65rem; font-weight: 700;
  padding: 0.1rem 0.5rem; border-radius: 10px; letter-spacing: 0.5px;
}
.phase-adv { background: rgba(74,222,128,0.1); color: #4ade80; }
.phase-peak { background: rgba(212,168,75,0.1); color: #D4A84B; }
.phase-ret { background: rgba(255,165,0,0.1); color: #ffa500; }
.phase-dead { background: rgba(196,30,58,0.1); color: #C41E3A; }
.rikuyo-advance-text { font-size: 0.78rem; color: rgba(255,255,255,0.5); line-height: 1.6; margin: 0.4rem 0 0; }

/* 藏干 */
.df-rikuyo-hidden { padding: 0.9rem; }
.hidden-stems-grid { display: flex; flex-wrap: wrap; gap: 0.5rem; }
.hidden-stem-card {
  display: flex; flex-direction: column; align-items: center; gap: 0.15rem;
  padding: 0.6rem 0.8rem; border-radius: 8px; min-width: 60px;
  border: 1px solid rgba(255,255,255,0.05); background: rgba(255,255,255,0.02);
  transition: all 0.25s;
}
.hs-fav { border-color: rgba(74,222,128,0.15); background: rgba(74,222,128,0.03); }
.hs-dis { border-color: rgba(196,30,58,0.12); background: rgba(196,30,58,0.03); }
.hs-stem { font-family: var(--font-serif); font-size: 1.2rem; font-weight: 900; color: rgba(255,255,255,0.75); }
.hs-type { font-size: 0.55rem; color: rgba(255,255,255,0.2); }
.hs-god { font-size: 0.65rem; font-weight: 700; }
.hs-fav .hs-god { color: #4ade80; }
.hs-dis .hs-god { color: #C41E3A; }
.hs-elem { font-size: 0.55rem; color: rgba(255,255,255,0.25); }

/* 干支关系 */
.df-rikuyo-relations { padding: 0.9rem; }
.relations-list { display: flex; flex-direction: column; gap: 0.35rem; }
.relation-item {
  display: flex; align-items: center; gap: 0.5rem;
  padding: 0.4rem 0.6rem; border-radius: 6px;
  background: rgba(255,255,255,0.015); border: 1px solid rgba(255,255,255,0.03);
}
.rel-fav { border-color: rgba(74,222,128,0.1); }
.rel-dis { border-color: rgba(196,30,58,0.08); }
.rel-type-tag {
  font-size: 0.6rem; font-weight: 700; padding: 0.1rem 0.4rem;
  border-radius: 3px; flex-shrink: 0; letter-spacing: 0.5px;
}
.rel-fav .rel-type-tag { background: rgba(74,222,128,0.08); color: #4ade80; }
.rel-dis .rel-type-tag { background: rgba(196,30,58,0.08); color: #C41E3A; }
.rel-detail { font-size: 0.72rem; color: rgba(255,255,255,0.6); }
.rel-note { font-size: 0.65rem; color: #D4A84B; font-style: italic; margin-left: auto; }

/* 神煞 */
.df-rikuyo-shensha { padding: 0.9rem; }
.shensha-list { display: flex; flex-direction: column; gap: 0.5rem; }
.shensha-item {
  padding: 0.6rem 0.8rem; border-radius: 8px;
  border: 1px solid rgba(255,255,255,0.04); background: rgba(255,255,255,0.015);
}
.ss-ji { border-color: rgba(74,222,128,0.12); }
.ss-xiong { border-color: rgba(196,30,58,0.1); }
.ss-name { font-size: 0.82rem; font-weight: 800; color: rgba(255,255,255,0.8); margin-right: 0.5rem; }
.ss-type-tag {
  font-size: 0.58rem; padding: 0.1rem 0.4rem; border-radius: 3px;
}
.ss-ji .ss-type-tag { background: rgba(74,222,128,0.08); color: #4ade80; }
.ss-xiong .ss-type-tag { background: rgba(196,30,58,0.08); color: #C41E3A; }
.ss-desc { font-size: 0.72rem; color: rgba(255,255,255,0.45); line-height: 1.5; margin: 0.3rem 0 0; }
.ss-activation { font-size: 0.65rem; color: rgba(255,255,255,0.25); margin: 0.2rem 0 0; font-style: italic; }

/* 大运流年 */
.df-rikuyo-yun { padding: 0.9rem; }
.yun-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 0.75rem; }
.yun-item {
  display: flex; flex-direction: column; gap: 0.2rem;
  padding: 0.7rem; background: rgba(255,255,255,0.02);
  border-radius: 8px; border: 1px solid rgba(255,255,255,0.04);
}
.yun-label { font-size: 0.58rem; color: rgba(255,255,255,0.2); text-transform: uppercase; letter-spacing: 0.5px; }
.yun-pillar { font-family: var(--font-serif); font-size: 1.3rem; font-weight: 900; color: rgba(255,255,255,0.8); letter-spacing: 2px; }
.yun-god { font-size: 0.75rem; font-weight: 700; }
.yun-age { font-size: 0.6rem; color: rgba(255,255,255,0.25); }
.yun-taisui { font-size: 0.65rem; color: #C41E3A; margin: 0.2rem 0 0; font-style: italic; }
.yun-desc { font-size: 0.7rem; color: rgba(255,255,255,0.4); line-height: 1.5; margin: 0.2rem 0 0; }

/* 用神影响 */
.df-rikuyo-yongshen { padding: 0.9rem; }
.yongshen-items { display: flex; flex-direction: column; gap: 0.4rem; }
.yongshen-item {
  display: flex; align-items: center; gap: 0.5rem;
  padding: 0.5rem 0.7rem; border-radius: 6px;
  background: rgba(255,255,255,0.015); border: 1px solid rgba(255,255,255,0.03);
  transition: all 0.25s;
}
.ys-hit { border-color: rgba(74,222,128,0.15); background: rgba(74,222,128,0.03); }
.ys-label { font-size: 0.6rem; color: rgba(255,255,255,0.25); min-width: 56px; flex-shrink: 0; }
.ys-elem { font-size: 0.82rem; font-weight: 700; color: rgba(255,255,255,0.7); }
.ys-status { font-size: 0.6rem; margin-left: auto; padding: 0.1rem 0.4rem; border-radius: 3px; }
.ys-hit .ys-status { background: rgba(74,222,128,0.1); color: #4ade80; }
.yongshen-desc { font-size: 0.72rem; color: rgba(255,255,255,0.4); margin: 0.5rem 0 0; line-height: 1.5; }
</style>