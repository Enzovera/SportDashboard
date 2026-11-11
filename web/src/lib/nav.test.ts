import { describe, expect, it } from 'vitest'
import { TABS, coverTabs, isTabActive } from './nav'

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

describe('coverTabs', () => {
  it('keeps the daily views in the bar and the rest under "Więcej", decisions first', () => {
    const { bar, more } = coverTabs()
    expect(bar.map((t) => t.label)).toEqual(['Dziś', 'Trening', 'Żywienie', 'Plan'])
    expect(more.map((t) => t.label)).toEqual(['Propozycje', 'Ciało'])
  })

  it('leaves room for "Więcej" as the fifth slot', () => {
    expect(coverTabs().bar.length).toBeLessThanOrEqual(4)
  })

  it('sends a new non-primary view to the sheet', () => {
    const tabs = [...TABS, { to: '/raport', label: 'Raport', icon: 'history' as const, primary: false }]
    expect(coverTabs(tabs).bar).toHaveLength(4)
    expect(coverTabs(tabs).more.map((t) => t.label)).toEqual(['Propozycje', 'Ciało', 'Raport'])
  })
})
