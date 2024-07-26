package gowq

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDynamicWQ(t *testing.T) {
	t.Run("basic setup", func(t *testing.T) {
		wq := New[*FakeResult](3)

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
			wq.Schedule(func(ctx context.Context) (*FakeResult, error) {
				mtx.Lock()
				checkvalue++
				mtx.Unlock()
				return nil, nil
			})
		}

		terminated := wq.Shutdown()

		require.True(t, terminated, "Start has not terminated yet")

		mtx.Lock()
		require.Equal(t, 100, checkvalue, "Unexpected number of jobs executed")
		mtx.Unlock()
	})

	t.Run("error management", func(t *testing.T) {
		wq := New[*FakeResult](10)
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
			wq.Schedule(func(ctx context.Context) (*FakeResult, error) {
				time.Sleep(time.Duration(rand.Intn(10)*10) * time.Millisecond)

				if j%2 == 0 {
					return nil, fmt.Errorf("error %d", j)
				}
				mtx.Lock()
				checkvalue++
				mtx.Unlock()
				return &FakeResult{id: j}, nil
			})
		}

		terminated := wq.Shutdown()
		time.Sleep(1 * time.Second)
		require.True(t, terminated, "Start has not terminated yet")

		mtx.Lock()
		require.Equal(t, 50, checkvalue, "Unexpected number of jobs executed")
		mtx.Unlock()

		errorsList := wq.GetErrors(false)
		require.Equal(t, 50, len(errorsList), "Unexpected number of errors")
		for i, err := range errorsList {
			require.True(t, errors.Is(err, ErrJobFailed), "Unexpected error type for error[%d]: %s", i, err.Error())
		}

		resultsList := wq.GetResults()
		require.Len(t, resultsList, 50)
		expectedResults := make([]int, 0)
		for i := 0; i < 100; i++ {
			if i%2 != 0 {
				expectedResults = append(expectedResults, i)
			}
		}
		resultIds := make([]int, 0)
		for _, result := range resultsList {
			resultIds = append(resultIds, result.id)
		}

		slices.Sort(resultIds)
		require.Equal(t, expectedResults, resultIds)
	})
}

func TestSchedule(t *testing.T) {
	t.Run("can't Schedule before Start", func(t *testing.T) {
		var panicOccurred bool
		var panicError error
		defer func() {
			if r := recover(); r != nil {
				panicOccurred = true
				panicError = r.(error)
			}
		}()

		wq := New[*FakeResult](10)
		wq.Schedule(fakeJob)

		require.True(t, panicOccurred, "A panic was expected")
		require.True(t, errors.Is(panicError, ErrQueueNotStarted), "Unexpected error type")
	})

	t.Run("Job is actually enqueued", func(t *testing.T) {
		wq := New[*FakeResult](10).(*workQueue[*FakeResult])
		simulateStart(wq, 1)

		require.Equal(t, len(wq.dynamicJobQueue), 0, "Unexpected job queue before enqueue")
		wq.Schedule(fakeJob)
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

		wq := New[*FakeResult](10)
		wq.Shutdown()

		require.True(t, panicOccurred, "A panic was expected")
		require.True(t, errors.Is(panicError, ErrQueueNotStarted), "Unexpected error type")
	})
}
