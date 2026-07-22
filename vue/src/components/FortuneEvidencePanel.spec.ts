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
    evidence_basis: 'empirical',
    validation_status: 'not_validated',
    interpretation_status: 'not_adjudicated',
    is_outcome_conclusion: false,
  },
]
const counter: ScoreEvidence[] = [
  {
    code: 'relation.branch.clash',
    stage: 'relation',
    category: '地支关系',
    label: '地支六冲',
    impact: -30,
    description: '流日支与日支命中六冲结构。',
    source: '《协纪辨方书》地支关系规则',
    evidence_basis: 'empirical',
    validation_status: 'not_validated',
    interpretation_status: 'not_adjudicated',
    is_outcome_conclusion: false,
  },
]
const breakdown: FortuneScoreBreakdown = {
  pipeline_version: 'fortune-score-pipeline-test',
  score_kind: 'structural_relation_index',
  evidence_basis: 'empirical',
  validation_status: 'not_validated',
  interpretation_status: 'not_adjudicated',
  is_outcome_probability: false,
  base_score: 50,
  relation_score: 38,
  final_score: 38,
  evidence_completeness: 90,
  supporting_evidence: supporting,
  counter_evidence: counter,
}

describe('FortuneEvidencePanel', () => {
  it('separates supporting and counter evidence', () => {
    const wrapper = mount(FortuneEvidencePanel, {
      props: { level: 'advanced', supporting, counter, breakdown },
    })

    expect(wrapper.text()).toContain('生扶关系')
    expect(wrapper.text()).toContain('冲克关系')
    expect(wrapper.text()).toContain('流日五行生扶日主')
    expect(wrapper.text()).toContain('流日支与日支命中六冲结构')
    expect(wrapper.text()).not.toContain('证据完整度')
    expect(wrapper.text()).not.toContain('not_validated')
    expect(wrapper.text()).not.toContain('not_adjudicated')
  })

  it('keeps internal rule codes and versions hidden in every mode', () => {
    const basic = mount(FortuneEvidencePanel, {
      props: { level: 'basic', supporting, counter, breakdown },
    })
    expect(basic.find('[data-testid="professional-meta"]').exists()).toBe(false)
    expect(basic.text()).not.toContain('relation.stem.shengWo')

    const professional = mount(FortuneEvidencePanel, {
      props: {
        level: 'professional',
        supporting,
        counter,
        breakdown,
      },
    })
    expect(professional.get('[data-testid="professional-meta"]').attributes('data-testid')).toBe(
      'professional-meta',
    )
    expect(professional.text()).toContain('最终值 38')
    expect(professional.text()).toContain('依据：《三命通会》十神生克规则')
    expect(professional.text()).not.toContain('relation.stem.shengWo')
    expect(professional.text()).not.toContain('fortune-score-pipeline-test')
    expect(professional.text()).not.toContain('not_validated')
    expect(professional.text()).not.toContain('not_adjudicated')
  })
})
