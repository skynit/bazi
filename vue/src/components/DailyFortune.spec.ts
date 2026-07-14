import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it } from 'vitest'
import DailyFortune from './DailyFortune.vue'

const evidence = {
  code: 'relation.stem.shengWo',
  stage: 'relation',
  category: '天干生克',
  label: '流日生扶日主',
  impact: 18,
  description: '流日五行生扶日主。',
  source: '规则来源',
}

beforeEach(() => localStorage.clear())

describe('DailyFortune interpretation levels', () => {
  it('defaults to ordinary mode and progressively reveals advanced and professional content', async () => {
    const wrapper = mount(DailyFortune, {
      props: {
        solarDate: '2026-07-10',
        dayGanZhi: '甲子',
        fortuneScore: 68,
        evidenceCompleteness: 88,
        supportingEvidence: [evidence],
        counterEvidence: [],
        engineVersion: 'fortune-engine-test',
        ruleVersion: 'rules-test',
        scoreBreakdown: {
          pipeline_version: 'pipeline-test',
          base_score: 50,
          relation_score: 68,
          detail_score: 68,
          final_score: 68,
          evidence_completeness: 88,
          supporting_evidence: [evidence],
          counter_evidence: [],
        },
        fortuneCategories: [{ name: '事业', score: 70, weight: 18 }],
      },
      global: { stubs: { Teleport: true } },
    })

    expect(wrapper.text()).toContain('普通')
    expect(wrapper.text()).not.toContain('运势分析')
    expect(wrapper.find('[data-testid="professional-meta"]').exists()).toBe(false)

    await wrapper.get('[data-level="advanced"]').trigger('click')
    expect(wrapper.text()).toContain('运势分析')
    expect(wrapper.text()).not.toContain('日课推算')
    expect(wrapper.find('[aria-label="分项运势"]').exists()).toBe(true)

    await wrapper.get('[data-level="professional"]').trigger('click')
    expect(wrapper.text()).toContain('日课推算')
    expect(wrapper.get('[data-testid="professional-meta"]').attributes('data-testid')).toBe(
      'professional-meta',
    )
    expect(localStorage.getItem('fortune-interpretation-level')).toBe('professional')
  })
})
