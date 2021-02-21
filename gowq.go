package gowq

import (
	"context"
	"sync"
)

// A Job is a simple function.
type Job func(context.Context)

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
func (w *WorkQueue) RunAll(ctx context.Context) {
	workersQueue := make(chan bool, w.nWorkers)

	var waitGroup sync.WaitGroup
	waitGroup.Add(len(w.jobQueue))

	for i := 0; i < len(w.jobQueue); i++ {
		job := w.jobQueue[i]

		workersQueue <- true

		go func(c context.Context, i int) {
			job(ctx)

			waitGroup.Done()
			_ = <-workersQueue
		}(ctx, i)
	}
	waitGroup.Wait()
	return
}

// Schedule can be used to append a new job to the work queue.
func (w *WorkQueue) Schedule(job Job) {
	w.jobQueue = append(w.jobQueue, job)
}
