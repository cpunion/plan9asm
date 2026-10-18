package plan9asm

import (
	"fmt"
	"reflect"
)

// RIP constants are a decoder role, never a named-immediate broadcast form.
// Complete source replay below binds their original bytes. Block reads also
// retain the exact DATA identity checked by the independent raw-return guard.
func x86RawBroadcastLiteralForm(ins Instr, arch string) bool {
	if arch != "amd64" || !ins.x86Encoded || !ins.x86RIPLiteral ||
		ins.x86SourceRaw == nil || ins.x86RawNearReturn != nil || ins.x86AddressBits != 0 ||
		ins.x86UnboundAddressOverride || ins.x86SegmentPrefix != 0 {
		return false
	}
	form, known, err := x86BroadcastProperties(ins.Op, ins, arch)
	if !known || err != nil {
		return false
	}
	if ins.Args[0].Kind == OpImm {
		return ins.x86SourceRaw.op == ins.Op && reflect.DeepEqual(ins.x86SourceRaw.args, ins.Args) &&
			(form.kind == x86BroadcastInteger || form.kind == x86BroadcastFloat)
	}
	return form.kind == x86BroadcastBlock && x86RawReturnDecodedReadIdentity(ins) &&
		ins.x86RIPMemoryRead.source == 0 && ins.x86RIPMemoryRead.width == int64(form.sourceBytes)
}

// This bounded source uses the ordinary typed broadcast byte effects. It does
// not join the conditional native-read route or waive raw RET's separately
// checked frame, memory and continuation contracts. Other raw families retain
// their own consumers. Declaration-only queries acquire no literal permission.
func (c *amd64Ctx) checkRawBroadcastLiteralResults() (bool, error) {
	fn, arch := c.fpScalarSource, c.goarch
	selected := false
	for _, ins := range fn.Instrs {
		selected = selected || x86RawBroadcastLiteralForm(ins, arch)
	}
	if !selected {
		return false, nil
	}
	fail := func() (bool, error) {
		return true, fmt.Errorf("%w: raw broadcast requires its complete literal source and independent return contract", ErrProbeNeedsContext)
	}
	if fn.FrameSize != 0 || len(fn.X86RawText) != 0 || fn.x86SourceStack != nil ||
		fn.x86FixedLocal != nil || fn.x86DirectTailRoot != nil || fn.x86ContinuationFlow != nil ||
		fn.x86BoolCalls != nil || fn.x86RawReplay == nil {
		return fail()
	}
	if err := fn.x86RawReplay.validate(fn, arch); err != nil {
		return true, err
	}
	hasReturn := false
	for _, ins := range fn.Instrs {
		if form := ins.x86RawNearReturn; form != nil {
			if ins.Op != OpRET || !form.nativeWidth || form.cleanup != 0 {
				return fail()
			}
			hasReturn = true
		} else if !x86RawBroadcastLiteralForm(ins, arch) && !x86ResultNamedPredicateInstruction(ins) {
			return fail()
		}
	}
	if !hasReturn {
		return fail()
	}
	transfer := func(_ int, ins Instr, state *x86ResultState, owned []bool, resultOffset int64, arch string) bool {
		state.loweredStaticReads = c.fpScalarData
		defer func() { state.loweredStaticReads = nil }()
		if x86RawBroadcastLiteralForm(ins, arch) {
			accepted, _ := x86ResultBroadcastInputs(ins, state, owned, arch, false, true)
			return accepted
		}
		return x86ScalarFPResultTransfer(ins, state, owned, resultOffset, arch)
	}
	return true, checkX86ABI0ResultInitializationAtSource(fn, c.sig, arch, transfer, true)
}
