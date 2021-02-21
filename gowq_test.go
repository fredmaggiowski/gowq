package gowq

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test(t *testing.T) {
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
}

func TestWithErrors(t *testing.T) {
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
}
