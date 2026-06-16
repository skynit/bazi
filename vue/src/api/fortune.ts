import client from './client'

export interface FortuneGuideItem {
  label: string
  value: string
  element?: string
  reason: string
}

export interface FortuneGuide {
  precision_level: string
  confidence: number
  primary_element: string
  secondary_element: string
  avoid_element: string
  lucky_colors: FortuneGuideItem[]
  lucky_numbers: FortuneGuideItem[]
  face_direction: FortuneGuideItem
  wealth_direction: FortuneGuideItem
  avoid_direction: FortuneGuideItem
  recommended_actions: FortuneGuideItem[]
  cautions: FortuneGuideItem[]
  best_hours: FortuneGuideItem[]
  analysis: string
  strategy: string
}

/** Backend FortuneResponse — server JSON keys (snake_case). */
export interface FortuneDay {
  rule_version?: string
  school?: string
  rule_meta?: unknown
  solar_date: string
  lunar_date?: string
  day_gan_zhi: string
  week_day?: string
  sheng_xiao?: string
  yi_ji?: string
  ji_shen?: string
  xiong_shen?: string
  chong_sha?: string
  tai_shen?: string
  wu_xing?: string
  peng_zu?: string
  gua?: string
  jie_qi?: string
  score: number
  lucky_color: string
  lucky_number: number
  wealth_direction: string
  guide?: FortuneGuide
  clash_zodiac?: string
  auspicious_hours?: string[]
  yi?: string[]
  ji?: string[]
  today_elements?: Record<string, number>
  tiao_hou?: string
  season_element_advice?: string
  flow_impact?: string
  today_ten_god?: string
  ten_god_favorable?: boolean
  ten_god_desc?: string
  twelve_stage?: string
  stage_favorable?: boolean
  pattern_name?: string
  pattern_type?: string
  pattern_favorable?: string[]
  pattern_unfavorable?: string[]
  overall_verdict?: string
  favor_score?: number
  fortune_layers?: FortuneLayerSet
  element_images?: Array<{ element: string; image_url: string; description: string }>
}

export interface FortuneLayerSet {
  rule_version: string
  school: string
  dayun: FortuneLayer
  liunian: FortuneLayer
  liuyue: FortuneLayer
  xiaoyun: FortuneLayer
}

export interface FortuneLayer {
  key: string
  name: string
  pillar: string
  gan: string
  zhi: string
  start_age?: number
  end_age?: number
  age?: number
  year?: number
  month?: number
  ten_god: string
  favorable: boolean
  score: number
  relations: Array<{ target: string; type: string; detail: string; score: number }>
  activated_shen_sha: string[]
  element_change: Record<string, number>
  description: string
  evidence: string[]
}

export interface FortuneSummary {
  best_day: string
  worst_day: string
  best_score: number
  worst_score: number
  peak_days: string[]
  low_days: string[]
  element_distribution: Record<string, number> // keys 木 火 土 金 水
  dominant_element: string
  dominant_ten_god: string
  good_streak: number
  bad_streak: number
  average_score: number
  volatility: number
  key_advice: string
}

export interface WeeklyFortuneResponse {
  daily_fortunes: FortuneDay[]
  weekly_score: number
  element_trend: string
  summary: FortuneSummary
}

export interface MonthlyFortuneResponse {
  daily_fortunes: FortuneDay[]
  monthly_score: number
  element_trend: string
  summary: FortuneSummary
}

export interface ElementTrendPoint {
  date: string
  score: number
  metal: number
  wood: number
  water: number
  fire: number
  earth: number
}

export function parseTrend(json: string): ElementTrendPoint[] {
  if (!json) return []
  try { return JSON.parse(json) as ElementTrendPoint[] } catch { return [] }
}

export async function fetchDaily(chartId: number, queryDate?: string) {
  const { data } = await client.post<FortuneDay>('/fortune', {
    chart_id: chartId,
    query_date: queryDate,
  })
  return data
}

export async function fetchWeekly(chartId: number, startDate: string) {
  const { data } = await client.post<WeeklyFortuneResponse>('/fortune/weekly', {
    chart_id: chartId,
    start_date: startDate,
  })
  return data
}

export async function fetchMonthly(chartId: number, year: number, month: number) {
  const { data } = await client.post<MonthlyFortuneResponse>('/fortune/monthly', {
    chart_id: chartId,
    year,
    month,
  })
  return data
}
