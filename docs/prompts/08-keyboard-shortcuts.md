
❯ /clear

❯ <ide_opened_file>The user opened the file
  /Users/ranonbezerra/Projects/seezle-calculator/docs/PROMPTS.md in the IDE.
  This may or may not be related to the current task.</ide_opened_file>

❯ Read CLAUDE.md. Fix issue #8: the Esc and Enter shortcuts only work when
  focus is inside the form, because they are handled on the form's onKeyDown.

  Move both shortcuts to a single document-level keydown listener registered in
  an effect and cleaned up on unmount:
  - Esc clears from anywhere.
  - Enter calculates from anywhere, except when the event target is an
  interactive element outside the form (button, link, input, select, textarea,
  or anything with a role of button/radio/link), which keeps its native
  behavior. Inside the form, Enter keeps working as today without submitting
  twice.
  - Do nothing while a request is loading or when Calculate is disabled.
  - Ignore events with defaultPrevented.

  Regression tests, written first and shown failing before the fix:
  - Esc with focus on document.body clears inputs and result.
  - Enter with focus on document.body calculates.
  - Enter on a focused history item reuses it and does not calculate.
  - Enter inside an input sends exactly one request.
    Run make frontend-cover and make frontend-lint.

  Searched for 1 pattern, ran 4 shell commands (ctrl+o to expand)

⏺ Writing the four regression tests first, next to the existing Enter/Escape
  test.

⏺ Update(frontend/src/App.test.tsx)
  ⎿  Added 57 lines
      240      expect(within(screen.getByRole('region', { name: 'Recent' 
           })).getAllByRole('listitem')).toHaveLength(1)
      241    })
      242  
      243 +  describe('keyboard shortcuts outside the form', () => {     
      244 +    it('clears inputs and result on Escape with focus on the p
          +age body', async () => {                                      
      245 +      respondWithResult(5)                                    
      246 +      const user = setup()                                    
      247 +      await fillOperands(user, '2', '3')                      
      248 +      await user.click(calculateButton())                     
      249 +      await waitFor(() => expect(result()).toHaveTextContent('
          +5'))                                                          
      250 +                                                              
      251 +      await user.click(document.body)                         
      252 +      expect(document.body).toHaveFocus()                     
      253 +      await user.keyboard('{Escape}')                         
      254 +                                                              
      255 +      expect(firstInput()).toHaveValue('')                    
      256 +      expect(secondInput()).toHaveValue('')                   
      257 +      expect(result()).toHaveTextContent('0')                 
      258 +    })                                                        
      259 +                                                              
      260 +    it('calculates on Enter with focus on the page body', asyn
          +c () => {                                                     
      261 +      respondWithResult(5)                                    
      262 +      const user = setup()                                    
      263 +      await fillOperands(user, '2', '3')                      
      264 +                                                              
      265 +      await user.click(document.body)                         
      266 +      expect(document.body).toHaveFocus()                     
      267 +      await user.keyboard('{Enter}')                          
      268 +                                                              
      269 +      await waitFor(() => expect(result()).toHaveTextContent('
          +5'))                                                          
      270 +      expect(sentRequest()).toEqual({ url: '/api/v1/add', body
          +: { a: 2, b: 3 } })                                           
      271 +    })                                                        
      272 +                                                              
      273 +    it('reuses a focused history item on Enter without calcula
          +ting', async () => {                                          
      274 +      respondWithResult(5)                                    
      275 +      const user = setup()                                    
      276 +      await fillOperands(user, '2', '3')                      
      277 +      await user.click(calculateButton())                     
      278 +      const item = await screen.findByRole('button', { name: '
          +2 + 3 = 5, use as first number' })                            
      279 +                                                              
      280 +      item.focus()                                            
      281 +      await user.keyboard('{Enter}')                          
      282 +                                                              
      283 +      expect(firstInput()).toHaveValue('5')                   
      284 +      expect(secondInput()).toHaveFocus()                     
      285 +      expect(fetchMock).toHaveBeenCalledTimes(1)              
      286 +    })                                                        
      287 +                                                              
      288 +    it('sends exactly one request on Enter inside an input', a
          +sync () => {                                                  
      289 +      respondWithResult(5)                                    
      290 +      const user = setup()                                    
      291 +      await fillOperands(user, '2', '3')                      
      292 +                                                              
      293 +      await user.keyboard('{Enter}')                          
      294 +                                                              
      295 +      await waitFor(() => expect(result()).toHaveTextContent('
          +5'))                                                          
      296 +      expect(fetchMock).toHaveBeenCalledTimes(1)              
      297 +    })                                                        
      298 +  })                                                          
      299 +                                                              
      300    it('clears the inputs with the Clear button', async () => {
      301      const user = setup()
      302  

  Ran 2 shell commands (ctrl+o to expand)

⏺ The two focus-on-body tests fail as expected, while the history-item and
  in-input guards pass before the fix. Now I'm applying the fix.

⏺ Update(frontend/src/components/CalculatorCard.tsx)
  ⎿  Added 17 lines, removed 3 lines
      35      void calculator.submit()
      36    }
      37  
      38 -  function handleKeyDown(event: KeyboardEvent<HTMLFormElement>)
         - {                                                             
      38 +  const handleKeyDown = useEffectEvent((event: KeyboardEvent) =
         +> {                                                            
      39 +    if (event.defaultPrevented || loading) {                   
      40 +      return                                                   
      41 +    }                                                          
      42      if (event.key === 'Escape') {
      43        event.preventDefault()
      44        calculator.clear()
      45 +      return                                                   
      46      }
      43 -  }                                                            
      47 +    if (event.key === 'Enter' && canSubmit && !handlesEnterNati
         +vely(event.target, formRef.current)) {                         
      48 +      event.preventDefault()                                   
      49 +      void calculator.submit()                                 
      50 +    }                                                          
      51 +  })                                                           
      52  
      53 +  useEffect(() => {                                            
      54 +    const listener = (event: KeyboardEvent) => handleKeyDown(ev
         +ent)                                                           
      55 +    document.addEventListener('keydown', listener)             
      56 +    return () => document.removeEventListener('keydown', listen
         +er)                                                            
      57 +  }, [])                                                       
      58 +                                                               
      59    return (
      60      <Card className="rounded-2xl shadow-paper ring-border [--ca
          rd-spacing:--spacing(5)] sm:[--card-spacing:--spacing(6)]">
      61        <CardContent>
      48 -        <form aria-label="Calculator" noValidate onSubmit={hand
         -leSubmit} onKeyDown={handleKeyDown} className="grid gap-5">    
      62 +        <form ref={formRef} aria-label="Calculator" noValidate 
         +onSubmit={handleSubmit} className="grid gap-5">                
      63            <ResultDisplay result={result} />
      64            <OperationPicker value={operation} onChange={calculat
          or.setOperation} />
      65            <div className="grid gap-2">

⏺ Update(frontend/src/components/CalculatorCard.tsx)
  ⎿  Added 1 line
      29  export function CalculatorCard({ calculator, secondInputRef, ca
          lculateRef }: CalculatorCardProps) {
      30    const { operation, a, b, aError, bError, canSubmit, loading, 
          result, error } = calculator
      31    const labels = LABELS[operation]
      32 +  const formRef = useRef<HTMLFormElement>(null)                
      33  
      34    function handleSubmit(event: FormEvent<HTMLFormElement>) {
      35      event.preventDefault()

⏺ Update(frontend/src/components/CalculatorCard.tsx)
  ⎿  Added 1 line, removed 1 line
      1  import { cn } from 'cn'
      2  import { LoaderCircle } from 'lucide-react'
      3 -import type { FormEvent, KeyboardEvent, Ref } from 'react'      
      3 +import { type FormEvent, type Ref, useEffect, useEffectEvent, us
        +eRef } from 'react'                                             
      4  import type { Operation } from '@/api/calculator.ts'
      5  import { Button } from '@/components/ui/button.tsx'
      6  import { Card, CardContent } from '@/components/ui/card.tsx'

⏺ Update(frontend/src/components/CalculatorCard.tsx)
  ⎿  Added 8 lines
      20    sqrt: { a: 'Number' },
      21  }
      22  
      23 +const INTERACTIVE = 'button, a, input, select, textarea, [role=
         +"button"], [role="radio"], [role="link"]'                      
      24 +                                                               
      25 +// Inside the form, Enter already submits through the input or 
         +activates the focused control,                                 
      26 +// and outside it an interactive element keeps its own Enter be
         +havior.                                                        
      27 +function handlesEnterNatively(target: EventTarget | null, form:
         + HTMLFormElement | null): boolean {                            
      28 +  return target instanceof Element && (form?.contains(target) |
         +| target.matches(INTERACTIVE))                                 
      29 +}                                                              
      30 +                                                               
      31  type CalculatorCardProps = {
      32    calculator: Calculator
      33    secondInputRef: Ref<HTMLInputElement>

  Ran 2 shell commands (ctrl+o to expand)

⏺ Esc and Enter now work no matter where focus is. All 130 tests pass, and lint
  and typecheck are clean. Nothing is committed; the changes are on
  fix/8-keyboard-shortcuts.

  Tests first. I added the four regression tests to App.test.tsx
  (frontend/src/App.test.tsx) and ran them before changing any code:
   × clears inputs and result on Escape with focus on the page body
   × calculates on Enter with focus on the page body
   ✓ reuses a focused history item on Enter without calculating
   ✓ sends exactly one request on Enter inside an input
   Tests  2 failed | 2 passed
  The two body-focus tests failed, which reproduces the bug. The history-item
  and in-input tests passed even before the fix. That's expected: they guard
  against the new listener causing a double action, which the old form-only
  handler couldn't do. If the listener didn't skip the form or interactive
  elements, each would send a second request and fail.

  The fix (CalculatorCard.tsx (frontend/src/components/CalculatorCard.tsx)):
  - I removed the form's onKeyDown and added one keydown listener on document.
    It's registered once in an effect and removed on unmount.
  - The handler uses useEffectEvent, so it always sees the current calculator
    state without re-registering on every render.
  - It does nothing if the event is defaultPrevented or a request is loading.
  - Esc clears from anywhere.
  - Enter calculates only when Calculate is enabled. It's skipped when the
    target is inside the form or is an interactive element (button, a, input,
    select, textarea, or role button/radio/link). Inside the form, the browser's
    own Enter handling still does the work, so a request is sent once, not
    twice.

  Behaviour change: Esc used to clear during a pending request when focus was in
  the form. Now it does nothing while loading, as the issue asks. The Clear
  button still works during loading.

  make frontend-cover
   Test Files  6 passed (6)
        Tests  130 passed (130)
  Statements   : 100% ( 139/139 )
  Branches     : 98.05% ( 101/103 )
  Functions    : 100% ( 39/39 )
  Lines        : 100% ( 133/133 )
  The report lists only files below 100%. The one uncovered branch is in
  format.ts (lines 34–36) and predates this change. CalculatorCard.tsx isn't
  listed, so it's fully covered.

  make frontend-lint: oxlint and tsc -b both finished with no errors or
  warnings.