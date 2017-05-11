// Package once provides for running a running a function once, until successful.
package once

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
)

// Success is an object that will perform exactly one action if successful.
type Success struct {
	// This is an atomic instead of, say, a bool so that callers can hot-path without acquiring a lock.
	done uint32
	// This Cond protects the running bool, so that only one Do() execution is happening at once.
	sync.Cond
	// Running is set to true when a goroutine is calling the provided function.
	running bool
}

// New returns a Success, ready to use.
func New() *Success {
	o := &Success{}
	o.L = &sync.Mutex{}
	return o
}

// Do calls the function f if and only if Do is being called for the
// first time for this instance of Success and previous calls were not successful.
// In other words, given
//	var once *Success = New()
// if once.Do(ctx, f) is called multiple times, f will be invoked until it returns
// a non-nil error, even if f has a different value in each invocation.
// A new instance of Success is required for each function to execute.
//
// Do is intended for initialization that must be run exactly once if successful.
//
// Because no call to Do returns until the one call to f returns, if f causes
// Do to be called, it will deadlock.
//
// If f panics, Do considers it to have returned with an error, so future calls
// of Do will invoke f again.
//
// If the context has been canceled before f is called successfully,
// context.Canceled will be returned. Callers are responsible to gracefully handle
// this event.
func (o *Success) Do(ctx context.Context, f func() error) error {
	if atomic.LoadUint32(&o.done) != 0 {
		return nil
	}

	o.L.Lock()
	defer o.L.Unlock()

	for o.running {
		if err := ctx.Err(); err != nil {
			return err
		}
		o.Wait()
	}
	o.running = true
	defer func() {
		o.running = false
	}()

	if err := ctx.Err(); err != nil {
		return err
	}
	if atomic.LoadUint32(&o.done) != 0 {
		return nil
	}

	if err := o.invoke(f); err != nil {
		// Wake up just one goroutine to make the next attempt.
		o.Signal()
		return err
	}

	atomic.StoreUint32(&o.done, 1)
	o.Broadcast()
	return nil
}

func (o *Success) invoke(f func() error) (err error) {
	// This does a pointer to an interface so that the deferred func can change the
	// error when the stack gets unwound.
	defer func(e *error) {
		if r := recover(); r != nil {
			*e = fmt.Errorf("recovered: %v", r)
		}
	}(&err)
	err = f()
	return err
}
