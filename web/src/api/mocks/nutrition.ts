import type { DayType, NutritionDay, NutritionResponse } from '../nutrition'

// Values from design/screens/FoldZywienie.html and DesktopZywienie.html (12–25.10, today = long run).
const DAYS: [string, DayType, number, number, number | null, number | null][] = [
  // date, day type, eaten kcal, target kcal, vs target %, protein g/kg
  ['2026-10-12', 'strength_run', 2620, 2750, -5, 1.62],
  ['2026-10-13', 'run', 2480, 2700, -8, 1.71],
  ['2026-10-14', 'strength', 2390, 2600, -8, 1.85],
  ['2026-10-15', 'rest', 2210, 2300, -4, 1.58],
  ['2026-10-16', 'strength_run', 2700, 2750, -2, 1.92],
  ['2026-10-17', 'run', 2550, 2700, -6, 1.74],
  ['2026-10-18', 'long_run', 2610, 2950, -12, 1.66],
  ['2026-10-19', 'strength_run', 2580, 2750, -6, 1.8],
  ['2026-10-20', 'run', 2720, 2700, 1, 1.95],
  ['2026-10-21', 'strength', 2450, 2600, -6, 1.7],
  ['2026-10-22', 'rest', 2280, 2300, -1, 1.55],
  ['2026-10-23', 'strength_run', 2660, 2750, -3, 1.88],
  ['2026-10-24', 'run', 2610, 2700, -3, 1.53],
  ['2026-10-25', 'long_run', 820, 2950, null, null],
]

const days: NutritionDay[] = DAYS.map(([date, dayType, kcal, kcalTarget, pct, protein], i) => {
  const complete = i < DAYS.length - 1
  return {
    date,
    dayType,
    complete,
    kcal,
    kcalTarget,
    kcalVsTargetPct: complete ? pct : null,
    kcalTone: complete && dayType === 'long_run' && (pct ?? 0) < 0 ? 'warn' : 'neutral',
    proteinGPerKg: protein,
    proteinTone: protein !== null && protein < 1.6 ? 'warn' : 'neutral',
  }
})

export const nutritionFull: NutritionResponse = {
  range: '14d',
  from: '2026-10-12',
  to: '2026-10-25',
  sync: [{ source: 'fitatu', status: 'ok', lastSuccessAt: '2026-10-25T08:15:00+01:00' }],
  today: {
    dayType: 'long_run',
    kcal: { eaten: 820, target: 2950 },
    proteinG: { eaten: 38, target: 145 },
    carbsG: { eaten: 112, target: 412 },
    fatG: { eaten: 24, target: 80 },
  },
  balance: {
    from: '2026-10-12',
    to: '2026-10-24',
    kcalAvg: 2528,
    kcalTargetAvg: 2658,
    kcalVsTargetPct: -5,
    proteinAvgGPerKg: 1.73,
    proteinTargetGPerKg: 1.9,
    proteinMinGPerKg: 1.6,
    proteinDaysAtMin: 10,
    days: 13,
    longRunCarbs: { date: '2026-10-18', gPerKg: 4.8, targetGPerKg: 5.4, tone: 'warn' },
  },
  days,
  targets: [
    { dayType: 'rest', kcal: 2300, proteinG: 145, carbsG: 255, fatG: 80 },
    { dayType: 'strength', kcal: 2600, proteinG: 145, carbsG: 330, fatG: 80 },
    { dayType: 'run', kcal: 2700, proteinG: 145, carbsG: 355, fatG: 80 },
    { dayType: 'strength_run', kcal: 2750, proteinG: 145, carbsG: 370, fatG: 80 },
    { dayType: 'long_run', kcal: 2950, proteinG: 145, carbsG: 412, fatG: 80 },
  ],
  targetProposals: [{ id: 'p-carbs-long', dayType: 'long_run', field: 'carbsG', value: 460 }],
  meals: [
    {
      id: 'm1',
      slot: 'breakfast',
      at: '2026-10-25T07:40:00+01:00',
      title: 'Owsianka z bananem i odżywką białkową',
      kcal: 820,
      proteinG: 38,
      carbsG: 112,
      fatG: 24,
    },
  ],
  coachNote: {
    at: '2026-10-25T06:31:00+01:00',
    text: 'przed długim wybieganiem jeszcze ok. 60 g węglowodanów, po biegu posiłek z 40 g białka.',
  },
  supplements: { taken: 2, total: 7, next: 'Kofeina' },
}

/** Before stage 2: Fitatu is not integrated and there is no plan, so nothing to show. */
export const nutritionStage1: NutritionResponse = {
  range: '14d',
  from: '2026-10-12',
  to: '2026-10-25',
  sync: [],
  today: null,
  balance: null,
  days: [],
  targets: [],
  targetProposals: [],
  meals: [],
  coachNote: null,
  supplements: null,
}
