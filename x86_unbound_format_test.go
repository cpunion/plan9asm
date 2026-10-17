package plan9asm

import (
	"errors"
	"testing"
)

// An original format-only function can lack numerical inputs, native addresses
// or a closed physical register state. Preserve its complete Go object and
// public Context result instead of treating a direct emitter as a public ABI.
func x86UnboundVoidFormatSource(t *testing.T, source, name, arch, triple string) (*File, map[string]FuncSig) {
	t.Helper()
	x86LEAActualTargetGoObject(t, arch, triple, "#include \"textflag.h\"\n"+source)
	for _, api := range []string{"text", "module", "Go"} {
		ir, err := x86ScalarFPBitViewAPI(t, source, "func "+name+"()", arch, triple, api)
		if ir != "" || !errors.Is(err, ErrProbeNeedsContext) {
			t.Fatal("original format fixture invented its missing source inputs", api, err)
		}
	}
	file, err := Parse(ArchAMD64, source)
	if err != nil {
		t.Fatal(err)
	}
	return file, map[string]FuncSig{name: {Name: name, Ret: Void}}
}

func x86UnboundVoidScalarFormatObject(t *testing.T, source, name, arch, triple string) string {
	t.Helper()
	file, signatures := x86UnboundVoidFormatSource(t, source, name, arch, triple)
	return x86ScalarFormatObjectUnit(t, file, signatures, arch, triple, 0)
}
