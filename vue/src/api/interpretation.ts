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
  author: string
  edition: string
  volume: string
  chapter: string
  page: string
  locator: string
  path: string
  artifact_path: string
  artifact_sha256: string
  document_sha256: string
  quote: string
  quote_sha256: string
  source_tier: string
  verification_status: string
  artifact_kind: string
  provenance_status: string
  independence_status: string
  coverage_status: string
  catalog_schema: string
  catalog_version: string
  catalog_sha256: string
  claim_eligible: boolean
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
