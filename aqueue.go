// Copyright 2020 The AQueue Contributors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package aqueue

import (
	"context"
	"sync"
	"time"
)

var (
	errBusy            = NewError(StatusCodeBusy, "queue busy")
	errClosed          = NewError(StatusCodeClosed, "queue closed")
	errCancelled       = NewError(StatusCodeCancelled, "request cancelled")
	errInvalidArgument = NewError(StatusCodeInvalidArgument, "invalid arguments")
)

// PopFunc is a promise that returns value when data is available
type PopFunc func() (interface{}, error)

// PushFunc is promise that returns when data is pushed to the queue
type PushFunc func() error

// CancelFunc is called to cancel an asynchronous operation
type CancelFunc func()

// AQueue is opaque context
type AQueue struct {
	lock     *sync.Mutex
	cond     *sync.Cond
	val      interface{}
	hasValue bool
	closed   bool
}

// NewAQueue returns a new queue instance
func NewAQueue() *AQueue {
	ctx := &AQueue{
		lock: &sync.Mutex{},
	}
	ctx.cond = sync.NewCond(ctx.lock)
	return ctx
}

// Push adds an element to the queue
func (q *AQueue) Push(val interface{}) error {
	pushFunc, _ := q.PushAsync(val)
	return pushFunc()
}

// PushWithContext adds an element to the queue, giving up when ctx is done.
func (q *AQueue) PushWithContext(ctx context.Context, val interface{}) error {
	if ctx == nil {
		return errInvalidArgument
	}
	pushFunc, cancelFunc := q.PushAsync(val)
	stop := watchContext(ctx, cancelFunc)
	defer stop()
	return pushFunc()
}

// PushWithTimeout is a convenience method implementing Push() with a timeout
// on top of ctx.
func (q *AQueue) PushWithTimeout(ctx context.Context, val interface{}, d time.Duration) error {
	if ctx == nil {
		return errInvalidArgument
	}
	newCtx, cancel := context.WithTimeout(ctx, d)
	defer cancel()
	return q.PushWithContext(newCtx, val)
}

// watchContext calls cancelFunc if ctx is done before the returned stop
// function is invoked. Unlike a bare `go func() { <-ctx.Done(); cancel() }()`,
// the watcher goroutine always exits when the operation completes, so it
// cannot leak for long-lived or background contexts.
func watchContext(ctx context.Context, cancelFunc CancelFunc) (stop func()) {
	// Fast path: nothing to watch for a context that can never be done.
	if ctx.Done() == nil {
		return func() {}
	}
	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			cancelFunc()
		case <-done:
		}
	}()
	return func() { close(done) }
}

// PushAsync initiates an asynchronoush push and returns
// a future and a cancel function
func (q *AQueue) PushAsync(val interface{}) (PushFunc, CancelFunc) {
	cancelled := false
	return func() error {
			q.lock.Lock()
			defer q.lock.Unlock()
			for {
				if cancelled {
					return errCancelled
				}
				if err := q.tryPushUnsync(val); err == nil {
					q.cond.Broadcast()
					return nil
				} else if e, ok := err.(*Error); ok {
					if e.StatusCode() == StatusCodeClosed {
						return e
					}
				} else {
					panic("unknown error type")
				}
				q.cond.Wait()
			}
		}, func() {
			q.lock.Lock()
			defer q.lock.Unlock()
			if !cancelled {
				cancelled = true
				q.cond.Broadcast()
			}

		}
}

// TryPush attempts to add an element to the queue without blocking.
func (q *AQueue) TryPush(val interface{}) error {
	q.lock.Lock()
	defer q.lock.Unlock()
	return q.tryPushUnsync(val)
}

func (q *AQueue) tryPushUnsync(val interface{}) error {
	// check if queue is closed
	if q.closed {
		return errClosed
	}
	// update value or wait
	if !q.hasValue {
		q.val = val
		q.hasValue = true
		return nil
	}
	return errBusy
}

// Pop removes the first element from the queue
func (q *AQueue) Pop() (interface{}, error) {
	popFunc, _ := q.PopAsync()
	return popFunc()
}

// PopWithContext removes an element from the queue, giving up when ctx is done.
func (q *AQueue) PopWithContext(ctx context.Context) (interface{}, error) {
	if ctx == nil {
		return nil, errInvalidArgument
	}
	popFunc, cancelFunc := q.PopAsync()
	stop := watchContext(ctx, cancelFunc)
	defer stop()
	return popFunc()
}

// PopWithTimeout is a convenience method implementing Pop() with a timeout
// on top of ctx.
func (q *AQueue) PopWithTimeout(ctx context.Context, d time.Duration) (interface{}, error) {
	if ctx == nil {
		return nil, errInvalidArgument
	}
	newCtx, cancel := context.WithTimeout(ctx, d)
	defer cancel()
	return q.PopWithContext(newCtx)
}

// PopAsync initiates retrieval of the next a value from the queue and
// returns a future and a cancel function
func (q *AQueue) PopAsync() (PopFunc, CancelFunc) {
	cancelled := false
	return func() (interface{}, error) {
			q.lock.Lock()
			defer q.lock.Unlock()
			for {
				if cancelled {
					return nil, errCancelled
				}
				if val, err := q.tryPopUnsync(); err == nil {
					q.cond.Broadcast()
					return val, nil
				} else if e, ok := err.(*Error); ok {
					if e.StatusCode() == StatusCodeClosed {
						return nil, e
					}
				} else {
					panic("unknown error type")
				}
				q.cond.Wait()
			}
		}, func() {
			q.lock.Lock()
			defer q.lock.Unlock()
			if !cancelled {
				cancelled = true
				q.cond.Broadcast()
			}
		}
}

// TryPop attempts to remove the last element from the queue without blocking.
func (q *AQueue) TryPop() (interface{}, error) {
	q.lock.Lock()
	defer q.lock.Unlock()
	return q.tryPopUnsync()
}

func (q *AQueue) tryPopUnsync() (interface{}, error) {
	// A value that was successfully pushed before Close() must still be
	// delivered: Push already reported success to the producer, so dropping
	// it here would lose an acknowledged message. Only report closed once
	// the slot is empty.
	if !q.hasValue {
		if q.closed {
			return nil, errClosed
		}
		return nil, errBusy
	}
	val := q.val
	q.val = nil
	q.hasValue = false
	return val, nil
}

// Close closes the queue. Pending and subsequent Push calls fail with a
// StatusCodeClosed error. A value that was already pushed remains available
// to exactly one Pop; after that (or immediately, if the queue was empty)
// Pop fails with StatusCodeClosed. A closed queue cannot be reused.
func (q *AQueue) Close() {
	q.lock.Lock()
	defer q.lock.Unlock()
	if q.closed {
		return
	}
	q.closed = true
	q.cond.Broadcast()
}
