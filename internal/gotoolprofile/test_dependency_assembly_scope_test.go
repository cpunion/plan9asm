package gotoolprofile

import (
	"strings"
	"testing"
)

func testDependencyAssemblyScopeFixture() (*ConsumerInput, *TestDependencyProof) {
	const module = "example.invalid/scope"
	const file = "helper/asm.s"
	source := TestSourcePackage{
		Package: PackageProof{
			PackagePath: module + "/helper", ModulePath: module, ModuleVersion: "v1.0.0",
			SourceModule: module, SourceVersion: "v1.0.0", SourceRole: "module",
			GoFiles: []string{"helper/helper.go"}, CompiledGoFiles: []string{"helper/helper.go"}, SFiles: []string{file},
			SourceSHA256: map[string]string{
				"helper/helper.go": bytesSHA256([]byte("package helper\n")), file: bytesSHA256(nil),
			},
		},
		VariantID: module + "/helper", ImportMap: map[string]string{},
	}
	input := &ConsumerInput{
		ID:     strings.Repeat("a", 64),
		Module: module, Version: "v1.0.0", SourceModule: module,
		Observed: &Observation{
			GoVersion: "go1.27.1", Target: "darwin/amd64",
			Environment: map[string]string{"GOOS": "darwin", "GOARCH": "amd64"},
		},
		Sources:             source.Package.SourceSHA256,
		Headers:             map[string]string{"helper/helper.go": "package helper\n", file: ""},
		Directories:         map[string][]string{"helper": {"asm.s", "helper.go"}},
		OrdinaryAsmPackages: []string{"registered"},
	}
	return input, &TestDependencyProof{Packages: []TestSourcePackage{source}}
}

func TestTestDependencyAssemblyScopeDiagnosticBindsActualPair(t *testing.T) {
	input, dependency := testDependencyAssemblyScopeFixture()
	err := validateTestSourceClosure(input, dependency, map[string]bool{}, []string{"chosen"})
	if err == nil {
		t.Fatal("unregistered assembly acquired a source-only translation role")
	}
	for _, detail := range []string{
		"package=example.invalid/scope/helper", "module=example.invalid/scope@v1.0.0",
		"source_root=\"\"", "directory=helper", "selected_asm=[\"helper/asm.s\"]",
		"ordinary_asm_packages=[\"registered\"]",
		"target=darwin/amd64", "profile=" + input.ID, "tags=[\"chosen\"]",
	} {
		if !strings.Contains(err.Error(), detail) {
			t.Errorf("actual assembly-scope diagnostic lacks %s: %v", detail, err)
		}
	}
}
