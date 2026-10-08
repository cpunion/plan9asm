package plan9asm

import (
	"reflect"
	"testing"
	"text/scanner"
)

func TestSDKCPPPositionDefaultProjection(t *testing.T) {
	raw := `#include "physical.h"
TEXT ·F(SB),$0
INNER
#ifdef ABSENT
#line 91 "not-an-opened-source.s"
#endif
RET
`
	opened := 0
	opt := AssemblyPreprocessOptions{FileName: "original.s", ReadInclude: func(parent, name string) (string, []byte, error) {
		opened++
		if parent != "original.s" || name != "physical.h" {
			t.Fatal("logical metadata became an include origin", parent, name)
		}
		return "canonical/physical.h", []byte("#define INNER MOVL $7,AX\n"), nil
	}}
	plain, err := ObserveAssemblyCPP(raw, opt)
	if err != nil {
		t.Fatal(err)
	}
	actual, positions, err := observeSDKAssemblyCPP(raw, opt)
	if err != nil || !reflect.DeepEqual(actual, plain) || opened != 2 {
		t.Fatal("private position hook changed the existing CPP projection", err)
	}
	var macro, inactive bool
	for _, token := range positions.Tokens {
		if token.Text == "MOVL" {
			macro = token.Position.Expansion != 0 && token.Position.Base.File == "original.s" && token.Position.Line == 3
		}
		if token.Text == "RET" {
			base := token.Position.Base
			inactive = base.File == "not-an-opened-source.s" && base.Line == 91 && base.Anchor == 7
		}
		if token.Position.base != nil {
			t.Fatal("retained position tape aliases a private lexer reader")
		}
	}
	if !macro || !inactive {
		t.Fatal("macro invocation base or inactive #line observation was lost", positions)
	}
}

func TestSDKCPPPositionIncludeFinalBase(t *testing.T) {
	_, positions, err := observeSDKAssemblyCPP("#include \"literal.h\"\n\n", AssemblyPreprocessOptions{
		FileName: "original.s", ReadInclude: func(_, _ string) (string, []byte, error) {
			return "physical/literal.h", []byte("RET"), nil
		}})
	if err != nil {
		t.Fatal(err)
	}
	if len(positions.Tokens) != 2 || positions.Tokens[0].Position.File != "physical/literal.h" ||
		positions.Tokens[0].Position.Base.File != "literal.h" ||
		positions.Tokens[1].Position.FinalBase.File != "original.s" ||
		positions.Mapped != "#line 2 \"original.s\"\nRET \n" {
		t.Fatal("statement was mapped from its first token's base instead of the final lexer base", positions)
	}
}

func TestSDKCPPPositionScratchSharesEventBudgets(t *testing.T) {
	observer := newPPGoObservation()
	observer.positions = &sdkCPPPositionObservation{}
	observer.active = 0
	observer.result.Events = []AssemblyCPPEvent{{Ordinal: 1}}
	base := newSDKCPPFileBase("source.s")
	token := ppGoToken{kind: scanner.Ident, text: "RET", value: "RET", file: "source.s",
		position: &sdkCPPTokenPosition{File: "source.s", Line: 1, Base: *base, base: base}}
	observer.tokens = ppGoObservationTokens - 1
	if observer.capture(token) != nil || observer.emitSDKPosition(token, base) == nil || len(observer.positions.Tokens) != 0 {
		t.Fatal("private scratch obtained a second token quota")
	}
	observer.tokens, observer.tokenBytes = 0, ppGoObservationBytes-1
	if observer.emitSDKPosition(token, base) == nil {
		t.Fatal("private scratch obtained a second byte quota")
	}
	if _, err := sdkCPPMappedAssembly([]sdkCPPPositionToken{{Kind: scanner.Ident, Text: "RET", Position: *token.position},
		{Kind: '\n', Position: sdkCPPTokenPosition{FinalBase: *base}}}, 1); err == nil {
		t.Fatal("mapped bytes escaped the remaining shared budget")
	}
}
