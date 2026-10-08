//go:build go1.27

package plan9asm

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
	fixedLocalAPIs(t, pkg, file, bound, func(t *testing.T, ir string, err error) {
		if err != nil || !strings.Contains(ir, "fixed_local_native") || strings.Count(ir, "load volatile") < 22 {
			t.Fatal("complete frozen local/read consumer lost an independent owner or native read", err)
		}
	})
}

func numericalLocalFactory(t *testing.T) (context.Context, GoPackage, GoAMD64FrameCaptureRequest, *File, Options, *GoX86NumericalReadSource, context.CancelFunc) {
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
}
