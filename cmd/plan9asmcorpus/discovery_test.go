package main

import (
	"archive/zip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go/build"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestResolveModuleDownloadUsesCachedZipAfterToolchainVersionRejection(t *testing.T) {
	root := t.TempDir()
	zipPath := filepath.Join(root, "module.zip")
	zipFile, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(zipFile)
	for name, contents := range map[string]string{
		"example.com/lib@v1.0.0/go.mod":      "module example.com/lib\n\ngo 1.27.1\n",
		"example.com/lib@v1.0.0/asm_amd64.s": "TEXT ·f(SB),NOSPLIT,$0-0\nRET\n",
	} {
		entry, err := writer.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := entry.Write([]byte(contents)); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := zipFile.Close(); err != nil {
		t.Fatal(err)
	}
	downloadJSON, err := json.Marshal(moduleDownloadInfo{
		Path:    "example.com/lib",
		Version: "v1.0.0",
		GoMod:   filepath.Join(root, "cache.mod"),
		Zip:     zipPath,
		Error:   "example.com/lib@v1.0.0 requires go >= 1.27.1 (running go 1.27.0; GOTOOLCHAIN=local)",
	})
	if err != nil {
		t.Fatal(err)
	}
	download, err := resolveModuleDownload(downloadJSON, errors.New("exit status 1"), filepath.Join(root, "work"))
	if err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile(filepath.Join(download.Dir, "asm_amd64.s"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(contents), "TEXT ·f") {
		t.Fatalf("extracted assembly = %q", contents)
	}
}

func TestRunDiscoveryAsmDeclUsesCurrentGoTargetABI(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "go.mod"), "module example.com/asmdecl\n\ngo 1.20\n")
	writeTestFile(t, filepath.Join(dir, "decl.go"), "package asmdecl\n\nfunc f(x int64)\n")
	asm := filepath.Join(dir, "decl_amd64.s")
	writeTestFile(t, asm, "TEXT ·f(SB), $0-1\nRET\n")
	env := replaceEnv(os.Environ(), map[string]string{"GOFLAGS": "-mod=mod", "GOWORK": "off"})
	if err := runDiscoveryAsmDecl(context.Background(), dir, env, "linux/amd64", nil, []string{"example.com/asmdecl"}); err == nil {
		t.Fatal("runDiscoveryAsmDecl() succeeded for a wrong TEXT argument size")
	}
	writeTestFile(t, asm, "TEXT ·f(SB), $0-8\nRET\n")
	if err := runDiscoveryAsmDecl(context.Background(), dir, env, "linux/amd64", nil, []string{"example.com/asmdecl"}); err != nil {
		t.Fatalf("runDiscoveryAsmDecl() valid source error = %v", err)
	}
}

func TestRunDiscoveryAsmDeclStillFindsABIMismatchWhenTestsDoNotCompile(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "work")
	dependency := filepath.Join(root, "dependency")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dependency, 0755); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, filepath.Join(dir, "go.mod"), "module plan9asm.local/probe\n\ngo 1.20\n\nrequire example.com/asmdeclbroken v0.0.0\nreplace example.com/asmdeclbroken => ../dependency\n")
	writeTestFile(t, filepath.Join(dependency, "go.mod"), "module example.com/asmdeclbroken\n\ngo 1.20\n")
	writeTestFile(t, filepath.Join(dependency, "decl.go"), "package asmdeclbroken\n\nfunc f() uint32\n")
	writeTestFile(t, filepath.Join(dependency, "decl_386.s"), "TEXT ·f(SB), $0-4\nMOVL AX, ret+4(FP)\nRET\n")
	writeTestFile(t, filepath.Join(dependency, "decl_test.go"), "package asmdeclbroken\n\nvar _ uint32 = int64(1)\n")
	env := replaceEnv(os.Environ(), map[string]string{"GOFLAGS": "-mod=mod", "GOWORK": "off"})
	err := runDiscoveryAsmDecl(context.Background(), dir, env, "linux/386", nil, []string{"example.com/asmdeclbroken"})
	if err == nil || !isDiscoveryAsmDeclABIMismatch(err.Error()) {
		t.Fatalf("runDiscoveryAsmDecl() error = %v, want assembly ABI mismatch despite broken tests", err)
	}
}

func TestRunDiscoveryGoBuildChecksExactCurrentPackage(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "go.mod"), "module example.com/buildable\n\ngo 1.20\n")
	writeTestFile(t, filepath.Join(dir, "decl.go"), "package buildable\n\nfunc f()\n")
	writeTestFile(t, filepath.Join(dir, "decl_amd64.s"), "TEXT ·f(SB), $0-0\nRET\n")
	env := replaceEnv(os.Environ(), map[string]string{"GOFLAGS": "-mod=mod", "GOWORK": "off"})
	if err := runDiscoveryGoBuild(context.Background(), dir, env, "linux/amd64", nil, "example.com/buildable"); err != nil {
		t.Fatalf("runDiscoveryGoBuild() valid package error = %v", err)
	}
	writeTestFile(t, filepath.Join(dir, "decl.go"), "package buildable\n\nvar broken = missingIdentifier\n")
	if err := runDiscoveryGoBuild(context.Background(), dir, env, "linux/amd64", nil, "example.com/buildable"); err == nil {
		t.Fatal("runDiscoveryGoBuild() accepted source rejected by the current Go compiler")
	}
}

func TestDiscoveryGoBuildInfrastructureFailuresAreNotSourceNotApplicable(t *testing.T) {
	for _, diagnostic := range []string{
		"go build example.com/pkg: context deadline exceeded",
		"dial tcp: lookup proxy.golang.org: no such host",
		"Get https://proxy.golang.org: net/http: TLS handshake timeout",
		"write /tmp/go-build/object.o: no space left on device",
		"go build example.com/pkg: signal: killed",
	} {
		if !isDiscoveryGoBuildInfrastructureFailure(diagnostic) {
			t.Fatalf("infrastructure diagnostic was classified as source incompatibility: %q", diagnostic)
		}
	}
	for _, diagnostic := range []string{
		"pkg/file.go:12:2: undefined: removedSymbol",
		"use of internal package runtime/internal/sys not allowed",
		"package example.com/old imports C: build constraints exclude all Go files",
		"pkg/goid_386.s:13: unexpected EOF\nasm: assembly of pkg/goid_386.s failed",
	} {
		if isDiscoveryGoBuildInfrastructureFailure(diagnostic) {
			t.Fatalf("source diagnostic was classified as infrastructure failure: %q", diagnostic)
		}
	}
}

func TestDiscoveryAsmDeclOnlyClassifiesConcreteABIMismatches(t *testing.T) {
	for _, diagnostic := range []string{
		"file.s:1: [386] f: wrong argument size 0; expected $...-12",
		"file.s:2: [386] f: invalid offset ret+4(FP); expected ret+8(FP)",
		"file.s:3: [386] f: invalid MOVL of x+0(FP); uint64 is 8-byte value",
	} {
		if !isDiscoveryAsmDeclABIMismatch(diagnostic) {
			t.Fatalf("concrete ABI diagnostic was not classified: %q", diagnostic)
		}
	}
	for _, diagnostic := range []string{
		"file.s:1: [arm] f: unknown variable unnamed_lo; offset 36 is ret1_lo+36(FP)",
		"file.s:1: [amd64] f+0: function f+0 missing Go declaration",
		"package dependency is missing",
	} {
		if isDiscoveryAsmDeclABIMismatch(diagnostic) {
			t.Fatalf("non-ABI diagnostic was classified as N/A: %q", diagnostic)
		}
	}
}

func TestDiscoveryConstraintTagsReadsOnlyTheSourceHeader(t *testing.T) {
	file := filepath.Join(t.TempDir(), "generated.go")
	contents := "//go:build realtag\n\npackage generated\n\nvar template = `\n//go:build invalid%template\n`\n" + strings.Repeat("x", 128<<10)
	writeTestFile(t, file, contents)
	tags, err := discoveryConstraintTags(file)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"realtag"}; !reflect.DeepEqual(tags, want) {
		t.Fatalf("constraint tags = %#v, want %#v", tags, want)
	}
}

func TestDiscoveryBuildConfigurationsRecordsDirectoryWithoutCurrentGoPackage(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "misnamed.go"), ".LCPI0_0:\n  .byte 1\n")
	writeTestFile(t, filepath.Join(dir, "routine.s"), "TEXT ·routine(SB), $0-0\nRET\n")
	candidate := discoveryCandidate{Module: "example.com/broken", Version: "v1.0.0", AsmFiles: []string{"routine.s"}}
	var evidence []discoverySourceNotApplicableItem
	configs, err := discoveryBuildConfigurationsWithEvidence(candidate, dir, []string{"linux/amd64"}, &evidence)
	if err != nil {
		t.Fatal(err)
	}
	if len(configs) != 0 || len(evidence) != 1 || evidence[0].Kind != discoverySourceNotApplicableNoGoPackage || !strings.Contains(evidence[0].Reason, "misnamed.go") {
		t.Fatalf("configs = %#v, evidence = %#v", configs, evidence)
	}
}

func TestInferUnsuffixedAssemblyTargetsTreatsCurrentGoRejectedSourceAsInapplicable(t *testing.T) {
	file := filepath.Join(t.TempDir(), "fixture.s")
	writeTestFile(t, file, "this is deliberately not current Go assembly\n")
	contexts := []build.Context{build.Default, build.Default}
	contexts[0].GOOS, contexts[0].GOARCH = "linux", "amd64"
	contexts[1].GOOS, contexts[1].GOARCH = "linux", "arm64"
	eligible, restricted, reason, err := inferUnsuffixedAssemblyTargetsDetailed(file, contexts)
	if err != nil {
		t.Fatal(err)
	}
	if !restricted || len(eligible) != 0 {
		t.Fatalf("eligible = %#v, restricted = %v; want explicit empty target set", eligible, restricted)
	}
	if !strings.Contains(reason, "rejected") || !strings.Contains(reason, "linux/amd64") || !strings.Contains(reason, "linux/arm64") {
		t.Fatalf("reason = %q, want rejected targets", reason)
	}
}

func TestInferUnsuffixedAssemblyTargetsUsesEmptyGeneratedHeaderForArchitectureProbe(t *testing.T) {
	file := filepath.Join(t.TempDir(), "x86.s")
	writeTestFile(t, file, "#include \"go_asm.h\"\nTEXT ·x(SB), $0-0\nMOVQ AX, AX\nRET\n")
	contexts := []build.Context{build.Default, build.Default}
	contexts[0].GOOS, contexts[0].GOARCH = "linux", "amd64"
	contexts[1].GOOS, contexts[1].GOARCH = "linux", "arm64"
	eligible, restricted, err := inferUnsuffixedAssemblyTargets(file, contexts)
	if err != nil {
		t.Fatal(err)
	}
	if !restricted || !eligible["linux/amd64"] || eligible["linux/arm64"] || len(eligible) != 1 {
		t.Fatalf("eligible = %#v, restricted = %v; want only linux/amd64", eligible, restricted)
	}
}

func TestMissingGoAsmHeaderRecognizesHostDiagnostics(t *testing.T) {
	for _, message := range []string{
		`fatal error: go_asm.h: No such file or directory`,
		`open go_asm.h: The system cannot find the file specified.`,
		`could not find included file "go_asm.h"`,
	} {
		if !missingGoAsmHeader(message) {
			t.Errorf("missingGoAsmHeader(%q) = false, want true", message)
		}
	}
	if missingGoAsmHeader("missing unrelated.h: no such file or directory") {
		t.Fatal("missingGoAsmHeader accepted an unrelated missing include")
	}
}

func TestLoadDiscoveryCandidatesDeduplicatesAndMergesMatchedRecords(t *testing.T) {
	dir := t.TempDir()
	records := filepath.Join(dir, "records")
	if err := os.Mkdir(records, 0755); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, filepath.Join(records, "01.jsonl"), strings.Join([]string{
		`{"kind":"scanned","module":"example.com/noasm","version":"v1.0.0"}`,
		`{"kind":"matched","module":"example.com/asm","version":"v1.2.3","architectures":["amd64"],"asm_files":["root_amd64.s"]}`,
	}, "\n")+"\n")
	writeTestFile(t, filepath.Join(records, "02.jsonl"), strings.Join([]string{
		`{"kind":"failure","module":"example.com/fail","version":"@latest","error":"404"}`,
		`{"kind":"matched","module":"example.com/asm","version":"v1.2.3","architectures":["unknown","arm64","amd64"],"asm_files":["sub/future_riscv64.s","sub/asm_arm64.s","root_amd64.s"]}`,
	}, "\n")+"\n")
	// A previous per-run layout must not silently expand the active corpus.
	legacy := filepath.Join(dir, "runs", "old", "records")
	if err := os.MkdirAll(legacy, 0755); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, filepath.Join(legacy, "03.jsonl"),
		`{"kind":"matched","module":"example.com/legacy","version":"v9.9.9","asm_files":["legacy_amd64.s"]}`+"\n")

	got, err := loadDiscoveryCandidates(dir)
	if err != nil {
		t.Fatal(err)
	}
	want := []discoveryCandidate{{
		Module:        "example.com/asm",
		Version:       "v1.2.3",
		Architectures: []string{"amd64", "arm64", "unknown"},
		AsmFiles:      []string{"root_amd64.s", "sub/asm_arm64.s", "sub/future_riscv64.s"},
	}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("candidates = %#v, want %#v", got, want)
	}
}

func TestLoadDiscoveryCandidatesRejectsUnsafeAssemblyPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "records.jsonl")
	writeTestFile(t, path, `{"kind":"matched","module":"example.com/asm","version":"v1.0.0","asm_files":["../escape_amd64.s"]}`+"\n")
	if _, err := loadDiscoveryCandidates(path); err == nil || !strings.Contains(err.Error(), "unsafe assembly path") {
		t.Fatalf("loadDiscoveryCandidates() error = %v, want unsafe path failure", err)
	}
}

func TestDiscoveryShardCoversEveryCandidateExactlyOnce(t *testing.T) {
	candidates := make([]discoveryCandidate, 97)
	for i := range candidates {
		candidates[i] = discoveryCandidate{Module: fmt.Sprintf("example.com/module-%03d", i), Version: "v1.0.0"}
	}
	seen := make(map[string]int, len(candidates))
	for shard := 0; shard < 13; shard++ {
		for _, candidate := range selectDiscoveryShard(candidates, shard, 13) {
			seen[candidate.exactKey()]++
		}
	}
	for _, candidate := range candidates {
		if seen[candidate.exactKey()] != 1 {
			t.Fatalf("candidate %s occurred in %d shards, want exactly one", candidate.exactKey(), seen[candidate.exactKey()])
		}
	}
}

func TestFilterDiscoveryCandidatesForTargetsUsesRecordedAssemblyPaths(t *testing.T) {
	candidates := []discoveryCandidate{
		{
			Module:   "example.com/mixed",
			Version:  "v1.0.0",
			AsmFiles: []string{"asm_amd64.s", "asm_linux_riscv64.s", "asm_plan9_riscv64.s", "portable.s"},
		},
		{
			Module:   "example.com/amd64-only",
			Version:  "v1.0.0",
			AsmFiles: []string{"asm_amd64.s"},
		},
	}
	got, err := filterDiscoveryCandidatesForTargets(candidates, []string{"linux/riscv64"})
	if err != nil {
		t.Fatal(err)
	}
	want := []discoveryCandidate{{
		Module:        "example.com/mixed",
		Version:       "v1.0.0",
		Architectures: []string{"riscv64", "unknown"},
		AsmFiles:      []string{"asm_linux_riscv64.s", "portable.s"},
	}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("target candidates = %#v, want %#v", got, want)
	}
}

func TestFilterDiscoveryCandidatesForTargetsKeepsUnknownSuffixesConservatively(t *testing.T) {
	candidates := []discoveryCandidate{{
		Module:   "example.com/generated",
		Version:  "v1.0.0",
		AsmFiles: []string{"asm_amd64x.s", "asm_linux.s", "asm_windows.s"},
	}}
	got, err := filterDiscoveryCandidatesForTargets(candidates, []string{"linux/riscv64"})
	if err != nil {
		t.Fatal(err)
	}
	wantFiles := []string{"asm_amd64x.s", "asm_linux.s"}
	if len(got) != 1 || !reflect.DeepEqual(got[0].AsmFiles, wantFiles) {
		t.Fatalf("target candidates = %#v, want files %#v", got, wantFiles)
	}
}

func TestDiscoveryPackagePatternsUseOnlyAssemblyDirectories(t *testing.T) {
	candidate := discoveryCandidate{
		Module:   "example.com/root",
		Version:  "v1.0.0",
		AsmFiles: []string{"root_amd64.s", "internal/a_amd64.s", "internal/sub/b_arm64.s"},
	}
	want := []string{"example.com/root", "example.com/root/internal", "example.com/root/internal/sub"}
	if got := discoveryPackagePatterns(candidate); !reflect.DeepEqual(got, want) {
		t.Fatalf("patterns = %#v, want %#v", got, want)
	}
}

func TestDiscoveryExecutionPlanUsesDeclaredModulePathAndExactForkContent(t *testing.T) {
	candidate := discoveryCandidate{
		Module:   "example.com/fork",
		Version:  "v1.2.3",
		AsmFiles: []string{"crypto/hash_amd64.s"},
	}
	plan, err := makeDiscoveryExecutionPlan(candidate, "example.com/upstream")
	if err != nil {
		t.Fatal(err)
	}
	if plan.ModulePath != "example.com/upstream" {
		t.Fatalf("module path = %q, want declared path", plan.ModulePath)
	}
	if want := []string{"example.com/upstream/crypto"}; !reflect.DeepEqual(plan.Patterns, want) {
		t.Fatalf("patterns = %#v, want %#v", plan.Patterns, want)
	}
	for _, want := range []string{
		"require example.com/upstream v1.2.3",
		"replace example.com/upstream => example.com/fork v1.2.3",
	} {
		if !strings.Contains(plan.GoMod, want) {
			t.Fatalf("go.mod missing %q:\n%s", want, plan.GoMod)
		}
	}
}

func TestParseDeclaredModulePath(t *testing.T) {
	for _, test := range []struct {
		contents string
		want     string
	}{
		{contents: "module example.com/plain\n", want: "example.com/plain"},
		{contents: "// generated\nmodule \"example.com/quoted\" // comment\n", want: "example.com/quoted"},
	} {
		got, err := parseDeclaredModulePath([]byte(test.contents))
		if err != nil {
			t.Fatal(err)
		}
		if got != test.want {
			t.Fatalf("declared module = %q, want %q", got, test.want)
		}
	}
}

func TestApplicableDiscoveryAssemblyFilesUsesExactTargetBuildRules(t *testing.T) {
	dir := t.TempDir()
	write := func(name, contents string) {
		t.Helper()
		fullPath := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatal(err)
		}
		writeTestFile(t, fullPath, contents)
	}
	write("pkg.go", "package pkg\n")
	write("asm_linux_amd64.s", "TEXT ·amd64(SB),0,$0-0\nRET\n")
	write("asm_freebsd_amd64.s", "TEXT ·freebsd(SB),0,$0-0\nRET\n")
	write("tagged.s", "//go:build linux && arm64\n\nTEXT ·arm64(SB),0,$0-0\nRET\n")
	write("_hidden/hidden.go", "package hidden\n")
	write("_hidden/hidden_amd64.s", "TEXT ·hidden(SB),0,$0-0\nRET\n")
	write("nested/go.mod", "module example.com/nested\n\ngo 1.27\n")
	write("nested/nested.go", "package nested\n")
	write("nested/nested_amd64.s", "TEXT ·nested(SB),0,$0-0\nRET\n")

	candidate := discoveryCandidate{AsmFiles: []string{
		"_hidden/hidden_amd64.s",
		"asm_freebsd_amd64.s",
		"asm_linux_amd64.s",
		"nested/nested_amd64.s",
		"tagged.s",
	}}
	got, err := applicableDiscoveryAssemblyFiles(candidate, dir, []string{"linux/amd64", "linux/arm64"})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"asm_linux_amd64.s", "tagged.s"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("applicable assembly = %#v, want %#v", got, want)
	}
}

func TestDiscoveryBuildConfigurationsCoverMutuallyExclusiveCustomTags(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "pkg.go"), "package pkg\n")
	writeTestFile(t, filepath.Join(dir, "default.s"), "//go:build !special\n\nTEXT ·defaultImpl(SB),0,$0-0\nRET\n")
	writeTestFile(t, filepath.Join(dir, "special.s"), "//go:build special\n\nTEXT ·specialImpl(SB),0,$0-0\nRET\n")
	candidate := discoveryCandidate{
		Module:   "example.com/customtags",
		Version:  "v1.0.0",
		AsmFiles: []string{"default.s", "special.s"},
	}
	got, err := discoveryBuildConfigurations(candidate, dir, []string{"linux/amd64"})
	if err != nil {
		t.Fatal(err)
	}
	want := []discoveryBuildConfiguration{
		{Targets: []string{"linux/amd64"}, AsmFiles: []string{"default.s"}},
		{BuildTags: []string{"special"}, Targets: []string{"linux/amd64"}, AsmFiles: []string{"special.s"}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("build configurations = %#v, want %#v", got, want)
	}
}

func TestDiscoveryBuildConfigurationsCoverTargetSpecificTagAssignments(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "pkg.go"), "package pkg\n")
	writeTestFile(t, filepath.Join(dir, "hybrid.s"), "//go:build (amd64 && avx) || (arm64 && neon)\n\nTEXT ·hybrid(SB),0,$0-0\nRET\n")
	candidate := discoveryCandidate{
		Module:   "example.com/targettags",
		Version:  "v1.0.0",
		AsmFiles: []string{"hybrid.s"},
	}
	got, err := discoveryBuildConfigurations(candidate, dir, []string{"linux/amd64", "linux/arm64"})
	if err != nil {
		t.Fatal(err)
	}
	want := []discoveryBuildConfiguration{
		{BuildTags: []string{"avx"}, Targets: []string{"linux/amd64"}, AsmFiles: []string{"hybrid.s"}},
		{BuildTags: []string{"neon"}, Targets: []string{"linux/arm64"}, AsmFiles: []string{"hybrid.s"}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("build configurations = %#v, want %#v", got, want)
	}
}

func TestDiscoveryBuildConfigurationsDoNotEnableIgnoredGeneratorSources(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "pkg.go"), "package pkg\n")
	writeTestFile(t, filepath.Join(dir, "generated.s"), "//go:build ignore\n\nTEXT ·generated(SB),0,$0-0\nRET\n")
	candidate := discoveryCandidate{
		Module:   "example.com/generator",
		Version:  "v1.0.0",
		AsmFiles: []string{"generated.s"},
	}
	got, err := discoveryBuildConfigurations(candidate, dir, []string{"linux/amd64"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("build configurations = %#v, want ignored generator assembly to remain inapplicable", got)
	}
}

func TestDiscoveryBuildConfigurationsInferUnsuffixedAssemblyFromGoAssembler(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "pkg.go"), "package pkg\n")
	writeTestFile(t, filepath.Join(dir, "goid.s"), "TEXT ·goid(SB),0,$0-0\nMOVQ AX, BX\nLEAQ (BX), CX\nRET\n")
	candidate := discoveryCandidate{
		Module:   "example.com/goid",
		Version:  "v1.0.0",
		AsmFiles: []string{"goid.s"},
	}
	got, err := discoveryBuildConfigurations(candidate, dir, []string{"linux/386", "linux/amd64", "linux/arm64", "js/wasm"})
	if err != nil {
		t.Fatal(err)
	}
	want := []discoveryBuildConfiguration{{Targets: []string{"linux/amd64"}, AsmFiles: []string{"goid.s"}}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("build configurations = %#v, want %#v", got, want)
	}
}

func TestDiscoveryBuildConfigurationsKeepTrulyCrossArchUnsuffixedAssembly(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "pkg.go"), "package pkg\n")
	writeTestFile(t, filepath.Join(dir, "return.s"), "TEXT ·returnOnly(SB),0,$0-0\nRET\n")
	candidate := discoveryCandidate{
		Module:   "example.com/returnonly",
		Version:  "v1.0.0",
		AsmFiles: []string{"return.s"},
	}
	got, err := discoveryBuildConfigurations(candidate, dir, []string{"linux/386", "linux/amd64", "linux/arm", "linux/arm64"})
	if err != nil {
		t.Fatal(err)
	}
	want := []discoveryBuildConfiguration{{Targets: []string{"linux/386", "linux/amd64", "linux/arm", "linux/arm64"}, AsmFiles: []string{"return.s"}}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("build configurations = %#v, want %#v", got, want)
	}
}

func TestValidateDiscoveryReportRejectsZeroAssemblyAndSkippedTargets(t *testing.T) {
	candidate := discoveryCandidate{Module: "example.com/root", Version: "v1.0.0", AsmFiles: []string{"root_amd64.s"}}
	report := matrixReport{
		Targets:      []targetReport{{Goos: "linux", Goarch: "amd64"}},
		TotalTargets: 1,
	}
	err := validateDiscoveryReport([]string{"linux/amd64"}, candidate, report)
	if err == nil || !strings.Contains(err.Error(), "no assembly") {
		t.Fatalf("validateDiscoveryReport() error = %v, want no assembly failure", err)
	}

	report.Targets[0].TotalAsm = 1
	report.Targets[0].Success = 1
	report.TotalAsm = 1
	report.Success = 1
	err = validateDiscoveryReport([]string{"linux/amd64", "windows/amd64"}, candidate, report)
	if err == nil || !strings.Contains(err.Error(), "target coverage changed") {
		t.Fatalf("validateDiscoveryReport() error = %v, want target coverage failure", err)
	}
}

func TestValidateDiscoveryReportRejectsEveryUnobservedApplicableAssemblyFile(t *testing.T) {
	candidate := discoveryCandidate{
		Module:   "example.com/root",
		Version:  "v1.0.0",
		AsmFiles: []string{"root_amd64.s", "asm_s390x.s"},
	}
	report := matrixReport{
		Targets: []targetReport{{
			Goos:     "linux",
			Goarch:   "amd64",
			AsmFiles: []string{"/go/pkg/mod/example.com/root@v1.0.0/different_amd64.s"},
			TotalAsm: 1,
			Success:  1,
		}},
		TotalTargets: 1,
		TotalAsm:     1,
		Success:      1,
	}
	err := validateDiscoveryReport([]string{"linux/amd64"}, candidate, report)
	if err == nil || !strings.Contains(err.Error(), "assembly files were not exercised") {
		t.Fatalf("validateDiscoveryReport() error = %v, want unobserved file failure", err)
	}

	report.Targets[0].AsmFiles = []string{"/go/pkg/mod/example.com/root@v1.0.0/root_amd64.s"}
	if err := validateDiscoveryReport([]string{"linux/amd64"}, candidate, report); err == nil || !strings.Contains(err.Error(), "asm_s390x.s") {
		t.Fatalf("validateDiscoveryReport() error = %v, want every candidate file required", err)
	}

	filtered := candidate
	filtered.AsmFiles = []string{"root_amd64.s"}
	if err := validateDiscoveryReport([]string{"linux/amd64"}, filtered, report); err != nil {
		t.Fatalf("validateDiscoveryReport() with prefiltered candidate error = %v", err)
	}
}

func TestValidateDiscoveryReportAcceptsOnlyEvidenceBackedTargetNotApplicable(t *testing.T) {
	candidate := discoveryCandidate{
		Module:   "example.com/root",
		Version:  "v1.0.0",
		AsmFiles: []string{"cross_arch.s"},
	}
	report := matrixReport{
		Targets: []targetReport{
			{
				Goos:     "linux",
				Goarch:   "amd64",
				AsmFiles: []string{"/go/pkg/mod/example.com/root@v1.0.0/cross_arch.s"},
				TotalAsm: 1,
				Success:  1,
			},
			{
				Goos:          "linux",
				Goarch:        "386",
				AsmFiles:      []string{"/go/pkg/mod/example.com/root@v1.0.0/cross_arch.s"},
				TotalAsm:      1,
				NotApplicable: 1,
				NotApplicableItems: []targetNotApplicableItem{{
					PkgPath:         "example.com/root",
					AsmFile:         "/go/pkg/mod/example.com/root@v1.0.0/cross_arch.s",
					Kind:            targetNotApplicableGoTextArgSize,
					Symbol:          "example.com/root.StructFieldB",
					DeclaredArgSize: 25,
					ExpectedArgSize: 17,
					Reason:          "TEXT argument size is incompatible with the Go declaration on 386",
				}},
			},
		},
		TotalTargets:  2,
		TotalAsm:      2,
		Success:       1,
		NotApplicable: 1,
	}
	if err := validateDiscoveryReport([]string{"linux/amd64", "linux/386"}, candidate, report); err != nil {
		t.Fatalf("validateDiscoveryReport() error = %v", err)
	}
	details := collectMatrixNotApplicableItems(report)
	if len(details) != 1 || details[0].Target != "linux/386" || details[0].Symbol != "example.com/root.StructFieldB" {
		t.Fatalf("not-applicable details = %#v", details)
	}

	report.Targets[1].NotApplicableItems[0].Kind = "unsupported_instruction"
	if err := validateDiscoveryReport([]string{"linux/amd64", "linux/386"}, candidate, report); err == nil || !strings.Contains(err.Error(), "invalid not-applicable evidence") {
		t.Fatalf("validateDiscoveryReport() error = %v, want evidence rejection", err)
	}
}

func TestDiscoveryCorpusReportAccountsForEverySelectedCandidate(t *testing.T) {
	report := discoveryCorpusReport{Selected: 7, Passed: 3, Failed: 2, NotApplicable: 2}
	if err := validateDiscoveryCorpusAccounting(report); err != nil {
		t.Fatal(err)
	}
	report.NotApplicable--
	if err := validateDiscoveryCorpusAccounting(report); err == nil || !strings.Contains(err.Error(), "accounting mismatch") {
		t.Fatalf("validateDiscoveryCorpusAccounting() error = %v, want accounting mismatch", err)
	}
}

func TestDiscoveryConfigurationAsmFilesDeduplicatesTargets(t *testing.T) {
	configs := []discoveryBuildConfiguration{
		{Targets: []string{"linux/amd64"}, AsmFiles: []string{"pkg/a_amd64.s", "pkg/b.s"}},
		{Targets: []string{"windows/amd64"}, AsmFiles: []string{"pkg/a_amd64.s"}},
		{Targets: []string{"linux/386"}, AsmFiles: []string{"pkg/b.s"}},
	}
	want := []string{"pkg/a_amd64.s", "pkg/b.s"}
	if got := discoveryConfigurationAsmFiles(configs); !reflect.DeepEqual(got, want) {
		t.Fatalf("applicable assembly files = %#v, want %#v", got, want)
	}
}

func TestDiscoveryCorpusReportValidatesSourceNotApplicableEvidence(t *testing.T) {
	result := discoveryCorpusResult{
		Module:             "example.com/root",
		Version:            "v1.0.0",
		Status:             discoveryStatusNotApplicable,
		DiscoveredAsmFiles: []string{"pkg/a_amd64.s", "pkg/b.s"},
		SourceNotApplicableItems: []discoverySourceNotApplicableItem{
			{AsmFile: "pkg/a_amd64.s", Targets: []string{"linux/amd64"}, Kind: discoverySourceNotApplicableGoAssembler, Reason: "current Go assembler rejected the source"},
			{AsmFiles: []string{"pkg/b.s"}, Targets: []string{"windows/amd64"}, Kind: discoverySourceNotApplicableGoBuild, Reason: "current Go compiler rejected the exact package"},
		},
	}
	report := discoveryCorpusReport{Selected: 1, NotApplicable: 1, Results: []discoveryCorpusResult{result}}
	if err := validateDiscoveryCorpusAccounting(report); err != nil {
		t.Fatalf("valid source N/A evidence error = %v", err)
	}

	invalid := report
	invalid.Results = append([]discoveryCorpusResult(nil), report.Results...)
	invalid.Results[0].SourceNotApplicableItems = append([]discoverySourceNotApplicableItem(nil), result.SourceNotApplicableItems...)
	invalid.Results[0].SourceNotApplicableItems[1].AsmFiles = []string{"pkg/not-discovered.s"}
	if err := validateDiscoveryCorpusAccounting(invalid); err == nil || !strings.Contains(err.Error(), "invalid source not-applicable evidence") {
		t.Fatalf("validateDiscoveryCorpusAccounting() error = %v, want invalid source evidence", err)
	}

	invalid.Results[0].SourceNotApplicableItems[1] = discoverySourceNotApplicableItem{
		AsmFiles: []string{"pkg/b.s"}, Targets: []string{"linux/amd64"}, Kind: "unsupported_instruction", Reason: "not valid N/A evidence",
	}
	if err := validateDiscoveryCorpusAccounting(invalid); err == nil || !strings.Contains(err.Error(), "invalid source not-applicable evidence") {
		t.Fatalf("validateDiscoveryCorpusAccounting() error = %v, want invalid kind rejection", err)
	}
}

func TestTranslatorInvocationAcceptsExactPatterns(t *testing.T) {
	invocation := makeTranslatorInvocationForTargetsAndTags(
		"/corpus",
		"example.com/root",
		[]string{"example.com/root/internal/...", "example.com/root"},
		[]string{"avx", "sse"},
		[]string{"linux/amd64", "windows/amd64"},
		[]string{"internal/fast.s", "root_amd64.s"},
		"/tmp/out",
		"/repo",
		"/bin/llc-22",
		"/tmp/report.json",
	)
	if !containsString(invocation.Args, "-patterns=example.com/root/internal/...,example.com/root") {
		t.Fatalf("translator args = %#v, want exact package patterns", invocation.Args)
	}
	if containsString(invocation.Args, "-strict-load") {
		t.Fatalf("discovery translator args = %#v, must tolerate unrelated Go package errors", invocation.Args)
	}
	if !containsString(invocation.Args, "-tags=avx,sse") {
		t.Fatalf("discovery translator args = %#v, want explicit custom build tags", invocation.Args)
	}
	if !containsString(invocation.Args, "-targets=linux/amd64,windows/amd64") || containsString(invocation.Args, "-all-targets") {
		t.Fatalf("discovery translator args = %#v, want only applicable targets", invocation.Args)
	}
	if !containsString(invocation.Args, "-asm-files=internal/fast.s,root_amd64.s") {
		t.Fatalf("discovery translator args = %#v, want exact assembly allowlist", invocation.Args)
	}
}

func writeTestFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0644); err != nil {
		t.Fatal(err)
	}
}
