<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
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
    router.push('/')
  } catch (e: any) {
    error.value = e.response?.data?.error || '登录失败，请检查用户名和密码'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-page">
    <!-- Background constellation -->
    <div class="bg-constellation" aria-hidden="true">
      <svg viewBox="0 0 800 600" preserveAspectRatio="xMidYMid slice" class="constellation-svg">
        <defs>
          <radialGradient id="login-nebula" cx="50%" cy="50%" r="50%">
            <stop offset="0%" stop-color="#D4A84B" stop-opacity="0.06" />
            <stop offset="100%" stop-color="#D4A84B" stop-opacity="0" />
          </radialGradient>
        </defs>
        <ellipse cx="400" cy="300" rx="350" ry="250" fill="url(#login-nebula)" />
        <circle cx="100" cy="100" r="1" fill="#D4A84B" opacity="0.3" />
        <circle cx="650" cy="80" r="1.2" fill="#D4A84B" opacity="0.4" />
        <circle cx="700" cy="500" r="1" fill="#D4A84B" opacity="0.3" />
        <circle cx="80" cy="450" r="0.8" fill="#D4A84B" opacity="0.25" />
        <circle cx="400" cy="300" r="1.5" fill="#D4A84B" opacity="0.5" filter="url(#star-glow)" />
        <circle cx="250" cy="200" r="1" fill="#D4A84B" opacity="0.35" />
        <circle cx="550" cy="420" r="1.3" fill="#D4A84B" opacity="0.4" />
        <circle cx="600" cy="150" r="0.9" fill="#fff" opacity="0.2" />
        <line
          x1="100"
          y1="100"
          x2="250"
          y2="200"
          stroke="#D4A84B"
          stroke-width="0.4"
          opacity="0.08"
        />
        <line
          x1="650"
          y1="80"
          x2="550"
          y2="420"
          stroke="#D4A84B"
          stroke-width="0.4"
          opacity="0.06"
        />
      </svg>
    </div>

    <!-- Login card -->
    <div class="login-card animate-in">
      <!-- Card header ornament -->
      <div class="card-ornament" aria-hidden="true">
        <div class="ornament-ring"></div>
        <div class="ornament-symbol">☯</div>
      </div>

      <div class="card-inner">
        <!-- Title -->
        <div class="card-header">
          <div class="header-eyebrow">BaZi Fortune</div>
          <h1 class="header-title">登录</h1>
          <p class="header-sub">探索命运，从这里开始</p>
        </div>

        <!-- Form -->
        <form @submit.prevent="handleLogin" class="login-form">
          <div class="form-field">
            <label class="field-label">用户名</label>
            <div class="field-input-wrap">
              <span class="field-icon">✦</span>
              <input
                v-model="form.username"
                class="field-input"
                placeholder="请输入用户名"
                autocomplete="username"
              />
            </div>
          </div>

          <div class="form-field">
            <label class="field-label">密码</label>
            <div class="field-input-wrap">
              <span class="field-icon">◇</span>
              <input
                v-model="form.password"
                :type="showPassword ? 'text' : 'password'"
                class="field-input"
                placeholder="请输入密码"
                autocomplete="current-password"
              />
              <button
                type="button"
                class="toggle-password"
                @click="showPassword = !showPassword"
                :aria-label="showPassword ? '隐藏密码' : '显示密码'"
              >
                {{ showPassword ? '◉' : '○' }}
              </button>
            </div>
          </div>

          <!-- Error message -->
          <div v-if="error" class="error-msg">
            <span class="error-icon">⚠</span>
            {{ error }}
          </div>

          <!-- Submit -->
          <button type="submit" class="btn-submit" :disabled="loading">
            <span v-if="loading" class="loading-spinner"></span>
            <span v-else>登 录</span>
          </button>
        </form>

        <!-- Register link -->
        <div class="card-footer">
          <span class="footer-text">没有账户？</span>
          <router-link to="/register" class="footer-link">
            立即注册
            <span class="footer-arrow">→</span>
          </router-link>
        </div>
      </div>
    </div>
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

/* ── Background constellation ── */
.bg-constellation {
  position: absolute;
  inset: 0;
  pointer-events: none;
  z-index: 0;
}

.constellation-svg {
  width: 100%;
  height: 100%;
  position: absolute;
  inset: 0;
}

/* ── Card ── */
.login-card {
  position: relative;
  z-index: 1;
  width: 100%;
  max-width: 420px;
  background: linear-gradient(160deg, rgba(25, 20, 40, 0.95), rgba(12, 10, 22, 0.98));
  border: 1px solid rgba(212, 168, 75, 0.12);
  border-radius: 24px;
  box-shadow:
    0 0 80px rgba(212, 168, 75, 0.06),
    0 25px 60px rgba(0, 0, 0, 0.4);
  overflow: hidden;
}

/* ── Ornament ── */
.card-ornament {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 80px;
  position: relative;
  background: linear-gradient(180deg, rgba(212, 168, 75, 0.06), transparent);
  border-bottom: 1px solid rgba(212, 168, 75, 0.06);
}

.ornament-ring {
  position: absolute;
  width: 50px;
  height: 50px;
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
  font-size: 2rem;
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

/* ── Inner ── */
.card-inner {
  padding: 36px 40px 40px;
}

/* ── Header ── */
.card-header {
  text-align: center;
  margin-bottom: 36px;
}

.header-eyebrow {
  font-size: 10px;
  letter-spacing: 3px;
  color: rgba(212, 168, 75, 0.4);
  text-transform: uppercase;
  margin-bottom: 8px;
}

.header-title {
  font-family: var(--font-serif), serif;
  font-size: 2rem;
  font-weight: 700;
  color: var(--text);
  margin: 0 0 8px;
  letter-spacing: 4px;
}

.header-sub {
  font-size: 13px;
  color: var(--muted);
  margin: 0;
}

/* ── Form ── */
.login-form {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.form-field {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.field-label {
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 1px;
  color: rgba(212, 168, 75, 0.5);
  text-transform: uppercase;
}

.field-input-wrap {
  position: relative;
  display: flex;
  align-items: center;
}

.field-icon {
  position: absolute;
  left: 14px;
  font-size: 12px;
  color: rgba(212, 168, 75, 0.3);
  pointer-events: none;
  z-index: 1;
}

.field-input {
  width: 100%;
  padding: 13px 44px;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(212, 168, 75, 0.1);
  border-radius: 10px;
  color: var(--text);
  font-size: 14px;
  font-family: var(--font-sans);
  outline: none;
  transition: all 0.3s ease;
}

.field-input:focus {
  border-color: rgba(212, 168, 75, 0.4);
  box-shadow: 0 0 0 3px rgba(212, 168, 75, 0.06);
  background: rgba(255, 255, 255, 0.06);
}

.field-input::placeholder {
  color: rgba(255, 255, 255, 0.1);
}

.toggle-password {
  position: absolute;
  right: 14px;
  background: none;
  border: none;
  color: rgba(212, 168, 75, 0.3);
  cursor: pointer;
  font-size: 14px;
  padding: 0;
  transition: color 0.2s;
}

.toggle-password:hover {
  color: rgba(212, 168, 75, 0.6);
}

/* ── Error ── */
.error-msg {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 14px;
  background: rgba(196, 30, 58, 0.08);
  border: 1px solid rgba(196, 30, 58, 0.2);
  border-radius: 8px;
  font-size: 13px;
  color: #e05a5a;
}

.error-icon {
  font-size: 12px;
}

/* ── Submit button ── */
.btn-submit {
  width: 100%;
  padding: 14px;
  margin-top: 8px;
  background: linear-gradient(135deg, #d4a84b, #b8860b);
  color: #0a0815;
  font-weight: 700;
  font-size: 15px;
  letter-spacing: 3px;
  border: none;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.3s ease;
  box-shadow: 0 4px 20px rgba(212, 168, 75, 0.25);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.btn-submit:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 8px 32px rgba(212, 168, 75, 0.4);
}

.btn-submit:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.loading-spinner {
  display: inline-block;
  width: 18px;
  height: 18px;
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

/* ── Footer ── */
.card-footer {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  margin-top: 28px;
  padding-top: 24px;
  border-top: 1px solid rgba(212, 168, 75, 0.06);
}

.footer-text {
  font-size: 13px;
  color: var(--muted);
}

.footer-link {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  font-weight: 600;
  color: var(--gold);
  text-decoration: none;
  transition: all 0.2s ease;
}

.footer-link:hover {
  text-shadow: 0 0 12px rgba(212, 168, 75, 0.4);
}

.footer-arrow {
  font-size: 11px;
  transition: transform 0.2s ease;
}

.footer-link:hover .footer-arrow {
  transform: translateX(3px);
}
</style>
