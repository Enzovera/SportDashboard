import { onScopeDispose, ref, watch, type Ref } from 'vue'

/**
 * Live content width of an element (0 until it is mounted). For layout choices that CSS container queries
 * cannot make because they switch component props, e.g. a wide table vs its compact variant.
 */
export function useElementWidth(el: Ref<HTMLElement | null>): Ref<number> {
  const width = ref(0)
  const observer = new ResizeObserver((entries) => {
    const w = entries[0]?.contentRect.width
    if (w !== undefined) width.value = w
  })
  watch(
    el,
    (now, before) => {
      if (before) observer.unobserve(before)
      if (now) {
        width.value = now.getBoundingClientRect().width
        observer.observe(now)
      }
    },
    { immediate: true, flush: 'post' },
  )
  onScopeDispose(() => observer.disconnect())
  return width
}
