package calculator

import (
	"errors"
	"math"
	"strings"
	"testing"
)

func TestBasicOperations(t *testing.T) {
	if got := Add(8, 4); got != 12 {
		t.Errorf("Add() = %v, want 12", got)
	}
	if got := Subtract(8, 4); got != 4 {
		t.Errorf("Subtract() = %v, want 4", got)
	}
	if got := Multiply(8, 4); got != 32 {
		t.Errorf("Multiply() = %v, want 32", got)
	}
}

func TestDivideByZero(t *testing.T) {
	_, err := Divide(8, 0)
	if !errors.Is(err, ErrDivisionByZero) {
		t.Fatalf("Divide() error = %v, want ErrDivisionByZero", err)
	}
}

func TestOptionalOperations(t *testing.T) {
	if got := Power(2, 3); got != 8 {
		t.Errorf("Power() = %v, want 8", got)
	}
	if got, err := SquareRoot(81); err != nil || got != 9 {
		t.Errorf("SquareRoot() = %v, %v, want 9, nil", got, err)
	}
	if got := Percentage(25); got != 0.25 {
		t.Errorf("Percentage() = %v, want 0.25", got)
	}
}

func TestNegativeSquareRoot(t *testing.T) {
	_, err := SquareRoot(-1)
	if !errors.Is(err, ErrNegativeRoot) {
		t.Fatalf("SquareRoot() error = %v, want ErrNegativeRoot", err)
	}
}

func TestEvaluateExpression(t *testing.T) {
	tests := []struct {
		expression string
		want       float64
		err        error
	}{
		{"-2 + 3", 1, nil},
		{"2 * -3", -6, nil},
		{"3 - -2", 5, nil},
		{"--2", 2, nil},
		{"+5", 5, nil},
		{"-2 ^ 2", -4, nil},
		{"(-2) ^ 2", 4, nil},
		{"2 ^ -1", 0.5, nil},
		{"2 ^ 3 ^ 2", 512, nil},
		{"-(2 + 3)", -5, nil},
		{"4 / -2", -2, nil},
		{"1e3 + 1", 1001, nil},
		{".5 + .5", 1, nil},
		{"2 × 3", 6, nil},
		{"8 / 0", 0, ErrDivisionByZero},
		{"-", 0, ErrInvalidExpression},
		{"2 +", 0, ErrInvalidExpression},
		{"()", 0, ErrInvalidExpression},
		{"2 3", 0, ErrInvalidExpression},
		{"(2", 0, ErrInvalidExpression},
		{"2)", 0, ErrInvalidExpression},
		{"* 3", 0, ErrInvalidExpression},
	}
	for _, test := range tests {
		t.Run(test.expression, func(t *testing.T) {
			got, err := EvaluateExpression(test.expression)
			if test.err != nil {
				if !errors.Is(err, test.err) {
					t.Fatalf("EvaluateExpression() error = %v, want %v", err, test.err)
				}
				return
			}
			if err != nil || math.Abs(got-test.want) > 1e-9 {
				t.Fatalf("EvaluateExpression() = %v, %v, want %v", got, err, test.want)
			}
		})
	}
}

func TestExpressionLimits(t *testing.T) {
	if _, err := EvaluateExpression(strings.Repeat("1", MaxExpressionLength+1)); err == nil {
		t.Fatal("expected long expression error")
	}
	if _, err := EvaluateExpression(strings.Repeat("(", MaxNestingDepth+1) + "1" + strings.Repeat(")", MaxNestingDepth+1)); !errors.Is(err, ErrNestingTooDeep) {
		t.Fatalf("error = %v, want nesting limit", err)
	}
}
