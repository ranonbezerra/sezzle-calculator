
❯ /clear

❯ <ide_opened_file>The user opened the file
  /Users/ranonbezerra/Projects/seezle-calculator/docs/PROMPTS.md in the IDE.
  This may or may not be related to the current task.</ide_opened_file>

❯ Read CLAUDE.md. Implement issue #3 (REST API). Plan first, then wait for my
  approval.

  - backend/internal/httpapi: handler registered as POST /api/v1/{operation}
  using Go 1.22+ ServeMux patterns, plus GET /healthz (move it here from main).
  - Decode with DisallowUnknownFields and a 1 KB MaxBytesReader. Use *float64
  in the request struct so missing fields are detected. Validate arity with
  calculator.Lookup: sqrt requires only "a" and rejects "b"; binary operations
  require both.
  - Error codes and statuses, all in the CLAUDE.md error format:
    invalid_json (400), invalid_number (400, e.g. 1e309 or a string where a
  number is expected), missing_operand (400), unexpected_operand (400),
  request_too_large (413), unknown_operation (404), method_not_allowed (405),
  division_by_zero, negative_sqrt and non_finite_result (422).
    The default ServeMux 404/405 responses are plain text; make sure clients
  always get the JSON error format. Update the API contract section of
  CLAUDE.md if you add a status code.
  - Content-Type application/json on every response.
  - Request logging middleware with slog (method, path, status, duration).
  - http.Server with ReadHeaderTimeout, ReadTimeout, WriteTimeout and
  IdleTimeout. Graceful shutdown on SIGINT/SIGTERM.
  - Keep main.go thin: extract config loading (PORT and defaults) and server
  construction into testable functions and cover them with tests. main only
  wires things and starts the server.
  - httptest tests: each operation, each error code and status, wrong method,
  unknown route, malformed, oversized and out-of-range bodies, the logging
  middleware.
  - docs/openapi.yaml (OpenAPI 3.1) describing the endpoints, request and
  response schemas and every error code.
    Run make backend-cover and make backend-lint and show the output.

⏺ I'll start by reading the backend code and the issue context.

  Ran 4 shell commands (ctrl+o to expand)

⏺ The codebase is small, so I've read it all directly instead of sending out
  Explore agents. Before I write the plan, there are three error-code choices
  your list doesn't cover.

⏺ User answered Claude's questions:
  ⎿  · A body with an unknown field, e.g. {"a":1,"b":2,"c":3}, is rejected by
     DisallowUnknownFields. Which error code should it return (400 either way)?
     → invalid_json (Recommended)
     · Paths outside the API, e.g. GET /foo or POST /api/v1/add/extra, need a
     JSON 404 instead of ServeMux's plain-text one. Which code should they use?
     → New not_found code (Recommended)
     · Mapping calculator errors to responses needs a default for an error that
     isn't one of the three sentinels. Today it can't happen. What should that
     branch do? → 500 internal_error (Recommended)

⏺ Updated plan
  ⎿  /plan to preview

⏺ User approved Claude's plan
  ⎿  Plan saved to:
     ~/.claude/plans/pasted-content-id-8635-read-claude-md-enumerated-reef.md ·
     /plan to edit
     Issue #3: REST API (branch feat/3-rest-api, already checked out and clean)

     Context

     The calculator core (backend/internal/calculator, with Lookup(name)
     (Operation, bool) and the sentinel errors ErrDivisionByZero,
     ErrNegativeSqrt and ErrNonFinite) exists, but nothing serves it over HTTP.
     cmd/server/main.go only serves GET /healthz, has no graceful shutdown and
     only sets ReadHeaderTimeout. This issue adds the JSON API from CLAUDE.md, a
     consistent JSON error model, server hardening, request logging, tests and
     an OpenAPI document.

     Decisions made with you:
     - An unknown body field returns invalid_json (400), and the message
       includes the decoder's detail.
     - A path outside the API returns a new not_found (404) code.
     - A calculator error that isn't one of the three sentinels returns
       internal_error (500). This adds 413 and 500 to the CLAUDE.md contract.

     New package backend/internal/httpapi

     handler.go
     - NewHandler(logger *slog.Logger) http.Handler builds a ServeMux and wraps
       it in logRequests.
     - Routes (Go 1.22+ patterns):
       - POST /api/v1/{operation} → calculate
       - GET /healthz → healthz, which returns {"status":"ok"}. GET patterns
         also match HEAD.
       - /api/v1/{operation} (no method) → 405 method_not_allowed, Allow: POST
       - /healthz (no method) → 405 method_not_allowed, Allow: GET, HEAD
       - / (catch-all) → 404 not_found
       - A pattern with a method takes precedence over the same path without
         one, so these explicit fallbacks replace the mux's plain-text 404 and
         405 responses. A 405 response always sets Allow, as RFC 9110 requires.
     - calculate:
       a. Look up the operation with
          calculator.Lookup(r.PathValue("operation")). A miss returns 404
          unknown_operation. This runs before the body is read, because an
          unknown resource takes precedence over a bad body.
       b. Decode the body with decodeRequest (see request.go).
       c. Validate arity:
          - Arity 2: a nil a or b returns missing_operand, and the message names
            the field.
          - Arity 1: a nil a returns missing_operand, and a non-nil b returns
            unexpected_operand.
       d. Call op.Apply(*a, b), using 0 for b when it is nil. On error, return
          mathError(err). On success, return 200 {"operation": name, "result":
          r}.

     request.go
     - type request struct { A *float64 json:"a"; B *float64 json:"b" }
     - const maxBodyBytes = 1 << 10
     - decodeRequest(w, r) (request, *apiError): wraps the body in
       http.MaxBytesReader, decodes with a json.Decoder that has
       DisallowUnknownFields, then requires io.EOF from a second Decode so
       trailing data such as {"a":1}{} is rejected. Errors map as follows:
       - *http.MaxBytesError → 413 request_too_large
       - *json.UnmarshalTypeError with a non-empty Field, which covers "a":"5",
         "a":true and "a":1e309 (out of range) → 400 invalid_number
       - anything else (empty body, syntax error, unknown field, a top-level
         array, number or string, trailing data) → 400 invalid_json
     - {"a": null} leaves the pointer nil, so it returns missing_operand. This
       is documented in openapi.yaml.

     errors.go
     - type apiError struct { status int; code, message string }
     - writeJSON(w, status, v) sets Content-Type: application/json and encodes
       v. Every response goes through it, which is how the "Content-Type on
       every response" requirement is met.
     - writeError(w, *apiError) writes {"error":{"code","message"}}.
     - mathError(err) *apiError uses a switch with errors.Is:
       - division_by_zero, negative_sqrt and non_finite_result → 422
       - default → 500 internal_error

     middleware.go
     - logRequests(logger, next) wraps the writer in a small statusRecorder that
       captures WriteHeader and defaults to 200. It logs logger.Info("request",
       "method", ..., "path", ..., "status", ..., "duration",
       time.Since(start)).
     - It wraps the whole mux, so 404 and 405 responses are logged too.

     backend/cmd/server (main stays wiring only)

     - config.go: type config struct { Addr string; ReadHeaderTimeout,
       ReadTimeout, WriteTimeout, IdleTimeout, ShutdownTimeout time.Duration }.
       - loadConfig(getenv func(string) string) (config, error): PORT defaults
         to 8080, and a value that isn't an integer from 1 to 65535 is an error.
       - Defaults: ReadHeaderTimeout 2s, ReadTimeout 5s, WriteTimeout 10s,
         IdleTimeout 60s, ShutdownTimeout 10s.
     - server.go:
       - newServer(cfg, handler) *http.Server sets Addr and all four timeouts.
       - run(ctx, srv, ln net.Listener, shutdownTimeout) error runs
         srv.Serve(ln) in a goroutine. When ctx is done it calls srv.Shutdown
         with a timeout context and returns nil. If Serve fails first, it
         returns that error.
     - main.go: signal.NotifyContext(ctx, SIGINT, SIGTERM), then
       loadConfig(os.Getenv), httpapi.NewHandler(slog.Default()), newServer,
       net.Listen and run. Any error is logged and exits 1. healthz, newMux and
       port are removed from main.

     Tests (table-driven, httptest)

     - internal/httpapi/handler_test.go
       - Every operation succeeds: status, body and Content-Type.
       - An error table, where each case asserts status, error.code,
         Content-Type and, for 405, Allow:
         - malformed {, empty body, trailing data, a top-level array, an unknown
           field
         - a string, a bool and 1e309 → invalid_number
         - missing a, missing b, sqrt with no a, "a": null
         - sqrt with b
         - a 1025-byte body → 413, and an exactly 1024-byte valid body (padded
           with whitespace) → 200
         - /api/v1/modulo and /api/v1/ADD → unknown_operation
         - /foo and /api/v1/add/extra → not_found
         - GET /api/v1/add, POST /healthz
         - divide by 0, sqrt(-1), 1e308 * 10
       - healthz with GET and HEAD.
     - errors_test.go: mathError for each sentinel, a wrapped sentinel, and an
       unknown error → 500.
     - middleware_test.go: a JSON slog handler writes into a buffer. The tests
       parse the record and check method, path, status and that duration is
       present, for both a 200 and an error status. They also check that the
       default status is 200 when the handler never calls WriteHeader.
     - cmd/server/config_test.go: PORT unset, a valid value, abc, 0 and 70000,
       plus the default timeouts.
     - cmd/server/server_test.go:
       - newServer copies every field.
       - run with a 127.0.0.1:0 listener serves a real GET /healthz. Cancelling
         ctx makes it return nil.
       - run with an already-closed listener returns an error.
     - cmd/server/main_test.go is deleted, because its healthz test moves to
       httpapi.

     Docs

     - docs/openapi.yaml (OpenAPI 3.1) covers:
       - POST /api/v1/{operation}, with an operation enum of the 7 names
       - GET /healthz
       - the request schema (a and b as numbers, additionalProperties: false,
         and a note on sqrt arity)
       - the Result, Health and Error schemas, with a code enum listing all 13
         codes
       - responses 200, 400, 404, 405 (with an Allow header), 413, 422 and 500,
         each with named examples for every code it can carry
     - CLAUDE.md API contract: add 413 body too large and 500 internal error to
       the status line, and list GET /healthz.

     Known edge cases (no change planned)

     - ServeMux answers non-canonical paths such as /api//v1/add with a 301
       redirect that has no JSON body. It isn't an error response, and
       intercepting it would need a custom mux.
     - statusRecorder doesn't forward the internal "close the connection" hint
       that MaxBytesReader gives the server. The 413 response is still correct,
       and the server just drains up to 256 KB of the remaining body before
       reusing the connection.

     Verification

     - make backend-cover (target: 100% in httpapi, and every function in
       cmd/server except main)
     - make backend-lint
     - A manual check with make backend-run and curl, covering a success, a 400,
       a 404, a 405 and a 413. Then Ctrl-C to confirm a graceful shutdown.
     - I'll report all of the output. Nothing gets committed.

⏺ Plan approved. Before writing the decoder, I'll check how Go 1.27's
  encoding/json reports the error cases I'm going to map.

  Ran 1 shell command (ctrl+o to expand)

⏺ Every case matches the plan, but the decoder's messages expose Go type names,
  so I'll write the client-facing errors myself. Writing the package now.

⏺ Write(internal/httpapi/errors.go)
  ⎿  Wrote 52 lines to internal/httpapi/errors.go
      1 package httpapi
      2
      3 import (
      4   "encoding/json"
      5   "errors"
      6   "net/http"
      7
      8   "github.com/ranonbezerra/seezle-calculator/backend/internal/cal
        culator"
      9 )
     … +43 lines (ctrl+o to expand)

⏺ Write(internal/httpapi/request.go)
  ⎿  Wrote 62 lines to internal/httpapi/request.go
       1 package httpapi
       2
       3 import (
       4   "encoding/json"
       5   "errors"
       6   "fmt"
       7   "io"
       8   "net/http"
       9   "strings"
      10 )
     … +52 lines (ctrl+o to expand)

⏺ Write(internal/httpapi/middleware.go)
  ⎿  Wrote 33 lines to internal/httpapi/middleware.go
       1 package httpapi
       2
       3 import (
       4   "log/slog"
       5   "net/http"
       6   "time"
       7 )
       8
       9 // statusRecorder remembers the status code written by the wrapp
         ed handler.
     … +24 lines (ctrl+o to expand)

⏺ Write(internal/httpapi/handler.go)
  ⎿  Wrote 94 lines to internal/httpapi/handler.go
       1 // Package httpapi exposes the calculator over a JSON HTTP API.
       2 package httpapi
       3
       4 import (
       5   "fmt"
       6   "log/slog"
       7   "net/http"
       8
       9   "github.com/ranonbezerra/seezle-calculator/backend/internal/ca
         lculator"
     … +85 lines (ctrl+o to expand)

⏺ Now the httpapi tests.

⏺ Write(internal/httpapi/handler_test.go)
  ⎿  Wrote 199 lines to internal/httpapi/handler_test.go
       1 package httpapi
       2
       3 import (
       4   "encoding/json"
       5   "io"
       6   "log/slog"
       7   "net/http"
       8   "net/http/httptest"
       9   "strings"
      10   "testing"
     … +189 lines (ctrl+o to expand)

⏺ Write(internal/httpapi/errors_test.go)
  ⎿  Wrote 43 lines to internal/httpapi/errors_test.go
       1 package httpapi
       2
       3 import (
       4   "errors"
       5   "fmt"
       6   "net/http"
       7   "testing"
       8
       9   "github.com/ranonbezerra/seezle-calculator/backend/internal/ca
         lculator"
     … +34 lines (ctrl+o to expand)

⏺ Write(internal/httpapi/middleware_test.go)
  ⎿  Wrote 83 lines to internal/httpapi/middleware_test.go
       1 package httpapi
       2
       3 import (
       4   "bytes"
       5   "encoding/json"
       6   "log/slog"
       7   "net/http"
       8   "net/http/httptest"
       9   "testing"
      10 )
     … +73 lines (ctrl+o to expand)

  Ran 1 shell command (ctrl+o to expand)

⏺ The httpapi tests all pass, including /api/v1/ (empty operation), which the
  catch-all routes to not_found. Next, cmd/server.

⏺ Write(cmd/server/config.go)
  ⎿  Wrote 40 lines to cmd/server/config.go
      1 package main
      2
      3 import (
      4   "fmt"
      5   "strconv"
      6   "time"
      7 )
      8
      9 const defaultPort = "8080"
     … +31 lines (ctrl+o to expand)

⏺ Write(cmd/server/server.go)
  ⎿  Wrote 44 lines to cmd/server/server.go
      1 package main
      2
      3 import (
      4   "context"
      5   "errors"
      6   "net"
      7   "net/http"
      8   "time"
      9 )
     … +35 lines (ctrl+o to expand)
⏺ Write(cmd/server/main.go)
  ⎿  Added 24 lines, removed 27 lines
       1  package main
       2  
       3  import (
       4 +  "context"                                                    
       5    "log/slog"
       5 -  "net/http"                                                   
       6 +  "net"                                                        
       7    "os"
       7 -  "time"                                                       
       8 +  "os/signal"                                                  
       9 +  "syscall"                                                    
      10 +                                                               
      11 +  "github.com/ranonbezerra/seezle-calculator/backend/internal/h
         +ttpapi"                                                        
      12  )
      13  
      10 -const defaultPort = "8080"                                     
      11 -                                                               
      14  func main() {
      13 -  addr := ":" + port()                                         
      14 -  srv := &http.Server{                                         
      15 -    Addr:              addr,                                   
      16 -    Handler:           newMux(),                               
      17 -    ReadHeaderTimeout: 5 * time.Second,                        
      15 +  ctx, stop := signal.NotifyContext(context.Background(), sysca
         +ll.SIGINT, syscall.SIGTERM)                                    
      16 +  defer stop()                                                 
      17 +                                                               
      18 +  cfg, err := loadConfig(os.Getenv)                            
      19 +  if err != nil {                                              
      20 +    fatal("load config", err)                                  
      21    }
      22 +  srv := newServer(cfg, httpapi.NewHandler(slog.Default()))    
      23  
      20 -  slog.Info("server listening", "addr", addr)                  
      21 -  err := srv.ListenAndServe()                                  
      22 -  slog.Error("server stopped", "err", err)                     
      23 -  os.Exit(1)                                                   
      24 -}                                                              
      24 +  ln, err := net.Listen("tcp", cfg.Addr)                       
      25 +  if err != nil {                                              
      26 +    fatal("listen", err)                                       
      27 +  }                                                            
      28 +  slog.Info("server listening", "addr", ln.Addr().String())    
      29  
      26 -func port() string {                                           
      27 -  if p := os.Getenv("PORT"); p != "" {                         
      28 -    return p                                                   
      30 +  if err := run(ctx, srv, ln, cfg.ShutdownTimeout); err != nil 
         +{                                                              
      31 +    fatal("server stopped", err)                               
      32    }
      30 -  return defaultPort                                           
      33 +  slog.Info("server stopped")                                  
      34  }
      35  
      33 -func newMux() *http.ServeMux {                                 
      34 -  mux := http.NewServeMux()                                    
      35 -  mux.HandleFunc("GET /healthz", healthz)                      
      36 -  return mux                                                   
      36 +func fatal(msg string, err error) {                            
      37 +  slog.Error(msg, "err", err)                                  
      38 +  os.Exit(1)                                                   
      39  }
      38 -                                                               
      39 -func healthz(w http.ResponseWriter, _ *http.Request) {         
      40 -  w.Header().Set("Content-Type", "application/json")           
      41 -  _, _ = w.Write([]byte(`{"status":"ok"}`))                    
      42 -}                                                              

⏺ run waits on Serve's result after Shutdown, but after a Shutdown Serve always
  returns ErrServerClosed, so that branch can't be reached. I'm removing it.

⏺ Update(cmd/server/server.go)
  ⎿  Added 1 line, removed 7 lines
      34  
      35    shutdownCtx, cancel := context.WithTimeout(context.Background
          (), shutdownTimeout)
      36    defer cancel()
      37 -  if err := srv.Shutdown(shutdownCtx); err != nil {            
      38 -    return err                                                 
      39 -  }                                                            
      40 -  if err := <-serveErr; !errors.Is(err, http.ErrServerClosed) {
      41 -    return err                                                 
      42 -  }                                                            
      43 -  return nil                                                   
      37 +  return srv.Shutdown(shutdownCtx)                             
      38  }

  Ran 1 shell command (ctrl+o to expand)

⏺ git rm also staged the deletion. You manage the index yourself, so I'll
  unstage it and leave the file deleted only in the working tree.

  Ran 1 shell command (ctrl+o to expand)

⏺ Write(cmd/server/config_test.go)
  ⎿  Wrote 70 lines to cmd/server/config_test.go
       1 package main
       2
       3 import (
       4   "testing"
       5   "time"
       6 )
       7
       8 func TestLoadConfig(t *testing.T) {
       9   tests := []struct {
      10     name     string
     … +60 lines (ctrl+o to expand)

⏺ Write(cmd/server/server_test.go)
  ⎿  Wrote 151 lines to cmd/server/server_test.go
       1 package main
       2
       3 import (
       4   "context"
       5   "errors"
       6   "net"
       7   "net/http"
       8   "testing"
       9   "time"
      10 )
     … +141 lines (ctrl+o to expand)

⏺ The in-flight test releases the handler right after cancel(), so it could pass
  before Shutdown has even started. I'm changing it to wait until the listener
  refuses new connections, so it actually shows the request draining.

⏺ Update(cmd/server/server_test.go)
  ⎿  Added 1 line
       99  
      100    <-started
      101    cancel()
      102 +  waitUntilClosed(t, ln.Addr().String())                      
      103    close(release)
      104  
      105    if got := <-respStatus; got != http.StatusOK {

⏺ Update(cmd/server/server_test.go)
  ⎿  Added 16 lines
      42    return ln
      43  }
      44  
      45 +// waitUntilClosed blocks until addr refuses connections, which
         + shows that                                                    
      46 +// shutdown has started.                                       
      47 +func waitUntilClosed(t *testing.T, addr string) {              
      48 +  t.Helper()                                                   
      49 +  deadline := time.Now().Add(5 * time.Second)                  
      50 +  for time.Now().Before(deadline) {                            
      51 +    conn, err := net.Dial("tcp", addr)                         
      52 +    if err != nil {                                            
      53 +      return                                                   
      54 +    }                                                          
      55 +    _ = conn.Close()                                           
      56 +    time.Sleep(5 * time.Millisecond)                           
      57 +  }                                                            
      58 +  t.Fatal("listener still open after cancelling the context")  
      59 +}                                                              
      60 +                                                               
      61  func TestRunServesUntilContextIsDone(t *testing.T) {
      62    ln := listen(t)
      63    srv := &http.Server{

  Ran 1 shell command (ctrl+o to expand)

⏺ Every package passes over three runs under -race. Now the docs: the OpenAPI
  spec and the CLAUDE.md contract.

⏺ Write(~/Projects/seezle-calculator/docs/openapi.yaml)
  ⎿  Wrote 215 lines to ~/Projects/seezle-calculator/docs/openapi.yaml
       1 openapi: 3.1.0
       2 info:
       3   title: Seezle Calculator API
       4   version: 1.0.0
       5   description: |
       6     Arithmetic over IEEE 754 double-precision numbers.
       7
       8     Every response, including errors, has `Content-Type: applica
         tion/json`.
       9     Errors share one shape: `{"error": {"code": "...", "message"
     … +207 lines (~8 KB) (ctrl+o to expand)

⏺ Update(~/Projects/seezle-calculator/CLAUDE.md)
  ⎿  Added 4 lines, removed 1 line
      10  - percentage(a, b) = a percent of b = a*b/100.
      11  - Success: 200 {"operation": "...", "result": number}
      12  - Errors: {"error": {"code": "snake_case_code", "message": "...
          "}}
      13 -  400 invalid input, 422 math error, 404 unknown operation, 405
         - wrong method.                                                 
      13 +  400 invalid input, 422 math error, 404 unknown operation, 405
         + wrong method,                                                 
      14 +  413 body too large (limit 1 KB), 500 internal error.         
      15 +- GET /healthz returns 200 {"status": "ok"}.                   
      16 +- Full contract with every error code: docs/openapi.yaml.      
      17  
      18  ## Rules
      19  - One issue, one branch, one PR. Stay inside the issue scope; d
          o not add features.

⏺ The spec's code enum has 12 entries, not the 13 the plan said, since it was a
  miscount. I'm checking that the YAML parses and that the codes match what the
  Go code emits, then running the requested targets.

  Ran 3 shell commands (ctrl+o to expand)

⏺ Both pass. Last, a manual end-to-end check against the real binary, finishing
  with SIGINT.

  Ran 1 shell command (ctrl+o to expand)

⏺ Issue #3 is implemented. make backend-cover and make backend-lint both pass,
  and a manual run against the real binary behaved as specified. Nothing is
  committed or staged.

  What changed
  - New backend/internal/httpapi:
    - handler.go (backend/internal/httpapi/handler.go) sets up the routes, the
      arity check and healthz. It also registers fallback routes so that unknown
      paths and wrong methods get JSON 404 and 405 responses instead of
      ServeMux's plain text. Every 405 carries an Allow header.
    - request.go (backend/internal/httpapi/request.go) decodes the body with the
      1 KB limit and DisallowUnknownFields, and rejects trailing data such as
      {"a":1,"b":2}{}.
    - errors.go (backend/internal/httpapi/errors.go) holds the JSON writer
      (which sets the Content-Type on every response) and maps calculator errors
      to 422/500.
    - middleware.go (backend/internal/httpapi/middleware.go) logs method, path,
      status and duration for each request.
  - cmd/server:
    - config.go (backend/cmd/server/config.go): PORT defaults to 8080 and must
      be an integer from 1 to 65535; the timeouts are fixed defaults.
    - server.go (backend/cmd/server/server.go): newServer and run, which serves
      until SIGINT/SIGTERM and then shuts down gracefully.
    - main.go (backend/cmd/server/main.go) now only wires these together.
    - The old main_test.go is deleted, and its healthz test moved to httpapi.
  - Docs: docs/openapi.yaml documents both endpoints, all the schemas, and every
    status with an example per error code. CLAUDE.md now lists 413 and 500,
    /healthz, and points to the spec.

  Things to check
  - Error messages: I wrote my own messages instead of passing on
    encoding/json's, because its messages include internal Go type names (req.a
    of type float64). The messages are tested exactly.
  - 12 codes, not 13: the plan said 13, which was my miscount. A script
    confirmed the spec's code list matches the codes the Go code returns
    exactly.
  - {"a": null} and a bare null body both count as a missing operand
    (missing_operand).
  - The in-flight shutdown test waits until the listener stops accepting
    connections before letting the request finish, so it really shows a request
    completing during shutdown rather than before it started.
  - Two edge cases I left as they are: ServeMux's 301 redirect for odd paths
    like /api//v1/add has no JSON body, and the logging wrapper hides the
    server's hint to close the connection after a 413.

  make backend-cover
  ok  .../backend/cmd/server            coverage: 50.0% of statements
  ok  .../backend/internal/calculator   (cached) coverage: 100.0% of statements
  ok  .../backend/internal/httpapi      coverage: 100.0% of statements
  cmd/server/config.go:23:   loadConfig  100.0%
  cmd/server/main.go:14:     main          0.0%
  cmd/server/main.go:36:     fatal         0.0%
  cmd/server/server.go:11:   newServer   100.0%
  cmd/server/server.go:24:   run         100.0%
  internal/httpapi/*         every function 100.0%
  total:                     (statements) 88.4%
  Only main and fatal are untested, as the plan intended. The cmd/server package
  shows 50% because of those two functions.

  make backend-lint
  cd backend && go vet ./... && golangci-lint run ./...
  0 issues.

  Manual run (built binary, PORT=18080, curl)
  - Requests returned 200, 400 invalid_number, 404 unknown_operation, 405 with
    Allow: POST, 413, 404 not_found and a 200 from healthz.
  - Every response had Content-Type: application/json, and each request produced
    one log line.
  - SIGINT logged server stopped and exited cleanly.
  - PORT=abc logged invalid PORT "abc" and exited with status 1.