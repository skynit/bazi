import { mount } from '@vue/test-utils'
import { afterEach, describe, expect, it, vi } from 'vitest'
import ComputationLog from './ComputationLog.vue'

afterEach(() => {
  vi.clearAllTimers()
  vi.useRealTimers()
})

describe('ComputationLog evidence wording', () => {
  it('shows actual chart evidence without inventing useful-god or seasonal conclusions', () => {
    vi.useFakeTimers()
    const wrapper = mount(ComputationLog, {
      props: {
        chartData: {
          day_pillar: { gan: '甲', zhi: '子' },
          five_elements: { 金: 2, 木: 11, 水: 8, 火: 4, 土: 6 },
          body_strength: { score_band_candidate: '偏弱' },
        },
      },
    })

    const text = wrapper.text()
    expect(text).toContain('身强本地分段候选【偏弱】')
    expect(text).toContain('原始五行计分：金 2 | 木 11 | 水 8 | 火 4 | 土 6')
    expect(text).not.toContain('当前用神')
    expect(text).not.toContain('生于当令之月')
    expect(text).not.toContain('因果律报告')

    wrapper.unmount()
  })
})
