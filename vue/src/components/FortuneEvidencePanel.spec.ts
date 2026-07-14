import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import FortuneEvidencePanel from './FortuneEvidencePanel.vue'
import type { FortuneScoreBreakdown, ScoreEvidence } from '../api/fortune'

const supporting: ScoreEvidence[] = [
  {
    code: 'relation.stem.shengWo',
    stage: 'relation',
    category: '天干生克',
    label: '流日生扶日主',
    impact: 18,
    description: '流日五行生扶日主。',
    source: '《三命通会》十神生克规则',
  },
]
const counter: ScoreEvidence[] = [
  {
    code: 'relation.branch.clash',
    stage: 'relation',
    category: '地支关系',
    label: '地支六冲',
    impact: -30,
    description: '变动和对立信号较强。',
    source: '《协纪辨方书》地支关系规则',
  },
]
const breakdown: FortuneScoreBreakdown = {
  pipeline_version: 'fortune-score-pipeline-test',
  base_score: 50,
  relation_score: 38,
  detail_score: 65,
  final_score: 57,
  evidence_completeness: 90,
  supporting_evidence: supporting,
  counter_evidence: counter,
}

describe('FortuneEvidencePanel', () => {
  it('separates supporting and counter evidence', () => {
    const wrapper = mount(FortuneEvidencePanel, {
      props: { level: 'advanced', completeness: 90, supporting, counter, breakdown },
    })

    expect(wrapper.text()).toContain('支持证据')
    expect(wrapper.text()).toContain('反向证据')
    expect(wrapper.text()).toContain('流日五行生扶日主')
    expect(wrapper.text()).toContain('变动和对立信号较强')
    expect(wrapper.text()).toContain('证据完整度 90%')
    expect(wrapper.text()).not.toContain('置信度')
  })

  it('only exposes rule codes and versions in professional mode', () => {
    const basic = mount(FortuneEvidencePanel, {
      props: { level: 'basic', completeness: 90, supporting, counter, breakdown },
    })
    expect(basic.find('[data-testid="professional-meta"]').exists()).toBe(false)
    expect(basic.text()).not.toContain('relation.stem.shengWo')

    const professional = mount(FortuneEvidencePanel, {
      props: {
        level: 'professional',
        completeness: 90,
        supporting,
        counter,
        breakdown,
        engineVersion: 'fortune-engine-test',
        ruleVersion: 'bazi-rules-test',
      },
    })
    expect(professional.get('[data-testid="professional-meta"]').attributes('data-testid')).toBe(
      'professional-meta',
    )
    expect(professional.text()).toContain('relation.stem.shengWo')
    expect(professional.text()).toContain('fortune-engine-test')
    expect(professional.text()).toContain('bazi-rules-test')
    expect(professional.text()).toContain('最终分 57')
  })
})
