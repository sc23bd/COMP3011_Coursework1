package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// doRequest executes an HTTP request against the router and returns the recorder.
func doRequest(r *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// doRequestWithHeader executes an HTTP request with an additional header.
func doRequestWithHeader(r *gin.Engine, method, path string, body interface{}, headerKey, headerValue string) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if headerKey != "" {
		req.Header.Set(headerKey, headerValue)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// assertStatus is a convenience helper to check the HTTP status code.
func assertStatus(t interface {
	Helper()
	Fatalf(string, ...interface{})
}, w *httptest.ResponseRecorder, want int) {
	t.Helper()
	if w.Code != want {
		t.Fatalf("expected status %d, got %d: %s", want, w.Code, w.Body.String())
	}
}

// decodeJSON decodes the response body into dst.
func decodeJSON(t interface {
	Helper()
	Fatalf(string, ...interface{})
}, w *httptest.ResponseRecorder, dst interface{}) {
	t.Helper()
	if err := json.NewDecoder(w.Body).Decode(dst); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
}

// checkHeader verifies that a response header is non-empty.
func checkHeader(t interface {
	Helper()
	Fatalf(string, ...interface{})
}, w *httptest.ResponseRecorder, name string) {
	t.Helper()
	if w.Header().Get(name) == "" {
		t.Fatalf("expected response header %q to be set", name)
	}
}

// notFound sends a GET request and asserts a 404 response.
func notFound(t interface {
	Helper()
	Fatalf(string, ...interface{})
}, r *gin.Engine, path string) {
	t.Helper()
	w := doRequest(r, http.MethodGet, path, nil)
	assertStatus(t, w, http.StatusNotFound)
}
