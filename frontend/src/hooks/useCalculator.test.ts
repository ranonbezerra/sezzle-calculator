import { act, renderHook } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { useCalculator } from './useCalculator.ts'

const fetchMock = vi.fn<typeof fetch>()

beforeEach(() => {
  vi.stubGlobal('fetch', fetchMock)
})

afterEach(() => {
  fetchMock.mockReset()
  vi.unstubAllGlobals()
})

describe('useCalculator', () => {
  it.each<[string, string, string]>([
    ['empty operands', '', ''],
    ['an invalid operand', '1', '1..2'],
  ])('does not call the API with %s', async (_, a, b) => {
    const { result } = renderHook(() => useCalculator())
    act(() => {
      result.current.setA(a)
      result.current.setB(b)
    })

    await act(() => result.current.submit())

    expect(result.current.canSubmit).toBe(false)
    expect(fetchMock).not.toHaveBeenCalled()
  })
})
