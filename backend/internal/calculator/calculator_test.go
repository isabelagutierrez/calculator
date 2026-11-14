package calculator

import (
	"errors"
	"math"
	"strings"
	"testing"
)

func TestOperations(t *testing.T) {
	if Add(8, 4) != 12 || Subtract(8, 4) != 4 || Multiply(8, 4) != 32 || Percentage(25) != .25 {
		t.Fatal("basic operation result mismatch")
	}
	tests := []struct {
		name string
		call func() (float64, error)
		want float64
		err  error
	}{{"divide", func() (float64, error) { return Divide(8, 4) }, 2, nil},
		{"power", func() (float64, error) { return Power(2, 3) }, 8, nil},
		{"root", func() (float64, error) { return SquareRoot(81) }, 9, nil},
		{"zero root", func() (float64, error) { return SquareRoot(0) }, 0, nil},
		{"divide zero", func() (float64, error) { return Divide(8, 0) }, 0, ErrDivisionByZero},
		{"negative root", func() (float64, error) { return SquareRoot(-1) }, 0, ErrNegativeRoot},
		{"negative power zero", func() (float64, error) { return Power(0, -1) }, 0, ErrDivisionByZero},
		{"undefined power", func() (float64, error) { return Power(-8, .5) }, 0, ErrUndefinedResult},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.call()
			if test.err != nil {
				if !errors.Is(err, test.err) {
					t.Fatalf("error = %v, want %v", err, test.err)
				}
				return
			}
			if err != nil || math.Abs(got-test.want) > 1e-9 {
				t.Fatalf("got %v, %v, want %v", got, err, test.want)
			}
		})
	}
}

func TestEvaluateExpression(t *testing.T) {
	tests := []struct {
		expression string
		want       float64
		err        error
	}{{"-5 + 3", -2, nil}, {"2 * -3", -6, nil}, {"-2 ^ 2", -4, nil}, {"2 ^ -1", .5, nil},
		{"2 ^ 3 ^ 2", 512, nil}, {"--3", 3, nil}, {"5 - -3", 8, nil}, {"2. + .5", 2.5, nil},
		{"2 × (3 + 4)", 14, nil}, {"0.1 + 0.2", .3, nil}, {"1e3 / 2", 500, nil},
		{"(2 + 3", 0, ErrMismatchedParentheses}, {"2 + 3)", 0, ErrMismatchedParentheses},
		{"2 +", 0, ErrInvalidExpression}, {"+", 0, ErrInvalidExpression}, {"()", 0, ErrMismatchedParentheses},
		{"abc", 0, ErrInvalidExpression}, {"2 $ 3", 0, ErrInvalidExpression}, {"8 / 0", 0, ErrDivisionByZero},
		{"10 ^ 400", 0, ErrOverflow}, {"1e400", 0, ErrOverflow}, {"(0 - 8) ^ 0.5", 0, ErrUndefinedResult}}
	for _, test := range tests {
		t.Run(test.expression, func(t *testing.T) {
			got, err := EvaluateExpression(test.expression)
			if test.err != nil {
				if !errors.Is(err, test.err) {
					t.Fatalf("error = %v, want %v", err, test.err)
				}
				return
			}
			if err != nil || math.Abs(got-test.want) > 1e-9 {
				t.Fatalf("got %v, %v, want %v", got, err, test.want)
			}
		})
	}
}

func TestExpressionLimits(t *testing.T) {
	if _, err := EvaluateExpression(strings.Repeat("1", 501)); !errors.Is(err, ErrExpressionTooLong) {
		t.Fatal(err)
	}
	if _, err := EvaluateExpression(strings.Repeat("(", 101) + "1" + strings.Repeat(")", 101)); !errors.Is(err, ErrNestingTooDeep) {
		t.Fatal(err)
	}
	if _, err := EvaluateExpression(strings.Repeat("(", 1_000_000)); err == nil {
		t.Fatal("expected limit error")
	}
}
