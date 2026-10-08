package plan9asm

import (
	"context"
	"fmt"
)

type GoARM64ByteCarrierCaptureRequest struct {
	GoBinary, Directory, PackagePath, AssemblyFile, Target string
	Environment                                            []string
}

// GoARM64ByteCarrierSource observes a complete private numeric source, original
// compiler entries and same-file ABI0 caller/callee objects. Capture and its
// portable readback do not provide Options, writable padding or entry authority.
type GoARM64ByteCarrierSource struct{}

func arm64ByteCarrierContext(format string, arguments ...any) error {
	return fmt.Errorf("%w: actual ARM64 byte-carrier source: %s", ErrProbeNeedsContext, fmt.Sprintf(format, arguments...))
}

func CaptureGoARM64ByteCarrierSource(ctx context.Context, request GoARM64ByteCarrierCaptureRequest) (*GoARM64ByteCarrierSource, error) {
	return nil, arm64ByteCarrierContext("complete original caller/entry source observation is not implemented")
}

func (source *GoARM64ByteCarrierSource) Readback(ctx context.Context) ([]byte, error) {
	return nil, arm64ByteCarrierContext("missing live source observation")
}

func (source *GoARM64ByteCarrierSource) Verify(ctx context.Context) error {
	return arm64ByteCarrierContext("missing live source observation")
}

func (source *GoARM64ByteCarrierSource) Close() error {
	return nil
}
