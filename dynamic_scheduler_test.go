package gowq

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDynamicWQ(t *testing.T) {
	t.Run("basic setup", func(t *testing.T) {
		wq := NewWQ(3)

		var startWait sync.WaitGroup
		startWait.Add(1)
		go func() {
			wq.Start(context.TODO())
			startWait.Done()
		}()

		checkvalue := 0
		mtx := sync.Mutex{}
		time.Sleep(1 * time.Second)
		for i := 0; i < 100; i++ {
			wq.Enqueue(func(ctx context.Context) error {
				mtx.Lock()
				checkvalue++
				mtx.Unlock()
				return nil
			})
		}

		terminated := wq.Shutdown()

		require.True(t, terminated, "Start has not terminated yet")

		mtx.Lock()
		require.Equal(t, 100, checkvalue, "Unexpected number of jobs executed")
		mtx.Unlock()
	})

	t.Run("error management", func(t *testing.T) {
		wq := NewWQ(10)
		var startWait sync.WaitGroup

		startWait.Add(1)
		go func() {
			wq.Start(context.TODO())
			startWait.Done()
		}()

		checkvalue := 0
		mtx := sync.Mutex{}
		time.Sleep(1 * time.Second)
		for i := 0; i < 100; i++ {
			j := i
			wq.Enqueue(func(ctx context.Context) error {
				if j%2 == 0 {
					return fmt.Errorf("error %d", j)
				}
				mtx.Lock()
				checkvalue++
				mtx.Unlock()
				return nil
			})
		}

		terminated := wq.Shutdown()

		require.True(t, terminated, "Start has not terminated yet")

		mtx.Lock()
		require.Equal(t, 50, checkvalue, "Unexpected number of jobs executed")
		mtx.Unlock()

		errorsList := wq.GetErrors(false)
		require.Equal(t, 50, len(errorsList), "Unexpected number of errors")
	})
}

func TestEnqueue(t *testing.T) {
	t.Run("can't Enqueue before Start", func(t *testing.T) {
		var panicOccurred bool
		var panicError error
		defer func() {
			if r := recover(); r != nil {
				panicOccurred = true
				panicError = r.(error)
			}
		}()

		wq := NewWQ(10)
		wq.Enqueue(fakeJob)

		require.True(t, panicOccurred, "A panic was expected")
		require.True(t, errors.Is(panicError, ErrQueueNotStarted), "Unexpected error type")
	})

	t.Run("Job is actually enqueued", func(t *testing.T) {
		wq := NewWQ(10)
		simulateStart(wq, 1)

		require.Equal(t, len(wq.dynamicJobQueue), 0, "Unexpected job queue before enqueue")
		wq.Enqueue(fakeJob)
		require.Equal(t, len(wq.dynamicJobQueue), 1, "Unexpected job queue after enqueue")
	})
}

func TestShutdown(t *testing.T) {
	t.Run("can't Shutdown before Start", func(t *testing.T) {
		var panicOccurred bool
		var panicError error
		defer func() {
			if r := recover(); r != nil {
				panicOccurred = true
				panicError = r.(error)
			}
		}()

		wq := NewWQ(10)
		wq.Shutdown()

		require.True(t, panicOccurred, "A panic was expected")
		require.True(t, errors.Is(panicError, ErrQueueNotStarted), "Unexpected error type")
	})
}
