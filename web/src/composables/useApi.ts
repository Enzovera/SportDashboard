import { ref, shallowRef, watch, type Ref, type WatchSource } from 'vue'
import { syncGeneration } from '@/api/sync'

/** Loads data on mount, whenever `deps` change and after an on-demand sync; keeps the previous data while reloading. */
export function useApi<T>(load: () => Promise<T>, deps: WatchSource[] = []) {
  const data: Ref<T | null> = shallowRef(null)
  const error = ref<string | null>(null)
  const loading = ref(false)

  async function reload() {
    loading.value = true
    error.value = null
    try {
      data.value = await load()
    } catch (e) {
      error.value = e instanceof Error ? e.message : String(e)
    } finally {
      loading.value = false
    }
  }

  watch([...deps, syncGeneration], reload, { immediate: true })
  return { data, error, loading, reload }
}
