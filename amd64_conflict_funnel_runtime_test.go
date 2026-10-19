package plan9asm

import "testing"

// These are the unchanged original C numerical oracles, compiled for the
// actual host. They do not claim original AVX512 hardware execution.
func x86ConflictFunnelPortableRuntime(t *testing.T, llc, clang, name, triple, ir, mainC string) {
	t.Helper()
	for _, pipeline := range []struct {
		name      string
		optimized bool
		backend   int
	}{
		{name: "backend_O0"},
		{name: "backend_O2", backend: 2},
		{name: "IR_O2/backend_O2", optimized: true, backend: 2},
	} {
		t.Run(pipeline.name, func(t *testing.T) {
			text := ir
			if pipeline.optimized {
				text = x86DwordProductIROptimize(t, ir)
			}
			compileAndRunRuntimeTestWithCompiler(t, llc, []string{clang}, name, triple,
				text, mainC, nil, pipeline.backend)
		})
	}
}
