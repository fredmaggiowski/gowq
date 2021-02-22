package gowq

import (
	"context"
	"fmt"
)

// Start runs the job scheduler, it blocks unless a new job is available.
func (w *WorkQueue) Start(ctx context.Context) {
	w.dynamicJobQueueLock.Lock()
	w.dynamicJobQueue = make(chan Job, 0)
	w.shutdownChan = make(chan bool)
	w.dynamicJobQueueLock.Unlock()

	workersQueue := make(chan bool, w.nWorkers)

	for {
		w.shutdownRequiredLock.Lock()
		if w.shutdownRequired && len(w.dynamicJobQueue) == 0 {
			w.shutdownRequiredLock.Unlock()
			break
		}
		w.shutdownRequiredLock.Unlock()

		job := <-w.dynamicJobQueue
		workersQueue <- true

		go func(ctx context.Context) {
			if err := job(ctx); err != nil {
				w.appendError(fmt.Errorf("job failed: %w", err))
			}
			_ = <-workersQueue
		}(ctx)
	}

	close(w.dynamicJobQueue)
	w.shutdownChan <- true
}

// Enqueue sends a new job to the scheduler.
func (w *WorkQueue) Enqueue(job Job) {
	w.ensureQueueStarted("Enqueue")
	w.dynamicJobQueue <- job
}

// Shutdown signals the job scheduler that it should terminate execution and
// then waits until the scheduler actually ends.
// Note that the scheduler will run until the job queue is not empty.
func (w *WorkQueue) Shutdown() bool {
	w.ensureQueueStarted("Shutdown")

	w.shutdownRequiredLock.Lock()
	w.shutdownRequired = true
	w.shutdownRequiredLock.Unlock()

	return <-w.shutdownChan
}

func (w *WorkQueue) ensureQueueStarted(method string) {
	w.dynamicJobQueueLock.Lock()
	defer w.dynamicJobQueueLock.Unlock()
	if w.dynamicJobQueue == nil {
		panic(fmt.Errorf("%s failed: %w", method, ErrQueueNotStarted))
	}
}
