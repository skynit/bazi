import { onBeforeUnmount, onMounted, ref, watch, type Ref } from 'vue'

/**
 * useScrollSpy — 长滚动分区的当前位置追踪。
 * 以"分区顶部越过视口顶部偏移线的最后一个分区"为当前分区，随滚动 rAF 节流更新；
 * scrollTo 点击时立即高亮并平滑滚动（尊重 prefers-reduced-motion）。
 */
export function useScrollSpy(sectionIds: Ref<string[]>, options?: { offsetTop?: number }) {
  const activeId = ref('')
  const offsetTop = options?.offsetTop ?? 96
  let ticking = false

  function update() {
    ticking = false
    const ids = sectionIds.value
    if (!ids.length) {
      activeId.value = ''
      return
    }
    // 接近页面底部时，末尾分区可能永远越不过顶部偏移线，直接高亮末项
    if (window.innerHeight + window.scrollY >= document.documentElement.scrollHeight - 4) {
      activeId.value = ids[ids.length - 1]
      return
    }
    let current = ids[0]
    for (const id of ids) {
      const el = document.getElementById(id)
      if (el && el.getBoundingClientRect().top <= offsetTop + 8) current = id
    }
    activeId.value = current
  }

  function onScroll() {
    if (ticking) return
    ticking = true
    window.requestAnimationFrame(update)
  }

  function prefersReducedMotion(): boolean {
    return (
      typeof window.matchMedia === 'function' &&
      window.matchMedia('(prefers-reduced-motion: reduce)').matches
    )
  }

  function scrollTo(id: string) {
    const el = document.getElementById(id)
    if (!el) return
    activeId.value = id
    el.scrollIntoView({ behavior: prefersReducedMotion() ? 'auto' : 'smooth', block: 'start' })
  }

  onMounted(() => {
    update()
    window.addEventListener('scroll', onScroll, { passive: true })
    window.addEventListener('resize', onScroll, { passive: true })
  })

  watch(sectionIds, () => update(), { flush: 'post' })

  onBeforeUnmount(() => {
    window.removeEventListener('scroll', onScroll)
    window.removeEventListener('resize', onScroll)
  })

  return { activeId, scrollTo }
}
