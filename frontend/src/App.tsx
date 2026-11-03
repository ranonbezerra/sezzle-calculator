import { useRef } from 'react'
import { flushSync } from 'react-dom'
import { CalculatorCard } from '@/components/CalculatorCard.tsx'
import { History } from '@/components/History.tsx'
import { type Calculation, useCalculator } from '@/hooks/useCalculator.ts'

function App() {
  const calculator = useCalculator()
  const secondInputRef = useRef<HTMLInputElement>(null)
  const calculateRef = useRef<HTMLButtonElement>(null)

  function handleReuse(entry: Calculation) {
    // Render first: Calculate may only become enabled once the reused operand is in place.
    flushSync(() => calculator.reuse(entry))
    // sqrt renders no second input, so focus falls through to Calculate.
    const target = secondInputRef.current ?? calculateRef.current
    target?.focus()
  }

  return (
    <main className="mx-auto grid min-h-dvh w-full max-w-3xl content-center gap-6 px-4 py-10">
      <header className="grid gap-1">
        <h1 className="font-heading text-3xl font-semibold tracking-tight sm:text-4xl">Sezzle Calculator</h1>
        <p className="text-sm text-muted-foreground">Enter to calculate, Esc to clear.</p>
      </header>
      <div className="grid items-start gap-6 md:grid-cols-[minmax(0,1fr)_16rem]">
        <CalculatorCard calculator={calculator} secondInputRef={secondInputRef} calculateRef={calculateRef} />
        <History entries={calculator.history} onReuse={handleReuse} />
      </div>
    </main>
  )
}

export default App
