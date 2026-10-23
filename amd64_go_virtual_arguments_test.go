package plan9asm

import (
	"errors"
	"fmt"
	"go/ast"
	"strings"
	"testing"

	"github.com/xgo-dev/llvm"
)

const amd64VirtualArgumentsPackage = "example.com/virtualargs"

const amd64VirtualArgumentsDeclarations = `package virtualargs
func CompareSlices(a, b []byte) int
func CompareStrings(a, b string) int
`

// Two independent public declarations share the complete private continuation.
// The fixture observes numerical lengths only: pointer identity transport must
// not be mistaken for permission to dereference, cast or expose those pointers.
const amd64VirtualArgumentsSource = `
TEXT ·CompareSlices<ABIInternal>(SB),4,$0-56
	MOVQ SI, DX
	MOVQ AX, SI
	JMP compareLengths<>(SB)

TEXT ·CompareStrings<ABIInternal>(SB),4,$0-40
	MOVQ AX, SI
	MOVQ DI, DX
	MOVQ CX, DI
	JMP compareLengths<>(SB)

TEXT compareLengths<>(SB),4,$0-0
	CMPQ BX, DX
	JEQ equal
	JLT less
	MOVQ $1, AX
	RET
less:
	MOVQ $-1, AX
	RET
equal:
	XORQ AX, AX
	RET
`

func amd64VirtualArgumentsRequest(t *testing.T, pkg GoPackage, source []byte, triple, api string, preprocessed bool,
	configure func(*GoModuleOptions),
) (string, error) {
	t.Helper()
	// This helper only translates supplied declarations in the current process.
	// It captures no stock-Go SDK, compiler archive, source lease or native ABI.
	return importedGlobalReadRequest(t, pkg, source, "amd64", triple, api, preprocessed, configure)
}

func amd64VirtualArgumentsRequireContext(t *testing.T, ir string, err error) {
	t.Helper()
	if ir != "" || !errors.Is(err, ErrProbeNeedsContext) {
		t.Fatalf("virtual language arguments granted unrelated state: output=%t err=%v", ir != "", err)
	}
}

func TestGoModuleAMD64VirtualArgumentsCompleteContinuation(t *testing.T) {
	for _, triple := range []string{"x86_64-unknown-linux-gnu", "x86_64-apple-darwin", "x86_64-pc-windows-msvc"} {
		for _, api := range []string{"public", "owned-context"} {
			for _, preprocessed := range []bool{false, true} {
				for _, linkedMain := range []bool{false, true} {
					t.Run(fmt.Sprintf("%s/%s/preprocessed=%t/main=%t", triple, api, preprocessed, linkedMain), func(t *testing.T) {
						pkg := mustGoPackage(t, amd64VirtualArgumentsPackage, amd64VirtualArgumentsDeclarations)
						if linkedMain {
							pkg.Path = "main" // Logical linked namespace, not Types.Path().
						}
						ir, err := amd64VirtualArgumentsRequest(t, pkg, []byte(amd64VirtualArgumentsSource), triple, api, preprocessed, nil)
						if err != nil || ir == "" {
							t.Fatal("actual loaded Go declarations did not initialize their virtual ABIInternal arguments", err)
						}
						for _, fragment := range []string{
							"extractvalue { ptr, i64, i64 } %arg0, 1",
							"extractvalue { ptr, i64, i64 } %arg1, 1",
							"extractvalue { ptr, i64 } %arg0, 1",
							"extractvalue { ptr, i64 } %arg1, 1",
						} {
							if !strings.Contains(ir, fragment) {
								t.Fatal("grouped LLVM parameters did not feed the original register leaves", fragment, ir)
							}
						}
						for _, name := range []string{"CompareSlices", "CompareStrings"} {
							found := false
							for _, line := range strings.Split(ir, "\n") {
								if strings.HasPrefix(line, "define i64 ") && strings.Contains(line, pkg.Path+"."+name) {
									found = strings.Count(line, "{ ptr, i64") == 2
								}
							}
							if !found {
								t.Fatal("public grouped Go signature was flattened, renamed or omitted", name, ir)
							}
						}
						if strings.Contains(ir, " nest ") || strings.Contains(ir, " swiftself ") || strings.Contains(ir, " noalias ") {
							t.Fatal("ordinary virtual parameters acquired hidden/native alias authority", ir)
						}
					})
				}
			}
		}
	}
}

func TestGoModuleAMD64VirtualArgumentsDeclaredWidths(t *testing.T) {
	for _, row := range []struct{ name, parameters, body, extraction string }{
		{"slice-length", "a []byte", "MOVQ BX,AX", "extractvalue { ptr, i64, i64 } %arg0, 1"},
		{"slice-capacity", "a []byte", "MOVQ CX,AX", "extractvalue { ptr, i64, i64 } %arg0, 2"},
		{"second-slice-length", "a,b []byte", "MOVQ SI,AX", "extractvalue { ptr, i64, i64 } %arg1, 1"},
		{"second-slice-capacity", "a,b []byte", "MOVQ R8,AX", "extractvalue { ptr, i64, i64 } %arg1, 2"},
		{"string-length", "a string", "MOVQ BX,AX", "extractvalue { ptr, i64 } %arg0, 1"},
		{"mixed-string-byte", "a string,b byte", "MOVBQZX CL,AX", ""},
		{"mixed-string-bool", "a string,b bool", "MOVBQZX CL,AX", ""},
		{"mixed-string-float32", "a string,b float32", "MOVL X0,AX", ""},
		{"mixed-string-float64", "a string,b float64", "MOVQ X0,AX", ""},
	} {
		t.Run(row.name, func(t *testing.T) {
			pkg := mustGoPackage(t, amd64VirtualArgumentsPackage, "package virtualargs\nfunc Entry("+row.parameters+") uint64\n")
			source := []byte("TEXT ·Entry<ABIInternal>(SB),4,$0\n" + row.body + "\nRET\n")
			ir, err := amd64VirtualArgumentsRequest(t, pkg, source, "x86_64-unknown-linux-gnu", "owned-context", true, nil)
			if err != nil || ir == "" || row.extraction != "" && !strings.Contains(ir, row.extraction) {
				t.Fatal("the original grouped parameter leaf was not transported at its own width", err, ir)
			}
		})
	}
}

func TestGoModuleAMD64VirtualArgumentsDoNotGrantOtherState(t *testing.T) {
	for _, row := range []struct{ name, parameters, body string }{
		{"unused-GP", "a string", "MOVQ R11,AX"},
		{"unused-vector", "a string", "MOVQ X0,AX"},
		{"hidden-context", "a string", "MOVQ 8(DX),AX"},
		{"stack", "a string", "MOVQ (SP),AX"},
		{"ABI0-home", "a string", "MOVQ a+0(FP),AX"},
		{"unknown-FLAGS", "a string", "SETEQ AL\nMOVBQZX AL,AX"},
		{"pointer-not-NUM", "a string", "MOVQ AX,X0\nMOVQ X0,AX"},
		{"byte-upper", "a string,b byte", "MOVQ CX,AX"},
		{"bool-upper", "a string,b bool", "MOVQ CX,AX"},
		{"float32-upper", "a string,b float32", "MOVQ X0,AX"},
		{"whole-array-spill", "a [2]uint64", "MOVQ AX,AX"},
		{"bank-overflow-spill", "a,b,c,d,e,f,g,h,i,j uint64", "MOVQ AX,AX"},
	} {
		t.Run(row.name, func(t *testing.T) {
			pkg := mustGoPackage(t, amd64VirtualArgumentsPackage, "package virtualargs\nfunc Entry("+row.parameters+") uint64\n")
			source := []byte("TEXT ·Entry<ABIInternal>(SB),4,$0\n" + row.body + "\nRET\n")
			ir, err := amd64VirtualArgumentsRequest(t, pkg, source, "x86_64-unknown-linux-gnu", "owned-context", true, nil)
			amd64VirtualArgumentsRequireContext(t, ir, err)
		})
	}
}

func TestGoModuleAMD64VirtualArgumentsNeedActualDeclarations(t *testing.T) {
	for _, mutation := range []string{"missing-syntax", "missing-function", "wrong-parameter", "wrong-result", "other-package"} {
		t.Run(mutation, func(t *testing.T) {
			pkg := mustGoPackage(t, amd64VirtualArgumentsPackage, "package virtualargs\nfunc Entry(a []byte) uint64\n")
			declarations := map[string]string{
				"missing-function": "package virtualargs\nfunc Another(a []byte) uint64\n",
				"wrong-parameter":  "package virtualargs\nfunc Entry(a uint64) uint64\n",
				"wrong-result":     "package virtualargs\nfunc Entry(a []byte) uint32\n",
				"other-package":    "package foreign\nfunc Entry(a []byte) uint64\n",
			}
			pkg.Syntax = nil
			if replacement := declarations[mutation]; replacement != "" {
				pkg.Syntax = mustGoPackage(t, amd64VirtualArgumentsPackage, replacement).Syntax
			}
			source := []byte("TEXT ·Entry<ABIInternal>(SB),4,$0\nMOVQ BX,AX\nRET\n")
			ir, err := amd64VirtualArgumentsRequest(t, pkg, source, "x86_64-unknown-linux-gnu", "owned-context", true, nil)
			amd64VirtualArgumentsRequireContext(t, ir, err)
		})
	}
}

func TestGoModuleAMD64VirtualArgumentsCannotLendToRawAPI(t *testing.T) {
	pkg := mustGoPackage(t, amd64VirtualArgumentsPackage, "package virtualargs\nfunc Entry(a []byte) uint64\n")
	// Raw APIs still accept their existing low-level register model. This
	// independent scalar/SIMD observation invokes its complete byte proof,
	// where a grouped shape alone cannot initialize BX as a slice length.
	source := "TEXT ·Entry<ABIInternal>(SB),4,$0\nMOVQ BX,AX\nMOVQ AX,X15\nMOVQ X15,AX\nRET\n"
	for _, api := range []string{"text", "module"} {
		t.Run(api, func(t *testing.T) {
			ir, err := x86ScalarFPBitViewPackageAPI(t, source, pkg, "amd64", "x86_64-unknown-linux-gnu", api)
			amd64VirtualArgumentsRequireContext(t, ir, err)
		})
	}
	t.Run("returned-signatures", func(t *testing.T) {
		ctx := llvm.NewContext()
		defer ctx.Dispose()
		resolve := testResolveSym(pkg.Path)
		translated, err := translateGoModuleInContext(ctx, pkg, []byte(source), GoModuleOptions{
			GOARCH: "amd64", TargetTriple: "x86_64-unknown-linux-gnu", SourcePreprocessed: true, ResolveSym: resolve,
		})
		if err != nil || translated == nil {
			t.Fatal("complete GoModule source failed before the borrowing control", err)
		}
		defer translated.Module.Dispose()
		if err := llvm.VerifyModule(translated.Module, llvm.ReturnStatusAction); err != nil {
			t.Fatal("invalid virtual argument module", err)
		}
		file, err := Parse(ArchAMD64, source)
		if err != nil {
			t.Fatal(err)
		}
		ir, err := Translate(file, Options{Goarch: "amd64", TargetTriple: "x86_64-unknown-linux-gnu",
			ResolveSym: resolve, Sigs: translated.Signatures})
		amd64VirtualArgumentsRequireContext(t, ir, err)
	})
}

func TestGoModuleAMD64VirtualArgumentsRecheckCallbacks(t *testing.T) {
	for _, mutation := range []string{"control", "syntax", "source", "resolver"} {
		t.Run(mutation, func(t *testing.T) {
			pkg := mustGoPackage(t, amd64VirtualArgumentsPackage, "package virtualargs\nfunc Entry(a []byte) uint64\n")
			source := []byte("TEXT ·Entry<ABIInternal>(SB),4,$0\nMOVQ BX,AX\nRET\n")
			called := false
			ir, err := amd64VirtualArgumentsRequest(t, pkg, source, "x86_64-unknown-linux-gnu", "owned-context", true, func(options *GoModuleOptions) {
				resolve := options.ResolveSym
				options.ResolveSym = func(symbol string) string {
					if mutation == "resolver" && called {
						return "foreign.Entry"
					}
					return resolve(symbol)
				}
				options.KeepFunc = func(_, _ string) bool {
					called = true
					if mutation == "syntax" {
						ast.Inspect(pkg.Syntax[0], func(node ast.Node) bool {
							if ident, ok := node.(*ast.Ident); ok && ident.Name == "byte" {
								ident.Name = "uint64"
							}
							return true
						})
					} else if mutation == "source" {
						source[len(source)-1] = ' '
					}
					return true // No source or public entry is filtered out.
				}
			})
			if !called {
				t.Fatal("the real source callback was not reached")
			}
			if mutation == "control" {
				if err != nil || ir == "" {
					t.Fatal("unmodified virtual argument source failed", err)
				}
			} else {
				amd64VirtualArgumentsRequireContext(t, ir, err)
			}
		})
	}
}
