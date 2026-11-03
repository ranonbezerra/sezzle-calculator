import { describe, expect, it } from 'vitest'
import { formatExpression, formatResult } from './format.ts'

describe('formatResult', () => {
  it.each<[number, string]>([
    [0, '0'],
    [-0, '0'],
    [36, '36'],
    [-42.5, '-42.5'],
    [0.1 + 0.2, '0.3'],
    [1 / 3, '0.333333333333'],
    [-2 / 3, '-0.666666666667'],
    [123456789012.4, '123456789012'],
    [999999999999.9, '1e+12'],
    [1e12, '1e+12'],
    [1.23456789012345e21, '1.23456789012e+21'],
    [-9.87654321e15, '-9.87654321e+15'],
    [0.000001, '0.000001'],
    [1.5e-7, '1.5e-7'],
    [Number.MIN_VALUE, '5e-324'],
  ])('formats %d as %s', (value, expected) => {
    expect(formatResult(value)).toBe(expected)
  })
})

describe('formatExpression', () => {
  it.each<[Parameters<typeof formatExpression>, string]>([
    [['add', 2, 3], '2 + 3'],
    [['subtract', 2, -3], '2 − (-3)'],
    [['multiply', 12, 3], '12 × 3'],
    [['divide', 1, 3], '1 ÷ 3'],
    [['power', 2, 10], '2 ^ 10'],
    [['sqrt', 81], '√81'],
    [['percentage', 5, 200], '5% of 200'],
    [['add', 0.1 + 0.2, 1e21], '0.3 + 1e+21'],
  ])('formats %j as %s', (args, expected) => {
    expect(formatExpression(...args)).toBe(expected)
  })
})
