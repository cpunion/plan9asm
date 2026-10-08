package plan9asm

import (
	"crypto/sha256"
	"fmt"
)

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
	// Physical reader/expansion origin, not a logical #line diagnostic name.
	File string
	Line int
}

type AssemblyCPPObservedToken struct {
	// Kind and RecordedValue are the shared engine's token projection, not
	// the selected Go SDK's private ScanToken numbers or source acceptance.
	Kind                rune
	Text, RecordedValue string
	Position            AssemblyCPPObservedPosition
}

// Events are ordered by execution, not physical source traversal. Expansion is
// the invocation event which supplied a generated token; zero denotes physical
// source. Definition refers to the actual macro-definition event, including -D.
// Include refers to the active include which opened the physical reader.
// Invoked distinguishes macro expansion from a bare function-macro reference;
// Executed distinguishes a dispatched control/reference from a skipped control.
// Raw skipped-control payloads remain in the separate physical source inventory.
type AssemblyCPPEvent struct {
	Ordinal, Expansion, Definition, Include int
	Kind, Name, Directive                   string
	Position                                AssemblyCPPObservedPosition
	EnabledBefore, EnabledAfter             bool
	DepthBefore, DepthAfter                 int
	Function                                bool
	Executed, Invoked                       bool
	Parameters                              []string
	Tokens                                  []AssemblyCPPObservedToken
	Included                                string
}

// ObserveAssemblyCPP executes exactly the strict preprocessing engine on full
// source bytes. Failure returns no successful observation. The result retains
// active includes, inactive invocations and control transitions, not an inferred
// all-profile branch denominator or acceptance of isolated header fragments.
func ObserveAssemblyCPP(src string, opt AssemblyPreprocessOptions) (AssemblyCPPObservation, error) {
	// Capture the actual -D inputs before a resolver can mutate its caller's
	// slice. The same immutable copy supplies the execution and its record.
	opt.Defines = append([]string(nil), opt.Defines...)
	observer := newPPGoObservation()
	output, err := preprocessGoAssemblyObserved(src, opt, observer)
	if err != nil {
		return AssemblyCPPObservation{}, err
	}
	observer.result.Output = output
	observer.result.Defines = opt.Defines
	return observer.result, nil
}

// These observation bounds do not enlarge any existing source, include,
// recursion, output or corpus budget. Exhaustion is an error, never truncation.
const (
	ppGoObservationEvents = 100000
	ppGoObservationTokens = 1000000
	ppGoObservationBytes  = 16 << 20
	ppGoObservationFiles  = 4096
)

type ppGoObservation struct {
	result                     AssemblyCPPObservation
	sources                    map[string]AssemblyCPPObservedSource
	active, tokens, tokenBytes int
	eventLimit                 int
	positions                  *sdkCPPPositionObservation
}

func newPPGoObservation() *ppGoObservation {
	return &ppGoObservation{
		sources: make(map[string]AssemblyCPPObservedSource), active: -1,
		eventLimit: ppGoObservationEvents,
	}
}

func (observer *ppGoObservation) source(file, source string) error {
	identity := AssemblyCPPObservedSource{
		File: file, Bytes: len(source), SHA256: fmt.Sprintf("%x", sha256.Sum256([]byte(source))),
	}
	if previous, exists := observer.sources[file]; exists {
		if previous != identity {
			return fmt.Errorf("CPP observation source %q changed during execution", file)
		}
		return nil
	}
	if len(observer.sources) >= ppGoObservationFiles {
		return fmt.Errorf("CPP observation exceeds the explicit %d-source bound", ppGoObservationFiles)
	}
	observer.sources[file] = identity
	observer.result.Sources = append(observer.result.Sources, identity)
	return nil
}

func (input *ppGoInput) beginObservation(kind string, origin ppGoToken) (int, error) {
	observer := input.observer
	if observer == nil {
		return -1, nil
	}
	if len(observer.result.Events) >= observer.eventLimit {
		return -1, input.errorf("CPP observation exceeds the explicit %d-event bound", observer.eventLimit)
	}
	if observer.active != -1 {
		return -1, input.errorf("CPP observation has overlapping raw-token captures")
	}
	index := len(observer.result.Events)
	observer.result.Events = append(observer.result.Events, AssemblyCPPEvent{
		Ordinal: index + 1, Kind: kind, Expansion: origin.expansion, Include: origin.includeEvent,
		Position:      AssemblyCPPObservedPosition{File: origin.file, Line: origin.line},
		EnabledBefore: input.enabled(), DepthBefore: len(input.conditions),
	})
	observer.active = index
	return index, nil
}

func (input *ppGoInput) finishObservation(index int) {
	if index < 0 {
		return
	}
	event := &input.observer.result.Events[index]
	event.EnabledAfter, event.DepthAfter = input.enabled(), len(input.conditions)
	input.observer.active = -1
}

func (input *ppGoInput) observedEvent() *AssemblyCPPEvent {
	if input.observer == nil || input.observer.active < 0 {
		return nil
	}
	return &input.observer.result.Events[input.observer.active]
}

func (observer *ppGoObservation) capture(token ppGoToken) error {
	if observer.active < 0 {
		return nil
	}
	if err := observer.chargeToken(len(token.text) + len(token.value) + len(token.file)); err != nil {
		return err
	}
	event := &observer.result.Events[observer.active]
	event.Tokens = append(event.Tokens, AssemblyCPPObservedToken{
		Kind: token.kind, Text: token.text, RecordedValue: token.value,
		Position: AssemblyCPPObservedPosition{File: token.file, Line: token.line},
	})
	return nil
}

func (observer *ppGoObservation) chargeToken(bytes int) error {
	if observer.tokens >= ppGoObservationTokens || bytes > ppGoObservationBytes-observer.tokenBytes {
		return fmt.Errorf("CPP observation exceeds its explicit token/byte bound")
	}
	observer.tokens++
	observer.tokenBytes += bytes
	return nil
}

func ppGoCopyParameters(parameters []string) []string {
	if parameters == nil {
		return nil
	}
	result := make([]string, len(parameters))
	for index, parameter := range parameters {
		result[index] = parameter
	}
	return result
}
