package plan9asm

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/xgo-dev/llvm"
)

// The original Go programs contain KMOVQ, but their runtime owners execute
// translated LLVM, not the Go machine objects. Observe that actual backend
// output before inferring a BW requirement from a source spelling.
func x86OriginalFloatingOpcodeListing(t *testing.T, triple, level, ir string, masked bool, fmas int) {
	t.Helper()
	llc, objdump := findLLVM22Tool("llc"), findLLVM22Tool("llvm-objdump")
	if llc == "" || objdump == "" {
		t.Fatal("actual floating runtime ISA requires LLVM22 llc and llvm-objdump")
	}
	listing := x86FourSourceNativeObjectListing(t, llc, objdump, triple, level, ir)
	if strings.Contains(listing, "<unknown>") || strings.Contains(listing, "invalid") {
		t.Fatal("actual original floating runtime object contains an undecoded instruction", listing)
	}
	rows := regexp.MustCompile(`(?m)^\s*[0-9a-f]+:\s+((?:[0-9a-f]{2}\s+)+)([a-z][a-z0-9.]*)`).FindAllStringSubmatch(listing, -1)
	if len(rows) == 0 {
		t.Fatal("actual original floating runtime object has no independent machine instructions", listing)
	}
	seen, kmovw, fmaCount := make(map[string]bool), 0, 0
	for _, row := range rows {
		op := row[2]
		seen[op] = true
		if op == "kmovw" {
			kmovw++
		}
		if strings.HasPrefix(op, "vfm") || strings.HasPrefix(op, "vfnm") {
			fmaCount++
		}
		// BW/DQ opmask transports are not interchangeable with the F-only
		// same-call KMOVW used by these floating native emitters.
		if strings.HasPrefix(op, "k") && !strings.HasSuffix(op, "w") {
			t.Fatal("actual LLVM runtime requires an unobserved non-F opmask ISA", row[0])
		}
		if strings.Fields(row[1])[0] == "62" && !x86OriginalFloatingEVEXWithinProfile(op) {
			t.Fatal("actual LLVM runtime contains an EVEX opcode outside its independently bounded F/VL profile", row[0])
		}
	}
	if masked && kmovw == 0 || fmaCount != fmas {
		t.Fatal("actual original runtime lost its same-call mask transport or original native operations", kmovw, fmaCount, fmas, listing)
	}
	var inventory []string
	for op := range seen {
		inventory = append(inventory, op)
	}
	sort.Strings(inventory)
	t.Logf("actual LLVM22 %s %s: %d instructions, KMOVW=%d, FMA=%d, opcodes=%s; object evidence, not CPU execution",
		triple, level, len(rows), kmovw, fmaCount, strings.Join(inventory, ","))
}

// This is a deliberately bounded independent ISA contract for these two
// original runtime programs, not a general instruction/CPUID classifier.
// Unknown backend choices require review instead of silently adding a CPU bit.
func x86OriginalFloatingEVEXWithinProfile(op string) bool {
	for _, member := range append(go127FMA3PackedOps(), go127FMA3ScalarOps()...) {
		if op == strings.ToLower(member) {
			return true
		}
	}
	switch op {
	case "vcvtsd2ss", "vcvtss2sd", "vmovaps", "vmovapd", "vmovups", "vmovupd", "vmovss", "vmovsd",
		"vmovdqa32", "vmovdqa64", "vmovdqu32", "vmovdqu64", "vpxord", "vpxorq",
		"vpandd", "vpandq", "vpandnd", "vpandnq", "vpord", "vporq", "vpternlogd", "vpternlogq",
		"vbroadcastss", "vbroadcastsd", "vextractf32x4", "vextracti32x4", "vextractf64x4", "vextracti64x4",
		"vinsertf32x4", "vinserti32x4", "vinsertf64x4", "vinserti64x4", "vshuff32x4", "vshufi32x4",
		"vshufps", "vshufpd", "vpermilps", "vpermilpd":
		return true
	}
	return false
}

func x86OriginalFloatingRuntimeObjectPipelines(t *testing.T, triple, ir string, masked bool, fmas int) {
	t.Helper()
	for _, pipeline := range []struct{ name, body, level string }{
		{"O0", ir, "-O0"}, {"O2", ir, "-O2"}, {"IR_O2_backend_O2", x86DwordProductIROptimize(t, ir), "-O2"},
	} {
		t.Run(pipeline.name, func(t *testing.T) {
			x86OriginalFloatingOpcodeListing(t, triple, pipeline.level, pipeline.body, masked, fmas)
		})
	}
}

func TestX86OriginalScalarPrecisionRuntimeActualBackendISA(t *testing.T) {
	source := x86ScalarPrecisionOriginalProgram(t)
	if strings.Count(source, "KMOVQ mask+40(FP), K1") != 1 {
		t.Fatal("original Go KMOVQ memory transport changed")
	}
	for _, triple := range []string{"x86_64-unknown-linux-gnu", "x86_64-apple-darwin", "x86_64-pc-windows-msvc"} {
		t.Run(triple, func(t *testing.T) {
			x86LEAActualTargetGoObject(t, "amd64", triple, "#include \"textflag.h\"\n"+source)
			file, err := Parse(ArchAMD64, source)
			if err != nil {
				t.Fatal(err)
			}
			// Keep the original owner's exact five pointer/one I64 manual ABI0
			// declaration, not a different typed or caller translation route.
			frame := FrameLayout{Params: []FrameSlot{
				{Offset: 0, Type: Ptr, Index: 0, Field: -1},
				{Offset: 8, Type: Ptr, Index: 1, Field: -1},
				{Offset: 16, Type: Ptr, Index: 2, Field: -1},
				{Offset: 24, Type: Ptr, Index: 3, Field: -1},
				{Offset: 32, Type: Ptr, Index: 4, Field: -1},
				{Offset: 40, Type: I64, Index: 5, Field: -1},
			}}
			ir, err := Translate(file, Options{TargetTriple: triple, Goarch: "amd64", Sigs: map[string]FuncSig{
				"scalarPrecisionConvertSemantics": {Name: "scalarPrecisionConvertSemantics", Args: []LLVMType{Ptr, Ptr, Ptr, Ptr, Ptr, I64}, Ret: Void, Frame: frame},
			}})
			if err != nil {
				t.Fatal("original scalar runtime translation failed before actual backend ISA observation", err)
			}
			x86OriginalFloatingRuntimeObjectPipelines(t, triple, ir, true, 0)
		})
	}
}

func x86OriginalFMACallerObjectSource(t *testing.T, formats []string, first int) (string, string) {
	t.Helper()
	data, file, set := x86OriginalFloatingPolicyFile(t, "amd64_go_caller_fma3_runtime_test.go")
	batch := x86OriginalFloatingPolicyFunction(t, file, "runX86FMA3GoCallerNumericalBatch")
	if fmt.Sprintf("%x", sha256.Sum256(data[set.Position(batch.Pos()).Offset:set.Position(batch.End()).Offset])) !=
		"671c54ae2be89372425614cf9d9ef893080a25f929ad8e31d3087d69bbb927b7" || len(x86FMA3NamedFormats("amd64")) != 1836 {
		t.Fatal("complete original FMA runtime builder or 1836-format inventory changed")
	}
	var goSource, assembly strings.Builder
	goSource.WriteString("package caller\nimport \"unsafe\"\n")
	for index, instruction := range formats {
		name := fmt.Sprintf("fmacase%d", first+index)
		base := strings.Split(strings.Fields(instruction)[0], ".")[0]
		wide := strings.HasSuffix(base, "PD") || strings.HasSuffix(base, "SD")
		declaration := x86CallerFixtureSource[strings.Index(x86CallerFixtureSource, "//go:noescape"):]
		declaration = strings.ReplaceAll(declaration, "kernel", name)
		declaration = strings.Replace(declaration, "func Call(", "func Call"+name+"(", 1)
		if wide {
			declaration = strings.ReplaceAll(declaration, "float32", "float64")
		}
		goSource.WriteString(declaration)
		bank := "X"
		if strings.Contains(instruction, "Y1") {
			bank = "Y"
		} else if strings.Contains(instruction, "Z1") {
			bank = "Z"
		}
		body := fmt.Sprintf("VMOVUPS (DI),%s0\nVMOVUPS (SI),%s1\nVMOVUPS 64(SI),%s2\nKMOVQ AX,K1\n", bank, bank, bank)
		body += strings.ReplaceAll(instruction, "(AX)", "(DI)") + "\n"
		store := "VMOVSS X2,(DX)"
		if wide {
			store = "VMOVSD X2,(DX)"
		}
		body += store
		machine := strings.ReplaceAll(x86CallerFixtureAssembly, "·kernel", "·"+name)
		machine = strings.Replace(machine, "MOVL $7,AX", "MOVQ (CX),AX", 1)
		assembly.WriteString(strings.Replace(machine, "MOVL AX,(DX)", body, 1))
	}
	return goSource.String(), assembly.String()
}

func TestX86OriginalFMACallerRuntimeActualBackendISA(t *testing.T) {
	for _, op := range []string{
		"vmovdqu8", "vmovdqu16", "vpmovb2m", "vpmovw2m", "vpaddb", "vpsllw",
		"vxorps", "vxorpd", "vextractf64x2", "vinserti64x2", "vfmadd132ph", "vfmadd132sh",
	} {
		if x86OriginalFloatingEVEXWithinProfile(op) {
			t.Fatal("object opcode contract admitted an independent BW, DQ or FP16 dependency", op)
		}
	}
	formats := make(map[string]bool)
	for _, row := range x86FMA3NamedFormats("amd64") {
		formats[row] = true
	}
	var rows []string
	for _, row := range x86FMA3InitializedAxisRows("amd64") {
		if formats[row] { // Exclude eight high-bank axes absent from the original owner.
			rows = append(rows, row)
		}
	}
	if len(rows) != 170 {
		t.Fatal("original FMA runtime's complete orthogonal physical/transport inventory changed", len(rows))
	}
	for _, triple := range []string{"x86_64-unknown-linux-gnu", "x86_64-apple-darwin", "x86_64-pc-windows-msvc"} {
		t.Run(triple, func(t *testing.T) {
			for first := 0; first < len(rows); first += 64 {
				end := first + 64
				if end > len(rows) {
					end = len(rows)
				}
				t.Run(fmt.Sprintf("batch%d", first/64), func(t *testing.T) {
					goSource, source := x86OriginalFMACallerObjectSource(t, rows[first:end], first)
					x86LEAActualTargetGoObject(t, "amd64", triple, source)
					pkg := x86CallerRuntimePackage(t, goSource, source)
					ctx := llvm.NewContext()
					defer ctx.Dispose()
					translation, err := translateGoModuleInContext(ctx, pkg, []byte(source), GoModuleOptions{
						FileName: "value_amd64.s", GOARCH: "amd64", TargetTriple: triple,
						ResolveSym: func(symbol string) string { return strings.Replace(symbol, "·", "caller.", 1) },
					})
					if err != nil {
						t.Fatal("original FMA caller translation failed before backend ISA observation", err)
					}
					defer translation.Module.Dispose()
					masked := false
					for _, row := range rows[first:end] {
						masked = masked || strings.Contains(row, "K1")
					}
					x86OriginalFloatingRuntimeObjectPipelines(t, triple, translation.Module.String(), masked, end-first)
				})
			}
		})
	}
	t.Log("170 original-owner orthogonal entries, not a new 1836-format Cartesian matrix or native CPU execution receipt")
}
