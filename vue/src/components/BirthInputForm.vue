<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import client from '../api/client'

const router = useRouter()
const errMsg = ref('')
const loading = ref(false)
const form = ref({ year: new Date().getFullYear(), month: 1, day: 1, shichen: 4, gender: 'male' as string })
const shichenOptions = [
  { label: '子时 (23-01)', value: 0 },{ label: '丑时 (01-03)', value: 1 },{ label: '寅时 (03-05)', value: 2 },
  { label: '卯时 (05-07)', value: 3 },{ label: '辰时 (07-09)', value: 4 },{ label: '巳时 (09-11)', value: 5 },
  { label: '午时 (11-13)', value: 6 },{ label: '未时 (13-15)', value: 7 },{ label: '申时 (15-17)', value: 8 },
  { label: '酉时 (17-19)', value: 9 },{ label: '戌时 (19-21)', value: 10 },{ label: '亥时 (21-23)', value: 11 },
]
async function handleSubmit() {
  if (!form.value.year || !form.value.month || !form.value.day) { errMsg.value = '请填写完整的出生日期'; return }
  loading.value = true; errMsg.value = ''
  try {
    const res = await client.post('/chart', {
      birth_year: form.value.year, birth_month: form.value.month, birth_day: form.value.day,
      birth_hour: form.value.shichen * 2, birth_min: 0, calendar_type: 'SOLAR',
      gender: form.value.gender.toUpperCase(), name: '',
    })
    sessionStorage.setItem('lastChart', JSON.stringify(res.data))
    router.push('/chart/new?_t=' + Date.now())
  } catch (err: any) {
    errMsg.value = err.response?.data?.error || err.message || '排盘失败，请稍后重试'
  } finally { loading.value = false }
}
</script>
<template>
  <div class="glass-panel" style="padding:32px;max-width:500px;margin:0 auto">
    <h2 style="text-align:center;font-size:1.4rem;font-weight:700;color:#EAE6DD;margin-bottom:28px">输入出生信息</h2>
    <div style="display:flex;gap:12px;margin-bottom:24px">
      <button :class="form.gender==='male'?'btn-gold':'btn-ghost'" style="flex:1;padding:10px" @click="form.gender='male'">男</button>
      <button :class="form.gender==='female'?'btn-gold':'btn-ghost'" style="flex:1;padding:10px" @click="form.gender='female'">女</button>
    </div>
    <div style="display:grid;grid-template-columns:1fr 1fr 1fr;gap:12px;margin-bottom:12px">
      <div><label style="font-size:12px;color:#7A756B">年</label><input v-model.number="form.year" class="input-dark" placeholder="1990" type="number"></div>
      <div><label style="font-size:12px;color:#7A756B">月</label><input v-model.number="form.month" class="input-dark" placeholder="6" type="number"></div>
      <div><label style="font-size:12px;color:#7A756B">日</label><input v-model.number="form.day" class="input-dark" placeholder="15" type="number"></div>
    </div>
    <div style="margin-bottom:24px">
      <label style="font-size:12px;color:#7A756B">时辰</label>
      <select v-model.number="form.shichen" class="input-dark" style="margin-top:4px">
        <option v-for="s in shichenOptions" :key="s.value" :value="s.value">{{ s.label }}</option>
      </select>
    </div>
    <p v-if="errMsg" style="color:#C41E3A;font-size:13px;text-align:center;margin-bottom:16px">{{ errMsg }}</p>
    <button class="btn-gold" :disabled="loading" style="width:100%;padding:14px" @click="handleSubmit">{{ loading?'排盘中...':'开始排盘' }}</button>
  </div>
</template>