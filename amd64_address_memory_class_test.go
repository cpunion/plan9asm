package plan9asm

import (
	"fmt"
	"strings"
	"testing"

	"github.com/xgo-dev/llvm"
)

func TestX86MemoryClassKeepsAddressImmediatesDistinct(t *testing.T) {
	for _, source := range []string{
		"$·cell(SB)", "$·cell+8(SB)", "$16(AX)", "$16(AX)(CX*4)",
	} {
		t.Run(source, func(t *testing.T) {
			operand, err := parseOperandForArch(ArchAMD64, source)
			if err != nil {
				t.Fatal(err)
			}
			if operand.Kind != OpSym || !strings.HasPrefix(operand.Sym, "$") {
				t.Fatalf("source address identity lost: %#v", operand)
			}
			if isAMD64MemoryOperand(operand) {
				t.Fatalf("Go TYPE_ADDR became a memory operand: %#v", operand)
			}
			// A caller-built operand must obey the same lexical boundary.
			operand.Sym = " \t" + operand.Sym
			if isAMD64MemoryOperand(operand) {
				t.Fatal("whitespace granted address-immediate memory permission")
			}
		})
	}

	for _, source := range []string{
		"·cell(SB)", "·cell+8(SB)", "(AX)", "16(AX)(CX*4)",
		"arg+0(FP)", "local-8(SP)", "(FS)", "(GS)",
	} {
		t.Run("memory/"+source, func(t *testing.T) {
			operand, err := parseOperandForArch(ArchAMD64, source)
			if err != nil || !isAMD64MemoryOperand(operand) {
				t.Fatalf("ordinary memory class changed: %#v: %v", operand, err)
			}
		})
	}
}

func TestX86AddressImmediateMemoryOnlyFormsRejectAcrossAPIs(t *testing.T) {
	// Independent witnesses from Go's Ym/Ymb/Yml, yxchg, SSE memory,
	// scalar-memory and non-temporal rows. The source contains a TYPE_ADDR
	// where a real memory operand is required; relocation is not a load/store.
	rows := []string{
		"XADDB AL,%s", "XADDW CX,%s", "XADDL CX,%s",
		"XCHGB AL,%s", "XCHGW CX,%s", "XCHGL CX,%s",
		"INCB %s", "INCW %s", "INCL %s",
		"DECB %s", "DECW %s", "DECL %s",
		"NEGB %s", "NEGW %s", "NEGL %s",
		"NOTB %s", "NOTW %s", "NOTL %s",
		"BTCL $1,%s", "BTRL $1,%s", "BTSL $1,%s",
		"STMXCSR %s", "LDMXCSR %s", "CLFLUSH %s",
		"MOVNTIL CX,%s", "MOVNTO X0,%s", "MOVNTPS X0,%s",
		"ADDSUBPS %s,X0", "ADDSUBPD %s,X0",
		"AESENC %s,X0", "PSHUFB %s,X0", "LDDQU %s,X0",
	}
	for _, arch := range []string{"386", "amd64"} {
		for _, row := range rows {
			for _, address := range []string{"$·cell(SB)", "$16(AX)"} {
				instruction := fmt.Sprintf(row, address)
				t.Run(arch+"/"+instruction, func(t *testing.T) {
					memorySource := "TEXT ·Value(SB),4,$0-0\n" +
						fmt.Sprintf(row, strings.TrimPrefix(address, "$")) + "\nRET\n"
					goAssemblerOracleOutput(t, arch, memorySource, "", true)
					source := "TEXT ·Value(SB),4,$0-0\n" + instruction + "\nRET\n"
					goAssemblerOracleOutput(t, arch, source, "", false)
					for _, target := range x86AddressMemoryTargets(arch) {
						for _, api := range []string{"text", "module", "go"} {
							t.Run(target+"/"+api, func(t *testing.T) {
								_, err := x86AddressMemoryAPI(t, api, source, arch, target)
								if err == nil {
									t.Fatal("Go-rejected address immediate acquired memory semantics")
								}
							})
						}
					}
				})
			}
		}
	}
}

func x86AddressMemoryTargets(arch string) []string {
	if arch == "386" {
		return []string{"i386-unknown-linux-gnu", "i686-pc-windows-msvc"}
	}
	return []string{
		"x86_64-apple-darwin", "x86_64-unknown-linux-gnu",
		"x86_64-pc-windows-msvc",
	}
}

func x86AddressMemoryAPI(t *testing.T, api, source, arch, target string) (string, error) {
	t.Helper()
	file, err := Parse(ArchAMD64, source)
	if err != nil {
		t.Fatal(err)
	}
	resolve := func(symbol string) string { return strings.TrimPrefix(symbol, "·") }
	options := Options{
		Goarch: arch, TargetTriple: target, ResolveSym: resolve,
		Sigs: map[string]FuncSig{"Value": {Name: "Value", Ret: Void}},
	}
	if api == "text" {
		return Translate(file, options)
	}
	ctx := llvm.NewContext()
	defer ctx.Dispose()
	if api == "module" {
		module, err := TranslateModuleInContext(ctx, file, options)
		if !module.IsNil() {
			defer module.Dispose()
		}
		if err != nil {
			return "", err
		}
		return module.String(), nil
	}
	if api != "go" {
		t.Fatalf("unregistered API %q", api)
	}
	pkg := mustGoPackage(t, "example.invalid/addressclass", "package addressclass\nfunc Value()\n")
	translated, err := translateGoModuleInContext(ctx, pkg, []byte(source), GoModuleOptions{
		GOARCH: arch, TargetTriple: target, ResolveSym: resolve,
	})
	if translated != nil && !translated.Module.IsNil() {
		defer translated.Module.Dispose()
	}
	if err != nil {
		return "", err
	}
	return translated.Module.String(), nil
}
