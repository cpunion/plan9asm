package plan9asm

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/xgo-dev/plan9asm/internal/gostaticstorage"
	"github.com/xgo-dev/plan9asm/internal/gotoolchain"
	"github.com/xgo-dev/plan9asm/internal/sourcepath"
)

func (source *GoX86SDKASMDataSource) verifySDKDirectory() error {
	root := source.guard.env["GOROOT"]
	wanted := filepath.Join(root, "src", filepath.FromSlash(source.request.PackagePath))
	resolved, err := sourcepath.Resolve(wanted)
	actual, actualErr := os.Lstat(source.request.Directory)
	physical, physicalErr := os.Lstat(wanted)
	if err != nil || resolved != wanted || actualErr != nil || physicalErr != nil ||
		!actual.IsDir() || !physical.IsDir() || !os.SameFile(actual, physical) || source.request.Directory != wanted {
		return sdkASMDataContext("source is not the invoked SDK's exact physical package")
	}
	return nil
}

func (source *GoX86SDKASMDataSource) capture(ctx context.Context, provided map[string]string) error {
	source.observation.Protocol = "actual_go_sdk_asm_data_live_observation_v1"
	source.observation.Request = source.request
	if _, err := source.guard.read(source.request.GoBinary); err != nil {
		return err
	}
	raw, err := source.run(ctx, "env", "-json", "GOROOT", "GOTOOLDIR", "GOVERSION", "GOOS", "GOARCH",
		"GOAMD64", "GO386", "GOFLAGS", "GOEXPERIMENT", "GOCACHE", "GOMODCACHE", "CGO_ENABLED")
	if err != nil {
		return err
	}
	var env map[string]string
	if json.Unmarshal(raw, &env) != nil || !strings.HasPrefix(env["GOVERSION"], "go1.27.") ||
		env["GOOS"]+"/"+env["GOARCH"] != source.request.Target ||
		filepath.Join(env["GOROOT"], "bin", gotoolchain.HostExecutable("go")) != source.request.GoBinary ||
		env["GOAMD64"] != "v1" || env["GOARCH"] == "386" && env["GO386"] != "sse2" ||
		env["GOFLAGS"] != "" || env["GOEXPERIMENT"] != "" ||
		env["GOCACHE"] != provided["GOCACHE"] || env["GOMODCACHE"] != provided["GOMODCACHE"] ||
		env["CGO_ENABLED"] != provided["CGO_ENABLED"] {
		return sdkASMDataContext("actual SDK/compiler/profile is outside this registered observation")
	}
	source.guard.env = env
	source.observation.RawEnvironment = raw
	if err := source.verifySDKDirectory(); err != nil {
		return err
	}
	for _, directory := range []string{env["GOROOT"], env["GOTOOLDIR"], filepath.Join(env["GOROOT"], "pkg", "include")} {
		if err := source.retainDirectory(directory); err != nil {
			return err
		}
	}
	version, err := source.guard.read(filepath.Join(env["GOROOT"], "VERSION"))
	if err != nil || strings.Split(string(version), "\n")[0] != env["GOVERSION"] {
		return sdkASMDataContext("invoked compiler differs from its SDK source VERSION: %v", err)
	}
	for _, tool := range []string{"asm", "compile"} {
		route, err := source.run(ctx, "tool", "-n", tool)
		if err != nil {
			return err
		}
		if strings.TrimSpace(string(route)) != filepath.Join(env["GOTOOLDIR"], gotoolchain.HostExecutable(tool)) {
			return sdkASMDataContext("actual %s route differs from the invoked SDK", tool)
		}
		if _, err := source.guard.read(strings.TrimSpace(string(route))); err != nil {
			return err
		}
	}
	for _, role := range []string{"src/cmd/internal/goobj/builtinlist.go", "src/cmd/internal/obj/x86/asm6.go",
		"src/cmd/internal/obj/x86/obj6.go", "src/cmd/go/internal/work/gc.go", "src/cmd/go/internal/load/pkg.go",
		"src/cmd/go/internal/work/init.go", "src/internal/platform/supported.go",
		"src/cmd/compile/internal/gc/obj.go", "src/cmd/link/internal/ld/go.go",
		"src/cmd/asm/main.go", "src/cmd/asm/internal/lex/input.go", "src/cmd/asm/internal/lex/stack.go",
		"src/cmd/asm/internal/lex/tokenizer.go", "src/cmd/compile/internal/typecheck/subr.go"} {
		if _, err := source.guard.read(filepath.Join(env["GOROOT"], filepath.FromSlash(role))); err != nil {
			return err
		}
	}
	raw, err = source.run(ctx, "list", "-json", "-compiled", "-export", "-mod=readonly", ".")
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	var selected goAMD64FrameSelection
	var extra any
	if decoder.Decode(&selected) != nil || decoder.Decode(&extra) != io.EOF ||
		selected.ImportPath != source.request.PackagePath || selected.Dir != source.request.Directory ||
		!selected.Standard || selected.Module != nil || selected.ForTest != "" || selected.nonordinaryRole() ||
		len(selected.Error)+len(selected.DepsErrors) != 0 || len(selected.GoFiles) == 0 ||
		len(selected.SFiles) == 0 || len(selected.SFiles) > 62 {
		return sdkASMDataContext("actual selection is not one complete ordinary SDK assembly package")
	}
	compiled := append([]string(nil), selected.CompiledGoFiles...)
	for index, name := range compiled {
		if filepath.IsAbs(name) {
			if filepath.Dir(name) != selected.Dir {
				return sdkASMDataContext("effective compiler source escaped its original directory")
			}
			compiled[index] = filepath.Base(name)
		}
	}
	if !reflect.DeepEqual(compiled, selected.GoFiles) {
		return sdkASMDataContext("generated compiler source mapping is not original SDK Go")
	}
	export, exportErr := sourcepath.Resolve(selected.Export)
	if exportErr != nil || selected.Export != export || !goAssemblyPathWithinRoot(env["GOCACHE"], export) {
		return sdkASMDataContext("actual selected export is not its canonical compiler cache object")
	}
	source.guard.selection = selected
	source.observation.RawSelection = raw
	graph, err := source.run(ctx, "list", "-deps", "-compiled", "-export", "-json", "-mod=readonly", ".")
	if err != nil {
		return err
	}
	imports, err := source.guard.bindOriginalDeclarationImports(graph)
	if err != nil {
		return err
	}
	source.guard.imports = imports
	source.observation.Graph, source.observation.Imports = graph, imports
	referenced := make(map[string]gostaticstorage.SDKReferencedArchive)
	paths := make([]string, 0, len(imports))
	for path := range imports {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		if path == "unsafe" {
			continue
		}
		dependency := imports[path]
		archive, err := source.guard.read(dependency.Export)
		if err != nil {
			return err
		}
		referenced[path] = gostaticstorage.SDKReferencedArchive{PackagePath: path,
			Sources: append([]string(nil), dependency.SFiles...), Archive: archive}
	}
	archive, err := source.guard.read(export)
	if err != nil {
		return err
	}
	builtin, err := source.guard.read(filepath.Join(env["GOROOT"], "src", "cmd", "internal", "goobj", "builtinlist.go"))
	if err != nil {
		return err
	}
	if err := source.captureHeader(ctx); err != nil {
		return err
	}
	objectInputs := gostaticstorage.SDKASMDataObservationRequest{PackagePath: source.request.PackagePath,
		Target: source.request.Target, Archive: archive, BuiltinSource: builtin, Referenced: referenced}
	for ordinal, name := range selected.SFiles {
		observed, cpp, err := source.captureAssembly(ctx, ordinal, name)
		if err != nil {
			return err
		}
		objectInputs.Sources = append(objectInputs.Sources, observed)
		source.observation.CPP = append(source.observation.CPP, cpp)
	}
	source.observation.ObjectInputs = objectInputs
	objects, err := gostaticstorage.ObserveX86SDKCPPASMData(objectInputs)
	if err != nil {
		return sdkASMDataContext("complete original/expanded/compiler DATA/TEXT join differs: %v", err)
	}
	source.observation.ObjectInputs, source.observation.Objects = objectInputs, objects
	return nil
}
