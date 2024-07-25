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
	// ErrJobFailed error is used to wrap errors provided by failing jobs.
	ErrJobFailed = fmt.Errorf("job failed")
)

type DynamicScheduler interface {
	Start(ctx context.Context)
	Schedule(job Job)
	Shutdown() bool
}

type StaticScheduler interface {
	RunAll(context.Context) []error
	Push(job Job)
}

type Queue interface {
	GetErrors(flush bool) []error
	FlushErrors()
	DynamicScheduler
	StaticScheduler
}

// workQueue represents the instance of a queue of jobs.
type workQueue struct {
	// General properties.
	nWorkers       int
	errorsList     []error
	errorsListLock sync.Mutex

	// Static scheduler properties.
	staticJobQueue []Job

	// Dynamic scheduler properties.
	dynamicJobQueue      chan Job
	dynamicJobQueueLock  sync.Mutex
	shutdownRequired     bool
	shutdownRequiredLock sync.Mutex
	shutdownChan         chan bool
}

// New creates a new WorkQueue instance to schedule jobs.
func New(workers int) Queue {
	return &workQueue{
		nWorkers: workers,
	}
}

// GetErrors returns the list of errors that occurred during job execution.
// If flush argument is set to true the internal error list gets flushed.
func (w *workQueue) GetErrors(flush bool) []error {
	w.errorsListLock.Lock()
	list := w.errorsList[:]
	w.errorsListLock.Unlock()

	if flush {
		w.FlushErrors()
	}
	return list
}

// FlushErrors can be used to clean error list.
func (w *workQueue) FlushErrors() {
	w.errorsListLock.Lock()
	defer w.errorsListLock.Unlock()
	w.errorsList = nil
}

func (w *workQueue) appendError(err error) {
	w.errorsListLock.Lock()
	defer w.errorsListLock.Unlock()

	w.errorsList = append(w.errorsList, err)
}
