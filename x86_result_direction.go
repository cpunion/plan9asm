package plan9asm

// Must-defined DF with an independent value domain. A mixed forward/backward
// join is defined but not a chosen direction. Unknown entry DF stays unknown;
// no ABI, declaration or initialized six-status mask seeds it.
type x86ResultDirection uint8

const (
	x86ResultDirectionUnknown  x86ResultDirection = 0
	x86ResultDirectionForward  x86ResultDirection = 1
	x86ResultDirectionBackward x86ResultDirection = 2
	x86ResultDirectionEither   x86ResultDirection = 3
)

func (direction x86ResultDirection) intersect(other x86ResultDirection) x86ResultDirection {
	if direction == x86ResultDirectionUnknown || other == x86ResultDirectionUnknown {
		return x86ResultDirectionUnknown
	}
	return direction | other
}

func x86ResultDirectionStep(ins Instr, state *x86ResultState, arch string) (bool, bool) {
	spec, known, err := x86DirectionControlProperties(ins.Op, ins, arch)
	if !known {
		return false, false
	}
	if err != nil || !x86ResultNamedPredicateInstruction(ins) {
		return false, true
	}
	state.direction = x86ResultDirectionForward
	if spec.backward {
		state.direction = x86ResultDirectionBackward
	}
	return true, true
}

// Applied after every independently accepted source transfer, including
// private CALL/stack roles. This never authorizes those operations. Unknown
// physical effects and callee/FLAGS restoration cannot retain a prior DF fact.
func x86ResultDirectionAfter(ins Instr, direction x86ResultDirection) x86ResultDirection {
	if x86MaskLogicalPreservesStorage(ins) {
		return direction
	}
	if x86RegisterQueryPreservesSourceState(ins) || x86WaitPackagePreservesStorage(ins) || x86NOPPreservesSourceState(ins) {
		return direction
	}
	if !x86ResultNamedPredicateInstruction(ins) || ins.Op == OpCALL {
		return x86ResultDirectionUnknown
	}
	if spec, known := amd64StackWidthSpecs[ins.Op]; known && spec.mode == amd64StackPopFlags {
		return x86ResultDirectionUnknown
	}
	if spec, known := amd64FarReturnSpecs[ins.Op]; known && spec.kind == amd64InterruptReturn {
		return x86ResultDirectionUnknown
	}
	return direction
}
