package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ranonbezerra/sezzle-calculator/backend/internal/calculator"
)

// apiError is an error response: an HTTP status plus the code and message
// sent to the client.
type apiError struct {
	status  int
	code    string
	message string
}

type errorBody struct {
	Error errorDetail `json:"error"`
}

type errorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// writeJSON writes v as the JSON response body with the given status.
// Every response goes through here, so all of them are application/json.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, e *apiError) {
	writeJSON(w, e.status, errorBody{Error: errorDetail{Code: e.code, Message: e.message}})
}

// mathError maps an error returned by a calculator operation to a response.
func mathError(err error) *apiError {
	switch {
	case errors.Is(err, calculator.ErrDivisionByZero):
		return &apiError{http.StatusUnprocessableEntity, "division_by_zero", err.Error()}
	case errors.Is(err, calculator.ErrNegativeSqrt):
		return &apiError{http.StatusUnprocessableEntity, "negative_sqrt", err.Error()}
	case errors.Is(err, calculator.ErrNonFinite):
		return &apiError{http.StatusUnprocessableEntity, "non_finite_result", err.Error()}
	default:
		return &apiError{http.StatusInternalServerError, "internal_error", "internal error"}
	}
}
