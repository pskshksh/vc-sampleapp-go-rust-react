import { describe, expect, it } from 'vitest'
import { formatDate } from './format'

describe('formatDate', () => {
  it('renders a YYYY-MM-DD string as a long, human-readable date', () => {
    // Force a stable locale/timezone-independent assertion by checking the
    // parts we control rather than the exact localized string.
    const out = formatDate('2026-08-17')
    expect(out).toContain('2026')
    expect(out).toContain('August')
    expect(out).toContain('17')
  })

  it('parses the date at local midnight (no off-by-one day)', () => {
    // '2026-01-01' must not roll back to Dec 31 in negative-offset timezones.
    const out = formatDate('2026-01-01')
    expect(out).toContain('January')
    expect(out).toContain('2026')
    expect(out).not.toContain('December')
  })
})
