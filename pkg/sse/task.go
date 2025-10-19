package sse

import "sync"

type AsyncTaskPool struct {
	tasks chan func()
	wg    sync.WaitGroup
}

// NewAsyncTaskPool creates a task pool with a fixed capacity
func NewAsyncTaskPool(maxWorkers int) *AsyncTaskPool {
	pool := &AsyncTaskPool{
		tasks: make(chan func(), 10000), // default capacity is 10000
	}
	for i := 0; i < maxWorkers; i++ {
		go pool.worker()
	}
	return pool
}

func (p *AsyncTaskPool) worker() {
	for task := range p.tasks {
		task()
		p.wg.Done()
	}
}

func (p *AsyncTaskPool) Submit(task func()) {
	p.wg.Add(1)
	p.tasks <- task
}

func (p *AsyncTaskPool) Wait() {
	p.wg.Wait()
}

func (p *AsyncTaskPool) Stop() {
	close(p.tasks)
}
