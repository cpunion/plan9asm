package plan9asm

import "strings"

// This map describes actual same-module numeric storage, not external Go
// variables or a reusable source/return certificate. All typed read consumers
// still request their complete native width, independently of a selected lane.
func x86LoweredScalarDataReads(file *File, resolve func(string) string, sigs map[string]FuncSig) map[string]int64 {
	if file == nil || file.Arch != ArchAMD64 || resolve == nil {
		return nil
	}
	type object struct {
		size    int64
		invalid bool
	}
	objects := make(map[string]object)
	for _, global := range file.Globl {
		name := resolveAssemblyDataSymbol(global.Sym, resolve)
		current := objects[name]
		if global.SizeRaw != "" || global.Size < 0 || global.Size > maxDataGlobalSize {
			current.invalid = true
		}
		if global.Size > current.size {
			current.size = global.Size
		}
		objects[name] = current
	}
	for _, datum := range file.Data {
		name := resolveAssemblyDataSymbol(datum.Sym, resolve)
		current := objects[name]
		end, err := dataStmtEnd(datum)
		if err != nil || datum.Addr != "" {
			// Reject relocations in the entire resolved object, not merely
			// the selected raw name or currently read range.
			current.invalid = true
		}
		if end > current.size {
			current.size = end
		}
		objects[name] = current
	}
	for _, function := range file.Funcs {
		name := resolve(function.Sym)
		current := objects[name]
		current.invalid = true
		objects[name] = current
	}
	for name, signature := range sigs {
		for _, symbol := range []string{name, funcSigSymbol(name, signature)} {
			current := objects[symbol]
			current.invalid = true
			objects[symbol] = current
		}
	}
	reads := make(map[string]int64)
	for _, function := range file.Funcs {
		for _, ins := range function.Instrs {
			for _, operand := range ins.Args {
				if operand.Kind != OpSym || !strings.HasSuffix(strings.TrimSpace(operand.Sym), "(SB)") {
					continue
				}
				base, offset, valid := parseSBRef(operand.Sym)
				if !valid || offset < 0 || strings.IndexAny(base, "$+(),[] \t") >= 0 {
					continue
				}
				current := objects[resolveAssemblyDataSymbol(base, resolve)]
				if !current.invalid && current.size > 0 && current.size <= maxDataGlobalSize && offset < current.size {
					reads[operand.Sym] = current.size - offset
				}
			}
		}
	}
	return reads
}
