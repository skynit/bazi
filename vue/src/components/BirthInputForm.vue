<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import client from '../api/client'

const router = useRouter()
const errMsg = ref('')
const loading = ref(false)
const form = ref({
  year: new Date().getFullYear(),
  month: 1,
  day: 1,
  shichen: 4,
  gender: 'male' as string,
})
onMounted(() => {
  const saved = localStorage.getItem('bazi_last_birth')
  if (saved) {
    try {
      const b = JSON.parse(saved)
      form.value.year = b.year
      form.value.month = b.month
      form.value.day = b.day
      form.value.shichen = b.shichen ?? 4
      form.value.gender = b.gender ?? 'male'
    } catch {}
  }
})
const shichenOptions = [
  { label: '子时 (23-01)', value: 0 },
  { label: '丑时 (01-03)', value: 1 },
  { label: '寅时 (03-05)', value: 2 },
  { label: '卯时 (05-07)', value: 3 },
  { label: '辰时 (07-09)', value: 4 },
  { label: '巳时 (09-11)', value: 5 },
  { label: '午时 (11-13)', value: 6 },
  { label: '未时 (13-15)', value: 7 },
  { label: '申时 (15-17)', value: 8 },
  { label: '酉时 (17-19)', value: 9 },
  { label: '戌时 (19-21)', value: 10 },
  { label: '亥时 (21-23)', value: 11 },
]
async function handleSubmit() {
  if (!form.value.year || !form.value.month || !form.value.day) {
    errMsg.value = '请填写完整的出生日期'
    return
  }
  loading.value = true
  errMsg.value = ''
  try {
    const res = await client.post('/chart', {
      birth_year: form.value.year,
      birth_month: form.value.month,
      birth_day: form.value.day,
      birth_hour: form.value.shichen * 2,
      birth_min: 0,
      calendar_type: 'SOLAR',
      gender: form.value.gender.toUpperCase(),
      name: '',
    })
    sessionStorage.setItem('lastChart', JSON.stringify(res.data))
    const birthInfo = {
      year: form.value.year,
      month: form.value.month,
      day: form.value.day,
      shichen: form.value.shichen,
      gender: form.value.gender,
      chartId: res.data.id,
    }
    localStorage.setItem('bazi_last_birth', JSON.stringify(birthInfo))
    router.push('/chart/new?_t=' + Date.now())
  } catch (err: any) {
    errMsg.value = err.response?.data?.error || err.message || '排盘失败，请稍后重试'
  } finally {
    loading.value = false
  }
}
</script>
<template>
  <div class="birth-form">
    <div class="form-card glass-panel">
      <div class="card-ornament">
        <div class="ornament-ring"></div>
        <div class="ornament-symbol">☯</div>
      </div>
      <div class="card-inner">
        <h2 class="form-title">输入出生信息</h2>

        <!-- Gender toggle -->
        <div class="gender-toggle">
          <button
            :class="form.gender === 'male' ? 'btn-gold' : 'btn-ghost'"
            class="toggle-btn"
            @click="form.gender = 'male'"
          >
            <span class="toggle-icon">♂</span>
            男
          </button>
          <button
            :class="form.gender === 'female' ? 'btn-gold' : 'btn-ghost'"
            class="toggle-btn"
            @click="form.gender = 'female'"
          >
            <span class="toggle-icon">♀</span>
            女
          </button>
        </div>

        <!-- Date inputs -->
        <div class="date-grid">
          <div class="field-group">
            <label class="field-label">年</label>
            <input v-model.number="form.year" class="input-dark" placeholder="1990" type="number" />
          </div>
          <div class="field-group">
            <label class="field-label">月</label>
            <input v-model.number="form.month" class="input-dark" placeholder="6" type="number" />
          </div>
          <div class="field-group">
            <label class="field-label">日</label>
            <input v-model.number="form.day" class="input-dark" placeholder="15" type="number" />
          </div>
        </div>

        <!-- Time select -->
        <div class="field-group">
          <label class="field-label">时辰</label>
          <select v-model.number="form.shichen" class="input-dark">
            <option v-for="s in shichenOptions" :key="s.value" :value="s.value">
              {{ s.label }}
            </option>
          </select>
        </div>

        <!-- Error -->
        <p v-if="errMsg" class="error-msg">{{ errMsg }}</p>

        <!-- Submit -->
        <button class="btn-gold btn-submit" :disabled="loading" @click="handleSubmit">
          <span v-if="loading" class="loading-spinner"></span>
          <span v-else>✦ {{ loading ? '排盘中...' : '开始排盘' }}</span>
        </button>
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
  max-width: 440px;
  overflow: hidden;
}

/* Ornament */
.card-ornament {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 70px;
  position: relative;
  background: linear-gradient(180deg, rgba(212, 168, 75, 0.06), transparent);
  border-bottom: 1px solid rgba(212, 168, 75, 0.06);
}

.ornament-ring {
  position: absolute;
  width: 44px;
  height: 44px;
  border: 1px solid rgba(212, 168, 75, 0.12);
  border-radius: 50%;
  animation: ring-pulse 3s ease-in-out infinite;
}

@keyframes ring-pulse {
  0%,
  100% {
    transform: scale(1);
    opacity: 0.15;
  }
  50% {
    transform: scale(1.1);
    opacity: 0.3;
  }
}

.ornament-symbol {
  font-size: 1.75rem;
  color: var(--gold);
  text-shadow: 0 0 20px rgba(212, 168, 75, 0.4);
  animation: symbol-glow 3s ease-in-out infinite;
}

@keyframes symbol-glow {
  0%,
  100% {
    text-shadow: 0 0 20px rgba(212, 168, 75, 0.4);
  }
  50% {
    text-shadow: 0 0 35px rgba(212, 168, 75, 0.6);
  }
}

/* Inner */
.card-inner {
  padding: 28px 36px 32px;
}

.form-title {
  text-align: center;
  font-family: var(--font-serif), serif;
  font-size: 1.3rem;
  font-weight: 700;
  color: var(--text);
  margin: 0 0 24px;
  letter-spacing: 3px;
}

/* Gender toggle */
.gender-toggle {
  display: flex;
  gap: 12px;
  margin-bottom: 20px;
}

.toggle-btn {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 10px;
  border-radius: 8px;
  font-size: 0.9rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.toggle-icon {
  font-size: 1rem;
}

/* Date grid */
.date-grid {
  display: grid;
  grid-template-columns: 1fr 1fr 1fr;
  gap: 12px;
  margin-bottom: 16px;
}

.field-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.field-label {
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 1px;
  color: rgba(212, 168, 75, 0.5);
  text-transform: uppercase;
}

/* Error */
.error-msg {
  color: var(--crimson);
  font-size: 13px;
  text-align: center;
  margin: 0 0 14px;
}

/* Submit */
.btn-submit {
  width: 100%;
  padding: 13px;
  font-size: 0.95rem;
  letter-spacing: 2px;
}

.btn-submit :deep(.loading-spinner) {
  display: inline-block;
  width: 16px;
  height: 16px;
  border: 2px solid rgba(10, 8, 21, 0.3);
  border-top-color: #0a0815;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
