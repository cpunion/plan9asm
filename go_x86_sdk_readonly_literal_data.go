package plan9asm

import "context"

// GoX86SDKReadonlyLiteralDataOwner is an opaque, explicit initializer owner.
// It is distinct from a Go numerical variable and supplies no pointer, read,
// native body, Frame, Options or ABI permission.
type GoX86SDKReadonlyLiteralDataOwner struct{}

// CaptureGoX86SDKReadonlyLiteralDataOwner requires an actual live SDK ASM
// observation. Public reports, names, hashes and types are not constructors.
func CaptureGoX86SDKReadonlyLiteralDataOwner(ctx context.Context, source *GoX86SDKASMDataSource) (*GoX86SDKReadonlyLiteralDataOwner, error) {
	if source == nil {
		return nil, sdkASMDataContext("missing actual SDK literal DATA source")
	}
	if err := source.Verify(ctx); err != nil {
		return nil, err
	}
	return nil, sdkASMDataContext("private readonly literal DATA initializer role is not implemented")
}
