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
		wq := New[*FakeResult](10)

		checkvalue := 0
		mtx := sync.Mutex{}
		for i := 0; i < 100; i++ {
			wq.Push(func(ctx context.Context) (*FakeResult, error) {
				mtx.Lock()
				defer mtx.Unlock()

				checkvalue++
				return nil, nil
			})
		}

		_, errors := wq.RunAll(context.TODO())
		require.Equal(t, 100, checkvalue, "Unexpected number of jobs executed")
		require.Equal(t, 0, len(errors), "Some unexpected errors occurred")
	})

	t.Run("error management", func(t *testing.T) {
		wq := New[*FakeResult](10)

		checkvalue := 0
		mtx := sync.Mutex{}
		for i := 0; i < 100; i++ {
			j := i
			wq.Push(func(ctx context.Context) (*FakeResult, error) {
				if j%2 == 0 {
					return nil, fmt.Errorf("error %d", j)
				}

				mtx.Lock()
				defer mtx.Unlock()

				checkvalue++
				return nil, nil
			})
		}

		_, errorsList := wq.RunAll(context.TODO())
		require.Equal(t, 50, checkvalue, "Unexpected number of jobs executed")
		require.Equal(t, 50, len(errorsList), "Some unexpected errors occurred")
		for i, err := range errorsList {
			require.True(t, errors.Is(err, ErrJobFailed), "Unexpected error type for error[%d]: %s", i, err.Error())
		}
	})

	t.Run("with random sleeps", func(t *testing.T) {
		wq := New[*FakeResult](5)

		checkvalue := 0
		mtx := sync.Mutex{}
		for i := 0; i < 6; i++ {
			wq.Push(func(ctx context.Context) (*FakeResult, error) {
				mtx.Lock()
				defer mtx.Unlock()

				time.Sleep(time.Duration(rand.Intn(10)*50) * time.Millisecond)
				checkvalue++
				return &FakeResult{}, nil
			})
		}

		_, errors := wq.RunAll(context.TODO())
		require.Equal(t, 6, checkvalue, "Unexpected number of jobs executed")
		require.Equal(t, 0, len(errors), "Some unexpected errors occurred")
	})

	t.Run("verify results", func(t *testing.T) {
		wq := New[*FakeResult](5)

		checkvalue := 0
		mtx := sync.Mutex{}
		for i := 0; i < 6; i++ {
			wq.Push(func(ctx context.Context) (*FakeResult, error) {
				mtx.Lock()
				defer mtx.Unlock()

				checkvalue++
				return &FakeResult{id: i}, nil
			})
		}

		results, errors := wq.RunAll(context.TODO())
		require.Equal(t, 6, checkvalue, "Unexpected number of jobs executed")
		require.Equal(t, 0, len(errors), "Some unexpected errors occurred")

		require.Equal(t, 6, len(results))
		foundIds := make([]int, 6)
		for _, result := range results {
			foundIds[result.id] = result.id
		}
		require.Equal(t, []int{0, 1, 2, 3, 4, 5}, foundIds)
	})
}

func TestStaticWQInRangeShouldCopyItemToPreventScopeShadowing(t *testing.T) {
	wq := New[*FakeResult](2)

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
		wq.Push(func(ctx context.Context) (*FakeResult, error) {
			mtx.Lock()
			defer mtx.Unlock()

			checkvalue += itemCopy.val
			return nil, nil
		})
	}

	_, errors := wq.RunAll(context.TODO())
	require.Equal(t, 100+90+80+70+60+50+40+30+20+10, checkvalue, "Unexpected number of jobs executed")
	require.Equal(t, 0, len(errors), "Some unexpected errors occurred")
}
