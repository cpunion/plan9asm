package plan9asm

import (
	"path/filepath"
	"reflect"
)

// Fixed-local physical SP and its initialized result already have a complete
// native source/CFG proof. Revalidate that exact role instead of reinterpreting
// its private access annotations as logical AUTO bytes. A proof pointer alone
// is not permission, and this route supplies no new native/caller capability.
func (c *amd64Ctx) checkFixedLocalResultInitialization() error {
	fn, sig, proof := c.fpScalarSource, c.fpSourceSig, c.fixedLocal
	binding := sig.x86FixedLocal
	fail := func(reason string) error { return goAMD64FrameContext("fixed local result consumer: %s", reason) }
	if binding == nil || proof == nil || fn.x86FixedLocal != proof || proof.binding != binding ||
		c.goarch != "amd64" || c.targetTriple != binding.target || fn.FrameSize != proof.frame || c.frameSize != fn.FrameSize {
		return fail("source/proof/target/frame role differs")
	}
	if err := binding.verify(); err != nil {
		return err
	}
	emitted := c.sig
	emitted.Attrs = sig.Attrs // Checked against the actual complete renderer inventory below.
	if !reflect.DeepEqual(emitted, sig) {
		return fail("emitted logical signature differs from the original declaration")
	}
	expected, declared := binding.signatures[binding.symbols[fn.Sym]]
	sig.x86FixedLocal = nil
	if !declared || !reflect.DeepEqual(sig, expected) {
		return fail("actual declaration/signature differs")
	}
	sig.x86FixedLocal = binding
	// Replay the actual guarded CPP, not a mutually consistent replacement of
	// the binding's private assembly and fingerprint. No caller callback runs
	// under the source mutex; the complete live Verify guards surround this.
	binding.source.mu.Lock()
	original, err := binding.source.read(filepath.Join(binding.source.request.Directory, binding.source.request.AssemblyFile))
	var expanded string
	if err == nil {
		expanded, err = binding.source.preprocess(binding.source.request.AssemblyFile, original)
	}
	binding.source.mu.Unlock()
	if err != nil || expanded != binding.assembly {
		return fail("expanded assembly differs from the actual guarded source")
	}
	actual, err := Parse(ArchAMD64, expanded)
	if err != nil {
		return fail("complete original assembly cannot be parsed")
	}
	fingerprint, err := x86CallerSourceFingerprint(actual)
	if err != nil || fingerprint != binding.file {
		return fail("complete original assembly fingerprint differs")
	}
	attributes, inferred, valid := fixedLocalResultFeatureAttrs(actual, binding, sig.Name, c.goarch)
	if !valid || c.sig.Attrs != attributes || c.fpFeatureAttrs != inferred {
		return fail("emitted attributes are not the exact complete-source renderer projection")
	}
	fingerprint, err = x86CallerSourceFingerprint(fn)
	if err != nil || fingerprint != proof.function {
		return fail("complete normalized function fingerprint differs")
	}
	plain := fn
	plain.Instrs = append([]Instr(nil), fn.Instrs...)
	for ordinal, ins := range fn.Instrs {
		if len(ins.x86FixedLocalMemory) != len(proof.accesses[ordinal]) {
			return fail("instruction access inventory differs")
		}
		for operand, access := range ins.x86FixedLocalMemory {
			if access.ordinal != ordinal || !c.fixedLocalMemoryCertified(ins, operand) {
				return fail("instruction access owner/shape/ordinal differs")
			}
		}
		plain.Instrs[ordinal].x86FixedLocalMemory = nil
	}
	for ordinal := range proof.accesses {
		if ordinal < 0 || ordinal >= len(fn.Instrs) {
			return fail("access ordinal is outside the complete function")
		}
	}
	instructions, err := x86ResultAUTOReplayInstructions(plain, c.goarch)
	if err != nil {
		return err
	}
	matches := 0
	for _, original := range actual.Funcs {
		if original.Sym == fn.Sym {
			if original.FrameSize != fn.FrameSize || original.ArgSize != fn.ArgSize ||
				!reflect.DeepEqual(original.Instrs, instructions) {
				return fail("original TEXT/source instruction roles differ")
			}
			matches++
		}
	}
	if matches != 1 {
		return fail("original TEXT is missing or ambiguous")
	}
	replayed, err := proveX86FixedLocal(plain, sig)
	if err != nil {
		return err
	}
	for ordinal, accesses := range proof.accesses {
		if len(accesses) != len(replayed.x86FixedLocal.accesses[ordinal]) {
			return fail("recomputed access inventory differs")
		}
		for operand, access := range accesses {
			recomputed, found := replayed.x86FixedLocal.accesses[ordinal][operand]
			access.owner, recomputed.owner = nil, nil // Compare effects, not new builder identity.
			if !found || !reflect.DeepEqual(access, recomputed) {
				return fail("recomputed access byte effects differ")
			}
		}
	}
	if len(proof.accesses) != len(replayed.x86FixedLocal.accesses) {
		return fail("recomputed function access inventory differs")
	}
	return binding.verify()
}

// Reproduce the renderer's complete source-order feature allocation. The
// reference spelling alone is not proof, and supplied attributes never pass
// through the inference registry or become a removable source difference.
func fixedLocalResultFeatureAttrs(file *File, binding *x86FixedLocalBinding, name, arch string) (string, string, bool) {
	registry := newFeatureAttrRegistry()
	var attributes, inferred string
	matches := 0
	for _, fn := range file.Funcs {
		resolved, found := binding.symbols[fn.Sym]
		sig, declared := binding.signatures[resolved]
		if !found || !declared || len(fn.X86RawText) != 0 || amd64IsRelocationAnchor(fn, sig) || file.x86DirectTailHelper(fn.Sym) {
			return "", "", false
		}
		derived := ""
		if sig.Attrs == "" {
			derived = registry.ref(inferFuncTargetFeaturesForGOARCH(file.Arch, arch, fn))
		}
		if resolved == name {
			attributes, inferred = sig.Attrs, derived
			if attributes == "" {
				attributes = derived
			}
			matches++
		}
	}
	return attributes, inferred, matches == 1
}
