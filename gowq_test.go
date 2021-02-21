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

func TestInRange(t *testing.T) {
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
