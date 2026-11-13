package calculator

import (
	"errors"
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
	}{
		{"8 × 8 + 2", 66},
		{"2 ^ 3 × (4 + 1)", 40},
		{"81 / 9", 9},
	}
	for _, test := range tests {
		t.Run(test.expression, func(t *testing.T) {
			got, err := EvaluateExpression(test.expression)
			if err != nil || got != test.want {
				t.Fatalf("EvaluateExpression() = %v, %v, want %v, nil", got, err, test.want)
			}
		})
	}
}
