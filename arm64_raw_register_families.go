package plan9asm

import "fmt"

type arm64RawRegisterControlEffects struct {
	gpReads  uint32
	gpWrites uint32
}

// These families are register-only. Their concrete decoder determines whether
// a scalar register belongs to the floating/vector bank or the GP bank. There
// are no memory operations, implicit GP clobbers or control transfers here.
type arm64RawRegisterFamily struct {
	effects func(uint32) (arm64RawRegisterControlEffects, bool)
	lower   func(*arm64Ctx, uint32) error
}

func arm64RawRegisterForm[T any](
	decode func(uint32) (T, bool),
	lower func(*arm64Ctx, T) error,
	effects func(T) arm64RawRegisterControlEffects,
) arm64RawRegisterFamily {
	return arm64RawRegisterFamily{
		effects: func(word uint32) (arm64RawRegisterControlEffects, bool) {
			form, ok := decode(word)
			if !ok {
				return arm64RawRegisterControlEffects{}, false
			}
			return effects(form), true
		},
		lower: func(c *arm64Ctx, word uint32) error {
			form, ok := decode(word)
			if !ok {
				return fmt.Errorf("ARM64 raw register family rejected its selected encoding %#08x", word)
			}
			return lower(c, form)
		},
	}
}

// Only explicitly registered vector-only forms use this constructor.
// GP conversions must instead supply their typed source/destination effects.
func arm64RawVectorOnlyForm[T any](decode func(uint32) (T, bool), lower func(*arm64Ctx, T) error) arm64RawRegisterFamily {
	return arm64RawRegisterForm(decode, lower, func(T) arm64RawRegisterControlEffects {
		return arm64RawRegisterControlEffects{}
	})
}

func arm64RawGPBit(index int) uint32 {
	if index >= 0 && index < 31 {
		return 1 << uint(index)
	}
	return 0 // Encoding 31 is WZR/XZR, not SP.
}

// Register-only families cannot store to a saved stack continuation. Unknown
// encodings still fall through to the conservative control-effect analysis.
func arm64RawRegisterEffects(word uint32) (arm64RawRegisterControlEffects, bool) {
	if _, effects, ok := decodeARM64RawFloatFamily(word); ok {
		return effects, true
	}
	if _, effects, ok := decodeARM64RawVectorFamily(word); ok {
		return effects, true
	}
	return arm64RawRegisterControlEffects{}, false
}
