package plan9asm

import (
	"fmt"
	"go/format"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCrossLinuxRuntimeMatrixX86GoCallerCountedLoopOriginalInitializedValues(t *testing.T) {
	crossRosetta := runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" && rosettaAvailable()
	if runtime.GOARCH != "amd64" && !crossRosetta {
		t.Skip("requires the separately required Linux/amd64 caller runtime counterpart")
	}
	var source, assembly, calls strings.Builder
	source.WriteString("package caller\nimport \"unsafe\"\n")
	for _, raw := range []bool{false, true} {
		for _, separateExit := range []bool{false, true} {
			name := fmt.Sprintf("counted%t%t", raw, separateExit)
			declaration := x86CallerCountedLoopScalarSource()
			declaration = declaration[strings.Index(declaration, "//go:noescape"):]
			declaration = strings.ReplaceAll(declaration, "kernel", name)
			declaration = strings.Replace(declaration, "func Call(", "func Call"+name+"(", 1)
			source.WriteString(declaration)
			machine := x86CallerCountedLoopScalarAssembly(raw)
			if separateExit {
				machine = x86CallerCountedLoopSeparateExitAssembly(raw)
			}
			assembly.WriteString(strings.ReplaceAll(machine, "·kernel", "·"+name))
			fmt.Fprintf(&calls, `
 if got:=Call%s(x.data[:length],y.data[:]); got!=want ||
  snapshot(&x)!=originalX || snapshot(&y)!=originalY {
  t.Fatalf("%s length=%%d got=%%#x want=%%#x",length,got,want)
 }
`, name, name)
		}
	}
	goTest := `package caller
import "testing"

type guardedInput struct {
 before [4]uint32
 data [16]uint32
 after [4]uint32
}

//go:noinline
func snapshot(input *guardedInput) guardedInput { return *input }

func TestOriginalInitializedCountedValues(t *testing.T) {
 for repeat:=0; repeat<4; repeat++ {
  for _, value:=range []uint32{0,1,7,0x80000000,0xffffffff,0x12345678} {
   for length:=0; length<16; length++ {
    x := guardedInput{
     before: [4]uint32{11,17,29,53},
     data: [16]uint32{value,value^0xaabbccdd,19,23,29,31,37,41,43,47,53,59,61,67,71,73},
     after: [4]uint32{0xaabbccdd,0x12345678,0x87654321,0xffffffff},
    }
    y := guardedInput{
     before: x.after,
     data: [16]uint32{31,37,value,43,47,53,59,61,67,71,73,79,83,89,97,101},
     after: x.before,
    }
    originalX,originalY:=snapshot(&x),snapshot(&y)
    var want uint32
    for index:=0;index<length&3;index++ { want+=x.data[index]+y.data[index] }
` + calls.String() + `
   }
  }
 }
}
`
	dir := t.TempDir()
	for name, content := range map[string]string{
		"go.mod": "module caller\n\ngo 1.20\n", "value.go": source.String(),
		"value_amd64.s": assembly.String(), "value_test.go": goTest,
	} {
		bytes := []byte(content)
		if strings.HasSuffix(name, ".go") {
			var err error
			bytes, err = format.Source(bytes)
			if err != nil {
				t.Fatalf("format generated %s: %v", name, err)
			}
		}
		if err := os.WriteFile(filepath.Join(dir, name), bytes, 0600); err != nil {
			t.Fatal(err)
		}
	}
	ctx, cancel := runtimeOracleStageContext(t)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "test", "-p=2", "-vet=off", "-count=1", ".")
	command.Dir = dir
	command.Env = append(os.Environ(), "GOENV=off", "GOWORK=off", "GOTOOLCHAIN=local", "GOARCH=amd64", "CGO_ENABLED=0", "GOMAXPROCS=2")
	if output, err := runtimeOracleCombinedOutput(ctx, command); err != nil {
		t.Fatalf("first-party original Go counted integer/canary oracle: %v\n%s", err, output)
	}
	t.Log("original Go scalar oracle: 1536 initialized calls; bound zero/one/two/three, u32 wrap, both shared/separate exits; noinline full-memory snapshots compare both data buffers and adjacent canaries, not a physical source-stack witness; Darwin/Rosetta is diagnostic, not Intel hardware evidence")
}
