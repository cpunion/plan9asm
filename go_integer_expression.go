package plan9asm

import (
	"strconv"
	"strings"
	"text/scanner"
	"unicode/utf8"
)

const (
	goIntegerExpressionMaxBytes  = 64 << 10
	goIntegerExpressionMaxTokens = 4096
	goIntegerExpressionMaxDepth  = 128
)

// Integer immediates use cmd/asm's expr/term/factor grammar, not Go's AST
// grammar. In particular, SHR consumes a term, whereas SHL consumes a factor
// and finishes its term. The caller must consume the complete expression.
func parseImmExpr(text string) (uint64, bool) {
	if len(text) > goIntegerExpressionMaxBytes {
		return 0, false
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return 0, false
	}
	if value, err := strconv.ParseUint(text, 0, 64); err == nil {
		return value, true
	}
	parser := goIntegerExpression{reader: newPPGoReader("<integer-expression>", text), valid: true}
	parser.next()
	value := parser.expression()
	return value, parser.valid && parser.token.kind == scanner.EOF
}

type goIntegerExpression struct {
	reader *ppGoReader
	token  ppGoToken
	tokens int
	depth  int
	valid  bool
}

func (parser *goIntegerExpression) next() {
	if !parser.valid {
		return
	}
	parser.tokens++
	if parser.tokens > goIntegerExpressionMaxTokens {
		parser.valid = false
		return
	}
	token, err := parser.reader.next()
	parser.token = token
	parser.valid = err == nil
}

func (parser *goIntegerExpression) enter() bool {
	if !parser.valid || parser.depth >= goIntegerExpressionMaxDepth {
		parser.valid = false
		return false
	}
	parser.depth++
	return true
}

func (parser *goIntegerExpression) expression() uint64 {
	value := parser.term()
	for parser.valid {
		switch parser.token.text {
		case "+", "-", "|", "^":
			operator := parser.token.text
			parser.next()
			right := parser.term()
			switch operator {
			case "+":
				value += right
			case "-":
				value -= right
			case "|":
				value |= right
			case "^":
				value ^= right
			}
		default:
			return value
		}
	}
	return 0
}

func (parser *goIntegerExpression) term() uint64 {
	if !parser.enter() {
		return 0
	}
	defer func() { parser.depth-- }()
	value := parser.factor()
	for parser.valid {
		operator := parser.token.text
		switch operator {
		case "*", "/", "%", "&":
			parser.next()
			right := parser.factor()
			switch operator {
			case "*":
				value *= right
			case "/", "%":
				if int64(value) < 0 || right == 0 {
					parser.valid = false
					return 0
				}
				if operator == "/" {
					value /= right
				} else {
					value %= right
				}
			case "&":
				value &= right
			}
		case "<<", ">>":
			parser.next()
			var shift uint64
			if operator == "<<" {
				shift = parser.factor()
			} else {
				shift = parser.term()
			}
			if !parser.valid || int64(shift) < 0 || operator == ">>" && int64(value) < 0 {
				parser.valid = false
				return 0
			}
			if operator == "<<" {
				return value << shift
			}
			value >>= shift
		default:
			return value
		}
	}
	return 0
}

func (parser *goIntegerExpression) factor() uint64 {
	if !parser.enter() {
		return 0
	}
	defer func() { parser.depth-- }()
	token := parser.token
	parser.next()
	switch token.kind {
	case scanner.Int:
		value, err := strconv.ParseUint(token.text, 0, 64)
		parser.valid = parser.valid && err == nil
		return value
	case scanner.Char:
		text, err := strconv.Unquote(token.text)
		value, width := utf8.DecodeRuneInString(text)
		parser.valid = parser.valid && err == nil && text != "" && width == len(text) &&
			!(value == utf8.RuneError && width == 1)
		return uint64(value)
	case '+':
		return parser.factor()
	case '-':
		return -parser.factor()
	case '~':
		return ^parser.factor()
	case '(':
		value := parser.expression()
		if parser.token.kind != ')' {
			parser.valid = false
			return 0
		}
		parser.next()
		return value
	default:
		parser.valid = false
		return 0
	}
}
