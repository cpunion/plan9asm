//go:build go1.27

package plan9asm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xgo-dev/llvm"
)

func TestGoX86NumericalSBReadComposesOnlyWithActualFixedLocalOwner(t *testing.T) {
	ctx, pkg, request, file, bound, _, _ := numericalLocalFactory(t)
	request.NumericalReads = bound.NumericalReads
	local, err := CaptureGoAMD64FrameSource(ctx, pkg, request)
	if err != nil {
		t.Fatal("independent actual Go framed object precondition failed", err)
	}
	defer local.Close()
	bound, err = local.BindFixedLocal(ctx, pkg, file, bound)
	if err != nil {
		t.Fatal("certified numerical reads did not compose with complete actual local proof", err)
	}
	numericalLocalAPIs(t, pkg, file, bound, func(t *testing.T, ir string, err error) {
		if err != nil || !strings.Contains(ir, "fixed_local_native") || strings.Count(ir, "load volatile") < 22 {
			t.Fatal("complete frozen local/read consumer lost an independent owner or native read", err)
		}
	})
}

func numericalLocalAPIs(t *testing.T, pkg GoPackage, file *File, options Options, check func(*testing.T, string, error)) {
	t.Helper()
	for _, api := range []string{"text", "module", "context", "go-module"} {
		t.Run(api, func(t *testing.T) {
			if api == "text" {
				ir, err := Translate(file, options)
				check(t, ir, err)
				return
			}
			context := llvm.NewContext()
			defer context.Dispose()
			var module llvm.Module
			var err error
			switch api {
			case "module":
				module, err = TranslateModule(file, options)
			case "context":
				module, err = TranslateModuleInContext(context, file, options)
			case "go-module":
				source := options.NumericalReads.source
				data, readErr := os.ReadFile(filepath.Join(source.request.Directory, source.request.AssemblyFile))
				if readErr != nil {
					t.Fatal(readErr)
				}
				result, translateErr := TranslateGoModule(pkg, data, GoModuleOptions{
					GOARCH: options.Goarch, GOOS: "linux", TargetTriple: options.TargetTriple,
					FileName: source.request.AssemblyFile, NumericalReads: options.NumericalReads,
					ResolveSym: options.ResolveSym,
				})
				err = translateErr
				if result != nil {
					module = result.Module
				}
			}
			ir := ""
			if !module.IsNil() {
				ir = module.String()
				module.Dispose()
			}
			check(t, ir, err)
		})
	}
}

func numericalLocalFactory(t *testing.T) (context.Context, GoPackage, GoAMD64FrameCaptureRequest, *File, Options, *GoX86NumericalReadSource, context.CancelFunc) {
	return numericalLocalFactoryWithContext(t, nil)
}

func numericalLocalFactoryWithContext(t *testing.T, wrap func(context.Context) context.Context) (context.Context, GoPackage, GoAMD64FrameCaptureRequest, *File, Options, *GoX86NumericalReadSource, context.CancelFunc) {
	t.Helper()
	pkg, request, initial, options := fixedLocalFrameFixture(t, "PADDQ 16(SP),X0")
	_, numericalProfile, _, _ := numericalSBFixture(t, "amd64")
	request.GoBinary, request.Environment = numericalProfile.GoBinary, numericalProfile.Environment
	options.ResolveSym = func(symbol string) string { return numericalNativeName(symbol, pkg.Path) }
	var err error
	request.Directory, err = filepath.EvalSymlinks(request.Directory)
	if err != nil {
		t.Fatal(err)
	}
	var reads strings.Builder
	for _, row := range numericalSBOracleRows {
		partner := "AX"
		if strings.HasPrefix(row.op, "CMP") {
			partner = "$1"
		}
		fmt.Fprintf(&reads, "%s internal∕cpu·X86+64(SB),%s\n", row.op, partner)
	}
	assembly := strings.Replace(fixedLocalVoidAssembly(initial), "MOVQ AX,16(SP)", reads.String()+"MOVQ AX,16(SP)", 1)
	if err := os.WriteFile(filepath.Join(request.Directory, request.AssemblyFile), []byte(assembly), 0600); err != nil {
		t.Fatal(err)
	}
	file, err := Parse(ArchAMD64, assembly)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := runtimeOracleStageContext(t)
	t.Cleanup(cancel)
	if wrap != nil {
		ctx = wrap(ctx)
	}
	owner, err := CaptureGoNumericalStorageOwner(ctx, GoNumericalStorageCaptureRequest{
		GoBinary: request.GoBinary, Directory: filepath.Join(filepath.Dir(filepath.Dir(request.GoBinary)), "src", "internal", "cpu"),
		PackagePath: "internal/cpu", Target: request.Target, Environment: request.Environment, SDKCPU: true,
		Variables: []GoNumericalStorageReference{{PackagePath: "internal/cpu", Name: "X86"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(owner.Close)
	numerical, err := CaptureGoX86NumericalReadSource(ctx, GoX86NumericalReadCaptureRequest{
		GoBinary: request.GoBinary, Directory: request.Directory, PackagePath: pkg.Path,
		AssemblyFile: request.AssemblyFile, Target: request.Target, Environment: request.Environment,
		Owners: []*GoNumericalStorageOwner{owner},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := numerical.Close(); err != nil {
			t.Error(err)
		}
	})
	if len(numerical.reads) != 22 {
		t.Fatal("actual complete original family did not bind all numerical reads", len(numerical.reads))
	}
	bound, err := numerical.BindNumericalReads(ctx, pkg, file, options)
	if err != nil {
		t.Fatal("independent original numerical read precondition failed", err)
	}
	return ctx, pkg, request, file, bound, numerical, cancel
}

func TestGoX86NumericalSBReadFrameProducerExactObjectJoin(t *testing.T) {
	ctx, pkg, request, _, bound, numerical, _ := numericalLocalFactory(t)
	unbound, err := CaptureGoAMD64FrameSource(ctx, pkg, request)
	if unbound != nil {
		unbound.Close()
	}
	if unbound != nil || !errors.Is(err, ErrProbeNeedsContext) {
		t.Fatal("default frame observation acquired numerical relocations", err)
	}
	request.NumericalReads = bound.NumericalReads
	frame, err := CaptureGoAMD64FrameSource(ctx, pkg, request)
	if err != nil {
		t.Fatal("independent complete original frame/numerical object join", err)
	}
	defer frame.Close()
	if err := frame.Verify(ctx); err != nil || len(frame.functions) != len(numerical.native.OriginalText) {
		t.Fatal("complete joined actual frame observation failed verification", err)
	}
	retainNumericalRead(t, numerical, "frame-join")
	retainNumericalFrame(t, frame)
}

func retainNumericalFrame(t *testing.T, frame *GoAMD64FrameSource) {
	t.Helper()
	artifact := os.Getenv("PLAN9ASM_STATIC_STORAGE_ARTIFACT_DIR")
	if artifact == "" {
		return
	}
	frame.mu.Lock()
	defer frame.mu.Unlock()
	folder := filepath.Join(artifact, "linux-amd64", "numerical-read-frame-join")
	if err := os.MkdirAll(folder, 0700); err != nil {
		t.Fatal(err)
	}
	for _, suffix := range []string{".o", ".o.listing"} {
		data, err := os.ReadFile(filepath.Join(frame.owned, frame.request.AssemblyFile+suffix))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(folder, "frame-original"+suffix), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	type function struct {
		ObjectSymbol, ListingSymbol string
		Locals, Args                int64
		Code                        []byte
		Positions                   map[int]string
		Signature                   FuncSig
		Transport                   goAMD64FrameTransport
		LanguageType                string
	}
	functions := make(map[string]function, len(frame.functions))
	for name, fn := range frame.functions {
		functions[name] = function{fn.objectSymbol, fn.listingSymbol, fn.locals, fn.args, fn.code,
			fn.positions, fn.signature, fn.transport, fn.languageType}
	}
	data, err := json.MarshalIndent(struct {
		Directory, GoBinary, AssemblyFile, Target string
		Environment                               []string
		ActualEnvironment                         map[string]string
		Selection                                 goAMD64FrameSelection
		Inputs                                    map[string][32]byte
		Functions                                 map[string]function
		Commands                                  []numericalSBCommand
		NumericalSeal                             [32]byte
	}{frame.request.Directory, frame.request.GoBinary, frame.request.AssemblyFile, frame.request.Target,
		frame.request.Environment, frame.env, frame.selection, frame.inputs, functions,
		frame.numericalCommands, frame.numericalSeal}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(folder, "frame-observation.json"), data, 0600); err != nil {
		t.Fatal(err)
	}
}
