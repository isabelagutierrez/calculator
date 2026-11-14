package calculator

import (
	"errors"
	"math"
	"strconv"
	"strings"
	"unicode"
)

const (
	MaxExpressionLength = 500
	MaxNestingDepth     = 100
)

var (
	ErrDivisionByZero    = errors.New("cannot divide by zero")
	ErrNegativeRoot      = errors.New("cannot calculate the square root of a negative number")
	ErrInvalidExpression = errors.New("invalid expression")
	ErrNestingTooDeep    = errors.New("parentheses are nested too deeply (maximum 100 levels)")
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

type token struct {
	value    string
	position int
}

func EvaluateExpression(expression string) (float64, error) {
	if len([]rune(expression)) > MaxExpressionLength {
		return 0, errors.New("expression exceeds the maximum length of 500 characters")
	}
	tokens, err := lex(expression)
	if err != nil {
		return 0, err
	}
	if len(tokens) == 0 {
		return 0, ErrInvalidExpression
	}
	parser := expressionParser{tokens: tokens}
	result, err := parser.parseExpression()
	if err != nil {
		return 0, err
	}
	if parser.position != len(tokens) || !finite(result) {
		return 0, ErrInvalidExpression
	}
	return result, nil
}

func lex(expression string) ([]token, error) {
	runes := []rune(strings.NewReplacer("×", "*", "÷", "/", "−", "-", "xʸ", "^").Replace(expression))
	tokens := make([]token, 0, len(runes)/2)
	for index := 0; index < len(runes); {
		if unicode.IsSpace(runes[index]) {
			index++
			continue
		}
		position := index + 1
		character := runes[index]
		if strings.ContainsRune("()+-*/^", character) {
			tokens = append(tokens, token{value: string(character), position: position})
			index++
			continue
		}
		if character == '.' || character >= '0' && character <= '9' {
			start := index
			digits := 0
			for index < len(runes) && runes[index] >= '0' && runes[index] <= '9' {
				index++
				digits++
			}
			if index < len(runes) && runes[index] == '.' {
				index++
				for index < len(runes) && runes[index] >= '0' && runes[index] <= '9' {
					index++
					digits++
				}
			}
			if digits == 0 {
				return nil, ErrInvalidExpression
			}
			if index < len(runes) && (runes[index] == 'e' || runes[index] == 'E') {
				index++
				if index < len(runes) && (runes[index] == '+' || runes[index] == '-') {
					index++
				}
				exponentStart := index
				for index < len(runes) && runes[index] >= '0' && runes[index] <= '9' {
					index++
				}
				if index == exponentStart {
					return nil, ErrInvalidExpression
				}
			}
			tokens = append(tokens, token{value: string(runes[start:index]), position: position})
			continue
		}
		return nil, ErrInvalidExpression
	}
	return tokens, nil
}

type expressionParser struct {
	tokens   []token
	position int
	depth    int
}

func (parser *expressionParser) peek() string {
	if parser.position >= len(parser.tokens) {
		return ""
	}
	return parser.tokens[parser.position].value
}

func (parser *expressionParser) consume() token {
	token := parser.tokens[parser.position]
	parser.position++
	return token
}

func (parser *expressionParser) parseExpression() (float64, error) {
	value, err := parser.parseTerm()
	if err != nil {
		return 0, err
	}
	for parser.peek() == "+" || parser.peek() == "-" {
		operation := parser.consume().value
		right, err := parser.parseTerm()
		if err != nil {
			return 0, err
		}
		if operation == "+" {
			value = Add(value, right)
		} else {
			value = Subtract(value, right)
		}
	}
	return value, nil
}

func (parser *expressionParser) parseTerm() (float64, error) {
	value, err := parser.parseUnary()
	if err != nil {
		return 0, err
	}
	for parser.peek() == "*" || parser.peek() == "/" {
		operation := parser.consume().value
		right, err := parser.parseUnary()
		if err != nil {
			return 0, err
		}
		if operation == "*" {
			value = Multiply(value, right)
		} else {
			value, err = Divide(value, right)
			if err != nil {
				return 0, err
			}
		}
	}
	return value, nil
}

func (parser *expressionParser) parseUnary() (float64, error) {
	sign := 1.0
	for parser.peek() == "+" || parser.peek() == "-" {
		if parser.consume().value == "-" {
			sign = -sign
		}
	}
	value, err := parser.parsePower()
	if err != nil {
		return 0, err
	}
	return sign * value, nil
}

func (parser *expressionParser) parsePower() (float64, error) {
	value, err := parser.parsePrimary()
	if err != nil {
		return 0, err
	}
	if parser.peek() != "^" {
		return value, nil
	}
	parser.consume()
	exponent, err := parser.parseUnary()
	if err != nil {
		return 0, err
	}
	return Power(value, exponent), nil
}

func (parser *expressionParser) parsePrimary() (float64, error) {
	if parser.peek() == "(" {
		parser.consume()
		parser.depth++
		if parser.depth > MaxNestingDepth {
			return 0, ErrNestingTooDeep
		}
		value, err := parser.parseExpression()
		if err != nil {
			return 0, err
		}
		if parser.peek() != ")" {
			return 0, ErrInvalidExpression
		}
		parser.consume()
		parser.depth--
		return value, nil
	}
	if parser.peek() == "" || parser.peek() == ")" {
		return 0, ErrInvalidExpression
	}
	value, err := strconv.ParseFloat(parser.consume().value, 64)
	if err != nil {
		return 0, ErrInvalidExpression
	}
	return value, nil
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
