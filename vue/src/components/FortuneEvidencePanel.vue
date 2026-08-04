<script setup lang="ts">
import { computed } from 'vue'
import type { FortuneScoreBreakdown, InterpretationLevel, ScoreEvidence } from '../api/fortune'

const props = withDefaults(
  defineProps<{
    level: InterpretationLevel
    supporting?: ScoreEvidence[]
    counter?: ScoreEvidence[]
    breakdown?: FortuneScoreBreakdown
  }>(),
  {
    supporting: () => [],
    counter: () => [],
    breakdown: undefined,
  },
)

const allEvidence = computed(() => {
  const seen = new Set<string>()
  return [...props.supporting, ...props.counter].filter((item) => {
    if (seen.has(item.code)) return false
    seen.add(item.code)
    return true
  })
})

const visibleEvidence = computed(() =>
  props.level === 'basic' ? allEvidence.value.slice(0, 2) : allEvidence.value,
)

const countText = computed(() => {
  const total = allEvidence.value.length
  if (props.level === 'basic' && visibleEvidence.value.length < total) {
    return `已展示 ${visibleEvidence.value.length}/${total} 条`
  }
  return `${total} 条`
})

function categoryLabel(item: ScoreEvidence) {
  if (item.code.includes('.stem.')) return '天干关系'
  if (item.code.includes('.branch.')) return '地支关系'
  return item.category || '结构关系'
}

function formatImpact(value: number) {
  return value > 0 ? `+${value}` : String(value)
}
</script>

<template>
  <section class="evidence-panel glass-card" aria-label="运势证据">
    <header class="evidence-head">
      <div>
        <span class="evidence-eyebrow">今日干支结构</span>
        <h2>今天记录到的结构关系</h2>
      </div>
      <span class="evidence-count">{{ countText }}</span>
    </header>

    <p class="evidence-note">
      以下内容只记录今日干支与命盘之间命中的规则关系。名称和说明来自原始证据，按原文展示；它们不表示吉凶、现实结果或发生概率，也不是行动建议。
    </p>

    <ul v-if="visibleEvidence.length" class="evidence-list">
      <li v-for="item in visibleEvidence" :key="item.code">
        <div class="evidence-item-head">
          <div>
            <span class="relation-kind">{{ categoryLabel(item) }}</span>
            <strong>{{ item.label }}</strong>
          </div>
          <em v-if="level === 'professional'">本地权重 {{ formatImpact(item.impact) }}</em>
        </div>
        <p>{{ item.description }}</p>
        <small v-if="level === 'professional'">规则口径：{{ item.source }}</small>
      </li>
    </ul>
    <p v-else class="empty-evidence">今天没有记录到可展示的干支关系。</p>

    <p v-if="level === 'basic' && visibleEvidence.length < allEvidence.length" class="more-note">
      切换到“详细”可查看其余关系。
    </p>

    <div v-if="level === 'professional'" class="professional-meta" data-testid="professional-meta">
      <div class="score-flow" v-if="breakdown">
        <span>比较基准 {{ breakdown.base_score }}</span>
        <span>关系阶段比较值 {{ breakdown.relation_score }}</span>
        <strong>内部比较值 {{ breakdown.final_score }}</strong>
      </div>
      <p class="calculation-note">
        ‘本地权重’是当前关系计算阶段用于内部比较值的加减点数。证据基于经验规则，尚未验证，尚未对现实结果含义作裁决；每条证据不是现实结果结论，内部比较值也不是结果概率。来源仅用于复核规则口径，不表示吉凶、可靠性或行动建议。
      </p>
    </div>
  </section>
</template>

<style scoped>
.evidence-panel {
  margin-top: 0.75rem;
  padding: 1.25rem;
}
.evidence-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
}
.evidence-count {
  color: var(--text-muted);
  font-size: var(--fs-xs, 0.76rem);
}
.evidence-eyebrow {
  color: var(--text-muted);
  font-size: var(--fs-2xs, 0.68rem);
  letter-spacing: 0;
}
h2 {
  margin: 0.18rem 0 0;
  font-size: var(--fs-lg, 1.1rem);
  letter-spacing: 0.03em;
  font-family: var(--font-serif), serif;
}
.evidence-note {
  max-width: 72ch;
  margin: 0.55rem 0 1rem;
  color: var(--text-muted);
  font-size: var(--fs-xs, 0.76rem);
  line-height: 1.6;
}
.evidence-item-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.7rem;
}
.evidence-item-head > div {
  display: grid;
  gap: 0.2rem;
}
.evidence-list {
  display: grid;
  gap: 0.55rem;
  margin: 0;
  padding: 0;
  list-style: none;
}
.evidence-list li {
  padding: 0.8rem;
  border: 1px solid var(--line-subtle);
  border-radius: 8px;
  background: color-mix(in oklab, var(--surface-1) 72%, transparent);
}
.evidence-list li strong {
  font-size: var(--fs-sm, 0.88rem);
}
.evidence-list li em {
  color: color-mix(in oklab, var(--accent) 48%, var(--text));
  font-style: normal;
  font-variant-numeric: tabular-nums;
  font-size: var(--fs-xs, 0.76rem);
}
.relation-kind {
  color: var(--text-soft, var(--text-muted));
  font-size: var(--fs-2xs, 0.68rem);
}
.evidence-list li p,
.empty-evidence {
  margin: 0.32rem 0 0;
  color: var(--text-muted);
  font-size: var(--fs-xs, 0.76rem);
  line-height: 1.55;
}
.evidence-list li small {
  display: block;
  margin-top: 0.3rem;
  color: var(--text-soft, var(--text-muted));
  overflow-wrap: anywhere;
}
.more-note {
  margin: 0.65rem 0 0;
  color: var(--text-soft, var(--text-muted));
  font-size: var(--fs-xs, 0.76rem);
}
.professional-meta {
  margin-top: 0.8rem;
  padding-top: 0.8rem;
  border-top: 1px solid color-mix(in oklab, var(--text) 9%, transparent);
}
.score-flow {
  display: flex;
  flex-wrap: wrap;
  gap: 0.45rem;
}
.score-flow span,
.score-flow strong {
  padding: 0.28rem 0.5rem;
  border-radius: 6px;
  background: color-mix(in oklab, var(--accent) 8%, transparent);
  font-size: var(--fs-xs, 0.76rem);
}
.calculation-note {
  margin: 0.65rem 0 0;
  color: var(--text-muted);
  font-size: var(--fs-xs, 0.76rem);
  line-height: 1.55;
}
@media (max-width: 680px) {
  .evidence-item-head {
    align-items: flex-start;
  }
}
@media (max-width: 480px) {
  .evidence-panel {
    padding: 1rem;
  }
  .evidence-head {
    align-items: flex-start;
    flex-direction: column;
    gap: 0.65rem;
  }
}
</style>
