package api

import (
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"strings"

	"github.com/isabelagutierrez/calculator/backend/internal/calculator"
)

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

func NewHandler() http.Handler {
	mux := http.NewServeMux()
	register := func(path, method string, handler http.HandlerFunc) {
		mux.HandleFunc(path, func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Allow", method)
			writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		})
		mux.HandleFunc(method+" "+path, handler)
	}
	register("/health", http.MethodGet, health)
	register("/api/add", http.MethodPost, binary(add))
	register("/api/subtract", http.MethodPost, binary(subtract))
	register("/api/multiply", http.MethodPost, binary(multiply))
	register("/api/divide", http.MethodPost, binaryWithError(divide))
	register("/api/power", http.MethodPost, binary(power))
	register("/api/square-root", http.MethodPost, unary(squareRoot))
	register("/api/percentage", http.MethodPost, unary(percentage))
	register("/api/calculate", http.MethodPost, calculateExpression)
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		writeError(w, http.StatusNotFound, errors.New("resource not found"))
	})
	return withCORS(mux)
}

func health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func binary(operation func(float64, float64) float64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		request, err := decodeBinary(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeResult(w, operation(*request.Left, *request.Right))
	}
}

func binaryWithError(operation func(float64, float64) (float64, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		request, err := decodeBinary(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		result, err := operation(*request.Left, *request.Right)
		if err != nil {
			writeError(w, http.StatusUnprocessableEntity, err)
			return
		}
		writeResult(w, result)
	}
}

func unary(operation func(float64) (float64, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		request, err := decodeUnary(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		result, err := operation(*request.Value)
		if err != nil {
			writeError(w, http.StatusUnprocessableEntity, err)
			return
		}
		writeResult(w, result)
	}
}

func add(left, right float64) float64 {
	return calculator.Add(left, right)
}

func subtract(left, right float64) float64 {
	return calculator.Subtract(left, right)
}

func multiply(left, right float64) float64 {
	return calculator.Multiply(left, right)
}

func divide(left, right float64) (float64, error) {
	return calculator.Divide(left, right)
}

func power(left, right float64) float64 {
	return calculator.Power(left, right)
}

func squareRoot(value float64) (float64, error) {
	return calculator.SquareRoot(value)
}

func percentage(value float64) (float64, error) {
	return calculator.Percentage(value), nil
}

func calculateExpression(w http.ResponseWriter, r *http.Request) {
	var request expressionRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if request.Expression == nil || strings.TrimSpace(*request.Expression) == "" {
		writeError(w, http.StatusBadRequest, errors.New("expression is required"))
		return
	}
	result, err := calculator.EvaluateExpression(*request.Expression)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	writeResult(w, result)
}

func decodeBinary(r *http.Request) (binaryRequest, error) {
	var request binaryRequest
	if err := decodeJSON(r, &request); err != nil {
		return request, err
	}
	if request.Left == nil || request.Right == nil {
		return request, errors.New("left and right are required")
	}
	if !finite(*request.Left) || !finite(*request.Right) {
		return request, errors.New("values must be finite numbers")
	}
	return request, nil
}

func decodeUnary(r *http.Request) (unaryRequest, error) {
	var request unaryRequest
	if err := decodeJSON(r, &request); err != nil {
		return request, err
	}
	if request.Value == nil {
		return request, errors.New("value is required")
	}
	if !finite(*request.Value) {
		return request, errors.New("value must be a finite number")
	}
	return request, nil
}

func decodeJSON(r *http.Request, destination interface{}) error {
	if r.Body == nil {
		return errors.New("request body is required")
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return errors.New("request body must be valid JSON")
	}
	var extra interface{}
	if err := decoder.Decode(&extra); err != io.EOF {
		return errors.New("request body must contain one JSON object")
	}
	return nil
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func writeResult(w http.ResponseWriter, result float64) {
	if !finite(result) {
		writeError(w, http.StatusUnprocessableEntity, errors.New("result is not finite"))
		return
	}
	writeJSON(w, http.StatusOK, resultResponse{Result: result})
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, errorResponse{Error: err.Error()})
}

func writeJSON(w http.ResponseWriter, status int, value interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
