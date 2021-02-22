package gowq

import "context"

func fakeJob(ctx context.Context) error {
	return nil
}

func simulateStart(wq *WorkQueue, size int) {
	wq.dynamicJobQueue = make(chan Job, size)
}
