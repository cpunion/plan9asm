package plan9asm

import (
	"context"
	"errors"
)

// Context implementations are caller-controlled callbacks. Sample their
// methods before taking operation locks, then bridge only the captured Done
// channels/deadlines to a standard private command context. The caller's actual
// cancellation cause is still checked outside locks before and after the work.
func sdkASMDataPrivateContext(caller, original context.Context) (context.Context, func(), error) {
	if caller == nil || original == nil {
		return nil, nil, sdkASMDataContext("missing verification/capture context")
	}
	if err := errors.Join(caller.Err(), original.Err()); err != nil {
		return nil, nil, err
	}
	callerDeadline, callerBounded := caller.Deadline()
	deadline, bounded := original.Deadline()
	callerDone, originalDone := caller.Done(), original.Done()
	if callerBounded && (!bounded || callerDeadline.Before(deadline)) {
		deadline, bounded = callerDeadline, true
	}
	if !bounded {
		return nil, nil, sdkASMDataContext("original bounded context lost its deadline")
	}
	if err := errors.Join(caller.Err(), original.Err()); err != nil {
		return nil, nil, err
	}
	private, cancel := context.WithDeadline(context.Background(), deadline)
	finished, stopped := make(chan struct{}), make(chan struct{})
	go func() {
		defer close(stopped)
		select {
		case <-callerDone:
			cancel()
		case <-originalDone:
			cancel()
		case <-private.Done():
		case <-finished:
		}
	}()
	return private, func() {
		close(finished)
		cancel()
		<-stopped
	}, nil
}
