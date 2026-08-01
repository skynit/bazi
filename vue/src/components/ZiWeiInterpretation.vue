<script setup lang="ts">
import { computed, ref } from 'vue'

interface SectionData {
  title: string
  content: string
  tags: string[]
}

interface ReadingEvidence {
  type: string
  label: string
  value: string
  basis: string
}

interface SanfangContext {
  opposite: string
  trine1: string
  trine2: string
  opposite_stars: string[]
  trine1_stars: string[]
  trine2_stars: string[]
  notes: string[]
}

interface PatternDetail {
  name: string
  palace: string
  stars: string[]
  basis: string
  structure_status: string
  validation_status: string
}

interface PalaceReading {
  palaceName: string
  palaceFocus?: string
  summary?: string
  keyPoints?: string[]
  evidence?: ReadingEvidence[]
  sanfangContext?: SanfangContext | null
  patternDetails?: PatternDetail[]
  reviewNotes?: string[]
  limitations?: string[]
  evidenceBasis?: string
  validationStatus?: string
  mainStarAnalysis: SectionData
  auxStarInfluence: SectionData
  sihuaInfluence: SectionData
  sanFangSiZheng: SectionData
  patternAnnotations: SectionData
}

const props = defineProps<{
  palaceReading: PalaceReading
}>()

const expanded = ref(true)

const groupedEvidence = computed(() => {
  const groups: Record<string, ReadingEvidence[]> = {}
  for (const item of props.palaceReading.evidence || []) {
    const key = item.type || 'other'
    if (!groups[key]) groups[key] = []
    groups[key].push(item)
  }
  return groups
})

const evidenceOrder = [
  'palace',
  'body_palace',
  'main_star',
  'borrowed_star',
  'four_hua',
  'soft_star',
  'tough_star',
  'aux_star',
  'adjective_star',
  'twelve_shen',
  'sanfang',
]

const orderedEvidence = computed(() => {
  const known = evidenceOrder
    .filter((type) => groupedEvidence.value[type]?.length)
    .flatMap((type) => groupedEvidence.value[type])
  const extra = Object.entries(groupedEvidence.value)
    .filter(([type]) => !evidenceOrder.includes(type))
    .flatMap(([, items]) => items)
  return [...known, ...extra]
})

const legacySections = computed(() =>
  [
    props.palaceReading.mainStarAnalysis,
    props.palaceReading.auxStarInfluence,
    props.palaceReading.sihuaInfluence,
    props.palaceReading.sanFangSiZheng,
    props.palaceReading.patternAnnotations,
  ].filter((section) => section?.content),
)

function toggle() {
  expanded.value = !expanded.value
}

function evidenceClass(type: string) {
  return `evidence-${type.replaceAll('_', '-')}`
}

function readableBasis(value: string): string {
  if (!value) return '按本命星曜与宫位关系整理'
  if (/[_=/]|rule|profile|hash|not_|\.ts|\.go|github/i.test(value)) {
    return '按本命星曜、宫位与三方四正关系整理'
  }
  return value
}
</script>

<template>
  <article class="interpretation-panel" :class="{ collapsed: !expanded }">
    <button class="panel-header" type="button" @click="toggle">
      <span class="header-kicker">宫位说明</span>
      <span class="header-title">{{ palaceReading.palaceName }}</span>
      <span v-if="palaceReading.palaceFocus" class="focus-pill">{{
        palaceReading.palaceFocus
      }}</span>
      <span class="toggle-icon">{{ expanded ? '收起' : '展开' }}</span>
    </button>

    <transition name="expand">
      <div v-if="expanded" class="panel-body">
        <section class="overview-band">
          <p class="overview-text">
            {{ palaceReading.summary || palaceReading.mainStarAnalysis.content }}
          </p>
          <div v-if="palaceReading.keyPoints?.length" class="key-point-grid">
            <div v-for="point in palaceReading.keyPoints" :key="point" class="key-point">
              {{ point }}
            </div>
          </div>
        </section>

        <section v-if="orderedEvidence.length" class="reading-block">
          <div class="block-title-row">
            <h4 class="block-title">计算依据</h4>
            <span class="block-count">{{ orderedEvidence.length }} 项</span>
          </div>
          <div class="evidence-grid">
            <div
              v-for="item in orderedEvidence"
              :key="item.type + item.label + item.value"
              class="evidence-card"
              :class="evidenceClass(item.type)"
            >
              <div class="evidence-head">
                <span class="evidence-label">{{ item.label }}</span>
                <strong class="evidence-value">{{ item.value }}</strong>
              </div>
              <p class="evidence-basis">{{ readableBasis(item.basis) }}</p>
            </div>
          </div>
        </section>

        <section v-if="palaceReading.sanfangContext" class="reading-block">
          <h4 class="block-title">三方四正</h4>
          <div class="sanfang-layout">
            <div class="sanfang-node opposite">
              <span class="node-label">对宫</span>
              <strong>{{ palaceReading.sanfangContext.opposite }}</strong>
              <p>{{ palaceReading.sanfangContext.opposite_stars?.join('、') || '无主辅星' }}</p>
            </div>
            <div class="sanfang-node">
              <span class="node-label">三合</span>
              <strong>{{ palaceReading.sanfangContext.trine1 }}</strong>
              <p>{{ palaceReading.sanfangContext.trine1_stars?.join('、') || '无主辅星' }}</p>
            </div>
            <div class="sanfang-node">
              <span class="node-label">三合</span>
              <strong>{{ palaceReading.sanfangContext.trine2 }}</strong>
              <p>{{ palaceReading.sanfangContext.trine2_stars?.join('、') || '无主辅星' }}</p>
            </div>
          </div>
          <ul v-if="palaceReading.sanfangContext.notes?.length" class="note-list">
            <li v-for="note in palaceReading.sanfangContext.notes" :key="note">{{ note }}</li>
          </ul>
        </section>

        <section v-if="palaceReading.patternDetails?.length" class="reading-block">
          <h4 class="block-title">格局依据</h4>
          <div class="pattern-list">
            <div
              v-for="pattern in palaceReading.patternDetails"
              :key="pattern.name + pattern.basis"
              class="pattern-card"
            >
              <div class="pattern-head">
                <strong>{{ pattern.name }}</strong>
                <span>结构线索</span>
              </div>
              <p>{{ readableBasis(pattern.basis) }}</p>
              <div v-if="pattern.stars?.length" class="mini-tags">
                <span v-for="star in pattern.stars" :key="star">{{ star }}</span>
              </div>
            </div>
          </div>
        </section>

        <p class="reading-boundary-note">宫位解读用于理解星曜与宫位结构，不直接判断具体事件。</p>

        <section
          v-if="!orderedEvidence.length && legacySections.length"
          class="reading-block legacy-block"
        >
          <h4 class="block-title">基础解读</h4>
          <div class="legacy-section" v-for="section in legacySections" :key="section.title">
            <strong>{{ section.title }}</strong>
            <p>{{ section.content }}</p>
            <div v-if="section.tags?.length" class="mini-tags">
              <span v-for="tag in section.tags" :key="tag">{{ tag }}</span>
            </div>
          </div>
        </section>
      </div>
    </transition>
  </article>
</template>

<style scoped>
@reference "tailwindcss";

.interpretation-panel {
  @apply overflow-hidden;
  border: 1px solid var(--line-strong);
  border-radius: 8px;
  background:
    linear-gradient(135deg, color-mix(in oklab, var(--accent) 8%, transparent), transparent 38%),
    color-mix(in oklab, var(--surface-1) 90%, transparent);
  box-shadow: var(--shadow-sm);
}

.panel-header {
  @apply w-full border-0 cursor-pointer text-left;
  display: grid;
  grid-template-columns: auto 1fr auto auto auto;
  gap: 0.75rem;
  align-items: center;
  padding: 0.875rem 1rem;
  background: color-mix(in oklab, var(--surface-2) 72%, transparent);
  border-bottom: 1px solid var(--line-subtle);
  transition: background 0.18s ease;
}

.panel-header:hover {
  background: color-mix(in oklab, var(--surface-3) 78%, transparent);
}

.panel-header:focus-visible {
  outline: 2px solid var(--line-focus);
  outline-offset: -2px;
}

.header-kicker {
  @apply text-xs font-semibold;
  color: var(--text-muted);
}

.header-title {
  @apply text-base font-bold;
  font-family: var(--font-serif), serif;
  letter-spacing: 0.04em;
  color: var(--accent);
}

.validation-pill {
  @apply text-xs font-semibold;
  color: #0f766e;
  background: rgba(15, 118, 110, 0.1);
  border: 1px solid rgba(15, 118, 110, 0.18);
  padding: 0.15rem 0.45rem;
  border-radius: 999px;
}

.focus-pill {
  @apply text-xs font-semibold;
  color: var(--text);
  background: color-mix(in oklab, var(--accent) 8%, transparent);
  border: 1px solid var(--line-subtle);
  padding: 0.15rem 0.45rem;
  border-radius: 999px;
  overflow-wrap: anywhere;
}

.toggle-icon {
  @apply text-xs font-semibold;
  color: var(--text-muted);
}

.panel-body {
  @apply flex flex-col gap-4 p-4;
}

.overview-band {
  @apply flex flex-col gap-3;
}

.overview-text {
  @apply m-0 text-sm leading-relaxed;
  color: var(--text);
}

.key-point-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 0.5rem;
}

.key-point {
  @apply text-xs leading-relaxed;
  padding: 0.55rem 0.65rem;
  color: var(--text);
  background: color-mix(in oklab, var(--accent) 5%, transparent);
  border: 1px solid var(--line-subtle);
  border-radius: 8px;
}

.reading-block {
  @apply flex flex-col gap-3;
  padding-top: 0.75rem;
  border-top: 1px solid var(--line-subtle);
}

.reading-boundary-note {
  margin: 0;
  color: var(--text-muted);
  font-size: var(--fs-xs);
  line-height: 1.6;
}

.block-title-row {
  @apply flex items-center justify-between gap-2;
}

.block-title {
  @apply m-0 text-sm font-bold;
  color: var(--text);
}

.block-count {
  @apply text-xs;
  color: var(--text-muted);
}

.evidence-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(190px, 1fr));
  gap: 0.5rem;
}

.evidence-card {
  @apply flex flex-col gap-2;
  min-width: 0;
  padding: 0.65rem;
  border: 1px solid var(--line-subtle);
  border-radius: 8px;
  background: color-mix(in oklab, var(--surface-0) 72%, transparent);
}

.evidence-head {
  @apply flex items-start justify-between gap-2;
}

.evidence-label {
  @apply text-xs font-semibold;
  flex: 0 0 auto;
  color: var(--text-muted);
}

.evidence-value {
  @apply text-sm text-right;
  color: var(--accent);
  overflow-wrap: anywhere;
}

.evidence-basis {
  @apply m-0 text-xs leading-relaxed;
  color: var(--text-soft);
}

.evidence-tough-star,
.evidence-four-hua {
  border-color: rgba(220, 38, 38, 0.18);
  background: rgba(220, 38, 38, 0.04);
}

.evidence-soft-star,
.evidence-body-palace {
  border-color: rgba(37, 99, 235, 0.18);
  background: rgba(37, 99, 235, 0.04);
}

.evidence-main-star {
  border-color: rgba(161, 98, 7, 0.2);
  background: rgba(161, 98, 7, 0.05);
}

.sanfang-layout {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.5rem;
}

.sanfang-node {
  @apply flex flex-col gap-1;
  min-width: 0;
  padding: 0.65rem;
  border: 1px solid var(--line-subtle);
  border-radius: 8px;
  background: color-mix(in oklab, var(--accent) 4%, transparent);
}

.sanfang-node.opposite {
  background: rgba(220, 38, 38, 0.04);
}

.node-label {
  @apply text-xs font-semibold;
  color: var(--text-muted);
}

.sanfang-node strong {
  @apply text-sm;
  color: var(--text);
}

.sanfang-node p {
  @apply m-0 text-xs leading-relaxed;
  color: var(--text-soft);
  overflow-wrap: anywhere;
}

.pattern-list {
  @apply flex flex-col gap-2;
}

.pattern-card {
  padding: 0.65rem;
  border: 1px solid var(--line-subtle);
  border-radius: 8px;
  background: color-mix(in oklab, var(--accent) 5%, transparent);
}

.pattern-head {
  @apply flex items-center justify-between gap-2;
}

.pattern-head strong {
  @apply text-sm;
  color: var(--accent);
}

.pattern-head span {
  @apply text-xs font-semibold;
  color: var(--text-muted);
}

.pattern-card p {
  @apply m-0 mt-2 text-xs leading-relaxed;
  color: var(--text);
}

.advice-risk-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(230px, 1fr));
  gap: 0.75rem;
}

.advice-block,
.risk-block {
  border: 1px solid var(--line-subtle);
  border-radius: 8px;
  padding: 0.75rem;
}

.advice-block {
  background: rgba(15, 118, 110, 0.04);
}

.risk-block {
  background: rgba(220, 38, 38, 0.04);
}

.note-list {
  @apply m-0 pl-4 text-xs leading-relaxed;
  color: var(--text);
}

.note-list li + li {
  margin-top: 0.4rem;
}

.legacy-block {
  opacity: 0.9;
}

.legacy-section {
  @apply flex flex-col gap-1;
}

.legacy-section strong {
  @apply text-xs;
  color: var(--accent);
}

.legacy-section p {
  @apply m-0 text-xs leading-relaxed;
  color: var(--text-soft);
}

.mini-tags {
  @apply flex flex-wrap gap-1.5 mt-2;
}

.mini-tags span {
  @apply text-xs font-semibold;
  color: var(--accent);
  background: var(--accent-dim);
  border: 1px solid var(--line-subtle);
  border-radius: 999px;
  padding: 0.1rem 0.45rem;
}

.expand-enter-active,
.expand-leave-active {
  transition:
    opacity 0.18s ease,
    transform 0.18s ease;
}

.expand-enter-from,
.expand-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}

:global(.dark) {
  .validation-pill {
    color: #86efac;
    background: rgba(74, 222, 128, 0.1);
    border-color: rgba(74, 222, 128, 0.18);
  }

  .evidence-soft-star,
  .evidence-body-palace {
    border-color: rgba(96, 165, 250, 0.18);
    background: rgba(96, 165, 250, 0.06);
  }

  .evidence-tough-star,
  .evidence-four-hua,
  .sanfang-node.opposite,
  .risk-block {
    background: rgba(251, 113, 133, 0.06);
    border-color: rgba(251, 113, 133, 0.16);
  }
}

@media (max-width: 640px) {
  .panel-header {
    grid-template-columns: 1fr auto;
  }

  .header-kicker,
  .focus-pill,
  .validation-pill {
    grid-column: 1 / -1;
  }

  .sanfang-layout {
    grid-template-columns: 1fr;
  }
}
</style>
