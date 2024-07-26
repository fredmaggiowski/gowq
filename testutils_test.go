package gowq

import "context"

func fakeJob(ctx context.Context) (*FakeResult, error) {
	return nil, nil
}

func simulateStart(wq *workQueue[*FakeResult], size int) {
	wq.dynamicJobQueue = make(chan Job[*FakeResult], size)
}

type FakeResult struct {
	id int
}
