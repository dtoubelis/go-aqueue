//
// Copyright (c) Dmitri Toubelis
//

package aqueue

import (
	"os"
	"sync"
	"testing"

	"github.com/inconshreveable/log15"
	"github.com/stretchr/testify/assert"
)

func init() {
	log15.Root().SetHandler(log15.LvlFilterHandler(log15.LvlDebug, log15.CallerFileHandler(log15.StreamHandler(os.Stdout, log15.LogfmtFormat()))))
}

func TestPushAsyncPopNil(t *testing.T) {
	var wg sync.WaitGroup
	var err error

	q := New()

	// try to  send
	err = q.TryPush(nil)
	assert.NoError(t, err)

	// now we expect to block
	waitFunc, _ := q.pushAsync(nil)
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
	q := New()

	// try to  send
	err = q.TryPush(refVal)
	assert.NoError(t, err)

	// now we expect to block
	waitFunc, _ := q.pushAsync(refVal)
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
	refErr := NewQueueError(StatusCodeCancelled, "request cancelled")
	q := New()

	pushFunc, cancelFunc := q.pushAsync(refVal)
	go func() {
		err := pushFunc()
		assert.Error(t, err)
		assert.IsType(t, refErr, err)
		assert.Equal(t, refErr.Error(), err.Error())
		assert.Equal(t, refErr.StatusCode(), err.(*QueueError).StatusCode())
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

	rounds := 1000
	refVal := "string1"
	refErr := NewQueueError(StatusCodeClosed, "queue closed")
	q := New()

	// fill the buffer
	err = q.TryPush(refVal)
	assert.NoError(t, err)

	// check that we are blocking
	err = q.TryPush(refVal)
	assert.Error(t, err)

	wg.Add(rounds)
	for i := 0; i < rounds; i++ {
		pushFunc, _ := q.pushAsync(refVal)
		go func() {
			err := pushFunc()
			assert.Error(t, err)
			assert.IsType(t, refErr, err)
			assert.Equal(t, refErr.Error(), err.Error())
			assert.Equal(t, refErr.StatusCode(), err.(*QueueError).StatusCode())
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
	refErr := NewQueueError(StatusCodeBusy, "queue busy")
	q := New()

	// try to  send
	err = q.TryPush(refVal)
	assert.NoError(t, err)

	// try to send again
	err = q.TryPush("string2")
	assert.Error(t, err)
	assert.IsType(t, refErr, err)
	assert.Equal(t, refErr.Error(), err.Error())
	assert.Equal(t, refErr.StatusCode(), err.(*QueueError).StatusCode())

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

	q := New()
	popFunc, _ := q.popAsync()

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
	q := New()
	popFunc, _ := q.popAsync()

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

	refErr := NewQueueError(StatusCodeCancelled, "request cancelled")
	q := New()
	popFunc, cancelFunc := q.popAsync()

	wg.Add(1)
	go func() {
		val, err := popFunc()
		assert.Nil(t, val)
		assert.Error(t, err)
		assert.IsType(t, refErr, err)
		assert.Equal(t, refErr.Error(), err.Error())
		assert.Equal(t, refErr.StatusCode(), err.(*QueueError).StatusCode())
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
	refErr := NewQueueError(StatusCodeClosed, "queue closed")
	q := New()

	wg.Add(rounds)
	for i := 0; i < rounds; i++ {
		popFunc, _ := q.popAsync()
		go func() {
			val, err := popFunc()
			assert.Nil(t, val)
			assert.Error(t, err)
			assert.IsType(t, refErr, err)
			assert.Equal(t, refErr.Error(), err.Error())
			assert.Equal(t, refErr.StatusCode(), err.(*QueueError).StatusCode())
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
	refErr := NewQueueError(StatusCodeBusy, "queue busy")
	q := New()

	// try to pop
	val, err = q.TryPop()
	assert.Nil(t, val)
	assert.Error(t, err)
	assert.IsType(t, refErr, err)
	assert.Equal(t, refErr.Error(), err.Error())
	assert.Equal(t, refErr.StatusCode(), err.(*QueueError).StatusCode())

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
	assert.Equal(t, refErr.StatusCode(), err.(*QueueError).StatusCode())
}
