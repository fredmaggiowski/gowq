# goWQ

[![Build][build-svg]][build-link]
[![Go Report Card][report-card-svg]][report-card-link]

A simple work queue manager to schedule jobs and then execute them on a defined pool of goroutines.

It can be used in two modes:

- *static*: allows to create a job queue and then run it waiting for all jobs to complete;
- *dynamic*: allows to run a job scheduler and then enqueue new jobs while the scheduler runs.

## Examples

### Static Queue Usage

```go
wq := NewWQ[MyResult](2)

wq.Push(func(ctx context.Context) (MyResult, error) {
    // do something...
    return MyResult{}, nil
})
errors := wq.RunAll(context.TODO())
```

### Dynamic Queue Manager

```go
wq := NewWQ[MyResult](2)

go func(ctx context.Context) {
    wq.Start(ctx)
}(context.TODO())

wq.Schedule(func(ctx context.Context) (MyResult, error) {
    // do something...
    return nil
})

// Wait until all jobs have been completed.
_ := wq.Shutdown()
```

[build-svg]: https://github.com/fredmaggiowski/gowq/actions/workflows/go.yml/badge.svg
[build-link]: https://github.com/fredmaggiowski/gowq/actions/workflows/go.yml
[report-card-svg]: https://goreportcard.com/badge/github.com/fredmaggiowski/gowq
[report-card-link]: https://goreportcard.com/report/github.com/fredmaggiowski/gowq
