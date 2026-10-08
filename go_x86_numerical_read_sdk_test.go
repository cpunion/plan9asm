//go:build go1.27

package plan9asm

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xgo-dev/plan9asm/internal/goasmcmd"
)

func numericalSDKCRC32Request(t *testing.T) (context.Context, GoX86NumericalReadCaptureRequest, *GoNumericalStorageOwner) {
	t.Helper()
	_, profile, _, _ := numericalSBFixture(t, "amd64")
	ctx, cancel := runtimeOracleStageContext(t)
	t.Cleanup(cancel)
	root := filepath.Dir(filepath.Dir(profile.GoBinary))
	owner, err := CaptureGoNumericalStorageOwner(ctx, GoNumericalStorageCaptureRequest{
		GoBinary: profile.GoBinary, Directory: filepath.Join(root, "src", "internal", "cpu"),
		PackagePath: "internal/cpu", Target: profile.Target, Environment: profile.Environment,
		SDKCPU: true, Variables: []GoNumericalStorageReference{{PackagePath: "internal/cpu", Name: "X86"}},
	})
	if err != nil {
		t.Fatal("actual SDK CPU owner precondition", err)
	}
	t.Cleanup(owner.Close)
	request := GoX86NumericalReadCaptureRequest{
		GoBinary: profile.GoBinary, Directory: filepath.Join(root, "src", "hash", "crc32"),
		PackagePath: "hash/crc32", AssemblyFile: "crc32_amd64.s", Target: profile.Target,
		Environment: profile.Environment, Owners: []*GoNumericalStorageOwner{owner},
	}
	data, err := os.ReadFile(filepath.Join(request.Directory, request.AssemblyFile))
	if err != nil || !strings.Contains(string(data), "CMPB internal∕cpu·X86+const_offsetX86HasAVX512VPCLMULQDQL(SB), $1") {
		t.Fatal("actual unchanged SDK crc32 counterexample is absent", err)
	}
	actual, err := goasmcmd.Run(ctx, request.Directory, request.Environment, request.GoBinary,
		"list", "-json", "-compiled", "-export", "-mod=readonly", ".")
	if err != nil {
		t.Fatal("stock Go failed the original complete CRC32 package", err)
	}
	var selected goAMD64FrameSelection
	if err := json.Unmarshal(actual, &selected); err != nil || !selected.Standard || selected.Module != nil ||
		selected.ImportPath != request.PackagePath || selected.Dir != request.Directory || selected.Export == "" {
		t.Fatal("actual SDK selection differs from the exact package", err)
	}
	archive, err := os.ReadFile(selected.Export)
	if err != nil || len(archive) == 0 {
		t.Fatal("actual selected SDK compiler archive is absent", err)
	}
	if artifacts := os.Getenv("PLAN9ASM_STATIC_STORAGE_ARTIFACT_DIR"); artifacts != "" {
		folder := filepath.Join(artifacts, "sdk-crc32-original")
		if err := os.MkdirAll(folder, 0700); err != nil {
			t.Fatal(err)
		}
		for name, bytes := range map[string][]byte{"source.s": data, "selection.json": actual, "original.a": archive} {
			if err := os.WriteFile(filepath.Join(folder, name), bytes, 0600); err != nil {
				t.Fatal(err)
			}
		}
	}
	return ctx, request, owner
}

func TestGoX86SDKNumericalReadDefaultOwnedRemainsContext(t *testing.T) {
	ctx, request, owner := numericalSDKCRC32Request(t)
	source, err := CaptureGoX86NumericalReadSource(ctx, request)
	if source != nil {
		source.Close()
	}
	if source != nil || !errors.Is(err, ErrProbeNeedsContext) || !strings.Contains(err.Error(), "owned source role") {
		t.Fatal("default owned factory borrowed SDK source identity", err)
	}
	if err := owner.Verify(ctx); err != nil {
		t.Fatal("refusing SDK source revoked the independent data owner", err)
	}
}

func TestGoX86SDKNumericalReadOriginalCRC32KeepsHeaderContext(t *testing.T) {
	ctx, request, owner := numericalSDKCRC32Request(t)
	source, err := CaptureGoX86SDKNumericalReadSource(ctx, request)
	if source != nil {
		source.Close()
	}
	if source != nil || !errors.Is(err, ErrProbeNeedsContext) || !strings.Contains(err.Error(), "generated/relative header") {
		t.Fatal("explicit SDK source authentication did not reach the unchanged generated-header boundary", err)
	}
	if err := owner.Verify(ctx); err != nil {
		t.Fatal("later header Context revoked the independent SDK CPU data owner", err)
	}
}
