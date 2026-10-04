package plan9asm

import (
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestX86FloatingCompareRuntimeRequiresExplicitISAAxes(t *testing.T) {
	for _, fixture := range []struct{ file, root, helper, axis string }{
		{"amd64_float_compare_fp_runtime_test.go", "TestCrossLinuxRuntimeMatrixX86FloatingCompareFPScalarByteSpans", "x86FloatCompareFPScalarByteSpansRuntime", "x86FloatCompareSSE"},
		{"amd64_float_compare_numerical_test.go", "TestCrossLinuxRuntimeMatrixX86FloatingCompareIndependentIEEEAndMXCSR", "x86FloatCompareIndependentRuntime", "x86FloatCompareSSE"},
		{"amd64_float_compare_numerical_test.go", "TestCrossLinuxRuntimeMatrixX86FloatingCompareDeadResultEffects", "x86FloatCompareDeadResultRuntime", "x86FloatCompareSSE"},
		{"amd64_float_compare_runtime_linux_test.go", "TestCrossLinuxRuntimeMatrixX86FloatingCompareFPScalarByteSpansVEX", "x86FloatCompareFPScalarByteSpansRuntime", "x86FloatCompareVEX"},
		{"amd64_float_compare_runtime_linux_test.go", "TestCrossLinuxRuntimeMatrixX86FloatingCompareIndependentIEEEAndMXCSRVEX", "x86FloatCompareIndependentRuntime", "x86FloatCompareVEX"},
		{"amd64_float_compare_runtime_linux_test.go", "TestCrossLinuxRuntimeMatrixX86FloatingCompareDeadResultEffectsVEX", "x86FloatCompareDeadResultRuntime", "x86FloatCompareVEX"},
		{"amd64_float_compare_runtime_linux_amd64_test.go", "TestCrossLinuxRuntimeMatrixX86FloatingCompareFPScalarByteSpansEVEX", "x86FloatCompareFPScalarByteSpansRuntime", "x86FloatCompareEVEX"},
		{"amd64_float_compare_runtime_linux_amd64_test.go", "TestCrossLinuxRuntimeMatrixX86FloatingCompareIndependentIEEEAndMXCSREVEX", "x86FloatCompareIndependentRuntime", "x86FloatCompareEVEX"},
		{"amd64_float_compare_runtime_linux_amd64_test.go", "TestCrossLinuxRuntimeMatrixX86FloatingCompareDeadResultEffectsEVEX", "x86FloatCompareDeadResultRuntime", "x86FloatCompareEVEX"},
	} {
		t.Run(fixture.root, func(t *testing.T) {
			file, err := parser.ParseFile(token.NewFileSet(), fixture.file, nil, 0)
			if err != nil {
				t.Fatal(err)
			}
			var body *ast.BlockStmt
			for _, declaration := range file.Decls {
				if fn, ok := declaration.(*ast.FuncDecl); ok && fn.Name.Name == fixture.root {
					body = fn.Body
				}
			}
			if body == nil || len(body.List) != 1 {
				t.Fatal("required runtime root must directly consume its complete explicit ISA fixture")
			}
			statement, ok := body.List[0].(*ast.ExprStmt)
			if !ok {
				t.Fatal("required runtime root cannot return or substitute a skip")
			}
			call, ok := statement.X.(*ast.CallExpr)
			if !ok || len(call.Args) != 2 {
				t.Fatal("required runtime fixture must receive the explicit ISA axis")
			}
			callee, ok := call.Fun.(*ast.Ident)
			axis, axisOK := call.Args[1].(*ast.Ident)
			if !ok || callee.Name != fixture.helper || !axisOK || axis.Name != fixture.axis {
				t.Fatalf("required runtime root does not consume %s with %s", fixture.helper, fixture.axis)
			}
		})
	}
}

func TestX86FloatingCompareRequiredLinuxBuildSelection(t *testing.T) {
	for _, target := range []struct{ os, arch string }{{"linux", "amd64"}, {"linux", "arm64"}, {"darwin", "arm64"}, {"windows", "amd64"}} {
		for _, file := range []string{"amd64_float_compare_runtime_linux_test.go", "amd64_float_compare_runtime_linux_amd64_test.go", "amd64_insert_extract_native_runtime_linux_amd64_test.go"} {
			ctx := build.Default
			ctx.GOOS, ctx.GOARCH = target.os, target.arch
			actual, err := ctx.MatchFile(".", file)
			if err != nil {
				t.Fatal(err)
			}
			want := target.os == "linux" && (target.arch == "amd64" || file == "amd64_float_compare_runtime_linux_test.go" && target.arch == "arm64")
			if actual != want {
				t.Fatalf("%s on %s/%s selected=%v want=%v", file, target.os, target.arch, actual, want)
			}
		}
	}
	contents, err := os.ReadFile(".github/workflows/go-ci.yml")
	if err != nil {
		t.Fatal(err)
	}
	jobs := strings.SplitN(string(contents), "\n  cross-runtime:\n", 2)
	if len(jobs) != 2 {
		t.Fatal("required Linux runtime job is absent")
	}
	job := strings.SplitN(jobs[1], "\n  test:\n", 2)[0]
	command := regexp.MustCompile(`(?m)^        run: go test \. -run '([^']+)'`).FindStringSubmatch(job)
	if len(command) != 2 || !strings.Contains(job, "ubuntu-latest") {
		t.Fatal("required Linux native AMD64 runtime job/actual selector is absent")
	}
	selector, err := regexp.Compile(command[1])
	if err != nil {
		t.Fatal(err)
	}
	for _, axis := range []string{"VEX", "EVEX"} {
		for _, family := range []string{"FPScalarByteSpans", "IndependentIEEEAndMXCSR", "DeadResultEffects"} {
			if !selector.MatchString("TestCrossLinuxRuntimeMatrixX86FloatingCompare" + family + axis) {
				t.Fatal("required Linux CI does not select complete compare fixture", family, axis)
			}
		}
	}
	if !selector.MatchString("TestCrossLinuxRuntimeMatrixAMD64ConformanceNativeAVX512") {
		t.Fatal("required Linux CI does not select original-Go/LLVM native wide conformance")
	}
	if !selector.MatchString("TestCrossLinuxRuntimeMatrixX86InsertExtractMaskedNativeMemory") {
		t.Fatal("required Linux CI does not select all masked native memory extract cases")
	}
}

func TestAMD64ConformanceWideRuntimeIsRequired(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "amd64_float_compare_runtime_linux_amd64_test.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var body *ast.BlockStmt
	for _, declaration := range file.Decls {
		if fn, ok := declaration.(*ast.FuncDecl); ok && fn.Name.Name == "TestCrossLinuxRuntimeMatrixAMD64ConformanceNativeAVX512" {
			body = fn.Body
		}
	}
	if body == nil || len(body.List) != 3 {
		t.Fatal("original Go and LLVM wide conformance must both consume required native preflight")
	}
	assignment, ok := body.List[0].(*ast.AssignStmt)
	if !ok || len(assignment.Rhs) != 1 {
		t.Fatal("native wide preflight is missing")
	}
	call, ok := assignment.Rhs[0].(*ast.CallExpr)
	if !ok || len(call.Args) != 2 {
		t.Fatal("native wide preflight must bind the explicit EVEX execution axis")
	}
	callee, ok := call.Fun.(*ast.Ident)
	axis, axisOK := call.Args[1].(*ast.Ident)
	if !ok || callee.Name != "x86FloatCompareRuntimeTools" || !axisOK || axis.Name != "x86FloatCompareEVEX" {
		t.Fatal("native wide conformance cannot borrow ordinary/guest tools")
	}
	for index, name := range []string{"amd64ConformanceOriginalWideGo", "amd64ConformanceLLVMRuntime"} {
		statement, ok := body.List[index+1].(*ast.ExprStmt)
		if !ok {
			t.Fatal("native wide conformance cannot substitute an early return or skip")
		}
		call, ok := statement.X.(*ast.CallExpr)
		if !ok {
			t.Fatal("native wide conformance is not consumed")
		}
		callee, ok := call.Fun.(*ast.Ident)
		if !ok || callee.Name != name || len(call.Args) < 2 {
			t.Fatal("native wide conformance consumer is missing", name)
		}
		tools, ok := call.Args[1].(*ast.Ident)
		if !ok || tools.Name != "execution" {
			t.Fatal("native wide conformance dropped its validated execution route")
		}
		if index == 1 {
			if len(call.Args) != 3 {
				t.Fatal("native wide LLVM conformance must execute its full wide assertions")
			}
			wide, ok := call.Args[2].(*ast.Ident)
			if !ok || wide.Name != "true" {
				t.Fatal("native wide LLVM conformance was disabled")
			}
		}
	}
}
