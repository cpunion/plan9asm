package plan9asm

import (
	"fmt"
	"strings"
)

type wasmValue struct {
	typ       LLVMType
	val       string
	stackAddr bool
}

// translateFuncWASM lowers Go's WebAssembly Plan 9 stack instructions into
// ordinary LLVM SSA. FP operands are mapped through FuncSig.Frame, so the
// translated function keeps the signature selected by its caller instead of
// recreating the official compiler's linear-memory Go stack ABI.
func translateFuncWASM(b *strings.Builder, fn Func, sig FuncSig, annotateSource bool) error {
	fmt.Fprintf(b, "define %s %s(", sig.Ret, llvmGlobal(sig.Name))
	for i, typ := range sig.Args {
		if i != 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(b, "%s %%arg%d", typ, i)
	}
	b.WriteString(")")
	if sig.Attrs != "" {
		b.WriteString(" " + sig.Attrs)
	}
	b.WriteString(" {\nentry:\n")

	var (
		stack      []wasmValue
		results    = make([]wasmValue, len(sig.Frame.Results))
		haveResult = make([]bool, len(sig.Frame.Results))
		tmp        int
		terminated bool
	)
	newTmp := func() string {
		tmp++
		return fmt.Sprintf("wasm%d", tmp)
	}
	push := func(v wasmValue) { stack = append(stack, v) }
	pop := func(op string) (wasmValue, error) {
		if len(stack) == 0 {
			return wasmValue{}, fmt.Errorf("%s: operand stack underflow", op)
		}
		v := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		return v, nil
	}
	frameParam := func(arg Operand) (wasmValue, error) {
		for _, slot := range sig.Frame.Params {
			if slot.Offset != arg.FPOffset {
				continue
			}
			if slot.Index < 0 || slot.Index >= len(sig.Args) {
				return wasmValue{}, fmt.Errorf("FP parameter %s has invalid argument index %d", arg, slot.Index)
			}
			value := fmt.Sprintf("%%arg%d", slot.Index)
			if slot.Field >= 0 {
				name := newTmp()
				fmt.Fprintf(b, "  %%%s = extractvalue %s %s, %d\n", name, sig.Args[slot.Index], value, slot.Field)
				value = "%" + name
			}
			return wasmValue{typ: slot.Type, val: value}, nil
		}
		return wasmValue{}, fmt.Errorf("unknown FP parameter %s", arg)
	}
	storeResult := func(arg Operand, value wasmValue) error {
		for i, slot := range sig.Frame.Results {
			if slot.Offset != arg.FPOffset {
				continue
			}
			if value.typ != slot.Type {
				return fmt.Errorf("FP result %s has type %s, want %s", arg, value.typ, slot.Type)
			}
			results[i] = value
			haveResult[i] = true
			return nil
		}
		return fmt.Errorf("unknown FP result %s", arg)
	}
	emitReturn := func() error {
		if sig.Ret == Void {
			b.WriteString("  ret void\n")
			terminated = true
			return nil
		}
		if len(results) == 1 && haveResult[0] {
			if results[0].typ != sig.Ret {
				return fmt.Errorf("return type %s does not match stored result %s", sig.Ret, results[0].typ)
			}
			fmt.Fprintf(b, "  ret %s %s\n", sig.Ret, results[0].val)
			terminated = true
			return nil
		}
		if len(stack) != 0 {
			v := stack[len(stack)-1]
			if !v.stackAddr && v.typ == sig.Ret {
				fmt.Fprintf(b, "  ret %s %s\n", sig.Ret, v.val)
				terminated = true
				return nil
			}
		}
		return fmt.Errorf("missing %s return value", sig.Ret)
	}

	for _, ins := range fn.Instrs {
		op := normalizeInstructionOpcode(ins.Op)
		if op == "TEXT" {
			continue
		}
		if annotateSource && ins.Raw != "" {
			fmt.Fprintf(b, "  ; %s\n", strings.ReplaceAll(ins.Raw, "\n", " "))
		}
		switch op {
		case "NOP", "NO_LOCAL_POINTERS":
			continue
		case "GET":
			if len(ins.Args) != 1 || ins.Args[0].Kind != OpReg {
				return fmt.Errorf("Get expects one register")
			}
			if ins.Args[0].Reg != SP {
				return fmt.Errorf("Get %s is not implemented", ins.Args[0].Reg)
			}
			push(wasmValue{typ: I32, val: "0", stackAddr: true})
		case "F64LOAD":
			if len(ins.Args) != 1 || ins.Args[0].Kind != OpFP {
				return fmt.Errorf("F64Load expects an FP slot")
			}
			v, err := frameParam(ins.Args[0])
			if err != nil {
				return err
			}
			if v.typ != LLVMType("double") {
				return fmt.Errorf("F64Load %s has type %s", ins.Args[0], v.typ)
			}
			push(v)
		case "F64FLOOR", "F64CEIL", "F64TRUNC":
			v, err := pop(op)
			if err != nil {
				return err
			}
			if v.typ != LLVMType("double") || v.stackAddr {
				return fmt.Errorf("%s expects f64, got %s", op, v.typ)
			}
			intrinsic := strings.ToLower(strings.TrimPrefix(op, "F64"))
			name := newTmp()
			fmt.Fprintf(b, "  %%%s = call double @llvm.%s.f64(double %s)\n", name, intrinsic, v.val)
			push(wasmValue{typ: LLVMType("double"), val: "%" + name})
		case "F64STORE":
			if len(ins.Args) != 1 || ins.Args[0].Kind != OpFP {
				return fmt.Errorf("F64Store expects an FP slot")
			}
			if len(stack) < 2 {
				return fmt.Errorf("F64Store requires a preceding Get SP")
			}
			v, err := pop(op)
			if err != nil {
				return err
			}
			addr, err := pop(op)
			if err != nil {
				return err
			}
			if !addr.stackAddr {
				return fmt.Errorf("F64Store requires a preceding Get SP")
			}
			if err := storeResult(ins.Args[0], v); err != nil {
				return err
			}
		case "RET", "RETURN":
			if err := emitReturn(); err != nil {
				return err
			}
		case "UNDEF", "UNREACHABLE":
			b.WriteString("  unreachable\n")
			terminated = true
		default:
			return fmt.Errorf("unsupported wasm instruction %s", ins.Op)
		}
		if terminated {
			break
		}
	}
	if !terminated {
		if err := emitReturn(); err != nil {
			return err
		}
	}
	b.WriteString("}\n")
	return nil
}
