//go:build go1.27

package plan9asm

import (
	"bytes"
	"crypto/sha256"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func numericalSDKReferencedInputSource(t *testing.T, arch string) *GoX86NumericalReadSource {
	t.Helper()
	ctx, request, _ := numericalSDKCPURequest(t, arch)
	root := filepath.Dir(filepath.Dir(request.GoBinary))
	request.Directory = filepath.Join(root, "src", "internal", "bytealg")
	request.PackagePath = "internal/bytealg"
	request.AssemblyFile = "compare_" + arch + ".s"
	return numericalSDKInputSource(t, ctx, request)
}

func TestGoX86SDKNumericalReadActualReferencedArchiveHandoff(t *testing.T) {
	for _, arch := range []string{"386", "amd64"} {
		t.Run(arch, func(t *testing.T) {
			source := numericalSDKReferencedInputSource(t, arch)
			selected, exists := source.sdk.Imports["internal/cpu"]
			archive, retained := source.sdk.Referenced["internal/cpu"]
			if !exists || !retained || selected.Export == "" || archive.PackagePath != selected.ImportPath ||
				!reflect.DeepEqual(archive.Sources, selected.SFiles) || len(archive.Archive) == 0 {
				t.Fatal("actual SDK graph did not hand off the exact referenced compiler archive")
			}
			actual, err := os.ReadFile(selected.Export)
			if err != nil || !bytes.Equal(actual, archive.Archive) ||
				source.sdk.Inputs[selected.Export] != sha256.Sum256(actual) {
				t.Fatal("referenced SDK compiler bytes escaped their complete original input seal", err)
			}
			if source.ready || source.file != nil || len(source.reads) != 0 {
				t.Fatal("SDK indexed metadata handed off a numerical/native permission")
			}
			if err := source.verifySDKObservation(source.sdk.originalContext); err != nil {
				t.Fatal("actual same-SDK referenced graph cannot verify", err)
			}
		})
	}
}
