package plan9asm

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Match Go's TYPE_CONST and TYPE_FCONST Prog formatting without converting
// integer source tokens into FLOAT or granting instruction/lowering permission.
// The FLOAT frontend consumes one leaf with only unary signs and parentheses.
func goAssemblerOracleX86OperandListing(original, listing []string) bool {
	if reflect.DeepEqual(original, listing) {
		return true
	}
	if len(original) < 2 || original[0] != "$" {
		return false
	}
	// TYPE_CONST is rendered as signed decimal, independently of an opcode
	// alias. Bind exactly one integer token and its optional unary sign;
	// FLOAT, expressions, symbols and another constant cannot borrow it.
	integer := original[1:]
	negative := len(integer) == 2 && integer[0] == "-"
	if len(integer) == 2 && (integer[0] == "+" || negative) {
		integer = integer[1:]
	}
	if len(integer) == 1 {
		if value, err := strconv.ParseUint(integer[0], 0, 64); err == nil {
			if negative {
				value = -value
			}
			printed := strconv.FormatInt(int64(value), 10)
			want := []string{"$"}
			if strings.HasPrefix(printed, "-") {
				want, printed = append(want, "-"), strings.TrimPrefix(printed, "-")
			}
			return reflect.DeepEqual(append(want, printed), listing)
		}
	}
	leaves := 0
	for _, token := range original[1:] {
		switch token {
		case "+", "-", "(", ")":
		default:
			if _, valid := parseGoFloatLiteral(token); !valid {
				return false
			}
			leaves++
		}
	}
	value, valid := parseGoFloatLiteral(strings.Join(original[1:], ""))
	if leaves != 1 || !valid {
		return false
	}
	// obj.Dconv uses 17 significant digits and forces a decimal point for
	// integral values, including negative zero. Require the exact canonical
	// token sequence, rather than accepting numerically equivalent output.
	printed := fmt.Sprintf("%.17g", value)
	if !strings.ContainsAny(printed, ".e") {
		printed += ".0"
	}
	want := []string{"$", "("}
	if strings.HasPrefix(printed, "-") {
		want = append(want, "-")
		printed = strings.TrimPrefix(printed, "-")
	}
	want = append(want, printed, ")")
	return reflect.DeepEqual(want, listing)
}
