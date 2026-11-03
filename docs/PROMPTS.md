# Prompts

How AI was used: planning in a Claude chat, implementation with Claude Code
guided by CLAUDE.md. Every change was reviewed, tested and committed by me.
Full session transcripts are in [docs/prompts/](prompts/).

## 0. Planning

To be added with the export of the planning conversation.

## 1. Bootstrap monorepo and CI (#1)

```
Read CLAUDE.md. Implement issue #1 (bootstrap monorepo and CI).
- backend/: go mod init github.com/ranonbezerra/seezle-calculator/backend, cmd/server/main.go that starts an HTTP server on PORT (default 8080) exposing GET /healthz, and one test for it.
- frontend/: scaffold Vite React TypeScript, add Vitest, Testing Library, jsdom and @vitest/coverage-v8. Configure the dev server to proxy /api to http://localhost:8080. Remove the Vite boilerplate UI and leave a minimal App with one passing test. Do not set up Tailwind or shadcn/ui yet; that belongs to issue #5.
- frontend/.nvmrc with 22.
- Root Makefile with backend-test, backend-lint (go vet + golangci-lint), backend-run, backend-cover.
- .golangci.yml in golangci-lint v2 format with gosec enabled.
- .github/workflows/ci.yml with two jobs, backend and frontend, running lint, typecheck where relevant, and tests with coverage. Read the Go version from backend/go.mod and the Node version from frontend/.nvmrc. Cache Go modules and npm.
- .github/pull_request_template.md with sections: Summary, Changes, How to test, Checklist.
- MIT LICENSE, .gitignore for Go and Node, .editorconfig.
Run every test and lint command locally and show me the output.
```

Follow-up prompts:
- "Add to the root Makefile: frontend-test (npm run coverage inside frontend/), frontend-lint (lint + typecheck), and a test target that runs backend and frontend tests. Update the Commands section of CLAUDE.md accordingly."
- "the frontend-run and cover as well."

Manual changes after review (landed in the #2 PR as a separate chore commit):
frontend-test switched to plain `npm test`, aggregate `lint` target,
`.gitignore` scoped to `frontend/coverage/` (so `docs/coverage` is tracked) and the `backend/server` binary.

Transcript: [01-bootstrap.md](prompts/01-bootstrap.md)

## 2. Calculator core (#2)

```
Read CLAUDE.md. Implement issue #2 (calculator core).
Create backend/internal/calculator with pure functions over float64 returning (float64, error): Add, Subtract, Multiply, Divide, Power, Sqrt, Percentage.
- Sentinel errors: ErrDivisionByZero, ErrNegativeSqrt, ErrNonFinite (any NaN or ±Inf result, e.g. overflow in Multiply or Power).
- Percentage(a, b) means "a percent of b" = a*b/100.
- Expose a lookup from operation name to its function and arity (1 or 2), so the HTTP layer has no switch statements. Unknown names must be detectable by the caller.
- Table-driven tests: normal cases, zeros, negatives, decimals, very large values, overflow, 0^0, negative base with fractional exponent, and one test documenting that 0.1 + 0.2 returns the raw float64 (formatting is the frontend's job).
- No HTTP code in this package.
Run make backend-cover and show per-function coverage. Target 100% for this package.
```

Manual changes after review: normalized `-0` to `0` in `finite` (JSON would
otherwise serialize `-0`), and added a test asserting the sign with `math.Signbit`.

Transcript: [02-calculator-core.md](prompts/02-calculator-core.md)