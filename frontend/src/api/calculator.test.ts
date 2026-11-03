import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { calculate } from './calculator.ts'

const fetchMock = vi.fn<typeof fetch>()

beforeEach(() => {
  vi.stubGlobal('fetch', fetchMock)
})

afterEach(() => {
  fetchMock.mockReset()
  vi.unstubAllGlobals()
  vi.useRealTimers()
})

function jsonResponse(status: number, body: unknown): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  })
}

function sentRequest(): { url: string; init: RequestInit; body: unknown } {
  const [url, init] = fetchMock.mock.calls[0]
  return { url: String(url), init: init ?? {}, body: JSON.parse(String(init?.body)) }
}

describe('calculate', () => {
  it('posts both operands for a binary operation and returns the result', async () => {
    fetchMock.mockResolvedValue(jsonResponse(200, { operation: 'add', result: 5 }))

    const result = await calculate('add', 2, 3)

    expect(result).toEqual({ ok: true, value: 5 })
    const { url, init, body } = sentRequest()
    expect(url).toBe('/api/v1/add')
    expect(init.method).toBe('POST')
    expect(init.headers).toEqual({ 'Content-Type': 'application/json' })
    expect(body).toEqual({ a: 2, b: 3 })
  })

  it('sends only "a" for sqrt', async () => {
    fetchMock.mockResolvedValue(jsonResponse(200, { operation: 'sqrt', result: 9 }))

    const result = await calculate('sqrt', 81, 4)

    expect(result).toEqual({ ok: true, value: 9 })
    const { url, body } = sentRequest()
    expect(url).toBe('/api/v1/sqrt')
    expect(body).toEqual({ a: 81 })
    expect(body).not.toHaveProperty('b')
  })

  it.each([
    [400, 'invalid_json'],
    [400, 'invalid_number'],
    [400, 'missing_operand'],
    [400, 'unexpected_operand'],
    [404, 'unknown_operation'],
    [404, 'not_found'],
    [405, 'method_not_allowed'],
    [413, 'request_too_large'],
    [422, 'division_by_zero'],
    [422, 'negative_sqrt'],
    [422, 'non_finite_result'],
    [500, 'internal_error'],
  ])('returns the API error for status %i with code %s', async (status, code) => {
    fetchMock.mockResolvedValue(jsonResponse(status, { error: { code, message: 'server message' } }))

    const result = await calculate('divide', 1, 0)

    expect(result).toEqual({ ok: false, error: { code, message: 'server message' } })
  })

  it.each<[string, Response]>([
    ['an unknown error code', jsonResponse(400, { error: { code: 'rate_limited', message: 'slow down' } })],
    ['an error without a message', jsonResponse(400, { error: { code: 'invalid_json' } })],
    ['an error status without an error body', jsonResponse(500, { result: 1 })],
    ['a success status without a numeric result', jsonResponse(200, { result: '5' })],
    ['a success status with an error body', jsonResponse(200, { error: { code: 'internal_error', message: 'x' } })],
    ['a JSON null body', jsonResponse(200, null)],
  ])('returns unexpected_response for %s', async (_, response) => {
    fetchMock.mockResolvedValue(response)

    const result = await calculate('add', 1, 2)

    expect(result).toEqual({
      ok: false,
      error: { code: 'unexpected_response', message: 'response does not match the API contract' },
    })
  })

  it('returns network_error for a non-JSON response', async () => {
    fetchMock.mockResolvedValue(new Response('<html>Bad Gateway</html>', { status: 502 }))

    const result = await calculate('add', 1, 2)

    expect(result).toEqual({ ok: false, error: { code: 'network_error', message: 'request failed' } })
  })

  it('returns network_error when fetch fails', async () => {
    fetchMock.mockRejectedValue(new TypeError('Failed to fetch'))

    const result = await calculate('add', 1, 2)

    expect(result).toEqual({ ok: false, error: { code: 'network_error', message: 'request failed' } })
  })

  it('aborts and returns network_error after 5 seconds', async () => {
    vi.useFakeTimers()
    fetchMock.mockImplementation(
      (_, init) =>
        new Promise((_, reject) => {
          init?.signal?.addEventListener('abort', () => reject(new DOMException('aborted', 'AbortError')))
        }),
    )
    let settled = false
    const pending = calculate('add', 1, 2).finally(() => {
      settled = true
    })

    await vi.advanceTimersByTimeAsync(4999)
    expect(settled).toBe(false)
    await vi.advanceTimersByTimeAsync(1)

    expect(await pending).toEqual({ ok: false, error: { code: 'network_error', message: 'request timed out' } })
    expect(sentRequest().init.signal?.aborted).toBe(true)
  })
})
