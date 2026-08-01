import type { Directive } from 'vue'

/**
 * v-reveal — 一次性进入视口动效（fade + translateY(12px)）。
 *
 * 用法：`v-reveal` 或 `v-reveal="40"`（binding 为延迟毫秒，建议区间 stagger ≤ 60ms）。
 * - 仅使用 transform/opacity，不造成布局位移（CLS = 0）。
 * - 尊重 prefers-reduced-motion：reduced 时不加任何过渡，直接呈现。
 * - 一次性：进入视口后断开观察，不重复播放。
 */

const DURATION = 480 // ms，reveal 上限 500ms
const MAX_DELAY = 240 // ms，防误传超大延迟

function prefersReducedMotion(): boolean {
  return (
    typeof window !== 'undefined' &&
    typeof window.matchMedia === 'function' &&
    window.matchMedia('(prefers-reduced-motion: reduce)').matches
  )
}

const observerMap = new WeakMap<HTMLElement, IntersectionObserver>()

export const vReveal: Directive<HTMLElement, number | undefined> = {
  mounted(el, binding) {
    if (prefersReducedMotion()) return
    const delay = typeof binding.value === 'number' ? Math.min(Math.max(binding.value, 0), MAX_DELAY) : 0

    el.style.opacity = '0'
    el.style.transform = 'translateY(12px)'
    el.style.transition = `opacity ${DURATION}ms ease, transform ${DURATION}ms ease`
    if (delay) el.style.transitionDelay = `${delay}ms`
    el.style.willChange = 'opacity, transform'

    const io = new IntersectionObserver(
      (entries) => {
        for (const entry of entries) {
          if (entry.isIntersecting) {
            el.style.opacity = '1'
            el.style.transform = 'translateY(0)'
            io.disconnect()
            observerMap.delete(el)
            // 播放结束后清理 inline 样式，避免影响后续交互态（如 hover transform）
            window.setTimeout(() => {
              el.style.willChange = ''
              el.style.transition = ''
              el.style.transitionDelay = ''
              el.style.transform = ''
            }, DURATION + delay + 50)
            break
          }
        }
      },
      { threshold: 0.06, rootMargin: '0px 0px -6% 0px' },
    )
    io.observe(el)
    observerMap.set(el, io)
  },
  unmounted(el) {
    observerMap.get(el)?.disconnect()
    observerMap.delete(el)
  },
}
