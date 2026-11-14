package api

import (
	"errors"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/isabelagutierrez/calculator/backend/internal/calculator"
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

func TestHTTPValidationAndRouting(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
		body   string
		header string
		status int
	}{
		{"not found", http.MethodGet, "/missing", "", "", http.StatusNotFound},
		{"method", http.MethodGet, "/api/add", "", "", http.StatusMethodNotAllowed},
		{"media type", http.MethodPost, "/api/add", `{}`, "text/plain", http.StatusUnsupportedMediaType},
		{"empty", http.MethodPost, "/api/add", "", "application/json", http.StatusBadRequest},
		{"wrong type", http.MethodPost, "/api/add", `{"left":"8","right":2}`, "application/json", http.StatusBadRequest},
		{"two objects", http.MethodPost, "/api/add", `{"left":1,"right":2}{"left":3}`, "application/json", http.StatusBadRequest},
		{"empty expression", http.MethodPost, "/api/calculate", `{"expression":" "}`, "application/json", http.StatusBadRequest},
		{"charset", http.MethodPost, "/api/add", `{"left":1,"right":2}`, "application/json; charset=utf-8", http.StatusOK},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.path, strings.NewReader(test.body))
			if test.header != "" {
				request.Header.Set("Content-Type", test.header)
			}
			response := httptest.NewRecorder()
			NewHandler().ServeHTTP(response, request)
			if response.Code != test.status {
				t.Fatalf("status = %d, want %d", response.Code, test.status)
			}
			if !strings.HasPrefix(response.Header().Get("Content-Type"), "application/json") && test.status != http.StatusNoContent {
				t.Fatal("expected JSON response")
			}
		})
	}
}

func TestBodyLimitAndCORS(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/api/add", strings.NewReader(`{"left":`+strings.Repeat("1", 9000)+`,"right":1}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	NewHandler(apiConfig()).ServeHTTP(response, request)
	if response.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d", response.Code)
	}
	options := httptest.NewRequest(http.MethodOptions, "/api/add", nil)
	optionsResponse := httptest.NewRecorder()
	NewHandler(apiConfig()).ServeHTTP(optionsResponse, options)
	if optionsResponse.Code != http.StatusNoContent || optionsResponse.Header().Get("Access-Control-Max-Age") == "" {
		t.Fatal("invalid CORS preflight")
	}
}

func TestErrorMappingAndRecovery(t *testing.T) {
	for _, err := range []error{calculator.ErrInvalidExpression, calculator.ErrMismatchedParentheses, calculator.ErrDivisionByZero, calculator.ErrNegativeRoot, calculator.ErrOverflow, calculator.ErrUndefinedResult, calculator.ErrExpressionTooLong, calculator.ErrNestingTooDeep} {
		if statusFor(err) != http.StatusUnprocessableEntity {
			t.Fatalf("statusFor(%v)", err)
		}
	}
	if statusFor(requestError{message: "Content-Type must be application/json", kind: errUnsupportedMediaType}) != http.StatusUnsupportedMediaType {
		t.Fatal("content type mapping")
	}
	if statusFor(errors.New("unexpected")) != http.StatusInternalServerError {
		t.Fatal("unexpected errors must be internal")
	}
	internalRecorder := httptest.NewRecorder()
	writeMappedError(internalRecorder, errors.New("secret"))
	if internalRecorder.Code != http.StatusInternalServerError || internalRecorder.Body.String() != "{\"error\":\"internal server error\"}\n" {
		t.Fatal("internal error must not leak details")
	}
	recorder := httptest.NewRecorder()
	withRecovery(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic("test") })).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("recovery status = %d", recorder.Code)
	}
	resultRecorder := httptest.NewRecorder()
	writeResult(resultRecorder, math.Inf(1))
	if resultRecorder.Code != http.StatusUnprocessableEntity {
		t.Fatalf("result status = %d", resultRecorder.Code)
	}
}

func apiConfig() Config { return Config{AllowedOrigin: "https://example.test"} }
