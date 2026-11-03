// Package httpapi exposes the calculator over a JSON HTTP API.
package httpapi

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/ranonbezerra/sezzle-calculator/backend/internal/calculator"
)

type result struct {
	Operation string  `json:"operation"`
	Result    float64 `json:"result"`
}

// NewHandler returns the API handler with request logging.
//
// The method-less and catch-all patterns replace the plain-text 404 and 405
// responses of http.ServeMux with JSON errors. A pattern with a method takes
// precedence over the same path without one.
func NewHandler(logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/{operation}", calculate)
	mux.HandleFunc("/api/v1/{operation}", methodNotAllowed("POST"))
	mux.HandleFunc("GET /healthz", healthz)
	mux.HandleFunc("/healthz", methodNotAllowed("GET, HEAD"))
	mux.HandleFunc("/", notFound)
	return logRequests(logger, mux)
}

func calculate(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("operation")
	op, ok := calculator.Lookup(name)
	if !ok {
		writeError(w, &apiError{http.StatusNotFound, "unknown_operation", fmt.Sprintf("unknown operation %q", name)})
		return
	}

	req, apiErr := decodeRequest(w, r)
	if apiErr != nil {
		writeError(w, apiErr)
		return
	}
	if apiErr := checkArity(op.Arity, req); apiErr != nil {
		writeError(w, apiErr)
		return
	}

	var b float64
	if req.B != nil {
		b = *req.B
	}
	res, err := op.Apply(*req.A, b)
	if err != nil {
		writeError(w, mathError(err))
		return
	}
	writeJSON(w, http.StatusOK, result{Operation: name, Result: res})
}

// checkArity reports a missing operand, or "b" sent to a unary operation.
func checkArity(arity int, req request) *apiError {
	if req.A == nil {
		return missingOperand("a")
	}
	if arity == 1 && req.B != nil {
		return &apiError{http.StatusBadRequest, "unexpected_operand", `this operation takes only "a"`}
	}
	if arity == 2 && req.B == nil {
		return missingOperand("b")
	}
	return nil
}

func missingOperand(field string) *apiError {
	return &apiError{http.StatusBadRequest, "missing_operand", fmt.Sprintf("field %q is required", field)}
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func methodNotAllowed(allow string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Allow", allow)
		msg := fmt.Sprintf("method %s is not allowed, use %s", r.Method, allow)
		writeError(w, &apiError{http.StatusMethodNotAllowed, "method_not_allowed", msg})
	}
}

func notFound(w http.ResponseWriter, r *http.Request) {
	writeError(w, &apiError{http.StatusNotFound, "not_found", fmt.Sprintf("no route for %s", r.URL.Path)})
}
