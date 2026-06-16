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
