import client from './client'
import type { FortuneLayerSet, RuleMeta, ShenShaActivation } from './chart'

export type InterpretationLevel = 'basic' | 'advanced' | 'professional'

export interface ScoreEvidence {
  code: string
  stage: string
  category: string
  label: string
  impact: number
  description: string
  source: string
  evidence_basis: 'empirical'
  validation_status: 'not_validated'
  interpretation_status: 'not_adjudicated'
  is_outcome_conclusion: false
}

export interface FortuneScoreBreakdown {
  pipeline_version: string
  score_kind: 'structural_relation_index'
  evidence_basis: 'empirical'
  validation_status: 'not_validated'
  interpretation_status: 'not_adjudicated'
  is_outcome_probability: false
  base_score: number
  relation_score: number
  final_score: number
  evidence_completeness: number
  supporting_evidence: ScoreEvidence[]
  counter_evidence: ScoreEvidence[]
}

export interface TwelveStageEvidence {
  rule_id: string
  reference_stem: string
  query_branch: string
  name: string
  basis: string
  status: 'observed' | 'unavailable'
  interpretation_status: 'not_adjudicated'
}

export interface TraditionalCalendarEvidence {
  rule_id: string
  month_branch: string
  query_branch: string
  name: string
  basis: string
  status: 'observed' | 'unavailable'
  interpretation_status: 'not_adjudicated'
}

export interface TenGodEvidence {
  rule_id: string
  reference_stem: string
  query_stem: string
  name: string
  basis: string
  status: 'observed' | 'unavailable'
  interpretation_status: 'not_adjudicated'
}

export interface SeasonElementEvidence {
  rule_id: string
  reference_stem: string
  reference_element: string
  query_month_branch: string
  season: string
  basis: string
  status: 'observed' | 'unavailable'
  interpretation_status: 'not_adjudicated'
}

export interface SeasonalStateEvidence {
  rule_id: string
  query_stem: string
  query_element: string
  query_month_branch: string
  season: string
  state: string
  basis: string
  status: 'observed' | 'unavailable'
  interpretation_status: 'not_adjudicated'
}

export interface ShengKeAnalysis {
  day_stem_relation: string
  day_branch_relation: string
  summary: string
}

export interface HiddenStemGod {
  rule_id: string
  query_branch: string
  reference_stem: string
  stem: string
  type: string
  element: string
  ten_god: string
  basis: string
  status: 'observed'
  interpretation_status: 'not_adjudicated'
}

export interface StemRelation {
  rule_id: string
  query_stem: string
  target_pillar: string
  target_stem: string
  type: string
  name: string
  combined_element?: string
  basis: string
  status: 'observed'
  transformation_status: string
  interpretation_status: 'not_adjudicated'
}

export interface BranchRelation {
  rule_id: string
  query_branch: string
  target_pillar: string
  target_branch: string
  type: string
  name: string
  basis: string
  status: 'observed'
  transformation_status: string
  interpretation_status: 'not_adjudicated'
}

/** Backend FortuneResponse — server JSON keys (snake_case). */
export interface FortuneDay {
  engine_version: string
  rule_version?: string
  school?: string
  rule_meta?: RuleMeta
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
  score_breakdown: FortuneScoreBreakdown
  evidence_completeness: number
  supporting_evidence: ScoreEvidence[]
  counter_evidence: ScoreEvidence[]
  clash_zodiac?: string
  today_elements?: Record<string, number>
  sheng_ke_analysis?: ShengKeAnalysis
  season_element?: SeasonElementEvidence
  ten_god?: TenGodEvidence
  twelve_stage?: TwelveStageEvidence
  jian_chu?: TraditionalCalendarEvidence
  huang_dao?: TraditionalCalendarEvidence
  hidden_stems?: HiddenStemGod[]
  stem_relations?: StemRelation[]
  branch_relations?: BranchRelation[]
  activated_shen_sha?: ShenShaActivation[]
  seasonal_state?: SeasonalStateEvidence
  fortune_layers?: FortuneLayerSet
  element_images?: Array<{ element: string; image_url: string; description: string }>
}

export interface FortuneSummary {
  highest_index_day: string
  lowest_index_day: string
  highest_index: number
  lowest_index: number
  element_distribution: Record<string, number> // keys 木 火 土 金 水
  dominant_element: string
  dominant_ten_god: string
  average_index: number
  index_standard_deviation: number
}

export interface WeeklyFortuneResponse {
  daily_fortunes: FortuneDay[]
  structural_relation_index: number
  element_trend: string
  summary: FortuneSummary
}

export interface MonthlyFortuneResponse {
  daily_fortunes: FortuneDay[]
  structural_relation_index: number
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
  try {
    return JSON.parse(json) as ElementTrendPoint[]
  } catch {
    return []
  }
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
