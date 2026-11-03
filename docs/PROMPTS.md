# Prompts

How AI was used: planning in a Claude chat, implementation with Claude Code
guided by CLAUDE.md. Every change was reviewed, tested and committed by me.
Full session transcripts are in [docs/prompts/](prompts/).

Numbers in section headings are issue numbers. Issues and PRs share one sequence on GitHub,
so the PRs are #8 to #15.

## 0. Planning

Planned in a Claude chat before writing any code, and used the same chat to review each step.
Decisions made there:
- Go for the backend, as the assignment prefers, using only the standard library.
- Private repository during development, made public with full history for submission.
- Issue breakdown (#1 to #7), API contract and CLAUDE.md conventions defined before implementation.
- Edge cases added to the plan: float precision, out-of-range numbers,
  server and client timeouts, decimal input handling.
- Tailwind + shadcn/ui for a modern, accessible UI without a heavy component library.

My prompts from that conversation: [00-planning.md](prompts/00-planning.md)

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

Manual changes after review (landed in the PR for issue #2 as a separate chore commit):
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

Landed in the PR for issue #4 (#11), as it was missed in the PR for #3.

Transcript: [03-rest-api.md](prompts/03-rest-api.md)

## 4. Typed API client (#4)

```
Read CLAUDE.md and docs/openapi.yaml. Implement issue #4 (typed API client). Treat the OpenAPI spec as the source of truth for operations, request bodies, responses and error codes.

- frontend/src/api/calculator.ts:
  - An Operation union type for the seven operations and a CalculateResult discriminated union:
    { ok: true; value: number } | { ok: false; error: { code: ErrorCode; message: string } }.
  - calculate(operation, a, b?) that POSTs to /api/v1/{operation} with Content-Type application/json, sending only "a" for sqrt.
  - A 5 second timeout with AbortController.
  - Network failures, timeouts and non-JSON or unexpected responses all become code "network_error".
  - Never throw: every outcome is returned as a CalculateResult.
- frontend/src/api/errors.ts: map every error code in the spec (plus network_error) to a short, friendly English message for the UI. Unknown codes fall back to a generic message.
- Unit tests with fetch mocked via vi.stubGlobal: success for a binary operation and for sqrt (asserting the request body has no "b"), each error code, unknown error code, network failure, timeout (fake timers), non-JSON response.
- No UI changes.
Run make frontend-cover and make frontend-lint and show the output.
```

Claude Code flagged a trade-off: mapping an unknown error code to `network_error` would tell
the user the server was unreachable when it actually replied. Instead of either option it
offered, I split the two meanings.

Follow-up prompt:
```
Neither option: split the two meanings. Add a client-only code "unexpected_response" to ErrorCode, with its own message in errors.ts ("The server returned an unexpected response. Please try again.").
- network_error: fetch rejected, timeout, or a body that is not JSON (e.g. an HTML 502 from a proxy).
- unexpected_response: valid JSON with the wrong shape, a 2xx without a numeric result, or an error code not in the spec.
Keep ErrorCode a closed union. Update the tests accordingly and rerun make frontend-cover and make frontend-lint.
```

Transcript: [04-api-client.md](prompts/04-api-client.md)

## 5. Calculator UI (#5)

```
Read CLAUDE.md and frontend/src/api/. Implement issue #5 (calculator UI). Use Plan mode: before any code, propose two distinct visual directions (palette, typography, layout, how the operation picker, result display and history look, light and dark) and wait for me to pick one. If the frontend-design skill is available, use it.

After I choose:
- Set up Tailwind CSS and shadcn/ui, adding only the components the UI uses. Exclude src/components/ui from coverage in the Vitest config.
- Use calculate() and errorMessage() from src/api. The result must never be computed on the client; the frontend only validates input and formats what the API returns.
- Layout: a centered card with a large monospace result display using tabular numbers; a segmented operation picker with icons (+ − × ÷ xʸ √ %); the second input hidden when sqrt is selected; a history panel of the last 5 calculations (side on desktop, below on mobile), where clicking an entry reuses its result as operand A.
- Inputs: type="text" with inputMode="decimal"; validate sign, decimal point and scientific notation; show inline messages and disable Calculate while invalid.
- Format results for display with up to 12 significant digits.
- States: loading prevents double submission; API errors appear in an aria-live region; every input has a label; visible focus states.
- Keyboard: Enter calculates, Esc clears.
- Light/dark following the system; motion respects prefers-reduced-motion. Responsive down to 360px.
- Keep state in a useCalculator hook and split into small components (CalculatorCard, OperandInput, OperationPicker, ResultDisplay, History).
- Testing Library tests for: successful calculation, validation messages, server error display, sqrt hiding the second input, loading state, keyboard shortcuts, reusing a history item, result formatting.
Run make frontend-cover and make frontend-lint and show the output.
```

Decisions made during plan review: of the two visual directions Claude Code proposed,
I chose "Warm Paper" (warm stone neutrals, terracotta accent, Fraunces / Inter / JetBrains Mono,
self-hosted fonts so the app makes no external requests).

After checking the running app in the browser, I asked for a round of visual adjustments.
The most important one was context-aware labels: "First number / Second number" did not
make the percentage semantics clear ("a percent of b"), so the labels now change per operation.

Follow-up prompt:
```
Visual review adjustments:
- Context-aware input labels per operation: power "Base" / "Exponent", percentage "Percent" / "Of", divide "Dividend" / "Divisor", sqrt a single "Number"; add/subtract/multiply keep "First number" / "Second number".
- After a calculation, show the expression (e.g. "200 × 123") in a smaller muted line above the result in the display.
- Make the enabled Calculate button clearly distinct from its disabled state.
- Add a muted "0" placeholder to the inputs.
- Reserve space for inline validation messages without leaving a large gap when there are none.
- Give history items clear hover and focus-visible states.
Update the tests for the labels and the expression line, then rerun make frontend-cover and make frontend-lint.
```

Second follow-up, so the reuse flow continues naturally (click history, type the other number, Enter):
```
After reusing a history entry, keep filling the first operand as today, then move focus to the second input, or to the Calculate button when the operation is sqrt. Add a test for it and rerun make frontend-cover and make frontend-lint.
```

Focus handling stays out of the hook (App passes refs to the card). The reuse update uses
`flushSync` so a previously disabled Calculate button is enabled before it receives focus;
removing it makes the sqrt focus test fail.

Considered and rejected: filling the last focused input when reusing a history entry. Once
the entry is clicked, focus has already left the inputs, so the target would be invisible
state. Reuse always fills the first operand, like a calculator's ANS.

Transcript: [05-calculator-ui.md](prompts/05-calculator-ui.md)

## 6. Docker and docker-compose (#6)

```
Read CLAUDE.md. Implement issue #6 (Docker and docker-compose).

- backend/Dockerfile: multi-stage. Build a static binary (CGO_ENABLED=0, -trimpath, -ldflags "-s -w") with the Go version from go.mod; final image gcr.io/distroless/static:nonroot. Expose 8080.
- Distroless has no shell or curl, so add a -healthcheck flag to the server binary: it GETs http://localhost:$PORT/healthz with a short timeout and exits 0 or 1. Keep it small and testable, with a unit test, and keep main thin.
- frontend/Dockerfile: multi-stage. Build with the Node version from .nvmrc using npm ci; serve dist/ with nginxinc/nginx-unprivileged (non-root).
- frontend/nginx.conf: SPA fallback to index.html; proxy /api and /healthz to http://backend:8080; long cache headers for hashed assets in /assets, no cache for index.html; gzip on.
- .dockerignore for both (node_modules, dist, coverage, .git, test files where safe).
- docker-compose.yml at the root: backend with a healthcheck using the -healthcheck flag; frontend depends_on backend with condition service_healthy; app published on http://localhost:3000. The backend port does not need to be published.
- Add a docker job to CI that builds both images, runs docker compose up -d --wait, curls POST /api/v1/add through http://localhost:3000 and checks the JSON result, then runs docker compose down. It should run after the backend and frontend jobs.
- Add make targets: docker-up (docker compose up --build) and docker-down.
Run docker compose up --build -d --wait, then curl a success, a 422 and the SPA index through port 3000, show the output, and bring it down. Also run make backend-cover and make backend-lint.
```

Claude Code raised two open points; I accepted both improvements:

```
Both decisions:
- Add a frontend healthcheck in docker-compose.yml using the busybox wget in the image against http://localhost:8080/ (the nginx-unprivileged port), so docker compose up --wait waits for nginx too. Then remove the connection-refused retry from the CI smoke test if it is no longer needed.
- Pin nginx-unprivileged to the current stable version with alpine (exact version tag, not just alpine), matching how Go and Node are pinned.
Rerun docker compose up --build -d --wait, the same curl checks through port 3000, and bring it down.
```

Accepted deviation: the frontend healthcheck uses `127.0.0.1` instead of `localhost`. Busybox
wget resolves `localhost` to `::1` without falling back to IPv4, and nginx listens only on IPv4.
Adding an IPv6 listener would break nginx on hosts with IPv6 disabled.

Result: backend image 16.3 MB (distroless, nonroot), frontend image on a pinned nginx-unprivileged,
both containers healthy before `--wait` returns, CI smoke test without retries.

While testing the app manually in the Docker setup, I found that the Esc and Enter shortcuts
only work with focus inside the form. Since it belongs to the #5 scope, it became its own bug (#13).

Transcript: [06-docker.md](prompts/06-docker.md)

## 7. Docs: README, design decisions, coverage, prompts (#7)

```
Read CLAUDE.md, docs/openapi.yaml, docs/PROMPTS.md and the whole repository. Implement issue #7 (docs).

README.md, replacing the placeholder, with these sections:
- Overview: what it is, the stack, a short text architecture diagram (browser -> nginx -> Go API; Vite proxy in dev).
- Prerequisites: Go 1.27, Node 22 (.nvmrc), Docker with Compose.
- Quick start with Docker: make docker-up, open http://localhost:3000.
- Run locally: backend and frontend separately with the make targets.
- Tests and coverage: make test, make lint, make backend-cover, make frontend-cover; a table with the current coverage numbers per package/area, and a note that main/fatal are untested by design and src/components/ui is vendored shadcn code excluded from coverage.
- API reference: endpoints, request/response shapes, status codes, and curl examples for every operation and for each error code reachable over HTTP, with the REAL responses. Start the backend, run every curl, paste the actual output, then stop it. Link docs/openapi.yaml.
- Design decisions: Go stdlib only; one endpoint per operation; error model and status codes (400/404/405/413/422/500); float64 with non-finite results as 422 and -0 normalized; raw float in the API with 12 significant digits formatting in the UI; percentage as "a percent of b" computed as a*b/100 (exact for common values, explicit error only near the float64 limit); *float64 and DisallowUnknownFields for strict input; 1 KB body limit; same-origin proxy instead of CORS (Vite in dev, nginx in Docker); form UI instead of a keypad, with per-operation labels; results computed only by the backend; history reuse fills the first operand, like a calculator's ANS, and moves focus to the next input; typed client separating network_error from unexpected_response; Tailwind + shadcn/ui; oxlint (create-vite default); distroless non-root backend with a -healthcheck flag, pinned nginx-unprivileged, healthcheck on 127.0.0.1; pinned toolchain versions.
- Assumptions and intended behaviors: after an error, the display keeps the last successful result and the error appears above Calculate; Clear clears the inputs and the current result, but a request already sent still shows its answer; the frontend validates input before sending, so input errors like invalid_number mainly matter for direct API clients; ServeMux answers non-canonical paths with a 301 without a JSON body.
- Optional items: all implemented (power, sqrt, percentage, Docker).
- Process: one issue per branch and PR, squash merges, CI on every PR; issues and PRs share numbering. AI usage: link to docs/PROMPTS.md.
- What I would do next: a short list (e.g. E2E tests with Playwright, choosing which operand to reuse, decimal arithmetic for exact results, locale-aware decimal input).
- Time spent: leave a line "Time spent: TODO" for me to fill in.

Coverage reports:
- Generate the Go HTML report (go tool cover -html) and the function summary into docs/coverage/backend/.
- Generate the Vitest HTML report and a text summary into docs/coverage/frontend/ (configure the reporters if needed).
- In CI, upload the backend and frontend coverage reports as artifacts with actions/upload-artifact (latest major; check its inputs like the other actions).

Also: add the missing final newline to CLAUDE.md and .gitignore.
Do not edit docs/PROMPTS.md or docs/prompts/.
Run make test and make lint at the end and show the output.
```

Accepted additions from Claude Code: a `make coverage-report` target to regenerate the committed
snapshot, a "Timeouts" design decision, and a note that Esc does nothing while a request is pending.
Every curl in the README was run against the real backend and pasted unedited.

Manual changes after review:
- Committed only the coverage summaries (`docs/coverage/*/summary.txt`). The HTML reports added
  about 30 generated files to the PR and GitHub shows them as source, so they stay as CI artifacts.
- Added a screenshot of the UI and the CI badge at the top of the README, and filled in the time spent.

Transcript: [07-docs.md](prompts/07-docs.md)

## 13. Fix: keyboard shortcuts outside the form (#13)

Found by me while testing the app manually in the Docker setup (#6): Esc and Enter did
nothing when focus was outside the form, because both were handled on the form's onKeyDown.
Enter needs care, since on a focused button (e.g. a history item) it already has a native meaning.

```
Read CLAUDE.md. Fix issue #8: the Esc and Enter shortcuts only work when focus is inside the form, because they are handled on the form's onKeyDown.

Move both shortcuts to a single document-level keydown listener registered in an effect and cleaned up on unmount:
- Esc clears from anywhere.
- Enter calculates from anywhere, except when the event target is an interactive element outside the form (button, link, input, select, textarea, or anything with a role of button/radio/link), which keeps its native behavior. Inside the form, Enter keeps working as today without submitting twice.
- Do nothing while a request is loading or when Calculate is disabled.
- Ignore events with defaultPrevented.

Regression tests, written first and shown failing before the fix:
- Esc with focus on document.body clears inputs and result.
- Enter with focus on document.body calculates.
- Enter on a focused history item reuses it and does not calculate.
- Enter inside an input sends exactly one request.
Run make frontend-cover and make frontend-lint.
```

The prompt says "issue #8" by mistake: issues and PRs share numbering, and #8 is the bootstrap PR.
The bug is issue #13, fixed in PR #15.

Tests were written before the fix. The two body-focus tests failed, reproducing the bug.
The history-item and in-input tests already passed: they guard against the new global listener
causing a double action, which the old form-only handler could not do.

Accepted behavior change: Esc no longer clears while a request is pending (the Clear button still does).

Transcript: [13-keyboard-shortcuts.md](prompts/13-keyboard-shortcuts.md)

## 17. Fix company name (#17)

The company is Sezzle; I had typed "Seezle" from the start, and it spread to the repository name,
the Go module path, the UI and the docs. I renamed the repository on GitHub, which redirects the
old URL, and fixed the rest in one PR.

```
Read CLAUDE.md. Fix issue #<number>: the company is Sezzle, not Seezle. The GitHub repository was renamed to ranonbezerra/sezzle-calculator.
- Change the Go module path to github.com/ranonbezerra/sezzle-calculator/backend in go.mod and every import.
- Fix every user-facing and documentation mention: the page title, the h1 and its test, the OpenAPI title, the README (title, badge URLs, image alt text), the docker-compose project name and the CLAUDE.md title.
- Do not edit docs/prompts/ (verbatim transcripts). In docs/PROMPTS.md, leave the text inside code blocks unchanged.
- Regenerate the coverage summaries with make coverage-report.
- At the end, grep -ri seezle excluding docs/prompts/ and node_modules: only prompt text inside PROMPTS.md code blocks may remain.
Run make test and make lint and show the output.
```

The prompt went out with the `#<number>` placeholder unfilled; Claude Code identified issue #17
as the only open one. It also kept the committed frontend coverage summary, since the regenerated
table only differed in width. Prompt transcripts and the prompt text above stay verbatim, so the
old spelling remains there.

Manual changes after review: new screenshot with the corrected title, and "Sezzle" in the
translated planning prompts.

Transcript: [17-company-name.md](prompts/17-company-name.md)
