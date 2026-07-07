import type { FortuneGuide } from '../api/fortune'

export interface LegacyLuckyGuide {
  colors?: string
  numbers?: string
  actions?: string
  outfit?: string
  favorable_elems?: string[]
  unfavorable_elems?: string[]
}

export interface FortuneGuideSource {
  analysis?: {
    lucky_guide?: LegacyLuckyGuide
  }
  auspicious_hours?: string[]
  guide?: FortuneGuide
  wealth_direction?: string
}

export function todayString(date = new Date()) {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
}

export function legacyGuide(fortuneData: FortuneGuideSource): FortuneGuide | undefined {
  const old = fortuneData.analysis?.lucky_guide
  if (!old) return undefined

  return {
    precision_level: 'legacy',
    confidence: 50,
    primary_element: old.favorable_elems?.[0] ?? '',
    secondary_element: old.favorable_elems?.[1] ?? '',
    avoid_element: old.unfavorable_elems?.[0] ?? '',
    lucky_colors: old.colors ? [{ label: '幸运色', value: old.colors, reason: '按旧版喜用五行规则生成。' }] : [],
    lucky_numbers: old.numbers ? [{ label: '幸运数', value: old.numbers, reason: '按旧版喜用五行规则生成。' }] : [],
    face_direction: { label: '朝向', value: '', reason: '' },
    wealth_direction: { label: '财位', value: fortuneData.wealth_direction ?? '', reason: '按流日天干财位生成。' },
    avoid_direction: { label: '避开', value: '', reason: '' },
    recommended_actions: old.actions ? [{ label: '动作', value: old.actions, reason: '按旧版开运动作生成。' }] : [],
    cautions: [],
    best_hours: (fortuneData.auspicious_hours ?? []).slice(0, 3).map((hour) => ({
      label: '吉时',
      value: hour,
      reason: '按流日地支取吉时。',
    })),
    analysis: '',
    strategy: old.outfit ? `穿搭建议：${old.outfit}` : '',
  }
}

export function activeGuide(fortuneData: FortuneGuideSource): FortuneGuide | undefined {
  return fortuneData.guide ?? legacyGuide(fortuneData)
}
