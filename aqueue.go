//
// Copyright (c) Dmitri Toubelis
//

package aqueue

import (
	"sync"
)

var (
	errBusy      = NewError(StatusCodeBusy, "queue busy")
	errClosed    = NewError(StatusCodeClosed, "queue closed")
	errCancelled = NewError(StatusCodeCancelled, "request cancelled")
)

type popFunc func() (interface{}, error)
type pushFunc func() error
type cancelFunc func()

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

// Push adds an element to the end of the queue.
func (c *AQueue) Push(val interface{}) error {
	pushFunc, _ := c.pushAsync(val)
	return pushFunc()
}

func (c *AQueue) pushAsync(val interface{}) (pushFunc, cancelFunc) {
	cancelled := false
	return func() error {
			c.lock.Lock()
			defer c.lock.Unlock()
			for {
				if cancelled {
					return errCancelled
				}
				if err := c.tryPushUnsync(val); err == nil {
					c.cond.Broadcast()
					return nil
				} else if e, ok := err.(*Error); ok {
					if e.StatusCode() == StatusCodeClosed {
						return e
					}
				} else {
					panic("unknown error type")
				}
				c.cond.Wait()
			}
		}, func() {
			c.lock.Lock()
			defer c.lock.Unlock()
			if !cancelled {
				cancelled = true
				c.cond.Broadcast()
			}

		}
}

// TryPush attempts to add an element to the queue without blocking.
func (c *AQueue) TryPush(val interface{}) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.tryPushUnsync(val)
}

func (c *AQueue) tryPushUnsync(val interface{}) error {
	// check if queue is closed
	if c.closed {
		return errClosed
	}
	// update value or wait
	if !c.hasValue {
		c.val = val
		c.hasValue = true
		return nil
	}
	return errBusy
}

// Pop removes the first element from the queue
func (c *AQueue) Pop() (interface{}, error) {
	popFunc, _ := c.popAsync()
	return popFunc()
}

func (c *AQueue) popAsync() (popFunc, cancelFunc) {
	cancelled := false
	return func() (interface{}, error) {
			c.lock.Lock()
			defer c.lock.Unlock()
			for {
				if cancelled {
					return nil, errCancelled
				}
				if val, err := c.tryPopUnsync(); err == nil {
					c.cond.Broadcast()
					return val, nil
				} else if e, ok := err.(*Error); ok {
					if e.StatusCode() == StatusCodeClosed {
						return nil, e
					}
				} else {
					panic("unknown error type")
				}
				c.cond.Wait()
			}
		}, func() {
			c.lock.Lock()
			defer c.lock.Unlock()
			if !cancelled {
				cancelled = true
				c.cond.Broadcast()
			}
		}
}

// TryPop attemts to remove the last element from the queue without blocking.
func (c *AQueue) TryPop() (interface{}, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.tryPopUnsync()
}

func (c *AQueue) tryPopUnsync() (interface{}, error) {
	// check if que is closed
	if c.closed {
		return nil, errClosed
	}
	// update value or wait
	if !c.hasValue {
		return nil, errBusy
	}
	val := c.val
	c.val = nil
	c.hasValue = false
	return val, nil
}

// Close closes the queue causing any pending Pop/Push calls to exit with EOF error
// and any subsequent requests to fail as well. Closed queue cannot be reused.
func (c *AQueue) Close() {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.closed {
		return
	}
	c.val = nil // release any references
	c.hasValue = false
	c.closed = true
	c.cond.Broadcast()
}
