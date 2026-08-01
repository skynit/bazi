<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  createChart,
  previewChart,
  type ChartCreateRequest,
  type ChartPreviewResponse,
} from '../api/chart'
import { Button } from '@/components/ui/button'
import { getApiErrorMessage } from '../api/client'

const router = useRouter()
const errMsg = ref('')
const loading = ref(false)
const confirming = ref(false)
const preview = ref<ChartPreviewResponse | null>(null)
const selectedCandidateId = ref('')

const browserTimezone = Intl.DateTimeFormat().resolvedOptions().timeZone || 'Asia/Shanghai'
const now = new Date()
const form = ref({
  name: '',
  calendarType: 'SOLAR' as 'SOLAR' | 'LUNAR',
  lunarLeapMonth: false,
  year: now.getFullYear(),
  month: now.getMonth() + 1,
  day: now.getDate(),
  hour: 8,
  minute: 0,
  second: 0,
  gender: 'MALE' as 'MALE' | 'FEMALE',
  ziHourPolicy: 'late_zi_next_day' as 'late_zi_next_day' | 'late_zi_same_day',
  birthPlace: '',
  timezone: browserTimezone,
  birthUTCOffsetSeconds: '' as number | '',
  longitude: '' as number | '',
  useTrueSolarTime: false,
  timeUncertain: false,
  uncertaintySeconds: 900,
})

const activeCandidate = computed(() => {
  const candidates = preview.value?.candidate_charts ?? []
  return candidates.find((candidate) => candidate.candidate_id === selectedCandidateId.value)
})

const fieldLabels: Record<string, string> = {
  year_pillar: '年柱',
  month_pillar: '月柱',
  day_pillar: '日柱',
  hour_pillar: '时柱',
}

const stableFieldLabels = computed(() =>
  (preview.value?.stable_fields ?? []).map((field) => fieldLabels[field] ?? field),
)
const unstableFieldLabels = computed(() =>
  (preview.value?.unstable_fields ?? []).map((field) => fieldLabels[field] ?? field),
)

function boundaryLabel(type: string, name?: string) {
  const labels: Record<string, string> = {
    hour_branch: '时辰交界',
    zi_hour_day_boundary: '子初换日边界',
    civil_day: '公历日期交界',
    solar_term: name ? `节气交界（${name}）` : '节气交界',
  }
  return labels[type] ?? name ?? '排盘边界'
}

const crossedBoundaryLabels = computed(() =>
  (preview.value?.uncertainty?.crossed_boundaries ?? []).map((boundary) =>
    boundaryLabel(boundary.type, boundary.name),
  ),
)

const pillars = computed(() => {
  if (!preview.value) return []
  if (preview.value.requires_candidate_selection && !activeCandidate.value) return []
  const chart = activeCandidate.value ?? preview.value
  return [
    { label: '年柱', value: chart.year_pillar },
    { label: '月柱', value: chart.month_pillar },
    { label: '日柱', value: chart.day_pillar },
    { label: '时柱', value: chart.hour_pillar },
  ]
})

onMounted(() => {
  const saved = localStorage.getItem('bazi_last_birth')
  if (!saved) return
  try {
    const b = JSON.parse(saved)
    form.value.name = b.name ?? ''
    form.value.calendarType = b.calendarType === 'LUNAR' ? 'LUNAR' : 'SOLAR'
    form.value.lunarLeapMonth = Boolean(b.lunarLeapMonth)
    form.value.year = Number(b.year) || form.value.year
    form.value.month = Number(b.month) || form.value.month
    form.value.day = Number(b.day) || form.value.day
    form.value.hour = Number.isFinite(Number(b.hour)) ? Number(b.hour) : 8
    form.value.minute = Number.isFinite(Number(b.minute)) ? Number(b.minute) : 0
    form.value.second = Number.isFinite(Number(b.second)) ? Number(b.second) : 0
    form.value.gender = b.gender === 'FEMALE' || b.gender === 'female' ? 'FEMALE' : 'MALE'
    form.value.ziHourPolicy =
      b.ziHourPolicy === 'late_zi_same_day' ? 'late_zi_same_day' : 'late_zi_next_day'
    form.value.birthPlace = b.birthPlace ?? ''
    form.value.timezone = b.timezone || browserTimezone
    form.value.birthUTCOffsetSeconds =
      typeof b.birthUTCOffsetSeconds === 'number' ? b.birthUTCOffsetSeconds : ''
    form.value.longitude = typeof b.longitude === 'number' ? b.longitude : ''
    form.value.useTrueSolarTime = Boolean(b.useTrueSolarTime)
    form.value.timeUncertain = Boolean(b.timeUncertain)
    form.value.uncertaintySeconds = Number.isFinite(Number(b.uncertaintySeconds))
      ? Number(b.uncertaintySeconds)
      : 900
  } catch {
    localStorage.removeItem('bazi_last_birth')
  }
})

function requestPayload(includeSelection = false): ChartCreateRequest {
  const payload: ChartCreateRequest = {
    birth_year: Number(form.value.year),
    birth_month: Number(form.value.month),
    birth_day: Number(form.value.day),
    birth_hour: Number(form.value.hour),
    birth_min: Number(form.value.minute),
    birth_sec: Number(form.value.second),
    calendar_type: form.value.calendarType,
    lunar_leap_month: form.value.calendarType === 'LUNAR' && form.value.lunarLeapMonth,
    gender: form.value.gender,
    zi_hour_policy: form.value.ziHourPolicy,
    name: form.value.name.trim(),
    birth_place: form.value.birthPlace.trim(),
    timezone: form.value.timezone.trim() || browserTimezone,
    use_true_solar_time: form.value.useTrueSolarTime,
    time_uncertain: form.value.timeUncertain,
    uncertainty_seconds: form.value.timeUncertain ? Number(form.value.uncertaintySeconds) : 0,
  }
  if (form.value.longitude !== '' && Number.isFinite(Number(form.value.longitude))) {
    payload.longitude = Number(form.value.longitude)
  }
  if (
    form.value.birthUTCOffsetSeconds !== '' &&
    Number.isFinite(Number(form.value.birthUTCOffsetSeconds))
  ) {
    payload.birth_utc_offset_seconds = Number(form.value.birthUTCOffsetSeconds)
  }
  if (includeSelection && selectedCandidateId.value) {
    payload.candidate_id = selectedCandidateId.value
  }
  return payload
}

function validateLocally(): string {
  if (!form.value.year || !form.value.month || !form.value.day) return '请填写完整的出生日期'
  if (form.value.hour < 0 || form.value.hour > 23) return '小时应在 0–23 之间'
  if (form.value.minute < 0 || form.value.minute > 59) return '分钟应在 0–59 之间'
  if (form.value.second < 0 || form.value.second > 59) return '秒应在 0–59 之间'
  if (
    form.value.timeUncertain &&
    (form.value.uncertaintySeconds < 1 || form.value.uncertaintySeconds > 86400)
  )
    return '时间误差应在 1–86400 秒之间'
  if (!form.value.timezone.trim()) return '请填写有效的 IANA 时区，例如 Asia/Shanghai'
  if (form.value.useTrueSolarTime && form.value.longitude === '')
    return '启用真太阳时需要填写出生地经度'
  return ''
}

async function handlePreview() {
  const validationError = validateLocally()
  if (validationError) {
    errMsg.value = validationError
    return
  }
  loading.value = true
  errMsg.value = ''
  try {
    preview.value = await previewChart(requestPayload())
    const candidates = preview.value.candidate_charts ?? []
    selectedCandidateId.value = candidates.length === 1 ? candidates[0].candidate_id : ''
  } catch (reason: unknown) {
    errMsg.value = getApiErrorMessage(reason, '校验失败，请检查出生信息。')
  } finally {
    loading.value = false
  }
}

function editInput() {
  preview.value = null
  selectedCandidateId.value = ''
  errMsg.value = ''
}

async function confirmCreate() {
  if (preview.value?.requires_candidate_selection && !selectedCandidateId.value) {
    errMsg.value = '时间区间跨越四柱边界，请先选择一个候选命盘'
    return
  }
  confirming.value = true
  errMsg.value = ''
  try {
    const payload = requestPayload(true)
    const data = await createChart(payload)
    sessionStorage.setItem('lastChart', JSON.stringify(data))
    localStorage.setItem(
      'bazi_last_birth',
      JSON.stringify({
        name: form.value.name,
        calendarType: form.value.calendarType,
        lunarLeapMonth: form.value.lunarLeapMonth,
        year: form.value.year,
        month: form.value.month,
        day: form.value.day,
        hour: form.value.hour,
        minute: form.value.minute,
        second: form.value.second,
        gender: form.value.gender,
        ziHourPolicy: form.value.ziHourPolicy,
        birthPlace: form.value.birthPlace,
        timezone: form.value.timezone,
        birthUTCOffsetSeconds:
          form.value.birthUTCOffsetSeconds === ''
            ? undefined
            : Number(form.value.birthUTCOffsetSeconds),
        longitude: form.value.longitude === '' ? undefined : Number(form.value.longitude),
        useTrueSolarTime: form.value.useTrueSolarTime,
        timeUncertain: form.value.timeUncertain,
        uncertaintySeconds: form.value.timeUncertain ? form.value.uncertaintySeconds : 0,
        chartId: data.id,
      }),
    )
    router.push('/chart/new?_t=' + Date.now())
  } catch (reason: unknown) {
    errMsg.value = getApiErrorMessage(reason, '保存命盘失败，请稍后重试。')
  } finally {
    confirming.value = false
  }
}
</script>

<template>
  <div class="birth-form">
    <div class="form-card glass-panel" :class="{ 'validation-card': preview }">
      <div v-if="!preview" class="card-inner">
        <div class="step-label">第一步 · 录入出生信息</div>
        <h2 class="form-title">输入出生信息</h2>

        <div class="field-group full-field">
          <label class="field-label" for="birth-name">命盘名称（可选）</label>
          <input
            id="birth-name"
            v-model="form.name"
            :aria-invalid="Boolean(errMsg)"
            aria-errormessage="birth-form-error"
            class="input-dark"
            maxlength="80"
            placeholder="例如：我的命盘"
          />
        </div>

        <div class="calendar-toggle">
          <button
            type="button"
            :aria-pressed="form.calendarType === 'SOLAR'"
            :class="form.calendarType === 'SOLAR' ? 'active' : ''"
            @click="form.calendarType = 'SOLAR'"
          >
            公历
          </button>
          <button
            type="button"
            :aria-pressed="form.calendarType === 'LUNAR'"
            :class="form.calendarType === 'LUNAR' ? 'active' : ''"
            @click="form.calendarType = 'LUNAR'"
          >
            农历
          </button>
        </div>

        <label v-if="form.calendarType === 'LUNAR'" class="check-row">
          <input v-model="form.lunarLeapMonth" type="checkbox" />
          <span>该月为闰月</span>
        </label>

        <div class="date-grid">
          <div class="field-group">
            <label class="field-label" for="birth-year">年</label
            ><input
              id="birth-year"
              v-model.number="form.year"
              :aria-invalid="Boolean(errMsg)"
              aria-errormessage="birth-form-error"
              class="input-dark"
              type="number"
            />
          </div>
          <div class="field-group">
            <label class="field-label" for="birth-month">月</label
            ><input
              id="birth-month"
              v-model.number="form.month"
              :aria-invalid="Boolean(errMsg)"
              aria-errormessage="birth-form-error"
              class="input-dark"
              min="1"
              max="12"
              type="number"
            />
          </div>
          <div class="field-group">
            <label class="field-label" for="birth-day">日</label
            ><input
              id="birth-day"
              v-model.number="form.day"
              :aria-invalid="Boolean(errMsg)"
              aria-errormessage="birth-form-error"
              class="input-dark"
              min="1"
              max="31"
              type="number"
            />
          </div>
        </div>

        <div class="time-grid">
          <div class="field-group">
            <label class="field-label" for="birth-hour">小时（0–23）</label
            ><input
              id="birth-hour"
              v-model.number="form.hour"
              :aria-invalid="Boolean(errMsg)"
              aria-errormessage="birth-form-error"
              class="input-dark"
              min="0"
              max="23"
              type="number"
            />
          </div>
          <div class="field-group">
            <label class="field-label" for="birth-minute">分钟（0–59）</label
            ><input
              id="birth-minute"
              v-model.number="form.minute"
              :aria-invalid="Boolean(errMsg)"
              aria-errormessage="birth-form-error"
              class="input-dark"
              min="0"
              max="59"
              type="number"
            />
          </div>
          <div class="field-group">
            <label class="field-label" for="birth-second">秒（0–59）</label
            ><input
              id="birth-second"
              v-model.number="form.second"
              :aria-invalid="Boolean(errMsg)"
              aria-errormessage="birth-form-error"
              class="input-dark"
              min="0"
              max="59"
              type="number"
            />
          </div>
        </div>

        <div class="gender-toggle">
          <button
            type="button"
            :aria-pressed="form.gender === 'MALE'"
            :class="form.gender === 'MALE' ? 'btn-tech' : 'btn-ghost'"
            class="toggle-btn"
            @click="form.gender = 'MALE'"
          >
            <span>♂</span>男
          </button>
          <button
            type="button"
            :aria-pressed="form.gender === 'FEMALE'"
            :class="form.gender === 'FEMALE' ? 'btn-tech' : 'btn-ghost'"
            class="toggle-btn"
            @click="form.gender = 'FEMALE'"
          >
            <span>♀</span>女
          </button>
        </div>

        <div class="field-group full-field">
          <label class="field-label" for="birth-place">出生地（可选）</label
          ><input
            id="birth-place"
            v-model="form.birthPlace"
            :aria-invalid="Boolean(errMsg)"
            aria-errormessage="birth-form-error"
            class="input-dark"
            placeholder="例如：上海市"
          />
        </div>

        <div class="check-list">
          <label class="check-row"
            ><input v-model="form.timeUncertain" type="checkbox" /><span
              >出生时间不确定（生成边界候选盘）</span
            ></label
          >
        </div>
        <div v-if="form.timeUncertain" class="field-group full-field">
          <label class="field-label" for="birth-uncertainty"
            >中心时刻前后误差（秒，最多 86400）</label
          >
          <input
            id="birth-uncertainty"
            v-model.number="form.uncertaintySeconds"
            :aria-invalid="Boolean(errMsg)"
            aria-errormessage="birth-form-error"
            class="input-dark"
            min="1"
            max="86400"
            step="1"
            type="number"
          />
        </div>

        <details class="advanced-settings">
          <summary>排盘口径与时间校正</summary>
          <p>默认设置适合大多数用户；仅在明确知道出生时区或需要真太阳时校正时调整。</p>
          <div class="field-group full-field">
            <span class="field-label">晚子时日柱口径</span>
            <div class="policy-toggle">
              <button
                type="button"
                :aria-pressed="form.ziHourPolicy === 'late_zi_next_day'"
                :class="form.ziHourPolicy === 'late_zi_next_day' ? 'active' : ''"
                @click="form.ziHourPolicy = 'late_zi_next_day'"
              >
                子初换日（23:00）
              </button>
              <button
                type="button"
                :aria-pressed="form.ziHourPolicy === 'late_zi_same_day'"
                :class="form.ziHourPolicy === 'late_zi_same_day' ? 'active' : ''"
                @click="form.ziHourPolicy = 'late_zi_same_day'"
              >
                午夜换日（00:00）
              </button>
            </div>
          </div>
          <div class="field-group full-field">
            <label class="field-label" for="birth-timezone">出生地时区</label>
            <input
              id="birth-timezone"
              v-model="form.timezone"
              :aria-invalid="Boolean(errMsg)"
              aria-errormessage="birth-form-error"
              class="input-dark"
              placeholder="Asia/Shanghai"
            />
          </div>
          <div class="field-group full-field">
            <label class="field-label" for="birth-utc-offset"
              >UTC 偏移秒（仅夏令时重复时刻需要）</label
            >
            <input
              id="birth-utc-offset"
              v-model.number="form.birthUTCOffsetSeconds"
              :aria-invalid="Boolean(errMsg)"
              aria-errormessage="birth-form-error"
              class="input-dark"
              min="-50400"
              max="50400"
              step="1"
              type="number"
              placeholder="例如：-14400"
            />
          </div>
          <label class="check-row"
            ><input v-model="form.useTrueSolarTime" type="checkbox" /><span
              >按经度与均时差换算真太阳时</span
            ></label
          >
          <div v-if="form.useTrueSolarTime" class="field-group full-field">
            <label class="field-label" for="birth-longitude">出生地经度（-180 到 180）</label>
            <input
              id="birth-longitude"
              v-model.number="form.longitude"
              :aria-invalid="Boolean(errMsg)"
              aria-errormessage="birth-form-error"
              class="input-dark"
              min="-180"
              max="180"
              step="0.0001"
              type="number"
              placeholder="例如：121.4737"
            />
          </div>
        </details>

        <p class="privacy-note">出生地点仅用于时区与真太阳时校正。请先核对转换结果，再保存命盘。</p>
        <p v-if="errMsg" id="birth-form-error" class="error-msg" role="alert">{{ errMsg }}</p>
        <Button
          :disabled="loading"
          class="w-full h-12 rounded-full text-base font-semibold bg-foreground text-background"
          @click="handlePreview"
        >
          {{ loading ? '正在校验…' : '校验命盘' }}
        </Button>
      </div>

      <div v-else class="card-inner validation-inner">
        <div class="step-label">第二步 · 命盘校验</div>
        <h2 class="form-title">请确认采用的出生时间</h2>
        <p class="validation-intro">
          以下仅展示历法转换与八字四柱，不包含紫微基础盘。确认无误后再保存。
        </p>

        <div class="validation-grid">
          <div class="validation-item">
            <span>原始输入</span
            ><strong
              >{{ preview.birth_validation.original_date_time }}（{{
                preview.birth_validation.input_calendar === 'LUNAR' ? '农历' : '公历'
              }}）</strong
            >
          </div>
          <div class="validation-item">
            <span>转换后公历</span
            ><strong>{{ preview.birth_validation.converted_solar_date_time }}</strong>
          </div>
          <div class="validation-item emphasis">
            <span>最终计算时间</span
            ><strong>{{
              activeCandidate?.birth_validation.calculation_date_time ??
              preview.birth_validation.calculation_date_time
            }}</strong>
          </div>
          <div class="validation-item">
            <span>对应农历</span><strong>{{ preview.birth_validation.lunar_date }}</strong>
          </div>
        </div>

        <details class="validation-details">
          <summary>查看历法与时间换算</summary>
          <div class="validation-grid">
            <div class="validation-item">
              <span>当前节气</span>
              <strong>
                {{ preview.birth_validation.current_solar_term }} ·
                {{ preview.birth_validation.current_solar_term_started_at }}
              </strong>
            </div>
            <div class="validation-item">
              <span>时区与 UTC</span>
              <strong>
                {{ preview.birth_validation.timezone }} ·
                {{ preview.birth_validation.utc_date_time }}
              </strong>
            </div>
            <div class="validation-item">
              <span>晚子时换日方式</span>
              <strong>{{
                preview.birth_validation.zi_hour_policy === 'late_zi_same_day'
                  ? '午夜换日（晚子时日柱算当天）'
                  : '子初换日（晚子时日柱算次日）'
              }}</strong>
            </div>
            <div v-if="preview.birth_validation.true_solar_time_applied" class="validation-item">
              <span>真太阳时校正</span>
              <strong>
                {{ preview.birth_validation.true_solar_adjustment_minutes >= 0 ? '+' : ''
                }}{{ preview.birth_validation.true_solar_adjustment_minutes }} 分钟
              </strong>
            </div>
            <div v-if="preview.birth_validation.local_time_ambiguous" class="validation-item">
              <span>重复时刻采用时区偏移</span>
              <strong>
                UTC {{ preview.birth_validation.timezone_offset_seconds / 3600 >= 0 ? '+' : ''
                }}{{ preview.birth_validation.timezone_offset_seconds / 3600 }}
              </strong>
            </div>
          </div>
        </details>

        <div v-if="preview.uncertainty" class="uncertainty-box">
          <strong>时间评估范围</strong>
          <p>
            输入范围：{{ preview.uncertainty.input_range_start }} 至
            {{ preview.uncertainty.input_range_end }}
          </p>
          <p v-if="crossedBoundaryLabels.length">跨越：{{ crossedBoundaryLabels.join('、') }}。</p>
          <p v-if="preview.candidate_charts.length > 1">
            稳定字段：{{ stableFieldLabels.join('、') || '无' }}；会变化：{{
              unstableFieldLabels.join('、') || '无'
            }}。
          </p>
          <p v-if="preview.candidate_charts.length > 1">
            请优先依据出生证明或家人记录选择；无法确认时可返回修改，缩小时间范围后再保存。
          </p>
        </div>

        <div v-if="preview.candidate_charts.length > 1" class="candidate-list">
          <div class="pillar-title">请选择要保存的候选命盘</div>
          <button
            v-for="(candidate, index) in preview.candidate_charts"
            :key="candidate.candidate_id"
            type="button"
            class="candidate-card"
            :class="{ selected: selectedCandidateId === candidate.candidate_id }"
            :aria-pressed="selectedCandidateId === candidate.candidate_id"
            @click="selectedCandidateId = candidate.candidate_id"
          >
            <span class="candidate-heading">候选 {{ index + 1 }}</span>
            <span
              >{{ candidate.calculation_range_start }} 至
              {{ candidate.calculation_range_end }}</span
            >
            <strong>
              {{ candidate.year_pillar.gan }}{{ candidate.year_pillar.zhi }} ·
              {{ candidate.month_pillar.gan }}{{ candidate.month_pillar.zhi }} ·
              {{ candidate.day_pillar.gan }}{{ candidate.day_pillar.zhi }} ·
              {{ candidate.hour_pillar.gan }}{{ candidate.hour_pillar.zhi }}
            </strong>
            <small>
              起运范围：{{ candidate.da_yun_start_at_min }} 至 {{ candidate.da_yun_start_at_max }}
            </small>
          </button>
        </div>

        <div v-if="pillars.length" class="pillar-title">用于计算的四柱</div>
        <div v-if="pillars.length" class="pillar-grid">
          <div v-for="pillar in pillars" :key="pillar.label" class="pillar-item">
            <span>{{ pillar.label }}</span>
            <strong>{{ pillar.value.gan }}{{ pillar.value.zhi }}</strong>
          </div>
        </div>

        <div v-if="preview.birth_validation.notices?.length" class="notice-box">
          <strong>校验提示</strong>
          <ul>
            <li v-for="notice in preview.birth_validation.notices" :key="notice">{{ notice }}</li>
          </ul>
        </div>
        <p v-if="errMsg" class="error-msg" role="alert">{{ errMsg }}</p>
        <div class="action-grid">
          <Button
            variant="outline"
            class="h-12 rounded-full"
            :disabled="confirming"
            @click="editInput"
            >返回修改</Button
          >
          <Button
            class="h-12 rounded-full font-semibold bg-foreground text-background"
            :disabled="confirming || (preview.requires_candidate_selection && !selectedCandidateId)"
            @click="confirmCreate"
            >{{ confirming ? '正在保存…' : '确认并保存命盘' }}</Button
          >
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.birth-form {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 60vh;
  padding: 2rem;
}
.form-card {
  width: 100%;
  max-width: 620px;
  overflow: hidden;
}
.validation-card {
  max-width: 760px;
}
.card-inner {
  padding: 32px 36px 32px;
}
.step-label {
  text-align: center;
  color: var(--accent);
  font-size: var(--fs-2xs);
  letter-spacing: 2px;
  margin-bottom: 8px;
}
.form-title {
  text-align: center;
  font-family: var(--font-serif), serif;
  font-size: var(--fs-2xl);
  font-weight: 700;
  color: var(--text);
  margin: 0 0 24px;
  letter-spacing: 2px;
}
.field-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.full-field {
  margin-bottom: 16px;
}
.field-label {
  font-size: var(--fs-2xs);
  font-weight: 600;
  letter-spacing: 1px;
  color: var(--text-muted);
}
.date-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
  margin-bottom: 16px;
}
.time-grid,
.location-grid {
  display: grid;
  gap: 12px;
  margin-bottom: 16px;
}
.time-grid {
  grid-template-columns: repeat(3, 1fr);
}
.location-grid {
  grid-template-columns: 1fr 1fr;
}
.calendar-toggle,
.policy-toggle {
  display: flex;
  gap: 4px;
  margin-bottom: 16px;
  padding: 4px;
  border: 1px solid var(--line-subtle);
  border-radius: 10px;
  background: var(--surface-2);
}
.gender-toggle {
  display: flex;
  gap: 10px;
  margin-bottom: 16px;
}
.calendar-toggle button,
.policy-toggle button {
  flex: 1;
  padding: 9px 10px;
  border: 1px solid transparent;
  border-radius: 7px;
  color: var(--text-muted);
  background: transparent;
  cursor: pointer;
  font-size: var(--fs-sm);
  transition:
    color 0.18s,
    background 0.18s,
    border-color 0.18s,
    box-shadow 0.18s;
}
.calendar-toggle button:hover,
.policy-toggle button:hover {
  color: var(--text);
}
.calendar-toggle button:focus-visible,
.policy-toggle button:focus-visible,
.toggle-btn:focus-visible {
  outline: 2px solid var(--line-focus);
  outline-offset: 1px;
}
.calendar-toggle button.active,
.policy-toggle button.active {
  background: var(--surface-0);
  border-color: var(--line-strong);
  color: var(--text);
  font-weight: 700;
  box-shadow: var(--shadow-xs);
}
.toggle-btn {
  flex: 1;
  padding: 10px;
  border: 1px solid var(--line-subtle);
  border-radius: 8px;
  color: var(--text-muted);
  background: transparent;
  cursor: pointer;
  transition:
    color 0.18s,
    background 0.18s,
    border-color 0.18s;
}
.toggle-btn.btn-tech {
  background: var(--accent);
  border-color: var(--accent);
  color: var(--bg);
  font-weight: 700;
}
.check-list {
  display: grid;
  gap: 10px;
  margin: 18px 0;
}
.advanced-settings {
  margin: 18px 0;
  padding: 0 14px;
  border: 1px solid var(--line-subtle);
  border-radius: 10px;
  background: var(--surface-1);
}
.advanced-settings summary {
  padding: 14px 0;
  color: var(--text-muted);
  font-size: var(--fs-sm);
  font-weight: 700;
  cursor: pointer;
}
.advanced-settings > p {
  margin: 0 0 14px;
  color: var(--text-soft);
  font-size: var(--fs-xs);
  line-height: 1.6;
}
.advanced-settings[open] {
  padding-bottom: 14px;
}
.check-row {
  display: flex;
  align-items: center;
  gap: 9px;
  color: var(--text-muted);
  font-size: var(--fs-sm);
  cursor: pointer;
}
.check-row input {
  accent-color: var(--accent);
}
.privacy-note,
.validation-intro {
  color: var(--text-soft);
  font-size: var(--fs-xs);
  line-height: 1.7;
  text-align: center;
  margin: 16px 0;
}
.error-msg {
  color: var(--crimson);
  font-size: var(--fs-xs);
  text-align: center;
  margin: 0 0 14px;
}
.validation-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
}
.validation-item {
  padding: 13px 14px;
  border: 1px solid var(--line-subtle);
  border-radius: 10px;
  background: var(--surface-1);
  display: flex;
  flex-direction: column;
  gap: 5px;
}
.validation-details {
  margin-top: 12px;
  border-top: 1px solid var(--line-subtle);
}
.validation-details summary {
  padding: 12px 0;
  color: var(--text-muted);
  font-size: var(--fs-xs);
  font-weight: 600;
  cursor: pointer;
}
.validation-details .validation-grid {
  padding-bottom: 4px;
}
.validation-item span,
.pillar-item span {
  color: var(--text-soft);
  font-size: var(--fs-2xs);
}
.validation-item strong {
  color: var(--text);
  font-size: var(--fs-sm);
  line-height: 1.5;
}
.validation-item.emphasis {
  border-color: var(--line-focus);
  background: var(--accent-dim);
}
.pillar-title {
  margin: 22px 0 10px;
  color: var(--text-muted);
  font-size: var(--fs-xs);
  letter-spacing: 1px;
}
.pillar-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 10px;
}
.pillar-item {
  text-align: center;
  padding: 14px 8px;
  border: 1px solid var(--line-focus);
  border-radius: 10px;
  background: color-mix(in oklab, var(--accent-dim) 55%, var(--surface-1));
}
.pillar-item strong {
  display: block;
  margin-top: 6px;
  color: var(--accent);
  font-family: var(--font-serif), serif;
  font-size: var(--fs-xl);
}
.notice-box {
  margin-top: 18px;
  padding: 14px 16px;
  border-radius: 10px;
  background: color-mix(in oklab, var(--crimson) 8%, transparent);
  border: 1px solid color-mix(in oklab, var(--crimson) 30%, transparent);
  color: var(--text-muted);
  font-size: var(--fs-xs);
  line-height: 1.7;
}
.notice-box ul {
  margin: 6px 0 0;
  padding-left: 18px;
}
.uncertainty-box {
  margin-top: 18px;
  padding: 14px 16px;
  border: 1px solid var(--line-focus);
  border-radius: 10px;
  background: var(--accent-dim);
  color: var(--text-muted);
  font-size: var(--fs-xs);
  line-height: 1.6;
}
.uncertainty-box p {
  margin: 5px 0 0;
}
.candidate-list {
  display: grid;
  gap: 10px;
}
.candidate-card {
  display: grid;
  gap: 5px;
  width: 100%;
  padding: 14px 16px;
  border: 1px solid var(--line-subtle);
  border-radius: 10px;
  background: var(--surface-1);
  color: var(--text-muted);
  text-align: left;
  cursor: pointer;
  transition:
    border-color 0.18s,
    background 0.18s,
    box-shadow 0.18s;
}
.candidate-card:hover {
  border-color: var(--line-strong);
  box-shadow: var(--shadow-xs);
}
.candidate-card:focus-visible {
  outline: 2px solid var(--line-focus);
  outline-offset: 2px;
}
.candidate-card.selected {
  border-color: var(--accent);
  background: var(--accent-dim);
  box-shadow: 0 0 0 1px var(--accent);
}
.candidate-card strong {
  color: var(--accent);
  font-family: var(--font-serif), serif;
  font-size: var(--fs-lg);
}
.candidate-heading {
  color: var(--text);
  font-weight: 700;
}
.action-grid {
  display: grid;
  grid-template-columns: 1fr 1.4fr;
  gap: 12px;
  margin-top: 22px;
}
@media (max-width: 640px) {
  .birth-form {
    padding: 1rem;
  }
  .card-inner {
    padding: 24px 18px;
  }
  .location-grid,
  .time-grid,
  .validation-grid {
    grid-template-columns: 1fr;
  }
  .pillar-grid {
    grid-template-columns: repeat(2, 1fr);
  }
  .action-grid {
    grid-template-columns: 1fr;
  }
  .form-title {
    font-size: var(--fs-xl);
  }
}
</style>
