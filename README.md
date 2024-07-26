# goWQ

[![Build](https://github.com/fredmaggiowski/gowq/actions/workflows/go.yml/badge.svg)](https://github.com/fredmaggiowski/gowq/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/fredmaggiowski/gowq)](https://goreportcard.com/report/github.com/fredmaggiowski/gowq)

A simple work queue manager to schedule jobs and then execute them on a defined pool of goroutines.

It can be used in two modes: 

- *static*: allows to create a job queue and then run it waiting for all jobs to complete;
- *dynamic*: allows to run a job scheduler and then enqueue new jobs while the scheduler runs.

## Examples

### Static Queue Usage

```go
wq := NewWQ(2)

wq.Push(func(ctx context.Context) error {
    // do something...
    return nil
})
errors := wq.RunAll(context.TODO())
```

### Dynamic Queue Manager

```go
wq := NewWQ(2)

go func(ctx context.Context) {
    wq.Start(ctx)
}(context.TODO())

wq.Schedule(func(ctx context.Context) error {
    // do something...
    return nil
})

// Wait until all jobs have been completed.
_ := wq.Shutdown()
```
