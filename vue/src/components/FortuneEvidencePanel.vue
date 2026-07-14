<script setup lang="ts">
import { computed } from 'vue'
import type { FortuneScoreBreakdown, InterpretationLevel, ScoreEvidence } from '../api/fortune'

const props = withDefaults(
  defineProps<{
    level: InterpretationLevel
    completeness?: number
    supporting?: ScoreEvidence[]
    counter?: ScoreEvidence[]
    breakdown?: FortuneScoreBreakdown
    engineVersion?: string
    ruleVersion?: string
  }>(),
  {
    completeness: 0,
    supporting: () => [],
    counter: () => [],
    breakdown: undefined,
    engineVersion: '',
    ruleVersion: '',
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
        <span class="evidence-eyebrow">可解释评分</span>
        <h2>支持证据与反向证据</h2>
      </div>
      <span class="completeness">证据完整度 {{ completeness }}%</span>
    </header>

    <p class="evidence-note">证据完整度表示规则所需资料的覆盖程度，不是事件发生概率。</p>

    <div class="evidence-grid">
      <article class="evidence-column support">
        <div class="evidence-title">
          <strong>支持证据</strong><span>{{ supporting.length }} 条</span>
        </div>
        <ul v-if="visibleSupporting.length">
          <li v-for="item in visibleSupporting" :key="item.code">
            <div>
              <strong>{{ item.label }}</strong
              ><em>+{{ item.impact }}</em>
            </div>
            <p v-if="level !== 'basic'">{{ item.description }}</p>
            <small v-if="level === 'professional'"
              >{{ item.code }} · {{ item.stage }} · {{ item.source }}</small
            >
          </li>
        </ul>
        <p v-else class="empty-evidence">当前规则未识别到明确支持项。</p>
      </article>

      <article class="evidence-column counter">
        <div class="evidence-title">
          <strong>反向证据</strong><span>{{ counter.length }} 条</span>
        </div>
        <ul v-if="visibleCounter.length">
          <li v-for="item in visibleCounter" :key="item.code">
            <div>
              <strong>{{ item.label }}</strong
              ><em>{{ item.impact }}</em>
            </div>
            <p v-if="level !== 'basic'">{{ item.description }}</p>
            <small v-if="level === 'professional'"
              >{{ item.code }} · {{ item.stage }} · {{ item.source }}</small
            >
          </li>
        </ul>
        <p v-else class="empty-evidence">当前规则未识别到明确反向项，仍需结合现实条件判断。</p>
      </article>
    </div>

    <div v-if="level === 'professional'" class="professional-meta" data-testid="professional-meta">
      <div class="score-flow" v-if="breakdown">
        <span>中性起分 {{ breakdown.base_score }}</span>
        <span>关系分 {{ breakdown.relation_score }}</span>
        <span>细项分 {{ breakdown.detail_score }}</span>
        <strong>最终分 {{ breakdown.final_score }}</strong>
      </div>
      <dl>
        <div>
          <dt>评分流水线</dt>
          <dd>{{ breakdown?.pipeline_version || '—' }}</dd>
        </div>
        <div>
          <dt>引擎版本</dt>
          <dd>{{ engineVersion || '—' }}</dd>
        </div>
        <div>
          <dt>规则版本</dt>
          <dd>{{ ruleVersion || '—' }}</dd>
        </div>
      </dl>
    </div>
  </section>
</template>

<style scoped>
.evidence-panel {
  margin-top: 0.75rem;
  padding: 1rem;
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
  letter-spacing: 0.12em;
}
h2 {
  margin: 0.18rem 0 0;
  font-size: var(--fs-lg, 1.1rem);
}
.completeness {
  white-space: nowrap;
  padding: 0.26rem 0.55rem;
  border-radius: 999px;
  color: var(--accent);
  background: color-mix(in oklab, var(--accent) 10%, transparent);
  font-size: var(--fs-xs, 0.76rem);
}
.evidence-note {
  margin: 0.55rem 0 0.8rem;
  color: var(--text-muted);
  font-size: var(--fs-xs, 0.76rem);
}
.evidence-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.7rem;
}
.evidence-column {
  padding: 0.75rem;
  border: 1px solid color-mix(in oklab, var(--text) 9%, transparent);
  border-radius: 10px;
  background: color-mix(in oklab, var(--surface, #111827) 72%, transparent);
}
.evidence-column.support {
  border-color: color-mix(in oklab, #38b98a 24%, transparent);
}
.evidence-column.counter {
  border-color: color-mix(in oklab, #e67878 22%, transparent);
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
  color: #38b98a;
  font-style: normal;
  font-variant-numeric: tabular-nums;
}
.counter li em {
  color: #e67878;
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
dl {
  display: grid;
  gap: 0.4rem;
  margin: 0.75rem 0 0;
}
dl div {
  display: grid;
  grid-template-columns: 6rem minmax(0, 1fr);
  gap: 0.6rem;
}
dt {
  color: var(--text-muted);
}
dd {
  margin: 0;
  overflow-wrap: anywhere;
}
@media (max-width: 680px) {
  .evidence-grid {
    grid-template-columns: 1fr;
  }
  .evidence-head {
    flex-direction: column;
  }
}
</style>
