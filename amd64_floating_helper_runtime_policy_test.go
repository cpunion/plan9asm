package plan9asm

import (
	"crypto/sha256"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strconv"
	"strings"
	"testing"
)

// Bind the existing complete assembly builders and independent C references.
// A transport repair must not replace them with a smaller passing program.
var x86FloatingHelperOriginalRuntimeOwners = []struct {
	root, file, sourceSHA, cSHA string
}{
	{
		root: "TestAMD64GetExpRuntimeSemantics", file: "amd64_lower_getexp_test.go",
		sourceSHA: "e984e60362c11b28189a43d2e24efc2a25d9ef7df5d7ba750632dffc2ada582c",
		cSHA:      "755b471dca11961ef94e8103f354d27b1b737d265886465f6cca9ce18a953f88",
	},
	{
		root: "TestAMD64GetMantRuntimeSemantics", file: "amd64_lower_getmant_test.go",
		sourceSHA: "631e669e6d8ff1c65192ad74a54f7d5ba3bc847e3b0c64dac64d94f44c1dcbbe",
		cSHA:      "c9bf0ab8f11e0317569d87c4ec72a9815afddcdf599c6b05195f4ecd9f94a19e",
	},
	{
		root: "TestAMD64ScaleFRuntimeSemantics", file: "amd64_lower_scalef_test.go",
		sourceSHA: "9e0547ef11ff18488d9f093641beff6f2f88405e6c6aa8884d7b79bde75307ab",
		cSHA:      "3cec89474463e1884cc2979aea89520dd011589ab73d3c6051bfebc61bcc31c9",
	},
	{
		root: "TestAMD64FloatingRangeRuntimeSemantics", file: "amd64_lower_range_test.go",
		sourceSHA: "5ec0b3b10d651607d57d18991cb9dd1d91aeea18c6bbbaaa540c0c2970370cbb",
		cSHA:      "f8b2c3f77cb8f4c6902308df47a1a832c3d03a981471ffe44a837c93b457b359",
	},
	{
		root: "TestAMD64FloatingRangeDoubleAndScalarRuntimeSemantics", file: "amd64_lower_range_test.go",
		sourceSHA: "a5fac93bf84e828725448f9d6fdef423baf0ac0f1e1bdc7ce30537bbeb4f4377",
		cSHA:      "5dfe99148f760b1df90c2633526437cf02e932f2f98068d6fdb15eb6cee67409",
	},
	{
		root: "TestAMD64FixupImmediateRuntimeSemantics", file: "amd64_lower_fixupimm_test.go",
		sourceSHA: "6ba976350729649f866150d52bc624bc5e5628524295d544c3d5df2bbe79b3fd",
		cSHA:      "147083a3eb84d95f929c67ac71d7ef42932b2bbdc4c8d5b2d3eee81ed7ac5a27",
	},
	{
		root: "TestAMD64ScaledRoundRuntimeSemantics", file: "amd64_lower_scaled_round_test.go",
		sourceSHA: "c53a42b278d22ca13a3144f75dd58be991b8971ab70754f4bcffbbe952f2c413",
		cSHA:      "6777ff33b41286f21812a46efea484c3d368d58e66018feac6b194b2d8eaa228",
	},
	{
		root: "TestAMD64FPClassRuntimeSemantics", file: "amd64_lower_fpclass_test.go",
		sourceSHA: "754b23dc03e28506f8036180859683c6a07f49355752f846d8f5afcb56e4f078",
		cSHA:      "92e86c32eddff2266c607b5ca64b9616e86c46be09e1a9ca2948d1e51306bff8",
	},
}

func TestX86FloatingHelperRuntimeNativeOwnerAndPipelinePolicy(t *testing.T) {
	if len(x86FloatingHelperOriginalRuntimeOwners) != 8 {
		t.Fatal("complete original runtime owner inventory changed")
	}
	for _, owner := range x86FloatingHelperOriginalRuntimeOwners {
		t.Run(owner.root, func(t *testing.T) {
			data, err := os.ReadFile(owner.file)
			if err != nil {
				t.Fatal(err)
			}
			positions := token.NewFileSet()
			file, err := parser.ParseFile(positions, owner.file, data, 0)
			if err != nil {
				t.Fatal(err)
			}
			var function *ast.FuncDecl
			for _, declaration := range file.Decls {
				if candidate, ok := declaration.(*ast.FuncDecl); ok && candidate.Name.Name == owner.root {
					if function != nil {
						t.Fatal("duplicate original runtime owner")
					}
					function = candidate
				}
			}
			if function == nil || function.Body == nil {
				t.Fatal("missing original runtime owner")
			}
			start, end := token.NoPos, token.NoPos
			cSource := ""
			calls := make(map[string]int)
			for _, statement := range function.Body.List {
				ast.Inspect(statement, func(node ast.Node) bool {
					if identifier, ok := node.(*ast.Ident); ok && identifier.Name == "source" && start == token.NoPos {
						start = statement.Pos()
					}
					if call, ok := node.(*ast.CallExpr); ok {
						if identifier, ok := call.Fun.(*ast.Ident); ok {
							calls[identifier.Name]++
							if identifier.Name == "requireX86GoAssemblerResult" && end == token.NoPos {
								end = statement.Pos()
							}
						}
					}
					if specification, ok := node.(*ast.ValueSpec); ok && len(specification.Names) == 1 &&
						specification.Names[0].Name == "mainC" && len(specification.Values) == 1 {
						if literal, ok := specification.Values[0].(*ast.BasicLit); ok && literal.Kind == token.STRING {
							cSource, err = strconv.Unquote(literal.Value)
						}
					}
					return true
				})
			}
			if start == token.NoPos || end <= start || cSource == "" || err != nil {
				t.Fatal("original complete source/reference boundaries lost", err)
			}
			assemblyBuilder := strings.TrimSpace(string(data[positions.Position(start).Offset:positions.Position(end).Offset]))
			if actual := fmt.Sprintf("%x", sha256.Sum256([]byte(assemblyBuilder))); actual != owner.sourceSHA {
				t.Fatal("original complete assembly source builder changed", actual, owner.sourceSHA)
			}
			if actual := fmt.Sprintf("%x", sha256.Sum256([]byte(cSource))); actual != owner.cSHA {
				t.Fatal("independent original C values/reference changed", actual, owner.cSHA)
			}
			if calls["x86FloatingHelperRuntimeTools"] != 1 || calls["x86FloatingHelperRunNativePipelines"] != 1 ||
				calls["rosettaAvailable"] != 0 || calls["compileAndRunRuntimeTestForTarget"] != 0 {
				t.Fatal("original runtime must use its required native EVEX/DQ owner and all three execution pipelines", calls)
			}
		})
	}
}
