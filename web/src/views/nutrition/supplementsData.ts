// Pure helpers of the Supplements view: grouping, labels, optimistic updates. No DOM, unit-tested.
import type { DayType } from '@/api/nutrition'
import type {
  DoseUnit,
  SupplementItem,
  SupplementLogStatus,
  SupplementRegularity,
  SupplementSlot,
  SupplementStock,
  SupplementWhen,
} from '@/api/supplements'
import { dayShort, num } from '@/lib/format'

export interface SlotGroup {
  key: 'morning' | 'session' | 'after' | 'evening'
  label: string
  items: SupplementItem[]
}

const GROUP_OF: Record<SupplementSlot, SlotGroup['key']> = {
  morning: 'morning',
  before: 'session',
  during: 'session',
  after: 'after',
  evening: 'evening',
}

const GROUP_ORDER: SlotGroup['key'][] = ['morning', 'session', 'after', 'evening']

function isRunDay(dayType: DayType | null): boolean {
  return dayType === 'run' || dayType === 'long_run' || dayType === 'strength_run'
}

/** Checklist in time-of-day groups: Rano, Bieg/Trening (before + during), Po biegu/treningu, Wieczór. */
export function slotGroups(items: SupplementItem[], dayType: DayType | null): SlotGroup[] {
  const run = isRunDay(dayType)
  const label: Record<SlotGroup['key'], string> = {
    morning: 'Rano',
    session: run ? 'Bieg' : 'Trening',
    after: run ? 'Po biegu' : 'Po treningu',
    evening: 'Wieczór',
  }
  return GROUP_ORDER.map((key) => ({ key, label: label[key], items: items.filter((i) => GROUP_OF[i.slot] === key) })).filter(
    (g) => g.items.length > 0,
  )
}

const SLOT_WORD: Record<SupplementSlot, string> = {
  morning: 'rano',
  before: 'przed',
  during: 'w trakcie',
  after: 'po',
  evening: 'wieczorem',
}

/** Right-hand word of a checklist row: status once logged, otherwise reminder time or slot. */
export function whenLabel(item: SupplementItem): string {
  if (item.status === 'taken') return 'przyjęte'
  if (item.status === 'skipped') return 'pominięte'
  return item.remindAt ?? SLOT_WORD[item.slot]
}

const UNIT_SHORT: Record<DoseUnit, string> = {
  IU: 'IU',
  mg: 'mg',
  g: 'g',
  capsule: 'kaps.',
  tablet: 'tabl.',
  sachet: 'sasz.',
}

/** { 2000, IU } -> "2 000 IU", { 1, sachet } -> "1 saszetka", { 2, capsule } -> "2 kaps.". */
export function doseLabel(dose: { amount: number; unit: DoseUnit }): string {
  if (dose.unit === 'sachet' && dose.amount === 1) return '1 saszetka'
  return `${num(dose.amount)} ${UNIT_SHORT[dose.unit]}`
}

/** "2000 IU · z posiłkiem"; whey: "30 g · liczy się w Fitatu". */
export function itemSub(item: SupplementItem): string {
  const note = item.countedInFitatu ? 'liczy się w Fitatu' : item.note
  return [doseLabel(item.dose), note].filter(Boolean).join(' · ')
}

export function counts(items: SupplementItem[]): { taken: number; total: number } {
  return { taken: items.filter((i) => i.status === 'taken').length, total: items.length }
}

/** One tap: tick = taken; unticking a taken item logs it as skipped (the latest event of the day wins). */
export function nextStatus(item: SupplementItem): SupplementLogStatus {
  return item.status === 'taken' ? 'skipped' : 'taken'
}

/** Copy of the list with one item's status changed (optimistic update before the server answers). */
export function withStatus(items: SupplementItem[], id: string, status: SupplementItem['status']): SupplementItem[] {
  return items.map((i) => (i.id === id ? { ...i, status } : i))
}

export function replaceItem(items: SupplementItem[], item: SupplementItem): SupplementItem[] {
  return items.map((i) => (i.id === item.id ? item : i))
}

const WHEN_SUFFIX: Record<SupplementWhen, string> = {
  daily: '',
  run_day: 'bieg.',
  long_run: 'długie',
  quality_run: 'jakość.',
  strength_day: 'siłowe',
  training_day: 'tren.',
}

/** "29 / 30", "4 / 4 długie", "22 / 26 tren.". */
export function regularityDays(r: SupplementRegularity): string {
  const suffix = WHEN_SUFFIX[r.when]
  return `${num(r.taken)} / ${num(r.applicable)}${suffix ? ' ' + suffix : ''}`
}

/** "12 kaps.", "6 sasz.", "205 g". */
export function stockLeft(s: SupplementStock): string {
  return `${num(s.left.amount)} ${UNIT_SHORT[s.left.unit]}`
}

const LASTS_UNIT: Record<SupplementWhen, string> = {
  daily: 'dni',
  run_day: 'dni bieg.',
  long_run: 'długich',
  quality_run: 'jakość.',
  strength_day: 'dni siłowych',
  training_day: 'tren.',
}

/** "~12 dni · zamów", "~6 długich". */
export function stockLasts(s: SupplementStock): string {
  const base = `~${num(s.lasts.count)} ${LASTS_UNIT[s.lasts.per]}`
  return s.reorder ? `${base} · zamów` : base
}

/** "Suplementy · Nd 25.10". */
export function supplementsMeta(date: string): string {
  return `Suplementy · ${dayShort(date)}`
}

/** Square strip of a regularity row, 13 px per day as in the desktop mockup (viewBox width = 13·n − 3). */
export function stripCells(days: SupplementRegularity['days']): { x: number; kind: SupplementRegularity['days'][number] }[] {
  return days.map((kind, i) => ({ x: i * 13, kind }))
}
