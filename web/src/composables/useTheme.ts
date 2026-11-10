import { ref } from 'vue'
import { applyTheme, nextTheme, readTheme, type ThemeMode } from '@/lib/theme'

// Module-level: the toggle appears in both headers, they share one state.
const mode = ref<ThemeMode>(readTheme())

export function useTheme() {
  const cycle = () => {
    mode.value = nextTheme(mode.value)
    applyTheme(mode.value)
  }
  return { mode, cycle }
}
