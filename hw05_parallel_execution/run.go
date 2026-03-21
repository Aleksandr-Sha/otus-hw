package hw05parallelexecution

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task func() error

type TaskExecutor struct {
	errorCount    int64
	maxErrorCount int
	workersCount  int

	chTask           chan Task
	workersWaitGroup *sync.WaitGroup

	chError               chan struct{}
	errorHandlerWaitGroup *sync.WaitGroup
}

func (e *TaskExecutor) run(tasks []Task) error {
	e.initChan()

	e.createWorkers()
	e.createErrorHandler()

	err := e.loadTasks(tasks)
	if err != nil {
		e.finishProcess()
		return fmt.Errorf("load tasks: %w", err)
	}

	e.finishProcess()

	if atomic.LoadInt64(&e.errorCount) >= int64(e.maxErrorCount) {
		return ErrErrorsLimitExceeded
	}

	return nil
}

func (e *TaskExecutor) initChan() {
	e.chTask = make(chan Task)
	e.chError = make(chan struct{})
}

func (e *TaskExecutor) createWorkers() {
	e.workersWaitGroup = &sync.WaitGroup{}

	for i := 0; i < e.workersCount; i++ {
		e.workersWaitGroup.Add(1)
		go func() {
			defer e.workersWaitGroup.Done()
			for {
				task, ok := <-e.chTask
				if !ok {
					return
				}

				err := task()
				if err != nil {
					e.chError <- struct{}{}
				}
			}
		}()
	}
}

func (e *TaskExecutor) createErrorHandler() {
	e.errorHandlerWaitGroup = &sync.WaitGroup{}

	e.errorHandlerWaitGroup.Add(1)
	go func() {
		defer e.errorHandlerWaitGroup.Done()
		for {
			_, ok := <-e.chError
			if !ok {
				return
			}

			atomic.AddInt64(&e.errorCount, 1)
		}
	}()
}

func (e *TaskExecutor) loadTasks(tasks []Task) error {
	for i := 0; i < len(tasks); {
		if atomic.LoadInt64(&e.errorCount) >= int64(e.maxErrorCount) {
			return ErrErrorsLimitExceeded
		}

		select {
		case e.chTask <- tasks[i]:
			i++
		case <-e.chError:
			atomic.AddInt64(&e.errorCount, 1)
		}
	}

	return nil
}

func (e *TaskExecutor) finishProcess() {
	close(e.chTask)
	e.workersWaitGroup.Wait()

	close(e.chError)
	e.errorHandlerWaitGroup.Wait()
}

func newTaskExecutor(maxErrorCount, workersCount int) *TaskExecutor {
	return &TaskExecutor{maxErrorCount: maxErrorCount, workersCount: workersCount}
}

func Run(tasks []Task, n, m int) error {
	executor := newTaskExecutor(m, n)

	err := executor.run(tasks)
	if err != nil {
		return err
	}

	return nil
}
