//
// Copyright (c) Dmitri Toubelis
//

package async

import (
	"context"
	"sync"
)

// Ctx is opaque context
type Ctx struct {
	sync.Mutex
}

// New returns a new queue instance
func New() Queue {
	return &Ctx{}
}

// Push implements Queue.Push() call
func (ctx *Ctx) Push(interface{}) error {
	panic("not implemented")
}

// PushWithContext implements Queue.PushWithContext() call
func (ctx *Ctx) PushWithContext(context.Context, interface{}) error {
	panic("not implemented")
}

// TryPush implements Queue.TryPush() call
func (ctx *Ctx) TryPush(interface{}) error {
	panic("not implemented")
}

// Pop implements Queue.Pop() call
func (ctx *Ctx) Pop() (interface{}, error) {
	panic("not implemented")
}

// PopWithContext implements Queue.PopWithContext() call
func (ctx *Ctx) PopWithContext(context.Context) (interface{}, error) {
	panic("not implemented")
}

// TryPop implements Queue.TryPop() call
func (ctx *Ctx) TryPop() (interface{}, error) {
	panic("not implemented")
}

// Close implements Queue.Close() call
func (ctx *Ctx) Close() error {
	panic("not implemented")
}
