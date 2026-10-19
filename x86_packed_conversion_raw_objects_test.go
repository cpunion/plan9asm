package plan9asm

import (
	"fmt"
	"strings"
	"testing"
)

func x86PackedConversionContinuationObjectRows() []string {
	bank := map[int]string{16: "X", 32: "Y", 64: "Z"}
	var rows []string
	for _, spec := range x86PackedConversionOracleSpecs() {
		for _, shape := range spec.shapes() {
			destination := bank[shape.destination] + "1"
			rows = append(rows, spec.name+" "+bank[shape.source]+"0,"+destination)
			if spec.legacy {
				rows = append(rows, spec.name+" (AX),"+destination)
			} else {
				rows = append(rows, spec.name+" (AX),K1,"+destination,
					spec.name+".BCST.Z 3(BX)(CX*4),K7,"+destination)
			}
		}
	}
	for _, row := range x86SignedPackedValueRows() {
		if strings.Contains(row.op, ".") || row.mask != "" {
			continue
		}
		rows = append(rows, row.op+" "+row.input+","+row.bank+"1")
	}
	return rows
}

func TestX86PackedConversionRawInitializedFiveTargetObjects(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM22 llc required")
	}
	rows := x86PackedConversionContinuationObjectRows()
	for _, target := range x86RawReturnTargets {
		t.Run(target.triple, func(t *testing.T) {
			rawRows := x86PackedConversionContinuationRawRows(t, target.arch, rows)
			const batchSize = 64
			consumed := 0
			for start := 0; start < len(rows); start += batchSize {
				end := start + batchSize
				if end > len(rows) {
					end = len(rows)
				}
				t.Run(fmt.Sprintf("batch-%04d", start), func(t *testing.T) {
					var source, declarations strings.Builder
					for index := start; index < end; index++ {
						name := fmt.Sprintf("RawConversion%d", index)
						body, declaration := x86PackedConversionContinuationFixture(t, name, target.arch, rows[index])
						source.WriteString(strings.Replace(body, rows[index]+"\n", rawRows[index]+"\n", 1))
						declarations.WriteString(declaration)
					}
					x86LEAActualTargetGoObject(t, target.arch, target.triple, source.String())
					for _, api := range []string{"text", "module", "Go"} {
						ir, err := x86ScalarFPBitViewAPI(t, source.String(), declarations.String(), target.arch, target.triple, api)
						if err != nil || ir == "" {
							t.Fatal("initialized original raw conversion lost its public object path", api, err)
						}
						optimized := x86DwordProductIROptimize(t, ir)
						for _, pipeline := range []struct{ ir, backend string }{
							{ir, "-O0"}, {ir, "-O2"}, {optimized, "-O2"},
						} {
							if strings.Count(pipeline.ir, "asm sideeffect") != end-start ||
								strings.Contains(pipeline.ir, " fptosi ") || strings.Contains(pipeline.ir, " fptoui ") {
								t.Fatal("original raw conversions lost native exception/MXCSR or dead effects", api, pipeline.backend)
							}
							x86PackedStringCompileObject(t, llc, target.triple, pipeline.ir, pipeline.backend)
						}
					}
					consumed += end - start
				})
			}
			if consumed != len(rows) {
				t.Fatal("incomplete raw conversion object partition", consumed, len(rows))
			}
			t.Logf("%d raw functions: all42 members and each physical shape, register/native tuple, merge and zero-broadcast; 3 APIs x 3 optimization paths, objects only", consumed)
		})
	}
}
