//go:build go1.27

package plan9asm

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xgo-dev/plan9asm/internal/gostaticstorage"
)

func TestGoAMD64WindowsUnwindCompleteObjectListingJoinGuards(t *testing.T) {
	goAMD64WindowsUnwindCompleteObjectListingJoinGuards(t)
}

func TestGoAMD64WindowsUnwindCompleteJoinTemporaryParentAlias(t *testing.T) {
	goAMD64WindowsUnwindTemporaryParentAlias(t)
	goAMD64WindowsUnwindCompleteObjectListingJoinGuards(t)
}

func goAMD64WindowsUnwindCompleteObjectListingJoinGuards(t *testing.T) {
	t.Helper()
	pkg, request := goAMD64FrameSourceFixture(t)
	request.Target = "windows/amd64"
	source := requireGoAMD64FrameCapture(t, pkg, request)
	t.Logf("ACTUAL_WINDOWS_SOURCE_CAPTURE request=%q observed=%q functions=%d",
		request.Directory, source.request.Directory, len(source.functions))
	ctx, cancel := runtimeOracleStageContext(t)
	defer cancel()
	object, err := source.read(filepath.Join(source.owned, request.AssemblyFile+".o"))
	if err != nil {
		t.Fatal(err)
	}
	listing, err := source.read(filepath.Join(source.owned, request.AssemblyFile+".o.listing"))
	if err != nil {
		t.Fatal(err)
	}
	original, err := source.read(filepath.Join(request.Directory, request.AssemblyFile))
	if err != nil {
		t.Fatal(err)
	}
	expanded, err := source.preprocess(request.AssemblyFile, original)
	if err != nil {
		t.Fatal(err)
	}
	file, err := Parse(ArchAMD64, expanded)
	if err != nil {
		t.Fatal(err)
	}
	observed, err := source.observeWindowsUnwind(object)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := goAMD64FrameWindowsUnwindListing(listing, file, pkg, source.origins, request.Directory, observed); err != nil {
		t.Fatal("independent actual original metadata join control failed", err)
	}
	serialized, err := json.Marshal(observed)
	if err != nil {
		t.Fatal(err)
	}
	// These are ordinary read-only observations, deliberately copied only for
	// decoder negatives. None are accepted as a public Source/factory input.
	type mutation func([]byte, *gostaticstorage.WindowsAMD64AssemblyUnwind) []byte
	metadataStart := bytes.Index(listing, []byte(observed.Unwind[0].Name+" SSEHUNWINDINFO"))
	if metadataStart < 0 {
		t.Fatal("original listing omitted its actual complete metadata section")
	}
	mutations := map[string]mutation{
		"wrong metadata name": func(v []byte, _ *gostaticstorage.WindowsAMD64AssemblyUnwind) []byte {
			return bytes.Replace(v, []byte("go:sehuw.12."), []byte("go:sehuw.13."), 1)
		},
		"wrong metadata kind": func(v []byte, _ *gostaticstorage.WindowsAMD64AssemblyUnwind) []byte {
			return bytes.Replace(v, []byte("SSEHUNWINDINFO"), []byte("SRODATA"), 1)
		},
		"wrong metadata flags": func(v []byte, _ *gostaticstorage.WindowsAMD64AssemblyUnwind) []byte {
			return bytes.Replace(v, []byte("SSEHUNWINDINFO dupok"), []byte("SSEHUNWINDINFO"), 1)
		},
		"wrong metadata alignment": func(v []byte, _ *gostaticstorage.WindowsAMD64AssemblyUnwind) []byte {
			return bytes.Replace(v, []byte("size=12 align=0x4"), []byte("size=12 align=0x8"), 1)
		},
		"wrong metadata byte": func(v []byte, _ *gostaticstorage.WindowsAMD64AssemblyUnwind) []byte {
			return bytes.Replace(v, []byte("01 04 02 05"), []byte("02 04 02 05"), 1)
		},
		"wrong metadata ASCII": func(v []byte, _ *gostaticstorage.WindowsAMD64AssemblyUnwind) []byte {
			return bytes.Replace(v, []byte(".......P...."), []byte("............"), 1)
		},
		"missing metadata section": func(v []byte, _ *gostaticstorage.WindowsAMD64AssemblyUnwind) []byte { return v[:metadataStart] },
		"duplicate metadata section": func(v []byte, _ *gostaticstorage.WindowsAMD64AssemblyUnwind) []byte {
			return append(v, v[metadataStart:]...)
		},
		"unknown trailing record": func(v []byte, _ *gostaticstorage.WindowsAMD64AssemblyUnwind) []byte {
			return append(v, []byte("UNKNOWN_NATIVE_RECORD\n")...)
		},
		"truncated metadata EOF": func(v []byte, _ *gostaticstorage.WindowsAMD64AssemblyUnwind) []byte { return v[:len(v)-5] },
		"different actual TEXT bytes": func(v []byte, o *gostaticstorage.WindowsAMD64AssemblyUnwind) []byte {
			o.Texts[0].Code[0] ^= 1
			return v
		},
		"different actual TEXT ABI": func(v []byte, o *gostaticstorage.WindowsAMD64AssemblyUnwind) []byte { o.Texts[0].ABI = 1; return v },
		"extra actual TEXT": func(v []byte, o *gostaticstorage.WindowsAMD64AssemblyUnwind) []byte {
			o.Texts = append(o.Texts, o.Texts[0])
			return v
		},
		"missing actual TEXT": func(v []byte, o *gostaticstorage.WindowsAMD64AssemblyUnwind) []byte { o.Texts = nil; return v },
		"wrong object source line": func(v []byte, o *gostaticstorage.WindowsAMD64AssemblyUnwind) []byte {
			o.Texts[0].Positions[0].Line++
			return v
		},
		"wrong object source file": func(v []byte, o *gostaticstorage.WindowsAMD64AssemblyUnwind) []byte {
			o.Texts[0].Positions[0].File = "other.s"
			return v
		},
		"object source-PC gap": func(v []byte, o *gostaticstorage.WindowsAMD64AssemblyUnwind) []byte {
			o.Texts[0].Positions[0].Start = 1
			return v
		},
		"object source-PC overlap": func(v []byte, o *gostaticstorage.WindowsAMD64AssemblyUnwind) []byte {
			o.Texts[0].Positions = append(o.Texts[0].Positions, o.Texts[0].Positions[0])
			return v
		},
		"object source-PC missing tail": func(v []byte, o *gostaticstorage.WindowsAMD64AssemblyUnwind) []byte {
			o.Texts[0].Positions = o.Texts[0].Positions[:len(o.Texts[0].Positions)-1]
			return v
		},
	}
	for _, role := range []string{"wrong push end", "wrong prologue end"} {
		mutations[role] = func(v []byte, o *gostaticstorage.WindowsAMD64AssemblyUnwind) []byte {
			metadata := &o.Unwind[0]
			oldName := metadata.Name
			if role == "wrong push end" {
				metadata.PushEnd, metadata.Data[6] = 2, 2
			} else {
				metadata.PrologueEnd, metadata.Data[1], metadata.Data[4] = 5, 5, 5
			}
			metadata.Name = "go:sehuw.12." + base64.StdEncoding.EncodeToString(metadata.Data)
			for index := range o.Texts[0].Auxiliary {
				if o.Texts[0].Auxiliary[index].Type == 14 {
					o.Texts[0].Auxiliary[index].Target.Name = metadata.Name
				}
			}
			v = bytes.ReplaceAll(v, []byte(oldName), []byte(metadata.Name))
			rowStart := bytes.Index(v[metadataStart:], []byte("\n")) + metadataStart + 1
			return append(append(v[:rowStart:rowStart], goAMD64WindowsUnwindTestByteRow(metadata.Data)...), '\n')
		}
	}
	for name, change := range mutations {
		t.Run(name, func(t *testing.T) {
			var independent gostaticstorage.WindowsAMD64AssemblyUnwind
			if err := json.Unmarshal(serialized, &independent); err != nil {
				t.Fatal(err)
			}
			changed := change(append([]byte(nil), listing...), &independent)
			if _, err := goAMD64FrameWindowsUnwindListing(changed, file, pkg, source.origins, request.Directory, independent); !errors.Is(err, ErrProbeNeedsContext) {
				t.Fatal("changed original Windows binary/listing join was accepted", err)
			}
		})
	}
	if err := source.Verify(ctx); err != nil {
		t.Fatal("read-only metadata negative controls changed original source/artifacts", err)
	}
}

func goAMD64WindowsUnwindTestByteRow(data []byte) []byte {
	var result, ascii strings.Builder
	result.WriteString("\t0x0000")
	for _, value := range data {
		fmt.Fprintf(&result, " %02x", value)
		if value < ' ' || value > 0x7e {
			value = '.'
		}
		ascii.WriteByte(value)
	}
	fmt.Fprintf(&result, "  %s", ascii.String())
	return []byte(result.String())
}
