//
// Copyright (c) Dmitri Toubelis
//

package aqueue

import (
	"sync"
)

var (
	errBusy      = NewQueueError(StatusCodeBusy, "queue busy")
	errClosed    = NewQueueError(StatusCodeClosed, "queue closed")
	errCancelled = NewQueueError(StatusCodeCancelled, "request cancelled")
)

type popFunc func() (interface{}, error)
type pushFunc func() error
type cancelFunc func()

// Ctx is opaque context
type Ctx struct {
	lock     *sync.Mutex
	cond     *sync.Cond
	val      interface{}
	hasValue bool
	closed   bool
}

// New returns a new queue instance
func New() *Ctx {
	ctx := &Ctx{
		lock: &sync.Mutex{},
	}
	ctx.cond = sync.NewCond(ctx.lock)
	return ctx
}

// Push implements Queue.Push() call
func (c *Ctx) Push(val interface{}) error {
	pushFunc, _ := c.pushAsync(val)
	return pushFunc()
}

func (c *Ctx) pushAsync(val interface{}) (pushFunc, cancelFunc) {
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
				} else if e, ok := err.(*QueueError); ok {
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

// TryPush implements Queue.TryPush() call
func (c *Ctx) TryPush(val interface{}) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.tryPushUnsync(val)
}

func (c *Ctx) tryPushUnsync(val interface{}) error {
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

// Pop implements Queue.Pop() call
func (c *Ctx) Pop() (interface{}, error) {
	popFunc, _ := c.popAsync()
	return popFunc()
}

func (c *Ctx) popAsync() (popFunc, cancelFunc) {
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
				} else if e, ok := err.(*QueueError); ok {
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

// TryPop implements Queue.TryPop() call
func (c *Ctx) TryPop() (interface{}, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.tryPopUnsync()
}

func (c *Ctx) tryPopUnsync() (interface{}, error) {
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

// Close implements Queue.Close() call
func (c *Ctx) Close() {
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
