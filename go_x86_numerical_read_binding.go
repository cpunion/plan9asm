package plan9asm

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"go/format"
	"go/parser"
	"go/token"
	"path/filepath"
	"reflect"
	"strings"
)

// GoX86NumericalReadBinding is a live factory token, not a serializable report.
// Its capture context and every source/storage owner must remain live until
// translation completes. Failure/Close revokes the originating source.
type GoX86NumericalReadBinding struct {
	source         *GoX86NumericalReadSource
	originalSource *GoX86NumericalReadSource
	context        context.Context
	packageInput   GoPackage
	file           [32]byte
	options        Options
	symbols        map[string]string
	declarations   map[string]goAMD64FrameDeclaration
	reads          []numericalSBReadFact
	seal           [32]byte
}

// BindNumericalReads joins complete caller declarations/options to the sealed
// original source and actual native reads. It returns independent Options;
// declarations alone, JSON, pointer types and a numerical owner cannot bind.
func (source *GoX86NumericalReadSource) BindNumericalReads(ctx context.Context, pkg GoPackage, file *File, options Options) (Options, error) {
	if err := source.Verify(ctx); err != nil {
		return Options{}, err
	}
	if file == nil || options.ResolveSym == nil || pkg.Types == nil || pkg.Path != source.request.PackagePath || pkg.Types.Path() != pkg.Path {
		return Options{}, numericalSBContext("binding lacks complete original declarations/source/resolver")
	}
	arch := strings.TrimPrefix(source.request.Target, "linux/")
	triple := map[string]string{"386": "i386-unknown-linux-gnu", "amd64": "x86_64-unknown-linux-gnu"}[arch]
	if options.Goarch != arch || options.TargetTriple != triple || options.WASMABI != WASMABIDirect || len(options.X86TailGroups) != 0 || options.NumericalReads != nil {
		return Options{}, numericalSBContext("binding profile or existing authority differs")
	}
	provided, err := x86CallerSourceFingerprint(file)
	if err != nil {
		return Options{}, err
	}
	actual, err := x86CallerSourceFingerprint(source.file)
	if err != nil || provided != actual || len(file.Funcs) != len(options.Sigs) {
		return Options{}, numericalSBContext("binding requires the exact complete original parsed file")
	}
	// Resolve outside observer locks, then verify the complete live source again.
	symbols := make(map[string]string)
	for _, fn := range file.Funcs {
		symbols[fn.Sym] = options.ResolveSym(fn.Sym)
		if symbols[fn.Sym] != numericalNativeName(strings.TrimSuffix(fn.Sym, "(SB)"), pkg.Path) {
			return Options{}, numericalSBContext("function resolver differs from actual ABI0 symbol")
		}
	}
	for _, read := range source.reads {
		base, _, valid := parseSBRef(read.original.Args[read.operand].Sym)
		if !valid || options.ResolveSym(base) != read.variable.NativeName {
			return Options{}, numericalSBContext("storage resolver differs from actual typed data symbol")
		}
		symbols[base] = read.variable.NativeName
		if _, collision := options.Sigs[read.variable.NativeName]; collision {
			return Options{}, numericalSBContext("typed data symbol collides with a function")
		}
	}
	if err := source.Verify(ctx); err != nil {
		return Options{}, err
	}
	source.mu.Lock()
	declarations, err := source.guard.typedDeclarationsForArch(pkg, false, arch)
	if err == nil {
		err = numericalDeclarationSyntax(source, pkg)
	}
	source.mu.Unlock()
	if err != nil {
		return Options{}, err
	}
	sigs, err := numericalCloneSignatures(options.Sigs)
	if err != nil {
		return Options{}, err
	}
	for _, fn := range file.Funcs {
		name := symbols[fn.Sym]
		declaration, found := declarations[name]
		if !found || !reflect.DeepEqual(numericalSignatureShape(sigs[name]), declaration.signature) {
			return Options{}, numericalSBContext("signature is not the actual source declaration: %s", name)
		}
	}
	options.Sigs, options.ResolveSym = sigs, nil
	binding := &GoX86NumericalReadBinding{source: source, originalSource: source, context: ctx, packageInput: pkg, file: actual,
		options: options, symbols: symbols, declarations: declarations, reads: source.readFacts()}
	binding.seal, err = binding.digest()
	if err != nil {
		return Options{}, err
	}
	if err := binding.verify(); err != nil {
		return Options{}, err
	}
	bound := options
	bound.Sigs, err = numericalCloneSignatures(options.Sigs)
	if err != nil {
		return Options{}, err
	}
	bound.ResolveSym = binding.resolve
	bound.NumericalReads = binding
	return bound, nil
}

func numericalSignatureShape(sig FuncSig) FuncSig {
	sig.x86CallerRegions, sig.x86FixedLocal = nil, nil
	sig.arm64FPReadDeclaration, sig.arm64ABI0Declaration, sig.arm64ABI0TailDeclaration, sig.arm64NativeScratch = nil, nil, nil, nil
	return sig
}

func numericalCloneSignatures(input map[string]FuncSig) (map[string]FuncSig, error) {
	result, err := cloneFixedLocalSigs(input)
	if err != nil {
		return nil, err
	}
	// Preserve separately established private contracts, not reconstructed JSON.
	for name, old := range input {
		sig := result[name]
		sig.x86CallerRegions, sig.x86FixedLocal = old.x86CallerRegions, old.x86FixedLocal
		sig.arm64FPReadDeclaration, sig.arm64ABI0Declaration, sig.arm64ABI0TailDeclaration, sig.arm64NativeScratch = old.arm64FPReadDeclaration, old.arm64ABI0Declaration, old.arm64ABI0TailDeclaration, old.arm64NativeScratch
		result[name] = sig
	}
	return result, nil
}

func numericalDeclarationSyntax(source *GoX86NumericalReadSource, pkg GoPackage) error {
	if len(pkg.Syntax) != len(source.guard.selection.GoFiles) {
		return numericalSBContext("declaration syntax inventory differs")
	}
	for i, name := range source.guard.selection.GoFiles {
		data, err := source.guard.read(filepath.Join(source.request.Directory, name))
		if err != nil {
			return err
		}
		actual, err := parser.ParseFile(token.NewFileSet(), name, data, parser.ParseComments)
		if err != nil || pkg.Syntax[i] == nil {
			return numericalSBContext("declaration syntax is missing or unparseable")
		}
		var a, b bytes.Buffer
		if format.Node(&a, token.NewFileSet(), actual) != nil || format.Node(&b, token.NewFileSet(), pkg.Syntax[i]) != nil || !bytes.Equal(a.Bytes(), b.Bytes()) {
			return numericalSBContext("supplied complete declaration syntax differs from original source")
		}
	}
	return nil
}

func (binding *GoX86NumericalReadBinding) resolve(symbol string) string {
	if name, found := binding.symbols[symbol]; found {
		return name
	}
	return symbol // No address/call role is created for unknown symbols.
}

func (binding *GoX86NumericalReadBinding) digest() ([32]byte, error) {
	data, err := json.Marshal(struct {
		File         [32]byte
		Target, Arch string
		WASM         WASMABI
		X87          X87Mode
		Annotate     bool
		Sigs         map[string]FuncSig
		Symbols      map[string]string
		Reads        []numericalSBReadFact
	}{binding.file, binding.options.TargetTriple, binding.options.Goarch, binding.options.WASMABI, binding.options.X87Mode,
		binding.options.AnnotateSource, binding.options.Sigs, binding.symbols, binding.reads})
	return sha256.Sum256(data), err
}

func (binding *GoX86NumericalReadBinding) verify() error {
	if binding == nil || binding.source == nil || binding.source != binding.originalSource {
		return numericalSBContext("missing live factory binding")
	}
	fail := func(err error) error {
		binding.source.mu.Lock()
		binding.source.ready = false
		binding.source.mu.Unlock()
		return err
	}
	if err := binding.source.Verify(binding.context); err != nil {
		return err
	}
	sum, err := binding.digest()
	if err != nil || sum != binding.seal {
		return fail(numericalSBContext("private binding shape changed"))
	}
	binding.source.mu.Lock()
	declarations, err := binding.source.guard.typedDeclarationsForArch(binding.packageInput, false, binding.options.Goarch)
	if err == nil {
		err = numericalDeclarationSyntax(binding.source, binding.packageInput)
	}
	valid := binding.source.ready && reflect.DeepEqual(binding.declarations, declarations) && reflect.DeepEqual(binding.reads, binding.source.readFacts())
	binding.source.mu.Unlock()
	if err != nil {
		return fail(err)
	}
	if !valid {
		return fail(numericalSBContext("declaration/source/owner read join changed"))
	}
	return binding.source.Verify(binding.context)
}
