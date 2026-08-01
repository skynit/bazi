<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { getApiErrorMessage } from '../api/client'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
  CardFooter,
} from '@/components/ui/card'
const router = useRouter()
const route = useRoute()
const auth = useAuthStore()
const form = ref({ username: '', email: '', password: '', confirm: '' })
const loading = ref(false)
const error = ref('')
const showPassword = ref(false)
const redirectTarget = computed(() => {
  const value = typeof route.query.redirect === 'string' ? route.query.redirect : ''
  return value.startsWith('/') ? value : '/'
})

async function handleRegister() {
  const { username, email, password, confirm } = form.value
  if (!username || !email || !password) {
    error.value = '请填写完整信息'
    return
  }
  if (!/^[A-Za-z0-9_.-]{3,32}$/.test(username)) {
    error.value = '用户名需为 3–32 位，只能包含字母、数字、下划线、短横线或点。'
    return
  }
  if (Array.from(password).length < 8) {
    error.value = '密码至少需要 8 个字符。'
    return
  }
  if (password !== confirm) {
    error.value = '两次密码不一致'
    return
  }
  loading.value = true
  error.value = ''
  try {
    await auth.register(username, email, password)
    router.push(redirectTarget.value)
  } catch (reason: unknown) {
    error.value = getApiErrorMessage(reason, '注册失败，请检查填写内容。')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="register-page">
    <Card
      class="register-card animate-in w-full max-w-[var(--container-sm)] relative z-10 rounded-[1.5rem] border-0 [box-shadow:0_0_0_1px_var(--line-strong),var(--shadow-lg)] bg-[var(--glass-bg)] backdrop-blur-xl"
    >
      <div class="card-ornament" aria-hidden="true">
        <div class="ornament-ring"></div>
        <div class="ornament-symbol">☯</div>
      </div>

      <CardHeader class="text-center pt-2 pb-0">
        <div class="text-[var(--fs-2xs)] tracking-[3px] text-[var(--text-soft)] uppercase mb-2">
          BaZi Fortune
        </div>
        <CardTitle
          as="h1"
          class="font-[family-name:var(--font-serif)] text-[var(--fs-stat)] font-bold tracking-[4px]"
        >
          注册
        </CardTitle>
        <CardDescription class="text-[var(--fs-xs)]">创建账户 · 开启命理探索</CardDescription>
      </CardHeader>

      <CardContent class="px-10 pb-10">
        <form @submit.prevent="handleRegister" class="flex flex-col gap-5">
          <div class="flex flex-col gap-2">
            <Label
              for="register-username"
              class="text-xs font-semibold tracking-[1px] uppercase text-[var(--text-muted)]"
              >用户名</Label
            >
            <div class="relative">
              <span
                class="absolute left-3.5 top-1/2 -translate-y-1/2 text-xs text-[var(--icon-muted)] z-10 pointer-events-none"
                >✦</span
              >
              <Input
                id="register-username"
                v-model="form.username"
                class="pl-10 h-11 rounded-[10px] bg-[var(--glass-bg)] border-[var(--line-strong)] text-sm"
                placeholder="请输入用户名"
                autocomplete="username"
                aria-describedby="username-rule"
              />
            </div>
            <p id="username-rule" class="m-0 text-[var(--fs-xs)] text-[var(--text-soft)]">
              3–32 位，可使用字母、数字、下划线、短横线和点
            </p>
          </div>

          <div class="flex flex-col gap-2">
            <Label
              for="register-email"
              class="text-xs font-semibold tracking-[1px] uppercase text-[var(--text-muted)]"
              >邮箱</Label
            >
            <div class="relative">
              <span
                class="absolute left-3.5 top-1/2 -translate-y-1/2 text-xs text-[var(--icon-muted)] z-10 pointer-events-none"
                >✧</span
              >
              <Input
                id="register-email"
                v-model="form.email"
                type="email"
                class="pl-10 h-11 rounded-[10px] bg-[var(--glass-bg)] border-[var(--line-strong)] text-sm"
                placeholder="请输入邮箱"
                autocomplete="email"
              />
            </div>
          </div>

          <div class="flex flex-col gap-2">
            <Label
              for="register-password"
              class="text-xs font-semibold tracking-[1px] uppercase text-[var(--text-muted)]"
              >密码</Label
            >
            <div class="relative">
              <span
                class="absolute left-3.5 top-1/2 -translate-y-1/2 text-xs text-[var(--icon-muted)] z-10 pointer-events-none"
                >◇</span
              >
              <Input
                id="register-password"
                v-model="form.password"
                :type="showPassword ? 'text' : 'password'"
                class="pl-10 pr-10 h-11 rounded-[10px] bg-[var(--glass-bg)] border-[var(--line-strong)] text-sm"
                placeholder="请输入密码"
                autocomplete="new-password"
                aria-describedby="password-rule"
              />
              <button
                type="button"
                class="absolute right-3.5 top-1/2 -translate-y-1/2 text-[var(--icon-muted)] hover:text-[var(--accent)] transition-colors"
                :aria-label="showPassword ? '隐藏密码' : '显示密码'"
                @click="showPassword = !showPassword"
              >
                {{ showPassword ? '◉' : '○' }}
              </button>
            </div>
            <p id="password-rule" class="m-0 text-[var(--fs-xs)] text-[var(--text-soft)]">
              至少 8 个字符
            </p>
          </div>

          <div class="flex flex-col gap-2">
            <Label
              for="register-confirm"
              class="text-xs font-semibold tracking-[1px] uppercase text-[var(--text-muted)]"
              >确认密码</Label
            >
            <div class="relative">
              <span
                class="absolute left-3.5 top-1/2 -translate-y-1/2 text-xs text-[var(--icon-muted)] z-10 pointer-events-none"
                >◇</span
              >
              <Input
                id="register-confirm"
                v-model="form.confirm"
                type="password"
                class="pl-10 h-11 rounded-[10px] bg-[var(--glass-bg)] border-[var(--line-strong)] text-sm"
                placeholder="请再次输入密码"
                autocomplete="new-password"
              />
            </div>
          </div>

          <div
            v-if="error"
            role="alert"
            class="flex items-center gap-2 text-xs text-[var(--danger)] bg-[rgba(251,113,133,0.08)] rounded-lg px-3 py-2.5"
          >
            <span>⚠</span> {{ error }}
          </div>

          <Button
            type="submit"
            :disabled="loading"
            class="w-full h-11 rounded-full text-sm font-semibold tracking-[1px] bg-foreground text-background hover:bg-foreground/90 shadow-[0_1px_2px_rgba(0,0,0,0.06)]"
          >
            <span v-if="loading" class="loading-spinner"></span>
            <span v-else>注 册</span>
          </Button>
        </form>
      </CardContent>

      <CardFooter class="justify-center pb-10 pt-0">
        <span class="text-[var(--fs-xs)] text-[var(--text-muted)]">已有账户？</span>
        <router-link
          :to="{
            path: '/login',
            query: redirectTarget !== '/' ? { redirect: redirectTarget } : {},
          }"
          class="text-[var(--fs-xs)] font-medium text-[var(--primary)] hover:underline ml-1"
        >
          立即登录 →
        </router-link>
      </CardFooter>
    </Card>
  </div>
</template>

<style scoped>
.register-page {
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
  font-size: var(--fs-stat);
  color: var(--accent);
  text-shadow: 0 0 20px var(--brand-glow);
  animation: symbol-glow 3s ease-in-out infinite;
}

@keyframes symbol-glow {
  0%,
  100% {
    text-shadow: 0 0 20px var(--brand-glow);
  }
  50% {
    text-shadow: 0 0 35px var(--accent-glow);
  }
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
  to {
    transform: rotate(360deg);
  }
}
</style>
