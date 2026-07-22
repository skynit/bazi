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
        <span class="evidence-eyebrow">结构关系指数</span>
        <h2>正向权重与负向权重</h2>
      </div>
      <span class="completeness">证据完整度 {{ completeness }}%</span>
    </header>

    <p class="evidence-note">
      权重为本地启发式配置，未经 Gold 验证；完整度只表示计算输入是否齐备，不是事件发生概率。
    </p>

    <div class="evidence-grid">
      <article class="evidence-column support">
        <div class="evidence-title">
          <strong>正向权重</strong><span>{{ supporting.length }} 条</span>
        </div>
        <ul v-if="visibleSupporting.length">
          <li v-for="item in visibleSupporting" :key="item.code">
            <div>
              <strong>{{ item.label }}</strong
              ><em>+{{ item.impact }}</em>
            </div>
            <p v-if="level !== 'basic'">{{ item.description }}</p>
            <small v-if="level === 'professional'"
              >{{ item.code }} · {{ item.validation_status }} · {{ item.interpretation_status }} ·
              {{ item.source }}</small
            >
          </li>
        </ul>
        <p v-else class="empty-evidence">当前结构规则没有正向权重项。</p>
      </article>

      <article class="evidence-column counter">
        <div class="evidence-title">
          <strong>负向权重</strong><span>{{ counter.length }} 条</span>
        </div>
        <ul v-if="visibleCounter.length">
          <li v-for="item in visibleCounter" :key="item.code">
            <div>
              <strong>{{ item.label }}</strong
              ><em>{{ item.impact }}</em>
            </div>
            <p v-if="level !== 'basic'">{{ item.description }}</p>
            <small v-if="level === 'professional'"
              >{{ item.code }} · {{ item.validation_status }} · {{ item.interpretation_status }} ·
              {{ item.source }}</small
            >
          </li>
        </ul>
        <p v-else class="empty-evidence">当前结构规则没有负向权重项。</p>
      </article>
    </div>

    <div v-if="level === 'professional'" class="professional-meta" data-testid="professional-meta">
      <div class="score-flow" v-if="breakdown">
        <span>中性起分 {{ breakdown.base_score }}</span>
        <span>关系分 {{ breakdown.relation_score }}</span>
        <strong>结构指数 {{ breakdown.final_score }}</strong>
      </div>
      <dl>
        <div>
          <dt>评分流水线</dt>
          <dd>{{ breakdown?.pipeline_version || '—' }}</dd>
        </div>
        <div>
          <dt>指数类型</dt>
          <dd>{{ breakdown?.score_kind || '—' }}</dd>
        </div>
        <div>
          <dt>验证状态</dt>
          <dd>
            {{ breakdown?.evidence_basis || '—' }} · {{ breakdown?.validation_status || '—' }} ·
            {{ breakdown?.interpretation_status || '—' }}
          </dd>
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
.completeness {
  white-space: nowrap;
  padding: 0.26rem 0.55rem;
  border: 1px solid var(--line-focus);
  border-radius: var(--radius-sm);
  color: color-mix(in oklab, var(--accent) 48%, var(--text));
  background: var(--accent-dim);
  font-size: var(--fs-xs, 0.76rem);
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
  border-top: 1px solid var(--line-subtle);
  border-bottom: 1px solid var(--line-subtle);
}
.evidence-column {
  min-width: 0;
  padding: 1rem 0.75rem 1rem 0;
  background: transparent;
}
.evidence-column.counter {
  padding-right: 0;
  padding-left: 0.75rem;
  border-left: 1px solid var(--line-subtle);
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
  .evidence-column {
    padding-right: 0;
  }
  .evidence-column.counter {
    padding-top: 1rem;
    padding-left: 0;
    border-top: 1px solid var(--line-subtle);
    border-left: 0;
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
