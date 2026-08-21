package plan9asm

import (
	"strings"
	"testing"
)

func TestTranslateRetjmp(t *testing.T) {
	for _, tc := range retjmpArchitectures {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			file, err := Parse(tc.arch, "TEXT ·f(SB),NOSPLIT,$0-0\n\tRET ·next(SB)\n")
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			ir, err := Translate(file, retjmpOptions(tc.goarch))
			if err != nil {
				t.Fatalf("Translate() error = %v", err)
			}
			if !strings.Contains(ir, "call void @example.next()") {
				t.Fatalf("RET target was not lowered as a tail jump:\n%s", ir)
			}
		})
	}
}

var retjmpArchitectures = []struct {
	name   string
	arch   Arch
	goarch string
}{
	{name: "arm", arch: ArchARM, goarch: "arm"},
	{name: "arm64", arch: ArchARM64, goarch: "arm64"},
	{name: "amd64", arch: ArchAMD64, goarch: "amd64"},
}

func retjmpOptions(goarch string) Options {
	return Options{
		ResolveSym: func(sym string) string { return "example." + strings.TrimPrefix(sym, "·") },
		Goarch:     goarch,
		Sigs: map[string]FuncSig{
			"example.f":    {Name: "example.f", Ret: Void},
			"example.next": {Name: "example.next", Ret: Void},
		},
	}
}

func TestTranslateRetRegister(t *testing.T) {
	for _, tc := range retjmpArchitectures {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			file, err := Parse(tc.arch, "TEXT ·f(SB),NOSPLIT,$0-0\n\tRET (R27)\n")
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			ir, err := Translate(file, retjmpOptions(tc.goarch))
			if err != nil {
				t.Fatalf("Translate() error = %v", err)
			}
			if !strings.Contains(ir, "ret void") {
				t.Fatalf("register RET was not lowered as a function return:\n%s", ir)
			}
		})
	}
}

func TestTranslateRetjmpRejectsMultipleTargets(t *testing.T) {
	for _, tc := range retjmpArchitectures {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			file, err := Parse(tc.arch, "TEXT ·f(SB),NOSPLIT,$0-0\n\tRET ·one(SB), ·two(SB)\n")
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			_, err = Translate(file, retjmpOptions(tc.goarch))
			if err == nil || !strings.Contains(err.Error(), "RET expects at most 1 operand") {
				t.Fatalf("Translate() error = %v, want RET operand count error", err)
			}
		})
	}
}
