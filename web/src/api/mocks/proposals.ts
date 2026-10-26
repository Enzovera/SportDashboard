import type { DayPlan, ProposalsResponse } from '../proposals'

const day = (session: string, runKm: number | null, kcal: number, carbsG: number): DayPlan => ({
  session,
  runKm,
  kcal,
  carbsG,
})

/** Values from design/screens/FoldPropozycje.html and DesktopPropozycje.html: two pending, one blocked, history. */
export const proposalsFull: ProposalsResponse = {
  date: '2026-10-25',
  pending: [
    {
      id: 'p-2026-10-25-carbs',
      kind: 'nutrition',
      origin: 'daily_run',
      proposedAt: '2026-10-25T06:31:00+01:00',
      title: 'Dzień długiego wybiegania: węgle 412 → 460 g',
      rationale:
        'Na dwóch ostatnich długich wybieganiach tempo spadało po 13 km, a węgle wyniosły 4,8 i 5,0 g/kg. ' +
        'Wyższy cel i posiłek przed biegiem powinny utrzymać tempo.',
      effectiveFrom: '2026-11-01',
      diff: {
        heading: 'cel dnia',
        rows: [
          { label: 'Węglowodany', before: 412, after: 460, unit: 'g', note: '6,0 g/kg' },
          { label: 'Kalorie', before: 2950, after: 3140, unit: 'kcal', note: null },
          { label: 'Białko', before: 145, after: 145, unit: 'g', note: null },
          { label: 'Tłuszcz', before: 80, after: 80, unit: 'g', note: null },
          { label: 'Przed biegiem', before: null, after: '80 g węgli', unit: null, note: '2–3 h przed' },
        ],
        total: null,
      },
      checks: [
        { rule: 'kcal_floor', status: 'pass', text: 'kcal powyżej dolnego progu' },
        { rule: 'long_run_deficit', status: 'pass', text: 'brak deficytu w dniu długiego' },
        { rule: 'protein_min', status: 'pass', text: 'białko ≥ 1,6 g/kg' },
      ],
    },
    {
      id: 'p-2026-10-24-legs',
      kind: 'strength',
      origin: 'daily_run',
      proposedAt: '2026-10-24T06:30:00+02:00',
      title: 'Śr 28.10, Legs: 16 → 12 serii',
      rationale:
        'W bloku Baza kilometraż rośnie, a Legs obciąża te same mięśnie co bieg. ' +
        '12 serii utrzyma bodziec przy mniejszym zmęczeniu przed sobotnim tempem.',
      effectiveFrom: null,
      diff: {
        heading: 'Serie robocze',
        rows: [
          { label: 'Przysiad', before: 4, after: 3, unit: null, note: null },
          { label: 'Martwy ciąg rumuński', before: 4, after: 3, unit: null, note: null },
          { label: 'Wykroki', before: 4, after: 3, unit: null, note: null },
          { label: 'Łydki', before: 4, after: 3, unit: null, note: null },
        ],
        total: { label: 'Razem', before: 16, after: 12, unit: null, note: null },
      },
      checks: [
        { rule: 'rest_day', status: 'pass', text: 'dzień wolny: czwartek' },
        { rule: 'km_increase', status: 'pass', text: 'kilometraż bez zmian (46 km)' },
      ],
    },
  ],
  blocked: [
    {
      id: 'p-2026-10-25-long',
      kind: 'run',
      origin: 'daily_run',
      proposedAt: '2026-10-25T06:31:00+01:00',
      title: 'Nd 1.11: długie 16 → 22 km',
      checks: [
        { rule: 'km_increase', status: 'fail', text: 'przyrost 46 → 52 km (+13%, limit 10%)' },
        { rule: 'long_run_share', status: 'warn', text: 'długie = 42% tygodnia (limit 35%)' },
      ],
    },
  ],
  history: [
    {
      id: 'p-2026-10-18-week',
      kind: 'run',
      origin: 'weekly_run',
      title: 'Tydzień 19–25.10: 43 → 46 km',
      proposedAt: '2026-10-18T20:04:00+02:00',
      decidedAt: '2026-10-18T21:15:00+02:00',
      decision: 'accepted',
      comment: null,
    },
    {
      id: 'p-2026-10-14-easy',
      kind: 'run',
      origin: 'daily_run',
      title: 'Sb 17.10: tempo → bieg łatwy (HRV 4 dni pod bazą)',
      proposedAt: '2026-10-14T06:30:00+02:00',
      decidedAt: '2026-10-14T07:02:00+02:00',
      decision: 'accepted',
      comment: null,
    },
    {
      id: 'p-2026-10-11-extra',
      kind: 'run',
      origin: 'daily_run',
      title: 'Dodatkowy bieg w czwartek 15.10',
      proposedAt: '2026-10-11T06:30:00+02:00',
      decidedAt: '2026-10-11T08:40:00+02:00',
      decision: 'rejected',
      comment: 'czwartek zostaje wolny',
    },
  ],
  week: {
    from: '2026-10-26',
    to: '2026-11-01',
    block: 'Baza · tydz. 1',
    days: [
      { date: '2026-10-26', now: day('Push + bieg 6 km', 6, 2750, 370), changes: [] },
      { date: '2026-10-27', now: day('Interwały 10 km', 10, 2700, 355), changes: [] },
      {
        date: '2026-10-28',
        now: day('Legs 16 serii', null, 2600, 330),
        changes: [{ proposalId: 'p-2026-10-24-legs', after: day('Legs 12 serii', null, 2600, 330) }],
      },
      { date: '2026-10-29', now: day('Odpoczynek', null, 2300, 255), changes: [] },
      { date: '2026-10-30', now: day('Pull + bieg 6 km', 6, 2750, 370), changes: [] },
      { date: '2026-10-31', now: day('Tempo 8 km', 8, 2700, 355), changes: [] },
      {
        date: '2026-11-01',
        now: day('Długie 16 km', 16, 2950, 412),
        changes: [{ proposalId: 'p-2026-10-25-carbs', after: day('Długie 16 km', 16, 3140, 460) }],
      },
    ],
  },
}

/** Stage 1: no plan and no coach yet, so nothing to decide. */
export const proposalsStage1: ProposalsResponse = {
  date: '2026-10-25',
  pending: [],
  blocked: [],
  history: [],
  week: null,
}
