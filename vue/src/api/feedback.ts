import client from './client'

export type FeedbackRating = 'accurate' | 'inaccurate' | 'too_generic' | 'confusing' | 'helpful'

export interface FeedbackRequest {
  chart_id: number
  target_type?: string
  target_id?: string
  rating: FeedbackRating
  tags?: string[]
  comment?: string
  event_year?: number
  event_category?: string
  consent_research?: boolean
  consent_training?: boolean
}

export interface FeedbackResponse {
  id: number
  status: string
}

export interface FeedbackSummaryItem {
  rating: FeedbackRating
  count: number
}

export interface FeedbackSummaryResponse {
  chart_id: number
  total: number
  items: FeedbackSummaryItem[]
}

export async function submitFeedback(payload: FeedbackRequest) {
  const { data } = await client.post<FeedbackResponse>('/feedback', payload)
  return data
}

export async function fetchFeedbackSummary(chartId: number) {
  const { data } = await client.get<FeedbackSummaryResponse>('/feedback/summary', {
    params: { chart_id: chartId },
  })
  return data
}
