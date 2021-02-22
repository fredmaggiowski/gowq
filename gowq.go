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

// Start runs the job scheduler, it blocks unless a new job is available.
func (w *WorkQueue) Start(ctx context.Context) {
	w.dynamicJobQueueLock.Lock()
	w.dynamicJobQueue = make(chan Job, w.nWorkers)
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

		nextJob := <-w.dynamicJobQueue
		workersQueue <- true

		go func(ctx context.Context) {
			nextJob(ctx)
			_ = <-workersQueue
		}(ctx)
	}

	close(w.dynamicJobQueue)
	w.shutdownChan <- true
}

func (w *WorkQueue) ensureQueueStarted(method string) {
	w.dynamicJobQueueLock.Lock()
	defer w.dynamicJobQueueLock.Unlock()
	if w.dynamicJobQueue == nil {
		panic(fmt.Errorf("%s failed: %w", method, ErrQueueNotStarted))
	}
}

// Enqueue sends a new job to the scheduler, it blocks when the job queue is
// full until there's space available for a new job.
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
	close(workersQueue)
	return errorList
}

// Schedule can be used to append a new job to the work queue.
func (w *WorkQueue) Schedule(job Job) {
	if w.staticJobQueue == nil {
		w.staticJobQueue = make([]Job, 0)
	}
	w.staticJobQueue = append(w.staticJobQueue, job)
}
