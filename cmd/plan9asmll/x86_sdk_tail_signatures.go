package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/types"
	"strings"

	"github.com/xgo-dev/plan9asm"
)

// This is an inventory adapter, not a source or entry producer. The live SDK
// binder subsequently authenticates every original declaration/node and the
// direct-tail graph proves each root's state and terminal compatibility. Generic
// CLI fallback signatures remain unchanged outside this complete-source route.
func x86SDKDirectTailDeclarationOptions(pkg plan9asm.GoPackage, file *plan9asm.File, options plan9asm.Options) (plan9asm.Options, error) {
	needed := false
	if file != nil && file.Arch == plan9asm.ArchAMD64 {
		for _, fn := range file.Funcs {
			for _, ins := range fn.Instrs {
				if ins.Op == plan9asm.OpJMP && len(ins.Args) == 1 && strings.Contains(ins.Args[0].Sym, "<>") {
					needed = true // Cost query only; no operand/target is admitted here.
				}
			}
		}
	}
	if !needed {
		return options, nil
	}
	contextError := func(reason string) error {
		return fmt.Errorf("%w: SDK numerical tail declarations: %s", plan9asm.ErrProbeNeedsContext, reason)
	}
	if pkg.Types == nil || pkg.Types.Path() != pkg.Path || options.ResolveSym == nil ||
		(options.Goarch != "386" && options.Goarch != "amd64") {
		return plan9asm.Options{}, contextError("missing actual typed package/resolver/mode")
	}
	sizes := types.SizesFor("gc", options.Goarch)
	linknames := linknameRemoteToLocal(pkg.Syntax)
	declared := make(map[string]plan9asm.FuncSig)
	helpers := make(map[string]bool)
	texts := make(map[string]bool)
	add := func(symbol string) error {
		logical := stripABISuffix(symbol)
		name := options.ResolveSym(logical)
		if name == "" {
			return contextError("empty original declaration identity")
		}
		actual, _, present, err := tryDeclSig(pkg.Types.Scope(), logical, name, linknames, options.Goarch, sizes)
		if err != nil || !present {
			return contextError("root/ordinary terminal lacks its actual Go declaration: " + logical)
		}
		provided, present := options.Sigs[name]
		// JSON is solely an equality projection. Keep the original value, including
		// its private caller/local capabilities; never reconstruct one from bytes.
		before, actualErr := json.Marshal(actual)
		after, providedErr := json.Marshal(provided)
		if !present || actualErr != nil || providedErr != nil || !bytes.Equal(before, after) {
			return contextError("supplied root/terminal transport differs from its actual declaration: " + name)
		}
		declared[name] = provided
		return nil
	}
	for _, fn := range file.Funcs {
		if texts[fn.Sym] {
			return plan9asm.Options{}, contextError("duplicate original TEXT")
		}
		texts[fn.Sym] = true
		if strings.HasSuffix(fn.Sym, "<>") {
			helpers[options.ResolveSym(fn.Sym)] = true
			continue
		}
		if err := add(fn.Sym); err != nil {
			return plan9asm.Options{}, err
		}
	}
	for _, fn := range file.Funcs {
		for _, ins := range fn.Instrs {
			if ins.Op != plan9asm.OpJMP || len(ins.Args) != 1 || ins.Args[0].Kind != plan9asm.OpSym {
				continue
			}
			operand := strings.TrimSpace(ins.Args[0].Sym)
			if !strings.HasSuffix(operand, "(SB)") {
				continue // Local source labels retain the graph's separate CFG audit.
			}
			symbol, offset := splitSymPlusOff(strings.TrimSuffix(operand, "(SB)"))
			if symbol == "" || offset != 0 {
				return plan9asm.Options{}, contextError("offset/unknown ordinary terminal")
			}
			if strings.HasSuffix(symbol, "<>") {
				if !texts[symbol] {
					return plan9asm.Options{}, contextError("private terminal lacks its complete original TEXT")
				}
				continue
			}
			if err := add(symbol); err != nil {
				return plan9asm.Options{}, err
			}
		}
	}
	for name := range options.Sigs {
		if _, present := declared[name]; !present && !helpers[name] {
			return plan9asm.Options{}, contextError("unassociated callable signature: " + name)
		}
	}
	options.Sigs = declared
	return options, nil
}
