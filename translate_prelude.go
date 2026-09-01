package plan9asm

import "strings"

func emitArchPrelude(b *strings.Builder, file *File, goarch string) {
	switch file.Arch {
	case ArchARM:
		emitARMPrelude(b)
	case ArchARM64:
		emitARM64Prelude(b)
	case ArchAMD64:
		// Keep historical gating to avoid injecting x86-only prelude when
		// parsing x86 syntax for unrelated targets in tests.
		if goarch == "amd64" || goarch == "386" {
			emitAMD64Prelude(b, goarch, file)
		}
	case ArchWASM:
		emitWASMPrelude(b, file)
	}
}

func emitWASMPrelude(b *strings.Builder, file *File) {
	want := make(map[string]bool)
	for _, fn := range file.Funcs {
		for _, ins := range fn.Instrs {
			switch normalizeInstructionOpcode(ins.Op) {
			case "F64FLOOR":
				want["floor"] = true
			case "F64CEIL":
				want["ceil"] = true
			case "F64TRUNC":
				want["trunc"] = true
			}
		}
	}
	for _, name := range []string{"floor", "ceil", "trunc"} {
		if want[name] {
			b.WriteString("declare double @llvm." + name + ".f64(double)\n")
		}
	}
	if len(want) != 0 {
		b.WriteString("\n")
	}
}
