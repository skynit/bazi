<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from './stores/auth'
import { useThemeStore } from './stores/theme'
import type { WuxingKey } from './composables/useWuxingThemes'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const themeStore = useThemeStore()
const scrolled = ref(false)
const isDark = ref(document.documentElement.classList.contains('dark'))
const mobileOpen = ref(false)

const wuxingCycle: Array<{ key: WuxingKey; label: string }> = [
  { key: 'mu', label: '木' },
  { key: 'huo', label: '火' },
  { key: 'tu', label: '土' },
  { key: 'jin', label: '金' },
  { key: 'shui', label: '水' }
]

function applyTheme(nextDark: boolean) {
  isDark.value = nextDark
  document.documentElement.classList.toggle('dark', nextDark)
  document.documentElement.style.colorScheme = nextDark ? 'dark' : 'light'
  localStorage.setItem('theme', nextDark ? 'dark' : 'light')
}

function onScroll() {
  scrolled.value = window.scrollY > 20
}

function toggleTheme() {
  applyTheme(!isDark.value)
}

function closeMobileMenu() {
  mobileOpen.value = false
}

function toggleMobileMenu() {
  mobileOpen.value = !mobileOpen.value
}

function logout() {
  closeMobileMenu()
  authStore.logout()
  router.push('/')
}

/**
 * 紫微斗数路由需要 chartId。
 * 优先使用最近一次排盘 (bazi_last_birth)，否则引导用户先排盘。
 */
function goZiwei() {
  closeMobileMenu()
  try {
    const raw = localStorage.getItem('bazi_last_birth')
    const chartId = raw ? JSON.parse(raw).chartId : null
    if (chartId) {
      router.push(`/ziwei/${chartId}`)
      return
    }
  } catch { /* ignore */ }
  router.push('/chart/new')
}

function goBaziChart() {
  closeMobileMenu()
  try {
    const raw = localStorage.getItem('bazi_last_birth')
    const chartId = raw ? JSON.parse(raw).chartId : null
    if (chartId) {
      router.push(`/chart/${chartId}`)
      return
    }
  } catch { /* ignore */ }
  router.push('/chart/new')
}

function onKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape') {
    closeMobileMenu()
  }
}

watch(() => route.fullPath, closeMobileMenu)

onMounted(() => {
  document.documentElement.style.colorScheme = isDark.value ? 'dark' : 'light'
  window.addEventListener('scroll', onScroll, { passive: true })
  window.addEventListener('keydown', onKeydown)
})
onUnmounted(() => {
  window.removeEventListener('scroll', onScroll)
  window.removeEventListener('keydown', onKeydown)
})
</script>
<template>
  <div class="app-root">
    <header class="fixed inset-x-0 top-4 z-50">
      <div class="mx-auto w-full max-w-[82rem] px-4 sm:px-8">
        <nav
          class="relative flex items-center h-14 rounded-full border px-4 sm:px-5 lg:px-6 transition-all duration-300"
          :class="scrolled
            ? 'bg-[var(--nav-bg)] backdrop-blur-md border-[var(--nav-border)] shadow-[var(--shadow-sm)]'
            : 'bg-[var(--nav-bg-idle)] backdrop-blur-md border-transparent'"
        >
          <!-- Logo (left) -->
          <router-link to="/" class="flex items-center gap-3 text-decoration-none shrink-0 mr-auto" @click="closeMobileMenu">
            <div class="relative flex items-center justify-center w-8 h-8">
              <div class="absolute -inset-[3px] border border-[var(--line-focus)] rounded-full animate-[spin_12s_linear_infinite]"></div>
              <span class="text-[var(--fs-2xl)] text-[var(--accent)] [text-shadow:0_0_12px_var(--brand-glow)]">☯</span>
            </div>
            <div class="hidden sm:flex flex-col gap-0">
              <span class="font-[family-name:var(--font-serif)] text-[var(--fs-sm)] font-bold text-[var(--text)] tracking-[2px] leading-[1.1]">八字命理</span>
              <span class="text-[var(--fs-2xs)] tracking-[1.6px] text-[var(--text-soft)] uppercase">BaZi Fortune</span>
            </div>
          </router-link>

          <!-- Nav links (absolutely centered, optical balance) -->
          <nav class="hidden md:flex absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 items-center gap-5 lg:gap-7 text-[var(--fs-xs)] font-medium text-[var(--text-muted)] whitespace-nowrap">
            <router-link to="/" class="transition-colors hover:text-[var(--accent)] aria-[current=page]:text-[var(--accent)]" exact-active-class="text-[var(--accent)]">首页</router-link>
            <template v-if="authStore.isLoggedIn()">
              <router-link to="/history" class="transition-colors hover:text-[var(--accent)]">历史</router-link>
              <button
                type="button"
                @click="goBaziChart"
                class="transition-colors hover:text-[var(--accent)] bg-transparent border-0 p-0 cursor-pointer font-medium text-[var(--fs-xs)] text-[var(--text-muted)]"
                :class="{ 'text-[var(--accent)]': $route.name === 'Chart' }"
              >
                八字命盘
              </button>
              <div class="relative group">
                <router-link to="/fortune" class="transition-colors hover:text-[var(--accent)] flex items-center gap-1">
                  运势 <span class="text-[var(--fs-2xs)] opacity-50">▾</span>
                </router-link>
                <div class="hidden group-hover:block absolute top-full left-1/2 -translate-x-1/2 min-w-[136px] py-1.5 mt-1 bg-[var(--surface-1)]/95 border border-[var(--line-strong)] rounded-[10px] shadow-[0_12px_40px_rgba(0,0,0,0.18)] backdrop-blur-[20px] z-[100]">
                  <router-link to="/fortune" class="block px-4 py-2 text-xs text-[var(--text-muted)] hover:text-[var(--accent)] hover:bg-[var(--menu-hover)] transition-colors">今日运势</router-link>
                  <router-link to="/fortune/blessing" class="block px-4 py-2 text-xs text-[var(--text-muted)] hover:text-[var(--accent)] hover:bg-[var(--menu-hover)] transition-colors">运势加持</router-link>
                  <router-link to="/fortune/weekly" class="block px-4 py-2 text-xs text-[var(--text-muted)] hover:text-[var(--accent)] hover:bg-[var(--menu-hover)] transition-colors">本周运势</router-link>
                  <router-link to="/fortune/monthly" class="block px-4 py-2 text-xs text-[var(--text-muted)] hover:text-[var(--accent)] hover:bg-[var(--menu-hover)] transition-colors">本月运势</router-link>
                </div>
              </div>
              <router-link to="/buyi" class="transition-colors hover:text-[var(--accent)]" :class="{ 'text-[var(--accent)]': $route.name === 'Buyi' }">卜易</router-link>
              <button
                @click="goZiwei"
                class="transition-colors hover:text-[var(--accent)] flex items-center gap-1 bg-transparent border-0 p-0 cursor-pointer font-medium text-[var(--fs-xs)] text-[var(--text-muted)]"
                :class="{ 'text-[var(--accent)]': $route.name === 'ZiWei' }"
              >
                <span class="text-[var(--fs-2xs)] opacity-60">✦</span>
                紫微斗数
              </button>
            </template>
          </nav>

          <!-- Right side -->
          <div class="flex items-center gap-3 shrink-0 ml-auto">
            <!-- Five-element switcher -->
            <div class="hidden md:flex items-center gap-1 px-1.5 py-1 rounded-full border border-[var(--line-subtle)] bg-[var(--glass-bg)] backdrop-blur-md">
              <button
                v-for="el in wuxingCycle"
                :key="el.key"
                type="button"
                @click="themeStore.setElementTheme(el.key)"
                :class="[
                  'w-6 h-6 flex items-center justify-center rounded-full text-[var(--fs-2xs)] font-[family-name:var(--font-serif)] transition-all',
                  themeStore.elementTheme === el.key
                    ? 'bg-[var(--accent)] text-[var(--bg)] shadow-[0_0_10px_var(--brand-glow)]'
                    : 'text-[var(--text-muted)] hover:text-[var(--accent)]'
                ]"
                :aria-label="`切换至${el.label}主题`"
                :title="`${el.label}行主题`"
              >{{ el.label }}</button>
            </div>
            <button
              @click="toggleTheme"
              class="w-8 h-8 flex items-center justify-center rounded-full text-sm transition-colors hover:bg-[var(--glass-bg)]"
              :aria-label="isDark ? '切换到白天模式' : '切换到暗黑模式'"
            >
              <span v-if="isDark" class="text-[var(--fs-sm)] text-[var(--icon)]">☀</span>
              <span v-else class="text-[var(--fs-sm)] text-[var(--icon)]">☽</span>
            </button>
            <template v-if="authStore.isLoggedIn()">
              <div class="hidden sm:flex items-center gap-2">
                <div class="w-7 h-7 rounded-full bg-[var(--accent)]/10 flex items-center justify-center text-[var(--fs-2xs)] font-semibold text-[var(--accent)]">
                  {{ authStore.user?.username?.charAt(0).toUpperCase() }}
                </div>
                <span class="text-[var(--fs-2xs)] font-medium text-[var(--text-muted)] hidden lg:inline">{{ authStore.user?.username }}</span>
              </div>
              <button @click="logout" class="text-[var(--fs-2xs)] text-[var(--text-muted)] hover:text-[var(--accent)] transition-colors hidden sm:inline">退出</button>
            </template>
            <template v-else>
              <router-link to="/login" class="text-[var(--fs-xs)] font-medium text-[var(--text-muted)] hover:text-[var(--accent)] transition-colors">登录</router-link>
              <router-link to="/register" class="inline-flex items-center justify-center h-8 px-4 text-[var(--fs-2xs)] font-medium rounded-full bg-[var(--text)] text-[var(--bg)] hover:bg-[var(--text)]/90 transition-colors shadow-[0_1px_2px_rgba(0,0,0,0.06)]">
                注册
              </router-link>
            </template>
            <button
              type="button"
              class="md:hidden w-9 h-9 flex items-center justify-center rounded-full border border-[var(--line-subtle)] text-[var(--text-muted)] hover:text-[var(--accent)] hover:bg-[var(--glass-bg)] transition-colors"
              :aria-expanded="mobileOpen"
              aria-controls="mobile-navigation"
              :aria-label="mobileOpen ? '关闭导航菜单' : '打开导航菜单'"
              @click="toggleMobileMenu"
            >
              <span class="relative block h-4 w-4" aria-hidden="true">
                <span
                  class="absolute left-0 top-0.5 h-px w-4 bg-current transition-transform duration-200"
                  :class="mobileOpen ? 'translate-y-[5px] rotate-45' : ''"
                ></span>
                <span
                  class="absolute left-0 top-[7px] h-px w-4 bg-current transition-opacity duration-200"
                  :class="mobileOpen ? 'opacity-0' : ''"
                ></span>
                <span
                  class="absolute left-0 top-[13px] h-px w-4 bg-current transition-transform duration-200"
                  :class="mobileOpen ? '-translate-y-[5px] -rotate-45' : ''"
                ></span>
              </span>
            </button>
          </div>
        </nav>

        <Transition
          enter-active-class="transition duration-200 ease-out"
          enter-from-class="opacity-0 -translate-y-2 scale-[0.98]"
          enter-to-class="opacity-100 translate-y-0 scale-100"
          leave-active-class="transition duration-150 ease-in"
          leave-from-class="opacity-100 translate-y-0 scale-100"
          leave-to-class="opacity-0 -translate-y-2 scale-[0.98]"
        >
          <div
            v-if="mobileOpen"
            id="mobile-navigation"
            class="md:hidden mt-2 overflow-hidden rounded-[22px] border border-[var(--nav-border)] bg-[var(--nav-bg)]/95 p-3 shadow-[0_18px_50px_rgba(0,0,0,0.22)] backdrop-blur-xl"
          >
            <div class="grid grid-cols-2 gap-2 text-[var(--fs-sm)]">
              <router-link to="/" class="mobile-nav-item col-span-2" exact-active-class="mobile-nav-item-active">
                <span aria-hidden="true">⌂</span><span>首页</span>
              </router-link>

              <template v-if="authStore.isLoggedIn()">
                <router-link to="/history" class="mobile-nav-item" active-class="mobile-nav-item-active">
                  <span aria-hidden="true">◷</span><span>历史</span>
                </router-link>
                <button type="button" class="mobile-nav-item" :class="{ 'mobile-nav-item-active': route.name === 'Chart' }" @click="goBaziChart">
                  <span aria-hidden="true">八</span><span>八字命盘</span>
                </button>
                <router-link to="/fortune" class="mobile-nav-item" active-class="mobile-nav-item-active">
                  <span aria-hidden="true">今</span><span>今日运势</span>
                </router-link>
                <router-link to="/fortune/blessing" class="mobile-nav-item" active-class="mobile-nav-item-active">
                  <span aria-hidden="true">福</span><span>运势加持</span>
                </router-link>
                <router-link to="/fortune/weekly" class="mobile-nav-item" active-class="mobile-nav-item-active">
                  <span aria-hidden="true">周</span><span>本周运势</span>
                </router-link>
                <router-link to="/fortune/monthly" class="mobile-nav-item" active-class="mobile-nav-item-active">
                  <span aria-hidden="true">月</span><span>本月运势</span>
                </router-link>
                <router-link to="/buyi" class="mobile-nav-item" active-class="mobile-nav-item-active">
                  <span aria-hidden="true">卦</span><span>卜易</span>
                </router-link>
                <button type="button" class="mobile-nav-item" :class="{ 'mobile-nav-item-active': route.name === 'ZiWei' }" @click="goZiwei">
                  <span aria-hidden="true">✦</span><span>紫微斗数</span>
                </button>
                <button type="button" class="mobile-nav-item col-span-2 justify-center text-[var(--text-soft)]" @click="logout">
                  退出登录
                </button>
              </template>

              <template v-else>
                <router-link to="/login" class="mobile-nav-item" active-class="mobile-nav-item-active">登录</router-link>
                <router-link to="/register" class="mobile-nav-item justify-center bg-[var(--accent)] text-[var(--bg)]" active-class="mobile-nav-item-active">注册</router-link>
              </template>
            </div>
          </div>
        </Transition>
      </div>
    </header>
    <main class="app-main"><router-view /></main>
  </div>
</template>

<style scoped>
.app-root {
  min-height: 100vh;
  background: var(--bg);
  display: flex;
  flex-direction: column;
}

.app-main {
  flex: 1;
  padding-top: 80px;
}

.router-link-active {
  color: var(--accent) !important;
}

.mobile-nav-item {
  min-height: 44px;
  display: flex;
  align-items: center;
  gap: 10px;
  border: 1px solid var(--line-subtle);
  border-radius: 14px;
  padding: 10px 14px;
  background: var(--glass-bg);
  color: var(--text-muted);
  text-align: left;
  transition: color 160ms ease, border-color 160ms ease, background 160ms ease;
}

.mobile-nav-item:hover,
.mobile-nav-item-active {
  border-color: var(--line-focus);
  background: var(--menu-hover);
  color: var(--accent) !important;
}
</style>
