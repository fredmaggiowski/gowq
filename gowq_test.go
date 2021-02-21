package gowq

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test(t *testing.T) {
	wq := NewWQ(10)

	checkvalue := 0
	mtx := sync.Mutex{}

	for i := 0; i < 100; i++ {
		wq.Schedule(func(ctx context.Context) {
			mtx.Lock()
			defer mtx.Unlock()

			checkvalue++
		})
	}

	wq.RunAll(context.TODO())
	require.Equal(t, 100, checkvalue)
}
