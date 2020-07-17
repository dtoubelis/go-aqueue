//
// Copyright (c) Dmitri Toubelis
//

package aqueue

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPushAsyncPopNil(t *testing.T) {
	var wg sync.WaitGroup
	var err error

	q := NewAQueue()

	// try to  send
	err = q.TryPush(nil)
	assert.NoError(t, err)

	// now we expect to block
	waitFunc, _ := q.PushAsync(nil)
	wg.Add(1)
	go func() {
		err := waitFunc()
		assert.NoError(t, err)
		wg.Done()
	}()

	val, err := q.Pop()
	assert.NoError(t, err)
	assert.Nil(t, val)
	assert.IsType(t, nil, val)
	assert.Equal(t, nil, val)
	wg.Wait()
}

func TestPushAsyncPopString(t *testing.T) {
	var wg sync.WaitGroup
	var err error

	refVal := "string1"
	q := NewAQueue()

	// try to  send
	err = q.TryPush(refVal)
	assert.NoError(t, err)

	// now we expect to block
	waitFunc, _ := q.PushAsync(refVal)
	wg.Add(1)
	go func() {
		err := waitFunc()
		assert.NoError(t, err)
		wg.Done()
	}()

	val, err := q.Pop()
	assert.NoError(t, err)
	assert.NotNil(t, val)
	assert.IsType(t, refVal, val)
	assert.Equal(t, refVal, val.(string))
	wg.Wait()
}

func TestPushAsyncCancel(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	refVal := "string1"
	refErr := NewError(StatusCodeCancelled, "request cancelled")
	q := NewAQueue()

	pushFunc, cancelFunc := q.PushAsync(refVal)
	go func() {
		err := pushFunc()
		assert.Error(t, err)
		assert.IsType(t, refErr, err)
		assert.Equal(t, refErr.Error(), err.Error())
		assert.Equal(t, refErr.StatusCode(), err.(*Error).StatusCode())
		wg.Done()
	}()

	cancelFunc()
	// call second time to check for panics
	cancelFunc()
	wg.Wait()
}

func TestPushAsyncClose(t *testing.T) {
	var wg sync.WaitGroup
	var err error

	sources := 1000
	refVal := "string1"
	refErr := NewError(StatusCodeClosed, "queue closed")
	q := NewAQueue()

	// fill the buffer
	err = q.TryPush(refVal)
	assert.NoError(t, err)

	// check that we are blocking
	err = q.TryPush(refVal)
	assert.Error(t, err)

	wg.Add(sources)
	for i := 0; i < sources; i++ {
		pushFunc, _ := q.PushAsync(refVal)
		go func() {
			err := pushFunc()
			assert.Error(t, err)
			assert.IsType(t, refErr, err)
			assert.Equal(t, refErr.Error(), err.Error())
			assert.Equal(t, refErr.StatusCode(), err.(*Error).StatusCode())
			wg.Done()
		}()
	}
	q.Close()
	// call second time to check for panics
	q.Close()
	wg.Wait()
}

func TestTryPush(t *testing.T) {
	var err error

	refVal := "string1"
	refErr := NewError(StatusCodeBusy, "queue busy")
	q := NewAQueue()

	// try to  send
	err = q.TryPush(refVal)
	assert.NoError(t, err)

	// try to send again
	err = q.TryPush("string2")
	assert.Error(t, err)
	assert.IsType(t, refErr, err)
	assert.Equal(t, refErr.Error(), err.Error())
	assert.Equal(t, refErr.StatusCode(), err.(*Error).StatusCode())

	// clear the queue
	val, err := q.TryPop()
	assert.NoError(t, err)
	assert.NotNil(t, val)
	assert.IsType(t, refVal, val)
	assert.Equal(t, refVal, val)

	// try to send again
	err = q.TryPush(refVal)
	assert.NoError(t, err)
}

func TestPopAsyncPushNil(t *testing.T) {
	var wg sync.WaitGroup

	q := NewAQueue()
	popFunc, _ := q.PopAsync()

	wg.Add(1)
	go func() {
		val, err := popFunc()
		assert.NoError(t, err)
		assert.Nil(t, val)
		wg.Done()
	}()

	err := q.Push(nil)
	assert.NoError(t, err)
	wg.Wait()
}

func TestPopAsyncPushString(t *testing.T) {
	var wg sync.WaitGroup

	refVal := "string1"
	q := NewAQueue()
	popFunc, _ := q.PopAsync()

	wg.Add(1)
	go func() {
		val, err := popFunc()
		assert.NoError(t, err)
		assert.NotNil(t, val)
		assert.IsType(t, refVal, val)
		assert.Equal(t, refVal, val)
		wg.Done()
	}()
	err := q.Push(refVal)
	assert.NoError(t, err)
	wg.Wait()
}

func TestPopAsyncCancel(t *testing.T) {
	var wg sync.WaitGroup

	refErr := NewError(StatusCodeCancelled, "request cancelled")
	q := NewAQueue()
	popFunc, cancelFunc := q.PopAsync()

	wg.Add(1)
	go func() {
		val, err := popFunc()
		assert.Nil(t, val)
		assert.Error(t, err)
		assert.IsType(t, refErr, err)
		assert.Equal(t, refErr.Error(), err.Error())
		assert.Equal(t, refErr.StatusCode(), err.(*Error).StatusCode())
		wg.Done()
	}()

	cancelFunc()
	// call second time to check for panics
	cancelFunc()
	wg.Wait()
}

func TestPopAsyncClose(t *testing.T) {
	var wg sync.WaitGroup

	rounds := 1000
	refErr := NewError(StatusCodeClosed, "queue closed")
	q := NewAQueue()

	wg.Add(rounds)
	for i := 0; i < rounds; i++ {
		popFunc, _ := q.PopAsync()
		go func() {
			val, err := popFunc()
			assert.Nil(t, val)
			assert.Error(t, err)
			assert.IsType(t, refErr, err)
			assert.Equal(t, refErr.Error(), err.Error())
			assert.Equal(t, refErr.StatusCode(), err.(*Error).StatusCode())
			wg.Done()
		}()
	}

	q.Close()
	// call second time to check for panics
	q.Close()
	wg.Wait()
}

func TestTryPop(t *testing.T) {
	var err error
	var val interface{}

	refVal := "string1"
	refErr := NewError(StatusCodeBusy, "queue busy")
	q := NewAQueue()

	// try to pop
	val, err = q.TryPop()
	assert.Nil(t, val)
	assert.Error(t, err)
	assert.IsType(t, refErr, err)
	assert.Equal(t, refErr.Error(), err.Error())
	assert.Equal(t, refErr.StatusCode(), err.(*Error).StatusCode())

	// send data
	err = q.TryPush(refVal)
	assert.NoError(t, err)

	// try to pop
	val, err = q.TryPop()
	assert.NoError(t, err)
	assert.NotNil(t, val)
	assert.IsType(t, refVal, val)
	assert.Equal(t, refVal, val.(string))

	// try to pop again
	val, err = q.TryPop()
	assert.Nil(t, val)
	assert.Error(t, err)
	assert.IsType(t, refErr, err)
	assert.Equal(t, refErr.Error(), err.Error())
	assert.Equal(t, refErr.StatusCode(), err.(*Error).StatusCode())
}

func BenchmarkPushThroughQueue(b *testing.B) {
	q := NewAQueue()

	concurrent := 97
	for i := 0; i < concurrent; i++ {
		go func(idx int) {
			for {
				if err := q.Push(idx); err != nil {
					break
				}
			}
		}(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.Pop()
	}
	q.Close()
}

func BenchmarkPushThroughChannel(b *testing.B) {
	c := make(chan int)

	concurrent := 97
	for i := 0; i < concurrent; i++ {
		go func(idx int) {
			for {
				c <- idx
			}
		}(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		<-c
	}
}

func ExampleAQueue_PushAsync_withContext() {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	q := NewAQueue()
	// add a value for next call to block
	q.TryPush(nil)
	// initiate asynchronous call
	funcPush, funcCancel := q.PushAsync(nil)
	// wait for context in a separate thread
	go func() {
		<-ctx.Done()
		funcCancel()
	}()
	// wait for push to timeout
	if err := funcPush(); err != nil {
		fmt.Println(err.Error())
	}
	// Output:
	// request cancelled
}

func ExampleAQueue_PopAsync_withContext() {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	q := NewAQueue()
	// initiate asynchronous call
	funcPop, funcCancel := q.PopAsync()
	// wait for context in a separate thread
	go func() {
		<-ctx.Done()
		funcCancel()
	}()
	// wait for push to timeout
	if _, err := funcPop(); err != nil {
		fmt.Println(err.Error())
	}
	// Output:
	// request cancelled
}
