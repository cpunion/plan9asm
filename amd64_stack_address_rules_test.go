package plan9asm

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func x86StackAddressEncoderCarrier(source []byte) error {
	file, err := parser.ParseFile(token.NewFileSet(), "asm6.go", source, 0)
	if err != nil {
		return err
	}
	canonical := func(node ast.Node) string {
		var output bytes.Buffer
		if err := format.Node(&output, token.NewFileSet(), node); err != nil {
			return ""
		}
		return output.String()
	}
	variables := make(map[string]*ast.CompositeLit)
	functions := make(map[string]*ast.FuncDecl)
	for _, declaration := range file.Decls {
		if function, ok := declaration.(*ast.FuncDecl); ok && function.Recv == nil {
			functions[function.Name.Name] = function
		}
		general, ok := declaration.(*ast.GenDecl)
		if !ok || general.Tok != token.VAR {
			continue
		}
		for _, spec := range general.Specs {
			value := spec.(*ast.ValueSpec)
			for index, name := range value.Names {
				if index < len(value.Values) {
					if literal, ok := value.Values[index].(*ast.CompositeLit); ok {
						variables[name.Name] = literal
					}
				}
			}
		}
	}
	const table = `[]ytab{
		{Zrp_, 1, argList{Yrl}},
		{Zm_o, 2, argList{Ym}},
		{Zib_, 1, argList{Yi8}},
		{Zil_, 1, argList{Yi32}},
	}`
	wantTable, err := parser.ParseExpr(table)
	if err != nil || variables["ypushl"] == nil || canonical(variables["ypushl"]) != canonical(wantTable) {
		return fmt.Errorf("complete ordered PUSH operand/encoder rows changed")
	}
	registrations := map[string]string{
		"APUSHL": `{APUSHL, ypushl, P32, opBytes{0x50, 0xff, 06, 0x6a, 0x68}}`,
		"APUSHQ": `{APUSHQ, ypushl, Py, opBytes{0x50, 0xff, 06, 0x6a, 0x68}}`,
		"APUSHW": `{APUSHW, ypushl, Pe, opBytes{0x50, 0xff, 06, 0x6a, 0x68}}`,
	}
	optab := variables["optab"]
	if optab == nil {
		return fmt.Errorf("missing complete Go opcode table")
	}
	seen := make(map[string]bool)
	for _, entry := range optab.Elts {
		row, ok := entry.(*ast.CompositeLit)
		if !ok || len(row.Elts) == 0 {
			continue
		}
		name := canonical(row.Elts[0])
		text, relevant := registrations[name]
		if !relevant {
			continue
		}
		want, err := parser.ParseExpr("[]Optab{" + text + "}")
		if err != nil || seen[name] || canonical(row) != canonical(want.(*ast.CompositeLit).Elts[0]) {
			return fmt.Errorf("complete PUSH registration changed: %s", name)
		}
		seen[name] = true
	}
	if len(seen) != len(registrations) {
		return fmt.Errorf("incomplete PUSH opcode inventory")
	}
	const absolute = `package reference
func useAbs(ctxt *obj.Link, s *obj.LSym) bool {
	if ctxt.Headtype == objabi.Hsolaris {
		return strings.HasPrefix(s.Name, "libc_")
	}
	return ctxt.Arch.Family == sys.I386 && !ctxt.Flag_shared
}`
	wantAbsolute, err := parser.ParseFile(token.NewFileSet(), "reference.go", absolute, 0)
	if err != nil || functions["useAbs"] == nil ||
		canonical(functions["useAbs"]) != canonical(wantAbsolute.Decls[0]) {
		return fmt.Errorf("absolute-relocation source mode changed")
	}
	const address = `package reference
func oclass() { switch kind {
case obj.TYPE_ADDR:
	switch a.Name {
	case obj.NAME_GOTREF:
		ctxt.Diag("unexpected TYPE_ADDR with NAME_GOTREF")
		return Yxxx
	case obj.NAME_EXTERN, obj.NAME_STATIC:
		if a.Sym != nil && useAbs(ctxt, a.Sym) { return Yi32 }
		return Yiauto
	case obj.NAME_AUTO, obj.NAME_PARAM:
		return Yiauto
	}
	if a.Sym != nil && strings.HasPrefix(a.Sym.Name, "runtime.duff") { return Yi32 }
	if a.Sym != nil || a.Name != obj.NAME_NONE { ctxt.Diag("unexpected addr: %v", obj.Dconv(p, a)) }
	fallthrough
} }`
	wantAddress, err := parser.ParseFile(token.NewFileSet(), "reference.go", address, 0)
	if err != nil || functions["oclass"] == nil {
		return fmt.Errorf("missing complete address classifier")
	}
	findAddress := func(function *ast.FuncDecl) *ast.CaseClause {
		var result *ast.CaseClause
		ast.Inspect(function.Body, func(node ast.Node) bool {
			clause, ok := node.(*ast.CaseClause)
			if !ok {
				return true
			}
			for _, expression := range clause.List {
				if canonical(expression) == "obj.TYPE_ADDR" {
					result = clause
				}
			}
			return true
		})
		return result
	}
	actualAddress := findAddress(functions["oclass"])
	if actualAddress == nil || canonical(actualAddress) != canonical(findAddress(wantAddress.Decls[0].(*ast.FuncDecl))) {
		return fmt.Errorf("complete named address classification changed")
	}
	return nil
}

func TestX86StackNamedAddressCompleteGoRuleCarrier(t *testing.T) {
	source, err := os.ReadFile(filepath.Join(runtime.GOROOT(), "src/cmd/internal/obj/x86/asm6.go"))
	if err != nil {
		t.Fatal(err)
	}
	if err := x86StackAddressEncoderCarrier(source); err != nil {
		t.Fatal(err)
	}
	for _, mutation := range []struct{ name, from, to string }{
		{"class", "{Zil_, 1, argList{Yi32}}", "{Zil_, 1, argList{Yiauto}}"},
		{"arity", "{Zil_, 1, argList{Yi32}}", "{Zil_, 1, argList{Yi32, Yrl}}"},
		{"encoder", "{Zil_, 1, argList{Yi32}}", "{Zib_, 1, argList{Yi32}}"},
		{"offset", "{Zil_, 1, argList{Yi32}}", "{Zil_, 2, argList{Yi32}}"},
		{"mode", "{APUSHL, ypushl, P32,", "{APUSHL, ypushl, Py,"},
		{"bytes", "{APUSHL, ypushl, P32, opBytes{0x50,", "{APUSHL, ypushl, P32, opBytes{0x51,"},
		{"shared", "return ctxt.Arch.Family == sys.I386 && !ctxt.Flag_shared", "return ctxt.Arch.Family == sys.I386"},
		{"absolute", "if a.Sym != nil && useAbs(ctxt, a.Sym) {", "if a.Sym != nil {"},
	} {
		t.Run(mutation.name, func(t *testing.T) {
			if !strings.Contains(string(source), mutation.from) {
				t.Fatal("mutation missed its actual source rule")
			}
			changed := strings.Replace(string(source), mutation.from, mutation.to, 1)
			if x86StackAddressEncoderCarrier([]byte(changed)) == nil {
				t.Fatal("changed source rule retained its family certificate")
			}
		})
	}
}

func TestX86StackNamedAddressSourceClassDoesNotGrantMemory(t *testing.T) {
	for _, spelling := range []string{
		"$·cell(SB)", "$cell<>+8(SB)", " \t$runtime·mainPC(SB)",
		"$·cell(SB)(AX)", "$·cell(SB)(R8)(X1*4)",
	} {
		operand, err := parseOperandForArch(ArchAMD64, strings.TrimSpace(spelling))
		if err != nil || !x86StackAbsoluteRelocationSource(operand, "386", 32) || isAMD64MemoryOperand(operand) {
			t.Fatalf("independent absolute source class changed: %s: %v", spelling, err)
		}
		before := operand
		for _, arch := range []string{"386", "amd64", "arm64"} {
			for _, bits := range []int{8, 16, 32, 64} {
				want := arch == "386" && bits == 32
				if x86StackAbsoluteRelocationSource(operand, arch, bits) != want || operand != before {
					t.Fatal("pure source classification changed operand identity or source mode")
				}
			}
		}
	}
	for _, spelling := range []string{
		"·cell(SB)", "$16(AX)", "$arg+0(FP)", "$local-8(SP)",
		"$16(SB)", "$·cell*(SB)", "$$·cell(SB)",
	} {
		operand, err := parseOperandForArch(ArchAMD64, spelling)
		if err == nil && x86StackAbsoluteRelocationSource(operand, "386", 32) {
			t.Fatal("unregistered source role acquired relocation permission:", spelling)
		}
	}
}
