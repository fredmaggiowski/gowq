package gowq

import (
	"context"
	"fmt"
	"sync"
)

type JobResult interface{}

// A Job is a simple function that should be executed in a limited set of routines.
type Job[T JobResult] func(context.Context) (T, error)

var (
	// ErrQueueNotStarted error used whenever an action is required on a dynamic queue
	// that has not been started yet.
	ErrQueueNotStarted = fmt.Errorf("Start must be called before")
	// ErrJobFailed error is used to wrap errors provided by failing jobs.
	ErrJobFailed = fmt.Errorf("job failed")
)

type DynamicScheduler[Result JobResult] interface {
	Start(ctx context.Context)
	Schedule(job Job[Result])
	Shutdown() bool
}

type StaticScheduler[Result JobResult] interface {
	RunAll(context.Context) ([]Result, []error)
	Push(job Job[Result])
}

type Queue[Result JobResult] interface {
	GetResults() []Result
	GetErrors(flush bool) []error
	FlushErrors()
	DynamicScheduler[Result]
	StaticScheduler[Result]
}

// workQueue represents the instance of a queue of jobs.
type workQueue[Result JobResult] struct {
	// General properties.
	nWorkers       int
	errorsList     []error
	errorsListLock sync.Mutex

	jobResults     []Result
	jobResultsLock sync.Mutex

	// Static scheduler properties.
	staticJobQueue []Job[Result]

	// Dynamic scheduler properties.
	dynamicJobQueue      chan Job[Result]
	dynamicJobQueueLock  sync.Mutex
	shutdownRequired     bool
	shutdownRequiredLock sync.Mutex
	shutdownChan         chan bool
}

// New creates a new WorkQueue instance to schedule jobs.
func New[Result JobResult](workers int) Queue[Result] {
	return &workQueue[Result]{
		nWorkers: workers,
	}
}

// GetErrors returns the list of errors that occurred during job execution.
// If flush argument is set to true the internal error list gets flushed.
func (w *workQueue[Result]) GetErrors(flush bool) []error {
	w.errorsListLock.Lock()
	list := w.errorsList[:]
	w.errorsListLock.Unlock()

	if flush {
		w.FlushErrors()
	}
	return list
}

// GetResult returns the list of results that have been returned during job execution.
func (w *workQueue[Result]) GetResults() []Result {
	w.jobResultsLock.Lock()
	list := w.jobResults[:]
	w.jobResultsLock.Unlock()
	return list
}

// FlushErrors can be used to clean error list.
func (w *workQueue[Result]) FlushErrors() {
	w.errorsListLock.Lock()
	defer w.errorsListLock.Unlock()
	w.errorsList = nil
}

func (w *workQueue[Result]) appendError(err error) {
	w.errorsListLock.Lock()
	defer w.errorsListLock.Unlock()

	w.errorsList = append(w.errorsList, err)
}

func (w *workQueue[Result]) appendResult(result Result) {
	w.jobResultsLock.Lock()
	defer w.jobResultsLock.Unlock()
	if w.jobResults == nil {
		w.jobResults = make([]Result, 0)
	}
	w.jobResults = append(w.jobResults, result)
}
