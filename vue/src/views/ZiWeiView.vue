<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { fetchChart, type ChartDetail } from '../api/chart'
import { getApiErrorMessage } from '../api/client'
import {
  fetchZiWeiChart,
  fetchZiWeiOverlay,
  fetchZiWeiPeriod,
  type ZiWeiDayunStageAnalysis,
  type ZiWeiOverlayAnalysis,
  type ZiWeiPeriodAnalysis,
  type ZiWeiSihuaProjectionItem,
} from '../api/ziwei'
import ZiWeiInterpretation from '../components/ZiWeiInterpretation.vue'
import ZiWeiOverlay from '../components/ZiWeiOverlay.vue'
import ZiWeiPeriodAnalysisPanel from '../components/ZiWeiPeriodAnalysisPanel.vue'

const route = useRoute()

interface BirthInfo {
  name: string
  gender: string
  yearBranch: string
  solarDate: string
  lunarDate: string
  baziChartId: number
}

interface StarInfo {
  name: string
  type: string
  scope: string
  brightness: string
}

interface PalaceData {
  name: string
  branch: string
  heavenly_stem: string
  is_body_palace: boolean
  stars: StarInfo[]
  four_hua: string[]
  adjective_stars?: string[]
  changsheng_12?: string
  boshi_12?: string
  jiang_qian_12?: string
  sui_qian_12?: string
  sanfang_sizheng?: { opposite?: string; trine1?: string; trine2?: string }
}

interface ZiWeiChartData {
  profile_id: string
  engine_version: string
  rule_version: string
  rule_school: string
  rule_sources: Array<{
    rule_id: string
    repository: string
    commit: string
    path: string
    sha256: string
    license: string
    source_tier: string
    validation_status: string
  }>
  plugin_manifest: Array<{ id: string; version: string }>
  plugin_manifest_hash: string
  calculation_input: {
    calendar_type: 'SOLAR'
    year: number
    month: number
    day: number
    hour: number
    minute: number
    gender: '男' | '女'
    basis: 'normalized_solar_minute'
  }
  input_fingerprint: string
  content_hash: string
  palaces: PalaceData[]
  life_master: string
  body_master: string
  five_bureau: string
  earthly_branch_of_soul_palace: string
  earthly_branch_of_body_palace: string
  body_palace: string
  patterns: string[]
}

interface SectionData {
  title: string
  content: string
  tags: string[]
}

interface ReadingEvidence {
  type: string
  label: string
  value: string
  basis: string
}

interface SanfangContext {
  opposite: string
  trine1: string
  trine2: string
  opposite_stars: string[]
  trine1_stars: string[]
  trine2_stars: string[]
  notes: string[]
}

interface PatternDetail {
  name: string
  palace: string
  stars: string[]
  basis: string
  structure_status: string
  validation_status: string
}

interface PalaceReading {
  palaceName: string
  palaceFocus?: string
  summary?: string
  keyPoints?: string[]
  evidence?: ReadingEvidence[]
  sanfangContext?: SanfangContext | null
  patternDetails?: PatternDetail[]
  reviewNotes?: string[]
  limitations?: string[]
  evidenceBasis?: string
  placementBasis?: string
  interpretationBasis?: string
  interpretationStatus?: string
  validationStatus?: string
  isOutcomeConclusion?: boolean
  mainStarAnalysis: SectionData
  auxStarInfluence: SectionData
  sihuaInfluence: SectionData
  sanFangSiZheng: SectionData
  patternAnnotations: SectionData
}

interface LiunianChartData {
  palaces: PalaceData[]
  year: number
  liu_nian_stars?: string[][]
  liu_nian_four_hua?: string[][]
  overlay_analysis?: ZiWeiOverlayAnalysis
}

interface SihuaGroup {
  type: string
  css: string
  items: ZiWeiSihuaProjectionItem[]
}

// State
const loading = ref(true)
const error = ref('')
const activeTab = ref('mingpan')

const birthInfo = ref<BirthInfo>()
const chartData = ref<ZiWeiChartData>()
const selectedPalace = ref<PalaceReading | null>(null)

const liunianData = ref<any[]>([])
const liuyueData = ref<any[]>([])
const liuriData = ref<any[]>([])
const sihuaData = ref<any>({})
const sihuaChainData = ref<any>({})

const dayunAnalysis = ref<ZiWeiPeriodAnalysis | null>(null)
const dayunNominalAge = ref<number | null>(null)
const liunianAnalysis = ref<ZiWeiPeriodAnalysis | null>(null)
const liuyueAnalysis = ref<ZiWeiPeriodAnalysis | null>(null)
const liuriAnalysis = ref<ZiWeiPeriodAnalysis | null>(null)
const dayunStageList = computed<ZiWeiDayunStageAnalysis[]>(
  () => dayunAnalysis.value?.dayun_stages || [],
)

const liunianOverlay = ref<LiunianChartData>()
const availableYears = ref<number[]>([])
const selectedLiunianYear = ref<number>(new Date().getFullYear())
const loadingTab = ref(false)
const tabError = ref('')

function list<T>(items: T[] | null | undefined): T[] {
  return Array.isArray(items) ? items : []
}

function mapBirthInfo(chart: ChartDetail): BirthInfo {
  const month = String(chart.birth_month).padStart(2, '0')
  const day = String(chart.birth_day).padStart(2, '0')
  const hour = String(chart.birth_hour).padStart(2, '0')
  const minute = String(chart.birth_min || 0).padStart(2, '0')
  return {
    name: chart.name || '未命名',
    gender: chart.gender,
    yearBranch: chart.year_pillar?.zhi || '',
    solarDate: `${chart.birth_year}-${month}-${day} ${hour}:${minute}`,
    lunarDate: chart.calendar_type === 'LUNAR' ? '农历生日' : '',
    baziChartId: chart.id,
  }
}

// Load chart data
async function loadZiWeiChart() {
  loading.value = true
  error.value = ''
  try {
    const chartId = route.params.chartId
    // First fetch the chart to get birth info
    const chart = await fetchChart(String(chartId))
    if (!chart || !chart.birth_year) {
      error.value = '未找到命盘数据，请先生成八字命盘后再查看紫微斗数。'
      loading.value = false
      return
    }
    // Then calculate ziwei by chart id so the backend can cache the result.
    const data = await fetchZiWeiChart({
      chart_id: Number(chartId),
      profile: 'ziwei-local-composite-v2',
    })

    chartData.value = data
    birthInfo.value = mapBirthInfo(chart)

    // Let the backend resolve today's lunar-year label, especially before lunar new year.
    const defaultYear = (await loadOverlay()) || new Date().getFullYear()
    selectedLiunianYear.value = defaultYear
    availableYears.value = Array.from({ length: 11 }, (_, i) => defaultYear - 5 + i)
    await ensureDayunAnalysis(Number(chartId), true)
  } catch (reason: unknown) {
    const status = (reason as { response?: { status?: number } }).response?.status
    if (status === 404) {
      error.value = '该命盘不存在或已被删除，请重新创建。'
    } else {
      error.value = getApiErrorMessage(reason, '紫微命盘加载失败，请稍后重试。')
    }
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadZiWeiChart()
})

async function loadOverlay(year?: number): Promise<number | undefined> {
  try {
    const chartId = route.params.chartId
    const data = await fetchZiWeiOverlay({
      chart_id: Number(chartId),
      year,
    })
    const resolvedYear = Number(data.year)
    if (!Number.isInteger(resolvedYear)) return undefined
    liunianOverlay.value = { ...data, year: resolvedYear }
    return resolvedYear
  } catch {
    // Overlay data optional, don't block
    return undefined
  }
}

async function ensureDayunAnalysis(chartId: number, silent = false) {
  if (dayunAnalysis.value) return
  try {
    const data = await fetchZiWeiPeriod({
      chart_id: chartId,
      period_type: 'dayun',
    })
    dayunAnalysis.value = data.analysis || null
    dayunNominalAge.value = data.nominal_age ?? null
  } catch (err) {
    if (!silent) throw err
    console.warn('Failed to load dayun labels:', err)
  }
}

// Tab switching with data loading
async function switchTab(tab: string) {
  activeTab.value = tab
  loadingTab.value = true
  tabError.value = ''

  try {
    const chartId = route.params.chartId
    switch (tab) {
      case 'dayun':
        await ensureDayunAnalysis(Number(chartId))
        break
      case 'liunian':
        {
          const year = selectedLiunianYear.value
          const data = await fetchZiWeiPeriod({
            chart_id: Number(chartId),
            period_type: 'liunian',
            year,
          })
          liunianData.value = list(data.periods)
          liunianAnalysis.value = data.analysis || null
        }
        break
      case 'liuyue':
        {
          const year = selectedLiunianYear.value
          const month = new Date().getMonth() + 1
          const day = new Date().getDate()
          const data = await fetchZiWeiPeriod({
            chart_id: Number(chartId),
            period_type: 'liuyue',
            year,
            month,
            day,
          })
          liuyueData.value = list(data.periods)
          liuyueAnalysis.value = data.analysis || null
        }
        break
      case 'liuri':
        {
          const year = selectedLiunianYear.value
          const month = new Date().getMonth() + 1
          const day = new Date().getDate()
          const data = await fetchZiWeiPeriod({
            chart_id: Number(chartId),
            period_type: 'liuri',
            year,
            month,
            day,
          })
          liuriData.value = list(data.periods)
          liuriAnalysis.value = data.analysis || null
        }
        break
      case 'sihua':
        if (!Object.keys(sihuaData.value).length) {
          const data = await fetchZiWeiPeriod({
            chart_id: Number(chartId),
            period_type: 'sihua_feixing',
          })
          sihuaData.value = data.periods || {}
        }
        if (!Object.keys(sihuaChainData.value).length) {
          const chainData = await fetchZiWeiPeriod({
            chart_id: Number(chartId),
            period_type: 'sihua_chain',
          })
          sihuaChainData.value = chainData.chain || {}
        }
        break
    }
  } catch (err: unknown) {
    tabError.value = getApiErrorMessage(err, '加载周期分析失败')
  } finally {
    loadingTab.value = false
  }
}

async function onPalaceClick(palace: PalaceData, palaceIdx: number) {
  if (!route.params.chartId) return
  try {
    const data = await fetchZiWeiPeriod({
      chart_id: Number(route.params.chartId),
      period_type: 'palace_reading',
      palace_idx: palaceIdx,
    })
    const reading = data.reading
    selectedPalace.value = {
      palaceName: reading.palace_name || palace.name,
      palaceFocus: reading.palace_focus || '',
      summary: reading.summary || '',
      keyPoints: list(reading.key_points),
      evidence: list(reading.evidence),
      sanfangContext: reading.sanfang_context || null,
      patternDetails: list(reading.pattern_details),
      reviewNotes: list(reading.review_notes),
      limitations: list(reading.limitations),
      evidenceBasis: reading.evidence_basis || '',
      placementBasis: reading.placement_basis || '',
      interpretationBasis: reading.interpretation_basis || '',
      interpretationStatus: reading.interpretation_status || '',
      validationStatus: reading.validation_status || '',
      isOutcomeConclusion: reading.is_outcome_conclusion === true,
      mainStarAnalysis: {
        title: '主星特性',
        content: reading.main_star_analysis || '',
        tags: list(palace.stars)
          .filter((s) => s.type === 'major')
          .map((s) => s.name),
      },
      auxStarInfluence: {
        title: '辅星影响',
        content: reading.aux_star_influence || '',
        tags: list(palace.stars)
          .filter((s) => s.type !== 'major')
          .map((s) => s.name),
      },
      sihuaInfluence: {
        title: '本命四化结构',
        content: reading.sihua_influence || '',
        tags: list(palace.four_hua),
      },
      sanFangSiZheng: {
        title: '三方四正',
        content: reading.sanfang_analysis || '',
        tags: [],
      },
      patternAnnotations: {
        title: '格局标注',
        content: reading.pattern_notes || '',
        tags: list(chartData.value?.patterns),
      },
    }
  } catch (e) {
    console.error('Failed to load palace reading:', e)
  }
}

async function onYearChange(year: number) {
  selectedLiunianYear.value = year
  await loadOverlay(year)
  if (['liunian', 'liuyue', 'liuri'].includes(activeTab.value)) {
    await switchTab(activeTab.value)
  }
}

function palaceMajorSignal(p: PalaceData): string {
  const stars = list(p?.stars)
    .filter((s) => s.type === 'major')
    .map((s) => (s.brightness ? `${s.name}${s.brightness}` : s.name))
  if (!stars.length) return '空宫'
  return stars.slice(0, 3).join('、')
}

function palaceSupportSignal(p: PalaceData): string {
  const stars = list(p?.stars)
  const soft = stars.filter((s) => ['soft', 'lucun', 'tianma'].includes(s.type)).length
  const tough = stars.filter((s) => s.type === 'tough').length
  const parts = []
  if (soft) parts.push(`辅${soft}`)
  if (tough) parts.push(`煞${tough}`)
  return parts.length ? parts.join(' / ') : '辅煞少'
}

function palaceFourHuaLabel(p: PalaceData): string {
  const count = list(p?.four_hua).length
  return count ? `四化 ${count}` : '无四化'
}

function palaceFourHuaTitle(p: PalaceData): string {
  const fourHua = list(p?.four_hua)
  return fourHua.length ? fourHua.join('、') : '本宫无四化'
}

const chartOverviewItems = computed(() => {
  const chart = chartData.value
  if (!chart) return []
  return [
    { label: '命主', value: chart.life_master || '未定', hint: '先天性格取象' },
    { label: '身主', value: chart.body_master || '未定', hint: '后天行动取象' },
    { label: '五行局', value: chart.five_bureau || '未定', hint: '命盘局数底色' },
    {
      label: '身宫',
      value: chart.body_palace || chart.earthly_branch_of_body_palace || '未定',
      hint: '现实行动重心',
    },
  ]
})

function getPalacesFromPeriod(p: any): PalaceData[] {
  return list<PalaceData>(p?.palaces)
}

function majorStars(p: any): StarInfo[] {
  return list<StarInfo>(p?.stars).filter((s) => s.type === 'major')
}

function auxStars(p: any): StarInfo[] {
  return list<StarInfo>(p?.stars).filter((s) => s.type !== 'major')
}

const sihuaFlyGroups = computed<SihuaGroup[]>(() => {
  const data = sihuaData.value as any
  if (!data || !data.hua_lu) return []
  return [
    { type: '化禄', css: 'sihua-lu', items: list<ZiWeiSihuaProjectionItem>(data.hua_lu) },
    { type: '化权', css: 'sihua-quan', items: list<ZiWeiSihuaProjectionItem>(data.hua_quan) },
    { type: '化科', css: 'sihua-ke', items: list<ZiWeiSihuaProjectionItem>(data.hua_ke) },
    { type: '化忌', css: 'sihua-ji', items: list<ZiWeiSihuaProjectionItem>(data.hua_ji) },
  ]
})

const sihuaChainGroups = computed<SihuaGroup[]>(() => {
  const chain = sihuaChainData.value as any
  if (!chain || !chain.hua_lu) return []
  return [
    { type: '化禄', css: 'sihua-lu', items: list<ZiWeiSihuaProjectionItem>(chain.hua_lu) },
    { type: '化权', css: 'sihua-quan', items: list<ZiWeiSihuaProjectionItem>(chain.hua_quan) },
    { type: '化科', css: 'sihua-ke', items: list<ZiWeiSihuaProjectionItem>(chain.hua_ke) },
    { type: '化忌', css: 'sihua-ji', items: list<ZiWeiSihuaProjectionItem>(chain.hua_ji) },
  ]
})
</script>

<template>
  <div class="ziwei-page">
    <!-- Loading -->
    <div v-if="loading" class="loading-state">
      <el-skeleton animated>
        <template #template>
          <div
            style="
              display: grid;
              grid-template-columns: repeat(4, 1fr);
              gap: 8px;
              max-width: 400px;
              margin: 0 auto;
            "
          >
            <el-skeleton-item
              v-for="i in 12"
              :key="i"
              variant="rect"
              style="aspect-ratio: 1; border-radius: 8px"
            />
          </div>
        </template>
      </el-skeleton>
    </div>

    <!-- Error -->
    <div v-else-if="error" class="error-state">
      <el-result icon="error" title="加载失败" sub-title="请检查网络连接后重试">
        <template #extra>
          <el-button type="primary" @click="loadZiWeiChart">重试</el-button>
        </template>
      </el-result>
    </div>

    <!-- Content -->
    <div v-else class="page-content">
      <!-- Birth info bar -->
      <div class="birth-bar">
        <div class="birth-info-items">
          <span class="birth-item" v-if="birthInfo">
            <span class="birth-label">姓名</span>
            <span class="birth-val">{{ birthInfo.name }}</span>
          </span>
          <span class="birth-item" v-if="birthInfo">
            <span class="birth-label">性别</span>
            <span class="birth-val">{{ birthInfo.gender }}</span>
          </span>
          <span class="birth-item" v-if="birthInfo">
            <span class="birth-label">公历</span>
            <span class="birth-val">{{ birthInfo.solarDate }}</span>
          </span>
          <span class="birth-item" v-if="birthInfo?.lunarDate">
            <span class="birth-label">农历</span>
            <span class="birth-val">{{ birthInfo.lunarDate }}</span>
          </span>
          <span class="birth-item" v-if="chartData">
            <span class="birth-label">排盘口径</span>
            <span class="birth-val">传统紫微排盘</span>
          </span>
          <span
            class="birth-item"
            v-for="source in chartData?.rule_sources || []"
            :key="source.rule_id"
          >
            <span class="birth-label">{{
              source.rule_id.includes('sihua') ? '四化来源' : '亮度来源'
            }}</span>
            <a
              class="birth-val source-link"
              :href="`${source.repository}/blob/${source.commit}/${source.path}`"
              target="_blank"
              rel="noopener noreferrer"
            >
              iztro@{{ source.commit.slice(0, 7) }}
            </a>
          </span>
        </div>
        <router-link v-if="birthInfo" :to="'/chart/' + birthInfo.baziChartId" class="bazi-link">
          查看八字命盘 →
        </router-link>
      </div>

      <!-- Overlay section (本命盘/流年叠盘 toggle) -->
      <div class="overlay-section" v-if="chartData && liunianOverlay">
        <ZiWeiOverlay
          :base-chart="{
            palaces: chartData.palaces,
            life_master: chartData.life_master,
            body_master: chartData.body_master,
            five_bureau: chartData.five_bureau,
            earthly_branch_of_soul_palace: chartData.earthly_branch_of_soul_palace,
            earthly_branch_of_body_palace: chartData.earthly_branch_of_body_palace,
          }"
          :liunian-chart="liunianOverlay"
          :dayun-stages="dayunStageList"
          :available-years="availableYears"
          :birth-year-branch="birthInfo?.yearBranch || ''"
          :gender="birthInfo?.gender || ''"
          :overlay-analysis="liunianOverlay.overlay_analysis"
          @year-change="onYearChange"
        />
      </div>

      <!-- Tabs -->
      <div class="tabs-section">
        <div class="tab-bar">
          <button
            v-for="tab in [
              { key: 'mingpan', label: '命盘详解' },
              { key: 'dayun', label: '大限分析' },
              { key: 'liunian', label: '流年分析' },
              { key: 'liuyue', label: '流月分析' },
              { key: 'liuri', label: '流日分析' },
              { key: 'sihua', label: '四化飞星' },
            ]"
            :key="tab.key"
            class="tab-btn"
            :class="{ active: activeTab === tab.key }"
            @click="switchTab(tab.key)"
          >
            {{ tab.label }}
          </button>
        </div>

        <!-- Tab content -->
        <div class="tab-content">
          <!-- 命盘详解 -->
          <div v-if="activeTab === 'mingpan'" class="mingpan-tab">
            <p class="tab-desc">总览命盘底色，选择宫位查看主星、辅星、四化和三方四正依据</p>

            <section class="chart-overview-panel" v-if="chartData">
              <div class="overview-items">
                <div v-for="item in chartOverviewItems" :key="item.label" class="overview-item">
                  <span class="overview-label">{{ item.label }}</span>
                  <strong>{{ item.value }}</strong>
                  <small>{{ item.hint }}</small>
                </div>
              </div>
              <div class="overview-patterns">
                <span class="overview-label">格局</span>
                <div v-if="list(chartData.patterns).length" class="overview-pattern-list">
                  <span v-for="(pattern, index) in list(chartData.patterns)" :key="index">{{
                    pattern
                  }}</span>
                </div>
                <p v-else>当前没有可直接列出的格局线索</p>
              </div>
            </section>

            <div class="palace-quick-grid">
              <button
                v-for="(palace, idx) in list(chartData?.palaces)"
                :key="palace.branch"
                class="palace-pill"
                :class="{
                  active: selectedPalace?.palaceName === palace.name,
                  'body-palace':
                    chartData && (palace.name === chartData.body_palace || palace.is_body_palace),
                }"
                @click="onPalaceClick(palace, idx)"
              >
                <span class="palace-pill-name">{{ palace.name }}</span>
                <span class="palace-pill-branch">{{ palace.branch }}</span>
                <span class="palace-pill-stars">{{ palaceMajorSignal(palace) }}</span>
                <span class="palace-pill-meta" :title="palaceFourHuaTitle(palace)">{{
                  palaceFourHuaLabel(palace)
                }}</span>
                <span class="palace-pill-meta support">{{ palaceSupportSignal(palace) }}</span>
                <span
                  v-if="
                    chartData && (palace.name === chartData.body_palace || palace.is_body_palace)
                  "
                  class="body-badge"
                  >身</span
                >
              </button>
            </div>

            <ZiWeiInterpretation v-if="selectedPalace" :palace-reading="selectedPalace" />
            <div v-else class="empty-state-inline">
              <svg width="40" height="40" viewBox="0 0 40 40" fill="none">
                <circle
                  cx="20"
                  cy="20"
                  r="16"
                  stroke="currentColor"
                  stroke-width="0.5"
                  stroke-dasharray="2 3"
                  opacity="0.3"
                />
                <circle cx="20" cy="20" r="3" fill="currentColor" opacity="0.2" />
              </svg>
              <p>选择一个宫位查看详细解读</p>
            </div>
          </div>

          <!-- 大限分析 -->
          <div v-else-if="activeTab === 'dayun'" class="data-tab">
            <p class="tab-desc">先用十年大限定底色，再叠加流年、流月、流日看触发点</p>
            <div v-if="loadingTab" class="tab-loading">
              <div class="loading-dots"><span></span><span></span><span></span></div>
            </div>
            <div v-else-if="tabError" class="tab-error">
              <strong>加载失败</strong>
              <p>{{ tabError }}</p>
            </div>
            <div v-else-if="!dayunAnalysis" class="empty-state-inline">
              <p>暂无可显示的大限分析</p>
            </div>
            <div v-else>
              <ZiWeiPeriodAnalysisPanel
                :analysis="dayunAnalysis"
                title="大限分析"
                :subtitle="`当前虚岁 ${dayunNominalAge} 岁`"
              />
            </div>
          </div>

          <!-- 流年分析 -->
          <div v-else-if="activeTab === 'liunian'" class="data-tab">
            <p class="tab-desc">
              {{
                liunianData[0]?.year ? liunianData[0].year + '年' : ''
              }}先看年度干支、四化、流禄羊陀马，再落到重点宫位
            </p>
            <div v-if="loadingTab" class="tab-loading">
              <div class="loading-dots"><span></span><span></span><span></span></div>
            </div>
            <div v-else-if="tabError" class="tab-error">
              <strong>加载失败</strong>
              <p>{{ tabError }}</p>
            </div>
            <div v-else-if="!liunianData.length" class="empty-state-inline">
              <p>暂无可显示的流年数据</p>
            </div>
            <div v-else>
              <ZiWeiPeriodAnalysisPanel
                :analysis="liunianAnalysis"
                title="流年分析"
                :subtitle="`${selectedLiunianYear}年年度触发`"
              />
              <details class="period-detail-fold">
                <summary>宫位明细</summary>
                <div
                  v-for="p in getPalacesFromPeriod(liunianData[0])"
                  :key="p.branch"
                  class="palace-strip"
                >
                  <div class="palace-strip-header">
                    <span class="strip-name">{{ p.name }}</span
                    ><span class="strip-branch">{{ p.branch }}</span>
                  </div>
                  <div class="palace-strip-stars">
                    <template v-if="majorStars(p).length"
                      ><span
                        v-for="s in majorStars(p)"
                        :key="s.name"
                        class="strip-main-star"
                        :class="{ dim: !s.brightness }"
                        >{{ s.name }}<small v-if="s.brightness">·{{ s.brightness }}</small></span
                      ></template
                    >
                    <span v-if="!majorStars(p).length" class="strip-empty">无主星</span>
                    <template v-if="auxStars(p).length"
                      ><span
                        v-for="s in auxStars(p).slice(0, 4)"
                        :key="s.name"
                        class="strip-aux-star"
                        >{{ s.name }}</span
                      ></template
                    >
                    <template v-if="list(p.adjective_stars).length"
                      ><span v-for="s in list(p.adjective_stars)" :key="s" class="strip-adj-star">{{
                        s
                      }}</span></template
                    >
                    <template v-if="list(p.four_hua).length"
                      ><span v-for="s in list(p.four_hua)" :key="s" class="strip-sihua">{{
                        s
                      }}</span></template
                    >
                  </div>
                  <div
                    v-if="p.changsheng_12 || p.boshi_12 || p.jiang_qian_12 || p.sui_qian_12"
                    class="strip-twelve-stars"
                  >
                    <span v-if="p.changsheng_12" class="twelve-tag twelve-cs">{{
                      p.changsheng_12
                    }}</span>
                    <span v-if="p.boshi_12" class="twelve-tag twelve-bs">{{ p.boshi_12 }}</span>
                    <span v-if="p.jiang_qian_12" class="twelve-tag twelve-jq">{{
                      p.jiang_qian_12
                    }}</span>
                    <span v-if="p.sui_qian_12" class="twelve-tag twelve-sq">{{
                      p.sui_qian_12
                    }}</span>
                  </div>
                  <div v-if="p.sanfang_sizheng" class="strip-sanfang">
                    <span class="sf-label">三方四正</span>
                    <span v-if="p.sanfang_sizheng.opposite" class="sf-item"
                      >对{{ p.sanfang_sizheng.opposite }}</span
                    >
                    <span v-if="p.sanfang_sizheng.trine1" class="sf-item"
                      >三{{ p.sanfang_sizheng.trine1 }}</span
                    >
                    <span v-if="p.sanfang_sizheng.trine2" class="sf-item"
                      >三{{ p.sanfang_sizheng.trine2 }}</span
                    >
                  </div>
                </div>
              </details>
            </div>
          </div>

          <!-- 流月分析 -->
          <div v-else-if="activeTab === 'liuyue'" class="data-tab">
            <p class="tab-desc">
              {{
                liuyueData[0]?.year ? liuyueData[0].year + '年' + liuyueData[0].month + '月' : ''
              }}本月看哪一宫被点亮，再回头看整体月份主题
            </p>
            <div v-if="loadingTab" class="tab-loading">
              <div class="loading-dots"><span></span><span></span><span></span></div>
            </div>
            <div v-else-if="tabError" class="tab-error">
              <strong>加载失败</strong>
              <p>{{ tabError }}</p>
            </div>
            <div v-else-if="!liuyueData.length" class="empty-state-inline">
              <p>暂无可显示的流月数据</p>
            </div>
            <div v-else>
              <ZiWeiPeriodAnalysisPanel
                :analysis="liuyueAnalysis"
                title="流月分析"
                :subtitle="`${liuyueData[0].year}年${liuyueData[0].month}月月度触发`"
              />
              <details class="period-detail-fold">
                <summary>宫位明细</summary>
                <div
                  v-for="p in getPalacesFromPeriod(liuyueData[0])"
                  :key="'ly-' + p.branch"
                  class="palace-strip"
                >
                  <div class="palace-strip-header">
                    <span class="strip-name">{{ p.name }}</span
                    ><span class="strip-branch">{{ p.branch }}</span>
                  </div>
                  <div class="palace-strip-stars">
                    <template v-if="majorStars(p).length"
                      ><span
                        v-for="s in majorStars(p)"
                        :key="s.name"
                        class="strip-main-star"
                        :class="{ dim: !s.brightness }"
                        >{{ s.name }}<small v-if="s.brightness">·{{ s.brightness }}</small></span
                      ></template
                    >
                    <span v-if="!majorStars(p).length" class="strip-empty">无主星</span>
                    <template v-if="auxStars(p).length"
                      ><span
                        v-for="s in auxStars(p).slice(0, 4)"
                        :key="s.name"
                        class="strip-aux-star"
                        >{{ s.name }}</span
                      ></template
                    >
                    <template v-if="list(p.adjective_stars).length"
                      ><span v-for="s in list(p.adjective_stars)" :key="s" class="strip-adj-star">{{
                        s
                      }}</span></template
                    >
                    <template v-if="list(p.four_hua).length"
                      ><span v-for="s in list(p.four_hua)" :key="s" class="strip-sihua">{{
                        s
                      }}</span></template
                    >
                  </div>
                  <div
                    v-if="p.changsheng_12 || p.boshi_12 || p.jiang_qian_12 || p.sui_qian_12"
                    class="strip-twelve-stars"
                  >
                    <span v-if="p.changsheng_12" class="twelve-tag twelve-cs">{{
                      p.changsheng_12
                    }}</span>
                    <span v-if="p.boshi_12" class="twelve-tag twelve-bs">{{ p.boshi_12 }}</span>
                    <span v-if="p.jiang_qian_12" class="twelve-tag twelve-jq">{{
                      p.jiang_qian_12
                    }}</span>
                    <span v-if="p.sui_qian_12" class="twelve-tag twelve-sq">{{
                      p.sui_qian_12
                    }}</span>
                  </div>
                  <div v-if="p.sanfang_sizheng" class="strip-sanfang">
                    <span class="sf-label">三方四正</span>
                    <span v-if="p.sanfang_sizheng.opposite" class="sf-item"
                      >对{{ p.sanfang_sizheng.opposite }}</span
                    >
                    <span v-if="p.sanfang_sizheng.trine1" class="sf-item"
                      >三{{ p.sanfang_sizheng.trine1 }}</span
                    >
                    <span v-if="p.sanfang_sizheng.trine2" class="sf-item"
                      >三{{ p.sanfang_sizheng.trine2 }}</span
                    >
                  </div>
                </div>
              </details>
            </div>
          </div>

          <!-- 流日分析 -->
          <div v-else-if="activeTab === 'liuri'" class="data-tab">
            <p class="tab-desc">
              {{
                liuriData[0]?.year
                  ? liuriData[0].year + '年' + liuriData[0].month + '月' + liuriData[0].day + '日'
                  : ''
              }}先看当天触发宫位，再看时辰窗口
            </p>
            <div v-if="loadingTab" class="tab-loading">
              <div class="loading-dots"><span></span><span></span><span></span></div>
            </div>
            <div v-else-if="tabError" class="tab-error">
              <strong>加载失败</strong>
              <p>{{ tabError }}</p>
            </div>
            <div v-else-if="!liuriData.length" class="empty-state-inline">
              <p>暂无可显示的流日数据</p>
            </div>
            <div v-else>
              <ZiWeiPeriodAnalysisPanel
                :analysis="liuriAnalysis"
                title="流日分析"
                :subtitle="`${liuriData[0].year}年${liuriData[0].month}月${liuriData[0].day}日当日触发`"
              />
              <details class="period-detail-fold">
                <summary>宫位明细</summary>
                <div
                  v-for="p in getPalacesFromPeriod(liuriData[0])"
                  :key="'lr-' + p.branch"
                  class="palace-strip palace-strip-sm"
                >
                  <div class="palace-strip-header">
                    <span class="strip-name">{{ p.name }}</span
                    ><span class="strip-branch">{{ p.branch }}</span>
                  </div>
                  <div class="palace-strip-stars">
                    <template v-if="majorStars(p).length"
                      ><span
                        v-for="s in majorStars(p)"
                        :key="s.name"
                        class="strip-main-star"
                        :class="{ dim: !s.brightness }"
                        >{{ s.name }}<small v-if="s.brightness">·{{ s.brightness }}</small></span
                      ></template
                    >
                    <span v-if="!majorStars(p).length" class="strip-empty">无主星</span>
                    <template v-if="auxStars(p).length"
                      ><span
                        v-for="s in auxStars(p).slice(0, 4)"
                        :key="s.name"
                        class="strip-aux-star"
                        >{{ s.name }}</span
                      ></template
                    >
                    <template v-if="list(p.adjective_stars).length"
                      ><span v-for="s in list(p.adjective_stars)" :key="s" class="strip-adj-star">{{
                        s
                      }}</span></template
                    >
                    <template v-if="list(p.four_hua).length"
                      ><span v-for="s in list(p.four_hua)" :key="s" class="strip-sihua">{{
                        s
                      }}</span></template
                    >
                  </div>
                  <div
                    v-if="p.changsheng_12 || p.boshi_12 || p.jiang_qian_12 || p.sui_qian_12"
                    class="strip-twelve-stars"
                  >
                    <span v-if="p.changsheng_12" class="twelve-tag twelve-cs">{{
                      p.changsheng_12
                    }}</span>
                    <span v-if="p.boshi_12" class="twelve-tag twelve-bs">{{ p.boshi_12 }}</span>
                    <span v-if="p.jiang_qian_12" class="twelve-tag twelve-jq">{{
                      p.jiang_qian_12
                    }}</span>
                    <span v-if="p.sui_qian_12" class="twelve-tag twelve-sq">{{
                      p.sui_qian_12
                    }}</span>
                  </div>
                  <div v-if="p.sanfang_sizheng" class="strip-sanfang">
                    <span class="sf-label">三方四正</span>
                    <span v-if="p.sanfang_sizheng.opposite" class="sf-item"
                      >对{{ p.sanfang_sizheng.opposite }}</span
                    >
                    <span v-if="p.sanfang_sizheng.trine1" class="sf-item"
                      >三{{ p.sanfang_sizheng.trine1 }}</span
                    >
                    <span v-if="p.sanfang_sizheng.trine2" class="sf-item"
                      >三{{ p.sanfang_sizheng.trine2 }}</span
                    >
                  </div>
                </div>
              </details>
            </div>
          </div>

          <!-- 四化飞星 -->
          <div v-else-if="activeTab === 'sihua'" class="data-tab">
            <div v-if="loadingTab" class="tab-loading">
              <div class="loading-dots"><span></span><span></span><span></span></div>
            </div>
            <div v-else-if="tabError" class="tab-error">
              <strong>加载失败</strong>
              <p>{{ tabError }}</p>
            </div>
            <div v-else-if="!sihuaFlyGroups.length" class="empty-state-inline">
              <p>暂无可显示的四化飞星数据</p>
            </div>
            <div v-else class="sihua-groups">
              <div v-for="grp in sihuaFlyGroups" :key="grp.type" class="sihua-group">
                <span class="sihua-group-badge" :class="grp.css">{{ grp.type }}</span>
                <div v-if="grp.items.length" class="sihua-group-items">
                  <div v-for="(it, i) in grp.items" :key="i" class="sihua-fly-item">
                    <span class="fly-star">{{ it.transformed_star }}</span
                    ><span class="fly-arrow">→</span>
                    <span class="fly-palace">{{ it.target_palace }}</span>
                  </div>
                </div>
                <p v-else class="sihua-empty-group">无</p>
              </div>
              <div v-if="sihuaChainGroups.length" class="sihua-chain-section">
                <h4 class="chain-title">宫干四化飞行</h4>
                <div v-for="grp in sihuaChainGroups" :key="'chain-' + grp.type" class="sihua-group">
                  <span class="sihua-group-badge" :class="grp.css">{{ grp.type }}</span>
                  <div v-if="grp.items.length" class="sihua-group-items">
                    <div v-for="(it, i) in grp.items" :key="'c-' + i" class="sihua-fly-item">
                      <span class="fly-star">{{ it.transformed_star }}</span
                      ><span class="fly-arrow">→</span>
                      <span class="fly-palace">{{ it.target_palace }}</span>
                      <span v-if="it.source_palace" class="fly-from"
                        >源{{ it.source_palace }}{{ it.source_palace_stem }}</span
                      >
                      <span class="fly-chain">{{ it.is_self_mutagen ? '自化' : '跨宫' }}</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
@reference "tailwindcss";
.ziwei-page {
  @apply min-h-screen relative;
  background-color: transparent;
}

.loading-state {
  @apply flex items-center justify-center min-h-[60vh];
  position: relative;
  z-index: 1;
}

.loading-spinner {
  width: 2rem;
  height: 2rem;
  border-radius: 50%;
  border: 2px dashed;
  animation: spin 1s linear infinite;
  border-color: var(--color-bazi-blue);
  border-top-color: transparent;
}

.error-state {
  @apply flex items-center justify-center min-h-[60vh];
  position: relative;
  z-index: 1;
}

.error-card {
  @apply rounded-lg shadow-md p-8 text-center max-w-sm;
  background: var(--surface-0);
}

.page-content {
  @apply max-w-5xl mx-auto px-4 py-6;
  position: relative;
  z-index: 1;
}

/* Birth info bar */
.birth-bar {
  @apply flex items-center justify-between flex-wrap gap-3 mb-6 p-4 rounded-lg;
  background: color-mix(in oklab, var(--surface-1) 84%, transparent);
  border: 1px solid var(--line-strong);
  box-shadow: var(--shadow-sm);
}

.birth-info-items {
  @apply flex flex-wrap gap-4;
}

.birth-item {
  @apply flex items-center gap-1.5;
}

.birth-label {
  @apply text-xs;
  color: var(--text-muted);
}

.birth-val {
  @apply text-sm font-medium;
  color: var(--color-bazi-blue);
}

.source-link {
  text-decoration: none;
}

.source-link:hover {
  text-decoration: underline;
}

.bazi-link {
  @apply text-sm font-medium no-underline;
  color: var(--color-bazi-red);
}

.bazi-link:hover {
  text-decoration: underline;
}

/* Chart section */
.chart-section {
  @apply mb-6;
}

.overlay-section {
  @apply mb-6;
}

/* Tabs */
.tabs-section {
  @apply rounded-lg overflow-hidden;
  background: color-mix(in oklab, var(--surface-1) 84%, transparent);
  border: 1px solid var(--line-strong);
  box-shadow: var(--shadow-sm);
}

.tab-bar {
  @apply flex overflow-x-auto;
  border-bottom: 2px solid var(--line-subtle);
}

.tab-btn {
  @apply flex-shrink-0 px-4 py-3 text-sm font-medium cursor-pointer border-0 bg-transparent transition-colors;
  color: var(--text-muted);
  border-bottom: 2px solid transparent;
  margin-bottom: -2px;
}

.tab-btn:hover {
  color: var(--color-bazi-blue);
}

.tab-btn.active {
  color: var(--color-bazi-red);
  border-bottom-color: var(--color-bazi-red);
}

.tab-content {
  @apply p-4 min-h-[200px];
}

.period-detail-fold {
  margin-top: 0.6rem;
  padding: 0.55rem 0.65rem;
  border: 1px solid var(--line-subtle);
  border-radius: 8px;
  background: color-mix(in oklab, var(--surface-0) 68%, transparent);
}

.period-detail-fold > summary {
  cursor: pointer;
  list-style: none;
  font-size: var(--fs-2xs);
  font-weight: 700;
  color: var(--accent);
}

.period-detail-fold > summary::-webkit-details-marker {
  display: none;
}

.tab-hint {
  font-size: var(--fs-xs);
  color: var(--text-muted);
  margin-bottom: 0.75rem;
}

.tab-loading {
  text-align: center;
  font-size: var(--fs-sm);
  color: var(--text-muted);
  padding-top: 2rem;
  padding-bottom: 2rem;
}

.tab-error {
  padding: 0.75rem 0.85rem;
  border: 1px solid rgba(220, 38, 38, 0.18);
  border-radius: 8px;
  background: rgba(220, 38, 38, 0.06);
  color: var(--danger);
}

.tab-error strong {
  display: block;
  margin-bottom: 0.2rem;
  font-size: var(--fs-xs);
}

.tab-error p {
  margin: 0;
  font-size: var(--fs-2xs);
  line-height: 1.5;
  color: var(--text-muted);
}

.empty-hint {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  padding-top: 2.5rem;
  padding-bottom: 2.5rem;
  color: var(--text-muted);
}

.chart-overview-panel {
  display: grid;
  grid-template-columns: minmax(0, 1.35fr) minmax(220px, 0.65fr);
  gap: 0.75rem;
  margin-bottom: 1rem;
  padding: 0.75rem;
  border: 1px solid var(--line-subtle);
  border-radius: 8px;
  background: color-mix(in oklab, var(--surface-2) 66%, transparent);
}

.overview-items {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(118px, 1fr));
  gap: 0.5rem;
}

.overview-item {
  min-width: 0;
  padding: 0.55rem 0.65rem;
  border: 1px solid var(--line-subtle);
  border-radius: 8px;
  background: color-mix(in oklab, var(--surface-0) 70%, transparent);
}

.overview-label {
  display: block;
  font-size: var(--fs-2xs);
  font-weight: 700;
  color: var(--text-muted);
}

.overview-item strong {
  display: block;
  margin-top: 0.18rem;
  font-size: var(--fs-sm);
  color: var(--accent);
}

.overview-item small {
  display: block;
  margin-top: 0.12rem;
  font-size: var(--fs-2xs);
  line-height: 1.3;
  color: var(--text-muted);
}

.overview-patterns {
  min-width: 0;
  padding: 0.55rem 0.65rem;
  border: 1px solid var(--line-subtle);
  border-radius: 8px;
  background: color-mix(in oklab, var(--accent) 4%, transparent);
}

.overview-pattern-list {
  display: flex;
  flex-wrap: wrap;
  gap: 0.3rem;
  margin-top: 0.45rem;
}

.overview-pattern-list span {
  max-width: 100%;
  padding: 0.08rem 0.4rem;
  border: 1px solid var(--line-subtle);
  border-radius: 999px;
  background: color-mix(in oklab, var(--surface-1) 72%, transparent);
  color: var(--accent);
  font-size: var(--fs-2xs);
  font-weight: 600;
  overflow-wrap: anywhere;
}

.overview-pattern-list .more-patterns {
  color: var(--text-muted);
}

.overview-patterns p {
  margin: 0.45rem 0 0;
  font-size: var(--fs-2xs);
  color: var(--text-muted);
}

/* Palace quick grid */
.palace-quick-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(132px, 1fr));
  gap: 0.5rem;
  margin-bottom: 1.25rem;
}
.palace-pill {
  display: grid;
  grid-template-columns: 1fr auto;
  align-items: center;
  gap: 0.2rem 0.4rem;
  min-height: 92px;
  padding: 0.55rem 0.6rem;
  background: var(--accent-dim);
  border: 1px solid var(--line-subtle);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
  font-family: var(--font-sans);
  text-align: left;
}
.palace-pill:hover {
  background: color-mix(in oklab, var(--accent) 8%, transparent);
  border-color: var(--line-focus);
  transform: translateY(-1px);
}
.palace-pill.active {
  border-color: var(--text-soft);
  background: var(--line-strong);
}
.palace-pill.body-palace {
  border-color: rgba(139, 75, 75, 0.4);
  background: rgba(139, 75, 75, 0.06);
}
.palace-pill-name {
  font-size: var(--fs-xs);
  font-weight: 600;
  color: var(--text);
  letter-spacing: 0;
}
.palace-pill-branch {
  font-size: var(--fs-2xs);
  color: var(--text-muted);
  text-align: right;
}
.palace-pill-stars {
  grid-column: 1 / -1;
  min-height: 1rem;
  font-size: var(--fs-2xs);
  line-height: 1.25;
  color: var(--accent);
  overflow-wrap: anywhere;
}
.palace-pill-meta {
  justify-self: start;
  font-size: var(--fs-2xs);
  color: var(--text-muted);
  background: color-mix(in oklab, var(--surface-2) 72%, transparent);
  border: 1px solid var(--line-subtle);
  border-radius: 999px;
  padding: 0.05rem 0.35rem;
}
.palace-pill-meta.support {
  justify-self: end;
}
.body-badge {
  justify-self: end;
  font-size: var(--fs-2xs);
  background: rgba(251, 113, 133, 0.12);
  color: var(--danger);
  padding: 0.05rem 0.25rem;
  border-radius: 3px;
  font-weight: 600;
}
.body-palace .palace-pill-name {
  color: var(--danger);
}

/* Period lists */
.period-list {
  @apply flex flex-col gap-3;
}

.period-card {
  @apply rounded-md border p-3;
  background: color-mix(in oklab, var(--accent) 5%, transparent);
  border-color: var(--line-subtle);
}

.period-header {
  @apply flex items-center justify-between mb-1;
}

.period-age {
  @apply text-sm font-bold;
  color: var(--color-bazi-red);
}

.period-palace {
  @apply text-sm font-medium;
  color: var(--color-bazi-blue);
}

.period-desc {
  font-size: var(--fs-xs);
  color: var(--text-muted);
  margin: 0;
}

/* Sihua table */
.sihua-table-wrap {
  @apply overflow-x-auto;
}

.sihua-table {
  @apply w-full text-sm border-collapse;
}

.sihua-table th {
  @apply text-left p-2 font-bold text-xs;
  background: var(--surface-2);
  color: var(--color-bazi-blue);
  border-bottom: 2px solid var(--line-strong);
}

.sihua-table td {
  @apply p-2 border-b;
  border-color: var(--line-subtle);
}

.sihua-badge {
  @apply inline-block rounded-full px-2 py-0.5 text-xs font-semibold;
}

.sihua-lu {
  background: rgba(22, 163, 74, 0.1);
  color: #16a34a;
}

.sihua-quan {
  background-color: var(--line-strong);
  color: var(--color-bazi-blue);
}

.sihua-ke {
  background-color: rgba(37, 99, 235, 0.1);
  color: #2563eb;
}

.sihua-ji {
  background-color: rgba(220, 38, 38, 0.08);
  color: #dc2626;
}

.tab-desc {
  font-size: var(--fs-xs);
  color: var(--text-muted);
  margin: 0 0 1rem;
  font-style: italic;
}
.dayun-timeline {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}
.dayun-card {
  display: flex;
  gap: 0.75rem;
  padding: 0.625rem 0.75rem;
  background: color-mix(in oklab, var(--accent) 4%, transparent);
  border: 1px solid var(--line-subtle);
  border-radius: 8px;
}
.dayun-card.is-current {
  border-color: var(--line-focus);
  background: var(--line-subtle);
}
.dayun-age-badge {
  min-width: 60px;
  text-align: center;
}
.age-primary {
  display: block;
  font-size: var(--fs-sm);
  font-weight: 700;
  color: var(--accent);
}
.age-unit {
  font-size: var(--fs-2xs);
  color: var(--text-muted);
}
.dayun-body {
  flex: 1;
}
.dayun-palace {
  font-size: var(--fs-sm);
  font-weight: 600;
  color: var(--text);
  margin-bottom: 0.125rem;
}
.dayun-desc {
  font-size: var(--fs-2xs);
  color: var(--text-muted);
  margin: 0 0 0.25rem;
}
.dayun-stars {
  display: flex;
  flex-wrap: wrap;
  gap: 0.2rem;
}
.star-chip {
  padding: 0.08rem 0.35rem;
  font-size: var(--fs-2xs);
  background: rgba(220, 38, 38, 0.08);
  color: var(--danger);
  border-radius: 3px;
  border: 1px solid rgba(220, 38, 38, 0.12);
}
.palace-strip {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  padding: 0.4rem 0.6rem;
  border-bottom: 1px solid var(--line-subtle);
}
.palace-strip-sm {
  padding: 0.25rem 0.4rem;
}
.palace-strip-header {
  display: flex;
  align-items: center;
  gap: 0.3rem;
  min-width: 68px;
}
.strip-name {
  font-size: var(--fs-xs);
  font-weight: 600;
  color: var(--text);
}
.strip-branch {
  font-size: var(--fs-2xs);
  color: var(--text-muted);
}
.palace-strip-stars {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  flex-wrap: wrap;
  flex: 1;
}
.strip-main-star {
  padding: 0.06rem 0.3rem;
  font-size: var(--fs-2xs);
  font-weight: 600;
  background: var(--line-strong);
  color: var(--accent);
  border-radius: 3px;
}
.strip-main-star.dim {
  color: var(--text-muted);
  background: var(--accent-dim);
}
.strip-main-star small {
  font-size: var(--fs-2xs);
  opacity: 0.5;
}
.strip-aux-star {
  font-size: var(--fs-2xs);
  color: var(--text-soft);
}
.strip-sihua {
  padding: 0.06rem 0.3rem;
  font-size: var(--fs-2xs);
  background: rgba(220, 38, 38, 0.08);
  color: var(--danger);
  border-radius: 3px;
}
.strip-empty {
  font-size: var(--fs-2xs);
  color: var(--text-muted);
  opacity: 0.3;
}
.sihua-groups {
  display: flex;
  flex-direction: column;
  gap: 0.625rem;
}
.sihua-group-badge {
  display: inline-block;
  padding: 0.15rem 0.5rem;
  font-size: var(--fs-2xs);
  font-weight: 700;
  border-radius: 4px;
  margin-bottom: 0.25rem;
}
.sihua-lu {
  background: rgba(22, 163, 74, 0.1);
  color: #16a34a;
}
.sihua-quan {
  background: rgba(71, 85, 105, 0.1);
  color: #475569;
}
.sihua-ke {
  background: rgba(37, 99, 235, 0.1);
  color: #2563eb;
}
.sihua-ji {
  background: rgba(220, 38, 38, 0.08);
  color: #dc2626;
}
.sihua-group-items {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}
.sihua-fly-item {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.3rem 0.5rem;
  background: color-mix(in oklab, var(--accent) 3%, transparent);
  border-radius: 5px;
  font-size: var(--fs-xs);
}
.fly-star {
  font-weight: 600;
  color: var(--text);
}
.fly-arrow {
  color: var(--text-muted);
  font-size: var(--fs-2xs);
}
.fly-palace {
  color: var(--accent);
  font-weight: 500;
}
.fly-effect {
  color: var(--text-muted);
  font-size: var(--fs-2xs);
  flex: 1;
}
.fly-from {
  font-size: var(--fs-2xs);
  color: #2563eb;
  background: rgba(37, 99, 235, 0.08);
  padding: 0.05rem 0.25rem;
  border-radius: 3px;
}
.fly-chain {
  font-size: var(--fs-2xs);
  color: #16a34a;
  background: rgba(22, 163, 74, 0.08);
  padding: 0.05rem 0.25rem;
  border-radius: 3px;
}
.fly-affinity {
  font-size: var(--fs-2xs);
  color: #a16207;
  background: rgba(161, 98, 7, 0.08);
  padding: 0.05rem 0.25rem;
  border-radius: 3px;
}
.sihua-chain-section {
  margin-top: 1rem;
  padding-top: 0.75rem;
  border-top: 1px solid var(--line-subtle);
}
.chain-title {
  font-size: var(--fs-xs);
  color: var(--text-muted);
  margin: 0 0 0.5rem;
  font-weight: 600;
}
.sihua-empty-group {
  font-size: var(--fs-2xs);
  color: var(--text-muted);
  padding: 0.2rem 0.4rem;
  opacity: 0.4;
}

/* Interpretation tab styles */
.interp-tab {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}
.interp-card {
  background: var(--glass);
  border: 1px solid var(--line-strong);
  border-radius: 12px;
  overflow: hidden;
}
.interp-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem 1rem;
  background: var(--line-subtle);
  border-bottom: 1px solid var(--line-strong);
}
.interp-year {
  font-size: var(--fs-sm);
  font-weight: 700;
  color: var(--accent);
  font-family: var(--font-serif);
}
.interp-ganZhi {
  font-size: var(--fs-sm);
  color: var(--text);
  font-weight: 600;
}
.interp-score {
  font-size: var(--fs-sm);
  font-weight: 700;
  padding: 0.15rem 0.5rem;
  border-radius: 6px;
  margin-left: auto;
}
.score-good {
  background: rgba(22, 163, 74, 0.1);
  color: #16a34a;
}
.score-bad {
  background: rgba(220, 38, 38, 0.08);
  color: #dc2626;
}
.interp-section {
  padding: 0.75rem 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}
.interp-row {
  display: flex;
  gap: 0.5rem;
  font-size: var(--fs-xs);
  line-height: 1.5;
}
.interp-label {
  min-width: 70px;
  font-weight: 600;
  color: var(--text-muted);
}
.interp-value {
  color: var(--text);
  flex: 1;
}
.interp-value.danger {
  color: #dc2626;
}
.interp-row.tip .interp-value {
  color: var(--accent);
  font-style: italic;
}
.interp-subtitle {
  font-size: var(--fs-xs);
  font-weight: 700;
  color: var(--accent);
  margin: 0 0 0.5rem;
}
.interp-summary {
  font-size: var(--fs-xs);
  color: var(--text-muted);
  padding: 0.5rem 1rem;
  background: var(--line-subtle);
  border-top: 1px dashed var(--line-strong);
  font-style: italic;
}
.hourly-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(110px, 1fr));
  gap: 0.375rem;
}
.hour-block {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
  padding: 0.4rem 0.5rem;
  border-radius: 6px;
  font-size: var(--fs-2xs);
}
.hour-good {
  background: rgba(22, 163, 74, 0.08);
  border: 1px solid rgba(22, 163, 74, 0.15);
}
.hour-neutral {
  background: var(--line-subtle);
  border: 1px solid var(--line-strong);
}
.hour-bad {
  background: rgba(220, 38, 38, 0.06);
  border: 1px solid rgba(220, 38, 38, 0.12);
}
.hour-time {
  font-weight: 700;
  color: var(--accent);
}

/* Adjective stars 形容星 */
.strip-adj-star {
  font-size: var(--fs-2xs);
  color: var(--text-muted);
  padding: 0.05rem 0.25rem;
  background: rgba(109, 40, 217, 0.06);
  border-radius: 3px;
}

/* Twelve stars 十二星 */
.strip-twelve-stars {
  display: flex;
  gap: 0.25rem;
  flex-wrap: wrap;
  margin-top: 0.15rem;
}
.twelve-tag {
  font-size: var(--fs-2xs);
  padding: 0.04rem 0.25rem;
  border-radius: 3px;
}
.twelve-cs {
  color: #0f766e;
  background: rgba(15, 118, 110, 0.08);
}
.twelve-bs {
  color: #2563eb;
  background: rgba(37, 99, 235, 0.08);
}
.twelve-jq {
  color: #16a34a;
  background: rgba(22, 163, 74, 0.08);
}
.twelve-sq {
  color: #dc2626;
  background: rgba(220, 38, 38, 0.08);
}

/* Sanfang Sizheng 三方四正 */
.strip-sanfang {
  display: flex;
  align-items: center;
  gap: 0.3rem;
  margin-top: 0.15rem;
  font-size: var(--fs-2xs);
}
.sf-label {
  color: var(--text-muted);
  font-weight: 600;
}
.sf-item {
  color: #2563eb;
  padding: 0.03rem 0.2rem;
  background: rgba(37, 99, 235, 0.08);
  border-radius: 3px;
}
.hour-effect {
  color: var(--text);
  line-height: 1.3;
}
.hour-score {
  font-size: var(--fs-2xs);
  margin-top: 0.1rem;
}
.hour-good .hour-score {
  color: #16a34a;
}
.hour-neutral .hour-score {
  color: var(--text-muted);
}
.hour-bad .hour-score {
  color: #dc2626;
}

/* ── Dark-mode overrides ── */
:global(.dark) {
  .sihua-lu {
    background: rgba(74, 222, 128, 0.1);
    color: #4ade80;
  }
  .sihua-quan {
    background: rgba(203, 213, 225, 0.1);
    color: var(--accent);
  }
  .sihua-ke {
    background: rgba(96, 165, 250, 0.1);
    color: #60a5fa;
  }
  .sihua-ji {
    background: rgba(251, 113, 133, 0.1);
    color: var(--danger);
  }

  .star-chip {
    background: rgba(251, 113, 133, 0.08);
    border-color: rgba(251, 113, 133, 0.12);
  }
  .strip-sihua {
    background: rgba(251, 113, 133, 0.08);
  }
  .strip-adj-star {
    background: rgba(161, 130, 207, 0.06);
  }

  .twelve-cs {
    color: var(--accent);
    background: rgba(203, 213, 225, 0.08);
  }
  .twelve-bs {
    color: #60a5fa;
    background: rgba(96, 165, 250, 0.08);
  }
  .twelve-jq {
    color: #4ade80;
    background: rgba(74, 222, 128, 0.08);
  }
  .twelve-sq {
    color: #f08080;
    background: rgba(251, 113, 133, 0.08);
  }

  .sf-item {
    color: #2563eb;
    background: rgba(96, 165, 250, 0.08);
  }

  .fly-from {
    color: #93c5fd;
    background: rgba(96, 165, 250, 0.1);
  }
  .fly-chain {
    color: #86efac;
    background: rgba(74, 222, 128, 0.1);
  }
  .fly-affinity {
    color: #fde68a;
    background: rgba(253, 230, 138, 0.1);
  }

  .hour-good {
    background: rgba(74, 222, 128, 0.08);
    border-color: rgba(74, 222, 128, 0.15);
  }
  .hour-bad {
    background: rgba(251, 113, 133, 0.08);
    border-color: rgba(251, 113, 133, 0.15);
  }
  .hour-good .hour-score {
    color: #4ade80;
  }
  .hour-bad .hour-score {
    color: #f08080;
  }

  .score-good {
    background: rgba(74, 222, 128, 0.12);
    color: #4ade80;
  }
  .score-bad {
    background: rgba(251, 113, 133, 0.12);
    color: #f08080;
  }

  .interp-value.danger {
    color: #f08080;
  }
}

@media (max-width: 720px) {
  .chart-overview-panel {
    grid-template-columns: 1fr;
  }

  .palace-quick-grid {
    grid-template-columns: repeat(auto-fill, minmax(128px, 1fr));
  }
}
</style>
