package plan9asm

import (
	"fmt"
	"testing"
)

func TestX86TypedCodeEntryOriginalTypeBoundary(t *testing.T) {
	// A one-field aggregate result is flattened to an LLVM scalar by the
	// ordinary ABI0 signature builder. That does not establish a scalar code
	// entry contract for the original Go declaration: it needs an adapter.
	cases := []struct {
		name, params, result, transport string
		parameterBytes                  int64 // -1 is one native pointer word.
		accepted                        bool
	}{
		{name: "struct-result", result: "struct{ X int64 }", transport: "wide"},
		{name: "array-result", result: "[1]int64", transport: "wide"},
		{name: "named-struct-result", result: "Aggregate", transport: "wide"},
		{name: "alias-struct-result", result: "AggregateAlias", transport: "wide"},
		{name: "named-array-result", result: "Array", transport: "wide"},
		{name: "alias-array-result", result: "ArrayAlias", transport: "wide"},
		{name: "multiple-results", result: "(int64, struct{})", transport: "wide"},
		{name: "struct-parameter", params: "value Aggregate", result: "int32", transport: "i32", parameterBytes: 8},
		{name: "array-parameter", params: "value ArrayAlias", result: "int32", transport: "i32", parameterBytes: 8},
		{name: "bool-result", result: "bool", transport: "bool"},
		{name: "bool-parameter", params: "flag bool", result: "int32", transport: "i32", parameterBytes: 1},
		{name: "function-handle", params: "fn func()", result: "int32", transport: "i32", parameterBytes: -1},
		{name: "map-handle", params: "value map[int]int", result: "int32", transport: "i32", parameterBytes: -1},
		{name: "channel-handle", params: "value chan int", result: "int32", transport: "i32", parameterBytes: -1},
		{name: "scalar-result", result: "int64", transport: "wide", accepted: true},
		{name: "named-scalar-result", result: "Integer", transport: "wide", accepted: true},
		{name: "pointer-parameter", params: "p *Aggregate", result: "int32", transport: "i32", parameterBytes: -1, accepted: true},
		{name: "pointer-result", params: "p *Aggregate", result: "*Aggregate", transport: "ptr", parameterBytes: -1, accepted: true},
		{name: "named-pointer-result", params: "p AggregatePointer", result: "AggregatePointer", transport: "ptr", parameterBytes: -1, accepted: true},
		{name: "alias-pointer-result", params: "p PointerAlias", result: "PointerAlias", transport: "ptr", parameterBytes: -1, accepted: true},
		{name: "void-result", params: "p *Aggregate", transport: "void", parameterBytes: -1, accepted: true},
	}
	for _, arch := range []string{"amd64", "386"} {
		word, move, triple := x86TypedCodeEntryMode(arch)
		for _, test := range cases {
			t.Run(arch+"/"+test.name, func(t *testing.T) {
				declarations := fmt.Sprintf(`package entry
type Aggregate struct { X int64 }
type AggregateAlias = Aggregate
type Array [1]int64
type ArrayAlias = Array
type AggregatePointer *Aggregate
type PointerAlias = *Aggregate
type Integer int64
func PC() uintptr
func Target(%s) %s
`, test.params, test.result)
				pkg := mustGoPackage(t, "entry", declarations)
				parameterBytes := test.parameterBytes
				if parameterBytes < 0 {
					parameterBytes = word
				}
				resultOffset := (parameterBytes + word - 1) / word * word
				var body string
				var resultBytes int64
				switch test.transport {
				case "wide":
					resultBytes = 8
					body = fmt.Sprintf("MOVQ $17, ret+%d(FP)\n", resultOffset)
					if arch == "386" {
						body = fmt.Sprintf("MOVL $17, ret+%d(FP)\nMOVL $0, ret+%d(FP)\n", resultOffset, resultOffset+4)
					}
				case "i32":
					resultBytes = 4
					body = fmt.Sprintf("MOVL $17, ret+%d(FP)\n", resultOffset)
				case "bool":
					resultBytes = 1
					body = fmt.Sprintf("MOVB $1, ret+%d(FP)\n", resultOffset)
				case "ptr":
					resultBytes = word
					body = fmt.Sprintf("%s p+0(FP), AX\n%s AX, ret+%d(FP)\n", move, move, resultOffset)
				case "void":
				default:
					t.Fatal("missing test transport")
				}
				source := fmt.Sprintf(`TEXT ·PC(SB), 4, $0-%d
%s $·Target(SB), AX
%s AX, ret+0(FP)
RET
TEXT ·Target(SB), 4, $0-%d
%sRET
`, word, move, move, resultOffset+resultBytes, body)
				options := GoModuleOptions{GOARCH: arch, TargetTriple: triple, ResolveSym: testResolveSym("entry")}
				module, err := TranslateGoModule(pkg, []byte(source), options)
				if module != nil {
					module.Module.Dispose()
				}
				if (err == nil) != test.accepted {
					t.Fatalf("original declaration accepted=%v, want %v: %v", err == nil, test.accepted, err)
				}
			})
		}
	}
}
