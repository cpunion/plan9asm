//go:build unix

package plan9asm

import (
	"fmt"
	"go/ast"
	"go/token"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func fixedLocalPackedRequireNativeBatchOwners(t *testing.T) {
	t.Helper()
	functions := fixedLocalLeaseFunctions(t, "amd64_fixed_local_packed_native_linux_amd64_test.go")
	const prefix = "TestCrossLinuxRuntimeMatrixGoAMD64FixedLocalPackedNativeWide"
	expected := make(map[string]int)
	for batch := 0; batch < 6; batch++ {
		expected[fmt.Sprintf("%sBatch%02d", prefix, batch)] = batch
	}
	for name := range functions {
		if strings.HasPrefix(name, prefix) {
			if _, ok := expected[name]; !ok {
				t.Error("unknown or monolithic native source owner", name)
			}
		}
	}
	for name, batch := range expected {
		function := functions[name]
		if function == nil {
			t.Error("missing complete native source owner", name)
			continue
		}
		if len(function.Body.List) != 1 {
			t.Fatal("native owner must delegate exactly once", name)
		}
		statement, ok := function.Body.List[0].(*ast.ExprStmt)
		if !ok {
			t.Fatal("native owner lost its complete source call", name)
		}
		call, ok := statement.X.(*ast.CallExpr)
		if !ok || fixedLocalLeaseCallName(call) != "fixedLocalPackedNativeWideBatch" || len(call.Args) != 2 {
			t.Fatal("native owner must retain the shared source/preflight helper", name)
		}
		fixedLocalLeaseRequireArguments(t, &ast.CallExpr{Fun: call.Fun, Args: call.Args[:1]}, "t")
		literal, ok := call.Args[1].(*ast.BasicLit)
		if !ok || literal.Kind != token.INT || literal.Value != strconv.Itoa(batch) {
			t.Fatal("native owner substituted or repeated a source batch", name)
		}
	}
	helper := functions["fixedLocalPackedNativeWideBatch"]
	if helper == nil {
		t.Error("missing independently preflighted native source helper")
		return
	}
	calls := fixedLocalLeaseCalls(helper.Body)
	if calls["x86PackedDoubleDwordNativeTools"] != 1 || calls["fixedLocalPackedRuntimeAPIs"] != 1 || calls["t.Skip"] != 0 {
		t.Fatal("every source owner needs the unchanged actual native preflight and API plan")
	}
	ast.Inspect(helper.Body, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok || fixedLocalLeaseCallName(call) != "fixedLocalPackedRuntimeAPIs" {
			return true
		}
		if len(call.Args) != 6 {
			t.Fatal("native source owner must select one complete batch")
		}
		fixedLocalLeaseRequireArguments(t, &ast.CallExpr{Fun: call.Fun, Args: call.Args[:4]}, "t", "execution", "true", "true")
		fixedLocalLeaseRequireArguments(t, &ast.CallExpr{Fun: call.Fun, Args: call.Args[5:]}, "batch")
		apis, ok := call.Args[4].(*ast.CompositeLit)
		if !ok || len(apis.Elts) != 3 {
			t.Fatal("native source owner lost a public API")
		}
		for index, want := range []string{"Translate", "Module", "GoModule"} {
			literal, ok := apis.Elts[index].(*ast.BasicLit)
			if !ok || literal.Value != strconv.Quote(want) {
				t.Fatal("native source owner changed its complete three-API plan")
			}
		}
		return true
	})
}

func TestFixedLocalPackedNativeBatchSelectionPreservesCompleteSourceUnion(t *testing.T) {
	for _, all := range []bool{false, true} {
		batches := fixedLocalPackedLeaseBatches(t, all, "x86_64-unknown-linux-gnu")
		indices, err := fixedLocalPackedBatchSelection(len(batches))
		if err != nil || len(indices) != len(batches) {
			t.Fatal("default compile-only/SSE source selection changed", err)
		}
		var assembly, declarations, rows, prototypes strings.Builder
		for index, batch := range batches {
			selected, err := fixedLocalPackedBatchSelection(len(batches), index)
			if err != nil || !reflect.DeepEqual(selected, []int{index}) || indices[index] != index {
				t.Fatal("one required owner must retain exactly its complete original source batch", index, selected, err)
			}
			assembly.WriteString(batch.assembly)
			declarations.WriteString(batch.declarations)
			rows.WriteString(batch.rows)
			prototypes.WriteString(batch.prototypes)
		}
		original, decls, cases := fixedLocalPackedRuntimeProgram(all)
		_, oracle := fixedLocalPackedRuntimeOracles(cases, "x86_64-unknown-linux-gnu")
		if assembly.String() != original || declarations.String() != decls ||
			!strings.Contains(oracle, rows.String()) || !strings.Contains(oracle, prototypes.String()) {
			t.Fatal("independent source owners cropped, renamed or duplicated the original whole program")
		}
	}
}

func TestFixedLocalPackedNativeBatchSelectionRejectsInvalidOwners(t *testing.T) {
	for _, selected := range [][]int{{-1}, {6}, {0, 0}, {0, 1}, {5, 4}} {
		if _, err := fixedLocalPackedBatchSelection(6, selected...); err == nil {
			t.Error("invalid or repeated source selection became an executable owner", selected)
		}
	}
	for _, total := range []int{-1, 0} {
		if _, err := fixedLocalPackedBatchSelection(total); err == nil {
			t.Error("empty source inventory became a successful owner", total)
		}
	}
}
