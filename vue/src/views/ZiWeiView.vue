<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { fetchChart, type ChartDetail } from '../api/chart'
import { fetchZiWeiChart, fetchZiWeiOverlay, fetchZiWeiPeriod } from '../api/ziwei'
import ZiWeiInterpretation from '../components/ZiWeiInterpretation.vue'
import ZiWeiOverlay from '../components/ZiWeiOverlay.vue'

const route = useRoute()

interface BirthInfo {
  name: string
  gender: string
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
  impact: string
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
  confidence: number
}

interface PalaceReading {
  palaceName: string
  palaceFocus?: string
  summary?: string
  keyPoints?: string[]
  evidence?: ReadingEvidence[]
  sanfangContext?: SanfangContext | null
  patternDetails?: PatternDetail[]
  advice?: string[]
  riskFlags?: string[]
  confidence?: number
  mainStarAnalysis: SectionData
  auxStarInfluence: SectionData
  sihuaInfluence: SectionData
  sanFangSiZheng: SectionData
  patternAnnotations: SectionData
}

interface LiunianChartData {
  palaces: PalaceData[]
  year: number
}

// State
const loading = ref(true)
const error = ref('')
const activeTab = ref('mingpan')

const birthInfo = ref<BirthInfo>()
const chartData = ref<ZiWeiChartData>()
const selectedPalace = ref<PalaceReading | null>(null)

const dayunData = ref<any[]>([])
const liunianData = ref<any[]>([])
const liuyueData = ref<any[]>([])
const liuriData = ref<any>({})
const sihuaData = ref<any>({})
const sihuaChainData = ref<any>({})

// Interpretation data (loaded per year, not cached across year changes)
const liunianInterp = ref<Record<string, any>>({})
const liuyueInterp = ref<Record<string, any>>({})
const liuriInterp = ref<Record<string, any>>({})

const liunianOverlay = ref<LiunianChartData>()
const availableYears = ref<number[]>([])
const selectedLiunianYear = ref<number>(new Date().getFullYear())
const loadingTab = ref(false)

function mapBirthInfo(chart: ChartDetail): BirthInfo {
  const month = String(chart.birth_month).padStart(2, '0')
  const day = String(chart.birth_day).padStart(2, '0')
  const hour = String(chart.birth_hour).padStart(2, '0')
  const minute = String(chart.birth_min || 0).padStart(2, '0')
  return {
    name: chart.name || '未命名',
    gender: chart.gender,
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
    })

    chartData.value = data
    birthInfo.value = mapBirthInfo(chart)

    // Generate available years for overlay (current year ± 5)
    const currentYear = new Date().getFullYear()
    availableYears.value = Array.from({length: 11}, (_, i) => currentYear - 5 + i)

    // Load initial liunian overlay for the middle year (current year)
    await loadOverlay(availableYears.value[5])
  } catch (err: any) {
    if (err.response?.status === 404) {
      error.value = '该命盘不存在或已被删除，请重新创建。'
    } else {
      error.value = err.response?.data?.message || err.message || '加载命盘失败'
    }
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadZiWeiChart()
})

async function loadOverlay(year: number) {
  try {
    const chartId = route.params.chartId
    const data = await fetchZiWeiOverlay({
      chart_id: Number(chartId),
      year,
    })
    // Always update with fresh data keyed by year for caching
    liunianOverlay.value = { ...data, year }
  } catch {
    // Overlay data optional, don't block
  }
}

// Tab switching with data loading
async function switchTab(tab: string) {
  activeTab.value = tab
  loadingTab.value = true

  try {
    const chartId = route.params.chartId
    switch (tab) {
      case 'dayun':
        if (!dayunData.value.length) {
          const data = await fetchZiWeiPeriod({
            chart_id: Number(chartId),
            period_type: 'dayun',
          })
          dayunData.value = data.periods || []
        }
        break
      case 'liunian':
        {
          const year = selectedLiunianYear.value
          const data = await fetchZiWeiPeriod({
            chart_id: Number(chartId),
            period_type: 'liunian',
            year,
          })
          liunianData.value = data.periods || []
          // Also fetch interpretation
          const interpData = await fetchZiWeiPeriod({
            chart_id: Number(chartId),
            period_type: 'liunian_interpretation',
            year,
          })
          liunianInterp.value[year] = interpData.periods?.[0] || null
        }
        break
      case 'liuyue':
        {
          const year = selectedLiunianYear.value
          const month = new Date().getMonth() + 1
          const key = `${year}-${month}`
          if (!liuyueInterp.value[key]) {
            const data = await fetchZiWeiPeriod({
              chart_id: Number(chartId),
              period_type: 'liuyue',
              year,
              month,
            })
            liuyueData.value = data.periods || []
            const interpData = await fetchZiWeiPeriod({
              chart_id: Number(chartId),
              period_type: 'liuyue_interpretation',
              year,
              month,
            })
            liuyueInterp.value[key] = interpData.periods?.[0] || null
          }
        }
        break
      case 'liuri':
        {
          const year = selectedLiunianYear.value
          const month = new Date().getMonth() + 1
          const day = new Date().getDate()
          const key = `${year}-${month}-${day}`
          if (!liuriInterp.value[key]) {
            const data = await fetchZiWeiPeriod({
              chart_id: Number(chartId),
              period_type: 'liuri',
              year,
              month,
              day,
            })
            liuriData.value = data.periods || []
            const interpData = await fetchZiWeiPeriod({
              chart_id: Number(chartId),
              period_type: 'liuri_interpretation',
              year,
              month,
              day,
            })
            liuriInterp.value[key] = interpData.periods?.[0] || null
          }
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
  } catch (err: any) {
    // Tab data optional
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
      keyPoints: reading.key_points || [],
      evidence: reading.evidence || [],
      sanfangContext: reading.sanfang_context || null,
      patternDetails: reading.pattern_details || [],
      advice: reading.advice || [],
      riskFlags: reading.risk_flags || [],
      confidence: reading.confidence || 0,
      mainStarAnalysis: {
        title: '主星特性',
        content: reading.main_star_analysis || '',
        tags: palace.stars.filter(s => s.type === 'major').map(s => s.name),
      },
      auxStarInfluence: {
        title: '辅星影响',
        content: reading.aux_star_influence || '',
        tags: palace.stars.filter(s => s.type !== 'major').map(s => s.name),
      },
      sihuaInfluence: {
        title: '四化影响',
        content: reading.sihua_influence || '',
        tags: palace.four_hua || [],
      },
      sanFangSiZheng: {
        title: '三方四正',
        content: reading.sanfang_analysis || '',
        tags: [],
      },
      patternAnnotations: {
        title: '格局标注',
        content: reading.pattern_notes || '',
        tags: chartData.value?.patterns || [],
      },
    }
  } catch (e) {
    console.error('Failed to load palace reading:', e)
  }
}

function onYearChange(year: number) {
  loadOverlay(year)
}

const currentAge = computed(() => {
  if (!birthInfo.value) return 0
  const parts = birthInfo.value.solarDate.split('-')
  return new Date().getFullYear() - Number(parts[0])
})

function getPalacesFromPeriod(p: any) { return p?.palaces || [] }

function majorStars(p: any): StarInfo[] {
  return (p?.stars || []).filter((s: StarInfo) => s.type === 'major')
}
function auxStars(p: any): StarInfo[] {
  return (p?.stars || []).filter((s: StarInfo) => s.type !== 'major')
}

function palaceMajorSignal(p: PalaceData): string {
  const stars = (p?.stars || [])
    .filter(s => s.type === 'major')
    .map(s => s.brightness ? `${s.name}${s.brightness}` : s.name)
  if (!stars.length) return '空宫'
  return stars.slice(0, 3).join('、')
}

function palaceSupportSignal(p: PalaceData): string {
  const soft = (p?.stars || []).filter(s => ['soft', 'lucun', 'tianma'].includes(s.type)).length
  const tough = (p?.stars || []).filter(s => s.type === 'tough').length
  const parts = []
  if (soft) parts.push(`辅${soft}`)
  if (tough) parts.push(`煞${tough}`)
  return parts.length ? parts.join(' / ') : '辅煞少'
}

function palaceFourHuaLabel(p: PalaceData): string {
  const count = p?.four_hua?.length || 0
  return count ? `四化 ${count}` : '无四化'
}

function palaceFourHuaTitle(p: PalaceData): string {
  return p?.four_hua?.length ? p.four_hua.join('、') : '本宫无四化'
}

const chartOverviewItems = computed(() => {
  const chart = chartData.value
  if (!chart) return []
  return [
    { label: '命主', value: chart.life_master || '未定', hint: '先天性格取象' },
    { label: '身主', value: chart.body_master || '未定', hint: '后天行动取象' },
    { label: '五行局', value: chart.five_bureau || '未定', hint: '命盘局数底色' },
    { label: '身宫', value: chart.body_palace || chart.earthly_branch_of_body_palace || '未定', hint: '现实行动重心' },
  ]
})

const chartPatternPreview = computed(() => {
  const patterns = chartData.value?.patterns || []
  return patterns.slice(0, 6)
})

const currentLiuyueInterp = computed(() => {
  if (!liuyueData.value[0]) return null
  const key = `${selectedLiunianYear.value}-${liuyueData.value[0].month}`
  return liuyueInterp.value[key] || null
})

const currentLiuriInterp = computed(() => {
  if (!liuriData.value[0]) return null
  const d = liuriData.value[0]
  const key = `${selectedLiunianYear.value}-${d.month}-${d.day}`
  return liuriInterp.value[key] || null
})

const sihuaFlyGroups = computed(() => {
  const data = sihuaData.value as any
  if (!data || !data.hua_lu) return []
  return [
    { type: '化禄', css: 'sihua-lu', items: data.hua_lu || [] },
    { type: '化权', css: 'sihua-quan', items: data.hua_quan || [] },
    { type: '化科', css: 'sihua-ke', items: data.hua_ke || [] },
    { type: '化忌', css: 'sihua-ji', items: data.hua_ji || [] },
  ]
})

const sihuaChainGroups = computed(() => {
  const chain = sihuaChainData.value as any
  if (!chain || !chain.hua_lu) return []
  return [
    { type: '化禄', css: 'sihua-lu', items: chain.hua_lu || [] },
    { type: '化权', css: 'sihua-quan', items: chain.hua_quan || [] },
    { type: '化科', css: 'sihua-ke', items: chain.hua_ke || [] },
    { type: '化忌', css: 'sihua-ji', items: chain.hua_ji || [] },
  ]
})
</script>

<template>
  <div class="ziwei-page">
    <!-- Loading -->
    <div v-if="loading" class="loading-state">
      <el-skeleton animated>
        <template #template>
          <div style="display: grid; grid-template-columns: repeat(4, 1fr); gap: 8px; max-width: 400px; margin: 0 auto;">
            <el-skeleton-item v-for="i in 12" :key="i" variant="rect" style="aspect-ratio: 1; border-radius: 8px;" />
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
        </div>
        <router-link
          v-if="birthInfo"
          :to="'/chart/' + birthInfo.baziChartId"
          class="bazi-link"
        >
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
          :available-years="availableYears"
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
                <div v-if="chartPatternPreview.length" class="overview-pattern-list">
                  <span v-for="pattern in chartPatternPreview" :key="pattern">{{ pattern }}</span>
                  <span v-if="(chartData.patterns?.length || 0) > chartPatternPreview.length" class="more-patterns">
                    +{{ (chartData.patterns?.length || 0) - chartPatternPreview.length }}
                  </span>
                </div>
                <p v-else>暂无可由当前规则直接验证的格局标签</p>
              </div>
            </section>

            <div class="palace-quick-grid">
              <button
                v-for="(palace, idx) in (chartData?.palaces || [])"
                :key="palace.branch"
                class="palace-pill"
                :class="{
                  active: selectedPalace?.palaceName === palace.name,
                  'body-palace': chartData && (palace.name === chartData.body_palace || palace.is_body_palace)
                }"
                @click="onPalaceClick(palace, idx)"
              >
                <span class="palace-pill-name">{{ palace.name }}</span>
                <span class="palace-pill-branch">{{ palace.branch }}</span>
                <span class="palace-pill-stars">{{ palaceMajorSignal(palace) }}</span>
                <span class="palace-pill-meta" :title="palaceFourHuaTitle(palace)">{{ palaceFourHuaLabel(palace) }}</span>
                <span class="palace-pill-meta support">{{ palaceSupportSignal(palace) }}</span>
                <span v-if="chartData && (palace.name === chartData.body_palace || palace.is_body_palace)" class="body-badge">身</span>
              </button>
            </div>

            <ZiWeiInterpretation v-if="selectedPalace" :palace-reading="selectedPalace" />
            <div v-else class="empty-state-inline">
              <svg width="40" height="40" viewBox="0 0 40 40" fill="none">
                <circle cx="20" cy="20" r="16" stroke="currentColor" stroke-width="0.5" stroke-dasharray="2 3" opacity="0.3"/>
                <circle cx="20" cy="20" r="3" fill="currentColor" opacity="0.2"/>
              </svg>
              <p>选择一个宫位查看详细解读</p>
            </div>
          </div>

          <!-- 大限分析 -->
          <div v-else-if="activeTab === 'dayun'" class="data-tab">
            <p class="tab-desc">人生各阶段十年大限，展示每阶段主要星曜和宫位变化</p>
            <div v-if="loadingTab" class="tab-loading"><div class="loading-dots"><span></span><span></span><span></span></div></div>
            <div v-else-if="!dayunData.length" class="empty-state-inline"><p>暂无可显示的大限数据</p></div>
            <div v-else class="dayun-timeline">
              <div v-for="(item, idx) in dayunData" :key="idx" class="dayun-card" :class="{ 'is-current': item.start_age <= currentAge && item.end_age >= currentAge }">
                <div class="dayun-age-badge">
                  <span class="age-primary">{{ item.start_age }}–{{ item.end_age }}</span>
                  <span class="age-unit">岁</span>
                </div>
                <div class="dayun-body">
                  <div class="dayun-palace">{{ item.palace_name || item.palace }}</div>
                  <p class="dayun-desc">{{ item.description }}</p>
                  <div v-if="item.stars?.length" class="dayun-stars"><span v-for="s in item.stars" :key="s" class="star-chip">{{ s }}</span></div>
                </div>
              </div>
            </div>
          </div>

          <!-- 流年分析 -->
          <div v-else-if="activeTab === 'liunian'" class="data-tab">
            <p class="tab-desc">{{ liunianData[0]?.year ? liunianData[0].year + '年' : '' }}流年各宫星曜分布，每年依次轮换</p>
            <div v-if="loadingTab" class="tab-loading"><div class="loading-dots"><span></span><span></span><span></span></div></div>
            <div v-else-if="!liunianData.length" class="empty-state-inline"><p>暂无可显示的流年数据</p></div>
            <div v-else>
              <div v-for="p in getPalacesFromPeriod(liunianData[0])" :key="p.branch" class="palace-strip">
                <div class="palace-strip-header"><span class="strip-name">{{ p.name }}</span><span class="strip-branch">{{ p.branch }}</span></div>
                <div class="palace-strip-stars">
                  <template v-if="majorStars(p).length"><span v-for="s in majorStars(p)" :key="s.name" class="strip-main-star" :class="{ dim: !s.brightness }">{{ s.name }}<small v-if="s.brightness">·{{s.brightness}}</small></span></template>
                  <span v-if="!majorStars(p).length" class="strip-empty">无主星</span>
                  <template v-if="auxStars(p).length"><span v-for="s in auxStars(p).slice(0,4)" :key="s.name" class="strip-aux-star">{{ s.name }}</span></template>
                  <template v-if="p.adjective_stars?.length"><span v-for="s in p.adjective_stars" :key="s" class="strip-adj-star">{{ s }}</span></template>
                  <template v-if="p.four_hua?.length"><span v-for="s in p.four_hua" :key="s" class="strip-sihua">{{ s }}</span></template>
                </div>
                <div v-if="p.changsheng_12 || p.boshi_12 || p.jiang_qian_12 || p.sui_qian_12" class="strip-twelve-stars">
                  <span v-if="p.changsheng_12" class="twelve-tag twelve-cs">{{ p.changsheng_12 }}</span>
                  <span v-if="p.boshi_12" class="twelve-tag twelve-bs">{{ p.boshi_12 }}</span>
                  <span v-if="p.jiang_qian_12" class="twelve-tag twelve-jq">{{ p.jiang_qian_12 }}</span>
                  <span v-if="p.sui_qian_12" class="twelve-tag twelve-sq">{{ p.sui_qian_12 }}</span>
                </div>
                <div v-if="p.sanfang_sizheng" class="strip-sanfang">
                  <span class="sf-label">三方四正</span>
                  <span v-if="p.sanfang_sizheng.opposite" class="sf-item">对{{ p.sanfang_sizheng.opposite }}</span>
                  <span v-if="p.sanfang_sizheng.trine1" class="sf-item">三{{ p.sanfang_sizheng.trine1 }}</span>
                  <span v-if="p.sanfang_sizheng.trine2" class="sf-item">三{{ p.sanfang_sizheng.trine2 }}</span>
                </div>
              </div>
            </div>
            <!-- 流年详解析 -->
            <div v-if="liunianInterp[selectedLiunianYear]" class="interp-card">
              <div class="interp-header">
                <span class="interp-year">{{ selectedLiunianYear }}年</span>
                <span class="interp-ganZhi">{{ liunianInterp[selectedLiunianYear].gan_zhi }}</span>
                <div class="interp-score" :class="liunianInterp[selectedLiunianYear].score >= 60 ? 'score-good' : 'score-bad'">{{ liunianInterp[selectedLiunianYear].score }}分</div>
              </div>
              <div class="interp-section">
                <div class="interp-row"><span class="interp-label">干支释义</span><span class="interp-value">{{ liunianInterp[selectedLiunianYear].gan_zhi_desc }}</span></div>
                <div class="interp-row"><span class="interp-label">十神</span><span class="interp-value">{{ liunianInterp[selectedLiunianYear].shi_shen }}</span></div>
                <div class="interp-row"><span class="interp-label">与命局关系</span><span class="interp-value danger">{{ liunianInterp[selectedLiunianYear].relation_to_ming }}</span></div>
                <div class="interp-row"><span class="interp-label">全年基调</span><span class="interp-value">{{ liunianInterp[selectedLiunianYear].overall_tone }}</span></div>
                <div class="interp-row tip"><span class="interp-label">重点提示</span><span class="interp-value">{{ liunianInterp[selectedLiunianYear].key_tips }}</span></div>
              </div>
            </div>
          </div>

          <!-- 流月分析 -->
          <div v-else-if="activeTab === 'liuyue'" class="data-tab">
            <p class="tab-desc">{{ liuyueData[0]?.year ? liuyueData[0].year + '年' + liuyueData[0].month + '月' : '' }}流月各宫星曜分布，每月依次轮换</p>
            <div v-if="loadingTab" class="tab-loading"><div class="loading-dots"><span></span><span></span><span></span></div></div>
            <div v-else-if="!liuyueData.length" class="empty-state-inline"><p>暂无可显示的流月数据</p></div>
            <div v-else>
              <div v-for="p in getPalacesFromPeriod(liuyueData[0])" :key="'ly-' + p.branch" class="palace-strip">
                <div class="palace-strip-header"><span class="strip-name">{{ p.name }}</span><span class="strip-branch">{{ p.branch }}</span></div>
                <div class="palace-strip-stars">
                  <template v-if="majorStars(p).length"><span v-for="s in majorStars(p)" :key="s.name" class="strip-main-star" :class="{ dim: !s.brightness }">{{ s.name }}<small v-if="s.brightness">·{{s.brightness}}</small></span></template>
                  <span v-if="!majorStars(p).length" class="strip-empty">无主星</span>
                  <template v-if="auxStars(p).length"><span v-for="s in auxStars(p).slice(0,4)" :key="s.name" class="strip-aux-star">{{ s.name }}</span></template>
                  <template v-if="p.adjective_stars?.length"><span v-for="s in p.adjective_stars" :key="s" class="strip-adj-star">{{ s }}</span></template>
                  <template v-if="p.four_hua?.length"><span v-for="s in p.four_hua" :key="s" class="strip-sihua">{{ s }}</span></template>
                </div>
                <div v-if="p.changsheng_12 || p.boshi_12 || p.jiang_qian_12 || p.sui_qian_12" class="strip-twelve-stars">
                  <span v-if="p.changsheng_12" class="twelve-tag twelve-cs">{{ p.changsheng_12 }}</span>
                  <span v-if="p.boshi_12" class="twelve-tag twelve-bs">{{ p.boshi_12 }}</span>
                  <span v-if="p.jiang_qian_12" class="twelve-tag twelve-jq">{{ p.jiang_qian_12 }}</span>
                  <span v-if="p.sui_qian_12" class="twelve-tag twelve-sq">{{ p.sui_qian_12 }}</span>
                </div>
                <div v-if="p.sanfang_sizheng" class="strip-sanfang">
                  <span class="sf-label">三方四正</span>
                  <span v-if="p.sanfang_sizheng.opposite" class="sf-item">对{{ p.sanfang_sizheng.opposite }}</span>
                  <span v-if="p.sanfang_sizheng.trine1" class="sf-item">三{{ p.sanfang_sizheng.trine1 }}</span>
                  <span v-if="p.sanfang_sizheng.trine2" class="sf-item">三{{ p.sanfang_sizheng.trine2 }}</span>
                </div>
              </div>
            </div>
            <!-- 流月详解析 -->
            <div v-if="liuyueData[0]" class="interp-card">
              <div class="interp-header">
                <span class="interp-year">{{ liuyueData[0].year }}年{{ liuyueData[0].month }}月</span>
                <span class="interp-ganZhi">{{ currentLiuyueInterp?.gan_zhi || '—' }}</span>
                <div class="interp-score" :class="(currentLiuyueInterp?.score || 0) >= 60 ? 'score-good' : 'score-bad'">{{ currentLiuyueInterp?.score || '—' }}分</div>
              </div>
              <div class="interp-section">
                <div class="interp-row"><span class="interp-label">干支释义</span><span class="interp-value">{{ currentLiuyueInterp?.gan_zhi_desc || '—' }}</span></div>
                <div class="interp-row"><span class="interp-label">十神</span><span class="interp-value">{{ currentLiuyueInterp?.shi_shen || '—' }}</span></div>
                <div class="interp-row"><span class="interp-label">与命局关系</span><span class="interp-value danger">{{ currentLiuyueInterp?.relation_to_ming || '—' }}</span></div>
                <div class="interp-row"><span class="interp-label">作用特点</span><span class="interp-value">{{ currentLiuyueInterp?.effect || '—' }}</span></div>
                <div class="interp-row"><span class="interp-label">健康提示</span><span class="interp-value">{{ currentLiuyueInterp?.health || '—' }}</span></div>
              </div>
            </div>
          </div>

          <!-- 流日分析 -->
          <div v-else-if="activeTab === 'liuri'" class="data-tab">
            <p class="tab-desc">{{ liuriData[0]?.year ? liuriData[0].year + '年' + liuriData[0].month + '月' + liuriData[0].day + '日' : '' }}流日各宫星曜分布，每日依次轮换</p>
            <div v-if="loadingTab" class="tab-loading"><div class="loading-dots"><span></span><span></span><span></span></div></div>
            <div v-else-if="!liuriData.length" class="empty-state-inline"><p>暂无可显示的流日数据</p></div>
            <div v-else>
              <div v-for="p in getPalacesFromPeriod(liuriData[0])" :key="'lr-' + p.branch" class="palace-strip palace-strip-sm">
                <div class="palace-strip-header"><span class="strip-name">{{ p.name }}</span><span class="strip-branch">{{ p.branch }}</span></div>
<div class="palace-strip-stars">
                  <template v-if="majorStars(p).length"><span v-for="s in majorStars(p)" :key="s.name" class="strip-main-star" :class="{ dim: !s.brightness }">{{ s.name }}<small v-if="s.brightness">·{{s.brightness}}</small></span></template>
                  <span v-if="!majorStars(p).length" class="strip-empty">无主星</span>
                  <template v-if="auxStars(p).length"><span v-for="s in auxStars(p).slice(0,4)" :key="s.name" class="strip-aux-star">{{ s.name }}</span></template>
                  <template v-if="p.adjective_stars?.length"><span v-for="s in p.adjective_stars" :key="s" class="strip-adj-star">{{ s }}</span></template>
                  <template v-if="p.four_hua?.length"><span v-for="s in p.four_hua" :key="s" class="strip-sihua">{{ s }}</span></template>
                </div>
                <div v-if="p.changsheng_12 || p.boshi_12 || p.jiang_qian_12 || p.sui_qian_12" class="strip-twelve-stars">
                  <span v-if="p.changsheng_12" class="twelve-tag twelve-cs">{{ p.changsheng_12 }}</span>
                  <span v-if="p.boshi_12" class="twelve-tag twelve-bs">{{ p.boshi_12 }}</span>
                  <span v-if="p.jiang_qian_12" class="twelve-tag twelve-jq">{{ p.jiang_qian_12 }}</span>
                  <span v-if="p.sui_qian_12" class="twelve-tag twelve-sq">{{ p.sui_qian_12 }}</span>
                </div>
                <div v-if="p.sanfang_sizheng" class="strip-sanfang">
                  <span class="sf-label">三方四正</span>
                  <span v-if="p.sanfang_sizheng.opposite" class="sf-item">对{{ p.sanfang_sizheng.opposite }}</span>
                  <span v-if="p.sanfang_sizheng.trine1" class="sf-item">三{{ p.sanfang_sizheng.trine1 }}</span>
                  <span v-if="p.sanfang_sizheng.trine2" class="sf-item">三{{ p.sanfang_sizheng.trine2 }}</span>
                </div>
              </div>
            </div>
            <!-- 流日详解析 -->
            <div v-if="liuriData[0]" class="interp-card">
              <div class="interp-header">
                <span class="interp-year">{{ liuriData[0].year }}年{{ liuriData[0].month }}月{{ liuriData[0].day }}日</span>
                <span class="interp-ganZhi">{{ currentLiuriInterp?.gan_zhi || '—' }}</span>
                <div class="interp-score" :class="(currentLiuriInterp?.score || 0) >= 60 ? 'score-good' : 'score-bad'">{{ currentLiuriInterp?.score || '—' }}分</div>
              </div>
              <div class="interp-section">
                <div class="interp-row"><span class="interp-label">干支释义</span><span class="interp-value">{{ currentLiuriInterp?.gan_zhi_desc || '—' }}</span></div>
                <div class="interp-row"><span class="interp-label">十神</span><span class="interp-value">{{ currentLiuriInterp?.shi_shen || '—' }}</span></div>
                <div class="interp-row"><span class="interp-label">与命局关系</span><span class="interp-value danger">{{ currentLiuriInterp?.relation_to_ming || '—' }}</span></div>
                <div class="interp-row"><span class="interp-label">七杀作用</span><span class="interp-value">{{ currentLiuriInterp?.qi_zi_effect || '—' }}</span></div>
                <div class="interp-row"><span class="interp-label">情绪状态</span><span class="interp-value">{{ currentLiuriInterp?.emotional_state || '—' }}</span></div>
                <div class="interp-row"><span class="interp-label">健康提示</span><span class="interp-value">{{ currentLiuriInterp?.health || '—' }}</span></div>
              </div>
              <div class="interp-section">
                <p class="interp-subtitle">时辰分析</p>
                <div class="hourly-grid">
                  <div v-for="(h, i) in (currentLiuriInterp?.hourly_analysis || [])" :key="i" class="hour-block" :class="h.score >= 65 ? 'hour-good' : h.score < 45 ? 'hour-bad' : 'hour-neutral'">
                    <span class="hour-time">{{ h.stem_branch }}</span>
                    <span class="hour-effect">{{ h.effect }}</span>
                    <span class="hour-score">{{ h.score }}分</span>
                  </div>
                </div>
              </div>
              <div class="interp-summary">{{ currentLiuriInterp?.summary || '—' }}</div>
            </div>
          </div>

          <!-- 四化飞星 -->
          <div v-else-if="activeTab === 'sihua'" class="data-tab">
            <p class="tab-desc">四化飞星在各宫的分布，展示星曜化禄/化权/化科/化忌的飞入情况及链式分析</p>
            <div v-if="loadingTab" class="tab-loading"><div class="loading-dots"><span></span><span></span><span></span></div></div>
            <div v-else-if="!sihuaFlyGroups.length" class="empty-state-inline"><p>暂无可显示的四化飞星数据</p></div>
            <div v-else class="sihua-groups">
              <div v-for="grp in sihuaFlyGroups" :key="grp.type" class="sihua-group">
                <span class="sihua-group-badge" :class="grp.css">{{ grp.type }}</span>
                <div v-if="grp.items.length" class="sihua-group-items">
                  <div v-for="(it, i) in grp.items" :key="i" class="sihua-fly-item">
                    <span class="fly-star">{{ it.from_star }}</span><span class="fly-arrow">→</span>
                    <span class="fly-palace">{{ it.to_palace }}</span>
                    <span v-if="it.from_palace" class="fly-from">源{{ it.from_palace }}</span>
                    <span v-if="it.chain_depth > 0" class="fly-chain">链{{ it.chain_depth }}</span>
                    <span class="fly-effect">{{ it.effect }}</span>
                  </div>
                </div>
                <p v-else class="sihua-empty-group">无</p>
              </div>
              <div v-if="sihuaChainGroups.length" class="sihua-chain-section">
                <h4 class="chain-title">四化链式分析</h4>
                <div v-for="grp in sihuaChainGroups" :key="'chain-' + grp.type" class="sihua-group">
                  <span class="sihua-group-badge" :class="grp.css">{{ grp.type }}</span>
                  <div v-if="grp.items.length" class="sihua-group-items">
                    <div v-for="(it, i) in grp.items" :key="'c-' + i" class="sihua-fly-item">
                      <span class="fly-star">{{ it.from_star }}</span><span class="fly-arrow">→</span>
                      <span class="fly-palace">{{ it.to_palace }}</span>
                      <span v-if="it.from_palace" class="fly-from">源{{ it.from_palace }}</span>
                      <span v-if="it.chain_depth > 0" class="fly-chain">链{{ it.chain_depth }}</span>
                      <span v-if="it.star_affinity > 0" class="fly-affinity">辅{{ it.star_affinity }}</span>
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
  width:2rem;height:2rem;border-radius:50%;border:2px dashed;animation:spin 1s linear infinite;
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

.tab-hint {
  font-size: 0.75rem; color: var(--text-muted); margin-bottom: 0.75rem;
}

.tab-loading {
  text-align: center; font-size: 0.875rem; color: var(--text-muted); padding-top: 2rem; padding-bottom: 2rem;
}

.empty-hint {
  display: flex; flex-direction: column; align-items: center; gap: 0.5rem; padding-top: 2.5rem; padding-bottom: 2.5rem; color: var(--text-muted);
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
  font-size: 0.58rem;
  font-weight: 700;
  color: var(--text-muted);
}

.overview-item strong {
  display: block;
  margin-top: 0.18rem;
  font-size: 0.86rem;
  color: var(--accent);
}

.overview-item small {
  display: block;
  margin-top: 0.12rem;
  font-size: 0.56rem;
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
  font-size: 0.58rem;
  font-weight: 600;
  overflow-wrap: anywhere;
}

.overview-pattern-list .more-patterns {
  color: var(--text-muted);
}

.overview-patterns p {
  margin: 0.45rem 0 0;
  font-size: 0.62rem;
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
  cursor: pointer; transition: all 0.2s;
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
  font-size: 0.75rem; font-weight: 600; color: var(--text);
  letter-spacing: 0;
}
.palace-pill-branch {
  font-size: 0.58rem; color: var(--text-muted); text-align:right;
}
.palace-pill-stars {
  grid-column: 1 / -1;
  min-height: 1rem;
  font-size: 0.64rem;
  line-height: 1.25;
  color: var(--accent);
  overflow-wrap: anywhere;
}
.palace-pill-meta {
  justify-self: start;
  font-size: 0.55rem;
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
  font-size: 0.5rem; background: rgba(251, 113, 133, 0.12); color: var(--danger);
  padding: 0.05rem 0.25rem; border-radius: 3px; font-weight: 600;
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
  font-size: 0.75rem; color: var(--text-muted); margin: 0;
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
  background: rgba(22,163,74,0.1);
  color: #16a34a;
}

.sihua-quan {
  background-color: var(--line-strong);
  color: var(--color-bazi-blue);
}

.sihua-ke {
  background-color: rgba(37,99,235,0.1);
  color: #2563eb;
}

.sihua-ji {
  background-color: rgba(220,38,38,0.08);
  color: #dc2626;
}

.tab-desc { font-size:0.72rem; color:var(--text-muted); margin:0 0 1rem; font-style:italic; }
.dayun-timeline { display:flex; flex-direction:column; gap:0.5rem; }
.dayun-card { display:flex; gap:0.75rem; padding:0.625rem 0.75rem; background:color-mix(in oklab, var(--accent) 4%, transparent); border:1px solid var(--line-subtle); border-radius:8px; }
.dayun-card.is-current { border-color:var(--line-focus); background:var(--line-subtle); }
.dayun-age-badge { min-width:60px; text-align:center; }
.age-primary { display:block; font-size:0.8rem; font-weight:700; color:var(--accent); }
.age-unit { font-size:0.58rem; color:var(--text-muted); }
.dayun-body { flex:1; }
.dayun-palace { font-size:0.82rem; font-weight:600; color:var(--text); margin-bottom:0.125rem; }
.dayun-desc { font-size:0.68rem; color:var(--text-muted); margin:0 0 0.25rem; }
.dayun-stars { display:flex; flex-wrap:wrap; gap:0.2rem; }
.star-chip { padding:0.08rem 0.35rem; font-size:0.62rem; background:rgba(220,38,38,0.08); color:var(--danger); border-radius:3px; border:1px solid rgba(220,38,38,0.12); }
.palace-strip { display:flex; align-items:center; gap:0.6rem; padding:0.4rem 0.6rem; border-bottom:1px solid var(--line-subtle); }
.palace-strip-sm { padding:0.25rem 0.4rem; }
.palace-strip-header { display:flex; align-items:center; gap:0.3rem; min-width:68px; }
.strip-name { font-size:0.72rem; font-weight:600; color:var(--text); }
.strip-branch { font-size:0.58rem; color:var(--text-muted); }
.palace-strip-stars { display:flex; align-items:center; gap:0.25rem; flex-wrap:wrap; flex:1; }
.strip-main-star { padding:0.06rem 0.3rem; font-size:0.62rem; font-weight:600; background:var(--line-strong); color:var(--accent); border-radius:3px; }
.strip-main-star.dim { color:var(--text-muted); background:var(--accent-dim); }
.strip-main-star small { font-size:0.5rem; opacity:0.5; }
.strip-aux-star { font-size:0.58rem; color: var(--text-soft); }
.strip-sihua { padding:0.06rem 0.3rem; font-size:0.58rem; background:rgba(220,38,38,0.08); color:var(--danger); border-radius:3px; }
.strip-empty { font-size:0.58rem; color:var(--text-muted); opacity:0.3; }
.sihua-groups { display:flex; flex-direction:column; gap:0.625rem; }
.sihua-group-badge { display:inline-block; padding:0.15rem 0.5rem; font-size:0.68rem; font-weight:700; border-radius:4px; margin-bottom:0.25rem; }
.sihua-lu { background:rgba(22,163,74,0.1); color:#16a34a; }
.sihua-quan { background:rgba(71,85,105,0.1); color:#475569; }
.sihua-ke { background:rgba(37,99,235,0.1); color:#2563eb; }
.sihua-ji { background:rgba(220,38,38,0.08); color:#dc2626; }
.sihua-group-items { display:flex; flex-direction:column; gap:0.2rem; }
.sihua-fly-item { display:flex; align-items:center; gap:0.4rem; padding:0.3rem 0.5rem; background:color-mix(in oklab, var(--accent) 3%, transparent); border-radius:5px; font-size:0.72rem; }
.fly-star { font-weight:600; color:var(--text); }
.fly-arrow { color:var(--text-muted); font-size:0.65rem; }
.fly-palace { color:var(--accent); font-weight:500; }
.fly-effect { color:var(--text-muted); font-size:0.65rem; flex:1; }
.fly-from { font-size:0.6rem; color:#2563eb; background:rgba(37,99,235,0.08); padding:0.05rem 0.25rem; border-radius:3px; }
.fly-chain { font-size:0.6rem; color:#16a34a; background:rgba(22,163,74,0.08); padding:0.05rem 0.25rem; border-radius:3px; }
.fly-affinity { font-size:0.6rem; color:#a16207; background:rgba(161,98,7,0.08); padding:0.05rem 0.25rem; border-radius:3px; }
.sihua-chain-section { margin-top:1rem; padding-top:0.75rem; border-top:1px solid var(--line-subtle); }
.chain-title { font-size:0.72rem; color:var(--text-muted); margin:0 0 0.5rem; font-weight:600; }
.sihua-empty-group { font-size:0.68rem; color:var(--text-muted); padding:0.2rem 0.4rem; opacity:0.4; }

/* Interpretation tab styles */
.interp-tab { display:flex; flex-direction:column; gap:0.75rem; }
.interp-card { background:var(--glass); border:1px solid var(--line-strong); border-radius:12px; overflow:hidden; }
.interp-header { display:flex; align-items:center; gap:0.75rem; padding:0.75rem 1rem; background:var(--line-subtle); border-bottom:1px solid var(--line-strong); }
.interp-year { font-size:0.9rem; font-weight:700; color:var(--accent); font-family:var(--font-serif); }
.interp-ganZhi { font-size:0.85rem; color:var(--text); font-weight:600; }
.interp-score { font-size:0.8rem; font-weight:700; padding:0.15rem 0.5rem; border-radius:6px; margin-left:auto; }
.score-good { background:rgba(22,163,74,0.1); color:#16a34a; }
.score-bad { background:rgba(220,38,38,0.08); color:#dc2626; }
.interp-section { padding:0.75rem 1rem; display:flex; flex-direction:column; gap:0.4rem; }
.interp-row { display:flex; gap:0.5rem; font-size:0.78rem; line-height:1.5; }
.interp-label { min-width:70px; font-weight:600; color:var(--text-muted); }
.interp-value { color:var(--text); flex:1; }
.interp-value.danger { color:#dc2626; }
.interp-row.tip .interp-value { color:var(--accent); font-style:italic; }
.interp-subtitle { font-size:0.75rem; font-weight:700; color:var(--accent); margin:0 0 0.5rem; }
.interp-summary { font-size:0.72rem; color:var(--text-muted); padding:0.5rem 1rem; background:var(--line-subtle); border-top:1px dashed var(--line-strong); font-style:italic; }
.hourly-grid { display:grid; grid-template-columns:repeat(auto-fill,minmax(110px,1fr)); gap:0.375rem; }
.hour-block { display:flex; flex-direction:column; gap:0.15rem; padding:0.4rem 0.5rem; border-radius:6px; font-size:0.68rem; }
.hour-good { background:rgba(22,163,74,0.08); border:1px solid rgba(22,163,74,0.15); }
.hour-neutral { background:var(--line-subtle); border:1px solid var(--line-strong); }
.hour-bad { background:rgba(220,38,38,0.06); border:1px solid rgba(220,38,38,0.12); }
.hour-time { font-weight:700; color:var(--accent); }

/* Adjective stars 形容星 */
.strip-adj-star { font-size:0.56rem; color:var(--text-muted); padding:0.05rem 0.25rem; background:rgba(109,40,217,0.06); border-radius:3px; }

/* Twelve stars 十二星 */
.strip-twelve-stars { display:flex; gap:0.25rem; flex-wrap:wrap; margin-top:0.15rem; }
.twelve-tag { font-size:0.52rem; padding:0.04rem 0.25rem; border-radius:3px; }
.twelve-cs { color: #0f766e; background:rgba(15,118,110,0.08); }
.twelve-bs { color:#2563eb; background:rgba(37,99,235,0.08); }
.twelve-jq { color:#16a34a; background:rgba(22,163,74,0.08); }
.twelve-sq { color:#dc2626; background:rgba(220,38,38,0.08); }

/* Sanfang Sizheng 三方四正 */
.strip-sanfang { display:flex; align-items:center; gap:0.3rem; margin-top:0.15rem; font-size:0.55rem; }
.sf-label { color:var(--text-muted); font-weight:600; }
.sf-item { color:#2563eb; padding:0.03rem 0.2rem; background:rgba(37,99,235,0.08); border-radius:3px; }
.hour-effect { color:var(--text); line-height:1.3; }
.hour-score { font-size:0.62rem; margin-top:0.1rem; }
.hour-good .hour-score { color:#16a34a; }
.hour-neutral .hour-score { color:var(--text-muted); }
.hour-bad .hour-score { color:#dc2626; }

/* ── Dark-mode overrides ── */
:global(.dark) {
  .sihua-lu { background:rgba(74,222,128,0.1); color:#4ade80; }
  .sihua-quan { background:rgba(203, 213, 225,0.1); color:var(--accent); }
  .sihua-ke { background:rgba(96,165,250,0.1); color:#60a5fa; }
  .sihua-ji { background:rgba(251, 113, 133,0.1); color:var(--danger); }

  .star-chip { background:rgba(251, 113, 133,0.08); border-color:rgba(251, 113, 133,0.12); }
  .strip-sihua { background:rgba(251, 113, 133,0.08); }
  .strip-adj-star { background:rgba(161,130,207,0.06); }

  .twelve-cs { color: var(--accent); background:rgba(203, 213, 225,0.08); }
  .twelve-bs { color:#60a5fa; background:rgba(96,165,250,0.08); }
  .twelve-jq { color:#4ade80; background:rgba(74,222,128,0.08); }
  .twelve-sq { color:#f08080; background:rgba(251, 113, 133,0.08); }

  .sf-item { color:#2563eb; background:rgba(96,165,250,0.08); }

  .fly-from { color:#93c5fd; background:rgba(96,165,250,0.1); }
  .fly-chain { color:#86efac; background:rgba(74,222,128,0.1); }
  .fly-affinity { color:#fde68a; background:rgba(253,230,138,0.1); }

  .hour-good { background:rgba(74,222,128,0.08); border-color:rgba(74,222,128,0.15); }
  .hour-bad { background:rgba(251, 113, 133,0.08); border-color:rgba(251, 113, 133,0.15); }
  .hour-good .hour-score { color:#4ade80; }
  .hour-bad .hour-score { color:#f08080; }

  .score-good { background:rgba(74,222,128,0.12); color:#4ade80; }
  .score-bad { background:rgba(251, 113, 133,0.12); color:#f08080; }

  .interp-value.danger { color:#f08080; }
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
