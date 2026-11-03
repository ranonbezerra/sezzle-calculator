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

## 3. REST API (#3)

```
Read CLAUDE.md. Implement issue #3 (REST API). Plan first, then wait for my approval.

- backend/internal/httpapi: handler registered as POST /api/v1/{operation} using Go 1.22+ ServeMux patterns, plus GET /healthz (move it here from main).
- Decode with DisallowUnknownFields and a 1 KB MaxBytesReader. Use *float64 in the request struct so missing fields are detected. Validate arity with calculator.Lookup: sqrt requires only "a" and rejects "b"; binary operations require both.
- Error codes and statuses, all in the CLAUDE.md error format:
  invalid_json (400), invalid_number (400, e.g. 1e309 or a string where a number is expected), missing_operand (400), unexpected_operand (400), request_too_large (413), unknown_operation (404), method_not_allowed (405), division_by_zero, negative_sqrt and non_finite_result (422).
  The default ServeMux 404/405 responses are plain text; make sure clients always get the JSON error format. Update the API contract section of CLAUDE.md if you add a status code.
- Content-Type application/json on every response.
- Request logging middleware with slog (method, path, status, duration).
- http.Server with ReadHeaderTimeout, ReadTimeout, WriteTimeout and IdleTimeout. Graceful shutdown on SIGINT/SIGTERM.
- Keep main.go thin: extract config loading (PORT and defaults) and server construction into testable functions and cover them with tests. main only wires things and starts the server.
- httptest tests: each operation, each error code and status, wrong method, unknown route, malformed, oversized and out-of-range bodies, the logging middleware.
- docs/openapi.yaml (OpenAPI 3.1) describing the endpoints, request and response schemas and every error code.
Run make backend-cover and make backend-lint and show the output.
```

Decisions made during plan review (Claude Code asked, I chose):
- Unknown body field returns `invalid_json` (400), with the decoder's detail in the message.
- Paths outside the API return a new `not_found` (404) code instead of ServeMux's plain-text 404.
- A calculator error that is not one of the sentinels returns `internal_error` (500). This added 413 and 500 to the contract in CLAUDE.md.

Accepted edge cases (documented, not changed): ServeMux's 301 redirect for non-canonical
paths such as `/api//v1/add` has no JSON body.

Landed in the #4 PR, as it was missed in the #3 PR.

Transcript: [03-rest-api.md](prompts/03-rest-api.md)