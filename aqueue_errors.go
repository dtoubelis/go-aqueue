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

// Error is opaque error context
type Error struct {
	status StatusCode
	msg    string
}

// NewError create a new
func NewError(status StatusCode, msg string) *Error {
	return &Error{
		status: status,
		msg:    msg,
	}
}

// Error implements error.Error() call
func (e *Error) Error() string {
	return e.msg
}

// StatusCode return a short operation status
func (e *Error) StatusCode() StatusCode {
	return e.status
}
