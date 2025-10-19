package sse

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestAsyncTaskPoolBasic(t *testing.T) {
	pool := NewAsyncTaskPool(3) // 3 workers
	defer pool.Stop()

	var count int32
	tasks := 100

	for i := 0; i < tasks; i++ {
		pool.Submit(func() {
			atomic.AddInt32(&count, 1)
			time.Sleep(10 * time.Millisecond)
		})
	}

	pool.Wait() // should block until all tasks are done

	if count != int32(tasks) {
		t.Errorf("expected %d tasks to complete, got %d", tasks, count)
	}
}

func TestAsyncTaskPoolStop(t *testing.T) {
	pool := NewAsyncTaskPool(2)

	done := make(chan struct{})
	pool.Submit(func() {
		time.Sleep(50 * time.Millisecond)
		close(done)
	})

	pool.Stop() // stop early
	select {
	case <-done:
		// ok
	case <-time.After(100 * time.Millisecond):
		t.Error("expected task to complete before timeout")
	}
}
