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
  algorithm?: 'default' | 'zhongzhou'
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

export interface ZiWeiOverlayRequest {
  chart_id: number
  year: number
}

export async function fetchZiWeiChart(payload: ZiWeiChartRequest) {
  const { data } = await client.post('/ziwei/chart', payload)
  return data
}

export async function fetchZiWeiPeriod(payload: ZiWeiPeriodRequest) {
  const { data } = await client.post('/ziwei/period', payload)
  return data
}

export async function fetchZiWeiOverlay(payload: ZiWeiOverlayRequest) {
  const { data } = await client.post('/ziwei/overlay', payload)
  return data
}
