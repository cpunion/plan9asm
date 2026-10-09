package plan9asm

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/xgo-dev/plan9asm/internal/gostaticstorage"
	"golang.org/x/arch/x86/x86asm"
)

type numericalSBNativeProof struct {
	Archive, OriginalObject, ExpandedObject []byte
	OriginalListing, ExpandedListing        []byte
	OriginalText, ExpandedText              []gostaticstorage.TextRecord
	ArchiveMembers                          []gostaticstorage.TextMember
	Sources                                 map[string][]byte
}

func (source *GoX86NumericalReadSource) bindNative(ctx context.Context, variables map[string]numericalSBRead, environment map[string]string) error {
	if err := source.preprocess(); err != nil {
		return err
	}
	guard := source.guard
	selected := guard.selection
	_, arch, _ := strings.Cut(source.request.Target, "/")
	builtin, err := guard.read(filepath.Join(guard.env["GOROOT"], "src", "cmd", "internal", "goobj", "builtinlist.go"))
	if err != nil {
		return err
	}
	proof := &source.native
	proof.Archive, err = guard.read(selected.Export)
	if err != nil {
		return err
	}
	if source.role == numericalSBSDKSource {
		if source.sdk == nil || source.sdk.origin != source || source.sdk.closed {
			return numericalSBContext("native observation lacks the live exact SDK input graph")
		}
		if source.literals != nil {
			if !bytes.Equal(proof.Archive, source.literals.Observation.ObjectInputs.Archive) {
				return numericalSBContext("literal native archive changed")
			}
			proof.ArchiveMembers = source.sdk.ArchiveMembers
		} else {
			proof.ArchiveMembers, err = gostaticstorage.ObserveX86SDKAssemblyArchiveReferences(proof.Archive, builtin, source.request.Target,
				selected.SFiles, source.sdk.Referenced)
		}
	} else {
		proof.ArchiveMembers, err = gostaticstorage.ObserveX86AssemblyArchive(proof.Archive, builtin, source.request.Target, selected.SFiles)
	}
	if err != nil {
		return numericalSBContext("complete actual compiler/archive role: %v", err)
	}
	found := false
	for _, name := range selected.SFiles {
		found = found || name == source.request.AssemblyFile
	}
	if !found {
		return numericalSBContext("assembly is not actually selected")
	}
	expanded := filepath.Join(guard.owned, "expanded.s")
	expandedProgram := source.cpp.Output
	var sourcePositions [][]string
	if source.literals != nil {
		positioned := source.literals.Observation.Positions[source.literals.SourceOrdinal]
		expandedProgram = positioned.Mapped
		sourcePositions, err = numericalSDKLiteralSourcePositions(source.file, source.cpp.Output, positioned)
		if err != nil {
			return err
		}
	}
	if err := os.WriteFile(expanded, []byte(expandedProgram), 0600); err != nil {
		return err
	}
	if _, err := guard.read(expanded); err != nil {
		return err
	}
	assemble := func(input, output string) ([]byte, []byte, error) {
		arguments := []string{"tool", "asm", "-S", "-p", source.request.PackagePath, "-I", source.request.Directory, "-I", filepath.Join(guard.env["GOROOT"], "pkg", "include")}
		if source.literals != nil {
			// Match the independent SDK header producer's complete source and
			// generated-header route; no user header or substituted constant.
			arguments = []string{"tool", "asm", "-S", "-p", source.request.PackagePath, "-std",
				"-I", source.literals.source.guard.owned, "-I", filepath.Join(guard.env["GOROOT"], "pkg", "include")}
		}
		for _, define := range GoAssemblerDefinesWithEnv("linux", arch, environment) {
			arguments = append(arguments, "-D", define)
		}
		arguments = append(arguments, "-o", output, input)
		listing, err := source.run(ctx, arguments...)
		if err != nil {
			return nil, listing, err
		}
		object, err := guard.read(output)
		return object, listing, err
	}
	proof.OriginalObject, proof.OriginalListing, err = assemble(filepath.Join(source.request.Directory, source.request.AssemblyFile), filepath.Join(guard.owned, "original.o"))
	if err != nil {
		return err
	}
	proof.ExpandedObject, proof.ExpandedListing, err = assemble(expanded, filepath.Join(guard.owned, "expanded.o"))
	if err != nil {
		return err
	}
	if source.literals != nil {
		proof.OriginalText, proof.ExpandedText, err = gostaticstorage.ObserveX86SDKASMDataListingObjects(
			source.literals.Observation.ObjectInputs, source.literals.SourceOrdinal, proof.OriginalObject, proof.ExpandedObject)
	} else {
		proof.OriginalText, err = gostaticstorage.ObserveX86AssemblyText(proof.OriginalObject, builtin, source.request.Target)
		if err == nil {
			proof.ExpandedText, err = gostaticstorage.ObserveX86AssemblyText(proof.ExpandedObject, builtin, source.request.Target)
		}
	}
	if err != nil || len(proof.OriginalText) != len(proof.ExpandedText) || len(proof.OriginalText) != len(source.file.Funcs) {
		return numericalSBContext("original/expanded complete TEXT differs: %v", err)
	}
	originalRows, err := numericalListingRows(proof.OriginalListing)
	if err != nil {
		return err
	}
	expandedRows, err := numericalListingRows(proof.ExpandedListing)
	if err != nil {
		return err
	}
	for index, fn := range source.file.Funcs {
		text, expandedText := proof.OriginalText[index], proof.ExpandedText[index]
		name := numericalNativeName(fn.Sym, source.request.PackagePath)
		if text.Name != name || text.ABI != 0 || expandedText.Name != name || expandedText.ABI != 0 ||
			!bytes.Equal(text.Code, expandedText.Code) || text.Size != expandedText.Size || !numericalRelocationEqual(text.Relocations, expandedText.Relocations) {
			return numericalSBContext("whole original/expanded TEXT/ABI/code/relocations differ")
		}
		memberMatches := 0
		for _, member := range proof.ArchiveMembers {
			if member.Name != strings.TrimSuffix(source.request.AssemblyFile, ".s")+".o" {
				continue
			}
			for _, record := range member.Texts {
				if record.Name == text.Name && record.ABI == text.ABI && record.Size == text.Size && bytes.Equal(record.Code, text.Code) && numericalRelocationEqual(record.Relocations, text.Relocations) {
					memberMatches++
				}
			}
		}
		if memberMatches != 1 {
			return numericalSBContext("original assembler TEXT does not match its actual selected archive member")
		}
		lines := strings.Split(source.cpp.Output, "\n")
		textLine := 0
		for line, physical := range lines {
			if strings.TrimSpace(physical) == fn.Instrs[0].Raw {
				if textLine != 0 {
					return numericalSBContext("ambiguous expanded physical TEXT identity")
				}
				textLine = line + 1
			}
		}
		if textLine == 0 {
			return numericalSBContext("missing expanded physical TEXT identity")
		}
		textEnd := len(lines) + 1
		for line, physical := range lines[textLine:] {
			for _, other := range source.file.Funcs {
				if strings.TrimSpace(physical) == other.Instrs[0].Raw {
					textEnd = textLine + line + 1
					break
				}
			}
			if textEnd != len(lines)+1 {
				break
			}
		}
		mode := 64
		if arch == "386" {
			mode = 32
		}
		// The full native stream must have trusted instruction boundaries,
		// including vector instructions and unreachable bytes. Only the
		// independent scalar decoder map can supply a numerical read below.
		_, decoded, err := x86NumericalNativeBoundaries(text.Code, mode)
		if err != nil {
			return err
		}
		for ordinal, ins := range fn.Instrs {
			containsSymbol := false
			for operandIndex, operand := range ins.Args {
				local, err := source.sdkLiteralRead(ins, operandIndex, arch)
				if err != nil {
					return err
				}
				containsSymbol = containsSymbol || operand.Kind == OpSym && !local
				if operand.Kind == OpMem && operand.Mem.Sym != "" {
					return numericalSBContext("indexed/alternate SB memory role needs separate binding")
				}
			}
			if !containsSymbol {
				continue
			}
			operandIndex, width, valid := x86NumericalSBReadShape(ins, arch)
			if !valid {
				return numericalSBContext("symbol address/write/unregistered numerical read form")
			}
			base, offset, valid := parseSBRef(ins.Args[operandIndex].Sym)
			if !valid || strings.Contains(base, "<") {
				return numericalSBContext("unknown source SB storage identity")
			}
			owner, exists := variables[numericalNativeName(base, source.request.PackagePath)]
			if !exists || !numericalOwnerContains(owner.variable, offset, width) {
				return numericalSBContext("read bytes lack exact numerical owner span")
			}
			matches := 0
			for pc, native := range decoded {
				for _, row := range expandedRows[name][pc] {
					if !numericalSameInstruction(ins, row.ins, source.request.PackagePath, arch) {
						continue
					}
					if source.literals != nil {
						if !numericalSDKLiteralReadOrdinal(fn, ordinal, sourcePositions[index], pc, row,
							expandedRows[name], source.request.PackagePath, arch) {
							continue
						}
					} else {
						// The expanded physical line must select this exact ordinal;
						// repeated identical instructions cannot borrow another row.
						separator := strings.LastIndexByte(row.position, ':')
						if separator < 0 || row.position[:separator] != expanded {
							return numericalSBContext("expanded listing source role differs")
						}
						line, err := strconv.Atoi(row.position[separator+1:])
						if err != nil || line <= textLine || line >= textEnd || strings.TrimSpace(lines[line-1]) != ins.Raw {
							return numericalSBContext("expanded source ordinal/physical line differs")
						}
						priorCount := 0
						for _, prior := range fn.Instrs[:ordinal] {
							if prior.Raw == ins.Raw {
								priorCount++
							}
						}
						physicalCount := 0
						for _, physical := range lines[textLine : line-1] {
							if strings.TrimSpace(physical) == ins.Raw {
								physicalCount++
							}
						}
						if priorCount != physicalCount {
							continue
						}
					}
					originalMatches := 0
					for _, original := range originalRows[name][pc] {
						if numericalSameInstruction(ins, original.ins, source.request.PackagePath, arch) {
							for _, position := range text.Positions {
								path := original.position
								objectFile := position.File
								if source.literals != nil {
									var valid bool
									path, valid = numericalSDKListingLogicalPosition(path)
									if !valid {
										return numericalSBContext("unknown SDK original source-position display")
									}
									objectFile, valid = source.sdkLiteralObjectSource(objectFile)
									if !valid {
										continue
									}
								}
								colon := strings.LastIndexByte(path, ':')
								line, err := strconv.Atoi(path[colon+1:])
								if colon >= 0 && err == nil && position.Start <= pc && pc < position.End && objectFile == path[:colon] && int32(line) == position.Line {
									for _, input := range source.cpp.Sources {
										if input.File == objectFile {
											originalMatches++
										}
									}
								}
							}
						}
					}
					if originalMatches != 1 || native.MemBytes != int(width) || native.AddrSize != mode {
						var locations []string
						for _, position := range text.Positions {
							if position.Start <= pc && pc < position.End {
								locations = append(locations, fmt.Sprintf("%s:%d", position.File, position.Line))
							}
						}
						return numericalSBContext("native read width/source-PC role differs: source matches=%d width=%d/%d address=%d/%d listing=%s binary=%v", originalMatches, native.MemBytes, width, native.AddrSize, mode, row.position, locations)
					}
					for _, prefix := range native.Prefix {
						value := uint8(prefix)
						if value != 0 && value != 0x66 && !(arch == "amd64" && value >= 0x40 && value <= 0x4f) {
							return numericalSBContext("read has an unregistered native prefix role")
						}
					}
					memory, memories := x86asm.Mem{}, 0
					for _, argument := range native.Args {
						if value, ok := argument.(x86asm.Mem); ok {
							memory, memories = value, memories+1
						}
					}
					if memories != 1 || memory.Segment != 0 || memory.Index != 0 || memory.Scale != 0 ||
						arch == "amd64" && memory.Base != x86asm.RIP || arch == "386" && memory.Base != 0 {
						return numericalSBContext("read has unbound segment/index/base role")
					}
					relocation, physicalOffset, err := numericalInstructionRelocation(text, pc, native, arch)
					if err != nil || physicalOffset != offset || relocation.Target.Name != owner.variable.NativeName {
						return numericalSBContext("whole native address/addend/owner differs: %v", err)
					}
					matches++
					read := owner
					read.function, read.instruction, read.original = name, ordinal, ins
					read.operand = operandIndex
					read.pc, read.offset, read.width, read.relocation = pc, offset, width, relocation
					read.code = append([]byte(nil), text.Code[pc:pc+int64(native.Len)]...)
					source.reads = append(source.reads, read)
				}
			}
			if matches != 1 {
				if source.literals != nil {
					var positions []string
					for pc, rows := range expandedRows[name] {
						for _, row := range rows {
							if numericalSameInstruction(ins, row.ins, source.request.PackagePath, arch) {
								positions = append(positions, fmt.Sprintf("%d@%s", pc, row.position))
							}
						}
					}
					return numericalSBContext("missing/ambiguous mapped source/native read: %s; expected %s, observed %v", ins.Raw, sourcePositions[index][ordinal], positions)
				}
				return numericalSBContext("missing/ambiguous whole source instruction/native read join: %s", ins.Raw)
			}
		}
	}
	if len(source.reads) == 0 {
		return numericalSBContext("source has no certified numerical SB reads")
	}
	proof.Sources = make(map[string][]byte)
	for name := range guard.inputs {
		if goAssemblyPathWithinRoot(source.request.Directory, name) || strings.HasPrefix(name, filepath.Join(guard.env["GOROOT"], "pkg", "include")+string(filepath.Separator)) {
			proof.Sources[name], err = guard.read(name)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
