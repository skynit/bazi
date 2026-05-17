<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import client from '../api/client'
import ZiWeiChart from '../components/ZiWeiChart.vue'
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
const liuriData = ref<any[]>([])
const sihuaData = ref<any[]>([])

const liunianOverlay = ref<LiunianChartData>()
const availableYears = ref<number[]>([])
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
    chartData.value = chartValues  // backend returns ZiWeiChart directly
    birthInfo.value = chart  // chart from /charts/:id has birth info

    // Generate available years for overlay (current year ± 5)
    const currentYear = new Date().getFullYear()
    availableYears.value = Array.from({length: 11}, (_, i) => currentYear - 5 + i)

    // Load initial liunian overlay
    if (availableYears.value.length > 0) {
      await loadOverlay(availableYears.value[0])
    }
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
    liunianOverlay.value = resp.data
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
        if (!liunianData.value.length) {
          const resp = await client.post('/ziwei/period', {
            chart_id: Number(chartId),
            period_type: 'liunian',
          })
          liunianData.value = resp.data.periods || []
        }
        break
      case 'liuyue':
        if (!liuyueData.value.length) {
          const resp = await client.post('/ziwei/period', {
            chart_id: Number(chartId),
            period_type: 'liuyue',
          })
          liuyueData.value = resp.data.periods || []
        }
        break
      case 'liuri':
        if (!liuriData.value.length) {
          const resp = await client.post('/ziwei/period', {
            chart_id: Number(chartId),
            period_type: 'liuri',
          })
          liuriData.value = resp.data.periods || []
        }
        break
      case 'sihua':
        if (!sihuaData.value.length) {
          const resp = await client.post('/ziwei/period', {
            chart_id: Number(chartId),
            period_type: 'sihua_feixing',
          })
          sihuaData.value = resp.data.periods || []
        }
        break
    }
  } catch (err: any) {
    // Tab data optional
  } finally {
    loadingTab.value = false
  }
}

function onPalaceClick(palace: PalaceData) {
  // Build a mock PalaceReading from the palace data
  selectedPalace.value = {
    palaceName: palace.name,
    mainStarAnalysis: {
      title: '主星特性',
      content: palace.mainStars.length
        ? `${palace.name}主星：${palace.mainStars.map((s) => s.name + '(' + s.brightness + ')').join('、')}。主星坐守影响该宫位的基本特质。`
        : `${palace.name}无主星，借对宫安星论之。`,
      tags: palace.mainStars.map((s) => s.name),
    },
    auxStarInfluence: {
      title: '辅星影响',
      content: palace.auxStars.length
        ? `辅星增强或削弱主星力量，影响宫位细节表现。`
        : '此宫无辅星影响。',
      tags: palace.auxStars.map((s) => s.name),
    },
    sihuaInfluence: {
      title: '四化影响',
      content: palace.sihua.length
        ? `四化飞入${palace.name}，带来化禄/化权/化科/化忌的特定影响。`
        : '此宫无四化飞入。',
      tags: palace.sihua,
    },
    sanFangSiZheng: {
      title: '三方四正',
      content: '三方四正宫位相互对照，形成完整的影响力网络。',
      tags: [],
    },
    patternAnnotations: {
      title: '格局标注',
      content: chartData.value?.patterns?.length
        ? `此命盘含：${chartData.value.patterns.join('、')}`
        : '未检测到特殊格局。',
      tags: chartData.value?.patterns || [],
    },
  }
}

function onYearChange(year: number) {
  loadOverlay(year)
}
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

      <!-- ZiWei Chart -->
      <div class="chart-section" v-if="chartData">
        <div v-if="chartData.palaces && chartData.palaces.length > 0">
          <ZiWeiChart
            :palaces="chartData.palaces"
            :ming-zhu="chartData.mingZhu"
            :shen-zhu="chartData.shenZhu"
            :wuxing-ju="chartData.wuxingJu"
            :patterns="chartData.patterns"
          />
        </div>
        <div v-else class="empty-hint py-8">
          <span class="text-3xl">📋</span>
          <p>暂无命盘数据，请稍后重试</p>
        </div>
      </div>

      <!-- Overlay section -->
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
          <!-- 命盘详解: click a palace to see interpretation -->
          <div v-if="activeTab === 'mingpan'" class="mingpan-tab">
            <p class="tab-hint">点击上方命盘中的宫位查看详解</p>

            <!-- Palace click simulation via data display -->
            <div class="palace-quick-list">
              <button
                v-for="palace in chartData?.palaces || []"
                :key="palace.branch"
                class="palace-quick-btn"
                @click="onPalaceClick(palace)"
              >
                {{ palace.name }}
              </button>
            </div>

            <!-- Interpretation card -->
            <ZiWeiInterpretation
              v-if="selectedPalace"
              :palace-reading="selectedPalace"
            />
            <div v-else class="empty-hint">
              <span class="text-3xl">👆</span>
              <p>请选择一个宫位查看详细解读</p>
            </div>
          </div>

          <!-- 大限分析 -->
          <div v-else-if="activeTab === 'dayun'" class="data-tab">
            <div v-if="loadingTab" class="tab-loading">加载中...</div>
            <div v-else-if="!dayunData.length" class="empty-hint">
              <p>暂无可显示的大限数据</p>
            </div>
            <div v-else class="period-list">
              <div
                v-for="(item, idx) in dayunData"
                :key="idx"
                class="period-card"
              >
                <div class="period-header">
                  <span class="period-age">{{ item.start_age }}–{{ item.start_age + 9 }}岁</span>
                  <span class="period-palace">{{ item.palace_name }}</span>
                </div>
                <p class="period-desc">{{ item.description || item.palace_name + '大限' }}</p>
              </div>
            </div>
          </div>

          <!-- 流年分析 -->
          <div v-else-if="activeTab === 'liunian'" class="data-tab">
            <div v-if="loadingTab" class="tab-loading">加载中...</div>
            <div v-else-if="!liunianData.length" class="empty-hint">
              <p>暂无可显示的流年数据</p>
            </div>
            <div v-else class="period-list">
              <div
                v-for="(item, idx) in liunianData"
                :key="idx"
                class="period-card"
              >
                <div class="period-header">
                  <span class="period-age">{{ item.year }}年</span>
                  <span class="period-palace">{{ item.palace_name }}</span>
                </div>
                <p class="period-desc">{{ item.description || '流年运势' }}</p>
              </div>
            </div>
          </div>

          <!-- 流月分析 -->
          <div v-else-if="activeTab === 'liuyue'" class="data-tab">
            <div v-if="loadingTab" class="tab-loading">加载中...</div>
            <div v-else-if="!liuyueData.length" class="empty-hint">
              <p>暂无可显示的流月数据</p>
            </div>
            <div v-else class="period-list">
              <div
                v-for="(item, idx) in liuyueData"
                :key="idx"
                class="period-card"
              >
                <div class="period-header">
                  <span class="period-age">{{ item.year }}年{{ item.month }}月</span>
                  <span class="period-palace">{{ item.palace_name }}</span>
                </div>
                <p class="period-desc">{{ item.description || '流月运势' }}</p>
              </div>
            </div>
          </div>

          <!-- 流日分析 -->
          <div v-else-if="activeTab === 'liuri'" class="data-tab">
            <div v-if="loadingTab" class="tab-loading">加载中...</div>
            <div v-else-if="!liuriData.length" class="empty-hint">
              <p>暂无可显示的流日数据</p>
            </div>
            <div v-else class="period-list">
              <div
                v-for="(item, idx) in liuriData"
                :key="idx"
                class="period-card"
              >
                <div class="period-header">
                  <span class="period-age">{{ item.date }}</span>
                  <span class="period-palace">{{ item.palace_name }}</span>
                </div>
                <p class="period-desc">{{ item.description || '流日运势' }}</p>
              </div>
            </div>
          </div>

          <!-- 四化飞星 -->
          <div v-else-if="activeTab === 'sihua'" class="data-tab">
            <div v-if="loadingTab" class="tab-loading">加载中...</div>
            <div v-else-if="!sihuaData.length" class="empty-hint">
              <p>暂无可显示的四化飞星数据</p>
            </div>
            <div v-else class="sihua-table-wrap">
              <table class="sihua-table">
                <thead>
                  <tr>
                    <th>四化</th>
                    <th>星曜</th>
                    <th>飞入宫位</th>
                    <th>影响</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(item, idx) in sihuaData" :key="idx">
                    <td>
                      <span
                        class="sihua-badge"
                        :class="{
                          'sihua-lu': item.type === '化禄',
                          'sihua-quan': item.type === '化权',
                          'sihua-ke': item.type === '化科',
                          'sihua-ji': item.type === '化忌',
                        }"
                      >
                        {{ item.type }}
                      </span>
                    </td>
                    <td>{{ item.star }}</td>
                    <td>{{ item.target_palace }}</td>
                    <td class="text-sm text-gray-600">{{ item.description || '' }}</td>
                  </tr>
                </tbody>
              </table>
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

/* Palace quick list */
.palace-quick-list {
  @apply flex flex-wrap gap-2 mb-4;
}

.palace-quick-btn {
  @apply rounded-md border px-3 py-1.5 text-xs font-medium cursor-pointer transition-colors;
  background-color: var(--color-bazi-paper);
  border-color: var(--color-bazi-blue);
  color: var(--color-bazi-blue);
}

.palace-quick-btn:hover {
  background-color: var(--color-bazi-blue);
  color: #fff;
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
</style>
