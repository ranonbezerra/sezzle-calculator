export type Operation = 'add' | 'subtract' | 'multiply' | 'divide' | 'power' | 'sqrt' | 'percentage'

// Error codes from docs/openapi.yaml (components.schemas.Error).
const API_ERROR_CODES = [
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
] as const

type ApiErrorCode = (typeof API_ERROR_CODES)[number]

type ClientErrorCode = 'network_error' | 'unexpected_response'

export type ErrorCode = ApiErrorCode | ClientErrorCode

export type CalculateResult =
  | { ok: true; value: number }
  | { ok: false; error: { code: ErrorCode; message: string } }

const TIMEOUT_MS = 5000

export async function calculate(operation: Operation, a: number, b?: number): Promise<CalculateResult> {
  const controller = new AbortController()
  const timer = setTimeout(() => controller.abort(), TIMEOUT_MS)
  try {
    const response = await fetch(`/api/v1/${operation}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(operation === 'sqrt' ? { a } : { a, b }),
      signal: controller.signal,
    })
    return toResult(response.ok, await response.json())
  } catch {
    return clientError('network_error', controller.signal.aborted ? 'request timed out' : 'request failed')
  } finally {
    clearTimeout(timer)
  }
}

function toResult(ok: boolean, body: unknown): CalculateResult {
  if (ok && isRecord(body) && typeof body.result === 'number') {
    return { ok: true, value: body.result }
  }
  if (!ok && isRecord(body) && isApiError(body.error)) {
    return { ok: false, error: { code: body.error.code, message: body.error.message } }
  }
  return clientError('unexpected_response', 'response does not match the API contract')
}

function clientError(code: ClientErrorCode, message: string): CalculateResult {
  return { ok: false, error: { code, message } }
}

function isApiError(value: unknown): value is { code: ApiErrorCode; message: string } {
  return isRecord(value) && isApiErrorCode(value.code) && typeof value.message === 'string'
}

function isApiErrorCode(value: unknown): value is ApiErrorCode {
  return (API_ERROR_CODES as readonly unknown[]).includes(value)
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null
}
