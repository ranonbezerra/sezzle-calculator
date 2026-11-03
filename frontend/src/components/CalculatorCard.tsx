import { cn } from 'cn'
import { LoaderCircle } from 'lucide-react'
import type { FormEvent, KeyboardEvent, Ref } from 'react'
import type { Operation } from '@/api/calculator.ts'
import { Button } from '@/components/ui/button.tsx'
import { Card, CardContent } from '@/components/ui/card.tsx'
import type { Calculator } from '@/hooks/useCalculator.ts'
import { OperandInput } from './OperandInput.tsx'
import { OperationPicker } from './OperationPicker.tsx'
import { ResultDisplay } from './ResultDisplay.tsx'

// sqrt has no second label because it takes a single operand.
const LABELS: Record<Operation, { a: string; b?: string }> = {
  add: { a: 'First number', b: 'Second number' },
  subtract: { a: 'First number', b: 'Second number' },
  multiply: { a: 'First number', b: 'Second number' },
  divide: { a: 'Dividend', b: 'Divisor' },
  power: { a: 'Base', b: 'Exponent' },
  percentage: { a: 'Percent', b: 'Of' },
  sqrt: { a: 'Number' },
}

type CalculatorCardProps = {
  calculator: Calculator
  secondInputRef: Ref<HTMLInputElement>
  calculateRef: Ref<HTMLButtonElement>
}

export function CalculatorCard({ calculator, secondInputRef, calculateRef }: CalculatorCardProps) {
  const { operation, a, b, aError, bError, canSubmit, loading, result, error } = calculator
  const labels = LABELS[operation]

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    void calculator.submit()
  }

  function handleKeyDown(event: KeyboardEvent<HTMLFormElement>) {
    if (event.key === 'Escape') {
      event.preventDefault()
      calculator.clear()
    }
  }

  return (
    <Card className="rounded-2xl shadow-paper ring-border [--card-spacing:--spacing(5)] sm:[--card-spacing:--spacing(6)]">
      <CardContent>
        <form aria-label="Calculator" noValidate onSubmit={handleSubmit} onKeyDown={handleKeyDown} className="grid gap-5">
          <ResultDisplay result={result} />
          <OperationPicker value={operation} onChange={calculator.setOperation} />
          <div className="grid gap-2">
            <div className={labels.b ? 'grid gap-x-4 gap-y-1 sm:grid-cols-2' : 'grid'}>
              <OperandInput id="operand-a" label={labels.a} value={a} error={aError} onChange={calculator.setA} />
              {labels.b && (
                <OperandInput
                  ref={secondInputRef}
                  id="operand-b"
                  label={labels.b}
                  value={b}
                  error={bError}
                  onChange={calculator.setB}
                />
              )}
            </div>
            {/* Always mounted so screen readers announce errors; takes no space while empty. */}
            <p role="alert" className="mb-1 text-sm text-destructive empty:mb-0">
              {error}
            </p>
            <div className="flex items-center gap-2">
              <Button
                ref={calculateRef}
                type="submit"
                disabled={!canSubmit}
                aria-busy={loading}
                className={cn(
                  'h-11 flex-1 rounded-full text-base font-semibold shadow-paper hover:bg-primary/90 disabled:opacity-100',
                  // Invalid or empty input: a flat, muted button, clearly not clickable.
                  !canSubmit && !loading && 'bg-muted text-muted-foreground shadow-none',
                  loading && 'opacity-80',
                )}
              >
                {loading ? (
                  <>
                    <LoaderCircle aria-hidden className="motion-safe:animate-spin" />
                    Calculating…
                  </>
                ) : (
                  'Calculate'
                )}
              </Button>
              <Button
                type="button"
                variant="ghost"
                onClick={calculator.clear}
                aria-keyshortcuts="Escape"
                className="h-11 rounded-full px-4 text-base"
              >
                Clear
                <kbd aria-hidden className="rounded-md border px-1.5 font-mono text-xs text-muted-foreground">
                  Esc
                </kbd>
              </Button>
            </div>
          </div>
        </form>
      </CardContent>
    </Card>
  )
}
