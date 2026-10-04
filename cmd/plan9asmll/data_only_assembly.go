package main

import (
	"fmt"
	"strings"

	"github.com/xgo-dev/plan9asm"
	"github.com/xgo-dev/plan9asm/internal/gotoolprofile"
	"golang.org/x/tools/go/packages"
)

func proveDataOnlyAssembly(pkg *packages.Package, task asmTask, file *plan9asm.File, target string, defines []string, cfg compileConfig) (*gotoolprofile.DataOnlyAssemblyProof, error) {
	if file == nil || len(file.Funcs) != 0 || len(file.Globl) == 0 {
		return nil, fmt.Errorf("data-only source requires global declarations and no TEXT")
	}
	var symbols []gotoolprofile.DataOnlySymbol
	native, err := proveAssemblyListing(pkg, task, target, defines, cfg, func(listing []byte) error {
		var err error
		symbols, err = gotoolprofile.DataOnlySymbolsFromListing(listing)
		return err
	})
	if err != nil {
		return nil, err
	}
	wanted := make(map[string]int64)
	for _, global := range file.Globl {
		name := strings.TrimSuffix(global.Sym, "<>")
		if strings.HasPrefix(name, "·") {
			name = nativeGoPackagePrefix(pkg.PkgPath) + name
		}
		name = strings.ReplaceAll(strings.ReplaceAll(name, "·", "."), "∕", "/")
		if strings.HasSuffix(global.Sym, "<>") {
			name += "<>"
		}
		if _, duplicate := wanted[name]; duplicate || global.SizeRaw != "" {
			return nil, fmt.Errorf("data-only source has duplicate or unresolved GLOBL")
		}
		wanted[name] = global.Size
	}
	if len(wanted) != len(symbols) {
		return nil, fmt.Errorf("Go native data symbols differ from translated global declarations")
	}
	for _, symbol := range symbols {
		size, present := wanted[symbol.Key()]
		if !present || size != symbol.Size {
			return nil, fmt.Errorf("Go native data symbol name/size differs from translated GLOBL")
		}
	}
	native.Protocol = gotoolprofile.DataOnlyAssemblyProtocol
	return &gotoolprofile.DataOnlyAssemblyProof{Native: *native, Symbols: symbols}, nil
}

// cmd/asm qualifies local symbols with objabi.PathToPrefix, not the logical
// import path used by LLVM. Keep the original native listing spelling intact.
// This is Go's byte-level rule; URL escaping changes different characters.
func nativeGoPackagePrefix(packagePath string) string {
	lastSlash := strings.LastIndexByte(packagePath, '/')
	const hex = "0123456789abcdef"
	var prefix strings.Builder
	for index := 0; index < len(packagePath); index++ {
		value := packagePath[index]
		if value <= ' ' || value == '.' && index > lastSlash ||
			value == '%' || value == '"' || value >= 0x7f {
			prefix.WriteByte('%')
			prefix.WriteByte(hex[value>>4])
			prefix.WriteByte(hex[value&0xf])
		} else {
			prefix.WriteByte(value)
		}
	}
	return prefix.String()
}
