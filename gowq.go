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

// WorkQueue represents the instance of a queue of jobs.
type WorkQueue struct {
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

// NewWQ creates a new WorkQueue instance to schedule jobs.
func NewWQ(workers int) *WorkQueue {
	w := &WorkQueue{
		nWorkers: workers,
	}
	w.unsafeInitErrorsList()
	return w
}

// GetErrors returns the list of errors that occurred during job execution.
// If flush argument is set to true the internal error list gets flushed.
func (w *WorkQueue) GetErrors(flush bool) []error {
	w.errorsListLock.Lock()
	list := w.errorsList[:]

	if flush {
		w.unsafeFlushErrors()
		w.unsafeInitErrorsList()
	}
	w.errorsListLock.Unlock()
	return list
}

// FlushErrors can be used to clean error list.
func (w *WorkQueue) FlushErrors() {
	w.errorsListLock.Lock()
	defer w.errorsListLock.Unlock()
	w.unsafeFlushErrors()
	w.unsafeInitErrorsList()
}

func (w *WorkQueue) appendError(err error) {
	w.errorsListLock.Lock()
	defer w.errorsListLock.Unlock()

	w.errorsList = append(w.errorsList, err)
}

func (w *WorkQueue) unsafeFlushErrors() {
	w.errorsList = nil
}

func (w *WorkQueue) unsafeInitErrorsList() {
	w.errorsList = make([]error, 0)
}
