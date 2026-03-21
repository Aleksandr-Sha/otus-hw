package hw05parallelexecution

import (
	"errors"
	"fmt"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestRun(t *testing.T) {
	defer goleak.VerifyNone(t)

	t.Run("if were errors in first M tasks, than finished not more N+M tasks when all task with error",
		func(t *testing.T) {
			tasksCount := 50
			tasks := make([]Task, 0, tasksCount)

			var runTasksCount int32

			for i := 0; i < tasksCount; i++ {
				taskSleep := time.Millisecond * time.Duration(rand.Intn(100))
				tasks = append(tasks, createTaskError(&runTasksCount, i, taskSleep))
			}

			workersCount := 10
			maxErrorsCount := 23
			err := Run(tasks, workersCount, maxErrorsCount)

			require.Truef(t, errors.Is(err, ErrErrorsLimitExceeded), "actual err - %v", err)
			require.LessOrEqual(t, runTasksCount, int32(workersCount+maxErrorsCount), "extra tasks were started")
		})

	t.Run("tasks without errors", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32
		var sumTime time.Duration

		for i := 0; i < tasksCount; i++ {
			taskSleep := time.Millisecond * time.Duration(rand.Intn(100))
			sumTime += taskSleep

			tasks = append(tasks, createTaskSuccess(&runTasksCount, taskSleep))
		}

		workersCount := 5
		maxErrorsCount := 1

		start := time.Now()
		err := Run(tasks, workersCount, maxErrorsCount)
		elapsedTime := time.Since(start)
		require.NoError(t, err)

		require.Equal(t, int32(tasksCount), runTasksCount, "not all tasks were completed")
		require.LessOrEqual(t, int64(elapsedTime), int64(sumTime/2), "tasks were run sequentially?")
	})

	t.Run( // todo ошибки в середине процесса
		"if were errors in first M tasks, than finished not more N+M tasks when not all tasks with errors",
		func(t *testing.T) {
			tasksCount := 30
			tasks := make([]Task, 0, tasksCount)

			var runTasksCount int32
			successTaskBeforeErrorCount := 10

			// Задачи, которые должны выполниться успешно в начале
			for i := 0; i < successTaskBeforeErrorCount; i++ {
				tasks = append(tasks, createTaskSuccess(&runTasksCount, time.Millisecond*50))
			}

			var maxErrorsCount int
			indexLastError := 12

			for i := successTaskBeforeErrorCount; i < indexLastError; i++ {
				maxErrorsCount++
				tasks = append(tasks, createTaskError(&runTasksCount, i, time.Millisecond*50))
			}

			for i := indexLastError; i < tasksCount; i++ {
				tasks = append(tasks, createTaskSuccess(&runTasksCount, time.Millisecond*50))
			}

			workersCount := 4

			err := Run(tasks, workersCount, maxErrorsCount)

			require.Truef(t, errors.Is(err, ErrErrorsLimitExceeded), "actual err - %v", err)
			require.Greater(t, runTasksCount, int32(maxErrorsCount+workersCount), "")
			require.Less(t, runTasksCount, int32(tasksCount), "")
		})

	t.Run( // todo выполняются все таски но всё равно ошибка так как последние таски будут с error
		"",
		func(t *testing.T) {
			tasksCount := 30
			tasks := make([]Task, 0, tasksCount)

			var runTasksCount int32
			successTaskCount := 28

			// Задачи, которые должны выполниться успешно в начале
			for i := 0; i < successTaskCount; i++ {
				tasks = append(tasks, createTaskSuccess(&runTasksCount, time.Millisecond*50))
			}

			var maxErrorsCount int

			for i := successTaskCount; i < tasksCount; i++ {
				maxErrorsCount++
				tasks = append(tasks, createTaskError(&runTasksCount, i, time.Millisecond*100))
			}

			workersCount := 4

			err := Run(tasks, workersCount, maxErrorsCount)

			require.Truef(t, errors.Is(err, ErrErrorsLimitExceeded), "actual err - %v", err)
			require.Equal(t, runTasksCount, int32(tasksCount), "")
		})

	t.Run( // TODO Колицество ошибок не привышает лимит
		"",
		func(t *testing.T) {
			tasksCount := 30
			tasks := make([]Task, 0, tasksCount)

			var runTasksCount int32

			// Задачи, которые должны выполниться успешно в начале
			for i := 0; i < 10; i++ {
				tasks = append(tasks, createTaskSuccess(&runTasksCount, time.Millisecond*50))
			}

			for i := 10; i < 12; i++ {
				tasks = append(tasks, createTaskError(&runTasksCount, i, time.Millisecond*50))
			}

			for i := 12; i < tasksCount; i++ {
				tasks = append(tasks, createTaskSuccess(&runTasksCount, time.Millisecond*50))
			}

			workersCount := 4
			maxErrorsCount := 3

			err := Run(tasks, workersCount, maxErrorsCount)

			require.NoError(t, err)
			require.Equal(t, runTasksCount, int32(tasksCount), "")
		})
}

func createTaskSuccess(runTasksCount *int32, d time.Duration) func() error {
	return func() error {
		time.Sleep(d)
		atomic.AddInt32(runTasksCount, 1)
		return nil
	}
}

func createTaskError(runTasksCount *int32, i int, d time.Duration) func() error {
	err := fmt.Errorf("error from task %d", i)
	return func() error {
		time.Sleep(d)
		atomic.AddInt32(runTasksCount, 1)
		return err
	}
}
