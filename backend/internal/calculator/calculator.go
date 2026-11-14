package calculator

import (
	"errors"
	"fmt"
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
	ErrInvalidExpression     = errors.New("invalid expression")
	ErrMismatchedParentheses = errors.New("mismatched parentheses")
	ErrDivisionByZero        = errors.New("cannot divide by zero")
	ErrNegativeRoot          = errors.New("cannot calculate the square root of a negative number")
	ErrOverflow              = errors.New("result is too large to represent")
	ErrUndefinedResult       = errors.New("result is undefined for these operands")
	ErrExpressionTooLong     = errors.New("expression exceeds the maximum length of 500 characters")
	ErrNestingTooDeep        = errors.New("parentheses are nested too deeply (maximum 100 levels)")
)

func Add(left, right float64) float64      { return left + right }
func Subtract(left, right float64) float64 { return left - right }
func Multiply(left, right float64) float64 { return left * right }

func Divide(left, right float64) (float64, error) {
	if right == 0 {
		return 0, ErrDivisionByZero
	}
	return checked(left / right)
}

func Power(base, exponent float64) (float64, error) {
	if base == 0 && exponent < 0 {
		return 0, ErrDivisionByZero
	}
	return checked(math.Pow(base, exponent))
}

func SquareRoot(value float64) (float64, error) {
	if value < 0 {
		return 0, ErrNegativeRoot
	}
	return checked(math.Sqrt(value))
}

func Percentage(value float64) float64 { return value / 100 }

func checked(value float64) (float64, error) {
	if math.IsNaN(value) {
		return 0, ErrUndefinedResult
	}
	if math.IsInf(value, 0) {
		return 0, ErrOverflow
	}
	if value == 0 {
		return 0, nil
	}
	return value, nil
}

type tokenKind uint8

const (
	tokenNumber tokenKind = iota
	tokenOperator
	tokenOpen
	tokenClose
)

type token struct {
	kind     tokenKind
	value    string
	position int
}

func EvaluateExpression(expression string) (float64, error) {
	if len([]rune(expression)) > MaxExpressionLength {
		return 0, ErrExpressionTooLong
	}
	tokens, err := lex(expression)
	if err != nil {
		return 0, err
	}
	if len(tokens) == 0 {
		return 0, fmt.Errorf("%w: unexpected end of expression", ErrInvalidExpression)
	}
	parser := expressionParser{tokens: tokens}
	result, err := parser.parseExpression()
	if err != nil {
		return 0, err
	}
	if parser.position < len(tokens) {
		t := parser.tokens[parser.position]
		if t.kind == tokenClose {
			return 0, fmt.Errorf("%w: unexpected closing parenthesis at position %d", ErrMismatchedParentheses, t.position)
		}
		return 0, fmt.Errorf("%w: unexpected token at position %d", ErrInvalidExpression, t.position)
	}
	return checked(result)
}

func lex(input string) ([]token, error) {
	runes := []rune(input)
	tokens := make([]token, 0, len(runes)/2)
	for index := 0; index < len(runes); {
		r := runes[index]
		position := index + 1
		if unicode.IsSpace(r) {
			index++
			continue
		}
		if r == '×' {
			tokens = append(tokens, token{kind: tokenOperator, value: "*", position: position})
			index++
			continue
		}
		if r == '÷' {
			tokens = append(tokens, token{kind: tokenOperator, value: "/", position: position})
			index++
			continue
		}
		if r == '−' {
			tokens = append(tokens, token{kind: tokenOperator, value: "-", position: position})
			index++
			continue
		}
		if r == 'x' && index+1 < len(runes) && runes[index+1] == 'ʸ' {
			tokens = append(tokens, token{kind: tokenOperator, value: "^", position: position})
			index += 2
			continue
		}
		if strings.ContainsRune("+-*/^", r) {
			tokens = append(tokens, token{kind: tokenOperator, value: string(r), position: position})
			index++
			continue
		}
		if r == '(' {
			tokens = append(tokens, token{kind: tokenOpen, value: "(", position: position})
			index++
			continue
		}
		if r == ')' {
			tokens = append(tokens, token{kind: tokenClose, value: ")", position: position})
			index++
			continue
		}
		if r == '.' || r >= '0' && r <= '9' {
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
				return nil, fmt.Errorf("%w: invalid number at position %d", ErrInvalidExpression, position)
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
				if exponentStart == index {
					return nil, fmt.Errorf("%w: invalid number at position %d", ErrInvalidExpression, position)
				}
			}
			tokens = append(tokens, token{kind: tokenNumber, value: string(runes[start:index]), position: position})
			continue
		}
		return nil, fmt.Errorf("%w: unexpected character %q at position %d", ErrInvalidExpression, r, position)
	}
	return tokens, nil
}

type expressionParser struct {
	tokens   []token
	position int
	depth    int
}

func (p *expressionParser) peek() token {
	if p.position >= len(p.tokens) {
		return token{position: len(p.tokens) + 1}
	}
	return p.tokens[p.position]
}

func (p *expressionParser) consume() token {
	t := p.peek()
	p.position++
	return t
}

func (p *expressionParser) parseExpression() (float64, error) {
	value, err := p.parseTerm()
	if err != nil {
		return 0, err
	}
	for {
		t := p.peek()
		if t.value != "+" && t.value != "-" {
			return value, nil
		}
		p.consume()
		right, err := p.parseTerm()
		if err != nil {
			return 0, err
		}
		if t.value == "+" {
			value = Add(value, right)
		} else {
			value = Subtract(value, right)
		}
		value, err = checked(value)
		if err != nil {
			return 0, err
		}
	}
}

func (p *expressionParser) parseTerm() (float64, error) {
	value, err := p.parseUnary()
	if err != nil {
		return 0, err
	}
	for {
		t := p.peek()
		if t.value != "*" && t.value != "/" {
			return value, nil
		}
		p.consume()
		right, err := p.parseUnary()
		if err != nil {
			return 0, err
		}
		if t.value == "*" {
			value = Multiply(value, right)
			value, err = checked(value)
		} else {
			value, err = Divide(value, right)
		}
		if err != nil {
			return 0, err
		}
	}
}

func (p *expressionParser) parseUnary() (float64, error) {
	sign := 1.0
	for {
		t := p.peek()
		if t.value != "+" && t.value != "-" {
			break
		}
		p.consume()
		if t.value == "-" {
			sign = -sign
		}
	}
	value, err := p.parsePower()
	if err != nil {
		return 0, err
	}
	return checked(sign * value)
}

func (p *expressionParser) parsePower() (float64, error) {
	value, err := p.parsePrimary()
	if err != nil {
		return 0, err
	}
	if p.peek().value != "^" {
		return value, nil
	}
	p.consume()
	exponent, err := p.parseUnary()
	if err != nil {
		return 0, err
	}
	return Power(value, exponent)
}

func (p *expressionParser) parsePrimary() (float64, error) {
	t := p.peek()
	if t.kind == tokenOpen {
		p.consume()
		p.depth++
		if p.depth > MaxNestingDepth {
			return 0, ErrNestingTooDeep
		}
		value, err := p.parseExpression()
		if err != nil {
			return 0, err
		}
		if p.peek().kind != tokenClose {
			return 0, fmt.Errorf("%w: missing closing parenthesis", ErrMismatchedParentheses)
		}
		p.consume()
		p.depth--
		return value, nil
	}
	if t.kind == tokenClose {
		return 0, fmt.Errorf("%w: unexpected closing parenthesis at position %d", ErrMismatchedParentheses, t.position)
	}
	if t.kind != tokenNumber {
		return 0, fmt.Errorf("%w: unexpected token at position %d", ErrInvalidExpression, t.position)
	}
	p.consume()
	value, err := strconv.ParseFloat(t.value, 64)
	if err != nil {
		if errors.Is(err, strconv.ErrRange) {
			return 0, ErrOverflow
		}
		return 0, fmt.Errorf("%w: invalid number at position %d", ErrInvalidExpression, t.position)
	}
	return checked(value)
}
