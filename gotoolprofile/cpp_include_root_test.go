package gotoolprofile

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCPPIncludeAbsenceFollowsPhysicalPackageCWD(t *testing.T) {
	base := t.TempDir()
	physical, tool := filepath.Join(base, "real/module"), filepath.Join(base, "tool")
	for name, content := range map[string]string{
		"real/module/pkg/native.s": "#include \"../../src/value.h\"\nTEXT ·Probe(SB),$0-0\nRET\nDATA ·value(SB)/4,$SOURCE_VALUE\nGLOBL ·value(SB),8,$4\n",
		"real/src/value.h":         "#define SOURCE_VALUE 99\n",
		"tool/src/value.h":         "#define SOURCE_VALUE 42\n",
	} {
		file := filepath.Join(base, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(file), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}
	alias := filepath.Join(base, "alias")
	if err := os.Symlink(physical, alias); err != nil {
		t.Fatal(err)
	}
	for _, target := range []string{"linux/386", "linux/amd64", "linux/arm", "linux/arm64", "js/wasm"} {
		parts := strings.Split(target, "/")
		command := exec.Command("go", "tool", "asm", "-S", "-I", filepath.Join(tool, "pkg/include"),
			"-o", filepath.Join(t.TempDir(), "go.o"), "native.s")
		command.Dir = filepath.Join(alias, "pkg")
		command.Env = append(os.Environ(), "GOOS="+parts[0], "GOARCH="+parts[1])
		out, err := command.CombinedOutput()
		if err != nil || !strings.Contains(string(out), "63 00 00 00") {
			t.Fatalf("actual Go physical package-CWD selected source value 99 (%s): %v\n%s", target, err, out)
		}
	}
	miss := CPPPackageSearchMiss{
		Protocol: CPPPackageSearchMissProtocol, Assembly: "pkg/native.s", Source: "module/pkg/native.s",
		Include: "../../src/value.h", PackageCandidate: "../src/value.h", ToolSource: "src/value.h",
	}
	if err := VerifyCPPPackageSearchMiss(miss, alias, tool); err == nil {
		t.Fatal("lexical symlink parent hid the actual package-CWD preferred source")
	}
	if err := os.Remove(filepath.Join(base, "real/src/value.h")); err != nil {
		t.Fatal(err)
	}
	if err := VerifyCPPPackageSearchMiss(miss, alias, tool); err != nil {
		t.Fatalf("physical CWD correctly missing its original candidate: %v", err)
	}
	for _, target := range []string{"linux/386", "linux/amd64", "linux/arm", "linux/arm64", "js/wasm"} {
		parts := strings.Split(target, "/")
		command := exec.Command("go", "tool", "asm", "-S", "-I", filepath.Join(tool, "pkg/include"),
			"-o", filepath.Join(t.TempDir(), "go.o"), "native.s")
		command.Dir = filepath.Join(alias, "pkg")
		command.Env = append(os.Environ(), "GOOS="+parts[0], "GOARCH="+parts[1])
		out, err := command.CombinedOutput()
		if err != nil || !strings.Contains(string(out), "2a 00 00 00") {
			t.Fatalf("actual Go fixed tool fallback selected value 42 (%s): %v\n%s", target, err, out)
		}
	}
}
