package api

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"math"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/isabelagutierrez/calculator/backend/internal/calculator"
)

const maxBodyBytes = 8 << 10

type Config struct {
	AllowedOrigin string
}

type binaryRequest struct {
	Left  *float64 `json:"left"`
	Right *float64 `json:"right"`
}

type unaryRequest struct {
	Value *float64 `json:"value"`
}

type expressionRequest struct {
	Expression *string `json:"expression"`
}

type resultResponse struct {
	Result float64 `json:"result"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type requestError struct {
	message string
	kind    error
}

func (e requestError) Error() string { return e.message }
func (e requestError) Unwrap() error { return e.kind }

var (
	errBadRequest           = errors.New("bad request")
	errUnsupportedMediaType = errors.New("unsupported media type")
)

func NewHandler(configs ...Config) http.Handler {
	config := Config{AllowedOrigin: "*"}
	if len(configs) > 0 && configs[0].AllowedOrigin != "" {
		config = configs[0]
	}
	mux := http.NewServeMux()
	register := func(path, method string, handler http.HandlerFunc) {
		mux.HandleFunc(path, methodNotAllowed(method))
		mux.HandleFunc(method+" "+path, handler)
	}
	register("/health", http.MethodGet, health)
	register("/api/add", http.MethodPost, binary(calculator.Add))
	register("/api/subtract", http.MethodPost, binary(calculator.Subtract))
	register("/api/multiply", http.MethodPost, binary(calculator.Multiply))
	register("/api/divide", http.MethodPost, binaryWithError(calculator.Divide))
	register("/api/power", http.MethodPost, binaryWithError(calculator.Power))
	register("/api/square-root", http.MethodPost, unary(calculator.SquareRoot))
	register("/api/percentage", http.MethodPost, unary(func(value float64) (float64, error) {
		return calculator.Percentage(value), nil
	}))
	register("/api/calculate", http.MethodPost, calculateExpression)
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		writeError(w, http.StatusNotFound, errors.New("resource not found"))
	})
	return withCORS(config.AllowedOrigin)(withLogging(withRecovery(mux)))
}

func methodNotAllowed(method string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Allow", method)
		writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
	}
}

func health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func binary(operation func(float64, float64) float64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		request, err := decodeBinary(w, r)
		if err != nil {
			writeMappedError(w, err)
			return
		}
		writeResult(w, operation(*request.Left, *request.Right))
	}
}

func binaryWithError(operation func(float64, float64) (float64, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		request, err := decodeBinary(w, r)
		if err != nil {
			writeMappedError(w, err)
			return
		}
		result, err := operation(*request.Left, *request.Right)
		if err != nil {
			writeMappedError(w, err)
			return
		}
		writeResult(w, result)
	}
}

func unary(operation func(float64) (float64, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		request, err := decodeUnary(w, r)
		if err != nil {
			writeMappedError(w, err)
			return
		}
		result, err := operation(*request.Value)
		if err != nil {
			writeMappedError(w, err)
			return
		}
		writeResult(w, result)
	}
}

func calculateExpression(w http.ResponseWriter, r *http.Request) {
	var request expressionRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeMappedError(w, err)
		return
	}
	if request.Expression == nil || strings.TrimSpace(*request.Expression) == "" {
		writeMappedError(w, requestError{message: "expression is required", kind: errBadRequest})
		return
	}
	result, err := calculator.EvaluateExpression(*request.Expression)
	if err != nil {
		writeMappedError(w, err)
		return
	}
	writeResult(w, result)
}

func decodeBinary(w http.ResponseWriter, r *http.Request) (binaryRequest, error) {
	var request binaryRequest
	if err := decodeJSON(w, r, &request); err != nil {
		return request, err
	}
	if request.Left == nil || request.Right == nil {
		return request, requestError{message: "left and right are required", kind: errBadRequest}
	}
	if !finite(*request.Left) || !finite(*request.Right) {
		return request, requestError{message: "values must be finite numbers", kind: errBadRequest}
	}
	return request, nil
}

func decodeUnary(w http.ResponseWriter, r *http.Request) (unaryRequest, error) {
	var request unaryRequest
	if err := decodeJSON(w, r, &request); err != nil {
		return request, err
	}
	if request.Value == nil {
		return request, requestError{message: "value is required", kind: errBadRequest}
	}
	if !finite(*request.Value) {
		return request, requestError{message: "value must be a finite number", kind: errBadRequest}
	}
	return request, nil
}

func decodeJSON(w http.ResponseWriter, r *http.Request, destination any) error {
	if !contentTypeIsJSON(r) {
		return requestError{message: "Content-Type must be application/json", kind: errUnsupportedMediaType}
	}
	if r.Body == nil {
		return requestError{message: "request body is required", kind: errBadRequest}
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		var maxError *http.MaxBytesError
		if errors.As(err, &maxError) {
			return maxError
		}
		return requestError{message: "request body must be valid JSON", kind: errBadRequest}
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		var maxError *http.MaxBytesError
		if errors.As(err, &maxError) {
			return maxError
		}
		return requestError{message: "request body must contain one JSON object", kind: errBadRequest}
	}
	return nil
}

func contentTypeIsJSON(r *http.Request) bool {
	value, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	return err == nil && value == "application/json"
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func writeResult(w http.ResponseWriter, result float64) {
	if !finite(result) {
		writeMappedError(w, calculator.ErrOverflow)
		return
	}
	writeJSON(w, http.StatusOK, resultResponse{Result: result})
}

func statusFor(err error) int {
	var maxError *http.MaxBytesError
	if errors.As(err, &maxError) {
		return http.StatusRequestEntityTooLarge
	}
	if errors.Is(err, calculator.ErrInvalidExpression) ||
		errors.Is(err, calculator.ErrMismatchedParentheses) ||
		errors.Is(err, calculator.ErrDivisionByZero) ||
		errors.Is(err, calculator.ErrNegativeRoot) ||
		errors.Is(err, calculator.ErrOverflow) ||
		errors.Is(err, calculator.ErrUndefinedResult) ||
		errors.Is(err, calculator.ErrExpressionTooLong) ||
		errors.Is(err, calculator.ErrNestingTooDeep) {
		return http.StatusUnprocessableEntity
	}
	if errors.Is(err, errUnsupportedMediaType) {
		return http.StatusUnsupportedMediaType
	}
	if errors.Is(err, errBadRequest) {
		return http.StatusBadRequest
	}
	return http.StatusInternalServerError
}

func writeMappedError(w http.ResponseWriter, err error) {
	status := statusFor(err)
	message := err.Error()
	if status == http.StatusInternalServerError {
		message = "internal server error"
	}
	writeError(w, status, errors.New(message))
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, errorResponse{Error: err.Error()})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func withCORS(origin string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Max-Age", "600")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func withRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				writeError(w, http.StatusInternalServerError, errors.New("internal server error"))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (w *statusRecorder) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusRecorder) Write(body []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(body)
}

func withLogging(next http.Handler) http.Handler {
	logger := slog.Default()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &statusRecorder{ResponseWriter: w}
		next.ServeHTTP(recorder, r)
		status := recorder.status
		if status == 0 {
			status = http.StatusOK
		}
		logger.Info("http request", "method", r.Method, "path", r.URL.Path, "status", status, "duration", time.Since(start))
	})
}
