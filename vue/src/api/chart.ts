import client from './client'

export interface RuleMeta {
  rule_version: string
  school: string
  tables: RuleTableMeta[]
  body_strength?: BodyStrengthRuleConfig
}

export interface RuleTableMeta {
  key: string
  name: string
  version: string
  school: string
  source: string
  description: string
  count?: number
}

export interface BodyStrengthRuleConfig {
  weights: Record<string, number>
  normalizers: Record<string, number>
  adjustment_thresholds: Record<string, number>
}

export interface BodyStrengthResult {
  rule_version: string
  school: string
  verdict: string
  like: string[]
  dislike: string[]
  total_score: number
  ling_score: number
  di_score: number
  shi_score: number
  sheng_score: number
  lu_bonus: number
  components: BodyStrengthComponent[]
  evidence: BodyStrengthEvidence[]
  adjustments: BodyStrengthAdjustment[]
  summary: string
}

export interface BodyStrengthComponent {
  key: string
  name: string
  raw_score: number
  normalized_score: number
  weight: number
  weighted_score: number
  description: string
}

export interface BodyStrengthEvidence {
  component: string
  polarity: string
  source: string
  item: string
  score: number
  reason: string
}

export interface BodyStrengthAdjustment {
  name: string
  before: number
  after: number
  reason: string
  description: string
}

export interface ShenShaMeta {
  name: string
  category: string
  polarity: string
  priority: number
  source: string
  description: string
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
  shen_sha_details?: ShenShaActivation[]
  element_change: Record<string, number>
  description: string
  evidence: string[]
}

export interface ShenShaActivation {
  name: string
  type: string
  category?: string
  polarity?: string
  priority?: number
  description: string
  activation: string
}

export interface ChartSummary {
  id: number
  name: string
  gender: string
  birth_year: number
  birth_month: number
  birth_day: number
  birth_hour: number
  birth_min: number
  calendar_type: string
  created_at?: string
  updated_at?: string
}

export interface ChartDetail extends ChartSummary {
  rule_version?: string
  school?: string
  rule_meta?: RuleMeta
  year_pillar?: unknown
  month_pillar?: unknown
  day_pillar?: unknown
  hour_pillar?: unknown
  five_elements?: Record<string, number>
  element_detail?: unknown
  body_strength?: BodyStrengthResult
  ten_gods?: unknown
  na_yin?: unknown
  hidden_stems?: unknown
  da_yun_start?: unknown
  da_yun?: unknown
  clash_harmony?: unknown
  gan_zhi_analysis?: unknown
  pattern_analysis?: unknown
  ming_gong?: unknown
  ri_zhu_desc?: string
  pillar_details?: unknown
  tiao_hou?: string
  tiaohou?: unknown
  global_shen_sha?: string[]
  global_shen_sha_details?: ShenShaMeta[]
  jin_bu_huan?: string
  day_shen_sha?: string[]
  day_shen_sha_details?: ShenShaMeta[]
  season_text?: string
  season_text_month?: string
  ri_zhu_poem?: string
  ri_zhu_source?: string
  ri_zhu_comment?: string
  ri_zhu_hour_detail?: string
  shen_sha_by_pillar?: unknown
  shen_sha_summary?: unknown
  ten_god_proportion?: unknown
  ten_god_analysis?: unknown
  wuxing_season_note?: string
  wuxing_flow?: unknown
  tong_guan?: unknown
  missing_elements?: unknown
  flow_pattern_desc?: string
  dayun_flow?: unknown
  fortune_layers?: FortuneLayerSet
  ziwei_result?: unknown
  ziwei_computed?: boolean
}

export type BirthChart = ChartDetail

export interface ChartCreateRequest {
  birth_year: number
  birth_month: number
  birth_day: number
  birth_hour: number
  birth_min?: number
  calendar_type: 'SOLAR' | 'LUNAR' | 'BAZI'
  gender: string
  name?: string
}

export interface ChartListResponse {
  charts: ChartSummary[]
  total: number
  page: number
  page_size: number
}

export async function createChart(payload: ChartCreateRequest) {
  const { data } = await client.post('/chart', payload)
  return data
}

export async function fetchCharts(page = 1, pageSize = 10) {
  const { data } = await client.get<ChartListResponse>('/charts', {
    params: { page, page_size: pageSize },
  })
  return data
}

export async function fetchChart(id: number | string) {
  const { data } = await client.get<ChartDetail>(`/charts/${id}`)
  return data
}

export async function fetchFortuneHistory(chartId: number, page = 1, pageSize = 10) {
  const { data } = await client.get('/fortune/history', {
    params: { chart_id: chartId, page, page_size: pageSize },
  })
  return data
}
