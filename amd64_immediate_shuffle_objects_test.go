package plan9asm

import (
	"fmt"
	"strings"
	"testing"

	"github.com/xgo-dev/llvm"
)

func TestX86ImmediateShuffleOfficialRowsThreeAPIsFiveTargets(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM22 llc required")
	}
	for _, target := range []struct{ arch, triple string }{
		{"amd64", "x86_64-apple-darwin"}, {"amd64", "x86_64-unknown-linux-gnu"},
		{"amd64", "x86_64-pc-windows-msvc"}, {"386", "i386-unknown-linux-gnu"},
		{"386", "i686-pc-windows-msvc"},
	} {
		t.Run(target.triple, func(t *testing.T) {
			rows := x86OfficialImmediateShuffleRows(t)
			var selected []string
			for _, row := range rows {
				if target.arch == "386" {
					if strings.Count(row, ",") == 3 {
						// This is an actual frontend refusal, not a skipped test.
						goAssemblerOracleOutput(t, target.arch, "TEXT f(SB),4,$0-0\n"+row+"\nRET\n", "", false)
						continue
					}
					row = strings.NewReplacer("X17", "X1", "Y17", "Y1").Replace(row)
				}
				selected = append(selected, row)
				args := strings.Split(strings.Fields(row)[1], ",")
				args[1] = "8(AX)"
				selected = append(selected, strings.Fields(row)[0]+" "+strings.Join(args, ","))
			}
			for _, op := range []string{"PSHUFL", "PSHUFD", "PSHUFHW", "PSHUFLW"} {
				selected = append(selected, op+" $255,X0,X1", op+" $0,16(AX),X1")
			}
			selected = append(selected, "PSHUFW $-128,M0,M1", "PSHUFW $127,8(AX),M1")
			if target.arch == "386" {
				for _, op := range []string{"PSHUFW", "PSHUFL", "PSHUFD", "PSHUFHW", "PSHUFLW", "VPSHUFD", "VPSHUFHW", "VPSHUFLW"} {
					bank, values := "X", []int{0, 27, 255}
					if op == "PSHUFW" {
						bank, values = "M", []int{0, -1, 127}
					} else if strings.HasPrefix(op, "V") {
						values = []int{0, -1, 255}
					}
					for _, value := range values {
						selected = append(selected, fmt.Sprintf("%s $%d,%s0,%s1", op, int64(4294967296)+int64(value), bank, bank))
					}
				}
			}
			for _, bank := range []string{"X", "Y", "Z"} {
				selected = append(selected, "VPSHUFD.BCST $0,8(AX),"+bank+"1")
				if target.arch == "amd64" {
					selected = append(selected, "VPSHUFD.BCST.Z $255,8(AX),K1,"+bank+"1")
				}
			}
			want := 76
			if target.arch == "386" {
				want = 79
			}
			if len(selected) != want {
				t.Fatal("official row/operand variation denominator changed", len(selected), want)
			}
			var source, declarations strings.Builder
			declarations.WriteString("package shuffle\n")
			sigs := map[string]FuncSig{}
			for index, row := range selected {
				name := fmt.Sprintf("F%d", index)
				fmt.Fprintf(&source, "TEXT ·%s(SB),4,$0-0\n%s\nRET\n", name, row)
				fmt.Fprintf(&declarations, "func %s()\n", name)
				sigs[name] = FuncSig{Name: name, Ret: Void}
			}
			goAssemblerOracleOutput(t, target.arch, source.String(), "", true)
			file, err := Parse(ArchAMD64, source.String())
			if err != nil {
				t.Fatal(err)
			}
			resolve := func(symbol string) string { return strings.TrimPrefix(symbol, "·") }
			options := Options{Goarch: target.arch, TargetTriple: target.triple, Sigs: sigs, ResolveSym: resolve}
			before := fmt.Sprintf("%#v %#v", file, sigs)
			ir, err := Translate(file, options)
			if err != nil {
				t.Fatal(err)
			}
			compileLLVMToObject(t, llc, target.triple, "shuffle-text.ll", "shuffle-text.o", ir)
			ctx := llvm.NewContext()
			defer ctx.Dispose()
			module, err := TranslateModuleInContext(ctx, file, options)
			if !module.IsNil() {
				defer module.Dispose()
			}
			if err != nil {
				t.Fatal(err)
			}
			compileLLVMToObject(t, llc, target.triple, "shuffle-module.ll", "shuffle-module.o", module.String())
			pkg := mustGoPackage(t, "example.invalid/shuffle", declarations.String())
			translated, err := translateGoModuleInContext(ctx, pkg, []byte(source.String()), GoModuleOptions{
				GOARCH: target.arch, TargetTriple: target.triple, ResolveSym: resolve,
			})
			if translated != nil && !translated.Module.IsNil() {
				defer translated.Module.Dispose()
			}
			if err != nil {
				t.Fatal(err)
			}
			compileLLVMToObject(t, llc, target.triple, "shuffle-go.ll", "shuffle-go.o", translated.Module.String())
			if before != fmt.Sprintf("%#v %#v", file, sigs) {
				t.Fatal("object verification mutated original input or signatures")
			}
			t.Logf("%d original Go forms, three APIs and LLVM22 objects; not a caller/runtime certificate", len(selected))
		})
	}
}
