package plan9asm

import (
	"fmt"
	"go/build/constraint"
	"strings"
	"text/scanner"
)

const ppGoBuildComment rune = -2001

// CPP operates on tokens, not on independent physical lines. A replacement
// can open a conditional which its parent closes, or define a macro used by a
// later original instruction. Keep lexical origins separate from #line names.
type ppGoToken struct {
	kind         rune
	text, value  string
	file         string
	line, endCol int
}

type ppGoReader struct {
	file    string
	scanner *scanner.Scanner
	err     error
	tokens  []ppGoToken
	index   int
	line    int
	include bool
}

func newPPGoReader(file, source string) *ppGoReader {
	reader := &ppGoReader{file: file}
	var lexical scanner.Scanner
	lexical.Init(strings.NewReader(source))
	lexical.Filename = file
	lexical.Whitespace = 1<<'\t' | 1<<'\r' | 1<<' '
	lexical.Mode = scanner.ScanChars | scanner.ScanFloats | scanner.ScanIdents |
		scanner.ScanInts | scanner.ScanStrings | scanner.ScanComments
	lexical.IsIdentRune = ppIdentRune
	lexical.Error = func(scan *scanner.Scanner, message string) {
		if reader.err == nil {
			reader.err = fmt.Errorf("CPP %s: %s", scan.Position, message)
		}
	}
	reader.scanner = &lexical
	return reader
}

func (reader *ppGoReader) next() (ppGoToken, error) {
	if reader.scanner == nil {
		if reader.index == len(reader.tokens) {
			return ppGoToken{kind: scanner.EOF, file: reader.file, line: reader.line}, nil
		}
		token := reader.tokens[reader.index]
		reader.index++
		// Go's lex.Slice reports sequential positions even when a generated
		// #define originally contained a space before its parenthesis.
		token.file, token.line, token.endCol = reader.file, reader.line, reader.index
		return token, nil
	}
	var kind rune
	for {
		kind = reader.scanner.Scan()
		if kind != scanner.Comment {
			break
		}
		// Go preserves build comments for the assembly parser to check their
		// position. Ordinary comments disappear without inventing newline tokens.
		if constraint.IsGoBuild(reader.scanner.TokenText()) {
			kind = ppGoBuildComment
			break
		}
	}
	text := reader.scanner.TokenText()
	if reader.err != nil {
		return ppGoToken{}, reader.err
	}
	// cmd/asm combines these operators before recording macro tokens. Do not
	// manufacture a shift from two separately scanned angle brackets.
	if kind == '<' || kind == '>' || kind == '-' || kind == '@' {
		next := reader.scanner.Peek()
		if kind == '<' && next == '<' || kind == '>' && next == '>' || (kind == '-' || kind == '@') && next == '>' {
			reader.scanner.Next()
			text += string(next)
		}
	}
	return ppGoToken{
		kind: kind, text: text, value: text, file: reader.file,
		line: reader.scanner.Position.Line, endCol: reader.scanner.Pos().Column,
	}, nil
}

// cmd/asm's lex.Make canonicalizes recorded replacement, argument and -D
// tokens, not physical source tokens. Keep text separately so emitting an
// identifier containing a dot/slash can retain its original lexical boundary.
func ppGoRecordedToken(token ppGoToken) ppGoToken {
	token.value = strings.ReplaceAll(strings.ReplaceAll(token.value, "·", "."), "∕", "/")
	return token
}

type ppGoMacro struct {
	parameters []string
	body       []ppGoToken
}

type ppGoInput struct {
	options    AssemblyPreprocessOptions
	readers    []*ppGoReader
	macros     map[string]ppGoMacro
	conditions []bool
	bol        bool
	sawCode    bool
	includes   int
	last       ppGoToken
}

func (input *ppGoInput) raw() (ppGoToken, error) {
	for {
		reader := input.readers[len(input.readers)-1]
		token, err := reader.next()
		if err != nil {
			return ppGoToken{}, err
		}
		if token.kind != scanner.EOF || len(input.readers) == 1 {
			input.last = token
			return token, nil
		}
		if reader.include {
			input.includes--
		}
		input.readers = input.readers[:len(input.readers)-1]
	}
}

func (input *ppGoInput) enabled() bool {
	return len(input.conditions) == 0 || input.conditions[len(input.conditions)-1]
}

func (input *ppGoInput) errorf(format string, args ...any) error {
	return fmt.Errorf("%s:%d: %s", input.last.file, input.last.line, fmt.Sprintf(format, args...))
}

func (input *ppGoInput) push(reader *ppGoReader) error {
	if len(input.readers) >= 100 {
		return input.errorf("CPP token-stack recursion exceeds the explicit 100-reader bound")
	}
	input.readers = append(input.readers, reader)
	return nil
}

func (input *ppGoInput) name() (string, error) {
	token, err := input.raw()
	if err != nil {
		return "", err
	}
	if token.kind != scanner.Ident {
		return "", input.errorf("invalid CPP macro name: %q", token.text)
	}
	return token.value, nil
}

func (input *ppGoInput) newline(directive string) error {
	token, err := input.raw()
	if err != nil {
		return err
	}
	if token.kind != '\n' {
		return input.errorf("invalid #%s: expected terminating newline, got %q", directive, token.text)
	}
	return nil
}

func (input *ppGoInput) next() (ppGoToken, error) {
	for nesting := 0; nesting < 100; {
		token, err := input.raw()
		if err != nil {
			return ppGoToken{}, err
		}
		if token.kind == '#' {
			if !input.bol {
				return ppGoToken{}, input.errorf("'#' must be first item on line")
			}
			input.bol, err = input.control(token)
			if err != nil {
				return ppGoToken{}, err
			}
			// Go returns a '#' wisp even for disabled controls. Its parser
			// uses this to distinguish a header build comment from a late one.
			input.sawCode = true
			continue
		}
		if token.kind == scanner.Ident {
			if macro, exists := input.macros[token.value]; exists {
				expanded, err := input.invoke(token, macro)
				if err != nil {
					return ppGoToken{}, err
				}
				if expanded {
					nesting++
					continue
				}
			}
		}
		if token.kind == scanner.EOF && len(input.conditions) != 0 {
			return ppGoToken{}, input.errorf("unclosed #ifdef or #ifndef")
		}
		input.bol = token.kind == '\n'
		if input.enabled() {
			return token, nil
		}
	}
	return ppGoToken{}, input.errorf("recursive macro invocation")
}

func (input *ppGoInput) invoke(invocation ppGoToken, macro ppGoMacro) (bool, error) {
	body := macro.body
	if macro.parameters != nil {
		opening, err := input.raw()
		if err != nil {
			return false, err
		}
		if opening.kind != '(' {
			return false, input.push(&ppGoReader{
				file: opening.file, line: opening.line, tokens: []ppGoToken{opening},
			})
		}
		actuals, err := input.arguments(invocation.value, macro.parameters)
		if err != nil {
			return false, err
		}
		body = nil
		for _, token := range macro.body {
			if replacement := actuals[token.value]; token.kind == scanner.Ident && replacement != nil {
				body = append(body, replacement...)
			} else {
				body = append(body, token)
			}
		}
	}
	return true, input.push(&ppGoReader{file: invocation.file, line: invocation.line, tokens: body})
}

func (input *ppGoInput) arguments(name string, parameters []string) (map[string][]ppGoToken, error) {
	var arguments [][]ppGoToken
	var argument []ppGoToken
	nesting := 0
	for {
		token, err := input.raw()
		if err != nil {
			return nil, err
		}
		if token.kind == scanner.EOF || token.kind == '\n' {
			return nil, input.errorf("unterminated arg list invoking macro: %s", name)
		}
		if nesting == 0 && (token.kind == ')' || token.kind == ',') {
			arguments = append(arguments, argument)
			argument = nil
			if token.kind == ')' {
				break
			}
			continue
		}
		if token.kind == '(' {
			nesting++
		} else if token.kind == ')' {
			nesting--
		}
		argument = append(argument, ppGoRecordedToken(token))
	}
	if len(parameters) == 0 && len(arguments) == 1 && arguments[0] == nil {
		arguments = nil
	}
	if len(arguments) != len(parameters) {
		return nil, input.errorf("wrong arg count for macro: %s", name)
	}
	actuals := make(map[string][]ppGoToken, len(parameters))
	for index, parameter := range parameters {
		actuals[parameter] = arguments[index]
	}
	return actuals, nil
}
