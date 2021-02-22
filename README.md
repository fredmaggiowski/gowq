# goWQ

[![Build](https://github.com/fredmaggiowski/gowq/actions/workflows/go.yml/badge.svg)](https://github.com/fredmaggiowski/gowq/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/fredmaggiowski/gowq)](https://goreportcard.com/report/github.com/fredmaggiowski/gowq)

A simple static work queue implementation to schedule jobs and then execute them on a defined pool of goroutines.

## Usage

```go
wq := NewWQ(2)
wq.Schedule(func(ctx context.Context) error {
    // do something...
    return nil
})
errors := wq.RunAll(context.TODO())
```
