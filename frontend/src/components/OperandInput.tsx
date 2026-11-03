import type { Ref } from 'react'
import { Input } from '@/components/ui/input.tsx'
import { Label } from '@/components/ui/label.tsx'

type OperandInputProps = {
  id: string
  label: string
  value: string
  error: string | null
  onChange: (value: string) => void
  ref?: Ref<HTMLInputElement>
}

export function OperandInput({ id, label, value, error, onChange, ref }: OperandInputProps) {
  // An empty field only disables Calculate; complaining before the user types is noise.
  const showError = error !== null && value.trim() !== ''
  const errorId = `${id}-error`

  return (
    <div className="grid content-start gap-1.5">
      <Label htmlFor={id}>{label}</Label>
      <Input
        ref={ref}
        id={id}
        type="text"
        inputMode="decimal"
        autoComplete="off"
        spellCheck={false}
        placeholder="0"
        value={value}
        onChange={(event) => onChange(event.target.value)}
        aria-invalid={showError || undefined}
        aria-describedby={showError ? errorId : undefined}
        className="h-11 rounded-xl bg-background/60 px-3.5 font-mono text-base tabular-nums placeholder:text-muted-foreground/60 md:text-base"
      />
      {/* Always mounted with one line reserved, so a message appearing doesn't push the layout. */}
      <p id={errorId} className="min-h-4 text-xs leading-4 text-destructive">
        {showError && error}
      </p>
    </div>
  )
}
