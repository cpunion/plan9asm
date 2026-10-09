package plan9asm

import (
	"fmt"
	"strings"
)

// Go's complete F32/F64 comparison and min/max namespace consists of pure,
// zero-operand stack operators. Comparisons produce i32, never i1 or f64.
type wasmFloatBinaryOp struct {
	typ       LLVMType
	predicate string
	intrinsic string
}

var wasmFloatBinaryOps = map[string]wasmFloatBinaryOp{
	"F32EQ":  {typ: "float", predicate: "oeq"},
	"F32NE":  {typ: "float", predicate: "une"},
	"F32LT":  {typ: "float", predicate: "olt"},
	"F32GT":  {typ: "float", predicate: "ogt"},
	"F32LE":  {typ: "float", predicate: "ole"},
	"F32GE":  {typ: "float", predicate: "oge"},
	"F64EQ":  {typ: "double", predicate: "oeq"},
	"F64NE":  {typ: "double", predicate: "une"},
	"F64LT":  {typ: "double", predicate: "olt"},
	"F64GT":  {typ: "double", predicate: "ogt"},
	"F64LE":  {typ: "double", predicate: "ole"},
	"F64GE":  {typ: "double", predicate: "oge"},
	"F32MIN": {typ: "float", intrinsic: "minimum"},
	"F32MAX": {typ: "float", intrinsic: "maximum"},
	"F64MIN": {typ: "double", intrinsic: "minimum"},
	"F64MAX": {typ: "double", intrinsic: "maximum"},
}

func (op wasmFloatBinaryOp) resultType() LLVMType {
	if op.predicate != "" {
		return I32
	}
	return op.typ
}

func parseWASMFloatBinary(ins Instr) (wasmFloatBinaryOp, bool, error) {
	op := normalizeInstructionOpcode(ins.Op)
	spec, found := wasmFloatBinaryOps[op]
	if found && (strings.ToUpper(string(ins.Op)) != op || len(ins.Args) != 0) {
		return spec, true, fmt.Errorf("%s is a zero-operand WebAssembly stack instruction without suffixes", ins.Op)
	}
	return spec, found, nil
}
