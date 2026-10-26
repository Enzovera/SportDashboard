import type { BodyResponse, BodyWeek } from '../body'

const WEEK_STARTS = [
  '2026-08-03', '2026-08-10', '2026-08-17', '2026-08-24', '2026-08-31', '2026-09-07',
  '2026-09-14', '2026-09-21', '2026-09-28', '2026-10-05', '2026-10-12', '2026-10-19',
]

// [avg, min, max, count] per week, read off design/screens/DesktopCialo.html.
const FULL_WEEKS: [number, number, number, number][] = [
  [78.1, 77.5, 78.7, 7], [78.0, 77.4, 78.6, 6], [77.8, 77.2, 78.4, 7], [77.9, 77.3, 78.5, 6],
  [77.6, 77.0, 78.2, 7], [77.4, 76.8, 78.0, 7], [77.3, 76.7, 77.9, 6], [77.1, 76.6, 77.8, 7],
  [77.0, 76.4, 77.6, 7], [76.8, 76.2, 77.4, 6], [76.7, 76.4, 77.0, 7], [76.4, 76.1, 76.9, 6],
]

/** Values from design/screens/FoldCialo.html and DesktopCialo.html. */
export const bodyFull: BodyResponse = {
  date: '2026-10-25',
  from: '2026-08-03',
  to: '2026-10-25',
  rangeWeeks: 12,
  device: 'Xiaomi',
  sync: [{ source: 'garmin', status: 'ok', lastSuccessAt: '2026-10-25T07:12:00+01:00' }],
  composition: {
    measuredAt: '2026-10-25T07:05:00+01:00',
    lastKg: 76.2,
    avg7dKg: 76.4,
    avg7dChangeKg: -1.7,
    fatPct: 15.8,
    muscleKg: 61.3,
    muscleTrend: 'stable',
    weekCount: 6,
    weekDays: 7,
    weekMissing: ['2026-10-21'],
    rangeCount: 79,
  },
  trendKgPerWeek: -0.15,
  weeks: WEEK_STARTS.map((weekStart, i): BodyWeek => {
    const [avgKg, minKg, maxKg, count] = FULL_WEEKS[i]!
    return { weekStart, avgKg, minKg, maxKg, count }
  }),
  weighIns: [
    { measuredAt: '2026-10-25T07:05:00+01:00', weightKg: 76.2, fatPct: 15.7, muscleKg: 61.3, waterPct: 58.5 },
    { measuredAt: '2026-10-24T07:20:00+02:00', weightKg: 76.6, fatPct: 15.9, muscleKg: 61.2, waterPct: 58.3 },
    { measuredAt: '2026-10-23T06:58:00+02:00', weightKg: 76.3, fatPct: 15.8, muscleKg: 61.3, waterPct: 58.4 },
    { measuredAt: '2026-10-22T07:02:00+02:00', weightKg: 76.9, fatPct: 16.0, muscleKg: 61.2, waterPct: 58.2 },
    { measuredAt: '2026-10-20T07:10:00+02:00', weightKg: 76.4, fatPct: 15.8, muscleKg: 61.3, waterPct: 58.4 },
    { measuredAt: '2026-10-19T07:04:00+02:00', weightKg: 76.1, fatPct: 15.7, muscleKg: 61.4, waterPct: 58.5 },
    { measuredAt: '2026-10-18T08:15:00+02:00', weightKg: 76.5, fatPct: 15.8, muscleKg: 61.3, waterPct: 58.4 },
  ],
  goal: null,
  coachNote: {
    date: '2026-10-12',
    kind: 'rule',
    text:
      'Patrzę na średnią 7 dni — pojedyncze ważenie waha się nawet o 0,8 kg, a skład ciała z wagi to tylko szacunek. ' +
      'Dopóki nie ustawisz celu wagi, nie zmieniam kalorii pod redukcję ani masę.',
  },
}

// Stage 1: weight only through Garmin, entered by hand now and then (3 weigh-ins in 90 days), no composition,
// no profile and no coach. The view shows the few real points instead of an empty trend.
const SPARSE: Record<string, number> = { '2026-08-17': 78.4, '2026-09-14': 77.6, '2026-10-12': 77.0 }

export const bodyStage1: BodyResponse = {
  ...bodyFull,
  device: null,
  composition: {
    measuredAt: '2026-10-14T07:40:00+02:00',
    lastKg: 77.0,
    avg7dKg: null,
    avg7dChangeKg: null,
    fatPct: null,
    muscleKg: null,
    muscleTrend: null,
    weekCount: 0,
    weekDays: 7,
    weekMissing: ['2026-10-19', '2026-10-20', '2026-10-21', '2026-10-22', '2026-10-23', '2026-10-24', '2026-10-25'],
    rangeCount: 3,
  },
  trendKgPerWeek: null,
  weeks: WEEK_STARTS.map((weekStart): BodyWeek => {
    const kg = SPARSE[weekStart] ?? null
    return { weekStart, avgKg: kg, minKg: kg, maxKg: kg, count: kg === null ? 0 : 1 }
  }),
  weighIns: [
    { measuredAt: '2026-10-14T07:40:00+02:00', weightKg: 77.0, fatPct: null, muscleKg: null, waterPct: null },
    { measuredAt: '2026-09-16T08:05:00+02:00', weightKg: 77.6, fatPct: null, muscleKg: null, waterPct: null },
    { measuredAt: '2026-08-21T07:55:00+02:00', weightKg: 78.4, fatPct: null, muscleKg: null, waterPct: null },
  ],
  goal: null,
  coachNote: null,
}
