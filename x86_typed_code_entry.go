package plan9asm

import (
	"fmt"
	"go/ast"
	"go/types"
	"reflect"
	"strings"
)

// A code entry is neither numerical DATA nor a Go funcval. This private role
// closes one complete address-producing function, not a general GP value rule.
// Both declarations and the actual retained ordinary TEXT bodies are required.
// The address denotes the emitted typed LLVM entry: a consumer must have that
// same signature. It is NOT a physical Go ABI0 machine-code entry.
type x86TypedCodeReturn struct {
	arch, triple, target, symbol string
	producer, callee             [32]byte
	signature, targetSig         FuncSig
}

func x86OrdinaryTypedCodeText(fn Func) bool {
	if !strings.HasPrefix(fn.Sym, "·") || strings.Contains(fn.Sym, "<") || len(fn.X86RawText) != 0 ||
		fn.x86SourceStack != nil || fn.x86FixedLocal != nil || fn.x86DirectTailRoot != nil ||
		fn.x86ContinuationFlow != nil || len(fn.x86ContinuationAddresses) != 0 || len(fn.x86IndirectLabels) != 0 {
		return false
	}
	for _, ins := range fn.Instrs {
		if !x86ResultNamedPredicateInstruction(ins) {
			return false
		}
		switch ins.Op {
		case "BYTE", "WORD", "LONG", "QUAD":
			return false
		}
	}
	return true
}

// Native-word MOV (including Go's AMOVQ aliases) and LEA materialize exactly
// one entry, followed only by full-word GP copies and one uintptr FP result.
// Partial writes, address arithmetic, branches, memory reads and indirect
// calls cannot borrow this role. No numerical initialization bit is set.
func x86TypedCodeReturnTarget(fn Func, arch string) (string, bool) {
	word := x86ResultPointerWord(arch)
	if arch != "amd64" && arch != "386" || !x86OrdinaryTypedCodeText(fn) ||
		fn.FrameSize != 0 || fn.ArgSize != word || len(fn.Instrs) < 4 || fn.Instrs[0].Op != OpTEXT {
		return "", false
	}
	gp := func(reg Reg) (int, bool) {
		index, _, valid := x86ResultGP(reg, word, arch)
		return index, valid && reg != SP && reg != BP && (arch != "amd64" || index != 14)
	}
	first := fn.Instrs[1]
	address, immediate, index, valid := x86StaticAddressValueMove(first, arch)
	if !valid || !immediate {
		var memory MemRef
		var width int64
		index, _, width, memory, valid = x86SBLEAWriteShape(first, arch)
		if !valid || width != word || memory.Base != "" || memory.Index != "" || memory.Scale != 0 {
			return "", false
		}
		var offset int64
		address.symbol, offset, valid = parseSBRef(memory.Sym)
		address.offset = offset + memory.Off
	}
	if !valid || address.offset != 0 || len(first.Args) != 2 {
		return "", false
	}
	if _, valid := gp(first.Args[1].Reg); !valid {
		return "", false
	}
	var entries [16]bool
	entries[index] = true
	for _, original := range fn.Instrs[2 : len(fn.Instrs)-2] {
		ins := x86GoMOVQSourceInstruction(original)
		if x86RawReturnScalarMoveBytes(ins.Op) != word || len(ins.Args) != 2 ||
			ins.Args[0].Kind != OpReg || ins.Args[1].Kind != OpReg {
			return "", false
		}
		source, sourceOK := gp(ins.Args[0].Reg)
		destination, destinationOK := gp(ins.Args[1].Reg)
		if !sourceOK || !destinationOK || !entries[source] {
			return "", false
		}
		entries[destination] = true
	}
	store := x86GoMOVQSourceInstruction(fn.Instrs[len(fn.Instrs)-2])
	ret := fn.Instrs[len(fn.Instrs)-1]
	if x86RawReturnScalarMoveBytes(store.Op) != word || len(store.Args) != 2 ||
		store.Args[0].Kind != OpReg || store.Args[1].Kind != OpFP || store.Args[1].FPOffset != 0 ||
		ret.Op != "RET" || len(ret.Args) != 0 {
		return "", false
	}
	index, valid = gp(store.Args[0].Reg)
	return address.symbol, valid && entries[index]
}

func goX86TypedCodeDeclaration(pkg GoPackage, fn Func, name string, options GoModuleOptions) (*types.Signature, FuncSig, bool) {
	if !x86OrdinaryTypedCodeText(fn) {
		return nil, FuncSig{}, false
	}
	declName := strings.TrimPrefix(fn.Sym, "·")
	object, ok := pkg.Types.Scope().Lookup(declName).(*types.Func)
	if !ok {
		return nil, FuncSig{}, false
	}
	// Scope/type metadata alone may be synthesized by a caller. Require the
	// same actual bodyless declaration selected alongside the assembly.
	declared := false
	for _, file := range pkg.Syntax {
		for _, decl := range file.Decls {
			function, ok := decl.(*ast.FuncDecl)
			if ok && function.Name.Name == declName && function.Name.Pos() == object.Pos() &&
				function.Body == nil && function.Recv == nil {
				declared = true
			}
		}
	}
	if !declared {
		return nil, FuncSig{}, false
	}
	declaration := object.Type().(*types.Signature)
	if declaration.Results().Len() > 1 {
		return nil, FuncSig{}, false
	}
	// Check original Go types before ABI0 result flattening can turn a
	// one-field aggregate (or an aggregate plus zero-sized results) into i64.
	// Function/map/channel handles are not an ordinary typed pointer contract
	// either. Named scalars and pointers to aggregates keep their exact types.
	for _, tuple := range []*types.Tuple{declaration.Params(), declaration.Results()} {
		for index := 0; index < tuple.Len(); index++ {
			if !goX86TypedCodeScalarType(tuple.At(index).Type()) {
				return nil, FuncSig{}, false
			}
		}
	}
	sig, err := goFuncSigForDeclaredFunc(name, object, options.GOARCH, options.Sizes, options.FrameSizes, true)
	if err != nil {
		return nil, FuncSig{}, false
	}
	// These scalar transports retain their LLVM calling signature after the
	// compiler's aggregate ABI lowering. Aggregates need a separate adapter.
	for _, typ := range append(append([]LLVMType(nil), sig.Args...), sig.Ret) {
		switch typ {
		case I8, I16, I32, I64, Ptr, "float", "double", Void:
		default:
			return nil, FuncSig{}, false
		}
	}
	return declaration, sig, true
}

func goX86TypedCodeScalarType(typ types.Type) bool {
	switch underlying := typ.Underlying().(type) {
	case *types.Pointer:
		return true
	case *types.Basic:
		switch underlying.Kind() {
		case types.Int, types.Int8, types.Int16, types.Int32, types.Int64,
			types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Uint64, types.Uintptr,
			types.Float32, types.Float64, types.UnsafePointer:
			return true
		}
	}
	return false
}

func bindGoX86TypedCodeReturns(pkg GoPackage, file *File, sigs map[string]FuncSig, options GoModuleOptions, resolve func(string) string) map[string]*x86TypedCodeReturn {
	if !x86CodeAddressTargetProfile(options.GOARCH, options.TargetTriple) {
		return nil
	}
	byName := make(map[string]Func, len(file.Funcs))
	counts := make(map[string]int, len(file.Funcs))
	for _, fn := range file.Funcs {
		byName[fn.Sym] = fn
		counts[resolve(fn.Sym)]++
	}
	var returns map[string]*x86TypedCodeReturn
	for _, fn := range file.Funcs {
		target, candidate := x86TypedCodeReturnTarget(fn, options.GOARCH)
		callee, retained := byName[target]
		if !candidate || !retained {
			continue
		}
		name, symbol := resolve(fn.Sym), resolve(target)
		if name == "" || symbol == "" || counts[name] != 1 || counts[symbol] != 1 ||
			resolveAssemblyDataSymbol(target, resolve) != symbol {
			continue
		}
		declaration, signature, valid := goX86TypedCodeDeclaration(pkg, fn, name, options)
		if !valid || declaration.Params().Len() != 0 || declaration.Results().Len() != 1 ||
			!reflect.DeepEqual(signature, sigs[name]) {
			continue
		}
		basic, integer := declaration.Results().At(0).Type().Underlying().(*types.Basic)
		if !integer || basic.Kind() != types.Uintptr {
			continue
		}
		_, targetSig, valid := goX86TypedCodeDeclaration(pkg, callee, symbol, options)
		if !valid || !reflect.DeepEqual(targetSig, sigs[symbol]) {
			continue
		}
		producer, producerErr := x86CallerSourceFingerprint(fn)
		body, bodyErr := x86CallerSourceFingerprint(callee)
		if producerErr != nil || bodyErr != nil {
			continue
		}
		if returns == nil {
			returns = make(map[string]*x86TypedCodeReturn)
		}
		returns[name] = &x86TypedCodeReturn{arch: options.GOARCH, triple: options.TargetTriple, target: target, symbol: symbol,
			producer: producer, callee: body, signature: signature, targetSig: targetSig}
	}
	return returns
}

// Recheck after normalization: raw/address-sensitive entries, coalesced tails,
// filtered targets and declaration-only externals must not become code entries.
// Any target emission failure discards the entire module, including producers.
func validateX86TypedCodeReturns(file *File, options Options, resolve func(string) string) error {
	if len(options.x86TypedCodeReturns) == 0 {
		return nil
	}
	functions := make(map[string]Func, len(file.Funcs))
	counts := make(map[string]int, len(file.Funcs))
	for _, fn := range file.Funcs {
		name := resolve(fn.Sym)
		functions[name] = fn
		counts[name]++
	}
	for name, entry := range options.x86TypedCodeReturns {
		fn, target := functions[name], functions[entry.symbol]
		producer, producerErr := x86CallerSourceFingerprint(fn)
		callee, calleeErr := x86CallerSourceFingerprint(target)
		if entry.arch != options.Goarch || entry.triple != options.TargetTriple ||
			!x86CodeAddressTargetProfile(options.Goarch, options.TargetTriple) || counts[name] != 1 || counts[entry.symbol] != 1 ||
			!x86OrdinaryTypedCodeText(fn) || !x86OrdinaryTypedCodeText(target) ||
			file.x86DirectTailHelper(target.Sym) || amd64IsRelocationAnchor(target, options.Sigs[entry.symbol]) ||
			resolve(entry.target) != entry.symbol || resolveAssemblyDataSymbol(entry.target, resolve) != entry.symbol ||
			producerErr != nil || calleeErr != nil || producer != entry.producer || callee != entry.callee ||
			!reflect.DeepEqual(options.Sigs[name], entry.signature) || !reflect.DeepEqual(options.Sigs[entry.symbol], entry.targetSig) {
			return fmt.Errorf("%w: typed code entry lost its retained ordinary source/signature", ErrProbeNeedsContext)
		}
	}
	return nil
}

func (c *amd64Ctx) checkTypedCodeReturn() (bool, error) {
	entry := c.fpTypedCodeReturn
	if entry == nil {
		return false, nil
	}
	target, valid := x86TypedCodeReturnTarget(c.fpScalarSource, c.goarch)
	producer, err := x86CallerSourceFingerprint(c.fpScalarSource)
	if !valid || entry.arch != c.goarch || entry.triple != c.targetTriple ||
		!x86CodeAddressTargetProfile(c.goarch, c.targetTriple) || target != entry.target || err != nil || producer != entry.producer ||
		c.resolve(target) != entry.symbol || resolveAssemblyDataSymbol(target, c.resolve) != entry.symbol ||
		!reflect.DeepEqual(c.fpSourceSig, entry.signature) || !reflect.DeepEqual(c.sigs[entry.symbol], entry.targetSig) {
		return true, fmt.Errorf("%w: typed code return lost its exact producer or target", ErrProbeNeedsContext)
	}
	return true, nil
}
