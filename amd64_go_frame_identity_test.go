package plan9asm

import (
	"errors"
	"os"
	"reflect"
	"testing"
)

func TestGoAMD64FrameSourceCopyCannotBorrowOrCleanOriginal(t *testing.T) {
	for _, operation := range []string{"Verify", "Close"} {
		t.Run(operation, func(t *testing.T) {
			pkg, request := goAMD64FrameSourceFixture(t)
			ctx, cancel := runtimeOracleStageContext(t)
			defer cancel()
			source, err := CaptureGoAMD64FrameSource(ctx, pkg, request)
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := source.Close(); err != nil {
					t.Error(err)
				}
			}()
			if err := source.Verify(ctx); err != nil {
				t.Fatal("actual original public factory did not verify", err)
			}
			owned := source.owned
			before, err := os.Lstat(owned)
			if err != nil {
				t.Fatal(err)
			}
			// Deliberately model the public whole-value operation without hiding
			// a mutex copy behind a constructor or serializing private fields.
			copied := new(GoAMD64FrameSource)
			reflect.ValueOf(copied).Elem().Set(reflect.ValueOf(source).Elem())
			if operation == "Verify" {
				err = copied.Verify(ctx)
			} else {
				err = copied.Close()
			}
			if !errors.Is(err, ErrProbeNeedsContext) {
				t.Errorf("whole-value copied frame %s borrowed its original owner role: %v", operation, err)
			}
			after, err := os.Lstat(owned)
			if err != nil || !os.SameFile(before, after) {
				t.Error("copied frame operation removed or replaced original artifacts", err)
			}
			if err := source.Verify(ctx); err != nil {
				t.Error("copied frame operation revoked its original object", err)
			}
		})
	}
}
