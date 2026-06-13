<script setup lang="ts">
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardHeader, CardTitle, CardDescription, CardFooter } from '@/components/ui/card'
import ShaderBackground from '../components/ShaderBackground.vue'


const router = useRouter()
const route = useRoute()
const auth = useAuthStore()
const form = ref({ username: '', password: '' })
const loading = ref(false)
const error = ref('')
const showPassword = ref(false)

async function handleLogin() {
  if (!form.value.username || !form.value.password) {
    error.value = '请填写用户名和密码'
    return
  }
  loading.value = true
  error.value = ''
  try {
    await auth.login(form.value.username, form.value.password)
    const redirect = (route.query.redirect as string) || '/'
    router.push(redirect)
  } catch (e: any) {
    error.value = e.response?.data?.error || '登录失败，请检查用户名和密码'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-page">
    <ShaderBackground yongshen="mu" shader-type="grainGradient" :overlay-opacity="0.74" />

    <Card class="login-card animate-in w-full max-w-[var(--container-sm)] relative z-10
      rounded-[1.5rem] border-0
      [box-shadow:0_0_0_1px_var(--line-strong),var(--shadow-lg)]
      bg-[var(--surface-1)]">
      <!-- Card ornament -->
      <div class="card-ornament" aria-hidden="true">
        <div class="ornament-ring"></div>
        <div class="ornament-symbol">☯</div>
      </div>

      <CardHeader class="text-center pt-2 pb-0">
        <div class="text-[10px] tracking-[3px] text-[var(--text-soft)] uppercase mb-2">BaZi Fortune</div>
        <CardTitle class="font-[family-name:var(--font-serif)] text-[2rem] font-bold tracking-[4px]">
          登录
        </CardTitle>
        <CardDescription class="text-[13px]">探索命运，从这里开始</CardDescription>
      </CardHeader>

      <CardContent class="px-10 pb-10">
        <form @submit.prevent="handleLogin" class="flex flex-col gap-5">
          <div class="flex flex-col gap-2">
            <Label class="text-xs font-semibold tracking-[1px] uppercase text-[var(--text-muted)]">
              用户名
            </Label>
            <div class="relative">
              <span class="absolute left-3.5 top-1/2 -translate-y-1/2 text-xs text-[var(--icon-muted)] z-10 pointer-events-none">✦</span>
              <Input
                v-model="form.username"
                class="pl-10 h-11 rounded-[10px] bg-[var(--glass-bg)] border-[var(--line-strong)] text-sm"
                placeholder="请输入用户名"
                autocomplete="username"
              />
            </div>
          </div>

          <div class="flex flex-col gap-2">
            <Label class="text-xs font-semibold tracking-[1px] uppercase text-[var(--text-muted)]">
              密码
            </Label>
            <div class="relative">
              <span class="absolute left-3.5 top-1/2 -translate-y-1/2 text-xs text-[var(--icon-muted)] z-10 pointer-events-none">◇</span>
              <Input
                v-model="form.password"
                :type="showPassword ? 'text' : 'password'"
                class="pl-10 pr-10 h-11 rounded-[10px] bg-[var(--glass-bg)] border-[var(--line-strong)] text-sm"
                placeholder="请输入密码"
                autocomplete="current-password"
              />
              <button
                type="button"
                class="absolute right-3.5 top-1/2 -translate-y-1/2 text-[var(--icon-muted)] hover:text-[var(--accent)] transition-colors"
                @click="showPassword = !showPassword"
                :aria-label="showPassword ? '隐藏密码' : '显示密码'"
              >
                {{ showPassword ? '◉' : '○' }}
              </button>
            </div>
          </div>

          <div v-if="error" class="flex items-center gap-2 text-xs text-[var(--danger)] bg-[rgba(251,113,133,0.08)] rounded-lg px-3 py-2.5">
            <span>⚠</span> {{ error }}
          </div>

          <Button type="submit" :disabled="loading"
            class="w-full h-11 rounded-full text-sm font-semibold tracking-[1px]
            bg-foreground text-background hover:bg-foreground/90
            shadow-[0_1px_2px_rgba(0,0,0,0.06)]">
            <span v-if="loading" class="loading-spinner"></span>
            <span v-else>登 录</span>
          </Button>
        </form>
      </CardContent>

      <CardFooter class="justify-center pb-10 pt-0">
        <span class="text-[13px] text-[var(--text-muted)]">没有账户？</span>
        <router-link to="/register" class="text-[13px] font-medium text-[var(--primary)] hover:underline ml-1">
          立即注册 →
        </router-link>
      </CardFooter>
    </Card>
  </div>
</template>

<style scoped>
.login-page {
  min-height: calc(100vh - 60px);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
  position: relative;
  overflow: hidden;
}

.card-ornament {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 80px;
  position: relative;
  background: linear-gradient(180deg, var(--accent-dim), transparent);
  border-bottom: 1px solid var(--line-subtle);
}

.ornament-ring {
  position: absolute;
  width: 50px;
  height: 50px;
  border: 1px solid var(--line-focus);
  border-radius: 50%;
  animation: ring-pulse 3s ease-in-out infinite;
}

@keyframes ring-pulse {
  0%, 100% { transform: scale(1); opacity: 0.15; }
  50% { transform: scale(1.1); opacity: 0.3; }
}

.ornament-symbol {
  font-size: 2rem;
  color: var(--accent);
  text-shadow: 0 0 20px var(--brand-glow);
  animation: symbol-glow 3s ease-in-out infinite;
}

@keyframes symbol-glow {
  0%, 100% { text-shadow: 0 0 20px var(--brand-glow); }
  50% { text-shadow: 0 0 35px var(--accent-glow); }
}

.loading-spinner {
  width: 18px;
  height: 18px;
  border: 2px solid color-mix(in oklab, currentColor 20%, transparent);
  border-top-color: currentColor;
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
  display: inline-block;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}
</style>
