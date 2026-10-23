package plan9asm

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"reflect"
	"strings"

	"github.com/xgo-dev/plan9asm/internal/gosym"
)

// This declaration-only directory describes numerical reads of Go-defined
// variables. It is deliberately separate from emitted DATA, writable objects,
// address values and the SDK's native compiler-record/source capabilities.
type goReadOnlyStorage struct {
	pkg      GoPackage
	sizes    types.Sizes
	file     *File
	original [32]byte
	syntax   []string
	objects  map[string]goReadOnlyObject
	symbols  map[string]string
	resolve  func(string) string
	failure  error
}

type goReadOnlyObject struct {
	variable *types.Var
	native   string
	size     int64
	spans    [][2]int64 // half-open numerical byte intervals; padding is absent
	shape    string
}

type goReadOnlyRead struct {
	operand Operand
	width   int64
}

func (read *goReadOnlyRead) allows(operand Operand, width int64) bool {
	return read != nil && width == read.width && reflect.DeepEqual(operand, read.operand)
}

// Only the complete named comparison grammar has the paired volatile emitter
// here. A MOV/raw/vector operation at the same address cannot borrow this fact.
func (source *goReadOnlyStorage) readFor(ins Instr, arch string) *goReadOnlyRead {
	if source == nil || !x86ResultNamedPredicateInstruction(ins) || ins.x86Encoded ||
		ins.x86SourceRaw != nil || ins.x86AddressBits != 0 || ins.x86UnboundAddressOverride {
		return nil
	}
	form, known, err := x86ScalarCompareProperties(ins.Op, ins, arch)
	if !known || err != nil || form.memory < 0 {
		return nil
	}
	operand, width := form.args[form.memory], int64(form.bits/8)
	if !source.allows(operand, width, arch) {
		return nil
	}
	return &goReadOnlyRead{operand: operand, width: width}
}

func goReadOnlyContext(format string, args ...any) error {
	return fmt.Errorf("%w: ordinary Go global read: %s", ErrProbeNeedsContext, fmt.Sprintf(format, args...))
}

func goReadOnlyName(pkgPath, symbol string) string {
	if strings.HasPrefix(symbol, "·") {
		symbol = strings.TrimPrefix(symbol, "·")
	} else if strings.HasPrefix(symbol, pkgPath+".") {
		symbol = strings.TrimPrefix(symbol, pkgPath+".")
	}
	if !token.IsIdentifier(symbol) || symbol == "_" {
		return ""
	}
	return symbol
}

func goReadOnlySyntax(pkg GoPackage) ([]string, error) {
	if len(pkg.Syntax) == 0 || len(pkg.Syntax) > 256 {
		return nil, goReadOnlyContext("missing or excessive same-package declaration syntax")
	}
	result := make([]string, len(pkg.Syntax))
	for index, file := range pkg.Syntax {
		if file == nil || file.Name == nil || file.Name.Name != pkg.Types.Name() {
			return nil, goReadOnlyContext("declaration package differs")
		}
		var b bytes.Buffer
		if err := format.Node(&b, token.NewFileSet(), file); err != nil || b.Len() > 16<<20 {
			return nil, goReadOnlyContext("declaration syntax is invalid or excessive")
		}
		result[index] = b.String()
	}
	return result, nil
}

type goReadOnlyImporter map[string]*types.Package

func (imports goReadOnlyImporter) Import(path string) (*types.Package, error) {
	if path == "unsafe" {
		return types.Unsafe, nil
	}
	if pkg := imports[path]; pkg != nil && pkg.Path() == path && pkg.Complete() {
		return pkg, nil
	}
	return nil, goReadOnlyContext("missing actual typed import %q", path)
}

// Recheck already supplied syntax and imports, never go list, an ambient
// importer, a compiler archive capture or the execution of an initializer.
func goReadOnlyTypedSyntax(pkg GoPackage, syntax []string, sizes types.Sizes, objects map[string]goReadOnlyObject) (*types.Package, error) {
	set := token.NewFileSet()
	files := make([]*ast.File, len(syntax))
	for index, source := range syntax {
		file, err := parser.ParseFile(set, fmt.Sprintf("global-source-%d.go", index), source, parser.ParseComments)
		if err != nil {
			return nil, goReadOnlyContext("declaration syntax cannot be rechecked")
		}
		for _, comments := range file.Comments {
			for _, comment := range comments.List {
				if strings.HasPrefix(comment.Text, "//line ") {
					return nil, goReadOnlyContext("aliased or source-remapped declarations need their own binding")
				}
				words := strings.Fields(comment.Text)
				if len(words) >= 2 && words[0] == "//go:linkname" {
					if objects[words[1]].variable != nil {
						return nil, goReadOnlyContext("aliased global declaration needs its own binding")
					}
					if len(words) >= 3 {
						for _, object := range objects {
							if words[2] == object.native {
								return nil, goReadOnlyContext("another declaration aliases the requested global")
							}
						}
					}
				}
			}
		}
		files[index] = file
	}
	typed, err := (&types.Config{Importer: goReadOnlyImporter(pkg.Imports), Sizes: sizes}).Check(pkg.Path, set, files, nil)
	if err != nil {
		return nil, goReadOnlyContext("syntax does not match a complete typed package: %v", err)
	}
	return typed, nil
}

func goReadOnlyLayout(typ types.Type, sizes types.Sizes) (size int64, spans [][2]int64, shape string, resultErr error) {
	defer func() {
		if recover() != nil {
			size, spans, shape, resultErr = 0, nil, "", goReadOnlyContext("incomplete target type layout")
		}
	}()
	size = sizes.Sizeof(typ)
	if size <= 0 || size > 1<<20 {
		return 0, nil, "", goReadOnlyContext("global numerical extent is empty or excessive")
	}
	var description strings.Builder
	var visit func(types.Type, int64, int) error
	visit = func(current types.Type, base int64, depth int) error {
		width := sizes.Sizeof(current)
		if depth > 64 || width < 0 || base < 0 || base > size || width > size-base {
			return goReadOnlyContext("global type layout exceeds its numerical extent")
		}
		fmt.Fprintf(&description, "[%s:%d:%d]", types.TypeString(current, func(p *types.Package) string { return p.Path() }), base, width)
		switch underlying := current.Underlying().(type) {
		case *types.Basic:
			if underlying.Info()&(types.IsInteger|types.IsFloat|types.IsBoolean) == 0 || underlying.Info()&types.IsUntyped != 0 {
				return goReadOnlyContext("global contains pointer/header or nonnumerical storage")
			}
			fmt.Fprintf(&description, "basic:%d;", underlying.Kind())
			if width != 0 {
				end := base + width
				if len(spans) != 0 && spans[len(spans)-1][1] == base {
					spans[len(spans)-1][1] = end
				} else {
					spans = append(spans, [2]int64{base, end})
				}
			}
		case *types.Array:
			if underlying.Len() > 1<<20 {
				return goReadOnlyContext("global array exceeds its type bound")
			}
			stride := sizes.Sizeof(underlying.Elem())
			for index := int64(0); index < underlying.Len(); index++ {
				if err := visit(underlying.Elem(), base+index*stride, depth+1); err != nil {
					return err
				}
			}
		case *types.Struct:
			fields := make([]*types.Var, underlying.NumFields())
			for index := range fields {
				fields[index] = underlying.Field(index)
			}
			for index, offset := range sizes.Offsetsof(fields) {
				fmt.Fprintf(&description, "field:%s:%d;", fields[index].Name(), offset)
				if fields[index].Name() != "_" {
					if err := visit(fields[index].Type(), base+offset, depth+1); err != nil {
						return err
					}
				}
			}
		default:
			return goReadOnlyContext("global contains pointer/header or nonnumerical storage")
		}
		return nil
	}
	if err := visit(typ, 0, 0); err != nil {
		return 0, nil, "", err
	}
	return size, spans, description.String(), nil
}

func bindGoReadOnlyStorage(pkg GoPackage, file *File, sizes types.Sizes, arch string, resolve func(string) string) (*goReadOnlyStorage, error) {
	if file.Arch != ArchAMD64 || arch != "amd64" && arch != "386" {
		return nil, nil
	}
	if sizes == nil {
		sizes = types.SizesFor("gc", arch)
	}
	if sizes == nil || pkg.Types == nil || pkg.Types.Scope() == nil || pkg.Types.Path() != pkg.Path {
		return nil, goReadOnlyContext("missing actual package types or target sizes")
	}
	source := &goReadOnlyStorage{pkg: pkg, sizes: sizes, file: x86DirectTailSnapshot(file), resolve: resolve,
		objects: make(map[string]goReadOnlyObject), symbols: make(map[string]string)}
	for _, function := range file.Funcs {
		for _, instruction := range function.Instrs {
			for _, operand := range instruction.Args {
				address, valid := x86StaticSBSourceAddress(operand, arch)
				name := goReadOnlyName(pkg.Path, address.symbol)
				variable, declared := pkg.Types.Scope().Lookup(name).(*types.Var)
				if !valid || !declared || variable.IsField() || variable.Pkg() != pkg.Types {
					continue // Unknown/imported symbols never inherit another variable's type.
				}
				if _, present := source.objects[name]; present {
					continue
				}
				size, spans, shape, err := goReadOnlyLayout(variable.Type(), sizes)
				if err != nil {
					return nil, err
				}
				source.objects[name] = goReadOnlyObject{variable: variable, native: gosym.PackagePrefix(pkg.Path) + "." + name,
					size: size, spans: spans, shape: shape}
			}
		}
	}
	if len(source.objects) == 0 {
		return nil, nil
	}
	if len(source.objects) > 32 {
		return nil, goReadOnlyContext("global declaration inventory exceeds its bound")
	}
	for _, function := range file.Funcs {
		for _, instruction := range function.Instrs {
			for _, operand := range instruction.Args {
				address, valid := x86StaticSBSourceAddress(operand, arch)
				if valid && source.objects[goReadOnlyName(pkg.Path, address.symbol)].variable != nil && source.readFor(instruction, arch) == nil {
					return nil, goReadOnlyContext("global reference lacks its exact bounded volatile comparison read")
				}
			}
		}
	}
	var err error
	source.syntax, err = goReadOnlySyntax(pkg)
	if err != nil {
		return nil, err
	}
	typed, err := goReadOnlyTypedSyntax(pkg, source.syntax, sizes, source.objects)
	if err != nil {
		return nil, err
	}
	for name, object := range source.objects {
		actual, present := typed.Scope().Lookup(name).(*types.Var)
		if !present {
			return nil, goReadOnlyContext("global %q has no actual variable declaration", name)
		}
		size, spans, shape, err := goReadOnlyLayout(actual.Type(), sizes)
		if err != nil || size != object.size || shape != object.shape || !reflect.DeepEqual(spans, object.spans) {
			return nil, goReadOnlyContext("global %q type and syntax layouts differ", name)
		}
	}
	for _, global := range file.Globl {
		if source.objects[goReadOnlyName(pkg.Path, global.Sym)].variable != nil {
			return nil, goReadOnlyContext("Go variable also has an assembly storage definition")
		}
	}
	for _, datum := range file.Data {
		if source.objects[goReadOnlyName(pkg.Path, datum.Sym)].variable != nil {
			return nil, goReadOnlyContext("Go variable also has an assembly DATA definition")
		}
	}
	source.original, err = x86CallerSourceFingerprint(file)
	if err != nil {
		return nil, err
	}
	for _, function := range file.Funcs {
		for _, instruction := range function.Instrs {
			for _, operand := range instruction.Args {
				address, valid := x86StaticSBSourceAddress(operand, arch)
				object := source.objects[goReadOnlyName(pkg.Path, address.symbol)]
				if !valid || object.variable == nil {
					continue
				}
				// Match the exact same-package native symbol, not an arbitrary
				// resolver alias with a conveniently named Go declaration.
				symbol := address.symbol
				if !strings.ContainsAny(symbol, "·./") {
					symbol = "·" + symbol
				}
				source.symbols[symbol] = object.native
			}
		}
	}
	if err := source.verify(); err != nil {
		return nil, err
	}
	return source, nil
}

func (source *goReadOnlyStorage) verify() error {
	if source == nil {
		return nil
	}
	if source.failure != nil {
		return source.failure
	}
	check := func() error {
		actual, err := x86CallerSourceFingerprint(source.file)
		if err != nil || actual != source.original {
			return goReadOnlyContext("original assembly changed during binding")
		}
		syntax, err := goReadOnlySyntax(source.pkg)
		if err != nil || !reflect.DeepEqual(syntax, source.syntax) {
			return goReadOnlyContext("declaration syntax changed during binding")
		}
		for name, object := range source.objects {
			variable, _ := source.pkg.Types.Scope().Lookup(name).(*types.Var)
			if variable != object.variable {
				return goReadOnlyContext("global declaration identity changed during binding")
			}
			size, spans, shape, err := goReadOnlyLayout(variable.Type(), source.sizes)
			if err != nil || size != object.size || shape != object.shape || !reflect.DeepEqual(spans, object.spans) {
				return goReadOnlyContext("global target layout changed during binding")
			}
		}
		return nil
	}
	if err := check(); err != nil {
		return err
	}
	for symbol, expected := range source.symbols {
		if source.resolve(symbol) != expected {
			return goReadOnlyContext("global resolver differs from its same-package declaration")
		}
	}
	return check() // Resolvers are user callbacks and may have changed source.
}

func (source *goReadOnlyStorage) resolveSymbol(symbol string) string {
	actual := source.resolve(symbol)
	if expected, known := source.symbols[symbol]; known && actual != expected {
		source.failure = goReadOnlyContext("global resolver changed during consumption")
	}
	return actual
}

func (source *goReadOnlyStorage) allows(operand Operand, width int64, arch string) bool {
	if source == nil || source.failure != nil || width <= 0 {
		return false
	}
	address, valid := x86StaticSBReadAddress(operand, arch)
	object := source.objects[goReadOnlyName(source.pkg.Path, address.symbol)]
	if !valid || object.variable == nil || width > object.size || address.offset > object.size-width {
		return false
	}
	for _, span := range object.spans {
		if address.offset >= span[0] && address.offset <= span[1]-width {
			return true
		}
	}
	return false
}
