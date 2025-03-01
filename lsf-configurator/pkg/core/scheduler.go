package core

import (
	"context"
	"fmt"
	"sync"
)

type Task func() (interface{}, error)

type Result struct {
	Value interface{}
	Err   error
}

type WorkerPool struct {
	taskQueue chan Task
	wg        sync.WaitGroup
	once      sync.Once
	ctx       context.Context
	cancel    context.CancelFunc
}

type Scheduler interface {
	AddTask(task Task) <-chan Result
	Close()
}

func NewScheduler(numWorkers int, queueSize int) Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	w := WorkerPool{
		taskQueue: make(chan Task, queueSize),
		ctx:       ctx,
		cancel:    cancel,
	}

	for i := 0; i < numWorkers; i++ {
		w.wg.Add(1)
		go w.worker()
	}

	return &w
}

func (w *WorkerPool) AddTask(task Task) <-chan Result {
	resultChan := make(chan Result, 1)
	go func() {
		select {
		case <-w.ctx.Done():
			resultChan <- Result{Err: fmt.Errorf("thread pool shutting down")}
			close(resultChan)
		case w.taskQueue <- func() (interface{}, error) {
			defer close(resultChan)
			result, err := task()
			resultChan <- Result{Value: result, Err: err}
			return result, err
		}:
		}
	}()
	return resultChan
}

func (w *WorkerPool) Close() {
	w.once.Do(func() {
		w.cancel()
		close(w.taskQueue)
	})
	w.wg.Wait()
}

func (w *WorkerPool) worker() {
	defer w.wg.Done()

	for {
		select {
		case job, ok := <-w.taskQueue:
			if !ok {
				return
			}
			job()
		case <-w.ctx.Done():
			return
		}
	}
}
