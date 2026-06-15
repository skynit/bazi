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

export async function submitFeedback(payload: FeedbackRequest) {
  const { data } = await client.post<FeedbackResponse>('/feedback', payload)
  return data
}

