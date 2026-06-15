import client from './client'

export type InterpretationFocus = 'overview' | 'pattern' | 'tiaohou' | 'ten_gods'
export type InterpretationStatus = 'ok' | 'fallback'

export interface InterpretationSection {
  title: string
  content: string
  citations: number[]
}

export interface InterpretationCitation {
  id: number
  book: string
  chapter: string
  path: string
  quote: string
  score: number
}

export interface BaziInterpretationResponse {
  status: InterpretationStatus
  reason: string
  chart_id: number
  focus: InterpretationFocus
  summary: string
  sections: InterpretationSection[]
  citations: InterpretationCitation[]
}

export async function fetchBaziInterpretation(chartId: number, focus: InterpretationFocus = 'overview') {
  const { data } = await client.post<BaziInterpretationResponse>('/interpretation/bazi', {
    chart_id: chartId,
    focus,
  })
  return data
}
