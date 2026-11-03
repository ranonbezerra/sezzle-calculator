// Optional sign, digits with at most one decimal point, optional exponent: 1, -1.5, .5, 1., +2e-3.
const NUMBER_PATTERN = /^[+-]?(\d+\.?\d*|\.\d+)([eE][+-]?\d+)?$/

export function validateOperand(raw: string): string | null {
  const value = raw.trim()
  if (value === '') {
    return 'Enter a number.'
  }
  if (!NUMBER_PATTERN.test(value)) {
    return 'Enter a valid number, like -1.5 or 2e3.'
  }
  if (!Number.isFinite(Number(value))) {
    return 'That number is too large.'
  }
  return null
}
