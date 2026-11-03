
❯ /clear

❯ <ide_opened_file>The user opened the file
  /Users/ranonbezerra/Projects/seezle-calculator/.gitignore in the IDE. This
  may or may not be related to the current task.</ide_opened_file>

❯ Read CLAUDE.md. Implement issue #2 (calculator core).
  Create backend/internal/calculator with pure functions over float64 returning
  (float64, error): Add, Subtract, Multiply, Divide, Power, Sqrt, Percentage.
  - Sentinel errors: ErrDivisionByZero, ErrNegativeSqrt, ErrNonFinite (any NaN
  or ±Inf result, e.g. overflow in Multiply or Power).
  - Percentage(a, b) means "a percent of b" = a*b/100.
  - Expose a lookup from operation name to its function and arity (1 or 2), so
  the HTTP layer has no switch statements. Unknown names must be detectable by
  the caller.
  - Table-driven tests: normal cases, zeros, negatives, decimals, very large
  values, overflow, 0^0, negative base with fractional exponent, and one test
  documenting that 0.1 + 0.2 returns the raw float64 (formatting is the
  frontend's job).
  - No HTTP code in this package.
    Run make backend-cover and show per-function coverage. Target 100% for this
  package.

  Ran 2 shell commands (ctrl+o to expand)

⏺ Branch feat/2-calculator-core is already checked out and clean. Writing the
  package now.

⏺ Write(backend/internal/calculator/calculator.go)
  ⎿  Wrote 65 lines to backend/internal/calculator/calculator.go
       1 // Package calculator implements the arithmetic operations expos
         ed by the API
       2 // as pure functions over float64.
       3 package calculator
       4
       5 import (
       6   "errors"
       7   "math"
       8 )
       9
     … +56 lines (ctrl+o to expand)

⏺ Write(backend/internal/calculator/operations.go)
  ⎿  Wrote 26 lines to backend/internal/calculator/operations.go
      1 package calculator
      2
      3 // Operation describes a calculator operation by its arity and im
        plementation.
      4 type Operation struct {
      5   // Arity is the number of operands the operation uses: 1 or 2.
      6   Arity int
      7   // Apply runs the operation. When Arity is 1, b is ignored.
      8   Apply func(a, b float64) (float64, error)
      9 }
     … +17 lines (ctrl+o to expand)

⏺ Write(backend/internal/calculator/calculator_test.go)
  ⎿  Wrote 150 lines to backend/internal/calculator/calculator_test.go
       1 package calculator
       2
       3 import (
       4   "errors"
       5   "math"
       6   "testing"
       7 )
       8
       9 type binaryCase struct {
      10   name    string
     … +140 lines (ctrl+o to expand)

⏺ Write(backend/internal/calculator/operations_test.go)
  ⎿  Wrote 55 lines to backend/internal/calculator/operations_test.go
       1 package calculator
       2
       3 import (
       4   "errors"
       5   "testing"
       6 )
       7
       8 func TestLookup(t *testing.T) {
       9   tests := []struct {
      10     name      string
     … +45 lines (ctrl+o to expand)

  Ran 1 shell command (ctrl+o to expand)

⏺ All tests pass and the package is at 100%. Now running lint and the plain test
  target.

  Ran 1 shell command (ctrl+o to expand)

⏺ Issue #2 is done: the backend/internal/calculator package is written, all
  tests and lint pass, and coverage for the package is 100%. Nothing is
  committed; the new files are untracked on feat/2-calculator-core.

  What's in it
  - calculator.go (backend/internal/calculator/calculator.go): the seven
    functions and the three sentinel errors. Every result goes through one check
    that returns ErrNonFinite if it is NaN or ±Inf. Divide returns
    ErrDivisionByZero when b == 0, which includes -0 and 0/0. Sqrt returns
    ErrNegativeSqrt when a < 0.
  - operations.go (backend/internal/calculator/operations.go): Lookup(name)
    (Operation, bool), where Operation holds Arity and Apply(a, b). An unknown
    name returns false. Every operation uses the same two-argument Apply, and
    for sqrt the second argument is ignored. The lookup table itself is private,
    so callers can't change it.
  - Tests are table-driven and check errors with errors.Is. They cover
    everything you listed. TestAddReturnsRawFloat64 checks that Add(0.1, 0.2)
    returns 0.30000000000000004. The lookup tests check the arity and result for
    every name, including that sqrt ignores b. They also check that "",
    "modulo", "Add", " add" and "sqrt/" are not found. The large-value cases use
    powers of two such as 0x1p1000, so exact == comparisons are safe.

  Behaviour you may want to check
  - Negative base with fractional exponent: Power(-8, 1/3) returns ErrNonFinite.
    math.Pow gives NaN here, and the brief only defined three error types, so
    there is no more specific error.
  - Zero to a negative power: Power(0, -1) also returns ErrNonFinite, not
    ErrDivisionByZero, because the result is +Inf.
  - Percentage overflow: Percentage computes a*b before dividing by 100. If a*b
    is too large, it returns ErrNonFinite even when the true answer would fit. I
    kept a*b/100 because it is the contract formula and it is exact for typical
    inputs (10% of 30 is exactly 3).

  make backend-cover
  ok   .../backend/cmd/server              coverage: 35.7% of statements
  ok   .../backend/internal/calculator     coverage: 100.0% of statements
  .../internal/calculator/calculator.go:20:   Add         100.0%
  .../internal/calculator/calculator.go:25:   Subtract    100.0%
  .../internal/calculator/calculator.go:30:   Multiply    100.0%
  .../internal/calculator/calculator.go:35:   Divide      100.0%
  .../internal/calculator/calculator.go:43:   Power       100.0%
  .../internal/calculator/calculator.go:48:   Sqrt        100.0%
  .../internal/calculator/calculator.go:56:   Percentage  100.0%
  .../internal/calculator/calculator.go:60:   finite      100.0%
  .../internal/calculator/operations.go:23:   Lookup      100.0%
  The overall total is 70.0% because main and port in cmd/server already had no
  tests. This issue doesn't touch them.

  make backend-lint: 0 issues.
  make backend-test: both packages ok.