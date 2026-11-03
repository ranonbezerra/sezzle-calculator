import { cn } from 'cn'
import type { Calculation } from '@/hooks/useCalculator.ts'
import { formatExpression, formatResult } from '@/lib/format.ts'

type ResultDisplayProps = {
  result: Calculation | null
}

export function ResultDisplay({ result }: ResultDisplayProps) {
  const value = result ? formatResult(result.value) : '0'

  return (
    <div className="grid gap-1 text-right">
      <p id="result-expression" className="min-h-5 truncate font-mono text-sm text-muted-foreground tabular-nums">
        {result && formatExpression(result.operation, result.a, result.b)}
      </p>
      <output
        aria-live="polite"
        aria-label="Result"
        aria-describedby={result ? 'result-expression' : undefined}
        className={cn('font-mono leading-tight font-medium tracking-tight break-all tabular-nums', sizeFor(value))}
      >
        {value}
      </output>
    </div>
  )
}

// Steps the font down so the longest values (e.g. -1.23456789012e-100) fit a 360px screen.
function sizeFor(value: string): string {
  if (value.length <= 9) {
    return 'text-5xl sm:text-6xl'
  }
  if (value.length <= 14) {
    return 'text-3xl sm:text-4xl'
  }
  return 'text-2xl sm:text-3xl'
}
