// Copyright 2020 AQueue contributors
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

// StatusCode represents status of an operation on the queue
type StatusCode int

const (
	// StatusCodeClosed is returned when queue is closed
	StatusCodeClosed StatusCode = iota
	// StatusCodeBusy returned when performing non blocking operation and queue is blocked
	StatusCodeBusy
	// StatusCodeCancelled returned when blocking operation times outd
	StatusCodeCancelled
	// StatusCodeInvalidArgument ...
	StatusCodeInvalidArgument
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
