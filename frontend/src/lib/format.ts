import type { Operation } from '@/api/calculator.ts'

const SIGNIFICANT_DIGITS = 12

// Beyond these magnitudes a 12-digit plain number would show padding zeros, not precision.
const EXPONENT_ABOVE = 1e12
const EXPONENT_BELOW = 1e-6

export function formatResult(value: number): string {
  const rounded = Number(value.toPrecision(SIGNIFICANT_DIGITS))
  const magnitude = Math.abs(rounded)
  if (magnitude === 0) {
    return '0'
  }
  if (magnitude >= EXPONENT_ABOVE || magnitude < EXPONENT_BELOW) {
    return rounded.toExponential()
  }
  return String(rounded)
}

const SYMBOLS: Record<Exclude<Operation, 'sqrt' | 'percentage'>, string> = {
  add: '+',
  subtract: '−',
  multiply: '×',
  divide: '÷',
  power: '^',
}

export function formatExpression(operation: Operation, a: number, b?: number): string {
  if (operation === 'sqrt') {
    return `√${operand(a)}`
  }
  if (operation === 'percentage') {
    return `${operand(a)}% of ${operand(b ?? 0)}`
  }
  return `${operand(a)} ${SYMBOLS[operation]} ${operand(b ?? 0)}`
}

function operand(value: number): string {
  const text = formatResult(value)
  return value < 0 ? `(${text})` : text
}
