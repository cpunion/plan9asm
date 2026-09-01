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

type wasmControlFrame struct {
	kind       string
	name       string
	startLabel string
	elseLabel  string
	endLabel   string
	baseStack  []wasmValue
	resultType LLVMType
	resultSlot string
	inElse     bool
}

// translateFuncWASM lowers Go's WebAssembly Plan 9 stack instructions into
// ordinary LLVM SSA. FP operands are mapped through FuncSig.Frame, so the
// translated function keeps the signature selected by its caller instead of
// recreating the official compiler's linear-memory Go stack ABI.
func translateFuncWASM(b *strings.Builder, fn Func, sig FuncSig, resolve func(string) string, sigs map[string]FuncSig, annotateSource bool) error {
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
		stack           []wasmValue
		regSlots        = make(map[Reg]string)
		regInitialized  = make(map[Reg]bool)
		results         = make([]wasmValue, len(sig.Frame.Results))
		haveResult      = make([]bool, len(sig.Frame.Results))
		tmp             int
		blockTerminated bool
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
	cast := func(v wasmValue, to LLVMType, signed bool) (wasmValue, error) {
		if v.stackAddr {
			return wasmValue{}, fmt.Errorf("cannot cast stack-address marker to %s", to)
		}
		if v.typ == to {
			return v, nil
		}
		name := newTmp()
		switch {
		case v.typ == Ptr && (to == I32 || to == I64):
			fmt.Fprintf(b, "  %%%s = ptrtoint ptr %s to %s\n", name, v.val, to)
		case (v.typ == I32 || v.typ == I64) && to == Ptr:
			fmt.Fprintf(b, "  %%%s = inttoptr %s %s to ptr\n", name, v.typ, v.val)
		case (v.typ == I1 || v.typ == I8 || v.typ == I16 || v.typ == I32) && to == I64:
			op := "zext"
			if signed {
				op = "sext"
			}
			fmt.Fprintf(b, "  %%%s = %s %s %s to i64\n", name, op, v.typ, v.val)
		case (v.typ == I1 || v.typ == I8 || v.typ == I16) && to == I32:
			op := "zext"
			if signed {
				op = "sext"
			}
			fmt.Fprintf(b, "  %%%s = %s %s %s to i32\n", name, op, v.typ, v.val)
		case v.typ == I64 && (to == I32 || to == I16 || to == I8 || to == I1):
			fmt.Fprintf(b, "  %%%s = trunc i64 %s to %s\n", name, v.val, to)
		case v.typ == I32 && (to == I16 || to == I8 || to == I1):
			fmt.Fprintf(b, "  %%%s = trunc i32 %s to %s\n", name, v.val, to)
		case v.typ == I16 && (to == I8 || to == I1):
			fmt.Fprintf(b, "  %%%s = trunc i16 %s to %s\n", name, v.val, to)
		case v.typ == I8 && to == I1:
			fmt.Fprintf(b, "  %%%s = trunc i8 %s to i1\n", name, v.val)
		default:
			return wasmValue{}, fmt.Errorf("unsupported wasm cast %s to %s", v.typ, to)
		}
		return wasmValue{typ: to, val: "%" + name}, nil
	}
	wasmRType := I64
	for i, reg := range sig.ArgRegs {
		if i >= len(sig.Args) || !strings.HasPrefix(strings.ToUpper(string(reg)), "R") {
			continue
		}
		if sig.Args[i] == I32 || sig.Args[i] == I64 {
			wasmRType = sig.Args[i]
			break
		}
	}
	registerType := func(reg Reg) (LLVMType, error) {
		name := strings.ToUpper(string(reg))
		switch {
		case strings.HasPrefix(name, "R"), name == "CTXT", name == "G",
			strings.HasPrefix(name, "RET"), name == "PAUSE":
			return wasmRType, nil
		case name == "PC_B":
			return I32, nil
		case strings.HasPrefix(name, "F"):
			n := 0
			if _, err := fmt.Sscanf(name, "F%d", &n); err != nil {
				return "", fmt.Errorf("invalid wasm register %s", reg)
			}
			if n < 16 {
				return LLVMType("float"), nil
			}
			return LLVMType("double"), nil
		case strings.HasPrefix(name, "V"):
			return LLVMType("<4 x i32>"), nil
		default:
			return "", fmt.Errorf("unsupported wasm register %s", reg)
		}
	}
	for _, ins := range fn.Instrs {
		for _, arg := range ins.Args {
			var reg Reg
			switch arg.Kind {
			case OpReg:
				reg = arg.Reg
			case OpMem:
				reg = arg.Mem.Base
			default:
				continue
			}
			if reg == "" || reg == SP {
				continue
			}
			if _, ok := regSlots[reg]; ok {
				continue
			}
			typ, err := registerType(reg)
			if err != nil {
				return err
			}
			slot := "wasm_reg_" + strings.ReplaceAll(string(reg), ".", "_")
			regSlots[reg] = slot
			fmt.Fprintf(b, "  %%%s = alloca %s\n", slot, typ)
			fmt.Fprintf(b, "  store %s %s, ptr %%%s\n", typ, llvmZeroValue(typ), slot)
		}
	}
	for i, reg := range sig.ArgRegs {
		if i >= len(sig.Args) || reg == SP {
			continue
		}
		slot, ok := regSlots[reg]
		if !ok {
			continue
		}
		typ, err := registerType(reg)
		if err != nil {
			return err
		}
		v, err := cast(wasmValue{typ: sig.Args[i], val: fmt.Sprintf("%%arg%d", i)}, typ, false)
		if err != nil {
			return fmt.Errorf("initialize %s from argument %d: %w", reg, i, err)
		}
		fmt.Fprintf(b, "  store %s %s, ptr %%%s\n", typ, v.val, slot)
		regInitialized[reg] = true
	}
	loadRegister := func(reg Reg) (wasmValue, error) {
		slot, ok := regSlots[reg]
		if !ok {
			return wasmValue{}, fmt.Errorf("register %s is not available", reg)
		}
		if reg == Reg("CTXT") && !regInitialized[reg] {
			return wasmValue{}, fmt.Errorf("CTXT requires an explicit closure-environment ABI input")
		}
		typ, err := registerType(reg)
		if err != nil {
			return wasmValue{}, err
		}
		name := newTmp()
		fmt.Fprintf(b, "  %%%s = load %s, ptr %%%s\n", name, typ, slot)
		return wasmValue{typ: typ, val: "%" + name}, nil
	}
	storeRegister := func(reg Reg, value wasmValue) (wasmValue, error) {
		slot, ok := regSlots[reg]
		if !ok {
			return wasmValue{}, fmt.Errorf("register %s is not available", reg)
		}
		typ, err := registerType(reg)
		if err != nil {
			return wasmValue{}, err
		}
		value, err = cast(value, typ, false)
		if err != nil {
			return wasmValue{}, err
		}
		fmt.Fprintf(b, "  store %s %s, ptr %%%s\n", typ, value.val, slot)
		regInitialized[reg] = true
		return value, nil
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
				fromType := value.typ
				var err error
				value, err = cast(value, slot.Type, false)
				if err != nil {
					return fmt.Errorf("FP result %s has type %s, want %s: %w", arg, fromType, slot.Type, err)
				}
			}
			results[i] = value
			haveResult[i] = true
			return nil
		}
		return fmt.Errorf("unknown FP result %s", arg)
	}
	var addressPtr func(wasmValue, int64) (wasmValue, error)
	loadOperand := func(arg Operand, width LLVMType) (wasmValue, error) {
		var v wasmValue
		var err error
		switch arg.Kind {
		case OpFP:
			v, err = frameParam(arg)
		case OpReg:
			v, err = loadRegister(arg.Reg)
		case OpImm:
			v = wasmValue{typ: width, val: fmt.Sprintf("%d", arg.Imm)}
		case OpMem:
			if arg.Mem.Base == SP {
				return wasmValue{}, fmt.Errorf("linear-memory SP operands are not available in the direct Go ABI")
			}
			base, loadErr := loadRegister(arg.Mem.Base)
			if loadErr != nil {
				return wasmValue{}, fmt.Errorf("load %s: %w", arg, loadErr)
			}
			address, addressErr := addressPtr(base, arg.Mem.Off)
			if addressErr != nil {
				return wasmValue{}, addressErr
			}
			name := newTmp()
			fmt.Fprintf(b, "  %%%s = load %s, ptr %s, align 1\n", name, width, address.val)
			v = wasmValue{typ: width, val: "%" + name}
		default:
			return wasmValue{}, fmt.Errorf("unsupported wasm source operand %s", arg)
		}
		if err != nil {
			return wasmValue{}, err
		}
		if v.typ != width && width != "" {
			v, err = cast(v, width, false)
		}
		return v, err
	}
	storeOperand := func(arg Operand, value wasmValue) error {
		switch arg.Kind {
		case OpReg:
			_, err := storeRegister(arg.Reg, value)
			return err
		case OpFP:
			return storeResult(arg, value)
		case OpMem:
			if arg.Mem.Base == SP {
				return fmt.Errorf("linear-memory SP operands are not available in the direct Go ABI")
			}
			base, err := loadRegister(arg.Mem.Base)
			if err != nil {
				return fmt.Errorf("store %s: %w", arg, err)
			}
			address, err := addressPtr(base, arg.Mem.Off)
			if err != nil {
				return err
			}
			fmt.Fprintf(b, "  store %s %s, ptr %s, align 1\n", value.typ, value.val, address.val)
			return nil
		default:
			return fmt.Errorf("unsupported wasm destination operand %s", arg)
		}
	}
	emitReturn := func() error {
		if sig.Ret == Void {
			b.WriteString("  ret void\n")
			blockTerminated = true
			return nil
		}
		if len(results) == 1 && haveResult[0] {
			if results[0].typ != sig.Ret {
				return fmt.Errorf("return type %s does not match stored result %s", sig.Ret, results[0].typ)
			}
			fmt.Fprintf(b, "  ret %s %s\n", sig.Ret, results[0].val)
			blockTerminated = true
			return nil
		}
		if len(stack) != 0 {
			v := stack[len(stack)-1]
			if !v.stackAddr && v.typ == sig.Ret {
				fmt.Fprintf(b, "  ret %s %s\n", sig.Ret, v.val)
				blockTerminated = true
				return nil
			}
		}
		return fmt.Errorf("missing %s return value", sig.Ret)
	}
	emitTailCall := func(arg Operand) error {
		if arg.Kind != OpSym || !strings.HasSuffix(arg.Sym, "(SB)") {
			return fmt.Errorf("JMP expects a direct (SB) symbol")
		}
		callee := resolve(strings.TrimSuffix(arg.Sym, "(SB)"))
		calleeSig, ok := sigs[callee]
		if !ok {
			return fmt.Errorf("missing signature for tail target %q", callee)
		}
		if len(calleeSig.Args) != len(sig.Args) {
			return fmt.Errorf("tail target %q needs %d arguments, caller has %d", callee, len(calleeSig.Args), len(sig.Args))
		}
		for i := range sig.Args {
			if sig.Args[i] != calleeSig.Args[i] {
				return fmt.Errorf("tail target %q argument %d has type %s, caller has %s", callee, i, calleeSig.Args[i], sig.Args[i])
			}
		}
		if sig.Ret != calleeSig.Ret {
			return fmt.Errorf("tail target %q returns %s, caller returns %s", callee, calleeSig.Ret, sig.Ret)
		}
		var args strings.Builder
		for i, typ := range calleeSig.Args {
			if i != 0 {
				args.WriteString(", ")
			}
			fmt.Fprintf(&args, "%s %%arg%d", typ, i)
		}
		if calleeSig.Ret == Void {
			fmt.Fprintf(b, "  tail call void %s(%s)\n  ret void\n", llvmGlobal(funcSigSymbol(callee, calleeSig)), args.String())
		} else {
			name := newTmp()
			fmt.Fprintf(b, "  %%%s = tail call %s %s(%s)\n", name, calleeSig.Ret, llvmGlobal(funcSigSymbol(callee, calleeSig)), args.String())
			fmt.Fprintf(b, "  ret %s %%%s\n", calleeSig.Ret, name)
		}
		blockTerminated = true
		return nil
	}
	emitLocalCall := func(arg Operand) error {
		if arg.Kind != OpSym || !strings.HasSuffix(arg.Sym, "<>(SB)") {
			return fmt.Errorf("Call currently supports file-local wasm helpers only")
		}
		callee := resolve(strings.TrimSuffix(arg.Sym, "(SB)"))
		calleeSig, ok := sigs[callee]
		if !ok {
			return fmt.Errorf("missing signature for local call target %q", callee)
		}
		if len(stack) < len(calleeSig.Args) {
			return fmt.Errorf("local call %q needs %d stack arguments, have %d", callee, len(calleeSig.Args), len(stack))
		}
		args := make([]wasmValue, len(calleeSig.Args))
		for i := len(args) - 1; i >= 0; i-- {
			v, _ := pop("Call") // len(stack) was checked above.
			var err error
			v, err = cast(v, calleeSig.Args[i], false)
			if err != nil {
				return fmt.Errorf("local call %q argument %d: %w", callee, i, err)
			}
			args[i] = v
		}
		var argText strings.Builder
		for i, v := range args {
			if i != 0 {
				argText.WriteString(", ")
			}
			fmt.Fprintf(&argText, "%s %s", v.typ, v.val)
		}
		if calleeSig.Ret == Void {
			fmt.Fprintf(b, "  call void %s(%s)\n", llvmGlobal(funcSigSymbol(callee, calleeSig)), argText.String())
			return nil
		}
		name := newTmp()
		fmt.Fprintf(b, "  %%%s = call %s %s(%s)\n", name, calleeSig.Ret, llvmGlobal(funcSigSymbol(callee, calleeSig)), argText.String())
		push(wasmValue{typ: calleeSig.Ret, val: "%" + name})
		return nil
	}
	emitIntBinary := func(op string, typ LLVMType, llvmOp string) error {
		rhs, err := pop(op)
		if err != nil {
			return err
		}
		lhs, err := pop(op)
		if err != nil {
			return err
		}
		lhs, err = cast(lhs, typ, false)
		if err != nil {
			return err
		}
		rhs, err = cast(rhs, typ, false)
		if err != nil {
			return err
		}
		name := newTmp()
		fmt.Fprintf(b, "  %%%s = %s %s %s, %s\n", name, llvmOp, typ, lhs.val, rhs.val)
		push(wasmValue{typ: typ, val: "%" + name})
		return nil
	}
	emitIntCompare := func(op string, typ LLVMType, pred string, unary bool) error {
		rhs := wasmValue{typ: typ, val: "0"}
		var err error
		if !unary {
			rhs, err = pop(op)
			if err != nil {
				return err
			}
		}
		lhs, err := pop(op)
		if err != nil {
			return err
		}
		lhs, err = cast(lhs, typ, false)
		if err != nil {
			return err
		}
		rhs, err = cast(rhs, typ, false)
		if err != nil {
			return err
		}
		cmp := newTmp()
		fmt.Fprintf(b, "  %%%s = icmp %s %s %s, %s\n", cmp, pred, typ, lhs.val, rhs.val)
		out := newTmp()
		fmt.Fprintf(b, "  %%%s = zext i1 %%%s to i32\n", out, cmp)
		push(wasmValue{typ: I32, val: "%" + out})
		return nil
	}
	addressPtr = func(v wasmValue, offset int64) (wasmValue, error) {
		var err error
		v, err = cast(v, Ptr, false)
		if err != nil {
			return wasmValue{}, err
		}
		if offset == 0 {
			return v, nil
		}
		name := newTmp()
		fmt.Fprintf(b, "  %%%s = getelementptr i8, ptr %s, i32 %d\n", name, v.val, offset)
		return wasmValue{typ: Ptr, val: "%" + name}, nil
	}
	emitMemoryLoad := func(op string, offset int64, spec wasmMemoryLoadOp) error {
		address, err := pop(op)
		if err != nil {
			return err
		}
		address, err = addressPtr(address, offset)
		if err != nil {
			return err
		}
		name := newTmp()
		fmt.Fprintf(b, "  %%%s = load %s, ptr %s, align 1\n", name, spec.loadType, address.val)
		v := wasmValue{typ: spec.loadType, val: "%" + name}
		if spec.loadType != spec.resultType {
			v, err = cast(v, spec.resultType, spec.signed)
			if err != nil {
				return err
			}
		}
		push(v)
		return nil
	}
	emitMemoryStore := func(op string, offset int64, spec wasmMemoryStoreOp) error {
		value, err := pop(op)
		if err != nil {
			return err
		}
		address, err := pop(op)
		if err != nil {
			return err
		}
		address, err = addressPtr(address, offset)
		if err != nil {
			return err
		}
		value, err = cast(value, spec.storeType, false)
		if err != nil {
			return err
		}
		fmt.Fprintf(b, "  store %s %s, ptr %s, align 1\n", spec.storeType, value.val, address.val)
		return nil
	}
	root := &wasmControlFrame{kind: "function", endLabel: "wasm_function_return"}
	if sig.Ret != Void {
		root.resultType = sig.Ret
		root.resultSlot = "wasm_function_result"
		fmt.Fprintf(b, "  %%%s = alloca %s\n", root.resultSlot, root.resultType)
	}
	var (
		controls       = []*wasmControlFrame{root}
		pendingName    string
		rootBranchUsed bool
	)
	copyStack := func(values []wasmValue) []wasmValue {
		return append([]wasmValue(nil), values...)
	}
	newLabel := func(prefix string) string {
		return prefix + "_" + newTmp()
	}
	emitLabel := func(label string) {
		fmt.Fprintf(b, "%s:\n", label)
		blockTerminated = false
	}
	conditionValue := func(op string) (string, error) {
		condition, err := pop(op)
		if err != nil {
			return "", err
		}
		condition, err = cast(condition, I32, false)
		if err != nil {
			return "", err
		}
		name := newTmp()
		fmt.Fprintf(b, "  %%%s = icmp ne i32 %s, 0\n", name, condition.val)
		return "%" + name, nil
	}
	storeBlockResult := func(frame *wasmControlFrame, popValue bool) error {
		if frame.resultType == "" {
			return nil
		}
		if len(stack) <= len(frame.baseStack) {
			return fmt.Errorf("%s block is missing its %s result", frame.kind, frame.resultType)
		}
		v := stack[len(stack)-1]
		if popValue {
			stack = stack[:len(stack)-1]
		}
		var err error
		v, err = cast(v, frame.resultType, false)
		if err != nil {
			return err
		}
		fmt.Fprintf(b, "  store %s %s, ptr %%%s\n", frame.resultType, v.val, frame.resultSlot)
		return nil
	}
	finishControlPath := func(frame *wasmControlFrame) error {
		if blockTerminated {
			return nil
		}
		if err := storeBlockResult(frame, true); err != nil {
			return err
		}
		fmt.Fprintf(b, "  br label %%%s\n", frame.endLabel)
		blockTerminated = true
		return nil
	}
	controlTarget := func(arg Operand) (*wasmControlFrame, error) {
		if arg.Kind == OpImm {
			depth := int(arg.Imm)
			if depth < 0 || depth >= len(controls) {
				return nil, fmt.Errorf("branch depth %d exceeds %d active controls", depth, len(controls))
			}
			return controls[len(controls)-1-depth], nil
		}
		var name string
		switch arg.Kind {
		case OpIdent:
			name = arg.Ident
		case OpLabel, OpSym:
			name = strings.TrimSuffix(arg.Sym, ":")
		default:
			return nil, fmt.Errorf("unsupported branch target %s", arg)
		}
		for i := len(controls) - 1; i >= 0; i-- {
			if controls[i].name == name {
				return controls[i], nil
			}
		}
		return nil, fmt.Errorf("unknown structured branch target %q", name)
	}
	emitBranch := func(op string, arg Operand, conditional bool) error {
		target, err := controlTarget(arg)
		if err != nil {
			return err
		}
		var condition string
		if conditional {
			condition, err = conditionValue(op)
			if err != nil {
				return err
			}
		}
		if target.resultType != "" {
			if err := storeBlockResult(target, !conditional); err != nil {
				return err
			}
		}
		if target.kind == "function" {
			rootBranchUsed = true
		}
		label := target.endLabel
		if target.kind == "loop" {
			label = target.startLabel
		}
		if conditional {
			continuation := newLabel("wasm_br_cont")
			fmt.Fprintf(b, "  br i1 %s, label %%%s, label %%%s\n", condition, label, continuation)
			emitLabel(continuation)
			return nil
		}
		fmt.Fprintf(b, "  br label %%%s\n", label)
		blockTerminated = true
		return nil
	}

	for _, ins := range fn.Instrs {
		op := normalizeInstructionOpcode(ins.Op)
		if op == "TEXT" {
			continue
		}
		if annotateSource && ins.Raw != "" {
			fmt.Fprintf(b, "  ; %s\n", strings.ReplaceAll(ins.Raw, "\n", " "))
		}
		if blockTerminated && op != "ELSE" && op != "END" && op != "LABEL" {
			continue
		}
		if spec, ok := wasmIntegerBinaryOps[op]; ok {
			if err := emitIntBinary(op, spec.typ, spec.op); err != nil {
				return err
			}
			continue
		}
		if spec, ok := wasmIntegerCompareOps[op]; ok {
			if err := emitIntCompare(op, spec.typ, spec.op, false); err != nil {
				return err
			}
			continue
		}
		if spec, ok := wasmMemoryLoadOps[op]; ok {
			if len(ins.Args) != 1 {
				return fmt.Errorf("%s expects one operand", ins.Op)
			}
			if ins.Args[0].Kind == OpFP && (op == "I32LOAD" || op == "I64LOAD") {
				v, err := loadOperand(ins.Args[0], spec.resultType)
				if err != nil {
					return err
				}
				push(v)
				continue
			}
			if ins.Args[0].Kind == OpMem {
				v, err := loadOperand(ins.Args[0], spec.loadType)
				if err != nil {
					return err
				}
				if spec.loadType != spec.resultType {
					v, err = cast(v, spec.resultType, spec.signed)
					if err != nil {
						return err
					}
				}
				push(v)
				continue
			}
			if ins.Args[0].Kind != OpImm {
				return fmt.Errorf("%s expects an FP slot or immediate memory offset, got %s", ins.Op, ins.Args[0])
			}
			if err := emitMemoryLoad(op, ins.Args[0].Imm, spec); err != nil {
				return err
			}
			continue
		}
		if spec, ok := wasmMemoryStoreOps[op]; ok {
			if len(ins.Args) != 1 {
				return fmt.Errorf("%s expects one operand", ins.Op)
			}
			if ins.Args[0].Kind != OpFP {
				if ins.Args[0].Kind != OpImm {
					return fmt.Errorf("%s expects an FP slot or immediate memory offset, got %s", ins.Op, ins.Args[0])
				}
				if err := emitMemoryStore(op, ins.Args[0].Imm, spec); err != nil {
					return err
				}
				continue
			}
			if len(stack) < 2 {
				return fmt.Errorf("%s requires a preceding Get SP", ins.Op)
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
				return fmt.Errorf("%s requires a preceding Get SP", ins.Op)
			}
			v, err = cast(v, spec.storeType, false)
			if err != nil {
				return err
			}
			if err := storeResult(ins.Args[0], v); err != nil {
				return err
			}
			continue
		}
		switch op {
		case "NOP", "NO_LOCAL_POINTERS":
			continue
		case "LABEL":
			if len(ins.Args) != 1 {
				return fmt.Errorf("wasm label expects one name")
			}
			pendingName = strings.TrimSuffix(ins.Args[0].Sym, ":")
		case "BLOCK", "LOOP", "IF":
			resultType := LLVMType("")
			if len(ins.Args) == 1 {
				if ins.Args[0].Kind != OpImm {
					return fmt.Errorf("%s block type must be an immediate", ins.Op)
				}
				resultType = wasmBlockResultType(ins.Args[0].Imm)
				if resultType == "" {
					return fmt.Errorf("%s has unsupported block type $%d", ins.Op, ins.Args[0].Imm)
				}
			} else if len(ins.Args) != 0 {
				return fmt.Errorf("%s expects at most one block type", ins.Op)
			}
			frame := &wasmControlFrame{
				kind:       strings.ToLower(op),
				name:       pendingName,
				endLabel:   newLabel("wasm_control_end"),
				baseStack:  copyStack(stack),
				resultType: resultType,
			}
			pendingName = ""
			if resultType != "" {
				frame.resultSlot = newLabel("wasm_control_result")
				fmt.Fprintf(b, "  %%%s = alloca %s\n", frame.resultSlot, resultType)
			}
			switch op {
			case "LOOP":
				frame.startLabel = newLabel("wasm_loop")
				fmt.Fprintf(b, "  br label %%%s\n", frame.startLabel)
				emitLabel(frame.startLabel)
			case "IF":
				condition, err := conditionValue(op)
				if err != nil {
					return err
				}
				frame.baseStack = copyStack(stack)
				thenLabel := newLabel("wasm_if_then")
				frame.elseLabel = newLabel("wasm_if_else")
				fmt.Fprintf(b, "  br i1 %s, label %%%s, label %%%s\n", condition, thenLabel, frame.elseLabel)
				emitLabel(thenLabel)
			}
			controls = append(controls, frame)
		case "ELSE":
			if len(controls) == 0 || controls[len(controls)-1].kind != "if" {
				return fmt.Errorf("Else without matching If")
			}
			frame := controls[len(controls)-1]
			if frame.inElse {
				return fmt.Errorf("duplicate Else")
			}
			if err := finishControlPath(frame); err != nil {
				return err
			}
			stack = copyStack(frame.baseStack)
			emitLabel(frame.elseLabel)
			frame.inElse = true
		case "END":
			if len(controls) <= 1 {
				return fmt.Errorf("End without active control")
			}
			frame := controls[len(controls)-1]
			controls = controls[:len(controls)-1]
			if frame.kind == "if" && !frame.inElse {
				if err := finishControlPath(frame); err != nil {
					return err
				}
				stack = copyStack(frame.baseStack)
				emitLabel(frame.elseLabel)
				if frame.resultType != "" {
					return fmt.Errorf("result-producing If requires Else")
				}
			}
			if err := finishControlPath(frame); err != nil {
				return err
			}
			stack = copyStack(frame.baseStack)
			emitLabel(frame.endLabel)
			if frame.resultType != "" {
				name := newTmp()
				fmt.Fprintf(b, "  %%%s = load %s, ptr %%%s\n", name, frame.resultType, frame.resultSlot)
				push(wasmValue{typ: frame.resultType, val: "%" + name})
			}
		case "BR", "BRIF":
			if len(ins.Args) != 1 {
				return fmt.Errorf("%s expects one target", ins.Op)
			}
			if err := emitBranch(op, ins.Args[0], op == "BRIF"); err != nil {
				return err
			}
		case "GET":
			if len(ins.Args) != 1 || ins.Args[0].Kind != OpReg {
				return fmt.Errorf("Get expects one register")
			}
			if ins.Args[0].Reg == SP {
				push(wasmValue{typ: I32, val: "0", stackAddr: true})
				continue
			}
			v, err := loadRegister(ins.Args[0].Reg)
			if err != nil {
				return err
			}
			push(v)
		case "SET", "TEE":
			if len(ins.Args) != 1 || ins.Args[0].Kind != OpReg {
				return fmt.Errorf("%s expects one register", ins.Op)
			}
			v, err := pop(op)
			if err != nil {
				return err
			}
			v, err = storeRegister(ins.Args[0].Reg, v)
			if err != nil {
				return err
			}
			if op == "TEE" {
				push(v)
			}
		case "MOVB", "MOVW", "MOVD":
			if len(ins.Args) != 2 {
				return fmt.Errorf("%s expects source and destination", ins.Op)
			}
			width := map[string]LLVMType{"MOVB": I8, "MOVW": I32, "MOVD": I64}[op]
			v, err := loadOperand(ins.Args[0], width)
			if err != nil {
				return err
			}
			if err := storeOperand(ins.Args[1], v); err != nil {
				return err
			}
		case "I32CONST", "I64CONST":
			if len(ins.Args) != 1 || ins.Args[0].Kind != OpImm {
				return fmt.Errorf("%s expects one integer immediate", ins.Op)
			}
			typ := I32
			if op == "I64CONST" {
				typ = I64
			}
			push(wasmValue{typ: typ, val: fmt.Sprintf("%d", ins.Args[0].Imm)})
		case "F64LOAD":
			if len(ins.Args) != 1 || ins.Args[0].Kind != OpFP {
				return fmt.Errorf("F64Load expects an FP slot")
			}
			v, err := loadOperand(ins.Args[0], LLVMType("double"))
			if err != nil {
				return err
			}
			push(v)
		case "I32WRAPI64":
			v, err := pop(op)
			if err != nil {
				return err
			}
			v, err = cast(v, I32, false)
			if err != nil {
				return err
			}
			push(v)
		case "I64EXTENDI32U", "I64EXTENDI32S":
			v, err := pop(op)
			if err != nil {
				return err
			}
			v, err = cast(v, I64, op == "I64EXTENDI32S")
			if err != nil {
				return err
			}
			push(v)
		case "I32EQZ", "NOT":
			if err := emitIntCompare(op, I32, "eq", true); err != nil {
				return err
			}
		case "I64EQZ":
			if err := emitIntCompare(op, I64, "eq", true); err != nil {
				return err
			}
		case "SELECT":
			condition, err := pop(op)
			if err != nil {
				return err
			}
			ifFalse, err := pop(op)
			if err != nil {
				return err
			}
			ifTrue, err := pop(op)
			if err != nil {
				return err
			}
			if ifTrue.typ != ifFalse.typ {
				return fmt.Errorf("Select value type mismatch: %s and %s", ifTrue.typ, ifFalse.typ)
			}
			condition, err = cast(condition, I32, false)
			if err != nil {
				return err
			}
			cmp := newTmp()
			fmt.Fprintf(b, "  %%%s = icmp ne i32 %s, 0\n", cmp, condition.val)
			name := newTmp()
			fmt.Fprintf(b, "  %%%s = select i1 %%%s, %s %s, %s %s\n", name, cmp, ifTrue.typ, ifTrue.val, ifFalse.typ, ifFalse.val)
			push(wasmValue{typ: ifTrue.typ, val: "%" + name})
		case "DROP":
			if _, err := pop(op); err != nil {
				return err
			}
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
		case "CURRENTMEMORY":
			name := newTmp()
			fmt.Fprintf(b, "  %%%s = call i32 @llvm.wasm.memory.size.i32(i32 0)\n", name)
			push(wasmValue{typ: I32, val: "%" + name})
		case "GROWMEMORY":
			pages, err := pop(op)
			if err != nil {
				return err
			}
			pages, err = cast(pages, I32, false)
			if err != nil {
				return err
			}
			name := newTmp()
			fmt.Fprintf(b, "  %%%s = call i32 @llvm.wasm.memory.grow.i32(i32 0, i32 %s)\n", name, pages.val)
			push(wasmValue{typ: I32, val: "%" + name})
		case "MEMORYCOPY":
			length, err := pop(op)
			if err != nil {
				return err
			}
			src, err := pop(op)
			if err != nil {
				return err
			}
			dst, err := pop(op)
			if err != nil {
				return err
			}
			length, err = cast(length, I32, false)
			if err != nil {
				return err
			}
			src, err = cast(src, Ptr, false)
			if err != nil {
				return err
			}
			dst, err = cast(dst, Ptr, false)
			if err != nil {
				return err
			}
			fmt.Fprintf(b, "  call void @llvm.memmove.p0.p0.i32(ptr %s, ptr %s, i32 %s, i1 false)\n", dst.val, src.val, length.val)
		case "MEMORYFILL":
			length, err := pop(op)
			if err != nil {
				return err
			}
			fill, err := pop(op)
			if err != nil {
				return err
			}
			dst, err := pop(op)
			if err != nil {
				return err
			}
			length, err = cast(length, I32, false)
			if err != nil {
				return err
			}
			fill, err = cast(fill, I8, false)
			if err != nil {
				return err
			}
			dst, err = cast(dst, Ptr, false)
			if err != nil {
				return err
			}
			fmt.Fprintf(b, "  call void @llvm.memset.p0.i32(ptr %s, i8 %s, i32 %s, i1 false)\n", dst.val, fill.val, length.val)
		case "RET", "RETURN":
			if err := emitReturn(); err != nil {
				return err
			}
		case "JMP":
			if len(ins.Args) != 1 {
				return fmt.Errorf("JMP expects one target")
			}
			if err := emitTailCall(ins.Args[0]); err != nil {
				return err
			}
		case "CALL":
			if len(ins.Args) != 1 {
				return fmt.Errorf("Call expects one target")
			}
			if err := emitLocalCall(ins.Args[0]); err != nil {
				return err
			}
		case "UNDEF", "UNREACHABLE":
			b.WriteString("  unreachable\n")
			blockTerminated = true
		default:
			return fmt.Errorf("unsupported wasm instruction %s", ins.Op)
		}
	}
	if len(controls) != 1 {
		return fmt.Errorf("%d unterminated wasm control blocks", len(controls)-1)
	}
	if !blockTerminated {
		if err := emitReturn(); err != nil {
			return err
		}
	}
	if rootBranchUsed {
		emitLabel(root.endLabel)
		if root.resultType == "" {
			b.WriteString("  ret void\n")
		} else {
			name := newTmp()
			fmt.Fprintf(b, "  %%%s = load %s, ptr %%%s\n", name, root.resultType, root.resultSlot)
			fmt.Fprintf(b, "  ret %s %%%s\n", root.resultType, name)
		}
	}
	b.WriteString("}\n")
	return nil
}
