package httpapi

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func serve(t *testing.T, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	rec := httptest.NewRecorder()
	NewHandler(slog.New(slog.NewTextHandler(io.Discard, nil))).ServeHTTP(rec, req)
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	return rec
}

func TestOperations(t *testing.T) {
	tests := []struct {
		operation string
		body      string
		want      float64
	}{
		{operation: "add", body: `{"a": 2, "b": 3}`, want: 5},
		{operation: "subtract", body: `{"a": 2, "b": 3}`, want: -1},
		{operation: "multiply", body: `{"a": 2.5, "b": -4}`, want: -10},
		{operation: "divide", body: `{"a": 7, "b": 2}`, want: 3.5},
		{operation: "power", body: `{"a": 2, "b": 10}`, want: 1024},
		{operation: "sqrt", body: `{"a": 81}`, want: 9},
		{operation: "percentage", body: `{"a": 10, "b": 30}`, want: 3},
		{operation: "add", body: `{"a": 0, "b": 0}`, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.operation+" "+tt.body, func(t *testing.T) {
			rec := serve(t, http.MethodPost, "/api/v1/"+tt.operation, tt.body)

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200; body %s", rec.Code, rec.Body)
			}
			var got result
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if got.Operation != tt.operation || got.Result != tt.want {
				t.Errorf("body = %+v, want {Operation:%s Result:%v}", got, tt.operation, tt.want)
			}
		})
	}
}

// padTo pads a JSON body with trailing spaces to exactly n bytes.
func padTo(body string, n int) string {
	return body + strings.Repeat(" ", n-len(body))
}

func TestBodySizeLimit(t *testing.T) {
	tests := []struct {
		name       string
		size       int
		wantStatus int
	}{
		{name: "exactly the limit", size: maxBodyBytes, wantStatus: http.StatusOK},
		{name: "one byte over the limit", size: maxBodyBytes + 1, wantStatus: http.StatusRequestEntityTooLarge},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := serve(t, http.MethodPost, "/api/v1/add", padTo(`{"a": 1, "b": 2}`, tt.size))

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d; body %s", rec.Code, tt.wantStatus, rec.Body)
			}
		})
	}
}

func TestErrors(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
		wantCode   string
		wantAllow  string
	}{
		{name: "malformed JSON", path: "/api/v1/add", body: `{"a": 1,`, wantStatus: 400, wantCode: "invalid_json"},
		{name: "empty body", path: "/api/v1/add", body: ``, wantStatus: 400, wantCode: "invalid_json"},
		{name: "trailing object", path: "/api/v1/add", body: `{"a": 1, "b": 2}{}`, wantStatus: 400, wantCode: "invalid_json"},
		{name: "trailing garbage", path: "/api/v1/add", body: `{"a": 1, "b": 2} x`, wantStatus: 400, wantCode: "invalid_json"},
		{name: "top-level array", path: "/api/v1/add", body: `[1, 2]`, wantStatus: 400, wantCode: "invalid_json"},
		{name: "top-level number", path: "/api/v1/add", body: `5`, wantStatus: 400, wantCode: "invalid_json"},
		{name: "unknown field", path: "/api/v1/add", body: `{"a": 1, "b": 2, "c": 3}`, wantStatus: 400, wantCode: "invalid_json"},
		{name: "string operand", path: "/api/v1/add", body: `{"a": "1", "b": 2}`, wantStatus: 400, wantCode: "invalid_number"},
		{name: "bool operand", path: "/api/v1/add", body: `{"a": 1, "b": true}`, wantStatus: 400, wantCode: "invalid_number"},
		{name: "object operand", path: "/api/v1/add", body: `{"a": {}, "b": 2}`, wantStatus: 400, wantCode: "invalid_number"},
		{name: "out-of-range operand", path: "/api/v1/add", body: `{"a": 1e309, "b": 2}`, wantStatus: 400, wantCode: "invalid_number"},
		{name: "missing a", path: "/api/v1/add", body: `{"b": 2}`, wantStatus: 400, wantCode: "missing_operand"},
		{name: "missing b", path: "/api/v1/add", body: `{"a": 1}`, wantStatus: 400, wantCode: "missing_operand"},
		{name: "null a", path: "/api/v1/add", body: `{"a": null, "b": 2}`, wantStatus: 400, wantCode: "missing_operand"},
		{name: "empty object", path: "/api/v1/divide", body: `{}`, wantStatus: 400, wantCode: "missing_operand"},
		{name: "sqrt missing a", path: "/api/v1/sqrt", body: `{}`, wantStatus: 400, wantCode: "missing_operand"},
		{name: "sqrt with b", path: "/api/v1/sqrt", body: `{"a": 4, "b": 2}`, wantStatus: 400, wantCode: "unexpected_operand"},
		{name: "oversized body", path: "/api/v1/add", body: `{"a": 1, "b": 2, "pad": "` + strings.Repeat("x", maxBodyBytes) + `"}`, wantStatus: 413, wantCode: "request_too_large"},
		{name: "unknown operation", path: "/api/v1/modulo", body: `{"a": 1, "b": 2}`, wantStatus: 404, wantCode: "unknown_operation"},
		{name: "operation is case sensitive", path: "/api/v1/ADD", body: `{"a": 1, "b": 2}`, wantStatus: 404, wantCode: "unknown_operation"},
		{name: "unknown operation wins over bad body", path: "/api/v1/modulo", body: `{`, wantStatus: 404, wantCode: "unknown_operation"},
		{name: "unknown route", path: "/foo", wantStatus: 404, wantCode: "not_found"},
		{name: "extra path segment", path: "/api/v1/add/extra", body: `{"a": 1, "b": 2}`, wantStatus: 404, wantCode: "not_found"},
		{name: "missing operation", path: "/api/v1/", body: `{"a": 1, "b": 2}`, wantStatus: 404, wantCode: "not_found"},
		{name: "GET on operation", method: http.MethodGet, path: "/api/v1/add", wantStatus: 405, wantCode: "method_not_allowed", wantAllow: "POST"},
		{name: "PUT on operation", method: http.MethodPut, path: "/api/v1/add", body: `{"a": 1, "b": 2}`, wantStatus: 405, wantCode: "method_not_allowed", wantAllow: "POST"},
		{name: "POST on healthz", path: "/healthz", wantStatus: 405, wantCode: "method_not_allowed", wantAllow: "GET, HEAD"},
		{name: "division by zero", path: "/api/v1/divide", body: `{"a": 1, "b": 0}`, wantStatus: 422, wantCode: "division_by_zero"},
		{name: "negative sqrt", path: "/api/v1/sqrt", body: `{"a": -1}`, wantStatus: 422, wantCode: "negative_sqrt"},
		{name: "overflow", path: "/api/v1/multiply", body: `{"a": 1e308, "b": 10}`, wantStatus: 422, wantCode: "non_finite_result"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method := tt.method
			if method == "" {
				method = http.MethodPost
			}
			rec := serve(t, method, tt.path, tt.body)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			var got errorBody
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatalf("decode body %q: %v", rec.Body, err)
			}
			if got.Error.Code != tt.wantCode {
				t.Errorf("code = %q, want %q", got.Error.Code, tt.wantCode)
			}
			if got.Error.Message == "" {
				t.Error("message is empty")
			}
			if allow := rec.Header().Get("Allow"); allow != tt.wantAllow {
				t.Errorf("Allow = %q, want %q", allow, tt.wantAllow)
			}
		})
	}
}

func TestErrorMessages(t *testing.T) {
	tests := []struct {
		name string
		path string
		body string
		want string
	}{
		{name: "invalid number names the field", path: "/api/v1/add", body: `{"a": 1, "b": "2"}`, want: `field "b" must be a finite number`},
		{name: "missing operand names the field", path: "/api/v1/add", body: `{"a": 1}`, want: `field "b" is required`},
		{name: "unknown field names the field", path: "/api/v1/add", body: `{"a": 1, "b": 2, "c": 3}`, want: `invalid JSON body: unknown field "c"`},
		{name: "trailing data", path: "/api/v1/add", body: `{"a": 1, "b": 2}{}`, want: "invalid JSON body: unexpected data after the JSON object"},
		{name: "empty body", path: "/api/v1/add", body: ``, want: "request body is empty"},
		{name: "top-level array", path: "/api/v1/add", body: `[]`, want: "request body must be a JSON object"},
		{name: "too large", path: "/api/v1/add", body: strings.Repeat(" ", maxBodyBytes+1), want: "request body must not exceed 1024 bytes"},
		{name: "unknown operation", path: "/api/v1/modulo", body: `{}`, want: `unknown operation "modulo"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := serve(t, http.MethodPost, tt.path, tt.body)

			var got errorBody
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatalf("decode body %q: %v", rec.Body, err)
			}
			if got.Error.Message != tt.want {
				t.Errorf("message = %q, want %q", got.Error.Message, tt.want)
			}
		})
	}
}

func TestHealthz(t *testing.T) {
	for _, method := range []string{http.MethodGet, http.MethodHead} {
		t.Run(method, func(t *testing.T) {
			rec := serve(t, method, "/healthz", "")

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200", rec.Code)
			}
			if got := strings.TrimSpace(rec.Body.String()); got != `{"status":"ok"}` {
				t.Errorf("body = %q, want {\"status\":\"ok\"}", got)
			}
		})
	}
}
