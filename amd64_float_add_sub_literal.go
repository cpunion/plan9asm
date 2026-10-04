package plan9asm

import (
	"math"
	"strconv"
	"strings"
	"text/scanner"
)

// obj6.progedit rewrites only these legacy scalar arithmetic constants to
// Float32Sym/Float64Sym. A literal is not caller memory or a register input.
// Both reviewed Go encoders use the same arithmetic rewrite, including ±0.
func x86FloatAddSubLiteral(ins Instr, spec x86FloatAddSubSpec) bool {
	if !spec.legacy || !spec.scalar || len(ins.Args) != 2 {
		return false
	}
	operand := ins.Args[0]
	if operand.Kind != OpImm || !operand.ImmIsFloat || operand.ImmRaw != "" {
		return false
	}
	value := math.Float64frombits(uint64(operand.Imm))
	if math.IsInf(value, 0) || math.IsNaN(value) {
		return false // these are not stock Go FLOAT tokens
	}
	if ins.Raw == "" {
		return true // explicitly typed finite value, without source syntax
	}
	_, rest := splitOpcode(ins.Raw)
	arguments := splitTopLevelCSV(rest)
	if len(arguments) != 2 {
		return false
	}
	text := strings.TrimSpace(arguments[0])
	if !strings.HasPrefix(text, "$") {
		return false
	}
	parsed, ok := x86GoArithmeticFloatExpr(strings.TrimSpace(text[1:]))
	return ok && math.Float64bits(parsed) == uint64(operand.Imm)
}

// cmd/asm floatExpr is a FLOAT leaf wrapped in unary +/- or parentheses,
// not the wider binary constant-expression grammar accepted by our parser.
func x86GoArithmeticFloatExpr(text string) (float64, bool) {
	var lexer scanner.Scanner
	lexer.Init(strings.NewReader(text))
	lexer.Mode = scanner.ScanInts | scanner.ScanFloats
	valid := true
	lexer.Error = func(*scanner.Scanner, string) { valid = false }
	negative, depth := false, 0
	for next := lexer.Scan(); ; next = lexer.Scan() {
		switch next {
		case '+':
		case '-':
			negative = !negative
		case '(':
			depth++
		case scanner.Float:
			value, err := strconv.ParseFloat(lexer.TokenText(), 64)
			if err != nil {
				return 0, false
			}
			for depth > 0 {
				if lexer.Scan() != ')' {
					return 0, false
				}
				depth--
			}
			if negative {
				value = -value
			}
			return value, valid && lexer.Scan() == scanner.EOF
		default:
			return 0, false
		}
	}
}
