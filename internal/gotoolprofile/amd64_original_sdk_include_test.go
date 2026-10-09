package gotoolprofile

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// This low-level source-retention fixture copies only the original driver
// registration source files. It is not a public SDK observation or frame owner.
func sdkDependencyIncludeFixture(t *testing.T, assembly string) (amd64OriginalListPackage, string, *Observation, *amd64OriginalObservationFiles) {
	t.Helper()
	request := amd64OriginalObservationFixture(t, "bodyful")
	ctx, cancel := amd64OriginalObservationContext(t)
	defer cancel()
	profile, err := Capture(ctx, request.GoBinary, t.TempDir(), request.Environment, request.Target, nil, RunBounded)
	if err != nil {
		t.Fatal(err)
	}
	macros, err := CaptureAssemblerMacros(runtime.GOROOT(), profile, "fixture/sdk")
	if err != nil {
		t.Fatal(err)
	}
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for source := range macros.ToolSourceSHA256 {
		data, err := os.ReadFile(filepath.Join(runtime.GOROOT(), filepath.FromSlash(source)))
		if err != nil {
			t.Fatal(err)
		}
		writeSDKIncludeFixture(t, root, source, string(data))
	}
	writeSDKIncludeFixture(t, root, "src/fixture/sdk/value_amd64.s", assembly)
	writeSDKIncludeFixture(t, root, "src/fixture/sibling/value.h", "#define VALUE 7\n")
	writeSDKIncludeFixture(t, root, "pkg/include/shared.h", "#define SHARED 11\n")
	pkg := amd64OriginalListPackage{Standard: true}
	pkg.ImportPath, pkg.Dir, pkg.SFiles = "fixture/sdk", filepath.Join(root, "src", "fixture", "sdk"), []string{"value_amd64.s"}
	files := newAMD64OriginalObservationFiles()
	if err := files.directory(pkg.Dir, true); err != nil {
		t.Fatal(err)
	}
	return pkg, root, profile, files
}

func writeSDKIncludeFixture(t *testing.T, root, name, source string) {
	t.Helper()
	file := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(file), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
}

func TestAMD64OriginalSDKIncludeSiblingSourceClosure(t *testing.T) {
	assembly := "#include \"../sibling/value.h\"\nTEXT ·Value(SB),4,$0-0\nMOVQ $VALUE,AX\nRET\n"
	pkg, root, profile, files := sdkDependencyIncludeFixture(t, assembly)
	request := amd64OriginalObservationFixture(t, "bodyful")
	ctx, cancel := amd64OriginalObservationContext(t)
	defer cancel()
	object := filepath.Join(t.TempDir(), "original.o")
	output, diagnostic, err := RunBounded(ctx, pkg.Dir, request.Environment, request.GoBinary,
		"tool", "asm", "-p", pkg.ImportPath, "-o", object, "value_amd64.s")
	if err != nil {
		t.Fatalf("original Go sibling include: %v\n%s\n%s", err, output, diagnostic)
	}
	if info, err := os.Stat(object); err != nil || info.Size() == 0 {
		t.Fatalf("original Go did not produce the sibling-include object: %v", err)
	}
	if err := captureAMD64OriginalSDKDependency(pkg, root, profile, files); err != nil {
		t.Fatal(err)
	}
	header := filepath.Join(root, "src", "fixture", "sibling", "value.h")
	if files.hashes[header] == "" {
		t.Fatal("actual original Go include outside package/shared roots has no byte guard")
	}
	if len(files.cpp) != 0 {
		t.Fatal("dependency include retention minted ordinary CPP authority")
	}
	writeSDKIncludeFixture(t, root, "src/fixture/sibling/value.h", "#define VALUE 9\n")
	if err := files.verify(); err == nil {
		t.Fatal("changed SDK sibling include escaped original source verification")
	}
}
