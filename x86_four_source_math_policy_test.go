package plan9asm

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func x86FourSourceMathDriverPolicy(function *ast.FuncDecl) bool {
	if function.Body == nil {
		return false
	}
	calls := 0
	for _, statement := range function.Body.List {
		if declaration, ok := statement.(*ast.DeclStmt); ok {
			constants, ok := declaration.Decl.(*ast.GenDecl)
			if !ok || constants.Tok != token.CONST {
				return false
			}
			continue // Literal identity has its independent complete SHA contract.
		}
		expression, ok := statement.(*ast.ExprStmt)
		if !ok {
			return false // Conditional, deferred and asynchronous paths are not execution evidence.
		}
		call, ok := expression.X.(*ast.CallExpr)
		if !ok {
			return false
		}
		callee, ok := call.Fun.(*ast.Ident)
		if !ok || callee.Name != "x86FourSourceOriginalMathCompanion" || len(call.Args) != 3 {
			return false
		}
		for index, expected := range []string{"t", "source", "mainC"} {
			argument, ok := call.Args[index].(*ast.Ident)
			if !ok || argument.Name != expected {
				return false
			}
		}
		calls++
	}
	return calls == 1
}

func TestX86FourSourceOriginalMathDriverIsExplicitAndMandatory(t *testing.T) {
	x86FourSourceOriginalProgram(t)
	syntax, err := parser.ParseFile(token.NewFileSet(), "amd64_lower_four_source_test.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, declaration := range syntax.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if ok && function.Name.Name == "TestAMD64FourSourceRuntimeSemantics" {
			if !x86FourSourceMathDriverPolicy(function) {
				t.Fatal("the original runtime driver must invoke its explicit mathematical companion with unchanged complete ASM/C, without a native claim or optional skip")
			}
			return
		}
	}
	t.Fatal("the original four-source runtime driver is absent")
}

func TestX86FourSourceMathDriverPolicyRejectsFalseExecution(t *testing.T) {
	for _, scenario := range []struct {
		name, body string
		valid      bool
	}{
		{"actual", "x86FourSourceOriginalMathCompanion(t,source,mainC)", true},
		{"absent", "", false},
		{"conditioned", "if false { x86FourSourceOriginalMathCompanion(t,source,mainC) }", false},
		{"deferred", "defer x86FourSourceOriginalMathCompanion(t,source,mainC)", false},
		{"asynchronous", "go x86FourSourceOriginalMathCompanion(t,source,mainC)", false},
		{"twice", "x86FourSourceOriginalMathCompanion(t,source,mainC);x86FourSourceOriginalMathCompanion(t,source,mainC)", false},
		{"wrong ASM", "x86FourSourceOriginalMathCompanion(t,mainC,mainC)", false},
		{"wrong C", "x86FourSourceOriginalMathCompanion(t,source,source)", false},
		{"literal replacement", `x86FourSourceOriginalMathCompanion(t,"other",mainC)`, false},
		{"native claim", "compileAndRunRuntimeTest(t,source,mainC)", false},
		{"skip", "t.Skip(\"rare ISA absent\");x86FourSourceOriginalMathCompanion(t,source,mainC)", false},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			input := "package fixture\nfunc Observe(){const source=\"source\";const mainC=\"C\";" + scenario.body + "}"
			syntax, err := parser.ParseFile(token.NewFileSet(), "driver.go", strings.NewReader(input), 0)
			if err != nil {
				t.Fatal(err)
			}
			function, ok := syntax.Decls[0].(*ast.FuncDecl)
			if !ok || x86FourSourceMathDriverPolicy(function) != scenario.valid {
				t.Fatal("an independent false-execution policy control changed", scenario.valid)
			}
		})
	}
}
