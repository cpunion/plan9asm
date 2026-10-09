package plan9asm

import (
	"fmt"
	"reflect"
	"strings"
)

// Match Go's TYPE_FCONST Prog formatting without replacing integer tokens by
// floats or granting an instruction-format/lowering permission. The frontend
// consumes one FLOAT leaf, with only unary signs and parentheses around it.
func goAssemblerOracleX86OperandListing(original, listing []string) bool {
	if reflect.DeepEqual(original, listing) {
		return true
	}
	if len(original) < 2 || original[0] != "$" {
		return false
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
