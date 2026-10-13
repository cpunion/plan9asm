//go:build go1.27

package plan9asm

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGoAMD64SDKFrameOriginalSourceCapture(t *testing.T) {
	ctx, request := sdkASMDataRequest(t, "amd64", "crypto/md5")
	observed, err := CaptureGoAMD64SDKFrameObservation(ctx, request)
	if err != nil {
		t.Fatal("complete original SDK source/native frame precondition", err)
	}
	defer observed.Close()
	pkg := sdkAMD64FrameActualPackage(t, observed.source)
	const assembly = "md5block_amd64.s"
	source, err := CaptureGoAMD64SDKFrameSource(ctx, pkg, GoAMD64FrameCaptureRequest{
		GoBinary: request.GoBinary, Directory: request.Directory, AssemblyFile: assembly,
		Target: request.Target, Environment: request.Environment,
	})
	if err != nil {
		t.Fatal("independent SDK-primary native frame handoff is absent", err)
	}
	defer source.Close()
	fn, found := source.functions[request.PackagePath+".block"]
	if !found || len(source.functions) != 1 || fn.locals != 16 || fn.args != 32 {
		t.Fatal("complete original SDK frame/declaration/native layout", fn)
	}
	if err := source.Verify(ctx); err != nil {
		t.Fatal("live SDK-primary native frame", err)
	}
	// Even a fresh outer identity cannot borrow the SDK origin's live owner.
	copy := &GoAMD64FrameSource{sdkFrame: source.sdkFrame}
	copy.identity = copy
	if err := copy.Verify(ctx); !errors.Is(err, ErrProbeNeedsContext) {
		t.Fatal("copied SDK frame acquired the original native source", err)
	}
	if err := copy.Close(); !errors.Is(err, ErrProbeNeedsContext) {
		t.Fatal("copied SDK frame removed original artifacts", err)
	}
	if err := source.Verify(ctx); err != nil {
		t.Fatal("copied SDK frame revoked the authentic original", err)
	}
	file, err := Parse(ArchAMD64, source.sdkFrame.owner.source.observation.CPP[0].Output)
	if err != nil {
		t.Fatal(err)
	}
	options := Options{Goarch: "amd64", TargetTriple: "x86_64-unknown-linux-gnu",
		ResolveSym: func(symbol string) string { return request.PackagePath + "." + strings.TrimPrefix(symbol, "·") },
		Sigs:       map[string]FuncSig{request.PackagePath + ".block": fn.signature}}
	if _, err := source.BindFixedLocal(ctx, pkg, file, options); !errors.Is(err, ErrProbeNeedsContext) || !strings.Contains(err.Error(), "SHRQ") {
		t.Fatal("authentic frame supplied unrelated full body/caller permissions", err)
	}
	original := observed.source.observation.ObjectInputs.Sources[0].Original
	after, err := os.ReadFile(filepath.Join(request.Directory, assembly))
	if err != nil || !bytes.Equal(after, original) {
		t.Fatal("SDK source was replaced or cropped", err)
	}
	if err := source.Close(); err != nil {
		t.Fatal("SDK frame artifact cleanup", err)
	}
	if err := source.Verify(ctx); !errors.Is(err, ErrProbeNeedsContext) {
		t.Fatal("closed SDK frame retained its native source lifetime", err)
	}
}

func TestGoAMD64SDKFrameRejectsChangedProfile(t *testing.T) {
	ctx, request := sdkASMDataRequest(t, "amd64", "crypto/md5")
	observed, err := CaptureGoAMD64SDKFrameObservation(ctx, request)
	if err != nil {
		t.Fatal(err)
	}
	defer observed.Close()
	pkg := sdkAMD64FrameActualPackage(t, observed.source)
	source, err := CaptureGoAMD64SDKFrameSource(ctx, pkg, GoAMD64FrameCaptureRequest{GoBinary: request.GoBinary,
		Directory: request.Directory, AssemblyFile: "md5block_amd64.s", Target: request.Target, Environment: request.Environment})
	if err != nil {
		t.Fatal(err)
	}
	defer source.Close()
	// The factory owns an independent environment copy; mutating the caller's
	// request must neither update its sealed profile nor revoke its lifetime.
	request.Environment[0] += "changed"
	if err := source.Verify(ctx); err != nil {
		t.Fatal("caller environment aliases the private SDK profile", err)
	}
	original := source.request.Target
	source.request.Target = "windows/amd64"
	if err := source.Verify(ctx); !errors.Is(err, ErrProbeNeedsContext) {
		t.Fatal("changed target retained the original SDK frame", err)
	}
	source.request.Target = original
	if err := source.Verify(ctx); !errors.Is(err, ErrProbeNeedsContext) {
		t.Fatal("restoring a changed SDK profile revived its frame", err)
	}
}

func TestGoAMD64SDKFrameRejectsChangedNativeWitness(t *testing.T) {
	ctx, request := sdkASMDataRequest(t, "amd64", "crypto/md5")
	observed, err := CaptureGoAMD64SDKFrameObservation(ctx, request)
	if err != nil {
		t.Fatal(err)
	}
	defer observed.Close()
	pkg := sdkAMD64FrameActualPackage(t, observed.source)
	source, err := CaptureGoAMD64SDKFrameSource(ctx, pkg, GoAMD64FrameCaptureRequest{GoBinary: request.GoBinary,
		Directory: request.Directory, AssemblyFile: "md5block_amd64.s", Target: request.Target, Environment: request.Environment})
	if err != nil {
		t.Fatal(err)
	}
	defer source.Close()
	name := request.PackagePath + ".block"
	original := source.functions[name]
	changed := original
	changed.locals++
	source.functions[name] = changed
	if err := source.Verify(ctx); !errors.Is(err, ErrProbeNeedsContext) {
		t.Fatal("changed private native frame supplied backing", err)
	}
	source.functions[name] = original
	if err := source.Verify(ctx); !errors.Is(err, ErrProbeNeedsContext) {
		t.Fatal("restoring native metadata revived a revoked frame", err)
	}
}

func TestGoAMD64SDKFrameRejectsWholeOriginalPrivateHelper(t *testing.T) {
	ctx, request := sdkASMDataRequest(t, "amd64", "vendor/golang.org/x/crypto/chacha20poly1305")
	observed, err := CaptureGoAMD64SDKFrameObservation(ctx, request)
	if err != nil {
		t.Fatal(err)
	}
	defer observed.Close()
	pkg := sdkAMD64FrameActualPackage(t, observed.source)
	source, err := CaptureGoAMD64SDKFrameSource(ctx, pkg, GoAMD64FrameCaptureRequest{GoBinary: request.GoBinary,
		Directory: request.Directory, AssemblyFile: "chacha20poly1305_amd64.s", Target: request.Target, Environment: request.Environment})
	if source != nil {
		source.Close()
	}
	if source != nil || !errors.Is(err, ErrProbeNeedsContext) || !strings.Contains(err.Error(), "polyHashADInternal<>") ||
		!strings.Contains(err.Error(), "declared=false") || len(observed.source.observation.Frames[0].Records) != 3 {
		t.Fatal("unknown original helper became a Go-declared frame or whole source was cropped", err)
	}
}
