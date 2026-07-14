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

const router = useRouter()
const errMsg = ref('')
const loading = ref(false)
const confirming = ref(false)
const preview = ref<ChartPreviewResponse | null>(null)

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
  gender: 'MALE' as 'MALE' | 'FEMALE',
  birthPlace: '',
  timezone: browserTimezone,
  longitude: '' as number | '',
  useTrueSolarTime: false,
  timeUncertain: false,
})

const pillars = computed(() => {
  if (!preview.value) return []
  return [
    { label: '年柱', value: preview.value.year_pillar },
    { label: '月柱', value: preview.value.month_pillar },
    { label: '日柱', value: preview.value.day_pillar },
    { label: '时柱', value: preview.value.hour_pillar },
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
    form.value.gender = b.gender === 'FEMALE' || b.gender === 'female' ? 'FEMALE' : 'MALE'
    form.value.birthPlace = b.birthPlace ?? ''
    form.value.timezone = b.timezone || browserTimezone
    form.value.longitude = typeof b.longitude === 'number' ? b.longitude : ''
    form.value.useTrueSolarTime = Boolean(b.useTrueSolarTime)
    form.value.timeUncertain = Boolean(b.timeUncertain)
  } catch {
    localStorage.removeItem('bazi_last_birth')
  }
})

function requestPayload(): ChartCreateRequest {
  const payload: ChartCreateRequest = {
    birth_year: Number(form.value.year),
    birth_month: Number(form.value.month),
    birth_day: Number(form.value.day),
    birth_hour: Number(form.value.hour),
    birth_min: Number(form.value.minute),
    calendar_type: form.value.calendarType,
    lunar_leap_month: form.value.calendarType === 'LUNAR' && form.value.lunarLeapMonth,
    gender: form.value.gender,
    name: form.value.name.trim(),
    birth_place: form.value.birthPlace.trim(),
    timezone: form.value.timezone.trim() || browserTimezone,
    use_true_solar_time: form.value.useTrueSolarTime,
    time_uncertain: form.value.timeUncertain,
  }
  if (form.value.longitude !== '' && Number.isFinite(Number(form.value.longitude))) {
    payload.longitude = Number(form.value.longitude)
  }
  return payload
}

function validateLocally(): string {
  if (!form.value.year || !form.value.month || !form.value.day) return '请填写完整的出生日期'
  if (form.value.hour < 0 || form.value.hour > 23) return '小时应在 0–23 之间'
  if (form.value.minute < 0 || form.value.minute > 59) return '分钟应在 0–59 之间'
  if (!form.value.timezone.trim()) return '请填写有效的 IANA 时区，例如 Asia/Shanghai'
  if (form.value.useTrueSolarTime && form.value.longitude === '')
    return '启用真太阳时需要填写出生地经度'
  return ''
}

function apiError(err: any, fallback: string): string {
  return err.response?.data?.message || err.response?.data?.error || err.message || fallback
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
  } catch (err: any) {
    errMsg.value = apiError(err, '校验失败，请检查出生信息')
  } finally {
    loading.value = false
  }
}

function editInput() {
  preview.value = null
  errMsg.value = ''
}

async function confirmCreate() {
  confirming.value = true
  errMsg.value = ''
  try {
    const payload = requestPayload()
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
        gender: form.value.gender,
        birthPlace: form.value.birthPlace,
        timezone: form.value.timezone,
        longitude: form.value.longitude === '' ? undefined : Number(form.value.longitude),
        useTrueSolarTime: form.value.useTrueSolarTime,
        timeUncertain: form.value.timeUncertain,
        chartId: data.id,
      }),
    )
    router.push('/chart/new?_t=' + Date.now())
  } catch (err: any) {
    errMsg.value = apiError(err, '保存命盘失败，请稍后重试')
  } finally {
    confirming.value = false
  }
}
</script>

<template>
  <div class="birth-form">
    <div class="form-card glass-panel" :class="{ 'validation-card': preview }">
      <div class="card-ornament">
        <div class="ornament-ring"></div>
        <div class="ornament-symbol">☯</div>
      </div>

      <div v-if="!preview" class="card-inner">
        <div class="step-label">第一步 · 录入出生信息</div>
        <h2 class="form-title">输入出生信息</h2>

        <div class="field-group full-field">
          <label class="field-label">命盘名称（可选）</label>
          <input
            v-model="form.name"
            class="input-dark"
            maxlength="80"
            placeholder="例如：我的命盘"
          />
        </div>

        <div class="calendar-toggle">
          <button
            :class="form.calendarType === 'SOLAR' ? 'active' : ''"
            @click="form.calendarType = 'SOLAR'"
          >
            公历
          </button>
          <button
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
            <label class="field-label">年</label
            ><input v-model.number="form.year" class="input-dark" type="number" />
          </div>
          <div class="field-group">
            <label class="field-label">月</label
            ><input v-model.number="form.month" class="input-dark" min="1" max="12" type="number" />
          </div>
          <div class="field-group">
            <label class="field-label">日</label
            ><input v-model.number="form.day" class="input-dark" min="1" max="31" type="number" />
          </div>
        </div>

        <div class="time-grid">
          <div class="field-group">
            <label class="field-label">小时（0–23）</label
            ><input v-model.number="form.hour" class="input-dark" min="0" max="23" type="number" />
          </div>
          <div class="field-group">
            <label class="field-label">分钟（0–59）</label
            ><input
              v-model.number="form.minute"
              class="input-dark"
              min="0"
              max="59"
              type="number"
            />
          </div>
        </div>

        <div class="gender-toggle">
          <button
            :class="form.gender === 'MALE' ? 'btn-tech' : 'btn-ghost'"
            class="toggle-btn"
            @click="form.gender = 'MALE'"
          >
            <span>♂</span>男
          </button>
          <button
            :class="form.gender === 'FEMALE' ? 'btn-tech' : 'btn-ghost'"
            class="toggle-btn"
            @click="form.gender = 'FEMALE'"
          >
            <span>♀</span>女
          </button>
        </div>

        <div class="location-grid">
          <div class="field-group">
            <label class="field-label">出生地（可选）</label
            ><input v-model="form.birthPlace" class="input-dark" placeholder="例如：上海市" />
          </div>
          <div class="field-group">
            <label class="field-label">IANA 时区</label
            ><input v-model="form.timezone" class="input-dark" placeholder="Asia/Shanghai" />
          </div>
        </div>
        <div class="field-group full-field">
          <label class="field-label">出生地经度（真太阳时可选）</label>
          <input
            v-model.number="form.longitude"
            class="input-dark"
            min="-180"
            max="180"
            step="0.0001"
            type="number"
            placeholder="例如：121.4737"
          />
        </div>

        <div class="check-list">
          <label class="check-row"
            ><input v-model="form.useTrueSolarTime" type="checkbox" /><span
              >按经度与均时差换算真太阳时</span
            ></label
          >
          <label class="check-row"
            ><input v-model="form.timeUncertain" type="checkbox" /><span
              >出生时间不确定（结果中提示边界风险）</span
            ></label
          >
        </div>

        <p class="privacy-note">出生地点仅用于时区与真太阳时校正。请先核对转换结果，再保存命盘。</p>
        <p v-if="errMsg" class="error-msg">{{ errMsg }}</p>
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
            ><strong>{{ preview.birth_validation.calculation_date_time }}</strong>
          </div>
          <div class="validation-item">
            <span>对应农历</span><strong>{{ preview.birth_validation.lunar_date }}</strong>
          </div>
          <div class="validation-item">
            <span>当前节气</span
            ><strong
              >{{ preview.birth_validation.current_solar_term }} ·
              {{ preview.birth_validation.current_solar_term_started_at }}</strong
            >
          </div>
          <div class="validation-item">
            <span>时区与 UTC</span
            ><strong
              >{{ preview.birth_validation.timezone }} ·
              {{ preview.birth_validation.utc_date_time }}</strong
            >
          </div>
          <div v-if="preview.birth_validation.true_solar_time_applied" class="validation-item">
            <span>真太阳时校正</span
            ><strong
              >{{ preview.birth_validation.true_solar_adjustment_minutes >= 0 ? '+' : ''
              }}{{ preview.birth_validation.true_solar_adjustment_minutes }} 分钟</strong
            >
          </div>
        </div>

        <div class="pillar-title">用于计算的四柱</div>
        <div class="pillar-grid">
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
        <p v-if="errMsg" class="error-msg">{{ errMsg }}</p>
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
            :disabled="confirming"
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
.card-ornament {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 70px;
  position: relative;
  background: linear-gradient(180deg, var(--accent-dim), transparent);
  border-bottom: 1px solid var(--line-subtle);
}
.ornament-ring {
  position: absolute;
  width: 44px;
  height: 44px;
  border: 1px solid var(--line-focus);
  border-radius: 50%;
}
.ornament-symbol {
  font-size: var(--fs-4xl);
  color: var(--accent);
  text-shadow: 0 0 20px var(--accent-glow);
}
.card-inner {
  padding: 28px 36px 32px;
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
  grid-template-columns: 1fr 1fr;
  gap: 12px;
  margin-bottom: 16px;
}
.calendar-toggle,
.gender-toggle {
  display: flex;
  gap: 10px;
  margin-bottom: 16px;
}
.calendar-toggle button,
.toggle-btn {
  flex: 1;
  padding: 10px;
  border: 1px solid var(--line-subtle);
  border-radius: 8px;
  color: var(--text-muted);
  background: transparent;
}
.calendar-toggle button.active,
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
  background: color-mix(in oklab, var(--surface) 85%, transparent);
  display: flex;
  flex-direction: column;
  gap: 5px;
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
