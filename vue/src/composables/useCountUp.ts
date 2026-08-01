import { ref, watch, onUnmounted, type Ref } from 'vue'

/**
 * useCountUp — 数字首次进入视口时从 0 缓动到目标值（ease-out）。
 *
 * - 默认 620ms ease-out，仅首次进入视口播放一次。
 * - 尊重 prefers-reduced-motion：reduced 时直接显示终值。
 * - 目标值变化时（非首次）直接跳到新值，不重复播放。
 */
export function useCountUp(
  target: Ref<number>,
  host: Ref<HTMLElement | null | undefined>,
  options?: { duration?: number; decimals?: number },
) {
  const duration = Math.min(Math.max(options?.duration ?? 620, 150), 800)
  const decimals = options?.decimals ?? 1
  const display = ref((0).toFixed(decimals))
  const started = ref(false)
  let raf = 0
  let io: IntersectionObserver | null = null

  const reduced =
    typeof window !== 'undefined' &&
    typeof window.matchMedia === 'function' &&
    window.matchMedia('(prefers-reduced-motion: reduce)').matches

  function finish() {
    display.value = target.value.toFixed(decimals)
  }

  function play() {
    if (started.value) return
    started.value = true
    if (reduced || duration <= 0) {
      finish()
      return
    }
    const from = 0
    const to = target.value
    const t0 = performance.now()
    const tick = (now: number) => {
      const p = Math.min(1, (now - t0) / duration)
      const eased = 1 - Math.pow(1 - p, 3) // easeOutCubic
      display.value = (from + (to - from) * eased).toFixed(decimals)
      if (p < 1) raf = requestAnimationFrame(tick)
      else finish()
    }
    raf = requestAnimationFrame(tick)
  }

  function observe(el: HTMLElement | null | undefined) {
    io?.disconnect()
    io = null
    if (!el) return
    if (reduced) {
      started.value = true
      finish()
      return
    }
    io = new IntersectionObserver(
      (entries) => {
        if (entries.some((e) => e.isIntersecting)) {
          io?.disconnect()
          io = null
          play()
        }
      },
      { threshold: 0.2 },
    )
    io.observe(el)
  }

  watch(
    host,
    (el) => {
      if (el) observe(el)
    },
    { immediate: true, flush: 'post' },
  )

  // 播放完成后目标值再变化：直接同步，不再动画
  watch(target, () => {
    if (started.value) finish()
  })

  onUnmounted(() => {
    cancelAnimationFrame(raf)
    io?.disconnect()
  })

  return { display }
}
