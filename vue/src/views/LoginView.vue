<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const auth = useAuthStore()
const form = ref({ username: '', password: '' })
const loading = ref(false)
const error = ref('')

async function handleLogin() {
  if (!form.value.username || !form.value.password) {
    error.value = '请填写用户名和密码'; return
  }
  loading.value = true; error.value = ''
  try {
    await auth.login(form.value.username, form.value.password)
    router.push('/')
  } catch (e: any) {
    error.value = e.response?.data?.error || '登录失败，请检查用户名和密码'
  } finally { loading.value = false }
}
</script>
<template>
  <div style="min-height:calc(100vh - 56px);display:flex;align-items:center;justify-content:center;padding:24px">
    <div class="glass-panel" style="width:100%;max-width:420px;padding:48px 40px">
      <div style="text-align:center;margin-bottom:40px">
        <div style="font-size:3rem;color:#D4A84B;margin-bottom:12px">☯</div>
        <h1 style="font-size:1.8rem;font-weight:700;color:#EAE6DD">登录</h1>
        <p style="font-size:14px;color:#7A756B;margin-top:8px">八字命理 · 探索命运</p>
      </div>
      <form @submit.prevent="handleLogin" style="display:flex;flex-direction:column;gap:20px">
        <div>
          <label style="display:block;font-size:13px;color:#7A756B;margin-bottom:8px">用户名</label>
          <input v-model="form.username" class="input-dark" placeholder="请输入用户名" autocomplete="username">
        </div>
        <div>
          <label style="display:block;font-size:13px;color:#7A756B;margin-bottom:8px">密码</label>
          <input v-model="form.password" type="password" class="input-dark" placeholder="请输入密码" autocomplete="current-password">
        </div>
        <p v-if="error" style="color:#C41E3A;font-size:13px;text-align:center">{{ error }}</p>
        <button type="submit" class="btn-gold" :disabled="loading" style="width:100%;margin-top:8px">
          {{ loading ? '登录中...' : '登 录' }}
        </button>
      </form>
      <div style="text-align:center;margin-top:24px">
        <router-link to="/register" style="color:#7A756B;font-size:14px;text-decoration:none">
          没有账户？<span class="text-gold">立即注册</span>
        </router-link>
      </div>
    </div>
  </div>
</template>
