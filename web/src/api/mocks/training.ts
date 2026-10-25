import type { TrainingResponse } from '../types'

const weekStarts = [
  '2026-08-03', '2026-08-10', '2026-08-17', '2026-08-24', '2026-08-31', '2026-09-07',
  '2026-09-14', '2026-09-21', '2026-09-28', '2026-10-05', '2026-10-12', '2026-10-19',
]
const km = [31, 34, 36, 33, 37, 39, 41, 36, 40, 42, 43, 30]
const planned = [32, 34, 37, 32, 38, 40, 42, 36, 40, 42, 44, 46]
const ctl = [30, 31, 32, 32, 33, 34, 35, 35, 36, 38, 39, 41]
const atl = [33, 36, 37, 29, 38, 40, 42, 31, 40, 43, 42, 37]

/** Values from design/screens/FoldTrening.html. */
export const trainingFull: TrainingResponse = {
  range: '12w',
  from: '2026-08-03',
  to: '2026-10-25',
  sync: [
    { source: 'garmin', status: 'ok', lastSuccessAt: '2026-10-25T06:12:00+01:00' },
    { source: 'hevy', status: 'error', lastSuccessAt: '2026-10-24T18:02:00+02:00' },
  ],
  load: {
    weekFrom: '2026-10-19',
    weekTo: '2026-10-25',
    weekKm: 46,
    prevWeekKm: 43,
    ctl: 41,
    ctlDeltaRange: 11,
    atl: 37,
    atlDeltaWeek: -5,
    tsb: 4,
    form: 'fresh',
  },
  weeklyKm: weekStarts.map((weekStart, i) => ({
    weekStart,
    km: km[i]!,
    plannedKm: planned[i]!,
    deload: i === 3 || i === 7,
  })),
  todayPlannedKm: 16,
  loadSeries: weekStarts.map((date, i) => ({ date, ctl: ctl[i]!, atl: atl[i]! })),
  coachNote: {
    date: '2026-10-18',
    text: 'Legs w środę daje 4 dni regeneracji przed długim wybieganiem. Zostaw ten układ na cały blok Baza.',
  },
  recentRuns: [
    { id: 'g5', date: '2026-10-24', title: 'Tempo', distanceKm: 8.0, paceSPerKm: 298, avgHr: 162, load: 118 },
    { id: 'g4', date: '2026-10-23', title: 'Bieg łatwy', distanceKm: 6.1, paceSPerKm: 362, avgHr: 138, load: 52 },
    { id: 'g3', date: '2026-10-20', title: 'Interwały 6×800', distanceKm: 10.2, paceSPerKm: 312, avgHr: 155, load: 131 },
    { id: 'g2', date: '2026-10-19', title: 'Bieg łatwy', distanceKm: 6.0, paceSPerKm: 365, avgHr: 136, load: 50 },
    { id: 'g1', date: '2026-10-18', title: 'Długie', distanceKm: 15.0, paceSPerKm: 358, avgHr: 145, load: 142 },
  ],
  setsPerMuscle: {
    weekFrom: '2026-10-19',
    weekTo: '2026-10-25',
    items: [
      { muscle: 'upper_back', sets: 12 },
      { muscle: 'chest', sets: 9 },
      { muscle: 'quadriceps', sets: 7 },
      { muscle: 'shoulders', sets: 6 },
      { muscle: 'biceps', sets: 5 },
      { muscle: 'hamstrings', sets: 5 },
      { muscle: 'glutes', sets: 4 },
      { muscle: 'triceps', sets: 3 },
    ],
  },
  strengthSessions: [
    { id: 'h1', date: '2026-10-19', title: 'Push', durationMin: 58, sets: 18, tonnageKg: 8420 },
    { id: 'h2', date: '2026-10-21', title: 'Legs', durationMin: 64, sets: 16, tonnageKg: 11950 },
    { id: 'h3', date: '2026-10-23', title: 'Pull', durationMin: 52, sets: 17, tonnageKg: 7380 },
  ],
}

/** Stage 1: no plan and no coach. */
export const trainingStage1: TrainingResponse = {
  ...trainingFull,
  sync: [
    { source: 'garmin', status: 'ok', lastSuccessAt: '2026-10-25T06:12:00+01:00' },
    { source: 'hevy', status: 'ok', lastSuccessAt: '2026-10-25T06:05:00+01:00' },
  ],
  weeklyKm: trainingFull.weeklyKm.map((w) => ({ ...w, plannedKm: null })),
  todayPlannedKm: null,
  coachNote: null,
}
