<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const auth = useAuthStore()
const form = ref({ username: '', email: '', password: '', confirm: '' })
const loading = ref(false)
const error = ref('')

async function handleRegister() {
  const { username, email, password, confirm } = form.value
  if (!username || !email || !password) { error.value = '请填写完整信息'; return }
  if (password !== confirm) { error.value = '两次密码不一致'; return }
  loading.value = true; error.value = ''
  try {
    await auth.register(username, email, password)
    router.push('/')
  } catch (e: any) {
    error.value = e.response?.data?.error || '注册失败'
  } finally { loading.value = false }
}
</script>
<template>
  <div style="min-height:calc(100vh - 56px);display:flex;align-items:center;justify-content:center;padding:24px">
    <div class="glass-panel" style="width:100%;max-width:420px;padding:48px 40px">
      <div style="text-align:center;margin-bottom:36px">
        <div style="font-size:3rem;color:#D4A84B;margin-bottom:12px">☯</div>
        <h1 style="font-size:1.8rem;font-weight:700;color:#EAE6DD">注册</h1>
        <p style="font-size:14px;color:#7A756B;margin-top:8px">创建账户 · 开启命理探索</p>
      </div>
      <form @submit.prevent="handleRegister" style="display:flex;flex-direction:column;gap:18px">
        <div>
          <label style="display:block;font-size:13px;color:#7A756B;margin-bottom:6px">用户名</label>
          <input v-model="form.username" class="input-dark" placeholder="请输入用户名">
        </div>
        <div>
          <label style="display:block;font-size:13px;color:#7A756B;margin-bottom:6px">邮箱</label>
          <input v-model="form.email" class="input-dark" placeholder="请输入邮箱" type="email">
        </div>
        <div>
          <label style="display:block;font-size:13px;color:#7A756B;margin-bottom:6px">密码</label>
          <input v-model="form.password" type="password" class="input-dark" placeholder="请输入密码">
        </div>
        <div>
          <label style="display:block;font-size:13px;color:#7A756B;margin-bottom:6px">确认密码</label>
          <input v-model="form.confirm" type="password" class="input-dark" placeholder="请再次输入密码">
        </div>
        <p v-if="error" style="color:#C41E3A;font-size:13px;text-align:center">{{ error }}</p>
        <button type="submit" class="btn-gold" :disabled="loading" style="width:100%;margin-top:4px">
          {{ loading ? '注册中...' : '注 册' }}
        </button>
      </form>
      <div style="text-align:center;margin-top:24px">
        <router-link to="/login" style="color:#7A756B;font-size:14px;text-decoration:none">
          已有账户？<span class="text-gold">立即登录</span>
        </router-link>
      </div>
    </div>
  </div>
</template>
