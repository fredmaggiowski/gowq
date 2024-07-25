package gowq

import (
	"context"
	"fmt"
	"sync"
)

// RunAll runs all scheduled jobs and returns when all jobs terminated.
func (w *workQueue) RunAll(ctx context.Context) []error {
	workersQueue := make(chan bool, w.nWorkers)

	var waitGroup sync.WaitGroup
	waitGroup.Add(len(w.staticJobQueue))

	w.errorsList = make([]error, 0)

	for i := 0; i < len(w.staticJobQueue); i++ {
		job := w.staticJobQueue[i]

		workersQueue <- true

		go func(c context.Context, i int) {
			if err := job(ctx); err != nil {
				w.appendError(fmt.Errorf("%w [%d]: %s", ErrJobFailed, i, err.Error()))
			}

			waitGroup.Done()
			<-workersQueue
		}(ctx, i)
	}
	waitGroup.Wait()
	close(workersQueue)

	return w.GetErrors(true)
}

// Push can be used to append a new job to the work queue.
func (w *workQueue) Push(job Job) {
	if w.staticJobQueue == nil {
		w.staticJobQueue = make([]Job, 0)
	}
	w.staticJobQueue = append(w.staticJobQueue, job)
}
