import { CornerLeftUp } from 'lucide-react'
import { Card, CardContent, CardHeader } from '@/components/ui/card.tsx'
import type { Calculation } from '@/hooks/useCalculator.ts'
import { formatExpression, formatResult } from '@/lib/format.ts'

type HistoryProps = {
  entries: Calculation[]
  onReuse: (entry: Calculation) => void
}

export function History({ entries, onReuse }: HistoryProps) {
  return (
    <Card
      role="region"
      aria-labelledby="history-title"
      className="rounded-2xl shadow-paper ring-border [--card-spacing:--spacing(5)]"
    >
      <CardHeader>
        <h2 id="history-title" className="font-heading text-lg font-medium">
          Recent
        </h2>
      </CardHeader>
      <CardContent>
        {entries.length === 0 ? (
          <p className="text-sm text-muted-foreground">No calculations yet.</p>
        ) : (
          <ol className="grid gap-2">
            {entries.map((entry) => (
              <li key={entry.id}>
                <HistoryItem entry={entry} onReuse={onReuse} />
              </li>
            ))}
          </ol>
        )}
      </CardContent>
    </Card>
  )
}

function HistoryItem({ entry, onReuse }: { entry: Calculation; onReuse: HistoryProps['onReuse'] }) {
  const expression = formatExpression(entry.operation, entry.a, entry.b)
  const value = formatResult(entry.value)

  return (
    <button
      type="button"
      onClick={() => onReuse(entry)}
      aria-label={`${expression} = ${value}, use as first number`}
      className="group grid w-full grid-cols-[minmax(0,1fr)_auto] items-center gap-x-2 rounded-xl border bg-background/60 px-3 py-2 text-left font-mono tabular-nums outline-none hover:border-primary/60 hover:bg-accent focus-visible:border-ring focus-visible:bg-accent focus-visible:ring-3 focus-visible:ring-ring/40 motion-safe:transition-colors"
    >
      <span className="truncate text-xs text-muted-foreground">{expression}</span>
      <CornerLeftUp
        aria-hidden
        className="row-span-2 size-4 text-primary opacity-0 group-hover:opacity-100 group-focus-visible:opacity-100 motion-safe:transition-opacity"
      />
      <span className="truncate text-base group-hover:text-primary group-focus-visible:text-primary">= {value}</span>
    </button>
  )
}
