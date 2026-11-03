import { Divide, Minus, Percent, Plus, Radical, X } from 'lucide-react'
import type { ReactNode } from 'react'
import type { Operation } from '@/api/calculator.ts'
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group.tsx'

const OPERATIONS: { id: Operation; label: string; icon: ReactNode }[] = [
  { id: 'add', label: 'Add', icon: <Plus /> },
  { id: 'subtract', label: 'Subtract', icon: <Minus /> },
  { id: 'multiply', label: 'Multiply', icon: <X /> },
  { id: 'divide', label: 'Divide', icon: <Divide /> },
  {
    id: 'power',
    label: 'Power',
    icon: (
      <span className="font-heading text-base leading-none italic">
        x<sup className="text-[0.7em]">y</sup>
      </span>
    ),
  },
  { id: 'sqrt', label: 'Square root', icon: <Radical /> },
  { id: 'percentage', label: 'Percentage', icon: <Percent /> },
]

type OperationPickerProps = {
  value: Operation
  onChange: (operation: Operation) => void
}

export function OperationPicker({ value, onChange }: OperationPickerProps) {
  const index = OPERATIONS.findIndex((operation) => operation.id === value)

  return (
    <div className="grid gap-2">
      <span id="operation-label" className="text-sm font-medium">
        Operation
      </span>
      <ToggleGroup
        type="single"
        value={value}
        // Re-selecting the active item reports "", which would leave no operation selected.
        onValueChange={(next) => next && onChange(next as Operation)}
        aria-labelledby="operation-label"
        className="relative grid w-full grid-cols-7 gap-0 rounded-full bg-muted p-1"
      >
        <span
          aria-hidden
          className="absolute inset-y-1 left-1 w-[calc((100%-0.5rem)/7)] rounded-full bg-card shadow-paper motion-safe:transition-transform motion-safe:duration-300 motion-safe:ease-out"
          style={{ transform: `translateX(${index * 100}%)` }}
        />
        {OPERATIONS.map((operation) => (
          <ToggleGroupItem
            key={operation.id}
            value={operation.id}
            aria-label={operation.label}
            title={operation.label}
            className="relative h-10 min-w-0 rounded-full px-0 text-muted-foreground hover:bg-transparent hover:text-foreground data-[state=on]:bg-transparent data-[state=on]:text-primary motion-safe:transition-colors"
          >
            {operation.icon}
          </ToggleGroupItem>
        ))}
      </ToggleGroup>
    </div>
  )
}
