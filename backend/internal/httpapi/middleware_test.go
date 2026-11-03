package httpapi

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLogRequests(t *testing.T) {
	tests := []struct {
		name       string
		handler    http.HandlerFunc
		method     string
		path       string
		wantStatus int
	}{
		{
			name:       "explicit status",
			handler:    func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusTeapot) },
			method:     http.MethodPost,
			path:       "/api/v1/add",
			wantStatus: http.StatusTeapot,
		},
		{
			name:       "implicit 200",
			handler:    func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("ok")) },
			method:     http.MethodGet,
			path:       "/healthz",
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&buf, nil))
			rec := httptest.NewRecorder()

			logRequests(logger, tt.handler).ServeHTTP(rec, httptest.NewRequest(tt.method, tt.path, nil))

			if rec.Code != tt.wantStatus {
				t.Errorf("response status = %d, want %d", rec.Code, tt.wantStatus)
			}
			var entry struct {
				Msg      string   `json:"msg"`
				Method   string   `json:"method"`
				Path     string   `json:"path"`
				Status   int      `json:"status"`
				Duration *float64 `json:"duration"`
			}
			if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
				t.Fatalf("decode log %q: %v", buf.String(), err)
			}
			if entry.Msg != "request" || entry.Method != tt.method || entry.Path != tt.path || entry.Status != tt.wantStatus {
				t.Errorf("log = %+v, want msg=request method=%s path=%s status=%d", entry, tt.method, tt.path, tt.wantStatus)
			}
			if entry.Duration == nil || *entry.Duration < 0 {
				t.Errorf("duration = %v, want a non-negative value", entry.Duration)
			}
		})
	}
}

func TestNewHandlerLogsErrorResponses(t *testing.T) {
	var buf bytes.Buffer
	handler := NewHandler(slog.New(slog.NewJSONHandler(&buf, nil)))

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/nope", nil))

	var entry struct {
		Path   string `json:"path"`
		Status int    `json:"status"`
	}
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("decode log %q: %v", buf.String(), err)
	}
	if entry.Path != "/nope" || entry.Status != http.StatusNotFound {
		t.Errorf("log = %+v, want path=/nope status=404", entry)
	}
}
