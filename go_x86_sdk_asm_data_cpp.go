package plan9asm

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/xgo-dev/plan9asm/internal/gostaticstorage"
)

func (source *GoX86SDKASMDataSource) captureAssembly(ctx context.Context, ordinal int, name string) (gostaticstorage.SDKASMDataSource, AssemblyCPPObservation, error) {
	var result gostaticstorage.SDKASMDataSource
	var cpp AssemblyCPPObservation
	if filepath.Base(name) != name {
		return result, cpp, sdkASMDataContext("selected ASM escaped its exact package-local route")
	}
	file := filepath.Join(source.request.Directory, name)
	raw, err := source.guard.read(file)
	if err != nil {
		return result, cpp, err
	}
	cpp, positions, err := observeSDKAssemblyCPP(string(raw), AssemblyPreprocessOptions{FileName: file,
		Defines: GoAssemblerDefinesWithEnv(source.guard.env["GOOS"], source.guard.env["GOARCH"], source.guard.env),
		ReadInclude: func(parent, header string) (string, []byte, error) {
			if filepath.Base(header) != header {
				return "", nil, sdkASMDataContext("relative/foreign include needs another source route")
			}
			// Match cmd/asm's actual lookup: CWD, generated object directory,
			// then the same SDK include directory. Retain every preferred miss.
			for _, directory := range []string{source.request.Directory, source.guard.owned,
				filepath.Join(source.guard.env["GOROOT"], "pkg", "include")} {
				candidate := filepath.Join(directory, header)
				_, err := os.Lstat(candidate)
				if os.IsNotExist(err) {
					source.misses[candidate] = true
					continue
				}
				if err != nil {
					return "", nil, err
				}
				if header == "go_asm.h" && directory != source.guard.owned {
					return "", nil, sdkASMDataContext("generated header is not the observed compiler artifact")
				}
				data, err := source.guard.read(candidate)
				return candidate, data, err
			}
			return "", nil, sdkASMDataContext("active header has no exact SDK/compiler source route")
		}})
	if err != nil {
		return result, cpp, err
	}
	source.observation.Positions = append(source.observation.Positions, positions)
	expanded, err := source.artifact(fmt.Sprintf("expanded-%d.s", ordinal), []byte(positions.Mapped))
	if err != nil {
		return result, cpp, err
	}
	originalObject := filepath.Join(source.guard.owned, fmt.Sprintf("original-%d.o", ordinal))
	expandedObject := filepath.Join(source.guard.owned, fmt.Sprintf("expanded-%d.o", ordinal))
	for _, command := range [][2]string{{file, originalObject}, {expanded, expandedObject}} {
		args := append(source.asmArguments(source.guard.owned), "-o", command[1], command[0])
		if _, err := source.run(ctx, args...); err != nil {
			return result, cpp, err
		}
	}
	originalBytes, err := source.guard.read(originalObject)
	if err != nil {
		return result, cpp, err
	}
	expandedBytes, err := source.guard.read(expandedObject)
	if err != nil {
		return result, cpp, err
	}
	return gostaticstorage.SDKASMDataSource{Name: name, Original: raw, Expanded: []byte(cpp.Output),
		OriginalObject: originalBytes, ExpandedObject: expandedBytes}, cpp, nil
}
