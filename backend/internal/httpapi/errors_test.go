package httpapi

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/ranonbezerra/sezzle-calculator/backend/internal/calculator"
)

func TestMathError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{name: "division by zero", err: calculator.ErrDivisionByZero, wantStatus: http.StatusUnprocessableEntity, wantCode: "division_by_zero"},
		{name: "negative sqrt", err: calculator.ErrNegativeSqrt, wantStatus: http.StatusUnprocessableEntity, wantCode: "negative_sqrt"},
		{name: "non-finite result", err: calculator.ErrNonFinite, wantStatus: http.StatusUnprocessableEntity, wantCode: "non_finite_result"},
		{name: "wrapped sentinel", err: fmt.Errorf("divide: %w", calculator.ErrDivisionByZero), wantStatus: http.StatusUnprocessableEntity, wantCode: "division_by_zero"},
		{name: "unknown error", err: errors.New("boom"), wantStatus: http.StatusInternalServerError, wantCode: "internal_error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mathError(tt.err)

			if got.status != tt.wantStatus || got.code != tt.wantCode {
				t.Errorf("mathError() = %d %s, want %d %s", got.status, got.code, tt.wantStatus, tt.wantCode)
			}
		})
	}
}

func TestMathErrorHidesUnknownDetails(t *testing.T) {
	got := mathError(errors.New("secret detail"))

	if got.message != "internal error" {
		t.Errorf("message = %q, want %q", got.message, "internal error")
	}
}
