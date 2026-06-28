package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		wantStatus int
		wantBody   string
	}{
		{name: "GET", method: http.MethodGet, wantStatus: http.StatusOK, wantBody: "ok"},
		{name: "HEAD", method: http.MethodHead, wantStatus: http.StatusOK, wantBody: ""},
		{name: "POST", method: http.MethodPost, wantStatus: http.StatusMethodNotAllowed, wantBody: "Methode nicht erlaubt\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, "/healthz", nil)
			response := httptest.NewRecorder()

			HealthHandler(response, request)

			if response.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", response.Code, tt.wantStatus)
			}
			if response.Body.String() != tt.wantBody {
				t.Errorf("body = %q, want %q", response.Body.String(), tt.wantBody)
			}
		})
	}
}
