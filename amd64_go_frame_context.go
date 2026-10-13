package plan9asm

import (
	"context"
	"errors"
	"time"
)

// Caller contexts are callbacks, not safe command contexts under an observer
// lock. Capture only their actual lifetime outside locks; commands consume a
// standard private context with the same (never extended) deadline/Done signal.
// The original context remains the published binding's lifetime authority.
type goAMD64FramePrivateContext struct {
	context.Context
	original context.Context
	deadline time.Time
	bounded  bool
	done     <-chan struct{}
	cancel   context.CancelFunc
	stop     func()
}

func captureGoAMD64FramePrivateContext(original context.Context) (*goAMD64FramePrivateContext, error) {
	if original == nil {
		return nil, goAMD64FrameContext("missing caller context")
	}
	if err := original.Err(); err != nil {
		return nil, err
	}
	deadline, bounded := original.Deadline()
	done := original.Done()
	if err := original.Err(); err != nil {
		return nil, err
	}
	var private context.Context
	var cancel context.CancelFunc
	if bounded {
		private, cancel = context.WithDeadline(context.Background(), deadline)
	} else {
		private, cancel = context.WithCancel(context.Background())
	}
	finished, stopped := make(chan struct{}), make(chan struct{})
	go func() {
		defer close(stopped)
		select {
		case <-done:
			cancel()
		case <-private.Done():
		case <-finished:
		}
	}()
	return &goAMD64FramePrivateContext{Context: private, original: original,
		deadline: deadline, bounded: bounded, done: done, cancel: cancel,
		stop: func() {
			close(finished)
			cancel()
			<-stopped
		}}, nil
}

// Always called outside observer locks. Retain actual cancellation causes and
// reject a caller that changes its captured lifetime while work is in flight.
func (ctx *goAMD64FramePrivateContext) verify() error {
	err := ctx.original.Err()
	deadline, bounded := ctx.original.Deadline()
	done := ctx.original.Done()
	if bounded != ctx.bounded || bounded && !deadline.Equal(ctx.deadline) || done != ctx.done {
		err = errors.Join(err, goAMD64FrameContext("caller context lifetime changed during observation"))
	}
	select {
	case <-ctx.done:
		ctx.cancel()
	default:
	}
	return errors.Join(err, ctx.original.Err(), ctx.Context.Err())
}
