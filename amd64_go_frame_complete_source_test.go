package plan9asm

import (
	"crypto/sha256"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/packages"
)

// Original Go compilation is the acceptance oracle. Additional ordinary Go
// declarations do not themselves grant caller memory or fixed-local backing.
func TestGoAMD64FrameSourceCompleteOrdinaryGoDeclarations(t *testing.T) {
	for _, fixture := range []struct {
		name, imports, declarations string
	}{
		{"body", "", "func Reference(a uint64) uint64 { return a + 1 }"},
		{"values", "", "const Bias uint64 = 1\nvar Counter uint64"},
		{"unused-type", "", "type Record struct { Bytes []byte; Value uint64 }"},
		{"sdk-import", "import \"math/bits\"", "func Reference(a uint64) uint64 { return bits.RotateLeft64(a, 1) }"},
		{"unrelated-body-transport", "", "func Reference(a float64) float64 { return a + 1 }"},
		{"combined", "import \"math/bits\"", "const Bias = 1\nvar Counter uint64\nfunc Reference(a uint64) uint64 { return bits.RotateLeft64(a, Bias) }"},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			_, request := goAMD64FrameSourceFixture(t)
			text := "package framefixture\n" + fixture.imports + "\nfunc Value(a uint64) uint64\n" + fixture.declarations + "\n"
			if err := os.WriteFile(filepath.Join(request.Directory, "value.go"), []byte(text), 0600); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := runtimeOracleStageContext(t)
			defer cancel()
			env := append(append([]string(nil), request.Environment...),
				"GOOS=linux", "GOARCH=amd64", "CGO_ENABLED=0", "GOTOOLCHAIN=local",
				"GOENV=off", "GOWORK=off", "GOFLAGS=", "GOPROXY=off")
			selected, err := packages.Load(&packages.Config{
				Context: ctx, Dir: request.Directory, Env: env,
				Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles |
					packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesSizes | packages.NeedExportFile,
			}, ".")
			if err != nil || len(selected) != 1 {
				t.Fatalf("original Go complete-source selection: %v; packages=%d", err, len(selected))
			}
			actual := selected[0]
			if len(actual.Errors) != 0 || actual.IllTyped || actual.Types == nil || len(actual.Syntax) != 1 || actual.ExportFile == "" {
				t.Fatalf("original Go rejected complete ordinary source: %+v", actual.Errors)
			}
			export, err := os.ReadFile(actual.ExportFile)
			if err != nil || len(export) == 0 {
				t.Fatalf("original Go omitted compiled package/assembly export: %v", err)
			}
			t.Logf("ORIGINAL_GO_COMPLETE_FRAME_SOURCE role=%s go_sha=%x export_sha=%x", fixture.name,
				sha256.Sum256([]byte(text)), sha256.Sum256(export))
			pkg := GoPackage{Path: actual.PkgPath, Types: actual.Types, Syntax: actual.Syntax}
			source, err := CaptureGoAMD64FrameSource(ctx, pkg, request)
			if err != nil {
				t.Fatalf("Go-accepted ordinary declarations lost original frame observation: %v", err)
			}
			defer func() {
				if err := source.Close(); err != nil {
					t.Error(err)
				}
			}()
			function, found := source.functions[pkg.Path+".Value"]
			if !found || function.locals != 168 || function.args != 16 || function.transport != goAMD64FrameNumericTransport {
				t.Fatalf("complete source changed the independently captured numeric ASM frame: %+v", function)
			}
			if len(source.functions) != 1 || len(source.declarations) != 1 {
				t.Fatal("unrelated ordinary Go declarations became ASM frame capabilities")
			}
			if err := source.Verify(ctx); err != nil {
				t.Fatal(err)
			}
		})
	}
}
