package gowq

import (
	"context"
	"fmt"
)

// Start runs the job scheduler, it blocks unless a new job is available.
func (w *WorkQueue) Start(ctx context.Context) {
	w.shutdownChan = make(chan bool)

	w.dynamicJobQueueLock.Lock()
	w.dynamicJobQueue = make(chan Job, 0)
	w.dynamicJobQueueLock.Unlock()

	workersQueue := make(chan bool, w.nWorkers)

	for {
		// w.dynamicJobQueueLock.Lock()
		jobQueueLength := len(w.dynamicJobQueue)
		// w.dynamicJobQueueLock.Unlock()

		w.shutdownRequiredLock.Lock()
		fmt.Printf("+++++++++++++++++++++ SHUTDOWN? %t %d\n", w.shutdownRequired, jobQueueLength)

		if w.shutdownRequired && jobQueueLength == 0 {
			w.shutdownRequiredLock.Unlock()
			fmt.Printf("+++++++++++++++++++++ STOPPING ITERATION\n")
			break
		}

		if w.shutdownRequired {
			fmt.Printf("SHUT DOWN HAS BEEN REQUIRED BUT %d %d", jobQueueLength, len(w.dynamicJobQueue))
		}
		w.shutdownRequiredLock.Unlock()

		fmt.Printf("WAITING FOR JOB %d %d - %d\n", jobQueueLength, len(w.dynamicJobQueue), len(workersQueue))
		job := <-w.dynamicJobQueue
		fmt.Printf("Waiting for worker for job\n")
		workersQueue <- true
		fmt.Printf("Found worker for job\n")

		go func(ctx context.Context) {
			if err := job(ctx); err != nil {
				w.appendError(fmt.Errorf("%w: %s", ErrJobFailed, err.Error()))
			}

			_ = <-workersQueue
			fmt.Printf("worker released\n")
		}(ctx)
	}

	fmt.Printf("+++++++++++++++++++++ ITERATION STOPPED\n")

	close(w.dynamicJobQueue)
	fmt.Printf("CHAN CLOSED\n")

	w.shutdownChan <- true
}

// Schedule sends a new job to the scheduler.
func (w *WorkQueue) Schedule(job Job) error {
	w.shutdownRequiredLock.Lock()

	if w.shutdownRequired {
		w.shutdownRequiredLock.Unlock()
		return fmt.Errorf("shutdown has been required")
	}
	w.shutdownRequiredLock.Unlock()

	w.ensureQueueStarted("Schedule")
	w.dynamicJobQueue <- job
	return nil
}

// Shutdown signals the job scheduler that it should terminate execution and
// then waits until the scheduler actually ends.
// Note that the scheduler will run until the job queue is not empty.
func (w *WorkQueue) Shutdown() bool {
	w.ensureQueueStarted("Shutdown")
	fmt.Printf("Required shutdown signal still %d\n", len(w.dynamicJobQueue))

	w.shutdownRequiredLock.Lock()
	w.shutdownRequired = true
	w.shutdownRequiredLock.Unlock()

	fmt.Printf("Issued shutdown signal, waiting\n")
	return <-w.shutdownChan
}

func (w *WorkQueue) ensureQueueStarted(method string) {
	w.dynamicJobQueueLock.Lock()
	defer w.dynamicJobQueueLock.Unlock()

	if w.dynamicJobQueue == nil {
		panic(fmt.Errorf("%s failed: %w", method, ErrQueueNotStarted))
	}
}
