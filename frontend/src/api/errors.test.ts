import { describe, expect, it } from 'vitest'
import { errorMessage } from './errors.ts'

const FALLBACK = 'Something went wrong. Please try again.'

describe('errorMessage', () => {
  it.each([
    'invalid_json',
    'invalid_number',
    'missing_operand',
    'unexpected_operand',
    'request_too_large',
    'unknown_operation',
    'not_found',
    'method_not_allowed',
    'division_by_zero',
    'negative_sqrt',
    'non_finite_result',
    'internal_error',
    'network_error',
    'unexpected_response',
  ])('has a specific message for %s', (code) => {
    const message = errorMessage(code)

    expect(message).not.toBe('')
    expect(message).not.toBe(FALLBACK)
  })

  it.each(['rate_limited', '', 'toString'])('falls back to a generic message for unknown code %j', (code) => {
    expect(errorMessage(code)).toBe(FALLBACK)
  })
})
