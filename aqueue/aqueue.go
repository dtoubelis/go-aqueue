//
// Copyright (c) Dmitri Toubelis
//

package aqueue

// Queue is doble ended queue
type Queue interface {

	// Push adds an element to the end of the queue.
	Push(interface{}) error

	// TryPush attempts to add an element to the queue without blocking.
	TryPush(interface{}) error

	// Pop removes the first element from the queue
	Pop() (interface{}, error)

	// TryPop attemts to remove the last element from the queue without blocking.
	TryPop() (interface{}, error)

	// Close closes the queue causing any pending Pop/Push calls to exit with EOF error
	// and any subsequent requests to fail as well. Closed queue cannot be reused.
	Close()
}

