package gowq

import "context"

func fakeJob(ctx context.Context) error {
	return nil
}

func simulateStart(wq *workQueue, size int) {
	wq.dynamicJobQueue = make(chan Job, size)
}
