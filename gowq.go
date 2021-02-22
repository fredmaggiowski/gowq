package gowq

import (
	"context"
	"fmt"
	"sync"
)

// A Job is a simple function that should be executed in a limited set of routines.
type Job func(context.Context) error

var (
	// ErrQueueNotStarted error used whenever an action is required on a dynamic queue
	// that has not been started yet.
	ErrQueueNotStarted = fmt.Errorf("Start must be called before")
)

// WorkQueue represents the instance of a queue of jobs.
type WorkQueue struct {
	nWorkers int

	errorsList     []error
	errorsListLock sync.Mutex

	staticJobQueue []Job

	dynamicJobQueue      chan Job
	dynamicJobQueueLock  sync.Mutex
	shutdownRequired     bool
	shutdownRequiredLock sync.Mutex
	shutdownChan         chan bool
}

// NewWQ creates a new WorkQueue instance to schedule jobs.
func NewWQ(workers int) *WorkQueue {
	return &WorkQueue{
		nWorkers: workers,
	}
}

// GetErrors returns the list of errors that occurred during job execution.
// If flush argument is set to true the internal error list gets flushed.
func (w *WorkQueue) GetErrors(flush bool) []error {
	w.errorsListLock.Lock()
	list := w.errorsList[:]
	w.errorsListLock.Unlock()

	if flush {
		w.FlushErrors()
	}
	return list
}

// FlushErrors can be used to clean error list.
func (w *WorkQueue) FlushErrors() {
	w.errorsListLock.Lock()
	defer w.errorsListLock.Unlock()
	w.errorsList = nil
}

// RunAll runs all scheduled jobs and returns when all jobs terminated.
func (w *WorkQueue) RunAll(ctx context.Context) []error {
	workersQueue := make(chan bool, w.nWorkers)

	var waitGroup sync.WaitGroup
	waitGroup.Add(len(w.staticJobQueue))

	w.errorsList = make([]error, 0)

	for i := 0; i < len(w.staticJobQueue); i++ {
		job := w.staticJobQueue[i]

		workersQueue <- true

		go func(c context.Context, i int) {
			if err := job(ctx); err != nil {
				w.appendError(fmt.Errorf("job %d failed: %w", i, err))
			}

			waitGroup.Done()
			_ = <-workersQueue
		}(ctx, i)
	}
	waitGroup.Wait()
	close(workersQueue)

	return w.GetErrors(true)
}

// Schedule can be used to append a new job to the work queue.
func (w *WorkQueue) Schedule(job Job) {
	if w.staticJobQueue == nil {
		w.staticJobQueue = make([]Job, 0)
	}
	w.staticJobQueue = append(w.staticJobQueue, job)
}

func (w *WorkQueue) appendError(err error) {
	w.errorsListLock.Lock()
	defer w.errorsListLock.Unlock()

	w.errorsList = append(w.errorsList, err)
}
