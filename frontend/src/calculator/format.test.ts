import { describe, expect, it } from 'vitest'
import { formatInput, formatNumber } from './format'

describe('number formatting', () => {
  it.each([[0, '0'], [-0, '0'], [.1 + .2, '0.3'], [1 / 3, '0.333333333333'], [123456789, '123456789']])('formats %s', (value, expected) => {
    expect(formatNumber(value)).toBe(expected)
  })
  it('handles invalid and long input', () => {
    expect(formatNumber(Infinity)).toBe('Error')
    expect(formatInput('0.')).toBe('0.')
    expect(formatInput('-')).toBe('-')
    expect(formatInput('123456789012345')).toBe('123456789012345')
    expect(formatInput('1234567890123456')).toBe('1234567890120000')
  })
})
