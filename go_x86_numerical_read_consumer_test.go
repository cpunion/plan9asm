//go:build go1.27

package plan9asm

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/xgo-dev/llvm"
)

func TestGoX86NumericalSBReadBoundConsumer(t *testing.T) {
	for _, arch := range []string{"386", "amd64"} {
		t.Run(arch, func(t *testing.T) {
			pkg, original, file, options := numericalSBFixture(t, arch)
			ctx, cancel := runtimeOracleStageContext(t)
			defer cancel()
			owner, err := CaptureGoNumericalStorageOwner(ctx, GoNumericalStorageCaptureRequest{
				GoBinary: original.GoBinary, Directory: filepath.Join(filepath.Dir(filepath.Dir(original.GoBinary)), "src", "internal", "cpu"),
				PackagePath: "internal/cpu", Target: original.Target, Environment: original.Environment, SDKCPU: true,
				Variables: []GoNumericalStorageReference{{PackagePath: "internal/cpu", Name: "X86"}},
			})
			if err != nil {
				t.Fatal(err)
			}
			defer owner.Close()
			source, err := CaptureGoX86NumericalReadSource(ctx, GoX86NumericalReadCaptureRequest{
				GoBinary: original.GoBinary, Directory: original.Directory, PackagePath: pkg.Path,
				AssemblyFile: original.AssemblyFile, Target: original.Target, Environment: original.Environment,
				Owners: []*GoNumericalStorageOwner{owner},
			})
			if err != nil {
				t.Fatal(err)
			}
			defer source.Close()
			bound, err := source.BindNumericalReads(ctx, pkg, file, options)
			if err != nil {
				t.Fatal("successful actual source/object/owner precondition failed at binding", err)
			}
			for _, api := range []string{"text", "module", "context"} {
				t.Run(api, func(t *testing.T) {
					var ir string
					var err error
					if api == "text" {
						ir, err = Translate(file, bound)
					} else {
						context := llvm.NewContext()
						var module llvm.Module
						if api == "module" {
							module, err = TranslateModule(file, bound)
						} else {
							module, err = TranslateModuleInContext(context, file, bound)
						}
						if !module.IsNil() {
							ir = module.String()
							module.Dispose()
						}
						context.Dispose()
					}
					if err != nil {
						t.Fatal("live exact numerical read still lacks coherent consumer effects", err)
					}
					if !strings.Contains(ir, "load volatile i8") || !strings.Contains(ir, "external global [") || strings.Contains(ir, "constant [") {
						t.Fatal("mutable numerical read lost exact live-load/global role", ir)
					}
				})
			}
		})
	}
}
