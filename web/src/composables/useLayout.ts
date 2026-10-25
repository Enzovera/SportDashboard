import { onScopeDispose, ref } from 'vue'

/**
 * fold  = Fold6 inner screen (707 px) or wider: two columns split by the hinge gap
 * cover = outer screen (~369 px): one column
 */
export type Layout = 'fold' | 'cover'

export const COVER_MAX_WIDTH = 599

export function useLayout() {
  const query = window.matchMedia(`(max-width: ${COVER_MAX_WIDTH}px)`)
  const layout = ref<Layout>(query.matches ? 'cover' : 'fold')
  const update = () => (layout.value = query.matches ? 'cover' : 'fold')
  query.addEventListener('change', update)
  onScopeDispose(() => query.removeEventListener('change', update))
  return layout
}
