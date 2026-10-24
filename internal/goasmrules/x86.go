// Package goasmrules reads the selected Go assembler's source rules for
// independent coverage checks. These are inventories, not translation, runtime
// or ABI certificates.
package goasmrules

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// X86AVXRule retains a complete operand-class row and its opcode-specific
// encoding table. Shared ytabs alone cannot establish broadcast, rounding,
// SAE, zeroing, tuple widths or the actual encoding generation.
type X86AVXRule struct {
	Opcode         string   `json:"opcode"`
	Ytab           string   `json:"ytab"`
	Prefix         string   `json:"prefix"`
	Classes        []string `json:"operand_classes"`
	RowIndex       int      `json:"row_index"`
	ZCase          string   `json:"zcase"`
	ZOffset        int      `json:"zoffset"`
	EncodingOffset int      `json:"encoding_offset"`
	OpBytes        []string `json:"op_bytes"`
}

func LoadX86AVXRules(goroot string) ([]X86AVXRule, error) {
	path := filepath.Join(goroot, "src", "cmd", "internal", "obj", "x86", "avx_optabs.go")
	source, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read Go AVX rules: %w", err)
	}
	file, err := parser.ParseFile(token.NewFileSet(), "avx_optabs.go", source, 0)
	if err != nil {
		return nil, fmt.Errorf("parse Go AVX rules: %w", err)
	}
	return X86AVXRules(file)
}

// X86AVXRules preserves source row order, including rows with equal operand
// classes or zero offsets. It intentionally does not infer frontend acceptance
// or physical register identities from the table.
func X86AVXRules(file *ast.File) ([]X86AVXRule, error) {
	variables := map[string]*ast.CompositeLit{}
	for _, declaration := range file.Decls {
		general, ok := declaration.(*ast.GenDecl)
		if !ok || general.Tok != token.VAR {
			continue
		}
		for _, spec := range general.Specs {
			value := spec.(*ast.ValueSpec)
			for i, name := range value.Names {
				if i >= len(value.Values) {
					continue
				}
				if literal, ok := value.Values[i].(*ast.CompositeLit); ok {
					variables[name.Name] = literal
				}
			}
		}
	}
	entries := variables["avxOptab"]
	if entries == nil {
		return nil, fmt.Errorf("Go AVX source has no literal avxOptab")
	}
	var rules []X86AVXRule
	for index, entry := range entries.Elts {
		row, ok := entry.(*ast.CompositeLit)
		if !ok {
			return nil, fmt.Errorf("AVX entry %d is not a literal rule", index)
		}
		opcode := expression(field(row, "as", 0))
		ytab := expression(field(row, "ytab", 1))
		prefix := expression(field(row, "prefix", 2))
		encoding, ok := field(row, "op", 3).(*ast.CompositeLit)
		if !strings.HasPrefix(opcode, "A") || len(opcode) < 2 || prefix == "" ||
			!ok || expression(encoding.Type) != "opBytes" {
			return nil, fmt.Errorf("AVX entry %d has an unrecognized opcode/prefix/encoding", index)
		}
		operands := variables[ytab]
		if operands == nil || len(operands.Elts) == 0 {
			return nil, fmt.Errorf("%s has no complete literal ytab %q", opcode, ytab)
		}
		data := make([]string, 0, len(encoding.Elts))
		for _, value := range encoding.Elts {
			if !encodingExpression(value) {
				return nil, fmt.Errorf("%s has an unsupported encoding expression %q", opcode, expression(value))
			}
			data = append(data, expression(value))
		}
		offset := 0
		for operandIndex, entry := range operands.Elts {
			operand, ok := entry.(*ast.CompositeLit)
			if !ok {
				return nil, fmt.Errorf("%s/%s row %d is not literal", opcode, ytab, operandIndex)
			}
			args, ok := field(operand, "args", 2).(*ast.CompositeLit)
			zcase := expression(field(operand, "zcase", 0))
			advance, err := strconv.ParseUint(expression(field(operand, "zoffset", 1)), 0, 8)
			if !ok || expression(args.Type) != "argList" || zcase == "" || err != nil {
				return nil, fmt.Errorf("%s/%s row %d has unrecognized operand rules", opcode, ytab, operandIndex)
			}
			classes := make([]string, 0, len(args.Elts))
			for _, arg := range args.Elts {
				name, ok := arg.(*ast.Ident)
				if !ok {
					return nil, fmt.Errorf("%s/%s has a non-identifier operand class", opcode, ytab)
				}
				classes = append(classes, name.Name)
			}
			rules = append(rules, X86AVXRule{
				Opcode: strings.TrimPrefix(opcode, "A"), Ytab: ytab, Prefix: prefix,
				Classes: classes, RowIndex: operandIndex, ZCase: zcase,
				ZOffset: int(advance), EncodingOffset: offset,
				OpBytes: append([]string(nil), data...),
			})
			offset += int(advance)
		}
	}
	return rules, nil
}

func field(row *ast.CompositeLit, name string, position int) ast.Expr {
	for _, entry := range row.Elts {
		if pair, ok := entry.(*ast.KeyValueExpr); ok && expression(pair.Key) == name {
			return pair.Value
		}
	}
	if position < len(row.Elts) {
		if _, keyed := row.Elts[position].(*ast.KeyValueExpr); !keyed {
			return row.Elts[position]
		}
	}
	return nil
}

func expression(value ast.Expr) string {
	if value == nil {
		return ""
	}
	var output bytes.Buffer
	if err := format.Node(&output, token.NewFileSet(), value); err != nil {
		return ""
	}
	return output.String()
}

func encodingExpression(value ast.Expr) bool {
	switch node := value.(type) {
	case *ast.Ident:
		return true
	case *ast.BasicLit:
		_, err := strconv.ParseUint(node.Value, 0, 8)
		return node.Kind == token.INT && err == nil
	case *ast.ParenExpr:
		return encodingExpression(node.X)
	case *ast.BinaryExpr:
		return node.Op == token.OR && encodingExpression(node.X) && encodingExpression(node.Y)
	}
	return false
}
