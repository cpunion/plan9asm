package gotoolprofile

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/parser"
	"go/token"
	"go/types"
	"path"
	"sort"
	"strconv"
	"strings"
)

type nativeHeaderUnsupportedDomain struct{ reason string }

func (err *nativeHeaderUnsupportedDomain) Error() string { return err.reason }

// NativeHeaderSourceDomainUnsupported only identifies an unimplemented query
// domain. It cannot turn a source/compiler/registration failure into authority.
func NativeHeaderSourceDomainUnsupported(err error) bool {
	for depth := 0; err != nil && depth < 32; depth++ {
		if _, ok := err.(*nativeHeaderUnsupportedDomain); ok {
			return true
		}
		wrapped, ok := err.(interface{ Unwrap() error })
		if !ok {
			return false
		}
		err = wrapped.Unwrap()
	}
	return false
}

// ReplayNativeGeneratedHeader queries a closed source model. Its inputs must
// independently match actual selected compiler sources and the stock SDK; it
// grants neither source authentication nor ABI/native-memory permission.
func ReplayNativeGeneratedHeader(packagePath, target, language string, sources map[string]string) (map[string]string, error) {
	set, files, word, err := parseNativeHeaderSources(packagePath, target, language, sources)
	if err != nil {
		return nil, err
	}
	sizes := &nativeHeaderGCSizes{word: word, cache: make(map[types.Type]nativeHeaderLayout), active: make(map[types.Type]bool)}
	pkg, err := (&types.Config{GoVersion: language, Sizes: sizes, IgnoreFuncBodies: true}).Check(packagePath, set, files, nil)
	if err != nil || sizes.err != nil {
		return nil, fmt.Errorf("native header original declaration typecheck failed: %v %v", err, sizes.err)
	}
	return nativeGeneratedHeaderDefinitions(pkg, sizes)
}

// Both ordinary and same-mode imported queries use this one native property
// projection. It is not an ABI layout or an allocation permission.
func nativeGeneratedHeaderDefinitions(pkg *types.Package, sizes *nativeHeaderGCSizes) (map[string]string, error) {
	definitions := make(map[string]string)
	add := func(name, value string) error {
		if _, duplicate := definitions[name]; duplicate {
			return fmt.Errorf("native header generated-name collision: %s", name)
		}
		definitions[name] = value
		return nil
	}
	for _, name := range pkg.Scope().Names() {
		if name == "_" {
			continue
		}
		switch object := pkg.Scope().Lookup(name).(type) {
		case *types.Const:
			value := object.Val()
			if value == nil || value.Kind() == constant.Unknown {
				return nil, fmt.Errorf("native header source has an unresolved constant")
			}
			if value.Kind() != constant.Float && value.Kind() != constant.Complex {
				if err := add("const_"+name, value.ExactString()); err != nil {
					return nil, err
				}
			}
		case *types.TypeName:
			structure, ok := object.Type().Underlying().(*types.Struct)
			if !ok {
				continue
			}
			layout, err := sizes.layout(object.Type(), 0)
			if err != nil {
				return nil, err
			}
			if err := add(name+"__size", strconv.FormatInt(layout.size, 10)); err != nil {
				return nil, err
			}
			for index := 0; index < structure.NumFields(); index++ {
				field := structure.Field(index)
				if field.Name() != "_" {
					if err := add(name+"_"+field.Name(), strconv.FormatInt(layout.offsets[index], 10)); err != nil {
						return nil, err
					}
				}
			}
		}
	}
	if _, err := CanonicalGeneratedHeader(definitions); err != nil {
		return nil, err
	}
	return definitions, nil
}

// PreflightNativeGeneratedHeaderSource rejects already-unclosed query domains
// before an expensive planning compiler query. A nil result is not a source,
// compiler, header or object proof; all final consumers remain mandatory.
func PreflightNativeGeneratedHeaderSource(packagePath, target, language string, sources map[string]string) error {
	_, _, _, err := parseNativeHeaderSources(packagePath, target, language, sources)
	return err
}

func parseNativeHeaderSources(packagePath, target, language string, sources map[string]string) (*token.FileSet, []*ast.File, int64, error) {
	if packagePath == "sync/atomic" || packagePath == "internal/runtime/atomic" || packagePath == "simd/archsimd" {
		return nil, nil, 0, &nativeHeaderUnsupportedDomain{"native header requires compiler-special package layout rules"}
	}
	languageMinor, languageError := goMinor(language)
	if language == "go1" || languageError == nil && languageMinor < 20 {
		return nil, nil, 0, &nativeHeaderUnsupportedDomain{"native property query does not register this historical module language"}
	}
	goos, arch, ok := strings.Cut(target, "/")
	word := int64(8)
	switch arch {
	case "386", "arm":
		word = 4
	case "amd64", "arm64", "wasm":
	default:
		return nil, nil, 0, fmt.Errorf("native header source has an unregistered target")
	}
	if !ok || !CPPLocalSource(goos) || strings.Contains(goos, "/") ||
		packagePath == "" || strings.ContainsAny(packagePath, "\\\r\n\t =") ||
		len(sources) == 0 || len(sources) > 4096 ||
		language != "go1.20" && language != "go1.21" && language != "go1.22" && language != "go1.23" &&
			language != "go1.24" && language != "go1.25" && language != "go1.26" && language != "go1.27" {
		return nil, nil, 0, fmt.Errorf("native header source has an unclosed package/language scope")
	}
	set := token.NewFileSet()
	var names []string
	for name := range sources {
		names = append(names, name)
	}
	sort.Strings(names)
	var files []*ast.File
	total, nodes := 0, 0
	for _, name := range names {
		text := sources[name]
		if !CPPLocalSource(name) || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") ||
			strings.HasPrefix(path.Base(name), ".") || strings.HasPrefix(path.Base(name), "_") ||
			len(text) > 2<<20 || len(text) > 16<<20-total || strings.ContainsRune(text, '\x00') {
			return nil, nil, 0, fmt.Errorf("native header source is absent, nonordinary or unbounded")
		}
		total += len(text)
		file, err := parser.ParseFile(set, name, text, parser.ParseComments)
		if err != nil {
			return nil, nil, 0, fmt.Errorf("native header source is not closed import-free Go: %v", err)
		}
		if len(file.Imports) != 0 {
			return nil, nil, 0, &nativeHeaderUnsupportedDomain{"native header source query has unclosed imports"}
		}
		closed := true
		ast.Inspect(file, func(node ast.Node) bool {
			if node == nil {
				return false
			}
			nodes++
			if spec, ok := node.(*ast.TypeSpec); ok && spec.TypeParams != nil {
				closed = false
			}
			if function, ok := node.(*ast.FuncType); ok && function.TypeParams != nil {
				closed = false
			}
			return nodes <= 100000
		})
		for _, group := range file.Comments {
			for _, comment := range group.List {
				if strings.HasPrefix(comment.Text, "//go:") && !strings.HasPrefix(comment.Text, "//go:build ") {
					closed = false
				}
			}
		}
		if nodes > 100000 {
			return nil, nil, 0, fmt.Errorf("native header source requires unregistered generated/generic/compiler rules")
		}
		if !closed {
			return nil, nil, 0, &nativeHeaderUnsupportedDomain{"native header source requires generated/generic/compiler rules"}
		}
		files = append(files, file)
	}
	if err := boundNativeHeaderStrings(files); err != nil {
		return nil, nil, 0, err
	}
	return set, files, word, nil
}

// This native model deliberately does not use the framework's SizesFor: old
// StdSizes omits gc struct tail padding and padded array strides. It is only
// a closed basic/array/struct property model, not a native allocation contract.
type nativeHeaderLayout struct {
	size, align int64
	offsets     []int64
}

type nativeHeaderGCSizes struct {
	word   int64
	work   int
	err    error
	cache  map[types.Type]nativeHeaderLayout
	active map[types.Type]bool
}

func (sizes *nativeHeaderGCSizes) Alignof(typ types.Type) int64 {
	layout, err := sizes.layout(typ, 0)
	if err != nil {
		sizes.err = err
		return 1
	}
	return layout.align
}

func (sizes *nativeHeaderGCSizes) Sizeof(typ types.Type) int64 {
	layout, err := sizes.layout(typ, 0)
	if err != nil {
		sizes.err = err
		return -1
	}
	return layout.size
}

func (sizes *nativeHeaderGCSizes) Offsetsof(fields []*types.Var) []int64 {
	offsets := make([]int64, len(fields))
	var end int64
	for index, field := range fields {
		layout, err := sizes.layout(field.Type(), 0)
		if err == nil {
			offsets[index], err = sizes.round(end, layout.align)
		}
		if err != nil || layout.size >= sizes.limit()-offsets[index] {
			sizes.err = fmt.Errorf("native header field extent is unclosed or overflows")
			offsets[index] = -1
			return offsets
		}
		end = offsets[index] + layout.size
	}
	return offsets
}

func (sizes *nativeHeaderGCSizes) limit() int64 {
	if sizes.word == 4 {
		// Native calcStructOffset rejects field ends at the signed 32-bit
		// boundary. Use the stricter common bound, including array queries.
		return 1<<31 - 1
	}
	return 1 << 50
}

func (sizes *nativeHeaderGCSizes) round(value, align int64) (int64, error) {
	if value < 0 || align < 1 || align > 8 || value > sizes.limit()-align {
		return 0, fmt.Errorf("native header alignment/extent overflows")
	}
	return (value + align - 1) &^ (align - 1), nil
}

func (sizes *nativeHeaderGCSizes) layout(typ types.Type, depth int) (nativeHeaderLayout, error) {
	if cached, present := sizes.cache[typ]; present {
		return cached, nil
	}
	sizes.work++
	if depth > 128 || sizes.work > 100000 || sizes.active[typ] {
		return nativeHeaderLayout{}, fmt.Errorf("native header layout is recursive or exceeds its query bound")
	}
	sizes.active[typ] = true
	defer delete(sizes.active, typ)
	layout := nativeHeaderLayout{align: 1}
	switch value := typ.Underlying().(type) {
	case *types.Basic:
		switch value.Kind() {
		case types.Bool, types.Int8, types.Uint8:
			layout.size, layout.align = 1, 1
		case types.Int16, types.Uint16:
			layout.size, layout.align = 2, 2
		case types.Int32, types.Uint32, types.Float32:
			layout.size, layout.align = 4, 4
		case types.Int, types.Uint, types.Uintptr:
			layout.size, layout.align = sizes.word, sizes.word
		case types.Int64, types.Uint64, types.Float64:
			layout.size, layout.align = 8, sizes.word
		case types.Complex64:
			layout.size, layout.align = 8, 4
		case types.Complex128:
			layout.size, layout.align = 16, sizes.word
		default:
			return nativeHeaderLayout{}, &nativeHeaderUnsupportedDomain{"native header layout has an unclosed basic/descriptor type"}
		}
	case *types.Array:
		element, err := sizes.layout(value.Elem(), depth+1)
		if err != nil || value.Len() < 0 || element.size != 0 && value.Len() > (sizes.limit()-1)/element.size {
			return nativeHeaderLayout{}, fmt.Errorf("native header array layout is unclosed or overflows")
		}
		layout.size, layout.align = element.size*value.Len(), element.align
	case *types.Struct:
		layout.offsets = make([]int64, value.NumFields())
		var lastSize int64
		for index := 0; index < value.NumFields(); index++ {
			field, err := sizes.layout(value.Field(index).Type(), depth+1)
			if err != nil {
				return nativeHeaderLayout{}, err
			}
			offset, err := sizes.round(layout.size, field.align)
			if err != nil || field.size >= sizes.limit()-offset {
				return nativeHeaderLayout{}, fmt.Errorf("native header struct extent overflows")
			}
			layout.offsets[index], layout.size, lastSize = offset, offset+field.size, field.size
			if field.align > layout.align {
				layout.align = field.align
			}
		}
		if layout.size > 0 && lastSize == 0 {
			layout.size++
		}
		var err error
		layout.size, err = sizes.round(layout.size, layout.align)
		if err != nil {
			return nativeHeaderLayout{}, err
		}
	default:
		return nativeHeaderLayout{}, &nativeHeaderUnsupportedDomain{"native header layout requires an unclosed pointer/descriptor/type-parameter model"}
	}
	sizes.cache[typ] = layout
	return layout, nil
}
