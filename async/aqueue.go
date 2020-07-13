//
// Copyright (c) Dmitri Toubelis
//

package async

import "context"

// Queue is doble ended queue
type Queue interface {

	// Push adds an element to the end of the queue.
	Push(interface{}) error
	// PushWithContext adds an element to the end of the queue with the provided context.
	PushWithContext(context.Context, interface{}) error
	// TryPush attempts to add an element to the queue without blocking.
	TryPush(interface{}) error

	// Pop removes the first element from the queue
	Pop() (interface{}, error)
	// PopWithContext removes the first element of the queue with the provided context.
	PopWithContext(context.Context) (interface{}, error)
	// TryPop attemts to remove the last element from the queue without blocking.
	TryPop() (interface{}, error)

	// Close closes the queue causing any pending Pop/Push calls to exit with EOF error
	// and any subsequent requests to fail as well. Closed queue cannot be reused.
	Close() error
}
