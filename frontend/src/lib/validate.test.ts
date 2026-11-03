import { describe, expect, it } from 'vitest'
import { validateOperand } from './validate.ts'

describe('validateOperand', () => {
  it.each(['0', '42', '-1.5', '+3', '.5', '1.', '  7  ', '2e3', '-1.5E-3', '6.02e+23'])('accepts %j', (raw) => {
    expect(validateOperand(raw)).toBeNull()
  })

  it.each(['', '   '])('asks for a number when the input is %j', (raw) => {
    expect(validateOperand(raw)).toBe('Enter a number.')
  })

  it.each(['abc', '1.2.3', '--1', '+-1', '1-', '.', '-', 'e5', '1e', '1e+', '1e2.5', '1,5', '0x10', 'Infinity', 'NaN', '1 2'])(
    'rejects %j',
    (raw) => {
      expect(validateOperand(raw)).toBe('Enter a valid number, like -1.5 or 2e3.')
    },
  )

  it.each(['1e999', '-1e400'])('rejects %j as too large', (raw) => {
    expect(validateOperand(raw)).toBe('That number is too large.')
  })
})
