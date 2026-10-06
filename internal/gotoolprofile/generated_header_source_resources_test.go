package gotoolprofile

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"testing"
	"time"
)

func nativeHeaderGrowingStringSources(steps int) map[string]string {
	var first, second strings.Builder
	first.WriteString("package header\nconst A0 = \"x\"\n")
	second.WriteString("package header\n")
	for step := 1; step <= steps; step++ {
		file := &first
		if step%2 != 0 {
			file = &second
		}
		fmt.Fprintf(file, "const A%d = (A%d) + A%d\n", step, step-1, step-1)
	}
	fmt.Fprintf(&second, "const Alias = (A%d)\nvar Observed = len(Alias)\n", steps)
	return map[string]string{"first.go": first.String(), "second.go": second.String()}
}

func TestGeneratedHeaderSourceStringExpansionBoundBeforeTypecheck(t *testing.T) {
	const childKey = "PLAN9ASM_NATIVE_HEADER_RESOURCE_CHILD"
	if os.Getenv(childKey) == "1" {
		// This is a soft Go memory limit, not an OS allocation-security claim.
		// The fixed 24-step witness has a 16 MiB largest original string, so the
		// pre-fix child remains a deliberately bounded diagnostic, not an OOM test.
		debug.SetMemoryLimit(128 << 20)
		var before, after runtime.MemStats
		runtime.ReadMemStats(&before)
		definitions, err := ReplayNativeGeneratedHeader("example.invalid/header", "linux/amd64", "go1.20", nativeHeaderGrowingStringSources(24))
		runtime.ReadMemStats(&after)
		t.Logf("NATIVE_HEADER_RESOURCE steps=24 allocated=%d retained=%d definitions=%d error=%v", after.TotalAlloc-before.TotalAlloc,
			after.HeapAlloc, len(definitions), err)
		if definitions != nil || !NativeHeaderSourceDomainUnsupported(err) {
			t.Fatalf("bounded source model expanded a compact string DAG before its resource-domain refusal: definitions=%d error=%v", len(definitions), err)
		}
		if after.TotalAlloc-before.TotalAlloc > 8<<20 {
			t.Fatalf("string resource guard ran after expensive type/constant materialization: allocated=%d", after.TotalAlloc-before.TotalAlloc)
		}
		return
	}

	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	env := append(os.Environ(), childKey+"=1", "GOMEMLIMIT=128MiB")
	stdout, stderr, err := RunBounded(ctx, t.TempDir(), env, binary,
		"-test.run=^TestGeneratedHeaderSourceStringExpansionBoundBeforeTypecheck$", "-test.v", "-test.timeout=25s")
	if err != nil || len(stderr) != 0 || !strings.Contains(string(stdout), "NATIVE_HEADER_RESOURCE steps=24") {
		t.Fatalf("actual bounded model child failed its pre-typecheck resource contract: %v\n%s\n%s", err, stdout, stderr)
	}
	t.Logf("original resource child receipt:\n%s", stdout)
}

func TestGeneratedHeaderSourceOrdinaryStringConstantsRemainExact(t *testing.T) {
	sources := nativeHeaderGrowingStringSources(6)
	definitions, err := ReplayNativeGeneratedHeader("example.invalid/header", "linux/amd64", "go1.20", sources)
	if err != nil || definitions["const_Alias"] != fmt.Sprintf("%q", strings.Repeat("x", 64)) || definitions["const_A0"] != "\"x\"" {
		t.Fatalf("ordinary alias/parenthesis/cross-file string semantics changed: %v %v", definitions, err)
	}
	sources["conversion.go"] = "package header\ntype Text string\nconst Converted = Text(\"a\" + \"b\")\nconst RuneText = string(65)\nconst Repeated = Converted\n"
	definitions, err = ReplayNativeGeneratedHeader("example.invalid/header", "linux/amd64", "go1.20", sources)
	if err != nil || definitions["const_Converted"] != "\"ab\"" || definitions["const_RuneText"] != "\"A\"" || definitions["const_Repeated"] != "\"ab\"" {
		t.Fatalf("bounded named/rune conversion string semantics changed: %v %v", definitions, err)
	}
}
