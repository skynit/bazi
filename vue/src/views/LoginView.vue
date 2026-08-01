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
const form = ref({ username: '', password: '' })
const loading = ref(false)
const error = ref('')
const showPassword = ref(false)
const redirectTarget = computed(() => {
  const value = typeof route.query.redirect === 'string' ? route.query.redirect : ''
  return value.startsWith('/') ? value : '/'
})
const loginContext = computed(() => {
  if (route.query.expired === '1') return '登录状态已过期，请重新登录后继续。'
  if (redirectTarget.value !== '/') return '请先登录，完成后将继续访问刚才的页面。'
  return '登录后继续使用排盘与运势功能。'
})

async function handleLogin() {
  if (!form.value.username || !form.value.password) {
    error.value = '请填写用户名和密码'
    return
  }
  loading.value = true
  error.value = ''
  try {
    await auth.login(form.value.username, form.value.password)
    router.push(redirectTarget.value)
  } catch (reason: unknown) {
    error.value = getApiErrorMessage(reason, '登录失败，请检查用户名和密码。')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-page">
    <Card
      class="login-card animate-in w-full max-w-[var(--container-sm)] relative z-10 rounded-[1.25rem] border-0 [box-shadow:0_0_0_1px_var(--line-strong),var(--shadow-lg)] bg-[var(--bg-elevated)]"
    >
      <!-- Card ornament -->
      <div class="card-ornament" aria-hidden="true">
        <span class="ornament-rule"></span>
        <span class="ornament-seal">☯</span>
        <span class="ornament-rule"></span>
      </div>

      <CardHeader class="text-center pt-1 pb-0">
        <div
          class="text-[var(--fs-2xs)] tracking-[var(--tracking-eyebrow)] text-[var(--text-soft)] uppercase mb-2"
        >
          BaZi Fortune
        </div>
        <CardTitle
          as="h1"
          class="font-[family-name:var(--font-serif)] text-[var(--fs-stat)] font-bold tracking-[0.2em]"
        >
          登录
        </CardTitle>
        <CardDescription class="text-[var(--fs-xs)] mt-1">{{ loginContext }}</CardDescription>
      </CardHeader>

      <CardContent class="px-10 pb-8 pt-6">
        <form @submit.prevent="handleLogin" class="flex flex-col gap-5">
          <div class="flex flex-col gap-1.5">
            <Label
              for="login-username"
              class="text-[var(--fs-xs)] font-medium tracking-[0.08em] text-[var(--text-muted)]"
            >
              用户名
            </Label>
            <Input
              id="login-username"
              v-model="form.username"
              class="h-11 rounded-[var(--radius-md)] bg-transparent border-[var(--line-strong)] text-sm transition-colors hover:border-[var(--line-focus)]"
              placeholder="请输入用户名"
              autocomplete="username"
            />
          </div>

          <div class="flex flex-col gap-1.5">
            <Label
              for="login-password"
              class="text-[var(--fs-xs)] font-medium tracking-[0.08em] text-[var(--text-muted)]"
            >
              密码
            </Label>
            <div class="relative">
              <Input
                id="login-password"
                v-model="form.password"
                :type="showPassword ? 'text' : 'password'"
                class="pr-10 h-11 rounded-[var(--radius-md)] bg-transparent border-[var(--line-strong)] text-sm transition-colors hover:border-[var(--line-focus)]"
                placeholder="请输入密码"
                autocomplete="current-password"
              />
              <button
                type="button"
                class="absolute right-3 top-1/2 -translate-y-1/2 text-[var(--icon-subtle)] hover:text-[var(--text)] transition-colors"
                @click="showPassword = !showPassword"
                :aria-label="showPassword ? '隐藏密码' : '显示密码'"
              >
                {{ showPassword ? '◉' : '○' }}
              </button>
            </div>
          </div>

          <div
            v-if="error"
            role="alert"
            class="flex items-center gap-2 text-[var(--fs-xs)] text-[var(--danger)] bg-[color-mix(in_oklab,var(--danger)_8%,transparent)] border border-[color-mix(in_oklab,var(--danger)_24%,transparent)] rounded-[var(--radius-md)] px-3 py-2.5"
          >
            <span aria-hidden="true">⚠</span> {{ error }}
          </div>

          <Button
            type="submit"
            :disabled="loading"
            class="w-full h-11 rounded-full text-sm font-semibold tracking-[0.3em] indent-[0.3em] bg-foreground text-background hover:bg-foreground/90 active:scale-[0.99] transition-all shadow-[var(--shadow-sm)] disabled:opacity-60"
          >
            <span v-if="loading" class="loading-spinner"></span>
            <span v-else>登录</span>
          </Button>
        </form>
      </CardContent>

      <CardFooter class="justify-center py-5 border-t border-[var(--line-subtle)]">
        <span class="text-[var(--fs-xs)] text-[var(--text-muted)]">没有账户？</span>
        <router-link
          :to="{
            path: '/register',
            query: redirectTarget !== '/' ? { redirect: redirectTarget } : {},
          }"
          class="text-[var(--fs-xs)] font-medium text-[var(--primary)] hover:underline underline-offset-4 ml-1.5"
        >
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
  gap: 14px;
  padding: 26px 40px 20px;
}

.ornament-rule {
  flex: 1;
  max-width: 72px;
  height: 1px;
  background: var(--line-strong);
}

.ornament-seal {
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--line-strong);
  border-radius: 50%;
  font-size: var(--fs-lg);
  color: var(--accent);
  background: var(--accent-dim);
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
