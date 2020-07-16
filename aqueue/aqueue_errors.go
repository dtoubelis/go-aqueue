//
// Copyright (c) Dmitri Toubelis
//

package aqueue

// StatusCode represents status of an operation on the queue
type StatusCode int

const (
	// StatusCodeClosed is returned when queue is closed
	StatusCodeClosed StatusCode = iota
	// StatusCodeBusy returned when performing non blocking operation and queue is blocked
	StatusCodeBusy
	// StatusCodeCancelled returned when blocking operation times outd
	StatusCodeCancelled
)

// QueueError is opaque error context
type QueueError struct {
	status StatusCode
	msg    string
}

// NewQueueError create a new
func NewQueueError(status StatusCode, msg string) *QueueError {
	return &QueueError{
		status: status,
		msg:    msg,
	}
}

// Error implements error.Error() call
func (e *QueueError) Error() string {
	return e.msg
}

// StatusCode return a short operation status
func (e *QueueError) StatusCode() StatusCode {
	return e.status
}
