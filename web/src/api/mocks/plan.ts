import type { PlanDay, PlanResponse, PlanSession } from '../plan'

const run = (title: string, km: number, detail: string | null = null): PlanSession => ({
  kind: 'run',
  title,
  detail,
  distanceKm: km,
})
const strength = (title: string, detail: string | null = null): PlanSession => ({
  kind: 'strength',
  title,
  detail,
  distanceKm: null,
})

const T = {
  strengthRun: { kcal: 2750, proteinG: 145, carbsG: 370, fatG: 80 },
  run: { kcal: 2700, proteinG: 145, carbsG: 355, fatG: 80 },
  strength: { kcal: 2600, proteinG: 145, carbsG: 330, fatG: 80 },
  rest: { kcal: 2300, proteinG: 145, carbsG: 255, fatG: 80 },
  long: { kcal: 2950, proteinG: 145, carbsG: 412, fatG: 80 },
}

const day = (date: string, d: Omit<PlanDay, 'date' | 'done' | 'proposals'> & Partial<PlanDay>): PlanDay => ({
  date,
  done: null,
  proposals: [],
  ...d,
})

/** Values from design/screens/FoldPlan.html and DesktopPlan.html (today = Sunday 25.10, Baza from tomorrow). */
export const planFull: PlanResponse = {
  date: '2026-10-25',
  proposalsPending: 2,
  plan: {
    goal: {
      race: 'Maraton',
      date: '2027-03-14',
      weeksLeft: 20,
      targetTimeS: 4 * 3600,
      targetPaceSPerKm: 341,
      secondary: [
        { label: 'Półmaraton', value: '< 1:40', detail: 'najlepszy 1:47' },
        { label: 'Siła', value: 'Push / Pull / Legs', detail: '3× w Bazie' },
      ],
    },
    blocks: [
      { kind: 'base', from: '2026-10-26', to: '2026-12-06', weeks: 6, summary: '46 → ok. 60 km, głównie strefa 2. PPL 3×.' },
      { kind: 'build', from: '2026-12-07', to: '2027-01-17', weeks: 6, summary: 'Tempo i interwały 2×. PPL 2–3×.' },
      { kind: 'peak', from: '2027-01-18', to: '2027-02-21', weeks: 5, summary: 'Najdłuższe wybiegania, tempo 5:41 /km.' },
      { kind: 'taper', from: '2027-02-22', to: '2027-03-14', weeks: 3, summary: null },
    ],
    currentWeek: {
      from: '2026-10-19',
      to: '2026-10-25',
      block: null,
      weekInBlock: null,
      plannedKm: 46,
      plannedStrength: 3,
      doneKm: 30,
      doneStrength: 3,
      days: [
        day('2026-10-19', { sessions: [strength('Push'), run('bieg łatwy', 6)], runKm: 6, dayType: 'strength_run', target: T.strengthRun, done: { runKm: 6.0, strengthSets: 18 } }),
        day('2026-10-20', { sessions: [run('Interwały', 10, '6×800 m')], runKm: 10, dayType: 'run', target: T.run, done: { runKm: 10.2, strengthSets: null } }),
        day('2026-10-21', { sessions: [strength('Legs')], runKm: null, dayType: 'strength', target: T.strength, done: { runKm: null, strengthSets: 16 } }),
        day('2026-10-22', { sessions: [], runKm: null, dayType: 'rest', target: T.rest, done: { runKm: null, strengthSets: null } }),
        day('2026-10-23', { sessions: [strength('Pull'), run('bieg łatwy', 6)], runKm: 6, dayType: 'strength_run', target: T.strengthRun, done: { runKm: 6.1, strengthSets: 17 } }),
        day('2026-10-24', { sessions: [run('Tempo', 8)], runKm: 8, dayType: 'run', target: T.run, done: { runKm: 8.0, strengthSets: null } }),
        day('2026-10-25', { sessions: [run('Długie wybieganie', 16, 'strefa 2')], runKm: 16, dayType: 'long_run', target: T.long }),
      ],
    },
    nextWeek: {
      from: '2026-10-26',
      to: '2026-11-01',
      block: 'base',
      weekInBlock: 1,
      plannedKm: 46,
      plannedStrength: 3,
      doneKm: null,
      doneStrength: null,
      days: [
        day('2026-10-26', { sessions: [strength('Push'), run('bieg łatwy', 6)], runKm: 6, dayType: 'strength_run', target: T.strengthRun }),
        day('2026-10-27', { sessions: [run('Interwały', 10, '6×800 m')], runKm: 10, dayType: 'run', target: T.run }),
        day('2026-10-28', {
          sessions: [strength('Legs', '16 serii roboczych')],
          runKm: null,
          dayType: 'strength',
          target: T.strength,
          proposals: [{ id: 'p-2026-10-24-legs', summary: '12 serii' }],
        }),
        day('2026-10-29', { sessions: [], runKm: null, dayType: 'rest', target: T.rest }),
        day('2026-10-30', { sessions: [strength('Pull'), run('bieg łatwy', 6)], runKm: 6, dayType: 'strength_run', target: T.strengthRun }),
        day('2026-10-31', { sessions: [run('Tempo', 8)], runKm: 8, dayType: 'run', target: T.run }),
        day('2026-11-01', {
          sessions: [run('Długie wybieganie', 16, 'strefa 2')],
          runKm: 16,
          dayType: 'long_run',
          target: T.long,
          proposals: [{ id: 'p-2026-10-25-carbs', summary: 'W 460 g, 3 140 kcal' }],
        }),
      ],
    },
  },
}

/** Stage 1: no plan exists yet (plan_event is empty). */
export const planStage1: PlanResponse = {
  date: '2026-10-25',
  plan: null,
  proposalsPending: null,
}
