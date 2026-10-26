import { describe, expect, it } from 'vitest'
import { isTabActive } from './nav'

describe('isTabActive', () => {
  it.each([
    ['/', '/', true],
    ['/', '/trening', false],
    ['/zywienie', '/zywienie', true],
    ['/zywienie', '/zywienie/suplementy', true],
    ['/zywienie', '/zywieniex', false],
    ['/propozycje', '/plan', false],
    ['/plan', '/plan', true],
  ])('tab %s on %s -> %s', (tab, path, want) => expect(isTabActive(tab, path)).toBe(want))
})
