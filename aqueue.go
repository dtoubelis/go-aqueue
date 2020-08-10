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

// PushWithContext adds element to the queue with context
func (q *AQueue) PushWithContext(ctx context.Context, val interface{}) error {
	if ctx == nil {
		return errInvalidArgument
	}
	pushFunc, cancelFunc := q.PushAsync(val)
	go func() {
		<-ctx.Done()
		cancelFunc()
	}()
	return pushFunc()
}

// PushWithTimeout is a convenience method implementing Push() timeout
func (q *AQueue) PushWithTimeout(ctx context.Context, val interface{}, d time.Duration) error {
	if ctx == nil {
		return errInvalidArgument
	}
	// create new context with timeout
	newCtx, cancel := context.WithTimeout(ctx, d)
	defer cancel()
	// initiate asynchronous call
	funcPush, funcCancel := q.PushAsync(nil)
	// wait for context cancelation in a separate thread
	go func() {
		<-newCtx.Done()
		funcCancel()
	}()
	// wait for push to complete
	return funcPush()
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

// PopWithContext removes an element from the queue with context
func (q *AQueue) PopWithContext(ctx context.Context) (interface{}, error) {
	if ctx == nil {
		return nil, errInvalidArgument
	}
	popFunc, cancelFunc := q.PopAsync()
	go func() {
		<-ctx.Done()
		cancelFunc()
	}()
	return popFunc()
}

// PopWithTimeout removes an element from the queue with context and timeout
func (q *AQueue) PopWithTimeout(ctx context.Context, d time.Duration) (interface{}, error) {
	if ctx == nil {
		return nil, errInvalidArgument
	}
	// create new context with timeout
	newCtx, cancel := context.WithTimeout(ctx, d)
	defer cancel()
	// perform async Pop()
	popFunc, cancelFunc := q.PopAsync()
	go func() {
		<-newCtx.Done()
		cancelFunc()
	}()
	return popFunc()
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
	// check if que is closed
	if q.closed {
		return nil, errClosed
	}
	// update value or wait
	if !q.hasValue {
		return nil, errBusy
	}
	val := q.val
	q.val = nil
	q.hasValue = false
	return val, nil
}

// Close closes the queue causing any pending Pop/Push calls to exit with EOF error
// and any subsequent requests to fail as well. Closed queue cannot be reused.
func (q *AQueue) Close() {
	q.lock.Lock()
	defer q.lock.Unlock()
	if q.closed {
		return
	}
	q.val = nil // release any references
	q.hasValue = false
	q.closed = true
	q.cond.Broadcast()
}
