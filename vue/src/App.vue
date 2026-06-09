<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useAuthStore } from './stores/auth'

const router = useRouter()
const authStore = useAuthStore()
</script>
<template>
  <div class="app-root">
    <header class="app-header">
      <div class="header-inner">
        <!-- Logo -->
        <router-link to="/" class="logo-link">
          <div class="logo-symbol">
            <div class="logo-ring"></div>
            <span class="logo-icon">☯</span>
          </div>
          <div class="logo-text">
            <span class="logo-title">八字命理</span>
            <span class="logo-sub">BaZi Fortune</span>
          </div>
        </router-link>

        <!-- Nav -->
        <nav class="app-nav">
          <router-link to="/" class="nav-link">
            <span class="nav-dot"></span>
            首页
          </router-link>
          <template v-if="authStore.isLoggedIn()">
            <router-link to="/history" class="nav-link">
              <span class="nav-dot"></span>
              历史
            </router-link>
            <div class="nav-dropdown">
              <router-link to="/fortune" class="nav-link nav-dropdown-trigger">
                <span class="nav-dot"></span>
                运势
                <span class="dropdown-arrow">▾</span>
              </router-link>
              <div class="dropdown-menu">
                <router-link to="/fortune" class="dropdown-item">今日运势</router-link>
                <router-link to="/fortune/weekly" class="dropdown-item">本周运势</router-link>
                <router-link to="/fortune/monthly" class="dropdown-item">本月运势</router-link>
              </div>
            </div>
            <div class="nav-divider"></div>
            <div class="user-chip">
              <div class="user-avatar">{{ authStore.user?.username?.charAt(0).toUpperCase() }}</div>
              <span class="user-name">{{ authStore.user?.username }}</span>
            </div>
            <button @click="authStore.logout();router.push('/')" class="logout-btn">退出</button>
          </template>
          <template v-else>
            <router-link to="/login" class="nav-link">登录</router-link>
            <router-link to="/register" class="btn-gold nav-register">注册</router-link>
          </template>
        </nav>
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

/* ── Header ── */
.app-header {
  position: sticky;
  top: 0;
  z-index: 50;
  background: rgba(3, 4, 4, 0.85);
  backdrop-filter: blur(24px);
  -webkit-backdrop-filter: blur(24px);
  border-bottom: 1px solid rgba(203, 213, 225, 0.06);
  transition: border-color 0.3s ease;
}

.app-header:hover {
  border-bottom-color: rgba(203, 213, 225, 0.1);
}

.header-inner {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 32px;
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
}

/* ── Logo ── */
.logo-link {
  display: flex;
  align-items: center;
  gap: 12px;
  text-decoration: none;
}

.logo-symbol {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
}

.logo-ring {
  position: absolute;
  inset: -3px;
  border: 1px solid rgba(203, 213, 225, 0.2);
  border-radius: 50%;
  animation: logo-spin 12s linear infinite;
}

@keyframes logo-spin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }

.logo-icon {
  font-size: 1.4rem;
  color: var(--accent);
  text-shadow: 0 0 12px rgba(203, 213, 225, 0.4);
}

.logo-text {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.logo-title {
  font-family: var(--font-serif), serif;
  font-size: 1.05rem;
  font-weight: 700;
  color: var(--text);
  letter-spacing: 3px;
  line-height: 1.1;
}

.logo-sub {
  font-size: 9px;
  letter-spacing: 2px;
  color: rgba(203, 213, 225, 0.4);
  text-transform: uppercase;
}

/* ── Nav ── */
.app-nav {
  display: flex;
  align-items: center;
  gap: 8px;
}

.nav-link {
  position: relative;
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 14px;
  font-size: 13px;
  font-weight: 500;
  color: var(--muted);
  text-decoration: none;
  border-radius: 6px;
  transition: all 0.2s ease;
}

.nav-dot {
  display: block;
  width: 4px;
  height: 4px;
  border-radius: 50%;
  background: rgba(203, 213, 225, 0.3);
  transition: all 0.2s ease;
}

.nav-link:hover {
  color: var(--accent);
  background: rgba(203, 213, 225, 0.05);
}

.nav-link:hover .nav-dot {
  background: var(--accent);
  box-shadow: 0 0 6px rgba(203, 213, 225, 0.5);
}

.nav-link.router-link-active {
  color: var(--accent);
}

.nav-link.router-link-active .nav-dot {
  background: var(--accent);
}

.nav-divider {
  width: 1px;
  height: 20px;
  background: rgba(255,255,255,0.08);
  margin: 0 4px;
}

/* ── Dropdown ── */
.nav-dropdown {
  position: relative;
}

.dropdown-arrow {
  font-size: 10px;
  margin-left: 2px;
  opacity: 0.5;
}

.dropdown-menu {
  display: none;
  position: absolute;
  top: 100%;
  left: 0;
  min-width: 120px;
  padding: 6px 0;
  background: rgba(15, 12, 28, 0.96);
  border: 1px solid rgba(203, 213, 225, 0.12);
  border-radius: 10px;
  box-shadow: 0 12px 40px rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(20px);
  z-index: 100;
}

.nav-dropdown:hover .dropdown-menu {
  display: block;
}

.dropdown-item {
  display: block;
  padding: 8px 16px;
  font-size: 12px;
  color: var(--muted);
  text-decoration: none;
  transition: all 0.15s;
}

.dropdown-item:hover {
  color: var(--accent);
  background: rgba(203, 213, 225, 0.06);
}

.dropdown-item.router-link-active {
  color: var(--accent);
}

/* ── User chip ── */
.user-chip {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 10px 4px 4px;
  border-radius: 20px;
  background: rgba(255,255,255,0.04);
  border: 1px solid rgba(203, 213, 225, 0.06);
}

.user-avatar {
  width: 26px;
  height: 26px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--accent), #94a3b8);
  color: #0A0815;
  font-size: 11px;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
}

.user-name {
  font-size: 12px;
  color: rgba(203,213,225,0.6);
  font-weight: 500;
}

.logout-btn {
  padding: 6px 12px;
  font-size: 12px;
  color: var(--muted);
  background: transparent;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.logout-btn:hover {
  color: var(--danger);
  background: rgba(251,113,133,0.08);
}

.nav-register {
  padding: 7px 18px;
  font-size: 13px;
  border-radius: 8px;
}

/* ── Main ── */
.app-main {
  flex: 1;
}

/* ── Responsive ── */
@media (max-width: 640px) {
  .header-inner { padding: 0 16px; }
  .logo-text { display: none; }
  .user-name { display: none; }
  .nav-link { padding: 6px 10px; }
}
</style>