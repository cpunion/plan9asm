package plan9asm

// These are the existing, independently qualified code-address profiles from
// sealX86Continuations. A source-native word can carry a code pointer only with
// its matching explicit target ABI. Do not infer the width from an architecture
// prefix (x86_64 gnux32 has 32-bit pointers), nor from the process's host target.
// This restriction does not apply to ordinary target-independent numerical IR.
func x86CodeAddressTargetProfile(arch, triple string) bool {
	switch triple {
	case "x86_64-apple-darwin", "x86_64-unknown-linux-gnu", "x86_64-pc-windows-msvc":
		return arch == "amd64"
	case "i386-unknown-linux-gnu", "i686-pc-windows-msvc":
		return arch == "386"
	default:
		return false
	}
}
