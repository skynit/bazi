import client from './client'

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
  year_pillar?: unknown
  month_pillar?: unknown
  day_pillar?: unknown
  hour_pillar?: unknown
  five_elements?: Record<string, number>
  element_detail?: unknown
  body_strength?: unknown
  ten_gods?: unknown
  na_yin?: unknown
  da_yun_start?: unknown
  da_yun?: unknown
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
