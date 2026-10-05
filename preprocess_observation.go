package plan9asm

import "fmt"

// AssemblyCPPObservation describes one actual preprocessing execution. It is
// query-only: it supplies neither source authentication nor assembler, ABI or
// LLVM compilation evidence. Independently recapture the complete inputs before
// using it in a corpus proof.
type AssemblyCPPObservation struct {
	Output  string
	Defines []string
	Sources []AssemblyCPPObservedSource
	Events  []AssemblyCPPEvent
}

type AssemblyCPPObservedSource struct {
	File, SHA256 string
	Bytes        int
}

type AssemblyCPPObservedPosition struct {
	File string
	Line int
}

type AssemblyCPPObservedToken struct {
	Kind                rune
	Text, RecordedValue string
	Position            AssemblyCPPObservedPosition
}

// Events are ordered by execution, not physical source traversal. Expansion is
// the invocation event which supplied a generated token; zero denotes physical
// source. Definition refers to the actual macro-definition event, including -D.
// Raw skipped-control payloads remain in the separate physical source inventory.
type AssemblyCPPEvent struct {
	Ordinal, Expansion, Definition int
	Kind, Name, Directive          string
	Position                       AssemblyCPPObservedPosition
	EnabledBefore, EnabledAfter    bool
	DepthBefore, DepthAfter        int
	Function                       bool
	Parameters                     []string
	Tokens                         []AssemblyCPPObservedToken
	Included                       string
}

// ObserveAssemblyCPP executes exactly the strict preprocessing engine on full
// source bytes. Failure returns no successful observation. The result retains
// active includes, inactive invocations and control transitions, not an inferred
// all-profile branch denominator or acceptance of isolated header fragments.
func ObserveAssemblyCPP(src string, opt AssemblyPreprocessOptions) (AssemblyCPPObservation, error) {
	return AssemblyCPPObservation{}, fmt.Errorf("CPP observation is not implemented")
}
