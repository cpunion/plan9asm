//go:build go1.27 && unix

package plan9asm

import (
	"strings"
	"testing"
)

// Exercise the complete generated carriers, not isolated resource-policy text.
// Every fault child must disable both core limits before mprotect or a call;
// an expected SIGSEGV is an oracle result, not permission to create core files.
func TestGoX86NumericalSBReadRuntimeFaultCorePolicy(t *testing.T) {
	for _, arch := range []string{"386", "amd64"} {
		cases := numericalExecutionCases(arch)
		carriers := []struct {
			name   string
			source string
		}{
			{"mutable", numericalRuntimeC(cases, arch)},
			{"width", numericalRuntimeWidthFaultC(cases, arch)},
		}
		if arch == "amd64" {
			carriers = append(carriers, struct {
				name   string
				source string
			}{"local", numericalLocalRuntimeC(0x8877665544332211)})
		}
		for _, carrier := range carriers {
			t.Run(arch+"/"+carrier.name, func(t *testing.T) {
				if !strings.Contains(carrier.source, "#include <sys/resource.h>") {
					t.Error("complete C carrier must include its actual resource declarations")
				}
				if err := runtimeOracleFaultCorePolicy(carrier.source); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}
