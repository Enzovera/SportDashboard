import { ref, shallowRef, watch, type Ref, type WatchSource } from 'vue'

/** Loads data on mount and whenever `deps` change; keeps the previous data while reloading. */
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

  watch(deps, reload, { immediate: true })
  return { data, error, loading, reload }
}
