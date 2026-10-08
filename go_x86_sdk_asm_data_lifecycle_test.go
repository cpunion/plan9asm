//go:build go1.27

package plan9asm

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

type sdkASMDataReentrantContext struct {
	context.Context
	armed    atomic.Bool
	callback func()
}

func (ctx *sdkASMDataReentrantContext) Err() error {
	if ctx.armed.CompareAndSwap(true, false) {
		ctx.callback()
	}
	return ctx.Context.Err()
}

func TestGoX86SDKASMDataLiveValidationAPIReentry(t *testing.T) {
	for _, route := range []string{"verification", "original"} {
		for _, api := range []string{"Verify", "Readback", "CheckReadback"} {
			t.Run(route+"/"+api, func(t *testing.T) {
				ctx, request := sdkASMDataRequest(t, "amd64", "internal/cpu")
				reentrant := &sdkASMDataReentrantContext{Context: ctx}
				captureContext := ctx
				if route == "original" {
					captureContext = reentrant
				}
				source, err := CaptureGoX86SDKASMDataSource(captureContext, request)
				if err != nil {
					t.Fatal("actual SDK source precondition", err)
				}
				t.Cleanup(func() { source.Close() })
				raw, err := source.Readback(ctx)
				if err != nil {
					t.Fatal(err)
				}
				var nested error
				reentrant.callback = func() {
					switch api {
					case "Verify":
						nested = source.Verify(ctx)
					case "Readback":
						_, nested = source.Readback(ctx)
					case "CheckReadback":
						nested = source.CheckReadback(ctx, raw)
					}
				}
				reentrant.armed.Store(true)
				done := make(chan error, 1)
				go func() {
					err := source.Verify(reentrant)
					done <- errors.Join(err, nested)
				}()
				select {
				case err := <-done:
					if err != nil {
						t.Fatal("live original/nested observation failed", err)
					}
				case <-time.After(2 * time.Second):
					// A test-process-local functional RED: a bounded context
					// cannot interrupt a synchronous Err callback deadlock.
					t.Fatal(fmt.Sprintf("%s Err callback permanently blocked reentrant %s", route, api))
				}
			})
		}
	}
}
