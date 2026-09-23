package plan9asm

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// NativeData describes a Go global whose storage is supplied by native assembly.
type NativeData struct {
	Name string
	Size uint32
}

// ForeignARM64Functions identifies files containing only file-local TEXT symbols.
// This is a routing hint, NOT an inference of a C function signature. The native
// backend validates frames, flags, instructions, and references independently.
// Files with package-visible TEXT must use the typed backend and explicit Go
// declarations. Unsupported native forms must not fall back to signature guessing.
func ForeignARM64Functions(src []byte) map[string]bool {
	f, err := Parse(ArchARM64, string(src))
	if err != nil || len(f.Funcs) == 0 {
		return nil
	}
	result := make(map[string]bool)
	for _, fn := range f.Funcs {
		if !strings.HasSuffix(fn.Sym, "<>") {
			return nil
		}
		result[strings.TrimSuffix(fn.Sym, "<>")] = true
	}
	return result
}

// TranslateNativeARM64Source emits Darwin/ARM64 assembly directly from a bounded
// Plan 9 subset. It neither invokes cmd/asm nor reads Go object files. Entry and
// call registers are the physical registers written in the source; no function
// signature, Go stack adjustment, or register allocator is involved.
//
// All TEXT symbols must be file-local and NOSPLIT, with zero frame/argument sizes.
// Non-leaf code requires NOFRAME and must manage its C ABI frame itself. Only
// explicit registers, immediate/register arithmetic, base+offset memory operands,
// local branches, declared foreign calls and checked DATA/GLOBL are supported.
// See doc/native-arm64.md for the complete operand and directive contract.
func TranslateNativeARM64Source(src []byte, imports map[string]string, pkgPath string) (string, []NativeData, error) {
	clean, err := nativeSource(src)
	if err != nil {
		return "", nil, err
	}
	f, err := Parse(ArchARM64, clean)
	if err != nil {
		return "", nil, err
	}
	if len(f.Funcs) == 0 {
		return "", nil, fmt.Errorf("native assembly requires TEXT")
	}
	e := nativeARM64{funcs: map[string]string{}, globals: map[string]string{}, imports: imports}
	for name, alias := range imports {
		if !nativeName.MatchString(name) || !nativeName.MatchString(alias) {
			return "", nil, fmt.Errorf("unsupported native import %q -> %q", name, alias)
		}
	}
	for i, fn := range f.Funcs {
		name := strings.TrimSuffix(fn.Sym, "<>")
		if name == fn.Sym || !nativeName.MatchString(name) {
			return "", nil, fmt.Errorf("native TEXT must be file-local: %s", fn.Sym)
		}
		if _, ok := e.funcs[fn.Sym]; ok {
			return "", nil, fmt.Errorf("duplicate native TEXT %s", fn.Sym)
		}
		e.funcs[fn.Sym] = fmt.Sprintf("Lnative_func_%d", i)
		parts := strings.Split(fn.Instrs[0].Raw, ",")
		if len(parts) != 3 || (strings.TrimSpace(parts[2]) != "$0" && strings.TrimSpace(parts[2]) != "$0-0") {
			return "", nil, fmt.Errorf("native TEXT requires zero Go frame and arguments: %s", fn.Sym)
		}
		flags, err := nativeFlags(parts[1], "NOSPLIT", "NOFRAME")
		if err != nil {
			return "", nil, err
		}
		if !flags["NOSPLIT"] {
			return "", nil, fmt.Errorf("native TEXT requires NOSPLIT: %s", fn.Sym)
		}
		for _, ins := range fn.Instrs {
			if (ins.Op == "BL" || ins.Op == "CALL") && !flags["NOFRAME"] {
				return "", nil, fmt.Errorf("native non-leaf TEXT requires NOFRAME: %s", fn.Sym)
			}
		}
	}
	var data []NativeData
	for _, g := range f.Globl {
		name := strings.TrimPrefix(g.Sym, "·")
		if name == g.Sym || !nativeName.MatchString(name) {
			return "", nil, fmt.Errorf("native GLOBL must name a package global: %s", g.Sym)
		}
		if _, ok := e.globals[g.Sym]; ok {
			return "", nil, fmt.Errorf("duplicate native GLOBL %s", g.Sym)
		}
		if g.Size <= 0 || g.Size > 64<<20 {
			return "", nil, fmt.Errorf("unsupported native GLOBL size %d", g.Size)
		}
		if _, err := nativeFlags(g.Flags, "RODATA", "NOPTR"); err != nil {
			return "", nil, err
		}
		e.globals[g.Sym] = "_" + pkgPath + "." + name
		data = append(data, NativeData{pkgPath + "." + name, uint32(g.Size)})
	}
	var out strings.Builder
	for i, fn := range f.Funcs {
		e.labels = map[string]string{}
		for _, ins := range fn.Instrs {
			if ins.Op == OpLABEL {
				name := ins.Args[0].Sym
				if !nativeName.MatchString(name) {
					return "", nil, fmt.Errorf("unsupported native label %s", name)
				}
				if _, ok := e.labels[name]; ok {
					return "", nil, fmt.Errorf("duplicate native label %s", name)
				}
				e.labels[name] = fmt.Sprintf("Lnative_%d_label_%d", i, len(e.labels))
			}
		}
		fmt.Fprintf(&out, ".text\n.p2align 2\n%s:\n", e.funcs[fn.Sym])
		for _, ins := range fn.Instrs {
			if err := e.instruction(&out, ins); err != nil {
				return "", nil, fmt.Errorf("native %s: %s: %w", fn.Sym, ins.Raw, err)
			}
		}
	}
	for _, d := range f.Data {
		if _, ok := e.globals[d.Sym]; !ok {
			return "", nil, fmt.Errorf("native DATA has no GLOBL: %s", d.Sym)
		}
	}
	for _, g := range f.Globl {
		flags, _ := nativeFlags(g.Flags, "RODATA", "NOPTR")
		if flags["RODATA"] {
			out.WriteString(".section __DATA_CONST,__const\n")
		} else {
			out.WriteString(".data\n")
		}
		label := strconv.Quote(e.globals[g.Sym])
		fmt.Fprintf(&out, ".p2align 3\n.globl %s\n%s:\n", label, label)
		var values []DataStmt
		for _, d := range f.Data {
			if d.Sym == g.Sym {
				values = append(values, d)
			}
		}
		sort.Slice(values, func(i, j int) bool { return values[i].Off < values[j].Off })
		pos := int64(0)
		for _, d := range values {
			if d.Off < pos || d.Width <= 0 || d.Off > g.Size || d.Width > g.Size-d.Off {
				return "", nil, fmt.Errorf("overlapping or out-of-bounds native DATA for %s", g.Sym)
			}
			if d.Off > pos {
				fmt.Fprintf(&out, ".zero %d\n", d.Off-pos)
			}
			if d.Addr != "" {
				if d.Width != 8 {
					return "", nil, fmt.Errorf("native address DATA requires width 8")
				}
				target, err := e.symbol(d.Addr, true)
				if err != nil {
					return "", nil, err
				}
				fmt.Fprintf(&out, ".quad %s\n", target)
			} else {
				if d.Payload != nil {
					return "", nil, fmt.Errorf("native string DATA is unsupported")
				}
				directive := map[int64]string{1: ".byte", 2: ".short", 4: ".long", 8: ".quad"}[d.Width]
				if directive == "" {
					return "", nil, fmt.Errorf("unsupported native DATA width %d", d.Width)
				}
				mask := ^uint64(0)
				if d.Width < 8 {
					mask = (uint64(1) << (8 * d.Width)) - 1
				}
				fmt.Fprintf(&out, "%s %#x\n", directive, d.Value&mask)
			}
			pos = d.Off + d.Width
		}
		if pos < g.Size {
			fmt.Fprintf(&out, ".zero %d\n", g.Size-pos)
		}
	}
	return out.String(), data, nil
}

var nativeName = regexp.MustCompile(`^[A-Za-z_][A-Za-z_0-9]*$`)
var nativePseudoRegister = regexp.MustCompile(`\b(?:SP|FP|g|G|R18_PLATFORM|R18|R31|W[0-9]+)\b`)

// The general parser deliberately tolerates ignored includes and symbolic
// placeholders. The native backend must reject these before parsing, rather
// than silently emitting guessed offsets or dropping conditional code.
func nativeSource(src []byte) (string, error) {
	var b strings.Builder
	for _, line := range strings.Split(string(src), "\n") {
		line, _, _ = strings.Cut(line, "//")
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			if line != `#include "textflag.h"` {
				return "", fmt.Errorf("unsupported native preprocessor directive: %s", line)
			}
			continue
		}
		if strings.Contains(line, "/*") || strings.ContainsAny(line, "\"\\") {
			return "", fmt.Errorf("unsupported native source syntax: %s", line)
		}
		if nativePseudoRegister.MatchString(line) {
			return "", fmt.Errorf("unsupported native register or Go stack operand: %s", line)
		}
		for _, stmt := range splitSemicolons(line) {
			op, rest := splitOpcode(stmt)
			// Enforce integer expressions before the permissive parser can invent a
			// symbolic GLOBL size or interpret an integer operand as float bits.
			if op == "GLOBL" {
				parts := strings.Split(rest, ",")
				if len(parts) != 3 {
					return "", fmt.Errorf("invalid native GLOBL")
				}
				if _, ok := nativeInteger(strings.TrimSpace(parts[2])); !ok {
					return "", fmt.Errorf("native GLOBL requires a constant integer size")
				}
			}
			if op == "DATA" {
				_, rhs, ok := strings.Cut(rest, ",")
				rhs = strings.TrimSpace(rhs)
				if !ok {
					return "", fmt.Errorf("invalid native DATA")
				}
				if _, ok := nativeInteger(rhs); !ok && !(strings.HasPrefix(rhs, "$") && strings.HasSuffix(rhs, "(SB)")) {
					return "", fmt.Errorf("unsupported native DATA initializer: %s", rhs)
				}
			}
		}
		b.WriteString(line)
		b.WriteByte('\n')
	}
	return b.String(), nil
}
func nativeInteger(s string) (uint64, bool) {
	if !strings.HasPrefix(s, "$") {
		return 0, false
	}
	return parseImmExpr(strings.TrimPrefix(s, "$"))
}
func nativeFlags(s string, allowed ...string) (map[string]bool, error) {
	flags := map[string]bool{}
	for _, flag := range strings.Split(strings.TrimSpace(s), "|") {
		flag = strings.TrimSpace(flag)
		if flag == "0" {
			continue
		}
		ok := false
		for _, a := range allowed {
			ok = ok || flag == a
		}
		if !ok || flags[flag] {
			return nil, fmt.Errorf("unsupported native flag %q", flag)
		}
		flags[flag] = true
	}
	return flags, nil
}

type nativeARM64 struct{ funcs, globals, imports, labels map[string]string }

func (e *nativeARM64) symbol(s string, data bool) (string, error) {
	if !strings.HasSuffix(s, "(SB)") {
		return "", fmt.Errorf("unsupported native symbol %s", s)
	}
	name := strings.TrimSuffix(s, "(SB)")
	if label, ok := e.funcs[name]; ok {
		return label, nil
	}
	if data {
		if label, ok := e.globals[name]; ok {
			return strconv.Quote(label), nil
		}
	}
	if alias, ok := e.imports[name]; ok {
		return strconv.Quote("_" + alias), nil
	}
	return "", fmt.Errorf("undeclared foreign symbol or unsupported native reference %s", s)
}
func nativeReg(o Operand, bits int, sp bool) (string, error) {
	if o.Kind != OpReg {
		return "", fmt.Errorf("expected native register")
	}
	if o.Reg == SP && sp && bits == 64 {
		return "sp", nil
	}
	if o.Reg == ZR {
		if bits == 32 {
			return "wzr", nil
		}
		return "xzr", nil
	}
	name := string(o.Reg)
	if strings.HasPrefix(name, "R") {
		if n, err := strconv.Atoi(name[1:]); err == nil && n >= 0 && n <= 30 && n != 18 {
			prefix := "x"
			if bits == 32 {
				prefix = "w"
			}
			return fmt.Sprintf("%s%d", prefix, n), nil
		}
	}
	return "", fmt.Errorf("unsupported native register %s", o.Reg)
}
func nativeFPReg(o Operand) (string, error) {
	if o.Kind == OpReg && strings.HasPrefix(string(o.Reg), "F") {
		if n, err := strconv.Atoi(string(o.Reg)[1:]); err == nil && n >= 0 && n < 32 {
			return fmt.Sprintf("d%d", n), nil
		}
	}
	return "", fmt.Errorf("expected native floating register")
}
func nativeMemory(o Operand, width int) (string, bool, error) {
	m := o.Mem
	if o.Kind != OpMem || m.Sym != "" || m.OffRaw != "" || m.Index != "" || m.Segment != "" {
		return "", false, fmt.Errorf("native memory requires a constant offset and one base register")
	}
	base, err := nativeReg(Operand{Kind: OpReg, Reg: m.Base}, 64, true)
	if err != nil || base == "xzr" {
		return "", false, fmt.Errorf("invalid native memory base")
	}
	unscaled := false
	if m.Off < 0 || m.Off%int64(width) != 0 || m.Off/int64(width) > 4095 {
		if m.Off < -256 || m.Off > 255 {
			return "", false, fmt.Errorf("native memory offset requires unsupported address expansion")
		}
		unscaled = true
	}
	return fmt.Sprintf("[%s, #%d]", base, m.Off), unscaled, nil
}
func (e *nativeARM64) branch(o Operand, external bool) (string, error) {
	if external && o.Kind == OpSym {
		return e.symbol(o.Sym, false)
	}
	name := o.Ident
	if o.Kind == OpLabel {
		name = o.Sym
	}
	if label, ok := e.labels[name]; ok {
		return label, nil
	}
	return "", fmt.Errorf("undefined native branch label %s", o.String())
}
func (e *nativeARM64) instruction(out *strings.Builder, ins Instr) error {
	a := ins.Args
	op := string(ins.Op)
	if ins.Op != OpTEXT && ins.Op != OpLABEL {
		_, rest := splitOpcode(ins.Raw)
		parts := splitTopLevelCSV(rest)
		if len(parts) != len(a) {
			return fmt.Errorf("invalid native operand list")
		}
		for i, operand := range a {
			if operand.Kind == OpImm {
				if _, ok := nativeInteger(strings.TrimSpace(parts[i])); !ok {
					return fmt.Errorf("native operand requires a constant integer")
				}
			}
		}
	}

	emit := func(op string, args ...string) { fmt.Fprintf(out, "\t%s %s\n", op, strings.Join(args, ", ")) }
	bad := func() error { return fmt.Errorf("unsupported native operand form for %s", op) }
	switch op {
	case "TEXT":
		return nil
	case "LABEL":
		fmt.Fprintf(out, "%s:\n", e.labels[a[0].Sym])
		return nil
	case "RET":
		if len(a) != 0 {
			return bad()
		}
		emit("ret")
		return nil
	case "JMP", "B", "BL", "CALL":
		if len(a) != 1 {
			return bad()
		}
		dst, err := e.branch(a[0], true)
		if err != nil {
			return err
		}
		inst := "b"
		if op == "BL" || op == "CALL" {
			inst = "bl"
		}
		emit(inst, dst)
		return nil
	case "BEQ", "BNE", "BLT", "BLE", "BGT", "BGE", "BHS", "BLO", "BHI", "BLS", "BMI", "BPL", "BVS", "BVC":
		if len(a) != 1 {
			return bad()
		}
		dst, err := e.branch(a[0], false)
		if err != nil {
			return err
		}
		emit("b."+strings.ToLower(op[1:]), dst)
		return nil
	case "CBZ", "CBNZ":
		if len(a) != 2 {
			return bad()
		}
		reg, err := nativeReg(a[0], 64, false)
		if err != nil {
			return err
		}
		dst, err := e.branch(a[1], false)
		if err != nil {
			return err
		}
		emit(strings.ToLower(op), reg, dst)
		return nil
	case "MOVD", "MOVW", "MOVWU":
		if len(a) != 2 {
			return bad()
		}
		width := 8
		if op != "MOVD" {
			width = 4
		}
		if a[0].Kind == OpMem || a[1].Kind == OpMem {
			mem, reg := a[0], a[1]
			load := true
			if a[1].Kind == OpMem {
				mem, reg = a[1], a[0]
				load = false
			}
			bits := width * 8
			if load && op == "MOVW" {
				bits = 64
			}
			r, err := nativeReg(reg, bits, false)
			if err != nil {
				return err
			}
			addr, unscaled, err := nativeMemory(mem, width)
			if err != nil {
				return err
			}
			inst := "ldr"
			if !load {
				inst = "str"
			}
			if unscaled {
				if load {
					inst = "ldur"
				} else {
					inst = "stur"
				}
			}
			if load && op == "MOVW" {
				inst += "sw"
			}
			emit(inst, r, addr)
			return nil
		}
		dst, err := nativeReg(a[1], 64, op == "MOVD")
		if err != nil {
			return err
		}
		if a[0].Kind == OpImm {
			if op != "MOVD" || dst == "sp" || a[0].ImmRaw != "" {
				return bad()
			}
			// MOVD constant expansion uses only the destination; it never clobbers
			// a hidden scratch register or the condition flags.
			v := uint64(a[0].Imm)
			emit("movz", dst, fmt.Sprintf("#%d", v&65535))
			for shift := 16; shift < 64; shift += 16 {
				if half := (v >> shift) & 65535; half != 0 {
					emit("movk", dst, fmt.Sprintf("#%d", half), fmt.Sprintf("lsl #%d", shift))
				}
			}
			return nil
		}
		bits := 64
		if op != "MOVD" {
			bits = 32
		}
		src, err := nativeReg(a[0], bits, op == "MOVD")
		if err != nil {
			return err
		}
		if op == "MOVW" {
			emit("sxtw", dst, src)
		} else if op == "MOVWU" {
			emit("uxtw", dst, src)
		} else {
			if (src == "sp" && dst == "xzr") || (src == "xzr" && dst == "sp") {
				return bad()
			}
			emit("mov", dst, src)
		}
		return nil
	case "FMOVD":
		if len(a) != 2 {
			return bad()
		}
		src, se := nativeFPReg(a[0])
		dst, de := nativeFPReg(a[1])
		if se != nil && de != nil {
			return bad()
		}
		if se != nil {
			src, se = nativeReg(a[0], 64, false)
		}
		if de != nil {
			dst, de = nativeReg(a[1], 64, false)
		}
		if se != nil || de != nil {
			return bad()
		}
		emit("fmov", dst, src)
		return nil
	case "ADD", "SUB", "CMP", "CMPW", "LSL":
		cmp := op == "CMP" || op == "CMPW"
		if len(a) != 2 && (cmp || len(a) != 3) {
			return bad()
		}
		bits := 64
		if op == "CMPW" {
			bits = 32
		}
		dstOp := a[len(a)-1]
		srcOp := dstOp
		if len(a) == 3 {
			srcOp = a[1]
		}
		allowSP := op == "ADD" || op == "SUB" || cmp
		src, err := nativeReg(srcOp, bits, allowSP)
		if err != nil {
			return err
		}
		dst, err := nativeReg(dstOp, bits, allowSP)
		if err != nil {
			return err
		}
		var rhs string
		if a[0].Kind == OpImm && a[0].ImmRaw == "" {
			limit := int64(4095)
			if op == "LSL" {
				limit = 63
			}
			if a[0].Imm < 0 || a[0].Imm > limit {
				return bad()
			}
			rhs = fmt.Sprintf("#%d", a[0].Imm)
			if src == "xzr" || src == "wzr" || (!cmp && dst == "xzr") {
				return bad()
			}
		} else {
			rhs, err = nativeReg(a[0], bits, false)
			if err != nil {
				return err
			}
			// Register ADD/SUB with SP needs explicit extended-register encoding.
			// Reject it rather than selecting an alias with different register-31 semantics.
			if src == "sp" || dst == "sp" {
				return bad()
			}
		}
		if cmp {
			emit("cmp", src, rhs)
		} else {
			emit(strings.ToLower(op), dst, src, rhs)
		}
		return nil
	}
	return fmt.Errorf("unsupported native instruction %s", op)
}
