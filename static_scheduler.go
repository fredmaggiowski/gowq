package gowq

import (
	"context"
	"fmt"
	"sync"
)

// RunAll runs all scheduled jobs and returns when all jobs terminated.
func (w *workQueue[Result]) RunAll(ctx context.Context) ([]Result, []error) {
	workersQueue := make(chan bool, w.nWorkers)

	var waitGroup sync.WaitGroup
	waitGroup.Add(len(w.staticJobQueue))

	w.errorsList = make([]error, 0)

	for i := 0; i < len(w.staticJobQueue); i++ {
		job := w.staticJobQueue[i]

		workersQueue <- true

		go func(c context.Context, i int) {
			result, err := job(ctx)
			if err != nil {
				w.appendError(fmt.Errorf("%w [%d]: %s", ErrJobFailed, i, err.Error()))
			} else {
				w.appendResult(result)
			}
			waitGroup.Done()
			<-workersQueue
		}(ctx, i)
	}
	waitGroup.Wait()
	close(workersQueue)

	return w.GetResults(), w.GetErrors(true)
}

// Push can be used to append a new job to the work queue.
func (w *workQueue[Result]) Push(job Job[Result]) {
	if w.staticJobQueue == nil {
		w.staticJobQueue = make([]Job[Result], 0)
	}
	w.staticJobQueue = append(w.staticJobQueue, job)
}
