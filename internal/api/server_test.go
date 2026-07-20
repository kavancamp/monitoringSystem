package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthz(t *testing.T) {
	server := NewServer(nil)
	handler := server.Routes()

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			recorder.Code,
		)
	}

	expectedBody := "ok"
	actualBody := recorder.Body.String()

	if actualBody != expectedBody {
		t.Errorf(
			"expected response body %q, got %q",
			expectedBody,
			actualBody,
		)
	}
}
