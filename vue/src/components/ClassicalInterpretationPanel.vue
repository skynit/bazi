<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import {
  BookOpenTextIcon,
  ChevronDownIcon,
  RefreshCwIcon,
  SparklesIcon,
  TriangleAlertIcon,
} from '@lucide/vue'
import {
  fetchBaziInterpretation,
  type BaziInterpretationResponse,
  type InterpretationFocus,
} from '../api/interpretation'
import { submitFeedback, type FeedbackRating } from '../api/feedback'
import { getApiErrorMessage } from '../api/client'

const props = defineProps<{
  chartId?: number
}>()

const focusOptions: Array<{ key: InterpretationFocus; label: string }> = [
  { key: 'overview', label: '综合' },
  { key: 'pattern', label: '格局' },
  { key: 'tiaohou', label: '调候' },
  { key: 'ten_gods', label: '十神' },
]

const focus = ref<InterpretationFocus>('overview')
const loading = ref(false)
const error = ref('')
const response = ref<BaziInterpretationResponse | null>(null)
const expandedCitations = ref<number[]>([])
const consentResearch = ref(false)
const feedbackState = ref<
  Record<string, { rating?: FeedbackRating; loading?: boolean; error?: string }>
>({})

const available = computed(() => !!props.chartId)

function toggleCitation(id: number) {
  if (expandedCitations.value.includes(id)) {
    expandedCitations.value = expandedCitations.value.filter((v) => v !== id)
  } else {
    expandedCitations.value = [...expandedCitations.value, id]
  }
}

async function loadInterpretation() {
  if (!props.chartId) {
    response.value = null
    error.value = ''
    return
  }
  loading.value = true
  error.value = ''
  try {
    response.value = await fetchBaziInterpretation(props.chartId, focus.value)
  } catch (reason: unknown) {
    response.value = null
    error.value = getApiErrorMessage(reason, '经典依据加载失败，请稍后重试。')
  } finally {
    loading.value = false
  }
}

const feedbackOptions: Array<{ rating: FeedbackRating; label: string }> = [
  { rating: 'accurate', label: '准' },
  { rating: 'inaccurate', label: '不准' },
  { rating: 'too_generic', label: '太泛' },
  { rating: 'confusing', label: '看不懂' },
  { rating: 'helpful', label: '有帮助' },
]

function sectionKey(title: string) {
  return `${focus.value}:${title}`
}

async function sendSectionFeedback(sectionTitle: string, rating: FeedbackRating) {
  if (!props.chartId) return
  const key = sectionKey(sectionTitle)
  feedbackState.value = {
    ...feedbackState.value,
    [key]: { ...feedbackState.value[key], loading: true, error: '' },
  }
  try {
    await submitFeedback({
      chart_id: props.chartId,
      target_type: 'interpretation_section',
      target_id: key,
      rating,
      tags: [focus.value, sectionTitle],
      consent_research: consentResearch.value,
      consent_training: false,
    })
    feedbackState.value = {
      ...feedbackState.value,
      [key]: { rating, loading: false, error: '' },
    }
  } catch (reason: unknown) {
    feedbackState.value = {
      ...feedbackState.value,
      [key]: {
        ...feedbackState.value[key],
        loading: false,
        error: getApiErrorMessage(reason, '反馈提交失败，请稍后重试。'),
      },
    }
  }
}

watch(() => props.chartId, loadInterpretation, { immediate: true })
watch(focus, loadInterpretation)

const statusText = computed(() => {
  if (!response.value) return ''
  if (response.value.status === 'fallback') {
    if (response.value.reason === 'disabled' || response.value.reason === 'not_configured') {
      return '经典依据暂未启用'
    }
    return '当前以规则解读为主'
  }
  return '已检索经典依据'
})
</script>

<template>
  <section class="ai-interpretation">
    <div class="ai-head">
      <div class="ai-title-wrap">
        <div class="ai-eyebrow">
          <BookOpenTextIcon class="ai-eyebrow-icon" />
          经典参考
        </div>
        <div class="ai-title-row">
          <h3 class="ai-title">八字经典参考</h3>
          <span v-if="statusText" class="ai-status">{{ statusText }}</span>
        </div>
      </div>
      <button class="ai-refresh" :disabled="loading || !available" @click="loadInterpretation">
        <RefreshCwIcon class="ai-btn-icon" :class="{ spinning: loading }" />
        重新生成
      </button>
    </div>

    <div class="ai-tabs" role="tablist" aria-label="经典依据解读维度">
      <button
        v-for="item in focusOptions"
        :key="item.key"
        class="ai-tab"
        :class="{ active: focus === item.key }"
        @click="focus = item.key"
      >
        {{ item.label }}
      </button>
    </div>

    <div v-if="!available" class="ai-empty">
      <TriangleAlertIcon class="ai-empty-icon" />
      <div>
        <p class="ai-empty-title">先创建或打开一个命盘</p>
        <p class="ai-empty-text">这里会显示与命盘结构相关的典籍条目。</p>
      </div>
    </div>

    <div v-else-if="loading && !response" class="ai-loading">
      <SparklesIcon class="ai-loading-icon" />
      <span>正在检索经典依据</span>
    </div>

    <div v-else-if="error" class="ai-error">
      <TriangleAlertIcon class="ai-error-icon" />
      <div>
        <p class="ai-error-title">{{ error }}</p>
        <p class="ai-error-text">不影响命盘主内容，稍后可重试。</p>
      </div>
    </div>

    <div v-else-if="response" class="ai-body">
      <p class="ai-summary">{{ response.summary }}</p>

      <label class="ai-research-consent">
        <input v-model="consentResearch" type="checkbox" />
        <span>允许将本次段落反馈用于内部解读准确性研究</span>
      </label>

      <div class="ai-sections">
        <article v-for="section in response.sections" :key="section.title" class="ai-section">
          <div class="ai-section-head">
            <h4 class="ai-section-title">{{ section.title }}</h4>
            <span v-if="section.citations?.length" class="ai-section-chip">
              {{ section.citations.length }} 引用
            </span>
          </div>
          <div class="ai-section-content">{{ section.content }}</div>
          <div class="ai-feedback" aria-label="段落反馈">
            <span class="ai-feedback-label">反馈</span>
            <button
              v-for="option in feedbackOptions"
              :key="option.rating"
              class="ai-feedback-btn"
              :class="{
                active: feedbackState[sectionKey(section.title)]?.rating === option.rating,
              }"
              :disabled="feedbackState[sectionKey(section.title)]?.loading"
              @click="sendSectionFeedback(section.title, option.rating)"
            >
              {{ option.label }}
            </button>
            <span v-if="feedbackState[sectionKey(section.title)]?.loading" class="ai-feedback-note">
              提交中
            </span>
            <span
              v-else-if="feedbackState[sectionKey(section.title)]?.rating"
              class="ai-feedback-note"
            >
              已记录
            </span>
            <span
              v-else-if="feedbackState[sectionKey(section.title)]?.error"
              class="ai-feedback-error"
            >
              {{ feedbackState[sectionKey(section.title)]?.error }}
            </span>
          </div>
        </article>
      </div>

      <div v-if="response.citations.length" class="ai-citations">
        <button
          class="ai-citations-toggle"
          @click="
            expandedCitations = expandedCitations.length ? [] : response.citations.map((c) => c.id)
          "
        >
          <ChevronDownIcon class="ai-btn-icon" :class="{ open: expandedCitations.length }" />
          引用依据
        </button>
        <div v-if="expandedCitations.length" class="ai-citation-list">
          <article v-for="citation in response.citations" :key="citation.id" class="ai-citation">
            <button class="ai-citation-head" @click="toggleCitation(citation.id)">
              <span class="ai-citation-book">{{ citation.book }}</span>
              <span class="ai-citation-meta">
                {{ citation.chapter }} · {{ citation.page || citation.locator || '定位缺失' }}
              </span>
            </button>
            <div v-if="expandedCitations.includes(citation.id)" class="ai-citation-body">
              <p class="ai-citation-path">{{ citation.author }} · {{ citation.edition }}</p>
              <p class="ai-citation-quote">{{ citation.quote }}</p>
            </div>
          </article>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.ai-interpretation {
  display: flex;
  flex-direction: column;
  gap: 0.9rem;
  margin-top: 1rem;
  padding: 1rem;
  border-radius: 0.75rem;
  background: var(--glass);
  border: 1px solid var(--glass-border);
  box-shadow: var(--shadow-md);
}

.ai-head {
  display: flex;
  gap: 0.75rem;
  justify-content: space-between;
  align-items: flex-start;
  flex-wrap: wrap;
}

.ai-title-wrap {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.ai-eyebrow {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  font-size: var(--fs-xs);
  color: var(--text-dim);
  letter-spacing: 0;
}

.ai-eyebrow-icon,
.ai-btn-icon,
.ai-empty-icon,
.ai-loading-icon,
.ai-error-icon {
  width: 0.9rem;
  height: 0.9rem;
  flex: 0 0 auto;
}

.ai-title-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.ai-title {
  margin: 0;
  font-size: var(--fs-lg);
  line-height: 1.35;
  font-weight: 700;
  color: var(--text);
}

.ai-status {
  font-size: var(--fs-xs);
  color: var(--accent);
  padding: 0.18rem 0.5rem;
  border-radius: 999px;
  background: var(--accent-dim);
}

.ai-refresh {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  border: 1px solid var(--glass-border);
  background: var(--surface-1);
  color: var(--text);
  border-radius: 0.625rem;
  padding: 0.48rem 0.7rem;
  font-size: var(--fs-xs);
  cursor: pointer;
}

.ai-refresh:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.ai-tabs {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem;
}

.ai-tab {
  border: 1px solid var(--glass-border);
  background: transparent;
  color: var(--text-dim);
  border-radius: 999px;
  padding: 0.4rem 0.75rem;
  font-size: var(--fs-xs);
  cursor: pointer;
}

.ai-tab.active {
  color: var(--accent);
  background: var(--accent-dim);
  border-color: var(--accent-dim);
}

.ai-empty,
.ai-loading,
.ai-error {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  min-height: 4rem;
  color: var(--text-dim);
}

.ai-empty-title,
.ai-error-title {
  margin: 0;
  font-weight: 600;
  color: var(--text);
}

.ai-empty-text,
.ai-error-text {
  margin: 0.2rem 0 0;
  font-size: var(--fs-sm);
  color: var(--text-dim);
}

.ai-summary {
  margin: 0;
  color: var(--text);
  font-size: var(--fs-sm);
  line-height: 1.7;
}

.ai-sections {
  display: grid;
  gap: 0.75rem;
}

.ai-section {
  border: 1px solid var(--glass-border);
  border-radius: 0.625rem;
  padding: 0.8rem;
  background: var(--surface-0);
}

.ai-section-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 0.6rem;
}

.ai-section-title {
  margin: 0;
  font-size: var(--fs-sm);
  font-weight: 700;
  color: var(--text);
}

.ai-section-chip {
  font-size: var(--fs-xs);
  color: var(--text-dim);
}

.ai-section-content {
  margin: 0.45rem 0 0;
  font-size: var(--fs-sm);
  line-height: 1.68;
  color: var(--text-muted);
  white-space: pre-line;
}

.ai-feedback {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.35rem;
  margin-top: 0.65rem;
  padding-top: 0.55rem;
  border-top: 1px solid var(--glass-border);
}

.ai-research-consent {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  color: var(--text-dim);
  font-size: var(--fs-xs);
  cursor: pointer;
}

.ai-research-consent input {
  width: 1rem;
  height: 1rem;
  accent-color: var(--accent);
}

.ai-feedback-label,
.ai-feedback-note,
.ai-feedback-error {
  font-size: var(--fs-xs);
  color: var(--text-dim);
}

.ai-feedback-btn {
  border: 1px solid var(--glass-border);
  background: transparent;
  color: var(--text-muted);
  border-radius: 999px;
  padding: 0.28rem 0.55rem;
  font-size: var(--fs-xs);
  line-height: 1.2;
  cursor: pointer;
}

.ai-feedback-btn.active {
  color: var(--accent);
  border-color: var(--accent);
  background: var(--accent-dim);
}

.ai-feedback-btn:disabled {
  opacity: 0.55;
  cursor: wait;
}

.ai-feedback-error {
  color: var(--danger, #c2410c);
}

.ai-citations {
  display: flex;
  flex-direction: column;
  gap: 0.45rem;
}

.ai-citations-toggle,
.ai-citation-head {
  width: 100%;
  display: inline-flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
  border: 1px solid var(--glass-border);
  background: transparent;
  color: var(--text);
  border-radius: 0.625rem;
  padding: 0.55rem 0.7rem;
  cursor: pointer;
}

.ai-citation-list {
  display: grid;
  gap: 0.4rem;
}

.ai-citation {
  border: 1px solid var(--glass-border);
  border-radius: 0.625rem;
  overflow: hidden;
}

.ai-citation-head {
  border: 0;
  border-radius: 0;
  text-align: left;
}

.ai-citation-book {
  font-size: var(--fs-sm);
  font-weight: 700;
}

.ai-citation-meta,
.ai-citation-score,
.ai-citation-path {
  font-size: var(--fs-xs);
  color: var(--text-dim);
}

.ai-citation-body {
  padding: 0.7rem;
  border-top: 1px solid var(--glass-border);
  background: var(--surface-1);
}

.ai-citation-quote {
  margin: 0.35rem 0 0;
  font-size: var(--fs-sm);
  line-height: 1.65;
  color: var(--text-muted);
}

.ai-citation-hash {
  margin: 0.35rem 0 0;
  font-size: var(--fs-xs);
  color: var(--text-dim);
  overflow-wrap: anywhere;
}

.spinning {
  animation: spin 1s linear infinite;
}

.open {
  transform: rotate(180deg);
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
