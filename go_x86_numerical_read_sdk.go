package plan9asm

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"

	"github.com/xgo-dev/plan9asm/internal/gostaticstorage"
)

type numericalSBSourceRole uint8

const (
	numericalSBOwnedSource numericalSBSourceRole = iota + 1
	numericalSBSDKSource
)

// This is input/object observation, not an instruction certificate. It cannot
// create a numerical source by deserialization, Standard, or a matching hash.
type numericalSBSDKObservation struct {
	Graph          []byte
	Archive        []byte
	ArchiveMembers []gostaticstorage.TextMember
	Imports        map[string]goAMD64FrameSelection
	Inputs         map[string][32]byte
	origin         *GoX86NumericalReadSource
	seal           [32]byte
	closed         bool
}

// CaptureGoX86SDKNumericalReadSource is an explicit SDK source entry point.
// SDK source identity is separate from the SDK CPU numerical storage owner;
// neither a Standard flag nor an assembly symbol supplies that identity.
func CaptureGoX86SDKNumericalReadSource(ctx context.Context, request GoX86NumericalReadCaptureRequest) (*GoX86NumericalReadSource, error) {
	return captureGoX86NumericalReadSource(ctx, request, numericalSBSDKSource, false)
}

func (source *GoX86NumericalReadSource) verifySDKDirectory(root string) error {
	path := source.request.PackagePath
	if source.role != numericalSBSDKSource || !filepath.IsLocal(filepath.FromSlash(path)) ||
		filepath.ToSlash(filepath.Clean(filepath.FromSlash(path))) != path {
		return numericalSBContext("SDK source needs an exact canonical package path")
	}
	wanted := filepath.Join(root, "src", filepath.FromSlash(path))
	resolved, err := filepath.EvalSymlinks(wanted)
	actual, actualErr := os.Lstat(source.request.Directory)
	physical, physicalErr := os.Lstat(wanted)
	if err != nil || resolved != wanted || actualErr != nil || physicalErr != nil ||
		!actual.IsDir() || !physical.IsDir() || !os.SameFile(actual, physical) || source.request.Directory != wanted {
		return numericalSBContext("SDK source is not this invoked installation's exact physical package")
	}
	return nil
}

func (source *GoX86NumericalReadSource) captureSourceRole(ctx context.Context) error {
	selected := source.guard.selection
	if source.role == numericalSBOwnedSource {
		if selected.Standard || selected.Module == nil || !selected.Module.Main || selected.Module.Version != "" || selected.Module.Path == "" ||
			len(selected.Module.Replace) != 0 && string(selected.Module.Replace) != "null" ||
			selected.Module.GoMod != filepath.Join(selected.Module.Dir, "go.mod") ||
			!goAssemblyPathWithinRoot(selected.Module.Dir, source.request.Directory) {
			return numericalSBContext("consumer needs complete original owned-main module/source role")
		}
		_, err := source.guard.read(selected.Module.GoMod)
		return err
	}
	if source.role != numericalSBSDKSource || !selected.Standard || selected.Module != nil {
		return numericalSBContext("SDK source must independently retain its actual standard/no-module selection")
	}
	if err := source.verifySDKDirectory(source.guard.env["GOROOT"]); err != nil {
		return err
	}
	compiled := append([]string(nil), selected.CompiledGoFiles...)
	for index, name := range compiled {
		if filepath.IsAbs(name) {
			relative, err := filepath.Rel(selected.Dir, name)
			if err != nil || filepath.Base(relative) != relative {
				return numericalSBContext("SDK effective/generated compiler sources need a separate mapping")
			}
			compiled[index] = relative
		}
	}
	if !reflect.DeepEqual(compiled, selected.GoFiles) {
		return numericalSBContext("SDK selection omitted complete ordinary compiler sources")
	}
	matched := 0
	for _, name := range selected.SFiles {
		if name == source.request.AssemblyFile {
			matched++
		}
	}
	if matched != 1 {
		return numericalSBContext("SDK assembly is not uniquely selected by the actual target")
	}
	graph, err := source.run(ctx, "list", "-deps", "-compiled", "-export", "-json", "-mod=readonly", ".")
	if err != nil {
		return err
	}
	imports, err := source.guard.bindStandardDeclarationImports(graph)
	if err != nil {
		return err
	}
	source.guard.imports = imports
	archive, err := source.guard.read(selected.Export)
	if err != nil {
		return err
	}
	builtin, err := source.guard.read(filepath.Join(source.guard.env["GOROOT"], "src", "cmd", "internal", "goobj", "builtinlist.go"))
	if err != nil {
		return err
	}
	members, err := gostaticstorage.ObserveX86AssemblyArchive(archive, builtin, source.request.Target, selected.SFiles)
	if err != nil {
		return numericalSBContext("SDK actual selected compiler/archive members differ: %v", err)
	}
	inputs := make(map[string][32]byte, len(source.guard.inputs))
	for name, digest := range source.guard.inputs {
		inputs[name] = digest
	}
	source.sdk = &numericalSBSDKObservation{Graph: graph, Archive: archive, ArchiveMembers: members,
		Imports: imports, Inputs: inputs, origin: source}
	source.sdk.seal, err = source.sdkDigest()
	if err != nil {
		return err
	}
	return nil
}

func (source *GoX86NumericalReadSource) sdkDigest() ([32]byte, error) {
	raw, err := json.Marshal(struct {
		Request   GoX86NumericalReadCaptureRequest
		Role      numericalSBSourceRole
		Selection goAMD64FrameSelection
		Env       map[string]string
		Entries   []string
		SDK       *numericalSBSDKObservation
	}{source.request, source.role, source.guard.selection, source.guard.env, source.guard.entries, source.sdk})
	return sha256.Sum256(raw), err
}

// The caller owns source.mu during a published Verify. Capture has not
// published this source. This method invokes no caller resolver or callback.
func (source *GoX86NumericalReadSource) verifySDKObservation(ctx context.Context) (result error) {
	defer func() {
		if result != nil && source != nil && source.identity == source && source.sdk != nil {
			source.sdk.closed = true
		}
	}()
	if ctx != nil && ctx.Err() != nil {
		return fmt.Errorf("verify SDK numerical source observation: %w", ctx.Err())
	}
	if ctx == nil || source == nil || source.identity != source || source.role != numericalSBSDKSource ||
		source.sdk == nil || source.sdk.origin != source || source.sdk.closed {
		return numericalSBContext("missing live independent SDK source observation")
	}
	before, err := source.sdkDigest()
	if err != nil || before != source.sdk.seal {
		return numericalSBContext("sealed independent SDK source/object graph changed")
	}
	if len(source.request.Owners) != len(source.originalOwners) {
		return numericalSBContext("SDK source owner inventory changed")
	}
	for ordinal, owner := range source.originalOwners {
		if owner != source.request.Owners[ordinal] {
			return numericalSBContext("SDK source owner identity/order changed")
		}
		if err := owner.Verify(ctx); err != nil {
			return err
		}
	}
	for name, input := range source.sdk.Inputs {
		if source.guard.inputs[name] != input {
			return numericalSBContext("SDK original source/tool/export input identity changed")
		}
	}
	if err := source.verifySDKDirectory(source.guard.env["GOROOT"]); err != nil {
		return err
	}
	if err := source.guard.verifyInputs(); err != nil {
		return err
	}
	graph, err := source.guard.run(ctx, "list", "-deps", "-compiled", "-export", "-json", "-mod=readonly", ".")
	if err != nil {
		return err
	}
	imports, err := source.guard.bindStandardDeclarationImports(graph)
	if err != nil || !reflect.DeepEqual(imports, source.sdk.Imports) || !reflect.DeepEqual(imports, source.guard.imports) ||
		!numericalSameRawSDKGraph(source.sdk.Graph, graph) {
		return numericalSBContext("complete SDK source/export/selection graph changed: %v", err)
	}
	if err := source.guard.verifyInputs(); err != nil {
		return err
	}
	after, err := source.sdkDigest()
	if err != nil || after != source.sdk.seal {
		return numericalSBContext("SDK source/object identity changed during observation")
	}
	return nil
}

func numericalSameRawSDKGraph(first, second []byte) bool {
	decode := func(data []byte) ([]json.RawMessage, bool) {
		if len(data) == 0 || len(data) > 16<<20 {
			return nil, false
		}
		decoder := json.NewDecoder(bytes.NewReader(data))
		var values []json.RawMessage
		for {
			var value json.RawMessage
			if err := decoder.Decode(&value); err == io.EOF {
				return values, len(values) != 0
			} else if err != nil || len(values) >= 4096 {
				return nil, false
			}
			values = append(values, value)
		}
	}
	a, validA := decode(first)
	b, validB := decode(second)
	if !validA || !validB || len(a) != len(b) {
		return false
	}
	for index := range a {
		if !numericalSameRawSelection(a[index], b[index]) {
			return false
		}
	}
	return true
}
