<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { use } from 'echarts/core'
import { BarChart } from 'echarts/charts'
import { GridComponent, TooltipComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import VChart from 'vue-echarts'
import ClassicalInterpretationPanel from './ClassicalInterpretationPanel.vue'
import ChartSection from './chart/ChartSection.vue'
import { useScrollSpy } from '@/composables/useScrollSpy'
import { vReveal } from '@/composables/useReveal'
import { TEN_GOD_TRAIT, absentTenGodSentence, tenGodKeyword } from '@/lib/tenGodInterpret'
import {
  PILLAR_ROLE,
  elementBalanceSentence,
  ganRelationPlain,
  shenShaMeaning,
  zhiRelationPlain,
} from '@/lib/baziGlossary'
import type {
  BodyStrengthResult,
  FortuneLayer,
  FortuneLayerSet,
  GanRelation,
  MonthCommandStemExposure,
  MonthSeasonEvidence,
  NaYinEvidence,
  PatternAnalysis,
  PillarShenShaGroup,
  RuleMeta,
  ShenShaMeta,
  TenGodAnalysis,
  TenGodRatio,
  ZhiRelation,
} from '@/api/chart'

type DayunStageView = {
  index: number
  gan: string
  zhi: string
  pillar: string
  startAge: number
  endAge: number
  startYear: number | null
  endYear: number | null
  ganElement: string
  zhiElement: string
  tenGod: string
  isCurrent: boolean
}
type DaYunInfoView = {
  start_age?: number
  start_age_detail?: { years?: number; months?: number; days?: number }
  start_at?: string
  direction?: string
  pillars?: Array<{ gan?: string; zhi?: string }>
}

const props = defineProps<{
  chart: {
    id?: number
    year_pillar: { gan: string; zhi: string }
    month_pillar: { gan: string; zhi: string }
    day_pillar: { gan: string; zhi: string }
    hour_pillar: { gan: string; zhi: string }
    rule_meta?: RuleMeta
    body_strength?: BodyStrengthResult
    pattern_analysis?: PatternAnalysis
    day_shen_sha_details?: ShenShaMeta[]
    global_shen_sha_details?: ShenShaMeta[]
    shen_sha_by_pillar?: PillarShenShaGroup[]
    fortune_layers?: FortuneLayerSet
    month_season?: MonthSeasonEvidence
    na_yin?: Record<string, NaYinEvidence>
    ten_god_proportion?: TenGodRatio[]
    ten_god_analysis?: TenGodAnalysis
    [key: string]: any
  }
}>()

const themeVersion = ref(0)
let themeObserver: MutationObserver | null = null

function cssVar(name: string, fallback: string): string {
  if (typeof window === 'undefined') return fallback
  return getComputedStyle(document.documentElement).getPropertyValue(name).trim() || fallback
}

onMounted(() => {
  themeObserver = new MutationObserver(() => {
    themeVersion.value += 1
  })
  themeObserver.observe(document.documentElement, { attributes: true, attributeFilter: ['class'] })
})

onUnmounted(() => {
  themeObserver?.disconnect()
  themeObserver = null
})

const ganElement: Record<string, { name: string; elemColor: string }> = {
  甲: { name: '木', elemColor: '#34d399' },
  乙: { name: '木', elemColor: '#34d399' },
  丙: { name: '火', elemColor: '#fb7185' },
  丁: { name: '火', elemColor: '#fb7185' },
  戊: { name: '土', elemColor: '#c76f12' },
  己: { name: '土', elemColor: '#c76f12' },
  庚: { name: '金', elemColor: '#cbd5e1' },
  辛: { name: '金', elemColor: '#cbd5e1' },
  壬: { name: '水', elemColor: '#22d3ee' },
  癸: { name: '水', elemColor: '#22d3ee' },
}

const zhiElement: Record<string, { name: string; elemColor: string }> = {
  寅: { name: '木', elemColor: '#34d399' },
  卯: { name: '木', elemColor: '#34d399' },
  巳: { name: '火', elemColor: '#fb7185' },
  午: { name: '火', elemColor: '#fb7185' },
  辰: { name: '土', elemColor: '#c76f12' },
  戌: { name: '土', elemColor: '#c76f12' },
  丑: { name: '土', elemColor: '#c76f12' },
  未: { name: '土', elemColor: '#c76f12' },
  申: { name: '金', elemColor: '#cbd5e1' },
  酉: { name: '金', elemColor: '#cbd5e1' },
  亥: { name: '水', elemColor: '#22d3ee' },
  子: { name: '水', elemColor: '#22d3ee' },
}

const pillars = computed(() => [
  { label: '年柱', key: 'year' as const, idx: 0, ...props.chart.year_pillar },
  { label: '月柱', key: 'month' as const, idx: 1, ...props.chart.month_pillar },
  { label: '日柱', key: 'day' as const, idx: 2, ...props.chart.day_pillar },
  { label: '时柱', key: 'hour' as const, idx: 3, ...props.chart.hour_pillar },
])

const daYun = computed<DaYunInfoView>(() => props.chart.da_yun || props.chart.da_yun_start || {})

const genderLabel = computed(() => {
  const gender = String(props.chart.gender || '').toUpperCase()
  if (gender === 'MALE' || gender === 'M' || gender === '男') return '男'
  if (gender === 'FEMALE' || gender === 'F' || gender === '女') return '女'
  return '未注明'
})
const calendarLabel = computed(() =>
  String(props.chart.calendar_type || '').toUpperCase() === 'LUNAR' ? '农历' : '公历',
)
const birthDateTimeLabel = computed(() => {
  const pad = (value: unknown) => String(Number(value || 0)).padStart(2, '0')
  const date = `${props.chart.birth_year || '----'}-${pad(props.chart.birth_month)}-${pad(props.chart.birth_day)}`
  const time = `${pad(props.chart.birth_hour)}:${pad(props.chart.birth_min)}`
  return `${date} ${time}`
})
const chartIdentityItems = computed(() => [
  { label: '姓名', value: String(props.chart.name || '未命名') },
  { label: '性别', value: genderLabel.value },
  { label: calendarLabel.value, value: birthDateTimeLabel.value },
  { label: '出生地', value: String(props.chart.birth_place || '未填写') },
  { label: '时区', value: String(props.chart.timezone || '未填写') },
  {
    label: '时间校准',
    value: props.chart.use_true_solar_time ? '已使用真太阳时' : '标准钟表时间',
  },
])

function parseLocalDateTime(value?: string): Date | null {
  if (!value) return null
  const parsed = new Date(value)
  return Number.isNaN(parsed.getTime()) ? null : parsed
}

function addYears(value: Date, years: number): Date {
  const next = new Date(value)
  next.setFullYear(next.getFullYear() + years)
  return next
}

const dayunStartLabel = computed(() => {
  const detail = daYun.value.start_age_detail
  const ageParts = [
    detail?.years ? `${detail.years}年` : '',
    detail?.months ? `${detail.months}个月` : '',
    detail?.days ? `${detail.days}天` : '',
  ].filter(Boolean)
  const startAt = daYun.value.start_at
  if (startAt) {
    return `${startAt.replace('T', ' ')}${ageParts.length ? `（约 ${ageParts.join('')}）` : ''}`
  }
  return `${daYun.value.start_age || 0}岁`
})

// --- 天干地支分析（从 API 数据读取）---
const ganZhi = computed(() => props.chart.gan_zhi_analysis)

function ganRelClass(type: string): string {
  if (type === '五合') return 'rel-he'
  if (type === '相克') return 'rel-ke'
  if (type === '相生') return 'rel-sheng'
  if (type === '比和') return 'rel-he'
  return ''
}

function zhiRelClass(type: string): string {
  if (type === '六冲') return 'rel-chong'
  if (['六合', '半合', '拱合', '三合局'].includes(type)) return 'rel-he'
  if (type === '六害') return 'rel-hai'
  if (type === '六破') return 'rel-hai'
  if (type === '相刑' || type === '三刑') return 'rel-xing'
  if (type === '半会' || type === '三会局') return 'rel-hui'
  return ''
}

function ganRelSymbol(type: string): string {
  if (type === '五合') return '合'
  if (type === '相克') return '克'
  if (type === '比和') return '同'
  return '生'
}

function zhiRelSymbol(type: string): string {
  if (type === '六冲') return '冲'
  if (['六合', '半合', '拱合', '三合局'].includes(type)) return '合'
  if (type === '六害') return '害'
  if (type === '六破') return '破'
  if (type === '相刑' || type === '三刑') return '刑'
  if (type === '伏吟') return '同'
  return '会'
}

function ganRelationSummary(relation: GanRelation): string {
  const [stemA = '', stemB = ''] = relation.stems || []
  const left = `${relation.pillar1 || ''}${stemA}`
  const right = `${relation.pillar2 || ''}${stemB}`
  const elementA = ganElement[stemA]?.name || ''
  const elementB = ganElement[stemB]?.name || ''

  if (relation.type === '相生') {
    return produces(elementA, elementB) ? `${left}生助${right}。` : `${right}生助${left}。`
  }
  if (relation.type === '相克') {
    return controls(elementA, elementB) ? `${left}克制${right}。` : `${right}克制${left}。`
  }
  if (relation.type === '比和') return `${left}与${right}同属${elementA}行。`
  if (relation.type === '五合') {
    const dispute = relation.status === 'disputed' ? '，同时参与其他五合' : ''
    return `${left}与${right}形成天干五合${dispute}。`
  }
  if (relation.type === '天干相冲') return `${left}与${right}形成天干相冲。`
  return `${left}与${right}形成${relation.type}关系。`
}

function zhiRelationSummary(relation: ZhiRelation): string {
  const parts = (relation.pillars || []).map(
    (pillar, index) => `${pillar}${relation.branches?.[index] || ''}`,
  )
  const subject = parts.length > 1 ? parts.join('与') : parts[0] || '相关地支'
  const subtype =
    relation.subtype && relation.subtype !== relation.type ? `（${relation.subtype}）` : ''
  return `${subject}形成${relation.type}${subtype}。`
}

const isDarkTheme = () => {
  themeVersion.value
  return typeof document !== 'undefined' && document.documentElement.classList.contains('dark')
}

const elemColor = (e: string) => {
  const isDark = isDarkTheme()
  const lightMap: Record<string, string> = {
    金: '#64748b',
    木: '#16a34a',
    水: '#0891b2',
    火: '#e11d48',
    土: '#c76f12',
  }
  const darkMap: Record<string, string> = {
    金: '#cbd5e1',
    木: '#34d399',
    水: '#22d3ee',
    火: '#fb7185',
    土: '#f59e0b',
  }
  return (isDark ? darkMap : lightMap)[e] || '#8a9a8e'
}

// 干支着色跟随 elemColor，避免浅色主题下金行文字几乎不可读
const ganzhiColor = (meta?: { name: string; elemColor: string }) =>
  meta ? elemColor(meta.name) : 'var(--text-muted)'

const pillarLabel = (k: string) =>
  ({ year: '年柱', month: '月柱', day: '日柱', hour: '时柱' })[k] || k

const ganPolarity: Record<string, 'yang' | 'yin'> = {
  甲: 'yang',
  丙: 'yang',
  戊: 'yang',
  庚: 'yang',
  壬: 'yang',
  乙: 'yin',
  丁: 'yin',
  己: 'yin',
  辛: 'yin',
  癸: 'yin',
}

const elementCycle = ['木', '火', '土', '金', '水']

function produces(source: string, target: string): boolean {
  const idx = elementCycle.indexOf(source)
  return idx >= 0 && elementCycle[(idx + 1) % elementCycle.length] === target
}

function controls(source: string, target: string): boolean {
  const idx = elementCycle.indexOf(source)
  return idx >= 0 && elementCycle[(idx + 2) % elementCycle.length] === target
}

function tenGodFor(dayGan?: string, targetGan?: string): string {
  if (!dayGan || !targetGan) return '未判'
  const dayElem = ganElement[dayGan]?.name
  const targetElem = ganElement[targetGan]?.name
  if (!dayElem || !targetElem) return '未判'
  const samePolarity = ganPolarity[dayGan] === ganPolarity[targetGan]

  if (targetElem === dayElem) return samePolarity ? '比肩' : '劫财'
  if (produces(targetElem, dayElem)) return samePolarity ? '偏印' : '正印'
  if (produces(dayElem, targetElem)) return samePolarity ? '食神' : '伤官'
  if (controls(dayElem, targetElem)) return samePolarity ? '偏财' : '正财'
  if (controls(targetElem, dayElem)) return samePolarity ? '七杀' : '正官'
  return '未判'
}

const currentAge = computed(() => {
  const birthYear = Number(props.chart.birth_year || 0)
  if (!birthYear) return null
  const now = new Date()
  const birthMonth = Number(props.chart.birth_month || 0)
  const birthDay = Number(props.chart.birth_day || 0)
  let age = now.getFullYear() - birthYear
  if (birthMonth && birthDay) {
    const monthPassed = now.getMonth() + 1 > birthMonth
    const dayPassed = now.getMonth() + 1 === birthMonth && now.getDate() >= birthDay
    if (!monthPassed && !dayPassed) age -= 1
  }
  return Math.max(0, age)
})

const dayunStages = computed<DayunStageView[]>(() => {
  const info = daYun.value
  const pillarsRaw: Array<{ gan?: string; zhi?: string }> = Array.isArray(info.pillars)
    ? info.pillars
    : []
  const source = pillarsRaw
  const startAgeBase = Number(info.start_age || 0)
  const birthYear = Number(props.chart.birth_year || 0)
  const age = currentAge.value
  const exactStart = parseLocalDateTime(info.start_at)
  const now = new Date()

  return source.map((p, index) => {
    const gan = String(p.gan || '')
    const zhi = String(p.zhi || '')
    const pillar = gan + zhi
    const startAge = startAgeBase + index * 10
    const endAge = startAge + 9
    const stageStart = exactStart ? addYears(exactStart, index * 10) : null
    const stageEnd = stageStart ? addYears(stageStart, 10) : null
    const startYear = stageStart?.getFullYear() || (birthYear ? birthYear + startAge : null)
    const endYear = stageEnd ? stageEnd.getFullYear() - 1 : startYear ? startYear + 9 : null

    return {
      index,
      gan,
      zhi,
      pillar,
      startAge,
      endAge,
      startYear,
      endYear,
      ganElement: ganElement[gan]?.name || '',
      zhiElement: zhiElement[zhi]?.name || '',
      tenGod: tenGodFor(props.chart.day_pillar?.gan, gan),
      isCurrent:
        stageStart && stageEnd
          ? now >= stageStart && now < stageEnd
          : age !== null && age >= startAge && age <= endAge,
    }
  })
})

const hasDaYun = computed(() => dayunStages.value.length > 0)
const currentDayunStage = computed(() => dayunStages.value.find((stage) => stage.isCurrent) || null)
const currentDayunLayer = computed(() => props.chart.fortune_layers?.dayun || null)

function parseShenSha(raw: string) {
  const [head] = raw.split('｜')
  const colonIndex = head.indexOf('：')
  if (colonIndex === -1) return { name: head, target: '', desc: '' }

  const name = head.slice(0, colonIndex)
  const target = head.slice(colonIndex + 1)
  return { name, target, desc: '' }
}

const parsedDayShenSha = computed(() => (props.chart.day_shen_sha || []).map(parseShenSha))

const groupedShenSha = computed(() => {
  const raw = props.chart.shen_sha_by_pillar
  if (!raw || !raw.length) return null
  return raw.map((g: any) => ({ ...g, items: (g.items || []).map(parseShenSha) }))
})

const globalShenSha = computed(() => (props.chart.global_shen_sha || []).map(parseShenSha))

const bodyStrengthComponents = computed(() => props.chart.body_strength?.components || [])

function monthCommandExposureLabel(exposures: MonthCommandStemExposure[]): string {
  return exposures.map((item) => `${item.pillar}${item.stem}`).join('、')
}

const shenShaDetails = computed(() => {
  const seen = new Set<string>()
  const details: ShenShaMeta[] = []
  const add = (items?: ShenShaMeta[]) => {
    for (const item of items || []) {
      if (!item?.name || seen.has(item.name)) continue
      seen.add(item.name)
      details.push(item)
    }
  }
  add(props.chart.day_shen_sha_details)
  add(props.chart.global_shen_sha_details)
  for (const group of props.chart.shen_sha_by_pillar || []) add(group.details)
  return details.slice(0, 8)
})

const fortuneLayerList = computed(() => {
  const layers = props.chart.fortune_layers
  if (!layers) return []
  return (['liunian', 'liuyue', 'xiaoyun'] as const)
    .map((key) => layers[key])
    .filter((layer): layer is FortuneLayer => Boolean(layer))
})

function calculationBasisLabel(value?: string): string {
  const labels: Record<string, string> = {
    exact_start_time_and_query_time: '依据出生时间与查询日期定位',
    period_pillar_and_natal_chart: '依据周期干支与本命四柱对照',
    period_layer_stem_pair: '依据周期天干关系',
    period_stem_and_target_stem_all_structures: '依据周期天干与本命天干关系',
    period_branch_and_target_branch_all_structures: '依据周期地支与本命地支关系',
  }
  return value ? labels[value] || '依据周期干支与本命结构对照' : '依据周期干支与本命结构对照'
}

function relationEndpointLabel(value?: string): string {
  const labels: Record<string, string> = {
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
  return value ? labels[value] || '命盘位置' : '命盘位置'
}

function relationDisplayLabel(type?: string, fallback?: string): string {
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

function componentWidth(score: number): string {
  const value = Number(score || 0)
  if (value >= 0.67) return '85%'
  if (value >= 0.34) return '60%'
  return '35%'
}

function componentBandLabel(score: number): string {
  const value = Number(score || 0)
  if (value >= 0.67) return '相对较多'
  if (value >= 0.34) return '中等'
  return '相对较少'
}

function strengthCandidateLabel(value?: string): string {
  const labels: Record<string, string> = {
    身旺: '日主力量偏强候选（传统称“身旺”）',
    身强: '日主力量偏强候选（传统称“身强”）',
    身弱: '日主力量偏弱候选（传统称“身弱”）',
    中和: '日主力量居中候选（传统称“中和”）',
  }
  return value ? labels[value] || `${value}（规则候选）` : '未形成候选'
}

const pillarShenShaColor = (p: string) =>
  ({ day: 'var(--accent)', year: '#5BA4CF', month: '#60B89A', hour: '#A182CF' })[p] || '#888'

const pillarShenShaBg = (p: string) =>
  ({
    day: 'var(--accent-dim)',
    year: 'rgba(91,164,207,0.06)',
    month: 'rgba(96,184,154,0.06)',
    hour: 'rgba(161,130,207,0.06)',
  })[p] || 'var(--accent-dim)'

function strengthLevel(total: number): string {
  if (total <= 0) return 'none'
  if (total <= 5) return 'weak'
  if (total <= 15) return 'medium'
  if (total <= 25) return 'strong'
  return 'very-strong'
}

function elementDistributionBand(value: number, maxValue: number): string {
  if (value <= 0 || maxValue <= 0) return '未见'
  const ratio = value / maxValue
  if (ratio >= 0.75) return '较多'
  if (ratio >= 0.4) return '中等'
  return '较少'
}

const fiveElementsOption = computed(() => {
  themeVersion.value
  const textColor = cssVar('--text', '#0f1712')
  const lineColor = cssVar('--line-subtle', 'rgba(15, 23, 18, 0.06)')
  const tooltipBg = cssVar('--surface-1', '#ffffff')
  const fe = props.chart.five_elements
  if (!fe) return null
  const total = Object.values(fe as Record<string, number>).reduce((s, v) => s + v, 0)
  if (total === 0) return null

  // 五行配色 — 与干支着色保持同一色板
  const labels = ['木', '火', '土', '金', '水']
  const barColors = labels.map((label) => elemColor(label))
  const maxValue = Math.max(...labels.map((label) => Number(fe[label] || 0)))

  return {
    backgroundColor: 'transparent',
    grid: { left: 8, right: 8, bottom: 28, top: 12, containLabel: true },
    xAxis: {
      type: 'category',
      data: labels,
      axisLine: { lineStyle: { color: lineColor } },
      axisTick: { show: false },
      axisLabel: {
        color: textColor,
        fontSize: 14,
        fontWeight: '600',
        fontFamily: 'Noto Serif SC, Songti SC, serif',
      },
      splitLine: { show: false },
    },
    yAxis: {
      type: 'value',
      max: 30,
      axisLabel: { show: false },
      splitLine: { lineStyle: { color: lineColor, type: 'dashed' } },
      axisLine: { show: false },
      axisTick: { show: false },
    },
    series: [
      {
        type: 'bar',
        data: labels.map((l, i) => ({
          value: fe[l] || 0,
          itemStyle: {
            color: {
              type: 'linear',
              x: 0,
              y: 0,
              x2: 0,
              y2: 1,
              colorStops: [
                { offset: 0, color: barColors[i] },
                { offset: 1, color: barColors[i] + '66' },
              ],
            },
            borderRadius: [4, 4, 0, 0],
          },
          emphasis: {
            itemStyle: {
              color: {
                type: 'linear',
                x: 0,
                y: 0,
                x2: 0,
                y2: 1,
                colorStops: [
                  { offset: 0, color: barColors[i] + 'EE' },
                  { offset: 1, color: barColors[i] + '99' },
                ],
              },
            },
          },
        })),
        barMaxWidth: 38,
        barCategoryGap: '30%',
        label: {
          show: true,
          position: 'top',
          formatter: (params: { value: number }) =>
            elementDistributionBand(Number(params.value || 0), maxValue),
          fontSize: 14,
          fontWeight: '600',
          color: textColor,
          fontFamily: 'Noto Sans SC, PingFang SC, sans-serif',
        },
      },
    ],
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'shadow' },
      backgroundColor: tooltipBg,
      borderColor: lineColor,
      borderWidth: 1,
      padding: [8, 14],
      textStyle: {
        color: textColor,
        fontSize: 14,
        fontFamily: 'Noto Serif SC, Songti SC, serif',
        fontWeight: '600',
      },
      formatter: (params: any[]) => {
        const p = params[0]
        const band = elementDistributionBand(Number(p.value || 0), maxValue)
        return `<span style="font-weight:700">${p.name}</span>：<span style="color:${textColor};font-weight:700">${band}</span><br><span style="font-size:14px;font-weight:400">固定权重下的命盘内比较，尚未独立验证</span>`
      },
    },
    animationDuration: 900,
    animationEasing: 'cubicOut' as const,
  }
})

const pillarDetails = computed(() => props.chart.pillar_details || [])

// tiaohouElem returns the element of the table-first candidate for color coding.
const tiaohouElem = computed(() => {
  const g = props.chart.tiaohou?.table_primary_candidate
  const map: Record<string, string> = {
    甲: '木',
    乙: '木',
    丙: '火',
    丁: '火',
    戊: '土',
    己: '土',
    庚: '金',
    辛: '金',
    壬: '水',
    癸: '水',
  }
  return g ? map[g] || '金' : '金'
})

function monthCommandSegmentLabel(startDay: number, endDay?: number): string {
  return endDay ? `第 ${startDay}-${endDay} 天` : `第 ${startDay} 天起`
}

use([BarChart, GridComponent, TooltipComponent, CanvasRenderer])

// 长滚动分区导航（替代原 tab 切换）：目录数据 + scroll-spy
const hasWuxingSection = computed(() => {
  const detail = props.chart.element_detail
  const missing = props.chart.missing_elements
  return Boolean(
    (Array.isArray(detail) && detail.length) ||
    missing?.missing_elements?.length ||
    missing?.weak_elements?.length,
  )
})
const hasShishenSection = computed(
  () =>
    Boolean(props.chart.ten_god_proportion?.length) ||
    props.chart.ten_god_analysis?.status === 'observed',
)
const hasPatternSection = computed(() =>
  Boolean(props.chart.pattern_analysis || props.chart.body_strength || props.chart.tiaohou),
)
const hasShenshaSection = computed(
  () =>
    shenShaDetails.value.length > 0 ||
    Boolean(groupedShenSha.value?.length) ||
    parsedDayShenSha.value.length > 0 ||
    globalShenSha.value.length > 0,
)
const hasFortuneSection = computed(
  () =>
    fortuneLayerList.value.length > 0 ||
    Boolean(props.chart.ming_gong?.gan_zhi) ||
    props.chart.month_season?.status === 'observed',
)

// 分区按普通用户阅读价值排序：总览/大运/十神在前，参考与深钻类在后
const chartSections = computed(() => {
  const sections = [{ key: 'overview', label: '命盘总览', sub: '十神 · 纳音 · 干支关系' }]
  if (hasDaYun.value) sections.push({ key: 'dayun', label: '大运', sub: '十年周期阶段' })
  if (hasShishenSection.value)
    sections.push({ key: 'shishen', label: '十神结构', sub: '生克制化统计' })
  if (hasWuxingSection.value)
    sections.push({ key: 'wuxing', label: '五行格局', sub: '占比与白话总结' })
  if (hasShenshaSection.value)
    sections.push({ key: 'shensha', label: '神煞', sub: '传统标记与寓意' })
  if (hasFortuneSection.value)
    sections.push({ key: 'fortune', label: '周期结构', sub: '流年 · 流月 · 命宫' })
  if (hasPatternSection.value)
    sections.push({ key: 'pattern', label: '传统参考', sub: '格局 · 强弱 · 调候' })
  if (props.chart.id) sections.push({ key: 'rules', label: '经典依据', sub: '古籍与计算日志' })
  return sections
})

const sectionIds = computed(() =>
  chartSections.value.map((section) => `chart-section-${section.key}`),
)
const { activeId: activeSectionId, scrollTo: scrollToSection } = useScrollSpy(sectionIds, {
  offsetTop: 92,
})

function sectionNo(key: string): string {
  const idx = chartSections.value.findIndex((section) => section.key === key)
  return idx >= 0 ? `SECTION ${String(idx + 1).padStart(2, '0')}` : ''
}

// 折叠分区的摘要句：不展开也能获得核心信息
const fortuneSectionSummary = computed(() => {
  const parts: string[] = []
  const pillar = currentDayunStage.value?.pillar || currentDayunLayer.value?.pillar
  if (pillar) parts.push(`当前大运 ${pillar}`)
  for (const layer of fortuneLayerList.value) {
    parts.push(`${layer.name} ${layer.pillar}`)
  }
  if (props.chart.ming_gong?.gan_zhi) parts.push(`命宫 ${props.chart.ming_gong.gan_zhi}`)
  return parts.length ? parts.join(' · ') : '周期干支与本命结构的对照记录'
})

// 传统参考分区折叠时的一句话摘要：格局线索 + 强弱候选 + 调候首项
const patternSectionSummary = computed(() => {
  const parts: string[] = []
  const candidates = props.chart.pattern_analysis?.candidates
  if (candidates?.length) {
    const names = candidates.slice(0, 3).map((candidate) => candidate.pattern_name)
    parts.push(
      `格局线索 ${names.join('、')}${candidates.length > 3 ? ` 等 ${candidates.length} 项` : ''}`,
    )
  }
  const band = props.chart.body_strength?.score_band_candidate
  if (band) parts.push(`强弱候选 ${band}`)
  const tiaohouPrimary = props.chart.tiaohou?.table_primary_candidate
  if (tiaohouPrimary) parts.push(`调候首项 ${tiaohouPrimary}`)
  return parts.length ? parts.join(' · ') : '格局、强弱与调候的传统参考口径'
})

// 十神配色：堆叠占比条与 chips 共用同一色板（中明度色，明暗主题均可辨）
const TEN_GOD_COLORS: Record<string, string> = {
  比肩: '#cbd5e1',
  劫财: '#94a3b8',
  食神: '#9B72CF',
  伤官: '#C85FCF',
  正财: '#60a5fa',
  偏财: '#3b82f6',
  正官: '#34d399',
  七杀: '#059669',
  正印: '#fb7185',
  偏印: '#e11d48',
}

type TenGodShareItem = { name: string; count: number; percent: number }

// 十神分布统一数据源：优先十神结构分析（含排名与次数），缺失时回退到占比接口
const tenGodShares = computed<TenGodShareItem[]>(() => {
  const analysis = props.chart.ten_god_analysis
  if (analysis?.status === 'observed' && analysis.ranked_gods?.length) {
    return analysis.ranked_gods.map((god) => ({
      name: god.god,
      count: god.count,
      percent: god.percent,
    }))
  }
  return (props.chart.ten_god_proportion || []).map((item) => ({
    name: item.name,
    count: item.count,
    percent: item.percent,
  }))
})

// 堆叠条只画占比 > 0 的分段；为 0 的十神在下方 chips 中弱化呈现
const tenGodStackSegs = computed(() => tenGodShares.value.filter((item) => item.percent > 0))

// 白话解读：占比最高的 1-2 个十神给出特质与留意；占比 0% 的十神一句带过
const tenGodInsightItems = computed(() =>
  [...tenGodShares.value]
    .filter((item) => item.percent > 0)
    .sort((a, b) => b.percent - a.percent)
    .slice(0, 2)
    .map((item) => ({
      name: item.name,
      keyword: tenGodKeyword(item.name),
      trait: TEN_GOD_TRAIT[item.name]?.trait ?? '',
      caution: TEN_GOD_TRAIT[item.name]?.caution ?? '',
    }))
    .filter((item) => item.trait),
)

const tenGodAbsentSentence = computed(() =>
  absentTenGodSentence(
    tenGodShares.value.filter((item) => item.percent <= 0).map((item) => item.name),
  ),
)

const tenGodColor = (name: string) => TEN_GOD_COLORS[name] || 'var(--text-soft)'

// ── 白话层：术语翻译 + 数据驱动小结 ─────────────────────────

/** 十神标签的白话短注：日主特殊处理，其余复用全站统一关键词 */
function tenGodTagPlain(name?: string): string {
  if (!name) return ''
  if (name === '日主') return '代表你自己'
  return tenGodKeyword(name)
}

// 命盘总览开头一句"这是什么"：日主定位 + 干支关系数量（全部由实际数据生成）
const overviewSummary = computed(() => {
  const dayGan = String(props.chart.day_pillar?.gan || '')
  const dayElem = ganElement[dayGan]?.name || ''
  const relCount =
    (ganZhi.value?.gan_relations?.length || 0) + (ganZhi.value?.zhi_relations?.length || 0)
  const parts: string[] = []
  if (dayGan && dayElem) {
    parts.push(`你的日主是「${dayGan}」，五行属${dayElem}——日主就是日柱的天干，代表你自己`)
  }
  if (relCount > 0) parts.push(`四柱之间共记录到 ${relCount} 组干支关系，下面逐条配有白话解释`)
  return parts.length ? `${parts.join('；')}。` : ''
})

// 大运当前步的白话小结：你现在走哪一步 + 该运十神关键词（hedged）
const dayunCurrentSummary = computed(() => {
  const stage = currentDayunStage.value
  const layer = currentDayunLayer.value
  const pillar = stage?.pillar || layer?.pillar || ''
  if (!pillar) return ''
  const god = stage?.tenGod || layer?.ten_god?.name || ''
  const keyword = tenGodKeyword(god)
  const range =
    stage && stage.startYear && stage.endYear
      ? `${stage.startYear}–${stage.endYear} · ${stage.startAge}–${stage.endAge} 岁`
      : stage
        ? `${stage.startAge}–${stage.endAge} 岁`
        : ''
  let text = `你现在正走「${pillar}」大运${range ? `（${range}）` : ''}。`
  if (god && keyword) {
    text += `这步运的十神是「${god}」，传统上这十年的整体节奏偏向「${keyword}」——说的是倾向，不代表必然发生什么。`
  }
  return text
})

// 五行格局开头一句数据驱动人话：按占比生成"谁偏多谁偏少"
const wuxingSummary = computed(() => {
  const detail = props.chart.element_detail
  let scores: Record<string, number> | null = null
  if (Array.isArray(detail) && detail.length) {
    scores = {}
    for (const row of detail as Array<{ element: string; total: number }>) {
      scores[row.element] = Number(row.total || 0)
    }
  } else if (props.chart.five_elements) {
    scores = props.chart.five_elements as Record<string, number>
  }
  return scores ? elementBalanceSentence(scores) : ''
})

// 神煞开头一句：命中数量 + 前几项名称，强调只是传统标记
const shenShaSummary = computed(() => {
  const names = shenShaDetails.value.map((item) => item.name)
  if (!names.length) return ''
  const shown = names.slice(0, 3).join('、')
  return `这张命盘共命中 ${names.length} 项神煞（${shown}${names.length > 3 ? ' 等' : ''}）。它们只是传统说法里的标记，供参考，不代表吉凶。`
})
</script>

<template>
  <div class="bazi-chart">
    <div class="chart-layout">
      <!-- 锚点目录：桌面左侧 sticky，移动端顶部横向 chip 条 -->
      <aside class="chart-toc" aria-label="命盘分区目录">
        <nav class="toc-inner">
          <button
            v-for="section in chartSections"
            :key="section.key"
            type="button"
            class="toc-item"
            :class="{ active: activeSectionId === `chart-section-${section.key}` }"
            @click="scrollToSection(`chart-section-${section.key}`)"
          >
            <span class="toc-label">{{ section.label }}</span>
            <span class="toc-sub">{{ section.sub }}</span>
          </button>
        </nav>
      </aside>

      <div class="chart-card glass-card overflow-hidden">
        <!-- Title -->
        <div class="chart-header">
          <div class="header-eyebrow">BaZi Fortune</div>
          <h1 class="chart-title">八字命盘</h1>
        </div>

        <dl class="chart-identity" aria-label="命盘身份摘要">
          <div v-for="item in chartIdentityItems" :key="item.label">
            <dt>{{ item.label }}</dt>
            <dd>{{ item.value }}</dd>
          </div>
        </dl>

        <!-- Four-pillar axis with centered five elements -->
        <div class="pillars-bento">
          <div
            v-for="(pillar, pi) in pillars"
            :key="pillar.key"
            class="bento-card bento-small"
            :class="[
              'bento-hover-' + ganElement[pillar.gan]?.name,
              { 'bento-day-card': pillar.key === 'day' },
            ]"
            :style="{ order: pillar.idx >= 2 ? pillar.idx + 2 : pillar.idx + 1 }"
          >
            <div class="bento-label">{{ pillar.label }}</div>
            <div class="bento-role">
              {{ PILLAR_ROLE[pillar.key] }}{{ pillar.key === 'day' ? ' · 日主 = 你自己' : '' }}
            </div>
            <div class="bento-body">
              <div class="bento-gan" :style="{ color: ganzhiColor(ganElement[pillar.gan]) }">
                <span class="bento-char">{{ pillar.gan }}</span>
                <span
                  class="elem-tag"
                  :style="{
                    background: ganzhiColor(ganElement[pillar.gan]) + '1f',
                    color: ganzhiColor(ganElement[pillar.gan]),
                    borderColor: ganzhiColor(ganElement[pillar.gan]) + '40',
                  }"
                  >{{ ganElement[pillar.gan]?.name }}</span
                >
              </div>
              <div class="bento-zhi" :style="{ color: ganzhiColor(zhiElement[pillar.zhi]) }">
                <span class="bento-char">{{ pillar.zhi }}</span>
                <span
                  class="elem-tag"
                  :style="{
                    background: ganzhiColor(zhiElement[pillar.zhi]) + '1f',
                    color: ganzhiColor(zhiElement[pillar.zhi]),
                    borderColor: ganzhiColor(zhiElement[pillar.zhi]) + '40',
                  }"
                  >{{ zhiElement[pillar.zhi]?.name }}</span
                >
              </div>
              <span v-if="chart.ten_gods?.[pillar.key]" class="bento-god-tag"
                >{{ chart.ten_gods[pillar.key]
                }}<template v-if="tenGodTagPlain(chart.ten_gods[pillar.key])">
                  · {{ tenGodTagPlain(chart.ten_gods[pillar.key]) }}</template
                ></span
              >
            </div>
            <div v-if="pillarDetails[pi]" class="bento-sub">
              <span class="sheng-xiao-tag">{{ pillarDetails[pi].sheng_xiao }}</span>
              <span
                v-if="pillarDetails[pi].empties[0]"
                class="empties-tag"
                title="空亡：传统上视为“落空、打折扣”的位置标记"
              >
                空{{ pillarDetails[pi].empties[0] }}{{ pillarDetails[pi].empties[1] }}
              </span>
            </div>
          </div>

          <div v-if="fiveElementsOption" class="bento-card bento-radar">
            <div class="bento-label">五行分布</div>
            <v-chart class="bento-radar-chart" :option="fiveElementsOption" autoresize />
            <p class="bento-radar-note">
              柱高仅表示固定权重下的命盘内相对分布；尚未独立验证，不代表强弱结论或现实结果。
            </p>
          </div>
        </div>

        <!-- Analysis sections：长滚动分区，替代原 tab 切换 -->
        <div class="analysis-section">
          <!-- ═══ Section: 命盘总览 (overview) ═══ -->
          <ChartSection
            id="chart-section-overview"
            :eyebrow="sectionNo('overview')"
            title="命盘总览"
            desc="四柱 = 出生年月日时换算成的四组干支，每柱两个字。这里看每柱的角色标签、纳音叫法和干支之间的关系。"
            :collapsible="false"
          >
            <div class="overview-layout">
              <p v-if="overviewSummary" class="plain-summary">{{ overviewSummary }}</p>
              <!-- Five Elements chart moved to Bento Grid -->

              <!-- Ten Gods -->
              <section
                v-if="chart.ten_gods"
                class="analysis-block overview-section overview-ten-gods"
              >
                <div class="overview-section-head">
                  <div>
                    <div class="block-title">十神</div>
                    <span class="block-desc"
                      >十神 =
                      以日主（你自己）为参照，给其他柱天干贴的角色标签；是关系名称，不直接代表性格或现实结果</span
                    >
                  </div>
                </div>
                <div class="overview-section-body">
                  <div class="ten-gods-grid">
                    <div v-for="(god, pillar) in chart.ten_gods" :key="pillar" class="god-item">
                      <span class="god-pillar">{{ pillarLabel(String(pillar)) }}</span>
                      <span class="god-name">{{ god }}</span>
                      <span v-if="tenGodTagPlain(String(god))" class="god-plain">{{
                        tenGodTagPlain(String(god))
                      }}</span>
                    </div>
                  </div>
                </div>
              </section>

              <!-- NaYin -->
              <section v-if="chart.na_yin" class="analysis-block overview-section overview-nayin">
                <div class="overview-section-head">
                  <div>
                    <div class="block-title">纳音</div>
                    <span class="block-desc"
                      >纳音 = 传统对干支组合的另一种分类叫法（如"路旁土"），供参考</span
                    >
                  </div>
                </div>
                <div class="overview-section-body">
                  <div class="nayin-list">
                    <span
                      v-for="(info, key) in chart.na_yin"
                      :key="key"
                      class="nayin-tag"
                      :style="{ borderColor: elemColor(info.element) }"
                    >
                      <span class="nayin-pillar">{{ pillarLabel(String(key)) }}</span>
                      <span class="nayin-ganzhi">{{ info.gan_zhi }}</span>
                      <span class="nayin-name" :style="{ color: elemColor(info.element) }">{{
                        info.name
                      }}</span>
                      <span class="nayin-element">{{ info.element }}</span>
                    </span>
                  </div>
                </div>
              </section>

              <section v-if="ganZhi" class="analysis-block overview-section overview-relations">
                <div class="overview-section-head">
                  <div>
                    <div class="block-title">天干地支关系</div>
                    <span class="block-desc"
                      >看四柱的天干地支之间如何互动：合、冲、生、克都是传统里的关系叫法，每条下面配一句白话</span
                    >
                  </div>
                </div>
                <div class="overview-relations-grid">
                  <div class="relation-card">
                    <div v-if="ganZhi.gan_relations?.length > 0" class="relations-section">
                      <div class="relations-title">
                        <span class="relations-title-dot"></span>
                        天干关系
                        <span class="relations-count">{{ ganZhi.gan_relations.length }}组</span>
                      </div>
                      <div class="ganzhi-compact">
                        <div
                          v-for="(rel, ri) in ganZhi.gan_relations"
                          :key="'g' + ri"
                          class="gz-item"
                          :class="ganRelClass(rel.type)"
                        >
                          <span class="gz-chars">
                            <span class="gz-c">{{ rel.pillar1 }}</span>
                            <span class="gz-sym" :class="'sym-' + ganRelClass(rel.type)">{{
                              ganRelSymbol(rel.type)
                            }}</span>
                            <span class="gz-c">{{ rel.pillar2 }}</span>
                          </span>
                          <span class="gz-tag" :class="'tag-' + ganRelClass(rel.type)">{{
                            rel.type + (rel.status === 'disputed' ? ' · 争议' : '')
                          }}</span>
                          <span class="gz-text">{{ ganRelationSummary(rel) }}</span>
                          <span v-if="ganRelationPlain(rel.type)" class="gz-plain">{{
                            ganRelationPlain(rel.type)
                          }}</span>
                        </div>
                      </div>
                    </div>
                    <div v-else class="no-relations">
                      <span class="no-rel-icon">◇</span> 天干无特殊关系
                    </div>
                  </div>

                  <div class="relation-card">
                    <div v-if="ganZhi.zhi_relations?.length > 0" class="relations-section">
                      <div class="relations-title">
                        <span class="relations-title-dot zhi-dot"></span>
                        地支关系
                        <span class="relations-count">{{ ganZhi.zhi_relations.length }}组</span>
                      </div>
                      <div class="ganzhi-compact">
                        <div
                          v-for="(rel, ri) in ganZhi.zhi_relations"
                          :key="'z' + ri"
                          class="gz-item"
                          :class="zhiRelClass(rel.type)"
                        >
                          <span class="gz-chars">
                            <template v-for="(pillar, pi) in rel.pillars" :key="pillar + pi">
                              <span class="gz-c">{{ pillar }}</span>
                              <span
                                v-if="pi < rel.pillars.length - 1"
                                class="gz-sym"
                                :class="'sym-' + zhiRelClass(rel.type)"
                                >{{ zhiRelSymbol(rel.type) }}</span
                              >
                            </template>
                          </span>
                          <span class="gz-tag" :class="'tag-' + zhiRelClass(rel.type)">{{
                            rel.type + (rel.status === 'disputed' ? ' · 争议' : '')
                          }}</span>
                          <span class="gz-text">{{ zhiRelationSummary(rel) }}</span>
                          <span v-if="zhiRelationPlain(rel.type)" class="gz-plain">{{
                            zhiRelationPlain(rel.type)
                          }}</span>
                        </div>
                      </div>
                    </div>
                    <div v-else class="no-relations">
                      <span class="no-rel-icon">◇</span> 地支无特殊关系
                    </div>
                  </div>
                </div>
                <p class="relation-boundary-note">
                  这里只展示命盘中的干支与五行关系，不据此判断具体事件。
                </p>
              </section>
            </div>
          </ChartSection>
          <!-- /overview section -->

          <!-- ═══ Section: 大运 (dayun) ═══ -->
          <ChartSection
            v-if="hasDaYun"
            id="chart-section-dayun"
            :eyebrow="sectionNo('dayun')"
            title="大运"
            desc="大运 = 十年一步的人生大周期，看的是每个阶段的整体节奏，而不是某一天的好坏。"
            v-reveal
          >
            <div class="dayun-detail-tab">
              <div v-if="hasDaYun" class="analysis-block dayun-overview-card">
                <div class="block-title">大运总览</div>
                <span class="block-desc"
                  >每步大运管十年：这里记录你的起运时间、排运方向，以及每一步的干支和十神</span
                >
                <div class="dayun-summary-grid">
                  <div class="dayun-summary-item">
                    <span class="dayun-summary-label">精确起运时间</span>
                    <strong>{{ dayunStartLabel }}</strong>
                  </div>
                  <div class="dayun-summary-item">
                    <span class="dayun-summary-label">排运方向</span>
                    <strong>{{ daYun?.direction || '未判' }}</strong>
                  </div>
                  <div class="dayun-summary-item">
                    <span class="dayun-summary-label">当前年龄</span>
                    <strong>{{ currentAge !== null ? `${currentAge}岁` : '未提供' }}</strong>
                  </div>
                  <div class="dayun-summary-item">
                    <span class="dayun-summary-label">当前大运</span>
                    <strong>{{
                      currentDayunStage?.pillar || currentDayunLayer?.pillar || '未定位'
                    }}</strong>
                  </div>
                </div>
                <div class="dayun-method-notes">
                  <span>依据出生时刻与节气距离定起运</span>
                  <span>按性别与年干阴阳定顺逆</span>
                  <span>每十年推进一组干支</span>
                </div>
              </div>

              <div
                v-if="currentDayunStage || currentDayunLayer"
                class="analysis-block dayun-current-card"
              >
                <div class="dayun-current-head">
                  <div>
                    <div class="block-title">你现在走哪一步</div>
                    <span class="block-desc"
                      >当前这步大运的干支、年龄段和十神，以及它与传统说法的对应</span
                    >
                  </div>
                  <span v-if="currentDayunLayer" class="dayun-current-score">结构已记录</span>
                </div>
                <div class="dayun-current-body">
                  <div class="dayun-current-pillar">
                    <span>{{ currentDayunStage?.gan || currentDayunLayer?.gan }}</span>
                    <span>{{ currentDayunStage?.zhi || currentDayunLayer?.zhi }}</span>
                  </div>
                  <div class="dayun-current-copy">
                    <p v-if="dayunCurrentSummary" class="dayun-current-plain">
                      {{ dayunCurrentSummary }}
                    </p>
                    <div class="dayun-current-tags">
                      <span v-if="currentDayunStage" class="dayun-mini-chip"
                        >{{ currentDayunStage.startAge }}-{{ currentDayunStage.endAge }}岁</span
                      >
                      <span v-if="currentDayunStage?.tenGod" class="dayun-mini-chip"
                        >十神 {{ currentDayunStage.tenGod
                        }}{{
                          tenGodKeyword(currentDayunStage.tenGod)
                            ? ` · ${tenGodKeyword(currentDayunStage.tenGod)}`
                            : ''
                        }}</span
                      >
                      <span v-else-if="currentDayunLayer?.ten_god.name" class="dayun-mini-chip"
                        >十神 {{ currentDayunLayer.ten_god.name
                        }}{{
                          tenGodKeyword(currentDayunLayer.ten_god.name)
                            ? ` · ${tenGodKeyword(currentDayunLayer.ten_god.name)}`
                            : ''
                        }}</span
                      >
                      <span v-if="currentDayunStage?.ganElement" class="dayun-mini-chip"
                        >天干 {{ currentDayunStage.ganElement }}</span
                      >
                    </div>
                    <p v-if="currentDayunLayer">
                      {{ calculationBasisLabel(currentDayunLayer.basis) }}
                    </p>
                    <p v-else>这里展示大运干支、年龄区间、五行与十神之间的对应关系。</p>
                    <div v-if="currentDayunLayer?.relations.length" class="dayun-evidence-list">
                      <span
                        v-for="item in currentDayunLayer.relations"
                        :key="item.rule_id + item.target"
                        class="dayun-evidence-chip"
                        >{{ relationEndpointLabel(item.source) }}{{ item.source_value }} ·
                        {{ relationDisplayLabel(item.type, item.name) }} ·
                        {{ relationEndpointLabel(item.target) }}{{ item.target_value }}</span
                      >
                    </div>
                  </div>
                </div>
              </div>

              <div v-if="dayunStages.length" class="analysis-block dayun-stage-block">
                <div class="block-title">十年阶段明细</div>
                <span class="block-desc"
                  >每一步展示年龄段、大致年份、干支五行与十神；十神后面的短词是它的白话关键词，不代表结果判断</span
                >
                <div class="dayun-stage-grid">
                  <article
                    v-for="stage in dayunStages"
                    :key="stage.index + stage.pillar"
                    class="dayun-stage-card"
                    :class="{ 'is-current': stage.isCurrent }"
                  >
                    <div class="dayun-stage-top">
                      <span v-if="stage.startYear && stage.endYear"
                        >{{ stage.startYear }}–{{ stage.endYear }} · {{ stage.startAge }}–{{
                          stage.endAge
                        }}
                        岁</span
                      >
                      <span v-else>{{ stage.startAge }}–{{ stage.endAge }} 岁</span>
                      <span v-if="stage.isCurrent" class="dayun-now-badge">当前</span>
                    </div>
                    <div class="dayun-stage-main">
                      <div class="dayun-stage-pillar">
                        <span
                          class="dayun-stage-gan"
                          :style="{ color: elemColor(stage.ganElement) }"
                          >{{ stage.gan }}</span
                        >
                        <span
                          class="dayun-stage-zhi"
                          :style="{ color: elemColor(stage.zhiElement) }"
                          >{{ stage.zhi }}</span
                        >
                      </div>
                      <div class="dayun-stage-info">
                        <div class="dayun-stage-title">{{ stage.pillar }}大运</div>
                        <div class="dayun-stage-tags">
                          <span
                            >十神 {{ stage.tenGod
                            }}{{
                              tenGodKeyword(stage.tenGod) ? ` · ${tenGodKeyword(stage.tenGod)}` : ''
                            }}</span
                          >
                          <span>干 {{ stage.ganElement || '未知' }}</span>
                          <span>支 {{ stage.zhiElement || '未知' }}</span>
                        </div>
                      </div>
                    </div>
                  </article>
                </div>
              </div>
            </div>
          </ChartSection>
          <!-- /dayun section -->

          <!-- ═══ Section: 十神结构 (shishen) ═══ -->
          <ChartSection
            v-if="hasShishenSection"
            id="chart-section-shishen"
            :eyebrow="sectionNo('shishen')"
            title="十神结构"
            desc="十神 = 以日主（你自己）为参照的角色标签；这里数每个标签出现几次，次数不代表性格、职业或具体事件。"
            v-reveal
          >
            <!-- 十神分布：堆叠占比条 + 紧凑 chips，合并原占比图与纵向列表两处重复表达 -->
            <div v-if="tenGodShares.length" class="analysis-block">
              <div class="block-title">十神分布</div>
              <span class="block-desc"
                >统计范围：浮在天干上的（透干）和藏在地支里的（藏干）都算一次；藏干深浅和月令强度尚未加权</span
              >
              <p v-if="chart.ten_god_analysis?.status === 'observed'" class="tg-summary-line">
                共记录 {{ chart.ten_god_analysis.total_occurrences }} 次 · 最高频
                {{ chart.ten_god_analysis.dominant_gods.join('、') }}（{{
                  chart.ten_god_analysis.dominant_percent
                }}%）
              </p>
              <div class="tg-stack" role="img" aria-label="十神占比堆叠条">
                <span
                  v-for="item in tenGodStackSegs"
                  :key="item.name"
                  class="tg-stack-seg"
                  :style="{ width: item.percent + '%', background: tenGodColor(item.name) }"
                  :title="`${item.name} ${item.percent}%`"
                ></span>
              </div>
              <div class="tg-chip-grid">
                <div
                  v-for="item in tenGodShares"
                  :key="item.name"
                  class="tg-chip"
                  :class="{ 'is-zero': item.percent <= 0 }"
                >
                  <span class="tg-chip-dot" :style="{ background: tenGodColor(item.name) }"></span>
                  <span class="tg-chip-name">{{ item.name }}</span>
                  <span class="tg-chip-count">{{ item.count }} 次</span>
                  <span class="tg-chip-pct">{{ item.percent }}%</span>
                </div>
              </div>
              <!-- 白话解读：主导特质 + 未出现十神 + 口径免责，与上方统计共用色板 -->
              <div v-if="tenGodInsightItems.length || tenGodAbsentSentence" class="tg-insight">
                <div class="tg-insight-title">白话解读</div>
                <div v-for="item in tenGodInsightItems" :key="item.name" class="tg-insight-item">
                  <span class="tg-insight-head">
                    <span
                      class="tg-chip-dot"
                      :style="{ background: tenGodColor(item.name) }"
                    ></span>
                    <span class="tg-insight-name">{{ item.name }}</span>
                    <span v-if="item.keyword" class="tg-insight-keyword">{{ item.keyword }}</span>
                  </span>
                  <p class="tg-insight-text">
                    {{ item.trait }}；<span class="tg-insight-caution"
                      >留意：{{ item.caution }}。</span
                    >
                  </p>
                </div>
                <p v-if="tenGodAbsentSentence" class="tg-insight-absent">
                  {{ tenGodAbsentSentence }}
                </p>
                <p class="tg-insight-note">
                  以上是按出现次数的传统取象解读，描述倾向而非定论；次数统计未加权藏干深浅与月令强度。
                </p>
              </div>
              <p class="section-boundary-note">
                次数只表示命盘中的分布，不直接代表性格、职业或具体事件。
              </p>
            </div>
          </ChartSection>
          <!-- /shishen section -->

          <!-- ═══ Section: 五行格局 (wuxing) ═══ -->
          <ChartSection
            v-if="hasWuxingSection"
            id="chart-section-wuxing"
            :eyebrow="sectionNo('wuxing')"
            title="五行格局"
            desc="木、火、土、金、水在你的命盘里各占多少——先看一句白话总结，再看技术明细。"
            v-reveal
          >
            <p v-if="wuxingSummary" class="plain-summary">
              {{ wuxingSummary }}传统认为五行均衡为佳，偏多偏少各有特点，不代表好坏。
            </p>

            <!-- Element Detail -->
            <div v-if="chart.element_detail && chart.element_detail.length" class="analysis-block">
              <div class="block-title">五行权重与藏干记录</div>
              <span class="block-desc"
                >技术细节：按固定权重计分（藏干 = 地支内部"藏着"的天干），仅用于命盘内部比较</span
              >
              <div class="element-detail-table">
                <div class="ed-header">
                  <span>五行</span>
                  <span>天干</span>
                  <span>地支藏干</span>
                  <span>权重合计</span>
                </div>
                <div
                  v-for="ed in chart.element_detail"
                  :key="ed.element"
                  class="ed-row"
                  :class="'level-' + strengthLevel(ed.total)"
                >
                  <span class="ed-elem" :style="{ color: elemColor(ed.element) }">{{
                    ed.element
                  }}</span>
                  <span class="ed-tg">{{ ed.tian_gan }}</span>
                  <span class="ed-zc">{{
                    ed.cang_gan_list ? ed.cang_gan_list.join('、') : '—'
                  }}</span>
                  <span class="ed-total">{{ ed.total }}</span>
                </div>
              </div>
            </div>

            <!-- MissingElements 原始五行分布 -->
            <div
              v-if="
                chart.missing_elements &&
                (chart.missing_elements.missing_elements?.length ||
                  chart.missing_elements.weak_elements?.length)
              "
              class="analysis-block"
            >
              <div class="block-title">五行分布</div>
              <span class="block-desc"
                >按原始计分看哪些五行没出现或偏弱；只是观察记录，不代表需要"补"什么</span
              >
              <div class="missing-elem-card">
                <div v-if="chart.missing_elements.missing_elements?.length" class="me-row">
                  <span class="me-label">原局未见</span>
                  <span
                    v-for="e in chart.missing_elements.missing_elements"
                    :key="'m' + e"
                    class="me-tag me-missing"
                    :style="{ color: elemColor(e), borderColor: elemColor(e) + '44' }"
                    >{{ e }}</span
                  >
                </div>
                <div v-if="chart.missing_elements.weak_elements?.length" class="me-row">
                  <span class="me-label">得分偏低</span>
                  <span
                    v-for="e in chart.missing_elements.weak_elements"
                    :key="'w' + e"
                    class="me-tag me-weak"
                    :style="{ color: elemColor(e), borderColor: elemColor(e) + '44' }"
                    >{{ e }}</span
                  >
                </div>
                <p class="me-note">{{ chart.missing_elements.note }}</p>
              </div>
            </div>
          </ChartSection>
          <!-- /wuxing section -->

          <!-- ═══ Section: 神煞 (shensha) ═══ -->
          <ChartSection
            v-if="hasShenshaSection"
            id="chart-section-shensha"
            :eyebrow="sectionNo('shensha')"
            title="神煞"
            desc="神煞 = 传统命理的查表符号，可理解为“传统说法里的标记”，不自动推断吉凶。"
            :summary="`规则命中 ${shenShaDetails.length} 项 · 按四柱干支对应传统表项`"
            v-reveal
          >
            <p v-if="shenShaSummary" class="plain-summary">{{ shenShaSummary }}</p>

            <section v-if="shenShaDetails.length" class="analysis-block shensha-overview-card">
              <div class="shensha-overview-head">
                <div>
                  <div class="block-title">神煞规则命中</div>
                  <span class="block-desc"
                    >每一项都是按传统查表命中的标记，后面附它的传统寓意；只是参考，不代表吉凶、性格或具体事件</span
                  >
                </div>
                <span class="shensha-overview-count">{{ shenShaDetails.length }} 项</span>
              </div>
              <div class="shensha-grid">
                <article
                  v-for="item in shenShaDetails"
                  :key="item.rule_id"
                  class="shen-sha-detail-row shensha-detail-card"
                >
                  <div class="shen-sha-detail-top">
                    <span class="shen-sha-detail-name">{{ item.name }}</span>
                    <span class="shen-sha-detail-meta">规则命中</span>
                  </div>
                  <span class="shen-sha-detail-desc">{{ shenShaMeaning(item.name) }}</span>
                </article>
              </div>
            </section>

            <!-- ShenSha (grouped by pillar when available) -->
            <template v-if="groupedShenSha">
              <template v-for="group in groupedShenSha" :key="group.pillar">
                <section
                  v-if="group.items && group.items.length"
                  class="analysis-block shensha-group-card"
                >
                  <div class="shen-sha-group-title shensha-group-head">
                    <span
                      class="shen-sha-group-dot"
                      :style="{ background: pillarShenShaColor(group.pillar) }"
                    ></span>
                    <span class="shen-sha-group-label">{{ group.label }}神煞</span>
                    <span class="shen-sha-group-role">· {{ group.gan }}{{ group.zhi }}</span>
                  </div>
                  <div class="shen-sha-list">
                    <article
                      v-for="sha in group.items"
                      :key="sha.name + sha.target + sha.desc"
                      class="shen-sha-row shensha-row-card"
                      :style="{ background: pillarShenShaBg(group.pillar) }"
                    >
                      <span
                        class="shen-sha-name"
                        :style="{ color: pillarShenShaColor(group.pillar) }"
                        >{{ sha.name }}</span
                      >
                      <span v-if="sha.target" class="shen-sha-target">{{ sha.target }}</span>
                      <span class="shen-sha-desc">{{ sha.desc || shenShaMeaning(sha.name) }}</span>
                    </article>
                  </div>
                </section>
              </template>
              <section v-if="globalShenSha.length" class="analysis-block shensha-group-card">
                <div class="shen-sha-group-title shensha-group-head">
                  <span class="shen-sha-group-dot" style="background: var(--accent)"></span>
                  <span class="shen-sha-group-label">全局组合神煞</span>
                  <span class="shen-sha-group-role">· 多柱配合</span>
                </div>
                <div class="shen-sha-list">
                  <article
                    v-for="sha in globalShenSha"
                    :key="sha.name + sha.target + sha.desc"
                    class="shen-sha-row shensha-row-card"
                    :style="{ background: 'var(--accent-dim)' }"
                  >
                    <span class="shen-sha-name" :style="{ color: 'var(--accent)' }">{{
                      sha.name
                    }}</span>
                    <span v-if="sha.target" class="shen-sha-target">{{ sha.target }}</span>
                    <span class="shen-sha-desc">{{ sha.desc || shenShaMeaning(sha.name) }}</span>
                  </article>
                </div>
              </section>
            </template>
            <section v-else-if="parsedDayShenSha.length" class="analysis-block shensha-group-card">
              <div class="block-title">日柱神煞</div>
              <span class="block-desc">按日柱相关查表规则记录命中的传统名称与目标支</span>
              <div class="shen-sha-list">
                <article
                  v-for="sha in parsedDayShenSha"
                  :key="sha.name + sha.target + sha.desc"
                  class="shen-sha-row shensha-row-card"
                >
                  <span class="shen-sha-name">{{ sha.name }}</span>
                  <span v-if="sha.target" class="shen-sha-target">{{ sha.target }}</span>
                  <span class="shen-sha-desc">{{ sha.desc || shenShaMeaning(sha.name) }}</span>
                </article>
              </div>
            </section>
          </ChartSection>
          <!-- /shensha section -->

          <!-- ═══ Section: 周期结构 (fortune) ═══ -->
          <ChartSection
            v-if="hasFortuneSection"
            id="chart-section-fortune"
            :eyebrow="sectionNo('fortune')"
            title="周期结构"
            desc="流年（这一年）、流月（这一月）、小运与命宫的干支记录，用来对照周期节奏，不直接生成吉凶结论。"
            :summary="fortuneSectionSummary"
            :default-open="false"
            v-reveal
          >
            <div class="fortune-detail-tab">
              <div v-if="fortuneLayerList.length" class="analysis-block">
                <div class="block-title">周期层依据</div>
                <span class="block-desc"
                  >流年 = 这一年的干支，流月 = 这一月的干支，小运 =
                  一年一换的辅助周期；这里记录它们与命局的结构关系，不直接生成吉凶结论</span
                >
                <div class="fortune-layer-list">
                  <div v-for="layer in fortuneLayerList" :key="layer.key" class="fortune-layer-row">
                    <div class="fortune-layer-top">
                      <span class="fortune-layer-name">{{ layer.name }}</span>
                      <span class="fortune-layer-pillar">{{ layer.pillar }}</span>
                      <span v-if="layer.ten_god.name" class="fortune-layer-god"
                        >十神 {{ layer.ten_god.name }}</span
                      >
                    </div>
                    <div class="fortune-layer-sub">
                      <span v-if="layer.start_age !== undefined" class="fortune-layer-chip"
                        >{{ layer.start_age }}岁起</span
                      >
                      <span v-if="layer.year" class="fortune-layer-chip">{{ layer.year }}年</span>
                      <span v-if="layer.month" class="fortune-layer-chip">{{ layer.month }}月</span>
                      <span class="fortune-layer-chip">结构参考</span>
                    </div>
                    <div class="fortune-layer-evidence">
                      {{ calculationBasisLabel(layer.basis)
                      }}<span v-if="layer.relations.length">
                        · {{ layer.relations.length }} 条关系</span
                      >
                    </div>
                  </div>
                </div>
              </div>

              <!-- MingGong -->
              <div v-if="chart.ming_gong && chart.ming_gong.gan_zhi" class="analysis-block">
                <div class="block-title">命宫 · 第五柱</div>
                <span class="block-desc"
                  >命宫 = 按出生时辰推算的"第五柱"，在四柱之外，传统上用来辅助看整体倾向</span
                >
                <div class="ming-gong-detail">
                  <div class="ming-gong-main">
                    <span class="ming-gong-ganzhi">{{ chart.ming_gong.gan_zhi }}</span>
                    <span v-if="chart.ming_gong.nayin" class="ming-gong-nayin"
                      >({{ chart.ming_gong.nayin }})</span
                    >
                  </div>
                  <div v-if="chart.ming_gong.shen_sha" class="ming-gong-shensha">
                    <span class="shensha-badge"> {{ chart.ming_gong.shen_sha }}星 </span>
                  </div>
                </div>
              </div>

              <div v-if="chart.month_season?.status === 'observed'" class="analysis-block">
                <div class="block-title">月令季节</div>
                <span class="block-desc"
                  >月令 =
                  出生月份的地支，传统上视为命盘的"季节背景"；这里按月柱地支记录传统月序与四季归属</span
                >
                <dl class="month-season-facts">
                  <div>
                    <dt>月支</dt>
                    <dd>{{ chart.month_season.month_branch }}</dd>
                  </div>
                  <div>
                    <dt>传统月序</dt>
                    <dd>第 {{ chart.month_season.traditional_month }} 月</dd>
                  </div>
                  <div>
                    <dt>季节</dt>
                    <dd>{{ chart.month_season.season }}</dd>
                  </div>
                  <div>
                    <dt>取值依据</dt>
                    <dd>月柱地支</dd>
                  </div>
                </dl>
              </div>
            </div>
          </ChartSection>
          <!-- /fortune section -->

          <!-- ═══ Section: 传统参考 (pattern) ═══ -->
          <ChartSection
            v-if="hasPatternSection"
            id="chart-section-pattern"
            :eyebrow="sectionNo('pattern')"
            title="传统参考"
            desc="格局（命盘结构的传统归类）、身强身弱（日主力量强弱）与调候（传统认为需要调和的五行）的候选参考口径。"
            :summary="patternSectionSummary"
            :default-open="false"
            v-reveal
          >
            <!-- Pattern Analysis -->
            <div v-if="chart.pattern_analysis" class="analysis-block">
              <div class="block-title">格局线索</div>
              <span class="block-desc"
                >格局 =
                传统对命盘整体结构的归类叫法；下面是根据四柱与月令整理出的候选线索，不是定论</span
              >
              <div class="pattern-detail">
                <div class="pattern-inputs">
                  <div class="pattern-input-row">
                    <span>四柱输入</span>
                    <strong>{{ chart.pattern_analysis.inputs.pillars.join('、') }}</strong>
                  </div>
                  <div class="pattern-input-row">
                    <span>月支</span>
                    <strong>{{ chart.pattern_analysis.inputs.month_branch }}</strong>
                  </div>
                </div>
                <div v-if="chart.pattern_analysis.candidates?.length" class="pattern-candidates">
                  <div class="pattern-candidates-title">结构线索</div>
                  <div
                    v-for="candidate in chart.pattern_analysis.candidates"
                    :key="candidate.rule_id"
                    class="pattern-candidate-row"
                  >
                    <div class="pattern-candidate-heading">
                      <strong>{{ candidate.pattern_name }}</strong>
                      <span>{{ candidate.category }}</span>
                    </div>
                    <small>按月令、藏干与透干关系整理</small>
                  </div>
                </div>
                <div
                  v-if="chart.pattern_analysis.month_command_evidence?.length"
                  class="pattern-candidates"
                >
                  <div class="pattern-candidates-title">月令藏干透出证据</div>
                  <div
                    v-for="evidence in chart.pattern_analysis.month_command_evidence"
                    :key="evidence.hidden_stem + monthCommandExposureLabel(evidence.exposures)"
                    class="pattern-candidate-row"
                  >
                    <div class="pattern-candidate-heading">
                      <strong>{{ evidence.candidate_names.join('、') }}</strong>
                      <span>月令线索</span>
                      <span v-if="evidence.month_special_structure">{{
                        evidence.month_special_structure
                      }}</span>
                    </div>
                    <small>
                      {{ evidence.month_branch }}中{{ evidence.hidden_stem }} ·
                      {{ evidence.hidden_stem_type }} · {{ evidence.hidden_ten_god }} · 透于
                      {{ monthCommandExposureLabel(evidence.exposures) }}
                    </small>
                    <small>按月令、藏干与透干关系整理</small>
                  </div>
                </div>
                <p class="section-boundary-note">
                  格局线索用于辅助理解命盘结构，不等同于已经确定格局或喜忌。
                </p>
              </div>
            </div>

            <!-- Body Strength -->
            <div v-if="chart.body_strength" class="analysis-block">
              <div class="block-title">五行强弱参考</div>
              <span class="block-desc"
                >身强 / 身弱 =
                日主（代表你自己）的力量偏强还是偏弱；下面是本地规则给出的候选判断，尚未独立验证</span
              >
              <div class="body-strength">
                <div class="bs-primary">
                  <span>规则候选区间</span>
                  <strong>{{
                    strengthCandidateLabel(chart.body_strength.score_band_candidate)
                  }}</strong>
                </div>
                <div class="bs-contract-grid">
                  <div>
                    <span>四柱输入</span>
                    <strong>{{ chart.body_strength.inputs.pillars.join('、') }}</strong>
                  </div>
                  <div>
                    <span>日主 / 月支</span>
                    <strong
                      >{{ chart.body_strength.inputs.day_stem
                      }}{{ chart.body_strength.inputs.day_element }} /
                      {{ chart.body_strength.inputs.month_branch }}</strong
                    >
                  </div>
                </div>
                <div v-if="bodyStrengthComponents.length" class="evidence-bars body-strength-bars">
                  <div
                    v-for="component in bodyStrengthComponents"
                    :key="component.key"
                    class="evidence-bar-row"
                  >
                    <span class="evidence-bar-label">{{ component.name }}</span>
                    <div class="evidence-bar-track">
                      <span
                        class="evidence-bar-fill"
                        :style="{ width: componentWidth(component.normalized_score) }"
                      ></span>
                    </div>
                    <span class="evidence-bar-value">{{
                      componentBandLabel(component.normalized_score)
                    }}</span>
                  </div>
                </div>
                <p class="section-boundary-note">
                  当前规则尚未完成独立验证；区间与百分比只用于复核本地计算，不直接决定身强身弱、喜忌五行或现实结果。
                </p>
              </div>
            </div>

            <!-- Tiaohou table evidence -->
            <div v-if="chart.tiaohou" class="analysis-block">
              <div class="block-title">
                调候参考 <span class="tiaohou-source">《穷通宝鉴》资料表</span>
              </div>
              <span class="block-desc"
                >调候 =
                传统上认为命局需要调和的五行；按日干与月支查阅《穷通宝鉴》资料表，结合出生时刻所在节令区间整理候选</span
              >
              <div class="tiaohou-card">
                <div class="tiaohou-header">
                  <span class="tiaohou-stem">{{ chart.tiaohou.stem }}</span>
                  <span class="tiaohou-arrow">生</span>
                  <span class="tiaohou-month">{{ chart.tiaohou.month }}</span>
                  <span class="tiaohou-divider">|</span>
                  <span class="tiaohou-label">资料表首项</span>
                  <span class="tiaohou-primary" :style="{ color: elemColor(tiaohouElem) }">{{
                    chart.tiaohou.table_primary_candidate
                  }}</span>
                </div>
                <div class="tiaohou-summary">
                  <template v-if="chart.tiaohou.depth_evidence.status === 'observed'">
                    {{ chart.tiaohou.depth_evidence.start_term }} 至
                    {{ chart.tiaohou.depth_evidence.end_term }} ·
                    {{ chart.tiaohou.depth_evidence.phase }} ·
                    {{ ((chart.tiaohou.depth_evidence.position || 0) * 100).toFixed(1) }}%
                  </template>
                  <template v-else>出生时刻不可定位，节令区间深浅不可用</template>
                </div>
                <div v-if="chart.tiaohou.matched_conditions?.length" class="tiaohou-chart-matches">
                  <div class="tiaohou-chart-heading">
                    <strong>完整命局条件已命中</strong>
                    <span>可参考 {{ chart.tiaohou.chart_candidates.join('、') }}</span>
                  </div>
                  <div
                    v-for="condition in chart.tiaohou.matched_conditions"
                    :key="condition.rule_id"
                    class="tiaohou-chart-match"
                  >
                    <div class="tiaohou-chart-candidates">
                      <span v-for="candidate in condition.candidates" :key="candidate">{{
                        candidate
                      }}</span>
                    </div>
                    <div>
                      <strong>{{ condition.condition }}</strong>
                      <p>{{ condition.source_text }}</p>
                      <small>{{ condition.evidence.join('；') }}</small>
                    </div>
                  </div>
                </div>
                <div v-if="chart.tiaohou.rules && chart.tiaohou.rules.length" class="tiaohou-rules">
                  <div v-for="rule in chart.tiaohou.rules" :key="rule.rule_id" class="tiaohou-rule">
                    <span class="tiaohou-xi">原表用字 {{ rule.xi_shen }}</span>
                    <span class="tiaohou-reason">{{ rule.source_text }}</span>
                  </div>
                </div>
                <div
                  v-if="chart.tiaohou.depth_evidence.month_command_candidates?.length"
                  class="month-command-candidates"
                >
                  <div class="month-command-heading">
                    <strong>四库月分日司令参考</strong>
                    <span>不同古籍口径并列</span>
                  </div>
                  <div
                    v-for="candidate in chart.tiaohou.depth_evidence.month_command_candidates"
                    :key="candidate.rule_id"
                    class="month-command-row"
                  >
                    <div class="month-command-current">
                      <strong>{{ candidate.commanding_stem }}</strong>
                      <span>{{ candidate.segment }}</span>
                      <small>传统分日资料口径</small>
                    </div>
                    <p>
                      入节约第 {{ Math.max(1, Math.round(candidate.position_day)) }} 天 ·
                      {{
                        monthCommandSegmentLabel(
                          candidate.segment_start_day,
                          candidate.segment_end_day,
                        )
                      }}
                      · {{ candidate.sequence }}
                    </p>
                  </div>
                </div>
                <p class="section-boundary-note">
                  调候条目尚未完成独立验证，是并列的传统查表候选，不代表唯一用神或现实吉凶；节令百分比仅表示时间位置。
                </p>
              </div>
            </div>
          </ChartSection>
          <!-- /pattern section -->

          <!-- ═══ Section: 经典依据 (rules) ═══ -->
          <ChartSection
            v-if="chart.id"
            id="chart-section-rules"
            :eyebrow="sectionNo('rules')"
            title="经典依据"
            desc="每条结论对应的古籍出处与本地计算过程，供想深究的人核对。"
            summary="按规则逐条对照古籍出处与本地计算日志"
            :default-open="false"
            v-reveal
          >
            <div class="classical-rules-block">
              <ClassicalInterpretationPanel :chart-id="chart.id" />
            </div>
          </ChartSection>
          <!-- /rules section -->
        </div>
        <!-- /analysis-section -->
      </div>
    </div>
  </div>
</template>

<style scoped>
.bazi-chart {
  position: relative;
}

/* 长滚动布局：左侧锚点目录 + 内容卡片 */
.chart-layout {
  display: grid;
  grid-template-columns: 176px minmax(0, 1fr);
  gap: 2.25rem;
  align-items: start;
}

.chart-toc {
  position: sticky;
  top: 96px;
  min-width: 0;
}

.toc-inner {
  display: flex;
  flex-direction: column;
  border-left: 1px solid var(--line-subtle);
}

.toc-item {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
  padding: 0.42rem 0 0.42rem 0.85rem;
  margin-left: -1px;
  background: none;
  border: 0;
  border-left: 1px solid transparent;
  font: inherit;
  text-align: left;
  cursor: pointer;
  color: var(--text-soft);
  transition:
    color 0.2s,
    border-color 0.2s;
}

.toc-item:hover {
  color: var(--text);
}

.toc-item:focus-visible {
  outline: 2px solid var(--line-focus);
  outline-offset: 2px;
  border-radius: 2px;
}

.toc-item.active {
  color: var(--accent);
  border-left-color: var(--accent);
}

.toc-label {
  font-size: var(--fs-xs);
  font-weight: 600;
  letter-spacing: 1px;
}

.toc-sub {
  font-size: var(--fs-2xs);
  color: var(--text-soft);
  letter-spacing: 0.3px;
  white-space: nowrap;
}

.toc-item.active .toc-sub {
  color: var(--text-muted);
}

.chart-card {
  position: relative;
  z-index: 1;
  min-width: 0;
}

.chart-card.glass-card:hover {
  box-shadow:
    0 0 0 1px var(--line-strong),
    var(--shadow-xs);
  transform: none;
}

/* Header */
.chart-header {
  background: linear-gradient(180deg, var(--line-subtle), transparent);
  border-bottom: 1px solid var(--line-subtle);
  padding: 1rem 1.25rem;
  text-align: center;
}

.header-eyebrow {
  font-size: var(--fs-2xs);
  letter-spacing: 3px;
  color: var(--text-soft);
  text-transform: uppercase;
  margin-bottom: 4px;
}

.chart-title {
  font-family: var(--font-serif), serif;
  font-size: var(--fs-2xl);
  font-weight: 700;
  color: var(--text);
  margin: 0;
  letter-spacing: 3px;
}

.chart-identity {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0;
  margin: 0;
  padding: 0 0.75rem 0.75rem;
  border-bottom: 1px solid var(--line-subtle);
}

.chart-identity > div {
  min-width: 0;
  padding: 0.65rem 0.75rem;
  border-right: 1px solid var(--line-subtle);
}

.chart-identity > div:nth-child(3n) {
  border-right: 0;
}

.chart-identity dt {
  color: var(--text-soft);
  font-size: var(--fs-2xs);
}

.chart-identity dd {
  margin: 0.2rem 0 0;
  overflow-wrap: anywhere;
  color: var(--text);
  font-size: var(--fs-xs);
  font-weight: 700;
}

/* Pillars bento grid */
.pillars-bento {
  display: grid;
  grid-template-columns:
    minmax(96px, 0.86fr)
    minmax(96px, 0.86fr)
    minmax(210px, 1.35fr)
    minmax(96px, 0.86fr)
    minmax(96px, 0.86fr);
  gap: 0.55rem;
  padding: 0.75rem;
  border-bottom: 1px solid var(--line-subtle);
  align-items: stretch;
}

.pillars-bento > .bento-card {
  min-height: 188px;
}

.bento-card {
  background: var(--accent-dim);
  border: 1px solid var(--line-subtle);
  border-radius: 10px;
  padding: 0.6rem 0.75rem;
  transition:
    box-shadow 0.3s,
    background 0.3s;
  position: relative;
  overflow: hidden;
}

.bento-small {
  display: flex;
  flex-direction: column;
}

.bento-card.bento-small:hover,
.bento-card.bento-day-card:hover {
  background: var(--surface-2);
}

/* Hover 五行微光 */
.bento-hover-木:hover {
  box-shadow: inset 0 0 30px color-mix(in oklab, var(--wuxing-mu) 7%, transparent);
}
.bento-hover-火:hover {
  box-shadow: inset 0 0 30px color-mix(in oklab, var(--wuxing-huo) 7%, transparent);
}
.bento-hover-土:hover {
  box-shadow: inset 0 0 30px color-mix(in oklab, var(--wuxing-tu) 7%, transparent);
}
.bento-hover-金:hover {
  box-shadow: inset 0 0 30px color-mix(in oklab, var(--wuxing-jin) 7%, transparent);
}
.bento-hover-水:hover {
  box-shadow: inset 0 0 30px color-mix(in oklab, var(--wuxing-shui) 7%, transparent);
}

.bento-label {
  font-size: var(--fs-2xs);
  color: var(--text-muted);
  letter-spacing: 1px;
  margin-bottom: 0.4rem;
  font-weight: 600;
}

/* 四柱白话定位小字：年柱「祖上与少年」等 */
.bento-role {
  margin: -0.3rem 0 0.35rem;
  font-size: var(--fs-2xs);
  color: var(--text-soft);
  letter-spacing: 0.3px;
  line-height: 1.4;
}

.bento-body {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.35rem;
  flex: 1;
  justify-content: center;
}

.bento-gan,
.bento-zhi {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.15rem;
}

.bento-char {
  font-family: var(--font-serif), serif;
  font-size: var(--fs-3xl);
  font-weight: 700;
  line-height: 1;
}

.bento-day-card {
  border-color: var(--line-focus);
  background: linear-gradient(180deg, var(--accent-dim), transparent 145%), var(--glass-bg);
}

/* 五行雷达图卡片 */
.bento-radar {
  display: flex;
  flex-direction: column;
  order: 3;
  min-height: 188px;
  padding-bottom: 0.35rem;
}

.bento-radar-chart {
  flex: 1;
  min-height: 148px;
  height: 148px;
  width: 100%;
}

.bento-radar-note {
  margin: 0.25rem 0 0;
  color: var(--text-muted);
  font-size: var(--fs-sm);
  line-height: 1.5;
}

.bento-sub {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.1rem;
  padding-top: 0.25rem;
  border-top: 1px solid var(--line-subtle);
  margin-top: 0.25rem;
}

.elem-tag {
  display: inline-block;
  font-size: var(--fs-2xs);
  padding: 0.1rem 0.35rem;
  border-radius: 3px;
  border: 1px solid;
  font-weight: 600;
  letter-spacing: 0.5px;
}

.bento-god-tag {
  font-size: var(--fs-2xs);
  padding: 0.1rem 0.4rem;
  border-radius: 4px;
  background: var(--surface-2);
  border: 1px solid var(--line-strong);
  color: var(--text-muted);
  font-weight: 600;
  letter-spacing: 0.5px;
  margin-top: 0.2rem;
}

.bento-god-tag-day {
  font-size: var(--fs-2xs);
  padding: 0.15rem 0.5rem;
  background: var(--line-strong);
  border-color: var(--line-focus);
  color: var(--accent);
}

/* ===== 天干地支关系 — information ledger ===== */

.ganzhi-analysis {
  margin-top: 2px;
}

.overview-relations {
  --rel-he: #16845a;
  --rel-ke: #b94a48;
  --rel-sheng: #2f6fb7;
  --rel-hai: #a86622;
  --rel-xing: #7e5aa8;
}

.relations-section {
  padding: 0.9rem 1rem 1rem;
  border-bottom: 1px solid var(--line-subtle);
}

.relation-boundary-note,
.section-boundary-note {
  margin: 0;
  padding: 0.72rem 1rem;
  border-top: 1px solid var(--line-subtle);
  color: var(--text-soft);
  font-size: var(--fs-2xs);
  line-height: 1.6;
}

.analysis-block .section-boundary-note {
  padding-right: 0;
  padding-left: 0;
}

.relations-title {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  padding-bottom: 0.68rem;
  border-bottom: 1px solid var(--line-subtle);
  font-size: var(--fs-xs);
  font-weight: 700;
  color: var(--text);
  letter-spacing: 0.04em;
  margin-bottom: 0.65rem;
}

.relations-title-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--rel-he);
  box-shadow: none;
}

.relations-title-dot.zhi-dot {
  background: var(--rel-sheng);
  box-shadow: none;
}

.relations-count {
  display: inline-flex;
  align-items: center;
  min-height: 22px;
  padding: 0 0.48rem;
  border: 1px solid var(--line-subtle);
  border-radius: 999px;
  background: color-mix(in oklab, var(--surface-0) 82%, transparent);
  font-size: var(--fs-2xs);
  font-weight: 600;
  color: var(--text-soft);
  margin-left: auto;
  letter-spacing: 0.02em;
}

.ganzhi-compact {
  display: flex;
  flex-direction: column;
  gap: 0.38rem;
}

.gz-item {
  --rel-color: var(--line-focus);
  display: grid;
  grid-template-columns: max-content max-content minmax(0, 1fr);
  align-items: center;
  gap: 0.55rem;
  min-height: 40px;
  padding: 0.48rem 0.62rem;
  border: 1px solid color-mix(in oklab, var(--rel-color) 26%, var(--line-subtle));
  border-radius: 7px;
  background: color-mix(in oklab, var(--surface-0) 58%, transparent);
  transition:
    background 0.18s ease,
    border-color 0.18s ease;
}

.gz-item:hover {
  background: color-mix(in oklab, var(--rel-color) 7%, var(--surface-0));
}

.gz-item.rel-he {
  --rel-color: var(--rel-he);
}
.gz-item.rel-ke {
  --rel-color: var(--rel-ke);
}
.gz-item.rel-sheng {
  --rel-color: var(--rel-sheng);
}
.gz-item.rel-chong {
  --rel-color: var(--rel-ke);
}
.gz-item.rel-hai {
  --rel-color: var(--rel-hai);
}
.gz-item.rel-xing {
  --rel-color: var(--rel-xing);
}
.gz-item.rel-hui {
  --rel-color: var(--rel-sheng);
}

.gz-chars {
  display: flex;
  align-items: center;
  gap: 0.15rem;
  min-width: 6.25rem;
  flex-shrink: 0;
}

.gz-c {
  font-family: var(--font-serif), serif;
  font-size: var(--fs-md);
  font-weight: 700;
  color: var(--text);
  line-height: 1.2;
}

.gz-sym {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 1.35rem;
  height: 1.35rem;
  border-radius: 999px;
  border: 1px solid transparent;
  font-size: var(--fs-2xs);
  font-weight: 700;
  line-height: 1;
}

.sym-rel-he {
  --rel-color: var(--rel-he);
}
.sym-rel-ke {
  --rel-color: var(--rel-ke);
}
.sym-rel-sheng {
  --rel-color: var(--rel-sheng);
}
.sym-rel-chong {
  --rel-color: var(--rel-ke);
}
.sym-rel-hai {
  --rel-color: var(--rel-hai);
}
.sym-rel-xing {
  --rel-color: var(--rel-xing);
}
.sym-rel-hui {
  --rel-color: var(--rel-sheng);
}

.gz-tag {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 24px;
  font-size: var(--fs-2xs);
  font-weight: 700;
  padding: 0.08rem 0.42rem;
  border: 1px solid transparent;
  border-radius: 6px;
  letter-spacing: 0.02em;
  flex-shrink: 0;
}

.tag-rel-he {
  --rel-color: var(--rel-he);
}
.tag-rel-ke {
  --rel-color: var(--rel-ke);
}
.tag-rel-sheng {
  --rel-color: var(--rel-sheng);
}
.tag-rel-chong {
  --rel-color: var(--rel-ke);
}
.tag-rel-hai {
  --rel-color: var(--rel-hai);
}
.tag-rel-xing {
  --rel-color: var(--rel-xing);
}
.tag-rel-hui {
  --rel-color: var(--rel-sheng);
}

.gz-sym,
.gz-tag {
  color: var(--rel-color);
  background: color-mix(in oklab, var(--rel-color) 11%, transparent);
  border-color: color-mix(in oklab, var(--rel-color) 18%, transparent);
}

.gz-text {
  color: var(--text-muted);
  font-size: var(--fs-xs);
  line-height: 1.5;
  min-width: 0;
  white-space: normal;
}

/* 干支关系的白话解释：独占一行的小字 */
.gz-plain {
  grid-column: 1 / -1;
  margin-top: -0.1rem;
  color: var(--text-soft);
  font-size: var(--fs-2xs);
  line-height: 1.5;
}

.no-relations {
  min-height: 132px;
  padding: 1rem;
  text-align: center;
  font-size: var(--fs-xs);
  color: var(--text-soft);
  border-bottom: 1px solid var(--line-subtle);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
}

.no-rel-icon {
  font-size: var(--fs-2xs);
  color: var(--text-soft);
}

/* Analysis sections：长滚动分区，间距由 ChartSection 的 hairline 与留白控制 */
.analysis-section {
  padding: 0.5rem 1.25rem 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 0;
}

.evidence-table-list,
.evidence-notes {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
}

.evidence-table-pill,
.evidence-note,
.fortune-layer-chip {
  display: inline-flex;
  align-items: center;
  min-height: 24px;
  padding: 0.15rem 0.48rem;
  border: 1px solid var(--line-subtle);
  border-radius: 4px;
  color: var(--text-muted);
  font-size: var(--fs-2xs);
  line-height: 1.2;
  background: var(--surface-0);
}

.evidence-bars {
  display: flex;
  flex-direction: column;
  gap: 0.42rem;
}

.evidence-bar-row {
  display: grid;
  grid-template-columns: 3.2rem minmax(80px, 1fr) 2.4rem;
  align-items: center;
  gap: 0.5rem;
}

.evidence-bar-label,
.evidence-bar-value {
  color: var(--text-muted);
  font-size: var(--fs-2xs);
}

.evidence-bar-value {
  text-align: right;
  font-family: var(--font-mono);
}

.evidence-bar-track {
  height: 6px;
  border-radius: 999px;
  background: var(--line-subtle);
  overflow: hidden;
}

.evidence-bar-fill {
  display: block;
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(90deg, var(--accent), #22d3ee);
}

.shen-sha-detail-list,
.fortune-layer-list {
  display: flex;
  flex-direction: column;
  gap: 0.45rem;
}

.shen-sha-detail-list {
  margin-top: 0.6rem;
}

.shen-sha-detail-row,
.fortune-layer-row {
  border-left: 2px solid var(--line-strong);
  padding-left: 0.55rem;
  min-width: 0;
}

.shen-sha-detail-row.sha-good,
.fortune-layer-chip.sha-good {
  border-color: rgba(22, 163, 74, 0.35);
  color: #16a34a;
}

.shen-sha-detail-row.sha-risk,
.fortune-layer-chip.sha-risk {
  border-color: rgba(220, 38, 38, 0.35);
  color: #dc2626;
}

.shen-sha-detail-row.sha-neutral {
  border-color: rgba(100, 149, 237, 0.35);
}

.shen-sha-detail-name,
.fortune-layer-name {
  display: inline-block;
  margin-right: 0.4rem;
  font-size: var(--fs-xs);
  font-weight: 700;
  color: var(--text);
}

.shen-sha-detail-meta,
.fortune-layer-god {
  font-size: var(--fs-2xs);
  color: var(--text-soft);
}

.shen-sha-detail-desc,
.fortune-layer-evidence {
  display: block;
  margin-top: 0.18rem;
  color: var(--text-muted);
  font-size: var(--fs-2xs);
  line-height: 1.45;
}

.shensha-overview-card,
.shensha-group-card {
  padding: 0.85rem 0.95rem;
  border: 1px solid var(--line-subtle);
  border-radius: 8px;
  background: color-mix(in oklab, var(--surface-2) 38%, transparent);
}

.shensha-overview-card {
  display: flex;
  flex-direction: column;
  gap: 0.8rem;
}

.shensha-overview-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.75rem;
  min-width: 0;
}

.shensha-overview-head .block-desc {
  max-width: 54rem;
}

.shensha-overview-count {
  display: inline-flex;
  align-items: center;
  min-height: 24px;
  padding: 0 0.48rem;
  border: 1px solid var(--line-subtle);
  border-radius: 999px;
  background: var(--surface-0);
  color: var(--text-soft);
  font-size: var(--fs-2xs);
  font-weight: 600;
  white-space: nowrap;
  flex-shrink: 0;
}

.shensha-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.42rem 0.55rem;
}

.shensha-detail-card {
  display: flex;
  flex-direction: column;
  gap: 0.22rem;
  padding: 0.48rem 0.55rem;
  border-radius: 8px;
  background: var(--surface-0);
}

.shen-sha-detail-top {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.3rem 0.45rem;
  min-width: 0;
}

.shensha-group-card {
  display: flex;
  flex-direction: column;
  gap: 0.65rem;
}

.shensha-group-head {
  margin-bottom: 0;
}

.shen-sha-group-label {
  min-width: 0;
}

.shensha-row-card {
  padding: 0.48rem 0.55rem;
  border-radius: 8px;
  border-left-width: 3px;
  background: var(--surface-0);
}

.fortune-layer-top {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  min-width: 0;
}

.fortune-layer-pillar {
  font-family: var(--font-serif), serif;
  color: var(--accent);
  font-weight: 700;
}

.fortune-layer-sub {
  display: flex;
  flex-wrap: wrap;
  gap: 0.3rem;
  margin-top: 0.35rem;
}

.rule-meta-card,
.rule-table-card {
  border: 1px solid var(--line-subtle);
  border-radius: 8px;
  background: color-mix(in oklab, var(--surface-2) 42%, transparent);
}

.rule-meta-card {
  display: grid;
  gap: 0.45rem;
  padding: 0.75rem;
}

.rule-meta-row,
.rule-table-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  min-width: 0;
}

.rule-meta-label,
.rule-table-source,
.rule-table-count {
  font-size: var(--fs-2xs);
  color: var(--text-soft);
}

.rule-meta-value {
  min-width: 0;
  text-align: right;
  color: var(--text);
  font-size: var(--fs-xs);
  font-family: var(--font-mono);
  overflow-wrap: anywhere;
}

.rule-weight-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem;
}

.rule-weight-chip {
  display: inline-flex;
  align-items: center;
  min-height: 24px;
  padding: 0.15rem 0.5rem;
  border: 1px solid var(--line-subtle);
  border-radius: 4px;
  background: var(--surface-0);
  color: var(--text-muted);
  font-size: var(--fs-2xs);
  font-family: var(--font-mono);
}

.rule-table-list {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.7rem;
}

.rule-table-card {
  position: relative;
  padding: 0.75rem;
  min-width: 0;
}

.rule-table-name {
  color: var(--text);
  font-size: var(--fs-xs);
  font-weight: 700;
}

.rule-table-version {
  flex-shrink: 0;
  color: var(--accent);
  font-size: var(--fs-2xs);
  font-family: var(--font-mono);
}

.rule-table-source {
  margin-top: 0.35rem;
}

.rule-table-desc {
  margin: 0.4rem 0 0;
  color: var(--text-muted);
  font-size: var(--fs-2xs);
  line-height: 1.55;
}

.rule-table-count {
  display: inline-flex;
  margin-top: 0.55rem;
  padding: 0.1rem 0.42rem;
  border: 1px solid var(--line-subtle);
  border-radius: 4px;
  background: var(--surface-0);
}

/* 分区内容容器 */
.overview-layout {
  padding-top: 0.15rem;
}

.overview-section {
  padding: 1.1rem 0 1.25rem;
}

.overview-section:first-child {
  padding-top: 0.35rem;
}

.overview-section + .overview-section {
  border-top: 1px solid var(--line-subtle);
}

.overview-section-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 0.9rem;
}

.overview-section .block-title {
  margin-bottom: 0.35rem;
  font-size: var(--fs-sm);
}

.overview-section .block-desc {
  max-width: 46rem;
  margin: 0;
  font-size: var(--fs-2xs);
  line-height: 1.7;
}

.overview-section-body {
  min-width: 0;
}

.overview-relations-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 1rem;
}

.relation-card {
  min-width: 0;
  border: 1px solid color-mix(in oklab, var(--text) 10%, transparent);
  border-radius: 10px;
  background:
    linear-gradient(
      180deg,
      color-mix(in oklab, var(--surface-0) 78%, transparent),
      color-mix(in oklab, var(--surface-2) 42%, transparent)
    ),
    var(--surface-0);
  box-shadow:
    inset 0 1px 0 color-mix(in oklab, var(--text) 4%, transparent),
    var(--shadow-xs);
  overflow: hidden;
}

.relation-card .relations-section,
.relation-card .no-relations {
  border-bottom: 0;
}

.block-title {
  font-size: var(--fs-xs);
  font-weight: 700;
  color: var(--accent);
  letter-spacing: 1px;
  margin-bottom: 0.5rem;
}

.block-desc {
  display: block;
  font-size: var(--fs-2xs);
  color: var(--text-soft);
  line-height: 1.5;
  margin-top: -0.3rem;
  margin-bottom: 0.5rem;
  letter-spacing: 0.3px;
}

/* 分区开头的数据驱动白话小结（"这是什么 / 这对你意味着什么"） */
.plain-summary {
  margin: 0 0 0.95rem;
  padding: 0.55rem 0.75rem;
  border-left: 2px solid var(--line-strong);
  background: color-mix(in oklab, var(--surface-2) 42%, transparent);
  border-radius: 0 6px 6px 0;
  font-size: var(--fs-xs);
  color: var(--text);
  line-height: 1.7;
  overflow-wrap: anywhere;
}

/* Ten gods */
.ten-gods-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 0.75rem;
}

.god-item {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  justify-content: space-between;
  min-height: 76px;
  padding: 0.72rem 0.78rem;
  background: color-mix(in oklab, var(--surface-2) 55%, transparent);
  border-radius: 8px;
  border: 1px solid var(--line-subtle);
  box-shadow: inset 0 1px 0 color-mix(in oklab, var(--text) 4%, transparent);
}

.god-pillar {
  font-size: var(--fs-2xs);
  color: var(--text-muted);
  margin-bottom: 0.35rem;
}

.god-name {
  font-size: var(--fs-sm);
  font-weight: 700;
  line-height: 1.2;
  color: var(--crimson);
}

/* 十神格的白话关键词小字 */
.god-plain {
  margin-top: 0.2rem;
  font-size: var(--fs-2xs);
  color: var(--text-soft);
  line-height: 1.4;
}

/* NaYin */
.nayin-list {
  display: flex;
  flex-wrap: wrap;
  gap: 0.65rem;
}

.nayin-tag {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  min-height: 38px;
  padding: 0.45rem 0.72rem;
  background: color-mix(in oklab, var(--accent-dim) 72%, transparent);
  border: 1px solid;
  border-radius: 8px;
  color: var(--text);
  line-height: 1.2;
}

.nayin-pillar {
  opacity: 0.6;
  font-size: var(--fs-2xs);
}

.nayin-name {
  font-weight: 600;
}

.nayin-ganzhi,
.nayin-element {
  font-size: var(--fs-2xs);
  color: var(--text-muted);
  letter-spacing: 0;
}

/* DaYun — 时间轴风格 */
.dayun-timeline {
  display: flex;
  flex-wrap: nowrap;
  gap: 0.25rem;
  overflow-x: auto;
  padding: 0.2rem 0.1rem 0.75rem;
  scrollbar-width: thin;
}

.dayun-node {
  display: flex;
  flex-direction: column;
  align-items: center;
  min-width: 76px;
  flex: 0 0 76px;
}

.dayun-age {
  font-size: var(--fs-2xs);
  color: var(--text-dim);
  margin-bottom: 0.3rem;
  white-space: nowrap;
}

.dayun-dot-line {
  display: flex;
  align-items: center;
  width: 100%;
  position: relative;
}

.dayun-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--accent);
  box-shadow: 0 0 8px var(--accent-glow);
  flex-shrink: 0;
  margin: 0 auto;
  position: relative;
  z-index: 1;
}

.dayun-line {
  position: absolute;
  top: 50%;
  left: 50%;
  width: 100%;
  height: 1px;
  background: linear-gradient(90deg, var(--line-focus), var(--line-subtle));
  transform: translateY(-50%);
}

.dayun-pillar-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0;
  min-width: 48px;
  margin-top: 0.5rem;
  padding: 0.48rem 0.58rem;
  background: color-mix(in oklab, var(--accent-dim) 76%, transparent);
  border: 1px solid var(--line-strong);
  border-radius: 8px;
}

.dayun-gan {
  font-family: var(--font-serif), serif;
  font-size: var(--fs-sm);
  font-weight: 700;
  color: var(--accent);
  line-height: 1.2;
}

.dayun-zhi {
  font-family: var(--font-serif), serif;
  font-size: var(--fs-sm);
  font-weight: 600;
  color: var(--text-muted);
  line-height: 1.2;
}

.dayun-detail-tab {
  gap: 0.95rem;
}

.dayun-overview-card,
.dayun-current-card,
.dayun-stage-block {
  border: 1px solid var(--line-subtle);
  border-radius: 10px;
  background: color-mix(in oklab, var(--surface-2) 42%, transparent);
  padding: 1rem;
}

.dayun-summary-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 0.75rem;
  margin-top: 0.9rem;
}

.dayun-summary-item {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  min-width: 0;
  padding: 0.8rem;
  border: 1px solid var(--line-subtle);
  border-radius: 8px;
  background: var(--surface-0);
}

.dayun-summary-label {
  color: var(--text-soft);
  font-size: var(--fs-2xs);
}

.dayun-summary-item strong {
  min-width: 0;
  color: var(--text);
  font-size: var(--fs-sm);
  line-height: 1.25;
  overflow-wrap: anywhere;
}

.dayun-method-notes {
  display: flex;
  flex-wrap: wrap;
  gap: 0.45rem;
  margin-top: 0.85rem;
}

.dayun-method-notes span,
.dayun-mini-chip,
.dayun-evidence-chip {
  display: inline-flex;
  align-items: center;
  min-height: 24px;
  padding: 0.14rem 0.5rem;
  border: 1px solid var(--line-subtle);
  border-radius: 4px;
  background: var(--surface-0);
  color: var(--text-muted);
  font-size: var(--fs-2xs);
  line-height: 1.25;
}

.dayun-current-head {
  display: flex;
  justify-content: space-between;
  gap: 0.9rem;
  align-items: flex-start;
  margin-bottom: 0.9rem;
}

.dayun-current-score {
  display: inline-flex;
  flex-shrink: 0;
  align-items: center;
  min-height: 28px;
  padding: 0.18rem 0.65rem;
  border-radius: 999px;
  border: 1px solid var(--line-subtle);
  font-size: var(--fs-2xs);
  font-weight: 700;
}

.dayun-current-score.score-good {
  color: #16a34a;
  background: rgba(22, 163, 74, 0.08);
  border-color: rgba(22, 163, 74, 0.18);
}

.dayun-current-score.score-watch {
  color: #dc2626;
  background: rgba(220, 38, 38, 0.08);
  border-color: rgba(220, 38, 38, 0.18);
}

.dayun-current-body {
  display: grid;
  grid-template-columns: 76px minmax(0, 1fr);
  gap: 0.9rem;
  align-items: stretch;
}

.dayun-current-pillar {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.05rem;
  border-radius: 9px;
  border: 1px solid var(--line-focus);
  background: var(--accent-dim);
  color: var(--accent);
  font-family: var(--font-serif), serif;
  font-size: var(--fs-3xl);
  font-weight: 800;
  line-height: 1.05;
}

.dayun-current-copy {
  min-width: 0;
}

.dayun-current-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem;
  margin-bottom: 0.55rem;
}

.dayun-current-copy p {
  margin: 0;
  color: var(--text-muted);
  font-size: var(--fs-xs);
  line-height: 1.65;
}

/* 当前大运的白话小结句：比依据句更醒目 */
.dayun-current-copy .dayun-current-plain {
  margin-bottom: 0.55rem;
  color: var(--text);
}

.dayun-evidence-list {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem;
  margin-top: 0.65rem;
}

.dayun-stage-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.8rem;
  margin-top: 0.9rem;
}

.dayun-stage-card {
  min-width: 0;
  padding: 0.85rem;
  border: 1px solid var(--line-subtle);
  border-left: 3px solid var(--line-focus);
  border-radius: 9px;
  background: var(--surface-0);
  transition:
    border-color 0.2s,
    background 0.2s,
    transform 0.2s;
}

.dayun-stage-card:hover {
  border-color: var(--line-strong);
  transform: translateY(-1px);
}

.dayun-stage-card.is-current {
  border-color: var(--line-focus);
  background: color-mix(in oklab, var(--accent-dim) 58%, var(--surface-0));
}

.dayun-stage-top {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.35rem;
  color: var(--text-soft);
  font-size: var(--fs-2xs);
  margin-bottom: 0.7rem;
}

.dayun-now-badge {
  color: var(--accent);
  background: var(--accent-dim);
  border: 1px solid var(--line-focus);
  border-radius: 999px;
  padding: 0.08rem 0.42rem;
  font-weight: 700;
}

.dayun-stage-main {
  display: grid;
  grid-template-columns: 48px minmax(0, 1fr);
  gap: 0.7rem;
  align-items: center;
  margin-bottom: 0.7rem;
}

.dayun-stage-pillar {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 60px;
  border-radius: 8px;
  border: 1px solid var(--line-subtle);
  background: color-mix(in oklab, var(--surface-2) 70%, transparent);
  font-family: var(--font-serif), serif;
  font-weight: 800;
  line-height: 1.05;
}

.dayun-stage-gan {
  font-size: var(--fs-2xl);
}

.dayun-stage-zhi {
  font-size: var(--fs-lg);
}

.dayun-stage-info {
  min-width: 0;
}

.dayun-stage-title {
  color: var(--text);
  font-size: var(--fs-sm);
  font-weight: 800;
  margin-bottom: 0.45rem;
}

.dayun-stage-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.32rem;
}

.dayun-stage-tags span {
  display: inline-flex;
  align-items: center;
  min-height: 22px;
  padding: 0.1rem 0.42rem;
  border: 1px solid var(--line-subtle);
  border-radius: 4px;
  color: var(--text-muted);
  background: color-mix(in oklab, var(--surface-2) 62%, transparent);
  font-size: var(--fs-2xs);
  line-height: 1.2;
}

.fortune-detail-tab {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.85rem;
}

.fortune-detail-tab > .analysis-block {
  min-width: 0;
  padding: 0.9rem;
  border: 1px solid var(--line-subtle);
  border-radius: 10px;
  background: color-mix(in oklab, var(--surface-2) 36%, transparent);
}

.fortune-detail-tab > .analysis-block:first-child {
  grid-column: 1 / -1;
}

.fortune-detail-tab .fortune-layer-list {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.65rem;
}

.fortune-detail-tab .fortune-layer-row {
  min-width: 0;
  padding: 0.7rem;
  border: 1px solid var(--line-subtle);
  border-left: 3px solid var(--line-focus);
  border-radius: 8px;
  background: var(--surface-0);
}

.classical-rules-block {
  margin-bottom: 0.95rem;
}

/* Element Detail Table */
.element-detail-table {
  border: 1px solid var(--line-strong);
  border-radius: 8px;
  overflow: hidden;
}

.ed-header {
  display: grid;
  grid-template-columns: 1fr 1fr 1.5fr 1fr;
  padding: 0.4rem 0.75rem;
  background: var(--surface-2);
  font-size: var(--fs-2xs);
  color: var(--text-muted);
  letter-spacing: 0.5px;
  border-bottom: 1px solid var(--line-subtle);
}

.ed-row {
  display: grid;
  grid-template-columns: 1fr 1fr 1.5fr 1fr;
  padding: 0.4rem 0.75rem;
  font-size: var(--fs-xs);
  border-bottom: 1px solid var(--line-subtle);
}

.ed-row:last-child {
  border-bottom: none;
}

.ed-elem {
  font-weight: 700;
}

.ed-tg,
.ed-zc,
.ed-total {
  color: var(--text);
}

.level-none {
  background: rgba(255, 255, 255, 0.01);
}
.level-weak .ed-total {
  color: var(--text-dim);
}
.level-medium .ed-total {
  color: var(--text-muted);
}
.level-strong .ed-total {
  color: #c76f12;
  font-weight: 700;
}
.level-very-strong .ed-total {
  color: #a6540a;
  font-weight: 800;
}

/* Body Strength */
.body-strength {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.bs-primary {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.45rem;
}

.bs-primary > span {
  color: var(--text-soft);
  font-size: var(--fs-2xs);
}

.bs-primary > strong {
  color: var(--accent);
  font-size: var(--fs-sm);
}

.bs-primary > code {
  padding: 0.12rem 0.45rem;
  border: 1px solid var(--line-subtle);
  border-radius: 3px;
  color: var(--text-muted);
  font-family: var(--font-mono);
  font-size: var(--fs-2xs);
}

.bs-contract-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  border-top: 1px solid var(--line-subtle);
  border-bottom: 1px solid var(--line-subtle);
}

.bs-contract-grid > div {
  min-width: 0;
  padding: 0.5rem;
  border-right: 1px solid var(--line-subtle);
}

.bs-contract-grid > div:last-child {
  border-right: 0;
}

.bs-contract-grid span,
.bs-contract-grid strong {
  display: block;
  overflow-wrap: anywhere;
}

.bs-contract-grid span {
  color: var(--text-soft);
  font-size: var(--fs-2xs);
}

.bs-contract-grid strong {
  margin-top: 0.2rem;
  color: var(--text);
  font-family: var(--font-mono);
  font-size: var(--fs-2xs);
  font-weight: 600;
}

.bs-band-rules,
.bs-limitations {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
}

.bs-band-rules span,
.bs-limitations span {
  padding: 0.14rem 0.4rem;
  border: 1px solid var(--line-subtle);
  border-radius: 3px;
  color: var(--text-soft);
  font-size: var(--fs-2xs);
}

.body-strength-bars {
  margin-top: 0.25rem;
  padding-top: 0.65rem;
  border-top: 1px solid var(--line-subtle);
}

.body-strength-notes {
  margin-top: 0.25rem;
}

.bs-adjustments {
  border-top: 1px solid var(--line-subtle);
}

.bs-adjustments > div {
  display: grid;
  grid-template-columns: minmax(5rem, auto) minmax(8rem, auto) 1fr;
  gap: 0.5rem;
  padding: 0.45rem 0;
  border-bottom: 1px solid var(--line-subtle);
  color: var(--text-muted);
  font-size: var(--fs-2xs);
}

.bs-adjustments p {
  margin: 0;
}

/* Pattern Analysis */
.pattern-detail {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.pattern-contract-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 0;
  border-top: 1px solid var(--line-subtle);
  border-bottom: 1px solid var(--line-subtle);
}

.pattern-contract-grid > div {
  min-width: 0;
  padding: 0.55rem 0.6rem;
  border-right: 1px solid var(--line-subtle);
}

.pattern-contract-grid > div:last-child {
  border-right: 0;
}

.pattern-contract-grid span,
.pattern-input-row span {
  display: block;
  color: var(--text-soft);
  font-size: var(--fs-2xs);
}

.pattern-contract-grid strong,
.pattern-input-row strong {
  display: block;
  min-width: 0;
  margin-top: 0.2rem;
  color: var(--text);
  font-family: var(--font-mono);
  font-size: var(--fs-2xs);
  font-weight: 600;
  overflow-wrap: anywhere;
}

.pattern-inputs {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.5rem 1rem;
  padding: 0.25rem 0;
}

.pattern-candidates {
  margin-top: 0.5rem;
  border-top: 1px solid var(--line-subtle);
}

.pattern-candidates-title {
  padding: 0.65rem 0 0.35rem;
  color: var(--text-soft);
  font-size: var(--fs-2xs);
  font-weight: 600;
}

.pattern-candidate-row {
  padding: 0.55rem 0;
  border-top: 1px solid color-mix(in oklab, var(--line-subtle) 65%, transparent);
}

.pattern-candidate-heading {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.4rem;
}

.pattern-candidate-heading strong {
  color: var(--text);
  font-size: var(--fs-xs);
}

.pattern-candidate-heading span {
  padding: 0.08rem 0.35rem;
  border: 1px solid var(--line-subtle);
  border-radius: 3px;
  color: var(--text-soft);
  font-size: var(--fs-2xs);
}

.pattern-candidate-row small {
  display: block;
  margin-top: 0.25rem;
  color: var(--text-soft);
  font-size: var(--fs-2xs);
  line-height: 1.5;
}

.pattern-candidate-row p {
  margin: 0.25rem 0 0;
  color: var(--crimson);
  font-size: var(--fs-2xs);
  line-height: 1.5;
}

.pattern-limitations {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
  padding-top: 0.4rem;
  border-top: 1px solid var(--line-subtle);
}

.pattern-limitations span {
  padding: 0.15rem 0.4rem;
  border: 1px solid var(--line-subtle);
  border-radius: 3px;
  color: var(--text-soft);
  font-size: var(--fs-2xs);
}

.ming-gong-detail {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.ming-gong-main {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.ming-gong-ganzhi {
  font-size: var(--fs-xl);
  font-weight: 700;
  color: var(--accent);
  letter-spacing: 0.15em;
}

.ming-gong-nayin {
  font-size: var(--fs-xs);
  color: var(--text-muted);
}

.ming-gong-shensha {
  display: flex;
  align-items: flex-start;
  gap: 0.5rem;
}

.shensha-badge {
  display: inline-block;
  padding: 0.12rem 0.45rem;
  border-radius: 3px;
  font-size: var(--fs-xs);
  font-weight: 600;
  white-space: nowrap;
  background: color-mix(in oklab, var(--text) 15%, transparent);
  color: var(--text);
}

.shensha-badge.shensha-ji {
  background: color-mix(in oklab, #b8860b 12%, transparent);
  color: var(--accent);
  border: 1px solid var(--line-focus);
}

.shensha-badge.shensha-xiong {
  background: color-mix(in oklab, var(--crimson) 10%, transparent);
  color: var(--crimson);
  border: 1px solid color-mix(in oklab, var(--crimson) 18%, transparent);
}

.shensha-desc {
  font-size: var(--fs-xs);
  color: var(--text-muted);
  line-height: 1.55;
}

.ming-gong-zhi {
  font-size: var(--fs-xs);
  color: var(--text-muted);
  line-height: 1.6;
  border-left: 2px solid var(--line-focus);
  padding-left: 0.6rem;
}

.tiaohou-source {
  font-size: var(--fs-2xs);
  color: var(--text-muted);
  font-weight: 400;
  margin-left: 0.4rem;
  font-style: italic;
}

.tiaohou-card {
  background: var(--accent-dim);
  border: 1px solid var(--line-focus);
  border-radius: 10px;
  padding: 1rem 1.1rem;
  margin-top: 0.4rem;
}

.tiaohou-header {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  margin-bottom: 0.7rem;
  flex-wrap: wrap;
}

.tiaohou-stem {
  font-size: var(--fs-2xl);
  font-weight: 700;
  color: var(--accent);
}

.tiaohou-arrow {
  color: var(--text-muted);
  font-size: var(--fs-sm);
}

.tiaohou-month {
  font-size: var(--fs-2xl);
  font-weight: 700;
  color: var(--text);
}

.tiaohou-divider {
  color: var(--text-soft);
  margin: 0 0.2rem;
}

.tiaohou-label {
  color: var(--text-soft);
  font-size: var(--fs-xs);
}

.tiaohou-primary {
  font-size: var(--fs-2xl);
  font-weight: 800;
}

.tiaohou-summary {
  font-size: var(--fs-sm);
  color: var(--text);
  line-height: 1.65;
  margin-bottom: 0.8rem;
}

.tiaohou-chart-matches {
  margin-bottom: 0.85rem;
  padding: 0.75rem;
  border: 1px solid color-mix(in oklab, var(--accent) 28%, var(--line-subtle));
  border-radius: 8px;
  background: color-mix(in oklab, var(--accent) 7%, transparent);
}

.tiaohou-chart-heading {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 0.5rem;
  flex-wrap: wrap;
  margin-bottom: 0.55rem;
}

.tiaohou-chart-heading strong {
  color: var(--text);
  font-size: var(--fs-sm);
}

.tiaohou-chart-heading span {
  color: var(--text-muted);
  font-size: var(--fs-2xs);
}

.tiaohou-chart-match {
  display: grid;
  grid-template-columns: auto 1fr;
  gap: 0.65rem;
  align-items: start;
}

.tiaohou-chart-candidates {
  display: flex;
  gap: 0.25rem;
}

.tiaohou-chart-candidates span {
  min-width: 1.8rem;
  padding: 0.2rem 0.35rem;
  border-radius: 6px;
  color: var(--accent);
  background: var(--surface-raised);
  border: 1px solid var(--line-focus);
  font-size: var(--fs-lg);
  font-weight: 800;
  text-align: center;
}

.tiaohou-chart-match strong {
  display: block;
  color: var(--text);
  font-size: var(--fs-xs);
  line-height: 1.5;
}

.tiaohou-chart-match p,
.tiaohou-chart-match small {
  display: block;
  margin: 0.2rem 0 0;
  color: var(--text-muted);
  font-size: var(--fs-2xs);
  line-height: 1.5;
}

.tiaohou-rules {
  display: flex;
  flex-direction: column;
  gap: 0.45rem;
}

.tiaohou-rule {
  display: flex;
  align-items: baseline;
  gap: 0.5rem;
  font-size: var(--fs-xs);
}

.tiaohou-xi {
  color: var(--accent);
  font-weight: 600;
  white-space: nowrap;
}

.tiaohou-ji {
  color: var(--text-muted);
  font-weight: 600;
  white-space: nowrap;
}

.tiaohou-reason {
  color: var(--text-muted);
  line-height: 1.5;
}

.month-command-candidates {
  margin-top: 0.8rem;
  border-top: 1px solid var(--line-subtle);
}

.month-command-heading,
.month-command-current {
  display: flex;
  align-items: baseline;
  gap: 0.45rem;
  flex-wrap: wrap;
}

.month-command-heading {
  padding: 0.7rem 0 0.45rem;
}

.month-command-heading strong {
  color: var(--text);
  font-size: var(--fs-xs);
}

.month-command-heading span {
  color: var(--text-muted);
  font-size: var(--fs-2xs);
}

.month-command-row {
  padding: 0.55rem 0;
  border-top: 1px dashed var(--line-subtle);
}

.month-command-current strong {
  color: var(--accent);
  font-size: var(--fs-lg);
}

.month-command-current span {
  color: var(--text);
  font-size: var(--fs-xs);
  font-weight: 700;
}

.month-command-current small {
  color: var(--text-muted);
  font-size: var(--fs-2xs);
}

.month-command-row p {
  margin: 0.2rem 0 0;
  color: var(--text-muted);
  font-size: var(--fs-2xs);
  line-height: 1.55;
}

.tiaohou-limitations {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  margin-top: 0.75rem;
  padding-top: 0.65rem;
  border-top: 1px solid var(--line-subtle);
  color: var(--text-muted);
  font-size: var(--fs-2xs);
  line-height: 1.5;
}

.season-text {
  font-size: var(--fs-sm);
  color: var(--text-muted);
  line-height: 1.7;
  white-space: pre-wrap;
  margin: 0;
}

.month-season-facts {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(110px, 1fr));
  gap: 0;
  margin: 0;
  border-top: 1px solid var(--line-subtle);
  border-bottom: 1px solid var(--line-subtle);
}

.month-season-facts > div {
  min-width: 0;
  padding: 0.65rem 0.75rem;
}

.month-season-facts dt {
  margin-bottom: 0.25rem;
  color: var(--text-soft);
  font-size: var(--fs-2xs);
  letter-spacing: 0;
}

.month-season-facts dd {
  margin: 0;
  color: var(--text);
  font-size: var(--fs-sm);
  font-weight: 700;
  letter-spacing: 0;
}

.sheng-xiao-tag {
  font-size: var(--fs-2xs);
  color: var(--text-muted);
  letter-spacing: 1px;
}

.empties-tag {
  font-size: var(--fs-2xs);
  color: color-mix(in oklab, var(--crimson) 78%, var(--text));
}

/* ShenSha list */
.shen-sha-list {
  display: flex;
  flex-direction: column;
  gap: 0.42rem;
}

.shen-sha-row {
  display: grid;
  grid-template-columns: 4.6rem auto 1fr;
  align-items: center;
  gap: 0.5rem;
  min-height: 2.1rem;
  padding: 0.35rem 0.65rem;
  background: rgba(255, 255, 255, 0.028);
  border: 1px solid var(--line-strong);
  border-radius: 7px;
  box-shadow: inset 0 1px 0 color-mix(in oklab, var(--text) 4%, transparent);
}

.shen-sha-name {
  font-family: var(--font-serif), serif;
  font-size: var(--fs-sm);
  font-weight: 700;
  color: var(--accent);
  letter-spacing: 0.06em;
  white-space: nowrap;
}

.shen-sha-target {
  min-width: 1.35rem;
  height: 1.35rem;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  background: var(--line-strong);
  border: 1px solid var(--line-focus);
  color: var(--text);
  font-size: var(--fs-2xs);
  font-weight: 700;
}

.shen-sha-desc {
  color: var(--text-muted);
  font-size: var(--fs-xs);
  line-height: 1.45;
}

/* ShenSha group title */
.shen-sha-group-title {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  font-family: var(--font-serif), serif;
  font-size: var(--fs-sm);
  font-weight: 700;
  color: var(--text);
  margin-bottom: 0.55rem;
  letter-spacing: 0.06em;
}

.shen-sha-group-dot {
  width: 0.42rem;
  height: 0.42rem;
  border-radius: 50%;
  flex-shrink: 0;
}

.shen-sha-group-role {
  font-family: var(--font-sans), sans-serif;
  font-size: var(--fs-2xs);
  font-weight: 400;
  color: var(--text-soft);
  letter-spacing: 0.03em;
}

/* ShenSha summary */
.shen-sha-summary-block {
  border-top: 1px solid var(--line-subtle);
  margin-top: 0.5rem;
  padding-top: 0.75rem;
}

.shen-sha-summary-title {
  font-size: var(--fs-xs);
  font-weight: 600;
  color: var(--text-muted);
  margin-bottom: 0.45rem;
}

.shen-sha-summary-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
}

.shen-sha-summary-list li {
  font-size: var(--fs-xs);
  color: var(--text-muted);
  line-height: 1.5;
  padding-left: 0.65rem;
  position: relative;
}

.shen-sha-summary-list li::before {
  content: '–';
  position: absolute;
  left: 0;
  color: var(--text-soft);
}

.tg-summary-line {
  margin: 0.55rem 0 0;
  font-size: var(--fs-xs);
  color: var(--text-muted);
  line-height: 1.6;
}
.tg-stack {
  display: flex;
  height: 12px;
  margin-top: 0.6rem;
  border-radius: 6px;
  overflow: hidden;
  background: var(--surface-2);
  border: 1px solid var(--line-subtle);
}
.tg-stack-seg {
  display: block;
  height: 100%;
}
.tg-chip-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.3rem 0.9rem;
  margin-top: 0.65rem;
}
.tg-chip {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  min-height: 32px;
  padding: 0.18rem 0.55rem;
  border: 1px solid var(--line-subtle);
  border-radius: 6px;
  background: color-mix(in oklab, var(--surface-0) 58%, transparent);
}
.tg-chip.is-zero {
  opacity: 0.45;
}
.tg-chip-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}
.tg-chip-name {
  font-size: var(--fs-xs);
  font-weight: 600;
  color: var(--text);
  letter-spacing: 0.5px;
}
.tg-chip-count {
  margin-left: auto;
  font-size: var(--fs-2xs);
  color: var(--text-soft);
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
}
.tg-chip-pct {
  min-width: 3.1em;
  text-align: right;
  font-size: var(--fs-xs);
  color: var(--text-muted);
  font-variant-numeric: tabular-nums;
}
@media (max-width: 480px) {
  .tg-chip-grid {
    grid-template-columns: 1fr;
  }
}
.tg-limitations {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  margin-top: 0.75rem;
  font-size: var(--fs-2xs);
  line-height: 1.6;
  color: var(--text-muted);
}

/* 十神白话解读：与统计块共用色点色板，紧凑列表不套卡片 */
.tg-insight {
  margin-top: 0.8rem;
  padding-top: 0.7rem;
  border-top: 1px solid var(--line-subtle);
}
.tg-insight-title {
  margin-bottom: 0.45rem;
  font-family: var(--font-serif);
  font-size: var(--fs-xs);
  font-weight: 700;
  color: var(--text);
  letter-spacing: 0.08em;
}
.tg-insight-item {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  column-gap: 0.6rem;
  padding: 0.22rem 0;
}
.tg-insight-head {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  flex-shrink: 0;
}
.tg-insight-name {
  font-size: var(--fs-xs);
  font-weight: 700;
  color: var(--text);
  letter-spacing: 0.5px;
}
.tg-insight-keyword {
  font-size: var(--fs-2xs);
  color: var(--text-soft);
}
.tg-insight-text {
  flex: 1 1 24rem;
  margin: 0;
  font-size: var(--fs-2xs);
  line-height: 1.65;
  color: var(--text-muted);
}
.tg-insight-caution {
  color: var(--text-soft);
}
.tg-insight-absent {
  margin: 0.5rem 0 0;
  font-size: var(--fs-2xs);
  line-height: 1.65;
  color: var(--text-muted);
}
.tg-insight-note {
  margin: 0.45rem 0 0;
  font-size: var(--fs-2xs);
  line-height: 1.6;
  color: var(--text-soft);
}

/* MissingElements 原始五行分布 */
.missing-elem-card {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
  padding: 0.75rem;
  background: color-mix(in oklab, var(--text) 2%, transparent);
  border-radius: 8px;
  border: 1px solid var(--line-subtle);
}
.me-row {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  flex-wrap: wrap;
}
.me-label {
  font-size: var(--fs-2xs);
  color: var(--text-muted);
  min-width: 60px;
}
.me-tag {
  font-size: var(--fs-xs);
  font-weight: 700;
  padding: 0.1rem 0.4rem;
  border: 1px solid;
  border-radius: 4px;
}
.me-missing {
  background: rgba(220, 38, 38, 0.06);
}
.me-weak {
  background: rgba(245, 158, 11, 0.06);
}
.me-note {
  margin: 0.15rem 0 0;
  color: var(--text-muted);
  font-size: var(--fs-2xs);
  line-height: 1.55;
}

/* ===== Dark mode overrides — keep relation colors readable without neon glare ===== */
:global(.dark) .overview-relations {
  --rel-he: #6fcf97;
  --rel-ke: #e58a8a;
  --rel-sheng: #8fb9ee;
  --rel-hai: #d8a56c;
  --rel-xing: #b9a0dd;
}

:global(.dark) .me-missing {
  background: rgba(251, 113, 133, 0.06);
}
:global(.dark) .me-weak {
  background: rgba(251, 191, 36, 0.06);
}

:global(.dark) .god-name {
  color: #fb7185;
}

:global(.dark) .level-strong .ed-total {
  color: #f59e0b;
}
:global(.dark) .level-very-strong .ed-total {
  color: #f97316;
}

:global(.dark) .bento-day-card {
  background:
    linear-gradient(180deg, rgba(var(--jade-accent-rgb), 0.12), transparent 150%), var(--glass-bg);
}

:global(.dark) .shensha-badge.shensha-xiong {
  color: #f87171;
}

:global(.dark) .shen-sha-detail-row.sha-good,
:global(.dark) .fortune-layer-chip.sha-good {
  border-color: rgba(74, 222, 128, 0.35);
  color: #4ade80;
}

:global(.dark) .shen-sha-detail-row.sha-risk,
:global(.dark) .fortune-layer-chip.sha-risk {
  border-color: rgba(248, 113, 113, 0.35);
  color: #f87171;
}

@media (max-width: 1023px) {
  /* 移动端：目录降级为顶部 sticky 横向 chip 条 */
  .chart-layout {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .chart-toc {
    top: 80px;
    z-index: 5;
    margin: 0 -0.25rem;
    padding: 0.35rem 0.25rem 0;
    background: var(--bg);
  }

  .toc-inner {
    flex-direction: row;
    gap: 0.25rem;
    border-left: 0;
    border-bottom: 1px solid var(--line-subtle);
    overflow-x: auto;
    scrollbar-width: none;
  }

  .toc-inner::-webkit-scrollbar {
    display: none;
  }

  .toc-item {
    flex-direction: row;
    padding: 0.5rem 0.65rem 0.55rem;
    margin-left: 0;
    border-left: 0;
    border-bottom: 2px solid transparent;
    margin-bottom: -1px;
    white-space: nowrap;
  }

  .toc-item.active {
    border-bottom-color: var(--accent);
  }

  .toc-sub {
    display: none;
  }
}

@media (max-width: 980px) {
  .pillars-bento {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .bento-radar {
    grid-column: 1 / -1;
    min-height: 184px;
  }

  .bento-radar-chart {
    min-height: 148px;
    height: 148px;
  }
}

@media (max-width: 860px) {
  .pillars-bento {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .bento-radar {
    grid-column: 1 / -1;
    order: 1;
  }

  .pillars-bento > .bento-small {
    order: 0 !important;
  }

  .ten-gods-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .overview-relations-grid,
  .dayun-summary-grid,
  .dayun-stage-grid,
  .fortune-detail-tab,
  .fortune-detail-tab .fortune-layer-list {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .shensha-grid {
    grid-template-columns: 1fr;
  }

  .rule-table-list {
    grid-template-columns: 1fr;
  }

  .pattern-contract-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .bs-contract-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .pattern-contract-grid > div:nth-child(2) {
    border-right: 0;
  }

  .bs-contract-grid > div:nth-child(2) {
    border-right: 0;
  }
}

@media (max-width: 560px) {
  .chart-identity {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .chart-identity > div:nth-child(3n) {
    border-right: 1px solid var(--line-subtle);
  }

  .chart-identity > div:nth-child(2n) {
    border-right: 0;
  }

  .pillars-bento {
    grid-template-columns: 1fr;
  }

  .pillars-bento > .bento-card {
    min-height: auto;
  }

  .bento-radar {
    grid-column: auto;
    min-height: 178px;
  }

  .bento-radar-chart {
    min-height: 142px;
    height: 142px;
  }

  .overview-section {
    padding: 0.95rem 0 1.05rem;
  }

  .overview-section-head {
    margin-bottom: 0.7rem;
  }

  .shensha-overview-head {
    flex-direction: column;
  }

  .shensha-overview-count {
    align-self: flex-start;
  }

  .shensha-grid {
    grid-template-columns: 1fr;
  }

  .ten-gods-grid {
    gap: 0.55rem;
  }

  .god-item {
    min-height: 68px;
    padding: 0.65rem;
  }

  .nayin-list {
    display: grid;
    grid-template-columns: 1fr;
    gap: 0.5rem;
  }

  .nayin-tag {
    width: 100%;
    justify-content: space-between;
  }

  .pattern-contract-grid,
  .pattern-inputs,
  .bs-contract-grid {
    grid-template-columns: 1fr;
  }

  .pattern-contract-grid > div {
    border-right: 0;
    border-bottom: 1px solid var(--line-subtle);
  }

  .pattern-contract-grid > div:last-child {
    border-bottom: 0;
  }

  .bs-contract-grid > div {
    border-right: 0;
    border-bottom: 1px solid var(--line-subtle);
  }

  .bs-contract-grid > div:last-child {
    border-bottom: 0;
  }

  .bs-adjustments > div {
    grid-template-columns: 1fr;
  }

  .dayun-node {
    min-width: 68px;
    flex-basis: 68px;
  }

  .dayun-pillar-card {
    min-width: 44px;
    padding: 0.42rem 0.52rem;
  }

  .dayun-overview-card,
  .dayun-current-card,
  .dayun-stage-block {
    padding: 0.8rem;
  }

  .overview-relations-grid,
  .dayun-summary-grid,
  .dayun-stage-grid,
  .fortune-detail-tab,
  .fortune-detail-tab .fortune-layer-list {
    grid-template-columns: 1fr;
    gap: 0.65rem;
  }

  .relations-section {
    padding: 0.82rem;
  }

  .gz-item {
    grid-template-columns: minmax(0, 1fr) max-content;
    align-items: start;
    gap: 0.42rem 0.55rem;
    padding: 0.58rem 0.62rem;
  }

  .gz-chars {
    min-width: 0;
  }

  .gz-text {
    grid-column: 1 / -1;
  }

  .dayun-current-head,
  .dayun-current-body {
    display: flex;
    flex-direction: column;
  }

  .dayun-current-score {
    align-self: flex-start;
  }

  .dayun-current-pillar {
    min-height: 72px;
    flex-direction: row;
    gap: 0.25rem;
  }

  .fortune-layer-top,
  .rule-meta-row,
  .rule-table-head {
    align-items: flex-start;
    flex-direction: column;
    gap: 0.35rem;
  }

  .rule-meta-value {
    text-align: left;
  }

  .evidence-bar-row {
    grid-template-columns: 2.8rem minmax(72px, 1fr) 2.2rem;
    gap: 0.38rem;
  }
}
</style>
