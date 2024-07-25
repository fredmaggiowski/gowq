package gowq

import (
	"context"
	"fmt"
	"time"
)

// Start runs the job scheduler, it blocks unless a new job is available.
func (w *workQueue) Start(ctx context.Context) {
	w.shutdownChan = make(chan bool)

	w.dynamicJobQueueLock.Lock()
	w.dynamicJobQueue = make(chan Job)
	w.dynamicJobQueueLock.Unlock()

	workersQueue := make(chan bool, w.nWorkers)

	for {
		w.shutdownRequiredLock.Lock()
		shutdownRequired := w.shutdownRequired
		w.shutdownRequiredLock.Unlock()

		if shutdownRequired && len(w.dynamicJobQueue) == 0 {
			break
		}

		var job Job

		select {
		case job = <-w.dynamicJobQueue:
		default:
			time.Sleep(1 * time.Millisecond)
			continue
		}

		workersQueue <- true

		go func(ctx context.Context) {
			if err := job(ctx); err != nil {
				w.appendError(fmt.Errorf("%w: %s", ErrJobFailed, err.Error()))
			}
			<-workersQueue
		}(ctx)
	}

	close(w.dynamicJobQueue)
	w.shutdownChan <- true
}

// Schedule sends a new job to the scheduler.
func (w *workQueue) Schedule(job Job) {
	w.ensureQueueStarted("Schedule")
	w.dynamicJobQueue <- job
}

// Shutdown signals the job scheduler that it should terminate execution and
// then waits until the scheduler actually ends.
// Note that the scheduler will run until the job queue is not empty.
func (w *workQueue) Shutdown() bool {
	w.ensureQueueStarted("Shutdown")

	w.shutdownRequiredLock.Lock()
	w.shutdownRequired = true
	w.shutdownRequiredLock.Unlock()

	return <-w.shutdownChan
}

func (w *workQueue) ensureQueueStarted(method string) {
	w.dynamicJobQueueLock.Lock()
	defer w.dynamicJobQueueLock.Unlock()

	if w.dynamicJobQueue == nil {
		panic(fmt.Errorf("%s failed: %w", method, ErrQueueNotStarted))
	}
}
