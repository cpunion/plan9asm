//go:build go1.27

package plan9asm

import (
	"bytes"
	"crypto/sha256"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/xgo-dev/plan9asm/internal/goasmcmd"
	"github.com/xgo-dev/plan9asm/internal/gostaticstorage"
)

// Keep the complete actual SDK source: the old failure occurs in the FMA
// branch, after valid scalar instructions and several original returns. A
// cropped live path or text assembled from a replacement fixture misses it.
func TestX86NumericalNativeBoundariesOriginalCompleteSDKMathExp(t *testing.T) {
	ctx, request := sdkASMDataRequest(t, "amd64", "math")
	root := filepath.Dir(filepath.Dir(request.GoBinary))
	path := filepath.Join(request.Directory, "exp_amd64.s")
	original, err := os.ReadFile(path)
	if err != nil || !strings.Contains(string(original), "VFNMADD231SD X2, X1, X0") ||
		!strings.Contains(string(original), "VFMADD213SD exprodata<>+56(SB), X0, X1") ||
		!strings.Contains(string(original), "VADDSD exprodata<>+16(SB), X0, X1") {
		t.Fatal("complete original SDK floating source is absent", err)
	}
	object := filepath.Join(t.TempDir(), "original-exp.o")
	listing, err := goasmcmd.Run(ctx, request.Directory, request.Environment, request.GoBinary,
		"tool", "asm", "-S", "-p", "math", "-I", filepath.Join(root, "pkg", "include"), "-o", object, path)
	if err != nil {
		t.Fatal("actual Go rejected the complete original SDK source", err)
	}
	objectBytes, err := os.ReadFile(object)
	if err != nil || len(objectBytes) == 0 {
		t.Fatal("complete actual original Go object is absent", err)
	}
	builtins, err := os.ReadFile(filepath.Join(root, "src", "cmd", "internal", "goobj", "builtinlist.go"))
	if err != nil {
		t.Fatal(err)
	}
	texts, err := gostaticstorage.ObserveX86AssemblyText(objectBytes, builtins, request.Target)
	if err != nil || len(texts) != 1 || texts[0].Name != "math.archExp" ||
		len(texts[0].Code) == 0 || texts[0].Size != int64(len(texts[0].Code)) {
		t.Fatal("independent whole original object/TEXT prerequisite failed", err, texts)
	}
	code := texts[0].Code
	var pcs []int64
	var listed []byte
	inText := false
	for _, line := range strings.Split(string(listing), "\n") {
		if strings.HasPrefix(line, "math.archExp STEXT ") {
			if inText {
				t.Fatal("duplicate original TEXT listing")
			}
			inText = true
			continue
		}
		if !inText {
			continue
		}
		if line != "" && line[0] != '\t' && line[0] != ' ' {
			break // DATA symbols are not executable instructions.
		}
		if chunk := x86InsertExtractListedBytes(line); len(chunk) != 0 {
			fields := strings.Fields(line)
			pc, parseErr := strconv.ParseInt(fields[0], 0, 64)
			if parseErr != nil || pc != int64(len(listed)) {
				t.Fatal("noncontiguous complete original Go TEXT listing", line)
			}
			listed = append(listed, chunk...)
		}
		match := goAMD64FrameRow.FindStringSubmatch(line)
		if match == nil || strings.HasPrefix(match[3], "TEXT") ||
			strings.HasPrefix(match[3], "PCDATA") || strings.HasPrefix(match[3], "FUNCDATA") {
			continue
		}
		pc, parseErr := strconv.ParseInt(match[1], 16, 64)
		if parseErr != nil || len(pcs) != 0 && pc <= pcs[len(pcs)-1] {
			t.Fatal("ambiguous original Go instruction PC", line)
		}
		pcs = append(pcs, pc)
	}
	if !inText || len(pcs) == 0 || pcs[0] != 0 || !bytes.Equal(code, listed) {
		t.Fatal("independent complete original object and listing bytes differ", len(code), len(listed), pcs)
	}
	after, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(original, after) {
		t.Fatal("complete original SDK source changed", err)
	}
	t.Logf("complete original ASM SHA256=%x objectSHA256=%x TEXTbytes=%d originalPCs=%d",
		sha256.Sum256(original), sha256.Sum256(objectBytes), len(code), len(pcs))
	boundaries, scalar, err := x86NumericalNativeBoundaries(code, 64)
	if err != nil {
		t.Fatal("original Go object lost its complete floating instruction boundaries", err)
	}
	if len(boundaries) != len(pcs) {
		t.Fatal("whole original source boundary count differs", len(boundaries), len(pcs))
	}
	vectorCount := 0
	for index, boundary := range boundaries {
		end := int64(len(code))
		if index+1 < len(pcs) {
			end = pcs[index+1]
		}
		if boundary.PC != pcs[index] || boundary.Length != int(end-pcs[index]) {
			t.Fatal("whole original source PC/byte boundary changed", index, boundary, pcs[index], end)
		}
		if x86NumericalVectorPrefix(code[boundary.PC:end], 64) {
			if _, exists := scalar[boundary.PC]; exists {
				t.Fatal("original floating opcode acquired a scalar NUM read", boundary)
			}
			vectorCount++
		}
	}
	if vectorCount == 0 || len(scalar)+vectorCount != len(boundaries) {
		t.Fatal("complete original FMA/native-vector branch was omitted", vectorCount, len(scalar), len(boundaries))
	}
}
