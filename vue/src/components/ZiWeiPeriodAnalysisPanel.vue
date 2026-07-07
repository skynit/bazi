<script setup lang="ts">
import { computed } from 'vue'
import type {
  ZiWeiDayunStageAnalysis,
  ZiWeiPeriodAnalysis,
  ZiWeiPeriodEvidence,
  ZiWeiPeriodHighlight,
  ZiWeiPeriodPalaceFocus,
} from '../api/ziwei'

interface Props {
  analysis: ZiWeiPeriodAnalysis | null
  title: string
  subtitle: string
}

const props = defineProps<Props>()

const stages = computed<ZiWeiDayunStageAnalysis[]>(() => list(props.analysis?.dayun_stages))
const focusPalaces = computed<ZiWeiPeriodPalaceFocus[]>(() => list(props.analysis?.focus_palaces))
const highlights = computed<ZiWeiPeriodHighlight[]>(() => list(props.analysis?.highlights))
const evidence = computed<ZiWeiPeriodEvidence[]>(() => list(props.analysis?.evidence))
const methods = computed<string[]>(() => list(props.analysis?.method))
const recommendations = computed<string[]>(() => list(props.analysis?.recommendations))
const risks = computed<string[]>(() => list(props.analysis?.risks))
const currentStage = computed(() => stages.value.find((stage) => stage.current))

function list<T>(items: T[] | null | undefined): T[] {
  return Array.isArray(items) ? items : []
}

function hasItems(items: unknown[] | null | undefined) {
  return list(items).length > 0
}

function scoreClass(score: number) {
  if (score >= 72) return 'is-strong'
  if (score >= 55) return 'is-mid'
  return 'is-low'
}

function shortLayer(layer: string) {
  switch (layer) {
    case 'dayun':
      return '十年'
    case 'liunian':
      return '年度'
    case 'liuyue':
      return '月度'
    case 'liuri':
      return '当日'
    default:
      return '周期'
  }
}

function starTags(item: ZiWeiPeriodPalaceFocus | ZiWeiDayunStageAnalysis) {
  return [
    ...list(item.main_stars).map((name) => ({ name, kind: 'main' })),
    ...list(item.aux_stars).map((name) => ({ name, kind: 'aux' })),
    ...list(item.four_hua).map((name) => ({ name, kind: 'hua' })),
  ]
}

function periodTags(item: ZiWeiPeriodPalaceFocus) {
  return list(item.period_stars).map((name) => ({ name, kind: 'period' }))
}

</script>

<template>
  <section v-if="analysis" class="zw-period-panel">
    <header class="zw-period-desk">
      <div class="zw-score-seal" :class="scoreClass(analysis.score)">
        <strong>{{ analysis.score }}</strong>
        <span>{{ analysis.tone }}</span>
      </div>
      <div class="zw-period-copy">
        <div class="zw-period-meta-line">
          <span>{{ shortLayer(analysis.layer) }}</span>
          <span>{{ analysis.time_label }}</span>
          <span v-if="analysis.gan_zhi">{{ analysis.gan_zhi }}</span>
          <span>{{ subtitle }}</span>
        </div>
        <h3>{{ title }}</h3>
        <p>{{ analysis.summary }}</p>
      </div>
    </header>

    <section class="zw-signal-strip" aria-label="周期提示">
      <article v-for="item in highlights" :key="`${item.label}-${item.value}`" class="zw-signal-item">
        <span>{{ item.label }}</span>
        <strong>{{ item.value }}</strong>
        <p>{{ item.note }}</p>
      </article>
      <article v-if="currentStage" class="zw-signal-item is-current">
        <span>当前大限</span>
        <strong>{{ currentStage.palace }} {{ currentStage.start_age }}-{{ currentStage.end_age }}岁</strong>
        <p>{{ currentStage.summary }}</p>
      </article>
    </section>

    <section v-if="stages.length" class="zw-stage-board">
      <div class="zw-section-head">
        <h4>大限轴</h4>
        <span v-if="currentStage">当前 {{ currentStage.palace }}</span>
        <span v-else>{{ stages.length }} 阶段</span>
      </div>
      <div class="zw-stage-rail">
        <article
          v-for="stage in stages"
          :key="`${stage.start_age}-${stage.end_age}-${stage.palace}`"
          class="zw-stage-row"
          :class="{ current: stage.current }"
        >
          <div class="zw-age-cell">
            <strong>{{ stage.start_age }}-{{ stage.end_age }}</strong>
            <span>岁</span>
          </div>
          <div class="zw-stage-main">
            <div class="zw-stage-title">
              <strong>{{ stage.palace }}</strong>
              <span>{{ stage.branch }} · {{ stage.tone }}</span>
            </div>
            <p>{{ stage.summary }}</p>
            <div class="zw-token-row">
              <span
                v-for="tag in starTags(stage)"
                :key="`${stage.palace}-${tag.kind}-${tag.name}`"
                class="zw-token"
                :class="`is-${tag.kind}`"
              >
                {{ tag.name }}
              </span>
            </div>
          </div>
          <div class="zw-row-score" :class="scoreClass(stage.score)">{{ stage.score }}</div>
        </article>
      </div>
    </section>

    <section class="zw-focus-board">
      <div class="zw-section-head">
        <h4>触发宫位</h4>
        <span>{{ focusPalaces.length }} 个重点</span>
      </div>

      <div v-if="focusPalaces.length" class="zw-focus-list">
        <article v-for="item in focusPalaces" :key="`${item.palace}-${item.branch}`" class="zw-focus-row">
          <div class="zw-focus-place">
            <strong>{{ item.palace }}</strong>
            <span>{{ item.branch }} · {{ item.level }}</span>
          </div>
          <div class="zw-focus-stars">
            <div class="zw-token-row">
              <span
                v-for="tag in [...starTags(item), ...periodTags(item)]"
                :key="`${item.palace}-${tag.kind}-${tag.name}`"
                class="zw-token"
                :class="`is-${tag.kind}`"
              >
                {{ tag.name }}
              </span>
              <span v-if="!hasItems(item.main_stars) && !hasItems(item.period_stars) && !hasItems(item.four_hua)" class="zw-token is-empty">无明显星曜</span>
            </div>
            <p>{{ item.reason }}</p>
          </div>
          <div class="zw-row-score" :class="scoreClass(item.score)">{{ item.score }}</div>
        </article>
      </div>

      <p v-else class="zw-empty-line">本层没有明显的流曜或四化集中触发。</p>
    </section>

    <div class="zw-reading-grid">
      <section class="zw-evidence-panel">
        <div class="zw-section-head">
          <h4>判断依据</h4>
          <span>{{ evidence.length }} 条</span>
        </div>
        <ul class="zw-evidence-list">
          <li v-for="item in evidence" :key="`${item.label}-${item.value}`">
            <span>{{ item.label }}</span>
            <strong>{{ item.value }}</strong>
            <p>{{ item.impact }}</p>
          </li>
        </ul>
      </section>

      <section v-if="recommendations.length || risks.length" class="zw-action-panel">
        <div class="zw-section-head">
          <h4>需要留意</h4>
          <span>{{ recommendations.length + risks.length }} 条</span>
        </div>
        <div v-if="recommendations.length" class="zw-note-block">
          <strong>提示</strong>
          <p v-for="item in recommendations" :key="item">{{ item }}</p>
        </div>
        <div v-if="risks.length" class="zw-note-block is-risk">
          <strong>风险</strong>
          <p v-for="item in risks" :key="item">{{ item }}</p>
        </div>
      </section>
    </div>

    <details class="zw-method-fold">
      <summary>
        <span>规则依据</span>
        <b>{{ analysis.rule_version }}</b>
      </summary>
      <div class="zw-method-body">
        <ol>
          <li v-for="item in methods" :key="item">{{ item }}</li>
        </ol>
        <p>{{ analysis.school }}</p>
      </div>
    </details>
  </section>
</template>

<style scoped>
.zw-period-panel {
  --period-ink: #17231f;
  --period-muted: color-mix(in oklab, var(--text-muted) 84%, transparent);
  --period-line: color-mix(in oklab, var(--line-subtle) 86%, transparent);
  --period-paper: color-mix(in oklab, var(--surface-0) 76%, transparent);
  --period-wash: color-mix(in oklab, #0f8b6d 7%, transparent);
  --period-jade: #0f8b6d;
  --period-cinnabar: #c2412d;
  --period-amber: #a16207;
  display: flex;
  flex-direction: column;
  gap: 0.85rem;
  margin-bottom: 1rem;
  color: var(--text);
}

:global(.dark) .zw-period-panel {
  --period-ink: #e5ece7;
  --period-paper: color-mix(in oklab, var(--surface-0) 70%, transparent);
  --period-wash: color-mix(in oklab, #16a37d 10%, transparent);
  --period-jade: #34d399;
  --period-cinnabar: #fb7185;
  --period-amber: #facc15;
}

.zw-period-desk {
  display: grid;
  grid-template-columns: 96px minmax(0, 1fr);
  gap: 0.9rem;
  align-items: stretch;
  padding: 0.85rem;
  border: 1px solid var(--period-line);
  border-radius: 8px;
  background:
    linear-gradient(90deg, var(--period-wash), transparent 42%),
    var(--period-paper);
}

.zw-score-seal {
  display: grid;
  place-content: center;
  min-height: 92px;
  border: 1px solid currentColor;
  border-radius: 8px;
  background: color-mix(in oklab, currentColor 7%, transparent);
  color: var(--period-jade);
  text-align: center;
}

.zw-score-seal.is-mid {
  color: var(--period-amber);
}

.zw-score-seal.is-low {
  color: var(--period-cinnabar);
}

.zw-score-seal strong {
  font-size: var(--fs-3xl);
  line-height: 1;
  font-weight: 800;
  font-family: Georgia, 'Times New Roman', serif;
}

.zw-score-seal span {
  margin-top: 0.2rem;
  font-size: var(--fs-2xs);
  font-weight: 700;
}

.zw-period-copy {
  min-width: 0;
}

.zw-period-meta-line {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
}

.zw-period-meta-line span {
  max-width: 100%;
  padding: 0.08rem 0.4rem;
  border: 1px solid var(--period-line);
  border-radius: 5px;
  background: color-mix(in oklab, var(--surface-1) 70%, transparent);
  color: var(--period-muted);
  font-size: var(--fs-2xs);
  overflow-wrap: anywhere;
}

.zw-period-copy h3,
.zw-section-head h4 {
  margin: 0;
  color: var(--period-ink);
  letter-spacing: 0;
}

.zw-period-copy h3 {
  margin-top: 0.42rem;
  font-size: var(--fs-body);
  font-weight: 800;
}

.zw-period-copy p {
  margin: 0.35rem 0 0;
  color: var(--text-muted);
  font-size: var(--fs-xs);
  line-height: 1.65;
}

.zw-signal-strip {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 0.55rem;
}

.zw-signal-item,
.zw-stage-board,
.zw-focus-board,
.zw-evidence-panel,
.zw-action-panel,
.zw-method-fold {
  border: 1px solid var(--period-line);
  border-radius: 8px;
  background: var(--period-paper);
}

.zw-signal-item {
  min-width: 0;
  padding: 0.6rem 0.65rem;
}

.zw-signal-item.is-current {
  border-color: color-mix(in oklab, var(--period-jade) 42%, var(--period-line));
  background: color-mix(in oklab, var(--period-jade) 7%, var(--period-paper));
}

.zw-signal-item span,
.zw-section-head span,
.zw-stage-title span,
.zw-focus-place span,
.zw-evidence-list span,
.zw-method-body p {
  display: block;
  color: var(--period-muted);
  font-size: var(--fs-2xs);
}

.zw-signal-item strong {
  display: block;
  margin-top: 0.18rem;
  color: var(--period-jade);
  font-size: var(--fs-xs);
  font-weight: 800;
  overflow-wrap: anywhere;
}

.zw-signal-item p {
  margin: 0.25rem 0 0;
  color: var(--text-muted);
  font-size: var(--fs-2xs);
  line-height: 1.45;
}

.zw-stage-board,
.zw-focus-board,
.zw-evidence-panel,
.zw-action-panel {
  padding: 0.75rem;
}

.zw-section-head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 0.75rem;
  margin-bottom: 0.55rem;
}

.zw-section-head h4 {
  font-size: var(--fs-xs);
  font-weight: 800;
}

.zw-stage-rail,
.zw-focus-list,
.zw-evidence-list,
.zw-reading-grid {
  display: grid;
  gap: 0.45rem;
}

.zw-stage-row,
.zw-focus-row {
  display: grid;
  align-items: center;
  gap: 0.65rem;
  min-width: 0;
  padding: 0.55rem 0.6rem;
  border: 1px solid var(--period-line);
  border-radius: 8px;
  background: color-mix(in oklab, var(--surface-1) 62%, transparent);
}

.zw-stage-row {
  grid-template-columns: 72px minmax(0, 1fr) 42px;
}

.zw-stage-row.current,
.zw-focus-row:hover {
  border-color: color-mix(in oklab, var(--period-jade) 40%, var(--period-line));
  background: color-mix(in oklab, var(--period-jade) 6%, var(--period-paper));
}

.zw-age-cell,
.zw-row-score {
  display: grid;
  place-content: center;
  min-height: 52px;
  border-radius: 7px;
  background: color-mix(in oklab, var(--surface-0) 74%, transparent);
  text-align: center;
}

.zw-age-cell strong,
.zw-row-score {
  color: var(--period-ink);
  font-size: var(--fs-xs);
  font-weight: 800;
  font-variant-numeric: tabular-nums;
}

.zw-age-cell span {
  color: var(--period-muted);
  font-size: var(--fs-2xs);
}

.zw-stage-main,
.zw-focus-stars {
  min-width: 0;
}

.zw-stage-title,
.zw-focus-place {
  min-width: 0;
}

.zw-stage-title strong,
.zw-focus-place strong {
  color: var(--period-ink);
  font-size: var(--fs-xs);
}

.zw-stage-main p,
.zw-focus-stars p,
.zw-empty-line,
.zw-note-block p,
.zw-evidence-list p {
  margin: 0.22rem 0 0;
  color: var(--text-muted);
  font-size: var(--fs-2xs);
  line-height: 1.5;
}

.zw-token-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0.24rem;
  margin-top: 0.42rem;
}

.zw-token {
  max-width: 100%;
  padding: 0.06rem 0.34rem;
  border: 1px solid var(--period-line);
  border-radius: 5px;
  background: color-mix(in oklab, var(--surface-0) 78%, transparent);
  color: var(--text-soft);
  font-size: var(--fs-2xs);
  font-weight: 700;
  overflow-wrap: anywhere;
}

.zw-token.is-main {
  color: var(--period-ink);
}

.zw-token.is-period,
.zw-token.is-hua {
  color: var(--period-jade);
  border-color: color-mix(in oklab, var(--period-jade) 25%, var(--period-line));
  background: color-mix(in oklab, var(--period-jade) 7%, transparent);
}

.zw-token.is-aux {
  color: var(--period-amber);
}

.zw-token.is-empty {
  color: var(--period-muted);
}

.zw-row-score {
  color: var(--period-jade);
}

.zw-row-score.is-mid {
  color: var(--period-amber);
}

.zw-row-score.is-low {
  color: var(--period-cinnabar);
}

.zw-focus-row {
  grid-template-columns: minmax(84px, 0.25fr) minmax(0, 1fr) 42px;
}

.zw-reading-grid {
  grid-template-columns: minmax(0, 1fr) minmax(260px, 0.72fr);
}

.zw-evidence-list {
  margin: 0;
  padding: 0;
}

.zw-evidence-list li {
  display: grid;
  grid-template-columns: minmax(72px, 0.28fr) minmax(90px, 0.34fr) minmax(0, 1fr);
  gap: 0.5rem;
  align-items: start;
  list-style: none;
  padding: 0.48rem 0;
  border-top: 1px solid var(--period-line);
}

.zw-evidence-list li:first-child {
  border-top: 0;
}

.zw-evidence-list strong {
  color: var(--period-jade);
  font-size: var(--fs-2xs);
  overflow-wrap: anywhere;
}

.zw-evidence-list p {
  margin: 0;
}

.zw-note-block + .zw-note-block {
  margin-top: 0.65rem;
  padding-top: 0.65rem;
  border-top: 1px solid var(--period-line);
}

.zw-note-block strong {
  display: block;
  color: var(--period-ink);
  font-size: var(--fs-2xs);
}

.zw-note-block.is-risk strong,
.zw-note-block.is-risk p {
  color: var(--period-cinnabar);
}

.zw-method-fold {
  padding: 0.55rem 0.65rem;
}

.zw-method-fold summary {
  display: flex;
  justify-content: space-between;
  gap: 0.75rem;
  cursor: pointer;
  color: var(--period-muted);
  font-size: var(--fs-2xs);
  list-style: none;
}

.zw-method-fold summary::-webkit-details-marker {
  display: none;
}

.zw-method-fold b {
  color: var(--period-jade);
  font-weight: 700;
}

.zw-method-body {
  margin-top: 0.55rem;
  padding-top: 0.55rem;
  border-top: 1px solid var(--period-line);
}

.zw-method-body ol {
  margin: 0;
  padding-left: 1rem;
  color: var(--text-muted);
  font-size: var(--fs-2xs);
  line-height: 1.55;
}

.zw-method-body p {
  margin: 0.45rem 0 0;
}

@media (max-width: 1024px) {
  .zw-reading-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 720px) {
  .zw-period-desk,
  .zw-stage-row,
  .zw-focus-row,
  .zw-evidence-list li {
    grid-template-columns: 1fr;
  }

  .zw-score-seal,
  .zw-age-cell,
  .zw-row-score {
    min-height: auto;
    justify-items: start;
    padding: 0.45rem 0.55rem;
  }
}
</style>
