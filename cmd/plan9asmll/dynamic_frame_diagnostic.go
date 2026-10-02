package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/xgo-dev/plan9asm"
)

// This is a failed-entry diagnostic, not signature inference or an exception.
// An omitted TEXT argument size and a zero-argument source declaration do not
// describe the caller-owned frame addressed by the assembly. Keep the original
// Context error and every ordinary source/translation applicability decision.
func annotateUnboundCallerFrame(file *plan9asm.File, resolve func(string) string,
	sigs map[string]plan9asm.FuncSig, declaredSizes map[string]int64, err error,
) error {
	if file == nil || resolve == nil || !errors.Is(err, plan9asm.ErrProbeNeedsContext) ||
		!strings.Contains(err.Error(), "bound typed frame storage") {
		return err
	}
	for _, fn := range file.Funcs {
		name := resolve(stripABISuffix(fn.Sym))
		sig, known := sigs[name]
		size, declared := declaredSizes[name]
		if !known || !declared || size != 0 || hasExplicitTextArgSize(fn) ||
			len(sig.Args) != 0 || sig.Ret != plan9asm.Void ||
			len(sig.Frame.Params) != 0 || len(sig.Frame.Results) != 0 ||
			!strings.HasPrefix(err.Error(), "translate: "+name+": ") {
			continue
		}
		for _, ins := range fn.Instrs {
			for _, operand := range ins.Args {
				if operand.Kind != plan9asm.OpFP && operand.Kind != plan9asm.OpFPAddr {
					continue
				}
				return fmt.Errorf("%w; actual source: empty Go declaration does not bind a dynamic caller frame; TEXT argument size is unspecified; local stack size, argument-only length and pointer bitmap are not a complete caller-frame contract", err)
			}
		}
	}
	return err
}
