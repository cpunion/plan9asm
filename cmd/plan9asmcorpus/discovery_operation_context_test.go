package main

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDiscoveryNestedOperationKeepsParentContext(t *testing.T) {
	for _, deadline := range []bool{false, true} {
		name := "cancelled"
		var parent context.Context
		var cancel context.CancelFunc
		if deadline {
			name = "expired"
			parent, cancel = context.WithTimeout(context.Background(), time.Millisecond)
			<-parent.Done()
		} else {
			parent, cancel = context.WithCancel(context.Background())
			cancel()
		}
		t.Run(name, func(t *testing.T) {
			defer cancel()
			called := false
			err := runDiscoveryOperationContext(parent, time.Minute, func(context.Context) error {
				called = true
				return nil
			})
			if called || !errors.Is(err, parent.Err()) {
				t.Fatalf("closed parent launched an operation: called=%v error=%v parent=%v", called, err, parent.Err())
			}
		})
	}

	t.Run("nil_parent", func(t *testing.T) {
		called := false
		err := runDiscoveryOperationContext(nil, time.Minute, func(context.Context) error {
			called = true
			return nil
		})
		if err == nil || called {
			t.Fatalf("missing operation owner accepted: called=%v error=%v", called, err)
		}
	})

	t.Run("values_and_deadline", func(t *testing.T) {
		type ownerKey struct{}
		parent, cancel := context.WithTimeout(context.WithValue(context.Background(), ownerKey{}, "actual-owner"), time.Minute)
		defer cancel()
		parentDeadline, _ := parent.Deadline()
		err := runDiscoveryOperationContext(parent, time.Hour, func(ctx context.Context) error {
			deadline, bounded := ctx.Deadline()
			if ctx.Value(ownerKey{}) != "actual-owner" || !bounded || !deadline.Equal(parentDeadline) {
				t.Errorf("nested operation lost its owner: value=%v deadline=%v bounded=%v want=%v", ctx.Value(ownerKey{}), deadline, bounded, parentDeadline)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("parent_cancel_during_operation", func(t *testing.T) {
		parent, cancel := context.WithCancel(context.Background())
		defer cancel()
		sourceErr := errors.New("independent source diagnostic")
		err := runDiscoveryOperationContext(parent, time.Minute, func(ctx context.Context) error {
			cancel()
			timer := time.NewTimer(200 * time.Millisecond)
			defer timer.Stop()
			select {
			case <-ctx.Done():
			case <-timer.C:
				t.Error("nested operation ignored its cancelled parent")
			}
			return sourceErr
		})
		if !errors.Is(err, context.Canceled) || !errors.Is(err, sourceErr) {
			t.Fatalf("parent cancellation or independent source error was hidden: %v", err)
		}
	})

	t.Run("independent_child_deadline", func(t *testing.T) {
		parent, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		err := runDiscoveryOperationContext(parent, time.Millisecond, func(ctx context.Context) error {
			<-ctx.Done()
			return nil
		})
		if !errors.Is(err, context.DeadlineExceeded) || parent.Err() != nil {
			t.Fatalf("child timeout failed or cancelled its parent: error=%v parent=%v", err, parent.Err())
		}
	})
}
