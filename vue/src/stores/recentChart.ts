import { computed, ref } from 'vue'
import { defineStore } from 'pinia'

const STORAGE_KEY = 'bazi_last_birth'

export interface RecentChart {
  name: string
  calendarType: 'SOLAR' | 'LUNAR'
  lunarLeapMonth: boolean
  year: number
  month: number
  day: number
  hour: number
  minute: number
  second: number
  gender: 'MALE' | 'FEMALE'
  ziHourPolicy: 'late_zi_next_day' | 'late_zi_same_day'
  birthPlace: string
  timezone: string
  birthUTCOffsetSeconds?: number
  longitude?: number
  useTrueSolarTime: boolean
  timeUncertain: boolean
  uncertaintySeconds: number
  chartId?: number
}

export interface RecentChartApiSource {
  id?: number
  name?: string
  birth_year?: number
  birth_month?: number
  birth_day?: number
  birth_hour?: number
  birth_min?: number
  birth_sec?: number
  calendar_type?: string
  lunar_leap_month?: boolean
  gender?: string
  zi_hour_policy?: string
  birth_place?: string
  timezone?: string
  birth_utc_offset_seconds?: number
  longitude?: number
  use_true_solar_time?: boolean
  time_uncertain?: boolean
  uncertainty_seconds?: number
}

function finiteNumber(value: unknown, fallback: number): number {
  const parsed = Number(value)
  return Number.isFinite(parsed) ? parsed : fallback
}

function optionalNumber(value: unknown): number | undefined {
  if (value === null || value === undefined || value === '') return undefined
  const parsed = Number(value)
  return Number.isFinite(parsed) ? parsed : undefined
}

function normalizeRecentChart(value: unknown): RecentChart | null {
  if (!value || typeof value !== 'object') return null
  const input = value as Record<string, unknown>
  const year = finiteNumber(input.year, 0)
  const month = finiteNumber(input.month, 0)
  const day = finiteNumber(input.day, 0)
  if (year < 1 || month < 1 || day < 1) return null

  const chartId = optionalNumber(input.chartId)
  return {
    name: typeof input.name === 'string' ? input.name : '',
    calendarType: input.calendarType === 'LUNAR' ? 'LUNAR' : 'SOLAR',
    lunarLeapMonth: Boolean(input.lunarLeapMonth),
    year,
    month,
    day,
    hour: finiteNumber(input.hour, 8),
    minute: finiteNumber(input.minute, 0),
    second: finiteNumber(input.second, 0),
    gender:
      input.gender === 'FEMALE' || input.gender === 'female' || input.gender === '女'
        ? 'FEMALE'
        : 'MALE',
    ziHourPolicy:
      input.ziHourPolicy === 'late_zi_same_day' ? 'late_zi_same_day' : 'late_zi_next_day',
    birthPlace: typeof input.birthPlace === 'string' ? input.birthPlace : '',
    timezone:
      typeof input.timezone === 'string' && input.timezone
        ? input.timezone
        : Intl.DateTimeFormat().resolvedOptions().timeZone || 'Asia/Shanghai',
    birthUTCOffsetSeconds: optionalNumber(input.birthUTCOffsetSeconds),
    longitude: optionalNumber(input.longitude),
    useTrueSolarTime: Boolean(input.useTrueSolarTime),
    timeUncertain: Boolean(input.timeUncertain),
    uncertaintySeconds: finiteNumber(input.uncertaintySeconds, 0),
    chartId: chartId && chartId > 0 ? chartId : undefined,
  }
}

function loadRecentChart(): RecentChart | null {
  const raw = localStorage.getItem(STORAGE_KEY)
  if (!raw) return null
  try {
    const parsed = normalizeRecentChart(JSON.parse(raw))
    if (!parsed) localStorage.removeItem(STORAGE_KEY)
    return parsed
  } catch {
    localStorage.removeItem(STORAGE_KEY)
    return null
  }
}

export function recentChartFromApi(source: RecentChartApiSource): RecentChart | null {
  return normalizeRecentChart({
    name: source.name,
    calendarType: source.calendar_type,
    lunarLeapMonth: source.lunar_leap_month,
    year: source.birth_year,
    month: source.birth_month,
    day: source.birth_day,
    hour: source.birth_hour,
    minute: source.birth_min,
    second: source.birth_sec,
    gender: source.gender,
    ziHourPolicy: source.zi_hour_policy,
    birthPlace: source.birth_place,
    timezone: source.timezone,
    birthUTCOffsetSeconds: source.birth_utc_offset_seconds,
    longitude: source.longitude,
    useTrueSolarTime: source.use_true_solar_time,
    timeUncertain: source.time_uncertain,
    uncertaintySeconds: source.uncertainty_seconds,
    chartId: source.id,
  })
}

export const useRecentChartStore = defineStore('recent-chart', () => {
  const recentChart = ref<RecentChart | null>(loadRecentChart())
  const chartId = computed(() => recentChart.value?.chartId ?? null)

  function save(value: RecentChart) {
    const normalized = normalizeRecentChart(value)
    if (!normalized) {
      clear()
      return
    }
    recentChart.value = normalized
    localStorage.setItem(STORAGE_KEY, JSON.stringify(normalized))
  }

  function saveFromApi(source: RecentChartApiSource) {
    const normalized = recentChartFromApi(source)
    if (normalized) save(normalized)
  }

  function clear() {
    recentChart.value = null
    localStorage.removeItem(STORAGE_KEY)
  }

  return { recentChart, chartId, save, saveFromApi, clear }
})
