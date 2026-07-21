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
  normalizers: {
    ling: number
    di: number
    shi_sigmoid_divisor: number
    sheng_sigmoid_divisor: number
    shi_formula: 'centered_logistic_v1'
    sheng_formula: 'zero_origin_logistic_v1'
  }
  adjustment_thresholds: Record<string, number>
  yue_ling: {
    rule_id: string
    profile: string
    hash_basis: string
    day_element_order: string[]
    month_branch_order: string[]
    scores: number[][]
    score_states: Array<{ state: string; score: number }>
    table_sha256: string
    earth_month_policy: string
    validation_status: string
  }
  root: BodyStrengthRootRuleConfig
  bonus: BodyStrengthBonusRuleConfig
  influence: BodyStrengthInfluenceRuleConfig
  adjustment_force: BodyStrengthAdjustmentForceConfig
}

export interface BodyStrengthRootRuleConfig {
  rule_id: string
  profile: string
  hide_stem_weights: { main: number; middle: number; residual: number }
  terrain_weights: Record<string, number>
  root_multiplier: number
  tou_gan_multiplier: number
  tou_gan_scope: string
  validation_status: 'not_validated'
}

export interface BodyStrengthBonusRuleConfig {
  rule_id: string
  profile: string
  hash_basis: string
  day_stem_order: string[]
  lu_branches: string[]
  yang_ren_stem_order: string[]
  yang_ren_branches: string[]
  scores: {
    day_lu: number
    month_lu: number
    day_yang_ren: number
    month_yang_ren: number
  }
  yin_stem_blade_policy: string
  table_sha256: string
  validation_status: 'not_validated'
}

export interface BodyStrengthInfluenceRuleConfig {
  rule_id: string
  profile: string
  visible_stem_scope: string
  same_polarity_peer_weight: number
  opposite_polarity_peer_weight: number
  officer_killer_weight: number
  output_weight: number
  wealth_weight: number
  hidden_branch_scope: string
  hidden_branch_multiplier: number
  same_element_root_ownership: 'di'
  seal_ownership: 'sheng'
  validation_status: 'not_validated'
}

export interface BodyStrengthAdjustmentForceConfig {
  rule_id: string
  profile: 'local_posterior_force_v1'
  stem_force: number
  hidden_stem_multiplier: number
  hidden_stem_weight_source: string
  shi_ling_support_basis: string
  neutral_target: number
  validation_status: 'not_validated'
}

export interface BodyStrengthResult {
  rule_id: 'bazi.body-strength-score-candidate-v3'
  schema_version: string
  rule_version: string
  school: string
  scoring_profile: 'local_fuyi_weighted_score_v3'
  yue_ling_rule_id: string
  yue_ling_profile: string
  yue_ling_table_sha256: string
  inputs: BodyStrengthInputSnapshot
  score_band_candidate: string
  band_selection_basis: 'ordered_fixed_local_thresholds_then_posterior_adjustments'
  band_rules: BodyStrengthBandRule[]
  total_score: number
  ling_score: number
  di_score: number
  shi_score: number
  sheng_score: number
  lu_bonus: number
  components: BodyStrengthComponent[]
  evidence: BodyStrengthEvidence[]
  adjustments: BodyStrengthAdjustment[]
  status: 'observed'
  validation_status: 'not_validated'
  interpretation_status: 'not_adjudicated'
  is_strength_conclusion: false
  limitations: string[]
}

export interface BodyStrengthInputSnapshot {
  pillars: string[]
  day_stem: string
  day_element: string
  month_branch: string
}

export interface BodyStrengthBandRule {
  candidate: string
  operator: 'gt' | 'otherwise'
  threshold?: number
}

export interface BodyStrengthComponent {
  rule_id: string
  key: string
  name: string
  raw_score: number
  normalized_score: number
  weight: number
  weighted_score: number
  basis: 'local_weighted_component_profile'
  status: 'observed'
  validation_status: 'not_validated'
  description: string
}

export interface BodyStrengthEvidence {
  rule_id: string
  component: string
  polarity: string
  source: string
  item: string
  score: number
  basis: 'local_component_scoring_rule'
  status: 'observed'
  interpretation_status: 'not_adjudicated'
  reason: string
}

export interface BodyStrengthAdjustment {
  rule_id: string
  name: string
  before: number
  after: number
  reason: string
  basis: 'local_posterior_force_v1'
  status: 'observed'
  validation_status: 'not_validated'
  description: string
}

export interface TenGodRatio {
  name: string
  count: number
  percent: number
}

export interface TenGodRank {
  rank: number
  god: string
  count: number
  percent: number
  basis: 'three_visible_stems_and_all_hidden_stems_counted_equally'
  status: 'observed'
  interpretation_status: 'not_adjudicated'
}

export interface TenGodAnalysis {
  rule_id: 'bazi.ten-god-occurrence-ranking-v1'
  calculation_method: 'three_visible_stems_and_all_hidden_stems_counted_equally'
  total_occurrences: number
  dominant_gods: string[]
  dominant_percent: number
  ranked_gods: TenGodRank[]
  status: 'observed' | 'unavailable'
  validation_status: 'not_validated'
  interpretation_status: 'not_adjudicated'
  limitations: string[]
}

export interface MissingElementAnalysis {
  status: 'observed'
  rule_id: string
  missing_elements: string[]
  weak_elements: string[]
  scores: Record<string, number>
  missing_count: number
  is_yongshen_conclusion: false
  remedy_status: 'not_adjudicated'
  note: string
}

export interface TiaohouRuleEvidence {
  rule_id: string
  xi_shen: string
  ji_shen: string
  ji_shen_status: 'not_adjudicated'
  source_text: string
  basis: 'day_stem_month_branch_table'
  status: 'observed'
  validation_status: 'not_validated'
  interpretation_status: 'not_adjudicated'
}

export interface MonthCommandDepthCandidate {
  rule_id: string
  profile_id: string
  source: string
  sequence: string
  commanding_stem: string
  segment: string
  segment_start_day: number
  segment_end_day?: number
  position_day: number
  basis: 'elapsed_since_month_jie'
  status: 'observed'
  interpretation_status: 'not_adjudicated'
}

export interface TiaohouDepthEvidence {
  rule_id: 'bazi.tiaohou.solar-term-depth-v1'
  status: 'observed' | 'unavailable'
  basis: 'solar_term_jie_interval' | 'birth_instant_unavailable'
  phase?: string
  start_term?: string
  start_at?: string
  end_term?: string
  end_at?: string
  elapsed_seconds?: number
  interval_seconds?: number
  position?: number
  month_command_candidates: MonthCommandDepthCandidate[]
  note: string
  interpretation_status: 'not_adjudicated'
}

export interface TiaohouBranchStructure {
  rule_id: string
  type: '三合局'
  branches: string[]
  target_element: string
  transformation_status: 'unadjudicated' | 'disputed'
}

export interface TiaohouChartEvidence {
  status: 'observed' | 'unavailable'
  basis: 'four_pillars_and_complete_branch_structures' | 'four_pillars_unavailable'
  visible_stems: string[]
  branches: string[]
  complete_branch_structures: TiaohouBranchStructure[]
}

export interface TiaohouConditionHit {
  rule_id: string
  candidates: string[]
  condition: string
  evidence: string[]
  source: string
  source_text: string
  status: 'matched'
  validation_status: 'classical_source_reviewed'
  interpretation_status: 'not_adjudicated'
}

export interface TiaohouResult {
  rule_id: 'bazi.tiaohou.table-candidates-v3'
  stem: string
  month: string
  rules: TiaohouRuleEvidence[]
  table_primary_candidate: string
  selection_basis: 'first_table_entry_candidate'
  depth_affects_selection: false
  depth_evidence: TiaohouDepthEvidence
  chart_evidence: TiaohouChartEvidence
  matched_conditions: TiaohouConditionHit[]
  chart_candidates: string[]
  chart_selection_basis:
    | 'chart_unavailable'
    | 'no_reviewed_chart_condition_match'
    | 'reviewed_four_pillar_condition_match'
  status: 'observed'
  validation_status: 'not_validated'
  interpretation_status: 'not_adjudicated'
  limitations: string[]
}

export interface GanRelation {
  id: string
  rule_id: string
  pillar1: string
  pillar2: string
  pillars: string[]
  stems: string[]
  type: string
  subtype?: string
  status: 'observed' | 'disputed'
  structure_status: string
  transformation_status: 'not_applicable' | 'unadjudicated' | 'disputed'
  target_element?: string
  direction?: string
  proximity: 'adjacent' | 'remote'
  priority: number
  conflicts_with: string[]
  dispute_reasons: string[]
  evidence: string[]
  transformation_evidence?: {
    month_branch: string
    month_element: string
    month_supports_target: boolean
    target_stem_exposed: boolean
    target_root_branches: string[]
    note: string
  }
  detail: string
}

export interface ZhiRelation {
  id: string
  rule_id: string
  pillar1: string
  pillar2: string
  pillars: string[]
  branches: string[]
  type: string
  subtype?: string
  status: 'observed' | 'partial' | 'complete' | 'disputed'
  structure_status: string
  transformation_status: 'not_applicable' | 'unadjudicated' | 'disputed'
  target_element?: string
  priority: number
  conflicts_with: string[]
  dispute_reasons: string[]
  evidence: string[]
  detail: string
}

export interface GanZhiAnalysis {
  gan_relations: GanRelation[]
  zhi_relations: ZhiRelation[]
}

export interface ShenShaMeta {
  name: string
  rule_id: string
  basis: string
  status: 'observed'
  interpretation_status: 'not_adjudicated'
}

export interface FortuneLayerSet {
  rule_version: string
  school: string
  dayun: FortuneLayer
  liunian: FortuneLayer
  liuyue: FortuneLayer
  xiaoyun: FortuneLayer
  inter_layer_relations: FortuneLayerRelation[]
}

export interface FortuneLayer {
  rule_id: string
  key: string
  name: string
  pillar: string
  gan: string
  zhi: string
  start_age?: number
  end_age?: number
  start_at?: string
  end_at_exclusive?: string
  age?: number
  year?: number
  month?: number
  ten_god: FortuneLayerTenGod
  relations: FortuneLayerRelation[]
  shen_sha_details: ShenShaActivation[]
  basis: string
  status: 'observed' | 'unavailable'
  interpretation_status: 'not_adjudicated'
}

export interface FortuneLayerTenGod {
  rule_id: string
  reference_stem: string
  query_stem: string
  name: string
  basis: string
  status: 'observed' | 'unavailable'
  interpretation_status: 'not_adjudicated'
}

export interface FortuneLayerRelation {
  rule_id: string
  source: string
  source_value: string
  target: string
  target_value: string
  type: string
  name: string
  basis: string
  status: 'observed'
  interpretation_status: 'not_adjudicated'
}

export interface ShenShaActivation {
  name: string
  rule_id: string
  basis: string
  status: 'observed'
  interpretation_status: 'not_adjudicated'
  activation: string
}

export interface BirthValidation {
  normalization_version: string
  input_calendar: 'SOLAR' | 'LUNAR'
  original_date_time: string
  converted_solar_date_time: string
  calculation_date_time: string
  lunar_date: string
  current_solar_term: string
  current_solar_term_started_at: string
  birth_place?: string
  timezone: string
  utc_date_time: string
  local_time_ambiguous: boolean
  possible_utc_offset_seconds: number[]
  longitude?: number
  true_solar_time_applied: boolean
  true_solar_adjustment_minutes: number
  timezone_offset_seconds: number
  mean_solar_adjustment_seconds: number
  equation_of_time_seconds: number
  true_solar_adjustment_seconds: number
  true_solar_algorithm?: string
  true_solar_source?: string
  true_solar_within_validated_range: boolean
  true_solar_uncertainty_seconds: number
  time_uncertain: boolean
  zi_hour_policy: 'late_zi_next_day' | 'late_zi_same_day'
  uncertainty_seconds: number
  notices: string[]
}

export interface ChartPillar {
  gan: string
  zhi: string
}

export interface PatternCandidate {
  rule_id: string
  pattern_name: string
  category: '结构格局' | '辅助特征' | string
  source: string
}

export interface PatternInputSnapshot {
  pillars: string[]
  month_branch: string
}

export interface MonthCommandStemExposure {
  pillar: '年干' | '月干' | '时干'
  stem: string
  ten_god: string
  exact_hidden_stem: true
}

export interface MonthCommandPatternEvidence {
  rule_id: 'bazi.pattern.month-command-exposure.v1'
  month_branch: string
  hidden_stem: string
  hidden_stem_type: string
  hidden_ten_god: string
  exposures: MonthCommandStemExposure[]
  candidate_names: string[]
  exposure_status: 'exact_hidden_stem_exposed'
  month_special_structure?: string
  source: string
  status: 'observed'
  interpretation_status: 'pattern_candidate_not_adjudicated'
  is_established_pattern: false
}

export interface PatternDetectorProfileDigest {
  rule_id: string
  algorithm_sha256: string
  behavior_sha256: string
  profile_sha256: string
}

export type PatternDetectorProfileChangeClass =
  | 'detector_added'
  | 'detector_removed'
  | 'algorithm_digest_changed'
  | 'behavior_evidence_digest_changed'
  | 'semantic_profile_digest_changed'
  | 'layered_digests_unchanged'

export interface PatternDetectorProfileChangeContract {
  scheme: 'layered_detector_digest_delta_v1'
  alignment_key: 'rule_id'
  classes: PatternDetectorProfileChangeClass[]
  behavior_evidence_scope: 'simple_full_truth_table_complex_partial_contract'
  inference_boundary: 'digest_evidence_only'
}

export interface PatternDetectorProfileMigrationReference {
  ledger_id: 'bazi.pattern-detector-profile-migrations'
  schema: 'pattern_detector_profile_migration_ledger_v2'
  sha256: 'a72422e12e07adae349c147b3581f8c4829368f134f00a4f229c9a1c29d21825'
  migration_count: number
  latest_migration_id: 'bazi.pattern-candidate-set-v33_to_v34'
  latest_from_snapshot_id: 'bazi.pattern-candidate-set-v33'
  latest_to_snapshot_id: 'bazi.pattern-candidate-set-v34'
  change_scheme: 'layered_detector_digest_delta_v1'
  claim_boundary: 'digest_evidence_only'
  chain_scheme: 'pattern_detector_profile_migration_chain_v1'
  chain_head_sha256: '07dc296ad9e5dd0f834e40256c1e0f6033eb0ded435d0c76be6a0602ae0113bd'
}

export interface PatternDetectorProfileReleaseAnchorReference {
  schema: 'pattern_detector_profile_release_anchor_v1'
  anchor_id: 'bazi.pattern-detector-profile-release-anchor-v34'
  artifact_path: 'release/pattern-detector-profile-anchor.json'
  sha256: 'ebd6323f28715695aa3c4ee9038e74d261c9fa34b422037266c4b097e3086a2e'
  verification_profile: 'repository_ci_cross_check_v1'
  trust_boundary: 'unsigned_repository_ci_artifact'
  claim_boundary: 'digest_evidence_only'
}

export interface PatternAnalysis {
  rule_id: 'bazi.pattern-candidate-set-v34'
  schema_version: string
  detector_profile: 'classical_structural_detectors_v45'
  detector_count: number
  detector_manifest_sha256: string
  detector_profiles: PatternDetectorProfileDigest[]
  detector_profile_change_contract: PatternDetectorProfileChangeContract
  detector_profile_migration: PatternDetectorProfileMigrationReference
  detector_profile_release_anchor: PatternDetectorProfileReleaseAnchorReference
  inputs: PatternInputSnapshot
  candidates: PatternCandidate[]
  month_command_evidence: MonthCommandPatternEvidence[]
  status: 'observed' | 'observed_without_structural_candidate' | 'invalid_input'
  validation_status: 'not_validated' | 'invalid_input'
  interpretation_status: 'not_adjudicated' | 'not_available'
  limitations: string[]
}

export interface ChartSummary {
  id: number
  name: string
  gender: string
  zi_hour_policy: 'late_zi_next_day' | 'late_zi_same_day'
  birth_year: number
  birth_month: number
  birth_day: number
  birth_hour: number
  birth_min: number
  birth_sec: number
  calendar_type: string
  lunar_leap_month?: boolean
  birth_place?: string
  timezone?: string
  birth_utc_offset_seconds?: number
  longitude?: number
  use_true_solar_time?: boolean
  time_uncertain?: boolean
  uncertainty_seconds?: number
  selected_candidate_id?: string
  engine_version?: string
  stored_rule_version?: string
  created_at?: string
  updated_at?: string
}

export interface MonthSeasonEvidence {
  rule_id: string
  month_branch: string
  traditional_month: number
  season: '春' | '夏' | '秋' | '冬'
  basis: 'month_pillar_branch'
  status: 'observed'
}

export interface NaYinEvidence {
  rule_id: 'nayin.sixty-cycle-v1'
  gan_zhi: string
  name: string
  element: string
  basis: 'pillar_gan_zhi'
  status: 'observed'
}

export interface ChartDetail extends ChartSummary {
  birth_validation?: BirthValidation
  rule_version?: string
  school?: string
  rule_meta?: RuleMeta
  year_pillar?: ChartPillar
  month_pillar?: unknown
  day_pillar?: unknown
  hour_pillar?: unknown
  five_elements?: Record<string, number>
  element_detail?: unknown
  body_strength?: BodyStrengthResult
  ten_gods?: unknown
  na_yin?: Record<string, NaYinEvidence>
  hidden_stems?: unknown
  da_yun_start?: unknown
  da_yun?: unknown
  gan_zhi_analysis?: GanZhiAnalysis
  pattern_analysis?: PatternAnalysis
  ming_gong?: unknown
  pillar_details?: unknown
  tiaohou?: TiaohouResult
  global_shen_sha?: string[]
  global_shen_sha_details?: ShenShaMeta[]
  day_shen_sha?: string[]
  day_shen_sha_details?: ShenShaMeta[]
  month_season?: MonthSeasonEvidence
  shen_sha_by_pillar?: unknown
  ten_god_proportion?: TenGodRatio[]
  ten_god_analysis?: TenGodAnalysis
  missing_elements?: MissingElementAnalysis
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
  birth_min: number
  birth_sec: number
  calendar_type: 'SOLAR' | 'LUNAR'
  lunar_leap_month?: boolean
  gender: 'MALE' | 'FEMALE'
  zi_hour_policy: 'late_zi_next_day' | 'late_zi_same_day'
  name?: string
  birth_place?: string
  timezone?: string
  birth_utc_offset_seconds?: number
  longitude?: number
  use_true_solar_time?: boolean
  time_uncertain?: boolean
  uncertainty_seconds?: number
  candidate_id?: string
}

export interface UncertaintyBoundary {
  type: 'hour_branch' | 'zi_hour_day_boundary' | 'civil_day' | 'solar_term' | string
  at: string
  name?: string
}

export interface BirthUncertainty {
  seconds: number
  algorithm_uncertainty_seconds: number
  effective_seconds: number
  input_range_start: string
  input_range_end: string
  evaluation_range_start: string
  evaluation_range_end: string
  calculation_range_start: string
  calculation_range_end: string
  crossed_boundaries: UncertaintyBoundary[]
}

export interface ChartCandidate {
  candidate_id: string
  input_range_start: string
  input_range_end: string
  calculation_range_start: string
  calculation_range_end: string
  representative_time: string
  birth_validation: BirthValidation
  year_pillar: ChartPillar
  month_pillar: ChartPillar
  day_pillar: ChartPillar
  hour_pillar: ChartPillar
  da_yun_start_at_min: string
  da_yun_start_at_max: string
}

export interface ChartPreviewResponse extends ChartDetail {
  birth_validation: BirthValidation
  year_pillar: ChartPillar
  month_pillar: ChartPillar
  day_pillar: ChartPillar
  hour_pillar: ChartPillar
  uncertainty: BirthUncertainty
  candidate_charts: ChartCandidate[]
  stable_fields: string[]
  unstable_fields: string[]
  requires_candidate_selection: boolean
  selected_candidate_id?: string
}

export interface ChartListResponse {
  charts: ChartSummary[]
  total: number
  page: number
  page_size: number
}

export async function previewChart(payload: ChartCreateRequest) {
  const { data } = await client.post<ChartPreviewResponse>('/chart/preview', payload)
  return data
}

export async function createChart(payload: ChartCreateRequest) {
  const { data } = await client.post<ChartPreviewResponse>('/chart', payload)
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
