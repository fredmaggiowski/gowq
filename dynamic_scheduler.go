package gowq

import (
	"context"
	"fmt"
	"time"
)

// Start runs the job scheduler, it blocks unless a new job is available.
func (w *workQueue[Result]) Start(ctx context.Context) {
	w.shutdownChan = make(chan bool)

	w.dynamicJobQueueLock.Lock()
	w.dynamicJobQueue = make(chan Job[Result])
	w.dynamicJobQueueLock.Unlock()

	workersQueue := make(chan bool, w.nWorkers)

	for {
		w.shutdownRequiredLock.Lock()
		shutdownRequired := w.shutdownRequired
		w.shutdownRequiredLock.Unlock()

		if shutdownRequired && len(w.dynamicJobQueue) == 0 {
			break
		}

		var job Job[Result]

		select {
		case job = <-w.dynamicJobQueue:
		default:
			time.Sleep(1 * time.Millisecond)
			continue
		}

		workersQueue <- true

		go func(ctx context.Context) {
			result, err := job(ctx)
			if err != nil {
				w.appendError(fmt.Errorf("%w: %s", ErrJobFailed, err.Error()))
			} else {
				w.appendResult(result)
			}
			<-workersQueue
		}(ctx)
	}

	close(w.dynamicJobQueue)
	w.shutdownChan <- true
}

// Schedule sends a new job to the scheduler.
func (w *workQueue[Result]) Schedule(job Job[Result]) {
	w.ensureQueueStarted("Schedule")
	w.dynamicJobQueue <- job
}

// Shutdown signals the job scheduler that it should terminate execution and
// then waits until the scheduler actually ends.
// Note that the scheduler will run until the job queue is not empty.
func (w *workQueue[Result]) Shutdown() bool {
	w.ensureQueueStarted("Shutdown")

	w.shutdownRequiredLock.Lock()
	w.shutdownRequired = true
	w.shutdownRequiredLock.Unlock()

	return <-w.shutdownChan
}

func (w *workQueue[Result]) ensureQueueStarted(method string) {
	w.dynamicJobQueueLock.Lock()
	defer w.dynamicJobQueueLock.Unlock()

	if w.dynamicJobQueue == nil {
		panic(fmt.Errorf("%s failed: %w", method, ErrQueueNotStarted))
	}
}
