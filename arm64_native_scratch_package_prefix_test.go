package plan9asm

import "testing"

func TestARM64NativeScratchActualPackageSymbolNamespaces(t *testing.T) {
	llc, _, _ := findLlcAndClang(t)
	for _, packagePath := range []string{
		"example.invalid/plain",
		"example.invalid/with.dot",
		"gopkg.in/scratch.v3",
		"example.invalid/deep.with.dot/plain",
		"example.invalid/plain/deep.with.dot",
	} {
		t.Run(packagePath, func(t *testing.T) {
			_, pkg, file, options, source := arm64ScratchActualBoundSourceWithPackage(t,
				"linux", "aarch64-unknown-linux-gnu", arm64ScratchSelectedSource, packagePath)
			for _, name := range []string{"Scratch", "Value"} {
				logical := pkg.Path + "." + name
				if source.abis[logical] != "ABI0" || len(source.native[logical]) < 2 {
					t.Fatalf("actual package ABI and all original native bodies lost logical identity %q: ABI=%q bodies=%d",
						logical, source.abis[logical], len(source.native[logical]))
				}
			}
			for _, api := range []string{"text", "module", "GoModule"} {
				t.Run(api, func(t *testing.T) {
					ir, err := arm64ScratchSourceAPI(t, pkg, file, options, api)
					if err != nil {
						t.Fatalf("actual stock-Go accepted package symbol identity prevented %s: %v", api, err)
					}
					compileLLVMToObject(t, llc, "aarch64-unknown-linux-gnu", "scratch.ll", "scratch.o", ir)
				})
			}
		})
	}
}
