# Seezle take-home: full-stack calculator

## Layout
- backend/: Go 1.27, standard library only (net/http, encoding/json, log/slog). No web frameworks. golangci-lint v2.
- frontend/: React + TypeScript + Vite, Tailwind CSS + shadcn/ui. Tests with Vitest + Testing Library.

## API contract
- POST /api/v1/{operation} with {"a": number, "b": number}; sqrt uses only "a".
- Operations: add, subtract, multiply, divide, power, sqrt, percentage.
- percentage(a, b) = a percent of b = a*b/100.
- Success: 200 {"operation": "...", "result": number}
- Errors: {"error": {"code": "snake_case_code", "message": "..."}}
  400 invalid input, 422 math error, 404 unknown operation, 405 wrong method.

## Rules
- One issue, one branch, one PR. Stay inside the issue scope; do not add features.
- Every change ships with unit tests. Prefer table-driven tests in Go.
- Small functions, no premature abstractions, no dead code.
- Do not commit or push. I review and commit myself.
- When done, run the relevant test and lint commands and report the output.

## Commands
- All: make test, make lint
- Backend: make backend-test, make backend-lint, make backend-cover, make backend-run
- Frontend: make frontend-test, make frontend-lint, make frontend-cover, make frontend-run
- Inside frontend/: npm test, npm run coverage, npm run lint, npm run typecheck, npm run dev