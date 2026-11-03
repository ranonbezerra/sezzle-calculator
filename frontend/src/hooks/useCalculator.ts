import { useRef, useState } from 'react'
import { calculate, type Operation } from '@/api/calculator.ts'
import { errorMessage } from '@/api/errors.ts'
import { validateOperand } from '@/lib/validate.ts'

export type Calculation = {
  id: number
  operation: Operation
  a: number
  b?: number
  value: number
}

export type Calculator = ReturnType<typeof useCalculator>

const HISTORY_SIZE = 5

export function useCalculator() {
  const [operation, setOperation] = useState<Operation>('add')
  const [a, setA] = useState('')
  const [b, setB] = useState('')
  const [loading, setLoading] = useState(false)
  const [result, setResult] = useState<Calculation | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [history, setHistory] = useState<Calculation[]>([])
  const nextId = useRef(1)

  const unary = operation === 'sqrt'
  const aError = validateOperand(a)
  const bError = unary ? null : validateOperand(b)
  const canSubmit = !loading && aError === null && bError === null

  async function submit() {
    if (!canSubmit) {
      return
    }
    const request = { operation, a: Number(a), b: unary ? undefined : Number(b) }
    setLoading(true)
    setError(null)
    const response = await calculate(request.operation, request.a, request.b)
    setLoading(false)
    if (!response.ok) {
      setError(errorMessage(response.error.code))
      return
    }
    const calculation = { id: nextId.current++, ...request, value: response.value }
    setResult(calculation)
    setHistory((previous) => [calculation, ...previous].slice(0, HISTORY_SIZE))
  }

  function clear() {
    setA('')
    setB('')
    setResult(null)
    setError(null)
  }

  function reuse(calculation: Calculation) {
    setA(String(calculation.value))
    setError(null)
  }

  return {
    operation,
    setOperation,
    a,
    setA,
    b,
    setB,
    aError,
    bError,
    canSubmit,
    loading,
    result,
    error,
    history,
    submit,
    clear,
    reuse,
  }
}
