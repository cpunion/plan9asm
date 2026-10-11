package plan9asm

import (
	"bytes"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/xgo-dev/plan9asm/internal/gostaticstorage"
)

func (source *GoAMD64FrameSource) frameDataObject(expanded string, object []byte) (*gostaticstorage.AMD64AssemblyData, error) {
	if source.env["GOVERSION"] != "go1.27.1" || source.numericalBinding != nil {
		return nil, goAMD64FrameContext("mixed DATA needs its independent actual Go object producer")
	}
	builtin, err := source.read(filepath.Join(source.env["GOROOT"], "src/cmd/internal/goobj/builtinlist.go"))
	if err != nil {
		return nil, err
	}
	observed, err := gostaticstorage.ObserveAMD64AssemblyData([]byte(expanded), object, builtin,
		source.selection.ImportPath, source.request.Target)
	if err != nil {
		return nil, goAMD64FrameContext("complete original DATA/source object: %v", err)
	}
	return &observed, nil
}

// Consume exact complete DATA blocks, never arbitrary listing lines. TEXT and
// Windows metadata remain in the original order for their independent parser.
// This observation grants no permission to use a DATA address or load its bytes.
func goAMD64FrameDataListing(listing []byte, observed *gostaticstorage.AMD64AssemblyData) ([]byte, error) {
	if observed == nil || len(observed.Layouts) == 0 || len(listing) == 0 || listing[len(listing)-1] != '\n' {
		return nil, goAMD64FrameContext("mixed DATA listing lacks a complete object/source inventory")
	}
	blocks := make(map[string][]string, len(observed.Layouts))
	var headers []string
	for _, layout := range observed.Layouts {
		block, err := goAMD64FrameDataBlock(layout)
		if err != nil {
			return nil, err
		}
		if _, duplicate := blocks[block[0]]; duplicate {
			return nil, goAMD64FrameContext("mixed DATA listing has an ambiguous complete symbol header")
		}
		blocks[block[0]] = block
		headers = append(headers, block[0])
	}
	lines := strings.Split(string(listing), "\n")
	var result strings.Builder
	next := 0
	for index := 0; index < len(lines)-1; index++ {
		block, found := blocks[lines[index]]
		if !found {
			result.WriteString(lines[index] + "\n")
			continue
		}
		if next >= len(headers) || block[0] != headers[next] ||
			index+len(block) > len(lines)-1 || !reflect.DeepEqual(lines[index:index+len(block)], block) {
			return nil, goAMD64FrameContext("mixed DATA listing bytes, extent or ordered relocations differ")
		}
		delete(blocks, block[0])
		next++
		index += len(block) - 1
	}
	if len(blocks) != 0 {
		return nil, goAMD64FrameContext("mixed DATA listing omits an original definition")
	}
	return []byte(result.String()), nil
}

func goAMD64FrameDataBlock(layout gostaticstorage.ASMGlobalLayout) ([]string, error) {
	definition := layout.Definition
	kinds := map[uint8]string{3: "SRODATA", 4: "SRODATAFIPS", 5: "SNOPTRDATA", 6: "SNOPTRDATAFIPS",
		7: "SDATA", 8: "SDATAFIPS", 9: "SBSS", 10: "SNOPTRBSS", 11: "STLSBSS"}
	kind := kinds[definition.Kind]
	if kind == "" || definition.Name == "" || definition.Name != layout.Name || definition.ABI != layout.ABI {
		return nil, goAMD64FrameContext("mixed DATA header lacks its exact source/object identity")
	}
	header := definition.Name + " " + kind + " "
	if layout.Static {
		header += "static "
	}
	if definition.Flags&1 != 0 {
		header += "dupok "
	}
	// Go's GloblPos consumes DUPOK/RODATA/NOPTR/TLSBSS only; unrelated
	// TEXT flags, including CFunc, do not change the DATA listing header.
	header += fmt.Sprintf("size=%d align=%#x", definition.Size, definition.Align)
	block := []string{header}
	for offset := 0; offset < len(definition.Data); offset += 16 {
		var row strings.Builder
		fmt.Fprintf(&row, "\t%#04x", offset)
		end := offset + 16
		if end > len(definition.Data) {
			end = len(definition.Data)
		}
		for _, value := range definition.Data[offset:end] {
			fmt.Fprintf(&row, " %02x", value)
		}
		row.WriteString(strings.Repeat("   ", offset+16-end) + "  ")
		for _, value := range definition.Data[offset:end] {
			if value < ' ' || value > '~' {
				value = '.'
			}
			row.WriteByte(value)
		}
		block = append(block, row.String())
	}
	for _, relocation := range definition.Relocations {
		if relocation.Type != 1 || !relocation.Target.ABIKnown {
			return nil, goAMD64FrameContext("mixed DATA initializer has an unbound relocation")
		}
		block = append(block, fmt.Sprintf("\trel %d+%d t=R_ADDR %s+%d",
			relocation.Offset, relocation.Size, relocation.Target.Name, relocation.Add))
	}
	return block, nil
}

func goAMD64FrameDataRelocationLine(relocation gostaticstorage.TextRelocation, observed *gostaticstorage.AMD64AssemblyData) string {
	if relocation.Type != 14 || relocation.Size != 4 || !relocation.Target.ABIKnown {
		return ""
	}
	packageIndices := map[gostaticstorage.TextSymbolNamespace]uint32{
		gostaticstorage.TextSymbolSelf:             2147483643,
		gostaticstorage.TextSymbolHashed64:         2147483646,
		gostaticstorage.TextSymbolHashed:           2147483645,
		gostaticstorage.TextSymbolNonpkgDefinition: 2147483647,
	}
	for _, layout := range observed.Layouts {
		definition := layout.Definition
		if definition.Name == relocation.Target.Name && definition.ABI == relocation.Target.ABI &&
			packageIndices[definition.Namespace] == relocation.Target.PackageIndex &&
			definition.Relative == relocation.Target.SymbolIndex {
			return fmt.Sprintf("rel %d+%d t=R_PCREL %s+%d",
				relocation.Offset, relocation.Size, relocation.Target.Name, relocation.Add)
		}
	}
	return ""
}

func goAMD64FrameDataTextJoin(functions map[string]goAMD64FrameFunction, observed *gostaticstorage.AMD64AssemblyData,
	directory string, origins map[string]int) error {
	if len(functions) != len(observed.Member.Texts) {
		return goAMD64FrameContext("mixed DATA object/listing complete TEXT inventories differ")
	}
	for _, text := range observed.Member.Texts {
		fn, found := functions[text.Name]
		if !found || text.ABI != 0 || text.Kind != 1 || text.Flags != 16 || text.Flags2 != 0 || !bytes.Equal(text.Code, fn.code) {
			return goAMD64FrameContext("mixed DATA object/listing TEXT ABI, flags or complete bytes differ")
		}
		if err := goAMD64WindowsUnwindSourcePCs(text, fn, directory, origins); err != nil {
			return err
		}
	}
	return nil
}
