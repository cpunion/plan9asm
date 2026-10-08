//go:build go1.27

package plan9asm

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/xgo-dev/plan9asm/internal/goasmcmd"
	"github.com/xgo-dev/plan9asm/internal/gostaticstorage"
)

// Explicit -std flags here exercise the offline decoder only. This owned
// first-party fixture is never passed off as a live authenticated SDK source.
func TestGoX86SDKASMDataCPPPositionFirstParty(t *testing.T) {
	for _, arch := range []string{"386", "amd64"} {
		t.Run(arch, func(t *testing.T) {
			pkg, profile, _, _ := numericalSBFixture(t, arch)
			ctx, cancel := runtimeOracleStageContext(t)
			defer cancel()
			raw := []byte(`#include "textflag.h"
#include "go_asm.h"
#include "position.h"

// Discarding this line changes compiler debug metadata, not just diagnostics.
TEXT ·Feature(SB),NOSPLIT,$0-1
INNER(const_Value)
#line 73 "logical-one.s"
MOVB AX,ret+0(FP)
#ifdef ABSENT
#line 137 "logical-inactive.s"
#endif
RET
DATA table<>(SB)/4,$const_Value
GLOBL table<>(SB),RODATA,$4
`)
			goSource := []byte("package numericalfixture\nconst Value=7\nfunc Feature() byte\n")
			for name, data := range map[string][]byte{"value.go": goSource, profile.AssemblyFile: raw,
				"position.h": []byte("#define INNER(v) MOVL $v,AX\n")} {
				if err := os.WriteFile(filepath.Join(profile.Directory, name), data, 0600); err != nil {
					t.Fatal(err)
				}
			}
			selection, err := goasmcmd.Run(ctx, profile.Directory, profile.Environment, profile.GoBinary,
				"list", "-compiled", "-export", "-json", "-gcflags=-std", "-asmflags=-std", ".")
			if err != nil {
				t.Fatal("actual complete first-party Go/ASM", err)
			}
			var selected goAMD64FrameSelection
			if json.Unmarshal(selection, &selected) != nil || selected.Standard || len(selected.SFiles) != 1 {
				t.Fatal("test -std flags changed source identity or complete member selection")
			}
			read := func(file string) []byte {
				data, err := os.ReadFile(file)
				if err != nil {
					t.Fatal(err)
				}
				return data
			}
			root := filepath.Dir(filepath.Dir(profile.GoBinary))
			header := filepath.Join(profile.Directory, "go_asm.h")
			if _, err := goasmcmd.Run(ctx, profile.Directory, profile.Environment, profile.GoBinary,
				"tool", "compile", "-std", "-p", pkg.Path, "-asmhdr", header,
				"-o", filepath.Join(profile.Directory, "header.o"), filepath.Join(profile.Directory, "value.go")); err != nil {
				t.Fatal(err)
			}
			cpp, positions, err := observeSDKAssemblyCPP(string(raw), AssemblyPreprocessOptions{
				FileName: filepath.Join(profile.Directory, profile.AssemblyFile),
				Defines:  GoAssemblerDefinesWithEnv("linux", arch, map[string]string{"GO386": "sse2", "GOAMD64": "v1"}),
				ReadInclude: func(parent, name string) (string, []byte, error) {
					file := filepath.Join(profile.Directory, name)
					if _, err := os.Stat(file); os.IsNotExist(err) {
						file = filepath.Join(root, "pkg", "include", name)
					}
					return file, read(file), nil
				}})
			if err != nil {
				t.Fatal(err)
			}
			expanded := filepath.Join(profile.Directory, "expanded.asm")
			if err := os.WriteFile(expanded, []byte(positions.Mapped), 0600); err != nil {
				t.Fatal(err)
			}
			input := gostaticstorage.SDKASMDataSource{Name: profile.AssemblyFile, Original: raw, Expanded: []byte(cpp.Output)}
			for _, role := range []string{"original", "expanded"} {
				file := filepath.Join(profile.Directory, profile.AssemblyFile)
				if role == "expanded" {
					file = expanded
				}
				object := filepath.Join(profile.Directory, role+".o")
				args := []string{"tool", "asm", "-std", "-p", pkg.Path, "-I", profile.Directory,
					"-I", filepath.Join(root, "pkg", "include"), "-o", object, file}
				if _, err := goasmcmd.Run(ctx, profile.Directory, profile.Environment, profile.GoBinary, args...); err != nil {
					t.Fatal(err)
				}
				if role == "original" {
					input.OriginalObject = read(object)
				} else {
					input.ExpandedObject = read(object)
				}
			}
			request := gostaticstorage.SDKASMDataObservationRequest{PackagePath: pkg.Path, Target: profile.Target,
				Archive: read(selected.Export), BuiltinSource: read(filepath.Join(root, "src", "cmd", "internal", "goobj", "builtinlist.go")),
				Sources: []gostaticstorage.SDKASMDataSource{input}}
			scratch := &GoX86SDKASMDataSource{observation: sdkASMDataLiveObservation{Request: GoX86SDKASMDataCaptureRequest{
				GoBinary: profile.GoBinary, Directory: profile.Directory, PackagePath: pkg.Path,
				Target: profile.Target, Environment: profile.Environment}, ObjectInputs: request,
				CPP: []AssemblyCPPObservation{cpp}, Positions: []sdkCPPPositionObservation{positions}}}
			sdkASMDataDiagnosticRetain(t, scratch)
			if _, err := gostaticstorage.ObserveX86SDKCPPASMData(request); err != nil {
				t.Fatal("complete first-party source/header/macro/lineage object join is not closed", err)
			}
		})
	}
}
