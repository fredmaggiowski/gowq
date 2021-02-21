package gowq

import (
	"context"
	"sync"
)

// Job .
type Job func(context.Context)

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

func (w *WorkQueue) Schedule(job Job) {
	w.jobQueue = append(w.jobQueue, job)
}
