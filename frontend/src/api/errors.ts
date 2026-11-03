import type { ErrorCode } from './calculator.ts'

const MESSAGES: Record<ErrorCode, string> = {
  invalid_json: "The request couldn't be read. Please try again.",
  invalid_number: 'Please enter valid numbers.',
  missing_operand: 'Please fill in every number.',
  unexpected_operand: 'Square root takes only one number.',
  request_too_large: 'The input is too long.',
  unknown_operation: "That operation isn't supported.",
  not_found: "The calculator service couldn't be found.",
  method_not_allowed: 'The server rejected the request.',
  division_by_zero: "You can't divide by zero.",
  negative_sqrt: "You can't take the square root of a negative number.",
  non_finite_result: 'The result is too large or not a real number.',
  internal_error: 'Something went wrong on our side. Please try again.',
  network_error: "Couldn't reach the calculator. Check your connection and try again.",
  unexpected_response: 'The server returned an unexpected response. Please try again.',
}

const FALLBACK_MESSAGE = 'Something went wrong. Please try again.'

export function errorMessage(code: string): string {
  return Object.hasOwn(MESSAGES, code) ? MESSAGES[code as ErrorCode] : FALLBACK_MESSAGE
}
