package calculator

import (
	"errors"
	"math"
	"regexp"
	"strconv"
	"strings"
)

var (
	ErrDivisionByZero = errors.New("cannot divide by zero")
	ErrNegativeRoot   = errors.New("cannot calculate the square root of a negative number")
)

func Add(left, right float64) float64 {
	return left + right
}

func Subtract(left, right float64) float64 {
	return left - right
}

func Multiply(left, right float64) float64 {
	return left * right
}

func Divide(left, right float64) (float64, error) {
	if right == 0 {
		return 0, ErrDivisionByZero
	}
	return left / right, nil
}

func Power(base, exponent float64) float64 {
	return math.Pow(base, exponent)
}

func SquareRoot(value float64) (float64, error) {
	if value < 0 {
		return 0, ErrNegativeRoot
	}
	return math.Sqrt(value), nil
}

func Percentage(value float64) float64 {
	return value / 100
}

func EvaluateExpression(expression string) (float64, error) {
	normalized := strings.NewReplacer("×", "*", "÷", "/", "−", "-", "xʸ", "^").Replace(expression)
	tokenPattern := regexp.MustCompile(`\d*\.?\d+(?:[eE][+-]?\d+)?|[()+\-*/^]`)
	tokens := tokenPattern.FindAllString(normalized, -1)
	if len(tokens) == 0 || strings.Join(tokens, "") != strings.Join(strings.Fields(normalized), "") {
		return 0, errors.New("invalid expression")
	}

	parser := expressionParser{tokens: tokens}
	result, err := parser.parseAdditive()
	if err != nil {
		return 0, err
	}
	if parser.position != len(tokens) || !finite(result) {
		return 0, errors.New("invalid expression")
	}
	return result, nil
}

type expressionParser struct {
	tokens   []string
	position int
}

func (parser *expressionParser) peek() string {
	if parser.position >= len(parser.tokens) {
		return ""
	}
	return parser.tokens[parser.position]
}

func (parser *expressionParser) consume() string {
	token := parser.peek()
	parser.position++
	return token
}

func (parser *expressionParser) parsePrimary() (float64, error) {
	if parser.peek() == "(" {
		parser.consume()
		value, err := parser.parseAdditive()
		if err != nil {
			return 0, err
		}
		if parser.consume() != ")" {
			return 0, errors.New("missing closing parenthesis")
		}
		return value, nil
	}
	value, err := strconv.ParseFloat(parser.consume(), 64)
	if err != nil {
		return 0, errors.New("invalid expression")
	}
	return value, nil
}

func (parser *expressionParser) parsePower() (float64, error) {
	left, err := parser.parsePrimary()
	if err != nil {
		return 0, err
	}
	if parser.peek() != "^" {
		return left, nil
	}
	parser.consume()
	right, err := parser.parsePower()
	if err != nil {
		return 0, err
	}
	return Power(left, right), nil
}

func (parser *expressionParser) parseMultiplicative() (float64, error) {
	value, err := parser.parsePower()
	if err != nil {
		return 0, err
	}
	for parser.peek() == "*" || parser.peek() == "/" {
		operation := parser.consume()
		right, rightErr := parser.parsePower()
		if rightErr != nil {
			return 0, rightErr
		}
		if operation == "/" {
			value, err = Divide(value, right)
		} else {
			value = Multiply(value, right)
		}
		if err != nil {
			return 0, err
		}
	}
	return value, nil
}

func (parser *expressionParser) parseAdditive() (float64, error) {
	value, err := parser.parseMultiplicative()
	if err != nil {
		return 0, err
	}
	for parser.peek() == "+" || parser.peek() == "-" {
		operation := parser.consume()
		right, rightErr := parser.parseMultiplicative()
		if rightErr != nil {
			return 0, rightErr
		}
		if operation == "+" {
			value = Add(value, right)
		} else {
			value = Subtract(value, right)
		}
	}
	return value, nil
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
