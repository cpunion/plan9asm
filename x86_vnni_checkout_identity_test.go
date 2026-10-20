package plan9asm

import (
	"bytes"
	"os"
	"testing"
)

func TestX86VNNIOriginalCheckoutLineEndingIdentity(t *testing.T) {
	assembly, err := os.ReadFile("testdata/conformance/amd64/conformance_amd64.s")
	if err != nil {
		t.Fatal(err)
	}
	reference, err := os.ReadFile("amd64_conformance_test.go")
	if err != nil {
		t.Fatal(err)
	}
	assembly = bytes.ReplaceAll(assembly, []byte("\r\n"), []byte("\n"))
	reference = bytes.ReplaceAll(reference, []byte("\r\n"), []byte("\n"))
	want, err := x86VNNIOriginalConformanceSource(assembly, reference)
	if err != nil {
		t.Fatal("complete LF originals lost their existing pinned identity", err)
	}
	for _, row := range []struct {
		name string
		asm  bool
		c    bool
	}{
		{"LF", false, false},
		{"assembly-CRLF", true, false},
		{"reference-CRLF", false, true},
		{"both-CRLF", true, true},
	} {
		t.Run(row.name, func(t *testing.T) {
			input, oracle := assembly, reference
			if row.asm {
				input = bytes.ReplaceAll(input, []byte("\n"), []byte("\r\n"))
			}
			if row.c {
				oracle = bytes.ReplaceAll(oracle, []byte("\n"), []byte("\r\n"))
			}
			got, err := x86VNNIOriginalConformanceSource(input, oracle)
			if err != nil || got != want {
				t.Fatal("repository checkout EOL changed the complete source/reference identity", err)
			}
		})
	}
	for _, row := range []struct {
		name string
		asm  []byte
		c    []byte
	}{
		{"instruction-changed", bytes.Replace(assembly, []byte("VPDPBUSD "), []byte("VPDPBUSDS "), 1), reference},
		{"reference-function-changed", assembly, bytes.Replace(reference,
			[]byte("func amd64ConformanceRuntimeC("), []byte("func foreignConformanceRuntimeC("), 1)},
	} {
		t.Run(row.name, func(t *testing.T) {
			if bytes.Equal(row.asm, assembly) && bytes.Equal(row.c, reference) {
				t.Fatal("independent source mutation did not change its actual input")
			}
			if _, err := x86VNNIOriginalConformanceSource(row.asm, row.c); err == nil {
				t.Fatal("checkout-EOL handling accepted a changed complete source/reference")
			}
		})
	}
}
