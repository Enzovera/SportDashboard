import type { FormState, Source, SyncStatus, Tone } from '@/api/types'
import { localDate, time } from './format'

export const SOURCE_LABEL: Record<Source, string> = {
  garmin: 'Garmin',
  hevy: 'Hevy',
  fitatu: 'Fitatu',
  hevy_garmin: 'Hevy → Garmin',
}

export const FORM_LABEL: Record<FormState, string> = {
  fresh: 'świeży',
  neutral: 'neutralnie',
  tired: 'zmęczony',
  overreached: 'przeciążony',
}

export const FORM_TONE: Record<FormState, Tone> = {
  fresh: 'good',
  neutral: 'neutral',
  tired: 'warn',
  overreached: 'warn',
}

/** Hevy muscle group keys -> Polish labels used in the UI. */
export const MUSCLE_LABEL: Record<string, string> = {
  abdominals: 'Brzuch',
  abductors: 'Odwodziciele',
  adductors: 'Przywodziciele',
  biceps: 'Biceps',
  calves: 'Łydki',
  cardio: 'Cardio',
  chest: 'Klatka',
  forearms: 'Przedramiona',
  full_body: 'Całe ciało',
  glutes: 'Pośladki',
  hamstrings: 'Dwugłowe uda',
  lats: 'Najszersze',
  lower_back: 'Dolny grzbiet',
  neck: 'Szyja',
  quadriceps: 'Czworogłowe',
  shoulders: 'Barki',
  traps: 'Kaptury',
  triceps: 'Triceps',
  upper_back: 'Plecy',
  other: 'Inne',
}

export function muscleLabel(key: string): string {
  return MUSCLE_LABEL[key] ?? key
}

/**
 * Short problem text for a source, or null when sync is fine.
 * error/partial -> "błąd sync"; stale -> "sync: wcz. 18:02" / "sync: 18:02" / "sync: 23.10".
 */
export function syncProblem(s: SyncStatus, now: Date = new Date()): string | null {
  if (s.status === 'ok') return null
  if (s.status === 'error') return 'błąd sync'
  if (s.status === 'partial') return 'sync częściowy'
  if (!s.lastSuccessAt) return 'brak sync'
  const today = localDate(now)
  const day = localDate(s.lastSuccessAt)
  const yesterday = localDate(new Date(now.getTime() - 86_400_000))
  if (day === today) return `sync: ${time(s.lastSuccessAt)}`
  if (day === yesterday) return `sync: wcz. ${time(s.lastSuccessAt)}`
  return `sync: ${day.slice(8, 10)}.${day.slice(5, 7)}`
}
