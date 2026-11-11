import type { SecretName, SecretStatus } from '@/api/settings'
import { dayMonth, localDate } from '@/lib/format'

export interface SecretSpec {
  name: SecretName
  title: string
  /** Who uses it; shown as the box meta when nothing is stored. */
  usedBy: string
  minLength?: number
  generate?: boolean
}

/** Display order and grouping: sources left, coach right (Fold6 3 | 3 halves). */
export const SOURCE_SECRETS: SecretSpec[] = [
  { name: 'HEVY_API_KEY', title: 'Hevy · klucz API', usedBy: 'ingest' },
  { name: 'FORMA_MCP_TOKEN', title: 'Token MCP', usedBy: 'serwer MCP i czat', minLength: 32, generate: true },
]
export const COACH_SECRETS: SecretSpec[] = [
  { name: 'CLAUDE_CODE_OAUTH_TOKEN', title: 'Claude · token subskrypcji', usedBy: 'czat i runy' },
  { name: 'ANTHROPIC_API_KEY', title: 'Claude · klucz API', usedBy: 'czat i runy' },
]

/** "…4f2a · 02.11" when stored in the app, "zmienna środowiskowa" when the services use the env (or nothing). */
export function secretMeta(s: SecretStatus | undefined): string {
  if (!s?.set) return 'zmienna środowiskowa'
  const parts = [s.hint ? `…${s.hint}` : 'ustawiony']
  if (s.updatedAt) parts.push(dayMonth(localDate(s.updatedAt)))
  return parts.join(' · ')
}

/** Client-side check mirroring the API (secrets.Validate); the API stays the authority. */
export function secretProblem(spec: SecretSpec, value: string): string | null {
  if (!value) return 'Puste pole.'
  if (/\s/.test(value)) return 'Wartość zawiera spacje lub nową linię.'
  if (value.length > 4096) return 'Za długa wartość.'
  if (spec.minLength && value.length < spec.minLength) return `Co najmniej ${spec.minLength} znaków.`
  return null
}
