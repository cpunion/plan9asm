package plan9asm

import "context"

// CaptureGoX86SDKNumericalReadSource is an explicit SDK source entry point.
// SDK source identity is separate from the SDK CPU numerical storage owner;
// neither a Standard flag nor an assembly symbol supplies that identity.
func CaptureGoX86SDKNumericalReadSource(ctx context.Context, request GoX86NumericalReadCaptureRequest) (*GoX86NumericalReadSource, error) {
	return CaptureGoX86NumericalReadSource(ctx, request)
}
