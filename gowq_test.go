package gowq

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestStaticWQ(t *testing.T) {
	t.Run("basic setup", func(t *testing.T) {
		wq := NewWQ(10)

		checkvalue := 0
		mtx := sync.Mutex{}
		for i := 0; i < 100; i++ {
			wq.Schedule(func(ctx context.Context) error {
				mtx.Lock()
				defer mtx.Unlock()

				checkvalue++
				return nil
			})
		}

		errors := wq.RunAll(context.TODO())
		require.Equal(t, 100, checkvalue, "Unexpected number of jobs executed")
		require.Equal(t, 0, len(errors), "Some unexpected errors occurred")
	})

	t.Run("error management", func(t *testing.T) {
		wq := NewWQ(10)

		checkvalue := 0
		mtx := sync.Mutex{}
		for i := 0; i < 100; i++ {
			j := i
			wq.Schedule(func(ctx context.Context) error {
				if j%2 == 0 {
					return fmt.Errorf("error %d", j)
				}

				mtx.Lock()
				defer mtx.Unlock()

				checkvalue++
				return nil
			})
		}

		errors := wq.RunAll(context.TODO())
		require.Equal(t, 50, checkvalue, "Unexpected number of jobs executed")
		require.Equal(t, 50, len(errors), "Some unexpected errors occurred")
	})

	t.Run("with random sleeps", func(t *testing.T) {
		wq := NewWQ(5)

		checkvalue := 0
		mtx := sync.Mutex{}
		for i := 0; i < 6; i++ {
			wq.Schedule(func(ctx context.Context) error {
				mtx.Lock()
				defer mtx.Unlock()

				time.Sleep(time.Duration(rand.Intn(10)*100) * time.Millisecond)
				checkvalue++
				return nil
			})
		}

		errors := wq.RunAll(context.TODO())
		require.Equal(t, 6, checkvalue, "Unexpected number of jobs executed")
		require.Equal(t, 0, len(errors), "Some unexpected errors occurred")
	})
}

func TestInRangeShouldCopyItemToPreventScopeShadowing(t *testing.T) {
	wq := NewWQ(2)

	checkvalue := 0
	mtx := sync.Mutex{}

	items := []struct {
		val int
	}{
		{val: 10}, {val: 20}, {val: 30}, {val: 40}, {val: 50},
		{val: 60}, {val: 70}, {val: 80}, {val: 90}, {val: 100},
	}

	for _, item := range items {
		itemCopy := item
		wq.Schedule(func(ctx context.Context) error {
			mtx.Lock()
			defer mtx.Unlock()

			checkvalue += itemCopy.val
			return nil
		})
	}

	errors := wq.RunAll(context.TODO())
	require.Equal(t, 100+90+80+70+60+50+40+30+20+10, checkvalue, "Unexpected number of jobs executed")
	require.Equal(t, 0, len(errors), "Some unexpected errors occurred")
}

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

func fakeJob(ctx context.Context) error {
	return nil
}

func simulateStart(wq *WorkQueue, size int) {
	wq.dynamicJobQueue = make(chan Job, size)
}
