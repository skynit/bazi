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

const visibleSupporting = computed(() =>
  props.level === 'basic' ? props.supporting.slice(0, 2) : props.supporting,
)
const visibleCounter = computed(() =>
  props.level === 'basic' ? props.counter.slice(0, 2) : props.counter,
)
</script>

<template>
  <section class="evidence-panel glass-card" aria-label="运势证据">
    <header class="evidence-head">
      <div>
        <span class="evidence-eyebrow">今日干支关系</span>
        <h2>生扶与冲克</h2>
      </div>
    </header>

    <p class="evidence-note">
      这里只展示今日干支与命盘之间的关系，不据此判断具体事件。
    </p>

    <div class="evidence-grid">
      <article class="evidence-column support">
        <div class="evidence-title">
          <strong>生扶关系</strong><span>{{ supporting.length }} 条</span>
        </div>
        <ul v-if="visibleSupporting.length">
          <li v-for="item in visibleSupporting" :key="item.code">
            <div>
              <strong>{{ item.label }}</strong
              ><em v-if="level === 'professional'">影响值 +{{ item.impact }}</em>
            </div>
            <p v-if="level !== 'basic'">{{ item.description }}</p>
            <small v-if="level === 'professional'">依据：{{ item.source }}</small>
          </li>
        </ul>
        <p v-else class="empty-evidence">今天没有记录到明显的生扶关系。</p>
      </article>

      <article class="evidence-column counter">
        <div class="evidence-title">
          <strong>冲克关系</strong><span>{{ counter.length }} 条</span>
        </div>
        <ul v-if="visibleCounter.length">
          <li v-for="item in visibleCounter" :key="item.code">
            <div>
              <strong>{{ item.label }}</strong
              ><em v-if="level === 'professional'">影响值 {{ item.impact }}</em>
            </div>
            <p v-if="level !== 'basic'">{{ item.description }}</p>
            <small v-if="level === 'professional'">依据：{{ item.source }}</small>
          </li>
        </ul>
        <p v-else class="empty-evidence">今天没有记录到明显的冲克关系。</p>
      </article>
    </div>

    <div v-if="level === 'professional'" class="professional-meta" data-testid="professional-meta">
      <div class="score-flow" v-if="breakdown">
        <span>起始值 {{ breakdown.base_score }}</span>
        <span>关系调整 {{ breakdown.relation_score }}</span>
        <strong>最终值 {{ breakdown.final_score }}</strong>
      </div>
      <p class="calculation-note">数值只用于比较同一命盘在不同日期的关系变化，不表示吉凶或事件发生概率。</p>
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
.evidence-eyebrow {
  color: var(--text-muted);
  font-size: var(--fs-2xs, 0.68rem);
  letter-spacing: 0;
}
h2 {
  margin: 0.18rem 0 0;
  font-size: var(--fs-lg, 1.1rem);
  letter-spacing: 0;
}
.evidence-note {
  max-width: 72ch;
  margin: 0.55rem 0 1rem;
  color: var(--text-muted);
  font-size: var(--fs-xs, 0.76rem);
  line-height: 1.6;
}
.evidence-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.75rem;
}
.evidence-column {
  min-width: 0;
  padding: 0.85rem;
  border: 1px solid var(--line-subtle);
  border-radius: 8px;
  background: var(--surface-0);
}
.evidence-column.counter {
  background: color-mix(in oklab, var(--crimson) 2%, var(--surface-0));
}
.evidence-title,
li > div {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.7rem;
}
.evidence-title span {
  color: var(--text-muted);
  font-size: var(--fs-xs, 0.76rem);
}
.evidence-title strong {
  display: inline-flex;
  align-items: center;
  gap: var(--space-sm);
}
.evidence-title strong::before {
  width: 0.45rem;
  height: 0.45rem;
  border-radius: 50%;
  background: var(--accent);
  content: '';
}
.counter .evidence-title strong::before {
  background: var(--crimson);
}
ul {
  display: grid;
  gap: 0.55rem;
  margin: 0.65rem 0 0;
  padding: 0;
  list-style: none;
}
li {
  padding-top: 0.55rem;
  border-top: 1px solid color-mix(in oklab, var(--text) 7%, transparent);
}
li strong {
  font-size: var(--fs-sm, 0.88rem);
}
li em {
  color: color-mix(in oklab, var(--accent) 48%, var(--text));
  font-style: normal;
  font-variant-numeric: tabular-nums;
}
.counter li em {
  color: color-mix(in oklab, var(--crimson) 58%, var(--text));
}
li p,
.empty-evidence {
  margin: 0.32rem 0 0;
  color: var(--text-muted);
  font-size: var(--fs-xs, 0.76rem);
  line-height: 1.55;
}
li small {
  display: block;
  margin-top: 0.3rem;
  color: var(--text-soft, var(--text-muted));
  overflow-wrap: anywhere;
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
  .evidence-grid {
    grid-template-columns: 1fr;
  }
  .evidence-column {
    padding: 0.85rem;
  }
  .evidence-column.counter {
    padding: 0.85rem;
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
