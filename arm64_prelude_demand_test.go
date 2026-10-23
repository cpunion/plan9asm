package plan9asm

import (
	"strings"
	"testing"
)

func TestARM64PreludeContainsOnlyUsedDeclarations(t *testing.T) {
	for _, row := range []struct {
		name string
		body string
		want []string
	}{
		{
			name: "scalar",
			body: "MOVD $7, R0\nMOVD R0, ret+0(FP)\nRET",
		},
		{
			name: "annotated-comment",
			body: "MOVD $7, R0 // @llvm.bitreverse.i64( is only a comment\nMOVD R0, ret+0(FP)\nRET",
		},
		{
			name: "scalar-intrinsics",
			body: "MOVD $7, R0\nRBIT R0, R1\nCLZW R1, R0\nMOVD R0, ret+0(FP)\nRET",
			want: []string{"llvm.bitreverse.i64", "llvm.ctlz.i32"},
		},
		{
			name: "scalable-intrinsic",
			body: "CNTB R0\nMOVD R0, ret+0(FP)\nRET",
			want: []string{"llvm.aarch64.sve.cntb"},
		},
	} {
		for _, api := range []string{"text", "module", "Go"} {
			t.Run(row.name+"/"+api, func(t *testing.T) {
				ir, err := targetBackendRequest(t, api, "aarch64-unknown-linux-gnu", targetBackendCase{
					arch: ArchARM64, goarch: "arm64", body: row.body, result: true,
				})
				if err != nil {
					t.Fatal(err)
				}
				var declarations []string
				for _, line := range strings.Split(ir, "\n") {
					if strings.HasPrefix(line, "declare ") {
						declarations = append(declarations, line)
					}
				}
				if len(declarations) != len(row.want) {
					t.Fatalf("generated %d declarations, want only %d used declarations (IR: %d bytes)",
						len(declarations), len(row.want), len(ir))
				}
				for _, name := range row.want {
					found := false
					for _, declaration := range declarations {
						found = found || strings.Contains(declaration, "@"+name+"(")
					}
					if !found {
						t.Errorf("missing used declaration %s", name)
					}
				}
			})
		}
	}
}
