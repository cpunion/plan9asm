package plan9asm

import (
	"fmt"
	"strings"
)

// These positions are a private, opt-in observation of the selected SDK's
// lexer. Logical filenames never select files or become physical source roles.
// All existing CPP entry points keep this tape nil and their schema unchanged.
type sdkCPPPositionBase struct {
	File         string
	Anchor, Line uint
}

type sdkCPPTokenPosition struct {
	File               string
	Line               uint
	Expansion, Include int
	Base, FinalBase    sdkCPPPositionBase
	base               *sdkCPPPositionBase
}

type sdkCPPPositionToken struct {
	Kind     rune
	Text     string
	Position sdkCPPTokenPosition
}

type sdkCPPPositionObservation struct {
	Tokens []sdkCPPPositionToken
	Mapped string
}

func newSDKCPPFileBase(file string) *sdkCPPPositionBase {
	return &sdkCPPPositionBase{File: file, Anchor: 1, Line: 1}
}

// cmd/internal/src.MakePos clamps physical lines before applying a pragma.
func sdkCPPClampLine(line uint) uint {
	const maximum = 1<<20 - 2
	if line > maximum {
		return maximum
	}
	return line
}

func (reader *ppGoReader) positionToken(token *ppGoToken) {
	if reader.positionBase == nil {
		return
	}
	token.position = &sdkCPPTokenPosition{File: reader.file, Line: reader.positionLine,
		Expansion: reader.expansion, Include: reader.includeEvent,
		Base: *reader.positionBase, base: reader.positionBase}
}

func (observer *ppGoObservation) emitSDKPosition(token ppGoToken, final *sdkCPPPositionBase) error {
	if token.position == nil || final == nil {
		return fmt.Errorf("SDK CPP position has no complete lexer base")
	}
	position := *token.position
	position.FinalBase = *final
	position.base = nil // The retained tape never aliases mutable reader state.
	if err := observer.chargeToken(len(token.text) + len(position.File) + len(position.Base.File) + len(final.File)); err != nil {
		return err
	}
	observer.positions.Tokens = append(observer.positions.Tokens, sdkCPPPositionToken{
		Kind: token.kind, Text: token.text, Position: position,
	})
	return nil
}

func observeSDKAssemblyCPP(src string, opt AssemblyPreprocessOptions) (AssemblyCPPObservation, sdkCPPPositionObservation, error) {
	opt.Defines = append([]string(nil), opt.Defines...)
	observer := newPPGoObservation()
	observer.positions = &sdkCPPPositionObservation{}
	output, err := preprocessGoAssemblyObserved(src, opt, observer)
	if err != nil {
		return AssemblyCPPObservation{}, sdkCPPPositionObservation{}, err
	}
	observer.result.Output, observer.result.Defines = output, opt.Defines
	mapped, err := sdkCPPMappedAssembly(observer.positions.Tokens, ppGoObservationBytes-observer.tokenBytes)
	if err != nil {
		return AssemblyCPPObservation{}, sdkCPPPositionObservation{}, err
	}
	// A mapped file is not a new program certificate. Independently replay its
	// complete grammar and retain the actual original/mapped/archive objects.
	replayed, err := preprocessGoAssembly(mapped, AssemblyPreprocessOptions{FileName: opt.FileName})
	if err != nil || replayed != output {
		return AssemblyCPPObservation{}, sdkCPPPositionObservation{}, fmt.Errorf("SDK mapped CPP grammar differs: %v", err)
	}
	observer.positions.Mapped = mapped
	observer.tokenBytes += len(mapped)
	return observer.result, *observer.positions, nil
}

func sdkCPPMappedAssembly(tokens []sdkCPPPositionToken, remainingBytes int) (string, error) {
	var mapped strings.Builder
	var statement []sdkCPPPositionToken
	for _, token := range tokens {
		if token.Kind != '\n' {
			statement = append(statement, token)
		}
		if token.Kind != '\n' && token.Kind != ';' && token.Kind != ':' {
			continue
		}
		if len(statement) == 0 {
			continue
		}
		// The parser saves the first instruction's raw lexer line and asks for
		// its current base only after consuming the statement's terminator.
		base := token.Position.FinalBase
		line := sdkCPPClampLine(statement[0].Position.Line)
		if base.Line == 0 {
			line = 0
		} else {
			// Match src.Pos.RelLine's unsigned arithmetic exactly, including
			// statements which consume a pragma after their first token.
			line = base.Line + (line - base.Anchor)
		}
		if line >= uint(^uint(0)>>1) {
			return "", fmt.Errorf("SDK mapped CPP line overflows")
		}
		fmt.Fprintf(&mapped, "#line %d %q\n", line+1, base.File)
		for _, item := range statement {
			mapped.WriteString(item.Text)
			mapped.WriteByte(' ')
		}
		mapped.WriteByte('\n')
		if mapped.Len() > remainingBytes {
			return "", fmt.Errorf("SDK mapped CPP exceeds the existing byte bound")
		}
		statement = nil
	}
	if len(statement) != 0 {
		return "", fmt.Errorf("SDK mapped CPP has an unterminated statement")
	}
	return mapped.String(), nil
}
