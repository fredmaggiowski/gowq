package gowq

import (
	"context"
	"fmt"
	"sync"
)

// A Job is a simple function that should be executed in a limited set of routines.
type Job func(context.Context) error

// WorkQueue represents the instance of a queue of jobs.
type WorkQueue struct {
	nWorkers int

	jobQueue []Job
}

// NewWQ creates a new WorkQueue instance to schedule jobs.
func NewWQ(workers int) *WorkQueue {
	return &WorkQueue{
		nWorkers: workers,
		jobQueue: make([]Job, 0),
	}
}

// RunAll runs all scheduled jobs and returns when all jobs terminated.
func (w *WorkQueue) RunAll(ctx context.Context) []error {
	workersQueue := make(chan bool, w.nWorkers)

	var waitGroup sync.WaitGroup
	waitGroup.Add(len(w.jobQueue))

	errorList := make([]error, 0)
	errorsLock := sync.Mutex{}

	for i := 0; i < len(w.jobQueue); i++ {
		job := w.jobQueue[i]

		workersQueue <- true

		go func(c context.Context, i int) {
			if err := job(ctx); err != nil {
				errorsLock.Lock()
				errorList = append(
					errorList,
					fmt.Errorf("job %d failed: %w", i, err),
				)
				errorsLock.Unlock()
			}

			waitGroup.Done()
			_ = <-workersQueue
		}(ctx, i)
	}
	waitGroup.Wait()
	return errorList
}

// Schedule can be used to append a new job to the work queue.
func (w *WorkQueue) Schedule(job Job) {
	w.jobQueue = append(w.jobQueue, job)
}
