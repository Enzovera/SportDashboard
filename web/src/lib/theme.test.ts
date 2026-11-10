import { describe, expect, it } from 'vitest'
import { nextTheme, parseTheme } from './theme'

describe('nextTheme', () => {
  it.each([
    ['auto', 'light'],
    ['light', 'dark'],
    ['dark', 'auto'],
  ] as const)('%s -> %s', (from, to) => {
    expect(nextTheme(from)).toBe(to)
  })
})

describe('parseTheme', () => {
  it.each([
    ['light', 'light'],
    ['dark', 'dark'],
    [null, 'auto'],
    ['garbage', 'auto'],
  ] as const)('%s -> %s', (value, mode) => {
    expect(parseTheme(value)).toBe(mode)
  })
})
