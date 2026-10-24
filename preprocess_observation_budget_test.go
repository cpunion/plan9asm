package plan9asm

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"text/scanner"
	"unsafe"
)

// A staging directory is not repeated source content. The libccallback
// regression charged its absolute filename once per emitted/captured token,
// so relocating unchanged assembly could exhaust the 16 MiB observation limit.
func TestPreprocessAssemblyCPPBudgetRelocation(t *testing.T) {
	const header = "#define VALUE PRESET\nMOVQ $PRESET, BX\n"
	source := "#include \"values.h\"\n" + strings.Repeat("MOVQ $VALUE, AX\n", 1024)
	prefix := strings.Repeat("stage/", 512)
	options := func(prefix string) AssemblyPreprocessOptions {
		return AssemblyPreprocessOptions{
			FileName: prefix + "entry.s", Defines: []string{"PRESET=17", "EMPTY="},
			ReadInclude: func(parent, name string) (string, []byte, error) {
				if parent != prefix+"entry.s" || name != "values.h" {
					return "", nil, fmt.Errorf("unexpected include %q from %q", name, parent)
				}
				return prefix + "values.h", []byte(header), nil
			},
		}
	}
	t.Run("captured-definition", func(t *testing.T) {
		definition := "#define MANY " + strings.Repeat("VALUE ", 8192) + "\n"
		short, err := ObserveAssemblyCPP(definition, options(""))
		if err != nil {
			t.Fatal(err)
		}
		long, err := ObserveAssemblyCPP(definition, options(prefix))
		if err != nil {
			t.Fatalf("relocating an unexpanded definition exhausted capture: %v", err)
		}
		ppTrimObservationPrefix(&long, prefix)
		if !reflect.DeepEqual(short, long) {
			t.Fatal("relocation changed the complete unexpanded macro capture")
		}
	})
	for _, retain := range []bool{false, true} {
		t.Run(fmt.Sprintf("output=%v", retain), func(t *testing.T) {
			short, shortTokens, err := observeAssemblyCPP(source, options(""), retain)
			if err != nil {
				t.Fatal(err)
			}
			long, longTokens, err := observeAssemblyCPP(source, options(prefix), retain)
			if err != nil {
				t.Fatalf("relocating unchanged source exhausted the observation: %v", err)
			}
			ppTrimObservationPrefix(&long, prefix)
			for index := range longTokens {
				longTokens[index].Position.File = strings.TrimPrefix(longTokens[index].Position.File, prefix)
			}
			if !reflect.DeepEqual(short, long) || !reflect.DeepEqual(shortTokens, longTokens) {
				t.Fatal("relocation changed source identities, output, event lineage or token positions")
			}
		})
	}
	t.Run("sdk-positions", func(t *testing.T) {
		short, shortPositions, err := observeSDKAssemblyCPP(source, options(""))
		if err != nil {
			t.Fatal(err)
		}
		long, longPositions, err := observeSDKAssemblyCPP(source, options(prefix))
		if err != nil {
			t.Fatalf("relocating unchanged SDK input exhausted the observation: %v", err)
		}
		ppTrimObservationPrefix(&long, prefix)
		for index := range longPositions.Tokens {
			position := &longPositions.Tokens[index].Position
			position.File = strings.TrimPrefix(position.File, prefix)
			position.Base.File = strings.TrimPrefix(position.Base.File, prefix)
			position.FinalBase.File = strings.TrimPrefix(position.FinalBase.File, prefix)
		}
		longPositions.Mapped = strings.ReplaceAll(longPositions.Mapped, prefix, "")
		if !reflect.DeepEqual(short, long) || !reflect.DeepEqual(shortPositions, longPositions) {
			t.Fatal("relocation changed SDK output, physical/logical origins or mapped grammar")
		}
	})
}

func ppTrimObservationPrefix(observation *AssemblyCPPObservation, prefix string) {
	for index := range observation.Sources {
		observation.Sources[index].File = strings.TrimPrefix(observation.Sources[index].File, prefix)
	}
	for index := range observation.Events {
		event := &observation.Events[index]
		event.Position.File = strings.TrimPrefix(event.Position.File, prefix)
		event.Included = strings.TrimPrefix(event.Included, prefix)
		for token := range event.Tokens {
			event.Tokens[token].Position.File = strings.TrimPrefix(event.Tokens[token].Position.File, prefix)
		}
	}
}

func TestPreprocessAssemblyCPPBudgetInternedOrigins(t *testing.T) {
	observer := newPPGoObservation()
	observer.retainOutput = true
	file := strings.Repeat("origin/", 16) + "entry.s"
	if err := observer.source(file, "RET\n"); err != nil {
		t.Fatal(err)
	}
	input := ppGoInput{observer: observer}
	event, err := input.beginObservation("macro", ppGoToken{file: strings.Clone(file)})
	if err != nil {
		t.Fatal(err)
	}
	token := ppGoToken{kind: scanner.Ident, text: "RET", value: "RET", file: strings.Clone(file)}
	if err := observer.capture(token); err != nil {
		t.Fatal(err)
	}
	input.finishObservation(event)
	for index := 0; index < 2; index++ {
		token.file = strings.Clone(file)
		if err := observer.emitOutput(token, 0, 3); err != nil {
			t.Fatal(err)
		}
	}
	if observer.tokenBytes != len(file)+12 || observer.tokens != 3 {
		t.Fatalf("shared origins were charged repeatedly: bytes=%d tokens=%d", observer.tokenBytes, observer.tokens)
	}
	origin := observer.result.Sources[0].File
	for _, retained := range []string{
		observer.result.Events[0].Position.File,
		observer.result.Events[0].Tokens[0].Position.File,
		observer.outputTokens[0].Position.File,
		observer.outputTokens[1].Position.File,
	} {
		if retained != origin || unsafe.StringData(retained) != unsafe.StringData(origin) {
			t.Fatal("single-charge origins still retain independent filename allocations")
		}
	}
	logical := "logical-only.s"
	observer.positions = &sdkCPPPositionObservation{}
	token.position = &sdkCPPTokenPosition{
		File: strings.Clone(file), Base: sdkCPPPositionBase{File: strings.Clone(logical)},
	}
	if err := observer.emitSDKPosition(token, &sdkCPPPositionBase{File: strings.Clone(logical)}); err != nil {
		t.Fatal(err)
	}
	position := observer.positions.Tokens[0].Position
	if observer.tokenBytes != len(file)+len(logical)+15 || observer.tokens != 4 ||
		unsafe.StringData(position.File) != unsafe.StringData(origin) ||
		unsafe.StringData(position.Base.File) != unsafe.StringData(position.FinalBase.File) {
		t.Fatal("SDK physical/logical names did not share the interned-name and token budgets")
	}
}

func TestPreprocessAssemblyCPPBudgetPredefineAndEmptySources(t *testing.T) {
	observer := newPPGoObservation()
	if _, err := preprocessGoAssemblyObserved("", AssemblyPreprocessOptions{
		FileName: "empty.s", Defines: []string{"EMPTY="},
	}, observer); err != nil {
		t.Fatal(err)
	}
	if observer.tokenBytes != len("empty.s")+len("<predefined>") {
		t.Fatalf("token-free source/event origins escaped accounting: %d", observer.tokenBytes)
	}
}

func TestPreprocessAssemblyCPPBudgetUniqueOriginsBound(t *testing.T) {
	observer := newPPGoObservation()
	observer.retainOutput = true
	for index := 0; index < ppGoObservationFiles; index++ {
		token := ppGoToken{file: fmt.Sprintf("origin-%d.s", index)}
		if err := observer.emitOutput(token, 0, 0); err != nil {
			t.Fatal(err)
		}
	}
	if err := observer.emitOutput(ppGoToken{file: "one-too-many.s"}, 0, 0); err == nil ||
		len(observer.outputTokens) != ppGoObservationFiles {
		t.Fatal("distinct origins escaped their bound or appended an invalid token")
	}
}

func TestPreprocessAssemblyCPPBudgetRejectsRealPayload(t *testing.T) {
	// Source and token count remain small; expansion genuinely retains more
	// than 16 MiB of payload even when every token has the same short origin.
	source := "#define LARGE \"" + strings.Repeat("x", 256<<10) + "\"\n" + strings.Repeat("LARGE\n", 70)
	observed, tokens, err := ObserveAssemblyCPPOutput(source, AssemblyPreprocessOptions{FileName: "s"})
	if err == nil || !strings.Contains(err.Error(), "token/byte bound") || tokens != nil ||
		!reflect.DeepEqual(observed, AssemblyCPPObservation{}) {
		t.Fatalf("over-budget payload received a successful observation: %v", err)
	}
	// Unique filename bytes also consume the same budget, even with no tokens.
	observed, tokens, err = ObserveAssemblyCPPOutput("", AssemblyPreprocessOptions{
		FileName: strings.Repeat("x", ppGoObservationBytes+1),
	})
	if err == nil || tokens != nil || !reflect.DeepEqual(observed, AssemblyCPPObservation{}) {
		t.Fatal("an oversized source identity escaped the byte bound")
	}
}
