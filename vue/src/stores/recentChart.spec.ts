import { beforeEach, describe, expect, it } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useRecentChartStore } from './recentChart'

describe('recent chart store', () => {
  beforeEach(() => {
    localStorage.clear()
    setActivePinia(createPinia())
  })

  it('normalizes API chart fields into one persistent contract', () => {
    const store = useRecentChartStore()
    store.saveFromApi({
      id: 42,
      birth_year: 2000,
      birth_month: 8,
      birth_day: 16,
      birth_hour: 3,
      gender: '女',
      calendar_type: 'LUNAR',
      zi_hour_policy: 'late_zi_same_day',
    })

    expect(store.chartId).toBe(42)
    expect(store.recentChart).toMatchObject({
      chartId: 42,
      year: 2000,
      month: 8,
      day: 16,
      gender: 'FEMALE',
      calendarType: 'LUNAR',
      ziHourPolicy: 'late_zi_same_day',
    })
    expect(JSON.parse(localStorage.getItem('bazi_last_birth') || '{}').chartId).toBe(42)
  })

  it('removes malformed persisted state', () => {
    localStorage.setItem('bazi_last_birth', '{broken')
    setActivePinia(createPinia())

    const store = useRecentChartStore()

    expect(store.recentChart).toBeNull()
    expect(localStorage.getItem('bazi_last_birth')).toBeNull()
  })
})
