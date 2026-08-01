import client from './client'

export interface BuyiRecord {
  id: number
  hexagram_number: number
  hexagram_name: string
  summary: string
  human_way: string
  image_reading: string
  advice: string
  source: string
  created_at: string
}

export interface BuyiTodayResponse {
  date: string
  has_record: boolean
  already_drawn: boolean
  record: BuyiRecord | null
}

export async function fetchBuyiToday() {
  const { data } = await client.get<BuyiTodayResponse>('/buyi/today')
  return data
}

export async function drawBuyiToday() {
  const { data } = await client.post<BuyiTodayResponse>('/buyi/today')
  return data
}
