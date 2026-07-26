//go:build !go1.23

package plan9asm

import "go/types"

func newAliasTypeForTest(_ *types.TypeName, rhs types.Type) types.Type {
	return rhs
}
