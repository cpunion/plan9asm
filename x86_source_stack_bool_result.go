package plan9asm

import (
	"reflect"
	"strings"
)

// This state belongs only to the already bounded physical source-stack CFG.
// It grants neither ordinary logical ABI0 entry nor caller/native memory. NUM
// and canonical bytes stay independent, including intermediate raw 2/3 values.
type x86SourceStackBoolState struct {
	bytes  x86ResultState
	owned  []bool // Immutable actual scalar layout.
	result int64
}

type x86SourceStackBoolProof struct {
	owner             *x86SourceStackProof
	minimum           int64
	word              int64
	stackIdentity     [32]byte
	countedIterations [4]int
	signature         [32]byte
	function          [32]byte
	original          []Instr
	normalized        []Instr
}

// A cost/denial predicate only. A stack-unobserved raw RET still transports
// physical ABI0 bool bytes; omitting explicit SP syntax cannot bypass the
// independent producer. Zero-argument/custom/native entries remain separate.
func x86SourceStackBoolCandidate(fn Func, sig FuncSig) bool {
	return fn.FrameSize == 0 && !strings.Contains(fn.Sym, "<") && sig.Ret == I1 &&
		len(sig.Args) != 0 && len(sig.ArgRegs) == 0 && len(sig.Frame.Results) == 1 && sig.Frame.Results[0].Type == I1
}

func x86SourceStackBoolArch(arch string) string {
	if arch == "" {
		return "amd64" // The existing ArchAMD64 renderer default, not an Options/source rewrite.
	}
	return arch
}

func x86SourceStackBoolInitial(fn Func, sig FuncSig, arch string) *x86SourceStackBoolState {
	arch = x86SourceStackBoolArch(arch)
	if (arch != "386" && arch != "amd64") || sig.Ret != I1 || fn.ArgSize > 4096 ||
		!x86SourceStackScalarABI0(sig, fn.ArgSize, arch) || sig.ARMEntry != nil ||
		sig.ARM64GoRegisterABI != nil || sig.ARM64ClosureABI != nil || sig.WASMContext != "" || sig.WASMNative ||
		sig.x86CallerRegions != nil || sig.x86FixedLocal != nil || sig.x86BorrowedBoolTail {
		return nil
	}
	result := sig.Frame.Results[0]
	if result.Index != 0 || result.Field != -1 || len(result.Fields) != 0 || result.Offset < 0 || result.Offset >= 4096 {
		return nil
	}
	state := &x86SourceStackBoolState{result: result.Offset, owned: make([]bool, result.Offset+1)}
	state.bytes.fp = make([]bool, len(state.owned))
	state.bytes.canonicalFP = make([]bool, len(state.owned))
	state.owned[result.Offset] = true
	for _, slot := range sig.Frame.Params {
		width := x86FrameScalarBytes(slot.Type, arch)
		if slot.Offset < 0 || slot.Offset > result.Offset-width {
			return nil
		}
		for offset := slot.Offset; offset < slot.Offset+width; offset++ {
			if state.owned[offset] {
				return nil
			}
			state.owned[offset], state.bytes.fp[offset] = true, true
			state.bytes.canonicalFP[offset] = slot.Type == I1
		}
	}
	return state
}

func (state *x86SourceStackBoolState) clone() *x86SourceStackBoolState {
	if state == nil {
		return nil
	}
	copy := *state
	copy.bytes = state.bytes.clone()
	return &copy
}

// Run only after the existing physical SP/BP, access and FLAGS producer has
// checked this exact instruction. Its access annotations are not public FP
// permission: only this local byte-effect copy uses their exact slot offsets.
func (state *x86SourceStackBoolState) step(ins Instr, arch string) bool {
	arch = x86SourceStackBoolArch(arch)
	if ins.Op == OpTEXT || ins.Op == OpLABEL {
		return true
	}
	if _, stack := amd64StackWidthSpecs[ins.Op]; stack {
		return true // Its saved NUM/canonical/FLAGS payload was transported together.
	}
	for _, operand := range ins.Args {
		if operand.Kind == OpReg && x86RawReturnStackRegister(operand.Reg) {
			// The producer allows only its exact SP/BP control effects here.
			// Physical address arithmetic cannot define numerical status bits.
			if effect, known, _ := x86TypedInstructionByteEffect(ins.Op, ins, arch); known && effect.flagsKilled != 0 {
				state.bytes.flags &^= effect.flagsKilled
			}
			return true
		}
	}
	local := ins
	local.Args = append([]Operand(nil), ins.Args...)
	for index, offset := range ins.x86SourceStackFP {
		if index < 0 || index >= len(local.Args) || local.Args[index].Kind != OpMem {
			return false
		}
		local.Args[index] = Operand{Kind: OpFP, FPOffset: offset}
	}
	beforeFP, beforeGP := state.bytes.canonicalFP, state.bytes.canonicalGP
	if !x86ResultTransfer(local, &state.bytes, state.owned, state.result, arch) {
		return false
	}
	state.bytes.canonicalFP, state.bytes.canonicalGP = x86ABI0CanonicalBytesAfter(local, beforeFP, beforeGP, arch)
	for offset := range state.bytes.canonicalFP {
		state.bytes.canonicalFP[offset] = state.bytes.canonicalFP[offset] && state.bytes.fp[offset]
	}
	for index := range state.bytes.canonicalGP {
		state.bytes.canonicalGP[index] &= state.bytes.gp[index]
	}
	return true
}

func sealX86SourceStackBool(fn, original Func, sig FuncSig, stack *x86SourceStackProof) *x86SourceStackBoolProof {
	function, err := x86CallerSourceFingerprint(fn)
	if err != nil {
		return nil
	}
	signature, err := x86CallerSourceFingerprint(sig)
	if err != nil {
		return nil
	}
	raw := original.Instrs
	if replay := original.x86RawReplay; replay != nil {
		if replay.validate(original, replay.goarch) != nil {
			return nil
		}
		raw = replay.original
	}
	return &x86SourceStackBoolProof{owner: stack, minimum: stack.minimum, word: stack.word,
		stackIdentity: stack.normalized, countedIterations: stack.countedIterations, signature: signature, function: function,
		original: cloneX86ReplayInstrs(raw), normalized: cloneX86ReplayInstrs(fn.Instrs)}
}

func (c *amd64Ctx) checkSourceStackBoolResultInitialization() error {
	fn, sig, stack := c.fpScalarSource, c.fpSourceSig, c.sourceStack
	fail := func() error {
		return goAMD64FrameContext("physical source-stack bool result lost its complete source/declaration/canonical proof")
	}
	if stack == nil || fn.x86SourceStack != stack || stack.boolResult == nil || stack.word != x86ResultPointerWord(c.goarch) ||
		fn.FrameSize != 0 || c.frameSize != fn.FrameSize || x86SourceStackBoolInitial(fn, sig, c.goarch) == nil {
		return fail()
	}
	proof := stack.boolResult
	if proof.owner != stack || proof.minimum != stack.minimum || proof.word != stack.word || stack.caller != nil ||
		proof.stackIdentity != stack.normalized || proof.countedIterations != stack.countedIterations {
		return fail()
	}
	function, err := x86CallerSourceFingerprint(fn)
	if err != nil || function != proof.function || !reflect.DeepEqual(fn.Instrs, proof.normalized) {
		return fail()
	}
	signature, err := x86CallerSourceFingerprint(sig)
	if err != nil || signature != proof.signature || fn.x86RawReplay == nil ||
		x86SourceStackBoolArch(fn.x86RawReplay.goarch) != x86SourceStackBoolArch(c.goarch) ||
		fn.x86RawReplay.validate(fn, fn.x86RawReplay.goarch) != nil || !reflect.DeepEqual(fn.x86RawReplay.original, proof.original) {
		return fail()
	}
	emitted := c.sig
	if sig.Attrs == "" && emitted.Attrs == c.fpFeatureAttrs {
		emitted.Attrs = "" // Only the renderer's separate, exact inference handoff.
	}
	if !reflect.DeepEqual(emitted, sig) {
		return fail()
	}
	return nil
}
