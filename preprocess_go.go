package plan9asm

import (
	"fmt"
	"strings"
	"text/scanner"
)

func preprocessGoAssembly(src string, options AssemblyPreprocessOptions) (string, error) {
	file := options.FileName
	if file == "" {
		file = "<assembly>"
	}
	input := ppGoInput{
		options: options, readers: []*ppGoReader{newPPGoReader(file, src)},
		macros: make(map[string]ppGoMacro), bol: true,
	}
	for _, definition := range options.Defines {
		name, value, assigned := strings.Cut(definition, "=")
		if !assigned {
			value = "1"
		}
		if !validPPIdentifier(name) {
			return "", fmt.Errorf("invalid predefined CPP macro %q", name)
		}
		reader := newPPGoReader("<predefined>", value)
		macro := ppGoMacro{}
		for {
			token, err := reader.next()
			if err != nil {
				return "", err
			}
			if token.kind == scanner.EOF {
				break
			}
			macro.body = append(macro.body, ppGoRecordedToken(token))
		}
		input.macros[name] = macro
	}
	var output strings.Builder
	var previous ppGoToken
	statementTokens := 0
	parserPositionInitialized := false
	for {
		token, err := input.next()
		if err != nil {
			return "", err
		}
		if token.kind == scanner.EOF {
			if statementTokens != 0 {
				return "", input.errorf("unexpected instruction EOF: missing newline or semicolon")
			}
			break
		}
		if token.kind == ppGoBuildComment {
			// Before the parser observes its first non-control token, Go's
			// diagnostic line is still zero and its errorf suppresses that
			// first-line CPP/build-comment diagnostic. Preserve this observed
			// acceptance instead of rejecting every comment after a control.
			if input.sawCode && parserPositionInitialized {
				return "", input.errorf("misplaced //go:build comment")
			}
			continue
		}
		parserPositionInitialized = true
		if token.kind != '\n' {
			input.sawCode = true
		}
		if token.kind == '\n' {
			if statementTokens != 0 {
				output.WriteByte('\n')
			}
			previous, statementTokens = ppGoToken{}, 0
			continue
		}
		if statementTokens != 0 && (ppGoTokenNeedsSpace(previous, token) ||
			statementTokens == 1 && previous.kind == scanner.Ident && token.kind != ':') {
			output.WriteByte(' ')
		}
		output.WriteString(token.text)
		previous = token
		statementTokens++
		if token.kind == ';' || token.kind == ':' {
			statementTokens = 0
		}
	}
	return output.String(), nil
}

func ppGoTokenNeedsSpace(left, right ppGoToken) bool {
	word := func(kind rune) bool {
		return kind == scanner.Ident || kind == scanner.Int || kind == scanner.Float
	}
	if word(left.kind) && word(right.kind) {
		return true
	}
	if left.text == "/" && (right.text == "*" || right.text == "/") {
		return true
	}
	// Removing whitespace must not turn separate source tokens into one
	// shift, rotate, arrow or numeric token on the second lexical pass.
	if (left.text == "<" && right.text == "<") || (left.text == ">" && right.text == ">") ||
		(left.text == "-" || left.text == "@") && right.text == ">" {
		return true
	}
	return (left.kind == scanner.Int || left.kind == scanner.Float) && right.text == "." ||
		left.text == "." && (right.kind == scanner.Int || right.kind == scanner.Float)
}
