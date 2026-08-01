import { defineStore } from 'pinia'
import { ref } from 'vue'
import { wuxingThemes, type WuxingKey } from '../composables/useWuxingThemes'

const STORAGE_KEY = 'bazi_element_theme'
const VALID_KEYS: WuxingKey[] = ['mu', 'huo', 'tu', 'jin', 'shui']
const LIGHT_ACCENTS: Record<WuxingKey, string> = {
  mu: '#047857',
  huo: '#be123c',
  tu: '#92400e',
  jin: '#475569',
  shui: '#0369a1',
}

function hexToRgb(value: string): string {
  const hex = value.replace('#', '')
  return [0, 2, 4].map((offset) => Number.parseInt(hex.slice(offset, offset + 2), 16)).join(', ')
}

function readStoredKey(): WuxingKey {
  const raw = localStorage.getItem(STORAGE_KEY)
  return raw && (VALID_KEYS as string[]).includes(raw) ? (raw as WuxingKey) : 'mu'
}

/**
 * Apply a wuxing theme to <html> by setting CSS custom properties.
 * Called from the store and from the inline <head> bootstrap to avoid FOUC.
 */
export function applyWuxingToRoot(key: WuxingKey) {
  const t = wuxingThemes[key]
  const root = document.documentElement
  const isDark = root.classList.contains('dark')
  const activeAccent = isDark ? t.accentHex : LIGHT_ACCENTS[key]
  const activeRgb = isDark ? t.accentRgb : hexToRgb(activeAccent)

  // 1. HomeView local tokens (kept for backward compat)
  root.style.setProperty('--jade-accent', t.accentHex)
  root.style.setProperty('--jade-accent-dark', t.accentDark)
  root.style.setProperty('--jade-accent-rgb', t.accentRgb)
  root.style.setProperty('--jade-button-text', t.buttonText)

  // 2. Global accent surface (--accent / --accent-dim / --accent-glow)
  root.style.setProperty('--accent', activeAccent)
  root.style.setProperty('--accent-dim', `rgba(${activeRgb}, 0.14)`)
  root.style.setProperty('--accent-glow', `rgba(${activeRgb}, 0.22)`)

  // 3. Focus ring & brand glow used in App.vue logo + buttons
  root.style.setProperty('--line-focus', activeAccent)
  root.style.setProperty('--brand-glow', `rgba(${activeRgb}, 0.42)`)
  root.style.setProperty('--menu-hover', `rgba(${activeRgb}, 0.10)`)

  // 4. shadcn primary token (used by status-dot, charts, sidebars, etc.)
  root.style.setProperty('--primary', isDark ? t.primaryOklchDark : t.primaryOklchLight)
  root.style.setProperty('--ring', activeAccent)

  // 5. Ambient glow used by body::before
  root.style.setProperty('--glow-primary', `rgba(${activeRgb}, ${isDark ? 0.14 : 0.1})`)

  // 6. Selection highlight (CSS already uses --accent-dim, no extra work needed)

  // 7. Data attribute so CSS can use [data-wuxing="huo"] selectors
  root.dataset.wuxing = key
}

export const useThemeStore = defineStore('theme', () => {
  const elementTheme = ref<WuxingKey>(readStoredKey())

  function setElementTheme(key: WuxingKey) {
    if (elementTheme.value === key) return
    elementTheme.value = key
    localStorage.setItem(STORAGE_KEY, key)
    applyWuxingToRoot(key)
  }

  /** Re-apply current theme — call after dark/light toggle so --primary picks up the new mode. */
  function refresh() {
    applyWuxingToRoot(elementTheme.value)
  }

  // Initial sync (the inline head script handles FOUC; this is the SPA-side source of truth)
  applyWuxingToRoot(elementTheme.value)

  // Cross-tab sync
  window.addEventListener('storage', (e) => {
    if (e.key === STORAGE_KEY && e.newValue && (VALID_KEYS as string[]).includes(e.newValue)) {
      elementTheme.value = e.newValue as WuxingKey
      applyWuxingToRoot(elementTheme.value)
    }
  })

  // Re-apply --primary when dark/light class flips
  const observer = new MutationObserver(() => {
    applyWuxingToRoot(elementTheme.value)
  })
  observer.observe(document.documentElement, { attributes: true, attributeFilter: ['class'] })

  return { elementTheme, setElementTheme, refresh }
})
