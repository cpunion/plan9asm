package plan9asm

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Layout constants from Go 1.27 cmd/internal/goobj/objfile.go (go120ld)
// and cmd/internal/objabi/reloctype.go. Reject other object formats below.
const (
	nativeBlockCount    = 19 // includes the end offset
	nativeHeaderSize    = 20 + nativeBlockCount*4
	nativeSymSize       = 21
	nativeSymFlags      = 11
	nativeSymSizeOffset = 13
	nativeSymAlign      = 17
	nativeRelocSize     = 23
	nativeRelocKind     = 5
	nativeRelocAddend   = 7
	nativeRelocPkg      = 15
	nativeRelocSym      = 19
	nativeNoSplit       = 1 << 4
	nativeRAddr         = 1
	nativeRCallARM64    = 9
	nativeBlkSymdef     = 3
	nativeBlkNonpkgref  = 7
	nativeBlkRelocIdx   = 11
	nativeBlkDataIdx    = 13
	nativeBlkReloc      = 14
	nativeBlkData       = 16
	nativePkgSelf       = 0x7ffffffb
	nativePkgHashed64   = 0x7ffffffe
	nativePkgHashed     = 0x7ffffffd
	nativePkgNone       = 0x7fffffff
)

// NativeData describes a Go global whose storage is supplied by native assembly.
type NativeData struct {
	Name string
	Size uint32
}

// TranslateNativeARM64Object preserves the machine code of foreign-ABI assembly. Unlike
// Go-declared functions, raw callbacks have no signature from which LLVM can
// reconstruct incoming registers. The Go assembler encodes those registers;
// this bridge only converts its address and branch relocations to Mach-O asm.
//
// The input must be a darwin/arm64 cmd/asm object in go120ld format. Only explicitly
// selected functions, their data, and declared dynamic imports are accepted.
// Unknown formats, symbol references and relocations are errors.
func TranslateNativeARM64Object(obj []byte, funcs map[string]bool, imports map[string]string, pkgPath string) (string, []NativeData, error) {
	r, err := readNativeObject(obj)
	if err != nil {
		return "", nil, err
	}
	labels := make(map[int]string)
	var data []NativeData
	for i, s := range r.syms[:r.ndef] {
		if funcs[s.name] {
			if s.flags&nativeNoSplit == 0 {
				return "", nil, fmt.Errorf("native function %s is not NOSPLIT", s.name)
			}
			labels[i] = fmt.Sprintf("Lllgo_native_%d", i)
		} else if strings.HasPrefix(s.name, pkgPath+".") {
			labels[i] = "_" + s.name
			data = append(data, NativeData{s.name, s.size})
		}
	}
	for name := range funcs {
		found := false
		for i := range labels {
			if r.syms[i].name == name {
				found = true
				break
			}
		}
		if !found {
			return "", nil, fmt.Errorf("missing native function %s", name)
		}
	}
	var out strings.Builder
	for i, s := range r.syms[:r.ndef] {
		label, ok := labels[i]
		if !ok {
			continue
		}
		code := r.payload(i)
		if uint64(len(code)) > uint64(s.size) {
			return "", nil, fmt.Errorf("oversized native symbol %s", s.name)
		}
		code = append(append([]byte(nil), code...), make([]byte, int(s.size)-len(code))...)
		if funcs[s.name] {
			out.WriteString(".text\n.p2align 2\n")
		} else {
			out.WriteString(".data\n")
			fmt.Fprintf(&out, ".globl %s\n", strconv.Quote(label))
			align := s.align
			if align == 0 {
				align = 8
			}
			if align&(align-1) != 0 {
				return "", nil, fmt.Errorf("invalid native alignment %d", align)
			}
			fmt.Fprintf(&out, ".balign %d\n", align)
		}
		fmt.Fprintf(&out, "%s:\n", strconv.Quote(label))
		rels := r.relocs(i)
		sort.Slice(rels, func(i, j int) bool { return rels[i].off < rels[j].off })
		pos := uint32(0)
		for _, rel := range rels {
			if rel.off < pos || uint64(rel.off)+uint64(rel.size) > uint64(len(code)) {
				return "", nil, fmt.Errorf("invalid native relocation in %s", s.name)
			}
			target, err := r.target(rel.pkg, rel.sym)
			if err != nil {
				return "", nil, err
			}
			dst, ok := labels[target]
			if !ok {
				alias, ok := imports[r.syms[target].name]
				if !ok {
					return "", nil, fmt.Errorf("native assembly references undeclared foreign symbol %s", r.syms[target].name)
				}
				dst = "_" + alias
			}
			expr := strconv.Quote(dst)
			if rel.add != 0 {
				expr += fmt.Sprintf("%+d", rel.add)
			}
			nativeBytes(&out, code[pos:rel.off])
			switch {
			case rel.kind == nativeRAddr && rel.size == 8: // R_ADDR
				fmt.Fprintf(&out, ".quad %s\n", expr)
			case rel.kind == nativeRCallARM64 && rel.size == 4 && funcs[s.name]: // R_CALLARM64: B or BL
				if rel.add != 0 {
					return "", nil, fmt.Errorf("unsupported native branch addend %d in %s", rel.add, s.name)
				}
				word := binary.LittleEndian.Uint32(code[rel.off:])
				op := "b"
				if word&0xfc000000 == 0x94000000 {
					op = "bl"
				} else if word&0xfc000000 != 0x14000000 {
					return "", nil, fmt.Errorf("invalid native branch in %s", s.name)
				}
				fmt.Fprintf(&out, "%s %s\n", op, expr)
			default:
				return "", nil, fmt.Errorf("unsupported native relocation %d/%d in %s", rel.kind, rel.size, s.name)
			}
			pos = rel.off + uint32(rel.size)
		}
		nativeBytes(&out, code[pos:])
	}
	return out.String(), data, nil
}

func nativeBytes(out *strings.Builder, b []byte) {
	var digits [3]byte // largest byte value is 255
	for len(b) > 0 {
		n := len(b)
		if n > 16 {
			n = 16
		}
		out.WriteString(".byte ")
		for i, x := range b[:n] {
			if i != 0 {
				out.WriteByte(',')
			}
			out.Write(strconv.AppendUint(digits[:0], uint64(x), 10))
		}
		out.WriteByte('\n')
		b = b[n:]
	}
}

type nativeSym struct {
	name        string
	size, align uint32
	flags       byte
}
type nativeReloc struct {
	off      uint32
	size     byte
	kind     uint16
	add      int64
	pkg, sym uint32
}
type nativeObject struct {
	b      []byte
	blocks [nativeBlockCount]uint32
	syms   []nativeSym
	ndef   int
	counts [5]int
}

func readNativeObject(obj []byte) (*nativeObject, error) {
	// cmd/asm prefixes the binary object with a target/version text header.
	h := bytes.Index(obj, []byte("\n!\n"))
	if h < 0 || !bytes.HasPrefix(obj, []byte("go object darwin arm64 ")) {
		return nil, fmt.Errorf("expected darwin/arm64 Go assembler object")
	}
	b := obj[h+3:]
	if len(b) < nativeHeaderSize || string(b[:8]) != "\x00go120ld" {
		return nil, fmt.Errorf("unsupported Go assembler object format")
	}
	r := &nativeObject{b: b}
	for i := range r.blocks {
		r.blocks[i] = binary.LittleEndian.Uint32(b[20+i*4:])
		if r.blocks[i] > uint32(len(b)) || (i > 0 && r.blocks[i] < r.blocks[i-1]) {
			return nil, fmt.Errorf("invalid native object block %d", i)
		}
	}
	if r.blocks[0] < nativeHeaderSize {
		return nil, fmt.Errorf("invalid native object header")
	}
	for block := nativeBlkSymdef; block <= nativeBlkNonpkgref; block++ {
		buf := r.block(block)
		if len(buf)%nativeSymSize != 0 {
			return nil, fmt.Errorf("invalid native symbol table")
		}
		r.counts[block-nativeBlkSymdef] = len(buf) / nativeSymSize
		for len(buf) > 0 {
			s := buf[:nativeSymSize]
			n, off := binary.LittleEndian.Uint32(s), binary.LittleEndian.Uint32(s[4:])
			if uint64(off)+uint64(n) > uint64(len(b)) {
				return nil, fmt.Errorf("invalid native symbol name")
			}
			size := binary.LittleEndian.Uint32(s[nativeSymSizeOffset:])
			if size > 64<<20 {
				return nil, fmt.Errorf("native symbol too large")
			}
			r.syms = append(r.syms, nativeSym{string(b[off : off+n]), size, binary.LittleEndian.Uint32(s[nativeSymAlign:]), s[nativeSymFlags]})
			buf = buf[nativeSymSize:]
		}
	}
	r.ndef = len(r.syms) - r.counts[4]
	if len(r.block(nativeBlkRelocIdx)) != (r.ndef+1)*4 || len(r.block(nativeBlkDataIdx)) != (r.ndef+1)*4 || len(r.block(nativeBlkReloc))%nativeRelocSize != 0 {
		return nil, fmt.Errorf("invalid native object indices")
	}
	for _, pair := range [][2]int{{nativeBlkRelocIdx, len(r.block(nativeBlkReloc)) / nativeRelocSize}, {nativeBlkDataIdx, len(r.block(nativeBlkData))}} {
		prev := uint32(0)
		buf := r.block(pair[0])
		for len(buf) > 0 {
			v := binary.LittleEndian.Uint32(buf)
			if v < prev || uint64(v) > uint64(pair[1]) {
				return nil, fmt.Errorf("invalid native object index")
			}
			prev = v
			buf = buf[4:]
		}
	}
	return r, nil
}

// block accepts only the block constants above, all strictly before the end
// offset. readNativeObject validates every offset before calling it.
func (r *nativeObject) block(i int) []byte { return r.b[r.blocks[i]:r.blocks[i+1]] }
func (r *nativeObject) payload(i int) []byte {
	idx := r.block(nativeBlkDataIdx)
	return r.block(nativeBlkData)[binary.LittleEndian.Uint32(idx[i*4:]):binary.LittleEndian.Uint32(idx[(i+1)*4:])]
}
func (r *nativeObject) relocs(i int) []nativeReloc {
	idx := r.block(nativeBlkRelocIdx)
	a, z := binary.LittleEndian.Uint32(idx[i*4:]), binary.LittleEndian.Uint32(idx[(i+1)*4:])
	var out []nativeReloc
	for n := a; n < z; n++ {
		v := r.block(nativeBlkReloc)[int(n)*nativeRelocSize:]
		out = append(out, nativeReloc{binary.LittleEndian.Uint32(v), v[4], binary.LittleEndian.Uint16(v[nativeRelocKind:]), int64(binary.LittleEndian.Uint64(v[nativeRelocAddend:])), binary.LittleEndian.Uint32(v[nativeRelocPkg:]), binary.LittleEndian.Uint32(v[nativeRelocSym:])})
	}
	return out
}
func (r *nativeObject) target(pkg, sym uint32) (int, error) {
	// Reserved package indices from cmd/internal/goobj. Package imports and
	// runtime builtins have Go ABI and cannot be called by these foreign stubs.
	var start, count int
	switch pkg {
	case nativePkgSelf:
		count = r.counts[0]
	case nativePkgHashed64:
		start = r.counts[0]
		count = r.counts[1]
	case nativePkgHashed:
		start = r.counts[0] + r.counts[1]
		count = r.counts[2]
	case nativePkgNone:
		start = r.counts[0] + r.counts[1] + r.counts[2]
		count = r.counts[3] + r.counts[4]
	default:
		return 0, fmt.Errorf("native assembly references Go package index %#x", pkg)
	}
	if uint64(sym) >= uint64(count) {
		return 0, fmt.Errorf("invalid native symbol reference")
	}
	return start + int(sym), nil
}

var nativeTextRE = regexp.MustCompile(`(?m)^\s*TEXT\s+([^\s(),]+)<>\(SB\),\s*([^,]+),\s*\$0(?:-0)?\s*(?://[^\n]*)?$`)

// ForeignARM64Functions selects files made entirely of raw foreign-ABI
// callbacks/trampolines. Go-declared functions continue through typed LLVM
// lowering. NOSPLIT and zero Go frames are required; the callback may manage
// its own native frame explicitly with NOFRAME.
func ForeignARM64Functions(src []byte) map[string]bool {
	matches := nativeTextRE.FindAllSubmatchIndex(src, -1)
	if len(matches) == 0 || len(reTextLines.FindAll(src, -1)) != len(matches) {
		return nil
	}
	result := make(map[string]bool)
	for i, m := range matches {
		flags := string(src[m[4]:m[5]])
		if !hasAsmFlag(flags, "NOSPLIT") {
			return nil
		}
		end := len(src)
		if i+1 < len(matches) {
			end = matches[i+1][0]
		}
		if !hasAsmFlag(flags, "NOFRAME") && nativeCallRE.Match(src[m[1]:end]) {
			return nil
		}
		result[string(src[m[2]:m[3]])] = true
	}
	return result
}

var reTextLines = regexp.MustCompile(`(?m)^\s*TEXT\b`)
var nativeCallRE = regexp.MustCompile(`(?m)^\s*(?:CALL|BL)\s`)

func hasAsmFlag(flags, want string) bool {
	for _, f := range strings.Split(flags, "|") {
		if strings.TrimSpace(f) == want {
			return true
		}
	}
	return false
}
