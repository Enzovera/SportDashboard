import { describe, expect, it } from 'vitest'
import { COACH_SECRETS, SOURCE_SECRETS, secretMeta, secretProblem } from './settingsData'

describe('secretMeta', () => {
  it.each([
    [undefined, 'zmienna środowiskowa'],
    [{ name: 'HEVY_API_KEY', set: false, hint: null, updatedAt: null }, 'zmienna środowiskowa'],
    [{ name: 'HEVY_API_KEY', set: true, hint: '4f2a', updatedAt: '2026-11-02T23:30:00Z' }, '…4f2a · 03.11'],
    [{ name: 'HEVY_API_KEY', set: true, hint: null, updatedAt: null }, 'ustawiony'],
  ] as const)('%j -> %s', (s, want) => expect(secretMeta(s)).toBe(want))
})

describe('secretProblem', () => {
  const [hevy, mcp] = SOURCE_SECRETS
  it.each([
    [hevy, 'abc', null],
    [hevy, '', 'Puste pole.'],
    [hevy, 'abc\n', 'Wartość zawiera spacje lub nową linię.'],
    [mcp, 'x'.repeat(31), 'Co najmniej 32 znaków.'],
    [mcp, 'x'.repeat(32), null],
    [COACH_SECRETS[0], 'x'.repeat(4097), 'Za długa wartość.'],
  ])('%#', (spec, value, want) => expect(secretProblem(spec!, value)).toBe(want))
})
