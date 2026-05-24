<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import client from '../api/client'
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
  brightness: string
}

interface PalaceData {
  branch: string
  name: string
  mainStars: StarInfo[]
  auxStars: StarInfo[]
  sihua: string[]
}

interface ZiWeiChartData {
  palaces: PalaceData[]
  mingZhu: string
  shenZhu: string
  wuxingJu: string
  patterns: string[]
  bodyPalace?: string
}

interface SectionData {
  title: string
  content: string
  tags: string[]
}

interface PalaceReading {
  palaceName: string
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

// Load chart data
async function loadZiWeiChart() {
  loading.value = true
  error.value = ''
  try {
    const chartId = route.params.chartId
    // First fetch the chart to get birth info
    const chartResp = await client.get(`/charts/${chartId}`)
    const chart = chartResp.data.chart || chartResp.data
    if (!chart || !chart.birth_year) {
      error.value = '未找到命盘数据，请先生成八字命盘后再查看紫微斗数。'
      loading.value = false
      return
    }
    // Then calculate ziwei from birth info
    const resp = await client.post('/ziwei/chart', {
      birth_year: chart.birth_year,
      birth_month: chart.birth_month,
      birth_day: chart.birth_day,
      birth_hour: chart.birth_hour,
      birth_min: chart.birth_minute || 0,
      calendar_type: chart.calendar_type || 'SOLAR',
      gender: chart.gender || 'MALE',
    })
    const data = resp.data

    const chartValues = data
    chartValues.mingZhu = data.mingZhu || data.life_master
    chartValues.shenZhu = data.shenZhu || data.body_master
    chartValues.wuxingJu = data.wuxingJu || data.five_bureau || data.fiveBureau || ''
    chartData.value = chartValues  // backend returns ZiWeiChart directly
    birthInfo.value = chart  // chart from /charts/:id has birth info

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
    const resp = await client.post('/ziwei/overlay', {
      chart_id: Number(chartId),
      year,
    })
    // Always update with fresh data keyed by year for caching
    liunianOverlay.value = { ...resp.data, year }
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
          const resp = await client.post('/ziwei/period', {
            chart_id: Number(chartId),
            period_type: 'dayun',
          })
          dayunData.value = resp.data.periods || []
        }
        break
      case 'liunian':
        {
          const year = selectedLiunianYear.value
          const resp = await client.post('/ziwei/period', {
            chart_id: Number(chartId),
            period_type: 'liunian',
            year,
          })
          liunianData.value = resp.data.periods || []
          // Also fetch interpretation
          const interpResp = await client.post('/ziwei/period', {
            chart_id: Number(chartId),
            period_type: 'liunian_interpretation',
            year,
          })
          liunianInterp.value[year] = interpResp.data.periods?.[0] || null
        }
        break
      case 'liuyue':
        {
          const year = selectedLiunianYear.value
          const month = new Date().getMonth() + 1
          const key = `${year}-${month}`
          if (!liuyueInterp.value[key]) {
            const resp = await client.post('/ziwei/period', {
              chart_id: Number(chartId),
              period_type: 'liuyue',
              year,
              month,
            })
            liuyueData.value = resp.data.periods || []
            const interpResp = await client.post('/ziwei/period', {
              chart_id: Number(chartId),
              period_type: 'liuyue_interpretation',
              year,
              month,
            })
            liuyueInterp.value[key] = interpResp.data.periods?.[0] || null
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
            const resp = await client.post('/ziwei/period', {
              chart_id: Number(chartId),
              period_type: 'liuri',
              year,
              month,
              day,
            })
            liuriData.value = resp.data.periods || []
            const interpResp = await client.post('/ziwei/period', {
              chart_id: Number(chartId),
              period_type: 'liuri_interpretation',
              year,
              month,
              day,
            })
            liuriInterp.value[key] = interpResp.data.periods?.[0] || null
          }
        }
        break
      case 'sihua':
        if (!Object.keys(sihuaData.value).length) {
          const resp = await client.post('/ziwei/period', {
            chart_id: Number(chartId),
            period_type: 'sihua_feixing',
          })
          sihuaData.value = resp.data.periods || {}
        }
        if (!Object.keys(sihuaChainData.value).length) {
          const chainResp = await client.post('/ziwei/period', {
            chart_id: Number(chartId),
            period_type: 'sihua_chain',
          })
          sihuaChainData.value = chainResp.data.chain || {}
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
    const resp = await client.post('/ziwei/period', {
      chart_id: Number(route.params.chartId),
      period_type: 'palace_reading',
      palace_idx: palaceIdx,
    })
    const reading = resp.data.reading
    selectedPalace.value = {
      palaceName: palace.name,
      mainStarAnalysis: {
        title: '主星特性',
        content: reading.main_star_analysis || '',
        tags: palace.mainStars.map((s) => s.name),
      },
      auxStarInfluence: {
        title: '辅星影响',
        content: reading.aux_star_influence || '',
        tags: palace.auxStars.map((s) => s.name),
      },
      sihuaInfluence: {
        title: '四化影响',
        content: reading.sihua_influence || '',
        tags: palace.sihua || [],
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
            mingZhu: chartData.mingZhu,
            shenZhu: chartData.shenZhu,
            wuxingJu: chartData.wuxingJu,
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
            <p class="tab-desc">选择一个宫位查看主星、辅星、四化详细解读</p>

            <div class="palace-quick-grid">
              <button
                v-for="(palace, idx) in (chartData?.palaces || [])"
                :key="palace.branch"
                class="palace-pill"
                :class="{
                  active: selectedPalace?.palaceName === palace.name,
                  'body-palace': chartData && palace.name === chartData.bodyPalace
                }"
                @click="onPalaceClick(palace, idx)"
              >
                <span class="palace-pill-name">{{ palace.name }}</span>
                <span class="palace-pill-branch">{{ palace.branch }}</span>
                <span v-if="chartData && palace.name === chartData.bodyPalace" class="body-badge">身</span>
              </button>
            </div>

            <ZiWeiInterpretation v-if="selectedPalace" :palace-reading="selectedPalace" />
            <div v-else class="empty-state-inline">
              <svg width="40" height="40" viewBox="0 0 40 40" fill="none">
                <circle cx="20" cy="20" r="16" stroke="#D4A84B" stroke-width="0.5" stroke-dasharray="2 3" opacity="0.3"/>
                <circle cx="20" cy="20" r="3" fill="#D4A84B" opacity="0.2"/>
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
                  <template v-if="p.mainStars?.length"><span v-for="s in p.mainStars" :key="s.name" class="strip-main-star" :class="{ dim: !s.brightness }">{{ s.name }}<small v-if="s.brightness">·{{s.brightness}}</small></span></template>
                  <span v-if="!p.mainStars?.length" class="strip-empty">无主星</span>
                  <template v-if="p.auxStars?.length"><span v-for="s in p.auxStars?.slice(0,4)" :key="s.name" class="strip-aux-star">{{ s.name }}</span></template>
                  <template v-if="p.sihua?.length"><span v-for="s in p.sihua" :key="s" class="strip-sihua">{{ s }}</span></template>
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
                  <template v-if="p.mainStars?.length"><span v-for="s in p.mainStars" :key="s.name" class="strip-main-star" :class="{ dim: !s.brightness }">{{ s.name }}<small v-if="s.brightness">·{{s.brightness}}</small></span></template>
                  <span v-if="!p.mainStars?.length" class="strip-empty">无主星</span>
                  <template v-if="p.auxStars?.length"><span v-for="s in p.auxStars?.slice(0,4)" :key="s.name" class="strip-aux-star">{{ s.name }}</span></template>
                  <template v-if="p.sihua?.length"><span v-for="s in p.sihua" :key="s" class="strip-sihua">{{ s }}</span></template>
                </div>
              </div>
            </div>
            <!-- 流月详解析 -->
            <div v-if="liuyueData[0]" class="interp-card">
              <div class="interp-header">
                <span class="interp-year">{{ liuyueData[0].year }}年{{ liuyueData[0].month }}月</span>
                <span class="interp-ganZhi">{{ liuyueInterp[selectedLiunianYear + '-' + liuyueData[0].month]?.gan_zhi || '—' }}</span>
                <div class="interp-score" :class="(liuyueInterp[selectedLiunianYear + '-' + liuyueData[0].month]?.score || 0) >= 60 ? 'score-good' : 'score-bad'">{{ liuyueInterp[selectedLiunianYear + '-' + liuyueData[0].month]?.score || '—' }}分</div>
              </div>
              <div class="interp-section">
                <div class="interp-row"><span class="interp-label">干支释义</span><span class="interp-value">{{ liuyueInterp[selectedLiunianYear + '-' + liuyueData[0].month]?.gan_zhi_desc || '—' }}</span></div>
                <div class="interp-row"><span class="interp-label">十神</span><span class="interp-value">{{ liuyueInterp[selectedLiunianYear + '-' + liuyueData[0].month]?.shi_shen || '—' }}</span></div>
                <div class="interp-row"><span class="interp-label">与命局关系</span><span class="interp-value danger">{{ liuyueInterp[selectedLiunianYear + '-' + liuyueData[0].month]?.relation_to_ming || '—' }}</span></div>
                <div class="interp-row"><span class="interp-label">作用特点</span><span class="interp-value">{{ liuyueInterp[selectedLiunianYear + '-' + liuyueData[0].month]?.effect || '—' }}</span></div>
                <div class="interp-row"><span class="interp-label">健康提示</span><span class="interp-value">{{ liuyueInterp[selectedLiunianYear + '-' + liuyueData[0].month]?.health || '—' }}</span></div>
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
                  <template v-if="p.mainStars?.length"><span v-for="s in p.mainStars" :key="s.name" class="strip-main-star" :class="{ dim: !s.brightness }">{{ s.name }}<small v-if="s.brightness">·{{s.brightness}}</small></span></template>
                  <span v-if="!p.mainStars?.length" class="strip-empty">无主星</span>
                </div>
              </div>
            </div>
            <!-- 流日详解析 -->
            <div v-if="liuriData[0]" class="interp-card">
              <div class="interp-header">
                <span class="interp-year">{{ liuriData[0].year }}年{{ liuriData[0].month }}月{{ liuriData[0].day }}日</span>
                <span class="interp-ganZhi">{{ liuriInterp[selectedLiunianYear + '-' + liuriData[0].month + '-' + liuriData[0].day]?.gan_zhi || '—' }}</span>
                <div class="interp-score" :class="(liuriInterp[selectedLiunianYear + '-' + liuriData[0].month + '-' + liuriData[0].day]?.score || 0) >= 60 ? 'score-good' : 'score-bad'">{{ liuriInterp[selectedLiunianYear + '-' + liuriData[0].month + '-' + liuriData[0].day]?.score || '—' }}分</div>
              </div>
              <div class="interp-section">
                <div class="interp-row"><span class="interp-label">干支释义</span><span class="interp-value">{{ liuriInterp[selectedLiunianYear + '-' + liuriData[0].month + '-' + liuriData[0].day]?.gan_zhi_desc || '—' }}</span></div>
                <div class="interp-row"><span class="interp-label">十神</span><span class="interp-value">{{ liuriInterp[selectedLiunianYear + '-' + liuriData[0].month + '-' + liuriData[0].day]?.shi_shen || '—' }}</span></div>
                <div class="interp-row"><span class="interp-label">与命局关系</span><span class="interp-value danger">{{ liuriInterp[selectedLiunianYear + '-' + liuriData[0].month + '-' + liuriData[0].day]?.relation_to_ming || '—' }}</span></div>
                <div class="interp-row"><span class="interp-label">七杀作用</span><span class="interp-value">{{ liuriInterp[selectedLiunianYear + '-' + liuriData[0].month + '-' + liuriData[0].day]?.qi_zi_effect || '—' }}</span></div>
                <div class="interp-row"><span class="interp-label">情绪状态</span><span class="interp-value">{{ liuriInterp[selectedLiunianYear + '-' + liuriData[0].month + '-' + liuriData[0].day]?.emotional_state || '—' }}</span></div>
                <div class="interp-row"><span class="interp-label">健康提示</span><span class="interp-value">{{ liuriInterp[selectedLiunianYear + '-' + liuriData[0].month + '-' + liuriData[0].day]?.health || '—' }}</span></div>
              </div>
              <div class="interp-section">
                <p class="interp-subtitle">时辰分析</p>
                <div class="hourly-grid">
                  <div v-for="(h, i) in (liuriInterp[selectedLiunianYear + '-' + liuriData[0].month + '-' + liuriData[0].day]?.hourly_analysis || [])" :key="i" class="hour-block" :class="h.score >= 65 ? 'hour-good' : h.score < 45 ? 'hour-bad' : 'hour-neutral'">
                    <span class="hour-time">{{ h.stem_branch }}</span>
                    <span class="hour-effect">{{ h.effect }}</span>
                    <span class="hour-score">{{ h.score }}分</span>
                  </div>
                </div>
              </div>
              <div class="interp-summary">{{ liuriInterp[selectedLiunianYear + '-' + liuriData[0].month + '-' + liuriData[0].day]?.summary || '—' }}</div>
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
  @apply min-h-screen;
  background-color: var(--bg);
}

.loading-state {
  @apply flex items-center justify-center min-h-[60vh];
}

.loading-spinner {
  width:2rem;height:2rem;border-radius:50%;border:2px dashed;animation:spin 1s linear infinite;
  border-color: var(--color-bazi-blue);
  border-top-color: transparent;
}

.error-state {
  @apply flex items-center justify-center min-h-[60vh];
}

.error-card {
  @apply bg-white rounded-lg shadow-md p-8 text-center max-w-sm;
}

.page-content {
  @apply max-w-5xl mx-auto px-4 py-6;
}

/* Birth info bar */
.birth-bar {
  @apply flex items-center justify-between flex-wrap gap-3 mb-6 p-4 rounded-lg;
  background-color: rgba(255,255,255,0.04);
  border: 1px solid rgba(212,168,75,0.08);
}

.birth-info-items {
  @apply flex flex-wrap gap-4;
}

.birth-item {
  @apply flex items-center gap-1.5;
}

.birth-label {
  @apply text-xs text-gray-400;
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
  background-color: rgba(255,255,255,0.04);
  border: 1px solid rgba(212,168,75,0.08);
}

.tab-bar {
  @apply flex overflow-x-auto;
  border-bottom: 2px solid rgba(255,255,255,0.06);
}

.tab-btn {
  @apply flex-shrink-0 px-4 py-3 text-sm font-medium cursor-pointer border-0 bg-transparent transition-colors;
  color: #999;
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
  @apply text-xs text-gray-400 mb-3;
}

.tab-loading {
  @apply text-center text-sm text-gray-400 py-8;
}

.empty-hint {
  @apply flex flex-col items-center gap-2 py-10 text-gray-400;
}

/* Palace quick grid */
.palace-quick-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(90px, 1fr));
  gap: 0.375rem;
  margin-bottom: 1.25rem;
}
.palace-pill {
  display: flex; flex-direction: column; align-items: center; gap: 0.125rem;
  padding: 0.5rem 0.5rem;
  background: rgba(255,255,255,0.02);
  border: 1px solid rgba(212,168,75,0.06);
  border-radius: 8px;
  cursor: pointer; transition: all 0.2s;
  font-family: var(--font-sans);
}
.palace-pill:hover {
  background: rgba(212,168,75,0.05);
  border-color: rgba(212,168,75,0.2);
  transform: translateY(-1px);
}
.palace-pill.active {
  border-color: rgba(212,168,75,0.3);
  background: rgba(212,168,75,0.08);
}
.palace-pill.body-palace {
  border-color: rgba(139, 75, 75, 0.4);
  background: rgba(139, 75, 75, 0.06);
}
.palace-pill-name {
  font-size: 0.75rem; font-weight: 600; color: var(--text);
  letter-spacing: 0.5px;
}
.palace-pill-branch {
  font-size: 0.58rem; color: var(--muted);
}
.body-badge {
  font-size: 0.5rem; background: rgba(139, 75, 75, 0.25); color: rgba(255, 200, 200, 0.8);
  padding: 0.05rem 0.25rem; border-radius: 3px; font-weight: 600;
}
.body-palace .palace-pill-name {
  color: rgba(255, 200, 200, 0.9);
}

/* Period lists */
.period-list {
  @apply flex flex-col gap-3;
}

.period-card {
  @apply rounded-md border p-3;
  background-color: var(--color-bazi-paper);
  border-color: rgba(43, 58, 66, 0.1);
  background: rgba(255,255,255,0.03);
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
  @apply text-xs text-gray-600 m-0;
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
  background-color: var(--color-bazi-paper);
  color: var(--color-bazi-blue);
  border-bottom: 2px solid rgba(43, 58, 66, 0.15);
}

.sihua-table td {
  @apply p-2 border-b;
  border-color: rgba(43, 58, 66, 0.05);
}

.sihua-badge {
  @apply inline-block rounded-full px-2 py-0.5 text-xs font-semibold;
}

.sihua-lu {
  background-color: rgba(196, 30, 58, 0.15);
  color: var(--color-bazi-red);
}

.sihua-quan {
  background-color: rgba(43, 58, 66, 0.12);
  color: var(--color-bazi-blue);
}

.sihua-ke {
  background-color: rgba(34, 139, 34, 0.12);
  color: #228B22;
}

.sihua-ji {
  background-color: rgba(26, 26, 26, 0.1);
  color: var(--color-bazi-ink);
}

.tab-desc { font-size:0.72rem; color:var(--muted); margin:0 0 1rem; font-style:italic; }
.dayun-timeline { display:flex; flex-direction:column; gap:0.5rem; }
.dayun-card { display:flex; gap:0.75rem; padding:0.625rem 0.75rem; background:rgba(255,255,255,0.015); border:1px solid rgba(212,168,75,0.05); border-radius:8px; }
.dayun-card.is-current { border-color:rgba(212,168,75,0.2); background:rgba(212,168,75,0.04); }
.dayun-age-badge { min-width:60px; text-align:center; }
.age-primary { display:block; font-size:0.8rem; font-weight:700; color:var(--gold); }
.age-unit { font-size:0.58rem; color:var(--muted); }
.dayun-body { flex:1; }
.dayun-palace { font-size:0.82rem; font-weight:600; color:var(--text); margin-bottom:0.125rem; }
.dayun-desc { font-size:0.68rem; color:var(--muted); margin:0 0 0.25rem; }
.dayun-stars { display:flex; flex-wrap:wrap; gap:0.2rem; }
.star-chip { padding:0.08rem 0.35rem; font-size:0.62rem; background:rgba(196,30,58,0.08); color:var(--crimson); border-radius:3px; border:1px solid rgba(196,30,58,0.12); }
.palace-strip { display:flex; align-items:center; gap:0.6rem; padding:0.4rem 0.6rem; border-bottom:1px solid rgba(255,255,255,0.025); }
.palace-strip-sm { padding:0.25rem 0.4rem; }
.palace-strip-header { display:flex; align-items:center; gap:0.3rem; min-width:68px; }
.strip-name { font-size:0.72rem; font-weight:600; color:var(--text); }
.strip-branch { font-size:0.58rem; color:var(--muted); }
.palace-strip-stars { display:flex; align-items:center; gap:0.25rem; flex-wrap:wrap; flex:1; }
.strip-main-star { padding:0.06rem 0.3rem; font-size:0.62rem; font-weight:600; background:rgba(212,168,75,0.08); color:var(--gold); border-radius:3px; }
.strip-main-star.dim { color:var(--muted); background:rgba(255,255,255,0.02); }
.strip-main-star small { font-size:0.5rem; opacity:0.5; }
.strip-aux-star { font-size:0.58rem; color:rgba(139,131,120,0.45); }
.strip-sihua { padding:0.06rem 0.3rem; font-size:0.58rem; background:rgba(196,30,58,0.08); color:var(--crimson); border-radius:3px; }
.strip-empty { font-size:0.58rem; color:var(--muted); opacity:0.3; }
.sihua-groups { display:flex; flex-direction:column; gap:0.625rem; }
.sihua-group-badge { display:inline-block; padding:0.15rem 0.5rem; font-size:0.68rem; font-weight:700; border-radius:4px; margin-bottom:0.25rem; }
.sihua-lu { background:rgba(74,222,128,0.1); color:#4ade80; }
.sihua-quan { background:rgba(212,168,75,0.1); color:var(--gold); }
.sihua-ke { background:rgba(96,165,250,0.1); color:#60a5fa; }
.sihua-ji { background:rgba(196,30,58,0.1); color:var(--crimson); }
.sihua-group-items { display:flex; flex-direction:column; gap:0.2rem; }
.sihua-fly-item { display:flex; align-items:center; gap:0.4rem; padding:0.3rem 0.5rem; background:rgba(255,255,255,0.012); border-radius:5px; font-size:0.72rem; }
.fly-star { font-weight:600; color:var(--text); }
.fly-arrow { color:var(--muted); font-size:0.65rem; }
.fly-palace { color:var(--gold); font-weight:500; }
.fly-effect { color:var(--muted); font-size:0.65rem; flex:1; }
.fly-from { font-size:0.6rem; color:#93c5fd; background:rgba(96,165,250,0.1); padding:0.05rem 0.25rem; border-radius:3px; }
.fly-chain { font-size:0.6rem; color:#86efac; background:rgba(74,222,128,0.1); padding:0.05rem 0.25rem; border-radius:3px; }
.fly-affinity { font-size:0.6rem; color:#c4a84b; background:rgba(212,168,75,0.1); padding:0.05rem 0.25rem; border-radius:3px; }
.sihua-chain-section { margin-top:1rem; padding-top:0.75rem; border-top:1px solid rgba(255,255,255,0.05); }
.chain-title { font-size:0.72rem; color:var(--muted); margin:0 0 0.5rem; font-weight:600; }
.sihua-empty-group { font-size:0.68rem; color:var(--muted); padding:0.2rem 0.4rem; opacity:0.4; }

/* Interpretation tab styles */
.interp-tab { display:flex; flex-direction:column; gap:0.75rem; }
.interp-card { background:var(--glass); border:1px solid rgba(212,168,75,0.1); border-radius:12px; overflow:hidden; }
.interp-header { display:flex; align-items:center; gap:0.75rem; padding:0.75rem 1rem; background:rgba(212,168,75,0.06); border-bottom:1px solid rgba(212,168,75,0.1); }
.interp-year { font-size:0.9rem; font-weight:700; color:var(--gold); font-family:var(--font-serif); }
.interp-ganZhi { font-size:0.85rem; color:var(--text); font-weight:600; }
.interp-score { font-size:0.8rem; font-weight:700; padding:0.15rem 0.5rem; border-radius:6px; margin-left:auto; }
.score-good { background:rgba(74,222,128,0.12); color:#4ade80; }
.score-bad { background:rgba(196,30,58,0.12); color:#f08080; }
.interp-section { padding:0.75rem 1rem; display:flex; flex-direction:column; gap:0.4rem; }
.interp-row { display:flex; gap:0.5rem; font-size:0.78rem; line-height:1.5; }
.interp-label { min-width:70px; font-weight:600; color:var(--muted); }
.interp-value { color:var(--text); flex:1; }
.interp-value.danger { color:#f08080; }
.interp-row.tip .interp-value { color:var(--gold); font-style:italic; }
.interp-subtitle { font-size:0.75rem; font-weight:700; color:var(--gold); margin:0 0 0.5rem; }
.interp-summary { font-size:0.72rem; color:var(--muted); padding:0.5rem 1rem; background:rgba(212,168,75,0.04); border-top:1px dashed rgba(212,168,75,0.1); font-style:italic; }
.hourly-grid { display:grid; grid-template-columns:repeat(auto-fill,minmax(110px,1fr)); gap:0.375rem; }
.hour-block { display:flex; flex-direction:column; gap:0.15rem; padding:0.4rem 0.5rem; border-radius:6px; font-size:0.68rem; }
.hour-good { background:rgba(74,222,128,0.08); border:1px solid rgba(74,222,128,0.15); }
.hour-neutral { background:rgba(212,168,75,0.05); border:1px solid rgba(212,168,75,0.1); }
.hour-bad { background:rgba(196,30,58,0.08); border:1px solid rgba(196,30,58,0.15); }
.hour-time { font-weight:700; color:var(--gold); }
.hour-effect { color:var(--text); line-height:1.3; }
.hour-score { font-size:0.62rem; margin-top:0.1rem; }
.hour-good .hour-score { color:#4ade80; }
.hour-neutral .hour-score { color:var(--muted); }
.hour-bad .hour-score { color:#f08080; }
</style>
