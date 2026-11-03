package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// maxBodyBytes is the largest request body accepted.
const maxBodyBytes = 1 << 10

// request is the body of a calculation. Pointers distinguish a missing
// operand from an explicit zero.
type request struct {
	A *float64 `json:"a"`
	B *float64 `json:"b"`
}

// decodeRequest reads a single JSON object with only the known fields from
// the body of r.
func decodeRequest(w http.ResponseWriter, r *http.Request) (request, *apiError) {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	dec.DisallowUnknownFields()

	var req request
	if err := dec.Decode(&req); err != nil {
		return request{}, decodeError(err)
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			err = errors.New("unexpected data after the JSON object")
		}
		return request{}, decodeError(err)
	}
	return req, nil
}

func decodeError(err error) *apiError {
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		msg := fmt.Sprintf("request body must not exceed %d bytes", maxBodyBytes)
		return &apiError{http.StatusRequestEntityTooLarge, "request_too_large", msg}
	}

	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &typeErr) {
		if typeErr.Field != "" {
			msg := fmt.Sprintf("field %q must be a finite number", typeErr.Field)
			return &apiError{http.StatusBadRequest, "invalid_number", msg}
		}
		return &apiError{http.StatusBadRequest, "invalid_json", "request body must be a JSON object"}
	}

	if errors.Is(err, io.EOF) {
		return &apiError{http.StatusBadRequest, "invalid_json", "request body is empty"}
	}
	msg := "invalid JSON body: " + strings.TrimPrefix(err.Error(), "json: ")
	return &apiError{http.StatusBadRequest, "invalid_json", msg}
}
