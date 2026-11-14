package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOperationEndpoints(t *testing.T) {
	tests := []struct {
		name string
		path string
		body string
		want string
	}{
		{"addition", "/api/add", `{"left":8,"right":4}`, `"result":12`},
		{"subtraction", "/api/subtract", `{"left":8,"right":4}`, `"result":4`},
		{"multiplication", "/api/multiply", `{"left":8,"right":4}`, `"result":32`},
		{"division", "/api/divide", `{"left":8,"right":4}`, `"result":2`},
		{"power", "/api/power", `{"left":2,"right":3}`, `"result":8`},
		{"square root", "/api/square-root", `{"value":81}`, `"result":9`},
		{"percentage", "/api/percentage", `{"value":25}`, `"result":0.25`},
		{"expression", "/api/calculate", `{"expression":"8 × 8 + 2"}`, `"result":66`},
	}

	handler := NewHandler()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, test.path, strings.NewReader(test.body))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)

			if response.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200", response.Code)
			}
			if !strings.Contains(response.Body.String(), test.want) {
				t.Fatalf("body = %q, want it to contain %q", response.Body.String(), test.want)
			}
		})
	}
}

func TestValidationErrors(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		body   string
		status int
	}{
		{"division by zero", "/api/divide", `{"left":8,"right":0}`, http.StatusUnprocessableEntity},
		{"negative square root", "/api/square-root", `{"value":-1}`, http.StatusUnprocessableEntity},
		{"missing operand", "/api/add", `{"left":8}`, http.StatusBadRequest},
		{"invalid JSON", "/api/add", `not-json`, http.StatusBadRequest},
		{"unknown field", "/api/add", `{"left":8,"right":4,"extra":1}`, http.StatusBadRequest},
		{"invalid expression", "/api/calculate", `{"expression":"8 ×"}`, http.StatusUnprocessableEntity},
	}

	handler := NewHandler()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, test.path, strings.NewReader(test.body))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)

			if response.Code != test.status {
				t.Fatalf("status = %d, want %d", response.Code, test.status)
			}
		})
	}
}

func TestHealthEndpoint(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	response := httptest.NewRecorder()
	NewHandler().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.Code)
	}
}

func TestCalculateNegativeExpression(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/api/calculate", strings.NewReader(`{"expression":"-2 + 3"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	NewHandler().ServeHTTP(response, request)

	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"result":1`) {
		t.Fatalf("status = %d, body = %q", response.Code, response.Body.String())
	}
}
