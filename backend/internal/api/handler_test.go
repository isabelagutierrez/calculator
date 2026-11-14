package api

import (
	"bytes"
	"encoding/json"
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

	handler := NewHandler(Config{})
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

	handler := NewHandler(Config{})
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
	NewHandler(Config{}).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.Code)
	}
}

func TestJSONRoutingErrors(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
		body   string
		status int
	}{
		{"invalid JSON", http.MethodPost, "/api/add", "not-json", http.StatusBadRequest},
		{"missing field", http.MethodPost, "/api/add", `{"left":8}`, http.StatusBadRequest},
		{"wrong type", http.MethodPost, "/api/add", `{"left":"8","right":4}`, http.StatusBadRequest},
		{"method not allowed", http.MethodGet, "/api/add", "", http.StatusMethodNotAllowed},
		{"not found", http.MethodGet, "/does-not-exist", "", http.StatusNotFound},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.path, strings.NewReader(test.body))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			NewHandler(Config{}).ServeHTTP(response, request)
			if response.Code != test.status {
				t.Fatalf("status = %d, want %d", response.Code, test.status)
			}
			var payload map[string]string
			if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if payload["error"] == "" {
				t.Fatalf("response has no error: %#v", payload)
			}
		})
	}
}

func TestCalculateNegativeExpression(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/api/calculate", strings.NewReader(`{"expression":"-2 + 3"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	NewHandler(Config{}).ServeHTTP(response, request)

	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"result":1`) {
		t.Fatalf("status = %d, body = %q", response.Code, response.Body.String())
	}
}

func TestRequestProtection(t *testing.T) {
	tests := []struct {
		name   string
		body   string
		header string
		status int
	}{
		{"missing content type", `{"left":8,"right":4}`, "", http.StatusUnsupportedMediaType},
		{"multiple JSON objects", `{"left":8,"right":4} {"left":1,"right":2}`, "application/json", http.StatusBadRequest},
	}
	handler := NewHandler(Config{CORSAllowedOrigin: "https://example.test"})
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/api/add", strings.NewReader(test.body))
			if test.header != "" {
				request.Header.Set("Content-Type", test.header)
			}
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != test.status {
				t.Fatalf("status = %d, want %d", response.Code, test.status)
			}
			var payload errorResponse
			if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
				t.Fatalf("decode error response: %v", err)
			}
			if payload.Error == "" {
				t.Fatal("expected JSON error")
			}
			if got := response.Header().Get("Access-Control-Allow-Origin"); got != "https://example.test" {
				t.Fatalf("CORS origin = %q", got)
			}
		})
	}
}

func TestLargeBodyReturns413AndHealthStillWorks(t *testing.T) {
	handler := NewHandler(Config{})
	body := `{"left":` + strings.Repeat("9", 6*1024*1024) + `,"right":1}`
	request := httptest.NewRequest(http.MethodPost, "/api/add", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("large body status = %d, want 413", response.Code)
	}

	health := httptest.NewRecorder()
	handler.ServeHTTP(health, httptest.NewRequest(http.MethodGet, "/health", nil))
	if health.Code != http.StatusOK {
		t.Fatalf("health status after large body = %d, want 200", health.Code)
	}
}

func TestMethodNotAllowedIncludesAllowHeader(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/add", nil)
	response := httptest.NewRecorder()
	NewHandler(Config{}).ServeHTTP(response, request)
	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", response.Code)
	}
	if got := response.Header().Get("Allow"); got != http.MethodPost {
		t.Fatalf("Allow = %q, want POST", got)
	}
}

func TestPanicRecovery(t *testing.T) {
	handler := withMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("private details")
	}), Config{})
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", response.Code)
	}
	if strings.Contains(response.Body.String(), "private details") {
		t.Fatal("panic details leaked in response")
	}
}

func TestBodyLimitIsAppliedToRawBody(t *testing.T) {
	body := bytes.NewReader(bytes.Repeat([]byte("x"), int(maxBodyBytes)+1))
	request := httptest.NewRequest(http.MethodPost, "/api/add", body)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	NewHandler(Config{}).ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest && response.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want body validation error", response.Code)
	}
}
