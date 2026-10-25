import type { TodayResponse } from '../types'

/** Values from design/screens/FoldDzis.html: every section filled. */
export const todayFull: TodayResponse = {
  date: '2026-10-25',
  raceWeeksLeft: 20,
  sync: [
    { source: 'garmin', status: 'ok', lastSuccessAt: '2026-10-25T06:12:00+01:00' },
    { source: 'hevy', status: 'error', lastSuccessAt: '2026-10-24T18:02:00+02:00' },
  ],
  coach: {
    at: '2026-10-25T06:31:00+01:00',
    text:
      'Sen 7 h 40 min i HRV 52 ms przy bazie 54 — jesteś wypoczęty. Wczorajsze tempo poszło zgodnie z planem, ' +
      'ale zabrakło 28 g białka. Długie wybieganie bez zmian; dołóż węglowodanów przed biegiem i po nim.',
    flags: [
      { tone: 'good', text: 'Gotowość w normie' },
      { tone: 'warn', text: 'Białko wczoraj −28 g' },
    ],
  },
  session: {
    title: 'Długie wybieganie',
    action: 'keep',
    distanceKm: 16,
    zone: 'strefa 2',
    paceSPerKm: [350, 370],
    durationMin: 96,
  },
  week: {
    from: '2026-10-19',
    to: '2026-10-25',
    days: [
      { date: '2026-10-19', runKm: 6, runPlanned: false, strength: 'Push' },
      { date: '2026-10-20', runKm: 10, runPlanned: false, strength: null },
      { date: '2026-10-21', runKm: null, runPlanned: false, strength: 'Legs' },
      { date: '2026-10-22', runKm: null, runPlanned: false, strength: null },
      { date: '2026-10-23', runKm: 6, runPlanned: false, strength: 'Pull' },
      { date: '2026-10-24', runKm: 8, runPlanned: false, strength: null },
      { date: '2026-10-25', runKm: 16, runPlanned: true, strength: null },
    ],
    runKmDone: 30,
    runKmPlanned: 46,
    strengthDone: 3,
    strengthPlanned: 3,
  },
  tomorrow: { title: 'Push', detail: '6 km spokojnie' },
  readiness: {
    measuredAt: '2026-10-25T06:12:00+01:00',
    hrv7dMs: 52,
    hrvBaselineMs: 54,
    hrvTone: 'good',
    hrvLabel: 'w normie',
    sleepS: 27600,
    sleepAvg7dS: 25800,
    restingHr: 48,
    restingHrDelta7d: -1,
    tsb: 4,
    form: 'fresh',
  },
  nutrition: {
    targetLabel: 'cel dnia długiego',
    kcal: { eaten: 820, target: 2950 },
    proteinG: { eaten: 38, target: 145 },
    carbsG: { eaten: 112, target: 412 },
    fatG: { eaten: 24, target: 80 },
    yesterday: {
      kcal: { eaten: 2610, target: 2700 },
      proteinG: { eaten: 117, target: 145 },
      proteinTone: 'warn',
    },
  },
  supplements: { taken: 2, total: 7, next: 'kofeina' },
  proposalsPending: 2,
  weight: { avg7dKg: 76.4, deltaPerWeekKg: -0.3 },
}

/** What stage 1 can actually serve: Garmin + Hevy only, no coach, plan, nutrition or supplements. */
export const todayStage1: TodayResponse = {
  ...todayFull,
  raceWeeksLeft: null,
  sync: [
    { source: 'garmin', status: 'ok', lastSuccessAt: '2026-10-25T06:12:00+01:00' },
    { source: 'hevy', status: 'ok', lastSuccessAt: '2026-10-25T06:05:00+01:00' },
  ],
  coach: null,
  session: null,
  week: {
    ...todayFull.week,
    days: todayFull.week.days.map((d) => (d.runPlanned ? { ...d, runKm: null, runPlanned: false } : d)),
    runKmPlanned: null,
    strengthPlanned: null,
  },
  tomorrow: null,
  nutrition: null,
  supplements: null,
  proposalsPending: null,
  weight: null,
}
