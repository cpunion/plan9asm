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
	}
}
