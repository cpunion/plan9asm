package plan9asm

import (
	"strings"
	"testing"
)

// Parsing is excluded. These fixed functions isolate family recognition and
// feature aggregation; they do not measure lowering or LLVM compilation.
func BenchmarkX86FeatureInference(b *testing.B) {
	for _, arch := range []string{"amd64", "386"} {
		for _, row := range []string{
			"MOVL AX,BX",
			"VPXOR X0,X1,X2",
			"VMOVUPS Y0,Y1",
			"PADDL X0,X1",
			"PSHUFB X0,X1",
			"VCVTSD2SS X0,X1,X2",
		} {
			b.Run(arch+"/"+row, func(b *testing.B) {
				file, err := Parse(ArchAMD64, "TEXT feature(SB),4,$0-0\n"+
					strings.Repeat(row+"\n", 64)+"RET\n")
				if err != nil {
					b.Fatal(err)
				}
				b.ReportAllocs()
				b.ResetTimer()
				for iteration := 0; iteration < b.N; iteration++ {
					x86FeatureInferenceBenchmarkSink = inferFuncTargetFeaturesForGOARCH(ArchAMD64, arch, file.Funcs[0])
				}
			})
		}
	}
}

var x86FeatureInferenceBenchmarkSink string
