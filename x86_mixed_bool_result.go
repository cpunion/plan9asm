package plan9asm

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"
)

// Initialized physical bytes are not a canonical bool witness. The existing
// scalar-I1 consumers retain their own entry/result proofs; an aggregate or
// contradictory result layout needs an independent complete canonical role.
// Inspect both declaration and modeled slots so dropping a bool field cannot
// turn an unregistered role into an ordinary numerical result.
func x86MixedBoolResultNeedsCanonicalRole(sig FuncSig, modeled []FrameSlot) bool {
	if sig.Ret == Void {
		return false
	}
	containsBool := func(typ LLVMType) bool {
		for _, token := range strings.FieldsFunc(string(typ), func(r rune) bool {
			return unicode.IsSpace(r) || strings.ContainsRune("{}[]<>,", r)
		}) {
			if token == string(I1) {
				return true
			}
		}
		return false
	}
	hasBool := containsBool(sig.Ret)
	for _, slots := range [][]FrameSlot{sig.Frame.Results, modeled} {
		for _, slot := range slots {
			hasBool = hasBool || containsBool(slot.Type)
		}
	}
	if !hasBool {
		return false
	}
	if sig.Ret != I1 || len(sig.Frame.Results) > 1 || len(modeled) > 1 {
		return true
	}
	if len(sig.Frame.Results) == 0 && len(modeled) == 0 {
		return false // A scalar register-result role has no logical FP slots.
	}
	return !reflect.DeepEqual(sig.Frame.Results, modeled) || sig.Frame.Results[0].Type != I1
}

func (c *amd64Ctx) validateMixedBoolResultRole() error {
	if !x86MixedBoolResultNeedsCanonicalRole(c.sig, c.fpResults) {
		return nil
	}
	if !reflect.DeepEqual(c.sig.Frame.Results, c.fpResults) || c.fpScalarSource.x86FixedLocal == nil {
		return fmt.Errorf("%w: mixed bool result requires an independent canonical result role", ErrProbeNeedsContext)
	}
	// This consumer rechecks the actual declaration, original/normalized source,
	// all field writes and before/final live binding guards. A pointer alone is
	// not authority, nor does this add a new aggregate or native-frame producer.
	return c.scalarFPResultInitialized()
}
