import client from './client'

export type ZiWeiPeriodType =
  | 'dayun'
  | 'liunian'
  | 'liuyue'
  | 'liuri'
  | 'sihua_feixing'
  | 'sihua_chain'
  | 'self_mutagen'
  | 'palace_reading'
  | 'heming'
  | 'liunian_interpretation'
  | 'liuyue_interpretation'
  | 'liuri_interpretation'
  | 'period_summary'
  | 'liu_nian_stars'
  | 'query_view'

export interface ZiWeiQueryView {
  rule_version: string
  school: string
  palaces: ZiWeiPalaceQuery[]
  star_index: Record<string, string[]>
  patterns: string[]
}

export interface ZiWeiPalaceQuery {
  name: string
  branch: string
  index: number
  is_body_palace: boolean
  main_stars: string[]
  aux_stars: string[]
  adjective_stars: string[]
  all_stars: string[]
  has_star: Record<string, boolean>
  four_hua: string[]
  sanfang_sizheng: {
    opposite: string
    trine1: string
    trine2: string
    opposite_stars: string[]
    trine1_stars: string[]
    trine2_stars: string[]
    all_stars: string[]
  }
  surrounded_palaces: Array<{
    name: string
    branch: string
    role: string
    stars: string[]
  }>
}

export interface ZiWeiChartRequest {
  chart_id?: number
  birth_year?: number
  birth_month?: number
  birth_day?: number
  birth_hour?: number
  birth_min?: number
  calendar_type?: string
  gender?: string
  name?: string
  profile?: string
}

export interface ZiWeiPeriodRequest {
  chart_id: number
  period_type: ZiWeiPeriodType
  year?: number
  month?: number
  day?: number
  palace_idx?: number
  chart_id2?: number
}

export interface ZiWeiDerivationInput {
  calendar_type: 'SOLAR' | 'LUNAR_YEAR'
  year: number
  month: number
  day: number
  basis:
    | 'target_lunar_year_label'
    | 'target_solar_date_resolved_to_lunar_month'
    | 'target_solar_date_resolved_to_lunar_day'
  boundary_policy: 'iztro_normal_lunar_boundaries_fix_leap_day_15'
  resolved_lunar_date: {
    year: number
    month: number
    day: number
    is_leap_month: boolean
  }
  period_gan_zhi: string
}

export interface ZiWeiDerivedChartContract {
  derivation_type: 'liunian' | 'liuyue' | 'liuri'
  derivation_input: ZiWeiDerivationInput
  derivation_fingerprint: string
  base_content_hash: string
  derived_content_hash: string
}

export interface ZiWeiTransitLayers {
  liu_nian_stars: string[][]
  liu_yue_stars: string[][]
  liu_ri_stars: string[][]
  liu_nian_four_hua: string[][]
  liu_yue_four_hua: string[][]
  liu_ri_four_hua: string[][]
  liu_nian_palaces: string[]
  liu_yue_palaces: string[]
  liu_ri_palaces: string[]
}

export interface ZiWeiOverlayRequest {
  chart_id: number
  year?: number
}

export interface ZiWeiRuleSource {
  rule_id: string
  repository: string
  commit: string
  path: string
  sha256: string
  license: string
  source_tier: string
  validation_status: string
}

export interface ZiWeiPluginRequirement {
  id: string
  version: string
}

export interface ZiWeiStar {
  name: string
  type: string
  scope: string
  brightness: string
}

export interface ZiWeiPalace {
  name: string
  branch: string
  heavenly_stem: string
  is_body_palace: boolean
  stars: ZiWeiStar[]
  four_hua: string[]
  adjective_stars?: string[]
  changsheng_12?: string
  boshi_12?: string
  jiang_qian_12?: string
  sui_qian_12?: string
  sanfang_sizheng?: {
    opposite: string
    trine1: string
    trine2: string
  }
}

export interface ZiWeiChartResponse extends ZiWeiTransitLayers {
  profile_id: string
  engine_version: string
  rule_version: string
  rule_school: string
  rule_sources: ZiWeiRuleSource[]
  runtime_rule_tables_schema: string
  runtime_rule_tables_hash: string
  plugin_manifest: ZiWeiPluginRequirement[]
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
  content_hash?: string
  derivation_type?: 'liunian' | 'liuyue' | 'liuri'
  derivation_input?: ZiWeiDerivationInput
  derivation_fingerprint?: string
  base_content_hash?: string
  derived_content_hash?: string
  palaces: ZiWeiPalace[]
  life_master: string
  body_master: string
  five_bureau: string
  body_palace: string
  earthly_branch_of_soul_palace: string
  earthly_branch_of_body_palace: string
  patterns: string[]
  query_view: ZiWeiQueryView
}

export interface ZiWeiPeriodChart extends ZiWeiChartResponse {
  year: number
  month?: number
  day?: number
  description: string
}

export interface ZiWeiDayunPeriod {
  start_age: number
  end_age: number
  palace: string
  stars: string[]
  description: string
}

export interface ZiWeiReadingEvidence {
  type: string
  label: string
  value: string
  basis: string
}

export interface ZiWeiReadingSanfangContext {
  opposite: string
  trine1: string
  trine2: string
  opposite_stars: string[]
  trine1_stars: string[]
  trine2_stars: string[]
  notes: string[]
}

export interface ZiWeiReadingPatternDetail {
  name: string
  palace: string
  stars: string[]
  basis: string
  structure_status: string
  validation_status: string
}

export interface ZiWeiPalaceReading {
  palace_name: string
  palace_focus: string
  main_star_analysis: string
  aux_star_influence: string
  sihua_influence: string
  sanfang_analysis: string
  pattern_notes: string
  brightness: string
  summary: string
  key_points: string[]
  evidence: ZiWeiReadingEvidence[]
  sanfang_context: ZiWeiReadingSanfangContext | null
  pattern_details: ZiWeiReadingPatternDetail[]
  review_notes: string[]
  limitations: string[]
  evidence_basis: string
  placement_basis: string
  interpretation_basis: string
  interpretation_status: string
  validation_status: string
  is_outcome_conclusion: boolean
}

export interface ZiWeiSihuaProjectionSemantics {
  rule_id: 'ziwei.sihua.ten-stem.iztro-v1'
  source_tier: 'silver_external'
  placement_basis: 'deterministic_rule_projection'
  validation_status: 'cross_checked_not_gold'
  is_outcome_conclusion: false
}

export interface ZiWeiSihuaProjectionItem extends ZiWeiSihuaProjectionSemantics {
  transformed_star: string
  hua_type: '化禄' | '化权' | '化科' | '化忌'
  target_palace: string
  source_palace?: string
  source_palace_stem?: string
  flight_scope?: 'same_palace' | 'cross_palace'
  is_self_mutagen?: boolean
}

export interface ZiWeiSihuaProjectionResult extends ZiWeiSihuaProjectionSemantics {
  analysis_kind: 'natal_year_stem_four_hua_projection' | 'direct_palace_stem_four_hua_flights'
  hua_lu: ZiWeiSihuaProjectionItem[]
  hua_quan: ZiWeiSihuaProjectionItem[]
  hua_ke: ZiWeiSihuaProjectionItem[]
  hua_ji: ZiWeiSihuaProjectionItem[]
}

export interface ZiWeiSelfMutagen extends ZiWeiSihuaProjectionSemantics {
  palace: string
  palace_stem: string
  transformed_star: string
  hua_type: '化禄' | '化权' | '化科' | '化忌'
  structure_status: 'same_palace_transformation'
  is_self_mutagen: true
}

export interface ZiWeiPeriodEvidenceSemantics {
  placement_basis: 'deterministic_rule_projection' | string
  interpretation_basis: 'traditional_rule_labels' | string
  interpretation_status: 'not_adjudicated' | string
  is_outcome_conclusion: boolean
}

export interface ZiWeiPeriodHighlight extends ZiWeiPeriodEvidenceSemantics {
  label: string
  value: string
  note: string
}

export interface ZiWeiPeriodEvidence extends ZiWeiPeriodEvidenceSemantics {
  type: string
  label: string
  value: string
  basis: string
}

export interface ZiWeiPeriodPalaceFocus extends ZiWeiPeriodEvidenceSemantics {
  palace: string
  period_palace: string
  branch: string
  main_stars: string[]
  aux_stars: string[]
  period_stars: string[]
  four_hua: string[]
  sanfang: string[]
  reason: string
  review_note: string
}

export interface ZiWeiDayunStageAnalysis extends ZiWeiPeriodEvidenceSemantics {
  start_age: number
  end_age: number
  palace: string
  branch: string
  heavenly_stem: string
  earthly_branch: string
  gan_zhi: string
  main_stars: string[]
  aux_stars: string[]
  period_stars: string[]
  four_hua: string[]
  sanfang: string[]
  summary: string
  review_notes: string[]
  current: boolean
  nominal_age?: number
  age_basis?: string
}

export interface ZiWeiPeriodAnalysis {
  rule_version: string
  school: string
  layer: 'dayun' | 'liunian' | 'liuyue' | 'liuri' | string
  title: string
  time_label: string
  gan_zhi?: string
  summary: string
  method: string[]
  highlights: ZiWeiPeriodHighlight[]
  focus_palaces: ZiWeiPeriodPalaceFocus[]
  evidence: ZiWeiPeriodEvidence[]
  cross_layer_relations: ZiWeiPeriodLayerRelation[]
  review_notes: string[]
  limitations: string[]
  evidence_basis: 'mixed_deterministic_projection_and_unadjudicated_traditional_labels' | string
  placement_basis: 'deterministic_rule_projection' | string
  interpretation_basis: 'traditional_rule_labels' | string
  interpretation_status: 'not_adjudicated' | string
  validation_status: 'not_adjudicated' | string
  is_outcome_conclusion: boolean
  dayun_stages?: ZiWeiDayunStageAnalysis[]
  dayun_context?: ZiWeiDayunStageAnalysis
}

export interface ZiWeiPeriodResponse<T = unknown> {
  periods?: T[]
  analysis?: ZiWeiPeriodAnalysis | null
  year?: number
  month?: number
  day?: number
  target_date?: string
  nominal_age?: number
  age_basis?: string
  boundary_policy?: string
  period_key?: string
  [key: string]: unknown
}

export interface ZiWeiDayunPeriodResponse extends ZiWeiPeriodResponse<ZiWeiDayunPeriod> {
  target_date: string
  nominal_age: number
  age_basis: string
  boundary_policy: string
}

export interface ZiWeiChartPeriodResponse extends ZiWeiPeriodResponse<ZiWeiPeriodChart> {
  periods: ZiWeiPeriodChart[]
}

export interface ZiWeiSihuaPeriodResponse {
  periods: ZiWeiSihuaProjectionResult
  description: string
}

export interface ZiWeiSihuaChainResponse {
  chain: ZiWeiSihuaProjectionResult
  description: string
}

export interface ZiWeiPalaceReadingResponse {
  reading: ZiWeiPalaceReading
}

export interface ZiWeiOverlayMethodStep extends ZiWeiPeriodEvidenceSemantics {
  label: string
  value: string
  meaning: string
}

export interface ZiWeiOverlayTrigger extends ZiWeiPeriodEvidenceSemantics {
  type: string
  star?: string
  palace: string
  branch: string
  meaning: string
  polarity: 'resource' | 'constraint' | 'movement' | 'neutral' | string
}

export interface ZiWeiOverlayFocusPalace extends ZiWeiPeriodEvidenceSemantics {
  palace: string
  branch: string
  triggers: ZiWeiOverlayTrigger[]
  main_stars: string[]
  review_note: string
}

export interface ZiWeiPeriodBranchRelation {
  period_branch: string
  natal_pillar: 'year' | 'month' | 'day' | 'hour' | string
  natal_branch: string
  relation: string
  subtype?: string
  rule_id: string
  structural_status: 'observed' | 'complete' | string
  transformation_status: 'not_applicable' | 'unadjudicated' | string
  target_element?: string
  evidence_basis: 'deterministic_rule_projection' | string
  interpretation_status: 'not_adjudicated' | string
  is_outcome_conclusion: boolean
}

export interface ZiWeiHourBlock {
  stem: string
  branch: string
  stem_branch: string
  interval_start_hour: number
  interval_end_hour_exclusive: number
  interval_label: string
  crosses_midnight: boolean
  day_stem_basis: 'period_derivation_day_stem'
  boundary_policy: 'traditional_two_hour_branch_slots_no_civil_date_assignment'
  rule_id: 'ziwei.period.hour-stem.five-rat-v1'
  shi_shen: string
  relation_to_ming: string
  relation_evidence: ZiWeiPeriodBranchRelation[]
  structural_summary: string
  evidence_basis: 'deterministic_rule_projection'
  validation_status: 'not_adjudicated'
  is_outcome_conclusion: false
}

export interface ZiWeiPeriodSummaryItem {
  gan_zhi: string
  shi_shen: string
  relation: string
  relation_evidence: ZiWeiPeriodBranchRelation[]
  structural_summary: string
}

export interface ZiWeiPeriodLayerRelation {
  source_layer: 'liuyue' | 'liuri'
  source_gan_zhi: string
  source_branch: string
  target_layer: 'liunian' | 'liuyue'
  target_gan_zhi: string
  target_branch: string
  relation: string
  subtype?: string
  rule_id: string
  structural_status: string
  transformation_status: string
  target_element?: string
  evidence_basis: 'deterministic_rule_projection'
  interpretation_status: 'not_adjudicated'
  is_outcome_conclusion: false
}

export interface ZiWeiPeriodSummary {
  liunian: ZiWeiPeriodSummaryItem
  liuyue: ZiWeiPeriodSummaryItem
  liuri: ZiWeiPeriodSummaryItem
  review_notes: {
    liunian: string[]
    liuyue: string[]
    liuri: string[]
  }
  evidence_basis: 'deterministic_rule_projection'
  validation_status: 'not_adjudicated'
  is_outcome_conclusion: false
}

export interface ZiWeiOverlayAnalysis {
  year: number
  gan_zhi: string
  stem: string
  branch: string
  shi_shen?: string
  relation_to_ming: string
  relation_evidence: ZiWeiPeriodBranchRelation[]
  review_note: string
  summary: string
  method: ZiWeiOverlayMethodStep[]
  four_hua: ZiWeiOverlayTrigger[]
  annual_stars: ZiWeiOverlayTrigger[]
  focus_palaces: ZiWeiOverlayFocusPalace[]
  dayun_context?: ZiWeiDayunStageAnalysis
  evidence_basis: 'mixed_deterministic_projection_and_unadjudicated_traditional_labels' | string
  placement_basis: 'deterministic_rule_projection' | string
  interpretation_basis: 'traditional_rule_labels' | string
  interpretation_status: 'not_adjudicated' | string
  validation_status: 'not_adjudicated' | string
  is_outcome_conclusion: boolean
}

export interface ZiWeiOverlayResponse extends ZiWeiChartResponse {
  year: number
  overlay_analysis?: ZiWeiOverlayAnalysis
}

export async function fetchZiWeiChart(payload: ZiWeiChartRequest): Promise<ZiWeiChartResponse> {
  const { data } = await client.post<ZiWeiChartResponse>('/ziwei/chart', payload)
  return data
}

export async function fetchZiWeiPeriod<TResponse = ZiWeiPeriodResponse>(
  payload: ZiWeiPeriodRequest,
): Promise<TResponse> {
  const { data } = await client.post<TResponse>('/ziwei/period', payload)
  return data
}

export async function fetchZiWeiOverlay(
  payload: ZiWeiOverlayRequest,
): Promise<ZiWeiOverlayResponse> {
  const { data } = await client.post<ZiWeiOverlayResponse>('/ziwei/overlay', payload)
  return data
}
