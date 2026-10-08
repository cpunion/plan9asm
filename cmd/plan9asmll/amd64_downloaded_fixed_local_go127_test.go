//go:build go1.27

package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/xgo-dev/plan9asm"
	"github.com/xgo-dev/plan9asm/internal/gotoolprofile"
	"golang.org/x/tools/go/packages"
)

// The generated file-proxy module is an owned mechanism fixture, not an
// authenticated external-library pass. Its original cache source stays intact.
func amd64DownloadedFixedLocalCLIFixture(t *testing.T) (*featureInput, string, compileConfig) {
	t.Helper()
	_, assembly, config := amd64FixedLocalCLIFixture(t, "PADDL 16(SP),X0")
	ctx := config.Context
	input := &featureInput{
		Protocol: featureInputProtocol,
		Module:   "example.invalid/fixedlocal", Version: "v1.0.0",
		SourceRoot: filepath.Dir(assembly),
		Sources:    make(map[string]string), Headers: make(map[string]string),
		Directories: make(map[string][]string), AsmFiles: []string{"value_amd64.s"},
	}
	for _, name := range []string{"go.mod", "value.go", "value_amd64.s"} {
		data, err := os.ReadFile(filepath.Join(input.SourceRoot, name))
		if err != nil {
			t.Fatal(err)
		}
		input.Sources[name] = featureBytesSHA256(data)
		input.Directories["."] = append(input.Directories["."], name)
		if name != "go.mod" {
			input.Headers[name] = string(data)
		}
	}
	sort.Strings(input.Directories["."])
	modeInput := &gotoolprofile.InstrumentationConsumerInput{Original: input}
	invocation, env := instrumentationVersionedFixture(t, ctx, modeInput)
	input = modeInput.Original
	env = featureEnvironment(env, map[string]string{
		"GOOS": "linux", "GOARCH": "amd64", "GO386": "", "GOARM": "", "GOARM64": "", "GOWASM": "",
	})
	for _, value := range env {
		key, value, valid := strings.Cut(value, "=")
		if !valid {
			t.Fatal("malformed exact-download fixture environment")
		}
		switch key {
		case "GOPROXY", "GOSUMDB", "GOPATH", "GOMODCACHE", "GOOS", "GOARCH", "GO386", "GOARM", "GOARM64", "GOWASM":
			t.Setenv(key, value)
		}
	}
	binary, err := exec.LookPath("go")
	if err != nil {
		t.Fatal(err)
	}
	binary, err = filepath.Abs(binary)
	if err != nil {
		t.Fatal(err)
	}
	root, stderr, err := gotoolprofile.RunBounded(ctx, invocation, env, binary, "env", "GOROOT")
	if err != nil || len(stderr) != 0 {
		t.Fatalf("actual exact-download Go source root: %v; %s", err, stderr)
	}
	textflag, err := gotoolprofile.FileSHA256(filepath.Join(strings.TrimSpace(string(root)), "pkg", "include", "textflag.h"))
	if err != nil {
		t.Fatal(err)
	}
	input.ToolSources = map[string]string{"pkg/include/textflag.h": textflag}
	input.Observed, err = gotoolprofile.Capture(ctx, binary, t.TempDir(), env,
		"linux/amd64", nil, gotoolprofile.RunBounded)
	if err != nil {
		t.Fatal(err)
	}
	input.ID = gotoolprofile.ProfileID(input.Observed)
	stdout, stderr, err := gotoolprofile.RunBounded(ctx, invocation, env, binary,
		"list", "-export", "-compiled", "-json", "-mod=readonly", input.Module)
	if err != nil || len(stderr) != 0 {
		t.Fatalf("actual exact-cache original Go export failed: %v\n%s%s", err, stdout, stderr)
	}
	t.Logf("ACTUAL_EXACT_CACHE_ORIGINAL_GO_EXPORT\n%s", stdout)
	for name, expected := range input.Sources {
		actual, err := gotoolprofile.FileSHA256(filepath.Join(input.SourceRoot, name))
		if err != nil || actual != expected {
			t.Fatalf("original cache source changed before CLI: %s: %v", name, err)
		}
	}
	return input, invocation, config
}

func TestAMD64DownloadedFixedLocalActualProfileCLI(t *testing.T) {
	input, invocation, config := amd64DownloadedFixedLocalCLIFixture(t)
	t.Chdir(invocation)
	config.FeaturePath = writeFeatureInput(t, input)
	output := t.TempDir()
	report, _, err := runOneTarget(targetSpec{Goos: "linux", Goarch: "amd64"},
		[]string{input.Module}, nil, input.AsmFiles, input.Module, output,
		false, 0, true, false, true, "", config)
	if err != nil || report.Success != 1 || report.Failed != 0 {
		t.Fatalf("actual original Go object compiled, but exact downloaded profile CLI has no fixed-local consumer: %v; %#v", err, report)
	}
	if err := gotoolprofile.ValidateSelection(input, report.FeatureSelection, nil, true); err != nil {
		t.Fatal(err)
	}
	if len(report.FeatureSelection.CPP) != 1 || len(report.FeatureSelection.Outputs) != 1 ||
		report.FeatureSelection.Outputs[0].IR == "" || report.FeatureSelection.Outputs[0].Object == "" {
		t.Fatal("exact downloaded CLI omitted its actual CPP/LLVM object consumption")
	}
	for name, expected := range input.Sources {
		actual, err := gotoolprofile.FileSHA256(filepath.Join(input.SourceRoot, name))
		if err != nil || actual != expected {
			t.Fatalf("CLI rewrote original cache source: %s: %v", name, err)
		}
	}
}

func TestAMD64DownloadedFixedLocalCLIRejectsUnboundAndChangedIdentity(t *testing.T) {
	for _, role := range []string{
		"no-explicit-request", "missing-version", "load-identity",
		"version-in-resolver", "source-module-in-resolver", "source-root-in-resolver", "input-in-resolver",
	} {
		t.Run(role, func(t *testing.T) {
			input, invocation, config := amd64DownloadedFixedLocalCLIFixture(t)
			t.Chdir(invocation)
			consumer, err := loadFeatureConsumer(config.Context, writeFeatureInput(t, input), t.TempDir(),
				"linux/amd64", input.Module, nil)
			if err != nil {
				t.Fatal(err)
			}
			pkgs, err := packages.Load(&packages.Config{
				Context: config.Context, Dir: invocation, Env: consumer.Env,
				Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles |
					packages.NeedModule | packages.NeedDeps | packages.NeedImports |
					packages.NeedTypes | packages.NeedTypesSizes,
			}, input.Module)
			if err != nil || len(pkgs) != 1 || len(pkgs[0].Errors) != 0 {
				t.Fatalf("actual exact-cache CLI load: %v; %+v", err, pkgs)
			}
			pkg := pkgs[0]
			if pkg.Module == nil || pkg.Module.Main || pkg.Module.Version != input.Version || pkg.Module.Replace != nil {
				t.Fatal("fixture did not retain its actual exact imported module role")
			}
			if err := consumer.capturePackages(pkgs); err != nil {
				t.Fatal(err)
			}
			config.Feature = consumer
			config.x86CallerSelections = captureX86CallerSelections(pkgs, "linux/amd64", consumer)
			assembly := filepath.Join(pkg.Dir, "value_amd64.s")
			data, err := os.ReadFile(assembly)
			if err != nil {
				t.Fatal(err)
			}
			file, err := plan9asm.Parse(plan9asm.ArchAMD64,
				strings.Replace(string(data), "#include \"textflag.h\"", "#define NOSPLIT 4", 1))
			if err != nil {
				t.Fatal(err)
			}
			calls := 0
			resolve := func(name string) string {
				calls++
				if calls == 1 {
					switch role {
					case "version-in-resolver":
						consumer.Input.Version = "v1.0.1"
					case "source-module-in-resolver":
						consumer.Input.SourceModule = "example.invalid/other"
					case "source-root-in-resolver":
						consumer.Input.SourceRoot = invocation
					case "input-in-resolver":
						copied := *consumer.Input
						consumer.Input = &copied
					}
				}
				return strings.Replace(name, "·", pkg.PkgPath+".", 1)
			}
			// Signature inference must not trigger the mutating binding callback.
			sigs, _, err := sigsForAsmFile(pkg, file,
				func(name string) string { return strings.Replace(name, "·", pkg.PkgPath+".", 1) }, "amd64")
			if err != nil {
				t.Fatal(err)
			}
			switch role {
			case "no-explicit-request":
				config.Feature = nil
			case "missing-version":
				consumer.Input.Version = ""
			case "load-identity":
				pkg.Module.Version = "v1.0.1"
			}
			bound, source, err := prepareAMD64FixedLocalSource(pkg, file, assembly, "linux/amd64",
				"x86_64-unknown-linux-gnu", "amd64", config, resolve, sigs)
			if source != nil {
				if closeErr := source.close(); closeErr != nil {
					t.Error(closeErr)
				}
			}
			if !errors.Is(err, plan9asm.ErrProbeNeedsContext) || source != nil || bound != nil {
				t.Fatalf("unbound/changed %s retained downloaded frame authority: %v", role, err)
			}
			if strings.HasSuffix(role, "-in-resolver") && calls == 0 {
				t.Fatal("identity control failed before its actual binding callback")
			}
		})
	}
}
