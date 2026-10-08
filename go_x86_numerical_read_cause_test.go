//go:build go1.27

package plan9asm

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestGoX86NumericalSBReadCaptureCancellationCause(t *testing.T) {
	for _, sdkCPU := range []bool{false, true} {
		for _, cause := range []error{context.Canceled, context.DeadlineExceeded} {
			t.Run(cause.Error()+map[bool]string{false: "/ordinary", true: "/SDK"}[sdkCPU], func(t *testing.T) {
				var ctx context.Context
				var cancel context.CancelFunc
				if cause == context.Canceled {
					ctx, cancel = context.WithCancel(context.Background())
					cancel()
				} else {
					ctx, cancel = context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
					defer cancel()
				}
				owner, err := CaptureGoNumericalStorageOwner(ctx, GoNumericalStorageCaptureRequest{SDKCPU: sdkCPU})
				if owner != nil || !errors.Is(err, cause) || errors.Is(err, ErrProbeNeedsContext) {
					t.Fatalf("cancelled infrastructure lost its cause or became source Context: owner=%v error=%v", owner, err)
				}
			})
		}
	}
}
